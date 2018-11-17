package resp

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArray_Decode(t *testing.T) {
	assert := assert.New(t)
	r := bytes.NewBufferString("*1\r\n$5\r\nhello\r\nabc")
	d := NewDecoder(r)

	size, err := d.Array()
	assert.NoError(err)
	assert.Equal(1, size)

	bs, err := d.BulkString()
	assert.NoError(err)
	assert.Equal("hello", bs)
}

func TestArray_Encode(t *testing.T) {
	assert := assert.New(t)
	out := bytes.NewBuffer(nil)
	e := NewEncoder(out)

	// Empty array
	err := e.Array(0)
	assert.NoError(err)
	assert.Equal("*0\r\n", out.String())

	// Array with one item
	out.Reset()
	err = e.Array(1)
	assert.NoError(err)
	assert.Equal("*1\r\n", out.String())

	// Array of large data
	out.Reset()
	err = e.Array(1000000000000)
	assert.NoError(err)
	assert.Equal("*1000000000000\r\n", out.String())
}

func TestSimpleString_Decode(t *testing.T) {
	assert := assert.New(t)
	d := NewDecoder(bytes.NewBufferString("+OK\r\n"))
	val, err := d.SimpleString()
	assert.NoError(err)
	assert.Equal("OK", val)
}

func TestSimpleString_Encode(t *testing.T) {
	assert := assert.New(t)
	out := bytes.NewBuffer(nil)
	e := NewEncoder(out)
	err := e.SimpleString("OK")
	assert.NoError(err)
	assert.Equal("+OK\r\n", out.String())
}

func TestBulkString_Decode(t *testing.T) {
	assert := assert.New(t)
	d := NewDecoder(bytes.NewBufferString("$4\r\ntest\r\n"))
	val, err := d.BulkString()
	assert.NoError(err)
	assert.Equal("test", val)

	// Truncated data
	d = NewDecoder(bytes.NewBufferString("$3\r\ntest\r\n"))
	val, err = d.BulkString()
	assert.NoError(err)
	assert.Equal("tes", val)

	// Invalid indicator
	d = NewDecoder(bytes.NewBufferString("*4\r\ntest\r\n"))
	val, err = d.BulkString()
	assert.Error(err)
	assert.Equal("", val)

	// Invalid delimiter
	d = NewDecoder(bytes.NewBufferString("*4\rtest\r\n"))
	val, err = d.BulkString()
	assert.Error(err)
	assert.Equal("", val)

	// Naughty string
	d = NewDecoder(bytes.NewBufferString("asdfghjk"))
	val, err = d.BulkString()
	assert.Error(err)
	assert.Equal("", val)
}

func TestBulkString_Encode(t *testing.T) {
	assert := assert.New(t)
	out := bytes.NewBuffer(nil)
	e := NewEncoder(out)
	err := e.BulkString("test")
	assert.NoError(err)
	assert.Equal("$4\r\ntest\r\n", out.String())
}

func TestError_Decode(t *testing.T) {
	assert := assert.New(t)
	d := NewDecoder(bytes.NewBufferString("-error\r\n"))
	val, err := d.Error()
	assert.NoError(err)
	assert.Equal("error", val)

	// Naughty string
	d = NewDecoder(bytes.NewBufferString("asdfghjk"))
	val, err = d.Error()
	assert.Error(err)
	assert.Equal("", val)
}

func TestError_Encode(t *testing.T) {
	assert := assert.New(t)
	out := bytes.NewBuffer(nil)
	e := NewEncoder(out)
	err := e.Error("error")
	assert.NoError(err)
	assert.Equal("-error\r\n", out.String())
}

func TestInteger_Decode(t *testing.T) {
	assert := assert.New(t)
	d := NewDecoder(bytes.NewBufferString(":1\r\n"))
	val, err := d.Integer()
	assert.NoError(err)
	assert.Equal(int64(1), val)

	// Naughty string
	d = NewDecoder(bytes.NewBufferString("asdfghjk"))
	val, err = d.Integer()
	assert.Error(err)
	assert.Equal(int64(-1), val)
}

func TestInteger_Encode(t *testing.T) {
	assert := assert.New(t)
	out := bytes.NewBuffer(nil)
	e := NewEncoder(out)
	err := e.Integer(1)
	assert.NoError(err)
	assert.Equal(":1\r\n", out.String())
}

func TestNullBulkString_Encode(t *testing.T) {
	assert := assert.New(t)
	out := bytes.NewBuffer(nil)
	e := NewEncoder(out)
	err := e.NullBulkString()
	assert.NoError(err)
	assert.Equal("$-1\r\n", out.String())
}
