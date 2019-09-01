package db

// List defines the list interface
type List interface {
	Index(n int64) (data []byte, err error)
	Insert(pivot, v []byte, before bool) error
	LPop() (data []byte, err error)
	LPush(data ...[]byte) (err error)
	RPop() (data []byte, err error)
	RPush(data ...[]byte) (err error)
	Range(left, right int64) (value [][]byte, err error)
	LRem(v []byte, n int64) (int, error)
	Set(n int64, data []byte) error
	LTrim(start int64, stop int64) error
	Length() int64
	Exist() bool
	Destory() error
}

// GetList returns a List object, it creates a new one if the key does not exist,
// when UseZip() is set, it will create a ziplist instead of a linklist
func GetList(txn *Transaction, key []byte, opts ...ListOption) (List, error) {
	opt := &listOption{}
	for _, o := range opts {
		if o == nil {
			continue
		}
		o(opt)
	}
	list := NewLList
	if opt.useZip {
		list = NewZList
	}

	metaKey := MetaKey(txn.db, key)
	val, err := txn.t.Get(metaKey)
	if err != nil {
		if IsErrNotFound(err) { // error NotFound
			return list(txn, key)
		}
		return nil, err
	}

	// exist
	obj, err := DecodeObject(val)
	if err != nil {
		return nil, err
	}
	if IsExpired(obj, Now()) {
		return list(txn, key)
	}

	if obj.Type != ObjectList {
		return nil, ErrTypeMismatch
	}

	if obj.Encoding == ObjectEncodingLinkedlist {
		return GetLList(txn, metaKey, obj, val)
	} else if obj.Encoding == ObjectEncodingZiplist {
		return GetZList(txn, metaKey, obj, val)
	}
	return nil, ErrEncodingMismatch
}
