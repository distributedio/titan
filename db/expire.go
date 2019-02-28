package db

import (
	"bytes"
	"context"
	"time"

	"github.com/meitu/titan/conf"
	"github.com/meitu/titan/db/store"
	"github.com/meitu/titan/metrics"
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

func expireAt(txn store.Transaction, mkey []byte, objID []byte, oldAt int64, newAt int64) error {
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
		isLeader, err := isLeader(db, sysExpireLeader, id, conf.LeaderLifeTime)
		if err != nil {
			zap.L().Error("[Expire] check expire leader failed", zap.Error(err))
			continue
		}
		if !isLeader {
			zap.L().Debug("[Expire] not expire leader")
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

func runExpire(db *DB, batchLimit int) {
	txn, err := db.Begin()
	if err != nil {
		zap.L().Error("[Expire] txn begin failed", zap.Error(err))
		return
	}
	iter, err := txn.t.Iter(expireKeyPrefix, nil)
	if err != nil {
		zap.L().Error("[Expire] seek failed", zap.ByteString("prefix", expireKeyPrefix), zap.Error(err))
		txn.Rollback()
		return
	}
	limit := batchLimit
	now := time.Now().UnixNano()
	for iter.Valid() && iter.Key().HasPrefix(expireKeyPrefix) && limit > 0 {
		key := iter.Key()
		val := iter.Value()
		mkey := key[expireMetakeyOffset:]
		namespace, dbid, rawkey := splitMetaKey(mkey)

		ts := DecodeInt64(key[expireTimestampOffset : expireTimestampOffset+8])
		if ts > now {
			break
		}

		//get obj info
		obj, err := getObject(txn, mkey)
		if err != nil {
			txn.Rollback()
			return
		}

		// Delete object meta
		if bytes.Equal(obj.ID, val) {
			if err := txn.t.Delete(mkey); err != nil {
				zap.L().Error("[Expire] delete failed",
					zap.ByteString("key", rawkey),
					zap.Error(err))
				return
			}
		}

		zap.L().Debug("[Expire] delete metakey", zap.ByteString("mkey", mkey), zap.String("key", string(rawkey)))
		// Remove from expire list
		if err := txn.t.Delete(key); err != nil {
			zap.L().Error("[Expire] delete failed",
				zap.ByteString("key", rawkey),
				zap.Error(err))
			txn.Rollback()
			return
		}

		//Need gc two types of data:
		//1.Normally expired data that requires gc to fall back to the composite data type
		//2.Overwritten Writing Requires GC to drop old data.(String override string will also be added to gc, even if string type data does not require gc data)
		if obj.Type != ObjectString || !bytes.Equal(obj.ID, val) {
			if err := gc(txn.t, toTikvDataKey(namespace, dbid, val)); err != nil {
				zap.L().Error("[Expire] gc failed",
					zap.ByteString("key", rawkey),
					zap.ByteString("namepace", namespace),
					zap.Int64("dbid", int64(dbid)),
					zap.ByteString("objid", val),
					zap.Error(err))
				txn.Rollback()
				return
			}
		}
		// Remove from expire list
		if err := iter.Next(); err != nil {
			zap.L().Error("[Expire] next failed",
				zap.ByteString("key", rawkey),
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
	metrics.GetMetrics().ExpireKeysTotal.WithLabelValues("expired").Add(float64(batchLimit - limit))
}
