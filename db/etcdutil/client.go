package etcdutil

import (
	"net/url"
	"strings"

	"github.com/distributedio/titan/db/store"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
)

var (
	etcdPrefix = []byte("/titan:")
)

func NewClient(addrs []string) (*clientv3.Client, error) {
	cli, err := clientv3.New(clientv3.Config{Endpoints: addrs})
	if err != nil {
		return nil, err
	}
	if logEnv := zap.L().Check(zap.DebugLevel, "new etcd client"); logEnv != nil {
		logEnv.Write(zap.Strings("endpoints", addrs))
	}
	return cli, nil
}

func PdAddrsToEtcd(pd string) []string {
	var etcdAddrs []string
	if strings.Contains(pd, store.MockAddr) {
		return etcdAddrs
	}
	u, err := url.Parse(pd)
	if err != nil {
		zap.L().Error("parase pd addrs to etcd err", zap.String("addrs", pd), zap.Error(err))
		return etcdAddrs
	}
	for _, v := range strings.Split(u.Host, ",") {
		etcdAddrs = append(etcdAddrs, "http://"+v)
	}
	return etcdAddrs
}
