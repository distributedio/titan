package command

import (
	"errors"
	"strconv"
	"strings"

	"gitlab.meitu.com/platform/thanos/db"
)

// LPush insert an entry to the head of the list
func LPush(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	args := ctx.Args

	// number of args should be checked by caller
	key := []byte(args[0])
	lst, err := txn.List(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !lst.Exist() {
		lst, err = txn.NewList(key, len(args)-1)
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
	}

	for _, val := range args[1:] {
		if err := lst.LPush([]byte(val)); err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
	}
	return Integer(ctx.Out, lst.Length()), nil
}

// Lpushx return ErrNotFound on key not found instead of create new list object
func LPushx(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	lst, err := txn.List(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !lst.Exist() {
		return Integer(ctx.Out, 0), nil
	}
	for _, val := range ctx.Args[1:] {
		if err := lst.LPush([]byte(val)); err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
	}
	return Integer(ctx.Out, lst.Length()), nil
}

// LPop removes and returns the first element of the list stored at key
func LPop(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	args := ctx.Args

	// number of args should be checked by caller
	key := []byte(args[0])

	lst, err := txn.List(key)

	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !lst.Exist() {
		return NullBulkString(ctx.Out), nil
	}

	val, err := lst.LPop()
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return BulkString(ctx.Out, string(val)), nil
}

// LRange returns the specified elements of the list stored at key
func LRange(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	args := ctx.Args
	key := []byte(args[0])

	start, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return nil, ErrInteger
	}
	stop, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return nil, ErrInteger
	}

	lst, err := txn.List(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !lst.Exist() {
		//TODO bug
		return BytesArray(ctx.Out, nil), nil
	}

	items, err := lst.Range(start, stop)
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
		return nil, ErrSyntax
	}

	pivot := []byte(ctx.Args[2])
	value := []byte(ctx.Args[3])

	lst, err := txn.List(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !lst.Exist() {
		return Integer(ctx.Out, -1), nil
	}

	err = lst.Insert(pivot, value, after)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, lst.Length()), nil
}

//LIndex
func LIndex(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	n, err := strconv.ParseInt(string(ctx.Args[1]), 10, 64)
	if err != nil {
		return nil, ErrInteger
	}

	lst, err := txn.List(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if !lst.Exist() {
		return NullBulkString(ctx.Out), nil
	}
	val, err := lst.Index(n)
	if err != nil {
		if err == db.ErrOutOfRange {
			return NullBulkString(ctx.Out), nil
		}
		return nil, errors.New("ERR " + err.Error())

	}
	return BulkString(ctx.Out, string(val)), nil
}

//LLen
func LLen(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	lst, err := txn.List(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !lst.Exist() {
		return Integer(ctx.Out, 0), nil
	}

	return Integer(ctx.Out, lst.Length()), nil
}

func LRem(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	n, err := strconv.ParseInt(string(ctx.Args[1]), 10, 64)
	if err != nil {
		return nil, ErrInteger
	}
	lst, err := txn.List(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !lst.Exist() {
		return Integer(ctx.Out, 0), nil
	}
	count, err := lst.LRem([]byte(ctx.Args[1]), n)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(count)), nil
}

func LTrim(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	start, err := strconv.ParseInt(ctx.Args[1], 10, 64)
	if err != nil {
		return nil, ErrInteger
	}
	stop, err := strconv.ParseInt(ctx.Args[2], 10, 64)
	if err != nil {
		return nil, ErrInteger
	}
	lst, err := txn.List(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !lst.Exist() {
		return BulkString(ctx.Out, "OK"), nil
	}
	if err = lst.LTrim(start, stop); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	return BulkString(ctx.Out, "OK"), nil

}

func LSet(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	lst, err := txn.List(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !lst.Exist() {
		return nil, ErrNoSuchKey
	}
	n, err := strconv.ParseInt(ctx.Args[1], 10, 64)
	if err != nil {
		return nil, ErrInteger
	}
	if err := lst.Set(n, []byte(ctx.Args[2])); err != nil {
		if err == db.ErrOutOfRange {
			return nil, ErrIndex
		}
		return nil, errors.New("ERR " + err.Error())
	}
	return BulkString(ctx.Out, "OK"), nil
}

func RPop(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	lst, err := txn.List(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !lst.Exist() {
		return NullBulkString(ctx.Out), nil
	}
	val, err := lst.RPop()
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return BulkString(ctx.Out, string(val)), nil
}

// RPopLPush return the value rpop from src and push into dest
func RPopLPush(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	listsrc, err := txn.List([]byte(ctx.Args[0]))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !listsrc.Exist() {
		return NullBulkString(ctx.Out), nil
	}
	val, err := listsrc.RPop()
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	// create dst list on not exist
	listdst, err := txn.List([]byte(ctx.Args[1]))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, ErrSyntax
	}

	if !listdst.Exist() {
		listdst, err = txn.NewList([]byte(ctx.Args[1]), len(ctx.Args)-1)
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
	}
	if err = listdst.LPush(val); err != nil {
		return nil, ErrTypeMismatch
	}
	return BulkString(ctx.Out, string(val)), nil
}

// Rpush push values into list from right
func RPush(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	lst, err := txn.List(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !lst.Exist() {
		lst, err = txn.NewList([]byte(ctx.Args[1]), len(ctx.Args)-1)
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
	}
	for _, val := range ctx.Args[1:] {
		if err := lst.LPush([]byte(val)); err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
	}
	return Integer(ctx.Out, lst.Length()), nil
}

// Rpushx return ErrNotFound on key not found instead of create new list object
func RPushx(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	lst, err := txn.List(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if !lst.Exist() {
		return Integer(ctx.Out, 0), nil
	}
	for _, val := range ctx.Args[1:] {
		if err := lst.RPush([]byte(val)); err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
	}
	return Integer(ctx.Out, lst.Length()), nil
}
