package regexsyntax

import "testing"

// The tests here exist because the ported fixture set does not reach three paths this package
// exports, and a suite that never touches a path cannot report when it breaks.
//
// Found by mutation rather than by reading: disabling the `u` flag, disabling the `v` flag, and
// making IsHexDigit always answer false each left the ported suite entirely green. Thirty-one
// passing tests, three surviving mutants. The ported file walks nesting thoroughly and calls
// ParseRegexFlags exactly never.

// TestParseRegexFlagsReadsBothModeFlags pins that `u` and `v` are read independently.
//
// They are separate fields rather than one mode because a pattern can be neither, and the class
// parser branches on each: `v` enables nested classes and set operations, `u` changes what an
// escape means. A parser that collapsed them would accept `v` syntax under `u`.
func TestParseRegexFlagsReadsBothModeFlags(t *testing.T) {
	t.Parallel()
	cases := []struct {
		flags               string
		unicode, unicodeSet bool
	}{
		{"", false, false},
		{"g", false, false},
		{"u", true, false},
		{"v", false, true},
		{"gu", true, false},
		{"gv", false, true},
		{"gimsuy", true, false},
	}
	for _, testCase := range cases {
		got := ParseRegexFlags(testCase.flags)
		if got.Unicode != testCase.unicode || got.UnicodeSets != testCase.unicodeSet {
			t.Errorf("ParseRegexFlags(%q) = {u:%v v:%v}, want {u:%v v:%v}",
				testCase.flags, got.Unicode, got.UnicodeSets, testCase.unicode, testCase.unicodeSet)
		}
		// UV is what every caller branches on, so it is asserted rather than assumed to follow.
		if want := testCase.unicode || testCase.unicodeSet; got.UV() != want {
			t.Errorf("ParseRegexFlags(%q).UV() = %v, want %v", testCase.flags, got.UV(), want)
		}
	}
}

// TestHexHelpers pins the three hex functions the class parser reaches for when reading `\xNN`,
// `\uNNNN` and `\u{NNNNN}`.
//
// The empty case is the one worth stating: AllHexDigits("") is false rather than vacuously true,
// because the callers are asking whether an escape has digits after it, and `\x` with nothing
// following is not a valid escape.
func TestHexHelpers(t *testing.T) {
	t.Parallel()
	for _, b := range []byte{'0', '9', 'a', 'f', 'A', 'F'} {
		if !IsHexDigit(b) {
			t.Errorf("IsHexDigit(%q) = false, want true", b)
		}
	}
	for _, b := range []byte{'g', 'G', 'z', ' ', '-', 0} {
		if IsHexDigit(b) {
			t.Errorf("IsHexDigit(%q) = true, want false", b)
		}
	}

	for _, s := range []string{"0", "ff", "FF", "1a2B3c"} {
		if !AllHexDigits(s) {
			t.Errorf("AllHexDigits(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"", "0g", "xyz", "1 2"} {
		if AllHexDigits(s) {
			t.Errorf("AllHexDigits(%q) = true, want false", s)
		}
	}

	for _, testCase := range []struct {
		text string
		want uint32
	}{
		{"0", 0}, {"f", 15}, {"F", 15}, {"ff", 255}, {"1F600", 0x1F600}, {"0000", 0},
	} {
		if got := ParseHexUint(testCase.text); got != testCase.want {
			t.Errorf("ParseHexUint(%q) = %d, want %d", testCase.text, got, testCase.want)
		}
	}
}

// TestPatternAndFlags pins the split on the last slash rather than the first, which is what makes
// an escaped slash inside a pattern come apart correctly.
func TestPatternAndFlags(t *testing.T) {
	t.Parallel()
	cases := []struct{ text, pattern, flags string }{
		{`/a/`, "a", ""},
		{`/a/g`, "a", "g"},
		{`/[a-z]+/gi`, "[a-z]+", "gi"},
		{`/a\/b/g`, `a\/b`, "g"},
		{`//`, "", ""},
		{"", "", ""},
		{"a", "", ""},
	}
	for _, testCase := range cases {
		pattern, flags := PatternAndFlags(testCase.text)
		if pattern != testCase.pattern || flags != testCase.flags {
			t.Errorf("PatternAndFlags(%q) = (%q, %q), want (%q, %q)",
				testCase.text, pattern, flags, testCase.pattern, testCase.flags)
		}
	}
}
