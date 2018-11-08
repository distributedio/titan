package db

import (
	"bytes"
	"context"
	"time"

	"go.uber.org/zap"
)

const (
	expireBatchLimit = 256
	expireTick       = time.Duration(time.Second)
)

var (
	expireKeyPrefix              = []byte("$sys:0:at:")
	sysExpireLeader              = []byte("$sys:0:EXL:EXLeader")
	sysExpireLeaderFlushInterval = 10

	// $sys:0:at:{ts}:{metaKey}
	expireTimestampOffset = len(expireKeyPrefix)
	expireMetakeyOffset   = expireTimestampOffset + 8 /*sizeof(int64)*/ + len(":")
)

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

func expireAt(txn *Transaction, mkey []byte, objID []byte, old int64, new int64) error {
	oldKey := expireKey(mkey, old)
	newKey := expireKey(mkey, new)

	if old > 0 {
		if err := txn.t.Delete(oldKey); err != nil {
			return err
		}
	}

	if new > 0 {
		if err := txn.t.Set(newKey, objID); err != nil {
			return err
		}
	}
	return nil
}

func unExpireAt(txn *Transaction, mkey []byte, expireAt int64) error {
	oldKey := expireKey(mkey, expireAt)
	if err := txn.t.Delete(oldKey); err != nil {
		return err
	}
	return nil
}

func StartExpire(db *DB) error {
	ticker := time.NewTicker(expireTick)
	defer ticker.Stop()
	for _ = range ticker.C {
		isLeader, err := isLeader(db, sysExpireLeader, time.Duration(sysExpireLeaderFlushInterval))
		if err != nil {
			zap.L().Error("check expire leader failed", zap.Error(err))
			continue
		}
		if !isLeader {
			zap.L().Debug("not expire leader")
			continue
		}
		runExpire(db)
	}
	return nil
}

// split a meta key with format: {namespace}:{id}:M:{key}
func splitMetaKey(key []byte) ([]byte, byte, []byte) {
	idx := bytes.Index(key, []byte{':'})
	namespace := key[:idx]
	id := key[idx+1]
	rawkey := key[idx+5:]
	return namespace, id, rawkey
}
func toTikvDataKey(namespace []byte, id byte, key []byte) []byte {
	var b []byte
	b = append(b, namespace...)
	b = append(b, ':', id)
	b = append(b, ':', 'D', ':')
	b = append(b, key...)
	return b
}

func runExpire(db *DB) {
	txn, err := db.Begin()
	if err != nil {
		zap.L().Error("transection begin failed", zap.Error(err))
		return
	}
	iter, err := txn.t.Seek(expireKeyPrefix)
	if err != nil {
		zap.L().Error("seek failed", zap.Error(err))
		txn.Rollback()
		return
	}
	limit := expireBatchLimit
	now := time.Now().UnixNano()
	for iter.Valid() && iter.Key().HasPrefix(expireKeyPrefix) && limit > 0 {
		key := iter.Key()
		objID := iter.Value()
		mkey := key[expireMetakeyOffset:]
		namespace, dbid, rawkey := splitMetaKey(mkey)

		ts := DecodeInt64(key[expireTimestampOffset : expireTimestampOffset+8])
		if ts > now {
			break
		}

		zap.L().Debug("expire", zap.String("key", string(rawkey)))
		// Delete object meta
		if err := txn.t.Delete(mkey); err != nil {
			zap.L().Error("delete failed", zap.Error(err))
			txn.Rollback()
			return
		}
		// Gc it if it is a complext data structure, the value of string is: []byte{'0'}
		if len(objID) > 1 {
			if err := gc(txn.t, toTikvDataKey(namespace, dbid, objID)); err != nil {
				zap.L().Error("gc failed", zap.Error(err))
				txn.Rollback()
				return
			}
		}
		// Remove from expire list
		if err := txn.t.Delete(iter.Key()); err != nil {
			zap.L().Error("delete failed", zap.Error(err))
			txn.Rollback()
			return
		}
		if err := iter.Next(); err != nil {
			zap.L().Error("next failed", zap.Error(err))
			txn.Rollback()
			return
		}
		limit--
	}

	if err := txn.Commit(context.Background()); err != nil {
		txn.Rollback()
		zap.L().Error("commit failed", zap.Error(err))
	}
}
