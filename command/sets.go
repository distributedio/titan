package command

import (
	"gitlab.meitu.com/platform/titan/db"
)

// SAdd adds the specified members to the set stored at key
func SAdd(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	set, err := txn.Set(key)
	if err != nil {
		return nil, err
	}

	var members [][]byte
	for _, member := range ctx.Args[1:] {
		members = append(members, []byte(member))
	}

	added, err := set.SAdd(members)
	if err != nil {
		return nil, err
	}
	return Integer(ctx.Out, added), nil
}

// SMembers returns all the members of the set value stored at key
func SMembers(ctx *Context, txn *db.Transaction) (OnCommit, error) {
	key := []byte(ctx.Args[0])

	set, err := txn.Set(key)
	if err != nil {
		return nil, err
	}

	members, err := set.SMembers()
	if err != nil {
		return nil, err
	}
	return BytesArray(ctx.Out, members), nil
}
