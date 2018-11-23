package command

import (
	"errors"
	"strconv"
	"strings"

	"github.com/meitu/titan/db"
)

var (
	// ListZipThreshold indicates to create a ziplist when it is exceeded to push elements
	ListZipThreshold = 100
)

// LPush inserts an entry to the head of the list
func LPush(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	args := ctx.Args

	// Create a ziplist if lpush with too much items
	var opts []db.ListOption
	if len(args[1:]) > ListZipThreshold {
		opts = append(opts, db.UseZip())
	}

	// Number of args should be checked by caller
	key := []byte(args[0])
	lst, err := txn.List(key, opts...)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	for _, val := range args[1:] {
		if err := lst.LPush([]byte(val)); err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
	}
	return Integer(ctx.Out, lst.Length()), nil
}

// LPushx prepend a value to a list, only if the list exists
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

// LRange get a range of elements from a list
func LRange(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	args := ctx.Args
	key := []byte(args[0])

	start, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, ErrInteger
	}
	stop, err := strconv.ParseInt(args[2], 10, 64)
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
		return BytesArray(ctx.Out, nil), nil
	}

	items, err := lst.Range(start, stop)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	if len(items) == 0 {
		return BytesArray(ctx.Out, nil), nil
	}
	return BytesArray(ctx.Out, items), nil
}

// LInsert insert an element before or after another element in a list
func LInsert(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	before := false
	switch strings.ToLower(ctx.Args[1]) {
	case "before":
		before = true
	case "after":
		before = false
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

	err = lst.Insert(pivot, value, before)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, lst.Length()), nil
}

//LIndex get an element from a list by its index
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

//LLen get the length of a list
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

//LRem remove elements from a list
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
	count, err := lst.LRem([]byte(ctx.Args[2]), n)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(count)), nil
}

//LTrim trim a list to the specified range
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

//LSet set the value of an element in a list by its index
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
	return SimpleString(ctx.Out, "OK"), nil
}

//RPop remove and get the last element in a list
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

// RPopLPush remove the last element in a list, prepend it to another list and return it
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
	if err = listdst.LPush(val); err != nil {
		return nil, ErrTypeMismatch
	}
	return BulkString(ctx.Out, string(val)), nil
}

// RPush append one or multiple values to a list
func RPush(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	args := ctx.Args
	key := []byte(args[0])

	// Create a ziplist if lpush with too much items
	var opts []db.ListOption
	if len(args[1:]) > ListZipThreshold {
		opts = append(opts, db.UseZip())
	}

	lst, err := txn.List(key, opts...)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	for _, val := range args[1:] {
		if err := lst.RPush([]byte(val)); err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
	}
	return Integer(ctx.Out, lst.Length()), nil
}

// RPushx append a value to a list, only if the list exists
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
