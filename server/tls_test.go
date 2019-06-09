package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	brokenFile = "testdata/broken"
	certFile   = "testdata/server.crt"
	keyFile    = "testdata/server.key"
)

func TestTLSConfig(t *testing.T) {
	// not found file #1
	_, err := TLSConfig("/not/found", keyFile)
	assert.Error(t, err)

	// not found file #2
	_, err = TLSConfig(certFile, "/not/found")
	assert.Error(t, err)

	// broken file #1
	_, err = TLSConfig(brokenFile, keyFile)
	assert.Error(t, err)

	// broken file #2
	_, err = TLSConfig(certFile, brokenFile)
	assert.Error(t, err)

	// success
	opts, err := TLSConfig(certFile, keyFile)
	assert.NoError(t, err)
	assert.NotNil(t, opts)
	assert.Len(t, opts.Certificates, 1)
}
