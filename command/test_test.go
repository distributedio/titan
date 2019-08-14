package command

import (
	"bytes"
	"io"
	"strings"

	"github.com/distributedio/titan/conf"
	"github.com/distributedio/titan/context"
	"github.com/distributedio/titan/db"
)

var Cfg = &conf.MockConf().Tikv
var mockdb *db.RedisStore

func init() {
	mockdb, _ = db.Open(Cfg)
}

func ContextTest(name string, args ...string) *Context {
	cliCtx := &context.ClientContext{
		DB: mockdb.DB("defalut", 1),
	}
	servCtx := &context.ServerContext{
		RequirePass:      "",
		ListZipThreshold: 100,
		Store:            mockdb,
	}
	rootCtx, _ := context.WithCancel(context.New(cliCtx, servCtx))
	return &Context{
		Name:    name,
		Args:    args,
		In:      &bytes.Buffer{},
		Out:     &bytes.Buffer{},
		Context: rootCtx,
	}
}

func CallTest(name string, args ...string) *bytes.Buffer {
	ctx := ContextTest(name, args...)
	Call(ctx)
	return ctx.Out.(*bytes.Buffer)
}

func ctxString(buf io.Writer) string {
	return buf.(*bytes.Buffer).String()
}

func ctxLines(buf io.Writer) []string {
	str := ctxString(buf)
	return strings.Split(str, "\r\n")
}
