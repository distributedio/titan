package db

import (
	"context"
	"time"

	log "gitlab.meitu.com/gocommons/logbunny"
	"gitlab.meitu.com/platform/thanos/conf"
	"gitlab.meitu.com/platform/titan/monitor"
)

var (
	sysZTLeader              = []byte("$sys:0:ZTL:ZTLeader")
	sysZTKeyPrefixLength     = len(toZTKey([]byte{}))
	sysZTLeaderFlushInterval = 10

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
		log.Error("[ZT] error in trans zlist, encoding type error", log.Err(err))
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
	//	b = append(b, sysNamespace...)
	//	b = append(b, ':', byte(sysDatabaseID))
	b = append(b, ':', 'Z', 'T', ':')
	b = append(b, metakey...)
	return b
}

// PutZList should be called after ZList created
func PutZList(txn *Transaction, metakey []byte) error {
	log.Debug("[ZT] Zlist recorded in txn", log.String("key", string(metakey)))
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
				log.Error("[ZT] error in remove ZTKkey", log.Err(err))
				return 0, err
			}
			return 0, nil
		}
		log.Error("[ZT] error in create zlist", log.Err(err))
		return 0, err
	}

	//llist, err := zlist.TransferToLList(splitMetaKey(metakey))
	llist, err := zlist.TransferToLList(nil, 1, nil)
	if err != nil {
		log.Error("[ZT] error in convert zlist", log.Err(err))
		return 0, err
	}
	// clean the zt key, after success
	if err = RemoveZTKey(txn, metakey); err != nil {
		log.Error("[ZT] error in remove ZTKkey", log.Err(err))
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
			log.Error("[ZT] error in commit transfer", log.Err(err))
			txn.Rollback()
		} else {
			monitor.WithLabelCounter(monitor.ZTInfoType, "zlist").Add(float64(batchCount))
			monitor.WithLabelCounter(monitor.ZTInfoType, "keys").Add(float64(sum))
			log.Debug("[ZT] transfer zlist succeed", log.Int("count", batchCount), log.Int("n", sum))
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
					log.Error("[ZT] zt worker error in kv begin", log.Err(err))
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
		log.Error("[ZT] error in kv begin", log.Err(err))
		return toZTKey(nil), nil
	}
	iter, err := txn.t.Seek(prefix)
	if err != nil {
		log.Error("[ZT] error in seek", log.Err(err))
		return toZTKey(nil), err
	}

	for ; iter.Valid() && iter.Key().HasPrefix(prefix); err = iter.Next() {
		if err != nil {
			log.Error("[ZT] error in iter next", log.Err(err))
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
	log.Debug("[ZT] no more ZT item, retrive iterator")
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
	tick := time.Tick(conf.Interval)
	for _ = range tick {
		isLeader, err := isLeader(db, sysZTLeader, time.Duration(sysZTLeaderFlushInterval))
		if err != nil {
			log.Error("[ZT] check ZT leader failed", log.Err(err))
			continue
		}
		if !isLeader {
			log.Debug("[ZT] not ZT leader")
			continue
		}

		if prefix, err = runZT(db, prefix, tick); err != nil {
			log.Error("[ZT] error in run ZT", log.Err(err))
			continue
		}
	}
}
