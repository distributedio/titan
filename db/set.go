package db

import (
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
func EncodeSetMeta(meta *SetMeta) []byte {
	b := EncodeObject(&meta.Object)
	m := make([]byte, 8)
	binary.BigEndian.PutUint64(m[:8], uint64(meta.Len))
	return append(b, m...)
}
func setItemKey(key []byte, member []byte) []byte {
	key = append(key, ':')
	return append(key, member...)
}
func (set *Set) updateMeta() error {
	meta := EncodeSetMeta(set.meta)
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
func (set *Set) SAdd(members [][]byte) (int64, error) {
	// Namespace:DBID:D:ObjectID
	dkey := DataKey(set.txn.db, set.meta.ID)
	// Remove the duplicate
	ms := removeRepByMap(members)
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
func removeRepByMap(members [][]byte) [][]byte {
	result := [][]byte{}
	// tempMap saves non-repeating primary keys
	//tempMap := make(map[string]int)
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

// Exists check hashes exist
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
