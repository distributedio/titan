package server

import (
	"crypto/rand"
	"crypto/tls"

	"github.com/shafreeck/continuous"
)

// GetTLSServerOpts loads the TLS certificate and key files, returning a
// continuous.ServerOption struct configured for TLS.
func GetTLSServerOpts(certFile, keyFile string) (continuous.ServerOption, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return continuous.TLSConfig(&tls.Config{
		Certificates: []tls.Certificate{cert},
		Rand:         rand.Reader,
	}), nil
}
