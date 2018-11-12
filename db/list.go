package db

var (
	// ListZipThreshould create the zlist type of list if count key > ListZipThreshould
	ListZipThreshould = 100
)

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

// GetList returns a List object
// if key is zip return zlist otherwise return llist
// if key is not found, return null object but return error nil
func GetList(txn *Transaction, key []byte) (List, error) {
	metaKey := MetaKey(txn.db, key)
	val, err := txn.t.Get(metaKey)
	if err != nil {
		if IsErrNotFound(err) { // error NotFound
			return &LList{}, nil
		}
		return nil, err
	}
	// exist
	obj, err := DecodeObject(val)
	if err != nil {
		return nil, err
	}
	if obj.Type != ObjectList {
		return nil, ErrTypeMismatch
	}

	if IsExpired(obj, Now()) {
		return &LList{}, nil
	}

	if obj.Encoding == ObjectEncodingLinkedlist {
		return GetLList(txn, metaKey, obj, val)
	} else if obj.Encoding == ObjectEncodingZiplist {
		return GetZList(txn, metaKey, obj, val)
	}
	return nil, ErrEncodingMismatch
}

//NewList create new list object
func NewList(txn *Transaction, key []byte, count int) (List, error) {
	if count < ListZipThreshould {
		return NewLList(txn, key)
	}
	return NewZList(txn, key)

}
