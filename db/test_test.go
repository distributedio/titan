package db

import (
	"github.com/distributedio/titan/conf"
	"github.com/distributedio/titan/db/store"
	"go.uber.org/zap"
)

func MockDB() *DB {
	zap.ReplaceGlobals(zap.NewNop())
	s, err := store.MockOpen("mocktikv://")
	if err != nil {
		panic(err)
	}
	mockConf := conf.MockConf()
	redis := &RedisStore{Storage: s, conf: &mockConf.TiKV}
	return &DB{Namespace: "ns", ID: DBID(1), kv: redis}
}
