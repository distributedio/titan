package db

import (
	"strconv"
)

//StringMeta string meta msg
type StringMeta struct {
	Object
	Value []byte
}

// String object operate tikv
type String struct {
	Meta StringMeta
	key  []byte
	txn  *Transaction
}

// GetString return string object ,
// if key is exist , object load meta
// otherwise object is null if key is not exist and err is not found
// otherwise  return err
func GetString(txn *Transaction, key []byte) (*String, error) {
	str := NewString(txn, key)
	mkey := MetaKey(txn.db, key)
	now := Now()
	Meta, err := txn.t.Get(mkey)
	if err != nil {
		if IsErrNotFound(err) {
			return str, nil
		}
		return nil, err
	}
	if err := str.decode(Meta); err != nil && err != ErrKeyNotFound {
		return nil, err
	}

	str.Meta.UpdatedAt = now
	return str, nil
}

// NewString  create new string object
func NewString(txn *Transaction, key []byte) *String {
	str := &String{txn: txn, key: key}
	now := Now()
	str.Meta.CreatedAt = now
	str.Meta.UpdatedAt = now
	str.Meta.ExpireAt = 0
	str.Meta.ID = UUID()
	str.Meta.Type = ObjectString
	str.Meta.Encoding = ObjectEncodingRaw
	return str
}

// Get the value information for the key from db
func (s *String) Get() ([]byte, error) {
	if !s.Exist() {
		return nil, ErrKeyNotFound
	}
	return s.Meta.Value, nil
}

// Set set the string value of a key
// the num of expire slice is not zero and expire[0] is not zero ,the key add exprie queue
// otherwise the delete expire queue
func (s *String) Set(val []byte, expire ...int64) error {
	timestamp := Now()
	mkey := MetaKey(s.txn.db, s.key)
	if len(expire) != 0 && expire[0] > 0 {
		old := s.Meta.ExpireAt
		s.Meta.ExpireAt = timestamp + expire[0]
		if err := expireAt(s.txn.t, mkey, s.Meta.ID, old, s.Meta.ExpireAt); err != nil {
			return err
		}
	} else {
		//maybe key is not expire queue,so unExpireAt will return err,but it is not relationship
		unExpireAt(s.txn.t, mkey, s.Meta.ExpireAt)
		s.Meta.ExpireAt = 0
	}
	s.Meta.Value = val
	return s.txn.t.Set(mkey, s.encode())
}

// Len value len
func (s *String) Len() (int, error) {
	return len(s.Meta.Value), nil
}

// Exist return ture if key exist
func (s *String) Exist() bool {
	if s.Meta.Value == nil {
		return false
	}
	return true
}

// Append append a value to key
func (s *String) Append(value []byte) (int, error) {
	s.Meta.Value = append(s.Meta.Value, value...)
	if err := s.txn.t.Set(MetaKey(s.txn.db, s.key), s.encode()); err != nil {
		return 0, err
	}
	return len(s.Meta.Value), nil
}

// GetSet return old value ,value replace old value
func (s *String) GetSet(value []byte) ([]byte, error) {
	v := s.Meta.Value
	if err := s.Set(value); err != nil {
		return nil, err
	}
	return v, nil
}

// GetRange return string from the absolute of start to the absolute of end
func (s *String) GetRange(start, end int) []byte {
	vlen := len(s.Meta.Value)
	if end < 0 {
		end = vlen + end
	}
	if start < 0 {
		start = vlen + start
	}
	if start > end || start > vlen || end < 0 {
		return nil
	}
	if end > vlen {
		end = vlen - 1
	}
	if start < 0 {
		start = 0
	}
	return s.Meta.Value[start : end+1]
}

// SetRange Overwrites part of the string stored at key, starting at the specified offset, for the entire length of value.
func (s *String) SetRange(offset int64, value []byte) ([]byte, error) {
	val := s.Meta.Value
	if int64(len(val)) < offset+int64(len(value)) {
		val = append(val, make([]byte, offset+int64(len(value))-int64(len(val)))...)
	}
	copy(val[offset:], value)
	if err := s.Set(val); err != nil {
		return nil, err
	}

	return val, nil
}

// Incr increment the integer value by the given amount
// the old value  must be integer
func (s *String) Incr(delta int64) (int64, error) {
	value := s.Meta.Value
	if value != nil {
		v, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return 0, ErrInteger
		}
		delta = v + delta
	}

	vs := strconv.FormatInt(delta, 10)
	if err := s.Set([]byte(vs)); err != nil {
		return 0, err
	}
	return delta, nil

}

// Incrf increment the float value by the given amount
// the old value  must be float
func (s *String) Incrf(delta float64) (float64, error) {
	value := s.Meta.Value
	if value != nil {
		v, err := strconv.ParseFloat(string(value), 64)
		if err != nil {
			return 0, ErrInteger
		}
		delta = v + delta
	}

	vs := strconv.FormatFloat(delta, 'e', -1, 64)
	if err := s.Set([]byte(vs)); err != nil {
		return 0, err
	}
	return delta, nil
}

// SetBit key offset bitvalue
// return the off postion of value
func (s *String) SetBit(offset, on int) (int, error) {
	val := s.Meta.Value
	bitoff := offset >> 3
	llen := int(bitoff) - len(val) + 1
	if llen > 0 {
		val = append(val, make([]byte, llen)...)
	}

	/* Get current values */
	byteval := int(val[bitoff])
	bit := uint(7 - (offset & 0x7))
	bitval := byteval & (1 << bit)

	/* Update byte with new bit value and return original value */
	byteval &= (^(1 << bit))
	byteval = byteval | ((on & 0x1) << bit)
	val[bitoff] = byte(byteval)
	if err := s.Set(val); err != nil {
		return 0, err
	}
	return bitval, nil
}

// GetBit key offset bitvalye
// offset / 8 > the index of value
// offset mod 8 +1
func (s *String) GetBit(offset int) (int, error) {
	val := s.Meta.Value
	bitoff := offset >> 3
	if int(bitoff) > len(val)-1 {
		return 0, nil
	}

	/* Get current values */
	byteval := int(val[bitoff])
	bit := uint(7 - (offset & 0x7))
	bitval := byteval & (1 << bit)

	return bitval, nil
}

// encode because of the value is small size , value and meta decode together
func (s *String) encode() []byte {
	b := EncodeObject(&s.Meta.Object)
	b = append(b, s.Meta.Value...)
	return b
}

// decode if obj has been existed , stop parse
func (s *String) decode(b []byte) error {
	obj, err := DecodeObject(b)
	if err != nil {
		return err
	}

	timestamp := Now()
	if obj.ExpireAt != 0 && obj.ExpireAt < timestamp {
		return ErrKeyNotFound
	}

	if obj.Type != ObjectString {
		return ErrTypeMismatch
	}

	if obj.Encoding != ObjectEncodingRaw {
		return ErrTypeMismatch
	}
	s.Meta.Object = *obj
	if len(b) > ObjectEncodingLength {
		s.Meta.Value = b[ObjectEncodingLength:]
	}
	return nil
}
