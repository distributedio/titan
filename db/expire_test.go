package db

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/distributedio/titan/db/store"
	"github.com/stretchr/testify/assert"
)

func getTxn(t *testing.T) *Transaction {
	txn, err := mockDB.Begin()
	assert.NotNil(t, txn)
	assert.NoError(t, err)
	return txn
}

func Test_runExpire(t *testing.T) {
	hashKey := []byte("TestExpiredHash")
	strKey := []byte("TestExpiredString")
	expireAt := (time.Now().Unix() - 30) * int64(time.Second)
	hashCall := func(t *testing.T, key []byte) []byte {
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

		hash, txn, err = getHash(t, []byte(key))
		newID := hash.meta.ID
		if bytes.Equal(oldID, newID) {
			assert.Fail(t, "old hash is not expired")
			return nil
		}
		assert.NoError(t, err)
		assert.NotNil(t, txn)
		assert.NotNil(t, hash)
		hash.HSet([]byte("field1"), []byte("val"))
		txn.Commit(context.TODO())
		return oldID
	}

	stringCall := func(t *testing.T, key []byte) []byte {
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

		txn = getTxn(t)
		s, err := GetString(txn, key)
		assert.NoError(t, err)
		newID := s.Meta.ID
		if bytes.Equal(oldID, newID) {
			assert.Fail(t, "old hash is not expired")
			return nil
		}
		err = s.Set([]byte("val"))
		assert.NoError(t, err)
		txn.Commit(context.TODO())
		return oldID
	}

	type args struct {
		key  []byte
		call func(*testing.T, []byte) []byte
	}
	type want struct {
		gckey bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "TestExpiredHash",
			args: args{
				key:  hashKey,
				call: hashCall,
			},
			want: want{
				gckey: true,
			},
		},
		{
			name: "TestExpiredString",
			args: args{
				key:  strKey,
				call: stringCall,
			},
			want: want{
				gckey: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.args.call(t, tt.args.key)
			txn := getTxn(t)
			for i := 0; i < EXPIRE_HASH_NUM; i++ {
				expireHash := fmt.Sprintf("%04d", i)
				runExpire(txn.db, 1, expireHash, 0)
			}
			txn.Commit(context.TODO())

			txn = getTxn(t)
			gcKey := toTikvGCKey(toTikvDataKey([]byte(txn.db.Namespace), txn.db.ID, id))

			_, err := txn.t.Get(gcKey)
			txn.Commit(context.TODO())
			if tt.want.gckey {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, true, store.IsErrNotFound(err))
			}
		})
	}

}

func Test_doExpire(t *testing.T) {
	initHash := func(t *testing.T, key []byte) []byte {
		hash, txn, err := getHash(t, key)
		assert.NoError(t, err)
		assert.NotNil(t, txn)
		assert.NotNil(t, hash)
		hash.HSet([]byte("field1"), []byte("val"))
		txn.Commit(context.TODO())
		return hash.meta.ID
	}

	expireAt := (time.Now().Unix() - 30) * int64(time.Second)
	hashCall := func(t *testing.T, key []byte) ([]byte, []byte) {
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

		hash, txn, err = getHash(t, []byte(key))
		newID := hash.meta.ID
		if bytes.Equal(oldID, newID) {
			assert.Fail(t, "old hash is not expired")
			return nil, nil
		}
		assert.NoError(t, err)
		assert.NotNil(t, txn)
		assert.NotNil(t, hash)
		hash.HSet([]byte("field1"), []byte("val"))
		txn.Commit(context.TODO())
		return oldID, newID
	}

	hashId := initHash(t, []byte("TestExpiredHash"))
	rHashId, nmateHashId := hashCall(t, []byte("TestExpiredRewriteHash"))

	dirtyDataHashID := initHash(t, []byte("TestExpiredHash_dirty_data"))
	rDHashId, _ := hashCall(t, []byte("TestExpiredRewriteHash_dirty_data"))
	txn := getTxn(t)
	type args struct {
		mkey []byte
		id   []byte
		tp   byte
	}
	type want struct {
		gckey bool
	}

	tests := []struct {
		name     string
		args     args
		want     want
		expireAt int64
	}{
		{
			name: "TestExpiredHash",
			args: args{
				mkey: MetaKey(txn.db, []byte("TestExpiredHash")),
				id:   hashId,
			},
			want: want{
				gckey: true,
			},
			expireAt: 0,
		},
		{
			name: "TestExpiredRewriteHash",
			args: args{
				mkey: MetaKey(txn.db, []byte("TestExpiredRewriteHash")),
				id:   rHashId,
			},
			want: want{
				gckey: true,
			},
			expireAt: expireAt,
		},
		{
			name: "TestExpiredNotExistsMeta",
			args: args{
				mkey: MetaKey(txn.db, []byte("TestExpiredRewriteHash")),
				id:   nmateHashId,
			},
			want: want{
				gckey: true,
			},
			expireAt: 0,
		},
		{
			name: "TestExpiredHash_dirty_data",
			args: args{
				mkey: MetaKey(txn.db, []byte("TestExpiredHash_dirty_data")),
				id:   dirtyDataHashID,
				tp:   byte(ObjectHash),
			},
			want: want{
				gckey: true,
			},
			expireAt: 0,
		},
		{
			name: "TestExpiredRewriteHash_dirty_data",
			args: args{
				mkey: MetaKey(txn.db, []byte("TestExpiredRewriteHash_dirty_data")),
				id:   rDHashId,
				tp:   byte(ObjectHash),
			},
			want: want{
				gckey: true,
			},
			expireAt: expireAt,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn := getTxn(t)
			id := tt.args.id
			if tt.args.tp == byte(ObjectHash) {
				id = append(id, tt.args.tp)
			}
			err := doExpire(txn, tt.args.mkey, id, "", tt.expireAt)
			txn.Commit(context.TODO())
			assert.NoError(t, err)

			txn = getTxn(t)
			gcKey := toTikvGCKey(toTikvDataKey([]byte(txn.db.Namespace), txn.db.ID, tt.args.id))

			_, err = txn.t.Get(gcKey)
			txn.Commit(context.TODO())
			if tt.want.gckey {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, true, store.IsErrNotFound(err))
			}
		})
	}

}
