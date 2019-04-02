package db

import (
	"encoding/binary"
	"errors"
	"strconv"

	"github.com/distributedio/titan/db/store"
	"github.com/pingcap/tidb/kv"
)

// HashMeta is the meta data of the hashtable
type HashMeta struct {
	Object
	Len int64
}

//EncodeHashMeta encodes meta data into byte slice
func EncodeHashMeta(meta *HashMeta) []byte {
	b := EncodeObject(&meta.Object)
	m := make([]byte, 8)
	binary.BigEndian.PutUint64(m[:8], uint64(meta.Len))
	return append(b, m...)
}

//DecodeHashMeta decode meta data into meta field
func DecodeHashMeta(b []byte) (*HashMeta, error) {
	if len(b[ObjectEncodingLength:]) != 8 {
		return nil, ErrInvalidLength
	}
	obj, err := DecodeObject(b)
	if err != nil {
		return nil, err
	}
	hmeta := &HashMeta{Object: *obj}
	m := b[ObjectEncodingLength:]
	hmeta.Len = int64(binary.BigEndian.Uint64(m[:8]))
	return hmeta, nil
}

// Hash implements the hashtable
type Hash struct {
	meta   *HashMeta
	key    []byte
	exists bool
	txn    *Transaction
}

// GetHash returns a hash object, create new one if nonexists
func GetHash(txn *Transaction, key []byte) (*Hash, error) {
	hash := newHash(txn, key)
	mkey := MetaKey(txn.db, key)
	meta, err := txn.t.Get(mkey)
	if err != nil {
		if IsErrNotFound(err) {
			return hash, nil
		}
		return nil, err
	}
	hmeta, err := DecodeHashMeta(meta)
	if err != nil {
		return nil, err
	}
	if hmeta.Type != ObjectHash {
		return nil, ErrTypeMismatch
	}
	if IsExpired(&hmeta.Object, Now()) {
		return hash, nil
	}
	hash.meta = hmeta
	hash.exists = true
	return hash, nil
}

//newHash creates a hash object
func newHash(txn *Transaction, key []byte) *Hash {
	now := Now()
	return &Hash{
		txn: txn,
		key: key,
		meta: &HashMeta{
			Object: Object{
				ID:        UUID(),
				CreatedAt: now,
				UpdatedAt: now,
				ExpireAt:  0,
				Type:      ObjectHash,
				Encoding:  ObjectEncodingHT,
			},
			Len: 0,
		},
	}
}

//hashItemKey spits field into metakey
func hashItemKey(key []byte, field []byte) []byte {
	var dkey []byte
	dkey = append(dkey, key...)
	dkey = append(dkey, ':')
	return append(dkey, field...)
}

// HDel removes the specified fields from the hash stored at key
func (hash *Hash) HDel(fields [][]byte) (int64, error) {
	var (
		fieldsMap  = make(map[string]bool, len(fields))
		num        int64
		retainMeta bool
	)

	if !hash.Exists() {
		return 0, nil
	}

	dkey := DataKey(hash.txn.db, hash.meta.ID)
	for _, f := range fields {
		field := hashItemKey(dkey, f)
		fieldsMap[string(field)] = true
	}
	prefix := kv.Key(hashItemKey(dkey, nil))
	endPrefix := prefix.PrefixNext()

	var delErr error
	callback := func(k kv.Key) bool {
		if _, ok := fieldsMap[string(k)]; ok {
			if delErr = hash.txn.t.Delete(k); delErr != nil {
				return true
			}
			num++
			return false
		}
		retainMeta = true
		if num == int64(len(fieldsMap)) {
			return true
		}
		return false
	}

	store.SetOption(hash.txn.t, store.KeyOnly, true)
	iter, err := hash.txn.t.Iter(prefix, endPrefix)
	if err != nil {
		return 0, err
	}
	if err := kv.NextUntil(iter, callback); err != nil {
		return 0, err
	}
	if delErr != nil {
		return 0, delErr
	}
	if !retainMeta {
		if err := hash.delMeta(); err != nil {
			return 0, err
		}
	}
	return num, nil
}

// HSet sets field in the hash stored at key to value
func (hash *Hash) HSet(field []byte, value []byte) (int, error) {
	dkey := DataKey(hash.txn.db, hash.meta.ID)
	ikey := hashItemKey(dkey, field)
	newField := false
	_, err := hash.txn.t.Get(ikey)
	if err != nil {
		if !IsErrNotFound(err) {
			return 0, err
		}
		newField = true
	}

	if err := hash.txn.t.Set(ikey, value); err != nil {
		return 0, err
	}

	if !hash.Exists() {
		if err := hash.setMeta(); err != nil {
			return 0, err
		}
	}
	if !newField {
		return 0, nil
	}

	return 1, nil
}

// HSetNX sets field in the hash stored at key to value, only if field does not yet exist
func (hash *Hash) HSetNX(field []byte, value []byte) (int, error) {
	dkey := DataKey(hash.txn.db, hash.meta.ID)
	ikey := hashItemKey(dkey, field)

	_, err := hash.txn.t.Get(ikey)
	if err == nil {
		return 0, nil
	}
	if !IsErrNotFound(err) {
		return 0, err
	}
	if err := hash.txn.t.Set(ikey, value); err != nil {
		return 0, err
	}

	//update and save meta
	if !hash.Exists() {
		if err := hash.setMeta(); err != nil {
			return 0, err
		}
	}
	return 1, nil
}

// HGet returns the value associated with field in the hash stored at key
func (hash *Hash) HGet(field []byte) ([]byte, error) {
	if !hash.Exists() {
		return nil, nil
	}
	dkey := DataKey(hash.txn.db, hash.meta.ID)
	ikey := hashItemKey(dkey, field)
	val, err := hash.txn.t.Get(ikey)
	if err != nil {
		if IsErrNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return val, nil
}

// HGetAll returns all fields and values of the hash stored at key
func (hash *Hash) HGetAll() ([][]byte, [][]byte, error) {
	if !hash.Exists() {
		return nil, nil, nil
	}
	dkey := DataKey(hash.txn.db, hash.meta.ID)
	prefix := hashItemKey(dkey, nil)
	endPrefix := kv.Key(prefix).PrefixNext()
	iter, err := hash.txn.t.Iter(prefix, endPrefix)
	if err != nil {
		return nil, nil, err
	}
	var fields [][]byte
	var vals [][]byte
	for iter.Valid() && iter.Key().HasPrefix(prefix) {
		fields = append(fields, []byte(iter.Key()[len(prefix):]))
		vals = append(vals, iter.Value())
		if err := iter.Next(); err != nil {
			return nil, nil, err
		}
	}
	return fields, vals, nil
}

// HExists returns if field is an existing field in the hash stored at key
func (hash *Hash) HExists(field []byte) (bool, error) {
	if !hash.Exists() {
		return false, nil
	}
	dkey := DataKey(hash.txn.db, hash.meta.ID)
	ikey := hashItemKey(dkey, field)
	if _, err := hash.txn.t.Get(ikey); err != nil {
		if IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// HIncrBy increments the number stored at field in the hash stored at key by increment
func (hash *Hash) HIncrBy(field []byte, v int64) (int64, error) {
	var n int64
	dkey := DataKey(hash.txn.db, hash.meta.ID)
	ikey := hashItemKey(dkey, field)

	if hash.Exists() {
		val, err := hash.txn.t.Get(ikey)
		if err != nil {
			if !IsErrNotFound(err) {
				return 0, err
			}
		} else {
			n, err = strconv.ParseInt(string(val), 10, 64)
			if err != nil {
				return 0, errors.New("hash value is not an integer")
			}

		}
	}
	n += v

	val := []byte(strconv.FormatInt(n, 10))
	if err := hash.txn.t.Set(ikey, val); err != nil {
		return 0, err
	}

	if !hash.Exists() {
		if err := hash.setMeta(); err != nil {
			return 0, err
		}
	}

	return n, nil
}

// HIncrByFloat increment the specified field of a hash stored at key,
// and representing a floating point number, by the specified increment
func (hash *Hash) HIncrByFloat(field []byte, v float64) (float64, error) {
	var n float64
	dkey := DataKey(hash.txn.db, hash.meta.ID)
	ikey := hashItemKey(dkey, field)
	if hash.Exists() {
		val, err := hash.txn.t.Get(ikey)
		if err != nil {
			if !IsErrNotFound(err) {
				return 0, err
			}
		} else {
			n, err = strconv.ParseFloat(string(val), 64)
			if err != nil {
				return 0, errors.New("hash value is not an float")
			}

		}
	}
	n += v

	val := []byte(strconv.FormatFloat(n, 'f', -1, 64))
	if err := hash.txn.t.Set(ikey, val); err != nil {
		return 0, err
	}

	if !hash.Exists() {
		if err := hash.setMeta(); err != nil {
			return 0, err
		}
	}

	return n, nil
}

// HLen returns the number of fields contained in the hash stored at key
func (hash *Hash) HLen() (int64, error) {
	if !hash.Exists() {
		return 0, nil
	}
	dkey := DataKey(hash.txn.db, hash.meta.ID)
	prefix := hashItemKey(dkey, nil)
	endPrefix := kv.Key(prefix).PrefixNext()
	var length int64
	callback := func(k kv.Key) bool {
		length++
		return false
	}
	store.SetOption(hash.txn.t, store.KeyOnly, true)
	iter, err := hash.txn.t.Iter(prefix, endPrefix)
	if err != nil {
		return 0, err
	}
	if err := kv.NextUntil(iter, callback); err != nil {
		return 0, err
	}

	return length, nil
}

// HMGet returns the values associated with the specified fields in the hash stored at key
func (hash *Hash) HMGet(fields [][]byte) ([][]byte, error) {
	values := make([][]byte, len(fields))
	if !hash.Exists() {
		return values, nil
	}
	ikeys := make([][]byte, len(fields))
	dkey := DataKey(hash.txn.db, hash.meta.ID)
	for i := range fields {
		ikeys[i] = hashItemKey(dkey, fields[i])
	}

	return BatchGetValues(hash.txn, ikeys)
}

// HMSet sets the specified fields to their respective values in the hash stored at key
func (hash *Hash) HMSet(fields, values [][]byte) error {
	var added int64
	oldValues, err := hash.HMGet(fields)
	if err != nil {
		return err
	}
	dkey := DataKey(hash.txn.db, hash.meta.ID)
	for i := range fields {
		ikey := hashItemKey(dkey, fields[i])
		if err := hash.txn.t.Set(ikey, values[i]); err != nil {
			return err
		}
		if oldValues[i] == nil {
			added++
		}
	}
	if added == 0 {
		return nil
	}
	if !hash.Exists() {
		if err := hash.setMeta(); err != nil {
			return err
		}
	}
	return nil
}

// HScan incrementally iterate hash fields and associated values
func (hash *Hash) HScan(cursor []byte, f func(key, val []byte) bool) error {
	if !hash.Exists() {
		return nil
	}
	dkey := DataKey(hash.txn.db, hash.meta.ID)
	prefix := hashItemKey(dkey, nil)
	endPrefix := kv.Key(prefix).PrefixNext()
	ikey := hashItemKey(dkey, cursor)
	iter, err := hash.txn.t.Iter(ikey, endPrefix)
	if err != nil {
		return err
	}
	for iter.Valid() && iter.Key().HasPrefix(prefix) {
		key := iter.Key()
		if !f(key[len(prefix):], iter.Value()) {
			break
		}
		if err := iter.Next(); err != nil {
			return err
		}
	}
	return nil
}

// Exists check hashes exist
func (hash *Hash) Exists() bool {
	return hash.exists
}

func (hash *Hash) setMeta() error {
	meta := EncodeHashMeta(hash.meta)
	err := hash.txn.t.Set(MetaKey(hash.txn.db, hash.key), meta)
	if err != nil {
		return err
	}
	if !hash.Exists() {
		hash.exists = true
	}
	return nil

}

func (hash *Hash) delMeta() error {
	err := hash.txn.t.Delete(MetaKey(hash.txn.db, hash.key))
	if err != nil {
		return err
	}
	if hash.Exists() {
		hash.exists = false
	}
	return nil

}
