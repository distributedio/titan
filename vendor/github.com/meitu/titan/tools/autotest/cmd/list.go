package cmd

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

//ExampleList the key command
//mapList record the key and value of the operation
type ExampleList struct {
	mapList map[string][]string
	conn    redis.Conn
}

//NewExampleList create list object
func NewExampleList(conn redis.Conn) *ExampleList {
	return &ExampleList{
		conn:    conn,
		mapList: make(map[string][]string),
	}
}

//LsetEqual verify that the return value of the lset key operation is correct
func (el *ExampleList) LsetEqual(t *testing.T, key string, index int, value string) {
	if _, ok := el.mapList[key]; !ok {
		el.mapList[key] = make([]string, 0, 10)
	}
	if index < 0 {
		index = -index
	}
	el.mapList[key][index] = value
	reply, err := redis.String(el.conn.Do("lset", key, index, value))
	assert.Equal(t, "OK", reply)
	assert.Nil(t, err)

	llen, err := redis.Int(el.conn.Do("llen", key))
	assert.Equal(t, len(el.mapList[key]), llen)
	assert.Nil(t, err)
}

//LsetEqualErr verify that the return err value of the Lset key operation is correct
func (el *ExampleList) LsetEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := el.conn.Do("lset", args...)
	assert.EqualError(t, err, errValue)
}

//LpushEqual verify that the return err value of the lpush key operation is correct
func (el *ExampleList) LpushEqual(t *testing.T, key string, values ...string) {
	req := make([]interface{}, 0, len(values))
	tmp := make([]string, 0, len(values))
	if _, ok := el.mapList[key]; !ok {
		el.mapList[key] = make([]string, 0, 0)
	}
	req = append(req, key)
	for _, value := range values {
		req = append(req, value)
	}
	for i := len(values) - 1; i >= 0; i-- {
		tmp = append(tmp, values[i])
	}
	tmp = append(tmp, el.mapList[key]...)
	el.mapList[key] = tmp

	reply, err := redis.Int(el.conn.Do("lpush", req...))
	assert.Equal(t, len(el.mapList[key]), reply)
	assert.Nil(t, err)
}

//LpushEqualErr verify that the return err value of the Lpush key operation is correct
func (el *ExampleList) LpushEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := el.conn.Do("lpush", args...)
	assert.EqualError(t, err, errValue)
}

//LpopEqual verify that the return err value of the Lpop key operation is correct
func (el *ExampleList) LpopEqual(t *testing.T, key string) {
	if vs, ok := el.mapList[key]; ok {
		v := vs[0]
		el.mapList[key] = el.mapList[key][1:]
		reply, err := redis.String(el.conn.Do("lpop", key))
		assert.Equal(t, v, reply)
		assert.Nil(t, err)
	} else {
		reply, err := redis.String(el.conn.Do("lpop", key))
		assert.Equal(t, "", reply)
		assert.EqualError(t, err, "redigo: nil returned")
	}
}

//LpopEqualErr verify that the return err value of the Lpop key operation is correct
func (el *ExampleList) LpopEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := el.conn.Do("lpop", args...)
	assert.EqualError(t, err, errValue)
}

//LindexEqual verify that the return err value of the Lindex key operation is correct
func (el *ExampleList) LindexEqual(t *testing.T, key string, index int) {
	if index <= 0 {
		index = -index
	}
	if vs, ok := el.mapList[key]; ok {
		v := vs[index]
		el.mapList[key] = el.mapList[key][:len(el.mapList[key])]
		reply, err := redis.String(el.conn.Do("lindex", key, index))
		assert.Equal(t, v, reply)
		assert.Nil(t, err)
	} else {
		reply, err := redis.String(el.conn.Do("lindex", key, index))
		assert.Equal(t, "", reply)
		assert.EqualError(t, err, "redigo: nil returned")
	}
}

//LindexEqualErr verify that the return err value of the Lindex key operation is correct
func (el *ExampleList) LindexEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := el.conn.Do("lindex", args...)
	assert.EqualError(t, err, errValue)
}

//LrangeEqual verify that the return value of the Lrange key operation is correct
func (el *ExampleList) LrangeEqual(t *testing.T, key string, start, end int) {
	if start > len(el.mapList[key]) {
		reply, err := redis.Strings(el.conn.Do("lrange", key, start, end))
		assert.Equal(t, []string{}, reply)
		assert.Nil(t, err)
		return
	}
	v, ok := el.mapList[key]
	if !ok {
		reply, err := redis.Strings(el.conn.Do("lrange", key, start, end))
		assert.Equal(t, v, reply)
		assert.Nil(t, err)
		return
	}
	if start < 0 {
		start = -start
	}
	if end < 0 {
		end = -end
	}
	if end <= len(el.mapList[key]) {
		v = v[start : end+1]
	}
	reply, err := redis.Strings(el.conn.Do("lrange", key, start, end))
	assert.Equal(t, v, reply)
	assert.Nil(t, err)
}

//LrangeEqualErr verify that the return err value of the lrange key operation is correct
func (el *ExampleList) LrangeEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := el.conn.Do("lrange", args...)
	assert.EqualError(t, err, errValue)
}

//RpushEqual verify that the return value of the Rpush key operation is correct
func (el *ExampleList) RpushEqual(t *testing.T, key string, values []string) {
	req := make([]interface{}, 0, len(values))
	req = append(req, key)
	if _, ok := el.mapList[key]; !ok {
		el.mapList[key] = make([]string, 0, 0)
	}
	for _, value := range values {
		el.mapList[key] = append(el.mapList[key], value)
		req = append(req, value)
	}
	reply, err := redis.Int(el.conn.Do("rpush", req...))
	assert.Equal(t, len(el.mapList[key]), reply)
	assert.Nil(t, err)
}

//RpushEqualErr verify that the return err value of the Rpush key operation is correct
func (el *ExampleList) RpushEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := el.conn.Do("Rpush", args...)
	assert.EqualError(t, err, errValue)
}

//RpopEqual verify that the return value of the rpop key operation is correct
func (el *ExampleList) RpopEqual(t *testing.T, key string) {
	v := el.mapList[key][len(el.mapList[key])-1]
	el.mapList[key] = el.mapList[key][:len(el.mapList[key])]
	reply, err := redis.String(el.conn.Do("rpop", key))
	assert.Equal(t, v, reply)
	assert.Nil(t, err)
}

//RpopEqualErr verify that the return err value of the rpop key operation is correct
func (el *ExampleList) RpopEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := el.conn.Do("Rpop", args...)
	assert.EqualError(t, err, errValue)
}

//LlenEqual verify that the return value of the Llen key operation is correct
func (el *ExampleList) LlenEqual(t *testing.T, key string) {
	reply, err := redis.Int(el.conn.Do("llen", key))
	assert.Equal(t, len(el.mapList[key]), reply)
	assert.Nil(t, err)
}

//LlenEqualErr verify that the return err value of the Llen key operation is correct
func (el *ExampleList) LlenEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := el.conn.Do("Llen", args...)
	assert.EqualError(t, err, errValue)
}
