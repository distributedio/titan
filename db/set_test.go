package db

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
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
		{
			name:    "set_duplicate_some",
			key:     testSetSAddKey,
			members: [][]byte{[]byte("value4"), []byte("value4"), []byte("value4"), []byte("value5"), []byte("value6")},
			want:    0,
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

func TestDecodeSetMeta(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *SetMeta
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeSetMeta(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeSetMeta() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecodeSetMeta() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodeSetMeta(t *testing.T) {
	type args struct {
		meta *SetMeta
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeSetMeta(tt.args.meta); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EncodeSetMeta() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setItemKey(t *testing.T) {
	type args struct {
		key    []byte
		member []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setItemKey(tt.args.key, tt.args.member); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("setItemKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSet_updateMeta(t *testing.T) {
	type fields struct {
		meta   *SetMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := &Set{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			if err := set.updateMeta(); (err != nil) != tt.wantErr {
				t.Errorf("Set.updateMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_removeRepByMap(t *testing.T) {
	type args struct {
		members [][]byte
	}
	tests := []struct {
		name string
		args args
		want [][]byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveRepByMap(tt.args.members); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeRepByMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSet_Exists(t *testing.T) {
	type fields struct {
		meta   *SetMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := &Set{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			if got := set.Exists(); got != tt.want {
				t.Errorf("Set.Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSet_SCard(t *testing.T) {
	var testSetSCardKey = []byte("SCardKey")
	testAddData(t, testSetSCardKey, [][]byte{[]byte("ExistsValue1")})
	testAddData(t, testSetSCardKey, [][]byte{[]byte("ExistsValue2")})
	tests := []struct {
		name string
		key  []byte
		want int64
	}{
		{
			name: "SCardKey",
			key:  testSetSCardKey,
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			assert.NotNil(t, txn)
			assert.NoError(t, err)

			set, err := GetSet(txn, tt.key)
			assert.NoError(t, err)
			assert.NotNil(t, set)

			got, err := set.SCard()
			assert.NotNil(t, got)
			assert.NoError(t, err)

			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("SIsmember() txn.Commit error = %v", err)
				return
			}
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestSet_SIsmember(t *testing.T) {
	var testSetSIsMembersKey = []byte("SIsmemberKey")
	testAddData(t, testSetSIsMembersKey, [][]byte{[]byte("ExistsValue")})
	type args struct {
		member []byte
	}
	tests := []struct {
		name string
		key  []byte
		args args
		want int64
	}{
		{
			name: "testExistMember",
			key:  testSetSIsMembersKey,
			args: args{
				member: []byte("ExistsValue"),
			},
			want: 1,
		},
		{
			name: "testExistMember",
			key:  testSetSIsMembersKey,
			args: args{
				member: []byte("NoExistsValue"),
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			assert.NotNil(t, txn)
			assert.NoError(t, err)

			set, err := GetSet(txn, tt.key)
			assert.NoError(t, err)
			assert.NotNil(t, set)

			got, err := set.SIsmember(tt.args.member)
			assert.NotNil(t, got)
			assert.NoError(t, err)

			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("SIsmember() txn.Commit error = %v", err)
				return
			}
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestSet_SPop(t *testing.T) {
	var testSPopKey = []byte("SPopKey")
	testAddData(t, testSPopKey, [][]byte{[]byte("1"), []byte("2"), []byte("3"), []byte("4"), []byte("5")})
	type args struct {
		count int64
	}
	tests := []struct {
		name             string
		key              []byte
		args             args
		wantMembersCount int64
	}{
		{
			name: "TestSPopZero",
			key:  testSPopKey,
			args: args{
				count: 0,
			},
			wantMembersCount: 1,
		},
		{
			name: "TestSPopNotZero",
			key:  testSPopKey,
			args: args{
				count: 2,
			},
			wantMembersCount: 2,
		},
		{
			name: "TestSPopBigCount",
			key:  testSPopKey,
			args: args{
				count: 6,
			},
			wantMembersCount: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			assert.NotNil(t, txn)
			assert.NoError(t, err)

			set, err := GetSet(txn, tt.key)
			assert.NoError(t, err)
			assert.NotNil(t, set)

			gotMembers, err := set.SPop(tt.args.count)
			assert.NoError(t, err)
			assert.NotNil(t, gotMembers)

			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("SIsmember() txn.Commit error = %v", err)
				return
			}
			assert.Equal(t, int64(len(gotMembers)), tt.wantMembersCount)
		})
	}
}

func TestSRem(t *testing.T) {
	var testSRemKey = []byte("testSRemKey")
	testAddData(t, testSRemKey, [][]byte{[]byte("1"), []byte("2"), []byte("3"), []byte("4"), []byte("5")})
	type args struct {
		members [][]byte
	}
	m1 := [][]byte{[]byte("1")}
	m2 := [][]byte{[]byte("4"), []byte("2"), []byte("3")}
	tests := []struct {
		name string
		key  []byte
		args args
		want int64
	}{
		{
			name: "testSRemOneMember",
			key:  testSRemKey,
			args: args{
				members: m1,
			},
			want: 1,
		},
		{
			name: "testSRemMoreMember",
			key:  testSRemKey,
			args: args{
				members: m2,
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			assert.NotNil(t, txn)
			assert.NoError(t, err)

			set, err := GetSet(txn, tt.key)
			assert.NoError(t, err)
			assert.NotNil(t, set)

			got, err := set.SRem(tt.args.members)
			assert.NoError(t, err)
			assert.NotNil(t, got)

			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("SIsmember() txn.Commit error = %v", err)
				return
			}
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestSet_SRem(t *testing.T) {
	type fields struct {
		meta   *SetMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		members [][]byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := &Set{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			got, err := set.SRem(tt.args.members)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set.SRem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Set.SRem() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
func TestSet_SInter(t *testing.T) {
	type fields struct {
		meta   *SetMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		keys [][]byte
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantMembers [][]byte
		wantErr     bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := &Set{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			gotMembers, err := set.SInter(tt.args.keys...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set.SInter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotMembers, tt.wantMembers) {
				t.Errorf("Set.SInter() = %v, want %v", gotMembers, tt.wantMembers)
			}
		})
	}
}

func TestSet_SDiff(t *testing.T) {
	type fields struct {
		meta   *SetMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		keys [][]byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    [][]byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := &Set{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			got, err := set.SDiff(tt.args.keys...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set.SDiff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Set.SDiff() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSet_SUion(t *testing.T) {
	var testSUionKey = []byte("testSUionKey")
	testAddData(t, testSUionKey, [][]byte{[]byte("2"), []byte("3"), []byte("4"), []byte("5")})

	var testKey1 = []byte("testSUionKey1")
	testAddData(t, testKey1, [][]byte{[]byte("1"), []byte("2")})

	var testKey2 = []byte("testSUionKey2")
	testAddData(t, testKey2, [][]byte{[]byte("1"), []byte("4"), []byte("5")})

	m1 := [][]byte{testKey1, testKey2}
	m2 := [][]byte{testKey2}
	type args struct {
		keys [][]byte
	}
	tests := []struct {
		name string
		key  []byte
		args args
		want [][]byte
	}{
		{
			name: "testSUionKey_1",
			key:  testSUionKey,
			args: args{
				keys: m1,
			},
			want: [][]byte{[]byte("2"), []byte("3"), []byte("4"), []byte("5"), []byte("1")},
		},
		{
			name: "testSUionKey_2",
			key:  testSUionKey,
			args: args{
				keys: m2,
			},
			want: [][]byte{[]byte("2"), []byte("3"), []byte("4"), []byte("5"), []byte("1")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			assert.NotNil(t, txn)
			assert.NoError(t, err)

			set, err := GetSet(txn, tt.key)
			assert.NoError(t, err)
			assert.NotNil(t, set)
			got, err := set.SUion(tt.args.keys...)
			assert.NoError(t, err)
			assert.NotNil(t, got)

			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("SIsmember() txn.Commit error = %v", err)
				return
			}
			assert.Equal(t, got, tt.want)
		})
	}
}
*/
