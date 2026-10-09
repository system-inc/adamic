package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// defaultCaseLastFile is where the fixtures pretend to live.
const defaultCaseLastFile = "/repository/source/DefaultCaseLast.ts"

// The corpus is upstream's, extracted from the tester rather than retyped.
//
// `/tmp/lint-sources/eslint/tests/lib/rules/default-case-last.js` carries 23 valid cases and 14
// invalid ones, every invalid case naming exactly one `notLast` and asserting the column of its
// `default` keyword. The cases below were rendered to Go literals by a script reading that file
// through a stub RuleTester, so no case was retyped and no escape could be cooked on the way.
func TestDefaultCaseLastStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		// The 23 clean cases, verbatim from upstream. They vary bodies, breaks and neighbour
		// counts while keeping the default clause last, which is how they pin that position is the
		// whole judgment.
		{"switch (foo) {}"},
		{"switch (foo) { case 1: bar(); break; }"},
		{"switch (foo) { case 1: break; }"},
		{"switch (foo) { case 1: }"},
		{"switch (foo) { case 1: bar(); break; case 2: baz(); break; }"},
		{"switch (foo) { case 1: break; case 2: break; }"},
		{"switch (foo) { case 1: case 2: break; }"},
		{"switch (foo) { case 1: case 2: }"},
		{"switch (foo) { default: bar(); break; }"},
		{"switch (foo) { default: bar(); }"},
		{"switch (foo) { default: break; }"},
		{"switch (foo) { default: }"},
		{"switch (foo) { case 1: break; default: break; }"},
		{"switch (foo) { case 1: break; default: }"},
		{"switch (foo) { case 1: default: break; }"},
		{"switch (foo) { case 1: default: }"},
		{"switch (foo) { case 1: baz(); break; case 2: quux(); break; default: quuux(); break; }"},
		{"switch (foo) { case 1: break; case 2: break; default: break; }"},
		{"switch (foo) { case 1: break; case 2: break; default: }"},
		{"switch (foo) { case 1: case 2: break; default: break; }"},
		{"switch (foo) { case 1: break; case 2: default: break; }"},
		{"switch (foo) { case 1: break; case 2: default: }"},
		{"switch (foo) { case 1: case 2: default: }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, DefaultCaseLast, defaultCaseLastFile, testCase.sourceText))
		})
	}
}

// Every reporting case names one `notLast` and no more, which is upstream's own count per input.
func TestDefaultCaseLastFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"switch (foo) { default: bar(); break; case 1: baz(); break; }"},  // errors=["notLast"] cols=[16]
		{"switch (foo) { default: break; case 1: break; }"},                // errors=["notLast"] cols=[16]
		{"switch (foo) { default: break; case 1: }"},                       // errors=["notLast"] cols=[16]
		{"switch (foo) { default: case 1: break; }"},                       // errors=["notLast"] cols=[16]
		{"switch (foo) { default: case 1: }"},                              // errors=["notLast"] cols=[16]
		{"switch (foo) { default: break; case 1: break; case 2: break; }"}, // errors=["notLast"] cols=[16]
		{"switch (foo) { default: case 1: break; case 2: break; }"},        // errors=["notLast"] cols=[16]
		{"switch (foo) { default: case 1: case 2: break; }"},               // errors=["notLast"] cols=[16]
		{"switch (foo) { default: case 1: case 2: }"},                      // errors=["notLast"] cols=[16]
		{"switch (foo) { case 1: break; default: break; case 2: break; }"}, // errors=["notLast"] cols=[31]
		{"switch (foo) { case 1: default: break; case 2: break; }"},        // errors=["notLast"] cols=[24]
		{"switch (foo) { case 1: break; default: case 2: break; }"},        // errors=["notLast"] cols=[31]
		{"switch (foo) { case 1: default: case 2: break; }"},               // errors=["notLast"] cols=[24]
		{"switch (foo) { case 1: default: case 2: }"},                      // errors=["notLast"] cols=[24]
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, DefaultCaseLast, defaultCaseLastFile, testCase.sourceText), "notLast")
		})
	}
}

// The span, which no message-id fixture above can see.
//
// Upstream asserts a column on all fourteen reporting cases and every one of them names the
// `default` keyword rather than the `switch`. A port reporting the whole statement passes both
// tests above and points a reader at the top of a switch that may be a hundred lines long. The
// three cases here are the three distinct columns the corpus asserts: 16 for a leading `default`,
// 24 for one after a bodyless `case 1:`, and 31 for one after `case 1: break;`.
//
// Asserted by slicing the source at the reported range rather than by comparing offsets, because an
// offset comparison is wrong in the same direction as the code that produced it. The prefix check
// rather than equality is because the reported node is the whole clause, whose text runs to the end
// of its body.
func TestDefaultCaseLastReportsTheDefaultClause(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantColumn int
	}{
		{"switch (foo) { default: break; case 1: break; }", 16},
		{"switch (foo) { case 1: default: case 2: break; }", 24},
		{"switch (foo) { case 1: break; default: case 2: break; }", 31},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, DefaultCaseLast, defaultCaseLastFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if !strings.HasPrefix(reported, "default") {
				t.Fatalf("reported %q, wanted a span starting at the default keyword", reported)
			}
			// Upstream's columns are one-based, so the offset it names is one less.
			if got := result.Diagnostics[0].Range.Pos() + 1; got != testCase.wantColumn {
				t.Fatalf("reported at column %d, upstream asserts %d", got, testCase.wantColumn)
			}
		})
	}
}

// A shape the corpus does not write, added from reading our own code.
//
// Two `default` clauses is a syntax error that the parser recovers from rather than refusing, so
// the shape reaches the rule. Upstream's `findIndex` takes the first and stops, reporting once; a
// port looping over every clause without returning reports twice. The imported corpus cannot see
// this because upstream never wrote it.
func TestDefaultCaseLastReportsOnceForTwoDefaults(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.Run(t, DefaultCaseLast, defaultCaseLastFile,
		"switch (foo) { default: break; default: break; case 1: break; }"), "notLast")
}
