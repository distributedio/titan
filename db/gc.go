package db

import (
	"context"
	"time"

	"github.com/meitu/titan/conf"
	"github.com/meitu/titan/db/store"
	"github.com/meitu/titan/metrics"
	"go.uber.org/zap"
)

var (
	sysGCLeader = []byte("$sys:0:GCL:GCLeader")
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
	metrics.GetMetrics().GCKeysCounterVec.WithLabelValues("add").Inc()
	return txn.Set(toTikvGCKey(prefix), []byte{0})
}

func gcGetPrefix(txn store.Transaction) ([]byte, error) {
	gcPrefix := []byte{}
	gcPrefix = append(gcPrefix, sysNamespace...)
	gcPrefix = append(gcPrefix, ':', byte(sysDatabaseID))
	gcPrefix = append(gcPrefix, ':', 'G', 'C', ':')
	itr, err := txn.Iter(gcPrefix, nil)
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

func gcDeleteRange(txn store.Transaction, prefix []byte, limit int) (int, error) {
	var count int
	itr, err := txn.Iter(prefix, nil)
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

func doGC(db *DB, limit int) error {
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
		count := 0
		zap.L().Debug("[GC] start to delete prefix", zap.String("prefix", string(prefix)), zap.Int("limit", limit))
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
		metrics.GetMetrics().GCKeysCounterVec.WithLabelValues("delete").Add(float64(count))
	}
	return nil
}

// StartGC start gc
//1.获取leader许可
//2.leader 执行清理任务
func StartGC(db *DB, conf *conf.GC) {
	ticker := time.NewTicker(conf.Interval)
	defer ticker.Stop()
	id := UUID()
	for range ticker.C {
		isLeader, err := isLeader(db, sysGCLeader, id, conf.LeaderLifeTime)
		if err != nil {
			zap.L().Error("[GC] check GC leader failed", zap.Error(err))
			continue
		}
		if !isLeader {
			zap.L().Debug("[GC] not GC leader")
			continue
		}
		if err := doGC(db, conf.BatchLimit); err != nil {
			zap.L().Error("[GC] do GC failed", zap.Error(err))
			continue
		}
	}
}
