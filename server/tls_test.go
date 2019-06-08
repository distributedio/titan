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

func TestGetTLSServerOpts(t *testing.T) {
	// not found
	_, err := GetTLSServerOpts("notfound", keyFile)
	assert.Error(t, err)

	// not found #2
	_, err = GetTLSServerOpts(certFile, "notfound")
	assert.Error(t, err)

	// broken file #1
	_, err = GetTLSServerOpts(brokenFile, keyFile)
	assert.Error(t, err)

	// broken file #2
	_, err = GetTLSServerOpts(certFile, brokenFile)
	assert.Error(t, err)

	// success
	opts, err := GetTLSServerOpts(certFile, keyFile)
	assert.NoError(t, err)
	assert.NotNil(t, opts)
}
