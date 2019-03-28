package titan

import (
	"net"
	"time"

	"github.com/meitu/titan/command"
	"github.com/meitu/titan/context"
	"github.com/meitu/titan/metrics"
	"go.uber.org/zap"
)

//Server implements the redis prototol server
type Server struct {
	servCtx *context.ServerContext
	lis     net.Listener
	idgen   func() int64
}

//New a server instance
func New(ctx *context.ServerContext) *Server {
	// id generator starts from 1(the first client's id is 2, the same as redis)
	return &Server{servCtx: ctx, idgen: GetClientID()}
}

//Serve the redis requests
func (s *Server) Serve(lis net.Listener) error {
	zap.L().Info("titan server start", zap.String("addr", lis.Addr().String()))
	s.servCtx.StartAt = time.Now()
	s.lis = lis
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
			metrics.GetMetrics().ConnectionOnlineGaugeVec.WithLabelValues(cli.cliCtx.Namespace).Inc()
			if err := cli.serve(conn); err != nil {
				zap.L().Error("serve conn failed", zap.String("addr", cli.cliCtx.RemoteAddr),
					zap.Int64("clientid", cliCtx.ID), zap.Error(err))
			}
			metrics.GetMetrics().ConnectionOnlineGaugeVec.WithLabelValues(cli.cliCtx.Namespace).Dec()
			s.servCtx.Clients.Delete(cli.cliCtx.ID)
		}(cli, conn)
	}
}

// ListenAndServe serves on a specified address
func (s *Server) ListenAndServe(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return s.Serve(lis)
}

//Stop the server
func (s *Server) Stop() error {
	zap.L().Info("titan serve stop", zap.String("addr", s.lis.Addr().String()))
	return s.lis.Close()
}

//GracefulStop the server, TODO close clients connections first
func (s *Server) GracefulStop() error {
	zap.L().Info("titan serve graceful", zap.String("addr", s.lis.Addr().String()))
	return s.lis.Close()
}
