package command

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	ctx := ContextTest("set", "key", "value")
	Call(ctx)
}

func EqualGet(t *testing.T, key string, value string, e error) {
	ctx := ContextTest("get", key)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), value)
}

func EqualStrlen(t *testing.T, key string, ll int) {
	ctx := ContextTest("strlen", key)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), strconv.Itoa(ll))
}

//bug,bug
func EqualMGet(t *testing.T, keys []string, values []string, e error) {
	ctx := ContextTest("mget", keys...)
	Call(ctx)
	for _, v := range values {
		assert.Contains(t, ctxString(ctx.Out), v)
	}

	// assert.Len(t, ctxString(ctx.Out), len(value))
}

var (
	value = "value"
)

func SetEXS(key string) []string {
	args := make([]string, 5)
	args[0] = key
	args[1] = value
	args[2] = "ex"
	args[3] = "1000"
	args[4] = "nx"
	return args
}

func SetPXS(key string) []string {
	args := make([]string, 5)
	args[0] = key
	args[1] = value
	args[2] = "px"
	args[3] = "1000"
	args[4] = "nx"
	return args
}

func SetFour(key string) []string {
	args := make([]string, 4)
	args[0] = key
	args[1] = "value"
	args[2] = "px"
	args[3] = "1000"
	return args
}

// test set ex NX|XX|未知
func TestStringSetEXS(t *testing.T) {

	key := "setexs"
	args := SetEXS(key)

	ctx := ContextTest("set", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "OK")
	EqualGet(t, key, value, nil)
	EqualStrlen(t, key, len(value))

	//修改key失败
	args = SetEXS(key)
	args[1] = "v2"
	ctx = ContextTest("set", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "-1")
	EqualGet(t, key, value, nil)

	args = SetEXS(key)
	args[1] = "v2"
	args[4] = "xx"

	ctx = ContextTest("set", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "OK")
	EqualGet(t, key, "v2", nil)
	EqualStrlen(t, key, len("v2"))
	// 测试nx
	// 修改key 失败
	args = SetEXS(key)
	args[1] = "value"
	ctx = ContextTest("set", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "-1")
	EqualGet(t, key, "v2", nil)

	//乱序测试
	args[0] = key
	args[1] = "v1"
	args[2] = "xx"
	args[3] = "ex"
	args[4] = "1000"

	ctx = ContextTest("set", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "OK")
	EqualGet(t, key, "v1", nil)

	//异常测试
	args = SetEXS(key)
	ctx = ContextTest("set", args[:3]...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), ErrSyntax.Error())

	//异常测试
	args = SetEXS(key)
	args[3] = "bx"
	ctx = ContextTest("set", args[:3]...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), ErrSyntax.Error())
}

// test set px NX|XX|未知
func TestStringSetPXS(t *testing.T) {

	key := "setpx"
	args := SetPXS(key)

	ctx := ContextTest("set", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "OK")
	EqualGet(t, key, value, nil)

	//修改key失败
	args = SetPXS(key)
	args[1] = "v2"

	ctx = ContextTest("set", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "-1")
	EqualGet(t, key, value, nil)

	// 测试nx
	args = SetPXS(key)
	args[1] = "v2"
	args[4] = "xx"

	ctx = ContextTest("set", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "OK")
	EqualGet(t, key, args[1], nil)

	// 修改key 失败
	// key =
	args = SetPXS("kpx2")
	args[4] = "xx"

	ctx = ContextTest("set", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "-1")
	EqualGet(t, key, "v2", nil)

	//异常测试
	ctx = ContextTest("set", args[:3]...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), ErrSyntax.Error())

	//乱序测试
	args[0] = key
	args[1] = "v1"
	args[2] = "xx"
	args[3] = "px"
	args[4] = "10000"

	ctx = ContextTest("set", args[:3]...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "OK")
	EqualGet(t, key, "v1", nil)

	//异常测试
	args = SetPXS(key)
	args[3] = "bx"
	args[4] = "xx"

	ctx = ContextTest("set", args[:3]...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), ErrSyntax.Error())
}

// test set px|ex|未知
func TestStringSetFour(t *testing.T) {
	key := "setpxex"
	args := SetFour(key)
	ctx := ContextTest("set", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "OK")
	EqualGet(t, key, value, nil)

	args[2] = "ex"
	ctx = ContextTest("set", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "OK")
	EqualGet(t, key, value, nil)

	args[3] = "x"
	ctx = ContextTest("set", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), ErrInteger.Error())

	args[2] = "zx"

	ctx = ContextTest("set", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), ErrSyntax.Error())
}

// test set NX|XX|未知
func TestStringSetThree(t *testing.T) {
	args := make([]string, 3)
	key := "setxxnxt"
	args[0] = key
	args[1] = value
	args[2] = "nx"
	ctx := ContextTest("set", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "OK")
	EqualGet(t, key, value, nil)

	args[1] = "v1"
	args[2] = "xx"

	ctx = ContextTest("set", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "OK")
	EqualGet(t, key, "v1", nil)

	args[2] = "zx"
	ctx = ContextTest("set", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), ErrSyntax.Error())
}

func TestStringSet(t *testing.T) {

	args := make([]string, 2)
	key := "set"
	args[0] = key
	args[1] = "value"
	ctx := ContextTest("lpush", args...)
	Call(ctx)

	ctx = ContextTest("set", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "OK")
	EqualGet(t, key, "value", nil)
	// EqualMGet(t, []string{key}, []string{"value"}, nil)
}

func TestStringSetEx(t *testing.T) {
	args := make([]string, 3)
	key := "setex"
	args[0] = key
	args[1] = "10000"
	args[2] = value

	ctx := ContextTest("setex", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "OK")
	EqualGet(t, key, value, nil)

	args[1] = "x"
	ctx = ContextTest("setex", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), ErrInteger.Error())
}

func TestStringSetNx(t *testing.T) {
	args := make([]string, 2)
	key := "setnx"
	args[0] = key
	args[1] = value

	ctx := ContextTest("setnx", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "1")
	EqualGet(t, key, value, nil)

	args[1] = "v1"
	ctx = ContextTest("setnx", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "0")
	EqualGet(t, key, value, nil)
}

func TestStringPSetEx(t *testing.T) {
	args := make([]string, 3)
	key := "psetex"
	args[0] = key
	args[1] = "100000"
	args[2] = value

	ctx := ContextTest("psetex", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "OK")
	EqualGet(t, key, value, nil)

	args[1] = "x"
	ctx = ContextTest("psetex", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), ErrInteger.Error())

}

func TestStringSetRange(t *testing.T) {
	args := make([]string, 3)
	key := "setrange"
	args[0] = key
	args[1] = "3"
	args[2] = value

	ctx := ContextTest("setrange", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "8")
	ctx = ContextTest("get", key)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), value)

	args[1] = "1"
	args[2] = "lll"
	ctx = ContextTest("setrange", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "8")
	ctx = ContextTest("get", key)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "lllalue")

	args[1] = "10"
	args[2] = value
	ctx = ContextTest("setrange", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "15")
	ctx = ContextTest("get", key)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "\x00lllalue\x00\x00value")

	args[1] = "s"
	ctx = ContextTest("setrange", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), ErrInteger.Error())

	args[1] = "-2"
	ctx = ContextTest("setrange", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), ErrMaximum.Error())
}
func TestStringIncr(t *testing.T) {
	args := make([]string, 1)
	args[0] = "incr"
	ctx := ContextTest("incr", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "1")

	args[0] = "setex"
	ctx = ContextTest("incr", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), ErrInteger.Error())
}

func TestStringIncrBy(t *testing.T) {
	args := make([]string, 2)
	args[0] = "incrby"
	args[1] = "2"
	ctx := ContextTest("incrby", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "2")

	args[1] = "-2"
	ctx = ContextTest("incrby", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "0")

	args[1] = "02"
	ctx = ContextTest("incrby", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "2")

}

//bug
func TestStringIncrByFloat(t *testing.T) {
	args := make([]string, 2)
	args[0] = "incrbyfloat"
	args[1] = "2.0e2"

	ctx := ContextTest("incrbyfloat", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "200")

	args[1] = "-2.0e2"
	ctx = ContextTest("incrbyfloat", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "0")

	args[1] = "02"
	ctx = ContextTest("incrbyfloat", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "0")
}

func TestStringDecr(t *testing.T) {
	args := make([]string, 1)
	args[0] = "decr"
	ctx := ContextTest("decr", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "-1")

	args[0] = "setex"
	ctx = ContextTest("decr", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), ErrInteger.Error())
}

func TestStringDecrBy(t *testing.T) {
	args := make([]string, 2)
	args[0] = "decrby"
	args[1] = "2"

	ctx := ContextTest("decrby", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "-2")

	args[1] = "-2"
	ctx = ContextTest("decrby", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "0")

	args[1] = "02"
	ctx = ContextTest("decrby", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "-2")
}

func TestStringMset(t *testing.T) {
	args := make([]string, 4)
	args[0] = "Mset1"
	args[1] = "Mset3"
	args[2] = "Mset2"
	args[3] = "Mset4"

	ctx := ContextTest("mset", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "OK")
	EqualMGet(t, []string{args[0], args[2]}, []string{args[1], args[3]}, nil)

	ctx = ContextTest("mset", args[:3]...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), ErrMSet.Error())
}

func TestStringMsetNx(t *testing.T) {
	args := make([]string, 4)
	args[0] = "MsetN1"
	args[1] = "MsetN3"
	args[2] = "MsetN2"
	args[3] = "MsetN4"

	ctx := ContextTest("msetnx", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "1")
	EqualMGet(t, []string{args[0], args[2]}, []string{args[1], args[3]}, nil)

	ctx = ContextTest("msetnx", args[:3]...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), ErrMSet.Error())

	args[2] = "MsetN5"
	ctx = ContextTest("msetnx", args...)
	Call(ctx)
	assert.Contains(t, ctxString(ctx.Out), "0")
	EqualGet(t, args[0], args[1], nil)
}

func TestStringAppend(t *testing.T) {
	args := make([]string, 2)
	args[0] = "Append"
	args[1] = "Ap"

	out := CallTest("append", args...)
	assert.Contains(t, out.String(), strconv.Itoa(len(args[1])))

	out = CallTest("append", args...)
	assert.Contains(t, out.String(), strconv.Itoa(len(args[1])*2))
}

func TestStringSetBit(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "1",
			args: []string{"setbit", "1", "0", "0"},
			want: "-ERR wrong number of arguments for 'setbit' command",
		},
		{
			name: "2",
			args: []string{"setbit", "x", "0"},
			want: ErrBitOffset.Error(),
		},
		{
			name: "3",
			args: []string{"setbit", "1", "x"},
			want: ErrBitInteger.Error(),
		},
		{
			name: "4",
			args: []string{"setbit", "1", "2"},
			want: ErrBitInteger.Error(),
		},
		{
			name: "5",
			args: []string{"setbit", "1", "1"},
			want: ":0",
		},
		{
			name: "6",
			args: []string{"setbit", "1", "0"},
			want: ":1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := CallTest("setbit", tt.args...)
			assert.Contains(t, out.String(), tt.want)
		})
	}
}

func TestStringGetBit(t *testing.T) {
	CallTest("setbit", "getbit", "5", "1")
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "1",
			args: []string{"getbit", "1", "0", "0"},
			want: "-ERR wrong number of arguments for 'getbit' command",
		},
		{
			name: "2",
			args: []string{"getbit", "x"},
			want: ErrBitOffset.Error(),
		},
		{
			name: "3",
			args: []string{"getbit", "1"},
			want: ":0",
		},
		{
			name: "5",
			args: []string{"getbit", "5"},
			want: ":1",
		},
		{
			name: "6",
			args: []string{"getbit", "10"},
			want: ":0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := CallTest("getbit", tt.args...)
			assert.Contains(t, out.String(), tt.want)
		})
	}
}
