package db

import (
	"bytes"
	"encoding/json"
	"log"
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
func (kv *Kv) Keys(onkey func(key []byte) bool) error {
	mkey := MetaKey(kv.txn.db, nil)
	iter, err := kv.txn.txn.Seek(mkey)
	if err != nil {
		return err
	}
	for iter.Valid() {
		key := iter.Key()
		if !bytes.HasPrefix(key, mkey) {
			break
		}
		if !onkey(key[len(mkey):]) {
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
	n := 0
	for _, key := range keys {
		mkey := MetaKey(kv.txn.db, key)

		obj, err := kv.txn.Object(key)
		if err != nil {
			if err == ErrKeyNotFound {
				continue
			}
			return -1, err
		}

		switch obj.Type {
		case ObjectString:
			if err := kv.txn.txn.Delete(mkey); err != nil {
				return -1, err
			}
		case ObjectList, ObjectHash:
			if err := kv.txn.Destory(obj, key); err != nil {
				return -1, err
			}
		}
		n++
	}
	return n, nil
}

// ExpireAt set a timeout on key
func (kv *Kv) ExpireAt(key []byte, at int64) error {
	mkey := MetaKey(kv.txn.db, key)
	meta, err := kv.txn.txn.Get(mkey)
	if err != nil {
		if IsErrNotFound(err) {
			return ErrKeyNotFound
		}
		return err
	}
	obj := &Object{}
	err = json.Unmarshal(meta, obj)
	if err != nil {
		return err
	}
	var updated []byte
	switch obj.Type {
	case ObjectList:
		listMeta := &ListMeta{}
		if err := json.Unmarshal(meta, listMeta); err != nil {
			return err
		}
		listMeta.ExpireAt = at
		updated, err = json.Marshal(listMeta)
		if err != nil {
			return err
		}
	case ObjectString:
		strMeta := &StringMeta{}
		if err := json.Unmarshal(meta, strMeta); err != nil {
			return err
		}
		strMeta.ExpireAt = at
		updated, err = json.Marshal(strMeta)
		if err != nil {
			return err
		}
	case ObjectSet:
	case ObjectZset:
	case ObjectHash:
	}

	if err := expireAt(kv.txn, key, obj.ID, obj.ExpireAt, at); err != nil {
		return err
	}

	return kv.txn.txn.Set(mkey, updated)
}

// FlushDB clear current db. FIXME one txn is limited for number of entries
func (kv *Kv) FlushDB() error {
	prefix := DBPrefix(kv.txn.db)
	txn := kv.txn.txn

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
	txn := kv.txn.txn

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
