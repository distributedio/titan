package db

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"
)

// compareGetString skip CreatedAt UpdatedAt ID compare
func compareGetString(want, get *String) error {
	switch {
	case !bytes.Equal(want.key, get.key):
		return fmt.Errorf("set key not equal, want=%s, get=%s", string(want.key), string(get.key))
	case want.Meta.ExpireAt != get.Meta.ExpireAt:
		return fmt.Errorf("meta.ExpireAt not equal, want=%v, get=%v", want.Meta.ExpireAt, get.Meta.ExpireAt)
	case want.Meta.Type != get.Meta.Type:
		return fmt.Errorf("meta.Type not equal, want=%v, get=%v", want.Meta.Type, get.Meta.Type)
	case want.Meta.Encoding != get.Meta.Encoding:
		return fmt.Errorf("meta.Encoding not equal, want=%v, get=%v", want.Meta.Encoding, get.Meta.Encoding)
	case !bytes.Equal(want.Meta.Value, get.Meta.Value):
		return fmt.Errorf("meta.Value not equal, want=%v, get=%v", want.Meta.Value, get.Meta.Value)
	}
	return nil
}
func setValue(t *testing.T, TestKey []byte, value []byte) {
	txn, err := mockDB.Begin()
	if err != nil {
		t.Errorf("db.Begin error %s", err)
	}
	s, err := GetString(txn, TestKey)
	if err != nil {
		t.Errorf("String.GetString() error = %v", err)
	}
	err = s.Set(value)
	if err != nil {
		t.Errorf("String_Set() error = %v", err)
	}
	if err = txn.Commit(context.TODO()); err != nil {
		t.Errorf("Set() txn.Commit error = %v", err)
		return
	}
}

var (
	TestExistKey       = []byte("StringKey")
	TestExpireExistKey = []byte("ExpireStringKey")
	TestNoExistKey     = []byte("NoExitStringKey")

	value = []byte("StringValue")

	NoExist = "No Exist Key"
	Exist   = "Exist Key"
)

func TestGetString(t *testing.T) {
	txn, err := mockDB.Begin()
	if err != nil {
		t.Errorf("db.Begin error %s", err)
	}
	NewString(txn, TestExistKey)
	type args struct {
		txn *Transaction
		key []byte
	}
	type want struct {
		key *String
		val []byte
		err error
	}
	tests := []struct {
		name    string
		args    args
		want    want
		err     error
		wantErr bool
	}{
		{
			name: "GetExistString",
			args: args{
				txn: txn,
				key: TestExistKey,
			},
			want: want{
				key: &String{
					key: TestExistKey,
					txn: txn,
				},
				val: value,
				err: nil,
			},
			err:     nil,
			wantErr: false,
		},
		{
			name: "GetNoExistString",
			args: args{
				txn: txn,
				key: TestNoExistKey,
			},
			want: want{
				key: &String{
					key: TestNoExistKey,
					txn: txn,
				},
				val: nil,
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
			got, err := GetString(tt.args.txn, tt.args.key)
			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("GetString() txn.Commit error = %v", err)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err := compareGetString(tt.want.key, got); err != nil {
				t.Errorf("GetString() = %v, want %v", got, tt.want)
			}
			if tt.want.err != tt.err {
				t.Errorf("GetString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func TestString_Get(t *testing.T) {
	setValue(t, TestExistKey, value)
	tests := []struct {
		name    string
		key     []byte
		value   []byte
		want    []byte
		wantErr bool
	}{
		{
			name:    "TestString_GetExitKey",
			key:     TestExistKey,
			value:   value,
			want:    value,
			wantErr: false,
		},
		{
			name:    "TestString_GetNoExitKey",
			key:     TestNoExistKey,
			value:   nil,
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			if err != nil {
				t.Errorf("db.Begin error %s", err)
			}
			s, err := GetString(txn, tt.key)
			if err != nil {
				t.Errorf("String.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := s.Get()
			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("Set() txn.Commit error = %v", err)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("String.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("String.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_Set(t *testing.T) {
	var exp = []int64{int64(2 * time.Second)}
	type args struct {
		val    []byte
		expire []int64
	}
	type want struct {
		val []byte
		err error
	}
	tests := []struct {
		name    string
		key     []byte
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "set no expire",
			key:  TestExistKey,
			args: args{
				val:    value,
				expire: nil,
			},
			want: want{
				val: value,
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "set expire",
			key:  TestExpireExistKey,
			args: args{
				val:    value,
				expire: exp,
			},
			want: want{
				val: nil,
				err: ErrKeyNotFound,
			},
			wantErr: false,
		},
	}
	var txn *Transaction
	var err error
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if txn, err = mockDB.Begin(); err != nil {
				t.Errorf("TestGetSet db.Begin error %s", err)
			}
			s, err := GetString(txn, tt.key)
			if err != nil {
				t.Errorf("String.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			err = s.Set(tt.args.val, tt.args.expire...)
			if err != nil {
				t.Errorf("String_Set() error = %v", err)
			}
			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("Set() txn.Commit error = %v", err)
				return
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("String.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(tt.args.expire) != 0 {
				if txn, err = mockDB.Begin(); err != nil {
					t.Errorf("TestGetSet db.Begin error %s", err)
				}
				s, err = GetString(txn, tt.key)
				if err != nil {
					t.Errorf("String.Set() error = %v, wantErr %v", err, tt.wantErr)
				}
				if val, err := s.Get(); !bytes.Equal(val, tt.args.val) || err != nil {
					t.Errorf("String.Get() key=%s error = %v,value = %v", string(tt.key), err, string(val))
				}
				if err = txn.Commit(context.TODO()); err != nil {
					t.Errorf("Set() txn.Commit error = %v", err)
					return
				}

				time.Sleep(3 * time.Second)
			}
			if txn, err = mockDB.Begin(); err != nil {
				t.Errorf("TestGetSet db.Begin error %s", err)
			}
			s, err = GetString(txn, tt.key)
			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("Set() txn.Commit error = %v", err)
				return
			}
			if val, err := s.Get(); !bytes.Equal(val, tt.want.val) || err != tt.want.err {
				t.Errorf("String.Get() key=%s error = %v,value = %v", string(tt.key), err, string(val))
			}
		})
	}
}

func TestString_Len(t *testing.T) {
	setValue(t, TestExistKey, value)
	tests := []struct {
		name    string
		key     []byte
		want    int
		wantErr bool
	}{
		{
			name:    "TestString_Len",
			key:     TestExistKey,
			want:    11,
			wantErr: false,
		},
		{
			name:    "TestString_Len",
			key:     TestNoExistKey,
			want:    0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			if err != nil {
				t.Errorf("TestGetSet db.Begin error %s", err)
			}
			s, err := GetString(txn, tt.key)
			if err != nil {
				t.Errorf("String.GetString error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := s.Len()
			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("Set() txn.Commit error = %v", err)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("String.Len() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("String.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_Exist(t *testing.T) {
	setValue(t, TestExistKey, value)
	// write exist string
	tests := []struct {
		name string
		key  []byte
		want bool
	}{
		{
			name: "Exist",
			key:  TestExistKey,
			want: true,
		},
		{
			name: "NoExist",
			key:  TestNoExistKey,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			if err != nil {
				t.Errorf("db.Begin error %s", err)
			}
			s, err := GetString(txn, tt.key)
			if err != nil {
				t.Errorf("String.GetString error = %v", err)
			}
			got := s.Exist()
			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("Set() txn.Commit error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("String.Exist() = %v, want %v", got, tt.want)
			}
		})
	}

}

func TestString_Append(t *testing.T) {
	setValue(t, TestExistKey, value)
	type args struct {
		value []byte
	}
	tests := []struct {
		name    string
		args    args
		key     []byte
		want    int
		wantErr bool
	}{
		{
			name: "TestString_AppendExistKey",
			args: args{
				value: value,
			},
			key:     TestExistKey,
			want:    22,
			wantErr: false,
		},
		{
			name: "TestString_AppendNoExistKey",
			args: args{
				value: value,
			},
			key:     TestNoExistKey,
			want:    11,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			if err != nil {
				t.Errorf("db.Begin error %s", err)
			}
			s, err := GetString(txn, tt.key)
			if err != nil {
				t.Errorf("String.GetString error = %v", err)
			}
			got, err := s.Append(tt.args.value)
			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("Set() txn.Commit error = %v", err)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("String.Append() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("String.Append() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_GetSet(t *testing.T) {
	setValue(t, []byte("GetSetExistKey"), value)
	type args struct {
		value []byte
	}
	tests := []struct {
		name    string
		args    args
		key     []byte
		want    []byte
		wantErr bool
	}{
		{
			name: "GetSet_ExistKey",
			args: args{
				value: []byte("NewVlaue"),
			},
			key:     []byte("GetSetExistKey"),
			want:    value,
			wantErr: false,
		},
		{
			name: "GetSet_NoExistKey",
			args: args{
				value: []byte("NewVlaue"),
			},
			key:     []byte("GetSetNoExistKey"),
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			if err != nil {
				t.Errorf("db.Begin error %s", err)
			}
			s, err := GetString(txn, tt.key)
			if err != nil {
				t.Errorf("String.GetString error = %v", err)
			}
			got, err := s.GetSet(tt.args.value)
			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("Set() txn.Commit error = %v", err)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("String.GetSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != string(tt.want) {
				t.Errorf("String.GetSet() = %v, want %v", got, tt.want)
			}
			value, err := s.Get()
			if err != nil || string(value) != "NewVlaue" {
				t.Error("get failed", err)
			}
		})
	}
}

func TestString_GetRange(t *testing.T) {
	type args struct {
		start int
		end   int
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "GetRange1",
			//"StringValue"
			args: args{
				start: 4,
				end:   -1,
			},
			want: []byte("ngValue"),
		},

		{
			name: "GetRange2",
			args: args{
				start: 4,
				end:   -20,
			},
			want: nil,
		},

		{
			name: "GetRange3",
			args: args{
				start: -1,
				end:   3,
			},
			want: nil,
		},
		{
			name: "GetRange4",
			args: args{
				start: 22,
				end:   3,
			},
			want: nil,
		},

		{
			name: "GetRange5",
			args: args{
				start: 3,
				end:   20,
			},
			want: []byte("ingValue"),
		},
		{
			name: "GetRange6",
			args: args{
				start: -22,
				end:   3,
			},
			want: []byte("Stri"),
		},
		{
			name: "GetRange7",
			args: args{
				start: 2,
				end:   10,
			},
			want: []byte("ringValue"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			if err != nil {
				t.Errorf("db.Begin error %s", err)
			}
			s, err := GetString(txn, []byte("GetRangeExistKey"))
			if err != nil {
				t.Errorf("String.GetString error = %v", err)
			}
			err = s.Set(value)
			if err != nil {
				t.Errorf("String.set() error = %v", err)
			}
			got := s.GetRange(tt.args.start, tt.args.end)
			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("Set() txn.Commit error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("String.GetRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_SetRange(t *testing.T) {
	setValue(t, []byte("SetRangeExistKey"), value)
	type args struct {
		offset int64
		value  []byte
	}

	tests := []struct {
		name    string
		args    args
		key     []byte
		want    []byte
		wantErr bool
	}{
		{
			name: "SetRange_ExistKey",
			args: args{
				offset: 6,
				value:  []byte("lllll"),
			},
			key:     []byte("SetRangeExistKey"),
			want:    []byte("Stringlllll"),
			wantErr: false,
		},
		{
			name: "SetRange_NoExistKey",
			args: args{
				offset: 6,
				value:  []byte("lllll"),
			},
			key:     []byte("SetRangeNoExistKey"),
			want:    []byte("\x00\x00\x00\x00\x00\x00lllll"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			if err != nil {
				t.Errorf("db.Begin error %s", err)
			}
			s, err := GetString(txn, tt.key)
			if err != nil {
				t.Errorf("String.GetString error = %v", err)
			}

			got, err := s.SetRange(tt.args.offset, tt.args.value)
			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("Set() txn.Commit error = %v", err)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("String.SetRange() error = %v, wantErr %v", err, tt.wantErr)
			}
			if string(got) != string(tt.want) {
				t.Errorf("String.SetRange() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestString_Incr(t *testing.T) {
	type args struct {
		delta int64
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "Incr",
			args: args{
				delta: 10,
			},
			want:    int64(20),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			if err != nil {
				t.Errorf("db.Begin error %s", err)
			}
			s, err := GetString(txn, value)
			if err != nil {
				t.Errorf("String.GetString error = %v", err)
			}
			err = s.Set([]byte("10"))
			if err != nil {
				t.Errorf("String.Set error = %v", err)
			}
			got, err := s.Incr(tt.args.delta)
			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("Set() txn.Commit error = %v", err)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("String.Incr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("String.Incr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_Incrf(t *testing.T) {
	type args struct {
		delta float64
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "Incr",
			args: args{
				delta: float64(0.2),
			},
			want:    float64(10.3),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txn, err := mockDB.Begin()
			if err != nil {
				t.Errorf("db.Begin error %s", err)
			}
			s, err := GetString(txn, value)
			if err != nil {
				t.Errorf("String.GetString error = %v", err)
			}
			err = s.Set([]byte("10.1"))
			if err != nil {
				t.Errorf("String.Set error = %v", err)
			}
			got, err := s.Incrf(tt.args.delta)
			if err = txn.Commit(context.TODO()); err != nil {
				t.Errorf("Set() txn.Commit error = %v", err)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("String.Incrf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got < tt.want-0.000000000000001 {
				t.Errorf("String.Incrf() = %v, want %v", got, tt.want)
			}
		})
	}
}
