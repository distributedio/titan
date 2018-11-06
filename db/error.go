package db

import (
	"errors"

	"gitlab.meitu.com/platform/titan/db/store"
)

var (
	//IsErrNotFound returns true if the key is not found, otherwise return false
	// IsErrNotFound    = store.IsErrNotFound
	KeyNotFoundError = errors.New("key is not exist")

	// IsRetryableError returns true if the error is temporary and can be retried
	IsRetryableError    = store.IsRetryableError
	ObjectTypeError     = errors.New("error object type")
	ObjectEncodingError = errors.New("error object encoding type")
	TxnNilError         = errors.New("txn is nil")

	// ErrTxnAlreadyBegin indicates that db is in a transaction now, you should not begin again
	ErrTxnAlreadyBegin = errors.New("transaction has already been begun")
	// ErrTxnNotBegin indicates that db is not in a transaction, and can not be commited or rollbacked
	ErrTxnNotBegin   = errors.New("transaction dose not begin")
	ErrOutOfRange    = errors.New("error index/offset out of range")
	ErrPrecision     = errors.New("error precision limitation, rebalance the index")
	ErrNotExist      = errors.New("error call EX function, key not exist")
	ErrInvalidLength = errors.New("error data length is invalid for unmarshaler")
	ErrInternal      = errors.New("error tikv internal")
	ErrStringType    = errors.New("error object of type is not string")
)
