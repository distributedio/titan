package db

import (
	"testing"
)

func TestDecodeFloat64(t *testing.T) {
	for _, f := range []float64{1, -1, 0.1, -0.1} {
		v := DecodeFloat64(EncodeFloat64(f))
		t.Log(v)
		if v != f {

			t.Fatal("decode failed", v)
		}
	}
}
