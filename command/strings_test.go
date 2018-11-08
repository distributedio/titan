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

/*
func EqualMGet(t *testing.T, key [][]byte, value [][]byte, e error) {
	cmdctx.Db.Begin()
	r, err := MGetHandler(key, cmdctx)
	cmdctx.Db.Commit()
	vs := make([]*protocol.ReplyData, len(value))
	for i, v := range value {
		vs[i] = &protocol.ReplyData{Type: protocol.REPLYBULK, Value: v}
	}
	re := &protocol.ReplyData{Type: protocol.REPLYARRAY, Value: vs}
	assert.Equal(t, re, r)
	assert.NoError(t, err)
}

*/
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

/*
// test set px NX|XX|未知
func TestStringSetPXS(t *testing.T) {

	key := []byte("setpx")
	args := SetPXS(key)
	cmdctx.Db.Begin()
	r, err := SetHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisOkResp, r)
	assert.NoError(t, err)
	EqualGet(t, key, value, nil)

	//修改key失败
	args = SetPXS(key)
	args[1] = []byte("v2")
	cmdctx.Db.Begin()
	r, err = SetHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisNilResp, r)
	assert.NoError(t, err)
	EqualGet(t, key, value, nil)

	// 测试nx
	args = SetPXS(key)
	args[1] = []byte("v2")
	args[4] = []byte("xx")
	cmdctx.Db.Begin()
	r, err = SetHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisOkResp, r)
	assert.NoError(t, err)
	EqualGet(t, key, args[1], nil)

	// 修改key 失败
	// key =
	args = SetPXS([]byte("kpx2"))
	args[4] = []byte("xx")
	cmdctx.Db.Begin()
	r, err = SetHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisNilResp, r)
	assert.NoError(t, err)
	EqualGet(t, key, []byte("v2"), nil)

	//异常测试
	cmdctx.Db.Begin()
	r, err = SetHandler(args[:3], cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisSyntaxResp, r)
	assert.NotNil(t, err)

	//乱序测试
	args[0] = key
	args[1] = []byte("v1")
	args[2] = []byte("xx")
	args[3] = []byte("px")
	args[4] = []byte("10000")

	cmdctx.Db.Begin()
	r, err = SetHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisOkResp, r)
	assert.NoError(t, err)
	EqualGet(t, key, []byte("v1"), nil)

	//异常测试
	args = SetPXS(key)
	args[3] = []byte("bx")
	args[4] = []byte("xx")
	cmdctx.Db.Begin()
	r, err = SetHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisIntegerResp, r)
	assert.NotNil(t, err)
}

// test set px|ex|未知
func TestStringSetFour(t *testing.T) {
	key := []byte("setpxex")
	args := SetFour(key)
	cmdctx.Db.Begin()
	r, err := SetHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisOkResp, r)
	assert.NoError(t, err)
	EqualGet(t, key, value, nil)

	// args = SetFour(key)
	args[2] = []byte("ex")
	cmdctx.Db.Begin()
	r, err = SetHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisOkResp, r)
	assert.NoError(t, err)
	EqualGet(t, key, value, nil)

	args[3] = []byte("x")
	cmdctx.Db.Begin()
	r, err = SetHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisIntegerResp, r)
	assert.NotNil(t, err)

	args[2] = []byte("zx")
	cmdctx.Db.Begin()
	r, err = SetHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisSyntaxResp, r)
	assert.NotNil(t, err)
}

// test set NX|XX|未知
func TestStringSetThree(t *testing.T) {
	args := make([][]byte, 3)
	key := []byte("setxxnxt")
	args[0] = key
	args[1] = value
	args[2] = []byte("nx")
	cmdctx.Db.Begin()
	r, err := SetHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisOkResp, r)
	assert.NoError(t, err)
	EqualGet(t, key, value, nil)

	args[1] = []byte("v1")
	args[2] = []byte("xx")
	cmdctx.Db.Begin()
	r, err = SetHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisOkResp, r)
	assert.NoError(t, err)
	EqualGet(t, key, []byte("v1"), nil)

	args[2] = []byte("zx")
	cmdctx.Db.Begin()
	r, err = SetHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisSyntaxResp, r)
	assert.NotNil(t, err)
}

func TestStringSet(t *testing.T) {
	args := make([][]byte, 2)
	key := []byte("set")
	args[0] = key
	args[1] = []byte("value")
	cmdctx.Db.Begin()
	r, err := SetHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisOkResp, r)
	assert.NoError(t, err)
	EqualGet(t, key, []byte("value"), nil)
}

func TestStringSetEx(t *testing.T) {
	args := make([][]byte, 3)
	key := []byte("setex")
	args[0] = key
	args[1] = []byte("10000")
	args[2] = value
	cmdctx.Db.Begin()
	r, err := SetExHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisOkResp, r)
	assert.NoError(t, err)
	EqualGet(t, key, value, nil)

	args[1] = []byte("x")
	cmdctx.Db.Begin()
	r, err = SetExHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisIntegerResp, r)
	assert.NotNil(t, err)
}

func TestStringSetNx(t *testing.T) {
	args := make([][]byte, 2)
	key := []byte("setnx")
	args[0] = key
	args[1] = value
	cmdctx.Db.Begin()
	r, err := SetNxHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisOneResp, r)
	assert.NoError(t, err)
	EqualGet(t, key, value, nil)

	args[1] = []byte("v1")
	cmdctx.Db.Begin()
	r, err = SetNxHandler(args, cmdctx)
	assert.Equal(t, RedisZeroResp, r)
	cmdctx.Db.Commit()
	assert.NoError(t, err)
	EqualGet(t, key, value, nil)
}

func TestStringPSetEx(t *testing.T) {
	args := make([][]byte, 3)
	key := []byte("psetex")
	args[0] = key
	args[1] = []byte("100000")
	args[2] = value
	cmdctx.Db.Begin()
	r, err := PSetExHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisOkResp, r)
	assert.NoError(t, err)
	EqualGet(t, key, value, nil)

	args[1] = []byte("x")
	cmdctx.Db.Begin()
	r, err = PSetExHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisIntegerResp, r)
	assert.NotNil(t, err)
}

func TestStringRange(t *testing.T) {
	args := make([][]byte, 3)
	key := []byte("range")
	args[0] = key
	args[1] = []byte("10")
	args[2] = value

	cmdctx.Db.Begin()
	r, err := SetRangeHandler(args, cmdctx)
	cmdctx.Db.Commit()
	rr := &protocol.ReplyData{Type: protocol.REPLYBINT, Value: int64(len(args[2]) + 10)}
	assert.Equal(t, rr, r)
	assert.NoError(t, err)

	cmdctx.Db.Begin()
	r, err = SetRangeHandler(args, cmdctx)
	cmdctx.Db.Commit()
	rr = &protocol.ReplyData{Type: protocol.REPLYBINT, Value: int64(len(args[2]) + 10)}
	assert.Equal(t, rr, r)
	assert.NoError(t, err)

	args[1] = []byte("1073741824")
	cmdctx.Db.Begin()
	r, err = SetRangeHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisMaximumResp, r)
	assert.NotNil(t, err)
}

func TestStringIncr(t *testing.T) {
	args := make([][]byte, 1)
	args[0] = []byte("incr")
	cmdctx.Db.Begin()
	r, err := IncrHandler(args, cmdctx)
	cmdctx.Db.Commit()
	rr := &protocol.ReplyData{Type: protocol.REPLYBINT, Value: int64(1)}
	assert.Equal(t, rr, r)
	assert.NoError(t, err)

	args[0] = []byte("setex")
	cmdctx.Db.Begin()
	r, err = IncrHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisIntegerResp, r)
	assert.NotNil(t, err)
}

func TestStringIncrBy(t *testing.T) {
	args := make([][]byte, 2)
	args[0] = []byte("incrby")
	args[1] = []byte("2")
	cmdctx.Db.Begin()
	r, err := IncrByHandler(cmdctx, args)
	cmdctx.Db.Commit()
	rr := &protocol.ReplyData{Type: protocol.REPLYBINT, Value: int64(2)}
	assert.Equal(t, rr, r)
	assert.NoError(t, err)

	args[1] = []byte("-2")
	cmdctx.Db.Begin()
	r, err = IncrByHandler(cmdctx, args)
	cmdctx.Db.Commit()
	rr = &protocol.ReplyData{Type: protocol.REPLYBINT, Value: int64(0)}
	assert.Equal(t, rr, r)
	assert.NoError(t, err)

	//TODO bug

		args[1] = []byte("02")
		cmdctx.Db.Begin()
		r, err = IncrByHandler(cmdctx, args)
		cmdctx.Db.Commit()
		assert.Equal(t, RedisIntegerResp, r)
		assert.NotNil(t, err)

}

//bug
func TestStringIncrByFloat(t *testing.T) {
	/*
		args := make([][]byte, 2)
		args[0] = []byte("incrbyfloat")
		args[1] = []byte("2.0e2")
		cmdctx.Db.Begin()
		r, err := IncrByFloatHandler(cmdctx, args)
		cmdctx.Db.Commit()
		rr := &protocol.ReplyData{Type: protocol.REPLYBINT, Value: int64(200)}
		assert.Equal(t, rr, r)
		assert.NoError(t, err)

		args[1] = []byte("2.0e2")
		cmdctx.Db.Begin()
		r, err = IncrByFloatHandler(cmdctx, args)
		cmdctx.Db.Commit()
		rr = &protocol.ReplyData{Type: protocol.REPLYBINT, Value: int64(0)}
		assert.Equal(t, rr, r)
		assert.NoError(t, err)

			args[1] = []byte("02")
			cmdctx.Db.Begin()
			r, err = IncrByFloatHandler(cmdctx, args)
			cmdctx.Db.Commit()
			assert.Equal(t, RedisIntegerResp, r)
			assert.NotNil(t, err)
	//
}

func TestStringDecr(t *testing.T) {
	args := make([][]byte, 1)
	args[0] = []byte("decr")
	cmdctx.Db.Begin()
	r, err := DecrHandler(cmdctx, args)
	cmdctx.Db.Commit()
	rr := &protocol.ReplyData{Type: protocol.REPLYBINT, Value: int64(-1)}
	assert.Equal(t, rr, r)
	assert.NoError(t, err)

	args[0] = []byte("setex")
	cmdctx.Db.Begin()
	r, err = DecrHandler(cmdctx, args)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisIntegerResp, r)
	assert.NotNil(t, err)
}

func TestStringDecrBy(t *testing.T) {
	args := make([][]byte, 2)
	args[0] = []byte("decrby")
	args[1] = []byte("2")
	cmdctx.Db.Begin()
	r, err := DecrByHandler(cmdctx, args)
	cmdctx.Db.Commit()
	rr := &protocol.ReplyData{Type: protocol.REPLYBINT, Value: int64(-2)}
	assert.Equal(t, rr, r)
	assert.NoError(t, err)

	args[1] = []byte("-2")
	cmdctx.Db.Begin()
	r, err = DecrByHandler(cmdctx, args)
	cmdctx.Db.Commit()
	rr = &protocol.ReplyData{Type: protocol.REPLYBINT, Value: int64(0)}
	assert.Equal(t, rr, r)
	assert.NoError(t, err)

	//bug

		args[1] = []byte("02")
		r, err = DecrByHandler(cmdctx, args)
		assert.Equal(t, RedisIntegerResp, r)
		assert.NotNil(t, err)

}

func TestStringMset(t *testing.T) {
	args := make([][]byte, 4)
	args[0] = []byte("Mset1")
	args[1] = []byte("Mset1")
	args[2] = []byte("Mset2")
	args[3] = []byte("Mset2")
	cmdctx.Db.Begin()
	r, err := MSetHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisOkResp, r)
	assert.NoError(t, err)
	EqualMGet(t, [][]byte{args[0], args[2]}, [][]byte{args[1], args[3]}, nil)

	r, err = MSetHandler(args[:3], cmdctx)
	assert.Equal(t, RedisMsetResp, r)
	assert.NotNil(t, err)
}

func TestStringAppend(t *testing.T) {
	args := make([][]byte, 2)
	args[0] = []byte("Append")
	args[1] = []byte("Ap")
	cmdctx.Db.Begin()
	r, err := AppendHandler(args, cmdctx)
	cmdctx.Db.Commit()
	rr := &protocol.ReplyData{Type: protocol.REPLYBINT, Value: int64(len(args[1]))}
	assert.Equal(t, rr, r)
	assert.NoError(t, err)

	cmdctx.Db.Begin()
	r, err = AppendHandler(args, cmdctx)
	cmdctx.Db.Commit()
	rr = &protocol.ReplyData{Type: protocol.REPLYBINT, Value: int64(len(args[1]) * 2)}
	assert.Equal(t, rr, r)
	assert.NoError(t, err)
}
*/
