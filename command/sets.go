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

	members := make([][]byte, len(ctx.Args[1:]))
	for i, member := range ctx.Args[1:] {
		members[i] = []byte(member)
	}
	set, err := txn.Set(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	added, err := set.SAdd(members...)
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

// SCard returns the set cardinality (number of elements) of the set stored at key
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

// SIsmember returns if member is a member of the set stored at key
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

// SPop removes and returns one or more random elements from the set value store at key
func SPop(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var count int
	var err error
	var members [][]byte
	var set *db.Set
	key := []byte(ctx.Args[0])

	if len(ctx.Args) == 2 {
		count, err = strconv.Atoi(ctx.Args[1])
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
	}
	set, err = txn.Set(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	members, err = set.SPop(int64(count))
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return BytesArray(ctx.Out, members), nil
}

// SRem removes the specified members from the set stored at key
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

// SMove movies member from the set at source to the set at destination
func SMove(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	member := make([]byte, 0, len(ctx.Args[2]))
	key := []byte(ctx.Args[0])
	destkey := []byte(ctx.Args[1])
	member = []byte(ctx.Args[2])

	set, err := txn.Set(key)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	count, err := set.SMove(destkey, member)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	return Integer(ctx.Out, int64(count)), nil
}

// GetPrefix gets prefix of the key
func GetPrefix(txn *db.Transaction, key []byte) ([]byte, error) {
	set, err := txn.Set([]byte(key))
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	dkey := set.GetDataKey(txn, []byte(key))
	prefix := append(dkey, ':')
	return prefix, nil
}

// SUnion returns the members of the set resulting from the union of all the given sets.
func SUnion(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var members [][]byte
	var setsIter = make([]db.Iterator, len(ctx.Args)) //存储每个set当前的迭代器位置
	var min []byte
	var count int
	var keys = make([][]byte, len(ctx.Args))

	for i, key := range ctx.Args {
		set, err := txn.Set([]byte(key))
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		dkey := set.GetDataKey(txn, []byte(key))
		prefix := append(dkey, ':')
		iter, err := set.GetIter(prefix)
		if err != nil {
			return nil, err
		}
		defer iter.Close()
		setsIter[i] = iter
		keys[i] = []byte(key)
	}
	prefix, _ := GetPrefix(txn, keys[0])
	min = setsIter[0].Key()[len(prefix):]
	for count < len(ctx.Args) {
		for i := 0; i < len(ctx.Args); i++ {
			prefix, _ := GetPrefix(txn, []byte(keys[i]))
			if !setsIter[i].Key().HasPrefix(prefix) {
				continue
			}
			iter := setsIter[i]
			if bytes.Compare(min, iter.Key()[len(prefix):]) == 1 || bytes.Equal(iter.Key()[len(prefix):], min) {
				min = iter.Key()[len(prefix):]
			}
		}
		for i := 0; i < len(ctx.Args); i++ {
			prefix, _ := GetPrefix(txn, []byte(keys[i]))
			if !setsIter[i].Key().HasPrefix(prefix) {
				continue
			}
			if bytes.Equal(setsIter[i].Key()[len(prefix):], min) {
				if err := setsIter[i].Next(); err != nil {
					return nil, err
				}
			}
			if !setsIter[i].Key().HasPrefix(prefix) {
				count++
			}

		}
		members = append(members, min)
		for i := 0; i < len(ctx.Args); i++ {
			prefix, _ := GetPrefix(txn, []byte(keys[i]))
			if setsIter[i].Key().HasPrefix(prefix) {
				min = setsIter[i].Key()[len(prefix):]
				break
			}
		}

	}
	return BytesArray(ctx.Out, members), nil
}

// SInter returns the members of the set resulting from the intersection of all the given sets.
func SInter(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var members [][]byte
	var setsIter = make([]db.Iterator, len(ctx.Args)) //存储每个set当前的迭代器位置
	var max []byte
	var keys = make([][]byte, len(ctx.Args))
	var mkeys = make([][]byte, len(ctx.Args))
	for i, key := range ctx.Args {
		set, err := txn.Set([]byte(key))
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		dkey := set.GetDataKey(txn, []byte(key))
		mkey := db.GetMetaKey(txn, []byte(key))
		prefix := append(dkey, ':')
		iter, err := set.GetIter(prefix)
		if err != nil {
			return nil, err
		}
		defer iter.Close()
		setsIter[i] = iter
		keys[i] = []byte(key)
		mkeys[i] = mkey
	}

	// Batch get meta information
	// If the set corresponding to key does not exist, it is processed as an empty set
	mval, err := db.BatchGetValues(txn, mkeys)
	if err != nil {
		return nil, errors.New("ERR " + err.Error())
	}
	for _, val := range mval {
		if val == nil {
			return nil, nil
		}
		smeta, _ := db.DecodeSetMeta(val)
		if smeta.Len == 0 {
			return nil, nil
		}
	}

	prefix, _ := GetPrefix(txn, keys[0])
	max = setsIter[0].Key()[len(prefix):]
	for {
		i := 0
	Loop:
		for ; i < len(ctx.Args); i++ {
			iter := setsIter[i]
			prefix, _ := GetPrefix(txn, []byte(keys[i]))
			for ; iter.Key().HasPrefix(prefix); iter.Next() {
				if bytes.Compare(iter.Key()[len(prefix):], max) == 1 {
					max = iter.Key()[len(prefix):]
					break Loop
				} else if bytes.Equal(iter.Key()[len(prefix):], max) {
					break
				}
			}
			if !iter.Key().HasPrefix(prefix) {
				return BytesArray(ctx.Out, members), nil
			}
			setsIter[i] = iter
		}
		if i == len(ctx.Args) {
			members = append(members, max)
			if err := setsIter[0].Next(); err != nil {
				return nil, err
			}
			prefix, _ := GetPrefix(txn, []byte(keys[0]))
			if !setsIter[0].Key().HasPrefix(prefix) {
				return BytesArray(ctx.Out, members), nil
			}
			max = setsIter[0].Key()[len(prefix):]
		}

	}

	return BytesArray(ctx.Out, members), nil
}

// SDiff returns the members of the set resulting from the difference between the first set and all the successive sets.
func SDiff(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var members [][]byte
	var setsIter = make([]db.Iterator, len(ctx.Args)) //存储每个set当前的迭代器位置
	var keys = make([][]byte, len(ctx.Args))
	var min []byte
	var count int
	for i, key := range ctx.Args {
		set, err := txn.Set([]byte(key))
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		dkey := set.GetDataKey(txn, []byte(key))
		prefix := append(dkey, ':')
		iter, err := set.GetIter(prefix)
		if err != nil {
			return nil, err
		}
		defer iter.Close()
		setsIter[i] = iter
		keys[i] = []byte(key)
	}

	iprefix, _ := GetPrefix(txn, keys[0])
	min = setsIter[0].Key()[len(iprefix):]
	for {
	Loop:
		// check to see if the same element exists as the current membet for the benchmark key
		for i := 0; i < len(ctx.Args); i++ {
			prefix, _ := GetPrefix(txn, keys[i])
			if !setsIter[i].Key().HasPrefix(prefix) {
				continue
			}
			if bytes.Equal(min, setsIter[i].Key()[len(prefix):]) {
				if i == 0 || !setsIter[i].Key().HasPrefix(prefix) {
					continue
				}
				if err := setsIter[i].Next(); err != nil {
					return nil, err
				}
				if err := setsIter[0].Next(); err != nil {
					return nil, err
				}
				if !setsIter[0].Key().HasPrefix(iprefix) {
					return BytesArray(ctx.Out, members), nil
				}
				min = setsIter[0].Key()[len(iprefix):]
				goto Loop
			}
		}
		//find min in members
		for i := 0; i < len(ctx.Args); i++ {
			prefix, _ := GetPrefix(txn, keys[i])
			if !setsIter[i].Key().HasPrefix(prefix) {
				if i == 0 {
					return BytesArray(ctx.Out, members), nil
				}
				continue
			}
			if bytes.Compare(min, setsIter[i].Key()[len(prefix):]) == 1 {
				min = setsIter[i].Key()[len(prefix):]
			}
		}
		//Find the smallest element in the current member and move the pointer back
		for i := 0; i < len(ctx.Args); i++ {
			prefix, _ := GetPrefix(txn, keys[i])
			if !setsIter[i].Key().HasPrefix(prefix) {
				continue
			}
			if bytes.Equal(min, setsIter[i].Key()[len(prefix):]) {
				if i == 0 {
					members = append(members, min)
					if err := setsIter[0].Next(); err != nil {
						return nil, err
					}
					for bytes.Equal(min, setsIter[0].Key()[len(iprefix):]) {
						if err := setsIter[0].Next(); err != nil {
							return nil, err
						}
					}
				} else if setsIter[i].Key().HasPrefix(prefix) {
					if err := setsIter[i].Next(); err != nil {
						return nil, err
					}
				}
			}
		}
		if setsIter[0].Key().HasPrefix(iprefix) {
			min = setsIter[0].Key()[len(iprefix):]
		}

		var j int
		for i := 0; i < len(ctx.Args); i++ {
			prefix, _ := GetPrefix(txn, keys[i])
			if !setsIter[i].Key().HasPrefix(prefix) {
				j++
			}
		}
		count = j
		if count == len(ctx.Args) {
			break
		}

	}
	return BytesArray(ctx.Out, members), nil
}

// SliceIntersect returns slice that are present in all the slice1 and slice2.
func sliceInter(slice1, slice2 [][]byte) (interslice [][]byte) {
	for _, v := range slice1 {
		if inSliceInter(v, slice2) {
			interslice = append(interslice, v)
		}
	}
	return
}

// InSliceInter checks given interface in interface slice.
func inSliceInter(v []byte, sl [][]byte) bool {
	for _, vv := range sl {
		if bytes.Equal(vv, v) {
			return true
		}
	}
	return false
}

// SliceIntersect returns all slices in slice1 that are not present in slice2.
func sliceDiff(slice1, slice2 [][]byte) [][]byte {
	var diffslice [][]byte
	for _, v := range slice1 {
		if inSliceDiff(v, slice2) {
			diffslice = append(diffslice, v)
		}
	}
	return diffslice
}

// InSliceDiff checks given interface in interface slice.
func inSliceDiff(v []byte, sl [][]byte) bool {
	for _, vv := range sl {
		if bytes.Equal(vv, v) {
			return false
		}
	}
	return true
}
