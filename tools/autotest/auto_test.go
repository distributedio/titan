package autotest

import (
	"testing"
)

func Test(t *testing.T) {
	at.SystemCase(t)
	at.StringCase(t)
	at.KeyCase(t)
	// at.ListCase(t)
	at.MultiCase(t)

	an.StringCase(t)
	// an.ListCase(t)
	an.KeyCase(t)
	an.MultiCase(t)
}
