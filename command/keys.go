package command

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitlab.meitu.com/platform/titan/db"
	"gitlab.meitu.com/platform/titan/protocol"
)

// scan iter max count
const ScanMaxCount = 1024
const defaultScanCount = 10

func Delete(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	kv := txn.Kv()
	keys := make([][]byte, len(ctx.Args))
	for i := range ctx.Args {
		keys[i] = []byte(ctx.Args[i])
	}
	count, err := kv.Delete(keys)
	if err != nil {
		return nil, err
	}

	return Intege(ctx.Out, int64(count))
}

func ExistsHandler(args [][]byte, cmdctx *CmdCtx) (*protocol.ReplyData, error) {
	kv, err := cmdctx.Db.Kv()
	if err != nil {
		return RedisTikvResp, err
	}

	count, err := kv.Exists(args...)
	if err != nil {
		return RedisTikvResp, err
	}
	rt := &protocol.ReplyData{
		Type:  protocol.REPLYBINT,
		Value: count,
	}
	return rt, nil
}

func expireAction(cmdctx *CmdCtx, key []byte, expireAt int64) error {
	now := time.Now().UnixNano()
	kv, err := cmdctx.Db.Kv()
	if err != nil {
		return err
	}
	if expireAt <= now {
		count, dErr := kv.Delete(key)
		if dErr == nil && count == 0 {
			return errors.New("delete not exists key")
		}
		return dErr
	} else {
		err = kv.ExpireAt(key, uint64(expireAt))
	}
	return err
}

func ExpireHandler(args [][]byte, cmdctx *CmdCtx) (*protocol.ReplyData, error) {
	key := args[0]
	seconds, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return RedisIntegerResp, err
	}
	rt := &protocol.ReplyData{
		Type:  protocol.REPLYBINT,
		Value: int64(0),
	}
	expireAt := time.Now().Add(time.Second * time.Duration(seconds)).UnixNano()
	err = expireAction(cmdctx, key, expireAt)
	if err == nil {
		rt.Value = int64(1)
	}
	return rt, err
}

func ExpireAtHandler(args [][]byte, cmdctx *CmdCtx) (*protocol.ReplyData, error) {
	key := args[0]
	timestamp, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return RedisIntegerResp, err
	}

	expireAt := int64(time.Second * time.Duration(timestamp))
	err = expireAction(cmdctx, key, expireAt)
	rt := &protocol.ReplyData{
		Type:  protocol.REPLYBINT,
		Value: int64(0),
	}
	if err == nil {
		rt.Value = int64(1)
	}
	return rt, err
}

func PersistHandeler(args [][]byte, cmdctx *CmdCtx) (*protocol.ReplyData, error) {
	key := args[0]
	kv, err := cmdctx.Db.Kv()
	if err != nil {
		return RedisTikvResp, err
	}
	rt := &protocol.ReplyData{
		Type:  protocol.REPLYBINT,
		Value: int64(0),
	}
	err = kv.Persist(key)
	if err == nil {
		rt.Value = int64(1)
	}
	return rt, err
}

func PExpireHandler(args [][]byte, cmdctx *CmdCtx) (*protocol.ReplyData, error) {
	key := args[0]
	millsecond, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return RedisIntegerResp, err
	}
	expireAt := time.Now().Add(time.Millisecond * time.Duration(millsecond)).UnixNano()
	err = expireAction(cmdctx, key, expireAt)
	rt := &protocol.ReplyData{
		Type:  protocol.REPLYBINT,
		Value: int64(0),
	}
	if err == nil {
		rt.Value = int64(1)
	}
	return rt, nil

}

func PExpireAtHandler(args [][]byte, cmdctx *CmdCtx) (*protocol.ReplyData, error) {
	key := args[0]
	millsecond, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return RedisIntegerResp, err
	}
	expireAt := int64(time.Millisecond * time.Duration(millsecond))
	err = expireAction(cmdctx, key, expireAt)
	rt := &protocol.ReplyData{
		Type:  protocol.REPLYBINT,
		Value: int64(0),
	}
	if err == nil {
		rt.Value = int64(1)
	}
	return rt, nil
}

func TTLHandler(args [][]byte, cmdctx *CmdCtx) (*protocol.ReplyData, error) {
	key := args[0]
	kv, err := cmdctx.Db.Kv()
	if err != nil {
		return RedisTikvResp, err
	}

	rt := &protocol.ReplyData{
		Type:  protocol.REPLYBINT,
		Value: int64(0),
	}
	now := time.Now().UnixNano()
	ttl, err := kv.TTL(key, uint64(now))
	if err != nil {
		return rt, err
	}

	if ttl <= 0 {
		rt.Value = ttl
	} else {
		rt.Value = ttl / int64(time.Second)
	}
	return rt, nil
}

func PTTLHandler(args [][]byte, cmdctx *CmdCtx) (*protocol.ReplyData, error) {
	key := args[0]
	kv, err := cmdctx.Db.Kv()
	if err != nil {
		return RedisTikvResp, err
	}
	rt := &protocol.ReplyData{
		Type:  protocol.REPLYBINT,
		Value: int64(0),
	}
	now := time.Now().UnixNano()

	ttl, err := kv.TTL(key, uint64(now))
	if err != nil {
		return rt, err
	}

	if ttl <= 0 {
		rt.Value = ttl
	} else {
		rt.Value = ttl / int64(time.Millisecond)
	}
	return rt, nil
}

func ObjectHandler(args [][]byte, cmdctx *CmdCtx) (*protocol.ReplyData, error) {
	argc := len(args)
	subCmd := strings.ToLower(string(args[0]))

	var objectErr protocol.RedisError = fmt.Errorf("ERR Unknown subcommand or wrong number of arguments for '%s'. Try OBJECT help", subCmd)
	rt := &protocol.ReplyData{}
	if argc == 1 && subCmd == "help" {
		rt.Type = protocol.REPLYARRAY
		helpInfo := []string{
			"OBJECT <subcommand> key. Subcommands:",
			"ENCODING <key> -- Return the kind of internal representation used in order to store the value associated with a key.",
			"FREQ <key> -- Return the access frequency index of the key. The returned integer is proportional to the logarithm of the recent access frequency of the key.",
			"IDLETIME <key> -- Return the idle time of the key, that is the approximated number of seconds elapsed since the last access to the key.",
			"REFCOUNT <key> -- Return the number of references of the value associated with the specified key.",
		}
		val := make([]*protocol.ReplyData, 0, len(helpInfo))
		for _, info := range helpInfo {
			val = append(val, &protocol.ReplyData{Type: protocol.REPLYSIMPLESTRING, Value: info})
		}
		rt.Value = val
		return rt, nil
	} else if argc == 2 {
		key := args[1]
		obj, err := cmdctx.Db.Object(key)
		if err != nil {
			if db.IsErrNotFound(err) {
				return RedisNilResp, nil
			}
			return RedisTikvResp, err
		}

		switch subCmd {
		case "refcount":
			rt.Type = protocol.REPLYBINT
			rt.Value = int64(0)
		case "idletime":
			sec := int64(time.Since(time.Unix(0, int64(obj.UpdatedAt))).Seconds())
			rt.Type = protocol.REPLYBINT
			rt.Value = sec
		case "encoding":
			rt.Type = protocol.REPLYSIMPLESTRING
			rt.Value = obj.Encoding.String()
		case "freq":
			rt.Type = protocol.REPLYBINT
			rt.Value = int64(0)
		default:
			rt.Type = protocol.REPLYERROR
			rt.Value = objectErr
			err = objectErr
		}
		return rt, err
	}
	rt.Type = protocol.REPLYERROR
	rt.Value = objectErr
	return rt, objectErr
}

func TypeHandler(args [][]byte, cmdctx *CmdCtx) (*protocol.ReplyData, error) {
	key := args[0]
	rt := &protocol.ReplyData{Type: protocol.REPLYSIMPLESTRING}
	obj, err := cmdctx.Db.Object(key)
	if err != nil {
		if db.IsErrNotFound(err) {
			rt.Value = "none"
			return rt, nil
		}
		return RedisTikvResp, err
	}
	rt.Value = obj.Type.String()

	return rt, nil
}

func KeysHandler(args [][]byte, cmdctx *CmdCtx) (*protocol.ReplyData, error) {
	pattern := args[0]
	allkey := (pattern[0] == '*' && len(pattern) == 1)
	prefixKey := PatternStringPrefix(pattern)
	kv, err := cmdctx.Db.Kv()
	if err != nil {
		return RedisTikvResp, err
	}

	keyList := make([]*protocol.ReplyData, 0)
	callback := func(key []byte) bool {
		if allkey || PatternMatch(pattern, key, false) {
			keyList = append(keyList, &protocol.ReplyData{Type: protocol.REPLYSIMPLESTRING, Value: string(key)})
		}
		return true
	}

	err = kv.Keys(prefixKey, prefixKey, callback)
	if err != nil {
		return RedisTikvResp, err
	}

	rt := &protocol.ReplyData{
		Type:  protocol.REPLYARRAY,
		Value: keyList,
	}
	return rt, nil
}

func ScanHandler(args [][]byte, cmdctx *CmdCtx) (*protocol.ReplyData, error) {
	var (
		cursor   []byte = args[0]
		count    uint64 = defaultScanCount
		pattern  []byte
		prefix   []byte
		matchAll bool
	)
	if string(cursor) == "0" {
		cursor = nil
	}

	if len(args)%2 == 0 {
		return RedisValueResp, protocol.ErrValue
	}

	var err error
	for i := 1; i < len(args); i += 2 {
		argv := strings.ToLower(string(args[i]))
		nextArgv := args[i+1]
		switch argv {
		case "count":
			if count, err = strconv.ParseUint(string(nextArgv), 10, 64); err != nil {
				return RedisIntegerResp, protocol.ErrValue
			}
			if count > ScanMaxCount {
				count = ScanMaxCount
			} else if count == 0 {
				count = defaultScanCount
			}
		case "match":
			pattern = nextArgv
			matchAll = string(pattern) == "*"
		}
	}

	if len(pattern) == 0 {
		matchAll = true
	} else {
		prefix = PatternStringPrefix(pattern)
		if cursor == nil && prefix != nil {
			cursor = prefix
		}
	}

	kv, err := cmdctx.Db.Kv()
	if err != nil {
		return RedisTikvResp, err
	}

	keyList := make([]*protocol.ReplyData, 0)
	lastCursor := &protocol.ReplyData{Type: protocol.REPLYBULK, Value: "0"}
	callback := func(key []byte) bool {
		strkey := string(key)
		if count <= 0 {
			lastCursor.Value = strkey
			return false
		}
		if matchAll || PatternMatch(pattern, key, false) {
			keyList = append(keyList, &protocol.ReplyData{Type: protocol.REPLYBULK, Value: strkey})
			count--
		}

		return true
	}

	err = kv.Keys(cursor, prefix, callback)
	if err != nil {
		return RedisTikvResp, err
	}

	val := make([]*protocol.ReplyData, 0, 2)
	val = append(val, lastCursor, &protocol.ReplyData{Type: protocol.REPLYARRAY, Value: keyList})

	rt := &protocol.ReplyData{
		Type:  protocol.REPLYARRAY,
		Value: val,
	}
	return rt, nil
}

// RandomKeyHandler return a random key from the currently selected database
func RandomKeyHandler(args [][]byte, cmdCtx *CmdCtx) (*protocol.ReplyData, error) {
	kv, _ := cmdCtx.Db.Kv()
	key, err := kv.RandomKey()
	if err != nil {
		return &protocol.ReplyData{Type: protocol.REPLYERROR, Value: []byte(err.Error())}, nil
	}
	if key == nil {
		return &protocol.ReplyData{Type: protocol.REPLYNIL, Value: nil}, nil
	}
	return &protocol.ReplyData{Type: protocol.REPLYBULK, Value: key}, err
}
