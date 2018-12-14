package metrics

import (
	"context"
	"net"
	"net/http"

	"go.uber.org/zap"

	"time"

	"github.com/meitu/titan/conf"
)

//Server status server
//export go pprof ane promtheus monitor
type Server struct {
	statusServer *http.Server
	addr         string
}

//NewServer creat status server
func NewServer(config *conf.Status) *Server {
	s := &Server{
		addr:         config.Listen,
		statusServer: &http.Server{Handler: http.DefaultServeMux},
	}
	return s
}

// Serve accepts incoming connections on the Listener l
func (s *Server) Serve(lis net.Listener) error {
	zap.L().Info("status server start", zap.String("addr", s.addr))
	MetricsHandle()
	return s.statusServer.Serve(lis)
}

//Stop Close serve fd
func (s *Server) Stop() error {
	zap.L().Info("status server stop", zap.String("addr", s.addr))
	if s.statusServer != nil {
		if err := s.statusServer.Close(); err != nil {
			zap.L().Error("status server stop failed", zap.String("addr", s.addr), zap.Error(err))
			return err
		}
	}
	zap.L().Info("status server stop sucess", zap.String("addr", s.addr))
	return nil
}

//GracefulStop serve graceful stop
func (s *Server) GracefulStop() error {
	zap.L().Info("status serve graceful stop", zap.String("addr", s.addr))
	if s.statusServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := s.statusServer.Shutdown(ctx); err != nil {
			zap.L().Error("status server gracestop failed", zap.String("addr", s.addr), zap.Error(err))
			return err
		}
	}
	zap.L().Info("status serve graceful sucess", zap.String("addr", s.addr))
	return nil
}

//ListenAndServe start the service by address
func (s *Server) ListenAndServe(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		zap.L().Error("status server liseten failed", zap.String("addr", s.addr), zap.Error(err))
		return err
	}
	zap.L().Info("status server start", zap.String("addr", s.addr))
	return s.Serve(lis)
}
