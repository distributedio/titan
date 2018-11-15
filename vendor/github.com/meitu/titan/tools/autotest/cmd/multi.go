package cmd

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

//ExampleMulti verify the multi command
type ExampleMulti struct {
	conn  redis.Conn
	value []interface{}
}

//NewExampleMulti create new multi object
func NewExampleMulti(conn redis.Conn) *ExampleMulti {
	return &ExampleMulti{
		conn: conn,
	}
}

//MultiEqual verify that the return value of the multi operation is correct
func (ms *ExampleMulti) MultiEqual(t *testing.T) {
	reply, err := redis.String(ms.conn.Do("multi"))
	assert.NoError(t, err)
	assert.Equal(t, "OK", reply)
}

//MultiEqualErr verify that the return err value of the multi operation is correct
func (ms *ExampleMulti) MultiEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ms.conn.Do("multi", args...)
	assert.EqualError(t, err, errValue)
}

//ExecEqual verify that the return err value of the exec operation is correct
func (ms *ExampleMulti) ExecEqual(t *testing.T) {
	reply, err := redis.MultiBulk(ms.conn.Do("exec"))
	assert.NoError(t, err)
	assert.Equal(t, ms.value, reply)
}

//ExecEqualErr verify that the return err value of the exec operation is correct
func (ms *ExampleMulti) ExecEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := ms.conn.Do("exec", args...)
	if errValue != "" {
		assert.EqualError(t, err, errValue)
	}
}

//Cmd analog message sending
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
