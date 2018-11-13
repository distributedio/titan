package db

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

// compareSet skip CreatedAt UpdatedAt ID compare
func compareSet(want, get *Set) error {
	switch {
	case !bytes.Equal(want.key, get.key):
		return fmt.Errorf("set key not equal, want=%s, get=%s", string(want.key), string(get.key))
	case want.meta.ExpireAt != get.meta.ExpireAt:
		return fmt.Errorf("meta.ExpireAt not equal, want=%v, get=%v", want.meta.ExpireAt, get.meta.ExpireAt)
	case want.meta.Type != get.meta.Type:
		return fmt.Errorf("meta.Type not equal, want=%v, get=%v", want.meta.Type, get.meta.Type)
	case want.meta.Encoding != get.meta.Encoding:
		return fmt.Errorf("meta.Encoding not equal, want=%v, get=%v", want.meta.Encoding, get.meta.Encoding)
	case want.meta.Len != get.meta.Len:
		return fmt.Errorf("meta.Len not equal, want=%v, get=%v", want.meta.Len, want.meta.Len)
	}
	return nil
}

func TestGetSet(t *testing.T) {
	txn, err := mockDB.Begin()
	if err != nil {
		t.Errorf("TestGetSet db.Begin error %s", err)
	}
	var notExistSetKey = []byte("not_exist_key")
	var existSetKey = []byte("exist_key")
	var setValue = [][]byte{[]byte("set value")}
	set, err := GetSet(txn, existSetKey)
	if err != nil {
		t.Errorf("GetSet failed. %s", err)
	}
	if _, err := set.SAdd(setValue); err != nil {
		t.Errorf("add value to set failed. %s", err)
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
				key: notExistSetKey,
			},
			want: &Set{
				key: notExistSetKey,
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
				key: existSetKey,
			},
			want: &Set{
				key: existSetKey,
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
		meta SetMeta
		key  []byte
		txn  *Transaction
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
				meta: tt.fields.meta,
				key:  tt.fields.key,
				txn:  tt.fields.txn,
			}
			if err := set.updateMeta(); (err != nil) != tt.wantErr {
				t.Errorf("Set.updateMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSet_SAdd(t *testing.T) {
	type fields struct {
		meta SetMeta
		key  []byte
		txn  *Transaction
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
				meta: tt.fields.meta,
				key:  tt.fields.key,
				txn:  tt.fields.txn,
			}
			got, err := set.SAdd(tt.args.members)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set.SAdd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Set.SAdd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSet_SMembers(t *testing.T) {
	type fields struct {
		meta SetMeta
		key  []byte
		txn  *Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		want    [][]byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := &Set{
				meta: tt.fields.meta,
				key:  tt.fields.key,
				txn:  tt.fields.txn,
			}
			got, err := set.SMembers()
			if (err != nil) != tt.wantErr {
				t.Errorf("Set.SMembers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Set.SMembers() = %v, want %v", got, tt.want)
			}
		})
	}
}
