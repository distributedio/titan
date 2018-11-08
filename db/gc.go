package db

import (
	"bytes"
	"context"
	"encoding/binary"
	"time"

	"gitlab.meitu.com/platform/thanos/db/store"
	"go.uber.org/zap"
)

var (
	sysNamespace  = []byte("$sys")
	sysDatabaseID = '0'
	sysGCLeader   = []byte("$sys:0:GCL:GCLeader")

	//TODO  使用配置
	gcTick = time.Duration(time.Second)
)

const (
	sysGCBurst              = 256
	sysGCLeaseFlushInterval = 10
)

func toTikvGCKey(key []byte) []byte {
	b := []byte{}
	b = append(b, sysNamespace...)
	b = append(b, ':', byte(sysDatabaseID))
	b = append(b, ':', 'G', 'C', ':')
	b = append(b, key...)
	return b
}

// {sys.ns}:{sys.id}:{GC}:{prefix}
// prefix: {user.ns}:{user.id}:{M/D}:{user.objectID}
func gc(txn store.Transaction, prefix []byte) error {
	zap.L().Debug("[GC] remove prefix", zap.String("prefix", string(prefix)))
	key := toTikvGCKey(prefix)
	return txn.Set(key, []byte{0})
}

func _doGC(txn store.Transaction, limit int64) (int64, error) {
	prefix, err := gcGetPrefix(txn)
	if err != nil {
		return 0, err
	}
	if prefix == nil {
		zap.L().Debug("[GC] no gc item")
		return 0, nil
	}

	zap.L().Debug("[GC] start to delete prefix", zap.String("prefix", string(prefix)), zap.Int64("limit", limit))
	count, err := gcDeleteRange(txn, prefix, limit)
	if err != nil {
		return 0, err
	}

	if count < limit {
		zap.L().Debug("[GC] delete prefix succeed", zap.String("prefix", string(prefix)))
		if err := gcComplete(txn, prefix); err != nil {
			return 0, err
		}
	}

	return count, nil
}

func gcGetPrefix(txn store.Transaction) ([]byte, error) {
	gcPrefix := []byte{}
	gcPrefix = append(gcPrefix, sysNamespace...)
	gcPrefix = append(gcPrefix, ':', byte(sysDatabaseID))
	gcPrefix = append(gcPrefix, ':', 'G', 'C', ':')
	itr, err := txn.Seek(gcPrefix)
	if err != nil {
		return nil, err
	}
	defer itr.Close()

	if !itr.Valid() {
		return nil, nil
	}

	key := itr.Key()

	if !key.HasPrefix(gcPrefix) {
		return nil, nil
	}

	return key[len(gcPrefix):], nil
}

func gcDeleteRange(txn store.Transaction, prefix []byte, limit int64) (int64, error) {
	count := int64(0)
	itr, err := txn.Seek(prefix)
	if err != nil {
		return count, err
	}
	defer itr.Close()

	for itr.Valid() {
		key := itr.Key()
		if !key.HasPrefix(prefix) {
			return count, nil
		}

		if err := txn.Delete(key); err != nil {
			return count, err
		}

		if limit > 0 {
			count++
			if count >= limit {
				return count, nil
			}
		}

		if err := itr.Next(); err != nil {
			return count, err
		}
	}

	return count, nil
}

func gcComplete(txn store.Transaction, prefix []byte) error {
	return txn.Delete(toTikvGCKey(prefix))
}

func flushLease(txn store.Transaction, key, id []byte, interval time.Duration) error {
	databytes := make([]byte, 24)
	copy(databytes, id)
	ts := uint64((time.Now().Add(interval * time.Second).Unix()))
	binary.BigEndian.PutUint64(databytes[16:], ts)

	if err := txn.Set(key, databytes); err != nil {
		return err
	}
	return nil
}

func checkLeader(txn store.Transaction, key, id []byte, interval time.Duration) (bool, error) {
	val, err := txn.Get(key)
	if err != nil {
		if !IsErrNotFound(err) {
			zap.L().Error("query leader message faild", zap.Error(err))
			return false, err
		}

		zap.L().Debug("no leader now, create new lease")
		if err := flushLease(txn, key, id, interval); err != nil {
			zap.L().Error("create lease failed", zap.Error(err))
			return false, err
		}

		return true, nil
	}

	curID := val[0:16]
	ts := int64(binary.BigEndian.Uint64(val[16:]))

	if time.Now().Unix() > ts {
		zap.L().Error("lease expire, create new lease")
		if err := flushLease(txn, key, id, interval); err != nil {
			zap.L().Error("create lease failed", zap.Error(err))
			return false, err
		}
		return true, nil
	}

	if bytes.Equal(curID, id) {
		if err := flushLease(txn, key, id, interval); err != nil {
			zap.L().Error("flush lease failed", zap.Error(err))
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func isLeader(s *RedisStore, leader []byte, interval time.Duration) (bool, error) {
	count := 0
	for {
		txn, err := s.Begin()
		if err != nil {
			return false, err
		}

		isLeader, err := checkLeader(txn, leader, gInstanceID.Bytes(), interval)
		if err != nil {
			txn.Rollback()
			if IsErrRetriable(err) {
				count++
				if count < 3 {
					continue
				}
			}
			return isLeader, err
		}

		if err := txn.Commit(context.Background()); err != nil {
			txn.Rollback()
			if IsErrRetriable(err) {
				count++
				if count < 3 {
					continue
				}
			}
			return isLeader, err
		}

		//TODO add monitor
		return isLeader, err
	}
}

func doGC(s *RedisStore, limit int64) error {
	left := limit
	for left > 0 {
		txn, err := s.Begin()
		if err != nil {
			return err
		}

		count, err := _doGC(txn, left)
		if err != nil {
			txn.Rollback()
			return err
		}

		if count == 0 {
			txn.Rollback()
			return nil
		}
		left -= count

		if err := txn.Commit(context.Background()); err != nil {
			txn.Rollback()
			return err
		}
	}
	return nil
}

func StartGC(s *RedisStore) {
	ticker := time.NewTicker(gcTick)
	for _ = range ticker.C {
		isLeader, err := isLeader(s, sysGCLeader, sysGCLeaseFlushInterval)
		if err != nil {
			zap.L().Error("[GC] check GC leader failed", zap.Error(err))
			continue
		}

		if !isLeader {
			zap.L().Debug("[GC] not GC leader")
			continue
		}

		if err := doGC(s, sysGCBurst); err != nil {
			zap.L().Error("[GC] do GC failed", zap.Error(err))
			continue
		}
	}
}
