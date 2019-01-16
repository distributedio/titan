package db

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// compareSet skip CreatedAt UpdatedAt ID compare
func compareSet(want, got *Set) error {
	switch {
	case !bytes.Equal(want.key, got.key):
		return fmt.Errorf("set key not equal, want=%s, got=%s", string(want.key), string(got.key))
	case want.meta.ExpireAt != got.meta.ExpireAt:
		return fmt.Errorf("meta.ExpireAt not equal, want=%v, got=%v", want.meta.ExpireAt, got.meta.ExpireAt)
	case want.meta.Type != got.meta.Type:
		return fmt.Errorf("meta.Type not equal, want=%v, got=%v", want.meta.Type, got.meta.Type)
	case want.meta.Encoding != got.meta.Encoding:
		return fmt.Errorf("meta.Encoding not equal, want=%v, got=%v", want.meta.Encoding, got.meta.Encoding)
	case want.meta.Len != got.meta.Len:
		return fmt.Errorf("meta.Len not equal, want=%v, got=%v", want.meta.Len, got.meta.Len)
	case want.exists != got.exists:
		return fmt.Errorf("exists not equal, want=%v, get=%v", want.exists, got.exists)
	}
	return nil
}

func testAddData(t *testing.T, key []byte, values [][]byte) {
	var txn *Transaction
	var err error
	var set *Set
	if txn, err = mockDB.Begin(); err != nil {
		t.Errorf("TestGetSet db.Begin error %s", err)
	}
	if set, err = GetSet(txn, key); err != nil {
		t.Errorf("Set.SAdd() error = %v", err)
	}
	_, err = set.SAdd(values)
	if err != nil {
		t.Errorf("Set.SAdd() error = %v", err)
		return
	}
	if err = txn.Commit(context.TODO()); err != nil {
		t.Errorf("Set.SAdd() txn.Commit error = %v", err)
		return
	}
}

func setSetMeta(t *testing.T, txn *Transaction, key []byte) error {
	h := newSet(txn, key)
	mkey := MetaKey(txn.db, key)
	sm := &SetMeta{
		Object: h.meta.Object,
		Len:    1,
	}
	meta := EncodeSetMeta(sm)
	err := txn.t.Set(mkey, meta)
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	return nil
}
func destorySetMeta(t *testing.T, txn *Transaction, key []byte) error {
	mkey := MetaKey(txn.db, key)
	if err := txn.t.Delete(mkey); err != nil {
		return err
	}
	return nil
}
func Test_newSet(t *testing.T) {
	txn, err := mockDB.Begin()
	assert.NotNil(t, txn)
	assert.NoError(t, err)
	type args struct {
		txn *Transaction
		key []byte
	}
	tests := []struct {
		name string
		args args
		want *Set
	}{
		{
			name: "TestNewSet",
			args: args{
				txn: txn,
				key: []byte("TestNewSet"),
			},
			want: &Set{
				meta: &SetMeta{
					Object: Object{
						ExpireAt: 0,
						Type:     ObjectSet,
						Encoding: ObjectEncodingHT,
					},
					Len: 0,
				},
				key: []byte("TestNewSet"),
				txn: txn,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			assert.NotNil(t, txn)
			assert.NoError(t, err)
			got := newSet(tt.args.txn, tt.args.key)
			txn.Commit(context.TODO())
			if compareSet(tt.want, got); err != nil {
				t.Errorf("newSet() = %v, want %v", got, tt.want)
			}
		})
	}
	txn.Commit(context.TODO())
}

func TestGetSet(t *testing.T) {
	var testNotExistSetKey = []byte("not_exist_key")
	var testExistSetKey = []byte("exist_key")

	txn, err := mockDB.Begin()
	assert.NotNil(t, txn)
	assert.NoError(t, err)
	setSetMeta(t, txn, testExistSetKey)
	type args struct {
		txn *Transaction
		key []byte
	}
	type want struct {
		set *Set
		err *error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "not exist set",
			args: args{
				txn: txn,
				key: testNotExistSetKey,
			},
			want: want{
				set: &Set{
					key: testNotExistSetKey,
					meta: &SetMeta{
						Object: Object{
							ExpireAt: 0,
							Type:     ObjectSet,
							Encoding: ObjectEncodingHT,
						},
						Len: 0,
					},
					exists: false,
					txn:    txn,
				},
				err: nil,
			},
		},
		{
			name: "exist set",
			args: args{
				txn: txn,
				key: testExistSetKey,
			},
			want: want{
				set: &Set{
					key: testExistSetKey,
					meta: &SetMeta{
						Object: Object{
							ExpireAt: 0,
							Type:     ObjectSet,
							Encoding: ObjectEncodingHT,
						},
						Len: 1,
					},
					exists: true,
					txn:    txn,
				},
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			assert.NotNil(t, txn)
			assert.NoError(t, err)
			got, err := GetSet(tt.args.txn, tt.args.key)
			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("GetString() txn.Commit error = %v", err)
				return
			}
			if err := compareSet(tt.want.set, got); err != nil {
				t.Errorf("GetSet() = %v, want %v", got, tt.want)
			}
		})
	}
	destorySetMeta(t, txn, testExistSetKey)
	txn.Commit(context.TODO())
}

func TestSet_SAdd(t *testing.T) {
	var testSetSAddKey = []byte("set_sadd_key")
	tests := []struct {
		name    string
		key     []byte
		members [][]byte
		want    int64
	}{
		{
			name:    "empty",
			key:     testSetSAddKey,
			members: [][]byte{[]byte("value1")},
			want:    1,
		},
		{
			name:    "duplicate",
			key:     testSetSAddKey,
			members: [][]byte{[]byte("value1")},
			want:    0,
		},
		{
			name:    "set_mutil",
			key:     testSetSAddKey,
			members: [][]byte{[]byte("value2"), []byte("value3"), []byte("value4")},
			want:    3,
		},
		{
			name:    "set_duplicate_some",
			key:     testSetSAddKey,
			members: [][]byte{[]byte("value4"), []byte("value5"), []byte("value6")},
			want:    2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			assert.NoError(t, err)
			assert.NotNil(t, txn)

			set, err := GetSet(txn, tt.key)
			assert.NoError(t, err)
			assert.NotNil(t, set)

			got, err := set.SAdd(tt.members)
			assert.NoError(t, err)
			assert.NotNil(t, got)

			txn.Commit(context.TODO())

			assert.Equal(t, got, tt.want)
		})
	}
}

func TestSet_SMembers(t *testing.T) {
	var testSetSMembersKeyEmpty = []byte("set_key_empty")
	var testSetSMembersKeyOne = []byte("set_key_one")
	var testSetSMembersKeyTwo = []byte("set_key_two")
	testAddData(t, testSetSMembersKeyOne, [][]byte{[]byte("value1")})
	testAddData(t, testSetSMembersKeyTwo, [][]byte{[]byte("value1"), []byte("value2")})
	tests := []struct {
		name string
		key  []byte
		want [][]byte
	}{
		{
			name: "empty",
			key:  testSetSMembersKeyEmpty,
			want: [][]byte{},
		},
		{
			name: "one",
			key:  testSetSMembersKeyOne,
			want: [][]byte{[]byte("value1")},
		},
		{
			name: "two",
			key:  testSetSMembersKeyTwo,
			want: [][]byte{[]byte("value1"), []byte("value2")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			assert.NoError(t, err)
			assert.NotNil(t, txn)

			set, err := GetSet(txn, tt.key)
			assert.NoError(t, err)
			assert.NotNil(t, set)

			got, err := set.SMembers()
			assert.NoError(t, err)

			txn.Commit(context.TODO())

			assert.Equal(t, len(got), len(tt.want))

			for i := range got {
				assert.Equal(t, got[i], tt.want[i])
			}
		})
	}
}
