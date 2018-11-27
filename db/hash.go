package db

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"math/rand"
	"strconv"

	"github.com/meitu/titan/db/store"
)

var (
	defaultHashSlots int64 = 0
)

type SlotMeta struct {
	ID        int64
	Len       int64
	UpdatedAt int64
}

func EncodeSlotMeta(s *SlotMeta) []byte {
	b := make([]byte, 24, 24)
	binary.BigEndian.PutUint64(b[:8], uint64(s.ID))
	binary.BigEndian.PutUint64(b[8:16], uint64(s.Len))
	binary.BigEndian.PutUint64(b[16:24], uint64(s.UpdatedAt))
	return b
}

func DecodeSlotMeta(b []byte) (*SlotMeta, error) {
	if len(b) != 24 {
		return nil, ErrInvalidLength
	}
	meta := &SlotMeta{}
	meta.ID = int64(binary.BigEndian.Uint64(b[:8]))
	meta.Len = int64(binary.BigEndian.Uint64(b[8:16]))
	meta.UpdatedAt = int64(binary.BigEndian.Uint64(b[16:24]))
	return meta, nil
}

// HashMeta is the meta data of the hashtable
type HashMeta struct {
	Object
	Len  int64
	Slot int64
}

func (hm *HashMeta) Encode() []byte {
	b := EncodeObject(&hm.Object)
	meta := make([]byte, 16, 16)
	binary.BigEndian.PutUint64(meta[:8], uint64(hm.Len))
	binary.BigEndian.PutUint64(meta[8:], uint64(hm.Slot))
	return append(b, meta...)
}

func (hm *HashMeta) Decode(b []byte) error {
	if len(b[ObjectEncodingLength:]) != 16 {
		return ErrInvalidLength
	}
	obj, err := DecodeObject(b)
	if err != nil {
		return err
	}
	hm.Object = *obj
	meta := b[ObjectEncodingLength:]
	hm.Len = int64(binary.BigEndian.Uint64(meta[:8]))
	hm.Slot = int64(binary.BigEndian.Uint64(meta[8:]))
	return nil
}

// Hash implements the hashtable
type Hash struct {
	meta HashMeta
	key  []byte
	txn  *Transaction
}

// GetHash returns a hash object, create new one if nonexists
func GetHash(txn *Transaction, key []byte) (*Hash, error) {
	hash := &Hash{txn: txn, key: key, meta: HashMeta{}}

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
			hash.meta.Slot = defaultHashSlots
			return hash, nil
		}
		return nil, err
	}
	if err := hash.meta.Decode(meta); err != nil {
		return nil, err
	}
	if hash.meta.Type != ObjectHash {
		return nil, ErrTypeMismatch
	}
	return hash, nil
}

func hashItemKey(key []byte, field []byte) []byte {
	key = append(key, []byte(Separator)...)
	return append(key, field...)
}

func hashSlotKey(key []byte, slot int64) []byte {
	key = append(key, []byte(Separator)...)
	return append(key, EncodeInt64(slot)...)
}

func (hash *Hash) calculateSlot(field []byte) int64 {
	if !hash.isSlot() {
		return 0
	}
	return int64(crc32.ChecksumIEEE(field)) % hash.meta.Slot
}

func (hash *Hash) isSlot() bool {
	if hash.meta.Slot != 0 {
		return true
	}
	return false
}

// HDel removes the specified fields from the hash stored at key
func (hash *Hash) HDel(fields [][]byte) (int64, error) {
	var keys [][]byte
	var num int64
	dkey := DataKey(hash.txn.db, hash.meta.ID)
	for _, field := range fields {
		keys = append(keys, hashItemKey(dkey, field))
	}
	kvMap, actionSlot, hlen, err := hash.delHash(keys)
	if err != nil {
		return 0, err
	}
	vlen := int64(len(kvMap))
	if vlen >= hlen {
		if err := hash.Destory(); err != nil {
			return 0, err
		}
		return vlen, nil
	}

	for k, v := range kvMap {
		if v == nil {
			continue
		}
		if err := hash.txn.t.Delete([]byte(k)); err != nil {
			return 0, err
		}
		num++
	}
	if num == 0 {
		return 0, nil
	}
	if actionSlot != nil && hash.isSlot() {
		actionSlot.Len = actionSlot.Len - num
		actionSlot.UpdatedAt = Now()
		if err := hash.updateSlot(actionSlot); err != nil {
			return 0, err
		}

	} else if err := hash.updateMeta(); err != nil {
		return 0, err
	}

	return num, nil
}

func (hash *Hash) delHash(keys [][]byte) (map[string][]byte, *SlotMeta, int64, error) {
	var (
		slots         [][]byte
		isSlot        = hash.isSlot()
		slotKeyPrefix = SlotKey(hash.txn.db, hash.meta.ID, nil)
	)
	if isSlot {
		slotKeys := hash.getSlotKeys()
		keys = append(slotKeys, keys...)
	}

	kvMap, err := store.BatchGetValues(hash.txn.t, keys)
	if err != nil {
		return nil, nil, 0, err
	}
	for k, v := range kvMap {
		if isSlot && bytes.Contains([]byte(k), slotKeyPrefix) {
			slots = append(slots, v)
			delete(kvMap, k)
		}
	}
	if isSlot && len(slots) > 0 {
		slot, err := hash.calculateSlotMeta(&slots)
		if err != nil {
			return nil, nil, 0, err
		}
		actionSlot, err := DecodeSlotMeta(slots[rand.Intn(len(slots))])
		if err != nil {
			return nil, nil, 0, err
		}
		return kvMap, actionSlot, slot.Len, nil
	}
	return kvMap, nil, hash.meta.Len, nil
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
	meta := hash.meta.Encode()
	return hash.txn.t.Set(MetaKey(hash.txn.db, hash.key), meta)
}

func (hash *Hash) updateSlot(slot *SlotMeta) error {
	slotKey := SlotKey(hash.txn.db, hash.meta.ID, EncodeInt64(slot.ID))
	smeta := EncodeSlotMeta(slot)
	return hash.txn.t.Set(slotKey, smeta)
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
func (hash *Hash) HLen() (int64, error) {
	if hash.isSlot() {
		skeys := hash.getSlotKeys()
		values, err := BatchGetValues(hash.txn, skeys)
		if err != nil {
			return 0, err
		}
		slotMeta, err := hash.calculateSlotMeta(&values)
		if err == nil {
			return 0, err
		}
		return slotMeta.Len, nil
	}
	return hash.meta.Len, nil

}

func (hash *Hash) getSlotKeys() [][]byte {
	slot := hash.meta.Slot
	keys := make([][]byte, slot, slot)
	for slot > 0 {
		keys = append(keys, SlotKey(hash.txn.db, hash.meta.ID, EncodeInt64(slot)))
		slot--
	}
	return keys
}

func (hash *Hash) calculateSlotMeta(vals *[][]byte) (*SlotMeta, error) {
	slot := &SlotMeta{}
	for _, val := range *vals {
		if val == nil {
			continue
		}
		meta, err := DecodeSlotMeta(val)
		if err != nil {
			return nil, err
		}
		slot.Len += meta.Len
		if meta.UpdatedAt > slot.UpdatedAt {
			slot.UpdatedAt = meta.UpdatedAt
		}
	}
	return slot, nil
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
