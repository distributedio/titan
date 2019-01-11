package command

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/meitu/titan/db"
	"github.com/meitu/titan/encoding/resp"
)

const (
	//ScanMaxCount is the max limitation of a single scan
	ScanMaxCount = 255
	// defautlScanCout is used when no hints being supplied by clients
	defaultScanCount = 10
)

// Delete removes the specified keys. A key is ignored if it does not exist
func Delete(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	keys := make([][]byte, len(ctx.Args))
	for i := range ctx.Args {
		keys[i] = []byte(ctx.Args[i])
	}
	c, err := kv.Delete(keys)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, c), nil
}

// Exists returns if key exists
func Exists(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	keys := make([][]byte, len(ctx.Args))
	for i := range ctx.Args {
		keys[i] = []byte(ctx.Args[i])
	}
	c, err := kv.Exists(keys)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, c), nil
}

// Expire sets a timeout on key
func Expire(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	key := []byte(ctx.Args[0])
	seconds, err := strconv.ParseInt(ctx.Args[1], 10, 64)
	if err != nil {
		return nil, ErrInteger
	}

	at := time.Now().Add(time.Second * time.Duration(seconds)).UnixNano()
	if err := kv.ExpireAt(key, at); err != nil {
		if err == db.ErrKeyNotFound {
			return Integer(ctx.Out, 0), nil
		}
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, 1), nil
}

// ExpireAt sets an absolute timestamp to expire on key
func ExpireAt(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	key := []byte(ctx.Args[0])
	timestamp, err := strconv.ParseInt(ctx.Args[1], 10, 64)
	if err != nil {
		return nil, ErrInteger
	}

	at := int64(time.Second * time.Duration(timestamp))
	if at <= 0 {
		at = 1
	}

	if err := kv.ExpireAt(key, at); err != nil {
		if err == db.ErrKeyNotFound {
			return Integer(ctx.Out, 0), nil
		}
		return nil, errors.New("ERR " + err.Error())
	}

	return Integer(ctx.Out, 1), nil
}

// Persist removes the existing timeout on key, turning the key from volatile to persistent
func Persist(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	key := []byte(ctx.Args[0])
	obj, err := txn.Object(key)
	if err != nil && err != db.ErrKeyNotFound {
		return nil, errors.New("ERR " + err.Error())
	}
	if err == db.ErrKeyNotFound || obj.ExpireAt == 0 {
		return Integer(ctx.Out, 0), nil
	}

	if err := kv.ExpireAt(key, 0); err != nil {
		if err == db.ErrKeyNotFound {
			return Integer(ctx.Out, 0), nil
		}
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, 1), nil
}

// PExpire works exactly like expire but the time to live of the key is specified in milliseconds instead of seconds
func PExpire(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	key := []byte(ctx.Args[0])
	ms, err := strconv.ParseInt(ctx.Args[1], 10, 64)
	if err != nil {
		return nil, ErrInteger
	}
	at := time.Now().Add(time.Millisecond * time.Duration(ms)).UnixNano()
	if err := kv.ExpireAt(key, at); err != nil {
		if err == db.ErrKeyNotFound {
			return Integer(ctx.Out, 0), nil
		}
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, 1), nil

}

// PExpireAt has the same effect and semantic as expireAt,
// but the Unix time at which the key will expire is specified in milliseconds instead of seconds
func PExpireAt(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	key := []byte(ctx.Args[0])
	ms, err := strconv.ParseInt(ctx.Args[1], 10, 64)
	if err != nil {
		return nil, ErrInteger
	}
	at := int64(time.Millisecond * time.Duration(ms))
	if at <= 0 {
		at = 1
	}
	if err := kv.ExpireAt(key, at); err != nil {
		if err == db.ErrKeyNotFound {
			return Integer(ctx.Out, 0), nil
		}
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, 1), nil
}

// TTL returns the remaining time to live of a key that has a timeout
func TTL(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	now := db.Now()
	obj, err := txn.Object(key)
	if err != nil {
		if err == db.ErrKeyNotFound {
			return Integer(ctx.Out, -2), nil
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if obj.ExpireAt == 0 {
		return Integer(ctx.Out, -1), nil
	}
	ttl := (obj.ExpireAt - now) / int64(time.Second)
	return Integer(ctx.Out, ttl), nil
}

// PTTL likes TTL this command returns the remaining time to live of a key that has an expire set,
// with the sole difference that TTL returns the amount of remaining time in seconds while PTTL returns it in milliseconds
func PTTL(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	now := db.Now()
	obj, err := txn.Object(key)
	if err != nil {
		if err == db.ErrKeyNotFound {
			return Integer(ctx.Out, -2), nil
		}
		return nil, errors.New("ERR " + err.Error())
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

// Object inspects the internals of Redis Objects
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
			return nil, errors.New("ERR " + err.Error())
		}
		switch subCmd {
		case "refcount", "freq":
			return Integer(ctx.Out, 0), nil
		case "idletime":
			if obj.Type == db.ObjectHash {
				hash, err := txn.Hash(key)
				if err != nil {
					return nil, errors.New("ERR " + err.Error())
				}
				obj, err = hash.Object()
				if err != nil {
					return nil, errors.New("ERR " + err.Error())
				}
			}
			sec := int64(time.Since(time.Unix(0, obj.UpdatedAt)).Seconds())
			return Integer(ctx.Out, sec), nil
		case "encoding":
			return SimpleString(ctx.Out, obj.Encoding.String()), nil
		}
	}
	return nil, cmdErr
}

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

// Keys returns all keys matching pattern
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
		return nil, errors.New("ERR " + err.Error())
	}
	return BytesArray(ctx.Out, list), nil
}

// Scan incrementally iterates the key space
func Scan(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var (
		start   []byte
		end            = []byte("0")
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

	if err := kv.Keys(start, f); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return func() {
		resp.ReplyArray(ctx.Out, 2)
		resp.ReplyBulkString(ctx.Out, string(end))
		resp.ReplyArray(ctx.Out, len(list))
		for i := range list {
			resp.ReplyBulkString(ctx.Out, string(list[i]))
		}
	}, nil

}

// RandomKey returns a random key from the currently selected database
func RandomKey(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	key, err := kv.RandomKey()
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	if key == nil {
		return NullBulkString(ctx.Out), nil
	}
	return BulkString(ctx.Out, string(key)), nil
}
