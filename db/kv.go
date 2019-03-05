package db

import (
	"bytes"
	"context"
	"errors"
	"math/rand"
	"sync"

	"github.com/meitu/titan/db/store"
	"github.com/pingcap/kvproto/pkg/kvrpcpb"
	"github.com/pingcap/kvproto/pkg/metapb"
	"github.com/pingcap/tidb/store/tikv"
	"github.com/pingcap/tidb/store/tikv/tikvrpc"
	"go.uber.org/zap"
)

// Kv supplies key releated operations
type Kv struct {
	txn *Transaction
}

// GetKv returns a Kv object
func GetKv(txn *Transaction) *Kv {
	return &Kv{txn}
}

// Keys iterator all keys in db
func (kv *Kv) Keys(start []byte, f func(key []byte) bool) error {
	mkey := MetaKey(kv.txn.db, start)
	prefix := MetaKey(kv.txn.db, nil)
	iter, err := kv.txn.t.Iter(mkey, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	now := Now()
	for iter.Valid() {
		key := iter.Key()
		if !bytes.HasPrefix(key, prefix) {
			break
		}

		obj, err := DecodeObject(iter.Value())
		if err != nil {
			return err
		}
		if !IsExpired(obj, now) && !f(key[len(prefix):]) {
			break
		}
		if err := iter.Next(); err != nil {
			return err
		}
	}
	return nil
}

// Delete specific keys, ignore if non exist
func (kv *Kv) Delete(keys [][]byte) (int64, error) {
	var (
		count    int64
		metaKeys [][]byte
		mapping  = make(map[string][]byte)
		now      = Now()
	)
	// use mapping to filter duplicate keys
	for _, key := range keys {
		mkey := MetaKey(kv.txn.db, key)
		if _, ok := mapping[string(mkey)]; !ok {
			mapping[string(mkey)] = key
			metaKeys = append(metaKeys, mkey)
		}
	}

	values, err := store.BatchGetValues(kv.txn.t, metaKeys)
	if err != nil {
		return count, err
	}
	for k, val := range values {
		if val != nil {
			obj, err := DecodeObject(val)
			if err != nil {
				return count, err
			}
			if IsExpired(obj, now) {
				continue
			}
			if obj.Type == ObjectHash {
				hash, err := kv.txn.Hash(mapping[k])
				if err != nil {
					return count, err
				}
				if err := hash.Destroy(); err != nil {
					return count, err
				}
			} else if err := kv.txn.Destory(obj, mapping[k]); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}

// ExpireAt set a timeout on key
func (kv *Kv) ExpireAt(key []byte, at int64) error {
	mkey := MetaKey(kv.txn.db, key)
	now := Now()

	meta, err := kv.txn.t.Get(mkey)
	if err != nil {
		if IsErrNotFound(err) {
			return ErrKeyNotFound
		}
		return err
	}
	obj, err := DecodeObject(meta)
	if err != nil {
		return err
	}
	if IsExpired(obj, now) {
		return ErrKeyNotFound
	}
	if at == 0 && obj.ExpireAt != 0 {
		if err = unExpireAt(kv.txn.t, mkey, obj.ExpireAt); err != nil {
			return err
		}
	}

	if at > 0 {
		if err := expireAt(kv.txn.t, mkey, obj.ID, obj.ExpireAt, at); err != nil {
			return err
		}
	}
	obj.ExpireAt = at
	updated := EncodeObject(obj)
	updated = append(updated, meta[ObjectEncodingLength:]...)
	return kv.txn.t.Set(mkey, updated)
}

//Exists check if the given keys exist
func (kv *Kv) Exists(keys [][]byte) (int64, error) {
	var count int64
	now := Now()
	mkeys := make([][]byte, len(keys))
	for i, key := range keys {
		mkeys[i] = MetaKey(kv.txn.db, key)
	}

	values, err := store.BatchGetValues(kv.txn.t, mkeys)
	if err != nil {
		return count, err
	}
	for _, val := range values {
		if val != nil {
			obj, err := DecodeObject(val)
			if err != nil {
				return count, err
			}
			if IsExpired(obj, now) {
				continue
			}
			count++
		}
	}
	return count, nil
}

// FlushDB clear current db.
func (kv *Kv) FlushDB(ctx context.Context) error {
	prefix := kv.txn.db.Prefix()
	nextID := kv.txn.db.ID + 1
	endKey := dbPrefix(kv.txn.db.Namespace, nextID.Bytes())

	if err := unsafeDeleteRange(ctx, kv.txn.db, prefix, endKey); err != nil {
		zap.L().Error("flushdb data unsafe clear err",
			zap.ByteString("start", prefix),
			zap.ByteString("end", endKey),
			zap.Error(err))

		return ErrStorageRetry
	}

	if err := clearSysRangeData(ctx, kv.txn.db, prefix, endKey); err != nil {
		return ErrStorageRetry
	}

	return nil
}

// FlushAll clean up all databases.
func (kv *Kv) FlushAll(ctx context.Context) error {
	prefix := kv.txn.db.Prefix()
	maxID := EncodeInt64(256)
	endKey := dbPrefix(kv.txn.db.Namespace, maxID)
	if err := unsafeDeleteRange(ctx, kv.txn.db, prefix, endKey); err != nil {
		zap.L().Error("flushall data unsafe clear err",
			zap.ByteString("start", prefix),
			zap.ByteString("end", endKey),
			zap.Error(err))
		return ErrStorageRetry
	}
	sysStart := sysPrefix(sysNamespace, byte(sysDatabaseID))
	sysEnd := sysPrefix(sysNamespace, byte(sysDatabaseID+1))
	if err := unsafeDeleteRange(ctx, kv.txn.db, sysStart, sysEnd); err != nil {
		zap.L().Error("flushall sys data unsafe clear err",
			zap.ByteString("start", sysStart),
			zap.ByteString("end", sysEnd),
			zap.Error(err))

		return ErrStorageRetry
	}

	return nil
}

// RandomKey return a key from current db randomly
// Now we use an static length(64) to generate the key spaces, it means it is random for keys
// that len(key) <= 64, it is enough for most cases
func (kv *Kv) RandomKey() ([]byte, error) {
	buf := make([]byte, 64)
	// Read for rand here always return a nil error
	rand.Read(buf)

	mkey := MetaKey(kv.txn.db, buf)
	prefix := MetaKey(kv.txn.db, nil)

	// Iter >= mkey
	iter, err := kv.txn.t.Iter(mkey, nil)
	if err != nil {
		return nil, err
	}

	if iter.Valid() && iter.Key().HasPrefix(prefix) {
		return iter.Key()[len(prefix):], nil
	}
	first := make([]byte, len(prefix)+1)
	copy(first, prefix)
	iter, err = kv.txn.t.Iter(first, nil)
	if err != nil {
		return nil, err
	}

	if iter.Valid() && iter.Key().HasPrefix(prefix) {
		return iter.Key()[len(prefix):], nil
	}
	return nil, err
}

//clear system range data(GC/ZT)
func clearSysRangeData(ctx context.Context, db *DB, startKey, endKey []byte) error {
	gcStart := toTikvGCKey(startKey)
	gcEnd := toTikvGCKey(endKey)
	if err := unsafeDeleteRange(ctx, db, gcStart, gcEnd); err != nil {
		zap.L().Error("[GC] unsafe clear err",
			zap.ByteString("start", gcStart),
			zap.ByteString("end", gcEnd),
			zap.Error(err))
		return err
	}

	ztStart := toZTKey(startKey)
	ztEnd := toZTKey(endKey)
	if err := unsafeDeleteRange(ctx, db, ztStart, ztEnd); err != nil {
		zap.L().Error("[ZT] unsafe clear err",
			zap.ByteString("start", ztStart),
			zap.ByteString("end", ztEnd),
			zap.Error(err))
		return err
	}
	return nil
}

func unsafeDeleteRange(ctx context.Context, db *DB, startKey, endKey []byte) error {
	storage, ok := db.kv.Storage.(tikv.Storage)
	if !ok {
		zap.L().Error("delete ranges: storage conversion PDClient failed")
		return errors.New("Storage not available")
	}
	stores, err := storage.GetRegionCache().PDClient().GetAllStores(ctx)
	if err != nil {
		zap.L().Error("delete ranges: got an error while trying to get store list from PD:", zap.Error(err))
		return err
	}

	req := &tikvrpc.Request{
		Type: tikvrpc.CmdUnsafeDestroyRange,
		UnsafeDestroyRange: &kvrpcpb.UnsafeDestroyRangeRequest{
			StartKey: startKey,
			EndKey:   endKey,
		},
	}
	tikvCli := storage.GetTiKVClient()

	var wg sync.WaitGroup
	for _, store := range stores {
		if store.State != metapb.StoreState_Up {
			continue
		}

		address := store.Address
		storeID := store.Id
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, storeErr := tikvCli.SendRequest(ctx, address, req, tikv.UnsafeDestroyRangeTimeout)
			if storeErr != nil {
				zap.L().Error("destroy range on store  failed with ",
					zap.Uint64("store_id", storeID),
					zap.String("addr", address),
					zap.ByteString("start", startKey),
					zap.ByteString("end", endKey),
					zap.Error(storeErr))
				err = storeErr
			}
		}()
	}
	wg.Wait()
	return err
}
