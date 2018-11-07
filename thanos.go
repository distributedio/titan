package thanos

import (
	"net"

	"go.uber.org/zap"
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
	zap.L().Info("thanos server start", zap.String("addr", lis.Addr().String()))
	for {
		conn, err := lis.Accept()
		if err != nil {
			zap.L().Error("server accept failed", zap.String("addr", lis.Addr().String()), zap.Error(err))
			return err
		}

		cliCtx := context.NewClientContext(s.idgen(), conn)
		cliCtx.DB = s.servCtx.Store.DB(cliCtx.Namespace, 0)
		s.servCtx.Clients.Store(cliCtx.ID, cliCtx)

		cli := newClient(cliCtx, s, command.NewExecutor())

		zap.L().Info("recv connection", zap.String("addr", cliCtx.RemoteAddr),
			zap.Int64("clientid", cliCtx.ID), zap.String("namespace", cliCtx.Namespace))

		go func(cli *client, conn net.Conn) {
			if err := cli.serve(conn); err != nil {
				zap.L().Error("serve conn failed", zap.String("addr", cli.cliCtx.RemoteAddr),
					zap.Int64("clientid", cliCtx.ID), zap.Error(err))
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
	zap.L().Info("titan serve stop", zap.String("addr", s.lis.Addr().String()))
	return s.lis.Close()
}

func (s *Server) GracefulStop() error {
	//TODO close client connections gracefully
	zap.L().Info("titan serve graceful", zap.String("addr", s.lis.Addr().String()))
	return s.lis.Close()
}
