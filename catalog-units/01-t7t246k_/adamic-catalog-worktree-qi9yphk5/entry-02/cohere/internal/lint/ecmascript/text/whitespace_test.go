package text

import (
	"slices"
	"testing"
	"unicode"
)

// javaScriptWhitespace is every code point JavaScript calls whitespace, as Node 24.14.1 (Unicode 17)
// printed it, scanning U+0000 to U+10FFFF for both `/\s/u.test` and `trim`, which agreed:
//
//	for (let c = 0; c <= 0x10FFFF; c++) { if (c >= 0xD800 && c <= 0xDFFF) continue;
//	  const ch = String.fromCodePoint(c); if (/\s/u.test(ch)) print(c); }
var javaScriptWhitespace = []rune{
	0x0009, 0x000A, 0x000B, 0x000C, 0x000D, 0x0020, 0x00A0, 0x1680, 0x2000, 0x2001, 0x2002, 0x2003,
	0x2004, 0x2005, 0x2006, 0x2007, 0x2008, 0x2009, 0x200A, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000,
	0xFEFF,
}

func TestIsWhitespaceMatchesJavaScript(t *testing.T) {
	t.Parallel()

	// Every code point, so a character added or dropped anywhere fails by name.
	matched := 0
	for character := rune(0); character <= unicode.MaxRune; character++ {
		want := slices.Contains(javaScriptWhitespace, character)
		if IsWhitespace(character) != want {
			t.Errorf("IsWhitespace(U+%04X) = %v, want %v", character, !want, want)
		}
		if want {
			matched++
		}
	}
	if matched != 25 {
		t.Fatalf("the scan matched %d code points, want JavaScript's 25", matched)
	}
}

// TestGoSpaceDiffersFromJavaScriptOnTwoCharacters records why the helper exists: Go's set and
// JavaScript's differ on exactly the next-line character and the byte order mark. If Go's set ever
// changes, this says so rather than leaving the helper's doc comment quietly wrong.
func TestGoSpaceDiffersFromJavaScriptOnTwoCharacters(t *testing.T) {
	t.Parallel()

	var differences []rune
	for character := rune(0); character <= unicode.MaxRune; character++ {
		if unicode.IsSpace(character) != IsWhitespace(character) {
			differences = append(differences, character)
		}
	}
	if !slices.Equal(differences, []rune{0x0085, 0xFEFF}) {
		t.Fatalf("Go and JavaScript whitespace differ on %U, want U+0085 and U+FEFF", differences)
	}
}

func TestTrimWhitespace(t *testing.T) {
	t.Parallel()

	cases := map[string]struct{ value, want string }{
		"spaces and tabs":                                              {" \tvalue\t ", "value"},
		"the byte order mark, which Go keeps":                          {"\ufeffvalue\ufeff", "value"},
		"the next-line character, which Go trims and JavaScript keeps": {"\u0085value\u0085", "\u0085value\u0085"},
		"a no-break space":                                             {"\u00a0value\u00a0", "value"},
		"line terminators":                                             {"\u2028\r\nvalue\u2029", "value"},
		"interior whitespace":                                          {" a \ufeff b ", "a \ufeff b"},
		"nothing but space":                                            {" \ufeff\u3000 ", ""},
		"empty":                                                        {"", ""},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := TrimWhitespace(testCase.value); got != testCase.want {
				t.Errorf("TrimWhitespace(%q) = %q, want %q", testCase.value, got, testCase.want)
			}
		})
	}
}

func TestWhitespaceFields(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		value string
		want  []string
	}{
		"runs of space":                        {"  a  b\tc ", []string{"a", "b", "c"}},
		"split on the byte order mark":         {"a\ufeffb", []string{"a", "b"}},
		"not split on the next-line character": {"a\u0085b", []string{"a\u0085b"}},
		"nothing but space":                    {" \u00a0 ", []string{}},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := WhitespaceFields(testCase.value); !slices.Equal(got, testCase.want) {
				t.Errorf("WhitespaceFields(%q) = %q, want %q", testCase.value, got, testCase.want)
			}
		})
	}
}
