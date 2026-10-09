package regexpattern

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
)

// collect walks a pattern and returns every character it reports.
func collect(t *testing.T, pattern string, flags string) []Character {
	t.Helper()
	var characters []Character
	if !Walk(pattern, regexsyntax.ParseRegexFlags(flags), func(character Character) bool {
		characters = append(characters, character)
		return true
	}) {
		t.Fatalf("walk refused %q with flags %q", pattern, flags)
	}
	return characters
}

// The kind is the whole reason this package reports more than values. `\0` and `\u0000` are one
// code point written two ways, and the rule this exists for reports the second and not the first.
// A walk returning values alone cannot express that difference, so it is asserted first.
func TestWalkReportsHowACharacterWasSpelled(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		pattern   string
		flags     string
		wantValue uint32
		wantKind  CharacterKind
	}{
		{"a plain character", "a", "", 'a', KindSymbol},
		{"a raw space", " ", "", ' ', KindSymbol},
		{"a null escape", `\0`, "", 0, KindNull},
		{"a hexadecimal escape", `\x00`, "", 0, KindHexadecimalEscape},
		{"a hexadecimal escape with a value", `\x1f`, "", 0x1F, KindHexadecimalEscape},
		{"a fixed-width unicode escape", `\u0000`, "", 0, KindUnicodeEscape},
		{"a braced unicode escape", `\u{1F}`, "u", 0x1F, KindUnicodeEscape},
		{"a named escape", `\t`, "", 0x09, KindSingleEscape},
		{"a newline escape", `\n`, "", 0x0A, KindSingleEscape},
		{"a control-letter escape", `\cA`, "", 1, KindControlLetter},
		{"a lowercase control-letter escape", `\ca`, "", 1, KindControlLetter},
		{"an octal escape", `\1`, "", 1, KindOctal},
		{"a multi-digit octal escape", `\101`, "", 65, KindOctal},
		{"an octal escape with a leading zero", `\01`, "", 1, KindOctal},
		{"an identity escape", `\-`, "", '-', KindIdentityEscape},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			characters := collect(t, testCase.pattern, testCase.flags)
			if len(characters) != 1 {
				t.Fatalf("want one character, got %d", len(characters))
			}
			if characters[0].Value != testCase.wantValue {
				t.Fatalf("want value %d, got %d", testCase.wantValue, characters[0].Value)
			}
			if characters[0].Kind != testCase.wantKind {
				t.Fatalf("want kind %d, got %d", testCase.wantKind, characters[0].Kind)
			}
		})
	}
}

// A quantifier contains the element it repeats, so the repeated character carries that depth.
//
// This was wrong in the first draft and a probe is what caught it. Reasoning said the character had
// already been emitted by the time the quantifier was seen, which is true of the scan order and
// false of the structure, and it produced a walk that reported every quantified character at depth
// zero. That is precisely the case the first caller exists to skip.
func TestWalkRaisesDepthForAQuantifiedCharacter(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		pattern   string
		wantDepth []int
	}{
		{"a star", "a*", []int{1}},
		{"a plus", "a+", []int{1}},
		{"an optional", "a?", []int{1}},
		{"a counted quantifier", "a{2}", []int{1}},
		{"an open-ended range", "a{2,}", []int{1}},
		{"a closed range", "a{2,4}", []int{1}},
		{"a lazy quantifier", "a+?", []int{1}},
		{"only the quantified character", "ab+", []int{0, 1}},
		{"two spaces where the second is quantified", "  +", []int{0, 1}},
		{"two spaces where the second is counted", "  {2}", []int{0, 1}},
		{"a brace that is not a quantifier", "a{x}", []int{0, 0, 0, 0}},
		{"nothing quantified", "ab", []int{0, 0}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			characters := collect(t, testCase.pattern, "")
			if len(characters) != len(testCase.wantDepth) {
				t.Fatalf("want %d characters, got %d", len(testCase.wantDepth), len(characters))
			}
			for index, want := range testCase.wantDepth {
				if characters[index].QuantifierDepth != want {
					t.Fatalf("character %d: want quantifier depth %d, got %d",
						index, want, characters[index].QuantifierDepth)
				}
			}
		})
	}
}

// A quantifier on a character class raises the depth for everything inside it.
//
// This exists because a mutation sweep found the class path unmeasured: raising quantifier depth
// only in the class branch left every other fixture green, since none of them put a quantifier on a
// class. A shared package with an unmeasured branch is worse than a duplicated helper, because the
// weakness is invisible from every rule that depends on it.
func TestWalkRaisesQuantifierDepthForAQuantifiedClass(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		pattern   string
		wantDepth []int
	}{
		{"a quantified class", "[ab]+", []int{1, 1}},
		{"a counted class", "[ab]{2}", []int{1, 1}},
		{"an unquantified class", "[ab]", []int{0, 0}},
		{"a quantified class beside a plain one", "[a][b]+", []int{0, 1}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			characters := collect(t, testCase.pattern, "")
			if len(characters) != len(testCase.wantDepth) {
				t.Fatalf("want %d characters, got %d", len(testCase.wantDepth), len(characters))
			}
			for index, want := range testCase.wantDepth {
				if characters[index].QuantifierDepth != want {
					t.Fatalf("character %d: want quantifier depth %d, got %d",
						index, want, characters[index].QuantifierDepth)
				}
			}
		})
	}
}

// The quantifier scan has three edges a sweep found unmeasured, and each is the kind that produces
// a plausible wrong answer rather than a crash.
func TestWalkConsumesQuantifiersExactly(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		pattern   string
		wantCount int
		wantDepth []int
	}{
		// A lazy quantifier's trailing question mark belongs to the quantifier. Left unconsumed it
		// is read as a second quantifier governing nothing, and the character after it inherits a
		// depth it should not have.
		{"a lazy star", "ab*?c", 3, []int{0, 1, 0}},
		{"a lazy counted quantifier", "ab{2}?c", 3, []int{0, 1, 0}},
		// A brace with no leading digits is a literal brace rather than a quantifier, so its
		// contents are matched characters.
		{"a brace holding no digits", "a{x}", 4, []int{0, 0, 0, 0}},
		{"a brace holding a trailing comma only", "a{,2}", 5, []int{0, 0, 0, 0, 0}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			characters := collect(t, testCase.pattern, "")
			if len(characters) != testCase.wantCount {
				t.Fatalf("want %d characters, got %d", testCase.wantCount, len(characters))
			}
			for index, want := range testCase.wantDepth {
				if characters[index].QuantifierDepth != want {
					t.Fatalf("character %d: want quantifier depth %d, got %d",
						index, want, characters[index].QuantifierDepth)
				}
			}
		})
	}
}

// An octal escape stops at three digits or at 255, whichever comes first, so the reported span
// equals the text the escape actually covers. Without the bound, a pattern reports one character
// where the language sees an escape plus a literal digit.
func TestWalkBoundsAnOctalEscape(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		pattern   string
		wantCount int
		wantFirst uint32
	}{
		{"three digits", `A`, 1, 65},
		{"two digits", `!`, 1, 33},
		{"a value above 255 stops early", `\400`, 2, 32},
		{"a fourth digit is literal", `A1`, 2, 65},
		{"a non-octal digit ends it", `8`, 2, 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			characters := collect(t, testCase.pattern, "")
			if len(characters) != testCase.wantCount {
				t.Fatalf("want %d characters, got %d", testCase.wantCount, len(characters))
			}
			if characters[0].Value != testCase.wantFirst {
				t.Fatalf("want first value %d, got %d", testCase.wantFirst, characters[0].Value)
			}
		})
	}
}

// Class depth is the other half of the structure the first caller reads, and it is separate from
// quantifier depth because a caller might one day want one without the other.
func TestWalkRaisesDepthInsideACharacterClass(t *testing.T) {
	t.Parallel()
	characters := collect(t, " [  ] ", "")
	if len(characters) != 4 {
		t.Fatalf("want four characters, got %d", len(characters))
	}
	wantClassDepth := []int{0, 1, 1, 0}
	for index, want := range wantClassDepth {
		if characters[index].ClassDepth != want {
			t.Fatalf("character %d: want class depth %d, got %d", index, want, characters[index].ClassDepth)
		}
	}
}

// Group syntax is structure rather than content, and a walk reporting matched characters must step
// over it. A probe caught this reporting the question mark and colon of a non-capturing group as
// things the pattern matches, which would have made every such group look like literal text.
func TestWalkSkipsGroupSyntax(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		pattern string
		want    string
	}{
		{"a capturing group", "(a)", "a"},
		{"a non-capturing group", "(?:a)", "a"},
		{"a named group", "(?<name>a)", "a"},
		{"a lookahead", "(?=a)", "a"},
		{"a negative lookahead", "(?!a)", "a"},
		{"a lookbehind", "(?<=a)", "a"},
		{"a negative lookbehind", "(?<!a)", "a"},
		{"several groups in a row", "(?:a)(?<name>b)(?=c)", "abc"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var got string
			for _, character := range collect(t, testCase.pattern, "") {
				got += string(rune(character.Value))
			}
			if got != testCase.want {
				t.Fatalf("want the walk to report %q, got %q", testCase.want, got)
			}
		})
	}
}

// Anchors, alternation and set escapes match positions or sets rather than single characters, so
// none of them has a value to report. Inventing one would put characters in the stream that the
// pattern never matches.
func TestWalkReportsNothingForNonCharacters(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		pattern string
	}{
		{"a start anchor", "^"},
		{"an end anchor", "$"},
		{"alternation", "|"},
		{"a digit class escape", `\d`},
		{"a word class escape", `\w`},
		{"a whitespace class escape", `\s`},
		{"a word boundary", `\b`},
		{"a negated class escape", `\D`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if characters := collect(t, testCase.pattern, ""); len(characters) != 0 {
				t.Fatalf("want no characters, got %d", len(characters))
			}
		})
	}
}

// The count has to be the total before any escape is judged, because a group may be referenced
// before it is defined. A running count calls the same escape a control character early in a
// pattern and a backreference later in it.
func TestCountCapturingGroups(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		pattern string
		want    int
	}{
		{"none", "abc", 0},
		{"one", "(a)", 1},
		{"two", "(a)(b)", 2},
		{"nested", "((a))", 2},
		{"non-capturing does not count", "(?:a)", 0},
		{"lookahead does not count", "(?=a)", 0},
		{"negative lookahead does not count", "(?!a)", 0},
		{"lookbehind does not count", "(?<=a)", 0},
		{"negative lookbehind does not count", "(?<!a)", 0},
		{"a named group counts", "(?<name>a)", 1},
		{"a paren inside a class does not count", "[(]", 0},
		{"an escaped paren does not count", `\(a\)`, 0},
		{"a mixture", "(a)(?:b)(?<c>d)(?=e)", 2},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got := CountCapturingGroups(testCase.pattern, regexsyntax.ParseRegexFlags(""))
			if got != testCase.want {
				t.Fatalf("want %d groups, got %d", testCase.want, got)
			}
		})
	}
}

// Spans are pattern-relative byte offsets, which is what lets one walk serve a regex literal and a
// string handed to the RegExp constructor. A caller adds its own offset; a walk reporting
// file-relative positions could not be reused.
func TestWalkReportsPatternRelativeSpans(t *testing.T) {
	t.Parallel()
	characters := collect(t, `a\x1fb`, "")
	if len(characters) != 3 {
		t.Fatalf("want three characters, got %d", len(characters))
	}
	wantSpans := [][2]int{{0, 1}, {1, 5}, {5, 6}}
	for index, want := range wantSpans {
		if characters[index].Start != want[0] || characters[index].End != want[1] {
			t.Fatalf("character %d: want span %d:%d, got %d:%d",
				index, want[0], want[1], characters[index].Start, characters[index].End)
		}
	}
}

// The callback returning false stops the walk, so a caller that has decided can avoid scanning the
// rest of a long pattern.
func TestWalkStopsWhenTheCallbackDeclines(t *testing.T) {
	t.Parallel()
	seen := 0
	completed := Walk("abcdef", regexsyntax.ParseRegexFlags(""), func(character Character) bool {
		seen++
		return seen < 2
	})
	if seen != 2 {
		t.Fatalf("want the walk to stop after two characters, saw %d", seen)
	}
	if completed {
		t.Fatal("want a stopped walk to report that it did not complete")
	}
}

// A pattern the scanner cannot finish stops the walk and says so, matching regexsyntax. An
// unterminated construct is a syntax error the parser has already refused, so a second opinion here
// would be noise on a file that does not compile.
func TestWalkRefusesMalformedInput(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		pattern string
	}{
		{"an unterminated class", "[abc"},
		{"a trailing backslash", `abc\`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if Walk(testCase.pattern, regexsyntax.ParseRegexFlags(""), func(Character) bool { return true }) {
				t.Fatalf("want the walk to refuse %q", testCase.pattern)
			}
		})
	}
}
