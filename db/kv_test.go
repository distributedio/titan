package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func SetVal(t *testing.T, db *DB, key, val []byte) {
	txn, err := db.Begin()
	assert.NoError(t, err)
	str := NewString(txn, key)
	assert.NoError(t, err)
	err = str.Set(val)
	assert.NoError(t, err)
	txn.Commit(context.Background())
}

func CheckNotFoundKey(t *testing.T, db *DB, key []byte) (bool, error) {
	txn, err := db.Begin()
	assert.NoError(t, err)
	obj, err := txn.Object(key)
	txn.Commit(context.Background())
	if obj != nil {
		return false, err
	}
	return true, err
}

func EqualExpireAt(t *testing.T, db *DB, key []byte, expected int64) {
	txn, err := db.Begin()
	assert.NoError(t, err)
	obj, err := txn.Object(key)
	txn.Commit(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, obj)
	assert.Equal(t, expected, obj.ExpireAt)
}

func TestDelete(t *testing.T) {
	key := []byte("keys-key-del")
	val := []byte("keys-val-del")
	db := MockDB()
	SetVal(t, db, key, val)
	txn, err := db.Begin()
	assert.NoError(t, err)
	kv := txn.Kv()
	assert.NoError(t, err)
	keys := [][]byte{key}
	_, err = kv.Delete(keys)
	assert.NoError(t, err)
	txn.Commit(context.Background())
	notFound, _ := CheckNotFoundKey(t, db, key)
	assert.Equal(t, true, notFound)
}

func TestExists(t *testing.T) {
	db := MockDB()
	key := []byte("key-ex")
	val := []byte("val-ex")
	SetVal(t, db, key, val)
	txn, err := db.Begin()
	assert.NoError(t, err)
	kv := txn.Kv()
	assert.NoError(t, err)
	keys := [][]byte{key}
	_, err = kv.Exists(keys)
	txn.Commit(context.Background())
	assert.NoError(t, err)
}

func TestExpireAt(t *testing.T) {
	db := MockDB()
	key := []byte("key-ex")
	val := []byte("val-ex")
	SetVal(t, db, key, val)
	now := time.Now().UnixNano()

	time1 := now + int64(100*time.Second)
	txn, err := db.Begin()
	assert.NoError(t, err)
	kv := txn.Kv()
	assert.NoError(t, err)
	err = kv.ExpireAt(key, time1)
	txn.Commit(context.Background())
	assert.NoError(t, err)
	EqualExpireAt(t, db, key, time1)

}

func TestKeys(t *testing.T) {
	list := [][]byte{
		[]byte("keys"),
		[]byte("keys12"),
		[]byte("keys13"),
		[]byte("keys14"),
		[]byte("keys15"),
	}

	db := MockDB()
	val := []byte("val")
	for _, key := range list {
		SetVal(t, db, key, val)
	}

	txn, err := db.Begin()
	assert.NoError(t, err)
	kv := txn.Kv()
	assert.NoError(t, err)
	var actualkeys [][]byte
	call := func(key []byte, obj *Object) bool {
		actualkeys = append(actualkeys, key)
		return true
	}
	err = kv.Keys([]byte("keys"), call)
	assert.NoError(t, err)
	txn.Commit(context.Background())
	assert.Equal(t, list, actualkeys)

}

func TestRandomKey(t *testing.T) {
	list := [][]byte{
		[]byte("randomkey1"),
		[]byte("randomkey2"),
		[]byte("randomkey3"),
		[]byte("randomkey4"),
		[]byte("randomkey5"),
	}

	db := MockDB()
	val := []byte("val")
	for _, key := range list {
		SetVal(t, db, key, val)
	}

	mapkey := make(map[string]int)
	for i := 0; i < 5; i++ {
		txn, err := db.Begin()
		assert.NoError(t, err)
		kv := txn.Kv()
		assert.NoError(t, err)

		tmp1, err := kv.RandomKey()
		assert.NoError(t, err)
		mapkey[string(tmp1)]++
		txn.Commit(context.Background())

	}
	// assert.NotEqual(t, 1, len(mapkey))
}
