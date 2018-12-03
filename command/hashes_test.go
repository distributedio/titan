package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func initHashes(t *testing.T, key string, n int) {
	args := []string{key}
	for ; n > 0; n-- {
		args = append(args, "field", "value")

	}
	ctx := ContextTest("hmset", args...)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])
}

func clearHashes(t *testing.T, key string) {
	ctx := ContextTest("del", key)
	Call(ctx)
	lines := ctxLines(ctx.Out)
	assert.Equal(t, ":1", lines[0])
}

func setHashes(t *testing.T, key, val string) []string {
	ctx := ContextTest("rpush", key, val)
	Call(ctx)
	return ctxLines(ctx.Out)
}

func TestHLen(t *testing.T) {
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
