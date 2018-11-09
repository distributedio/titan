package db

import (
	"bytes"
	"log"
	"math/rand"

	"gitlab.meitu.com/platform/thanos/db/store"
)

// Kv supplies key releated operations
type Kv struct {
	txn *Transaction
}

// GetKv returns a Kv object
func GetKv(txn *Transaction) *Kv {
	return &Kv{txn}
}

// Keys iterator all keys in db
func (kv *Kv) Keys(start []byte, f func(key []byte) bool) error {
	mkey := MetaKey(kv.txn.db, start)
	prefix := MetaKey(kv.txn.db, nil)
	iter, err := kv.txn.t.Seek(mkey)
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
		if !IsExpired(obj, now) && !f(key[len(prefix):]) {
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
	var count int64
	now := Now()
	metaKeys := make([][]byte, len(keys))
	for i, key := range keys {
		metaKeys[i] = MetaKey(kv.txn.db, key)
	}

	values, err := BatchGetValues(kv.txn, metaKeys)
	if err != nil {
		return count, err
	}
	for i, val := range values {
		if val != nil {
			obj, err := DecodeObject(val)
			if err != nil {
				return count, err
			}
			if IsExpired(obj, now) {
				continue
			}
			if err := kv.txn.Destory(obj, keys[i]); err != nil {
				continue
			}
			if obj.ExpireAt > now {
				if err := unExpireAt(kv.txn, keys[i], obj.ExpireAt); err != nil {
					return count, err
				}
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

	meta, err := kv.txn.t.Get(mkey)
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
	if err := expireAt(kv.txn, key, obj.ID, obj.ExpireAt, at); err != nil {
		return err
	}
	obj.ExpireAt = at
	updated := EncodeObject(obj)
	updated = append(updated, meta[ObjectEncodingLength:]...)
	return kv.txn.t.Set(mkey, updated)
}

//Exists check if the given keys exist
func (kv *Kv) Exists(keys [][]byte) (int64, error) {
	var count int64
	now := Now()
	mkeys := make([][]byte, len(keys))
	for i, key := range keys {
		mkeys[i] = MetaKey(kv.txn.db, key)
	}

	values, err := store.BatchGetValues(kv.txn.t, mkeys)
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

// FlushDB clear current db. FIXME one txn is limited for number of entries
func (kv *Kv) FlushDB() error {
	prefix := DBPrefix(kv.txn.db)
	txn := kv.txn.t

	iter, err := txn.Seek(prefix)
	if err != nil {
		return err
	}
	for iter.Valid() && iter.Key().HasPrefix(prefix) {
		if err := txn.Delete(iter.Key()); err != nil {
			return err
		}
		if err := iter.Next(); err != nil {
			return err
		}
	}
	return nil
}

// FlushAll clean up all databases. FIXME one txn is limited for number of entries
func (kv *Kv) FlushAll() error {
	prefix := []byte(kv.txn.db.Namespace + ":")
	txn := kv.txn.t

	iter, err := txn.Seek(prefix)
	if err != nil {
		return err
	}
	for iter.Valid() && iter.Key().HasPrefix(prefix) {
		log.Println(string(iter.Key()))
		if err := txn.Delete(iter.Key()); err != nil {
			return err
		}
		if err := iter.Next(); err != nil {
			return err
		}
	}
	return nil

}

// RandomKey return a key from current db randomly
// Now we use an static length(64) to generate the key spaces, it means it is random for keys
// that len(key) <= 64, it is enough for most cases
func (kv *Kv) RandomKey() ([]byte, error) {
	buf := make([]byte, 64)
	// Read for rand here always return a nil error
	rand.Read(buf)

	mkey := MetaKey(kv.txn.db, buf)
	prefix := MetaKey(kv.txn.db, nil)

	// Seek >= mkey
	iter, err := kv.txn.t.Seek(mkey)
	if err != nil {
		return nil, err
	}

	if iter.Valid() && iter.Key().HasPrefix(prefix) {
		return iter.Key()[len(prefix):], nil
	}
	first := make([]byte, len(prefix)+1)
	copy(first, prefix)
	iter, err = kv.txn.t.Seek(first)
	if err != nil {
		return nil, err
	}

	if iter.Valid() && iter.Key().HasPrefix(prefix) {
		return iter.Key()[len(prefix):], nil
	}
	return nil, err
}
