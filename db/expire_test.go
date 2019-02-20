package db

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getDB(t *testing.T) *Transaction {
	txn, err := mockDB.Begin()
	assert.NotNil(t, txn)
	assert.NoError(t, err)
	return txn
}

func Test_runExpire(t *testing.T) {

	//case 1 test hash
	// first add new expired hash
	key := "TestExpiredHash"
	expireAt := (time.Now().Unix() - 30) * int64(time.Second)
	hash, txn, err := getHash(t, []byte(key))
	oldID := hash.meta.ID
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.NotNil(t, hash)
	hash.HSet([]byte("field1"), []byte("val"))
	kv := GetKv(txn)
	err = kv.ExpireAt([]byte(key), expireAt)
	assert.NoError(t, err)
	txn.Commit(context.TODO())

	//second reset a expired hash
	hash, txn, err = getHash(t, []byte(key))
	newID := hash.meta.ID
	if bytes.Equal(oldID, newID) {
		assert.Fail(t, "old hash is not expired")
		return
	}
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.NotNil(t, hash)
	hash.HSet([]byte("field1"), []byte("val"))
	txn.Commit(context.TODO())

	txn = getDB(t)
	runExpire(txn.db, 1)
	gcKey := toTikvGCKey(toTikvDataKey([]byte(txn.db.Namespace), txn.db.ID, oldID))
	val, err := txn.t.Get(gcKey)
	assert.NoError(t, err)
	assert.Equal(t, val, []byte{0})

}
