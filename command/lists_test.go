package command

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func initList(t *testing.T, key string, length int) {
	args := []string{key}
	for i := length; i >= 1; i-- {
		args = append(args, strconv.Itoa(i))
	}
	ctx := ContextTest("lpush", args...)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	strlen := strconv.Itoa(length)
	assert.Equal(t, ":"+strlen, lines[0])
}

func clearList(t *testing.T, key string) {
	ctx := ContextTest("del", key)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])
}

func rpushList(t *testing.T, key, val string) []string {
	ctx := ContextTest("rpush", key, val)
	Call(ctx)
	return ctxLines(ctx.Out)

}

func TestLLen(t *testing.T) {
	// init
	key := "list-llen-key"
	initList(t, key, 3)

	// case 1
	ctx := ContextTest("llen", key)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":3", lines[0])

	// case 2
	lines = rpushList(t, key, "4")
	assert.Equal(t, ":4", lines[0])
	ctx = ContextTest("llen", key)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":4", lines[0])
	// end
	clearList(t, key)

	// zlist init
	key = "list-llen-zlistkey"
	initList(t, key, 600)

	ctx = ContextTest("llen", key)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":600", lines[0])

	lines = rpushList(t, key, "lastval")
	assert.Equal(t, ":601", lines[0])
	ctx = ContextTest("llen", key)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":601", lines[0])
	// end
	clearList(t, key)

}

func TestLIndex(t *testing.T) {
	// init
	key := "list-lindex-list"
	initList(t, key, 3)

	// case 1
	ctx := ContextTest("lindex", key, "0")
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, "1", lines[1])

	// case 2
	ctx = ContextTest("lindex", key, "-1")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "3", lines[1])

	// case 4
	ctx = ContextTest("lindex", key, "-100")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "$-1", lines[0])

	// end
	clearList(t, key)

	// init zlikst
	key = "list-lindex-zlist"
	initList(t, key, 600)

	// case 1
	ctx = ContextTest("lindex", key, "0")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "1", lines[1])

	// case 2
	ctx = ContextTest("lindex", key, "-1")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "600", lines[1])

	// case 4
	ctx = ContextTest("lindex", key, "-700")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "$-1", lines[0])

	// end
	clearList(t, key)

}

func TestLInsert(t *testing.T) {
	key := "list-linsert-list"
	initList(t, key, 3)

	// case 1
	ctx := ContextTest("linsert", key, "BeFoRe", "2", "a1")
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":4", lines[0])
	// case 2
	ctx = ContextTest("linsert", key, "afTER", "2", "b1")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":5", lines[0])

	// end
	ctx = ContextTest("lrange", key, "0", "-1")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "a1", lines[4])
	assert.Equal(t, "b1", lines[8])
	clearList(t, key)

	key = "list-linsert-zlist"
	initList(t, key, 600)

	// case 1
	ctx = ContextTest("linsert", key, "BeFoRe", "2", "a1")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":601", lines[0])
	// case 2
	ctx = ContextTest("linsert", key, "afTER", "2", "b1")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":602", lines[0])

	// end
	ctx = ContextTest("lrange", key, "0", "-1")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "a1", lines[4])
	assert.Equal(t, "b1", lines[8])
	clearList(t, key)
}

func TestLPop(t *testing.T) {
	// init
	key := "list-lpop-list"
	initList(t, key, 3)

	// case 1
	ctx := ContextTest("lpop", key)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, "1", lines[1])
	// case 2
	ctx = ContextTest("lpop", key)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "2", lines[1])
	// case 3
	ctx = ContextTest("lpop", key)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "3", lines[1])
	// case 4
	ctx = ContextTest("lpop", key)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "$-1", lines[0])

	key = "list-lpop-zlist"
	initList(t, key, 600)

	for i := 1; i <= 600; i++ {
		stri := strconv.Itoa(i)
		ctx = ContextTest("lpop", key)
		Call(ctx)
		lines = ctxLines(ctx.Out)
		assert.Equal(t, stri, lines[1], "test zlist lpop error")
	}
	ctx = ContextTest("lpop", key)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "$-1", lines[0])

}

func TestLPush(t *testing.T) {
	// init
	key := "list-lpush-list"
	initList(t, key, 3)
	ctx := ContextTest("lpush", key, "b")
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":4", lines[0])
	ctx = ContextTest("lrange", key, "0", "-1")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Contains(t, lines, "b")
	clearList(t, key)

	key = "list-lpush-zlist"
	initList(t, key, 600)
	ctx = ContextTest("lpush", key, "zlist")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":601", lines[0])
	ctx = ContextTest("lrange", key, "0", "-1")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Contains(t, lines, "zlist")
	clearList(t, key)
}

func TestLPushx(t *testing.T) {
	key := "list-lpushx-list"
	ctx := ContextTest("lpushx", key, "b")
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":0", lines[0])
	initList(t, key, 3)
	ctx = ContextTest("lpushx", key, "b")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":4", lines[0])

	ctx = ContextTest("lrange", key, "0", "-1")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Contains(t, lines, "b")
	clearList(t, key)

	key = "list-lpushx-zlist"
	initList(t, key, 600)
	ctx = ContextTest("lpushx", key, "zlist")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":601", lines[0])
	ctx = ContextTest("lrange", key, "0", "-1")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Contains(t, lines, "zlist")
	clearList(t, key)
}

func TestLRange(t *testing.T) {
	key := "list-lrange-list"
	initList(t, key, 3)
	ctx := ContextTest("lrange", key, "-6", "1")
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, "1", lines[2])
	assert.Equal(t, "2", lines[4])

	ctx = ContextTest("lrange", key, "-7", "-2")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "1", lines[2])
	assert.Equal(t, "2", lines[4])

	ctx = ContextTest("lrange", key, "-2", "-5")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "*0", lines[0])

	ctx = ContextTest("lrange", key, "0", "2")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "1", lines[2])
	assert.Equal(t, "2", lines[4])
	assert.Equal(t, "3", lines[6])

	ctx = ContextTest("lrange", key, "0", "0")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "1", lines[2])
	clearList(t, key)

	key = "list-lrange-zlist"
	initList(t, key, 600)
	ctx = ContextTest("lrange", key, "-6", "1")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "*0", lines[0])

	ctx = ContextTest("lrange", key, "-7", "-2")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Contains(t, lines, "599")

	ctx = ContextTest("lrange", key, "-2", "-5")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "*0", lines[0])

	ctx = ContextTest("lrange", key, "0", "2")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "1", lines[2])

	ctx = ContextTest("lrange", key, "0", "0")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "1", lines[2])
	clearList(t, key)

}

func TestLSet(t *testing.T) {
	key := "list-lset-list"
	initList(t, key, 3)

	ctx := ContextTest("lset", key, "0")
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, "-ERR wrong number of arguments for 'lset' command", lines[0])

	ctx = ContextTest("lset", key, "0", "testval")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "+OK", lines[0])

	ctx = ContextTest("lindex", key, "0")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "testval", lines[1])

	ctx = ContextTest("lset", key, "-1", "aa")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "+OK", lines[0])

	ctx = ContextTest("lset", key, "100", "bb")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "-ERR index out of range", lines[0])
	clearList(t, key)

	key = "list-lset-zlist"
	initList(t, key, 600)

	ctx = ContextTest("lset", key, "0")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "-ERR wrong number of arguments for 'lset' command", lines[0])

	ctx = ContextTest("lset", key, "0", "testval")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "+OK", lines[0])

	ctx = ContextTest("lindex", key, "0")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "testval", lines[1])

	ctx = ContextTest("lset", key, "-1", "aa")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "+OK", lines[0])

	ctx = ContextTest("lset", key, "10000", "bb")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "-ERR index out of range", lines[0])

}

func TestRPush(t *testing.T) {
	key := "list-rpush-list"
	ctx := ContextTest("rpush", key, "first")
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])
	ctx = ContextTest("lrange", key, "0", "-1")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Contains(t, lines, "first")
	clearList(t, key)

	initList(t, key, 3)
	ctx = ContextTest("rpush", key, "aa")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":4", lines[0])
	ctx = ContextTest("lrange", key, "0", "-1")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "aa", lines[8])
	clearList(t, key)

	key = "list-rpush-zlist"
	initList(t, key, 600)
	ctx = ContextTest("rpush", key, "aa")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":601", lines[0])
	ctx = ContextTest("lrange", key, "0", "-1")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "aa", lines[1202])
	clearList(t, key)

}

func TestRPushx(t *testing.T) {
	key := "list-rpushx-list"
	ctx := ContextTest("rpushx", key, "first")
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":0", lines[0])

	initList(t, key, 3)
	ctx = ContextTest("rpushx", key, "aa")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":4", lines[0])
	ctx = ContextTest("lrange", key, "0", "-1")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "aa", lines[8])
	clearList(t, key)

	key = "list-rpushx-zlist"
	initList(t, key, 600)
	ctx = ContextTest("rpushx", key, "aa")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":601", lines[0])
	ctx = ContextTest("lrange", key, "0", "-1")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "aa", lines[1202])
	clearList(t, key)

}
