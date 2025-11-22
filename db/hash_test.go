package db

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// compareGetString skip CreatedAt UpdatedAt ID compare
func compareGetHash(want, get *Hash) error {
	switch {
	case !bytes.Equal(want.key, get.key):
		return fmt.Errorf("set key not equal, want=%s, get=%s", string(want.key), string(get.key))
	case want.meta.ExpireAt != get.meta.ExpireAt:
		return fmt.Errorf("meta.expireAt not equal, want=%v, get=%v", want.meta.ExpireAt, get.meta.ExpireAt)
	case want.meta.Type != get.meta.Type:
		return fmt.Errorf("meta.Type not equal, want=%v, get=%v", want.meta.Type, get.meta.Type)
	case want.exists != get.exists:
		return fmt.Errorf("exists not equal, want=%v, get=%v", want.exists, get.exists)
	}
	return nil
}
func compareNewHash(want, get *Hash) error {
	switch {
	case !bytes.Equal(want.key, get.key):
		return fmt.Errorf("set key not equal, want=%s, get=%s", string(want.key), string(get.key))
	}
	return nil
}

var (
	TestHashExistKey = []byte("HashKey")
	TesyHashField    = []byte("HashField")
)

func Test_newHash(t *testing.T) {
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
		want *Hash
	}{
		{
			name: "TestNewHash",
			args: args{
				txn: txn,
				key: []byte("TestNewHash"),
			},
			want: &Hash{
				meta: &HashMeta{},
				key:  []byte("TestNewHash"),
				txn:  txn,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			assert.NotNil(t, txn)
			assert.NoError(t, err)
			got := newHash(tt.args.txn, tt.args.key)
			if err := txn.Commit(context.TODO()); err != nil {
				t.Errorf("txn.Commit error = %v", err)
			}
			if err := compareNewHash(tt.want, got); err != nil {
				t.Errorf("NewHash() = %v, want %v", got, tt.want)
			}
		})
	}
	if err := txn.Commit(context.TODO()); err != nil {
		t.Errorf("txn.Commit error = %v", err)
	}
}

func setHashMeta(t *testing.T, txn *Transaction, key []byte, metaSlot int64) error {
	h := newHash(txn, key)
	mkey := MetaKey(txn.db, key)
	hm := &HashMeta{
		Object: h.meta.Object,
	}
	meta := EncodeHashMeta(hm)
	err := txn.t.Set(mkey, meta)
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	return nil
}

//删除设置的meta信息
func destoryHashMeta(t *testing.T, txn *Transaction, key []byte) error {
	metakey := MetaKey(txn.db, key)
	if err := txn.t.Delete(metakey); err != nil {
		return err
	}
	return nil
}

func getHash(t *testing.T, key []byte) (*Hash, *Transaction, error) {
	txn, err := mockDB.Begin()
	assert.NotNil(t, txn)
	assert.NoError(t, err)
	hash, err := GetHash(txn, key)
	assert.NotNil(t, hash)
	assert.NoError(t, err)

	return hash, txn, nil
}

func TestGetHash(t *testing.T) {
	txn, err := mockDB.Begin()
	assert.NoError(t, err)
	assert.NotNil(t, txn)

	if err := setHashMeta(t, txn, []byte("TestGetHashExistKey"), 0); err != nil {
		t.Errorf("setHashMeta error = %v", err)
	}
	if err := setHashMeta(t, txn, []byte("TestGetHashSlotKey"), 100); err != nil {
		t.Errorf("setHashMeta error = %v", err)
	}
	type args struct {
		txn *Transaction
		key []byte
	}
	type want struct {
		hash *Hash
		err  error
	}
	tests := []struct {
		name    string
		args    args
		want    want
		err     error
		wantErr bool
	}{
		{
			name: "TestGetHashNoExistKey",
			args: args{
				txn: txn,
				key: []byte("TestGetHashNoExistKey"),
			},
			want: want{
				hash: &Hash{
					meta: &HashMeta{
						Object: Object{
							Type: ObjectHash,
						},
					},
					key:    []byte("TestGetHashNoExistKey"),
					exists: false,
					txn:    txn,
				},
				err: nil,
			},
			err:     nil,
			wantErr: false,
		},
		{
			name: "TestGetHashExistKey",
			args: args{
				txn: txn,
				key: []byte("TestGetHashExistKey")},
			want: want{
				hash: &Hash{meta: &HashMeta{
					Object: Object{
						Type: ObjectHash,
					},
				},
					key:    []byte("TestGetHashExistKey"),
					exists: true,
					txn:    txn,
				},
				err: nil,
			},
			err:     nil,
			wantErr: false,
		},
		{
			name: "TestGetHashSlotKey",
			args: args{
				txn: txn,
				key: []byte("TestGetHashSlotKey")},
			want: want{
				hash: &Hash{meta: &HashMeta{
					Object: Object{
						Type: ObjectHash,
					},
				},
					key:    []byte("TestGetHashSlotKey"),
					exists: true,
					txn:    txn,
				},
				err: nil,
			},
			err:     nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			if err != nil {
				t.Errorf("db.Begin error %s", err)
			}
			got, err := GetHash(tt.args.txn, tt.args.key)
			_ = err
			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("GetString() txn.Commit error = %v", err)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err := compareGetHash(tt.want.hash, got); err != nil {
				t.Errorf("GetHash() = %v, want %v", got, tt.want)
			}
			if tt.want.err != tt.err {
				t.Errorf("GetHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	if err := destoryHashMeta(t, txn, []byte("TestGetHashExistKey")); err != nil {
		t.Errorf("destoryHashMeta error = %v", err)
	}
	if err := destoryHashMeta(t, txn, []byte("TestGetHashSlotKey")); err != nil {
		t.Errorf("destoryHashMeta error = %v", err)
	}
	if err := txn.Commit(context.TODO()); err != nil {
		t.Errorf("txn.Commit error = %v", err)
	}
}

func TestHashHSet(t *testing.T) {
	type args struct {
		field []byte
		value []byte
	}
	type want struct {
		num int
		len int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "TestHashHSetNoExistKey",
			args: args{
				field: []byte("HashField"),
				value: []byte("HashValue"),
			},
			want: want{
				num: 1,
				len: 1,
			},
		},
		{
			name: "TestHashHSetExistKey",
			args: args{
				field: []byte("HashField"),
				value: []byte("HashValue2"),
			},
			want: want{
				num: 0,
				len: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, txn, err := getHash(t, []byte("TestHashHSet"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, hash)

			got, err := hash.HSet(tt.args.field, tt.args.value)
			assert.NoError(t, err)
			assert.NotNil(t, got)
			_ = txn.Commit(context.TODO())

			assert.Equal(t, got, tt.want.num)
		})
	}
}
func TestHashHDel(t *testing.T) {
	hash, txn, err := getHash(t, []byte("TestHashHDel"))
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.NotNil(t, hash)

	var fileds [][]byte
	fileds = append(fileds, []byte("TestHashDelFiled1"))
	fileds = append(fileds, []byte("TestHashDelFiled2"))
	fileds = append(fileds, []byte("TestHashDelFiled3"))

	_, _ = hash.HSet([]byte("TestHashDelFiled1"), []byte("TestDelHashValue1"))
	_, _ = hash.HSet([]byte("TestHashDelFiled2"), []byte("TestDelHashValue2"))
	_, _ = hash.HSet([]byte("TestHashDelFiled3"), []byte("TestDelHashValue3"))
	_ = txn.Commit(context.TODO())

	type args struct {
		fields [][]byte
	}

	type want struct {
		num int64
		len int64
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "TestHashDelCase1",
			args: args{
				fields: fileds[:1],
			},
			want: want{
				num: 1,
				len: 2,
			},
		},
		{
			name: "TestHashDelCase2",
			args: args{
				fields: fileds[1:],
			},
			want: want{
				num: 2,
				len: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//test hdel method
			hash, txn, err := getHash(t, []byte("TestHashHDel"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, hash)

			got, err := hash.HDel(tt.args.fields)
			assert.NoError(t, err)
			assert.NotNil(t, got)

			if err := txn.Commit(context.TODO()); err != nil {
				t.Errorf("txn.Commit error = %v", err)
			}

			assert.Equal(t, got, tt.want.num)

			//use hlen check
			hash, txn, _ = getHash(t, []byte("TestHashHDel"))
			hlen, err1 := hash.HLen()
			assert.NoError(t, err1)
			assert.Equal(t, hlen, tt.want.len)
			if err := txn.Commit(context.TODO()); err != nil {
				t.Errorf("txn.Commit error = %v", err)
			}
		})
	}
}

func TestHashHSetNX(t *testing.T) {

	type args struct {
		field []byte
		value []byte
	}
	tests := []struct {
		name string
		args args
		want int
	}{

		{
			name: "TestHash_HSetNXNoExist",
			args: args{
				field: []byte("TestHashSetNxField"),
				value: []byte("TestHashSetNxValue"),
			},
			want: 1,
		},

		{
			name: "TestHash_HSetNXExist",
			args: args{
				field: []byte("TestHashSetNxField"),
				value: []byte("TestHashSetNxValue"),
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, txn, err := getHash(t, []byte("TestHashHSetNX"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, hash)

			got, _ := hash.HSetNX(tt.args.field, tt.args.value)

			if err := txn.Commit(context.TODO()); err != nil {
				t.Errorf("txn.Commit error = %v", err)
			}
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestHash_HGet(t *testing.T) {
	hash, txn, err := getHash(t, []byte("TestHashHGet"))
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.NotNil(t, hash)

	if _, err := hash.HSet([]byte("TestHashHGetFiled"), []byte("TestHashHGetValue")); err != nil {
		t.Errorf("hash.HSet error = %v", err)
	}
	if err := txn.Commit(context.TODO()); err != nil {
		t.Errorf("txn.Commit error = %v", err)
	}
	type args struct {
		field []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "TestHashHGet",
			args: args{
				field: []byte("TestHashHGetFiled"),
			},
			want: []byte("TestHashHGetValue"),
		},
		{
			name: "TestHashHGetNoExist",
			args: args{
				field: []byte("TestHashHGetFiled1"),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			hash, txn, err := getHash(t, []byte("TestHashHGet"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, hash)

			got, _ := hash.HGet(tt.args.field)
			if err := txn.Commit(context.TODO()); err != nil {
				t.Errorf("txn.Commit error = %v", err)
			}
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestHashHGetAll(t *testing.T) {

	hash, txn, err := getHash(t, []byte("TestHashHGetAll"))
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.NotNil(t, hash)

	if _, err := hash.HSet([]byte("TestHashHGetAllFiled1"), []byte("TestHashHGetAllValue1")); err != nil {
		t.Errorf("hash.HSet error = %v", err)
	}
	if _, err := hash.HSet([]byte("TestHashHGetAllFiled2"), []byte("TestHashHGetAllValue2")); err != nil {
		t.Errorf("hash.HSet error = %v", err)
	}
	if _, err := hash.HSet([]byte("TestHashHGetAllFiled3"), []byte("TestHashHGetAllValue3")); err != nil {
		t.Errorf("hash.HSet error = %v", err)
	}
	if err := txn.Commit(context.TODO()); err != nil {
		t.Errorf("txn.Commit error = %v", err)
	}
	type want struct {
		fields [][]byte
		value  [][]byte
	}

	var fields [][]byte
	var value [][]byte

	fields = append(fields, []byte("TestHashHGetAllFiled1"))
	fields = append(fields, []byte("TestHashHGetAllFiled2"))
	fields = append(fields, []byte("TestHashHGetAllFiled3"))
	value = append(value, []byte("TestHashHGetAllValue1"))
	value = append(value, []byte("TestHashHGetAllValue2"))
	value = append(value, []byte("TestHashHGetAllValue3"))
	tests := []struct {
		name string
		want want
	}{

		{
			name: "TestHashHGetAll",
			want: want{
				fields: fields,
				value:  value,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, txn, err := getHash(t, []byte("TestHashHGetAll"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, hash)
			got, got1, err := hash.HGetAll()
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.NotNil(t, got1)
			if err := txn.Commit(context.TODO()); err != nil {
				t.Errorf("txn.Commit error = %v", err)
			}

			assert.Equal(t, got, tt.want.fields)
			assert.Equal(t, got1, tt.want.value)
		})
	}
}

func TestHash_HExists(t *testing.T) {
	hash, txn, err := getHash(t, []byte("TestHashExists"))
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.NotNil(t, hash)

	if _, err := hash.HSet([]byte("TestHashDestory"), []byte("TestHashDestoryValue1")); err != nil {
		t.Errorf("hash.HSet error = %v", err)
	}

	if err := txn.Commit(context.TODO()); err != nil {
		t.Errorf("txn.Commit error = %v", err)
	}

	type args struct {
		field []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "TestHashExists",
			args: args{
				field: []byte("TestHashDestory"),
			},
			want: true,
		},
		{
			name: "TestHashNoExists",
			args: args{
				field: []byte("TestHashNoExistsDestory"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			hash, txn, err := getHash(t, []byte("TestHashExists"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, hash)

			got, err := hash.HExists(tt.args.field)
			if err := txn.Commit(context.TODO()); err != nil {
				t.Errorf("txn.Commit error = %v", err)
			}
			assert.Equal(t, got, tt.want)
			assert.NoError(t, err)
		})
	}
}

func TestHashHIncrByFloat(t *testing.T) {
	hash, txn, err := getHash(t, []byte("TestHashHIncrByFloat"))
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.NotNil(t, hash)

	if _, err := hash.HSet([]byte("TestHashHIncrByFloat"), []byte("10.50")); err != nil {
		t.Errorf("hash.HSet error = %v", err)
	}

	if err := txn.Commit(context.TODO()); err != nil {
		t.Errorf("txn.Commit error = %v", err)
	}
	type args struct {
		field []byte
		v     float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "TestHashHIncrByFloat",
			args: args{
				field: []byte("TestHashHIncrByFloat"),
				v:     float64(0.1),
			},
			want: float64(10.6),
		},
		{
			name: "TestHashHIncrByFloat2",
			args: args{
				field: []byte("TestHashHIncrByFloat"),
				v:     float64(-5),
			},
			want: float64(5.6),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, txn, err := getHash(t, []byte("TestHashHIncrByFloat"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, hash)
			got, err := hash.HIncrByFloat(tt.args.field, tt.args.v)
			if err := txn.Commit(context.TODO()); err != nil {
				t.Errorf("txn.Commit error = %v", err)
			}
			assert.Equal(t, got, tt.want)
			assert.NoError(t, err)
		})
	}
}

func TestHashHLen(t *testing.T) {
	hash, txn, err := getHash(t, []byte("TestHashHLen"))
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.NotNil(t, hash)
	//	hashslot, err := getHash(t, []byte("TestHashSlotHLen"))
	//	hashslot.HMSlot(10)

	_, _ = hash.HSet([]byte("TestHashHlenField1"), []byte("TestHashHlenValue"))
	_, _ = hash.HSet([]byte("TestHashHlenField2"), []byte("TestHashHlenValue"))
	_, _ = hash.HSet([]byte("TestHashHlenField3"), []byte("TestHashHlenValue"))

	_ = txn.Commit(context.TODO())
	hashslot, txn, err := getHash(t, []byte("TestHashSlotHLen"))
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.NotNil(t, hashslot)
	//  hashslot, err := getHash(t, []byte("TestHashSlotHLen"))
	//  hashslot.HMSlot(10)

	_, _ = hashslot.HSet([]byte("TestHashHlenField1"), []byte("TestHashHlenValue"))

	_, _ = hashslot.HSet([]byte("TestHashHlenField2"), []byte("TestHashHlenValue"))
	_, _ = hashslot.HSet([]byte("TestHashHlenField3"), []byte("TestHashHlenValue"))

	_ = txn.Commit(context.TODO())
	//	hashslot.HSet([]byte("TestHashslotHlenField1"), []byte("TestHashHlenValue"))
	//	hashslot.HSet([]byte("TestHashslotHlenField2"), []byte("TestHashHlenValue"))
	//	hashslot.HSet([]byte("TestHashslotHlenField3"), []byte("TestHashHlenValue"))
	tests := []struct {
		name string

		want int64
	}{
		{
			name: "TestHashHLen",

			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, txn, err := getHash(t, []byte("TestHashHLen"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, hash)

			got, err := hash.HLen()
			if err := txn.Commit(context.TODO()); err != nil {
				t.Errorf("txn.Commit error = %v", err)
			}
			assert.Equal(t, got, tt.want)
			assert.NoError(t, err)

			hashslot, txn, err := getHash(t, []byte("TestHashSlotHLen"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, hashslot)

			got1, err := hashslot.HLen()
			if err := txn.Commit(context.TODO()); err != nil {
				t.Errorf("txn.Commit error = %v", err)
			}
			assert.Equal(t, got1, tt.want)
			assert.NoError(t, err)
		})
	}
}

func TestHash_HScan(t *testing.T) {
	hash, txn, err := getHash(t, []byte("TestHashHScan"))
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.NotNil(t, hash)

	if _, err := hash.HSet([]byte("TestHashHScanFiled1"), []byte("TestHashHScanValue1")); err != nil {
		t.Errorf("hash.HSet error = %v", err)
	}
	if _, err := hash.HSet([]byte("TestHashHScanFiled2"), []byte("TestHashHScanValue2")); err != nil {
		t.Errorf("hash.HSet error = %v", err)
	}
	if _, err := hash.HSet([]byte("TestHashHScanFiled3"), []byte("TestHashHScanValue3")); err != nil {
		t.Errorf("hash.HSet error = %v", err)
	}

	if err := txn.Commit(context.TODO()); err != nil {
		t.Errorf("txn.Commit error = %v", err)
	}

	type args struct {
		cursor []byte
		f      func(key, val []byte) bool
	}
	var value [][]byte
	count := 2

	tests := []struct {
		name string
		args args
		want [][]byte
	}{
		{
			name: "TestHashHScan",
			args: args{
				cursor: []byte("TestHashHScanFiled"),
				f: func(key, val []byte) bool {
					if count == 0 {
						return false
					}
					value = append(value, key, val)
					count--
					return true

				},
			},
			want: append(value, []byte("TestHashHScanFiled1"), []byte("TestHashHScanValue1"), []byte("TestHashHScanFiled2"), []byte("TestHashHScanValue2")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, txn, err := getHash(t, []byte("TestHashHScan"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, hash)

			err = hash.HScan(tt.args.cursor, tt.args.f)
			if err := txn.Commit(context.TODO()); err != nil {
				t.Errorf("txn.Commit error = %v", err)
			}

			assert.Equal(t, value, tt.want)
			assert.NoError(t, err)
		})
	}
}

func TestHash_HMGet(t *testing.T) {
	hash, txn, err := getHash(t, []byte("TestHashHMGet"))
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.NotNil(t, hash)

	if _, err := hash.HSet([]byte("TestHashHMGetFiled1"), []byte("TestHashHGetValue1")); err != nil {
		t.Errorf("hash.HSet error = %v", err)
	}
	if _, err := hash.HSet([]byte("TestHashHMGetFiled2"), []byte("TestHashHGetValue2")); err != nil {
		t.Errorf("hash.HSet error = %v", err)
	}
	if _, err := hash.HSet([]byte("TestHashHMGetFiled3"), []byte("TestHashHGetValue3")); err != nil {
		t.Errorf("hash.HSet error = %v", err)
	}

	if err := txn.Commit(context.TODO()); err != nil {
		t.Errorf("txn.Commit error = %v", err)
	}

	var fields [][]byte
	var value [][]byte

	fields = append(fields, []byte("TestHashHMGetFiled1"))
	fields = append(fields, []byte("TestHashHMGetFiled2"))
	fields = append(fields, []byte("TestHashHMGetFiled3"))
	value = append(value, []byte("TestHashHGetValue1"))
	value = append(value, []byte("TestHashHGetValue2"))
	value = append(value, []byte("TestHashHGetValue3"))
	type args struct {
		fields [][]byte
	}
	tests := []struct {
		name string
		args args
		want [][]byte
	}{
		{
			name: "TestHashHMGet",
			args: args{
				fields: fields,
			},
			want: value,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, txn, err := getHash(t, []byte("TestHashHMGet"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, hash)
			got, err := hash.HMGet(tt.args.fields)

			if err := txn.Commit(context.TODO()); err != nil {
				t.Errorf("txn.Commit error = %v", err)
			}
			assert.Equal(t, got, tt.want)
			assert.NoError(t, err)
		})
	}
}

func TestHashHMSet(t *testing.T) {

	var fields [][]byte
	var value [][]byte

	fields = append(fields, []byte("TestHashHMSetFiled1"))
	fields = append(fields, []byte("TestHashHMSetFiled2"))
	fields = append(fields, []byte("TestHashHMSetFiled3"))
	value = append(value, []byte("TestHashHSetValue1"))
	value = append(value, []byte("TestHashHSetValue2"))
	value = append(value, []byte("TestHashHSetValue3"))
	type args struct {
		fields [][]byte
		values [][]byte
	}
	type want struct {
		value [][]byte
		len   int64
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "TestHashHMSet",
			args: args{
				fields: fields,
				values: value,
			},
			want: want{
				value: value,
				len:   int64(3),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			hash, txn, err := getHash(t, []byte("TestHashHMSet"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, hash)
			err = hash.HMSet(tt.args.fields, tt.args.values)
			assert.NoError(t, err)
			if err := txn.Commit(context.TODO()); err != nil {
				t.Errorf("txn.Commit error = %v", err)
			}

			hash, txn, err = getHash(t, []byte("TestHashHMSet"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, hash)

			got, err := hash.HMGet(fields)

			assert.Equal(t, got, tt.want.value)
			assert.NoError(t, err)
		})
	}
}

func TestHashExpired(t *testing.T) {
	setExpireAt := func(t *testing.T, key []byte, expireAt int64) []byte {
		hash, txn, err := getHash(t, key)
		assert.NoError(t, err)
		assert.NotNil(t, txn)
		assert.NotNil(t, hash)
		id := hash.meta.ID
		hash.meta.ExpireAt = expireAt
		if _, err := hash.HSet([]byte("TestHashExpiredfield"), []byte("TestHashExpiredval")); err != nil {
			t.Errorf("hash.HSet error = %v", err)
		}
		if err := txn.Commit(context.TODO()); err != nil {
			t.Errorf("txn.Commit error = %v", err)
		}
		return id
	}

	ts := Now()
	type args struct {
		key      []byte
		expireAt int64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "TestHashExpired",
			args: args{
				key:      []byte("TestHashExpired"),
				expireAt: ts - 3*int64(time.Second),
			},
			want: false,
		},
		{
			name: "TestHashNotExpired",
			args: args{
				key:      []byte("TestHashNotExpired"),
				expireAt: ts + 10*int64(time.Second),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldID := setExpireAt(t, tt.args.key, tt.args.expireAt)
			hash, txn, err := getHash(t, tt.args.key)
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, hash)
			newID := hash.meta.ID
			if err := txn.Commit(context.TODO()); err != nil {
				t.Errorf("txn.Commit error = %v", err)
			}
			if tt.want {
				assert.Equal(t, newID, oldID)
			} else {
				assert.NotEqual(t, newID, oldID)
			}
		})
	}
}
