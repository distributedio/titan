package command

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/meitu/titan/context"
	"github.com/meitu/titan/db"
	"github.com/stretchr/testify/assert"
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
	// t.Log(out.String())
	if strings.Index(out.String(), "ERR") == 0 {
		t.Fail()
	}
}

func TestMonitor(t *testing.T) {
	assert := assert.New(t)
	cli := &context.ClientContext{
		Namespace:  "$unittest",
		ID:         1,
		RemoteAddr: "127.0.0.1",
		DB:         &db.DB{Namespace: "$unittest", ID: 0},
	}
	serv := &context.ServerContext{}

	out := bytes.NewBuffer(nil)
	ctx := &Context{
		Name:    "monitor",
		Args:    nil,
		In:      nil,
		Out:     out,
		Context: context.New(cli, serv),
	}

	Monitor(ctx)
	assert.Equal("+OK\r\n", out.String())

	out.Reset()
	ctx = &Context{
		Name:    "ping",
		Args:    nil,
		In:      nil,
		Out:     ioutil.Discard,
		Context: ctx.Context,
	}
	feedMonitors(ctx)
	assert.Contains(out.String(), "[0 127.0.0.1] ping \r\n")
}

func TestClient_List(t *testing.T) {
	assert := assert.New(t)
	now := time.Now()
	cli := &context.ClientContext{
		Namespace:  "$unittest",
		ID:         1,
		RemoteAddr: "127.0.0.1",
		DB:         &db.DB{Namespace: "$unittest", ID: 0},
		Created:    now,
		Updated:    now,
	}
	serv := &context.ServerContext{}
	serv.Clients.Store(cli.RemoteAddr, cli)

	out := bytes.NewBuffer(nil)
	ctx := &Context{
		Name:    "client",
		Args:    []string{"list"},
		In:      nil,
		Out:     out,
		Context: context.New(cli, serv),
	}

	Client(ctx)

	assert.Contains(out.String(), "id=1 addr=127.0.0.1")
}
