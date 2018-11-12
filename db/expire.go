package db

import (
	"bytes"
	"context"
	"time"

	"gitlab.meitu.com/platform/thanos/db/store"

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
	return nil
}

// StartExpire get leader from db
func StartExpire(db *DB) error {
	ticker := time.NewTicker(expireTick)
	defer ticker.Stop()
	for _ = range ticker.C {
		isLeader, err := isLeader(db, sysExpireLeader, time.Duration(sysExpireLeaderFlushInterval))
		if err != nil {
			zap.L().Error("[Expire] check expire leader failed", zap.Error(err))
			continue
		}
		if !isLeader {
			zap.L().Debug("[Expire] not expire leader")
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
		zap.L().Error("[Expire] txn begin failed", zap.Error(err))
		return
	}
	iter, err := txn.t.Seek(expireKeyPrefix)
	if err != nil {
		zap.L().Error("[Expire] seek failed", zap.ByteString("prefix", expireKeyPrefix), zap.Error(err))
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

		zap.L().Debug("[Expire] delete metakey", zap.ByteString("mkey", mkey), zap.String("key", string(rawkey)))
		// Delete object meta
		if err := txn.t.Delete(mkey); err != nil {
			zap.L().Error("[Expire] delete failed",
				zap.ByteString("key", rawkey),
				zap.Error(err))
			txn.Rollback()
			return
		}
		// Gc it if it is a complext data structure, the value of string is: []byte{'0'}
		if len(objID) > 1 {
			if err := gc(txn.t, toTikvDataKey(namespace, dbid, objID)); err != nil {
				zap.L().Error("[Expire] gc failed",
					zap.ByteString("key", rawkey),
					zap.ByteString("namepace", namespace),
					zap.Int64("dbid", int64(dbid)),
					zap.ByteString("objid", objID),
					zap.Error(err))
				txn.Rollback()
				return
			}
		}
		// Remove from expire list
		if err := txn.t.Delete(iter.Key()); err != nil {
			zap.L().Error("[Expire] delete failed",
				zap.ByteString("key", rawkey),
				zap.Error(err))
			txn.Rollback()
			return
		}
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
}
