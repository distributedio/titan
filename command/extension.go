package command

import (
	"fmt"
	"github.com/distributedio/titan/db"
	"github.com/distributedio/titan/encoding/resp"
	"math"
	"strconv"
)

// Escan scan the expiration list
// escan [from start] [to end] [count N]
func Escan(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var from, to, count int64
	to = math.MaxInt64
	count = 10
	args := ctx.Args
	if len(args)%2 != 0 {
		return nil, ErrWrongArgs("escan")
	}

	var p *int64
	for i := 0; i < len(args)-1; i++ {
		switch args[i] {
		case "from":
			p = &from
		case "to":
			p = &to
		case "count":
			p = &count
		default:
			return nil, ErrSyntax
		}
		i++
		val, err := strconv.ParseInt(args[i], 10, 64)
		if err != nil {
			return nil, err
		}
		*p = val
	}
	if from > to {
		return nil, db.ErrOutOfRange
	}
	at, keys, err := db.ScanExpiration(txn, from, to, count)
	if err != nil {
		return nil, err
	}
	n := len(keys)
	if n == 0 {
		return BytesArray(ctx.Out, nil), nil
	}
	// set the last ts as cursor
	cursor := fmt.Sprintf("%d", at[n-1])

	// set cursor to 0 if there is no more results
	if int64(n) < count {
		cursor = "0"
	}

	return func() {
		resp.ReplyArray(ctx.Out, 2)
		resp.ReplyBulkString(ctx.Out, cursor)
		resp.ReplyArray(ctx.Out, n)
		for i := 0; i < n; i++ {
			resp.ReplyBulkString(ctx.Out, fmt.Sprintf("%d %s", at[i], keys[i]))
		}
	}, nil
}
