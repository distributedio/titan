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

func TestExpireReliability(t *testing.T) {
	// Setup
	addr := "127.0.0.1:17371"
	auth := "background_pass"
	SetAddr(addr)
	SetAuth(auth)
	
	go Start()
	time.Sleep(500 * time.Millisecond)
	defer Close()

	token, err := command.Token([]byte(auth), []byte("bg_ns"), time.Now().Unix())
	assert.NoError(t, err)

	// Scenario: Create many short-lived keys while simultaneously updating other keys
	// to trigger conflicts if they land in same regions or just create general load.
	
	// Actually, to cause CONFLICT on expiration, we need to touch the SAME keys being expired.
	// But expiration deletes them.
	// If we SET a key that is about to expire, we might extend it or replace it.
	
	clientCount := 5
	keysPerClient := 50
	
	var wg sync.WaitGroup
	wg.Add(clientCount)
	
	for i := 0; i < clientCount; i++ {
		go func(id int) {
			defer wg.Done()
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				t.Errorf("connect failed: %v", err)
				return
			}
			defer func() {
				_ = c.Close()
			}()
			_, _ = c.Do("AUTH", token)

			for j := 0; j < keysPerClient; j++ {
				key := fmt.Sprintf("expire_key_%d_%d", id, j)
				// Set with 1 second expiry
				_, err := c.Do("SETEX", key, 1, "val")
				if err != nil {
					t.Errorf("SETEX failed: %v", err)
				}
				
				// Immediately try to update it to cause contention during expiration window?
				// 1s is long enough for set to finish.
				// We want to update it AROUND the time it expires.
				
				time.Sleep(900 * time.Millisecond)
				
				// Try to overwrite it (conflict with expire?)
				_, err = c.Do("SET", key, "new_val")
				// If expire happened first, this sets new key.
				// If set happened first, expire might kill this new key if ID didn't change? 
				// (Titan assigns new ID on SET usually).
				// If ID changed, old expire entry is stale (points to old ID).
				// So it handles it safely.
				
				if err != nil {
					t.Errorf("SET failed: %v", err)
				}
			}
		}(i)
	}
	
	wg.Wait()
	
	// Allow time for background tasks to settle
	time.Sleep(2 * time.Second)
	
	// Verification:
	// Keys should either be gone (if expire won last) or present (if set won last).
	// But we just want to ensure no panics or crashes, and that retry logic triggered if needed.
	// We can't easily assert state of keys because of race.
	
	// Check server is still responsive
	c, _ := redis.Dial("tcp", addr)
	defer func() {
		_ = c.Close()
	}()
	_, _ = c.Do("AUTH", token)
	_, err = c.Do("PING")
	assert.NoError(t, err, "Server should be alive")
}