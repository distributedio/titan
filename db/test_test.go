package db

import (
	"github.com/meitu/titan/conf"
	"github.com/pingcap/tidb/store/mockstore"
	"go.uber.org/zap"
)

func MockDB() *DB {
	zap.ReplaceGlobals(zap.NewNop())
	store, err := mockstore.NewMockTikvStore()
	if err != nil {
		panic(err)
	}
	redis := &RedisStore{Storage: store, conf: &conf.Tikv{}}
	return &DB{Namespace: "ns", ID: DBID(1), kv: redis}
}
