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
	_, err := getTLSServerOpts("notfound", "notfoundeither")
	assert.Error(t, err)

	_, err = getTLSServerOpts(brokenFile, keyFile)
	assert.Error(t, err)

	_, err = getTLSServerOpts(certFile, brokenFile)
	assert.Error(t, err)

	opts, err := getTLSServerOpts(certFile, keyFile)
	assert.NoError(t, err)
	assert.NotNil(t, opts)
}
