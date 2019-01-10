package db

import (
	"bytes"
	"encoding/binary"
	"math"
	"time"

	"github.com/meitu/titan/metrics"

	"github.com/pingcap/tidb/kv"
)

// LListMeta keeps all meta info of a list object
// after marshaled, listmeta raw data should organized like below
// |----object 34----|
// |------len 8------|
// |----linedx 8-----|
// |----rindex 8-----|
type LListMeta struct {
	Object
	Len    int64
	Lindex float64
	Rindex float64
}

// LList is a distributed object that works like a double link list
// {Object schema}:{index} -> value
type LList struct {
	LListMeta
	rawMetaKey       []byte
	rawDataKeyPrefix []byte
	txn              *Transaction
}

//GetLList returns a list
func GetLList(txn *Transaction, metaKey []byte, obj *Object, val []byte) (List, error) {
	l := &LList{
		txn:        txn,
		rawMetaKey: metaKey,
	}
	if err := l.LListMeta.Unmarshal(obj, val); err != nil {
		return nil, err
	}
	l.rawDataKeyPrefix = DataKey(txn.db, l.Object.ID)
	l.rawDataKeyPrefix = append(l.rawDataKeyPrefix, []byte(Separator)...)
	return l, nil
}

//NewLList creates a new list
func NewLList(txn *Transaction, key []byte) List {
	now := Now()
	metaKey := MetaKey(txn.db, key)
	obj := Object{
		ExpireAt:  0,
		CreatedAt: now,
		UpdatedAt: now,
		Type:      ObjectList,
		ID:        UUID(),
		Encoding:  ObjectEncodingLinkedlist,
	}
	l := &LList{
		LListMeta: LListMeta{
			Object: obj,
			Len:    0,
			Lindex: 0,
			Rindex: 0,
		},
		txn:        txn,
		rawMetaKey: metaKey,
	}
	l.rawDataKeyPrefix = DataKey(txn.db, l.Object.ID)
	l.rawDataKeyPrefix = append(l.rawDataKeyPrefix, []byte(Separator)...)
	return l

}

// Marshal encodes meta data into byte slice
func (l *LListMeta) Marshal() []byte {
	b := EncodeObject(&l.Object)
	meta := make([]byte, 32)
	binary.BigEndian.PutUint64(meta[0:8], uint64(l.Len))
	binary.BigEndian.PutUint64(meta[8:16], math.Float64bits(l.Lindex))
	binary.BigEndian.PutUint64(meta[16:32], math.Float64bits(l.Rindex))
	return append(b, meta...)
}

// Unmarshal parses meta data into meta field
func (l *LListMeta) Unmarshal(obj *Object, b []byte) (err error) {
	if len(b[ObjectEncodingLength:]) != 32 {
		return ErrInvalidLength
	}
	meta := b[ObjectEncodingLength:]
	l.Object = *obj
	l.Len = int64(binary.BigEndian.Uint64(meta[:8]))
	l.Lindex = math.Float64frombits(binary.BigEndian.Uint64(meta[8:16]))
	l.Rindex = math.Float64frombits(binary.BigEndian.Uint64(meta[16:32]))
	return nil
}

// Length returns length of the list
func (l *LList) Length() int64 { return l.LListMeta.Len }

// LPush adds new elements to the left
// 1. calculate index
// 2. encode object and call kv
// 3. modify the new index in meta
func (l *LList) LPush(data ...[]byte) (err error) {
	for i := range data {
		l.Lindex--
		if err = l.txn.t.Set(append(l.rawDataKeyPrefix, EncodeFloat64(l.Lindex)...), data[i]); err != nil {
			return err
		}
		l.Len++
		if l.Len == 1 {
			l.Rindex = l.Lindex
		}
	}
	return l.txn.t.Set(l.rawMetaKey, l.LListMeta.Marshal())
}

// RPush pushes elements into right side of list
func (l *LList) RPush(data ...[]byte) (err error) {
	for i := range data {
		l.Rindex++
		if err = l.txn.t.Set(append(l.rawDataKeyPrefix, EncodeFloat64(l.Rindex)...), data[i]); err != nil {
			return err
		}
		l.Len++
		if l.Len == 1 {
			l.Lindex = l.Rindex
		}
	}
	return l.txn.t.Set(l.rawMetaKey, l.LListMeta.Marshal())
}

// Set the index object with given value, return ErrIndex on out of range error.
func (l *LList) Set(n int64, data []byte) error {
	if n < 0 {
		n = l.Len + n
	}
	if n < 0 || n >= l.Len {
		return ErrOutOfRange
	}
	realidx, _, err := l.index(n)
	if err != nil {
		return err
	}
	return l.txn.t.Set(append(l.rawDataKeyPrefix, EncodeFloat64(realidx)...), data)
}

// Insert value in the list stored at key either before or after the reference value pivot
// 1. pivot berfore/ pivot/ next --> real indexs
func (l *LList) Insert(pivot, v []byte, before bool) error {
	idxs, err := l.indexValue(pivot)
	if err != nil {
		return err
	}

	var idx float64
	if before {
		if idxs[0] == math.MaxFloat64 { // LPUSH
			l.LListMeta.Lindex--
			idx = l.LListMeta.Lindex
		} else if idx, err = calculateIndex(idxs[0], idxs[1]); err != nil {
			return err
		}
	} else {
		if idxs[2] == math.MaxFloat64 { // RPUSH
			l.LListMeta.Rindex++
			idx = l.LListMeta.Rindex
		} else if idx, err = calculateIndex(idxs[1], idxs[2]); err != nil {
			return err
		}
	}
	l.Len++
	if err = l.txn.t.Set(append(l.rawDataKeyPrefix, EncodeFloat64(idx)...), v); err != nil {
		return err
	}
	return l.txn.t.Set(l.rawMetaKey, l.LListMeta.Marshal())
}

// Index returns the element at index n in the list stored at key
func (l *LList) Index(n int64) (data []byte, err error) {
	if n < 0 {
		n = l.Len + n
	}
	if n < 0 || n >= l.Len {
		return nil, ErrOutOfRange
	}
	_, val, err := l.index(n)
	if err != nil {
		return nil, err
	}
	return val, nil
}

// LPop returns and deletes the left most element
// 0. calculate data key
// 1. iterate to last value
// 2. get the key and call kv delete
// 3. modify the new index in meta
func (l *LList) LPop() (data []byte, err error) {
	if l.Len == 0 {
		// XXX should delete this?
		return nil, ErrKeyNotFound
	}
	leftKey := append(l.rawDataKeyPrefix, EncodeFloat64(l.LListMeta.Lindex)...)

	// find the left object
	iter, err := l.txn.t.Iter(leftKey, nil)
	if err != nil {
		return nil, err
	}
	if !iter.Valid() || !iter.Key().HasPrefix(l.rawDataKeyPrefix) {
		return nil, ErrKeyNotFound
	}
	val := iter.Value()

	if err = l.txn.t.Delete(iter.Key()); err != nil {
		return nil, err
	}

	if l.Len == 1 {
		return val, l.txn.t.Delete(l.rawMetaKey)
	}

	// get the next data object and check if get
	err = iter.Next()
	if err != nil {
		return nil, err
	}
	if !iter.Valid() || !iter.Key().HasPrefix(l.rawDataKeyPrefix) {
		return nil, ErrKeyNotFound
	}
	l.LListMeta.Len--
	l.LListMeta.Lindex = DecodeFloat64(iter.Key()[len(l.rawDataKeyPrefix):]) // trim prefix with list data key
	return val, l.txn.t.Set(l.rawMetaKey, l.LListMeta.Marshal())
}

// RPop returns and deletes the right most element
func (l *LList) RPop() ([]byte, error) {
	if l.Len == 0 {
		return nil, ErrKeyNotFound
	}
	// rightKey: {DB.ns}:{DB.id}:D:{linedx}
	rightKey := append(l.rawDataKeyPrefix, EncodeFloat64(l.LListMeta.Rindex)...)

	// find the left object
	iter, err := l.txn.t.IterReverse(rightKey)
	if err != nil {
		return nil, err
	}
	if !iter.Valid() || !iter.Key().HasPrefix(l.rawDataKeyPrefix) {
		return nil, ErrKeyNotFound
	}
	val := iter.Value()
	if err = l.txn.t.Delete(iter.Key()); err != nil {
		return nil, err
	}

	if l.Len == 1 {
		return val, l.txn.t.Delete(l.rawMetaKey)
	}

	// get the next data object and check if get
	err = iter.Next()
	if err != nil {
		return nil, err
	}
	if !iter.Valid() || !iter.Key().HasPrefix(l.rawDataKeyPrefix) {
		return nil, ErrKeyNotFound
	}
	l.LListMeta.Len--
	l.LListMeta.Rindex = DecodeFloat64(iter.Key()[len(l.rawDataKeyPrefix):]) // trim prefix with list data key
	return val, l.txn.t.Set(l.rawMetaKey, l.LListMeta.Marshal())
}

// Range returns the elements in [left, right]
func (l *LList) Range(left, right int64) (value [][]byte, err error) {
	if right < 0 {
		if right = l.Len + right; right < 0 {
			return [][]byte{}, nil
		}
	}
	if left < 0 {
		if left = l.Len + left; left < 0 {
			left = 0
		}
	}
	// return 0 elements
	if left > right {
		return [][]byte{}, nil
	}
	_, v, err := l.scan(left, right+1)
	return v, err
}

// LTrim an existing list so that it will contain only the specified range of elements specified
func (l *LList) LTrim(start int64, stop int64) error {
	if start < 0 {
		if start = l.Len + start; start < 0 {
			start = 0
		}
	}
	if stop < 0 {
		stop = l.Len + stop
	}
	if stop > l.Len-1 {
		stop = l.Len - 1
	}

	if start > stop {
		return l.Destory()
	}

	lIndex := l.Lindex
	rIndex := l.Rindex
	var err error
	// Iter from left to start, when start is 0, do not need to seek
	if start > 0 {
		lIndex, err = l.remove(l.LListMeta.Lindex, start)
		if err != nil {
			return err
		}
	}
	if stop+1 < l.Len {
		rIndex, _, err = l.index(stop) // stop is included to reserve
		if err != nil {
			return err
		}
		iter, err := l.seekIndex(stop)
		if err != nil {
			return err
		}
		defer iter.Close()
		rIndex = DecodeFloat64(iter.Key()[len(l.rawDataKeyPrefix):])
		if err := iter.Next(); err != nil {
			return err
		}

		toRemove := DecodeFloat64(iter.Key()[len(l.rawDataKeyPrefix):])
		_, err = l.remove(toRemove, l.Len-stop-1)
		if err != nil {
			return err
		}
	}
	l.Lindex = lIndex
	l.Rindex = rIndex
	l.Len = stop - start + 1
	if l.LListMeta.Len == 0 { // destory if len comes to 0
		return l.txn.t.Delete(l.rawMetaKey)
	}
	return l.txn.t.Set(l.rawMetaKey, l.LListMeta.Marshal())
}

// seekIndex will return till we get the last element not larger than index
func (l *LList) seekIndex(index int64) (Iterator, error) {
	key := append(l.rawDataKeyPrefix, EncodeFloat64(l.Lindex)...)
	iter, err := l.txn.t.Iter(key, nil)
	if err != nil {
		return nil, err
	}
	for ; iter.Valid() && iter.Key().HasPrefix(l.rawDataKeyPrefix) && index != 0 && err == nil; err = iter.Next() {
		index--
	}
	if err != nil {
		return nil, err
	}
	return iter, nil
}

// remove n elements from start, return next index after delete
func (l *LList) remove(start float64, n int64) (float64, error) {
	startKey := append(l.rawDataKeyPrefix, EncodeFloat64(start)...)
	iter, err := l.txn.t.Iter(startKey, nil)
	if err != nil {
		return 0, err
	}
	for iter.Valid() && iter.Key().HasPrefix(l.rawDataKeyPrefix) && n != 0 {
		if err := l.txn.t.Delete(iter.Key()); err != nil {
			return 0, err
		}
		if err := iter.Next(); err != nil {
			return 0, err
		}
		n--
	}
	if n > 0 {
		return l.Rindex, nil
	}

	if err := iter.Next(); err != nil {
		return 0, err
	}
	nextKey := iter.Key()
	if !nextKey.HasPrefix(l.rawDataKeyPrefix) {
		return l.Rindex, nil
	}

	nextIdx := DecodeFloat64(nextKey[len(l.rawDataKeyPrefix):])
	return nextIdx, nil
}

// index return the index and value of the n index list data
// n should be positive
func (l *LList) index(n int64) (realindex float64, value []byte, err error) {
	if n < 0 || n >= l.Len {
		return 0, nil, ErrOutOfRange
	}

	// case1: only 1 object in list
	if l.Len == 1 {
		val, err := l.txn.t.Get(append(l.rawDataKeyPrefix, EncodeFloat64(l.LListMeta.Lindex)...))
		if err != nil {
			return 0, nil, err
		}
		return l.Lindex, val, nil
	}

	// case2: only 2 object in list
	if l.Len == 2 {
		idx := l.LListMeta.Rindex
		if n == 0 {
			idx = l.LListMeta.Lindex
		}

		val, err := l.txn.t.Get(append(l.rawDataKeyPrefix, EncodeFloat64(idx)...))
		if err != nil {
			return 0, nil, err
		}
		return idx, val, nil
	}

	idxs, vals, err := l.scan(n, n+1)
	if err != nil {
		return 0, nil, err
	}
	return idxs[0], vals[0], nil
}

// LRem removes the first count occurrences of elements equal to value from the list stored at key
func (l *LList) LRem(v []byte, n int64) (int, error) {
	idxs, err := l.indexValueN(v, n)
	if err != nil {
		return 0, err
	}

	for i := range idxs {
		if err = l.txn.t.Delete(append(l.rawDataKeyPrefix, EncodeFloat64(idxs[i])...)); err != nil {
			return 0, err
		}
	}

	l.LListMeta.Len -= int64(len(idxs))
	if l.LListMeta.Len == 0 { // destory if len comes to 0
		return len(idxs), l.txn.t.Delete(l.rawMetaKey)
	}

	// TODO maybe we can find a new way to avoid these seek
	// update list index and left right index
	rightKey := append(l.rawDataKeyPrefix, EncodeFloat64(l.LListMeta.Rindex)...)
	iter, err := l.txn.t.IterReverse(rightKey)
	if err != nil {
		return 0, err
	}
	if !iter.Valid() || !iter.Key().HasPrefix(l.rawDataKeyPrefix) {
		return 0, ErrKeyNotFound
	}
	l.LListMeta.Rindex = DecodeFloat64(iter.Key()[len(l.rawDataKeyPrefix):]) // trim prefix with list data key

	leftKey := append(l.rawDataKeyPrefix, EncodeFloat64(l.LListMeta.Lindex)...)
	iter, err = l.txn.t.Iter(leftKey, nil)
	if err != nil {
		return 0, err
	}
	if !iter.Valid() || !iter.Key().HasPrefix(l.rawDataKeyPrefix) {
		return 0, ErrKeyNotFound
	}
	l.LListMeta.Lindex = DecodeFloat64(iter.Key()[len(l.rawDataKeyPrefix):]) // trim prefix with list data key

	return len(idxs), l.txn.t.Set(l.rawMetaKey, l.LListMeta.Marshal())
}

// indexValueN return the index of the given list data value.
func (l *LList) indexValueN(v []byte, n int64) (realidxs []float64, err error) {
	var iter kv.Iterator
	if n < 0 {
		n = -n
		if iter, err = l.txn.t.IterReverse(append(l.rawDataKeyPrefix, EncodeFloat64(l.LListMeta.Rindex)...)); err != nil {
			return nil, err
		}
	} else if n > 0 {
		if iter, err = l.txn.t.Iter(append(l.rawDataKeyPrefix, EncodeFloat64(l.LListMeta.Lindex)...), nil); err != nil {
			return nil, err
		}
	} else {
		n = l.Len
	}

	// for loop iterate all objects and check if valid until reach pivot value
	for count := int64(0); count < n && err == nil && iter.Valid() && iter.Key().HasPrefix(l.rawDataKeyPrefix); err = iter.Next() {
		// reset the rindex/lindex here
		if bytes.Equal(iter.Value(), v) { // found the value! now iter the next and return
			realidxs = append(realidxs, DecodeFloat64(iter.Key()[len(l.rawDataKeyPrefix):]))
			count++
		}
	}
	return realidxs, err
}

// indexValue return the [befor, real, after] index and value of the given list data value.
func (l *LList) indexValue(v []byte) (realidxs []float64, err error) {
	realidxs = []float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64}
	iter, err := l.txn.t.Iter(append(l.rawDataKeyPrefix, EncodeFloat64(l.LListMeta.Lindex)...), nil)
	if err != nil { // dup
		return nil, err
	}

	flag := false //if we find the v key
	rawDataKeyPrefixLen := len(l.rawDataKeyPrefix)

	// for loop iterate all objects and check if valid until reach pivot value
	for ; iter.Valid() && iter.Key().HasPrefix(l.rawDataKeyPrefix) && err == nil; err = iter.Next() {
		realidxs[0] = realidxs[1]
		realidxs[1] = realidxs[2]
		realidxs[2] = DecodeFloat64(iter.Key()[rawDataKeyPrefixLen:])
		if bytes.Equal(iter.Value(), v) { // found the value! now iter the next and return
			flag = true
			realidxs[0] = realidxs[1]
			realidxs[1] = realidxs[2]
			if err = iter.Next(); err == nil && iter.Valid() && iter.Key().HasPrefix(l.rawDataKeyPrefix) {
				// found next value index
				realidxs[2] = DecodeFloat64(iter.Key()[rawDataKeyPrefixLen:])
			} else {
				// did not found next value
				// 1.  {before, key, next}
				// 1.1 {NULL, key, next}
				// 2.  {before, key, NULL}
				// 2.1 {NULL, key, NULL}
				realidxs[2] = math.MaxFloat64
			}
			break
		}
	}
	if err != nil {
		return nil, err
	}
	if flag {
		return realidxs, nil
	}
	return nil, ErrKeyNotFound
}

// scan return objects between [left, right) range
// if left == right , we should return the firsh element
// else case: iterator all keys and get the object index value
// rightidx may not larger than right
func (l *LList) scan(left, right int64) (realidxs []float64, values [][]byte, err error) {
	realidxs = make([]float64, 0, right-left)
	values = make([][]byte, 0, right-left)

	// seek start indecate the seek first key start time.
	start := time.Now()
	iter, err := l.txn.t.Iter(append(l.rawDataKeyPrefix, EncodeFloat64(l.LListMeta.Lindex)...), nil)

	var idx int64
	// for loop iterate all objects to get the next data object and check if valid
	for idx = 0; idx < left && err == nil && iter.Valid() && iter.Key().HasPrefix(l.rawDataKeyPrefix); err = iter.Next() {
		idx++
	}
	if err != nil { // err tikv error
		return nil, nil, err
	}
	// if list not exist, return the 0
	if idx != left {
		return []float64{}, [][]byte{}, nil
	}

	//monitor the seek first key cost
	metrics.GetMetrics().LRangeSeekHistogram.Observe(time.Since(start).Seconds())

	// for loop iterate all objects to get objects and check if valid
	for ; idx < right && err == nil && iter.Valid() && iter.Key().HasPrefix(l.rawDataKeyPrefix); err = iter.Next() {
		// found
		realidxs = append(realidxs, DecodeFloat64(iter.Key()[len(l.rawDataKeyPrefix):]))
		values = append(values, iter.Value())
		idx++
	}
	if err != nil {
		return nil, nil, err
	}
	return realidxs, values, nil
}

// Exist checks if a list exists
func (l *LList) Exist() bool {
	if l.Len == 0 {
		return false
	}
	return true
}

// Destory the list
func (l *LList) Destory() error {
	// delete the meta data
	l.txn.t.Delete(l.rawMetaKey)
	// leaving the data to gc
	return gc(l.txn.t, l.rawDataKeyPrefix)
}

// calculateIndex return the real index between left and right, return ErrPerc=
func calculateIndex(left, right float64) (float64, error) {
	if f := (left + right) / 2; f != left && f != right {
		return f, nil
	}
	return 0, ErrPrecision
}
