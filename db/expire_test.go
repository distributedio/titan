package db

import (
	"bytes"
	"context"
	"fmt"
	"math"
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
			runExpire(txn.db, 1)
			txn.Commit(context.TODO())

			txn = getTxn(t)
			gcKey := toTiKVGCKey(toTiKVDataKey([]byte(txn.db.Namespace), txn.db.ID, id))

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
		name string
		args args
		want want
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn := getTxn(t)
			id := tt.args.id
			if tt.args.tp == byte(ObjectHash) {
				id = append(id, tt.args.tp)
			}
			err := doExpire(txn, tt.args.mkey, id)
			txn.Commit(context.TODO())
			assert.NoError(t, err)

			txn = getTxn(t)
			gcKey := toTiKVGCKey(toTiKVDataKey([]byte(txn.db.Namespace), txn.db.ID, tt.args.id))

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

func TestScanExpiration(t *testing.T) {
	var at []int64
	var mkeys [][]byte

	// setUp fill the expiration list
	now := Now()
	setUp := func() {
		txn, err := mockDB.Begin()
		assert.NoError(t, err)

		// cleanup the keys left by other tests(TODO these dirty data should be deleted where it is generated)
		iter, err := txn.t.Iter(expireKeyPrefix, nil)
		assert.NoError(t, err)
		defer iter.Close()
		for iter.Valid() && iter.Key().HasPrefix(expireKeyPrefix) {
			txn.t.Delete(iter.Key())
			iter.Next()
		}

		for i := 0; i < 10; i++ {
			ts := now - 10 + int64(i)*int64(time.Second)
			mkey := MetaKey(txn.db, []byte(fmt.Sprintf("expire_key_%d", i)))
			err := expireAt(txn.t, mkey, mkey, ObjectString, 0, ts)
			assert.NoError(t, err)

			at = append(at, ts)
			mkeys = append(mkeys, mkey)
		}
		assert.NoError(t, txn.Commit(context.Background()))
	}
	tearDown := func() {
		txn, err := mockDB.Begin()
		assert.NoError(t, err)
		for i := range at {
			assert.NoError(t, unExpireAt(txn.t, mkeys[i], at[i]))
		}
		assert.NoError(t, txn.Commit(context.Background()))
	}

	setUp()

	type args struct {
		from  int64
		to    int64
		count int64
	}
	type want struct {
		s int // start index of the result
		e int // end index of the result(not included)
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{"escan 0 max 10", args{0, math.MaxInt64, 10}, want{0, 10}},
		{"escan 0 max 1", args{0, math.MaxInt64, 1}, want{0, 1}},
		{"escan 0 0 1", args{0, 0, 1}, want{0, 0}},
		{"escan max max 1", args{math.MaxInt64, math.MaxInt64, 1}, want{0, 0}},
		{"escan 0 max 20", args{0, math.MaxInt64, 10}, want{0, 10}},
		{"escan at[2] at[8] 10", args{at[2], at[8], 10}, want{2, 8}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)

			txn, err := mockDB.Begin()
			assert.NoError(t, err)

			tses, keys, err := ScanExpiration(txn, tt.args.from, tt.args.to, tt.args.count)
			assert.NoError(t, err)
			assert.NoError(t, txn.Commit(context.Background()))

			assert.Equal(t, tt.want.e-tt.want.s, len(tses))
			assert.Equal(t, tt.want.e-tt.want.s, len(keys))
			for i := range tses {
				assert.Equal(t, at[tt.want.s+i], tses[i])
				assert.Equal(t, mkeys[tt.want.s+i], keys[i])
			}
		})
	}

	tearDown()
}
