package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWatch_EmptyKeys(t *testing.T) {
	// This test ensures that invoking Watch with no keys does not panic.
	ctx := ContextTest("watch")
	ctx.Args = []string{} // Explicitly empty args

	assert.NotPanics(t, func() {
		Watch(ctx)
	})
}

func TestWatch_WithKeys(t *testing.T) {
	// This test ensures that invoking Watch with keys works correctly (using optimistic locking reads).
	ctx := ContextTest("watch", "key1")

	assert.NotPanics(t, func() {
		Watch(ctx)
	})
}