package db

import (
	"encoding/json"
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
	meta SetMeta
	key  []byte
	txn  *Transaction
}

// GetSet returns a set object, create new one if nonexists
func GetSet(txn *Transaction, key []byte) (*Set, error) {
	set := &Set{txn: txn, key: key}

	mkey := MetaKey(txn.db, key)
	meta, err := txn.t.Get(mkey)
	if err != nil {
		if IsErrNotFound(err) {
			now := Now()
			set.meta.CreatedAt = now
			set.meta.UpdatedAt = now
			set.meta.ExpireAt = 0
			set.meta.ID = UUID()
			set.meta.Type = ObjectSet
			set.meta.Encoding = ObjectEncodingHT
			set.meta.Len = 0
			return set, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(meta, &set.meta); err != nil {
		return nil, err
	}
	if set.meta.Type != ObjectSet {
		return nil, ErrTypeMismatch
	}
	return set, nil
}

func setItemKey(key []byte, member []byte) []byte {
	key = append(key, ':')
	return append(key, member...)
}

func (set *Set) updateMeta() error {
	meta, err := json.Marshal(set.meta)
	if err != nil {
		return err
	}
	return set.txn.t.Set(MetaKey(set.txn.db, set.key), meta)
}

// SAdd adds the specified members to the set stored at key
func (set *Set) SAdd(members [][]byte) (int64, error) {
	dkey := DataKey(set.txn.db, set.meta.ID)
	ikeys := make([][]byte, len(members))

	for i := range members {
		ikeys[i] = setItemKey(dkey, members[i])
	}

	values, err := BatchGetValues(set.txn, ikeys)
	if err != nil {
		return 0, nil
	}

	added := int64(0)
	for i := range members {
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

// SMembers returns all the members of the set value stored at key
func (set *Set) SMembers() ([][]byte, error) {
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
