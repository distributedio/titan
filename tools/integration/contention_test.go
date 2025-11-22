package integration

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/distributedio/titan/command"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

func TestHighContention(t *testing.T) {
	// Setup
	addr := "127.0.0.1:17370"
	auth := "contention_pass"
	
	// Note: SetAddr and SetAuth modify package-level globals.
	// This test assumes sequential execution with other tests in this package.
	SetAddr(addr)
	SetAuth(auth)
	
	go Start()
	// Give server time to start
	time.Sleep(500 * time.Millisecond)
	defer Close()

	// Generate auth token
	token, err := command.Token([]byte(auth), []byte("contention_ns"), time.Now().Unix())
	assert.NoError(t, err)

	// Initialize key
	initConn, err := redis.Dial("tcp", addr)
	if !assert.NoError(t, err) {
		return
	}
	defer func() {
		_ = initConn.Close()
	}()

	if _, err := initConn.Do("AUTH", token); !assert.NoError(t, err) {
		return
	}
	if _, err := initConn.Do("SET", "counter", 0); !assert.NoError(t, err) {
		return
	}

	// Run concurrent increments
	numClients := 10
	incrementsPerClient := 50 // 500 total increments
	key := "counter"

	var wg sync.WaitGroup
	wg.Add(numClients)

	errCh := make(chan error, numClients*incrementsPerClient)

	start := time.Now()

	for i := 0; i < numClients; i++ {
		go func(id int) {
			defer wg.Done()
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				errCh <- fmt.Errorf("client %d connect failed: %v", id, err)
				return
			}
			defer func() {
				_ = c.Close()
			}()

			if _, err := c.Do("AUTH", token); err != nil {
				errCh <- fmt.Errorf("client %d auth failed: %v", id, err)
				return
			}

			for j := 0; j < incrementsPerClient; j++ {
				_, err := c.Do("INCR", key)
				if err != nil {
					// If exponential backoff and retry logic works, we should see very few errors here
					// unless contention is extreme and max retries are exhausted.
					errCh <- fmt.Errorf("client %d incr %d failed: %v", id, j, err)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errCh)
	duration := time.Since(start)
	t.Logf("Performed %d increments in %v", numClients*incrementsPerClient, duration)

	// Check for errors
	errorCount := 0
	for err := range errCh {
		t.Error(err)
		errorCount++
	}
	assert.Equal(t, 0, errorCount, "Expected 0 errors with retry logic enabled")

	// Verify final result
	val, err := redis.Int(initConn.Do("GET", key))
	assert.NoError(t, err)
	assert.Equal(t, numClients*incrementsPerClient, val, "Counter should match expected value")
}