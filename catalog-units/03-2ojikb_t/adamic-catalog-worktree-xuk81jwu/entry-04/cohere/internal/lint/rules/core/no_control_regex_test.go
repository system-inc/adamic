package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// controlRegexFile is where the fixtures pretend to live.
const controlRegexFile = "/repository/source/ControlRegex.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// That file carries FIVE Tester blocks, and four of them call `.test()` rather than
// `.test_and_snapshot()`. Reading only the last one, which is what a porter opening the file sees,
// imports 34 of 57 failing cases and silently drops the rest. The extractor counts the blocks first
// for exactly this reason.
//
// Only the regex-literal cases are taken here. Upstream's corpus is roughly half `new RegExp(...)`
// calls, and this port watches the literal form only; the constructor arm is a separate surface with
// its own discrimination and porting half of it silently would be worse than declining it.
func TestNoControlRegexFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a hex escape", "export const r = /\\x1f/;"},
		{"a hex escape among other characters", "export const r = /\\x1fFOO\\x00/;"},
		{"two in one pattern report once", "export const r = /FOO\\x1fFOO\\x1f/;"},
		{"inside a named group", "export const r = /(?<a>\\x1f)/;"},
		// The `u` flag is what makes this a code-point escape rather than a `\u` followed by
		// literal braces, which is the pass case directly below.
		{"a braced unicode escape under the u flag", "export const r = /\\u{1F}/u;"},
		{"the same with more flags", "export const r = /\\u{1F}/ugi;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoControlRegex, controlRegexFile, testCase.sourceText),
				"noControlRegex")
		})
	}
}

// The clean cases are the whole discrimination, and the flag pair is the sharpest.
//
// `/\u{1F}/` and `/\u{1F}/u` are the same six characters and only the second reports. Without the
// flag those braces are literal, so the pattern matches a `u` repeated once, and nothing in it is a
// control character at all. A rule reading the source text for `\u` would report both.
//
// The named escapes are the other half: `\t`, `\n`, `\r`, `\f`, `\v` are all control characters by
// value and none is reported, because the rule objects to spelling a code out by number rather than
// to the code point being present.
func TestNoControlRegexStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"text that merely looks like an escape", "export const r = /x1f/;"},
		{"a braced unicode escape without the u flag", "export const r = /\\u{1F}/;"},
		{"the same with an unrelated flag", "export const r = /\\u{1F}/g;"},
		{"a space escape under the u flag", "export const r = /\\u{20}/u;"},
		{"a tab", "export const r = /\\t/;"},
		{"a newline", "export const r = /\\n/g;"},
		{"the line-ending alternatives", "export const r = /\\r\\n|\\r|\\n/;"},
		{"a form feed", "export const r = /\\f/;"},
		{"a vertical tab", "export const r = /\\v/;"},
		{"a null written as itself", "export const r = /\\0\\t\\n\\r/;"},
		{"tabs in a quantified class", "export const r = /[\\n\\t]+/g;"},
		{"a leading tab class", "export const r = /^\\t+/;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoControlRegex, controlRegexFile, testCase.sourceText))
		})
	}
}

// A case written because somebody read our code rather than upstream's.
//
// Upstream batches every control character in one pattern into a single diagnostic, which is why
// its 57 failing inputs produce 34. Its corpus proves the batching only indirectly, through a
// snapshot count a porter has to go and read. This asserts it directly: three escapes, one finding.
func TestNoControlRegexReportsOncePerPattern(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoControlRegex, controlRegexFile, "export const r = /\\x01\\x02\\x03/;"),
		"noControlRegex")
}
