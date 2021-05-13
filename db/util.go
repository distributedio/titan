package db

import (
	"encoding/binary"
	"math"
	"math/bits"
	"time"

	uuid "github.com/satori/go.uuid"
)

// UUID allocates an unique object ID.
func UUID() []byte { return uuid.Must(uuid.NewV4()).Bytes() }

// UUIDString returns canonical string representation of UUID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
func UUIDString(id []byte) string { return uuid.FromBytesOrNil(id).String() }

// Now returns the current unix nano timestamp.
func Now() int64 { return time.Now().UnixNano() }

func redisPopcount(s []byte) int {
	// Count initial bytes not aligned to 32 bit.
	bitval := s
	begin := 0
	ll := 4 - len(bitval)%4
	if ll != 4 {
		bitval = append(bitval, make([]byte, ll)...)
	}

	sum := 0
	for begin < len(s) {
		num := binary.BigEndian.Uint32(bitval[begin : begin+4])
		sum += bits.OnesCount32(num)
		begin += 4
	}

	return sum
}

func initCursor(begin, end, llen int) (int, int) {
	if begin < 0 {
		begin = llen + begin
	}
	if end < 0 {
		end = llen + end
	}

	if begin < 0 {
		begin = 0
	}
	if end < 0 {
		end = 0
	}
	if end >= llen {
		end = llen - 1
	}
	return begin, end
}

func redisBitpos(val []byte, bit int) int {
	// fill in four bytes
	bitval := val
	ll := 4 - len(bitval)%4
	if ll != 4 {
		bitval = append(bitval, make([]byte, ll)...)
	}

	// calculated in term of four bytes
	var skipval uint32
	if bit == 0 {
		skipval = math.MaxUint32
	}
	pos := 0
	for pos < len(bitval) {
		num := binary.BigEndian.Uint32(bitval[pos : pos+4])
		if num != skipval {
			break
		}
		pos += 4
	}

	// calculated in term of one bytes
	var one uint8
	if bit == 0 {
		one = math.MaxUint8
	}
	for pos < len(bitval) {
		num := uint8(bitval[pos])
		if num != one {
			break
		}
		pos += 1
	}
	if pos == len(bitval) && bit == 1 {
		return -1
	}

	// find the corresponding bit
	one = math.MaxUint8 /* All bitval set to 1.*/
	one >>= 1           /* All bitval set to 1 but the MSB. */
	one = ^one          /* All bitval set to 0 but the MSB. */
	word := bitval[pos]
	pos = pos * 8
	bbit := false
	if bit == 1 {
		bbit = true
	}
	for one != 0 {
		// fmt.Println(one, word, (one & word), pos, uint8(bit))
		if ((one & word) != 0) == bbit {
			return pos
		}
		one >>= 1
		pos++
	}
	return pos
}
