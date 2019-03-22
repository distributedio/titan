package db

import (
	"context"
	"time"

	"github.com/meitu/titan/db/store"
	"github.com/meitu/titan/metrics"
	"go.uber.org/zap"
)

var (
	sysGCLeader = []byte("$sys:0:GCL:GCLeader")
	gcInterval  = time.Duration(1) //TODO 使用配置
)

const (
	sysGCBurst              = 257
	sysGCSeekNum            = 10
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
// prefix: {user.ns}:{user.id}:{M/D/S}:{user.objectID}
func gc(txn store.Transaction, prefixs [][]byte) error {
	var err error
	for _, prefix := range prefixs{
		zap.L().Debug("add to gc", zap.ByteString("prefix", prefix))
		metrics.GetMetrics().GCKeysCounterVec.WithLabelValues("add").Inc()
		if err = txn.Set(toTikvGCKey(prefix), []byte{0}); err != nil {
			return err
		}
	}
	return nil
}

func gcGetPrefix(txn store.Transaction) ([][]byte, error) {
	gcPrefix := []byte{}
	gcPrefix = append(gcPrefix, sysNamespace...)
	gcPrefix = append(gcPrefix, ':', byte(sysDatabaseID))
	gcPrefix = append(gcPrefix, ':', 'G', 'C', ':')
	itr, err := txn.Seek(gcPrefix)
	if err != nil {
		return nil, err
	}
	defer itr.Close()

	keys := make([][]byte, 0, sysGCSeekNum)
	for i := int64(0); i <= sysGCSeekNum && itr.Valid() && itr.Key().HasPrefix(gcPrefix); i++{
		keys = append(keys, itr.Key()[len(gcPrefix):])

		if err = itr.Next(); err != nil {
			break
		}
	}
	return keys, nil
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

		prefixs, err := gcGetPrefix(txn.t)
		if err != nil {
			return err
		}
		if len(prefixs) == 0 {
			zap.L().Debug("[GC] no gc item")
			return nil
		}
		count := int64(0)
		onceCount := int64(0)
		zap.L().Debug("[GC] doGC loop once")
		for i,num :=0,0 ; num<len(prefixs); num++ {
			prefix := prefixs[i]

			zap.L().Debug("[GC] start to delete prefix", zap.String("prefix", string(prefix)), zap.Int64("limit", limit))
			if onceCount, err = gcDeleteRange(txn.t, prefix, limit); err != nil {
				return err
			}
			count += onceCount

			if onceCount < limit {
				zap.L().Debug("[GC] delete prefix succeed", zap.String("prefix", string(prefix)), zap.Int64("deleted", onceCount))
				if err := gcComplete(txn.t, prefix); err != nil {
					txn.Rollback()
					return err
				}
				i++
			} else {
				if onceCount == 0 {
					txn.Rollback()
					return nil
				}
				//still need deleteRange for this prefix
				zap.L().Debug("[GC] part of delete prefix", zap.String("prefix", string(prefix)), zap.Int64("deleted", onceCount))
			}
		}
		//either < or >= limit, the "left" should be decreased count,
		// else current thread will busy in gc and don't update lease, another thread will also do gc
		left -= count

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
func StartGC(db *DB) {//empty
	ticker := time.Tick(gcInterval * time.Second)
	id := UUID()
	for range ticker {
		isLeader, err := isLeader(db, sysGCLeader, id, sysGCLeaseFlushInterval)
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
