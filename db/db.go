package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"gitlab.meitu.com/platform/thanos/db/store"
)

var (
	// IsErrNotFound returns true if the key is not found, otherwise return false
	IsErrNotFound = store.IsErrNotFound
	// IsRetryableError returns true if the error is temporary and can be retried
	IsRetryableError = store.IsRetryableError

	// ErrTxnAlreadyBegin indicates that db is in a transaction now, you should not begin again
	ErrTxnAlreadyBegin = errors.New("transaction has already been begun")
	// ErrTxnNotBegin indicates that db is not in a transaction, and can not be commited or rollbacked
	ErrTxnNotBegin = errors.New("transaction dose not begin")
	// ErrTypeMismatch indicates object type of key is not as expect
	ErrTypeMismatch = errors.New("type mismatch")
	// ErrKeyNotFound key not exist
	ErrKeyNotFound = errors.New("key not found")
	// ErrPrecision list index reach precision limitatin
	ErrPrecision = errors.New("list reaches precision limitation, rebalance now")

	ErrTxnNil           = errors.New("txn is nil")
	ErrOutOfRange       = errors.New("error index/offset out of range")
	ErrInvalidLength    = errors.New("error data length is invalid for unmarshaler")
	ErrEncodingMismatch = errors.New("error object encoding type")
	ErrInternal         = errors.New("error tikv internal")
)

type Iterator store.Iterator

// BatchGetValues issue batch requests to get values
func BatchGetValues(txn *Transaction, keys [][]byte) ([][]byte, error) {
	return store.BatchGetValues(txn.txn, keys)
}

type DBID byte

func (id DBID) String() string {
	return fmt.Sprintf("%03d", id)
}
func (id DBID) Bytes() []byte {
	return []byte(id.String())
}
func toDBID(v []byte) DBID {
	id, _ := strconv.Atoi(string(v))
	return DBID(id)
}

// DB is a redis compatible data structure storage
type DB struct {
	Namespace string
	ID        DBID
	kv        *RedisStore
}

type RedisStore struct {
	store.Storage
}

func Open(addr string) (*RedisStore, error) {
	s, err := store.Open(addr)
	if err != nil {
		return nil, err
	}
	rs := &RedisStore{s}
	go StartGC(rs)
	go StartExpire(rs)

	return rs, nil
}

func (rds *RedisStore) DB(namesapce string, id int) *DB {
	return &DB{Namespace: namesapce, ID: DBID(id), kv: rds}
}

func (rds *RedisStore) Close() error {
	return rds.Close()
}

// Transaction is the interface of store tranaction
type Transaction struct {
	txn store.Transaction
	db  *DB
}

// Begin a transaction
func (db *DB) Begin() (*Transaction, error) {
	txn, err := db.kv.Begin()
	if err != nil {
		return nil, err
	}
	return &Transaction{txn: txn, db: db}, nil
}

// Commit a transaction
func (txn *Transaction) Commit(ctx context.Context) error {
	return txn.txn.Commit(ctx)
}

// Rollback a transaction
func (txn *Transaction) Rollback() error {
	return txn.txn.Rollback()
}

// List return a list object, a new list is created if the key dose not exist.
func (txn *Transaction) List(key []byte) (*List, error) {
	return GetList(txn, key)
}

// String return a string object
func (txn *Transaction) String(key []byte) (*String, error) {
	return GetString(txn, key)
}

func (txn *Transaction) Kv() *Kv {
	return GetKv(txn)
}

func (txn *Transaction) Hash(key []byte) (*Hash, error) {
	return GetHash(txn, key)
}

// Set returns a set object
func (txn *Transaction) Set(key []byte) (*Set, error) {
	return GetSet(txn, key)
}

// LockKeys tries to lock the entries with the keys in KV store.
func (txn *Transaction) LockKeys(keys ...[]byte) error {
	return store.LockKeys(txn.txn, keys)
}

func MetaKey(db *DB, key []byte) []byte {
	var mkey []byte
	mkey = append(mkey, []byte(db.Namespace)...)
	mkey = append(mkey, ':')
	mkey = append(mkey, db.ID.Bytes()...)
	mkey = append(mkey, ':', 'M', ':')
	mkey = append(mkey, key...)
	return mkey
}
func DataKey(db *DB, key []byte) []byte {
	var dkey []byte
	dkey = append(dkey, []byte(db.Namespace)...)
	dkey = append(dkey, ':')
	dkey = append(dkey, db.ID.Bytes()...)
	dkey = append(dkey, ':', 'D', ':')
	dkey = append(dkey, key...)
	return dkey
}
func DBPrefix(db *DB) []byte {
	var prefix []byte
	prefix = append(prefix, []byte(db.Namespace)...)
	prefix = append(prefix, ':')
	prefix = append(prefix, db.ID.Bytes()...)
	prefix = append(prefix, ':')
	return prefix
}
