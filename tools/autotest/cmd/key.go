package cmd

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

//ExampleKey verify the key command
type ExampleKey struct {
	conn redis.Conn
}

//NewExampleKey create key object
func NewExampleKey(conn redis.Conn) *ExampleKey {
	return &ExampleKey{
		conn: conn,
	}
}

//DelEqual verify that the return value of the del key operation is correct
func (ek *ExampleKey) DelEqual(t *testing.T, expectReply int, keys ...string) {
	req := make([]interface{}, len(keys))
	for i, eky := range keys {
		req[i] = eky
	}
	reply, err := redis.Int(ek.conn.Do("Del", req...))
	assert.Equal(t, expectReply, reply)
	assert.NoError(t, err)
}

//DelEqualErr verify that the return error value of the del key operation is correct
func (ek *ExampleKey) DelEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ek.conn.Do("del", args...)
	assert.EqualError(t, err, errValue)
}

//ExistsEqual verify that the return value of the exists key operation is correct
func (ek *ExampleKey) ExistsEqual(t *testing.T, expectReply int, keys ...string) {
	req := make([]interface{}, len(keys))
	for i, eky := range keys {
		req[i] = eky
	}
	reply, err := redis.Int(ek.conn.Do("exists", req...))
	assert.Equal(t, expectReply, reply)
	assert.NoError(t, err)
}

//ExistsEqualErr verify that the return error value of the exists key operation is correct
func (ek *ExampleKey) ExistsEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ek.conn.Do("exists", args...)
	assert.EqualError(t, err, errValue)
}

//TTLEqual verify that the return value of the ttl key operation is correct
func (ek *ExampleKey) TTLEqual(t *testing.T, key string, expectReply int) {
	reply, err := redis.Int(ek.conn.Do("ttl", key))
	assert.Equal(t, expectReply, reply)
	assert.NoError(t, err)
}

//TTLEqualErr verify that the return error value of the ttl key operation is correct
func (ek *ExampleKey) TTLEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ek.conn.Do("ttl", args...)
	assert.EqualError(t, err, errValue)
}

//Info TODO
func (ek *ExampleKey) Info(t *testing.T, key string, expectReply interface{}) {

}

//InfoEqualErr TODO
func (ek *ExampleKey) InfoEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ek.conn.Do("info", args...)
	assert.EqualError(t, err, errValue)
}

//ScanEqual verify that the return value of the scan key operation is correct
//default scan all key in store
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

//ScanEqualErr verify that the return err value of the scan key operation is correct
func (ek *ExampleKey) ScanEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ek.conn.Do("scan", args...)
	assert.EqualError(t, err, errValue)
}

//RandomKeyEqual verify that the return value of the random key operation is correct
func (ek *ExampleKey) RandomKeyEqual(t *testing.T) {
	_, err := ek.conn.Do("RANDOMKEY")
	assert.NoError(t, err)
}

//RandomKeyEqualErr verify that the return err value of the random key operation is correct
func (ek *ExampleKey) RandomKeyEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ek.conn.Do("Randomkey", args...)
	assert.EqualError(t, err, errValue)
}

//ExpireEqual verify that the return value of the expire key operation is correct
func (ek *ExampleKey) ExpireEqual(t *testing.T, key string, value, expectValue int) {
	reply, err := redis.Int(ek.conn.Do("expire", key, value))
	assert.NoError(t, err)
	assert.Equal(t, expectValue, reply)
}

//ExpireEqualErr verify that the err return value of the expire key operation is correct
func (ek *ExampleKey) ExpireEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ek.conn.Do("expire", args...)
	assert.EqualError(t, err, errValue)
}
