package db

import (
	"reflect"
	"testing"
)

/*
func TestGetString(t *testing.T) {
	txn, err := mockDB.Begin()
	if err != nil {
		t.Errorf("db.Begin error %s", err)
	}

	type args struct {
		txn *Transaction
		key []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *String
		wantErr bool
	}{
		{
			name: "not exist key",
			args: args{
				txn: txn,
				key: []byte("key1"),
			},
			want: &String{
				key: []byte("key1"),
				txn: txn,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetString(tt.args.txn, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}

*/
func TestString_Get(t *testing.T) {
	type fields struct {
		Meta StringMeta
		key  []byte
		txn  *Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &String{
				Meta: tt.fields.Meta,
				key:  tt.fields.key,
				txn:  tt.fields.txn,
			}
			got, err := s.Get()
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
	type fields struct {
		Meta StringMeta
		key  []byte
		txn  *Transaction
	}
	type args struct {
		val    []byte
		expire []int64
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
			s := &String{
				Meta: tt.fields.Meta,
				key:  tt.fields.key,
				txn:  tt.fields.txn,
			}
			if err := s.Set(tt.args.val, tt.args.expire...); (err != nil) != tt.wantErr {
				t.Errorf("String.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestString_Len(t *testing.T) {
	type fields struct {
		Meta StringMeta
		key  []byte
		txn  *Transaction
	}
	tests := []struct {
		name    string
		fields  fields
		want    int
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &String{
				Meta: tt.fields.Meta,
				key:  tt.fields.key,
				txn:  tt.fields.txn,
			}
			got, err := s.Len()
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
	type fields struct {
		Meta StringMeta
		key  []byte
		txn  *Transaction
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
			s := &String{
				Meta: tt.fields.Meta,
				key:  tt.fields.key,
				txn:  tt.fields.txn,
			}
			if got := s.Exist(); got != tt.want {
				t.Errorf("String.Exist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_Append(t *testing.T) {
	type fields struct {
		Meta StringMeta
		key  []byte
		txn  *Transaction
	}
	type args struct {
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
			s := &String{
				Meta: tt.fields.Meta,
				key:  tt.fields.key,
				txn:  tt.fields.txn,
			}
			got, err := s.Append(tt.args.value)
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
	type fields struct {
		Meta StringMeta
		key  []byte
		txn  *Transaction
	}
	type args struct {
		value []byte
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
			s := &String{
				Meta: tt.fields.Meta,
				key:  tt.fields.key,
				txn:  tt.fields.txn,
			}
			got, err := s.GetSet(tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("String.GetSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("String.GetSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_GetRange(t *testing.T) {
	type fields struct {
		Meta StringMeta
		key  []byte
		txn  *Transaction
	}
	type args struct {
		start int
		end   int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &String{
				Meta: tt.fields.Meta,
				key:  tt.fields.key,
				txn:  tt.fields.txn,
			}
			if got := s.GetRange(tt.args.start, tt.args.end); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("String.GetRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_SetRange(t *testing.T) {
	type fields struct {
		Meta StringMeta
		key  []byte
		txn  *Transaction
	}
	type args struct {
		offset int64
		value  []byte
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
			s := &String{
				Meta: tt.fields.Meta,
				key:  tt.fields.key,
				txn:  tt.fields.txn,
			}
			if _, err := s.SetRange(tt.args.offset, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("String.SetRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestString_Incr(t *testing.T) {
	type fields struct {
		Meta StringMeta
		key  []byte
		txn  *Transaction
	}
	type args struct {
		delta int64
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
			s := &String{
				Meta: tt.fields.Meta,
				key:  tt.fields.key,
				txn:  tt.fields.txn,
			}
			got, err := s.Incr(tt.args.delta)
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
	type fields struct {
		Meta StringMeta
		key  []byte
		txn  *Transaction
	}
	type args struct {
		delta float64
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
			s := &String{
				Meta: tt.fields.Meta,
				key:  tt.fields.key,
				txn:  tt.fields.txn,
			}
			got, err := s.Incrf(tt.args.delta)
			if (err != nil) != tt.wantErr {
				t.Errorf("String.Incrf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("String.Incrf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_encode(t *testing.T) {
	type fields struct {
		Meta StringMeta
		key  []byte
		txn  *Transaction
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &String{
				Meta: tt.fields.Meta,
				key:  tt.fields.key,
				txn:  tt.fields.txn,
			}
			if got := s.encode(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("String.encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestString_decode(t *testing.T) {
	type fields struct {
		Meta StringMeta
		key  []byte
		txn  *Transaction
	}
	type args struct {
		b []byte
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
			s := &String{
				Meta: tt.fields.Meta,
				key:  tt.fields.key,
				txn:  tt.fields.txn,
			}
			if err := s.decode(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("String.decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
