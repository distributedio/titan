package db

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"testing"
)

// compareGetString skip CreatedAt UpdatedAt ID compare
func compareGetHash(want, get *Hash) error {
	switch {
	case !bytes.Equal(want.key, get.key):
		fmt.Println("bytes.Equal(want.key, get.key):")
		return fmt.Errorf("set key not equal, want=%s, get=%s", string(want.key), string(get.key))
	case want.meta.ExpireAt != get.meta.ExpireAt:
		fmt.Println("want.meta.ExpireAt!= get.meta.ExpireAt:")
		return fmt.Errorf("meta.Type not equal, want=%v, get=%v", want.meta.ExpireAt, get.meta.ExpireAt)
	case want.meta.Len != get.meta.Len:
		fmt.Println("  want.meta.Len != get.meta.Len:")
		return fmt.Errorf("meta.Type not equal, want=%v, get=%v", want.meta.Len, get.meta.Len)
	case want.meta.MetaSlot != get.meta.MetaSlot:
		fmt.Println(" want.meta.MetaSlot != get.meta.MetaSlot:")
		return fmt.Errorf("meta.Type not equal, want=%v, get=%v", want.meta.MetaSlot, get.meta.MetaSlot)

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
func setHashValue(t *testing.T, TestKey []byte, TestField []byte, value []byte) {
	txn, err := mockDB.Begin()
	if err != nil {
		t.Errorf("db.Begin error %s", err)
	}
	h, err := GetHash(txn, TestKey)
	if err != nil {
		t.Errorf("Hash.GetHash() error = %v", err)
	}
	_, err = h.HSet(TestField, value)
	if err != nil {
		t.Errorf("String_Set() error = %v", err)
	}
	if err = txn.Commit(context.TODO()); err != nil {
		t.Errorf("Set() txn.Commit error = %v", err)
		return
	}
}

var (
	TestHashExistKey = []byte("HashKey")
	TesyHashField    = []byte("HashField")
)

func TestGetHash(t *testing.T) {
	txn, err := mockDB.Begin()
	if err != nil {
		t.Errorf("db.Begin error %s", err)
	}
	NewHash(txn, TestHashExistKey)
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
			name: "GetExistHash",
			args: args{
				txn: txn,
				key: TestHashExistKey,
			},
			want: want{
				hash: &Hash{
					meta: HashMeta{
						Len:      0,
						MetaSlot: defaultHashMetaSlot,
					},
					key:    TestHashExistKey,
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
}

func TestNewHash(t *testing.T) {
	txn, err := mockDB.Begin()
	if err != nil {
		t.Errorf("db.Begin error %s", err)
	}
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
				meta: HashMeta{
					Len:      0,
					MetaSlot: defaultHashMetaSlot,
				},
				key: []byte("TestNewHash"),
				txn: txn,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewHash(tt.args.txn, tt.args.key)
			if err := compareNewHash(tt.want, got); err != nil {
				t.Errorf("NewHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hashItemKey(t *testing.T) {
	type args struct {
		key   []byte
		field []byte
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
			if got := hashItemKey(tt.args.key, tt.args.field); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("hashItemKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_slotGC(t *testing.T) {
	type args struct {
		txn   *Transaction
		objID []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := slotGC(tt.args.txn, tt.args.objID); (err != nil) != tt.wantErr {
				t.Errorf("slotGC() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHash_calculateSlotID(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		limit int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			if got := hash.calculateSlotID(tt.args.limit); got != tt.want {
				t.Errorf("Hash.calculateSlotID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_isMetaSlot(t *testing.T) {
	type fields struct {
		meta   HashMeta
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
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			if got := hash.isMetaSlot(); got != tt.want {
				t.Errorf("Hash.isMetaSlot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_HDel(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		fields [][]byte
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
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			got, err := hash.HDel(tt.args.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash.HDel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Hash.HDel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_delHash(t *testing.T) {
	type fields struct {
		meta   HashMeta
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
		want    map[string][]byte
		want1   int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			got, got1, err := hash.delHash(tt.args.keys)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash.delHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Hash.delHash() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Hash.delHash() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestHash_HSet(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		field []byte
		value []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			got, err := hash.HSet(tt.args.field, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash.HSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Hash.HSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_HSetNX(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		field []byte
		value []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			got, err := hash.HSetNX(tt.args.field, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash.HSetNX() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Hash.HSetNX() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_HGet(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		field []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			got, err := hash.HGet(tt.args.field)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash.HGet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Hash.HGet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_HGetAll(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		want    [][]byte
		want1   [][]byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			got, got1, err := hash.HGetAll()
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash.HGetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Hash.HGetAll() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Hash.HGetAll() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestHash_Destroy(t *testing.T) {
	type fields struct {
		meta   HashMeta
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
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			if err := hash.Destroy(); (err != nil) != tt.wantErr {
				t.Errorf("Hash.Destroy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHash_HExists(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		field []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			got, err := hash.HExists(tt.args.field)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash.HExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Hash.HExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_HIncrBy(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		field []byte
		v     int64
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
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			got, err := hash.HIncrBy(tt.args.field, tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash.HIncrBy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Hash.HIncrBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_HIncrByFloat(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		field []byte
		v     float64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    float64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			got, err := hash.HIncrByFloat(tt.args.field, tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash.HIncrByFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Hash.HIncrByFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_HLen(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			got, err := hash.HLen()
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash.HLen() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Hash.HLen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_Object(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		want    *Object
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			got, err := hash.Object()
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash.Object() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Hash.Object() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_HMGet(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		fields [][]byte
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
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			got, err := hash.HMGet(tt.args.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash.HMGet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Hash.HMGet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_HMSet(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		fields [][]byte
		values [][]byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			if err := hash.HMSet(tt.args.fields, tt.args.values); (err != nil) != tt.wantErr {
				t.Errorf("Hash.HMSet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHash_HMSlot(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		metaSlot int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			if err := hash.HMSlot(tt.args.metaSlot); (err != nil) != tt.wantErr {
				t.Errorf("Hash.HMSlot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHash_addLen(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		len int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			if err := hash.addLen(tt.args.len); (err != nil) != tt.wantErr {
				t.Errorf("Hash.addLen() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHash_autoUpdateSlot(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		metaSlot int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			if err := hash.autoUpdateSlot(tt.args.metaSlot); (err != nil) != tt.wantErr {
				t.Errorf("Hash.autoUpdateSlot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHash_clearSliceSlot(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		start int64
		end   int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			if err := hash.clearSliceSlot(tt.args.start, tt.args.end); (err != nil) != tt.wantErr {
				t.Errorf("Hash.clearSliceSlot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHash_addSlotLen(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		newID int64
		len   int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			if err := hash.addSlotLen(tt.args.newID, tt.args.len); (err != nil) != tt.wantErr {
				t.Errorf("Hash.addSlotLen() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHash_getSlot(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		slotID int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Slot
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			got, err := hash.getSlot(tt.args.slotID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash.getSlot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Hash.getSlot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_updateMeta(t *testing.T) {
	type fields struct {
		meta   HashMeta
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
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			if err := hash.updateMeta(); (err != nil) != tt.wantErr {
				t.Errorf("Hash.updateMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHash_updateSlot(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		slotID int64
		slot   *Slot
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			if err := hash.updateSlot(tt.args.slotID, tt.args.slot); (err != nil) != tt.wantErr {
				t.Errorf("Hash.updateSlot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHash_getMetaSlotKeys(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	tests := []struct {
		name   string
		fields fields
		want   [][]byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			if got := hash.getMetaSlotKeys(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Hash.getMetaSlotKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_getAllSlot(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		want    *Slot
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			got, err := hash.getAllSlot()
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash.getAllSlot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Hash.getAllSlot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_getSliceSlot(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		start int64
		end   int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Slot
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			got, err := hash.getSliceSlot(tt.args.start, tt.args.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash.getSliceSlot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Hash.getSliceSlot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_calculateSlot(t *testing.T) {
	type fields struct {
		meta   HashMeta
		key    []byte
		exists bool
		txn    *Transaction
	}
	type args struct {
		vals *[][]byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Slot
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{
				meta:   tt.fields.meta,
				key:    tt.fields.key,
				exists: tt.fields.exists,
				txn:    tt.fields.txn,
			}
			got, err := hash.calculateSlot(tt.args.vals)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hash.calculateSlot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Hash.calculateSlot() = %v, want %v", got, tt.want)
			}
		})
	}
}
