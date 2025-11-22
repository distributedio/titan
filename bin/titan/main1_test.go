package main

import (
	"os"
	"path"
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

	td, td_err := os.MkdirTemp("", "titan-test")
	assert.Nil(t, td_err)
	_, _err = Writer(path.Join(td, "titan-test-log"), "* * * * *", true)
	assert.Nil(t, _err)

	_, _err = Writer(path.Join(td, "titan-test-log"), "", true)
	assert.NotNil(t, _err)

}
