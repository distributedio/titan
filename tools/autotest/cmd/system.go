package cmd

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

type ExampleSystem struct {
	conn redis.Conn
}

func NewExampleSystem(conn redis.Conn) *ExampleSystem {
	return &ExampleSystem{
		conn: conn,
	}
}

func (es *ExampleSystem) AuthEqual(t *testing.T, password string) {
	reply, err := redis.String(es.conn.Do("auth", password))
	assert.NoError(t, err)
	assert.Equal(t, "OK", reply)
}
func (es *ExampleSystem) AuthEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := es.conn.Do("auth", args)
	assert.EqualError(t, err, errValue)
}

func (es *ExampleSystem) PingEqual(t *testing.T) {
	reply, err := redis.String(es.conn.Do("ping", "hello"))
	assert.NoError(t, err)
	assert.Equal(t, "hello", reply)

	reply, err = redis.String(es.conn.Do("ping"))
	assert.NoError(t, err)
	assert.Equal(t, "PONG", reply)
}

func (es *ExampleSystem) PingEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := es.conn.Do("ping", args)
	assert.EqualError(t, err, errValue)
}
