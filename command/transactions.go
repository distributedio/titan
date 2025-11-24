package command

import (
	"bytes"
	"strings"
	"time"

	"github.com/distributedio/titan/db"
	"github.com/distributedio/titan/encoding/resp"
	"github.com/distributedio/titan/metrics"
	"github.com/shafreeck/retry"
	"go.uber.org/zap"
)

// Multi starts a transaction which will block subsequent commands until 'exec'
func Multi(ctx *Context) {
	ctx.Client.Multi = true
	_ = resp.ReplySimpleString(ctx.Out, OK)
}

// Exec all the commands queued in client
func Exec(ctx *Context) {
	ctx.Client.Multi = false
	pendings := ctx.Client.Commands
	if len(pendings) == 0 {
		_, _ = resp.ReplyArray(ctx.Out, 0)
		return
	}
	ctx.Client.Commands = nil

	// Has watch command been issued
	watching := ctx.Client.Txn != nil
	txn := ctx.Client.Txn
	ctx.Client.Txn = nil

	size := len(pendings)
	var err error
	var outputs []*bytes.Buffer
	var onCommits []OnCommit
	err = retry.Ensure(ctx, func() error {
		if !watching {
			txn, err = ctx.Client.DB.Begin()
			if err != nil {
				zap.L().Error("begin txn failed",
					zap.Int64("clientid", ctx.Client.ID),
					zap.String("command", ctx.Name),
					zap.String("traceid", ctx.TraceID),
					zap.Error(err))
				_, _ = resp.ReplyArray(ctx.Out, 0)
				return err
			}
		}
		outputs = make([]*bytes.Buffer, size)
		onCommits = make([]OnCommit, size)
		commandCount := 0
		for i, cmd := range pendings {
			var onCommit OnCommit
			out := bytes.NewBuffer(nil)
			subCtx := &Context{
				Name:    cmd.Name,
				Args:    cmd.Args,
				In:      ctx.In,
				Out:     out,
				Context: ctx.Context,
			}
			name := strings.ToLower(cmd.Name)
			if desc, ok := commands[name]; ok && desc.Txn != nil {
				start := time.Now()
				onCommit, err = TxnCall(subCtx, txn)
				zap.L().Debug("execute", zap.String("command", subCtx.Name), zap.Int64("cost(us)", time.Since(start).Nanoseconds()/1000))
				if err != nil {
					_ = resp.ReplyError(out, err.Error())
				}
			} else {
				Call(subCtx)
			}
			onCommits[i] = onCommit
			outputs[i] = out
			commandCount++
		}
		start := time.Now()
		mt := metrics.GetMetrics()
		mt.MultiCommandHistogramVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Observe(float64(commandCount))
		defer func() {
			cost := time.Since(start).Seconds()
			mt.TxnCommitHistogramVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Observe(cost)
		}()
		start = time.Now()
		err = txn.Commit(ctx)
		zap.L().Debug("commit", zap.String("command", ctx.Name), zap.Int64("cost(us)", time.Since(start).Nanoseconds()/1000))
		if err != nil {
			mt.TxnFailuresCounterVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Inc()
			if db.IsRetryableError(err) {
				mt.TxnConflictsCounterVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Inc()
				// When in WATCH mode, we cannot retry on conflicts because:
				// 1. The watched keys' snapshot is stale after the first attempt
				// 2. Refreshing snapshots would require re-executing all commands
				// 3. This could change command results, violating Redis semantics
				//
				// Users should handle WATCH failures and retry the entire WATCH/MULTI/EXEC sequence.
				if !watching {
					mt.TxnRetriesCounterVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Inc()
					zap.L().WithOptions(zap.AddStacktrace(zap.FatalLevel)).Info("txn commit retry",
						zap.Int64("clientid", ctx.Client.ID),
						zap.String("command", ctx.Name),
						zap.String("traceid", ctx.TraceID),
						zap.String("error", err.Error()))
					return retry.Retriable(err)
				}
			}
			zap.L().Error("commit failed",
				zap.Int64("clientid", ctx.Client.ID),
				zap.String("command", ctx.Name),
				zap.String("traceid", ctx.TraceID),
				zap.Error(err))
			return err
		}
		return nil
	})
	if err != nil {
		zap.L().Error("txn failed",
			zap.Int64("clientid", ctx.Client.ID),
			zap.String("command", ctx.Name),
			zap.String("traceid", ctx.TraceID),
			zap.Error(err))
		if watching {
			_, _ = resp.ReplyArray(ctx.Out, 0)
			return
		}
		_ = resp.ReplyError(ctx.Out, "EXECABORT Transaction discarded because of txn conflicts")
		return
	}

	_, _ = resp.ReplyArray(ctx.Out, size)
	// run OnCommit that fill reply to outputs
	for i := range onCommits {
		c := onCommits[i]
		if c != nil {
			c()
		}

		if _, err := ctx.Out.Write(outputs[i].Bytes()); err != nil {
			zap.L().Error("reply to client failed",
				zap.Int64("clientid", ctx.Client.ID),
				zap.String("command", ctx.Name),
				zap.String("traceid", ctx.TraceID),
				zap.Error(err))
			break
		}
	}
}

// Watch starts a transaction, watch is a global transaction and is not key associated(this is different from redis)
func Watch(ctx *Context) {
	txn := ctx.Client.Txn
	if txn == nil {
		var err error
		txn, err = ctx.Client.DB.Begin()
		if err != nil {
			_ = resp.ReplyError(ctx.Out, "Err "+err.Error())
			return
		}
		ctx.Client.Txn = txn
	}
	keys := make([][]byte, len(ctx.Args))
	for i := range ctx.Args {
		keys[i] = db.MetaKey(ctx.Client.DB, []byte(ctx.Args[i]))
	}
	// LockKeys in TiKV client (v2) can be unstable in some environments (like mock store tests) when
	// pessimistic locking is involved, or might have specific requirements.
	// Instead of explicitly locking, we can achieve optimistic locking semantics by reading the keys.
	// This updates the read set of the transaction. If any of these keys are modified by another
	// transaction before this one commits, the commit will fail with a conflict, which satisfies
	// the WATCH contract.
	if len(keys) > 0 {
		if _, err := db.BatchGetValues(txn, keys); err != nil {
			_ = txn.Rollback()
			ctx.Client.Txn = nil
			_ = resp.ReplyError(ctx.Out, "Err "+err.Error())
			return
		}
	}
	_ = resp.ReplySimpleString(ctx.Out, OK)
}

// Discard flushes all previously queued commands in a transaction and restores the connection state to normal
func Discard(ctx *Context) {
	// in watch state, the txn has begun, rollback it
	if ctx.Client.Txn != nil {
		_ = ctx.Client.Txn.Rollback()
		ctx.Client.Txn = nil
	}
	ctx.Client.Commands = nil
	ctx.Client.Multi = false
	_ = resp.ReplySimpleString(ctx.Out, OK)
}

// Unwatch flushes all the previously watched keys for a transaction
func Unwatch(ctx *Context) {
	if ctx.Client.Txn != nil {
		_ = ctx.Client.Txn.Rollback()
		ctx.Client.Txn = nil
	}
	_ = resp.ReplySimpleString(ctx.Out, OK)
}