package db

import (
	"context"
	"log"
	"time"

	"gitlab.meitu.com/platform/thanos/db/store"
)

const gcTick = time.Duration(time.Second)
const gcBatchMaxRemove = 1000

var gcKeyPrefix = []byte("$sys:gc:")

func gcKey(key []byte) []byte {
	var buf []byte
	buf = append(buf, gcKeyPrefix...)
	buf = append(buf, key...)
	return buf
}

func gc(txn *Transaction, dataKey []byte) error {
	return txn.txn.Set(gcKey(dataKey), []byte{0})
}

func StartGC(s *RedisStore) error {
	ticker := time.NewTicker(gcTick)
	for {
		select {
		case <-ticker.C:
			runGC(s)
		}
	}
}

func runGC(s *RedisStore) {
	txn, err := s.Begin()
	if err != nil {
		log.Println(err)
		return
	}
	iter, err := txn.Seek(gcKeyPrefix)
	if err != nil {
		log.Println(err)
		return
	}
	limit := gcBatchMaxRemove
	for iter.Valid() && iter.Key().HasPrefix(gcKeyPrefix) && limit > 0 {
		dkey := iter.Key()[len(gcKeyPrefix):]

		count, err := removeWithPrefix(txn, dkey, limit)
		if err != nil {
			txn.Rollback()
			return
		}
		limit -= count
		// all elements with prefix dkey has been removed, remove the record from gc
		if limit > 0 {
			if err := txn.Delete(iter.Key()); err != nil {
				txn.Rollback()
				return
			}
		}

		if err := iter.Next(); err != nil {
			txn.Rollback()
			return
		}
	}
	if err := txn.Commit(context.Background()); err != nil {
		log.Println(err)
	}
}

// removeWithPrefix removes all keys with certain prefix and returns the count that being removed
func removeWithPrefix(txn store.Transaction, prefix []byte, limit int) (int, error) {
	log.Println("remove ", string(prefix))
	iter, err := txn.Seek(prefix)
	if err != nil {
		return 0, err
	}
	total := limit
	for iter.Valid() && iter.Key().HasPrefix(prefix) && limit != 0 {
		key := iter.Key()
		if err := txn.Delete(key); err != nil {
			return total - limit, err
		}
		if err := iter.Next(); err != nil {
			return total - limit, err
		}
		limit--
	}
	return total - limit, nil
}
