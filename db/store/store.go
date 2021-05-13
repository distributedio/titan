package store

import (
	"context"
	"strings"
	"unsafe"

	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/store/tikv"
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

//type rename tidb kv type
type (
	// Storage defines the interface for storage.
	Storage kv.Storage
	// Transaction defines the interface for operations inside a Transaction.
	Transaction kv.Transaction
	// Iterator is the interface for a iterator on KV store.
	Iterator kv.Iterator
	// Option is used for customizing kv store's behaviors during a transaction.
	Option kv.Option
)

//Open create tikv db ,create fake db if addr contains mockaddr
func Open(addrs string) (r Storage, e error) {
	if strings.Contains(addrs, MockAddr) {
		return MockOpen(addrs)
	}
	return tikv.Driver{}.Open(addrs)
}

// IsErrNotFound checks if err is a kind of NotFound error.
func IsErrNotFound(err error) bool {
	return kv.IsErrNotFound(err)
}

// IsRetryableError checks if err is a kind of RetryableError error.
func IsRetryableError(err error) bool {
	return kv.IsTxnRetryableError(err)
}

func IsConflictError(err error) bool {
	return kv.ErrWriteConflict.Equal(err)

}

func RunInNewTxn(store Storage, retryable bool, f func(txn kv.Transaction) error) error {
	return kv.RunInNewTxn(store, retryable, f)
}

// LockKeys tries to lock the entries with the keys in KV store.
func LockKeys(txn Transaction, keys [][]byte) error {
	kvKeys := make([]kv.Key, len(keys))
	for i := range keys {
		kvKeys[i] = kv.Key(keys[i])
	}
	return txn.LockKeys(context.Background(), &kv.LockCtx{}, kvKeys...)
}

// BatchGetValues issue batch requests to get values
func BatchGetValues(ctx context.Context, txn Transaction, keys [][]byte) (map[string][]byte, error) {
	kvkeys := *(*[]kv.Key)(unsafe.Pointer(&keys))
	return txn.BatchGet(ctx, kvkeys)
}

func SetOption(txn Transaction, opt Option, val interface{}) {
	txn.SetOption(kv.Option(opt), val)
}

func DelOption(txn Transaction, opt Option) {
	txn.DelOption(kv.Option(opt))
}
