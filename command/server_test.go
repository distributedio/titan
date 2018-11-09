package command

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"gitlab.meitu.com/platform/thanos/context"
)

func TestInfo(t *testing.T) {
	out := bytes.NewBuffer(nil)
	ctx := &Context{
		Name: "info",
		Args: nil,
		In:   nil,
		Out:  out,
		Context: context.New(&context.ClientContext{
			Namespace: "$unittest",
		}, &context.ServerContext{
			StartAt: time.Now(),
		}),
	}
	Info(ctx)
	t.Log(out.String())
	if strings.Index(out.String(), "ERR") == 0 {
		t.Fail()
	}
}
