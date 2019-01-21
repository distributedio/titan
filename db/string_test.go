package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// compareGetString skip CreatedAt UpdatedAt ID compare
func compareGetString(t *testing.T, want, get *String) {
	assert.Equal(t, want.key, get.key)
	assert.Equal(t, want.Meta.ExpireAt, get.Meta.ExpireAt)
	assert.Equal(t, want.Meta.Type, get.Meta.Type)
	assert.Equal(t, want.Meta.Encoding, get.Meta.Encoding)
	assert.Equal(t, want.Meta.Value, get.Meta.Value)
}

func setValue(t *testing.T, TestKey []byte, value []byte) {
	callFunc := func(txn *Transaction) {
		s, err := GetString(txn, TestKey)
		assert.NoError(t, err)
		err = s.Set(value)
	}
	MockTest(t, callFunc)
}

func getValue(t *testing.T, key []byte, value []byte) {
	callFunc := func(txn *Transaction) {
		s, err := GetString(txn, key)
		assert.NoError(t, err)
		val, err := s.Get()
		assert.NoError(t, err)
		assert.Equal(t, value, val)
	}
	MockTest(t, callFunc)
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
	assert.NoError(t, err)
	NewString(txn, TestExistKey)
	type args struct {
		txn *Transaction
		key []byte
	}
	type want struct {
		key *String
		val []byte
	}
	tests := []struct {
		name string
		args args
		want want
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
			},
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callFunc := func(txn *Transaction) {
				got, err := GetString(tt.args.txn, tt.args.key)
				assert.NoError(t, err)
				compareGetString(t, tt.want.key, got)
			}
			MockTest(t, callFunc)
		})
	}
}

func TestStringGet(t *testing.T) {
	setValue(t, TestExistKey, value)
	tests := []struct {
		name  string
		key   []byte
		value []byte
		want  []byte
	}{
		{
			name:  "TestString_GetExitKey",
			key:   TestExistKey,
			value: value,
			want:  value,
		},
		{
			name:  "TestString_GetNoExitKey",
			key:   TestNoExistKey,
			value: nil,
			want:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callFunc := func(txn *Transaction) {
				s, err := GetString(txn, tt.key)
				assert.NoError(t, err)
				got, err := s.Get()
				assert.Equal(t, got, tt.want)
			}
			MockTest(t, callFunc)
		})
	}
}

func TestStringSet(t *testing.T) {
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
		name string
		key  []byte
		args args
		want want
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callFunc := func(txn *Transaction) {
				s, err := GetString(txn, tt.key)
				assert.NoError(t, err)
				err = s.Set(tt.args.val, tt.args.expire...)
				assert.NoError(t, err)
			}
			MockTest(t, callFunc)

			if len(tt.args.expire) != 0 {
				callFunc = func(txn *Transaction) {
					s, err := GetString(txn, tt.key)
					assert.NoError(t, err)
					val, err := s.Get()
					assert.NoError(t, err)
					assert.Equal(t, tt.args.val, val)
				}
				MockTest(t, callFunc)
				time.Sleep(3 * time.Second)
			}
			callFunc = func(txn *Transaction) {
				s, err := GetString(txn, tt.key)
				assert.NoError(t, err)
				val, err := s.Get()
				assert.Equal(t, tt.want.err, err)
				assert.Equal(t, tt.want.val, val)
			}
			MockTest(t, callFunc)
		})
	}
}

func TestStringLen(t *testing.T) {
	setValue(t, TestExistKey, value)
	tests := []struct {
		name string
		key  []byte
		want int
	}{
		{
			name: "TestString_Len",
			key:  TestExistKey,
			want: 11,
		},
		{
			name: "TestString_Len",
			key:  TestNoExistKey,
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callFunc := func(txn *Transaction) {
				s, err := GetString(txn, tt.key)
				assert.NoError(t, err)
				got, err := s.Len()
				assert.Equal(t, tt.want, got)
			}
			MockTest(t, callFunc)
		})
	}
}

func TestStringExist(t *testing.T) {
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
			callFunc := func(txn *Transaction) {
				s, err := GetString(txn, tt.key)
				assert.NoError(t, err)
				got := s.Exist()
				assert.Equal(t, tt.want, got)
			}
			MockTest(t, callFunc)
		})
	}

}

func TestStringAppend(t *testing.T) {
	setValue(t, TestExistKey, value)
	type args struct {
		value []byte
	}
	tests := []struct {
		name string
		args args
		key  []byte
		want int
	}{
		{
			name: "TestString_AppendExistKey",
			args: args{
				value: value,
			},
			key:  TestExistKey,
			want: 22,
		},
		{
			name: "TestString_AppendNoExistKey",
			args: args{
				value: value,
			},
			key:  TestNoExistKey,
			want: 11,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callFunc := func(txn *Transaction) {
				s, err := GetString(txn, tt.key)
				assert.NoError(t, err)
				got, err := s.Append(tt.args.value)
				assert.Equal(t, tt.want, got)
			}
			MockTest(t, callFunc)
		})
	}
}

func TestStringGetSet(t *testing.T) {
	setValue(t, []byte("GetSetExistKey"), value)
	type args struct {
		value []byte
	}
	tests := []struct {
		name string
		args args
		key  []byte
		want []byte
	}{
		{
			name: "GetSet_ExistKey",
			args: args{
				value: []byte("NewVlaue"),
			},
			key:  []byte("GetSetExistKey"),
			want: value,
		},
		{
			name: "GetSet_NoExistKey",
			args: args{
				value: []byte("NewVlaue"),
			},
			key:  []byte("GetSetNoExistKey"),
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callFunc := func(txn *Transaction) {
				s, err := GetString(txn, tt.key)
				assert.NoError(t, err)
				got, err := s.GetSet(tt.args.value)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
				value, err := s.Get()
				assert.NoError(t, err)
				assert.Equal(t, value, []byte("NewVlaue"))
			}
			MockTest(t, callFunc)
		})
	}
}

func TestStringGetRange(t *testing.T) {
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
			callFunc := func(txn *Transaction) {
				s, err := GetString(txn, []byte("GetRangeExistKey"))
				assert.NoError(t, err)
				err = s.Set(value)
				assert.NoError(t, err)
				got := s.GetRange(tt.args.start, tt.args.end)
				assert.Equal(t, tt.want, got)
			}
			MockTest(t, callFunc)
		})
	}
}

func TestStringSetRange(t *testing.T) {
	setValue(t, []byte("SetRangeExistKey"), value)
	type args struct {
		offset int64
		value  []byte
	}

	tests := []struct {
		name string
		args args
		key  []byte
		want []byte
	}{
		{
			name: "SetRange_ExistKey",
			args: args{
				offset: 6,
				value:  []byte("lllll"),
			},
			key:  []byte("SetRangeExistKey"),
			want: []byte("Stringlllll"),
		},
		{
			name: "SetRange_NoExistKey",
			args: args{
				offset: 6,
				value:  []byte("lllll"),
			},
			key:  []byte("SetRangeNoExistKey"),
			want: []byte("\x00\x00\x00\x00\x00\x00lllll"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callFunc := func(txn *Transaction) {
				s, err := GetString(txn, tt.key)
				assert.NoError(t, err)
				got, err := s.SetRange(tt.args.offset, tt.args.value)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			MockTest(t, callFunc)
		})
	}
}

func TestStringIncr(t *testing.T) {
	type args struct {
		delta int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "Incr",
			args: args{
				delta: 10,
			},
			want: int64(20),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callFunc := func(txn *Transaction) {
				s, err := GetString(txn, value)
				assert.NoError(t, err)
				err = s.Set([]byte("10"))
				assert.NoError(t, err)
				got, err := s.Incr(tt.args.delta)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			MockTest(t, callFunc)
		})
	}
}

func TestStringIncrf(t *testing.T) {
	type args struct {
		delta float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "Incr",
			args: args{
				delta: float64(0.2),
			},
			want: float64(10.3),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callFunc := func(txn *Transaction) {
				s, err := GetString(txn, value)
				assert.NoError(t, err)
				err = s.Set([]byte("10.1"))
				assert.NoError(t, err)
				got, err := s.Incrf(tt.args.delta)
				assert.NoError(t, err)
				if got < tt.want-0.000000000000001 {
					t.Errorf("String.Incrf() = %v, want %v", got, tt.want)
				}
			}
			MockTest(t, callFunc)
		})
	}
}

func TestStringSetBit(t *testing.T) {
	type args struct {
		on  int
		off int
	}
	type want struct {
		retval int
		value  []byte
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "one",
			args: args{
				off: 1,
				on:  1,
			},
			want: want{
				retval: 0,
				value:  []byte{0x40},
			},
		},
		{
			name: "two",
			args: args{
				off: 5,
				on:  1,
			},
			want: want{
				retval: 0,
				value:  []byte{0x44},
			},
		},
		{
			name: "three",
			args: args{
				off: 5,
				on:  0,
			},
			want: want{
				retval: 0x4,
				value:  []byte{0x40},
			},
		},
		{
			name: "four",
			args: args{
				off: 12,
				on:  1,
			},
			want: want{
				retval: 0,
				value:  []byte{0x40, 0x08},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := []byte("set-bit")
			callFunc := func(txn *Transaction) {
				s, err := GetString(txn, key)
				assert.NoError(t, err)
				want, err := s.SetBit(tt.args.off, tt.args.on)
				assert.NoError(t, err)
				assert.Equal(t, tt.want.retval, want)
			}
			MockTest(t, callFunc)
			getValue(t, key, tt.want.value)
		})
	}
}

func TestStringGetBit(t *testing.T) {
	key := []byte("get-bits")
	callFunc := func(txn *Transaction) {
		s, err := GetString(txn, key)
		assert.NoError(t, err)
		s.SetBit(4, 1)
	}
	MockTest(t, callFunc)

	type args struct {
		off int
	}
	type want struct {
		retval int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "one",
			args: args{
				off: 1,
			},
			want: want{
				retval: 0,
			},
		},
		{
			name: "two",
			args: args{
				off: 100,
			},
			want: want{
				retval: 0,
			},
		},
		{
			name: "three",
			args: args{
				off: 4,
			},
			want: want{
				retval: 8,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callFunc := func(txn *Transaction) {
				s, err := GetString(txn, key)
				assert.NoError(t, err)
				want, err := s.GetBit(tt.args.off)
				assert.NoError(t, err)
				assert.Equal(t, tt.want.retval, want)
			}
			MockTest(t, callFunc)
		})
	}
}
