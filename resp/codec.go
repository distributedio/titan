package resp

type Encoder interface {
	Error(s string) error
	SimpleString(s string) error
	BulkString(s string) error
	NullBulkString() error
	Integer(v int64) error
	Array(size int) error
}

type Decoder interface {
	Error() (string, error)
	SimpleString() (string, error)
	BulkString() (string, error)
	Integer() (int64, error)
	Array(each func([]byte)) (int, error)
}
