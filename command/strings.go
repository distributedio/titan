package command

import (
	"errors"
	"github.com/meitu/titan/encoding/resp"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/meitu/titan/db"
)

var (
	//MaxRangeInteger max index in setrange command
	MaxRangeInteger = 2<<29 - 1
)

// Get the value of key
func Get(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	val, err := str.Get()
	if err != nil {
		if err == db.ErrKeyNotFound {
			return NullBulkString(ctx.Out), nil
		}
		return nil, errors.New("ERR " + err.Error())
	}
	return BulkString(ctx.Out, string(val)), nil
}

// Set key to hold the string value
func Set(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	value := []byte(ctx.Args[1])
	args := ctx.Args

	var next bool
	var flag int      // flag int // 0 -- null 1---nx  2---xx
	var unit int64    // time ms
	var expire string //expire = expire *unit
	for i := 2; i < len(args); i++ {
		if i+1 < len(args) {
			next = true
		}
		if (args[i][0] == 'n' || args[i][0] == 'N') &&
			(args[i][1] == 'x' || args[i][1] == 'X') &&
			len(args[i]) == 2 {
			flag = 1
		} else if (args[i][0] == 'x' || args[i][0] == 'X') &&
			(args[i][1] == 'x' || args[i][1] == 'X') &&
			len(args[i]) == 2 {
			flag = 2
		} else if (args[i][0] == 'p' || args[i][0] == 'P') &&
			(args[i][1] == 'x' || args[i][1] == 'X') &&
			len(args[i]) == 2 && next {
			expire = args[i+1]
			i = i + 1
			unit = int64(time.Millisecond)
		} else if (args[i][0] == 'e' || args[i][0] == 'E') &&
			(args[i][1] == 'x' || args[i][1] == 'X') &&
			len(args[i]) == 2 && next {
			expire = args[i+1]
			i = i + 1
			unit = int64(time.Second)
		} else {
			return nil, ErrSyntax
		}
	}

	//ex|px
	if expire != "" {
		ui, err := strconv.ParseInt(expire, 10, 64)
		if err != nil {
			return nil, ErrInteger
		}
		if ui == 0 {
			return nil, ErrExpire
		}
		unit = ui * unit
	}

	obj, err := txn.Object(key)
	if err != nil && err != db.ErrKeyNotFound {
		return nil, errors.New("ERR " + err.Error())
	}

	//xx
	if flag == 2 {
		if err == db.ErrKeyNotFound {
			return NullBulkString(ctx.Out), nil
		}
	}
	//nx
	if flag == 1 {
		if err != db.ErrKeyNotFound {
			return NullBulkString(ctx.Out), nil
		}
	}

	if err != db.ErrKeyNotFound {
		txn.Destory(obj, key)
	}

	s := db.NewString(txn, key)
	if err := s.Set(value, unit); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return SimpleString(ctx.Out, OK), nil
}

// MGet returns the values of all specified key
func MGet(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	count := len(ctx.Args)
	values := make([][]byte, count)

	keys := make([][]byte, count)
	for i := range ctx.Args {
		keys[i] = []byte(ctx.Args[i])
	}

	strs, err := txn.Strings(keys)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	for i, str := range strs {
		if str == nil || !str.Exist() {
			values[i] = nil
			continue
		}
		values[i], _ = str.Get()
	}
	return BytesArray(ctx.Out, values), nil
}

// MSet sets the given keys to their respective values
func MSet(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	argc := len(ctx.Args)
	args := ctx.Args
	if argc%2 != 0 {
		return nil, ErrMSet
	}
	for i := 2; i <= argc; i += 2 {
		ctx.Args = args[i-2 : i]
		if _, err := Set(ctx, txn); err != nil {
			return nil, err
		}
	}
	return SimpleString(ctx.Out, OK), nil
}

// MSetNx et multiple keys to multiple values,only if none of the keys exist
//TODO bug
func MSetNx(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	argc := len(ctx.Args)
	args := ctx.Args
	if argc%2 != 0 {
		return nil, ErrMSet
	}
	for i := 2; i <= argc; i += 2 {
		if str, _ := txn.String([]byte(ctx.Args[0])); !str.Exist() {
			ctx.Args = append(args[i-2:i], "nx")
			if _, err := Set(ctx, txn); err != nil {
				return nil, errors.New("ERR " + err.Error())
			}
			ctx.Args = ctx.Args[2:]
			continue
		}
		return Integer(ctx.Out, int64(0)), nil
	}
	return Integer(ctx.Out, int64(1)), nil
}

// Strlen returns the length of the string value stored at key
func Strlen(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !str.Exist() {
		return Integer(ctx.Out, int64(0)), nil
	}

	v, _ := str.Len()

	return Integer(ctx.Out, int64(v)), nil
}

// Append a value to a key
func Append(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	value := []byte(ctx.Args[1])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	llen, err := str.Append(value)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(llen)), nil
}

// GetSet sets the string value of a key and return its old value
func GetSet(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	v := []byte(ctx.Args[1])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	if !str.Exist() {
		return NullBulkString(ctx.Out), nil
	}

	value, err := str.GetSet(v)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return BulkString(ctx.Out, string(value)), nil
}

// GetRange increments the integer value of a keys by the given amount
func GetRange(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	start, err := strconv.Atoi(ctx.Args[1])
	if err != nil {
		return nil, ErrInteger
	}
	end, err := strconv.Atoi(ctx.Args[2])
	if err != nil {
		return nil, ErrInteger
	}

	if !str.Exist() {
		return NullBulkString(ctx.Out), nil
	}

	value := str.GetRange(start, end)
	if len(value) == 0 {
		return NullBulkString(ctx.Out), nil
	}

	return BulkString(ctx.Out, string(value)), nil
}

// SetNx sets the value of a key ,only if the key does not exist
func SetNx(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	//get the key
	key := []byte(ctx.Args[0])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return Integer(ctx.Out, int64(0)), nil
		}
		return nil, errors.New("ERR " + err.Error())
	}
	//key 不存在时，为 key 设置指定的值。设置成功，返回 1 。 设置失败，返回 0 。
	if str.Exist() {
		return Integer(ctx.Out, int64(0)), nil
	}

	if err := str.Set([]byte(ctx.Args[1])); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(1)), nil
}

// SetEx sets the value and expiration of a key KEY_NAME TIMEOUT VALUE
func SetEx(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	//get the key
	key := []byte(ctx.Args[0])
	obj, err := txn.Object(key)
	if err != nil && err != db.ErrKeyNotFound {
		return nil, errors.New("ERR " + err.Error())
	}
	if err != db.ErrKeyNotFound {
		txn.Destory(obj, key)
	}

	s := db.NewString(txn, key)
	ui, err := strconv.ParseInt(ctx.Args[1], 10, 64)
	if err != nil {
		return nil, ErrInteger
	}
	unit := ui * int64(time.Second)
	if err := s.Set([]byte(ctx.Args[2]), unit); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	return SimpleString(ctx.Out, OK), nil
}

// PSetEx set the value and expiration in milliseconds of a key
func PSetEx(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	//get the key
	key := []byte(ctx.Args[0])
	obj, err := txn.Object(key)
	if err != nil && err != db.ErrKeyNotFound {
		return nil, errors.New("ERR " + err.Error())
	}

	if err != db.ErrKeyNotFound {
		txn.Destory(obj, key)
	}

	s := db.NewString(txn, key)
	ui, err := strconv.ParseUint(ctx.Args[1], 10, 64)
	if err != nil {
		return nil, ErrInteger
	}
	unit := ui * uint64(time.Millisecond)
	if err := s.Set([]byte(ctx.Args[2]), int64(unit)); err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	return SimpleString(ctx.Out, OK), nil
}

//SetRange Overwrites part of the string stored at key, starting at the specified offset, for the entire length of value.
func SetRange(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	offset, err := strconv.Atoi(ctx.Args[1])
	if err != nil {
		return nil, ErrInteger
	}

	key := []byte(ctx.Args[0])
	if offset < 0 || offset > MaxRangeInteger {
		return nil, ErrMaximum
	}

	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	// If the offset is larger than the current length of the string at key, the string is padded with zero-bytes to make offset fit.
	val, err := str.SetRange(int64(offset), []byte(ctx.Args[2]))
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(len(val))), nil

}

// Incr increments the integer value of a key  by one
func Incr(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	delta, err := str.Incr(1)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(delta)), nil
}

// IncrBy increments the integer value of a key by the given amount
func IncrBy(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	delta, err := strconv.ParseInt(ctx.Args[1], 10, 0)
	if err != nil {
		return nil, ErrInteger
	}

	delta, err = str.Incr(delta)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(delta)), nil
}

// IncrByFloat increments the float value of a key by the given amount
func IncrByFloat(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	delta, err := strconv.ParseFloat(ctx.Args[1], 64)
	if err != nil {
		return nil, ErrInteger
	}
	delta, err = str.Incrf(delta)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return BulkString(ctx.Out, strconv.FormatFloat(delta, 'f', 17, 64)), nil
}

// Decr decrements the integer value of a key by one
func Decr(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	delta, err := str.Incr(-1)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(delta)), nil
}

// DecrBy decrements the integer value of a key by the given number
func DecrBy(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	delta, err := strconv.ParseInt(ctx.Args[1], 10, 64)
	if err != nil {
		return nil, ErrInteger
	}

	delta, err = str.Incr(-delta)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(delta)), nil
}

// SetBit Sets or clears the bit at offset in the string value stored at key.
func SetBit(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	offset, err := strconv.Atoi(ctx.Args[1])
	if err != nil {
		return nil, ErrBitOffset
	}
	if offset < 0 {
		return nil, ErrBitOffset
	}

	on, err := strconv.Atoi(ctx.Args[2])
	if err != nil {
		return nil, ErrBitInteger
	}

	// Bits can only be set or cleared...
	if (on & ^1) != 0 {
		return nil, ErrBitInteger
	}

	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	val, err := str.SetBit(offset, on)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	if val != 0 {
		return Integer(ctx.Out, 1), nil
	}
	return Integer(ctx.Out, 0), nil
}

// GetBit get the bit at offset in the string value stored at key.
func GetBit(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	offset, err := strconv.Atoi(ctx.Args[1])
	if err != nil || offset < 0 {
		return nil, ErrBitOffset
	}

	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !str.Exist() {
		return Integer(ctx.Out, 0), nil
	}

	val, err := str.GetBit(offset)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	if val != 0 {
		return Integer(ctx.Out, 1), nil
	}
	return Integer(ctx.Out, 0), nil
}

// BitCount Count the number of set bits (population counting) in a string.
func BitCount(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !str.Exist() {
		return Integer(ctx.Out, 0), nil
	}

	var begin, end int
	switch len(ctx.Args) {
	case 3:
		begin, err = strconv.Atoi(ctx.Args[1])
		if err != nil {
			return nil, ErrInteger
		}
		end, err = strconv.Atoi(ctx.Args[2])
		if err != nil {
			return nil, ErrInteger
		}
	case 1:
		begin = 0
		end = len(str.Meta.Value) - 1
	default:
		return nil, ErrSyntax
	}

	val, err := str.BitCount(begin, end)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(val)), nil
}

// BitPos find first bit set or clear in a string
func BitPos(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	bit, err := strconv.Atoi(string(ctx.Args[1]))
	if err != nil {
		return nil, ErrInteger
	}

	if (bit != 0) && (bit != 1) {
		return nil, ErrBitInvaild
	}

	key := []byte(ctx.Args[0])
	str, err := txn.String(key)
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}

	if !str.Exist() {
		if bit == 1 {
			return Integer(ctx.Out, -1), nil
		}
		return Integer(ctx.Out, 0), nil
	}

	var begin, end int
	switch len(ctx.Args) {
	case 4:
		begin, err = strconv.Atoi(ctx.Args[2])
		if err != nil {
			return nil, ErrInteger
		}
		end, err = strconv.Atoi(ctx.Args[3])
		if err != nil {
			return nil, ErrInteger
		}
	case 3:
		begin, err = strconv.Atoi(ctx.Args[2])
		if err != nil {
			return nil, ErrInteger
		}
		end = len(str.Meta.Value) - 1
	case 2:
		begin = 0
		end = len(str.Meta.Value) - 1
	default:
		return nil, ErrSyntax
	}

	val, err := str.BitPos(bit, begin, end)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(val)), nil
}

// BitOp perform bitwise operations between strings
func BitOp(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	if len(ctx.Args) < 3 {
		return nil, ErrBitOp
	}

	// Get source keys value
	keys := make([][]byte, len(ctx.Args[2:]))
	values := make([][]byte, len(ctx.Args[2:]))
	for i, _ := range ctx.Args[2:] {
		keys[i] = []byte(ctx.Args[2+i])
	}

	strs, err := txn.Strings(keys)
	if err != nil {
		return nil, errors.New("ERR" + err.Error())
	}
	for i, str := range strs {
		if str == nil || !str.Exist() {
			values[i] = nil
			continue
		}
		values[i], _ = str.Get()
	}

	var result []byte
	switch ctx.Args[0] {
	case "and", "AND", "or", "OR", "xor", "XOR":
		result = doBitOp(ctx.Args[0], values)
	case "not", "NOT":
		if len(ctx.Args) != 3 {
			return nil, ErrBitOp
		}
		if values[0] != nil {
			for i, _ := range values[0] {
				values[0][i] = ^values[0][i]
			}
		}
		result = values[0]
	default:
		return nil, ErrBitOp
	}

	destination := ctx.Args[1]
	str, err := txn.String([]byte(destination))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}

		return nil, errors.New("ERR " + err.Error())
	}
	err = str.Set(result)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	return Integer(ctx.Out, int64(len(result))), nil
}

// BitField perform arbitrary bitfield integer operations on strings
func BitField(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := ctx.Args[0]
	str, err := txn.String([]byte(key))
	if err != nil {
		if err == db.ErrTypeMismatch {
			return nil, ErrTypeMismatch
		}
		return nil, errors.New("ERR " + err.Error())
	}
	value, err := str.Get()
	if err != nil && err != db.ErrKeyNotFound {
		return nil, errors.New("ERR " + err.Error())
	}
	if err == db.ErrKeyNotFound {
		value = []byte{0x00}
	}

	var (
		currentOverflowType = "wrap"
		bitfieldOperations  []func()
		result              []resp.Value
		command             = ctx.Args[1:]
		isChanged           = false
	)
	for i := 0; i < len(command); {
		c := strings.ToLower(command[i])
		// BITFIELD GET command needs three arguments
		if c == "get" {
			if i+1 >= len(command) || i+2 >= len(command) {
				return nil, ErrSyntax
			}
			bitfieldType := command[i+1]
			bitfieldOffset := command[i+2]

			bits, err := checkAndGetBitfieldType(bitfieldType)
			if err != nil {
				return nil, err
			}

			offset, err := checkAndGetBitfieldOffset(bitfieldOffset)
			if err != nil {
				return nil, err
			}

			if bitfieldOffset[0] != '#' {
				offset, _ = strconv.ParseUint(bitfieldOffset[:], 10, 64)
			} else {
				offset, _ = strconv.ParseUint(bitfieldOffset[1:], 10, 64)
				offset = offset * bits
			}

			isSigned := false
			if bitfieldType[0] == 'i' {
				isSigned = true
			}

			bitfieldOperations = append(bitfieldOperations, func() {
				result = append(result, resp.IntegerValue(bitfieldGet(value, bits, offset, isSigned)))
			})
			i = i + 3
		} else if c == "set" {
			// BITFIELD SET command needs four arguments
			if i+1 >= len(command) || i+2 >= len(command) || i+3 >= len(command) {
				return nil, ErrSyntax
			}
			bitfieldType := command[i+1]
			bitfieldOffset := command[i+2]
			bitfieldNewValue := command[i+3]

			bits, err := checkAndGetBitfieldType(bitfieldType)
			if err != nil {
				return nil, err
			}

			offset, err := checkAndGetBitfieldOffset(bitfieldOffset)
			if err != nil {
				return nil, err
			}

			newValue, err := checkAndGetBitfieldNewValue(bitfieldNewValue)
			if err != nil {
				return nil, err
			}

			if bitfieldOffset[0] != '#' {
				offset, _ = strconv.ParseUint(bitfieldOffset[:], 10, 64)
			} else {
				offset, _ = strconv.ParseUint(bitfieldOffset[1:], 10, 64)
				offset = offset * bits
			}

			isSigned := false
			if bitfieldType[0] == 'i' {
				isSigned = true
			}

			bitfieldOperations = append(bitfieldOperations, func() {
				r, data := bitfieldSet(value, newValue, bits, offset, isSigned)
				result = append(result, resp.IntegerValue(r))
				value = data
				isChanged = true
			})
			i = i + 4
		} else if c == "incrby" {
			if i+1 >= len(command) || i+2 >= len(command) || i+3 > len(command) {
				return nil, ErrSyntax
			}

			bitfieldType := command[i+1]
			bitfieldOffset := command[i+2]
			bitfieldIncrement := command[i+3]

			bits, err := checkAndGetBitfieldType(bitfieldType)
			if err != nil {
				return nil, err
			}

			offset, err := checkAndGetBitfieldOffset(bitfieldOffset)
			if err != nil {
				return nil, err
			}

			incr, err := checkAndGetBitfieldNewValue(bitfieldIncrement)
			if err != nil {
				return nil, err
			}

			if bitfieldOffset[0] != '#' {
				offset, _ = strconv.ParseUint(bitfieldOffset[:], 10, 64)
			} else {
				offset, _ = strconv.ParseUint(bitfieldOffset[1:], 10, 64)
				offset = offset * bits
			}

			isSigned := false
			if bitfieldType[0] == 'i' {
				isSigned = true
			}

			bitfieldOperations = append(bitfieldOperations, func() {
				r, data := bitfieldIncrby(value, incr, bits, offset, currentOverflowType, isSigned)
				if data == nil {
					result = append(result, resp.NullValue())
				} else {
					result = append(result, resp.IntegerValue(r))
				}
				// Reset to the default type
				currentOverflowType = "wrap"
				if len(data) > 0 {
					value = data
				}
				isChanged = true
			})
			i = i + 4
		} else if c == "overflow" {
			if i+1 >= len(command) {
				return nil, ErrSyntax
			}
			overflowType := command[i+1]
			if err := checkBitfieldOverflowType(overflowType); err != nil {
				return nil, err
			}
			bitfieldOperations = append(bitfieldOperations, func() {
				currentOverflowType = overflowType
			})
			i = i + 2
			continue
		} else {
			return nil, ErrSyntax
		}
	}

	// After syntax check and argument format, we excuate all the bitfield subcommands.
	for _, op := range bitfieldOperations {
		op()
	}

	// Set the final value if value is changed.
	if isChanged {
		err = str.Set(value)
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
	}

	return Array(ctx.Out, result), nil
}

// The type MUST format as: [u|i][1-64].
// The supported types are up to 64 bits for signed integers,
// and up to 63 bits for unsigned integers.
// Therefore, bitfield only support u63 and i64.
func checkAndGetBitfieldType(bitfieldType string) (uint64, error) {
	if len(bitfieldType) != 2 && len(bitfieldType) != 3 {
		return 0, ErrInvalidBitfieldType
	}

	if bitfieldType[0] != 'u' && bitfieldType[0] != 'i' {
		return 0, ErrInvalidBitfieldType
	}

	bits, err := strconv.ParseUint(bitfieldType[1:], 10, 64)
	if err != nil {
		return 0, ErrInvalidBitfieldType
	}

	switch bitfieldType[0] {
	case 'u':
		if bits > 63 {
			return 0, ErrInvalidBitfieldType
		}
	case 'i':
		if bits > 64 {
			return 0, ErrInvalidBitfieldType
		}
	}

	return bits, nil
}

// The offset format is like '100' or '#100'.
func checkAndGetBitfieldOffset(bitfieldOffset string) (uint64, error) {
	if bitfieldOffset[0] == '#' {
		bitfieldOffset = bitfieldOffset[1:]
	}

	offset, err := strconv.ParseUint(bitfieldOffset, 10, 64)
	if err != nil {
		return 0, ErrInvalidBitfieldOffset
	}

	return offset, nil
}

func checkBitfieldOverflowType(bitfieldOverflowType string) error {
	switch strings.ToLower(bitfieldOverflowType) {
	case "wrap", "sat", "fail":
		return nil
	default:
		return ErrInvalidBitfieldOverflowType
	}
}

func checkAndGetBitfieldNewValue(newValue string) (uint64, error) {
	if newValue[0] == '-' {
		n, err := strconv.ParseInt(newValue, 10, 64)
		if err != nil {
			return 0, ErrInvalidBitfieldValue
		}
		return uint64(n), nil
	}

	n, err := strconv.ParseUint(newValue, 10, 64)
	if err != nil {
		return 0, ErrInvalidBitfieldValue
	}
	return n, nil
}

// Translate from redis-5.0.3 src/bitops.c
func checkSignedBitfieldOverflow(value int64, incr int64, bits uint64, overflowType string) (int64, int) {
	var (
		max, min, maxincr, minincr int64
	)

	if bits == 64 {
		max = math.MaxInt64
	} else {
		max = int64(1<<(bits-1)) - 1
	}
	min = -max - 1
	maxincr = max - value
	minincr = min - value

	// Overflow process
	if value > max || (bits != 64 && incr > maxincr) || (value >= 0 && incr > 0 && incr > maxincr) {
		switch overflowType {
		case "wrap", "fail":
			msb := uint64(1 << (bits - 1))
			mask := uint64(0xFFFFFFFFFFFFFFFF << (bits - 1))
			c := uint64(value) + uint64(incr)
			if c&msb > 0 {
				c |= mask
			} else {
				c &= ^mask
			}
			return int64(c), 1
		case "sat":
			return max, 1
		}
	}

	// Underflow process
	if value < min || (bits != 64 && incr < minincr) || (value < 0 && incr < 0 && incr < minincr) {
		switch overflowType {
		case "wrap", "fail":
			msb := uint64(1 << (bits - 1))
			mask := uint64(0xFFFFFFFFFFFFFFFF << (bits - 1))
			c := uint64(value) + uint64(incr)
			if c&msb > 0 {
				c |= mask
			} else {
				c &= ^mask
			}
			return int64(c), 1
		case "sat":
			return min, -1
		}
	}

	return incr + value, 0
}

func checkUnsignedBitfieldOverflow(value uint64, incr int64, bits uint64, overflowType string) (uint64, int) {
	var max uint64
	if bits == 64 {
		max = math.MaxUint64
	} else {
		max = (uint64)(1<<bits) - 1
	}

	maxincr := max - value
	minincr := -int64(value)

	// Overflow process
	if value > max || (incr > 0 && incr > int64(maxincr)) {
		switch overflowType {
		case "wrap", "fail":
			return uint64(int64(value)+incr) & ^uint64(0xFFFFFFFFFFFFFFFF<<bits), 1
		case "sat":
			return max, 1
		}
	}

	// Underflow process
	if incr < 0 && incr < minincr {
		switch overflowType {
		case "wrap", "fail":
			return uint64(int64(value)+incr) & ^uint64(0xFFFFFFFFFFFFFFFF<<bits), 1
		case "sat":
			return 0, -1
		}
	}

	return uint64(incr) + value, 0
}

func getUnsignedBitfield(data []byte, offset, bits uint64) uint64 {
	var value uint64
	for i := 0; i < int(bits); i++ {
		byteIndex := offset >> 3
		bit := 7 - (offset & 0x7)
		if byteIndex > uint64(len(data)-1) {
			value <<= 1
		} else {
			byteval := data[byteIndex]
			bitval := (byteval >> bit) & 1
			value = (value << 1) | uint64(bitval)
		}
		offset++
	}
	return value
}

func getSignedBitfield(data []byte, offset, bits uint64) int64 {
	unsignedValue := getUnsignedBitfield(data, offset, bits)
	sign := unsignedValue & (1 << uint(bits-1))

	// Get the 2's complement of effective data
	if sign > 0 {
		return int64((^unsignedValue)&(0xFFFFFFFFFFFFFFFF>>(64-bits))+1) * -1
	}
	return int64(unsignedValue)
}

func setUnsignedBitfield(data []byte, offset, bits, value uint64) []byte {
	value &= 0xFFFFFFFFFFFFFFFF >> (64 - bits)

	// If the offset+bits beyond the original data length, expand the array of bytes with 0x00
	if offset+bits > uint64(len(data)*8) {
		var numOfExpandBytes int
		numOfExpandBits := offset + bits - uint64(len(data)*8)
		if numOfExpandBits%8 > 0 {
			numOfExpandBytes = int(numOfExpandBits/8 + 1)
		} else {
			numOfExpandBytes = int(numOfExpandBits / 8)
		}
		for i := 0; i < numOfExpandBytes; i++ {
			data = append(data, byte(0x00))
		}
	}

	for i := 0; i < int(bits); i++ {
		var bitval byte = 0
		if (value & (uint64)(1<<(bits-1-uint64(i)))) > 0 {
			bitval = 1
		}
		byteIndex := offset >> 3
		bit := 7 - (offset & 0x7)
		byteval := data[byteIndex]
		byteval &= ^(1 << bit)
		byteval |= bitval << bit
		data[byteIndex] = byteval
		offset++
	}
	return data
}

func bitfieldGet(data []byte, bits, offset uint64, isSigned bool) int64 {
	if offset > uint64(len(data)*8) {
		return 0
	}

	if isSigned {
		return getSignedBitfield(data, offset, bits)
	}

	// It's safe to store result in int64.
	return int64(getUnsignedBitfield(data, offset, bits))
}

func bitfieldSet(data []byte, newValue, bits, offset uint64, isSigned bool) (int64, []byte) {
	if offset > uint64(len(data)*8) {
		return 0, data
	}

	oldValue := bitfieldGet(data, bits, offset, isSigned)
	return oldValue, setUnsignedBitfield(data, offset, bits, newValue)
}

func bitfieldIncrby(data []byte, incr, bits, offset uint64, bitfieldOverflowType string, isSigned bool) (int64, []byte) {
	if offset > uint64(len(data)*8) {
		return 0, data
	}

	if isSigned {
		oldValue := getUnsignedBitfield(data, offset, bits)
		newValue, overflowFlag := checkSignedBitfieldOverflow(int64(oldValue), int64(incr), bits, bitfieldOverflowType)
		if overflowFlag != 0 && strings.ToLower(bitfieldOverflowType) == "fail" {
			return 0, nil
		}
		return int64(newValue), setUnsignedBitfield(data, offset, bits, uint64(newValue))
	}

	oldValue := getUnsignedBitfield(data, offset, bits)
	newValue, overflowFlag := checkUnsignedBitfieldOverflow(oldValue, int64(incr), bits, bitfieldOverflowType)
	if overflowFlag != 0 && strings.ToLower(bitfieldOverflowType) == "fail" {
		return 0, nil
	}
	return int64(newValue), setUnsignedBitfield(data, offset, bits, newValue)
}

func doBitOp(operation string, values [][]byte) []byte {
	var result = values[0]
	for i := 1; i < len(values); i++ {
		next := values[i]

		if (result == nil) && (next == nil) {
			result = nil
		}

		if len(result) < len(next) {
			for i := len(result); i < len(next); i++ {
				result = append(result, byte(0))
			}
		}

		if len(result) > len(next) {
			for i := len(next); i < len(result); i++ {
				next = append(next, byte(0))
			}
		}

		for i, _ := range result {
			result[i] = byteBoolOp(operation)(result[i], next[i])
		}
	}
	return result
}

func byteBoolOp(operation string) func(a, b byte) byte {
	switch operation {
	case "and", "AND":
		return func(a, b byte) byte { return a & b }
	case "or", "OR":
		return func(a, b byte) byte { return a | b }
	case "xor", "XOR":
		return func(a, b byte) byte { return a ^ b }
	default:
		return nil
	}
}
