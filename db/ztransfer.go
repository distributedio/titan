package db

import (
	"context"
	"time"

	log "gitlab.meitu.com/gocommons/logbunny"
	"gitlab.meitu.com/platform/titan/conf"
	"gitlab.meitu.com/platform/titan/monitor"
)

var (
	sysZTLeader          = []byte("$sys:0:ZTL:ZTLeader")
	sysZTKeyPrefixLength = len(toZTKey([]byte{}))

	ztQueue chan []byte
)

// loadZList is read only, so ZList did not call Destroy()
func loadZList(txn Transaction, metaKey []byte) (*ZList, error) {
	val, err := txn.Get(metaKey)
	if err != nil {
		return nil, err
	}

	obj, err := DecodeObject(val)
	if err != nil {
		return nil, err
	}
	if obj.Type != ObjectList || obj.Encoding != ObjectEncodingZiplist {
		return nil, ObjectTypeError
	}
	if obj.ExpireAt != 0 && obj.ExpireAt < CurrentTimestamp() {
		return nil, ErrNotExist
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
func PutZList(txn Transaction, metakey []byte) error {
	log.Debug("[ZT] Zlist recorded in txn", log.String("key", string(metakey)))
	return txn.Set(toZTKey(metakey), []byte{0})
}

// RemoveZTKey remove an metakey from ZT
func RemoveZTKey(txn Transaction, metakey []byte) error {
	return txn.Delete(toZTKey(metakey))
}

// doZListTransfer get zt key, create zlist and transfer to llist, after that, delete zt key
func doZListTransfer(txn Transaction, metakey []byte) (int, error) {
	zlist, err := loadZList(txn, metakey)
	if err != nil {
		if IsErrNotFound(err) {
			if err = RemoveZTKey(txn, metakey); err != nil {
				log.Error("[ZT] error in remove ZTKkey", log.Err(err))
				return 0, err
			}
			return 0, nil
		}
		log.Error("[ZT] error in create zlist", log.Err(err))
		return 0, err
	}

	llist, err := zlist.TransferToLList(splitMetaKey(metakey))
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

func (db *DB) ztWorker(batch int, interval time.Duration) {
	var txn Transaction
	var err error
	var n int

	txnstart := false
	batchCount := 0
	sum := 0
	commit := func() {
		if err = txn.Commit(context.Background()); err != nil {
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
				if txn, err = db.kv.Begin(); err != nil {
					log.Error("[ZT] zt worker error in kv begin", log.Err(err))
					continue
				}
				txnstart = true
			}

			if n, err = doZListTransfer(txn, metakey); err != nil {
				log.Error("[ZT] error in trans", log.Err(err))
				txn.Rollback()
				txnstart = false
				continue
			}
			sum += n
			batchCount++
			if batchCount >= batch {
				commit()
			}
		default:
			if batchCount > 0 {
				commit()
			} else {
				time.Sleep(interval)
				txnstart = false
			}
		}
	}
}

// StartZT start ZT fill in the queue(channel), and start the worker to consume.
func (db *DB) StartZT(conf *conf.ZT) {
	ztQueue = make(chan []byte, conf.QueueDepth)
	for i := 0; i < conf.Wrokers; i++ {
		go db.ztWorker(conf.BatchCount, conf.Interval)
	}

	var txn Transaction
	var iter Iterator
	prefix := toZTKey([]byte{})

	// check leader and fill the channel
	tick := time.Tick(conf.Interval)
	for _ = range tick {
		isLeader, err := db.isLeader(sysZTLeader)
		if err != nil {
			log.Error("[ZT] check ZT leader failed", log.Err(err))
			continue
		}

		if !isLeader {
			log.Debug("[ZT] not ZT leader")
			continue
		}

		if txn, err = db.kv.Begin(); err != nil {
			log.Error("[ZT] error in kv begin", log.Err(err))
			continue
		}
		if iter, err = txn.Seek(prefix); err != nil {
			log.Error("[ZT] error in seek", log.Err(err))
			continue
		}
		if !iter.Valid() || !iter.Key().HasPrefix(prefix) {
			log.Debug("[ZT] no ZT item")
			continue
		}

		var e error
		for ; iter.Valid() && iter.Key().HasPrefix(prefix); e = iter.Next() {
			if e != nil {
				log.Error("[ZT] error in iter next", log.Err(e))
				break
			}
			select {
			case ztQueue <- iter.Key()[sysZTKeyPrefixLength:]:
			default:
				break
			}
		}
		txn.Commit(context.Background())
	}
}
