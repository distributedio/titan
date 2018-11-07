package logbunny

import (
	"fmt"
	"time"

	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Field zapcore.Field

// Recycle will put back the object into the pool. Make sure the object you wanna recycle
// has noone hold. By default you should not call this function only if you
// actually know the effection
func Recycle(obj *Field) {
	// zap.Recycle()
}

// Skip constructs a no-op Field.
func Skip() *Field {
	field := zap.Skip()
	return (*Field)(&field)
}

// Bool constructs a Field with the given key and value. Bools are marshaled
// lazily.
func Bool(key string, val bool) *Field {
	field := zap.Bool(key, val)
	return (*Field)(&field)
}

// Float64 constructs a Field with the given key and value. The way the
// floating-point value is represented is encoder-dependent, so marshaling is
// necessarily lazy.
func Float64(key string, val float64) *Field {
	field := zap.Float64(key, val)
	return (*Field)(&field)
}

// Int constructs a Field with the given key and value.
func Int(key string, val int) *Field {
	field := zap.Int(key, val)
	return (*Field)(&field)
}

// Int64 constructs a Field with the given key and value. Like ints
func Int64(key string, val int64) *Field {
	field := zap.Int64(key, val)
	return (*Field)(&field)
}

// Uint constructs a Field with the given key and value.
func Uint(key string, val uint) *Field {
	field := zap.Uint(key, val)
	return (*Field)(&field)
}

// Uint64 constructs a Field with the given key and value.
func Uint64(key string, val uint64) *Field {
	field := zap.Uint64(key, val)
	return (*Field)(&field)
}

// Uintptr constructs a Field with the given key and value.
func Uintptr(key string, val uintptr) *Field {
	field := zap.Uintptr(key, val)
	return (*Field)(&field)
}

// Strings constructs a field that carries a slice of strings.
func Strings(key string, val []string) *Field {
	field := zap.Strings(key, val)
	return (*Field)(&field)
}

// Bytes constructs a field that carries a slice of []byte, each of which
// must be UTF-8 encoded text.
func Bytes(key string, val []byte) *Field {
	field := zap.ByteString(key, val)
	return (*Field)(&field)
}

// Ints constructs a field that carries a slice of integers.
func Ints(key string, val []int) *Field {
	field := zap.Ints(key, val)
	return (*Field)(&field)
}

// Caller will add the line & file name to logger. The caller info is refer to goroutine
// String constructs a Field with the given key and value.
func String(key string, val string) *Field {
	field := zap.String(key, val)
	return (*Field)(&field)
}

// Stringer constructs a Field with the given key and the output of the value's String method
func Stringer(key string, val fmt.Stringer) *Field {
	field := zap.Stringer(key, val)
	return (*Field)(&field)
}

// Time constructs a Field with the given key and value. It represents a
// time.Time as a floating-point number of seconds since epoch. Conversion to a
// float64 happens eagerly.
func Time(key string, val time.Time) *Field {
	field := zap.Time(key, val)
	return (*Field)(&field)
}

// Err constructs a Field that lazily stores err.Error() under the key "error".
func Err(err error) *Field {
	field := zap.Error(err)
	return (*Field)(&field)
}

// Duration constructs a Field with the given key and value. It represents
// durations as an integer number of nanoseconds.
func Duration(key string, val time.Duration) *Field {
	field := zap.Duration(key, val)
	return (*Field)(&field)
}

// Object constructs a field with the given key and an arbitrary object. It uses
// an encoding-appropriate, reflection-based function to lazily serialize nearly
// any object into the logging context, but it's relatively slow and
// allocation-heavy.
func Object(key string, val interface{}) *Field {
	field := zap.Any(key, val)
	return (*Field)(&field)
}

// stack information which can be get by runtime.Caller(_skip). _callerSkip is a index
// defined the layer about the stack info which contains the file name and line number.
// This could be unstable while the call stack changed. But mostly we got a stable value in it.
func Caller() *Field {
	field := zap.Stack("Caller")
	lines := strings.Split(field.String, "\n")
	idx := 0
	for i := range lines {
		line := strings.Split(lines[i], "\t")
		if len(line) > 1 {
			idx += 1
			if idx > 1 {
				field.String = line[1]
				break
			}
		}
	}
	return (*Field)(&field)
}
