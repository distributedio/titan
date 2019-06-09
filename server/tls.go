package server

import (
	"crypto/rand"
	"crypto/tls"
)

// TLSConfig loads the TLS certificate and key files, returning a
// tls.Config.
func TLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		Rand:         rand.Reader,
	}, nil
}
