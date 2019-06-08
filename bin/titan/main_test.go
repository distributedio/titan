package main

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
	_, err := getTLSServerOpts("notfound", keyFile)
	assert.Error(t, err)

	// not found #2
	_, err = getTLSServerOpts(certFile, "notfound")
	assert.Error(t, err)

	// broken file #1
	_, err = getTLSServerOpts(brokenFile, keyFile)
	assert.Error(t, err)

	// broken file #2
	_, err = getTLSServerOpts(certFile, brokenFile)
	assert.Error(t, err)

	// success
	opts, err := getTLSServerOpts(certFile, keyFile)
	assert.NoError(t, err)
	assert.NotNil(t, opts)
}
