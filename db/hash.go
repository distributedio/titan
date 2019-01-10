package db

import (
	"encoding/json"
	"strconv"
)

// HashMeta is the meta data of the hashtable
type HashMeta struct {
	Object
	Len int64
}

// Hash implements the hashtable
type Hash struct {
	meta HashMeta
	key  []byte
	txn  *Transaction
}

// GetHash returns a hash object, create new one if nonexists
func GetHash(txn *Transaction, key []byte) (*Hash, error) {
	hash := &Hash{txn: txn, key: key}

	mkey := MetaKey(txn.db, key)
	meta, err := txn.t.Get(mkey)
	if err != nil {
		if IsErrNotFound(err) {
			now := Now()
			hash.meta.CreatedAt = now
			hash.meta.UpdatedAt = now
			hash.meta.ExpireAt = 0
			hash.meta.ID = UUID()
			hash.meta.Type = ObjectHash
			hash.meta.Encoding = ObjectEncodingHT
			hash.meta.Len = 0
			return hash, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(meta, &hash.meta); err != nil {
		return nil, err
	}
	if hash.meta.Type != ObjectHash {
		return nil, ErrTypeMismatch
	}
	return hash, nil
}
func hashItemKey(key []byte, field []byte) []byte {
	key = append(key, ':')
	return append(key, field...)
}

// HDel removes the specified fields from the hash stored at key
func (hash *Hash) HDel(fields [][]byte) (int64, error) {
	var keys [][]byte
	var num int64
	dkey := DataKey(hash.txn.db, hash.meta.ID)
	for _, field := range fields {
		keys = append(keys, hashItemKey(dkey, field))
	}
	values, err := BatchGetValues(hash.txn, keys)
	if err != nil {
		return 0, err
	}
	for i, val := range values {
		if val == nil {
			continue
		}
		if err := hash.txn.t.Delete(keys[i]); err != nil {
			return 0, err
		}
		num++
	}
	if num == 0 {
		return 0, nil
	}
	hash.meta.Len -= num
	if hash.meta.Len == 0 {
		return num, hash.Destory()
	}
	if err := hash.updateMeta(); err != nil {
		return 0, err
	}
	return num, nil
}

// HSet sets field in the hash stored at key to value
func (hash *Hash) HSet(field []byte, value []byte) (int, error) {
	dkey := DataKey(hash.txn.db, hash.meta.ID)
	ikey := hashItemKey(dkey, field)
	exist := true

	_, err := hash.txn.t.Get(ikey)
	if err != nil {
		if !IsErrNotFound(err) {
			return 0, err
		}
		exist = false
	}

	if err := hash.txn.t.Set(ikey, value); err != nil {
		return 0, err
	}

	if exist {
		return 0, nil
	}
	hash.meta.Len++
	if err := hash.updateMeta(); err != nil {
		return 0, err
	}
	return 1, nil
}

// HSetNX sets field in the hash stored at key to value, only if field does not yet exist
func (hash *Hash) HSetNX(field []byte, value []byte) (int, error) {
	dkey := DataKey(hash.txn.db, hash.meta.ID)
	ikey := hashItemKey(dkey, field)

	_, err := hash.txn.t.Get(ikey)
	if err != nil {
		if !IsErrNotFound(err) {
			return 0, err
		}
		return 0, nil
	}
	if err := hash.txn.t.Set(ikey, value); err != nil {
		return 0, err
	}

	hash.meta.Len++
	if err := hash.updateMeta(); err != nil {
		return 0, err
	}
	return 1, nil
}

// HGet returns the value associated with field in the hash stored at key
func (hash *Hash) HGet(field []byte) ([]byte, error) {
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
	dkey := DataKey(hash.txn.db, hash.meta.ID)
	prefix := append(dkey, ':')
	iter, err := hash.txn.t.Iter(prefix, nil)
	if err != nil {
		return nil, nil, err
	}
	var fields [][]byte
	var vals [][]byte
	count := hash.meta.Len
	for iter.Valid() && iter.Key().HasPrefix(prefix) && count != 0 {
		fields = append(fields, []byte(iter.Key()[len(prefix):]))
		vals = append(vals, iter.Value())
		if err := iter.Next(); err != nil {
			return nil, nil, err
		}
		count--
	}
	return fields, vals, nil
}

func (hash *Hash) updateMeta() error {
	meta, err := json.Marshal(hash.meta)
	if err != nil {
		return err
	}
	return hash.txn.t.Set(MetaKey(hash.txn.db, hash.key), meta)
}

// Destory the hash store
func (hash *Hash) Destory() error {
	return hash.txn.Destory(&hash.meta.Object, hash.key)
}

// HExists returns if field is an existing field in the hash stored at key
func (hash *Hash) HExists(field []byte) (bool, error) {
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
	var exist bool

	dkey := DataKey(hash.txn.db, hash.meta.ID)
	ikey := hashItemKey(dkey, field)
	val, err := hash.txn.t.Get(ikey)
	if err != nil && !IsErrNotFound(err) {
		return 0, err
	}
	if err == nil {
		exist = true
		n, err = strconv.ParseInt(string(val), 10, 64)
		if err != nil {
			return 0, err
		}
	}
	n += v

	val = []byte(strconv.FormatInt(n, 10))
	if err := hash.txn.t.Set(ikey, val); err != nil {
		return 0, err
	}

	if !exist {
		hash.meta.Len++
		if err := hash.updateMeta(); err != nil {
			return 0, err
		}
	}
	return n, nil
}

// HIncrByFloat increment the specified field of a hash stored at key,
// and representing a floating point number, by the specified increment
func (hash *Hash) HIncrByFloat(field []byte, v float64) (float64, error) {
	var n float64
	var exist bool

	dkey := DataKey(hash.txn.db, hash.meta.ID)
	ikey := hashItemKey(dkey, field)
	val, err := hash.txn.t.Get(ikey)
	if err != nil && !IsErrNotFound(err) {
		return 0, err
	}
	if err == nil {
		exist = true
		n, err = strconv.ParseFloat(string(val), 64)
		if err != nil {
			return 0, err
		}
	}
	n += v

	val = []byte(strconv.FormatFloat(n, 'f', -1, 64))
	if err := hash.txn.t.Set(ikey, val); err != nil {
		return 0, err
	}

	if !exist {
		hash.meta.Len++
		if err := hash.updateMeta(); err != nil {
			return 0, err
		}
	}
	return n, nil
}

// HLen returns the number of fields contained in the hash stored at key
func (hash *Hash) HLen() int64 {
	return hash.meta.Len
}

// HMGet returns the values associated with the specified fields in the hash stored at key
func (hash *Hash) HMGet(fields [][]byte) ([][]byte, error) {
	ikeys := make([][]byte, len(fields))
	dkey := DataKey(hash.txn.db, hash.meta.ID)
	for i := range fields {
		ikeys[i] = hashItemKey(dkey, fields[i])
	}

	return BatchGetValues(hash.txn, ikeys)
}

// HMSet sets the specified fields to their respective values in the hash stored at key
func (hash *Hash) HMSet(fields [][]byte, values [][]byte) error {
	added := int64(0)
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

	hash.meta.Len += added
	return hash.updateMeta()
}
