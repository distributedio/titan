package command

import (
	"errors"
	"strconv"
	"time"

	"gitlab.meitu.com/platform/thanos/db"
	"gitlab.meitu.com/platform/thanos/resp"
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
	var flag int // flag int // 0 -- null 1---nx  2---xx
	var unit int64
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

	s, err := txn.String(key)
	if err != nil && err != db.ErrKeyNotFound {
		return nil, errors.New("ERR " + err.Error())
	}

	//xx
	if flag == 2 {
		if !s.Exist() {
			return NullBulkString(ctx.Out), nil
		}
	}

	//nx
	if flag == 1 {
		if s.Exist() {
			return NullBulkString(ctx.Out), nil
		}
	}

	s = txn.NewString(key)
	if err := s.Set(value, unit); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return SimpleString(ctx.Out, "OK"), nil
}

// MGet returns the values of all specified key
// use BatchGetRequest to gain performance
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
			resp.ReplyNullBulkString(ctx.Out)
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
	return SimpleString(ctx.Out, "OK"), nil
}

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

//Append
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

	if !str.Exist() {
		str = txn.NewString(key)
	}

	llen, err := str.Append(value)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(llen)), nil
}

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

func SetNx(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	//get the key
	key := []byte(ctx.Args[0])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	//key 不存在时，为 key 设置指定的值。设置成功，返回 1 。 设置失败，返回 0 。
	if str.Exist() {
		return Integer(ctx.Out, int64(0)), nil
	}

	str = txn.NewString(key)
	if err := str.Set([]byte(ctx.Args[1])); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(1)), nil
}

//SETEX KEY_NAME TIMEOUT VALUE
func SetEx(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	//get the key
	key := []byte(ctx.Args[0])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !str.Exist() {
		str = txn.NewString(key)
	}

	ui, err := strconv.ParseInt(string(ctx.Args[1]), 10, 64)
	if err != nil {
		return nil, ErrInteger
	}
	unit := ui * int64(time.Second)
	if err := str.Set([]byte(ctx.Args[2]), unit); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	return SimpleString(ctx.Out, "OK"), nil
}

func PSetEx(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	//get the key
	key := []byte(ctx.Args[0])
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !str.Exist() {
		str = txn.NewString(key)
	}

	ui, err := strconv.ParseUint(string(ctx.Args[1]), 10, 64)
	if err != nil {
		return nil, ErrInteger
	}
	unit := ui * uint64(time.Microsecond)
	if err := str.Set([]byte(ctx.Args[2]), int64(unit)); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	return SimpleString(ctx.Out, "OK"), nil
}

/*
//setrange key offset value
TODO bug
func SetRange(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	offset, err := strconv.Atoi(string(ctx.Args[1]))
	if err != nil {
		return nil, ErrInteger
	}

	if offset < 0 || offset > MaxRangeInteger {
		return nil, ErrMaximum
	}

	if !str.Exist() {
		return NullBulkString(ctx.Out), nil
	}

	vlen := len(value)
	if vlen < offset+len(ctx.Args[2]) {
		value = append(value, make([]byte, len(ctx.Args[2])+offset-vlen)...)
	}
	copy(value[offset:], ctx.Args[2])
	return Integer(ctx.Out, int64(len(value))), nil
}
*/

func Incr(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if !str.Exist() {
		str = txn.NewString(key)
	}
	delta, err := str.Incr(1)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(delta)), nil
}

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

	if !str.Exist() {
		str = txn.NewString(key)
	}

	delta, err = str.Incr(delta)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(delta)), nil
}

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

	if !str.Exist() {
		str = txn.NewString(key)
	}

	delta, err = str.Incrf(delta)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return SimpleString(ctx.Out, strconv.FormatFloat(delta, 'f', 17, 64)), nil
}

func Decr(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if !str.Exist() {
		str = txn.NewString(key)
	}

	delta, err := str.Incr(-1)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(delta)), nil
}

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
	if !str.Exist() {
		str = txn.NewString(key)
	}

	delta, err = str.Incr(-delta)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(delta)), nil
}
