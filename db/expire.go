package db

import (
	"bytes"
	"context"
	"log"
	"time"
)

const expireBatchLimit = 100
const expireTick = time.Duration(time.Second)

var expireKeyPrefix = []byte("$sys:at:")

func expireKey(key []byte, ts int64) []byte {
	var buf []byte
	buf = append(buf, expireKeyPrefix...)
	buf = append(buf, EncodeInt64(ts)...)
	buf = append(buf, ':')
	buf = append(buf, key...)
	return buf
}

func expireAt(txn *Transaction, key []byte, objID []byte, old int64, new int64) error {
	mkey := MetaKey(txn.db, key)
	oldKey := expireKey(mkey, old)
	newKey := expireKey(mkey, new)

	if err := txn.txn.Delete(oldKey); err != nil {
		return err
	}

	if err := txn.txn.Set(newKey, objID); err != nil {
		return err
	}
	return nil
}

func StartExpire(s *RedisStore) error {
	ticker := time.NewTicker(expireTick)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			runExpire(s)
		}
	}
	return nil
}

func splitMetaKey(key []byte) ([]byte, []byte, []byte) {
	idx := bytes.Index(key, []byte{':'})
	namespace := key[0:idx]
	id := key[idx+1 : idx+4]
	rawkey := key[idx+7:]
	return namespace, id, rawkey
}

func runExpire(s *RedisStore) {
	txn, err := s.Begin()
	if err != nil {
		log.Println(err)
		return
	}
	iter, err := txn.Seek(expireKeyPrefix)
	if err != nil {
		log.Println(err)
		txn.Rollback()
	}
	limit := expireBatchLimit
	now := time.Now().UnixNano()
	for iter.Valid() && iter.Key().HasPrefix(expireKeyPrefix) && limit > 0 {
		key := iter.Key()
		objID := iter.Value()
		// prefix + sizeof(int64) + len(":")
		mkey := key[len(expireKeyPrefix)+9:]
		namespace, dbid, rawkey := splitMetaKey(mkey)

		ts := DecodeInt64(key[len(expireKeyPrefix) : len(expireKeyPrefix)+8])
		if ts > now {
			break
		}

		log.Println("expire ", string(rawkey))
		// Delete object meta
		if err := txn.Delete(mkey); err != nil {
			log.Println(err)
			txn.Rollback()
			return
		}
		// Gc it if it is a complext data structure
		if bytes.Compare(objID, rawkey) != 0 {
			dkey := DataKey(&DB{Namespace: string(namespace), ID: toDBID(dbid)}, objID)
			log.Println("add to gc ", string(dkey))
			if err := gc(&Transaction{txn: txn}, dkey); err != nil {
				log.Println(err)
				txn.Rollback()
				return
			}
		}
		// Remove from expire list
		if err := txn.Delete(iter.Key()); err != nil {
			log.Println(err)
			txn.Rollback()
			return
		}
		if err := iter.Next(); err != nil {
			log.Println(err)
			txn.Rollback()
			return
		}
		limit--
	}

	if err := txn.Commit(context.Background()); err != nil {
		txn.Rollback()
		log.Println(err)
	}
}
