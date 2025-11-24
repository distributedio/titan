package command

import (
	"strconv"

	"github.com/distributedio/titan/encoding/resp"
	"github.com/distributedio/titan/metrics"
)

// Auth verifies the client
func Auth(ctx *Context) {
	args := ctx.Args
	serverauth := []byte(ctx.Server.RequirePass)
	if len(serverauth) == 0 {
		if err := resp.ReplyError(ctx.Out, "ERR Client sent AUTH, but no password is set"); err != nil {
			return
		}
		return
	}

	token := []byte(args[0])
	namespace, err := Verify(token, serverauth)
	if err != nil {
		if err := resp.ReplyError(ctx.Out, "ERR invalid password"); err != nil {
			return
		}
		return
	}
	metrics.GetMetrics().ConnectionOnlineGaugeVec.WithLabelValues(ctx.Client.Namespace).Dec()
	metrics.GetMetrics().ConnectionOnlineGaugeVec.WithLabelValues(string(namespace)).Inc()
	ctx.Client.Namespace = string(namespace)
	ctx.Client.DB.Namespace = string(namespace)
	ctx.Client.Authenticated = true
	_ = resp.ReplySimpleString(ctx.Out, OK)
}

// Echo the given string
func Echo(ctx *Context) {
	_ = resp.ReplyBulkString(ctx.Out, ctx.Args[0])
}

// Ping the server
func Ping(ctx *Context) {
	args := ctx.Args
	if len(args) > 0 {
		_ = resp.ReplyBulkString(ctx.Out, args[0])
		return
	}
	_ = resp.ReplySimpleString(ctx.Out, "PONG")
}

// Select the logical database
func Select(ctx *Context) {
	args := ctx.Args
	idx, err := strconv.Atoi(args[0])
	if err != nil {
		if err := resp.ReplyError(ctx.Out, "ERR invalid DB index"); err != nil {
			return
		}
		return
	}
	if idx < 0 || idx > 255 {
		if err := resp.ReplyError(ctx.Out, "ERR invalid DB index"); err != nil {
			return
		}
		return
	}
	namespace := ctx.Client.Namespace
	ctx.Client.DB = ctx.Server.Store.DB(namespace, idx)
	if err := resp.ReplySimpleString(ctx.Out, OK); err != nil {
		return
	}
}

// Quit asks the server to close the connection
func Quit(ctx *Context) {
	close(ctx.Client.Done)
	if err := resp.ReplySimpleString(ctx.Out, OK); err != nil {
		return
	}
}

// SwapDB swaps two Redis databases
func SwapDB(ctx *Context) {
	if err := resp.ReplyError(ctx.Out, "ERR not supported"); err != nil {
		return
	}
}
