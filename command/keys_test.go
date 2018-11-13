package command

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func InitData(t *testing.T, keys []string, val string) {
	for _, key := range keys {
		ctx := ContextTest("set", key, val)
		Call(ctx)
	}
}

func AddList(t *testing.T, key string, val string) {
	ctx := ContextTest("lpush", key, val)
	Call(ctx)
}

func EquealKeyExists(t *testing.T, key string) {
	ctx := ContextTest("exists", key)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])

}

func NotEquealKeyExists(t *testing.T, key string) {
	ctx := ContextTest("exists", key)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":0", lines[0])
}

func TestDel(t *testing.T) {
	keys := []string{
		"keys-del-key1",
		"keys-del-key2",
		"keys-del-key3",
		"keys-del-key4",
		"keys-del-key5",
		"keys-del-key6",
	}

	InitData(t, keys, "val")
	ctx := ContextTest("del", keys[0], keys[1], keys[2], keys[3], keys[4], keys[5])
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":6", lines[0])
	NotEquealKeyExists(t, keys[0])

	InitData(t, keys, "val")
	ctx = ContextTest("del", keys[0], keys[1], keys[2], keys[3], keys[4], keys[5], "keys-del-faild")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":6", lines[0])
	NotEquealKeyExists(t, keys[0])
}

func TestExists(t *testing.T) {
	keys := []string{
		"keys-keyexists1",
		"keys-keyexists2",
		"keys-keyexists3",
		"keys-keyexists4",
		"keys-keyexists5",
		"keys-keyexists6",
	}

	InitData(t, keys, "val")
	ctx := ContextTest("exists", keys[0], keys[1], keys[2], keys[3], keys[4], keys[5])
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":6", lines[0])

	ctx = ContextTest("exists", keys[0], keys[1], keys[2], keys[3], keys[4], keys[5], "keys-keyexists-faild")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":6", lines[0])

	ctx = ContextTest("del", keys[0], keys[1], keys[2], keys[3], keys[4], keys[5])
	Call(ctx)
}

func TestExpireAt(t *testing.T) {
	keys := []string{"keys-expireat1", "keys-expireat2", "keys-expireat3"}
	InitData(t, keys, "val")

	now := time.Now().Unix()

	time1 := now + 10
	ctx := ContextTest("expireat", keys[0], strconv.FormatInt(time1, 10))
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])
	EquealKeyExists(t, keys[0])

	time2 := "0"
	ctx = ContextTest("expireat", keys[1], time2)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])
	NotEquealKeyExists(t, keys[1])

	time3 := "-1"
	ctx = ContextTest("expireat", keys[2], time3)
	Call(ctx)
	lines = ctxLines(ctx.Out)

	assert.Equal(t, ":1", lines[0])
	NotEquealKeyExists(t, keys[2])
}

func TestExpire(t *testing.T) {
	key1 := "keys-expire1"
	key2 := "keys-expire2"
	key3 := "keys-expire3"
	keys := []string{
		key1, key2, key3,
	}
	InitData(t, keys, "val")

	time1 := "10"
	ctx := ContextTest("expire", keys[0], time1)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])
	EquealKeyExists(t, keys[0])

	time2 := "0"
	ctx = ContextTest("expire", keys[1], time2)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])
	NotEquealKeyExists(t, keys[1])

	time3 := "-1"
	ctx = ContextTest("expire", keys[2], time3)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])
	NotEquealKeyExists(t, keys[2])
}

func TestPExpire(t *testing.T) {
	key1 := "keys-pexpire1"
	key2 := "keys-pexpire2"
	key3 := "keys-pexpire3"
	keys := []string{
		key1, key2, key3,
	}
	InitData(t, keys, "val")

	time1 := "10"
	ctx := ContextTest("pexpire", keys[0], time1)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])
	EquealKeyExists(t, keys[0])

	time2 := "0"
	ctx = ContextTest("pexpire", keys[1], time2)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])
	NotEquealKeyExists(t, keys[1])

	time3 := "-1"
	ctx = ContextTest("pexpire", keys[2], time3)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])
	NotEquealKeyExists(t, keys[2])

}

func TestPExpireAt(t *testing.T) {
	key1 := "keys-pexpireat1"
	key2 := "keys-pexpireat2"
	key3 := "keys-pexpireat3"
	keys := []string{
		key1, key2, key3,
	}
	InitData(t, keys, "val")

	now := time.Now().Unix()

	time1 := (now + 10) * int64(time.Second) / int64(time.Millisecond)
	ctx := ContextTest("pexpireat", keys[0], strconv.FormatInt(time1, 10))
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])
	EquealKeyExists(t, keys[0])

	time2 := "0"
	ctx = ContextTest("pexpireat", keys[1], time2)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])
	NotEquealKeyExists(t, keys[1])

	time3 := "-1"
	ctx = ContextTest("pexpireat", keys[2], time3)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])
	NotEquealKeyExists(t, keys[2])
}

func TestTTL(t *testing.T) {
	val := "val"
	key1 := "keys-ttl1"

	InitData(t, []string{key1}, val)

	ctx := ContextTest("ttl", key1)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":-1", lines[0])

	ctx = ContextTest("expire", key1, "2")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])

	ctx = ContextTest("ttl", key1)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.NotEqual(t, ":-1", lines[0])
	assert.NotEqual(t, ":-2", lines[0])
	assert.NotEqual(t, ":0", lines[0])

	time.Sleep(1 * time.Second)

	ctx = ContextTest("ttl", key1)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.NotEqual(t, ":-2", lines[0])
}

func TestPTTL(t *testing.T) {
	val := "val"
	key1 := "keys-pttl1"

	InitData(t, []string{key1}, val)

	ctx := ContextTest("pttl", key1)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":-1", lines[0])

	ctx = ContextTest("expire", key1, "2")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])

	ctx = ContextTest("pttl", key1)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.NotEqual(t, ":-1", lines[0])
	assert.NotEqual(t, ":-2", lines[0])
	assert.NotEqual(t, ":0", lines[0])

	time.Sleep(1 * time.Second)

	ctx = ContextTest("pttl", key1)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.NotEqual(t, ":-2", lines[0])
}

func TestPerist(t *testing.T) {
	key := "keys-pexist1"
	val := "val"
	InitData(t, []string{key}, val)

	ctx := ContextTest("expire", key, "5")
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])

	ctx = ContextTest("persist", key)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])

	ctx = ContextTest("ttl", key)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":-1", lines[0])
}

func TestType(t *testing.T) {
	key := "keys-type1"
	val := "val"
	InitData(t, []string{key}, val)

	lkey := "keys-type-list1"
	AddList(t, lkey, val)

	ctx := ContextTest("type", key)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, "+string", lines[0])

	ctx = ContextTest("type", lkey)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "+list", lines[0])

	ctx = ContextTest("type", "keys-type-faild")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "+none", lines[0])
}

func TestKeys(t *testing.T) {
	keys := []string{
		"keys-abc1:keys",
		"keys-acb1:keys",
		"keys-abc2:keys",
	}
	val := "val"
	InitData(t, keys, val)

	ctx := ContextTest("keys", "keys*:keys")
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, "*3", lines[0])
	assert.Contains(t, lines, "keys-abc1:keys")

	ctx = ContextTest("keys", "keys-ab*")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "*2", lines[0])
	assert.Contains(t, lines, "keys-abc1:keys")
}

func TestScan(t *testing.T) {
	keys := []string{
		"keys-scan1",
		"keys-scan2",
		"keys-scan3",
		"keys-scan4",
		"keys-sscan5",
	}
	val := "val"
	InitData(t, keys, val)

	ctx := ContextTest("scan", "0", "count", "4", "match", "keys-scan*")
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, "*2", lines[0])
	assert.Contains(t, lines, "keys-scan4")
	assert.Equal(t, "keys-sscan5", lines[2])
}

func TestObject(t *testing.T) {
	key := "keys-object1"
	val := "val"
	InitData(t, []string{key}, val)
	lkey := "keys-objectlist1"
	AddList(t, lkey, val)

	ctx := ContextTest("object", "encoding", key)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, "+raw", lines[0])

	ctx = ContextTest("object", "encoding", lkey)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "+linkedlist", lines[0])

	time.Sleep(time.Second)
	ctx = ContextTest("object", "idletime", lkey)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.NotEqual(t, ":0", lines[0])

}

func TestRandomkey(t *testing.T) {
	keys := []string{
		"keyscan1",
		"keyscan2",
		"keyscan3",
		"keyscan4",
		"skeyscan5",
	}
	val := "val"
	InitData(t, keys, val)

	ctx := ContextTest("randomkey")
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.NotEqual(t, 0, len(lines))
}
