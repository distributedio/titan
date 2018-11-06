package thanos

import (
	"bufio"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"sync"
	"time"

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
	cancel context.CancelFunc
}

func newClient(cliCtx *context.ClientContext, s *Server, exec *command.Executor) *client {
	return &client{
		cliCtx: cliCtx,
		server: s,
		exec:   exec,
	}
}

func (c *client) serve(conn net.Conn) error {
	log.Println("serve conn")
	c.conn = conn
	c.r = bufio.NewReader(conn)

	// Use a separate goroutine to keep reading commands
	// then we can detect a closed connection as soon as possible.
	// It only works when the cmd channel is not blocked
	cmdc := make(chan []string, 128)
	errc := make(chan error)
	go func() {
		for {
			log.Println("read command")
			cmd, err := c.readCommand()
			if err != nil {
				errc <- err
				cancel := c.cancel
				if cancel != nil {
					cancel()
				}
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
			Name: cmd[0],
			Args: cmd[1:],
			In:   c.r,
			Out:  c.conn,
		}
		innerCtx, cancel := context.WithCancel(context.New(c.cliCtx, c.server.servCtx))
		ctx.Context = innerCtx

		// Make sure only cancel once
		once := sync.Once{}
		c.cancel = func() {
			once.Do(cancel)
		}
		// Skip reply if necessary
		if c.cliCtx.SkipN != 0 {
			ctx.Out = ioutil.Discard
			if c.cliCtx.SkipN > 0 {
				c.cliCtx.SkipN--
			}
		}
		c.exec.Execute(ctx)
		c.cancel()
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
	log.Println(argv)
	return argv, nil
}
