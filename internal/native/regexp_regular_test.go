package native

import (
	"math/rand"
	"testing"
	"unicode/utf16"
)

// Cache exhaustion is a fallback, not a failed match. This regular language has
// exponentially many distinguishable suffixes, beyond the bounded lazy cache.
func TestRegExpRegularEngineNode(t *testing.T) {
	random := rand.New(rand.NewSource(875))
	input := make([]rune, 512)
	for k := range input {
		input[k] = rune('a' + random.Intn(2))
	}
	input[len(input)-9] = 'a'
	cases := []regexCase{
		{Pattern: `(?:a|b)*a(?:a|b){8}$`, Input: utf16.Encode(input)},
		{Pattern: `^(a)\1$`, Input: []uint16{'a'}},
		{Pattern: `^(a)\1$`, Input: []uint16{'a', 'a'}},
		{Pattern: `a.*z|b`, Input: []uint16{'a', 'b', 'z', 'z'}},
		{Pattern: `a.*z|b`, Input: []uint16{'a', 'b', 'x', 'x'}},
		{Pattern: `(a|aa)(a?)`, Input: []uint16{'a', 'a'}},
		{Pattern: `(a|(b))+`, Input: []uint16{'a', 'b', 'a'}},
		{Pattern: `a.*?b`, Input: []uint16{'a', 'x', 'b', 'b'}},
		{Pattern: `a.*b`, Input: []uint16{'a', 'a', 'x', 'b'}},
		{Pattern: `a.*b`, Input: []uint16{'a', 'x', 'b', 'b'}},
	}
	runRegexNodeCases(t, cases)
}
