package db

import (
	"context"
	"time"

	"gitlab.meitu.com/platform/thanos/db/store"
	"gitlab.meitu.com/platform/thanos/metrics"
	"go.uber.org/zap"
)

var (
	sysGCLeader = []byte("$sys:0:GCL:GCLeader")
	gcInterval  = time.Duration(1) //TODO 使用配置
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
	zap.L().Debug("add to gc", zap.ByteString("prefix", prefix))
	metrics.GetMetrics().RecycleInfoGaugeVec.WithLabelValues("gc_add_key").Inc()
	return txn.Set(toTikvGCKey(prefix), []byte{0})
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

func doGC(db *DB, limit int64) error {
	left := limit
	for left > 0 {
		txn, err := db.Begin()
		if err != nil {
			zap.L().Error("[GC] transection begin failed", zap.Error(err))
			return err
		}

		prefix, err := gcGetPrefix(txn.t)
		if err != nil {
			return err
		}
		if prefix == nil {
			zap.L().Debug("[GC] no gc item")
			return nil
		}
		count := int64(0)
		zap.L().Debug("[GC] start to delete prefix", zap.String("prefix", string(prefix)), zap.Int64("limit", limit))
		if count, err = gcDeleteRange(txn.t, prefix, limit); err != nil {
			return err
		}

		if count < limit {
			zap.L().Debug("[GC] delete prefix succeed", zap.String("prefix", string(prefix)))
			if err := gcComplete(txn.t, prefix); err != nil {
				txn.Rollback()
				return err
			}
		} else {
			if count == 0 {
				txn.Rollback()
				return nil
			}
			left -= count
		}

		if err := txn.Commit(context.Background()); err != nil {
			txn.Rollback()
			return err
		}
		metrics.GetMetrics().RecycleInfoGaugeVec.WithLabelValues("gc_delete_key").Add(float64(count))
	}
	return nil
}

func StartGC(db *DB) {
	ticker := time.Tick(gcInterval * time.Second)
	for _ = range ticker {
		isLeader, err := isLeader(db, sysGCLeader, sysGCLeaseFlushInterval)
		if err != nil {
			zap.L().Error("[GC] check GC leader failed", zap.Error(err))
			continue
		}
		if !isLeader {
			zap.L().Debug("[GC] not GC leader")
			continue
		}
		if err := doGC(db, sysGCBurst); err != nil {
			zap.L().Error("[GC] do GC failed", zap.Error(err))
			continue
		}
	}
}
