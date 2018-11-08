package db

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
	str.meta.ExpireAt = 0
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
	if len(expire) != 0 {
		old := s.meta.ExpireAt
		s.meta.ExpireAt = timestamp + expire[0]
		if err := expireAt(s.txn, s.key, s.key, old, s.meta.ExpireAt); err != nil {
			return err
		}
	}
	s.meta.Value = val
	return s.txn.t.Set(MetaKey(s.txn.db, s.key), s.encode())
}

//Len value len
func (s *String) Len() (int, error) {
	return len(s.meta.Value), nil
}

//Exist return ture if key exist
func (s *String) Exist() bool {
	if s.meta.Value == nil {
		return false
	}
	return true
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
