package command

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"gitlab.meitu.com/platform/thanos/db"
)

// Get the value of key
func Get(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]

	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR " + err.Error())
	}
	val, err := str.Get()
	if err == db.ErrKeyNotFound {
		return NullBulkString(ctx.Out), nil
	}
	return BulkString(ctx.Out, string(val)), nil
}

// Set key to hold the string value
func Set(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	val := ctx.Args[1]

	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR " + err.Error())
	}
	ttl := time.Duration(-1)
	options := ctx.Args[1:]
	for i, opt := range options {
		switch strings.ToLower(opt) {
		case "ex":
			if i+1 >= len(options) {
				return nil, errors.New("ERR syntax error")
			}
			val := options[i+1]
			sec, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, err
			}
			ttl = time.Duration(sec) * time.Second
		case "px":
			if i+1 >= len(options) {
				return nil, errors.New("ERR syntax error")
			}
			val := options[i+1]
			msec, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, err
			}
			ttl = time.Duration(msec) * time.Millisecond
		case "nx":
			if _, err := str.Get(); err != db.ErrKeyNotFound {
				return NullBulkString(ctx.Out), nil
			}
		case "xx":
			if _, err := str.Get(); err == db.ErrKeyNotFound {
				return NullBulkString(ctx.Out), nil
			}
		}
	}

	var at []int64
	if ttl > 0 {
		at = append(at, time.Now().Add(ttl).UnixNano())
	}
	if err := str.Set([]byte(val), at...); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	return SimpleString(ctx.Out, "OK"), nil
	// TODO handle other options
}

// MGet returns the values of all specified keys
// TODO use BatchGetRequest to gain performance
func MGet(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	count := len(ctx.Args)

	keys := make([][]byte, count)
	for i := range ctx.Args {
		keys[i] = []byte(ctx.Args[i])
	}

	values, err := db.BatchGetValues(txn, keys)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return BytesArray(ctx.Out, values), nil
}

// MSet sets the given keys to their respective values
func MSet(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	argc := len(ctx.Args)
	if argc%2 != 0 {
		return nil, errors.New("ERR wrong number of arguments for MSET")
	}
	for i := 0; i < argc-1; i += 2 {
		if _, err := Set(ctx, txn); err != nil {
			return nil, err
		}
		ctx.Args = ctx.Args[2:]
	}
	return SimpleString(ctx.Out, "OK"), nil
}

// Strlen returns the length of the string value stored at key
func Strlen(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR " + err.Error())
	}
	v, err := str.Len()
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(v)), nil
}
