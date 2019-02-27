package db

import (
	"math"
	"time"

	"github.com/satori/go.uuid"
)

// UUID allocates an unique object ID.
func UUID() []byte { return uuid.NewV4().Bytes() }

// UUIDString returns canonical string representation of UUID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
func UUIDString(id []byte) string { return uuid.FromBytesOrNil(id).String() }

// Now returns the current unix nano timestamp.
func Now() int64 { return time.Now().UnixNano() }

func EndKey(prefix []byte) []byte {
	if len(prefix) >= MAX_KEY_SIZE {
		return prefix
	}
	key := make([]byte, MAX_KEY_SIZE)
	for i := 0; i < MAX_KEY_SIZE; i++ {
		if i < len(prefix) {
			key[i] = prefix[i]
		} else {
			key[i] = math.MaxUint8
		}
	}
	return key
}
