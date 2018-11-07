package store

import (
	"strings"

	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/store/tikv"
)

type Storage kv.Storage
type Transaction kv.Transaction
type Iterator kv.Iterator

func Open(addrs string) (r Storage, e error) {
	if strings.Contains(addrs, MockAddr) {
		return MockOpen(addrs)
	}
	return tikv.Driver{}.Open(addrs)
}

func IsErrNotFound(err error) bool {
	return kv.IsErrNotFound(err)
}

func IsRetryableError(err error) bool {
	return kv.IsRetryableError(err)
}

func RunInNewTxn(store Storage, retryable bool, f func(txn kv.Transaction) error) error {
	return kv.RunInNewTxn(store, retryable, f)
}

func LockKeys(txn Transaction, keys [][]byte) error {
	kvKeys := make([]kv.Key, len(keys))
	for i := range keys {
		kvKeys[i] = kv.Key(keys[i])
	}
	return txn.LockKeys(kvKeys...)
}

// BatchGetValues issue batch requests to get values
func BatchGetValues(txn Transaction, keys [][]byte) ([][]byte, error) {
	kvkeys := make([]kv.Key, len(keys))
	for i := range keys {
		kvkeys[i] = keys[i]
	}
	kvs, err := kv.BatchGetValues(txn, kvkeys)
	if err != nil {
		return nil, err
	}

	values := make([][]byte, len(keys))
	for i := range keys {
		values[i] = kvs[string(keys[i])]
	}

	return values, nil
}
