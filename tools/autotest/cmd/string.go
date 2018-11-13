package cmd

import (
	"strconv"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

//ExampleString verify the string command
type ExampleString struct {
	values map[string]string
	conn   redis.Conn
}

//NewExampleString create new string object
func NewExampleString(conn redis.Conn) *ExampleString {
	return &ExampleString{
		values: make(map[string]string),
		conn:   conn,
	}
}

//SetEqual verify that the return value of the set operation is correct
func (es *ExampleString) SetEqual(t *testing.T, key string, value string) {
	es.values[key] = value
	reply, err := redis.String(es.conn.Do("SET", key, value))
	assert.Equal(t, "OK", reply)
	assert.NoError(t, err)
	data, err := redis.Bytes(es.conn.Do("GET", key))
	assert.Equal(t, value, string(data))
	assert.NoError(t, err)
}

//GetEqual verify that the return value of the get operation is correct
func (es *ExampleString) GetEqual(t *testing.T, key string) {
	data, err := redis.Bytes(es.conn.Do("GET", key))
	assert.Equal(t, es.values[key], string(data))
	assert.NoError(t, err)
}

//SetEqualErr verify that the return value of the set operation is correct
func (es *ExampleString) SetEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := es.conn.Do("set", args...)
	assert.EqualError(t, err, errValue)
}

//GetEqualErr verify that the return value of the set operation is correct
func (es *ExampleString) GetEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := es.conn.Do("get", args...)
	assert.EqualError(t, err, errValue)
}

//SetNxEqualErr verify that the return value of the set operation is correct
func (es *ExampleString) SetNxEqual(t *testing.T, key string, value string) {
	es.values[key] = value
	reply, err := redis.String(es.conn.Do("SETNX", key, value))
	assert.Equal(t, "OK", reply)
	assert.NoError(t, err)
	data, err := redis.Bytes(es.conn.Do("GET", key))
	assert.Equal(t, value, string(data))
	assert.NoError(t, err)
}

//SetExEqualErr verify that the return value of the set operation is correct
func (es *ExampleString) SetExEqual(t *testing.T, key string, value string, delta int) {
	es.values[key] = value
	reply, err := redis.String(es.conn.Do("SETEX", key, value, delta))
	assert.Equal(t, "OK", reply)
	assert.NoError(t, err)
	data, err := redis.Bytes(es.conn.Do("GET", key))
	assert.Equal(t, value, string(data))
	assert.NoError(t, err)
	time.Sleep(time.Second * time.Duration(delta))
	data, err = redis.Bytes(es.conn.Do("GET", key))
	assert.Equal(t, value, "")
	assert.NoError(t, err)
}

//PSetexEqualErr verify that the return value of the set operation is correct
func (es *ExampleString) PSetexEqual(t *testing.T, key string, value string, delta int) {
	es.values[key] = value
	reply, err := redis.String(es.conn.Do("PSETEX", key, value, delta))
	assert.Equal(t, "OK", reply)
	assert.NoError(t, err)
	data, err := redis.Bytes(es.conn.Do("GET", key))
	assert.Equal(t, value, string(data))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * time.Duration(delta))
	data, err = redis.Bytes(es.conn.Do("GET", key))
	assert.Equal(t, value, "")
	assert.NoError(t, err)
}

//SetRangeEqualErr verify that the return value of the set operation is correct
func (es *ExampleString) SetRangeEqual(t *testing.T, key string, value string) {
}

//MSetNxRangeEqualErr verify that the return value of the set operation is correct
func (es *ExampleString) MSetNxEqual(t *testing.T, expectValue int, args ...string) {
	as := make([]interface{}, len(args))
	for i := 0; i < len(args); i = i + 2 {
		es.values[args[i]] = args[i+1]
		as[i] = args[i]
		as[i+1] = args[i+1]
	}
	reply, err := redis.Int(es.conn.Do("MSETNx", as...))
	assert.NoError(t, err)
	assert.Equal(t, expectValue, reply)
}

//AppendEqual verify that the return value of the append operation is correct
func (es *ExampleString) AppendEqual(t *testing.T, key string, value string) {
	if v, ok := es.values[key]; ok {
		es.values[key] = v + value
	} else {
		es.values[key] = value
	}
	reply, err := redis.Int(es.conn.Do("Append", key, value))
	assert.Equal(t, len(es.values[key]), reply)
	assert.NoError(t, err)
	data, err := redis.Bytes(es.conn.Do("GET", key))
	assert.Equal(t, es.values[key], string(data))
	assert.NoError(t, err)
}

//AppendEqualErr TODO
func (es *ExampleString) AppendEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := es.conn.Do("append", args...)
	assert.EqualError(t, err, errValue)
}

//IncrEqual verify that the return value of the incr operation is correct
func (es *ExampleString) IncrEqual(t *testing.T, key string) {
	var vi int
	if v, ok := es.values[key]; ok {
		vi, _ = strconv.Atoi(v)
		vi = vi + 1
		es.values[key] = strconv.Itoa(vi)
	} else {
		vi = 1
		es.values[key] = strconv.Itoa(vi)
	}
	reply, err := redis.Int(es.conn.Do("incr", key))
	assert.Equal(t, vi, reply)
	assert.NoError(t, err)
}

//DecrEqual verify that the return value of the incr operation is correct
func (es *ExampleString) DecrEqual(t *testing.T, key string) {
}

//IncrByEqual verify that the return value of the incr operation is correct
func (es *ExampleString) IncrByEqual(t *testing.T, key string) {
}

//IncrByFloatEqual verify that the return value of the incr operation is correct
func (es *ExampleString) IncrByFloatEqual(t *testing.T, key string) {
}

//DecrbyEqual verify that the return value of the incr operation is correct
func (es *ExampleString) DecrByEqual(t *testing.T, key string) {
}

//IncrEqualErr verify that the return value of the incr operation is correct
func (es *ExampleString) IncrEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := es.conn.Do("incr", args...)
	assert.EqualError(t, err, errValue)
}

//StrlenEqual verify that the return value of the strlen operation is correct
func (es *ExampleString) StrlenEqual(t *testing.T, key string) {
	var vi int
	if v, ok := es.values[key]; ok {
		vi = len(v)
	} else {
		vi = 0
	}
	reply, err := redis.Int(es.conn.Do("strlen", key))
	assert.Equal(t, vi, reply)
	assert.NoError(t, err)
}

//StrlenEqualErr verify that the return value of the strlen operation is correct
func (es *ExampleString) StrlenEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := es.conn.Do("strlen", args...)
	assert.EqualError(t, err, errValue)
}

//MGetEqual verify that the return value of the mget operation is correct
func (es *ExampleString) MGetEqual(t *testing.T, args ...string) {
	as := make([]interface{}, len(args))
	for i, a := range args {
		as[i] = a
	}
	reply, err := redis.Strings(es.conn.Do("MGET", as...))
	assert.NoError(t, err)
	for i, a := range args {
		if expects, ok := es.values[a]; ok {
			assert.Equal(t, expects, reply[i])
		} else {
			assert.Equal(t, "", reply[i])
		}
	}
}

//MGetEqualErr verify that the return value of the mget operation is correct
func (es *ExampleString) MGetEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := es.conn.Do("Mget", args...)
	assert.EqualError(t, err, errValue)
}

//MSetEqual verify that the return value of the mset operation is correct
func (es *ExampleString) MSetEqual(t *testing.T, args ...string) {
	as := make([]interface{}, len(args))
	for i := 0; i < len(args); i = i + 2 {
		es.values[args[i]] = args[i+1]
		as[i] = args[i]
		as[i+1] = args[i+1]
	}
	reply, err := redis.String(es.conn.Do("MSET", as...))
	assert.NoError(t, err)
	assert.Equal(t, "OK", reply)

	for i := 0; i < len(args); i = i + 2 {
		es.GetEqual(t, args[i])
	}
}

//MSetEqualErr verify that the return value of the mset operation is correct
func (es *ExampleString) MSetEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := es.conn.Do("Mset", args)
	assert.EqualError(t, err, errValue)
}
