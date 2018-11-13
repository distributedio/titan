package db

import (
	"bytes"
	"context"
	"fmt"
	"testing"
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

func TestGetSet(t *testing.T) {
	var testNotExistSetKey = []byte("not_exist_key")
	var testExistSetKey = []byte("exist_key")
	var testSetValue = [][]byte{[]byte("set value")}
	testAddData(t, testExistSetKey, testSetValue)

	txn, err := mockDB.Begin()
	if err != nil {
		t.Errorf("TestGetSet db.Begin error %s", err)
	}
	type args struct {
		txn *Transaction
		key []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Set
		wantErr bool
	}{
		{
			name: "not exist set",
			args: args{
				txn: txn,
				key: testNotExistSetKey,
			},
			want: &Set{
				key: testNotExistSetKey,
				meta: SetMeta{
					Object: Object{
						ExpireAt: 0,
						Type:     ObjectSet,
						Encoding: ObjectEncodingHT,
					},
					Len: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "exist set",
			args: args{
				txn: txn,
				key: testExistSetKey,
			},
			want: &Set{
				key: testExistSetKey,
				meta: SetMeta{
					Object: Object{
						ExpireAt: 0,
						Type:     ObjectSet,
						Encoding: ObjectEncodingHT,
					},
					Len: 1,
				},
			},
			wantErr: false,
		},
		//TODO type mismatch
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSet(tt.args.txn, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err := compareSet(tt.want, got); err != nil {
				t.Errorf("%v", err)
			}
		})
	}
}

func TestSet_SAdd(t *testing.T) {
	var testSetSAddKey = []byte("set_sadd_key")
	tests := []struct {
		name    string
		key     []byte
		members [][]byte
		want    int64
		wantErr bool
	}{
		{
			name:    "empty",
			key:     testSetSAddKey,
			members: [][]byte{[]byte("value1")},
			want:    1,
			wantErr: false,
		},
		{
			name:    "duplicate",
			key:     testSetSAddKey,
			members: [][]byte{[]byte("value1")},
			want:    0,
			wantErr: false,
		},
		{
			name:    "set_mutil",
			key:     testSetSAddKey,
			members: [][]byte{[]byte("value2"), []byte("value3"), []byte("value4")},
			want:    3,
			wantErr: false,
		},
		{
			name:    "set_duplicate_some",
			key:     testSetSAddKey,
			members: [][]byte{[]byte("value4"), []byte("value5"), []byte("value6")},
			want:    2,
			wantErr: false,
		},
	}
	var txn *Transaction
	var err error
	var set *Set
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if txn, err = mockDB.Begin(); err != nil {
				t.Errorf("TestGetSet db.Begin error %s", err)
			}
			if set, err = GetSet(txn, tt.key); err != nil {
				t.Errorf("Set.SAdd() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := set.SAdd(tt.members)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set.SAdd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("Set.SAdd() txn.Commit error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Set.SAdd() = %v, want %v", got, tt.want)
			}
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
		name    string
		key     []byte
		want    [][]byte
		wantErr bool
	}{
		{
			name:    "empty",
			key:     testSetSMembersKeyEmpty,
			want:    [][]byte{},
			wantErr: false,
		},
		{
			name:    "one",
			key:     testSetSMembersKeyOne,
			want:    [][]byte{[]byte("value1")},
			wantErr: false,
		},
		{
			name:    "two",
			key:     testSetSMembersKeyTwo,
			want:    [][]byte{[]byte("value1"), []byte("value2")},
			wantErr: false,
		},
	}
	var txn *Transaction
	var err error
	var set *Set
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if txn, err = mockDB.Begin(); err != nil {
				t.Errorf("TestGetSet db.Begin error %s", err)
			}
			if set, err = GetSet(txn, tt.key); err != nil {
				t.Errorf("Set.Smembers() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := set.SMembers()
			if (err != nil) != tt.wantErr {
				t.Errorf("Set.Smembers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("Set.Smembers() txn.Commit error = %v", err)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("Set.Smembers() = %v, want %v", len(got), len(tt.want))
			}
			for i := range got {
				if !bytes.Equal(got[i], tt.want[i]) {
					t.Errorf("Set.Smembers() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
