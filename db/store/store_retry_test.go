package store

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	tikverr "github.com/tikv/client-go/v2/error"
)

func TestRunInNewTxn_RetryLimit(t *testing.T) {
	s, err := Open(MockAddr)
	assert.NoError(t, err)
	defer func() {
		_ = s.Close()
	}()

	// Simulate a retryable error
	retryErr := &tikverr.ErrRetryable{Retryable: "retry me"}

	count := 0
	txnFunc := func(txn Transaction) error {
		count++
		return retryErr
	}

	start := time.Now()
	// Use a small delay for testing
	// RunInNewTxn signature will be changed to:
	// func RunInNewTxn(store Storage, retryable bool, maxRetries int, baseDelay time.Duration, txnFunc func(txn Transaction) error) error
	err = RunInNewTxn(s, true, 3, 10*time.Millisecond, txnFunc)
	duration := time.Since(start)

	assert.Error(t, err)
	// assert.Contains(t, err.Error(), "transaction timeout") // We return original error on exhaustion
	assert.True(t, tikverr.IsErrWriteConflict(err) || IsRetryableError(err))
	
	// Should have retried 3 times (initial + 3 retries = 4 calls? No, usually maxRetries limits the attempts or the retries)
	// If loop is: for attempt := 0; attempt < maxRetries; attempt++ 
	// Then it runs maxRetries times.
	// If I pass 3, I expect 3 calls.
	assert.Equal(t, 3, count)
	
	// Expected delay:
	// Attempt 1: 0 delay (starts immediately) -> fails
	// Wait 10ms
	// Attempt 2: -> fails
	// Wait 20ms
	// Attempt 3: -> fails
	// Returns error.
	// Total wait: 10 + 20 = 30ms.
	assert.GreaterOrEqual(t, duration, 30*time.Millisecond)
}

func TestRunInNewTxn_SuccessAfterRetry(t *testing.T) {
	s, err := Open(MockAddr)
	assert.NoError(t, err)
	defer func() {
		_ = s.Close()
	}()

	retryErr := &tikverr.ErrRetryable{Retryable: "retry me"}

	count := 0
	txnFunc := func(txn Transaction) error {
		count++
		if count < 3 {
			return retryErr
		}
		return nil
	}

	err = RunInNewTxn(s, true, 5, 10*time.Millisecond, txnFunc)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestRunInNewTxn_NonRetryableError(t *testing.T) {
	s, err := Open(MockAddr)
	assert.NoError(t, err)
	defer func() {
		_ = s.Close()
	}()

	fatalErr := errors.New("fatal error")

	count := 0
	txnFunc := func(txn Transaction) error {
		count++
		return fatalErr
	}

	err = RunInNewTxn(s, true, 5, 10*time.Millisecond, txnFunc)
	assert.Equal(t, fatalErr, err)
	assert.Equal(t, 1, count)
}