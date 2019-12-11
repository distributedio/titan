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
	resp.ReplySimpleString(ctx.Out, OK)
}

// Exec all the commands queued in client
func Exec(ctx *Context) {
	ctx.Client.Multi = false
	commands := ctx.Client.Commands
	if len(commands) == 0 {
		resp.ReplyArray(ctx.Out, 0)
		return
	}
	ctx.Client.Commands = nil

	// Has watch command been issued
	watching := ctx.Client.Txn != nil
	txn := ctx.Client.Txn
	ctx.Client.Txn = nil

	size := len(commands)
	var err error
	var outputs []*bytes.Buffer
	var onCommits []OnCommit
	err = retry.Ensure(ctx, func() error {
		mt := metrics.GetMetrics()
		if !watching {
			start := time.Now()
			txn, err = ctx.Client.DB.Begin()
			cost := time.Since(start).Seconds()
			mt.TxnBeginHistogramVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Observe(cost)
			zap.L().Debug("transation begin", zap.String("name", ctx.Name), zap.Int64("cost(us)", int64(cost*1000000)))
			if err != nil {
				mt.TxnFailuresCounterVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Inc()
				zap.L().Error("begin txn failed",
					zap.Int64("clientid", ctx.Client.ID),
					zap.String("command", ctx.Name),
					zap.String("traceid", ctx.TraceID),
					zap.Error(err))
				resp.ReplyArray(ctx.Out, 0)
				return err
			}
		}
		outputs = make([]*bytes.Buffer, size)
		onCommits = make([]OnCommit, size)
		commandCount := 0
		for i, cmd := range commands {
			var onCommit OnCommit
			out := bytes.NewBuffer(nil)
			subCtx := &Context{
				Name:    cmd.Name,
				Args:    cmd.Args,
				In:      ctx.In,
				Out:     out,
				Context: ctx.Context,
			}
			if len(cmd.Args) > 0 {
				mt.CommandArgsNumHistogramVec.WithLabelValues(ctx.Client.Namespace, cmd.Name).Observe(float64(len(cmd.Args)))
			}
			name := strings.ToLower(cmd.Name)
			if _, ok := txnCommands[name]; ok {
				start := time.Now()
				onCommit, err = TxnCall(subCtx, txn)
				cost := time.Since(start).Seconds()
				mt.CommandFuncDoneHistogramVec.WithLabelValues(ctx.Client.Namespace, cmd.Name).Observe(cost)
				zap.L().Debug("execute", zap.String("command", cmd.Name), zap.Int64("cost(us)", int64(cost*1000000)))
				if err != nil {
					resp.ReplyError(out, err.Error())
				}
			} else {
				Call(subCtx)
			}
			onCommits[i] = onCommit
			outputs[i] = out
			commandCount++
		}
		start := time.Now()
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
			if db.IsRetryableError(err) && !watching {
				mt.TxnRetriesCounterVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Inc()
				mt.TxnConflictsCounterVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Inc()
				zap.L().Error("txn commit retry",
					zap.Int64("clientid", ctx.Client.ID),
					zap.String("command", ctx.Name),
					zap.String("traceid", ctx.TraceID),
					zap.Error(err))
				return retry.Retriable(err)
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
			resp.ReplyArray(ctx.Out, 0)
			return
		}
		resp.ReplyError(ctx.Out, "EXECABORT Transaction discarded because of txn conflicts")
		return
	}

	start := time.Now()
	resp.ReplyArray(ctx.Out, size)
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
	cost := time.Since(start).Seconds()
	metrics.GetMetrics().ReplyFuncDoneHistogramVec.WithLabelValues(ctx.Client.Namespace, ctx.Name).Observe(cost)
	zap.L().Debug("onCommit ", zap.String("name", ctx.Name), zap.Int64("cost(us)", int64(cost*1000000)))
}

// Watch starts a transaction, watch is a global transaction and is not key associated(this is different from redis)
func Watch(ctx *Context) {
	txn, err := ctx.Client.DB.Begin()
	if err != nil {
		resp.ReplyError(ctx.Out, "Err "+err.Error())
		return
	}
	keys := make([][]byte, len(ctx.Args))
	for i := range ctx.Args {
		keys[i] = []byte(ctx.Args[i])
	}
	if err := txn.LockKeys(keys...); err != nil {
		txn.Rollback()
		resp.ReplyError(ctx.Out, "Err "+err.Error())
		return
	}
	ctx.Client.Txn = txn
	resp.ReplySimpleString(ctx.Out, OK)
}

// Discard flushes all previously queued commands in a transaction and restores the connection state to normal
func Discard(ctx *Context) {
	// in watch state, the txn has begun, rollback it
	if ctx.Client.Txn != nil {
		ctx.Client.Txn.Rollback()
		ctx.Client.Txn = nil
	}
	ctx.Client.Commands = nil
	ctx.Client.Multi = false
	resp.ReplySimpleString(ctx.Out, OK)
}

// Unwatch flushes all the previously watched keys for a transaction
func Unwatch(ctx *Context) {
	if ctx.Client.Txn != nil {
		ctx.Client.Txn.Rollback()
		ctx.Client.Txn = nil
	}
	resp.ReplySimpleString(ctx.Out, OK)
}
