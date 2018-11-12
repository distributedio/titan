package thanos

// Encoder defines the interface of a RESP encoder
type Encoder interface {
	Error(s string) error
	SimpleString(s string) error
	BulkString(s string) error
	NullBulkString() error
	Integer(v int64) error
	Array(size int) error
}

//Decoder defines the interface of a RESP decoder
type Decoder interface {
	Error() (string, error)
	SimpleString() (string, error)
	BulkString() (string, error)
	Integer() (int64, error)
	Array(each func([]byte)) (int, error)
}
