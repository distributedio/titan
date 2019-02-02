package db

import (
	"encoding/binary"
	"time"

	"github.com/satori/go.uuid"
)

// UUID allocates an unique object ID.
func UUID() []byte { return uuid.NewV4().Bytes() }

// UUIDString returns canonical string representation of UUID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
func UUIDString(id []byte) string { return uuid.FromBytesOrNil(id).String() }

// Now returns the current unix nano timestamp.
func Now() int64 { return time.Now().UnixNano() }

func bytesToUint32s(s []byte) []uint32 {

	// Count initial bytes not aligned to 32 bit.
	bits := s
	begin := 0
	ll := 4 - len(bits)%4
	if ll != 4 {
		bits = append(bits, make([]byte, ll)...)
	}

	nums := make([]uint32, len(bits)/4)
	for begin < len(s) {
		nums = append(nums, binary.BigEndian.Uint32(bits[begin:begin+4]))
		begin += 4
	}
	return nums
}
