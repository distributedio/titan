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
func GetList(txn *Transaction, key []byte, count int) (List, error) {
	var lst List
	if count < ListZipThreshould {
		lst = NewLList(txn, key)
	} else {
		lst = NewZList(txn, key)
	}

	metaKey := MetaKey(txn.db, key)
	val, err := txn.t.Get(metaKey)
	if err != nil {
		if IsErrNotFound(err) { // error NotFound
			return lst, nil
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
		return lst, nil
	}

	if obj.Encoding == ObjectEncodingLinkedlist {
		return GetLList(txn, metaKey, obj, val)
	} else if obj.Encoding == ObjectEncodingZiplist {
		return GetZList(txn, metaKey, obj, val)
	}
	return nil, ErrEncodingMismatch
}
