package db

import (
	"bytes"
	"math"
	"reflect"
	"testing"
)

func TestCodecObject(t *testing.T) {
	now := Now()
	tests := []*Object{
		&Object{
			ID:        []byte("1234567890123456"),
			CreatedAt: now,
			UpdatedAt: now,
			ExpireAt:  0,
			Type:      ObjectString,
			Encoding:  ObjectEncodingRaw,
		},
	}
	for _, obj := range tests {
		got, err := DecodeObject(EncodeObject(obj))
		if err != nil {
			t.Fatalf("decode object get error: %s", err)
		}
		if !reflect.DeepEqual(got, obj) {
			t.Fatalf("decode failed want=%v, got=%v", obj, got)
		}
	}
}

func TestCodecEncodeInt64(t *testing.T) {
	values := []int64{math.MinInt64, -1, 0, 1, math.MaxInt64}
	for i := 0; i < len(values)-1; i++ {
		if bytes.Compare(EncodeInt64(values[i]), EncodeInt64(values[i+1])) >= 0 {
			t.Fatal("EncodeInt64 is not memcomparable")
		}
	}
}

func TestCodecEncodeFloat64(t *testing.T) {
	values := []float64{-1.0, 0.0, 1.0, math.MaxFloat64}
	for i := 0; i < len(values)-1; i++ {
		if bytes.Compare(EncodeFloat64(values[i]), EncodeFloat64(values[i+1])) >= 0 {
			t.Fatal("EncodeFloat64 is not memcomparable")
		}
	}
}

func TestCodecInt64(t *testing.T) {
	for _, i := range []int64{0, 1, -1, math.MaxInt64, math.MinInt64} {
		v := DecodeInt64(EncodeInt64(i))
		if v != i {
			t.Fatalf("decode failed want=%v, got=%v", i, v)
		}
	}
}

func TestCodecFloat64(t *testing.T) {
	for _, f := range []float64{math.MaxFloat64, math.SmallestNonzeroFloat64, 0, 1, -1, 0.1, -0.1} {
		v := DecodeFloat64(EncodeFloat64(f))
		if v != f {
			t.Fatalf("decode failed want=%v, got=%v", f, v)
		}
	}
}
