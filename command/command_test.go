package command

import (
	"testing"
)

func TestGlobMatch(t *testing.T) {
	if !globMatch([]byte("hello"), []byte("h?llo"), false) {
		t.Fatal()
	}
	if !globMatch([]byte("hello"), []byte("h??lo"), false) {
		t.Fatal()
	}
	if globMatch([]byte("healo"), []byte("h?llo"), false) {
		t.Fatal()
	}

	if !globMatch([]byte("hello"), []byte("h*o"), false) {
		t.Fatal()
	}
	if !globMatch([]byte("ho"), []byte("h*o"), false) {
		t.Fatal()
	}
	if !globMatch([]byte("lo"), []byte("*lo"), false) {
		t.Fatal()
	}
	if !globMatch([]byte("hellabcdlo"), []byte("h*lo"), false) {
		t.Fatal()
	}
	if globMatch([]byte("hellabcdlao"), []byte("h*lo"), false) {
		t.Fatal()
	}
	if !globMatch([]byte("hello"), []byte("**"), false) {
		t.Fatal()
	}
	if !globMatch([]byte("hello"), []byte("*?"), false) {
		t.Fatal()
	}

	if !globMatch([]byte("hello"), []byte("h[edf]llo"), false) {
		t.Fatal()
	}
	if !globMatch([]byte("hdllo"), []byte("h[edf]llo"), false) {
		t.Fatal()
	}
	if !globMatch([]byte("hfllo"), []byte("h[edf]llo"), false) {
		t.Fatal()
	}
	if globMatch([]byte("hallo"), []byte("h[edf]llo"), false) {
		t.Fatal()
	}

	if !globMatch([]byte("hallo"), []byte("h[^edf]llo"), false) {
		t.Fatal()
	}
	if globMatch([]byte("hello"), []byte("h[^edf]llo"), false) {
		t.Fatal()
	}

	if globMatch([]byte("hello"), []byte("h"), false) {
		t.Fatal()
	}
}
func BenchmarkGlobMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		globMatch([]byte("hellabcdlo"), []byte("h*lo"), false)
	}
}
