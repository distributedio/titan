package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	certFile = "testdata/server.crt"
	keyFile  = "testdata/server.key"
)

func TestGetTLSServerOpts(t *testing.T) {
	_, err := getTLSServerOpts("notfound", "notfoundeither")
	assert.Error(t, err)

	opts, err = getTLSServerOpts(certFile, keyFile)
	assert.NoError(t, err)
	assert.NotNil(t, opts)
}
