package thanos

import (
	"bufio"
	"io/ioutil"
	"net"
	"strings"
	"time"

	log "gitlab.meitu.com/gocommons/logbunny"
	"gitlab.meitu.com/platform/thanos/command"
	"gitlab.meitu.com/platform/thanos/context"
	"gitlab.meitu.com/platform/thanos/resp"
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

func (c *client) serve(conn net.Conn) error {
	c.conn = conn
	c.r = bufio.NewReader(conn)

	for {
		select {
		case <-c.cliCtx.Done:
			return c.conn.Close()
		default:
		}
		cmd, err := c.readCommand()
		if err != nil {
			log.Error("read command failed", log.String("addr", c.cliCtx.RemoteAddr),
				log.Int64("clientid", c.cliCtx.ID), log.Err(err))
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
			Out:     c.conn,
			TraceID: GenerateTraceID(),
		}
		innerCtx, cancel := context.WithCancel(context.New(c.cliCtx, c.server.servCtx))
		ctx.Context = innerCtx
		// Skip reply if necessary
		if c.cliCtx.SkipN != 0 {
			ctx.Out = ioutil.Discard
			if c.cliCtx.SkipN > 0 {
				c.cliCtx.SkipN--
			}
		}
		if env := log.Check(log.DebugLevel, "recv client command"); env != nil {
			env.Write(log.String("addr", c.cliCtx.RemoteAddr),
				log.Int64("clientid", c.cliCtx.ID),
				log.String("traceid", ctx.TraceID),
				log.String("command", ctx.Name))
		}
		c.exec.Execute(ctx)
		cancel()
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
