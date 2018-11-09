package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func matchPrefixCase() map[string]string {
	cs := map[string]string{
		"abc?[a-z]":  "abc",
		"?abc":       "",
		"\\*[^abc]?": "*",
	}
	return cs
}

type patternMap map[string]bool

func matchCase(nocase bool) map[string]*patternMap {
	var cs map[string]*patternMap
	if !nocase {
		cs = map[string]*patternMap{
			"*": &patternMap{
				"":     true,
				"abcd": true,
				"*[*]": true,
			},
			"******a": &patternMap{
				"a":     true,
				"***a":  true,
				"bcdea": true,
				"abcd":  false,
			},
			"\\*?aaa": &patternMap{
				"*caaa": true,
				"abc":   false,
			},
			"[a-z][^0-9][z-a]?[a-z": &patternMap{
				"abz.a": true,
				"a1z.*": false,
				"abz.e": true,
			},
			"[a-z]*cat*[h][^b]*eyes*": &patternMap{
				"my cat has very bright eyes": true,
				"my dog has very bright eyes": false,
			},
			"h?llo": &patternMap{
				"hello": true,
				"healo": false,
			},
			"h??lo": &patternMap{
				"hello": true,
			},
			"h*o": &patternMap{
				"hello": true,
				"ho":    true,
			},
		}

	} else {
		cs = map[string]*patternMap{
			"[A-Z][0-9]*": &patternMap{
				"B1":    true,
				"B2000": true,
				"b2000": false,
			},
			"*A": &patternMap{
				"abcdA": true,
				"abcda": false,
				"Ae":    false,
			},
			"?A*C": &patternMap{
				"1AbcdC":   true,
				"cA12344C": true,
				"1abcdc":   false,
			},
		}
	}

	return cs
}

func TestGlobMatchPrefix(t *testing.T) {
	list := matchPrefixCase()
	for match, exptected := range list {
		val := globMatchPrefix([]byte(match))
		assert.Equal(t, exptected, string(val))
	}
}

func TestPatternMatch(t *testing.T) {
	cs := matchCase(false)
	for pattern, vals := range cs {
		for val, expected := range map[string]bool(*vals) {
			actual := globMatch([]byte(pattern), []byte(val), false)
			assert.Equal(t, expected, actual, "err log:", pattern, val)
		}
	}

	// check upper string
	cs = matchCase(true)
	for pattern, vals := range cs {
		for val, expected := range map[string]bool(*vals) {
			actual := globMatch([]byte(pattern), []byte(val), true)
			assert.Equal(t, expected, actual, "err log:", pattern, val)
		}
	}
}

func BenchmarkGlobMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		globMatch([]byte("hellabcdlo"), []byte("h*lo"), false)
	}
}
