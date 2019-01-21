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
	key := "set-add"
	//initSets(t, key, 3)

	//case1
	ctx := ContextTest("sadd", key, "1", "2", "3")
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":3", lines[0])

	//case 2
	lines = setSets(t, key, "1", "2")
	assert.Equal(t, ":0", lines[0])
	/*
		//case 3
		lines = setSets(t, key, "3", "4")
		assert.Equal(t, ":1", lines[0])
	*/
	// end
	clearHashes(t, key)

}

/*
func TestSCard(t *testing.T) {
	// init
	key := "set-key"
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
	clearHashes(t, key)
}

/*
func TestSMembers(t *testing.T) {
	// init
	key := "set-key"
	initSets(t, key, 3)

	//case1
	ctx := ContextTest("smembers", key)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, "1", lines[0])
	assert.Equal(t, "2", lines[1])
	assert.Equal(t, "3", lines[2])
	//case2

	//case3

	//end

}*/

func TestIsmembers(t *testing.T) {

}
func TestSPop(t *testing.T) {

}
func TestSRem(t *testing.T) {

}
func TestSUion(t *testing.T) {

}
func TestSInter(t *testing.T) {

}
