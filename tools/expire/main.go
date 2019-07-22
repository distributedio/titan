package main

import (
	"context"
	"flag"

	"github.com/distributedio/titan/db"
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/store/tikv"
	"go.uber.org/zap"
)

var log = zap.L()

type option struct {
	db        int
	namespace string
	batch     int
}

func deletePrefix(txn kv.Transaction, prefix kv.Key) error {
	end := prefix.PrefixNext()
	iter, err := txn.Iter(prefix, end)
	if err != nil {
		return err
	}
	for iter.Valid() && iter.Key().HasPrefix(prefix) {
		if err := txn.Delete(iter.Key()); err != nil {
			return err
		}
		if err := iter.Next(); err != nil {
			return err
		}
	}
	return nil
}
func cleanUp(txn kv.Transaction, mkey kv.Key, database *db.DB,
	obj *db.Object) error {
	switch obj.Type {
	case db.ObjectString:
		return txn.Delete(mkey)
	case db.ObjectZSet:
		dkey := db.DataKey(database, obj.ID)
		skey := db.ZSetScorePrefix(database, obj.ID)
		deletePrefix(txn, dkey) // ignore its error for simplicity
		return deletePrefix(txn, skey)
	default:
		dkey := db.DataKey(database, obj.ID)
		return txn.Delete(dkey)
	}
	return nil
}

func doExpire(s kv.Storage, database *db.DB, start kv.Key, limit int) (kv.Key, error) {
	txn, err := s.Begin()
	if err != nil {
		return nil, err
	}
	iter, err := txn.Iter(start, nil)
	if err != nil {
		return nil, err
	}
	// tolerate 5mins
	now := db.Now() - 300
	// scan the whole database
	var end kv.Key
	for iter.Valid() && limit != 0 {
		obj, err := db.DecodeObject(iter.Value())
		if err != nil {
			return nil, err
		}
		if obj.ExpireAt < now {
			if err := cleanUp(txn, iter.Key(), database, obj); err != nil {
				return nil, err
			}
		}
		limit--
		end = iter.Key()
		iter.Next()
	}
	if err := txn.Commit(context.Background()); err != nil {
		return nil, err
	}
	// scan done
	if limit > 0 {
		return nil, nil
	}
	return end.Next(), nil
}
func expire(opt *option, addr string) error {
	log.Info("start to expire", zap.Int("db", opt.db),
		zap.String("namespace", opt.namespace))

	database := &db.DB{Namespace: opt.namespace, ID: db.DBID(opt.db)}

	store, err := tikv.Driver{}.Open(addr)
	if err != nil {
		return err
	}

	start := db.MetaKey(database, nil)
	for {
		next, err := doExpire(store, database, start, opt.batch)
		if err != nil {
			return err
		}
		if next == nil {
			break
		}
		start = next
	}
	log.Info("finish to expire", zap.Int("db", opt.db),
		zap.String("namespace", opt.namespace))
	return nil
}
func main() {
	opt := &option{}
	flag.IntVar(&opt.db, "db", 0, "db slot")
	flag.IntVar(&opt.batch, "batch", 10000, "number of objects to check in one txn")
	flag.StringVar(&opt.namespace, "namespace", "default", "namespace")
	flag.Parse()

	addr := flag.Arg(0)
	if err := expire(opt, addr); err != nil {
		log.Fatal("expire failed", zap.Error(err))
	}
}
