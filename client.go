package titan

import (
	"bufio"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"github.com/meitu/titan/command"
	"github.com/meitu/titan/context"
	"github.com/meitu/titan/encoding/resp"
	"go.uber.org/zap"
)

type client struct {
	cliCtx *context.ClientContext
	server *Server
	conn   net.Conn
	exec   *command.Executor
	r      *bufio.Reader
}

func newClient(cliCtx *context.ClientContext, s *Server, exec *command.Executor) *client {
	return &client{
		cliCtx: cliCtx,
		server: s,
		exec:   exec,
	}
}

// Write to conn and log error if needed
func (c *client) Write(p []byte) (int, error) {
	n, err := c.conn.Write(p)
	if err != nil {
		zap.L().Error("write net failed", zap.String("addr", c.cliCtx.RemoteAddr),
			zap.Int64("clientid", c.cliCtx.ID),
			zap.String("namespace", c.cliCtx.Namespace),
			zap.Bool("multi", c.cliCtx.Multi),
			zap.Bool("watching", c.cliCtx.Txn != nil),
			zap.String("command", c.cliCtx.LastCmd))
		c.conn.Close()
	}
	return n, err
}

func (c *client) serve(conn net.Conn) error {
	c.conn = conn
	c.r = bufio.NewReader(conn)

	rootCtx, rootCancel := context.WithCancel(context.New(c.cliCtx, c.server.servCtx))

	// Use a separate goroutine to keep reading commands
	// then we can detect a closed connection as soon as possible.
	// It only works when the cmd channel is not blocked
	cmdc := make(chan []string, 128)
	errc := make(chan error)
	go func() {
		for {
			cmd, err := c.readCommand()
			if err != nil {
				errc <- err
				rootCancel()
				return
			}
			cmdc <- cmd
		}
	}()

	var cmd []string
	var err error
	for {
		select {
		case <-c.cliCtx.Done:
			return c.conn.Close()
		case cmd = <-cmdc:
		case err = <-errc:
			zap.L().Error("read command failed", zap.String("addr", c.cliCtx.RemoteAddr),
				zap.Int64("clientid", c.cliCtx.ID), zap.Error(err))
			c.conn.Close()
			return err
		}

		if c.server.servCtx.Pause > 0 {
			time.Sleep(c.server.servCtx.Pause)
			c.server.servCtx.Pause = 0
		}

		c.cliCtx.Updated = time.Now()
		c.cliCtx.LastCmd = cmd[0]

		ctx := &command.Context{
			Name:    cmd[0],
			Args:    cmd[1:],
			In:      c.r,
			Out:     c,
			TraceID: GenerateTraceID(),
		}
		ctx.Context = rootCtx

		// Skip reply if necessary
		if c.cliCtx.SkipN != 0 {
			ctx.Out = ioutil.Discard
			if c.cliCtx.SkipN > 0 {
				c.cliCtx.SkipN--
			}
		}
		if env := zap.L().Check(zap.DebugLevel, "recv client command"); env != nil {
			env.Write(zap.String("addr", c.cliCtx.RemoteAddr),
				zap.Int64("clientid", c.cliCtx.ID),
				zap.String("traceid", ctx.TraceID),
				zap.String("command", ctx.Name))
		}
		c.exec.Execute(ctx)
	}
}

func (c *client) readInlineCommand() ([]string, error) {
	buf, err := c.r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	line := strings.TrimRight(string(buf), "\r\n")
	return strings.Fields(line), nil
}

func (c *client) readCommand() ([]string, error) {
	p, err := c.r.Peek(1)
	if err != nil {
		return nil, err
	}
	// not a bulk string
	if p[0] != '*' {
		return c.readInlineCommand()
	}

	argc, err := resp.ReadArray(c.r)
	if err != nil {
		return nil, err
	}
	if argc == 0 {
		return []string{}, nil
	}

	argv := make([]string, argc)
	for i := 0; i < argc; i++ {
		arg, err := resp.ReadBulkString(c.r)
		if err != nil {
			return nil, err
		}
		argv[i] = arg
	}
	return argv, nil
}
