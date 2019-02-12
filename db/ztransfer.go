package db

import (
	"context"
	"time"

	"github.com/meitu/titan/conf"
	"github.com/meitu/titan/metrics"
	"go.uber.org/zap"
)

var (
	sysZTLeader              = []byte("$sys:0:ZTL:ZTLeader")
	sysZTKeyPrefixLength     = len(toZTKey([]byte{}))
	sysZTLeaderFlushInterval = 10 * time.Second

	ztQueue chan []byte
)

// loadZList is read only, so ZList did not call Destroy()
func loadZList(txn *Transaction, metaKey []byte) (*ZList, error) {
	val, err := txn.t.Get(metaKey)
	if err != nil {
		return nil, err
	}

	obj, err := DecodeObject(val)
	if err != nil {
		return nil, err
	}
	if obj.Type != ObjectList {
		return nil, ErrTypeMismatch
	}
	if obj.Encoding != ObjectEncodingZiplist {
		zap.L().Error("[ZT] error in trans zlist, encoding type error", zap.Error(err))
		return nil, ErrEncodingMismatch
	}
	if obj.ExpireAt != 0 && obj.ExpireAt < Now() {
		return nil, ErrKeyNotFound
	}

	l := &ZList{
		rawMetaKey: metaKey,
		txn:        txn,
	}
	if err = l.Unmarshal(obj, val); err != nil {
		return nil, err
	}
	return l, nil
}

// toZTKey convert meta key to ZT key
// {sys.ns}:{sys.id}:{ZT}:{metakey}
// NOTE put this key to sys db.
func toZTKey(metakey []byte) []byte {
	b := []byte{}
	b = append(b, sysNamespace...)
	b = append(b, ':', byte(sysDatabaseID))
	b = append(b, ':', 'Z', 'T', ':')
	b = append(b, metakey...)
	return b
}

// PutZList should be called after ZList created
func PutZList(txn *Transaction, metakey []byte) error {
	zap.L().Debug("[ZT] Zlist recorded in txn", zap.String("key", string(metakey)))
	return txn.t.Set(toZTKey(metakey), []byte{0})
}

// RemoveZTKey remove an metakey from ZT
func RemoveZTKey(txn *Transaction, metakey []byte) error {
	return txn.t.Delete(toZTKey(metakey))
}

// doZListTransfer get zt key, create zlist and transfer to llist, after that, delete zt key
func doZListTransfer(txn *Transaction, metakey []byte) (int, error) {
	zlist, err := loadZList(txn, metakey)
	if err != nil {
		if err == ErrTypeMismatch || err == ErrEncodingMismatch || err == ErrKeyNotFound {
			if err = RemoveZTKey(txn, metakey); err != nil {
				zap.L().Error("[ZT] error in remove ZTKkey", zap.Error(err))
				return 0, err
			}
			return 0, nil
		}
		zap.L().Error("[ZT] error in create zlist", zap.Error(err))
		return 0, err
	}

	llist, err := zlist.TransferToLList(splitMetaKey(metakey))
	if err != nil {
		zap.L().Error("[ZT] error in convert zlist", zap.Error(err))
		return 0, err
	}
	// clean the zt key, after success
	if err = RemoveZTKey(txn, metakey); err != nil {
		zap.L().Error("[ZT] error in remove ZTKkey", zap.Error(err))
		return 0, err
	}

	return int(llist.Len), nil
}

func ztWorker(db *DB, batch int, interval time.Duration) {
	var txn *Transaction
	var err error
	var n int

	txnstart := false
	batchCount := 0
	sum := 0
	commit := func(t *Transaction) {
		if err = t.Commit(context.Background()); err != nil {
			zap.L().Error("[ZT] error in commit transfer", zap.Error(err))
			txn.Rollback()
		} else {
			metrics.GetMetrics().ZTInfoCounterVec.WithLabelValues("zlist").Add(float64(batchCount))
			metrics.GetMetrics().ZTInfoCounterVec.WithLabelValues("key").Add(float64(sum))
			zap.L().Debug("[ZT] transfer zlist succeed", zap.Int("count", batchCount), zap.Int("n", sum))
		}
		txnstart = false
		batchCount = 0
		sum = 0
	}

	// create zlist and transfer to llist, after that, delete zt key
	for {
		select {
		case metakey := <-ztQueue:
			if !txnstart {
				if txn, err = db.Begin(); err != nil {
					zap.L().Error("[ZT] zt worker error in kv begin", zap.Error(err))
					continue
				}
				txnstart = true
			}

			if n, err = doZListTransfer(txn, metakey); err != nil {
				txn.Rollback()
				txnstart = false
				continue
			}
			sum += n
			batchCount++
			if batchCount >= batch {
				commit(txn)
			}
		default:
			if batchCount > 0 {
				commit(txn)
			} else {
				time.Sleep(interval)
				txnstart = false
			}
		}
	}
}

func runZT(db *DB, prefix []byte, tick <-chan time.Time) ([]byte, error) {
	txn, err := db.Begin()
	if err != nil {
		zap.L().Error("[ZT] error in kv begin", zap.Error(err))
		return toZTKey(nil), nil
	}
	iter, err := txn.t.Iter(prefix, nil)
	if err != nil {
		zap.L().Error("[ZT] error in seek", zap.ByteString("prefix", prefix), zap.Error(err))
		return toZTKey(nil), err
	}

	for ; iter.Valid() && iter.Key().HasPrefix(prefix); err = iter.Next() {
		if err != nil {
			zap.L().Error("[ZT] error in iter next", zap.Error(err))
			return toZTKey(nil), err
		}
		select {
		case ztQueue <- iter.Key()[sysZTKeyPrefixLength:]:
		case <-tick:
			return iter.Key(), nil
		default:
			return iter.Key(), nil
		}
	}
	zap.L().Debug("[ZT] no more ZT item, retrive iterator")
	return toZTKey(nil), txn.Commit(context.Background())
}

// StartZT start ZT fill in the queue(channel), and start the worker to consume.
func StartZT(db *DB, conf *conf.ZT) {
	ztQueue = make(chan []byte, conf.QueueDepth)
	for i := 0; i < conf.Wrokers; i++ {
		go ztWorker(db, conf.BatchCount, conf.Interval)
	}

	// check leader and fill the channel
	prefix := toZTKey(nil)
	ticker := time.NewTicker(conf.Interval)
	defer ticker.Stop()
	id := UUID()
	for range ticker.C {
		isLeader, err := isLeader(db, sysZTLeader, id, sysZTLeaderFlushInterval)
		if err != nil {
			zap.L().Error("[ZT] check ZT leader failed",
				zap.Int64("dbid", int64(db.ID)),
				zap.Error(err))
			continue
		}
		if !isLeader {
			zap.L().Debug("[ZT] not ZT leader")
			continue
		}

		if prefix, err = runZT(db, prefix, ticker.C); err != nil {
			zap.L().Error("[ZT] error in run ZT",
				zap.Int64("dbid", int64(db.ID)),
				zap.ByteString("prefix", prefix),
				zap.Error(err))
			continue
		}
	}
}
