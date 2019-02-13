package db

import (
	"bytes"
	"encoding/binary"
)

// SetNilValue is the value set to a tikv key for tikv do not support a real empty value
var SetNilValue = []byte{0}

// SetMeta is the meta data of the set
type SetMeta struct {
	Object
	Len int64
}

// Set implements the set data structure
type Set struct {
	meta   *SetMeta
	key    []byte
	exists bool
	txn    *Transaction
}

// GetSet returns a set object, create new one if nonexists
func GetSet(txn *Transaction, key []byte) (*Set, error) {
	set := newSet(txn, key)
	mkey := MetaKey(txn.db, key)
	meta, err := txn.t.Get(mkey)
	if err != nil {
		if IsErrNotFound(err) {
			return set, nil
		}
		return nil, err
	}
	smeta, err := DecodeSetMeta(meta)
	if err != nil {
		return nil, err
	}
	if smeta.Type != ObjectSet {
		return nil, ErrTypeMismatch
	}
	if IsExpired(&set.meta.Object, Now()) {
		return set, nil
	}
	set.meta = smeta
	set.exists = true
	return set, nil
}

// GetMetaKey gets MetaKey based on the given key
func GetMetaKey(txn *Transaction, key []byte) (mkey []byte) {
	mkey = MetaKey(txn.db, key)
	return
}

//newSet  create new Set object
func newSet(txn *Transaction, key []byte) *Set {
	now := Now()
	return &Set{
		txn: txn,
		key: key,
		meta: &SetMeta{
			Object: Object{
				ID:        UUID(),
				CreatedAt: now,
				UpdatedAt: now,
				ExpireAt:  0,
				Type:      ObjectSet,
				Encoding:  ObjectEncodingHT,
			},
			Len: 0,
		},
	}
}

//DecodeSetMeta decode meta data into meta field
func DecodeSetMeta(b []byte) (*SetMeta, error) {
	if len(b[ObjectEncodingLength:]) != 8 {
		return nil, ErrInvalidLength
	}
	obj, err := DecodeObject(b)
	if err != nil {
		return nil, err
	}
	smeta := &SetMeta{Object: *obj}
	m := b[ObjectEncodingLength:]
	smeta.Len = int64(binary.BigEndian.Uint64(m[:8]))
	return smeta, nil
}

//EncodeSetMeta encodes meta data into byte slice
func encodeSetMeta(meta *SetMeta) []byte {
	b := EncodeObject(&meta.Object)
	m := make([]byte, 8)
	binary.BigEndian.PutUint64(m[:8], uint64(meta.Len))
	return append(b, m...)
}
func setItemKey(key []byte, member []byte) []byte {
	var ikeys []byte
	ikeys = append(ikeys, key...)
	ikeys = append(ikeys, ':')
	ikeys = append(ikeys, member...)
	return ikeys
}
func (set *Set) updateMeta() error {
	meta := encodeSetMeta(set.meta)
	err := set.txn.t.Set(MetaKey(set.txn.db, set.key), meta)
	if err != nil {
		return err
	}
	if !set.exists {
		set.exists = true
	}
	return nil
}

// SAdd adds the specified members to the set stored at key
func (set *Set) SAdd(members ...[]byte) (int64, error) {
	// Namespace:DBID:D:ObjectID
	dkey := DataKey(set.txn.db, set.meta.ID)
	// Remove the duplicate
	ms := RemoveRepByMap(members)
	ikeys := make([][]byte, len(ms))
	for i := range ms {
		ikeys[i] = setItemKey(dkey, ms[i])
	}
	// {Namespace}:{DBID}:{D}:{ObjectID}:{ms[i]}
	values, err := BatchGetValues(set.txn, ikeys)

	if err != nil {
		return 0, nil
	}
	added := int64(0)
	for i := range ikeys {
		if values[i] == nil {
			added++
		}
		if err := set.txn.t.Set(ikeys[i], SetNilValue); err != nil {
			return 0, err
		}
	}
	set.meta.Len += added
	if err := set.updateMeta(); err != nil {
		return 0, err
	}

	return added, nil
}

// RemoveRepByMap filters duplicate elements through the map's unique primary key feature
func RemoveRepByMap(members [][]byte) [][]byte {
	result := [][]byte{}
	// tempMap saves non-repeating primary keys
	tempMap := map[string]int{}
	for _, m := range members {
		l := len(tempMap)
		tempMap[string(m)] = 0
		if len(tempMap) != l {
			result = append(result, m)
		}
	}
	return result
}

// Exists check set exist
func (set *Set) Exists() bool {
	return set.exists
}

// SMembers returns all the members of the set value stored at key
func (set *Set) SMembers() ([][]byte, error) {
	if !set.Exists() {
		return nil, nil
	}
	dkey := DataKey(set.txn.db, set.meta.ID)
	prefix := append(dkey, ':')

	count := set.meta.Len
	members := make([][]byte, 0, count)

	iter, err := set.txn.t.Iter(prefix, nil)
	if err != nil {
		return nil, err
	}

	for iter.Valid() && iter.Key().HasPrefix(prefix) && count != 0 {
		members = append(members, iter.Key()[len(prefix):])
		if err := iter.Next(); err != nil {
			return nil, err
		}
		count--
	}
	return members, nil
}

// SCard returns the set cardinality (number of elements) of the set stored at key
func (set *Set) SCard() (int64, error) {
	if !set.Exists() {
		return 0, nil
	}
	return set.meta.Len, nil
}

// SIsmember returns if member is a member of the set stored at key
func (set *Set) SIsmember(member []byte) (int64, error) {
	if !set.Exists() {
		return 0, nil
	}
	dkey := DataKey(set.txn.db, set.meta.ID)
	ikey := setItemKey(dkey, member)

	value, err := set.txn.t.Get(ikey)
	if err != nil {
		if IsErrNotFound(err) {
			return 0, nil
		}
		return 0, err
	}
	if !bytes.Equal(value, SetNilValue) {
		return 0, nil
	}
	return 1, nil
}

// SPop removes and returns one or more random elements from the set value store at key.
func (set *Set) SPop(count int64) (members [][]byte, err error) {
	// TODO BUG  No rand
	if !set.Exists() {
		return nil, nil
	}
	dkey := DataKey(set.txn.db, set.meta.ID)
	prefix := append(dkey, ':')
	iter, err := set.txn.t.Iter(prefix, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var del int64
	var ms [][]byte
	if count == 0 {
		if iter.Valid() && iter.Key().HasPrefix(prefix) {
			ms = append(ms, iter.Key()[len(prefix):])
			if err := set.txn.t.Delete([]byte(iter.Key())); err != nil {
				return nil, err
			}
			del++
		}
	} else {
		for iter.Valid() && iter.Key().HasPrefix(prefix) && count != 0 {
			ms = append(ms, iter.Key()[len(prefix):])
			if err := set.txn.t.Delete([]byte(iter.Key())); err != nil {
				return nil, err
			}
			del++
			count--
			if err := iter.Next(); err != nil {
				return nil, err

			}
		}
	}
	if count < set.meta.Len {
		set.meta.Len -= del
	} else {
		set.meta.Len = 0
	}
	if err := set.updateMeta(); err != nil {
		return nil, err
	}

	return ms, nil
}

// SRem removes the specified members from the set stored at key
func (set *Set) SRem(members [][]byte) (int64, error) {
	var num int64
	if !set.Exists() {
		return 0, nil
	}
	dkey := DataKey(set.txn.db, set.meta.ID)
	ms := RemoveRepByMap(members)
	ikeys := make([][]byte, len(ms))
	for i := range ms {
		ikeys[i] = setItemKey(dkey, ms[i])
		value, err := set.txn.t.Get(ikeys[i])
		if err != nil {
			if IsErrNotFound(err) {
				continue
			}
			return 0, err
		}
		if bytes.Equal(value, SetNilValue) {
			if err := set.txn.t.Delete([]byte(ikeys[i])); err != nil {
				return 0, err
			}
			num++
		}
	}
	set.meta.Len -= num
	if err := set.updateMeta(); err != nil {
		return 0, err
	}
	return num, nil
}

// SMove movies member from the set at source to the set at destination
func (set *Set) SMove(destination []byte, member []byte) (int64, error) {

	if !set.Exists() {
		return 0, nil
	}
	res, err := set.SIsmember(member)
	if err != nil {
		return 0, err
	}
	if res == 0 {
		return 0, nil
	}
	destset, _ := GetSet(set.txn, destination)
	res, err = destset.SIsmember(member)
	if err != nil {
		return 0, err
	}
	if res == 0 {
		if _, err := destset.SAdd(member); err != nil {
			return 0, err
		}
		destset.meta.Len++
		if err := destset.updateMeta(); err != nil {
			return 0, err
		}
	}
	dkey := DataKey(set.txn.db, set.meta.ID)
	ikey := setItemKey(dkey, member)

	value, err := set.txn.t.Get(ikey)
	if err != nil {
		if IsErrNotFound(err) {
			return 0, nil
		}
		return 0, err
	}
	if bytes.Equal(value, SetNilValue) {
		if err := set.txn.t.Delete([]byte(ikey)); err != nil {
			return 0, err
		}
		set.meta.Len--
		if err := set.updateMeta(); err != nil {
			return 0, err
		}
		return 1, nil
	}
	return 0, nil
}
