package store

import (
	//nolint:depguard
	"github.com/tikv/client-go/v2/testutils"
	//nolint:depguard
	"github.com/tikv/client-go/v2/tikv"
	//nolint:depguard
	"github.com/tikv/client-go/v2/txnkv"
)

// MockAddr default mock tikv addr
const MockAddr = "mocktikv://"

// MockOpen create fake tikv db
//
//nolint:ireturn,wrapcheck
func MockOpen(_ string) (Storage, error) {
	mockClient, cluster, pdClient, err := testutils.NewMockTiKV("", nil)
	if err != nil {
		return nil, err
	}

	// Bootstrap with single store to ensure default region exists
	testutils.BootstrapWithSingleStore(cluster)

	tikvStore, err := tikv.NewTestTiKVStore(mockClient, pdClient, nil, nil, 0)
	if err != nil {
		return nil, err
	}

	return &storageWrapper{Client: &txnkv.Client{KVStore: tikvStore}}, nil
}