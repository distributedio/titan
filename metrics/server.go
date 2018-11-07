package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"gitlab.meitu.com/platform/thanos/conf"
)

type Server struct {
	statusServer *http.Server
	addr         string
}

func NewServer(config *conf.Status) *Server {
	s := &Server{
		addr:         config.Listen,
		statusServer: &http.Server{Handler: http.DefaultServeMux},
	}
	return s
}

func (s *Server) Serve(lis net.Listener) error {
	return s.statusServer.Serve(lis)
}

func (s *Server) Stop() error {
	if s.statusServer != nil {
		if err := s.statusServer.Close(); err != nil {
			fmt.Printf("status Server stop failed err:%s \n", err)
			return err
		}
	}
	return nil
}

func (s *Server) GracefulStop() error {
	if s.statusServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := s.statusServer.Shutdown(ctx); err != nil {
			fmt.Printf("status Server stop failed err:%s \n", err)
			return err
		}
	}
	return nil
}

func (s *Server) ListenAndServe(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return s.statusServer.Serve(lis)
}
