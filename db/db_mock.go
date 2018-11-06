package db

import (
	"context"

	"github.com/pingcap/tidb/store/mockstore"
)

var (
	db = MockDB()
)

func MockDB() *DB {
	store, err := mockstore.NewMockTikvStore()
	if err != nil {
		panic(err)
	}
	return &DB{
		kv:  store,
		ns:  []byte("ns"),
		id:  1,
		Ctx: context.TODO(),
	}
}

func GlobalDB() *DB {
	return db
}
