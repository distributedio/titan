package store

import "github.com/pingcap/tidb/store/mockstore"

var MockAddr = "mocktikv://"

func MockOpen(addrs string) (r Storage, e error) {
	var driver mockstore.MockDriver
	return driver.Open(MockAddr)
}
