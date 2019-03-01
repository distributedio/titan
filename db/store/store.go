package store

import (
	"strings"
	"unsafe"

	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/store/tikv"
)

// Transaction options
const (
	// PresumeKeyNotExists indicates that when dealing with a Get operation but failing to read data from cache,
	// we presume that the key does not exist in Store. The actual existence will be checked before the
	// transaction's commit.
	// This option is an optimization for frequent checks during a transaction, e.g. batch inserts.
	PresumeKeyNotExists Option = iota + 1
	// PresumeKeyNotExistsError is the option key for error.
	// When PresumeKeyNotExists is set and condition is not match, should throw the error.
	PresumeKeyNotExistsError
	// BinlogInfo contains the binlog data and client.
	BinlogInfo
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
	return kv.IsRetryableError(err)
}

func IsConflictError(err error) bool {
	return kv.ErrLockConflict.Equal(err)

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
	return txn.LockKeys(kvKeys...)
}

// BatchGetValues issue batch requests to get values
func BatchGetValues(txn Transaction, keys [][]byte) (map[string][]byte, error) {
	kvkeys := *(*[]kv.Key)(unsafe.Pointer(&keys))
	return txn.BatchGet(kvkeys)
}

func SetOption(txn Transaction, opt Option, val interface{}) {
	txn.SetOption(kv.Option(opt), val)
}

func DelOption(txn Transaction, opt Option) {
	txn.DelOption(kv.Option(opt))
}
