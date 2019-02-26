package command

import (
    "github.com/meitu/titan/db"
    "strconv"
    "errors"
)

// ZAdd adds the specified members with scores to the sorted set
func ZAdd(ctx *Context, txn *db.Transaction) (OnCommit, error) {
    key := []byte(ctx.Args[0])

    kvs := ctx.Args[1:]
    if len(kvs)%2 != 0 {
        return nil, errors.New("ERR wrong number of arguments for ZADD")
    }

    uniqueMembers := make(map[string]bool)
    count := len(kvs)/2
    members := make([][]byte, 0, count)
    scores := make([]float64, 0, count)
    for i := 0; i < len(kvs)-1; i += 2 {
        member := kvs[i+1]
        if _, ok := uniqueMembers[member]; ok{
            continue
        }

        members = append(members, []byte(member))
        score, err := strconv.ParseFloat(kvs[i], 64)
        if err != nil {
            return nil, errors.New("ERR " + err.Error())
        }
        scores = append(scores, score)

        uniqueMembers[member]=true
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
        if ctx.Args[3] == "WITHSCORES" {
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
        return BytesArray(ctx.Out, nil), nil
    }

    items, err := zset.ZAnyOrderRange(start, stop, withScore, positiveOrder)
    if err != nil {
        return nil, errors.New("ERR " + err.Error())
    }
    if len(items) == 0 {
        return BytesArray(ctx.Out, nil), nil
    }
    return BytesArray(ctx.Out, items), nil
}

func ZRem(ctx *Context, txn *db.Transaction)(OnCommit, error) {
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