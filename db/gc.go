package db

import (
	"context"
	"time"

	"github.com/meitu/titan/conf"
	"github.com/meitu/titan/db/store"
	"github.com/meitu/titan/metrics"
	"github.com/pingcap/tidb/kv"
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
	if logEnv := zap.L().Check(zap.DebugLevel, "[GC] add to gc"); logEnv != nil {
		logEnv.Write(zap.ByteString("prefix", prefix))
	}
	metrics.GetMetrics().GCKeysCounterVec.WithLabelValues("gc_add").Inc()
	return txn.Set(toTikvGCKey(prefix), []byte{0})
}

func gcDeleteRange(txn store.Transaction, prefix []byte, limit int) (int, error) {
	var (
		resultErr error
		count     int
	)
	endPrefix := kv.Key(prefix).PrefixNext()
	itr, err := txn.Iter(prefix, endPrefix)
	if err != nil {
		return count, err
	}
	defer itr.Close()
	callback := func(k kv.Key) bool {
		if resultErr = txn.Delete(itr.Key()); resultErr != nil {
			return true
		}
		count++
		if limit > 0 && count >= limit {
			return true
		}
		return false
	}
	if err := kv.NextUntil(itr, callback); err != nil {
		return 0, err
	}
	if resultErr != nil {
		return 0, resultErr
	}
	return count, nil
}

func doGC(db *DB, limit int) error {
	gcPrefix := toTikvGCKey(nil)
	endGCPrefix := kv.Key(gcPrefix).PrefixNext()
	dbTxn, err := db.Begin()
	if err != nil {
		zap.L().Error("[GC] transection begin failed",
			zap.ByteString("gcprefix", gcPrefix),
			zap.Int("limit", limit),
			zap.Error(err))
		return err
	}
	txn := dbTxn.t
	store.SetOption(txn, store.KeyOnly, true)

	itr, err := txn.Iter(gcPrefix, endGCPrefix)
	if err != nil {
		return err
	}
	defer itr.Close()
	if !itr.Valid() || !itr.Key().HasPrefix(gcPrefix) {
		if logEnv := zap.L().Check(zap.DebugLevel, "[GC] not need to gc item"); logEnv != nil {
			logEnv.Write(zap.ByteString("gcprefix", gcPrefix), zap.Int("limit", limit))
		}
		return nil
	}
	gcKeyCount := 0
	dataKeyCount := 0
	var resultErr error
	callback := func(k kv.Key) bool {
		dataPrefix := k[len(gcPrefix):]
		count := 0
		if logEnv := zap.L().Check(zap.DebugLevel, "[GC] start to delete prefix"); logEnv != nil {
			logEnv.Write(zap.ByteString("data-prefix", dataPrefix), zap.Int("limit", limit))
		}
		if count, resultErr = gcDeleteRange(txn, dataPrefix, limit); resultErr != nil {
			return true
		}

		//check and delete gc key
		if limit > 0 && count < limit || limit <= 0 && count > 0 {
			if logEnv := zap.L().Check(zap.DebugLevel, "[GC] delete prefix succeed"); logEnv != nil {
				logEnv.Write(zap.ByteString("data-prefix", dataPrefix), zap.Int("limit", limit))
			}

			if resultErr = txn.Delete(k); resultErr != nil {
				return true
			}
			gcKeyCount++
		}

		dataKeyCount += count
		if limit-(gcKeyCount+dataKeyCount) <= 0 {
			return true
		}
		return false
	}
	if err := kv.NextUntil(itr, callback); err != nil {
		zap.L().Error("[GC] iter prefix err", zap.ByteString("gc-prefix", gcPrefix), zap.Error(err))
		return err
	}
	if resultErr != nil {
		txn.Rollback()
		return resultErr
	}
	if err := txn.Commit(context.Background()); err != nil {
		txn.Rollback()
		return err
	}
	if logEnv := zap.L().Check(zap.DebugLevel, "[GC]  txn commit success"); logEnv != nil {
		logEnv.Write(zap.Int("limit", limit),
			zap.Int("gcKeyCount", gcKeyCount),
			zap.Int("dataKeyCount", dataKeyCount))
	}
	metrics.GetMetrics().GCKeysCounterVec.WithLabelValues("data_delete").Add(float64(dataKeyCount))
	metrics.GetMetrics().GCKeysCounterVec.WithLabelValues("gc_delete").Add(float64(gcKeyCount))
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
		if !conf.Enable {
			continue
		}
		isLeader, err := isLeader(db, sysGCLeader, id, conf.LeaderLifeTime)
		if err != nil {
			zap.L().Error("[GC] check GC leader failed",
				zap.ByteString("leader", sysGCLeader),
				zap.ByteString("uuid", id),
				zap.Duration("leader-life-time", conf.LeaderLifeTime),
				zap.Error(err))
			continue
		}
		if !isLeader {
			if logEnv := zap.L().Check(zap.DebugLevel, "[GC]  current is not gc leader"); logEnv != nil {
				logEnv.Write(zap.ByteString("leader", sysGCLeader),
					zap.ByteString("uuid", id),
					zap.Duration("leader-life-time", conf.LeaderLifeTime))
			}
			continue
		}
		if err := doGC(db, conf.BatchLimit); err != nil {
			zap.L().Error("[GC] do GC failed",
				zap.ByteString("leader", sysGCLeader),
				zap.ByteString("uuid", id),
				zap.Duration("leader-life-time", conf.LeaderLifeTime),
				zap.Error(err))
			continue
		}
	}
}
