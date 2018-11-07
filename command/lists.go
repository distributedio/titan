package command

import (
	"errors"
	"strconv"
	"strings"

	"gitlab.meitu.com/platform/thanos/db"
	"gitlab.meitu.com/platform/thanos/resp"
)

// LPush insert an entry to the head of the list
func LPush(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	args := ctx.Args

	// number of args should be checked by caller
	key := []byte(args[0])

	lst, err := txn.List(key)

	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR " + err.Error())
	}

	for _, val := range args[1:] {
		if err := lst.LPush([]byte(val)); err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
	}
	return SimpleString(ctx.Out, "OK"), nil
}

// LPop removes and returns the first element of the list stored at key
func LPop(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	args := ctx.Args

	// number of args should be checked by caller
	key := []byte(args[0])

	lst, err := txn.List(key)

	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR " + err.Error())
	}

	val, err := lst.LPop()
	if err != nil {
		if err == db.ErrKeyNotFound {
			return NullBulkString(ctx.Out), nil
		}
		return nil, errors.New("ERR " + err.Error())
	}
	return BulkString(ctx.Out, string(val)), nil
}

// LRange returns the specified elements of the list stored at key
func LRange(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	args := ctx.Args
	key := []byte(args[0])

	start, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}
	stop, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, errors.New("ERR value is not an integer or out of range")
	}

	lst, err := txn.List(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if lst == nil {
		resp.ReplyArray(ctx.Out, 0)
		return nil, nil
	}

	items, err := lst.Range(int64(start), int64(stop))
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return BytesArray(ctx.Out, items), nil
}

// LInsert inserts value in the list stored at key either before or after the reference value pivot
func LInsert(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	after := false
	switch strings.ToLower(ctx.Args[1]) {
	case "before":
		after = false
	case "after":
		after = true
	default:
		return nil, errors.New("ERR syntax error")
	}

	pivot := []byte(ctx.Args[2])
	value := []byte(ctx.Args[3])

	lst, err := txn.List(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR syntax error")
	}

	err = lst.Insert(pivot, value, after)
	if err != nil {
		if err == db.ErrPrecision {
			return nil, err
		}
		return nil, errors.New("ERR syntax error")
	}
	length := lst.Length()
	return Integer(ctx.Out, int64(length)), nil
}
