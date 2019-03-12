package store

import (
	"strings"
	"unsafe"

	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/store/tikv"
)

//type rename tidb kv type
type (
	// Storage defines the interface for storage.
	Storage kv.Storage
	// Transaction defines the interface for operations inside a Transaction.
	Transaction kv.Transaction
	// Iterator is the interface for a iterator on KV store.
	Iterator kv.Iterator
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
