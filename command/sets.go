package command

import (
	"bytes"
	"errors"
	"strconv"

	"github.com/distributedio/titan/db"
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
	count := 0
	var err error
	key := []byte(ctx.Args[0])
	if len(ctx.Args) == 2 {
		count, err = strconv.Atoi(ctx.Args[1])
		if err != nil {
			return nil, ErrInteger
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

// SUnion returns the members of the set resulting from the union of all the given sets.
func SUnion(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var members [][]byte
	setsIter := make([]*db.SetIter, len(ctx.Args))
	for i, key := range ctx.Args {
		set, err := txn.Set([]byte(key))
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		siter, err := set.Iter()
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		defer siter.Iter.Close()
		setsIter[i] = siter
	}
	sum := getNodeCount(len(setsIter))
	k := sum - len(setsIter) + 1
	ls := make([]int, k)
	idx := CreateLoserTree(ls, setsIter)
	for {
		sum := getNodeCount(len(setsIter))
		k := sum - len(setsIter) + 1
		ls := make([]int, k)
		min := adjustMin(ls, setsIter, idx)
		l := len(setsIter)
		for i, j := 0, 0; i < l; i, j = i+1, j+1 {
			if bytes.Equal(setsIter[j].Value(), min) {
				if err := setsIter[j].Iter.Next(); err != nil {
					return nil, errors.New("ERR " + err.Error())
				}
			}
			if !setsIter[j].Valid() {
				setsIter = append(setsIter[:j], setsIter[j+1:]...)
				j--
			}
		}
		members = append(members, min)
		if len(setsIter) <= 0 {
			break
		}
	}
	return BytesArray(ctx.Out, members), nil
}

// CreateLoserTree creates losertree
func CreateLoserTree(ls []int, sets []*db.SetIter) int {
	for i := len(ls) - 1; i >= 0; i-- {
		adjustMin(ls, sets, i)
	}
	//res := sets[ls[0]].Value()
	return ls[0]

}

// getNodeCount calculaties the total number of nodes in a complete binary tree
func getNodeCount(count int) (sum int) {
	if count == 2 {
		sum = 3
	} else if count%2 == 0 {
		if count%4 == 0 {
			sum = count*2 - 1
		} else {
			sum = count * 2
		}
	} else {
		sum = count*2 - 1
	}
	return
}

// adjustMin adjusts the loser tree and return the smallest element
func adjustMin(ls []int, sets []*db.SetIter, s int) []byte {
	t := (s + len(ls)) / 2
	for t > 0 {
		if bytes.Compare(sets[s].Value(), sets[ls[t]].Value()) == 1 {
			swap := s
			s = ls[t]
			ls[t] = swap
		}
		t = t / 2
	}
	res := sets[s].Value()
	return res
}

// SInter returns the members of the set resulting from the intersection of all the given sets.
func SInter(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var members [][]byte
	setsIter := make([]*db.SetIter, len(ctx.Args))
	for i, key := range ctx.Args {
		set, err := txn.Set([]byte(key))
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		setlen, err := set.SCard()
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		// If the set corresponding to key does not exist, it is processed as an empty set
		if !set.Exists() || setlen == 0 {
			return nil, nil
		}
		siter, err := set.Iter()
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		defer siter.Iter.Close()
		setsIter[i] = siter
	}
	for {
		max := getMaxMember(setsIter)
		for i := 0; i < len(ctx.Args); i++ {
			if !bytes.Equal(setsIter[i].Value(), max) {
				if err := setsIter[i].Iter.Next(); err != nil {
					return nil, errors.New("ERR " + err.Error())
				}
				if !setsIter[i].Valid() {
					return BytesArray(ctx.Out, members), nil
				}
			}
			if i == len(ctx.Args)-1 {
				members = append(members, max)
				for j := 0; j < len(ctx.Args); j++ {
					if err := setsIter[j].Iter.Next(); err != nil {
						return nil, errors.New("ERR " + err.Error())
					}
					if !setsIter[j].Valid() {
						return BytesArray(ctx.Out, members), nil
					}
				}
			}
		}
	}
	return BytesArray(ctx.Out, members), nil
}

// getMaxMember gets the maximum value in members
func getMaxMember(sets []*db.SetIter) []byte {
	var arr [][]byte
	for i := 0; i < len(sets); i++ {
		arr = append(arr, sets[i].Value())
	}
	for k := len(arr) / 2; k >= 0; k-- {
		adjustHeap(arr, k)
	}
	res := arr[0]
	return res
}

// adjustHeap adjust big root heap
func adjustHeap(arr [][]byte, k int) {
	for {
		i := 2 * k
		if i > len(arr)-1 { //保证该节点是非叶子节点
			break
		}
		if i+1 < len(arr) && bytes.Compare(arr[i+1], arr[i]) == 1 { //选择较大的子节点
			i++
		}
		if bytes.Compare(arr[k], arr[i]) == 1 || bytes.Equal(arr[k], arr[i]) {
			break
		}
		swap(arr, k, i)
		k = i
	}
}

// swap swaps two values
func swap(s [][]byte, i int, j int) {
	s[i], s[j] = s[j], s[i]
}

// SDiff returns the members of the set resulting from the difference between the first set and all the successive sets.
func SDiff(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var members [][]byte
	var err error
	setsIter := make([]*db.SetIter, len(ctx.Args))
	for i, key := range ctx.Args {
		set, err := txn.Set([]byte(key))
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		siter, err := set.Iter()
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		defer siter.Iter.Close()
		setsIter[i] = siter
	}
	min := setsIter[0].Value()
	for {
		// check to see if the same element exists as the current membet for the benchmark key
		for {
			match, err := check(setsIter, min, len(ctx.Args))
			if err != nil {
				return nil, errors.New("ERR " + err.Error())
			}
			if match {
				if err := setsIter[0].Iter.Next(); err != nil {
					return nil, errors.New("ERR " + err.Error())
				}
				if !setsIter[0].Valid() {
					return BytesArray(ctx.Out, members), nil
				}
				min = setsIter[0].Value()
			} else {
				break
			}
		}
		//find min in members
		for i := 0; i < len(ctx.Args); i++ {
			if !setsIter[i].Valid() {
				if i == 0 {
					return BytesArray(ctx.Out, members), nil
				}
				continue
			}
			if bytes.Compare(min, setsIter[i].Value()) == 1 {
				min = setsIter[i].Value()
			}
		}
		members, err = moveMembers(setsIter, min, members)
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}

		if setsIter[0].Valid() {
			min = setsIter[0].Value()
		}

		if stopcirculation(setsIter, len(ctx.Args)) == len(ctx.Args) {
			break
		}
	}
	return BytesArray(ctx.Out, members), nil
}

// stopcirculation determines when to stop the loop
func stopcirculation(sets []*db.SetIter, length int) (count int) {
	for i := 0; i < length; i++ {
		if !sets[i].Valid() {
			count++
		}
	}
	return
}

// check checks to see if the same element exists as the current membet for the benchmark key
func check(sets []*db.SetIter, min []byte, length int) (bool, error) {
	for i := 1; i < length; i++ {
		if !sets[i].Valid() {
			continue
		}
		if bytes.Equal(min, sets[i].Value()) {
			if err := sets[i].Iter.Next(); err != nil {
				return false, errors.New("ERR " + err.Error())
			}
			return true, nil
		}
	}
	return false, nil
}

// moveMembers finds the smallest element in the current member and move the pointer back
func moveMembers(setsIter []*db.SetIter, min []byte, members [][]byte) ([][]byte, error) {
	for i := 0; i < len(setsIter); i++ {
		if !setsIter[i].Valid() {
			continue
		}
		if bytes.Equal(min, setsIter[i].Value()) {
			if i == 0 {
				members = append(members, min)
				if err := setsIter[0].Iter.Next(); err != nil {
					return nil, errors.New("ERR " + err.Error())
				}
				for bytes.Equal(min, setsIter[0].Value()) {
					if err := setsIter[0].Iter.Next(); err != nil {
						return nil, errors.New("ERR " + err.Error())
					}
				}
			} else if setsIter[i].Valid() {
				if err := setsIter[i].Iter.Next(); err != nil {
					return nil, errors.New("ERR " + err.Error())
				}
			}
		}
	}
	return members, nil
}
