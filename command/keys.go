package command

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitlab.meitu.com/platform/thanos/db"
	"gitlab.meitu.com/platform/thanos/resp"
)

// scan iter max count
const ScanMaxCount = 255
const defaultScanCount = 10

func Delete(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	keys := make([][]byte, len(ctx.Args))
	for i := range ctx.Args {
		keys[i] = []byte(ctx.Args[i])
	}
	c, err := kv.Delete(keys)
	if err != nil {
		return nil, err
	}
	return Integer(ctx.Out, c), nil
}

//Exists check if the given keys exist
func Exists(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	keys := make([][]byte, len(ctx.Args))
	for i := range ctx.Args {
		keys[i] = []byte(ctx.Args[i])
	}
	c, err := kv.Exists(keys)
	if err != nil {
		return nil, err
	}
	return Integer(ctx.Out, c), nil
}

//Expire
func Expire(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	key := []byte(ctx.Args[0])
	seconds, err := strconv.Atoi(ctx.Args[1])
	if err != nil {
		return nil, ErrInteger
	}

	at := time.Now().Add(time.Second * time.Duration(seconds)).UnixNano()
	if err := kv.ExpireAt(key, at); err != nil {
		if err == db.ErrKeyNotFound {
			return Integer(ctx.Out, 0), nil
		}
		return nil, err
	}
	return Integer(ctx.Out, 1), nil
}

//ExpireAt
func ExpireAt(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	key := []byte(ctx.Args[0])
	timestamp, err := strconv.Atoi(ctx.Args[1])
	if err != nil {
		return nil, ErrInteger
	}

	at := int64(time.Second * time.Duration(timestamp))
	if at <= 0 {
		at = db.Now() - 1
	}

	if err := kv.ExpireAt(key, at); err != nil {
		if err == db.ErrKeyNotFound {
			return Integer(ctx.Out, 0), nil
		}
		return nil, err
	}

	return Integer(ctx.Out, 1), nil
}

//Persist
func Persist(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	key := []byte(ctx.Args[0])
	if err := kv.ExpireAt(key, 0); err != nil {
		if err == db.ErrKeyNotFound {
			return Integer(ctx.Out, 0), nil
		}
		return nil, err
	}
	return Integer(ctx.Out, 1), nil
}

//PExpire
func PExpire(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	key := []byte(ctx.Args[0])
	ms, err := strconv.Atoi(ctx.Args[1])
	if err != nil {
		return nil, ErrInteger
	}
	at := time.Now().Add(time.Millisecond * time.Duration(ms)).UnixNano()
	if err := kv.ExpireAt(key, at); err != nil {
		if err == db.ErrKeyNotFound {
			return Integer(ctx.Out, 0), nil
		}
		return nil, err
	}
	return Integer(ctx.Out, 1), nil

}

//PExpireAt
func PExpireAt(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	key := []byte(ctx.Args[0])
	ms, err := strconv.Atoi(ctx.Args[1])
	if err != nil {
		return nil, ErrInteger
	}
	at := int64(time.Millisecond * time.Duration(ms))
	if at <= 0 {
		at = db.Now()
	}
	if err := kv.ExpireAt(key, at); err != nil {
		if err == db.ErrKeyNotFound {
			return Integer(ctx.Out, 0), nil
		}
		return nil, err
	}
	return Integer(ctx.Out, 1), nil
}

//TTL
func TTL(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	now := db.Now()
	obj, err := txn.Object(key)
	if err != nil {
		if err == db.ErrKeyNotFound {
			return Integer(ctx.Out, -2), nil
		}
		return nil, err
	}
	if obj.ExpireAt == 0 {
		return Integer(ctx.Out, -1), nil
	}
	ttl := (obj.ExpireAt - now) / int64(time.Second)
	return Integer(ctx.Out, ttl), nil
}

//PTTL
func PTTL(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	now := db.Now()
	obj, err := txn.Object(key)
	if err != nil {
		if err == db.ErrKeyNotFound {
			return Integer(ctx.Out, -2), nil
		}
		return nil, err
	}
	if db.IsExpired(obj, now) {
		return Integer(ctx.Out, -2), nil
	}
	if obj.ExpireAt == 0 {
		return Integer(ctx.Out, -1), nil
	}
	ttl := (obj.ExpireAt - now) / int64(time.Millisecond)
	return Integer(ctx.Out, ttl), nil

}

//Object
func Object(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	argc := len(ctx.Args)
	subCmd := strings.ToLower(ctx.Args[0])
	cmdErr := fmt.Errorf("ERR Unknown subcommand or wrong number of arguments for '%s'. Try OBJECT help", subCmd)
	if argc == 1 && subCmd == "help" {

		helpInfo := [][]byte{
			[]byte("OBJECT <subcommand> key. Subcommands:"),
			[]byte("ENCODING <key> -- Return the kind of internal representation used in order to store the value associated with a key."),
			[]byte("FREQ <key> -- Return the access frequency index of the key. The returned integer is proportional to the logarithm of the recent access frequency of the key."),
			[]byte("IDLETIME <key> -- Return the idle time of the key, that is the approximated number of seconds elapsed since the last access to the key."),
			[]byte("REFCOUNT <key> -- Return the number of references of the value associated with the specified key."),
		}
		return BytesArray(ctx.Out, helpInfo), nil
	} else if argc == 2 {
		key := []byte(ctx.Args[1])
		obj, err := txn.Object(key)
		if err != nil {
			if err == db.ErrKeyNotFound {
				return NullBulkString(ctx.Out), nil
			}
			return nil, err
		}

		switch subCmd {
		case "refcount", "freq":
			return Integer(ctx.Out, 0), nil
		case "idletime":
			sec := int64(time.Since(time.Unix(0, obj.UpdatedAt)).Seconds())
			return Integer(ctx.Out, sec), nil
		case "encoding":
			return SimpleString(ctx.Out, obj.Encoding.String()), nil
		}
	}
	return nil, cmdErr
}

//Type
func Type(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	obj, err := txn.Object(key)
	if err != nil {
		if err == db.ErrKeyNotFound {
			return SimpleString(ctx.Out, "none"), nil
		}
		return nil, err
	}

	return SimpleString(ctx.Out, obj.Type.String()), nil
}

//Keys
func Keys(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	list := make([][]byte, 0)
	pattern := []byte(ctx.Args[0])
	all := (pattern[0] == '*' && len(pattern) == 1)
	prefix := globMatchPrefix(pattern)

	kv := txn.Kv()
	f := func(key []byte) bool {
		if all || globMatch(pattern, key, false) {
			list = append(list, key)
		}
		return true
	}

	if err := kv.Keys(prefix, f); err != nil {
		return nil, err
	}
	return BytesArray(ctx.Out, list), nil
}

//Scan
func Scan(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var (
		start   []byte
		end     []byte = []byte("0")
		count   uint64 = defaultScanCount
		pattern []byte
		prefix  []byte
		all     bool
		err     error
	)
	if strings.Compare(ctx.Args[0], "0") != 0 {
		start = []byte(ctx.Args[0])
	}

	if len(ctx.Args)%2 == 0 {
		return nil, ErrInteger
	}

	for i := 1; i < len(ctx.Args); i += 2 {
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
				count = defaultScanCount
			}
		case "match":
			pattern = []byte(next)
			all = (pattern[0] == '*' && len(pattern) == 1)
		}
	}

	if len(pattern) == 0 {
		all = true
	} else {
		prefix = globMatchPrefix(pattern)
		if start == nil && prefix != nil {
			start = prefix
		}
	}

	kv := txn.Kv()
	list := [][]byte{}
	f := func(key []byte) bool {
		if count <= 0 {
			end = key
			return false
		}
		if prefix != nil && !bytes.HasPrefix(key, prefix) {
			return false
		}
		if all || globMatch(pattern, key, false) {
			list = append(list, key)
			count--
		}
		return true
	}

	err = kv.Keys(start, f)
	if err != nil {
		return nil, err
	}
	return func() {
		resp.ReplyArray(ctx.Out, 2)
		resp.ReplyBulkString(ctx.Out, string(end))
		BytesArray(ctx.Out, list)
	}, nil

}

// RandomKey return a random key from the currently selected database
func RandomKey(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	key, err := kv.RandomKey()
	if err != nil {
		return nil, err
	}
	if key == nil {
		return NullBulkString(ctx.Out), nil
	}
	return BulkString(ctx.Out, string(key)), nil
}
