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
		return nil, errors.New("ERR syntax error")
	}

	pivot := []byte(ctx.Args[2])
	value := []byte(ctx.Args[3])

	lst, err := txn.List(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, ErrSyntax
	}

	if !lst.Exist() {
		return Integer(ctx.Out, 0), nil
	}

	err = lst.Insert(pivot, value, after)
	if err != nil {
		if err == db.ErrPrecision {
			//TODO check
			return nil, err
		}
		return nil, ErrSyntax
	}
	return Integer(ctx.Out, lst.Length()), nil
}
