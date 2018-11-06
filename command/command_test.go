package command

import (
	"testing"
)

func TestGlobMatch(t *testing.T) {
	if !globMatch("hello", "h?llo") {
		t.Fatal()
	}
	if !globMatch("hello", "h??lo") {
		t.Fatal()
	}
	if globMatch("healo", "h?llo") {
		t.Fatal()
	}

	if !globMatch("hello", "h*o") {
		t.Fatal()
	}
	if !globMatch("ho", "h*o") {
		t.Fatal()
	}
	if !globMatch("lo", "*lo") {
		t.Fatal()
	}
	if !globMatch("hellabcdlo", "h*lo") {
		t.Fatal()
	}
	if globMatch("hellabcdlao", "h*lo") {
		t.Fatal()
	}
	if !globMatch("hello", "**") {
		t.Fatal()
	}
	if !globMatch("hello", "*?") {
		t.Fatal()
	}

	if !globMatch("hello", "h[edf]llo") {
		t.Fatal()
	}
	if !globMatch("hdllo", "h[edf]llo") {
		t.Fatal()
	}
	if !globMatch("hfllo", "h[edf]llo") {
		t.Fatal()
	}
	if globMatch("hallo", "h[edf]llo") {
		t.Fatal()
	}

	if !globMatch("hallo", "h[^edf]llo") {
		t.Fatal()
	}
	if globMatch("hello", "h[^edf]llo") {
		t.Fatal()
	}

	if globMatch("hello", "h") {
		t.Fatal()
	}
}
func BenchmarkGlobMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		globMatch("hellabcdlo", "h*lo")
	}
}
