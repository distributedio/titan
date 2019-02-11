package command

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func initSets(t *testing.T, key string, length int) {
	args := []string{key}
	for i := length; i > 0; i-- {
		args = append(args, strconv.Itoa(i))
	}
	ctx := ContextTest("sadd", args...)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	strlen := strconv.Itoa(length)
	assert.Equal(t, ":"+strlen, lines[0])
}

func clearSets(t *testing.T, key string) {
	ctx := ContextTest("del", key)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])
}

func setSets(t *testing.T, args ...string) []string {
	ctx := ContextTest("sadd", args...)
	Call(ctx)
	return ctxLines(ctx.Out)
}

func TestAdd(t *testing.T) {
	key := "set-sadd"
	//initSets(t, key, 3)

	//case1
	ctx := ContextTest("sadd", key, "1", "2", "3")
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":3", lines[0])

	//case 2
	lines = setSets(t, key, "1", "2")
	assert.Equal(t, ":0", lines[0])

	//case 3
	lines = setSets(t, key, "3", "4")
	assert.Equal(t, ":1", lines[0])

	// end
	clearSets(t, key)

}

func TestSCard(t *testing.T) {
	// init
	key := "set-scard"
	initSets(t, key, 3)

	// case 1
	ctx := ContextTest("scard", key)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":3", lines[0])

	// case 2
	lines = setSets(t, key, "a", "b")
	assert.Equal(t, ":2", lines[0])
	ctx = ContextTest("scard", key)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":5", lines[0])

	// case 3
	lines = setSets(t, key, "a", "c")
	assert.Equal(t, ":1", lines[0])
	ctx = ContextTest("scard", key)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":6", lines[0])

	// end
	clearSets(t, key)
}

func TestSMembers(t *testing.T) {
	// init
	key := "set-smembers"

	//case1
	ctx := ContextTest("sadd", key, "1", "2", "3")
	Call(ctx)
	ctx = ContextTest("smembers", key)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, "*3", lines[0])
	assert.Equal(t, "$1", lines[1])
	assert.Equal(t, "1", lines[2])
	assert.Equal(t, "$1", lines[3])
	assert.Equal(t, "2", lines[4])
	assert.Equal(t, "$1", lines[5])
	assert.Equal(t, "3", lines[6])

	//end
	clearSets(t, key)
}

func TestSIsmember(t *testing.T) {

	key := "set-ismember"

	ctx := ContextTest("sadd", key, "1", "2", "3")
	Call(ctx)

	//case 1
	ctx = ContextTest("sismember", key, "1")
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])
	//case 2
	ctx = ContextTest("sismember", key, "4")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":0", lines[0])

	//end
	clearSets(t, key)
}
func TestSPop(t *testing.T) {
	key := "set-spop"

	ctx := ContextTest("sadd", key, "1", "2", "3", "4", "5")
	Call(ctx)

	//case 1
	ctx = ContextTest("spop", key)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, "*1", lines[0])
	assert.Equal(t, "$1", lines[1])
	assert.Equal(t, "1", lines[2])

	ctx = ContextTest("smembers", key)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "*4", lines[0])
	assert.Equal(t, "$1", lines[1])
	assert.Equal(t, "2", lines[2])
	assert.Equal(t, "$1", lines[3])
	assert.Equal(t, "3", lines[4])
	assert.Equal(t, "$1", lines[5])
	assert.Equal(t, "4", lines[6])
	assert.Equal(t, "$1", lines[7])
	assert.Equal(t, "5", lines[8])

	//case 2
	ctx = ContextTest("spop", key, "2")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "*2", lines[0])
	assert.Equal(t, "$1", lines[1])
	assert.Equal(t, "2", lines[2])
	assert.Equal(t, "$1", lines[3])
	assert.Equal(t, "3", lines[4])

	ctx = ContextTest("smembers", key)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "*2", lines[0])
	assert.Equal(t, "$1", lines[1])
	assert.Equal(t, "4", lines[2])
	assert.Equal(t, "$1", lines[3])
	assert.Equal(t, "5", lines[4])

	clearSets(t, key)
}
func TestSRem(t *testing.T) {

	key := "set-srem"

	ctx := ContextTest("sadd", key, "1", "2", "3")
	Call(ctx)

	//case 1
	ctx = ContextTest("srem", key, "1")
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])

	ctx = ContextTest("smembers", key)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "*2", lines[0])
	assert.Equal(t, "$1", lines[1])
	assert.Equal(t, "2", lines[2])
	assert.Equal(t, "$1", lines[3])
	assert.Equal(t, "3", lines[4])
	//case 2
	ctx = ContextTest("srem", key, "5")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":0", lines[0])

	ctx = ContextTest("smembers", key)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "*2", lines[0])
	assert.Equal(t, "$1", lines[1])
	assert.Equal(t, "2", lines[2])
	assert.Equal(t, "$1", lines[3])
	assert.Equal(t, "3", lines[4])

	//end
	clearSets(t, key)
}
func TestSMove(t *testing.T) {

	key := "set-smove"
	destkey := "set-dest-key"
	ctx := ContextTest("sadd", key, "1", "2", "3")
	Call(ctx)
	ctx = ContextTest("sadd", destkey, "3", "4")
	Call(ctx)

	//case 1
	ctx = ContextTest("smove", key, destkey, "1")
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])

	ctx = ContextTest("smembers", key)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "*2", lines[0])
	assert.Equal(t, "$1", lines[1])
	assert.Equal(t, "2", lines[2])
	assert.Equal(t, "$1", lines[3])
	assert.Equal(t, "3", lines[4])
	ctx = ContextTest("smembers", destkey)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "*3", lines[0])
	assert.Equal(t, "$1", lines[1])
	assert.Equal(t, "1", lines[2])
	assert.Equal(t, "$1", lines[3])
	assert.Equal(t, "3", lines[4])
	assert.Equal(t, "$1", lines[5])
	assert.Equal(t, "4", lines[6])
	//case 2
	ctx = ContextTest("smove", key, destkey, "5")
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, ":0", lines[0])

	ctx = ContextTest("smembers", key)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "*2", lines[0])
	assert.Equal(t, "$1", lines[1])
	assert.Equal(t, "2", lines[2])
	assert.Equal(t, "$1", lines[3])
	assert.Equal(t, "3", lines[4])

	//end
	clearSets(t, key)
	clearSets(t, destkey)
}
func TestSUnion(t *testing.T) {
	key1 := "set-sunion1"
	key2 := "set-sunion2"
	key3 := "set-sunion3"

	ctx := ContextTest("sadd", key1, "a", "b", "c", "d")
	Call(ctx)
	ctx = ContextTest("sadd", key2, "c", "d", "e")
	Call(ctx)
	ctx = ContextTest("sadd", key3, "")
	Call(ctx)

	//case 1
	ctx = ContextTest("sunion", key1, key2)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, "*5", lines[0])

	//case 2
	ctx = ContextTest("sunion", key1, key3)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "*5", lines[0])
	//case 3
	ctx = ContextTest("sunion", key2, key3)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "*4", lines[0])
	//end
	clearSets(t, key1)

	clearSets(t, key2)

	clearSets(t, key3)
}
func TestSInter(t *testing.T) {
	key1 := "set-sinter1"
	key2 := "set-sinter2"
	key3 := "set-sinter3"

	ctx := ContextTest("sadd", key1, "a", "b", "c", "d")
	Call(ctx)
	ctx = ContextTest("sadd", key2, "c")
	Call(ctx)
	ctx = ContextTest("sadd", key3, "a", "c", "e")
	Call(ctx)

	//case 1
	ctx = ContextTest("sinter", key1, key2, key3)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, "*1", lines[0])
	assert.Equal(t, "$1", lines[1])
	assert.Equal(t, "c", lines[2])

	//case 2
	ctx = ContextTest("sinter", key1, key3)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "*2", lines[0])
	assert.Equal(t, "$1", lines[1])
	assert.Equal(t, "a", lines[2])
	assert.Equal(t, "$1", lines[3])
	assert.Equal(t, "c", lines[4])

	//end
	clearSets(t, key1)

	clearSets(t, key2)

	clearSets(t, key3)

}
func TestSDiff(t *testing.T) {
	key1 := "set-sdiff1"
	key2 := "set-sdiff2"
	key3 := "set-sdiff3"

	ctx := ContextTest("sadd", key1, "a", "b", "c", "d")
	Call(ctx)
	ctx = ContextTest("sadd", key2, "c")
	Call(ctx)
	ctx = ContextTest("sadd", key3, "a", "c", "e")
	Call(ctx)

	//case 1
	ctx = ContextTest("sdiff", key1, key2)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, "*3", lines[0])
	assert.Equal(t, "$1", lines[1])
	assert.Equal(t, "a", lines[2])
	assert.Equal(t, "$1", lines[3])
	assert.Equal(t, "b", lines[4])
	assert.Equal(t, "$1", lines[5])
	assert.Equal(t, "d", lines[6])

	//case 2
	ctx = ContextTest("sdiff", key1, key2, key3)
	Call(ctx)
	lines = ctxLines(ctx.Out)
	assert.Equal(t, "*2", lines[0])
	assert.Equal(t, "$1", lines[1])
	assert.Equal(t, "b", lines[2])
	assert.Equal(t, "$1", lines[3])
	assert.Equal(t, "d", lines[4])
	//end
	clearSets(t, key1)

	clearSets(t, key2)

	clearSets(t, key3)

}
