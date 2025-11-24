package main

import (
	"context"
	"testing"
	"time"

	"github.com/distributedio/titan/db"
	"github.com/distributedio/titan/db/store"
	"github.com/stretchr/testify/assert"
)

func TestPrefixNext(t *testing.T) {
	tests := []struct {
		input    []byte
		expected []byte
	}{
		{[]byte{0x01}, []byte{0x02}},
		{[]byte{0x01, 0xFF}, []byte{0x02, 0x00}},
		{[]byte{0xFF}, nil},
		{[]byte{0xFF, 0xFF}, nil},
		{[]byte{}, nil},
	}

	for _, test := range tests {
		result := prefixNext(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func TestRunExpire(t *testing.T) {
	s, err := store.MockOpen("mocktikv://")
	assert.NoError(t, err)
	defer func() {
		_ = s.Close()
	}()

	opt := &option{
		db:         1,
		namespace:  "test",
		batch:      100,
		dry:        false,
		toleration: 0,
	}

	database := &db.DB{Namespace: opt.namespace, ID: db.DBID(opt.db)}

	txn, err := s.Begin()
	assert.NoError(t, err)

	now := db.Now()

	// Key1: Expired String
	key1 := []byte("key1")
	mkey1 := db.MetaKey(database, key1)
	obj1 := &db.Object{
		ID:        db.UUID(),
		CreatedAt: now,
		UpdatedAt: now,
		ExpireAt:  now - int64(time.Minute), // Expired
		Type:      db.ObjectString,
		Encoding:  db.ObjectEncodingRaw,
	}
	val1 := db.EncodeObject(obj1)
	val1 = append(val1, []byte("value1")...)
	err = txn.Set(mkey1, val1)
	assert.NoError(t, err)

	// Key2: Not expired String
	key2 := []byte("key2")
	mkey2 := db.MetaKey(database, key2)
	obj2 := &db.Object{
		ID:        db.UUID(),
		CreatedAt: now,
		UpdatedAt: now,
		ExpireAt:  now + int64(time.Minute), // Not expired
		Type:      db.ObjectString,
		Encoding:  db.ObjectEncodingRaw,
	}
	val2 := db.EncodeObject(obj2)
	val2 = append(val2, []byte("value2")...)
	err = txn.Set(mkey2, val2)
	assert.NoError(t, err)

	// Key3: No expire String
	key3 := []byte("key3")
	mkey3 := db.MetaKey(database, key3)
	obj3 := &db.Object{
		ID:        db.UUID(),
		CreatedAt: now,
		UpdatedAt: now,
		ExpireAt:  0, // No expire
		Type:      db.ObjectString,
		Encoding:  db.ObjectEncodingRaw,
	}
	val3 := db.EncodeObject(obj3)
	val3 = append(val3, []byte("value3")...)
	err = txn.Set(mkey3, val3)
	assert.NoError(t, err)

	// Key4: Expired Hash
	key4 := []byte("key4")
	mkey4 := db.MetaKey(database, key4)
	obj4 := &db.Object{
		ID:        db.UUID(),
		CreatedAt: now,
		UpdatedAt: now,
		ExpireAt:  now - int64(time.Minute),
		Type:      db.ObjectHash,
		Encoding:  db.ObjectEncodingHT,
	}
	val4 := db.EncodeObject(obj4)
	err = txn.Set(mkey4, val4)
	assert.NoError(t, err)

	// Add data fields for the hash
	dkey4 := db.DataKey(database, obj4.ID)
	dkey4Field := append(dkey4, []byte(":field1")...)
	err = txn.Set(dkey4Field, []byte("hashval"))
	assert.NoError(t, err)

	err = txn.Commit(context.Background())
	assert.NoError(t, err)

	// Run expire
	err = runExpire(s, opt)
	assert.NoError(t, err)

	// Verify
	txn, err = s.Begin()
	assert.NoError(t, err)
	defer func() {
		_ = txn.Rollback()
	}()

	// Key1 should be gone
	_, err = txn.Get(context.Background(), mkey1)
	assert.True(t, store.IsErrNotFound(err))

	// Key2 should exist
	_, err = txn.Get(context.Background(), mkey2)
	assert.NoError(t, err)

	// Key3 should exist
	_, err = txn.Get(context.Background(), mkey3)
	assert.NoError(t, err)

	// Key4 (Meta) should be gone
	_, err = txn.Get(context.Background(), mkey4)
	assert.True(t, store.IsErrNotFound(err))

	// Key4 (Data) should be gone
	_, err = txn.Get(context.Background(), dkey4Field)
	assert.True(t, store.IsErrNotFound(err))
}

func TestDryRun(t *testing.T) {
	s, err := store.MockOpen("mocktikv://")
	assert.NoError(t, err)
	defer func() {
		_ = s.Close()
	}()

	opt := &option{
		db:         1,
		namespace:  "test",
		batch:      100,
		dry:        true, // Dry run
		toleration: 0,
	}

	database := &db.DB{Namespace: opt.namespace, ID: db.DBID(opt.db)}

	txn, err := s.Begin()
	assert.NoError(t, err)

	now := db.Now()

	// Key1: Expired String
	key1 := []byte("key1")
	mkey1 := db.MetaKey(database, key1)
	obj1 := &db.Object{
		ID:        db.UUID(),
		CreatedAt: now,
		UpdatedAt: now,
		ExpireAt:  now - int64(time.Minute),
		Type:      db.ObjectString,
		Encoding:  db.ObjectEncodingRaw,
	}
	val1 := db.EncodeObject(obj1)
	val1 = append(val1, []byte("value1")...)
	err = txn.Set(mkey1, val1)
	assert.NoError(t, err)

	err = txn.Commit(context.Background())
	assert.NoError(t, err)

	// Run expire
	err = runExpire(s, opt)
	assert.NoError(t, err)

	// Verify - Should still exist because dry=true
	txn, err = s.Begin()
	assert.NoError(t, err)
	defer func() {
		_ = txn.Rollback()
	}()

	_, err = txn.Get(context.Background(), mkey1)
	assert.NoError(t, err)
}