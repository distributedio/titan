package command

import (
	"errors"
	"strconv"
	"time"

	"github.com/meitu/titan/db"
)

var (
	//MaxRangeInteger max index in setrange command
	MaxRangeInteger = 2<<29 - 1
)

// Get the value of key
func Get(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	val, err := str.Get()
	if err != nil {
		if err == db.ErrKeyNotFound {
			return NullBulkString(ctx.Out), nil
		}
		return nil, errors.New("ERR " + err.Error())
	}
	return BulkString(ctx.Out, string(val)), nil
}

// Set key to hold the string value
func Set(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	value := []byte(ctx.Args[1])
	args := ctx.Args

	var next bool
	var flag int      // flag int // 0 -- null 1---nx  2---xx
	var unit int64    // time ms
	var expire string //expire = expire *unit
	for i := 2; i < len(args); i++ {
		if i+1 < len(args) {
			next = true
		}
		if (args[i][0] == 'n' || args[i][0] == 'N') &&
			(args[i][1] == 'x' || args[i][1] == 'X') &&
			len(args[i]) == 2 {
			flag = 1
		} else if (args[i][0] == 'x' || args[i][0] == 'X') &&
			(args[i][1] == 'x' || args[i][1] == 'X') &&
			len(args[i]) == 2 {
			flag = 2
		} else if (args[i][0] == 'p' || args[i][0] == 'P') &&
			(args[i][1] == 'x' || args[i][1] == 'X') &&
			len(args[i]) == 2 && next {
			expire = args[i+1]
			i = i + 1
			unit = int64(time.Millisecond)
		} else if (args[i][0] == 'e' || args[i][0] == 'E') &&
			(args[i][1] == 'x' || args[i][1] == 'X') &&
			len(args[i]) == 2 && next {
			expire = args[i+1]
			i = i + 1
			unit = int64(time.Second)
		} else {
			return nil, ErrSyntax
		}
	}

	//ex|px
	if expire != "" {
		ui, err := strconv.ParseInt(expire, 10, 64)
		if err != nil {
			return nil, ErrInteger
		}
		if ui == 0 {
			return nil, ErrExpire
		}
		unit = ui * unit
	}

	obj, err := txn.Object(key)
	if err != nil && err != db.ErrKeyNotFound {
		return nil, errors.New("ERR " + err.Error())
	}

	//xx
	if flag == 2 {
		if err == db.ErrKeyNotFound {
			return NullBulkString(ctx.Out), nil
		}
	}
	//nx
	if flag == 1 {
		if err != db.ErrKeyNotFound {
			return NullBulkString(ctx.Out), nil
		}
	}

	if err != db.ErrKeyNotFound {
		txn.Destory(obj, key)
	}

	s := db.NewString(txn, key)
	if err := s.Set(value, unit); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return SimpleString(ctx.Out, OK), nil
}

// MGet returns the values of all specified key
func MGet(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	count := len(ctx.Args)
	values := make([][]byte, count)

	keys := make([][]byte, count)
	for i := range ctx.Args {
		keys[i] = []byte(ctx.Args[i])
	}

	strs, err := txn.Strings(keys)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	for i, str := range strs {
		if str == nil || !str.Exist() {
			values[i] = nil
			continue
		}
		values[i], _ = str.Get()
	}
	return BytesArray(ctx.Out, values), nil
}

// MSet sets the given keys to their respective values
func MSet(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	argc := len(ctx.Args)
	args := ctx.Args
	if argc%2 != 0 {
		return nil, ErrMSet
	}
	for i := 2; i <= argc; i += 2 {
		ctx.Args = args[i-2 : i]
		if _, err := Set(ctx, txn); err != nil {
			return nil, err
		}
	}
	return SimpleString(ctx.Out, OK), nil
}

// MSetNx et multiple keys to multiple values,only if none of the keys exist
//TODO bug
func MSetNx(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	argc := len(ctx.Args)
	args := ctx.Args
	if argc%2 != 0 {
		return nil, ErrMSet
	}
	for i := 2; i <= argc; i += 2 {
		if str, _ := txn.String([]byte(ctx.Args[0])); !str.Exist() {
			ctx.Args = append(args[i-2:i], "nx")
			if _, err := Set(ctx, txn); err != nil {
				return nil, errors.New("ERR " + err.Error())
			}
			ctx.Args = ctx.Args[2:]
			continue
		}
		return Integer(ctx.Out, int64(0)), nil
	}
	return Integer(ctx.Out, int64(1)), nil
}

// Strlen returns the length of the string value stored at key
func Strlen(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !str.Exist() {
		return Integer(ctx.Out, int64(0)), nil
	}

	v, _ := str.Len()

	return Integer(ctx.Out, int64(v)), nil
}

// Append a value to a key
func Append(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	value := []byte(ctx.Args[1])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	llen, err := str.Append(value)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(llen)), nil
}

// GetSet sets the string value of a key and return its old value
func GetSet(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	v := []byte(ctx.Args[1])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if !str.Exist() {
		return NullBulkString(ctx.Out), nil
	}

	value, err := str.GetSet(v)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return BulkString(ctx.Out, string(value)), nil
}

// GetRange increments the integer value of a keys by the given amount
func GetRange(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	start, err := strconv.Atoi(string(ctx.Args[1]))
	if err != nil {
		return nil, ErrInteger
	}
	end, err := strconv.Atoi(string(ctx.Args[2]))
	if err != nil {
		return nil, ErrInteger
	}

	if !str.Exist() {
		return NullBulkString(ctx.Out), nil
	}

	value := str.GetRange(start, end)
	if len(value) == 0 {
		return NullBulkString(ctx.Out), nil
	}

	return BulkString(ctx.Out, string(value)), nil
}

// SetNx sets the value of a key ,only if the key does not exist
func SetNx(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	//get the key
	key := []byte(ctx.Args[0])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return Integer(ctx.Out, int64(0)), nil
		}
		return nil, errors.New("ERR " + err.Error())
	}
	//key 不存在时，为 key 设置指定的值。设置成功，返回 1 。 设置失败，返回 0 。
	if str.Exist() {
		return Integer(ctx.Out, int64(0)), nil
	}

	if err := str.Set([]byte(ctx.Args[1])); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(1)), nil
}

// SetEx sets the value and expiration of a key KEY_NAME TIMEOUT VALUE
func SetEx(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	//get the key
	key := []byte(ctx.Args[0])
	obj, err := txn.Object(key)
	if err != nil && err != db.ErrKeyNotFound {
		return nil, errors.New("ERR " + err.Error())
	}
	if err != db.ErrKeyNotFound {
		txn.Destory(obj, key)
	}

	s := db.NewString(txn, key)
	ui, err := strconv.ParseInt(string(ctx.Args[1]), 10, 64)
	if err != nil {
		return nil, ErrInteger
	}
	unit := ui * int64(time.Second)
	if err := s.Set([]byte(ctx.Args[2]), unit); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	return SimpleString(ctx.Out, OK), nil
}

// PSetEx set the value and expiration in milliseconds of a key
func PSetEx(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	//get the key
	key := []byte(ctx.Args[0])
	obj, err := txn.Object(key)
	if err != nil && err != db.ErrKeyNotFound {
		return nil, errors.New("ERR " + err.Error())
	}

	if err != db.ErrKeyNotFound {
		txn.Destory(obj, key)
	}

	s := db.NewString(txn, key)
	ui, err := strconv.ParseUint(string(ctx.Args[1]), 10, 64)
	if err != nil {
		return nil, ErrInteger
	}
	unit := ui * uint64(time.Millisecond)
	if err := s.Set([]byte(ctx.Args[2]), int64(unit)); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	return SimpleString(ctx.Out, OK), nil
}

//SetRange Overwrites part of the string stored at key, starting at the specified offset, for the entire length of value.
func SetRange(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	offset, err := strconv.Atoi(string(ctx.Args[1]))
	if err != nil {
		return nil, ErrInteger
	}

	key := []byte(ctx.Args[0])
	if offset < 0 || offset > MaxRangeInteger {
		return nil, ErrMaximum
	}

	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	// If the offset is larger than the current length of the string at key, the string is padded with zero-bytes to make offset fit.
	val, err := str.SetRange(int64(offset), []byte(ctx.Args[2]))
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(len(val))), nil

}

// Incr increments the integer value of a key  by one
func Incr(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	delta, err := str.Incr(1)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(delta)), nil
}

// IncrBy increments the integer value of a key by the given amount
func IncrBy(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	delta, err := strconv.ParseInt(string(ctx.Args[1]), 10, 0)
	if err != nil {
		return nil, ErrInteger
	}

	delta, err = str.Incr(delta)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(delta)), nil
}

// IncrByFloat increments the float value of a key by the given amount
func IncrByFloat(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	delta, err := strconv.ParseFloat(string(ctx.Args[1]), 64)
	if err != nil {
		return nil, ErrInteger
	}
	delta, err = str.Incrf(delta)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return BulkString(ctx.Out, strconv.FormatFloat(delta, 'f', 17, 64)), nil
}

// Decr decrements the integer value of a key by one
func Decr(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	delta, err := str.Incr(-1)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(delta)), nil
}

// DecrBy decrements the integer value of a key by the given number
func DecrBy(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	delta, err := strconv.ParseInt(string(ctx.Args[1]), 10, 64)
	if err != nil {
		return nil, ErrInteger
	}

	delta, err = str.Incr(-delta)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(delta)), nil
}

// SetBit Sets or clears the bit at offset in the string value stored at key.
func SetBit(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	offset, err := strconv.Atoi(string(ctx.Args[1]))
	if err != nil {
		return nil, ErrBitOffset
	}
	if offset < 0 {
		return nil, ErrBitOffset
	}

	on, err := strconv.Atoi(string(ctx.Args[2]))
	if err != nil {
		return nil, ErrBitInteger
	}

	// Bits can only be set or cleared...
	if (on & ^1) != 0 {
		return nil, ErrBitInteger
	}

	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	val, err := str.SetBit(offset, on)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	if val != 0 {
		return Integer(ctx.Out, 1), nil
	}
	return Integer(ctx.Out, 0), nil
}

// GetBit get the bit at offset in the string value stored at key.
func GetBit(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	offset, err := strconv.Atoi(string(ctx.Args[1]))
	if err != nil || offset < 0 {
		return nil, ErrBitOffset
	}

	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	val, err := str.GetBit(offset)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	if val != 0 {
		return Integer(ctx.Out, 1), nil
	}
	return Integer(ctx.Out, 0), nil
}
