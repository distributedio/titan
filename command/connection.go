package command

import (
	"strconv"

	"github.com/meitu/titan/encoding/resp"
	"github.com/meitu/titan/metrics"
)

// Auth verifies the client
func Auth(ctx *Context) {
	args := ctx.Args
	serverauth := []byte(ctx.Server.RequirePass)
	if len(serverauth) == 0 {
		resp.ReplyError(ctx.Out, "ERR Client sent AUTH, but no password is set")
		return
	}

	token := []byte(args[0])
	namespace, err := Verify(token, serverauth)
	if err != nil {
		resp.ReplyError(ctx.Out, "ERR invalid password")
		return
	}
	metrics.GetMetrics().ConnectionOnlineGaugeVec.WithLabelValues(ctx.Client.Namespace).Dec()
	metrics.GetMetrics().ConnectionOnlineGaugeVec.WithLabelValues(string(namespace)).Inc()
	ctx.Client.Namespace = string(namespace)
	ctx.Client.DB.Namespace = string(namespace)
	ctx.Client.Authenticated = true
	resp.ReplySimpleString(ctx.Out, OK)
}

// Echo the given string
func Echo(ctx *Context) {
	resp.ReplyBulkString(ctx.Out, ctx.Args[0])
}

// Ping the server
func Ping(ctx *Context) {
	args := ctx.Args
	if len(args) > 0 {
		resp.ReplyBulkString(ctx.Out, args[0])
		return
	}
	resp.ReplyBulkString(ctx.Out, "PONG")
}

// Select the logical database
func Select(ctx *Context) {
	args := ctx.Args
	idx, err := strconv.Atoi(args[0])
	if err != nil {
		resp.ReplyError(ctx.Out, "ERR invalid DB index")
		return
	}
	if idx < 0 || idx > 255 {
		resp.ReplyError(ctx.Out, "ERR invalid DB index")
		return
	}
	namespace := ctx.Client.Namespace
	ctx.Client.DB = ctx.Server.Store.DB(namespace, idx)
	resp.ReplySimpleString(ctx.Out, OK)
}

// Quit asks the server to close the connection
func Quit(ctx *Context) {
	close(ctx.Client.Done)
	resp.ReplySimpleString(ctx.Out, OK)
}

// SwapDB swaps two Redis databases
func SwapDB(ctx *Context) {
	resp.ReplyError(ctx.Out, "ERR not supported")
}
