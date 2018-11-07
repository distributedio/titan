package cmd

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

//ExampleSystem verify the system command
type ExampleSystem struct {
	conn redis.Conn
}

//NewExampleSystem create new system object
func NewExampleSystem(conn redis.Conn) *ExampleSystem {
	return &ExampleSystem{
		conn: conn,
	}
}

//AuthEqual verify that the return value of the auth operation is correct
func (es *ExampleSystem) AuthEqual(t *testing.T, password string) {
	reply, err := redis.String(es.conn.Do("auth", password))
	assert.NoError(t, err)
	assert.Equal(t, "OK", reply)
}

//AuthEqualErr verify that the return value of the auth operation is correct
func (es *ExampleSystem) AuthEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := es.conn.Do("auth", args)
	assert.EqualError(t, err, errValue)
}

//PingEqual verify that the return value of the ping operation is correct
func (es *ExampleSystem) PingEqual(t *testing.T) {
	reply, err := redis.String(es.conn.Do("ping", "hello"))
	assert.NoError(t, err)
	assert.Equal(t, "hello", reply)

	reply, err = redis.String(es.conn.Do("ping"))
	assert.NoError(t, err)
	assert.Equal(t, "PONG", reply)
}

//PingEqualErr verify that the return value of the ping operation is correct
func (es *ExampleSystem) PingEqualErr(t *testing.T, errValue string, args ...interface{}) {
	_, err := es.conn.Do("ping", args)
	assert.EqualError(t, err, errValue)
}
