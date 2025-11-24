package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	//nolint:depguard
	tikverr "github.com/tikv/client-go/v2/error"
	//nolint:depguard
	"github.com/tikv/client-go/v2/txnkv"
	//nolint:depguard
	"github.com/tikv/client-go/v2/txnkv/txnutil"
)

// Transaction options
const (
	// BinlogInfo contains the binlog data and client.
	BinlogInfo Option = iota + 1
	// SchemaChecker is used for checking schema-validity.
	SchemaChecker
	// IsolationLevel sets isolation level for current transaction. The default level is SI.
	IsolationLevel
	// Priority marks the priority of this transaction.
	Priority
	// NotFillCache makes this request do not touch the LRU cache of the underlying storage.
	NotFillCache
	// SyncLog decides whether the WAL(write-ahead log) of this request should be synchronized.
	SyncLog
	// KeyOnly retrieve only keys, it can be used in scan now.
	KeyOnly
	// Pessimistic is defined for pessimistic lock
	Pessimistic
	// SnapshotTS is defined to set snapshot ts.
	SnapshotTS
	// Set replica read
	ReplicaRead
	// Set task ID
	TaskID
	// InfoSchema is schema version used by txn startTS.
	InfoSchema
	// CollectRuntimeStats is used to enable collect runtime stats.
	CollectRuntimeStats
	// SchemaAmender is used to amend mutations for pessimistic transactions
	SchemaAmender
	// SampleStep skips 'SampleStep - 1' number of keys after each returned key.
	SampleStep
	// CommitHook is a callback function called right after the transaction gets committed
	CommitHook
	// EnableAsyncCommit indicates whether async commit is enabled
	EnableAsyncCommit
	// Enable1PC indicates whether one-phase commit is enabled
	Enable1PC
	// GuaranteeExternalConsistency indicates whether to guarantee external consistency at the cost of an extra tso request before prewrite
	GuaranteeExternalConsistency
	// TxnScope indicates which @@txn_scope this transaction will work with.
	TxnScope
)

// Priority value for transaction priority.
const (
	PriorityNormal = iota
	PriorityLow
	PriorityHigh
)

// type rename tidb kv type
type (
	// Storage defines the interface for storage.
	Storage interface {
		Begin() (Transaction, error)
		Close() error
		CurrentTimestamp(scope string) (uint64, error)
	}
	// Transaction defines the interface for operations inside a Transaction.
	Transaction = *txnkv.KVTxn
	// Iterator is the interface for a iterator on KV store.
	Iterator interface {
		Next() error
		Value() []byte
		Key() []byte
		Valid() bool
		Close()
	}
	// Option is used for customizing kv store's behaviors during a transaction.
	Option int
)

// storageWrapper wrapper to match old Storage interface
type storageWrapper struct {
	*txnkv.Client
}

//nolint:wrapcheck
func (s *storageWrapper) Begin() (Transaction, error) {
	return s.Client.Begin()
}

//nolint:wrapcheck
func (s *storageWrapper) Close() error {
	return s.Client.Close()
}

//nolint:wrapcheck
func (s *storageWrapper) CurrentTimestamp(scope string) (uint64, error) {
	return s.Client.CurrentTimestamp(scope)
}

// Open create tikv db ,create fake db if addr contains mockaddr
//
//nolint:ireturn,wrapcheck
func Open(addrs string) (Storage, error) {
	if strings.Contains(addrs, MockAddr) {
		return MockOpen(addrs)
	}
	pdAddrs := strings.Split(strings.TrimPrefix(addrs, "tikv://"), ",")
	client, err := txnkv.NewClient(pdAddrs)
	if err != nil {
		return nil, err
	}
	return &storageWrapper{Client: client}, nil
}

// IsErrNotFound checks if err is a kind of NotFound error.
func IsErrNotFound(err error) bool {
	return tikverr.IsErrNotFound(err)
}

// IsRetryableError checks if err is a kind of RetryableError error.
func IsRetryableError(err error) bool {
	if tikverr.IsErrWriteConflict(err) {
		return true
	}
	var retryableErr *tikverr.ErrRetryable
	return errors.As(err, &retryableErr)
}

func IsConflictError(err error) bool {
	return tikverr.IsErrWriteConflict(err)
}

//nolint:wrapcheck
func RunInNewTxn(store Storage, retryable bool, maxRetries int, baseDelay time.Duration, txnFunc func(txn Transaction) error) error {
	timeout := time.After(30 * time.Second)
	for attempt := 0; attempt < maxRetries; attempt++ {
		txn, err := store.Begin()
		if err != nil {
			return err
		}

		err = txnFunc(txn)
		if err != nil {
			_ = txn.Rollback()
			if retryable && IsRetryableError(err) {
				if attempt < maxRetries-1 {
					delay := baseDelay * time.Duration(1<<uint(attempt))
					zap.L().Info("txn retry",
						zap.Int("attempt", attempt+1),
						zap.Duration("delay", delay),
						zap.Error(err))
					select {
					case <-timeout:
						return fmt.Errorf("transaction timeout after %d retries", attempt)
					case <-time.After(delay):
						continue
					}
				}
			}
			return err
		}

		if err = txn.Commit(context.Background()); err == nil {
			return nil
		}

		if retryable && IsRetryableError(err) {
			if attempt < maxRetries-1 {
				delay := baseDelay * time.Duration(1<<uint(attempt))
				zap.L().WithOptions(zap.AddStacktrace(zap.FatalLevel)).Info("txn commit retry",
					zap.Int("attempt", attempt+1),
					zap.Duration("delay", delay),
					zap.String("error", err.Error()))
				select {
				case <-timeout:
					return fmt.Errorf("transaction timeout after %d retries", attempt)
				case <-time.After(delay):
					continue
				}
			}
		}
		return err
	}
	return fmt.Errorf("transaction failed after %d attempts", maxRetries)
}

// LockKeys tries to lock the entries with the keys in KV store.
func LockKeys(txn Transaction, keys [][]byte) error {
	return txn.LockKeys(context.Background(), nil, keys...)
}

// BatchGetValues issue batch requests to get values
func BatchGetValues(ctx context.Context, txn Transaction, keys [][]byte) (map[string][]byte, error) {
	return txn.BatchGet(ctx, keys)
}

//nolint:cyclop,exhaustive
func SetOption(txn Transaction, opt Option, val interface{}) {
	switch opt {
	case Enable1PC:
		if v, ok := val.(bool); ok && v {
			txn.SetEnable1PC(true)
		}
	case EnableAsyncCommit:
		if v, ok := val.(bool); ok && v {
			txn.SetEnableAsyncCommit(true)
		}
	case GuaranteeExternalConsistency:
		if v, ok := val.(bool); ok && v {
			txn.SetCausalConsistency(true)
		}
	case Pessimistic:
		if v, ok := val.(bool); ok && v {
			txn.SetPessimistic(true)
		}
	case Priority:
		if v, ok := val.(int); ok {
			if v == PriorityHigh {
				txn.SetPriority(txnutil.PriorityHigh)
			}
		}
	}
}

func DelOption(_ Transaction, _ Option) {
	// DelOption is not supported in client-go v2 txn interface directly or not needed
}
