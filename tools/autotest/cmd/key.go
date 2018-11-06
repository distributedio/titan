package cmd

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

type ExampleKey struct {
	conn redis.Conn
}

func NewExampleKey(conn redis.Conn) *ExampleKey {
	return &ExampleKey{
		conn: conn,
	}
}

func (ek *ExampleKey) DelEqual(t *testing.T, expectReply int, keys ...string) {
	req := make([]interface{}, len(keys))
	for i, eky := range keys {
		req[i] = eky
	}
	reply, err := redis.Int(ek.conn.Do("Del", req...))
	assert.Equal(t, expectReply, reply)
	assert.NoError(t, err)
}

func (ek *ExampleKey) DelEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ek.conn.Do("del", args...)
	assert.EqualError(t, err, errValue)
}

func (ek *ExampleKey) ExistsEqual(t *testing.T, expectReply int, keys ...string) {
	req := make([]interface{}, len(keys))
	for i, eky := range keys {
		req[i] = eky
	}
	reply, err := redis.Int(ek.conn.Do("exists", req...))
	assert.Equal(t, expectReply, reply)
	assert.NoError(t, err)
}

func (ek *ExampleKey) ExistsEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ek.conn.Do("exists", args...)
	assert.EqualError(t, err, errValue)
}

func (ek *ExampleKey) TTLEqual(t *testing.T, key string, expectReply int) {
	reply, err := redis.Int(ek.conn.Do("ttl", key))
	assert.Equal(t, expectReply, reply)
	assert.NoError(t, err)
}

func (ek *ExampleKey) TTLEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ek.conn.Do("ttl", args...)
	assert.EqualError(t, err, errValue)
}

//TODO
func (ek *ExampleKey) Info(t *testing.T, key string, expectReply interface{}) {

}

func (ek *ExampleKey) InfoEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ek.conn.Do("info", args...)
	assert.EqualError(t, err, errValue)
}

// 默认扫全量数据
func (ek *ExampleKey) ScanEqual(t *testing.T, match string, expectCount int) {
	var reply interface{}
	var err error
	if match == "" {
		reply, err = ek.conn.Do("Scan", 0, "count", 10000)
	} else {
		reply, err = ek.conn.Do("Scan", 0, "match", match, "count", 10000)
	}
	r, _ := redis.MultiBulk(reply, err)
	strs, _ := redis.Strings(r[1], err)
	assert.Equal(t, expectCount, len(strs))
	assert.NoError(t, err)
}

func (ek *ExampleKey) ScanEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ek.conn.Do("scan", args...)
	assert.EqualError(t, err, errValue)
}

func (ek *ExampleKey) RandomKeyqual(t *testing.T) {
	_, err := ek.conn.Do("RANDOMKEY")
	assert.NoError(t, err)
}

func (ek *ExampleKey) RandomKeyEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ek.conn.Do("Randomkey", args...)
	assert.EqualError(t, err, errValue)
}

func (ek *ExampleKey) ExpireEqual(t *testing.T, key string, value, expectValue int) {
	reply, err := redis.Int(ek.conn.Do("expire", key, value))
	assert.NoError(t, err)
	assert.Equal(t, expectValue, reply)
}

func (ek *ExampleKey) ExpireEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ek.conn.Do("expire", args...)
	assert.EqualError(t, err, errValue)
}
