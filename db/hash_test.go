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
	case want.meta.Len != get.meta.Len:
		return fmt.Errorf("meta.Len not equal, want=%v, get=%v", want.meta.Len, get.meta.Len)
	case want.meta.MetaSlot != get.meta.MetaSlot:
		return fmt.Errorf("meta.MeyaSlot not equal, want=%v, get=%v", want.meta.MetaSlot, get.meta.MetaSlot)
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
	case want.meta.Len != get.meta.Len:
		return fmt.Errorf("meta.Len not equal, want=%v, get=%v", want.meta.Len, get.meta.Len)
	case want.meta.MetaSlot != get.meta.MetaSlot:
		return fmt.Errorf("meta.Type not equal, want=%v, get=%v", want.meta.MetaSlot, get.meta.MetaSlot)
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
				meta: &HashMeta{
					Len:      0,
					MetaSlot: 0,
				},
				key: []byte("TestNewHash"),
				txn: txn,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			assert.NotNil(t, txn)
			assert.NoError(t, err)
			got := newHash(tt.args.txn, tt.args.key)
			txn.Commit(context.TODO())
			if err := compareNewHash(tt.want, got); err != nil {
				t.Errorf("NewHash() = %v, want %v", got, tt.want)
			}
		})
	}
	txn.Commit(context.TODO())
}

func setHashMeta(t *testing.T, txn *Transaction, key []byte, metaSlot int64) error {
	h := newHash(txn, key)
	mkey := MetaKey(txn.db, key)
	hm := &HashMeta{
		Object:   h.meta.Object,
		Len:      1,
		MetaSlot: metaSlot,
	}
	meta := EncodeHashMeta(hm)
	err := txn.t.Set(mkey, meta)
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	return nil
}

func getHashMeta(t *testing.T, txn *Transaction, key []byte) *HashMeta {
	mkey := MetaKey(txn.db, key)
	rawMeta, err := txn.t.Get(mkey)
	assert.NoError(t, err)
	meta, err1 := DecodeHashMeta(rawMeta)
	assert.NoError(t, err1)
	return meta
}

func setSlotMeta(t *testing.T, txn *Transaction, key []byte, slotID int64) error {
	h := newHash(txn, key)
	slotKey := MetaSlotKey(txn.db, h.meta.ID, EncodeInt64(slotID))
	slot := &Slot{
		Len:       11,
		UpdatedAt: Now(),
	}
	metaSlot := EncodeSlot(slot)
	err := txn.t.Set(slotKey, metaSlot)
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

//删除设置的slotMeta信息
func destorySlotMeta(t *testing.T, txn *Transaction, key []byte, SlotID int64) error {
	hash := newHash(txn, key)
	metaSlotKey := MetaSlotKey(hash.txn.db, hash.meta.ID, EncodeInt64(SlotID))
	if err := hash.txn.t.Delete(metaSlotKey); err != nil {
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
func compareKvMap(t *testing.T, get, want map[string][]byte) error {
	switch {
	case !bytes.Equal(want["TestHashdelHashFiled1"], get["TestHashdelHashFiled1"]):
		return fmt.Errorf("set key not equal, want=%s, get=%s", string(want["TestHashdelHashFiled1"]), string(get["TestHashdelHashFiled1"]))
	case !bytes.Equal(want["TestHashdelHashFiled2"], get["TestHashdelHashFiled2"]):
		return fmt.Errorf("set key not equal, want=%s, get=%s", string(want["TestHashdelHashFiled2"]), string(get["Tes  tHashdelHashFiled2"]))
	case !bytes.Equal(want["TestHashdelHashFiled3"], get["TestHashdelHashFiled3"]):
		return fmt.Errorf("set key not equal, want=%s, get=%s", string(want["TestHashdelHashFiled3"]), string(get["Tes    tHashdelHashFiled3"]))
	}
	return nil
}
func TestGetHash(t *testing.T) {
	txn, err := mockDB.Begin()
	assert.NoError(t, err)
	assert.NotNil(t, txn)

	setHashMeta(t, txn, []byte("TestGetHashExistKey"), 0)
	setHashMeta(t, txn, []byte("TestGetHashSlotKey"), 100)
	setSlotMeta(t, txn, []byte("TestGetHashSlotKey"), 13)
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
						Len:      0,
						MetaSlot: 0,
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
					Len:      1,
					MetaSlot: 0,
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
					Len:      1,
					MetaSlot: 100,
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
	destoryHashMeta(t, txn, []byte("TestGetHashExistKey"))
	destoryHashMeta(t, txn, []byte("TestGetHashSlotKey"))
	destorySlotMeta(t, txn, []byte("TestGetHashSlotKey"), 13)
	txn.Commit(context.TODO())
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
			txn.Commit(context.TODO())

			assert.Equal(t, got, tt.want.num)
			assert.Equal(t, hash.meta.Len, int64(tt.want.len))
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

	hash.HSet([]byte("TestHashDelFiled1"), []byte("TestDelHashValue1"))
	hash.HSet([]byte("TestHashDelFiled2"), []byte("TestDelHashValue2"))
	hash.HSet([]byte("TestHashDelFiled3"), []byte("TestDelHashValue3"))
	txn.Commit(context.TODO())

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

			txn.Commit(context.TODO())

			assert.Equal(t, got, tt.want.num)

			//use hlen check
			hash, txn, err = getHash(t, []byte("TestHashHDel"))
			hlen, err1 := hash.HLen()
			assert.NoError(t, err1)
			assert.Equal(t, hlen, tt.want.len)
			txn.Commit(context.TODO())
		})
	}
}

func TestHashdelHash(t *testing.T) {

	hash, txn, err := getHash(t, []byte("TestHashdelHash"))
	assert.NotNil(t, txn)
	assert.NoError(t, err)
	assert.NotNil(t, hash)

	hash.HSet([]byte("TestHashdelHashFiled1"), []byte("TestHashdelHashValue1"))
	hash.HSet([]byte("TestHashdelHashFiled2"), []byte("TestHashdelHashValue2"))
	hash.HSet([]byte("TestHashdelHashFiled3"), []byte("TestHashdelHashValue3"))

	txn.Commit(context.TODO())

	var fileds [][]byte

	var keys [][]byte
	fileds = append(fileds, []byte("TestHashdelHashFiled1"))
	fileds = append(fileds, []byte("TestHashdelHashFiled2"))
	fileds = append(fileds, []byte("TestHashdelHashFiled3"))

	dkey := DataKey(hash.txn.db, hash.meta.ID)
	for _, field := range fileds {
		keys = append(keys, hashItemKey(dkey, field))
	}

	type args struct {
		keys [][]byte
	}
	type want struct {
		kvMap map[string][]byte
		len   int64
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "TestHashdelHash",
			args: args{
				keys: keys,
			},
			want: want{
				kvMap: map[string][]byte{
					"TestHashdelHashFiled1": []byte("TestHashdelHashValue1"),
					"TestHashdelHashFiled2": []byte("TestHashdelHashValue2"),
					"TestHashdelHashFiled3": []byte("TestHashdelHashValue3"),
				},
				len: int64(3),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, txn, err := getHash(t, []byte("TestHashdelHash"))
			assert.NotNil(t, txn)
			assert.NoError(t, err)
			assert.NotNil(t, hash)
			got, got1, err := hash.getHashFieldAndLength(tt.args.keys)
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.NotNil(t, got1)

			txn.Commit(context.TODO())
			compareKvMap(t, got, tt.want.kvMap)
			assert.Equal(t, got1, tt.want.len)
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

			got, err := hash.HSetNX(tt.args.field, tt.args.value)

			txn.Commit(context.TODO())
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestHash_HGet(t *testing.T) {
	hash, txn, err := getHash(t, []byte("TestHashHGet"))
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.NotNil(t, hash)

	hash.HSet([]byte("TestHashHGetFiled"), []byte("TestHashHGetValue"))
	txn.Commit(context.TODO())
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

			got, err := hash.HGet(tt.args.field)
			txn.Commit(context.TODO())
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestHashHGetAll(t *testing.T) {

	hash, txn, err := getHash(t, []byte("TestHashHGetAll"))
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.NotNil(t, hash)

	hash.HSet([]byte("TestHashHGetAllFiled1"), []byte("TestHashHGetAllValue1"))
	hash.HSet([]byte("TestHashHGetAllFiled2"), []byte("TestHashHGetAllValue2"))
	hash.HSet([]byte("TestHashHGetAllFiled3"), []byte("TestHashHGetAllValue3"))
	txn.Commit(context.TODO())
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
			txn.Commit(context.TODO())

			assert.Equal(t, got, tt.want.fields)
			assert.Equal(t, got1, tt.want.value)
		})
	}
}

func TestHashDestroy(t *testing.T) {

	hash, txn, err := getHash(t, []byte("TestHashDestory"))
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.NotNil(t, hash)

	hash.HSet([]byte("TestHashDestoryfiled"), []byte("TestHashDestoryValue1"))
	txn.Commit(context.TODO())
	type want struct {
		value []byte
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "TestHashDestory",
			want: want{
				value: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, txn, err := getHash(t, []byte("TestHashDestory"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, hash)

			err = hash.Destroy()
			assert.NoError(t, err)

			txn.Commit(context.TODO())
			hash, txn, err = getHash(t, []byte("TestHashDestory"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, hash)

			got, err := hash.HGet([]byte("TestHashDestoryfiled"))
			txn.Commit(context.TODO())
			assert.Equal(t, got, tt.want.value)
			assert.NoError(t, err)
		})
	}
}
func TestHash_HExists(t *testing.T) {
	hash, txn, err := getHash(t, []byte("TestHashExists"))
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.NotNil(t, hash)

	hash.HSet([]byte("TestHashDestory"), []byte("TestHashDestoryValue1"))

	txn.Commit(context.TODO())

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
			txn.Commit(context.TODO())
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

	hash.HSet([]byte("TestHashHIncrByFloat"), []byte("10.50"))

	txn.Commit(context.TODO())
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
			txn.Commit(context.TODO())
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

	hash.HSet([]byte("TestHashHlenField1"), []byte("TestHashHlenValue"))
	hash.HSet([]byte("TestHashHlenField2"), []byte("TestHashHlenValue"))
	hash.HSet([]byte("TestHashHlenField3"), []byte("TestHashHlenValue"))

	txn.Commit(context.TODO())
	hashslot, txn, err := getHash(t, []byte("TestHashSlotHLen"))
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.NotNil(t, hashslot)
	//  hashslot, err := getHash(t, []byte("TestHashSlotHLen"))
	//  hashslot.HMSlot(10)

	hashslot.HSet([]byte("TestHashHlenField1"), []byte("TestHashHlenValue"))

	hashslot.HSet([]byte("TestHashHlenField2"), []byte("TestHashHlenValue"))
	hashslot.HSet([]byte("TestHashHlenField3"), []byte("TestHashHlenValue"))

	txn.Commit(context.TODO())
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
			txn.Commit(context.TODO())
			assert.Equal(t, got, tt.want)
			assert.NoError(t, err)

			hashslot, txn, err := getHash(t, []byte("TestHashSlotHLen"))
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, hashslot)

			got1, err := hashslot.HLen()
			txn.Commit(context.TODO())
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

	hash.HSet([]byte("TestHashHScanFiled1"), []byte("TestHashHScanValue1"))
	hash.HSet([]byte("TestHashHScanFiled2"), []byte("TestHashHScanValue2"))
	hash.HSet([]byte("TestHashHScanFiled3"), []byte("TestHashHScanValue3"))

	txn.Commit(context.TODO())

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
			txn.Commit(context.TODO())

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

	hash.HSet([]byte("TestHashHMGetFiled1"), []byte("TestHashHGetValue1"))
	hash.HSet([]byte("TestHashHMGetFiled2"), []byte("TestHashHGetValue2"))
	hash.HSet([]byte("TestHashHMGetFiled3"), []byte("TestHashHGetValue3"))

	txn.Commit(context.TODO())

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

			txn.Commit(context.TODO())
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
			txn.Commit(context.TODO())
			got, err := hash.HMGet(fields)

			assert.Equal(t, got, tt.want.value)
			assert.Equal(t, hash.meta.Len, tt.want.len)
			assert.NoError(t, err)
		})
	}
}

func TestHashHMSlot(t *testing.T) {
	key := []byte("TestHashHMSlot")
	hash, txn, err := getHash(t, key)
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.NotNil(t, hash)

	hash.HSet([]byte("TestHashMSlotFiled1"), []byte("TestHashHMSlotValue1"))
	txn.Commit(context.TODO())

	type args struct {
		metaSlot int64
	}

	type want struct {
		len int64
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "TestHashHMSlot",
			args: args{
				metaSlot: int64(10),
			},
			want: want{
				len: int64(10),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, txn, err := getHash(t, key)
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.NotNil(t, hash)

			err = hash.HMSlot(tt.args.metaSlot)
			assert.NoError(t, err)
			txn.Commit(context.TODO())
			got := hash.meta.MetaSlot
			assert.Equal(t, got, tt.want.len)

			txn, err = mockDB.Begin()
			assert.NotNil(t, txn)
			assert.NoError(t, err)
			meta := getHashMeta(t, txn, key)
			assert.Equal(t, tt.args.metaSlot, meta.MetaSlot)
			txn.Commit(context.TODO())
		})
	}
}

func TestHashUpdateMeta(t *testing.T) {
	txn, err := mockDB.Begin()
	assert.NoError(t, err)
	assert.NotNil(t, txn)

	key := []byte("TestHashUpdateMeta")
	hash := newHash(txn, key)
	hash.HSet([]byte("TestHashdelHashFiled1"), []byte("TestHashdelHashValue1"))
	txn.Commit(context.TODO())

	type want struct {
		metaSlot int64
	}
	type args struct {
		metaSlot int64
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "TestHashUpdateMetaCase1",
			args: args{
				metaSlot: int64(256),
			},
			want: want{
				metaSlot: int64(256),
			},
		},
		{
			name: "TestHashUpdateMetaCase2",
			args: args{
				metaSlot: int64(500),
			},
			want: want{
				metaSlot: int64(256),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err = mockDB.Begin()
			assert.NoError(t, err)
			assert.NotNil(t, txn)

			txn.db.conf.Hash.MetaSlot = tt.args.metaSlot
			hash, err = GetHash(txn, key)
			assert.NoError(t, err)
			assert.NotNil(t, hash)
			hash.HSet([]byte(tt.name), []byte("TestHashdelHashValue1"))
			txn.Commit(context.TODO())

			txn, err = mockDB.Begin()
			assert.NotNil(t, txn)
			assert.NoError(t, err)
			meta := getHashMeta(t, txn, key)
			assert.Equal(t, tt.want.metaSlot, meta.MetaSlot)
			txn.Commit(context.TODO())
		})
	}
}

func TestHashExpired(t *testing.T) {
	setExpireAt := func(t *testing.T, key []byte, expireAt int64) []byte {
		hash, txn, err := getHash(t, key)
		assert.NoError(t, err)
		assert.NotNil(t, txn)
		assert.NotNil(t, hash)
		id := hash.meta.Object.ID
		hash.meta.Object.ExpireAt = expireAt
		hash.HSet([]byte("TestHashExpiredfield"), []byte("TestHashExpiredval"))
		txn.Commit(context.TODO())
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
			newID := hash.meta.Object.ID
			txn.Commit(context.TODO())
			if tt.want {
				assert.Equal(t, newID, oldID)
			} else {
				assert.NotEqual(t, newID, oldID)
			}
		})
	}
}
