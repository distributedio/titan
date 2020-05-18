package etcdutil

import (
	"strings"

	"go.etcd.io/etcd/clientv3"
)

func NewClient(addrs string) (*clientv3.Client, error) {
	endpoints := strings.Split(addrs, ",")
	cli, err := clientv3.New(clientv3.Config{Endpoints: endpoints})
	if err != nil {
		return nil, err
	}
	return cli, nil
}
