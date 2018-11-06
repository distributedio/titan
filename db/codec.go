package db

import (
	"bytes"
	"encoding/binary"
	"math"
)

func EncodeInt64(v int64) []byte {
	var buf bytes.Buffer

	// Ignore the error returned here, because buf is a memory io.Writer, can should not fail here
	binary.Write(&buf, binary.BigEndian, -v)
	return buf.Bytes()
}
func DecodeInt64(b []byte) int64 {
	v := int64(binary.BigEndian.Uint64(b))
	return -v
}

func EncodeFloat64(v float64) []byte {
	var buf bytes.Buffer
	// keep the same pattern of 0.0 and -0.0
	if v == 0.0 {
		v = 0.0
	}

	vi := int64(math.Float64bits(v))
	vi = ((vi ^ (vi >> 63)) | int64(uint64(^vi)&0x8000000000000000))

	// Ignore the error returned here, because buf is a memory io.Writer, can should not fail here
	binary.Write(&buf, binary.BigEndian, vi)
	return buf.Bytes()
}

func DecodeFloat64(d []byte) float64 {
	vi := int64(binary.BigEndian.Uint64(d))
	if vi == 0 {
		return 0.0
	}
	if vi > 0 {
		vi = ^vi
	} else {
		vi = (vi & 0x7FFFFFFFFFFFFFFF)
	}
	return math.Float64frombits(uint64(vi))
}
