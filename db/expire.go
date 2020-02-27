package db

import (
	"bytes"
	"context"
	"time"

	"github.com/distributedio/titan/conf"
	"github.com/distributedio/titan/db/store"
	"github.com/distributedio/titan/metrics"
	"go.uber.org/zap"
)

var (
	expireKeyPrefix = []byte("$sys:0:at:")
	sysExpireLeader = []byte("$sys:0:EXL:EXLeader")

	// $sys:0:at:{ts}:{metaKey}
	expireTimestampOffset = len(expireKeyPrefix)
	expireMetakeyOffset   = expireTimestampOffset + 8 /*sizeof(int64)*/ + len(":")
)

const (
	expire_worker = "expire"
)

// IsExpired judge object expire through now
func IsExpired(obj *Object, now int64) bool {
	if obj.ExpireAt == 0 || obj.ExpireAt > now {
		return false
	}
	return true
}

func expireKey(key []byte, ts int64) []byte {
	var buf []byte
	buf = append(buf, expireKeyPrefix...)
	buf = append(buf, EncodeInt64(ts)...)
	buf = append(buf, ':')
	buf = append(buf, key...)
	return buf
}

func expireAt(txn store.Transaction, mkey []byte, objID []byte, objType ObjectType, oldAt int64, newAt int64) error {
	oldKey := expireKey(mkey, oldAt)
	newKey := expireKey(mkey, newAt)

	if oldAt > 0 {
		if err := txn.Delete(oldKey); err != nil {
			return err
		}
	}

	if newAt > 0 {
		if err := txn.Set(newKey, objID); err != nil {
			return err
		}
	}
	action := ""
	if oldAt > 0 && newAt > 0 {
		action = "updated"
	} else if oldAt > 0 {
		action = "removed"
	} else if newAt > 0 {
		action = "added"
	}
	if action != "" {
		metrics.GetMetrics().ExpireKeysTotal.WithLabelValues(action).Inc()
	}
	return nil
}

func unExpireAt(txn store.Transaction, mkey []byte, expireAt int64) error {
	if expireAt == 0 {
		return nil
	}
	oldKey := expireKey(mkey, expireAt)
	if err := txn.Delete(oldKey); err != nil {
		return err
	}
	metrics.GetMetrics().ExpireKeysTotal.WithLabelValues("removed").Inc()
	return nil
}

// StartExpire get leader from db
func StartExpire(db *DB, conf *conf.Expire) error {
	ticker := time.NewTicker(conf.Interval)
	defer ticker.Stop()
	id := UUID()
	lastExpireEndTs := int64(0)
	for range ticker.C {
		if conf.Disable {
			continue
		}

		start := time.Now()
		isLeader, err := isLeader(db, sysExpireLeader, id, conf.LeaderLifeTime)
		if err != nil {
			zap.L().Error("[Expire] check expire leader failed", zap.Error(err))
			continue
		}
		if !isLeader {
			if logEnv := zap.L().Check(zap.DebugLevel, "[Expire] not expire leader"); logEnv != nil {
				logEnv.Write(zap.ByteString("leader", sysExpireLeader),
					zap.ByteString("uuid", id),
					zap.Duration("leader-life-time", conf.LeaderLifeTime))
			}
			continue
		}
		lastExpireEndTs = runExpire(db, conf.BatchLimit, lastExpireEndTs)
		metrics.GetMetrics().WorkerRoundCostHistogramVec.WithLabelValues(expire_worker).Observe(time.Since(start).Seconds())
	}
	return nil
}

// split a meta key with format: {namespace}:{id}:M:{key}
func splitMetaKey(key []byte) ([]byte, DBID, []byte) {
	idx := bytes.Index(key, []byte{':'})
	namespace := key[:idx]
	id := toDBID(key[idx+1 : idx+4])
	rawkey := key[idx+6:]
	return namespace, id, rawkey
}

func toTikvDataKey(namespace []byte, id DBID, key []byte) []byte {
	var b []byte
	b = append(b, namespace...)
	b = append(b, ':')
	b = append(b, id.Bytes()...)
	b = append(b, ':', 'D', ':')
	b = append(b, key...)
	return b
}

func toTikvScorePrefix(namespace []byte, id DBID, key []byte) []byte {
	var b []byte
	b = append(b, namespace...)
	b = append(b, ':')
	b = append(b, id.Bytes()...)
	b = append(b, ':', 'S', ':')
	b = append(b, key...)
	return b
}

func runExpire(db *DB, batchLimit int, lastExpireEndTs int64) int64 {
	txn, err := db.Begin()
	if err != nil {
		zap.L().Error("[Expire] txn begin failed", zap.Error(err))
		return 0
	}

	now := time.Now().UnixNano()
	//iter get keys [key, upperBound), so using now+1 as 2nd parameter will get "at:now:" prefixed keys
	//we seek end in "at:<now>" replace in "at;" , it can reduce the seek range and seek the deleted expired keys as little as possible.
	//the behavior should reduce the expire delay in days and get/mget timeout, which are caused by rocksdb tomstone problem
	var endPrefix []byte
	endPrefix = append(endPrefix, expireKeyPrefix...)
	endPrefix = append(endPrefix, EncodeInt64(now+1)...)

	var startPrefix []byte
	if lastExpireEndTs > 0 {
		startPrefix = append(startPrefix, expireKeyPrefix...)
		startPrefix = append(startPrefix, EncodeInt64(lastExpireEndTs)...)
		startPrefix = append(startPrefix, ':')
	} else {
		startPrefix = expireKeyPrefix
	}

	start := time.Now()
	iter, err := txn.t.Iter(startPrefix, endPrefix)
	metrics.GetMetrics().WorkerSeekCostHistogramVec.WithLabelValues(expire_worker).Observe(time.Since(start).Seconds())
	if logEnv := zap.L().Check(zap.DebugLevel, "[Expire] seek expire keys"); logEnv != nil {
		logEnv.Write(zap.Int64("[startTs", lastExpireEndTs), zap.Int64("endTs)", now+1))
	}
	if err != nil {
		zap.L().Error("[Expire] seek failed", zap.ByteString("prefix", expireKeyPrefix), zap.Error(err))
		txn.Rollback()
		return 0
	}
	limit := batchLimit

	thisExpireEndTs := int64(0)
	ts := now
	for iter.Valid() && iter.Key().HasPrefix(expireKeyPrefix) && limit > 0 {
		rawKey := iter.Key()
		ts = DecodeInt64(rawKey[expireTimestampOffset : expireTimestampOffset+8])
		if ts > now {
			if logEnv := zap.L().Check(zap.DebugLevel, "[Expire] not need to expire key"); logEnv != nil {
				logEnv.Write(zap.String("raw-key", string(rawKey)), zap.Int64("last-timestamp", ts))
			}
			break
		}
		mkey := rawKey[expireMetakeyOffset:]
		if err := doExpire(txn, mkey, iter.Value()); err != nil {
			txn.Rollback()
			return 0
		}

		// Remove from expire list
		if err := txn.t.Delete(rawKey); err != nil {
			zap.L().Error("[Expire] delete failed",
				zap.ByteString("mkey", mkey),
				zap.Error(err))
			txn.Rollback()
			return 0
		}

		if logEnv := zap.L().Check(zap.DebugLevel, "[Expire] delete expire list item"); logEnv != nil {
			logEnv.Write(zap.Int64("ts", ts), zap.ByteString("mkey", mkey))
		}

		start = time.Now()
		err := iter.Next()
		cost := time.Since(start)
		if cost >= time.Millisecond {
			metrics.GetMetrics().WorkerSeekCostHistogramVec.WithLabelValues(expire_worker).Observe(cost.Seconds())
		}
		if err != nil {
			zap.L().Error("[Expire] next failed",
				zap.ByteString("mkey", mkey),
				zap.Error(err))
			txn.Rollback()
			return 0
		}

		//just use the latest processed expireKey(don't include the last expire key in the loop which is > now) as next seek's start key
		thisExpireEndTs = ts
		limit--
	}

	now = time.Now().UnixNano()
	diff := (ts - now) / int64(time.Second)
	if diff >= 0 {
		metrics.GetMetrics().ExpireLeftSecondsVec.WithLabelValues("left").Set(float64(diff))
		metrics.GetMetrics().ExpireLeftSecondsVec.WithLabelValues("delay").Set(0)
	} else {
		metrics.GetMetrics().ExpireLeftSecondsVec.WithLabelValues("delay").Set(float64(-1 * diff))
		metrics.GetMetrics().ExpireLeftSecondsVec.WithLabelValues("left").Set(0)
	}

	start = time.Now()
	err = txn.Commit(context.Background())
	metrics.GetMetrics().WorkerCommitCostHistogramVec.WithLabelValues(expire_worker).Observe(time.Since(start).Seconds())
	if err != nil {
		txn.Rollback()
		zap.L().Error("[Expire] commit failed", zap.Error(err))
	}

	if logEnv := zap.L().Check(zap.DebugLevel, "[Expire] expired end"); logEnv != nil {
		logEnv.Write(zap.Int("expired_num", batchLimit-limit))
	}

	metrics.GetMetrics().ExpireKeysTotal.WithLabelValues("expired").Add(float64(batchLimit - limit))
	return thisExpireEndTs
}

func gcDataKey(txn *Transaction, namespace []byte, dbid DBID, key, id []byte) error {
	dkey := toTikvDataKey(namespace, dbid, id)
	if err := gc(txn.t, dkey); err != nil {
		zap.L().Error("[Expire] gc failed",
			zap.ByteString("key", key),
			zap.ByteString("namepace", namespace),
			zap.Int64("db_id", int64(dbid)),
			zap.ByteString("obj_id", id),
			zap.Error(err))
		return err
	}
	if logEnv := zap.L().Check(zap.DebugLevel, "[Expire] gc data key"); logEnv != nil {
		logEnv.Write(zap.ByteString("obj_id", id))
	}
	return nil
}
func doExpire(txn *Transaction, mkey, id []byte) error {
	namespace, dbid, key := splitMetaKey(mkey)
	obj, err := getObject(txn, mkey)
	// Check for dirty data due to copying or flushdb/flushall
	if err == ErrKeyNotFound {
		return gcDataKey(txn, namespace, dbid, key, id)
	}
	if err != nil {
		return err
	}
	idLen := len(obj.ID)
	if len(id) > idLen {
		id = id[:idLen]
	}
	if !bytes.Equal(obj.ID, id) {
		return gcDataKey(txn, namespace, dbid, key, id)
	}

	// Delete object meta
	if err := txn.t.Delete(mkey); err != nil {
		zap.L().Error("[Expire] delete failed",
			zap.ByteString("key", key),
			zap.Error(err))
		return err
	}

	if logEnv := zap.L().Check(zap.DebugLevel, "[Expire] delete metakey"); logEnv != nil {
		logEnv.Write(zap.ByteString("mkey", mkey))
	}
	if obj.Type == ObjectString {
		return nil
	}
	return gcDataKey(txn, namespace, dbid, key, id)
}
