package integration

import (
	"testing"
	"time"

	"github.com/distributedio/titan/command"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

func TestTitanIntegration(t *testing.T) {
	// Configure server
	addr := "127.0.0.1:18888"
	auth := "testpassword"
	SetAddr(addr)
	SetAuth(auth)

	// Start server in background
	go Start()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)
	defer Close()

	// Connect to server
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		// Try waiting a bit longer if first attempt fails
		time.Sleep(500 * time.Millisecond)
		conn, err = redis.Dial("tcp", addr)
	}
	assert.NoError(t, err, "Should connect to Titan server")
	if err != nil {
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	// Authenticate
	// Generate a valid token using the auth secret
	token, err := command.Token([]byte(auth), []byte("test_ns"), time.Now().Unix())
	assert.NoError(t, err, "Should generate token")

	_, err = conn.Do("AUTH", token)
	assert.NoError(t, err, "AUTH should succeed")

	// Ping
	pong, err := redis.String(conn.Do("PING"))
	assert.NoError(t, err, "PING should succeed")
	assert.Equal(t, "PONG", pong, "PING should return PONG")

	// Test Set/Get to verify basic functionality
	_, err = conn.Do("SET", "test_key", "test_value")
	assert.NoError(t, err, "SET should succeed")

	val, err := redis.String(conn.Do("GET", "test_key"))
	assert.NoError(t, err, "GET should succeed")
	assert.Equal(t, "test_value", val, "GET should return correct value")
}