package db

import (
	"context"
	"os"
	"testing"

	"github.com/distributedio/titan/conf"
	"github.com/distributedio/titan/db/store"
	"github.com/stretchr/testify/assert"
)

var mockDB *DB

func TestMain(m *testing.M) {
	s, err := store.MockOpen("mocktikv://")
	if err != nil {
		panic(err)
	}
	mockConf := conf.MockConf()
	// RedisStore expects store.Storage which is what MockOpen returns
	mockDB = &DB{
		Namespace: "mockdb-ns",
		ID:        1,
		kv:        &RedisStore{Storage: s, conf: &mockConf.TiKV},
	}

	os.Exit(m.Run())
}

func MockTest(t *testing.T, callFunc func(txn *Transaction)) {
	txn, err := mockDB.Begin()
	assert.NoError(t, err)
	callFunc(txn)
	err = txn.Commit(context.TODO())
	assert.NoError(t, err)
}
