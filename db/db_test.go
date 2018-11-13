package db

import (
	"os"
	"testing"

	"github.com/pingcap/tidb/store/mockstore"
)

var mockDB *DB

func TestMain(m *testing.M) {
	store, err := mockstore.NewMockTikvStore()
	if err != nil {
		panic(err)
	}
	mockDB = &DB{
		Namespace: "mockdb-ns",
		ID:        1,
		kv:        &RedisStore{store},
	}

	os.Exit(m.Run())
}
