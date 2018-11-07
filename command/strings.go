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
func MSetNx(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	argc := len(ctx.Args)
	if argc%2 != 0 {
		return nil, errors.New("ERR wrong number of arguments for MSET")
	}
	for i := 0; i < argc-1; i += 2 {
		if str, _ := txn.String([]byte(ctx.Args[0])); !str.Exist() {
			if _, err := Set(ctx, txn); err != nil {
				return nil, err
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
func Append(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR " + err.Error())
	}
	value, err := str.Get()

	if err == db.ErrKeyNotFound {
		return NullBulkString(ctx.Out), nil
	}

	value = append(value, ctx.Args[1]...)
	if err := str.Set(value); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(len(value))), nil
}

//TODO
func GetBit(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	return nil, nil
}
func GetSet(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR " + err.Error())
	}
	value, err := str.Get()
	if err == db.ErrKeyNotFound {
		return NullBulkString(ctx.Out), nil
	}

	if err := str.Set([]byte(ctx.Args[1])); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	if value != nil {
		return BulkString(ctx.Out, string(value)), nil
	}
	return nil, nil
}

func GetRange(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR " + err.Error())
	}

	start, err := strconv.Atoi(string(ctx.Args[1]))
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	end, err := strconv.Atoi(string(ctx.Args[2]))
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	value, err := str.Get()

	if err == db.ErrKeyNotFound {
		return NullBulkString(ctx.Out), nil
	}

	vlen := len(value)
	if end < 0 {
		end = vlen + end
	}
	if start < 0 {
		start = vlen + start
	}
	if start > end || start > vlen || end < 0 {
		return nil, errors.New("EmptyArray error")
	}
	if end > vlen {
		end = vlen
	}
	if start < 0 {
		start = 0
	}
	return BulkString(ctx.Out, string(value[start:][:end+1])), nil
}

func SetNx(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	//get the key
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
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

//SETEX KEY_NAME TIMEOUT VALUE
func SetEx(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	//get the key
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if []byte(ctx.Args[1]) != nil {
		ui, err := strconv.ParseUint(string(ctx.Args[1]), 10, 64)
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		unit := ui * uint64(time.Second)
		if err := str.Set([]byte(ctx.Args[2]), int64(unit)); err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
	}
	return SimpleString(ctx.Out, "OK"), nil

}

func PSetEx(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	//get the key
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if []byte(ctx.Args[1]) != nil {
		ui, err := strconv.ParseUint(string(ctx.Args[1]), 10, 64)
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		unit := ui * uint64(time.Microsecond)
		if err := str.Set([]byte(ctx.Args[2]), int64(unit)); err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
	}
	return SimpleString(ctx.Out, "OK"), nil
}

//setrange key offset value
func SetRange(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR " + err.Error())
	}

	offset, err := strconv.Atoi(string(ctx.Args[1]))
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	if offset < 0 || offset > MaxRangeInteger {
		return nil, errors.New("ERR string exceeds maximum allowed size")
	}

	value, err := str.Get()
	if err == db.ErrKeyNotFound {
		return NullBulkString(ctx.Out), nil
	}

	vlen := len(value)
	if vlen < offset+len(ctx.Args[2]) {
		value = append(value, make([]byte, len(ctx.Args[2])+offset-vlen)...)
	}
	copy(value[offset:], ctx.Args[2])
	return Integer(ctx.Out, int64(len(value))), nil
}

//TODO
func SetBit(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	return nil, nil
}
func Incr(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR " + err.Error())
	}
	return incr(1, str, ctx)
}

func IncrBy(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR " + err.Error())
	}
	delta, err := strconv.ParseInt(string(ctx.Args[1]), 10, 0)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return incr(delta, str, ctx)
}

func IncrByFloat(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR " + err.Error())
	}
	delta, err := strconv.ParseFloat(string(ctx.Args[1]), 64)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return incrf(delta, str, ctx)
}

func Decr(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR " + err.Error())
	}
	return incr(-1, str, ctx)

}

func DecrBy(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR " + err.Error())
	}
	delta, err := strconv.ParseInt(string(ctx.Args[1]), 10, 64)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return incr(-delta, str, ctx)
}

func DecrByFloat(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		return nil, errors.New("ERR " + err.Error())
	}
	delta, err := strconv.ParseFloat(string(ctx.Args[1]), 64)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return incrf(-delta, str, ctx)
}
func incr(delta int64, str *db.String, ctx *Context) (OnCommit, error) {
	value, err := str.Get()
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	if value != nil {
		v, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		delta = v + delta
	}

	vs := strconv.FormatInt(delta, 10)
	if err := str.Set([]byte(vs)); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(delta)), nil
}
func incrf(delta float64, str *db.String, ctx *Context) (OnCommit, error) {
	value, err := str.Get()
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	if value != nil {
		v, err := strconv.ParseFloat(string(value), 64)
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		delta = v + delta
	}

	vs := strconv.FormatFloat(delta, 'e', -1, 64)
	if err := str.Set([]byte(vs)); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return SimpleString(ctx.Out, strconv.FormatFloat(delta, 'f', 17, 64)), nil
}
