package resp

import (
	"bytes"
	"testing"
)

func TestArray(t *testing.T) {
	s := "*1\r\n$5\r\nhello\r\nabc"
	r := bytes.NewBufferString(s)
	t.Log(r.Len())
	d := NewRESPDecoder(r)
	size, err := d.Array()
	if err != nil {
		t.Fatal(err)
	}
	if size != 1 {
		t.Fatal("unexpected array size")
	}

	bs, err := d.BulkString()
	if err != nil {
		t.Fatal(err)
	}
	if bs != "hello" {
		t.Fatal("unexpected array item")
	}

}
