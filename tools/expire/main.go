package main

import (
	"bytes"
	"context"
	"flag"
	"time"

	"github.com/distributedio/titan/db"
	"github.com/distributedio/titan/db/store"
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

func prefixNext(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}
	next := make([]byte, len(prefix))
	copy(next, prefix)
	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] != 0 {
			return next
		}
	}
	return nil
}

func deletePrefix(txn store.Transaction, prefix []byte) error {
	end := prefixNext(prefix)
	iter, err := txn.Iter(prefix, end)
	if err != nil {
		return err
	}
	defer iter.Close()
	for iter.Valid() && bytes.HasPrefix(iter.Key(), prefix) {
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

func cleanUp(txn store.Transaction, mkey []byte, database *db.DB,
	obj *db.Object) error {
	if err := txn.Delete(mkey); err != nil {
		return err
	}
	switch obj.Type {
	case db.ObjectString:
		return nil
	default:
		dkey := db.DataKey(database, obj.ID)
		return deletePrefix(txn, dkey)
	}
}

func doExpire(s store.Storage, database *db.DB, prefix []byte,
	start []byte, opt *option) ([]byte, error) {
	txn, err := s.Begin()
	if err != nil {
		return nil, err
	}
	iter, err := txn.Iter(start, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	// tolerate certain times
	now := db.Now() - int64(opt.toleration)
	// scan the whole database
	var end []byte
	limit := opt.batch
	for iter.Valid() && bytes.HasPrefix(iter.Key(), prefix) && limit != 0 {
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
		if err := iter.Next(); err != nil {
			return nil, err
		}
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
	return append(end, 0), nil
}

func runExpire(s store.Storage, opt *option) error {
	log.Info("start to expire", zap.Int("db", opt.db),
		zap.String("namespace", opt.namespace))

	database := &db.DB{Namespace: opt.namespace, ID: db.DBID(opt.db)}

	start := db.MetaKey(database, nil)
	prefix := start
	for {
		next, err := doExpire(s, database, prefix, start, opt)
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

func expire(opt *option, addr string) error {
	s, err := store.Open(addr)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()

	return runExpire(s, opt)
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