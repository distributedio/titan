package db

import (
	"context"
	"time"

	"github.com/distributedio/titan/conf"
	"github.com/distributedio/titan/db/store"
	"github.com/distributedio/titan/metrics"
	"github.com/pingcap/tidb/kv"
	"go.uber.org/zap"
)

var (
	sysGCLeader = []byte("$sys:0:GCL:GCLeader")
)

func toTiKVGCKey(key []byte) []byte {
	b := []byte{}
	b = append(b, sysNamespace...)
	b = append(b, ':', byte(sysDatabaseID))
	b = append(b, ':', 'G', 'C', ':')
	b = append(b, key...)
	return b
}

// {sys.ns}:{sys.id}:{GC}:{prefix}
// prefix: {user.ns}:{user.id}:{M/D}:{user.objectID}
func gc(txn store.Transaction, prefixes ...[]byte) error {
	for _, prefix := range prefixes {
		if logEnv := zap.L().Check(zap.DebugLevel, "[GC] add to gc"); logEnv != nil {
			logEnv.Write(zap.ByteString("prefix", prefix))
		}
		metrics.GetMetrics().GCKeysCounterVec.WithLabelValues("gc_add").Inc()
		if err := txn.Set(toTiKVGCKey(prefix), []byte{0}); err != nil {
			return err
		}
	}
	return nil
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
	gcPrefix := toTiKVGCKey(nil)
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
	store.SetOption(txn, store.Priority, store.PriorityLow)

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
		return limit-(gcKeyCount+dataKeyCount) <= 0
	}
	if err := kv.NextUntil(itr, callback); err != nil {
		zap.L().Error("[GC] iter prefix err", zap.ByteString("gc-prefix", gcPrefix), zap.Error(err))
		return err
	}
	if resultErr != nil {
		if err := txn.Rollback(); err != nil {
			zap.L().Error("[GC] rollback err", zap.Error(err))
		}
		return resultErr
	}
	if err := txn.Commit(context.Background()); err != nil {
		if err := txn.Rollback(); err != nil {
			zap.L().Error("[GC] rollback err", zap.Error(err))
		}
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
func StartGC(task *Task) {
	conf := task.conf.(conf.GC)
	ticker := time.NewTicker(conf.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-task.Done():
			if logEnv := zap.L().Check(zap.DebugLevel, "[GC] current is not gc leader"); logEnv != nil {
				logEnv.Write(zap.ByteString("key", task.key),
					zap.ByteString("uuid", task.id),
					zap.String("lable", task.lable))
			}
			return
		case <-ticker.C:
		}

		if err := doGC(task.db, conf.BatchLimit); err != nil {
			zap.L().Error("[GC] do GC failed",
				zap.ByteString("leader", task.key),
				zap.ByteString("uuid", task.id),
				zap.Int("leader-ttl", conf.LeaderTTL),
				zap.Error(err))
			continue
		}
	}
}
