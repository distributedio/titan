package command

import (
	"bytes"
	"io"
	"strings"

	"gitlab.meitu.com/platform/thanos/conf"
	"gitlab.meitu.com/platform/thanos/context"
	"gitlab.meitu.com/platform/thanos/db"
	"gitlab.meitu.com/platform/thanos/db/store"
)

var cfg = &conf.Tikv{
	PdAddrs: store.MockAddr,
}
var mockdb *db.RedisStore

func init() {
	mockdb, _ = db.Open(cfg)
}

func ContextTest(name string, args ...string) *Context {
	cliCtx := &context.ClientContext{
		DB: mockdb.DB("defalut", 1),
	}
	servCtx := &context.ServerContext{
		RequirePass: "",
		Store:       mockdb,
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
