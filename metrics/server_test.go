package metrics

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.meitu.com/platform/thanos/conf"
)

var (
	cstatus = &conf.Status{
		Listen: ":32345",
	}
)

func TestListenAndServer(t *testing.T) {
	server := NewServer(cstatus)
	assert.NotNil(t, server)
	go server.ListenAndServe(cstatus.Listen)
	err := server.Stop()
	assert.NoError(t, err)
}

func TestServer(t *testing.T) {
	server := NewServer(cstatus)
	assert.NotNil(t, server)
	lis, err := net.Listen("tcp", cstatus.Listen)
	assert.NoError(t, err)
	go server.Serve(lis)
	err = server.GracefulStop()
	assert.NoError(t, err)
}
