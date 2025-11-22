package db

import (
	"bytes"
	"crypto/rand"
	"math/big"

	"github.com/distributedio/titan/db/store"
)

// Kv supplies key releated operations
type Kv struct {
	txn *Transaction
}

// GetKv returns a Kv object
func GetKv(txn *Transaction) *Kv {
	return &Kv{txn}
}

// Helper to replace sdk_kv.Key(prefix).PrefixNext()
func prefixNext(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}
	next := make([]byte, len(prefix))
	copy(next, prefix)
	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] != 0 {
			return next
		}
	}
	return nil
}

// Keys iterator all keys in db
func (kv *Kv) Keys(start []byte, f func(key []byte, obj *Object) bool) error {
	mkey := MetaKey(kv.txn.db, start)
	prefix := MetaKey(kv.txn.db, nil)
	endPrefix := prefixNext(prefix)
	iter, err := kv.txn.t.Iter(mkey, endPrefix)
	if err != nil {
		return err
	}
	defer iter.Close()

	now := Now()
	for iter.Valid() {
		key := iter.Key()
		if !bytes.HasPrefix(key, prefix) {
			break
		}

		obj, err := DecodeObject(iter.Value())
		if err != nil {
			return err
		}
		if !IsExpired(obj, now) && !f(key[len(prefix):], obj) {
			break
		}
		if err := iter.Next(); err != nil {
			return err
		}
	}
	return nil
}

// Delete specific keys, ignore if non exist
func (kv *Kv) Delete(keys [][]byte) (int64, error) {
	var (
		count    int64
		metaKeys [][]byte
		mapping  = make(map[string][]byte)
		now      = Now()
	)
	// use mapping to filter duplicate keys
	for _, key := range keys {
		mkey := MetaKey(kv.txn.db, key)
		if _, ok := mapping[string(mkey)]; !ok {
			mapping[string(mkey)] = key
			metaKeys = append(metaKeys, mkey)
		}
	}

	values, err := store.BatchGetValues(kv.txn.ctx, kv.txn.t, metaKeys)
	if err != nil {
		return count, err
	}
	for k, val := range values {
		if val != nil {
			obj, err := DecodeObject(val)
			if err != nil {
				return count, err
			}
			if IsExpired(obj, now) {
				continue
			}
			if err := kv.txn.Destory(obj, mapping[k]); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}

// ExpireAt set a timeout on key
func (kv *Kv) ExpireAt(key []byte, at int64) error {
	mkey := MetaKey(kv.txn.db, key)
	now := Now()

	meta, err := kv.txn.t.Get(kv.txn.ctx, mkey)
	if err != nil {
		if IsErrNotFound(err) {
			return ErrKeyNotFound
		}
		return err
	}
	obj, err := DecodeObject(meta)
	if err != nil {
		return err
	}
	if IsExpired(obj, now) {
		return ErrKeyNotFound
	}
	if at == 0 && obj.ExpireAt != 0 {
		if err = unExpireAt(kv.txn.t, mkey, obj.ExpireAt); err != nil {
			return err
		}
	}

	if at > 0 {
		if err := expireAt(kv.txn.t, mkey, obj.ID, obj.Type, obj.ExpireAt, at); err != nil {
			return err
		}
	}
	obj.ExpireAt = at
	updated := EncodeObject(obj)
	updated = append(updated, meta[ObjectEncodingLength:]...)
	return kv.txn.t.Set(mkey, updated)
}

// Exists check if the given keys exist
func (kv *Kv) Exists(keys [][]byte) (int64, error) {
	var count int64
	now := Now()
	mkeys := make([][]byte, len(keys))
	for i, key := range keys {
		mkeys[i] = MetaKey(kv.txn.db, key)
	}

	values, err := store.BatchGetValues(kv.txn.ctx, kv.txn.t, mkeys)
	if err != nil {
		return count, err
	}
	for _, val := range values {
		if val != nil {
			obj, err := DecodeObject(val)
			if err != nil {
				return count, err
			}
			if IsExpired(obj, now) {
				continue
			}
			count++
		}
	}
	return count, nil
}

// RandomKey return a key from current db randomly
// Now we use an static length(64) to generate the key spaces, it means it is random for keys
// that len(key) <= 64, it is enough for most cases
func (kv *Kv) RandomKey() ([]byte, error) {
	buf := make([]byte, 64)
	// Read for rand here always return a nil error
	_, _ = rand.Read(buf)

	mkey := MetaKey(kv.txn.db, buf)
	prefix := MetaKey(kv.txn.db, nil)

	randKeys := func(iter Iterator) []byte {
		keys := make([][]byte, 0)
		for iter.Valid() && bytes.HasPrefix(iter.Key(), prefix) {
			keys = append(keys, iter.Key()[len(prefix):])
			if err := iter.Next(); err != nil {
				return nil
			}
		}
		count := len(keys)
		if count == 0 {
			return nil
		}
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(count)))
		return keys[idx.Int64()]
	}

	iter, err := kv.txn.t.Iter(mkey, nil)
	if err != nil {
		return nil, err
	}

	if key := randKeys(iter); key != nil {
		return key, nil
	}

	iter, err = kv.txn.t.IterReverse(mkey)
	if err != nil {
		return nil, err
	}
	if key := randKeys(iter); key != nil {
		return key, nil
	}
	return nil, nil
}

// Touch alters the last access time of a key(s)
func (kv *Kv) Touch(keys [][]byte) (int64, error) {
	ts := Now()
	count := int64(0)

	mkeys := make([][]byte, len(keys))
	for i, key := range keys {
		mkeys[i] = MetaKey(kv.txn.db, key)
	}

	values, err := store.BatchGetValues(kv.txn.ctx, kv.txn.t, mkeys)
	if err != nil {
		return 0, err
	}

	for mkey, meta := range values {
		if meta == nil {
			continue
		}
		obj, err := DecodeObject(meta)
		if err != nil {
			return 0, err
		}
		if IsExpired(obj, ts) {
			continue
		}

		obj.UpdatedAt = ts
		count++
		updated := append(EncodeObject(obj), meta[ObjectEncodingLength:]...)
		if err := kv.txn.t.Set([]byte(mkey), updated); err != nil {
			return 0, err
		}
	}
	return count, nil
}