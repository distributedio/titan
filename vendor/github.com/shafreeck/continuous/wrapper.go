package continuous

import (
	"context"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"
)

type httpServer struct {
	*http.Server
}

func (s *httpServer) Stop() error {
	return s.Server.Close()
}
func (s *httpServer) GracefulStop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return s.Server.Shutdown(ctx)
}

func WrapHTTPServer(s *http.Server) Continuous {
	return &httpServer{s}
}

type httpServerTLS struct {
	*httpServer
	certFile string
	keyFile  string
}

func WrapHTTPServerTLS(s *http.Server, certFile, keyFile string) Continuous {
	return &httpServerTLS{httpServer: &httpServer{s}, certFile: certFile, keyFile: keyFile}
}
func (s *httpServerTLS) Serve(lis net.Listener) error {
	return s.ServeTLS(lis, s.certFile, s.keyFile)
}

type grpcServer struct {
	*grpc.Server
}

func (s *grpcServer) Stop() error {
	s.Server.Stop()
	return nil
}
func (s *grpcServer) GracefulStop() error {
	s.Server.GracefulStop()
	return nil
}

func WrapGRPCServer(s *grpc.Server) Continuous {
	return &grpcServer{s}
}
