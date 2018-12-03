package db

import (
	"bytes"
	"encoding/binary"
	"math"
)

// EncodeObject encode the object to binary
func EncodeObject(obj *Object) []byte {
	b := make([]byte, 42, 42)
	copy(b, obj.ID[:16])
	binary.BigEndian.PutUint64(b[16:], uint64(obj.CreatedAt))
	binary.BigEndian.PutUint64(b[24:], uint64(obj.UpdatedAt))
	binary.BigEndian.PutUint64(b[32:], uint64(obj.ExpireAt))
	b[40], b[41] = byte(obj.Type), byte(obj.Encoding)
	return b
}

// DecodeObject decode the object from binary
func DecodeObject(b []byte) (obj *Object, err error) {
	if len(b) < ObjectEncodingLength {
		return nil, ErrInvalidLength
	}
	obj = &Object{
		ID:        make([]byte, 16, 16),
		CreatedAt: int64(binary.BigEndian.Uint64(b[16:])),
		UpdatedAt: int64(binary.BigEndian.Uint64(b[24:])),
		ExpireAt:  int64(binary.BigEndian.Uint64(b[32:])), // 40 bit fields
		Type:      ObjectType(b[40]),                      // 41 bit
		Encoding:  ObjectEncoding(b[41]),                  // 42 bit
	}
	copy(obj.ID, b[0:16])
	return obj, nil
}

// EncodeInt64  encode the int64 object to binary
func EncodeInt64(v int64) []byte {
	var buf bytes.Buffer
	if v < 0 {
		v = int64(uint64(v) & 0x7FFFFFFFFFFFFFFF)
	} else if v >= 0 {
		v = int64(uint64(v) | 0x8000000000000000)
	}

	// Ignore the error returned here, because buf is a memory io.Writer, can should not fail here
	binary.Write(&buf, binary.BigEndian, v)
	return buf.Bytes()
}

// DecodeInt64 decode the int64 object from binary
func DecodeInt64(b []byte) int64 {
	v := int64(binary.BigEndian.Uint64(b))
	if v < 0 {
		v = int64(uint64(v) & 0x7FFFFFFFFFFFFFFF)
	} else if v >= 0 {
		v = int64(uint64(v) | 0x8000000000000000)
	}
	return v
}

// EncodeFloat64 encode the float64 object to binary
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

// DecodeFloat64 decode the float64 object from binary
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
