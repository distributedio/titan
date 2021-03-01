package store

import "github.com/pingcap/tidb/store/mockstore"

//MockAddr default mock tikv addr
var MockAddr = "mocktikv://"

// MockOpen create fake tikv db
func MockOpen(addrs string) (r Storage, e error) {
	var driver mockstore.MockTiKVDriver
	return driver.Open(MockAddr)
}
