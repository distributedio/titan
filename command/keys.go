package command

import (
	"errors"
	"strconv"
	"time"

	"gitlab.meitu.com/platform/thanos/db"
)

// Type returns the string representation of the type of the value stored at key
func Type(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	obj, err := txn.Object(key)
	if err != nil {
		if err == db.ErrKeyNotFound {
			return SimpleString(ctx.Out, "none"), nil
		}
		return nil, errors.New("ERR " + err.Error())
	}

	return SimpleString(ctx.Out, obj.Type.String()), nil
}

// Exists returns if key exists
func Exists(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	n := int64(0)
	for _, key := range ctx.Args {
		_, err := txn.Object([]byte(key))
		if err != nil {
			if err == db.ErrKeyNotFound {
				continue
			}
			return nil, errors.New("ERR " + err.Error())
		}
		n++
	}
	return Integer(ctx.Out, n), nil
}

// Keys returns all keys matching pattern.
func Keys(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	pattern := ctx.Args[0]
	kv := txn.Kv()
	var outputs [][]byte
	kv.Keys(func(key []byte) bool {
		if globMatch(string(key), pattern) {
			outputs = append(outputs, key)
		}
		return true
	})
	return BytesArray(ctx.Out, outputs), nil
}

// Delete removes the specified keys. A key is ignored if it does not exist
func Delete(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()

	keys := make([][]byte, len(ctx.Args))
	for i := range ctx.Args {
		keys[i] = []byte(ctx.Args[i])
	}

	n, err := kv.Delete(keys)
	if err != nil {
		return nil, err
	}

	return Integer(ctx.Out, int64(n)), nil
}

// Expire set a timeout on key
func Expire(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	key := []byte(ctx.Args[0])
	sec, err := strconv.ParseInt(ctx.Args[1], 10, 64)
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}

	at := time.Now().Add(time.Duration(sec) * time.Second).UnixNano()
	if err := kv.ExpireAt(key, at); err != nil {
		if err == db.ErrKeyNotFound {
			return Integer(ctx.Out, 0), nil
		}
		return nil, err
	}
	return Integer(ctx.Out, 1), nil
}

// ExpireAt set an absolute timestamp to expire on key
func ExpireAt(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	key := []byte(ctx.Args[0])
	ts, err := strconv.ParseInt(ctx.Args[1], 10, 64)
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}
	if err := kv.ExpireAt(key, ts); err != nil {
		if err == db.ErrKeyNotFound {
			return Integer(ctx.Out, 0), nil
		}
		return nil, err
	}
	return Integer(ctx.Out, 1), nil
}
