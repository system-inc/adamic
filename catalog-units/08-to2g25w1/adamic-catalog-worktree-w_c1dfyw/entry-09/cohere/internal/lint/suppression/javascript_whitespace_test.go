package suppression

import "testing"

// TestDirectivesReadWhitespaceAsJavaScriptDoes holds the directive grammar to ESLint wherever Go's
// whitespace and JavaScript's differ: Go counts the next-line character (U+0085) as space and
// JavaScript does not, and JavaScript counts the byte order mark (U+FEFF) as space and Go does not
// (#z4nssqs).
//
// The first eight sources are @system_adamic's stream P2's, verbatim, from porting this index to
// Adamic (adamic stage1/cohere/suppression/GAPS.md and testdata/eslint_whitespace.cjs), where 6 of 8
// differed from ESLint 10.12.0. The other eleven reach the rest of the grammar's whitespace: what may
// follow the directive word, a listed rule's name, the reason separator, and an enable. Every answer
// below is ESLint 10.8.1's, the version this tree gates with, run through its Linter API with
// `no-debugger` on, and 10.8.1 agrees with Adamic's 10.12.0 table on the eight they share.
//
// The no-break space (U+00A0) is whitespace to both, so its cases pass either way, and they are here
// so a fix that narrowed the set to Go's minus one character would fail.
func TestDirectivesReadWhitespaceAsJavaScriptDoes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		// source is a comment, then `debugger;` on the line debuggerLine.
		source       string
		debuggerLine int
		suppressed   bool
	}{
		{"plain", "/* eslint-disable no-debugger */\ndebugger;", 1, true},
		{"U+0085 before the word", "/*\u0085eslint-disable no-debugger */\ndebugger;", 1, false},
		{"U+FEFF before the word", "/*\ufeffeslint-disable no-debugger */\ndebugger;", 1, true},
		{"U+0085 after the rule", "/* eslint-disable no-debugger\u0085*/\ndebugger;", 1, false},
		{"U+FEFF after the rule", "/* eslint-disable no-debugger\ufeff*/\ndebugger;", 1, true},
		{"next-line, U+0085 before", "//\u0085eslint-disable-next-line no-debugger\ndebugger;", 1, false},
		{"next-line, U+FEFF before", "//\ufeffeslint-disable-next-line no-debugger\ndebugger;", 1, true},
		{"U+00A0 before the word", "/*\u00a0eslint-disable no-debugger */\ndebugger;", 1, true},
		{"U+00A0 after the word", "/* eslint-disable\u00a0no-debugger */\ndebugger;", 1, true},
		{"U+FEFF after the word", "/* eslint-disable\ufeffno-debugger */\ndebugger;", 1, true},
		{"U+0085 after the word", "/* eslint-disable\u0085no-debugger */\ndebugger;", 1, false},
		{"next-line, U+FEFF after the word", "// eslint-disable-next-line\ufeffno-debugger\ndebugger;", 1, true},
		{"line, U+00A0 after the word", "debugger; // eslint-disable-line\u00a0no-debugger", 0, true},
		{"U+FEFF before a listed rule", "/* eslint-disable no-console,\ufeffno-debugger */\ndebugger;", 1, true},
		{"U+0085 before a listed rule", "/* eslint-disable no-console,\u0085no-debugger */\ndebugger;", 1, false},
		{"U+FEFF around the reason separator", "/* eslint-disable no-debugger\ufeff--\ufeffwhy */\ndebugger;", 1, true},
		// An enable that ESLint reads closes the block, so the `debugger;` after it is reported.
		{"enable, U+00A0 after the word", "/* eslint-disable no-debugger */\n/* eslint-enable\u00a0no-debugger */\ndebugger;", 2, false},
		{"enable, U+FEFF after the word", "/* eslint-disable no-debugger */\n/* eslint-enable\ufeffno-debugger */\ndebugger;", 2, false},
		{"enable, U+0085 after the word", "/* eslint-disable no-debugger */\n/* eslint-enable\u0085no-debugger */\ndebugger;", 2, true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			index := Build(testCase.source)
			got := index.Suppresses("no-debugger", offsetOfLine(testCase.source, testCase.debuggerLine))
			if got != testCase.suppressed {
				t.Errorf("suppressed = %v, ESLint says %v", got, testCase.suppressed)
			}
		})
	}
}
