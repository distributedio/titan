package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitCursor(t *testing.T) {
	type args struct {
		begin int
		end   int
		llen  int
	}
	type want struct {
		begin int
		end   int
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "1",
			args: args{
				begin: -11,
				end:   -12,
				llen:  10,
			},
			want: want{
				begin: 0,
				end:   0,
			},
		},
		{
			name: "2",
			args: args{
				begin: -11,
				end:   12,
				llen:  10,
			},
			want: want{
				begin: 0,
				end:   9,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			begin, end := initCursor(tt.args.begin, tt.args.end, tt.args.llen)
			assert.Equal(t, tt.want.begin, begin)
			assert.Equal(t, tt.want.end, end)
		})
	}
}

func TestBitpos(t *testing.T) {
	type args struct {
		val []byte
		bit int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "1",
			args: args{
				val: []byte{byte(255), byte(255), byte(255), byte(255), byte(255), byte(255)},
				bit: 1,
			},
			want: 0,
		},
		{
			name: "2",
			args: args{
				val: []byte{byte(255), byte(255), byte(255), byte(255), byte(255), byte(255)},
				bit: 0,
			},
			want: 48,
		},
		{
			name: "3",
			args: args{
				val: []byte{byte(0), byte(0), byte(0), byte(0), byte(0)},
				bit: 1,
			},
			want: -1,
		},
		{
			name: "4",
			args: args{
				val: []byte{byte(0), byte(0), byte(0), byte(0), byte(0)},
				bit: 0,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, bitpos(tt.args.val, tt.args.bit))
		})
	}
}
