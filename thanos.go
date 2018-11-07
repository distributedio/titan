package thanos

import (
	"net"

	log "gitlab.meitu.com/gocommons/logbunny"
	"gitlab.meitu.com/platform/thanos/command"
	"gitlab.meitu.com/platform/thanos/context"
)

type Server struct {
	servCtx *context.ServerContext
	lis     net.Listener
	idgen   func() int64
}

func New(ctx *context.ServerContext) *Server {
	// id generator starts from 1(the first client's id is 2, the same as redis)
	return &Server{servCtx: ctx, idgen: GetClientID()}
}

func (s *Server) Serve(lis net.Listener) error {
	log.Info("thanos server start", log.String("addr", lis.Addr().String()))
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Error("server accept failed", log.String("addr", lis.Addr().String()), log.Err(err))
			return err
		}

		cliCtx := context.NewClientContext(s.idgen(), conn)
		cliCtx.DB = s.servCtx.Store.DB(cliCtx.Namespace, 0)
		s.servCtx.Clients.Store(cliCtx.ID, cliCtx)

		cli := newClient(cliCtx, s, command.NewExecutor())

		log.Info("recv connection", log.String("addr", cliCtx.RemoteAddr),
			log.Int64("clientid", cliCtx.ID), log.String("namespace", cliCtx.Namespace))

		go func(cli *client, conn net.Conn) {
			if err := cli.serve(conn); err != nil {
				log.Error("serve conn failed", log.String("addr", cli.cliCtx.RemoteAddr),
					log.Int64("clientid", cliCtx.ID), log.Err(err))
			}
			s.servCtx.Clients.Delete(cli.cliCtx.ID)
		}(cli, conn)
	}
	return nil
}

func (s *Server) ListenAndServe(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.lis = lis
	return s.Serve(lis)
}

func (s *Server) Stop() error {
	log.Info("titan serve stop", log.String("addr", s.lis.Addr().String()))
	return s.lis.Close()
}

func (s *Server) GracefulStop() error {
	//TODO close client connections gracefully
	log.Info("titan serve graceful", log.String("addr", s.lis.Addr().String()))
	return s.lis.Close()
}
