package main

import (
	"context"
	"flag"
	"time"

	"github.com/distributedio/titan/db"
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/store/tikv"
	"go.uber.org/zap"
)

var log, _ = zap.NewDevelopment()

type option struct {
	db         int
	namespace  string
	batch      int
	dry        bool
	toleration time.Duration
}

func deletePrefix(txn kv.Transaction, prefix kv.Key) error {
	end := prefix.PrefixNext()
	iter, err := txn.Iter(prefix, end)
	if err != nil {
		return err
	}
	for iter.Valid() && iter.Key().HasPrefix(prefix) {
		log.Debug("delete", zap.String("key", string(iter.Key())))
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
	if err := txn.Delete(mkey); err != nil {
		return err
	}
	switch obj.Type {
	case db.ObjectString:
		return nil
	case db.ObjectZSet:
		dkey := db.DataKey(database, obj.ID)
		skey := db.ZSetScorePrefix(database, obj.ID)
		deletePrefix(txn, dkey) // ignore its error for simplicity
		return deletePrefix(txn, skey)
	default:
		dkey := db.DataKey(database, obj.ID)
		return deletePrefix(txn, dkey)
	}
}

func doExpire(s kv.Storage, database *db.DB, prefix kv.Key,
	start kv.Key, opt *option) (kv.Key, error) {
	txn, err := s.Begin()
	if err != nil {
		return nil, err
	}
	iter, err := txn.Iter(start, nil)
	if err != nil {
		return nil, err
	}
	// tolerate certain times
	now := db.Now() - int64(opt.toleration)
	// scan the whole database
	var end kv.Key
	limit := opt.batch
	for iter.Valid() && iter.Key().HasPrefix(prefix) && limit != 0 {
		obj, err := db.DecodeObject(iter.Value())
		if err != nil {
			return nil, err
		}
		if obj.ExpireAt > 0 && obj.ExpireAt < now {
			log.Debug("expire", zap.String("key", string(iter.Key())))
			if err := cleanUp(txn, iter.Key(), database, obj); err != nil {
				return nil, err
			}
		}
		limit--
		end = iter.Key()
		iter.Next()
	}
	if opt.dry {
		log.Debug("rollback by dry option")
		if err := txn.Rollback(); err != nil {
			return nil, err
		}
	} else {
		if err := txn.Commit(context.Background()); err != nil {
			return nil, err
		}
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
	prefix := start
	for {
		next, err := doExpire(store, database, prefix, start, opt)
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
	flag.BoolVar(&opt.dry, "dry", false, "do not affect the database")
	flag.DurationVar(&opt.toleration, "toleration", 300*time.Second, "tolerate certain time to expire")
	flag.Parse()

	addr := flag.Arg(0)
	log.Debug("options", zap.Reflect("dry", opt.dry),
		zap.Int("batch", opt.batch), zap.String("addr", addr),
		zap.Duration("toleration", opt.toleration))
	if err := expire(opt, addr); err != nil {
		log.Fatal("expire failed", zap.Error(err))
	}
}
