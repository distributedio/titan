package command

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/shafreeck/retry"
	"gitlab.meitu.com/platform/thanos/context"
	"gitlab.meitu.com/platform/thanos/db"
	"gitlab.meitu.com/platform/thanos/resp"
)

// Context is the runtime context of a command
type Context struct {
	Name    string
	Args    []string
	In      io.Reader
	Out     io.Writer
	TraceID string
	*context.Context
}

const (
	BitMaxOffset = 232
	BitValueZero = 0
	BitValueOne  = 1

	MaxRangeInteger = 2<<29 - 1
)

// Command is a redis command implementation
type Command func(ctx *Context)

// OnCommit return by TxnCommand and will be called after a transaction commit
type OnCommit func()

// SimpleString reply simple string when commit
func SimpleString(w io.Writer, s string) OnCommit {
	return func() {
		resp.ReplySimpleString(w, s)
	}
}

// BulkString reply bulk string when commit
func BulkString(w io.Writer, s string) OnCommit {
	return func() {
		resp.ReplyBulkString(w, s)
	}
}

// NullBulkString reply null bulk string when commit
func NullBulkString(w io.Writer) OnCommit {
	return func() {
		resp.ReplyNullBulkString(w)
	}
}

// Integer reply integer when commit
func Integer(w io.Writer, v int64) OnCommit {
	return func() {
		resp.ReplyInteger(w, v)
	}
}

// BytesArray reply [][]byte when commit
func BytesArray(w io.Writer, a [][]byte) OnCommit {
	return func() {
		resp.ReplyArray(w, len(a))
		for i := range a {
			if a[i] == nil {
				resp.ReplyNullBulkString(w)
				continue
			}
			resp.ReplyBulkString(w, string(a[i]))
		}
	}
}

// TxnCommand runs a command in transaction
type TxnCommand func(ctx *Context, txn *db.Transaction) (OnCommit, error)

// Call a command
func Call(ctx *Context) {
	name := strings.ToLower(ctx.Name)

	if name != "auth" &&
		ctx.Server.RequirePass != "" &&
		ctx.Client.Authenticated == false {
		resp.ReplyError(ctx.Out, "NOAUTH Authentication required.")
		return
	}
	// Exec all queued commands if this is an exec command
	if name == "exec" {
		// Exec must begin with multi
		if !ctx.Client.Multi {
			resp.ReplyError(ctx.Out, "ERR EXEC without MULTI")
			return
		}

		Exec(ctx)
		feedMonitors(ctx)
		return
	}
	// Discard all queued commands and return
	if name == "discard" {
		if !ctx.Client.Multi {
			resp.ReplyError(ctx.Out, "ERR DISCARD without MULTI")
			return
		}

		Discard(ctx)
		feedMonitors(ctx)
		return
	}

	cmdInfo, ok := commands[name]
	if !ok {
		resp.ReplyError(ctx.Out, "ERR unknown command '"+ctx.Name+"'")
		return
	}
	argc := len(ctx.Args) + 1 // include the command name
	arity := cmdInfo.Cons.Arity
	if arity > 0 && argc != arity {
		resp.ReplyError(ctx.Out, "ERR wrong number of arguments for '"+ctx.Name+"' command")
		return
	}

	if arity < 0 && argc < -arity {
		resp.ReplyError(ctx.Out, "ERR wrong number of arguments for '"+ctx.Name+"' command")
		return
	}

	// We now in a multi block, queue the command and return
	if ctx.Client.Multi {
		commands := ctx.Client.Commands
		commands = append(commands, &context.Command{Name: ctx.Name, Args: ctx.Args})
		ctx.Client.Commands = commands
		resp.ReplySimpleString(ctx.Out, "QUEUED")
		return
	}

	feedMonitors(ctx)
	start := time.Now()
	cmdInfo.Proc(ctx)
	cost := time.Since(start)

	cmdInfo.Stat.Calls++
	cmdInfo.Stat.Microseconds += cost.Nanoseconds() / int64(1000)
}

// TxnCall call command with transaction, it is used with multi/exec
func TxnCall(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	name := strings.ToLower(ctx.Name)
	cmd, ok := txnCommands[name]
	if !ok {
		return nil, errors.New("ERR unknown command '" + ctx.Name + "'")
	}
	feedMonitors(ctx)
	return cmd(ctx, txn)
}

// AutoCommit commit to database after run a txn command
func AutoCommit(cmd TxnCommand) Command {
	return func(ctx *Context) {
		retry.Ensure(ctx, func() error {
			txn, err := ctx.Client.DB.Begin()
			if err != nil {
				resp.ReplyError(ctx.Out, "ERR "+string(err.Error()))
				return err
			}

			onCommit, err := cmd(ctx, txn)
			if err != nil {
				resp.ReplyError(ctx.Out, err.Error())
				txn.Rollback()
				return err
			}

			if err := txn.Commit(ctx); err != nil {
				txn.Rollback()
				if db.IsRetryableError(err) {
					return retry.Retriable(err)
				}
				resp.ReplyError(ctx.Out, "ERR "+err.Error())
				return err
			}
			if onCommit != nil {
				onCommit()
			}
			return nil
		})
	}
}

func feedMonitors(ctx *Context) {
	ctx.Server.Monitors.Range(func(k, v interface{}) bool {
		out := v.(io.Writer)

		now := time.Now().UnixNano() / 1000
		ts := strconv.FormatFloat(float64(now)/1000000, 'f', -1, 64)
		id := strconv.FormatInt(int64(ctx.Client.DB.ID), 10)

		line := ts + " [" + id + " " + ctx.Client.RemoteAddr + "]" + " " + ctx.Name + " " + strings.Join(ctx.Args, " ")
		err := resp.ReplySimpleString(out, line)
		if err != nil {
			ctx.Server.Monitors.Delete(k)
		}

		return true
	})
}

// globMatch matches s with pattern in glob-style
func globMatch(s string, pattern string) bool {
	i := 0
	j := 0
	for i < len(s) && j < len(pattern) {
		p := pattern[j]
		switch p {
		case '?':
			i++
			j++
			continue
		case '*':
			j++
			if j == len(pattern) {
				return true
			}
			p = pattern[j]
			for i < len(s) {
				if !globMatch(s[i:], pattern[j:]) {
					i++
					continue
				}
				return true
			}
			if i == len(s) {
				return false
			}
			continue
		case '[':
			exclude := false
			if pattern[j+1] == '^' {
				j++
				exclude = true
			}
			matched := false
			for j < len(pattern) && pattern[j] != ']' {
				p = pattern[j]
				if !matched && s[i] == p {
					matched = true
				}
				j++
			}
			if j == len(pattern) {
				return false
			}
			if matched == exclude {
				return false
			}
			j++
			i++
			continue
		case '\\':
			j++
			fallthrough
		default:
			if s[i] != p {
				return false
			}
			i++
			j++
			continue
		}
	}
	if i < len(s) {
		return false
	}
	return true
}

// Executor execute any command
type Executor struct {
	txnCommands map[string]TxnCommand
	commands    map[string]CommandInfo
}

// NewExecutor new a Executor object
func NewExecutor() *Executor {
	return &Executor{txnCommands: txnCommands, commands: commands}
}

// Execute a command
func (e *Executor) Execute(ctx *Context) {
	Call(ctx)
}

// CommandInfo combines command procedure, constraint and statistics
type CommandInfo struct {
	Proc Command
	Stat Statistic
	Cons Constraint
}
