package db

import (
	"encoding/json"
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

	mkey := MetaKey(txn.db, key)
	meta, err := txn.txn.Get(mkey)
	if err != nil {
		if IsErrNotFound(err) {
			now := Now()
			str.meta.CreatedAt = now
			str.meta.UpdatedAt = now
			str.meta.ExpireAt = 0
			str.meta.ID = key
			str.meta.Type = ObjectString
			str.meta.Encoding = ObjectEncodingRaw
			return str, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(meta, &str.meta); err != nil {
		return nil, err
	}
	if str.meta.Encoding != ObjectEncodingRaw {
		return nil, ErrTypeMismatch
	}
	return str, nil
}

func (s *String) Get() ([]byte, error) {
	if len(s.meta.Value) == 0 {
		return nil, ErrKeyNotFound
	}
	return s.meta.Value, nil
}

func (s *String) Set(val []byte, at ...int64) error {
	if len(at) != 0 {
		old := s.meta.ExpireAt
		s.meta.ExpireAt = at[0]
		if err := expireAt(s.txn, s.key, s.key, old, at[0]); err != nil {
			return err
		}
	}
	s.meta.Value = val
	return s.updateMeta()
}

func (s *String) Len() (int, error) {
	return len(s.meta.Value), nil
}

func (s *String) updateMeta() error {
	meta, err := json.Marshal(s.meta)
	if err != nil {
		return err
	}
	return s.txn.txn.Set(MetaKey(s.txn.db, s.key), meta)
}
func (s *String) Exist() bool {
	if s.meta.Value == nil {
		return false
	}
	return true
}
