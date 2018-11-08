package db

/*
func EqualGet(t *testing.T, db *DB, key []byte, value []byte) {
	txn, _ := db.Begin()
	obj, err := GetString(txn, key)
	assert.NoError(t, err)
	v, err := obj.Get()
	assert.Equal(t, value, v)
	assert.NoError(t, err)
	txn.Commit(context.TODO())
}

func EqualNotFound(t *testing.T, db *DB, key []byte) {
	// 不存在的key ，判断错误类型
	txn, _ := db.Begin()
	obj, err := GetString(txn, key)
	assert.NoError(t, err)
	v, err := obj.Get()
	assert.Nil(t, v)
	assert.Equal(t, false, obj.Exist())
	txn.Commit(context.TODO())
}

func TestStringSet(t *testing.T) {
	value := []byte("value")
	key := []byte("key")

	db := MockDB()
	txn, _ := db.Begin()
	obj, err := GetString(txn, key)
	assert.NoError(t, err)
	err = obj.Set(value)
	assert.NoError(t, err)
	txn.Commit(context.TODO())
	EqualGet(t, db, key, value)
}

//过期校验
func TestStringSetPx(t *testing.T) {
	value := []byte("value-px")
	key := []byte("key-px")

	db := MockDB()
	txn, _ := db.Begin()
	obj, err := GetString(txn, key)
	assert.NoError(t, err)

	err = obj.Set(value, int64(time.Millisecond*100))
	assert.NoError(t, err)
	txn.Commit(context.TODO())
	EqualGet(t, db, key, value)
	time.Sleep(time.Second)
	EqualNotFound(t, db, key)
}

//过期校验
func TestStringSetExpire(t *testing.T) {
	value := []byte("value-ex")
	key := []byte("key-ex")

	db := MockDB()
	txn, _ := db.Begin()
	obj, err := GetString(txn, key)
	assert.NoError(t, err)

	err = obj.Set(value, int64(time.Second))
	assert.NoError(t, err)
	txn.Commit(context.TODO())
	EqualGet(t, db, key, value)
	time.Sleep(time.Second)
	EqualNotFound(t, db, key)
}

/*
func TestStrings(t *testing.T) {
	keys := make([][]byte, 5)
	keys[0] = []byte("key")
	keys[1] = []byte("k2")
	keys[2] = []byte("key-ex")
	keys[3] = []byte("k3")
	keys[4] = []byte("key-px")
	value := []byte("value")

	db.Begin()
	objs, err := db.Strings(keys)

	assert.Len(t, keys, len(objs))
	assert.NoError(t, err)

	err = objs[3].Set(value, 0, 1)
	assert.NoError(t, err)

	err = objs[0].Set(value, 0, 1)
	assert.NoError(t, err)
	db.Commit()

	EqualGet(t, keys[0], value)
	EqualNotFound(t, keys[1])
	EqualNotFound(t, keys[2])
	EqualGet(t, keys[3], value)
	EqualNotFound(t, keys[4])
}
*/
