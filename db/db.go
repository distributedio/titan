package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"github.com/distributedio/titan/conf"
	"github.com/distributedio/titan/db/store"
)

var (
	// ErrTypeMismatch indicates object type of key is not as expect
	ErrTypeMismatch = errors.New("type mismatch")

	// ErrKeyNotFound key not exist
	ErrKeyNotFound = errors.New("key not found")

	// ErrInteger valeu is not interge
	ErrInteger = errors.New("value is not an integer or out of range")

	// ErrPrecision list index reach precision limitatin
	ErrPrecision = errors.New("list reaches precision limitation, rebalance now")

	// ErrOutOfRange index/offset out of range
	ErrOutOfRange = errors.New("error index/offset out of range")

	// ErrInvalidLength data length is invalid for unmarshaler"
	ErrInvalidLength = errors.New("error data length is invalid for unmarshaler")

	// ErrEncodingMismatch object encoding type
	ErrEncodingMismatch = errors.New("error object encoding type")

	// ErrStorageRetry storage err and try again later
	ErrStorageRetry = errors.New("Storage err and try again later")

	//ErrSetNilValue means the value corresponding to key is a non-zero value
	ErrSetNilValue = errors.New("The value corresponding to key is a non-zero value")

	// IsErrNotFound returns true if the key is not found, otherwise return false
	IsErrNotFound = store.IsErrNotFound

	// IsRetryableError returns true if the error is temporary and can be retried
	IsRetryableError = store.IsRetryableError

	// IsConflictError return true if the error is conflict
	IsConflictError = store.IsConflictError

	// sysNamespace default namespace
	sysNamespace = "$sys"

	// sysDatabaseID default db id
	sysDatabaseID = 0

	NilValue = []byte{0}
)

// Iterator store.Iterator
type Iterator store.Iterator

// DBID is the redis database ID
type DBID byte

// String returns the string format of DBID
func (id DBID) String() string {
	return fmt.Sprintf("%03d", id)
}

// Bytes DBID returns a byte slice
func (id DBID) Bytes() []byte {
	return []byte(id.String())
}

func toDBID(v []byte) DBID {
	id, _ := strconv.Atoi(string(v))
	return DBID(id)
}

// BatchGetValues issues batch requests to get values
func BatchGetValues(txn *Transaction, keys [][]byte) ([][]byte, error) {
	kvs, err := store.BatchGetValues(txn.t, keys)
	if err != nil {
		return nil, err
	}
	values := make([][]byte, len(keys))
	for i := range keys {
		values[i] = kvs[string(keys[i])]
	}
	return values, nil
}

// DB is a redis compatible data structure storage
type DB struct {
	Namespace string
	ID        DBID
	kv        *RedisStore
}

// RedisStore wraps store.Storage
type RedisStore struct {
	store.Storage
	conf *conf.TiKV
}

// Open a storage instance
func Open(conf *conf.TiKV) (*RedisStore, error) {
	s, err := store.Open(conf.PdAddrs)
	if err != nil {
		return nil, err
	}
	rds := &RedisStore{Storage: s, conf: conf}
	sysdb := rds.DB(sysNamespace, sysDatabaseID)

	if err := RegisterTask(sysdb, conf); err != nil {
		return nil, err
	}
	return rds, nil
}

// DB returns a DB object with sepcific ID
func (rds *RedisStore) DB(namesapce string, id int) *DB {
	return &DB{Namespace: namesapce, ID: DBID(id), kv: rds}
}

// Close the storage instance
func (rds *RedisStore) Close() error {
	return rds.Storage.Close()
}

// Transaction supplies transaction for data structures
type Transaction struct {
	t  store.Transaction
	db *DB
}

// Begin a transaction
func (db *DB) Begin() (*Transaction, error) {
	txn, err := db.kv.Begin()
	if err != nil {
		return nil, err
	}
	return &Transaction{t: txn, db: db}, nil
}

// Prefix returns the prefix of a DB object
func (db *DB) Prefix() []byte {
	return dbPrefix(db.Namespace, db.ID.Bytes())
}

// Commit a transaction
func (txn *Transaction) Commit(ctx context.Context) error {
	return txn.t.Commit(ctx)
}

// Rollback a transaction
func (txn *Transaction) Rollback() error {
	return txn.t.Rollback()
}

// listOption for get a list
type listOption struct {
	useZip bool
}

// ListOption customize how to get a list
type ListOption func(o *listOption)

// UseZip will create a ziplist if set when the key is missing
func UseZip() ListOption {
	return func(o *listOption) {
		o.useZip = true
	}
}

// List return a lists object, a new list is created if the key dose not exist.
func (txn *Transaction) List(key []byte, opts ...ListOption) (List, error) {
	return GetList(txn, key, opts...)
}

// String returns a string object, but the object is unsafe, maybe the object is expire,or not exist
func (txn *Transaction) String(key []byte) (*String, error) {
	return GetString(txn, key)
}

// Strings returns a slice of String
func (txn *Transaction) Strings(keys [][]byte) ([]*String, error) {
	sobjs := make([]*String, len(keys))
	tkeys := make([][]byte, len(keys))
	for i, key := range keys {
		tkeys[i] = MetaKey(txn.db, key)
	}
	mdata, err := store.BatchGetValues(txn.t, tkeys)
	if err != nil {
		return nil, err
	}
	for i, key := range keys {
		obj := NewString(txn, key)
		if data, ok := mdata[string(tkeys[i])]; ok {
			if err := obj.decode(data); err != nil {
				if logEnv := zap.L().Check(zap.DebugLevel, "Strings decoded value error"); logEnv != nil {
					logEnv.Write(zap.ByteString("key", key), zap.Error(err))
				}
			}
		}
		sobjs[i] = obj
	}
	return sobjs, nil
}

// Kv returns a kv object
func (txn *Transaction) Kv() *Kv {
	return GetKv(txn)
}

// Hash returns a hash object
func (txn *Transaction) Hash(key []byte) (*Hash, error) {
	return GetHash(txn, key)
}

// Set returns a set object
func (txn *Transaction) Set(key []byte) (*Set, error) {
	return GetSet(txn, key)
}

// ZSet returns a zset object
func (txn *Transaction) ZSet(key []byte) (*ZSet, error) {
	return GetZSet(txn, key)
}

// LockKeys tries to lock the entries with the keys in KV store.
func (txn *Transaction) LockKeys(keys ...[]byte) error {
	return store.LockKeys(txn.t, keys)
}

// MetaKey build to metakey from a redis key
func MetaKey(db *DB, key []byte) []byte {
	var mkey []byte
	mkey = append(mkey, []byte(db.Namespace)...)
	mkey = append(mkey, ':')
	mkey = append(mkey, db.ID.Bytes()...)
	mkey = append(mkey, ':', 'M', ':')
	mkey = append(mkey, key...)
	return mkey
}

// DataKey builds a datakey from a redis key
func DataKey(db *DB, key []byte) []byte {
	var dkey []byte
	dkey = append(dkey, []byte(db.Namespace)...)
	dkey = append(dkey, ':')
	dkey = append(dkey, db.ID.Bytes()...)
	dkey = append(dkey, ':', 'D', ':')
	dkey = append(dkey, key...)
	return dkey
}

func dbPrefix(ns string, id []byte) []byte {
	prefix := []byte(ns)
	prefix = append(prefix, ':')
	if id != nil {
		prefix = append(prefix, id...)
		prefix = append(prefix, ':')
	}
	return prefix
}
