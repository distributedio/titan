package titan

import (
	"bufio"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/distributedio/titan/command"
	"github.com/distributedio/titan/context"
	"github.com/distributedio/titan/encoding/resp"
	"go.uber.org/zap"
)

type client struct {
	cliCtx *context.ClientContext
	server *Server
	conn   net.Conn
	exec   *command.Executor
	r      *bufio.Reader

	eofLock sync.Mutex //the lock of reading_writing 'eof'
	eof     bool       //is over when read data from socket
}

func newClient(cliCtx *context.ClientContext, s *Server, exec *command.Executor) *client {
	return &client{
		cliCtx: cliCtx,
		server: s,
		exec:   exec,
		eof:    false,
	}
}

func (c *client) readEof() {
	c.eofLock.Lock()
	defer c.eofLock.Unlock()

	c.eof = true
}

func (c *client) isEof() bool {
	c.eofLock.Lock()
	defer c.eofLock.Unlock()

	return c.eof
}

// Write to conn and log error if needed
func (c *client) Write(p []byte) (int, error) {
	zap.L().Debug("write to client", zap.Int64("clientid", c.cliCtx.ID), zap.String("msg", string(p)))
	n, err := c.conn.Write(p)
	if err != nil {
		c.conn.Close()
		if err == io.EOF {
			zap.L().Info("close connection", zap.String("addr", c.cliCtx.RemoteAddr),
				zap.Int64("clientid", c.cliCtx.ID))
		} else {
			zap.L().Error("write net failed", zap.String("addr", c.cliCtx.RemoteAddr),
				zap.Int64("clientid", c.cliCtx.ID),
				zap.String("namespace", c.cliCtx.Namespace),
				zap.Bool("multi", c.cliCtx.Multi),
				zap.Bool("watching", c.cliCtx.Txn != nil),
				zap.String("command", c.cliCtx.LastCmd),
				zap.String("error", err.Error()))
			return 0, err
		}
	}
	return n, nil
}

func (c *client) serve(conn net.Conn) error {
	c.conn = conn
	c.r = bufio.NewReader(conn)

	var cmd []string
	var err error
	for {
		select {
		case <-c.cliCtx.Done:
			return c.conn.Close()
		default:
			cmd, err = c.readCommand()
			if err != nil {
				c.conn.Close()
				if err == io.EOF {
					zap.L().Info("close connection", zap.String("addr", c.cliCtx.RemoteAddr),
						zap.Int64("clientid", c.cliCtx.ID))
					return nil
				}
				zap.L().Error("read command failed", zap.String("addr", c.cliCtx.RemoteAddr),
					zap.Int64("clientid", c.cliCtx.ID), zap.Error(err))
				return err
			}
		}

		if c.server.servCtx.Pause > 0 {
			time.Sleep(c.server.servCtx.Pause)
			c.server.servCtx.Pause = 0
		}

		if len(cmd) <= 0 {
			err := command.ErrEmptyCommand
			zap.L().Error(err.Error(), zap.String("addr", c.cliCtx.RemoteAddr),
				zap.Int64("clientid", c.cliCtx.ID))
			resp.ReplyError(c, err.Error())
			c.conn.Close()
			return nil
		}

		c.cliCtx.Updated = time.Now()
		c.cliCtx.LastCmd = cmd[0]
		if !c.exec.CanExecute(c.cliCtx.LastCmd) {
			err := command.ErrUnKnownCommand(c.cliCtx.LastCmd)
			zap.L().Error(err.Error(), zap.String("addr", c.cliCtx.RemoteAddr),
				zap.Int64("clientid", c.cliCtx.ID))
			resp.ReplyError(c, err.Error())
			continue
		}

		ctx := &command.Context{
			Name:    cmd[0],
			Args:    cmd[1:],
			In:      c.r,
			Out:     c,
			TraceID: GenerateTraceID(),
		}

		ctx.Context = context.New(c.cliCtx, c.server.servCtx)
		zap.L().Debug("recv msg", zap.String("command", ctx.Name), zap.Strings("arguments", ctx.Args))

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
