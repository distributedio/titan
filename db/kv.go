package db

import (
	"bytes"
	"log"

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
func (kv *Kv) Keys(start []byte, filter func(key []byte) bool) error {
	mkey := MetaKey(kv.txn.db, start)
	prefixKey := MetaKey(kv.txn.db, nil)
	iter, err := kv.txn.t.Seek(mkey)
	if err != nil {
		return err
	}
	defer iter.Close()

	now := Now()
	for iter.Valid() {
		key := iter.Key()
		if !bytes.HasPrefix(key, prefixKey) {
			break
		}

		obj, err := DecodeObject(iter.Value())
		if err != nil {
			return err
		}
		if !IsExpired(obj, now) && !filter(key[len(prefixKey):]) {
			break
		}
		if err := iter.Next(); err != nil {
			return err
		}
	}
	return nil
}

// Delete specific keys, ignore if non exist
func (kv *Kv) Delete(keys [][]byte) (int, error) {
	count := 0
	now := Now()
	metaKeys := make([][]byte, len(keys))
	for i, key := range keys {
		metaKeys[i] = MetaKey(kv.txn.db, key)
	}

	dataBytes, err := store.BatchGetValues(kv.txn.t, metaKeys)
	if err != nil {
		return count, err
	}
	for i, data := range dataBytes {
		if data != nil {
			obj, err := DecodeObject(data)
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

//Exists
func (kv *Kv) Exists(keys [][]byte) (int64, error) {
	var (
		now   int64    = Now()
		count int64    = 0
		mkeys [][]byte = make([][]byte, len(keys))
	)
	for i, key := range keys {
		mkeys[i] = MetaKey(kv.txn.db, key)
	}

	dataBytes, err := store.BatchGetValues(kv.txn.t, mkeys)
	if err != nil {
		return count, err
	}
	for _, data := range dataBytes {
		if data != nil {
			obj, err := DecodeObject(data)
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
