package command

import (
	"testing"

	 "gitlab.meitu.com/platform/thanos/db"
)



func TestGet(t *testing.T){

}

func TestSet(t * testing.T){


}




















/*
//TODO
func EqualGet(t *testing.T, key []byte, value []byte, e error) {
	cmdctx.Db.Begin()
	r, err := GetHandler([][]byte{key}, cmdctx)
	cmdctx.Db.Commit()
	re := &protocol.ReplyData{Type: protocol.REPLYBULK, Value: value}
	assert.Equal(t, re, r)
	assert.Equal(t, e, err)
}

func EqualStrlen(t *testing.T, key []byte, ll int64) {
	cmdctx.Db.Begin()
	r, err := StrlenHandler([][]byte{key}, cmdctx)
	cmdctx.Db.Commit()
	re := &protocol.ReplyData{Type: protocol.REPLYBINT, Value: ll}
	assert.Equal(t, re, r)
	assert.NoError(t, err)
}

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

var (
	value = []byte("value")
)

func SetEXS(key []byte) [][]byte {
	args := make([][]byte, 5)
	args[0] = key
	args[1] = []byte(value)
	args[2] = []byte("ex")
	args[3] = []byte("1000")
	args[4] = []byte("nx")
	return args
}

func SetPXS(key []byte) [][]byte {
	args := make([][]byte, 5)
	args[0] = key
	args[1] = []byte(value)
	args[2] = []byte("px")
	args[3] = []byte("1000")
	args[4] = []byte("nx")
	return args
}

func SetFour(key []byte) [][]byte {
	args := make([][]byte, 4)
	args[0] = key
	args[1] = []byte("value")
	args[2] = []byte("px")
	args[3] = []byte("1000")
	return args
}

// test set ex NX|XX|未知
func TestStringSetEXS(t *testing.T) {

	key := []byte("setexs")
	args := SetEXS(key)

	cmdctx.Db.Begin()
	r, err := SetHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisOkResp, r)
	assert.NoError(t, err)
	EqualGet(t, key, value, nil)
	EqualStrlen(t, key, int64(len(value)))

	//修改key失败
	args = SetEXS(key)
	args[1] = []byte("v2")
	cmdctx.Db.Begin()
	r, err = SetHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisNilResp, r)
	assert.NoError(t, err)
	EqualGet(t, key, value, nil)

	args = SetEXS(key)
	args[1] = []byte("v2")
	args[4] = []byte("xx")
	cmdctx.Db.Begin()
	r, err = SetHandler(args, cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisOkResp, r)
	assert.NoError(t, err)
	EqualGet(t, key, []byte("v2"), nil)
	EqualStrlen(t, key, int64(len("v2")))

	// 测试nx
	// 修改key 失败
	args = SetEXS(key)
	args[1] = []byte("value")
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
	args[3] = []byte("ex")
	args[4] = []byte("1000")

	cmdctx.Db.Begin()
	r, err = SetHandler(args, cmdctx)
	cmdctx.Db.Begin()
	assert.Equal(t, RedisOkResp, r)
	cmdctx.Db.Commit()
	assert.NoError(t, err)
	EqualGet(t, key, []byte("v1"), nil)

	//异常测试
	args = SetEXS(key)
	cmdctx.Db.Begin()
	r, err = SetHandler(args[:3], cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisSyntaxResp, r)
	assert.NotNil(t, err)

	//异常测试
	args = SetEXS(key)
	args[3] = []byte("bx")
	cmdctx.Db.Begin()
	SetHandler(args[:3], cmdctx)
	cmdctx.Db.Commit()
	assert.Equal(t, RedisSyntaxResp, r)
	assert.NotNil(t, err)

}

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

