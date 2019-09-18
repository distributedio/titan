package db

import (
	"bytes"
	"context"
	"time"

	"github.com/distributedio/titan/conf"
	"github.com/distributedio/titan/db/store"
	"github.com/distributedio/titan/metrics"
	"github.com/pingcap/tidb/kv"
	"go.uber.org/zap"
)

var (
	expireKeyPrefix = []byte("$sys:0:at:")
	sysExpireLeader = []byte("$sys:0:EXL:EXLeader")

	// $sys:0:at:{ts}:{metaKey}
	expireTimestampOffset = len(expireKeyPrefix)
	expireMetakeyOffset   = expireTimestampOffset + 8 /*sizeof(int64)*/ + len(":")
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
	for range ticker.C {
		if conf.Disable {
			continue
		}
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
		runExpire(db, conf.BatchLimit)
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

func runExpire(db *DB, batchLimit int) {
	txn, err := db.Begin()
	if err != nil {
		zap.L().Error("[Expire] txn begin failed", zap.Error(err))
		return
	}
	endPrefix := kv.Key(expireKeyPrefix).PrefixNext()
	iter, err := txn.t.Iter(expireKeyPrefix, endPrefix)
	if err != nil {
		zap.L().Error("[Expire] seek failed", zap.ByteString("prefix", expireKeyPrefix), zap.Error(err))
		txn.Rollback()
		return
	}
	limit := batchLimit
	now := time.Now().UnixNano()

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
			return
		}

		// Remove from expire list
		if err := txn.t.Delete(rawKey); err != nil {
			zap.L().Error("[Expire] delete failed",
				zap.ByteString("mkey", mkey),
				zap.Error(err))
			txn.Rollback()
			return
		}

		if logEnv := zap.L().Check(zap.DebugLevel, "[Expire] delete expire list item"); logEnv != nil {
			logEnv.Write(zap.ByteString("mkey", mkey))
		}

		if err := iter.Next(); err != nil {
			zap.L().Error("[Expire] next failed",
				zap.ByteString("mkey", mkey),
				zap.Error(err))
			txn.Rollback()
			return
		}
		limit--
	}

	if err := txn.Commit(context.Background()); err != nil {
		txn.Rollback()
		zap.L().Error("[Expire] commit failed", zap.Error(err))
	}

	if logEnv := zap.L().Check(zap.DebugLevel, "[Expire] expired end"); logEnv != nil {
		logEnv.Write(zap.Int("expired_num", batchLimit-limit))
	}

	diff := (ts - now) / int64(time.Second)
	if diff >= 0 {
		metrics.GetMetrics().ExpireLeftSecondsVec.WithLabelValues("left").Set(float64(diff))
	} else {
		metrics.GetMetrics().ExpireLeftSecondsVec.WithLabelValues("delay").Set(float64(-1 * diff))
	}

	metrics.GetMetrics().ExpireKeysTotal.WithLabelValues("expired").Add(float64(batchLimit - limit))
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
