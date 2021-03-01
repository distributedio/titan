package autotest

import (
	"os"
	"testing"
	"time"

	"github.com/distributedio/titan/tools/integration"
	etcd_int "go.etcd.io/etcd/integration"
)

var (
	at *AutoClient
	an *Abnormal
)

func TestMain(m *testing.M) {
	t := &testing.T{}
	clus := etcd_int.NewClusterV3(t, &etcd_int.ClusterConfig{Size: 1, ClientTLS: nil})
	etcdAddrs := clus.RandClient().Endpoints()
	integration.SetAuth("titan")
	integration.SetEtcdAddrs(etcdAddrs)

	go integration.Start()
	time.Sleep(time.Second)
	at = NewAutoClient()
	an = NewAbnormal()
	//TODO
	at.Start(integration.ServerAddr)
	an.Start(integration.ServerAddr)
	// Pool = newPool(integration.ServerAddr)
	v := m.Run()
	an.Close()
	integration.Close()
	at.Close()

	os.Exit(v)
}
