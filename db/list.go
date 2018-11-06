package db

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
)

var ErrListEmpty = errors.New("list is empty")

type ListMeta struct {
	Object
	Lindex float64
	Rindex float64
	Len    int64
}

// List is a distributed object that works like a double link list
type List struct {
	meta ListMeta
	key  []byte
	txn  *Transaction
}

// GetList return the list of key, create new one if it is not exist.
func GetList(txn *Transaction, key []byte) (*List, error) {
	lst := &List{
		key: key,
		txn: txn,
	}
	mkey := MetaKey(txn.db, key)
	meta, err := txn.txn.Get(mkey)
	if err != nil {
		if IsErrNotFound(err) {
			now := Now()
			lst.meta.CreatedAt = now
			lst.meta.UpdatedAt = now
			lst.meta.ExpireAt = 0
			lst.meta.ID = UUID()
			lst.meta.Type = ObjectList
			lst.meta.Encoding = ObjectEncodingLinkedlist
			lst.meta.Lindex = 0
			lst.meta.Rindex = 0
			lst.meta.Len = 0
			return lst, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(meta, &lst.meta); err != nil {
		return nil, err
	}
	if lst.meta.Type != ObjectList {
		return nil, ErrTypeMismatch
	}
	return lst, nil
}

func listItemKey(key []byte, idx float64) []byte {
	key = append(key, ':')
	return append(key, EncodeFloat64(idx)...)
}

// LPush add new element to the left
func (lst *List) LPush(data []byte) error {
	lst.meta.Len++
	lst.meta.Lindex--
	lst.meta.UpdatedAt = Now()

	//data key
	dkey := DataKey(lst.txn.db, lst.meta.ID)

	// item key
	ikey := listItemKey(dkey, lst.meta.Lindex)

	if err := lst.txn.txn.Set(ikey, data); err != nil {
		return nil
	}
	return lst.updateMeta()
}

// LPop return and delete the left most element
func (lst *List) LPop() ([]byte, error) {
	if lst.meta.Len == 0 {
		if err := lst.Destory(); err != nil {
			return nil, err
		}
		return nil, ErrListEmpty
	}

	lst.meta.Len--
	lst.meta.UpdatedAt = Now()

	//data key
	dkey := DataKey(lst.txn.db, lst.meta.ID)

	// item key
	ikey := listItemKey(dkey, lst.meta.Lindex)

	// seek this key
	iter, err := lst.txn.txn.Seek(ikey)
	if err != nil {
		return nil, err
	}
	val := iter.Value()

	if err := lst.txn.txn.Delete(ikey); err != nil {
		return nil, err
	}

	// find next item
	if err := iter.Next(); err != nil {
		return nil, err
	}

	if iter.Key().HasPrefix(dkey) && lst.meta.Len > 0 {
		// Update Lindex to next
		lst.meta.Lindex = lst.index(iter.Key())
	} else {
		if err := lst.Destory(); err != nil {
			return nil, err
		}
		return val, nil
	}

	if err := lst.updateMeta(); err != nil && lst.meta.Len == 0 {
		return nil, err
	}
	return val, nil
}

// LRange returns the specified elements of the list stored at key
func (lst *List) LRange(start int, stop int) ([][]byte, error) {
	if start < 0 {
		start = int(lst.meta.Len) + start
		if start < 0 {
			start = 0
		}
	}
	if stop < 0 {
		stop = int(lst.meta.Len) + stop
		if start < 0 {
			start = 0
		}
	}
	if start > int(lst.meta.Len) {
		return nil, nil
	}
	if stop >= int(lst.meta.Len) {
		stop = int(lst.meta.Len) - 1
	}
	if start > stop {
		return nil, nil
	}

	dkey := DataKey(lst.txn.db, lst.meta.ID)
	leftKey := listItemKey(dkey, lst.meta.Lindex)

	// seek to start key
	iter, err := lst.txn.txn.Seek(leftKey)
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	idx := start
	for iter.Valid() && idx != 0 {
		idx--
		if err := iter.Next(); err != nil {
			return nil, err
		}
	}
	var items [][]byte
	items = append(items, iter.Value())

	count := stop - start
	for iter.Valid() && count != 0 {
		count--
		if err := iter.Next(); err != nil {
			return nil, err
		}
		items = append(items, iter.Value())
	}
	return items, nil
}

// LInsert inserts value in the list stored at key either before or after the reference value pivot
func (lst *List) LInsert(after bool, pivot []byte, val []byte) (int, error) {
	dkey := DataKey(lst.txn.db, lst.meta.ID)
	leftKey := listItemKey(dkey, lst.meta.Lindex)

	prev := lst.meta.Lindex
	next := lst.meta.Lindex
	// seek to start key
	iter, err := lst.txn.txn.Seek(leftKey)
	if err != nil {
		return -1, err
	}
	defer iter.Close()

	found := false
	count := lst.meta.Len
	for iter.Valid() && count != 0 {
		key := iter.Key()
		prev = next
		next = lst.index(key)
		log.Println(prev, next)
		if bytes.Compare(iter.Value(), pivot) == 0 {
			found = true
			break
		}
		if err := iter.Next(); err != nil {
			return -1, err
		}
		count--
	}
	if !found {
		return -1, ErrKeyNotFound
	}

	if after {
		prev = next
		if err := iter.Next(); err != nil {
		}
		count--
		if count != 0 {
			next = lst.index(iter.Key())
		}
	}

	idx := (prev + next) / 2
	if idx == prev || idx == next {
		return -1, ErrFullSlot
	}

	// only have on element here
	if prev == next {
		if after {
			idx = lst.meta.Rindex
			lst.meta.Rindex += 1
		} else {
			idx = lst.meta.Lindex - 1
			lst.meta.Lindex -= 1
			log.Println(lst.meta.Lindex)
		}
	}
	// item key
	ikey := listItemKey(dkey, idx)

	if err := lst.txn.txn.Set(ikey, val); err != nil {
		return -1, err
	}
	lst.meta.Len += 1
	lst.meta.UpdatedAt = Now()
	return int(lst.meta.Len), lst.updateMeta()
}

func (lst *List) updateMeta() error {
	meta, err := json.Marshal(lst.meta)
	if err != nil {
		return err
	}
	return lst.txn.txn.Set(MetaKey(lst.txn.db, lst.key), meta)
}

func (lst *List) index(key []byte) float64 {
	dkey := DataKey(lst.txn.db, lst.meta.ID)
	return DecodeFloat64(key[len(dkey)+1:])
}

// Destory the list
func (lst *List) Destory() error {
	return lst.txn.Destory(&lst.meta.Object, lst.key)
}
