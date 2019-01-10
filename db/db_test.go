package db

import (
	"os"
	"testing"

	"github.com/meitu/titan/conf"
	"github.com/pingcap/tidb/store/mockstore"
)

var mockDB *DB

func TestMain(m *testing.M) {
	store, err := mockstore.NewMockTikvStore()
	if err != nil {
		panic(err)
	}
	conf := &conf.Tikv{}
	mockDB = &DB{
		Namespace: "mockdb-ns",
		ID:        1,
		kv:        &RedisStore{Storage: store, conf: conf},
		conf:      &conf.DB,
	}

	os.Exit(m.Run())
}
