package command

import (
	"bytes"
	"container/heap"
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
	count := 1
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
	var setsIter []*db.SetIter
	for _, key := range ctx.Args {
		set, err := txn.Set([]byte(key))
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		if !set.Exists() {
			continue
		}
		siter, err := set.Iter()
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		defer siter.Iter.Close()
		setsIter = append(setsIter, siter)
	}

	h := MinHeap(setsIter)
	heap.Init(&h)
	var last []byte
	for len(h) != 0 {
		min := h[0].Value()

		// ignore the duplicated member
		if last == nil || !bytes.Equal(min, last) {
			members = append(members, min)
		}
		last = min

		if err := h[0].Iter.Next(); err != nil {
			return nil, err
		}
		if h[0].Valid() {
			heap.Fix(&h, 0)
		} else {
			heap.Remove(&h, 0)
		}
	}
	return BytesArray(ctx.Out, members), nil
}

type MinHeap []*db.SetIter

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return bytes.Compare(h[i].Value(), h[j].Value()) < 0 }
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *MinHeap) Push(x interface{}) {
	item := x.(*db.SetIter)
	*h = append(*h, item)
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
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
		// If the set corresponding to key does not exist, it is processed as an empty set
		if !set.Exists() {
			return BytesArray(ctx.Out, members), nil
		}
		siter, err := set.Iter()
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		defer siter.Iter.Close()
		setsIter[i] = siter
	}

	h := MinHeap(setsIter)
	heap.Init(&h)
	var last []byte
	k := len(h) //k-way merge
	n := k
	for len(h) != 0 {
		min := h[0].Value()
		if last == nil || bytes.Equal(last, min) {
			n--
		} else {
			n = k - 1
		}
		last = min

		// it is a member of intersection if there are k equal values continuously
		if n == 0 {
			members = append(members, min)
		}

		if err := h[0].Iter.Next(); err != nil {
			return nil, err
		}
		if h[0].Valid() {
			heap.Fix(&h, 0)
		} else {
			heap.Remove(&h, 0)
		}
	}
	return BytesArray(ctx.Out, members), nil
}

// SDiff returns the members of the set resulting from the difference between the first set and all the successive sets.
func SDiff(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	var members [][]byte
	var setsIter []*db.SetIter
	for i, key := range ctx.Args {
		set, err := txn.Set([]byte(key))
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		if !set.Exists() && i != 0 {
			continue
		}
		siter, err := set.Iter()
		if err != nil {
			return nil, errors.New("ERR " + err.Error())
		}
		defer siter.Iter.Close()
		setsIter = append(setsIter, siter)
	}
	iter := setsIter[0]
	h := MinHeap(setsIter[1:])
	heap.Init(&h)
	for iter.Valid() {
		member := iter.Value()
		if len(h) == 0 {
			members = append(members, member)
			if err := iter.Iter.Next(); err != nil {
				return nil, errors.New("ERR " + err.Error())
			}
			continue
		}

		min := h[0].Value()
		switch bytes.Compare(member, min) {
		case -1:
			members = append(members, member)
			if err := iter.Iter.Next(); err != nil {
				return nil, errors.New("ERR " + err.Error())
			}
		case 0:
			if err := iter.Iter.Next(); err != nil {
				return nil, errors.New("ERR " + err.Error())
			}
			fallthrough
		case 1:
			if err := h[0].Iter.Next(); err != nil {
				return nil, errors.New("ERR " + err.Error())
			}
			if h[0].Valid() {
				heap.Fix(&h, 0)
			} else {
				heap.Remove(&h, 0)
			}
		}
	}
	return BytesArray(ctx.Out, members), nil
}
