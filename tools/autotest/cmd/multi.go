package cmd

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

type ExampleMulti struct {
	conn  redis.Conn
	value []interface{}
}

func NewExampleMulti(conn redis.Conn) *ExampleMulti {
	return &ExampleMulti{
		conn: conn,
	}
}

func (ms *ExampleMulti) MultiEqual(t *testing.T) {
	reply, err := redis.String(ms.conn.Do("multi"))
	assert.NoError(t, err)
	assert.Equal(t, "OK", reply)
}

func (ms *ExampleMulti) MultiEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ms.conn.Do("multi", args...)
	assert.EqualError(t, err, errValue)
}

func (ms *ExampleMulti) ExecEqual(t *testing.T) {
	reply, err := redis.MultiBulk(ms.conn.Do("exec"))
	assert.NoError(t, err)
	assert.Equal(t, ms.value, reply)
}

func (ms *ExampleMulti) ExecEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ms.conn.Do("exec", args...)
	if errValue != "" {
		assert.EqualError(t, err, errValue)
	}
}

func (ms *ExampleMulti) Cmd(t *testing.T) {
	reply, err := redis.String(ms.conn.Do("SET", "key-mulit-string", "value"))
	assert.Equal(t, "QUEUED", reply)
	assert.NoError(t, err)
	ms.value = append(ms.value, "OK")

	reply, err = redis.String(ms.conn.Do("lpush", "key-mulit-list", "value", "value", "value"))
	assert.Equal(t, "QUEUED", reply)
	assert.Nil(t, err)
	ms.value = append(ms.value, int64(3))
}
