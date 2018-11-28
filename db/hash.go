package db

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"strconv"

	"github.com/meitu/titan/db/store"
)

var (
	defaultHashMetaSlot int64 = 0
)

//Slot slot information about hash meta
type Slot struct {
	Len       int64
	UpdatedAt int64
}

//EncodeSlot encodes slot data into byte slice
func EncodeSlot(s *Slot) []byte {
	b := make([]byte, 16)
	binary.BigEndian.PutUint64(b[:8], uint64(s.Len))
	binary.BigEndian.PutUint64(b[8:], uint64(s.UpdatedAt))
	return b
}

// DecodeSlot decode slot data into slot field
func DecodeSlot(b []byte) (*Slot, error) {
	if len(b) != 16 {
		return nil, ErrInvalidLength
	}
	meta := &Slot{}
	meta.Len = int64(binary.BigEndian.Uint64(b[:8]))
	meta.UpdatedAt = int64(binary.BigEndian.Uint64(b[8:]))
	return meta, nil
}

// HashMeta is the meta data of the hashtable
type HashMeta struct {
	Object
	Len      int64
	MetaSlot int64
}

//Encode encodes meta data into byte slice
func (hm *HashMeta) Encode() []byte {
	b := EncodeObject(&hm.Object)
	meta := make([]byte, 16)
	binary.BigEndian.PutUint64(meta[:8], uint64(hm.Len))
	binary.BigEndian.PutUint64(meta[8:], uint64(hm.MetaSlot))
	return append(b, meta...)
}

//Decode decode meta data into meta field
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
	hm.MetaSlot = int64(binary.BigEndian.Uint64(meta[8:]))
	return nil
}

// Hash implements the hashtable
type Hash struct {
	meta   HashMeta
	key    []byte
	exists bool
	txn    *Transaction
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
			hash.meta.MetaSlot = defaultHashMetaSlot
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
	hash.exists = true
	return hash, nil
}

//NewHash create new hashes object
func NewHash(txn *Transaction, key []byte) *Hash {
	hash := &Hash{txn: txn, key: key, meta: HashMeta{}}
	now := Now()
	hash.meta.CreatedAt = now
	hash.meta.UpdatedAt = now
	hash.meta.ExpireAt = 0
	hash.meta.ID = UUID()
	hash.meta.Type = ObjectHash
	hash.meta.Encoding = ObjectEncodingHT
	hash.meta.Len = 0
	hash.meta.MetaSlot = defaultHashMetaSlot
	return hash
}

func hashItemKey(key []byte, field []byte) []byte {
	var dkey []byte
	dkey = append(dkey, key...)
	dkey = append(dkey, ':')
	return append(dkey, field...)
}

func slotGC(txn *Transaction, objID []byte) error {
	key := MetaSlotKey(txn.db, objID, nil)
	if err := gc(txn.t, key); err != nil {
		return err
	}
	return nil
}

func (hash *Hash) calculateSlotID(limit int64) int64 {
	if !hash.isMetaSlot() || limit <= 1 {
		return 0
	}
	return rand.Int63n(limit - 1)
}

func (hash *Hash) isMetaSlot() bool {
	if hash.meta.MetaSlot != 0 {
		return true
	}
	return false
}

func slotGC(txn *Transaction, objID []byte) error {
	key := MetaSlotKey(txn.db, objID, nil)
	if err := gc(txn.t, key); err != nil {
		return err
	}
	return nil
}

func (hash *Hash) calculateSlotID(limit int64) int64 {
	if !hash.isMetaSlot() || limit <= 1 {
		return 0
	}
	return rand.Int63n(limit - 1)
}

func (hash *Hash) isMetaSlot() bool {
	if hash.meta.MetaSlot != 0 {
		return true
	}
	return false
}

// HDel removes the specified fields from the hash stored at key
func (hash *Hash) HDel(fields [][]byte) (int64, error) {
	var (
		keys [][]byte
		num  int64
	)
	dkey := DataKey(hash.txn.db, hash.meta.ID)
	for _, field := range fields {
		keys = append(keys, hashItemKey(dkey, field))
	}
	values, hlen, err := hash.delHash(keys)
	if err != nil {
		return 0, err
	}
	if hlen == 0 {
		return 0, nil
	}
	vlen := int64(len(values))
	if vlen >= hlen {
		if err := hash.Destroy(); err != nil {
			return 0, err
		}
		return vlen, nil
	}

	for k, v := range values {
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

	// update Len and UpdateAt
	if err := hash.addLen(-num); err != nil {
		return 0, err
	}
	return num, nil
}

func (hash *Hash) delHash(keys [][]byte) (map[string][]byte, int64, error) {
	var (
		slots       [][]byte
		isMetaSlot  = hash.isMetaSlot()
		metaSlotKey = MetaSlotKey(hash.txn.db, hash.meta.ID, nil)
	)
	if isMetaSlot {
		metaSlotKeys := hash.getMetaSlotKeys()
		keys = append(metaSlotKeys, keys...)
	}

	kvMap, err := store.BatchGetValues(hash.txn.t, keys)
	if err != nil {
		return nil, 0, err
	}
	for k, v := range kvMap {
		if isMetaSlot && bytes.Contains([]byte(k), metaSlotKey) {
			slots = append(slots, v)
			delete(kvMap, k)
		}
	}
	if isMetaSlot && len(slots) > 0 {
		slot, err := hash.calculateSlot(&slots)
		if err != nil {
			return nil, 0, err
		}
		return kvMap, slot.Len, nil
	}
	return kvMap, hash.meta.Len, nil
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
	if err := hash.addLen(1); err != nil {
		return 0, err
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
	} else if !IsErrNotFound(err) {
		return 0, err
	}
	if err := hash.txn.t.Set(ikey, value); err != nil {
		return 0, err
	}
	if err := hash.addLen(1); err != nil {
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
	count, err := hash.HLen()
	if err != nil {
		return nil, nil, err
	}
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

// Destroy the hash store
func (hash *Hash) Destroy() error {
	metaKey := MetaKey(hash.txn.db, hash.key)
	dataKey := DataKey(hash.txn.db, hash.meta.ID)
	if err := hash.txn.t.Delete(metaKey); err != nil {
		return err
	}
	if err := gc(hash.txn.t, dataKey); err != nil {
		return err
	}

	if hash.isMetaSlot() {
		if err := slotGC(hash.txn, hash.meta.ID); err != nil {
			return err
		}
	}

	if hash.meta.ExpireAt > 0 {
		if err := unExpireAt(hash.txn.t, metaKey, hash.meta.ExpireAt); err != nil {
			return err
		}
	}
	return nil
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
		if err := hash.addLen(1); err != nil {
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
		if err := hash.addLen(1); err != nil {
			return 0, err
		}
	}
	return n, nil
}

// HLen returns the number of fields contained in the hash stored at key
func (hash *Hash) HLen() (int64, error) {
	if hash.isMetaSlot() {
		slot, err := hash.getAllSlot()
		if err == nil {
			return slot.Len, nil
		}
		if err != ErrKeyNotFound {
			return 0, err
		}
	}
	return hash.meta.Len, nil
}

// Object new object from hash
func (hash *Hash) Object() (*Object, error) {
	obj := hash.meta.Object
	if hash.isMetaSlot() {
		slot, err := hash.getAllSlot()
		if err == nil {
			obj.UpdatedAt = slot.UpdatedAt
		}
		if err != ErrKeyNotFound {
			return nil, err
		}
	}
	return &obj, nil
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
func (hash *Hash) HMSet(fields, values [][]byte) error {
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
	return hash.addLen(added)
}

// HMSet sets meta slot num
func (hash *Hash) HMSlot(metaSlot int64) error {

	if hash.isMetaSlot() {
		slot, err := hash.getAllSlot()

		if err == ErrKeyNotFound {
			slot = &Slot{Len: 0, UpdatedAt: Now()}
		} else if err != nil {
			return err
		}
		hash.meta.Len = slot.Len
		hash.meta.UpdatedAt = slot.UpdatedAt
	}
	if err := hash.autoUpdateSlot(metaSlot); err != nil {
		return err
	}
	hash.meta.MetaSlot = metaSlot
	if err := hash.updateMeta(); err != nil {
		return err
	}
	return nil
}

func (hash *Hash) addLen(len int64) error {
	isDirty := false
	if hash.isMetaSlot() {
		slotID := hash.calculateSlotID(hash.meta.MetaSlot)
		if err := hash.addSlotLen(slotID, len); err != nil {
			return err
		}
	} else {
		hash.meta.Len += len
		hash.meta.UpdatedAt = Now()
		if err := hash.autoUpdateSlot(defaultHashMetaSlot); err == nil {
			hash.meta.MetaSlot = defaultHashMetaSlot
		}
		isDirty = true
	}
	if isDirty || !hash.exists {
		if err := hash.updateMeta(); err != nil {
			return err
		}
	}
	return nil
}

func (hash *Hash) autoUpdateSlot(metaSlot int64) error {
	isMetaSlot := hash.isMetaSlot()
	if metaSlot < 0 {
		return ErrInteger
	}
	if metaSlot == hash.meta.MetaSlot {
		return nil
	}
	if metaSlot > hash.meta.MetaSlot {
		if !isMetaSlot && hash.meta.Len > 0 {
			slot := &Slot{Len: hash.meta.Len, UpdatedAt: Now()}
			if err := hash.updateSlot(0, slot); err != nil {
				return err
			}
		}
		return nil
	}

	if metaSlot < hash.meta.MetaSlot {
		slot, err := hash.getSliceSlot(metaSlot-1, hash.meta.MetaSlot-1)
		if err != nil {
			if err == ErrKeyNotFound {
				return nil
			}
			return err
		}
		sid := hash.calculateSlotID(metaSlot)
		if err := hash.addSlotLen(sid, slot.Len); err != nil {
			return err
		}
		if err := hash.clearSliceSlot(metaSlot, hash.meta.MetaSlot-1); err != nil {
			return err
		}
	}
	return nil
}

func (hash *Hash) clearSliceSlot(start, end int64) error {
	if start >= end || start < 0 || end < 1 {
		return ErrOutOfRange
	}
	i := start
	for i <= end {
		metaSlotKey := MetaSlotKey(hash.txn.db, hash.meta.ID, EncodeInt64(i))
		if err := hash.txn.t.Delete(metaSlotKey); err != nil {
			return err
		}
		i++
	}
	return nil
}

func (hash *Hash) addSlotLen(newID int64, len int64) error {
	slot, err := hash.getSlot(newID)
	if err != nil {
		return err
	}
	slot.Len += len
	slot.UpdatedAt = Now()
	return hash.updateSlot(newID, slot)
}

func (hash *Hash) getSlot(slotID int64) (*Slot, error) {
	metaSlotKey := MetaSlotKey(hash.txn.db, hash.meta.ID, EncodeInt64(slotID))
	raw, err := hash.txn.t.Get(metaSlotKey)
	if err != nil {
		if IsErrNotFound(err) {
			return &Slot{UpdatedAt: Now()}, nil
		}
		return nil, err
	}
	slot, err := DecodeSlot(raw)
	if err != nil {
		return nil, err
	}
	return slot, nil
}

func (hash *Hash) updateMeta() error {
	meta := hash.meta.Encode()
	return hash.txn.t.Set(MetaKey(hash.txn.db, hash.key), meta)
}

func (hash *Hash) updateSlot(slotID int64, slot *Slot) error {
	slotKey := MetaSlotKey(hash.txn.db, hash.meta.ID, EncodeInt64(slotID))
	metaSlot := EncodeSlot(slot)
	return hash.txn.t.Set(slotKey, metaSlot)
}

func (hash *Hash) getMetaSlotKeys() [][]byte {
	metaSlot := hash.meta.MetaSlot
	keys := make([][]byte, metaSlot)
	for metaSlot > 0 {
		keys = append(keys, MetaSlotKey(hash.txn.db, hash.meta.ID, EncodeInt64(metaSlot)))
		metaSlot--
	}
	return keys
}

func (hash *Hash) getAllSlot() (*Slot, error) {
	return hash.getSliceSlot(0, hash.meta.MetaSlot-1)
}

func (hash *Hash) getSliceSlot(start, end int64) (*Slot, error) {
	if start >= end {
		return nil, ErrOutOfRange
	}
	var rawSlots [][]byte
	prefixKey := MetaSlotKey(hash.txn.db, hash.meta.ID, nil)
	startKey := MetaSlotKey(hash.txn.db, hash.meta.ID, EncodeInt64(start))
	iter, err := hash.txn.t.Seek(startKey)
	if err != nil {
		return nil, err
	}
	count := start
	for iter.Valid() && iter.Key().HasPrefix(prefixKey) && count <= end {
		rawSlots = append(rawSlots, iter.Value())
		if err := iter.Next(); err != nil {
			break
		}
		count++
	}
	if len(rawSlots) > 0 {
		slot, err := hash.calculateSlot(&rawSlots)
		if err != nil {
			return nil, err
		}
		return slot, nil
	}
	return nil, ErrKeyNotFound
}

func (hash *Hash) calculateSlot(vals *[][]byte) (*Slot, error) {
	slot := &Slot{}
	for _, val := range *vals {
		if val == nil {
			continue
		}
		meta, err := DecodeSlot(val)
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
