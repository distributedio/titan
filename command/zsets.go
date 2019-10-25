package command

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/distributedio/titan/db"
)

// ZAdd adds the specified members with scores to the sorted set
func ZAdd(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	kvs := ctx.Args[1:]
	if len(kvs)%2 != 0 {
		return nil, errors.New("ERR syntax error")
	}

	uniqueMembers := make(map[string]bool)
	count := len(kvs) / 2
	members := make([][]byte, 0, count)
	scores := make([]float64, 0, count)
	for i := 0; i < len(kvs)-1; i += 2 {
		member := kvs[i+1]
		if _, ok := uniqueMembers[member]; ok {
			continue
		}

		members = append(members, []byte(member))
		score, err := strconv.ParseFloat(kvs[i], 64)
		if err != nil || math.IsNaN(score) {
			return nil, ErrFloat
		}
		scores = append(scores, score)

		uniqueMembers[member] = true
	}

	zset, err := txn.ZSet(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	added, err := zset.ZAdd(members, scores)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	return Integer(ctx.Out, added), nil
}

func ZRange(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	return zAnyOrderRange(ctx, txn, true)
}

func ZRevRange(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	return zAnyOrderRange(ctx, txn, false)
}

func zAnyOrderRange(ctx *Context, txn *db.Transaction, positiveOrder bool) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	start, err := strconv.ParseInt(ctx.Args[1], 10, 64)
	if err != nil {
		return nil, ErrInteger
	}
	stop, err := strconv.ParseInt(ctx.Args[2], 10, 64)
	if err != nil {
		return nil, ErrInteger
	}
	withScore := bool(false)
	if len(ctx.Args) >= 4 {
		if strings.ToUpper(ctx.Args[3]) == "WITHSCORES" {
			withScore = true
		}
	}

	zset, err := txn.ZSet(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if !zset.Exist() {
		return BytesArrayOnce(ctx.Out, nil), nil
	}

	items, err := zset.ZAnyOrderRange(start, stop, withScore, positiveOrder)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	if len(items) == 0 {
		return BytesArrayOnce(ctx.Out, nil), nil
	}
	return BytesArrayOnce(ctx.Out, items), nil
}

func ZRangeByScore(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	return zAnyOrderRangeByScore(ctx, txn, true)
}

func ZRevRangeByScore(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	return zAnyOrderRangeByScore(ctx, txn, false)
}

func zAnyOrderRangeByScore(ctx *Context, txn *db.Transaction, positiveOrder bool) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	startScore, startInclude, err := getFloatAndInclude(ctx.Args[1])
	if err != nil {
		return nil, ErrMinOrMaxNotFloat
	}
	endScore, endInclude, err := getFloatAndInclude(ctx.Args[2])
	if err != nil {
		return nil, ErrMinOrMaxNotFloat
	}

	withScore := bool(false)
	offset := int64(0)
	count := int64(math.MaxInt64)
	for i := 3; i < len(ctx.Args); i++ {
		switch strings.ToUpper(ctx.Args[i]) {
		case "WITHSCORES":
			withScore = true
		case "LIMIT":
			if offset, count, err = getLimitParameters(ctx.Args[i+1:]); err != nil {
				return nil, err
			}
			i += 2
		default:
			return nil, ErrSyntax
		}
	}

	zset, err := txn.ZSet(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if !zset.Exist() {
		return BytesArrayOnce(ctx.Out, nil), nil
	}

	items, err := zset.ZAnyOrderRangeByScore(startScore, startInclude,
		endScore, endInclude,
		withScore,
		offset, count,
		positiveOrder)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	if len(items) == 0 {
		return BytesArrayOnce(ctx.Out, nil), nil
	}
	return BytesArrayOnce(ctx.Out, items), nil
}

func ZRem(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	uniqueMembers := make(map[string]bool)
	members := make([][]byte, 0, len(ctx.Args)-1)
	for _, member := range ctx.Args[1:] {
		if _, ok := uniqueMembers[member]; ok {
			continue
		}

		members = append(members, []byte(member))
		uniqueMembers[member] = true
	}

	zset, err := txn.ZSet(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if !zset.Exist() {
		return Integer(ctx.Out, 0), nil
	}

	deleted, err := zset.ZRem(members)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	return Integer(ctx.Out, deleted), nil
}

func ZCard(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	zset, err := txn.ZSet(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if !zset.Exist() {
		return Integer(ctx.Out, 0), nil
	}

	return Integer(ctx.Out, zset.ZCard()), nil
}

func ZScore(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	member := []byte(ctx.Args[1])

	zset, err := txn.ZSet(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if !zset.Exist() {
		return NullBulkString(ctx.Out), nil
	}

	score, err := zset.ZScore(member)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	if score == nil {
		return NullBulkString(ctx.Out), nil
	}

	return BulkString(ctx.Out, string(score)), nil
}
