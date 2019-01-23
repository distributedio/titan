package command

import (
	"bytes"
	"errors"
	"strconv"
	"strings"

	"github.com/meitu/titan/db"
	"github.com/meitu/titan/encoding/resp"
)

// HDel removes the specified fields from the hash stored at key
func HDel(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	hash, err := txn.Hash([]byte(ctx.Args[0]))
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	var fields [][]byte
	for _, field := range ctx.Args[1:] {
		fields = append(fields, []byte(field))
	}
	c, err := hash.HDel(fields)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
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
		return nil, errors.New("ERR " + err.Error())
	}

	status, err := hash.HSet(field, value)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
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
		return nil, errors.New("ERR " + err.Error())
	}

	status, err := hash.HSetNX(field, value)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(status)), nil
}

// HGet returns the value associated with field in the hash stored at key
func HGet(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	field := []byte(ctx.Args[1])

	hash, err := txn.Hash(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	val, err := hash.HGet(field)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
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
		return nil, errors.New("ERR " + err.Error())
	}
	fields, vals, err := hash.HGetAll()
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
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
		return nil, errors.New("ERR " + err.Error())
	}
	exist, err := hash.HExists(field)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
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
		return nil, ErrInteger
	}

	hash, err := txn.Hash(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	val, err := hash.HIncrBy(field, incr)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
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
		return nil, ErrFloat
	}

	hash, err := txn.Hash(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	val, err := hash.HIncrByFloat(field, incr)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return BulkString(ctx.Out, strconv.FormatFloat(val, 'f', -1, 64)), nil
}

// HKeys returns all field names in the hash stored at key
func HKeys(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	hash, err := txn.Hash(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	fields, _, err := hash.HGetAll()
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return BytesArray(ctx.Out, fields), nil
}

// HVals returns all values in the hash stored at key
func HVals(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	hash, err := txn.Hash(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	_, vals, err := hash.HGetAll()
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return BytesArray(ctx.Out, vals), nil

}

// HLen returns the number of fields contained in the hash stored at key
func HLen(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var size int64
	key := []byte(ctx.Args[0])
	hash, err := txn.Hash(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	size, err = hash.HLen()
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, size), nil
}

// HStrLen returns the string length of the value associated with field in the hash stored at key
func HStrLen(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	field := []byte(ctx.Args[1])
	hash, err := txn.Hash(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	val, err := hash.HGet(field)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
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
		return nil, errors.New("ERR " + err.Error())
	}

	vals, err := hash.HMGet(fields)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return BytesArray(ctx.Out, vals), nil
}

// HMSet sets the specified fields to their respective values in the hash stored at key
func HMSet(ctx *Context, txn *db.Transaction) (OnCommit, error) {

	var (
		key     = []byte(ctx.Args[0])
		kvs     = ctx.Args[1:]
		mapping = make(map[string][]byte)
		fields  [][]byte
		values  [][]byte
	)
	if len(kvs)%2 != 0 {
		return nil, errors.New("ERR wrong number of arguments for HMSET")
	}

	// When there are multiple groups of the same field/val,
	// take the last valid fields/val pair and save it in mapping
	for i := 0; i < len(kvs)-1; i += 2 {
		mapping[kvs[i]] = []byte(kvs[i+1])
	}

	// Iterate mapping to get the field and value after de-duplication
	for field, val := range mapping {
		fields = append(fields, []byte(field))
		values = append(values, val)
	}

	hash, err := txn.Hash(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	if err := hash.HMSet(fields, values); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return SimpleString(ctx.Out, "OK"), nil
}

// HMSlot specify the meta slot for hashes itself
func HMSlot(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	count, err := strconv.ParseInt(ctx.Args[1], 10, 64)
	if err != nil || count < 0 {
		return nil, ErrInteger
	}
	hash, err := txn.Hash(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	if err := hash.HMSlot(count); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return SimpleString(ctx.Out, "OK"), nil
}

//HScan incrementally iterate hash fields and associated values
func HScan(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var (
		key        []byte
		cursor     []byte
		lastCursor = []byte("0")
		count      = uint64(defaultScanCount)
		kvs        = [][]byte{}
		pattern    []byte
		prefix     []byte
		isAll      bool
		err        error
	)
	key = []byte(ctx.Args[0])
	if strings.Compare(ctx.Args[1], "0") != 0 {
		cursor = []byte(ctx.Args[1])
	}

	// define return result
	result := func() {
		resp.ReplyArray(ctx.Out, 2)
		resp.ReplyBulkString(ctx.Out, string(lastCursor))
		resp.ReplyArray(ctx.Out, len(kvs))
		for i := range kvs {
			resp.ReplyBulkString(ctx.Out, string(kvs[i]))
		}
	}
	hash, err := txn.Hash(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	//check if hash is not exits return result
	if !hash.Exists() {
		return result, nil
	}

	if len(ctx.Args)%2 != 0 {
		return nil, ErrSyntax
	}

	for i := 2; i < len(ctx.Args); i += 2 {
		arg := strings.ToLower(ctx.Args[i])
		next := ctx.Args[i+1]
		switch arg {
		case "count":
			if count, err = strconv.ParseUint(next, 10, 64); err != nil {
				return nil, ErrInteger
			}
			if count > ScanMaxCount {
				count = ScanMaxCount
			}
			if count == 0 {
				count = uint64(defaultScanCount)
			}
		case "match":
			pattern = []byte(next)
			isAll = (pattern[0] == '*' && len(pattern) == 1)
		}
	}

	if len(pattern) == 0 {
		isAll = true
	} else {
		prefix = globMatchPrefix(pattern)
		if cursor == nil && prefix != nil {
			cursor = prefix
		}
	}

	f := func(key, val []byte) bool {
		if count <= 0 {
			lastCursor = key
			return false
		}
		if prefix != nil && !bytes.HasPrefix(key, prefix) {
			return false
		}
		if isAll || globMatch(pattern, key, false) {
			kvs = append(kvs, key)
			kvs = append(kvs, val)
			count--
		}
		return true
	}

	if err := hash.HScan(cursor, f); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return result, nil

}
