package db

import (
	"time"

	"github.com/satori/go.uuid"
)

// UUID allocates an unique object ID.
func UUID() []byte {
	return uuid.NewV4().Bytes()
}

// UUIDString returns canonical string representation of UUID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
func UUIDString(id []byte) string {
	return uuid.FromBytesOrNil(id).String()
}

// Now returns the current unix nano timestamp.
func Now() int64 {
	return time.Now().UnixNano()
}
