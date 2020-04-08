package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriter(t *testing.T) {
	stream, _err := Writer("stdout", "", true)
	assert.Equal(t, stream, os.Stdout)
	assert.Nil(t, _err)

	stream, _err = Writer("stderr", "", true)
	assert.Equal(t, stream, os.Stderr)
	assert.Nil(t, _err)
}
