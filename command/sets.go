package command

import (
	"bytes"
	"errors"
	"strconv"

	"github.com/meitu/titan/db"
)

// SAdd adds the specified members to the set stored at key
func SAdd(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	var members [][]byte
	for _, member := range ctx.Args[1:] {
		members = append(members, []byte(member))
	}
	set, err := txn.Set(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	added, err := set.SAdd(members)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, added), nil
}

// SMembers returns all the members of the set value stored at key
func SMembers(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	set, err := txn.Set(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}

	members, err := set.SMembers()
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return BytesArray(ctx.Out, members), nil
}
func SCard(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	set, err := txn.Set(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	count, err := set.SCard()
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(count)), nil
}
func SIsmember(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])
	member := []byte(ctx.Args[1])
	set, err := txn.Set(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	count, err := set.SIsmember(member)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(count)), nil

}
func SPop(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var count int = 0
	var err error
	key := []byte(ctx.Args[0])

	if []byte(ctx.Args[1]) != nil {
		count, err = strconv.Atoi(ctx.Args[1])
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
	}
	set, err := txn.Set(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	members, err := set.SPop(int64(count))
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return BytesArray(ctx.Out, members), nil
}
func SRem(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var members [][]byte
	key := []byte(ctx.Args[0])
	for _, member := range ctx.Args[1:] {
		members = append(members, []byte(member))
	}
	set, err := txn.Set(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	count, err := set.SRem(members)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(count)), nil
}

// SUion returns the members of the set resulting from the union of all the given sets.
func SUion(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var members [][]byte
	var keys [][]byte
	for _, key := range ctx.Args {
		keys = append(keys, []byte(key))
	}

	for i := range keys {
		set, err := txn.Set(keys[i])
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		if !set.Exists() {
			continue
		}
		if n, _ := set.SCard(); n == 0 {
			continue
		}
		ms, err := set.SMembers()
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		for i := range ms {
			members = append(members, ms[i])
		}
	}
	return BytesArray(ctx.Out, db.RemoveRepByMap(members)), nil
}

// SInter returns the members of the set resulting from the intersection of all the given sets.
func SInter(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var keys [][]byte
	var members [][]byte

	//取第一个
	key := []byte(ctx.Args[0])
	set, err := txn.Set(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	if !set.Exists() {
		return nil, nil
	}
	if n, _ := set.SCard(); n == 0 {
		return nil, nil
	}

	members, err = set.SMembers()
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	//遍历之后的
	for _, key := range ctx.Args[1:] {
		keys = append(keys, []byte(key))
	}
	for i := range keys {
		set, err := txn.Set(keys[i])
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		if !set.Exists() {
			return nil, nil
		}
		if n, _ := set.SCard(); n == 0 {
			return nil, nil
		}
		ms, err := set.SMembers()
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		members = sliceIntersect(members, ms)

	}
	return BytesArray(ctx.Out, db.RemoveRepByMap(members)), nil
}

// SDiff returns the members of the set resulting from the difference between the first set and all the successive sets.
//func SDiff(ctx *Context, txn *db.Transaction) (OnCommit, error) {
//	return BytesArray(ctx.Out, db.RemoveRepByMap(members)), nil
//}

// InSliceIface checks given interface in interface slice.
func inSliceIface(v []byte, sl [][]byte) bool {
	for _, vv := range sl {
		if bytes.Equal(vv, v) {
			return true
		}
	}
	return false
}

// SliceIntersect returns slice that are present in all the slice1 and slice2.
func sliceIntersect(slice1, slice2 [][]byte) (diffslice [][]byte) {
	for _, v := range slice1 {
		if inSliceIface(v, slice2) {
			diffslice = append(diffslice, v)
		}
	}
	return
}
