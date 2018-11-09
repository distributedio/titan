package db

import (
	"strconv"
)

type StringMeta struct {
	Object
	Value []byte
}

type String struct {
	meta StringMeta
	key  []byte
	txn  *Transaction
}

func GetString(txn *Transaction, key []byte) (*String, error) {
	str := &String{txn: txn, key: key}
	now := Now()

	mkey := MetaKey(txn.db, key)
	meta, err := txn.t.Get(mkey)
	if err != nil {
		if IsErrNotFound(err) {
			str.meta.CreatedAt = now
			str.meta.UpdatedAt = now
			str.meta.ExpireAt = 0
			str.meta.ID = UUID()
			str.meta.Type = ObjectString
			str.meta.Encoding = ObjectEncodingRaw
			return str, nil
		}
		return nil, err
	}
	if err := str.decode(meta); err != nil {
		return nil, err
	}

	if str.meta.Type != ObjectString {
		return nil, ErrTypeMismatch
	}

	if str.meta.Encoding != ObjectEncodingRaw {
		return nil, ErrTypeMismatch
	}

	str.meta.UpdatedAt = now

	return str, nil
}

//NewString  create new string object
func NewString(txn *Transaction, key []byte) *String {
	str := &String{txn: txn, key: key}
	now := Now()
	str.meta.CreatedAt = now
	str.meta.UpdatedAt = now
	str.meta.ExpireAt = 0
	str.meta.ID = UUID()
	str.meta.Type = ObjectString
	str.meta.Encoding = ObjectEncodingRaw
	return str
}

//Gets the value information for the key from db
func (s *String) Get() ([]byte, error) {
	if !s.Exist() {
		return nil, ErrKeyNotFound
	}
	return s.meta.Value, nil
}

func (s *String) Set(val []byte, expire ...int64) error {
	timestamp := Now()
	if len(expire) != 0 && expire[0] > 0 {
		old := s.meta.ExpireAt
		s.meta.ExpireAt = timestamp + expire[0]
		if err := expireAt(s.txn, s.key, s.key, old, s.meta.ExpireAt); err != nil {
			return err
		}
	} else {
		s.meta.ExpireAt = 0
	}
	s.meta.Value = val
	return s.txn.t.Set(MetaKey(s.txn.db, s.key), s.encode())
}

//Len value len
func (s *String) Len() (int, error) {
	if !s.Exist() {
		return 0, ErrKeyNotFound
	}
	return len(s.meta.Value), nil
}

//Exist return ture if key exist
func (s *String) Exist() bool {
	if s.meta.Value == nil {
		return false
	}
	return true
}

func (s *String) Append(value []byte) (int, error) {
	s.meta.Value = append(s.meta.Value, value...)
	s.meta.ExpireAt = 0
	if err := s.txn.t.Set(MetaKey(s.txn.db, s.key), s.encode()); err != nil {
		return 0, err
	}
	return len(s.meta.Value), nil
}

func (s *String) GetSet(value []byte) ([]byte, error) {
	v := s.meta.Value
	if err := s.Set(value); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *String) GetRange(start, end int) []byte {
	vlen := len(s.meta.Value)
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
		end = vlen
	}
	if start < 0 {
		start = 0
	}
	return s.meta.Value[start:][:end+1]
}

//TODO bug
func (s *String) SetRange(offset int64, value []byte) error {
	/*
		vlen := len(value)
		if vlen < offset+len(ctx.Args[2]) {
			value = append(value, make([]byte, len(ctx.Args[2])+offset-vlen)...)
		}
		copy(value[offset:], ctx.Args[2])
	*/
	return s.Set(value)
}

func (s *String) Incr(delta int64) (int64, error) {
	value := s.meta.Value
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

func (s *String) Incrf(delta float64) (float64, error) {
	value := s.meta.Value
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

func (s *String) encode() []byte {
	b := EncodeObject(&s.meta.Object)
	b = append(b, s.meta.Value...)
	return b
}

func (s *String) decode(b []byte) error {
	obj, err := DecodeObject(b)
	if err != nil {
		return err
	}

	timestamp := Now()
	if obj.ExpireAt != 0 && obj.ExpireAt < timestamp {
		return nil
	}

	s.meta.Object = *obj
	if len(b) > ObjectEncodingLength {
		s.meta.Value = b[ObjectEncodingLength:]
	}
	return nil
}
