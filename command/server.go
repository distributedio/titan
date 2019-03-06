package command

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/meitu/titan/context"
	"github.com/meitu/titan/db"
	"github.com/meitu/titan/encoding/resp"
)

const sysAdminNamespace = "$sys.admin"

// Monitor streams back every command processed by the Titan server
func Monitor(ctx *Context) {
	ctx.Server.Monitors.Store(ctx.Client.RemoteAddr, ctx)
	resp.ReplySimpleString(ctx.Out, "OK")
}

// Client manages client connections
func Client(ctx *Context) {
	syntaxErr := "ERR Syntax error, try CLIENT (LIST | KILL | GETNAME | SETNAME | PAUSE | REPLY)"
	list := func(ctx *Context) {
		now := time.Now()
		var lines []string
		clients := &ctx.Server.Clients
		clients.Range(func(k, v interface{}) bool {
			client := v.(*context.ClientContext)
			if ctx.Client.Namespace != sysAdminNamespace && client.Namespace != ctx.Client.Namespace {
				return true
			}
			age := now.Sub(client.Created) / time.Second
			idle := now.Sub(client.Updated) / time.Second
			flags := "N"
			if client.Multi {
				flags = "x"
			}

			// id=2 addr=127.0.0.1:39604 fd=6 name= age=196 idle=2 flags=N db=0 sub=0 psub=0 multi=-1 qbuf=0 qbuf-free=0 obl=0 oll=0 omem=0 events=r cmd=client
			line := fmt.Sprintf("id=%d addr=%s fd=%d name=%s age=%d idle=%d "+
				"flags=%s db=%d sub=%d psub=%d multi=%d qbuf=%d qbuf-free=%d obl=%d oll=%d omem=%d events=%s cmd=%s\n",
				client.ID, client.RemoteAddr, 0, client.Name, age, idle, flags, client.DB.ID, 0, 0, len(client.Commands),
				0, 0, 0, 0, 0, "rw", client.LastCmd)
			lines = append(lines, line)
			return true
		})
		resp.ReplyBulkString(ctx.Out, strings.Join(lines, ""))
	}
	getname := func(ctx *Context) {
		name := ctx.Client.Name
		if len(name) != 0 {
			resp.ReplyBulkString(ctx.Out, name)
			return
		}
		resp.ReplyNullBulkString(ctx.Out)
	}
	setname := func(ctx *Context) {
		args := ctx.Args[1:]
		if len(args) != 1 {
			resp.ReplyError(ctx.Out, syntaxErr)
			return
		}
		ctx.Client.Name = args[0]
		resp.ReplySimpleString(ctx.Out, "OK")
	}
	pause := func(ctx *Context) {
		if ctx.Client.Namespace != sysAdminNamespace {
			resp.ReplyError(ctx.Out, "ERR client pause can be used by $sys.admin only")
			return
		}
		args := ctx.Args[1:]
		if len(args) != 1 {
			resp.ReplyError(ctx.Out, syntaxErr)
			return
		}
		msec, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			resp.ReplyError(ctx.Out, "ERR timeout is not an integer or out of range")
			return
		}
		ctx.Server.Pause = time.Duration(msec) * time.Millisecond
		resp.ReplySimpleString(ctx.Out, "OK")
	}
	reply := func(ctx *Context) {
		args := ctx.Args[1:]
		if len(args) != 1 {
			resp.ReplyError(ctx.Out, syntaxErr)
			return
		}
		switch strings.ToLower(args[0]) {
		case "on":
			ctx.Client.SkipN = 0
			resp.ReplySimpleString(ctx.Out, "OK")
		case "off":
			ctx.Client.SkipN = -1
		case "skip":
			ctx.Client.SkipN = 1
		}
	}
	kill := func(ctx *Context) {
		args := ctx.Args[1:]
		if len(args) < 1 {
			resp.ReplyError(ctx.Out, syntaxErr)
			return
		}
		var addr string
		var id int64
		var ctype string // client type
		var err error
		skipSelf := true
		if len(args) == 1 {
			addr = args[0]
			skipSelf = false // you can kill yourself in old fashion
		} else if len(args)%2 != 0 {
			resp.ReplyError(ctx.Out, syntaxErr)
			return
		}
		for i := 0; i < len(args)-1; i += 2 {
			switch strings.ToLower(string(args[i])) {
			case "addr":
				addr = string(args[i+1])
			case "type":
				ctype = string(args[i+1])
			case "id":
				id, err = strconv.ParseInt(string(args[i+1]), 10, 64)
				if err != nil {
					resp.ReplyError(ctx.Out, syntaxErr)
					return
				}
			case "skipme":
				answer := string(args[i+1])
				if strings.ToLower(answer) == "yes" {
					skipSelf = true
				}
				if strings.ToLower(answer) == "no" {
					skipSelf = false
				}
			}
		}
		// now kill clients with above rules
		killed := 0
		closeSelf := false
		ctx.Server.Clients.Range(func(k, v interface{}) bool {
			cli := v.(*context.ClientContext)

			if cli.Namespace != sysAdminNamespace && cli.Namespace != ctx.Client.Namespace {
				return true
			}

			if id != 0 && cli.ID != id {
				return true
			}
			if addr != "" && cli.RemoteAddr != addr {
				return true
			}
			if ctype != "" && strings.ToLower(ctype) != "normal" {
				return true
			}
			if ctx.Client.ID == cli.ID && skipSelf {
				return true
			}
			if ctx.Client.ID == cli.ID {
				killed++
				closeSelf = true
				return true
			}

			cli.Close()
			killed++
			return true
		})

		if len(args) == 1 {
			if killed == 0 {
				resp.ReplyError(ctx.Out, "ERR No such client")
			} else {
				resp.ReplySimpleString(ctx.Out, "OK")
			}
		} else {
			resp.ReplyInteger(ctx.Out, int64(killed))
		}
		if closeSelf {
			close(ctx.Client.Done)
		}
	}

	args := ctx.Args
	switch strings.ToLower(args[0]) {
	case "list":
		list(ctx)
	case "kill":
		kill(ctx)
	case "getname":
		getname(ctx)
	case "setname":
		setname(ctx)
	case "reply":
		reply(ctx)
	case "pause":
		pause(ctx)
	default:
		resp.ReplyError(ctx.Out, syntaxErr)
	}

}

// Debug the titan server
func Debug(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	switch strings.ToLower(ctx.Args[0]) {
	case "object":
		return debugObject(ctx, txn)
	default:
		return nil, errors.New("ERR not supported")
	}
}
func debugObject(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[1])
	obj, err := txn.Object(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	if obj.Type == db.ObjectHash {
		hash, err := txn.Hash(key)
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		obj, err = hash.Object()
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
	}
	return SimpleString(ctx.Out, obj.String()), nil
}

// RedisCommand returns Array reply of details about all Redis commands
func RedisCommand(ctx *Context) {
	count := func(ctx *Context) {
		resp.ReplyInteger(ctx.Out, int64(len(commands)))
	}
	getkeys := func(ctx *Context) {
		args := ctx.Args[1:]
		if len(args) == 0 {
			resp.ReplyError(ctx.Out, "ERR Unknown subcommand or wrong number of arguments.")
			return
		}
		name := args[0]
		cmdInfo, ok := commands[name]
		if !ok {
			resp.ReplyError(ctx.Out, "ERR Invalid command specified")
			return
		}

		if cmdInfo.Cons.Arity > 0 && len(args) != cmdInfo.Cons.Arity {
			resp.ReplyError(ctx.Out, "ERR Invalid number of arguments specified for command")
			return
		}
		if cmdInfo.Cons.Arity < 0 && len(args) < -cmdInfo.Cons.Arity {
			resp.ReplyError(ctx.Out, "ERR Invalid number of arguments specified for command")
			return
		}
		var keys []string
		last := cmdInfo.Cons.LastKey
		if last < 0 {
			last += len(args)
		}
		for i := cmdInfo.Cons.FirstKey; i <= last; i += cmdInfo.Cons.KeyStep {
			keys = append(keys, args[i])
		}
		resp.ReplyArray(ctx.Out, len(keys))
		for _, key := range keys {
			resp.ReplyBulkString(ctx.Out, key)
		}
	}
	info := func(ctx *Context) {
		names := ctx.Args[1:]
		resp.ReplyArray(ctx.Out, len(names))
		for _, name := range names {
			if cmd, ok := commands[name]; ok {
				resp.ReplyArray(ctx.Out, 6)
				resp.ReplyBulkString(ctx.Out, name)
				resp.ReplyInteger(ctx.Out, int64(cmd.Cons.Arity))

				flags := parseFlags(cmd.Cons.Flags)
				resp.ReplyArray(ctx.Out, len(flags))
				for i := range flags {
					resp.ReplyBulkString(ctx.Out, flags[i])
				}

				resp.ReplyInteger(ctx.Out, int64(cmd.Cons.FirstKey))
				resp.ReplyInteger(ctx.Out, int64(cmd.Cons.LastKey))
				resp.ReplyInteger(ctx.Out, int64(cmd.Cons.KeyStep))
			} else {
				resp.ReplyNullBulkString(ctx.Out)
			}
		}
	}
	list := func(ctx *Context) {
		resp.ReplyArray(ctx.Out, len(commands))
		for name, cmd := range commands {
			resp.ReplyArray(ctx.Out, 6)
			resp.ReplyBulkString(ctx.Out, name)
			resp.ReplyInteger(ctx.Out, int64(cmd.Cons.Arity))

			flags := parseFlags(cmd.Cons.Flags)
			resp.ReplyArray(ctx.Out, len(flags))
			for i := range flags {
				resp.ReplyBulkString(ctx.Out, flags[i])
			}

			resp.ReplyInteger(ctx.Out, int64(cmd.Cons.FirstKey))
			resp.ReplyInteger(ctx.Out, int64(cmd.Cons.LastKey))
			resp.ReplyInteger(ctx.Out, int64(cmd.Cons.KeyStep))
		}

	}
	args := ctx.Args
	if len(args) == 0 {
		list(ctx)
		return
	}
	switch strings.ToLower(args[0]) {
	case "count":
		count(ctx)
	case "getkeys":
		getkeys(ctx)
	case "info":
		info(ctx)
	default:
		resp.ReplyError(ctx.Out, "ERR Unknown subcommand or wrong number of arguments.")
	}
}

// FlushDB clears current db
// This function is **VERY DANGEROUS**. It's not only running on one single region, but it can
// delete a large range that spans over many regions, bypassing the Raft layer.
func FlushDB(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	if err := kv.FlushDB(ctx); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return SimpleString(ctx.Out, "OK"), nil
}

// FlushAll cleans up all databases
// This function is **VERY DANGEROUS**. It's not only running on one single region, but it can
// delete a large range that spans over many regions, bypassing the Raft layer.
func FlushAll(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	if err := kv.FlushAll(ctx); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return SimpleString(ctx.Out, "OK"), nil
}

// Time returns the server time
func Time(ctx *Context) {
	now := time.Now().UnixNano() / int64(time.Microsecond)
	sec := now / 1000000
	msec := now % sec
	resp.ReplyArray(ctx.Out, 2)
	resp.ReplyBulkString(ctx.Out, strconv.Itoa(int(sec)))
	resp.ReplyBulkString(ctx.Out, strconv.Itoa(int(msec)))
}

// Info returns information and statistics about the server in a format that is simple to parse by computers and easy to read by humans
func Info(ctx *Context) {
	exe, err := os.Executable()
	if err != nil {
		resp.ReplyError(ctx.Out, "ERR "+err.Error())
	}

	// count the number of clients
	var numberOfClients int
	ctx.Server.Clients.Range(func(k, v interface{}) bool {
		numberOfClients++
		return true
	})

	var lines []string
	lines = append(lines, "# Server")
	lines = append(lines, "titan_version:"+context.ReleaseVersion)
	lines = append(lines, "titan_git_sha1:"+context.GitHash)
	lines = append(lines, "titan_build_id:"+context.BuildTS)
	lines = append(lines, "os:"+runtime.GOOS)
	lines = append(lines, "arch_bits:"+runtime.GOARCH)
	lines = append(lines, "go_version:"+context.GolangVersion)
	lines = append(lines, "process_id:"+strconv.Itoa(os.Getpid()))
	lines = append(lines, "uptime_in_seconds:"+strconv.FormatInt(int64(time.Since(ctx.Server.StartAt)/time.Second), 10))
	lines = append(lines, "uptime_in_days:"+strconv.FormatInt(int64(time.Since(ctx.Server.StartAt)/time.Second/86400), 10))
	lines = append(lines, "executable:"+exe)

	lines = append(lines, "# Clients")
	lines = append(lines, "connected_clients:"+strconv.Itoa(numberOfClients))
	lines = append(lines, "client_longest_output_list:0")
	lines = append(lines, "client_biggest_input_buf:0")
	lines = append(lines, "blocked_clients:0")
	lines = append(lines, "client_namespace:"+ctx.Client.Namespace)

	resp.ReplyBulkString(ctx.Out, strings.Join(lines, "\n")+"\n")
	return
}
