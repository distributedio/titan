package db

import (
	"bytes"

	"github.com/golang/protobuf/proto"
	pb "gitlab.meitu.com/platform/titan/db/zlistproto"
)

// GetZList generate List objectm with auto reation, if zip is true, zipped list will be choose
func GetZList(txn *Transaction, metaKey []byte, obj *Object, val []byte) (*ZList, error) {
	l := &ZList{
		rawMetaKey: metaKey,
		txn:        txn,
	}
	if err := l.Unmarshal(obj, val); err != nil {
		return nil, err
	}
	return l, nil
}

//NewZList create new list object ,the key is not checked for presence
func NewZList(txn *Transaction, key []byte) (List, error) {
	metaKey := MetaKey(txn.db, key)
	ts := Now()
	obj := Object{
		ExpireAt:  0,
		CreatedAt: ts,
		UpdatedAt: ts,
		Type:      ObjectList,
		ID:        UUID(),
		Encoding:  ObjectEncodingZiplist,
	}
	l := &ZList{
		Object:     obj,
		value:      pb.Zlistvalue{},
		rawMetaKey: metaKey,
		txn:        txn,
	}
	b, err := l.Marshal()
	if err != nil {
		return nil, err
	}
	return l, txn.t.Set(metaKey, b)
}

// ZList ZListMeta defined zip list, with only objectMeta info.
type ZList struct {
	Object
	rawMetaKey []byte
	value      pb.Zlistvalue //[][]byte
	txn        *Transaction
}

//Exist if zlist is  effective return true ,otherwise return false
func (l *ZList) Exist() bool {
	if l.value.V == nil {
		return false
	}
	return true
}

// Marshal encode zlist into byte slice
func (l *ZList) Marshal() ([]byte, error) {
	b := EncodeObject(&l.Object)
	meta, err := proto.Marshal(&l.value)
	if err != nil {
		return nil, err
	}
	return append(b, meta...), nil
}

// zlistCommit try to marshal zlist values and then do set
func (l *ZList) zlistCommit() error {
	b, err := l.Marshal()
	if err != nil {
		return err
	}
	return l.txn.t.Set(l.rawMetaKey, b)
}

// Unmarshal parse meta data into meta field
func (l *ZList) Unmarshal(obj *Object, b []byte) (err error) {
	l.Object = *obj
	if err := proto.Unmarshal(b[ObjectEncodingLength:], &l.value); err != nil {
		return err
	}
	return nil
}

// Length return z list length
func (l *ZList) Length() int64 { return int64(len(l.value.V)) }

//LPush append new elements to the object values
func (l *ZList) LPush(data ...[]byte) (err error) {
	cv := make([][]byte, len(data), len(data)+len(l.value.V))

	j := 0 // data->[] lpush
	for i := len(data) - 1; i >= 0; i-- {
		cv[j] = data[i]
		j++
	}
	cv = append(cv, l.value.V...)
	l.value.V = cv
	return l.zlistCommit()
}

// RPush insert data befroe object values
func (l *ZList) RPush(data ...[]byte) (err error) {
	l.value.V = append(l.value.V, data...) // []<-data rpush
	return l.zlistCommit()
}

// Set the index object with given value, return ErrIndex on out of range error.
func (l *ZList) Set(n int64, data []byte) error {
	if n < 0 {
		n = int64(len(l.value.V)) + n
	}
	if n < 0 || n >= int64(len(l.value.V)) {
		return ErrOutOfRange
	}
	l.value.V[n] = data
	return l.zlistCommit()
}

// Insert v before/after pivot in zlist
func (l *ZList) Insert(pivot, v []byte, before bool) error {
	index := -1
	for index = range l.value.V {
		if bytes.Equal(l.value.V[index], pivot) {
			break
		}
	}
	// if pivot not exist, index will reach len(l.valus.V)
	if index == len(l.value.V) {
		return ErrKeyNotFound
	}

	if !before {
		index++
	}

	cv := make([][]byte, len(l.value.V)+1, len(l.value.V)+1)
	copy(cv[:index], l.value.V[:index])
	cv[index] = v
	copy(cv[index+1:], l.value.V[index:])

	l.value.V = cv
	return l.zlistCommit()
}

//Index return the value at index
func (l *ZList) Index(n int64) (data []byte, err error) {
	if n < 0 {
		n += int64(len(l.value.V))
	}
	if n < 0 || n >= int64(len(l.value.V)) {
		return nil, ErrOutOfRange
	}
	return l.value.V[n], nil
}

// LPop return and delete the left most element
func (l *ZList) LPop() (data []byte, err error) {
	v := l.value.V[0]
	l.value.V = l.value.V[1:]

	//destory on last key
	if len(l.value.V) == 0 {
		return v, l.Destory()
	}
	return v, l.zlistCommit()
}

// RPop return and delete the right most element
func (l *ZList) RPop() ([]byte, error) {
	v := l.value.V[len(l.value.V)-1]
	l.value.V = l.value.V[:len(l.value.V)-1]
	//destory on last key
	if len(l.value.V) == 0 {
		return v, l.Destory()
	}
	return v, l.zlistCommit()
}

// Range return the elements in [left, right]
func (l *ZList) Range(left, right int64) (value [][]byte, err error) {
	if right < 0 {
		if right = int64(len(l.value.V)) + right; right < 0 {
			return [][]byte{}, nil
		}
	}
	if left < 0 {
		if left = int64(len(l.value.V)) + left; left < 0 {
			left = 0
		}
	}
	if right >= int64(len(l.value.V)) {
		right = int64(len(l.value.V) - 1)
	}
	// return 0 elements
	if left > right {
		return [][]byte{}, nil
	}
	return l.value.V[left : right+1], nil
}

// LTrim get keys from start index to stop index
func (l *ZList) LTrim(start int64, stop int64) error {
	if start < 0 {
		if start = int64(len(l.value.V)) + start; start < 0 {
			start = 0
		}
	}
	if stop < 0 {
		stop += int64(len(l.value.V))
	}
	if stop > int64(len(l.value.V)-1) {
		stop = int64(len(l.value.V) - 1)
	}

	if start > stop {
		return l.Destory()
	}

	l.value.V = l.value.V[start : stop+1]
	return l.zlistCommit()
}

// LRem begin delete the count of n key from v
func (l *ZList) LRem(v []byte, n int64) (int, error) {
	cv := make([][]byte, len(l.value.V), len(l.value.V))
	count := 0
	if n < 0 { // delete from tail to head, then trim the cv
		j := len(l.value.V) - 1
		for i := j; i >= 0 && count < int(n); i-- {
			if !bytes.Equal(l.value.V[i], v) {
				cv[j] = l.value.V[i]
				j--
			} else {
				count++
			}
		}
		l.value.V = cv[j+1:]
	} else if n == 0 {
		j := 0
		for i := range l.value.V {
			if !bytes.Equal(l.value.V[i], v) {
				cv[j] = l.value.V[i]
				j++
			} else {
				count++
			}
		}
		l.value.V = cv[:j]
	} else {
		j := 0
		for i := j; i < len(l.value.V) && count < int(n); i-- {
			if !bytes.Equal(l.value.V[i], v) {
				cv[j] = l.value.V[i]
				j++
			} else {
				count++
			}
		}
		l.value.V = cv[:j]
	}
	return count, l.zlistCommit()
}

// Destory the zlist
func (l *ZList) Destory() error {
	// delete the meta data
	return l.txn.t.Delete(l.rawMetaKey)
}

// TransferToLList create an llist and put values into llist from zlist, LList will inheritance
// information from ZList
func (l *ZList) TransferToLList(dbns []byte, dbid byte, key []byte) (*LList, error) {
	ll := &LList{
		LListMeta: LListMeta{
			Object: Object{
				ExpireAt:  0,
				CreatedAt: l.CreatedAt,
				UpdatedAt: l.UpdatedAt,
				Type:      ObjectList,
				ID:        l.ID,
				Encoding:  ObjectEncodingLinkedlist,
			},
			Len:    0,
			Lindex: 0,
			Rindex: 0,
		},
		txn:        l.txn,
		rawMetaKey: l.rawMetaKey,
	}
	dataKeyPrefix := []byte{}
	dataKeyPrefix = append(dataKeyPrefix, dbns...)
	dataKeyPrefix = append(dataKeyPrefix, ':', dbid, ':', 'D', ':')
	dataKeyPrefix = append(dataKeyPrefix, ll.Object.ID...)
	ll.rawDataKeyPrefix = append(dataKeyPrefix, ':')
	return ll, ll.RPush(l.value.V...)
}
