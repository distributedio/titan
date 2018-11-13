package command

import (
	"errors"
	"strconv"

	"gitlab.meitu.com/platform/titan/db"
)

// HDel removes the specified fields from the hash stored at key
func HDel(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	hash, err := txn.Hash([]byte(ctx.Args[0]))
	if err != nil {
		return nil, err
	}

	var fields [][]byte
	for _, field := range ctx.Args[1:] {
		fields = append(fields, []byte(field))
	}
	c, err := hash.HDel(fields)
	if err != nil {
		return nil, err
	}
	return Integer(ctx.Out, c), nil
}

// HSet sets field in the hash stored at key to value
func HSet(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	field := []byte(ctx.Args[1])
	value := []byte(ctx.Args[2])

	hash, err := txn.Hash(key)
	if err != nil {
		return nil, err
	}

	status, err := hash.HSet(field, value)
	if err != nil {
		return nil, err
	}
	return Integer(ctx.Out, int64(status)), nil
}

// HSetNX sets field in the hash stored at key to value, only if field does not yet exist
func HSetNX(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	field := []byte(ctx.Args[1])
	value := []byte(ctx.Args[2])

	hash, err := txn.Hash(key)
	if err != nil {
		return nil, err
	}

	status, err := hash.HSetNX(field, value)
	if err != nil {
		return nil, err
	}
	return Integer(ctx.Out, int64(status)), nil
}

// HGet returns the value associated with field in the hash stored at key
func HGet(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	field := []byte(ctx.Args[1])

	hash, err := txn.Hash(key)
	if err != nil {
		return nil, err
	}
	val, err := hash.HGet(field)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return NullBulkString(ctx.Out), nil
	}
	return BulkString(ctx.Out, string(val)), nil
}

// HGetAll returns all fields and values of the hash stored at key
func HGetAll(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	hash, err := txn.Hash(key)
	if err != nil {
		return nil, err
	}
	fields, vals, err := hash.HGetAll()
	if err != nil {
		return nil, err
	}

	var results [][]byte
	for i := range fields {
		results = append(results, fields[i])
		results = append(results, vals[i])
	}

	return BytesArray(ctx.Out, results), nil
}

// HExists returns if field is an existing field in the hash stored at key
func HExists(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	field := []byte(ctx.Args[1])
	hash, err := txn.Hash(key)
	if err != nil {
		return nil, err
	}
	exist, err := hash.HExists(field)
	if err != nil {
		return nil, err
	}
	if exist {
		return Integer(ctx.Out, 1), nil
	}
	return Integer(ctx.Out, 0), nil
}

// HIncrBy increments the number stored at field in the hash stored at key by increment
func HIncrBy(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	field := []byte(ctx.Args[1])
	incr, err := strconv.ParseInt(ctx.Args[2], 10, 64)
	if err != nil {
		return nil, err
	}

	hash, err := txn.Hash(key)
	if err != nil {
		return nil, err
	}

	val, err := hash.HIncrBy(field, incr)
	if err != nil {
		return nil, err
	}
	return Integer(ctx.Out, val), err
}

// HIncrByFloat increment the specified field of a hash stored at key,
// and representing a floating point number, by the specified increment
func HIncrByFloat(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	field := []byte(ctx.Args[1])
	incr, err := strconv.ParseFloat(ctx.Args[2], 64)
	if err != nil {
		return nil, err
	}

	hash, err := txn.Hash(key)
	if err != nil {
		return nil, err
	}

	val, err := hash.HIncrByFloat(field, incr)
	if err != nil {
		return nil, err
	}
	return BulkString(ctx.Out, strconv.FormatFloat(val, 'f', -1, 64)), nil
}

// HKeys returns all field names in the hash stored at key
func HKeys(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	hash, err := txn.Hash(key)
	if err != nil {
		return nil, err
	}
	fields, _, err := hash.HGetAll()
	if err != nil {
		return nil, err
	}
	return BytesArray(ctx.Out, fields), nil
}

// HVals returns all values in the hash stored at key
func HVals(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	hash, err := txn.Hash(key)
	if err != nil {
		return nil, err
	}
	_, vals, err := hash.HGetAll()
	if err != nil {
		return nil, err
	}
	return BytesArray(ctx.Out, vals), nil

}

// HLen returns the number of fields contained in the hash stored at key
func HLen(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	hash, err := txn.Hash(key)
	if err != nil {
		return nil, err
	}
	return Integer(ctx.Out, hash.HLen()), nil
}

// HStrLen returns the string length of the value associated with field in the hash stored at key
func HStrLen(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	field := []byte(ctx.Args[1])
	hash, err := txn.Hash(key)
	if err != nil {
		return nil, err
	}
	val, err := hash.HGet(field)
	if err != nil {
		return nil, err
	}
	return Integer(ctx.Out, int64(len(val))), nil
}

// HMGet returns the values associated with the specified fields in the hash stored at key
func HMGet(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	var fields [][]byte
	for _, field := range ctx.Args[1:] {
		fields = append(fields, []byte(field))
	}

	hash, err := txn.Hash(key)
	if err != nil {
		return nil, err
	}

	vals, err := hash.HMGet(fields)
	if err != nil {
		return nil, err
	}
	return BytesArray(ctx.Out, vals), nil
}

// HMSet sets the specified fields to their respective values in the hash stored at key
func HMSet(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	kvs := ctx.Args[1:]
	if len(kvs)%2 != 0 {
		return nil, errors.New("ERR wrong number of arguments for HMSET")
	}

	count := len(kvs) / 2
	fields := make([][]byte, count)
	values := make([][]byte, count)
	j := 0
	for i := 0; i < len(kvs)-1; i += 2 {
		fields[j] = []byte(kvs[i])
		values[j] = []byte(kvs[i+1])
		j++
	}

	hash, err := txn.Hash(key)
	if err != nil {
		return nil, err
	}

	if err := hash.HMSet(fields, values); err != nil {
		return nil, err
	}
	return SimpleString(ctx.Out, "OK"), nil
}
