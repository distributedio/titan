package command

import (
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/shafreeck/retry"
	"github.com/meitu/titan/context"
	"github.com/meitu/titan/db"
	"github.com/meitu/titan/encoding/resp"
	"github.com/meitu/titan/metrics"
	"go.uber.org/zap"
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

// Command is a redis command implementation
type Command func(ctx *Context)

// Command2 is a redis command implementation and return error, for flushdb and flushall
type Command2 func(ctx *Context) error

// OnCommit returns by TxnCommand and will be called after a transaction being committed
type OnCommit func()

// SimpleString replies a simplestring when commit
func SimpleString(w io.Writer, s string) OnCommit {
	return func() {
		resp.ReplySimpleString(w, s)
	}
}

// BulkString replies a bulkstring when commit
func BulkString(w io.Writer, s string) OnCommit {
	return func() {
		resp.ReplyBulkString(w, s)
	}
}

// NullBulkString replies a null bulkstring when commit
func NullBulkString(w io.Writer) OnCommit {
	return func() {
		resp.ReplyNullBulkString(w)
	}
}

// Integer replies in integer when commit
func Integer(w io.Writer, v int64) OnCommit {
	return func() {
		resp.ReplyInteger(w, v)
	}
}

// BytesArray replies a [][]byte when commit
func BytesArray(w io.Writer, a [][]byte) OnCommit {
	return func() {
		start := time.Now()
		resp.ReplyArray(w, len(a))
		zap.L().Debug("reply array size", zap.Int64("cost(us)", time.Since(start).Nanoseconds()/1000))
		start = time.Now()
		for i := range a {
			if a[i] == nil {
				resp.ReplyNullBulkString(w)
				continue
			}
			resp.ReplyBulkString(w, string(a[i]))
			if i % 10 == 9 {
				zap.L().Debug("reply 10 bulk string", zap.Int64("cost(us)", time.Since(start).Nanoseconds()/1000))
				start = time.Now()
			}
		}
	}
}

func BytesArrayOnce(w io.Writer, a [][]byte) OnCommit {
    return func() {
        resp.ReplyStringArray(w, a)
    }
}

// TxnCommand runs a command in transaction
type TxnCommand func(ctx *Context, txn *db.Transaction) (OnCommit, error)

// Call a command
func Call(ctx *Context) {
	ctx.Name = strings.ToLower(ctx.Name)

	if ctx.Name != "auth" &&
		ctx.Server.RequirePass != "" &&
		ctx.Client.Authenticated == false {
		resp.ReplyError(ctx.Out, ErrNoAuth.Error())
		return
	}
	// Exec all queued commands if this is an exec command
	if ctx.Name == "exec" {
		if len(ctx.Args) != 0 {
			resp.ReplyError(ctx.Out, ErrWrongArgs(ctx.Name).Error())
			return
		}
		// Exec must begin with multi
		if !ctx.Client.Multi {
			resp.ReplyError(ctx.Out, ErrExec.Error())
			return
		}

		Exec(ctx)
		feedMonitors(ctx)
		return
	}
	// Discard all queued commands and return
	if ctx.Name == "discard" {
		if !ctx.Client.Multi {
			resp.ReplyError(ctx.Out, ErrDiscard.Error())
			return
		}

		Discard(ctx)
		feedMonitors(ctx)
		return
	}

	cmdInfoCommand, ok := commands[ctx.Name]
	if !ok {
		resp.ReplyError(ctx.Out, ErrUnKnownCommand(ctx.Name).Error())
		return
	}
	argc := len(ctx.Args) + 1 // include the command name
	arity := cmdInfoCommand.Cons.Arity

	if arity > 0 && argc != arity {
		resp.ReplyError(ctx.Out, ErrWrongArgs(ctx.Name).Error())
		return
	}

	if arity < 0 && argc < -arity {
		resp.ReplyError(ctx.Out, ErrWrongArgs(ctx.Name).Error())
		return
	}

	// We now in a multi block, queue the command and return
	if ctx.Client.Multi {
		if ctx.Name == "multi" {
			resp.ReplyError(ctx.Out, ErrMultiNested.Error())
			return
		}
		commands := ctx.Client.Commands
		commands = append(commands, &context.Command{Name: ctx.Name, Args: ctx.Args})
		ctx.Client.Commands = commands
		resp.ReplySimpleString(ctx.Out, "QUEUED")
		return
	}

	feedMonitors(ctx)
	start := time.Now()
	//wangzongsheng add below
	if ctx.Name == "flushdb" || ctx.Name == "flushall" {
		for {
			err := cmdInfoCommand.Proc2(ctx)
			if err != nil && err.Error() == db.ERR_MAX_FLUSH_COUNT {
				continue
			} else {
				break
			}
		}
	} else {
		cmdInfoCommand.Proc(ctx)
	}
	//////
	cost := time.Since(start)

	cmdInfoCommand.Stat.Calls++
	cmdInfoCommand.Stat.Microseconds += cost.Nanoseconds() / int64(1000)
}

// TxnCall calls a command with transaction, it is used with multi/exec
func TxnCall(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	name := strings.ToLower(ctx.Name)
	cmd, ok := txnCommands[name]
	if !ok {
		return nil, ErrUnKnownCommand(ctx.Name)
	}
	feedMonitors(ctx)
	return cmd(ctx, txn)
}

// AutoCommit commits to database after run a txn command
func AutoCommit(cmd TxnCommand) Command {
	return func(ctx *Context) {
		retry.Ensure(ctx, func() error {
			mt := metrics.GetMetrics()
            start := time.Now()
			txn, err := ctx.Client.DB.Begin()
			key := ""
			if len(ctx.Args) > 0 {
				key = ctx.Args[0]
			}
            zap.L().Debug("transation begin", zap.String("name", ctx.Name), zap.String("name", key), zap.Int64("cost(us)", time.Since(start).Nanoseconds()/1000))
			if err != nil {
				mt.TxnFailuresCounterVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Inc()
				resp.ReplyError(ctx.Out, "ERR "+err.Error())
				zap.L().Error("txn begin failed",
					zap.Int64("clientid", ctx.Client.ID),
					zap.String("command", ctx.Name),
					zap.String("traceid", ctx.TraceID),
					zap.Error(err))
				return err
			}

			start = time.Now()
			onCommit, err := cmd(ctx, txn)
			zap.L().Debug("command done", zap.String("name", ctx.Name), zap.Int64("cost(us)", time.Since(start).Nanoseconds()/1000))
			if err != nil {
				mt.TxnFailuresCounterVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Inc()
				resp.ReplyError(ctx.Out, err.Error())
				txn.Rollback()
				zap.L().Error("command process failed",
					zap.Int64("clientid", ctx.Client.ID),
					zap.String("command", ctx.Name),
					zap.String("traceid", ctx.TraceID),
					zap.Error(err))
				return err
			}

			start = time.Now()
			mtFunc := func() {
				cost := time.Since(start).Seconds()
				mt.TxnCommitHistogramVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Observe(cost)
			}
			if err := txn.Commit(ctx); err != nil {
				txn.Rollback()
				if db.IsConflictError(err) {
					mt.TxnConflictsCounterVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Inc()
				}
				if db.IsRetryableError(err) {
					mt.TxnRetriesCounterVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Inc()
					mtFunc()
					zap.L().Error("txn commit retry",
						zap.Int64("clientid", ctx.Client.ID),
						zap.String("command", ctx.Name),
						zap.String("traceid", ctx.TraceID),
						zap.Error(err))
					return retry.Retriable(err)
				}
				mt.TxnFailuresCounterVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Inc()
				resp.ReplyError(ctx.Out, "ERR "+err.Error())
				mtFunc()
				zap.L().Error("txn commit failed",
					zap.Int64("clientid", ctx.Client.ID),
					zap.String("command", ctx.Name),
					zap.String("traceid", ctx.TraceID),
					zap.Error(err))
				return err
			}
			zap.L().Debug("commit ", zap.String("name", ctx.Name), zap.Int64("cost(us)", time.Since(start).Nanoseconds()/1000))

			start = time.Now()
			if onCommit != nil {
				onCommit()
			}
			zap.L().Debug("onCommit ", zap.String("name", ctx.Name), zap.Int64("cost(us)", time.Since(start).Nanoseconds()/1000))
			mtFunc()
			return nil
		})
	}
}

// AutoCommit2 commits to database after run a txn command
func AutoCommit2(cmd TxnCommand) Command2 {
	return func(ctx *Context) error {
		return retry.Ensure(ctx, func() error {
			mt := metrics.GetMetrics()
			txn, err := ctx.Client.DB.Begin()
			if err != nil {
				mt.TxnFailuresCounterVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Inc()
				resp.ReplyError(ctx.Out, "ERR "+err.Error())
				zap.L().Error("txn begin failed",
					zap.Int64("clientid", ctx.Client.ID),
					zap.String("command", ctx.Name),
					zap.String("traceid", ctx.TraceID),
					zap.Error(err))
				return err
			}

			onCommit, err := cmd(ctx, txn)
			if err != nil {
				mt.TxnFailuresCounterVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Inc()
				resp.ReplyError(ctx.Out, err.Error())
				txn.Rollback()
				zap.L().Error("command process failed",
					zap.Int64("clientid", ctx.Client.ID),
					zap.String("command", ctx.Name),
					zap.String("traceid", ctx.TraceID),
					zap.Error(err))
				return err
			}

			start := time.Now()
			mtFunc := func() {
				cost := time.Since(start).Seconds()
				mt.TxnCommitHistogramVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Observe(cost)
			}
			if err := txn.Commit(ctx); err != nil {
				txn.Rollback()
				if db.IsConflictError(err) {
					mt.TxnConflictsCounterVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Inc()
				}
				if db.IsRetryableError(err) {
					mt.TxnRetriesCounterVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Inc()
					mtFunc()
					zap.L().Error("txn commit retry",
						zap.Int64("clientid", ctx.Client.ID),
						zap.String("command", ctx.Name),
						zap.String("traceid", ctx.TraceID),
						zap.Error(err))
					return retry.Retriable(err)
				}
				mt.TxnFailuresCounterVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Inc()
				resp.ReplyError(ctx.Out, "ERR "+err.Error())
				mtFunc()
				zap.L().Error("txn commit failed",
					zap.Int64("clientid", ctx.Client.ID),
					zap.String("command", ctx.Name),
					zap.String("traceid", ctx.TraceID),
					zap.Error(err))
				return err
			}

			if onCommit != nil {
				onCommit()
			}
			mtFunc()
			return nil
		})
	}
}

func feedMonitors(ctx *Context) {
	ctx.Server.Monitors.Range(func(k, v interface{}) bool {
		mCtx := v.(*Context)
		if mCtx.Client.Namespace != sysAdminNamespace && mCtx.Client.Namespace != ctx.Client.Namespace {
			return true
		}

		now := time.Now().UnixNano() / 1000
		ts := strconv.FormatFloat(float64(now)/1000000, 'f', -1, 64)
		id := strconv.FormatInt(int64(ctx.Client.DB.ID), 10)

		line := ts + " [" + id + " " + ctx.Client.RemoteAddr + "]" + " " + ctx.Name + " " + strings.Join(ctx.Args, " ")
        start := time.Now()
		err := resp.ReplySimpleString(mCtx.Out, line)
        zap.L().Debug("feedMonitors reply", zap.String("name", ctx.Name), zap.Int64("cost(us)", time.Since(start).Nanoseconds()/1000))
		if err != nil {
			ctx.Server.Monitors.Delete(k)
		}

		return true
	})
}

// Executor executes a command
type Executor struct {
	txnCommands map[string]TxnCommand
	commands    map[string]Desc
}

// NewExecutor news a Executor
func NewExecutor() *Executor {
	return &Executor{txnCommands: txnCommands, commands: commands}
}

// Execute a command
func (e *Executor) Execute(ctx *Context) {
	start := time.Now()
	Call(ctx)
	cost := time.Since(start).Seconds()
	metrics.GetMetrics().CommandCallHistogramVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Observe(cost)
}

// Desc describes a command with constraints
type Desc struct {
	Proc  Command
	Proc2 Command2
	Stat  Statistic
	Cons  Constraint
}
