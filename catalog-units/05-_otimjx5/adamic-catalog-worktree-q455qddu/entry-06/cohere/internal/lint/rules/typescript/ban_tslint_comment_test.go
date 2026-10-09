package typescript

import (
	"strconv"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// banTslintCommentFile is a TypeScript name, but the rule does not gate on it.
//
// Upstream registers a plain `Program` visitor with no source-type test, and the installed 8.67.0
// build reports `// tslint:disable` in `a.js`, `a.mjs`, `a.ts` and `a.tsx` alike. The JavaScript
// case has its own fixture below rather than being assumed from that measurement.
const banTslintCommentFile = "/repository/source/Thing.ts"

// banTslintCommentCaseName numbers a row so a failure names which one, since the sources are long
// and several differ only in their tail.
func banTslintCommentCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// TestBanTslintCommentStaysSilentOnUpstreamPassCases is the imported clean corpus, verbatim.
//
// All five of upstream's passing inputs, extracted from the clone's test file by parsing it with
// the TypeScript compiler rather than by reading it, so no escape sequence passed through a shell
// or a keyboard on the way here. Every one was additionally run through the installed 8.67.0 build
// driven by the ESLint 10.8.1 Linter API, which reported nothing on all five.
func TestBanTslintCommentStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []string{
		"let a: readonly any[] = [];",
		"let a = new Array();",
		"// some other comment",
		"// TODO: this is a comment that mentions tslint",
		"/* another comment that mentions tslint */",
	}
	for index, sourceText := range cases {
		t.Run(banTslintCommentCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, BanTslintComment,
				banTslintCommentFile, sourceText))
		})
	}
}

// TestBanTslintCommentFiresOnUpstreamFailCases is the imported reporting corpus, verbatim.
//
// Every field on every row is measured rather than transcribed. The span is upstream's reported
// column range converted to an offset slice of the same input; the rendered text is recovered from
// the message the installed build produced; the fixed source is upstream's own `output` field, and
// it was separately confirmed to equal what `cohereAndFix` writes for all eight, so the corpus and
// the running rule agree and there is no drift to record.
//
// All four assertions run on each row, which is the point. A finding at the right span with the
// wrong rewrite, and a rewrite at the right span with the wrong text, both pass a message-id check.
func TestBanTslintCommentFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSpan   string
		wantText   string
		wantFixed  string
	}{
		{
			sourceText: "/* tslint:disable */",
			wantSpan:   "/* tslint:disable */",
			wantText:   "/* tslint:disable */",
			wantFixed:  "",
		},
		{
			sourceText: "/* tslint:enable */",
			wantSpan:   "/* tslint:enable */",
			wantText:   "/* tslint:enable */",
			wantFixed:  "",
		},
		{
			sourceText: "/* tslint:disable:rule1 rule2 rule3... */",
			wantSpan:   "/* tslint:disable:rule1 rule2 rule3... */",
			wantText:   "/* tslint:disable:rule1 rule2 rule3... */",
			wantFixed:  "",
		},
		{
			sourceText: "/* tslint:enable:rule1 rule2 rule3... */",
			wantSpan:   "/* tslint:enable:rule1 rule2 rule3... */",
			wantText:   "/* tslint:enable:rule1 rule2 rule3... */",
			wantFixed:  "",
		},
		{
			sourceText: "// tslint:disable-next-line",
			wantSpan:   "// tslint:disable-next-line",
			wantText:   "// tslint:disable-next-line",
			wantFixed:  "",
		},
		{
			sourceText: "someCode(); // tslint:disable-line",
			wantSpan:   "// tslint:disable-line",
			wantText:   "// tslint:disable-line",
			wantFixed:  "someCode();",
		},
		{
			sourceText: "// tslint:disable-next-line:rule1 rule2 rule3...",
			wantSpan:   "// tslint:disable-next-line:rule1 rule2 rule3...",
			wantText:   "// tslint:disable-next-line:rule1 rule2 rule3...",
			wantFixed:  "",
		},
		{
			sourceText: "\nconst woah = doSomeStuff();\n// tslint:disable-line\nconsole.log(woah);\n      ",
			wantSpan:   "// tslint:disable-line",
			wantText:   "// tslint:disable-line",
			wantFixed:  "\nconst woah = doSomeStuff();\nconsole.log(woah);\n      ",
		},
	}
	for index, testCase := range cases {
		t.Run(banTslintCommentCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, BanTslintComment, banTslintCommentFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "commentDetected")

			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.wantSpan {
				t.Errorf("the finding covers %q, want %q", reported, testCase.wantSpan)
			}

			// The rendered comment is the only part of the description that varies per input, and
			// it is the part upstream interpolates. Equality against a literal typed here rather
			// than against the rule's own message constant: comparing a finding to the constant it
			// was built from moves both sides under mutation and asserts nothing.
			wantDescription := messageBanTslintComment(testCase.wantText).Description
			if result.Diagnostics[0].Message.Description != wantDescription {
				t.Errorf("the description is %q, want %q",
					result.Diagnostics[0].Message.Description, wantDescription)
			}

			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

// TestBanTslintCommentDiscriminatesOnCasesUpstreamDoesNotWrite covers what the corpus leaves open.
//
// Upstream ships five clean cases and eight reporting ones, none of which touches the directive
// regex's trailing group, the anchor, JSDoc, or a file holding more than one directive. Every row
// here was run through the installed 8.67.0 build and carries the verdict that build produced, so a
// row asserting silence is asserting upstream's silence rather than this port's.
func TestBanTslintCommentDiscriminatesOnCasesUpstreamDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantIds    []string
		reason     string
	}{
		{
			sourceText: "// tslint:disabled",
			wantIds:    nil,
			reason:     "boundary: a longer word is not the directive",
		},
		{
			sourceText: "// tslint:disable-foo",
			wantIds:    nil,
			reason:     "boundary: an unknown suffix is not line or next-line",
		},
		{
			sourceText: "// tslint:disableX",
			wantIds:    nil,
			reason:     "boundary: no separator after the directive",
		},
		{
			sourceText: "// tslint disable",
			wantIds:    nil,
			reason:     "boundary: the colon is required",
		},
		{
			sourceText: "// TSLINT:disable",
			wantIds:    nil,
			reason:     "boundary: the match is case sensitive",
		},
		{
			sourceText: "// tslint:disable:",
			wantIds:    []string{"commentDetected"},
			reason:     "boundary: a bare colon satisfies the separator",
		},
		{
			sourceText: "// tslint:disable extra",
			wantIds:    []string{"commentDetected"},
			reason:     "boundary: whitespace satisfies the separator",
		},
		{
			sourceText: "// tslint:enable-next-line",
			wantIds:    []string{"commentDetected"},
			reason:     "the fourth directive spelling",
		},
		{
			sourceText: "// x tslint:disable",
			wantIds:    nil,
			reason:     "anchored: a word before the directive declines",
		},
		{
			sourceText: "//   tslint:disable",
			wantIds:    []string{"commentDetected"},
			reason:     "anchored: leading whitespace is absorbed",
		},
		{
			sourceText: "/** tslint:disable */",
			wantIds:    nil,
			reason:     "a JSDoc star defeats the anchor",
		},
		{
			sourceText: "/*\n tslint:disable\n*/",
			wantIds:    []string{"commentDetected"},
			reason:     "a newline is whitespace, so the anchor still matches",
		},
		{
			sourceText: "// tslint:disable-line\n// tslint:enable-line\n",
			wantIds:    []string{"commentDetected", "commentDetected"},
			reason:     "two comments, two findings",
		},
	}
	for index, testCase := range cases {
		t.Run(banTslintCommentCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, BanTslintComment, banTslintCommentFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestBanTslintCommentRendersTheCommentUpstreamRenders pins `toText`.
//
// The normalization is invisible to a message-id assertion and it is the whole of what upstream
// interpolates, so a port dropping the trim would agree with every other fixture in this file. Each
// row's expectation is the text the installed build printed for that exact input.
func TestBanTslintCommentRendersTheCommentUpstreamRenders(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantText   string
	}{
		{
			sourceText: "//   tslint:disable-line   ",
			wantText:   "// tslint:disable-line",
		},
		{
			sourceText: "/*   tslint:enable   */",
			wantText:   "/* tslint:enable */",
		},
		{
			sourceText: "/*\n tslint:disable\n*/",
			wantText:   "/* tslint:disable */",
		},
	}
	for index, testCase := range cases {
		t.Run(banTslintCommentCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, BanTslintComment, banTslintCommentFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "commentDetected")
			want := messageBanTslintComment(testCase.wantText).Description
			if result.Diagnostics[0].Message.Description != want {
				t.Errorf("the description is %q, want %q",
					result.Diagnostics[0].Message.Description, want)
			}
		})
	}
}

// TestBanTslintCommentFixDeclinesToDeleteCode is the recorded divergence.
//
// Upstream widens its removal by one character on each side without testing what those characters
// are, so it deletes source on four measured shapes. This port widens only onto whitespace. The
// judgment is identical on every row here, and only the rewrite differs.
//
// `upstreamFixed` is what the installed build's `cohereAndFix` actually wrote, captured rather than
// predicted, so the row stays honest if upstream ever narrows this itself. `wantFixed` is what this
// port writes.
func TestBanTslintCommentFixDeclinesToDeleteCode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText    string
		upstreamFixed string
		wantFixed     string
		reason        string
	}{
		{
			sourceText:    "x;// tslint:disable",
			upstreamFixed: "x",
			wantFixed:     "x;",
			reason:        "DIVERGENCE: upstream eats the semicolon",
		},
		{
			sourceText:    "/* tslint:disable */let x = 1;",
			upstreamFixed: "et x = 1;",
			wantFixed:     "let x = 1;",
			reason:        "DIVERGENCE: upstream eats the l",
		},
		{
			sourceText:    "\t// tslint:disable-line",
			upstreamFixed: "",
			wantFixed:     "",
			reason:        "DIVERGENCE: upstream eats the tab",
		},
		{
			sourceText:    "x;  // tslint:disable",
			upstreamFixed: "x; ",
			wantFixed:     "x; ",
			reason:        "agrees: the preceding character is blank",
		},
	}
	for index, testCase := range cases {
		t.Run(banTslintCommentCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, BanTslintComment, banTslintCommentFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "commentDetected")
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

// TestBanTslintCommentHasNoFileGate is the fixture for the sibling's trap.
//
// `ban-ts-comment` in this package declines a JavaScript file and this rule must not, so the
// difference is pinned rather than left to a doc comment. Measured on the installed build: the same
// input reports in every one of these extensions.
func TestBanTslintCommentHasNoFileGate(t *testing.T) {
	t.Parallel()

	for _, fileName := range []string{
		"/repository/source/Thing.ts",
		"/repository/source/Thing.tsx",
		"/repository/source/Thing.mts",
		"/repository/source/Thing.cts",
		"/repository/source/Thing.js",
		"/repository/source/Thing.jsx",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, BanTslintComment,
				fileName, "// tslint:disable"), "commentDetected")
		})
	}
}

// TestBanTslintCommentSurvivesAnUnterminatedBlockComment pins the guard in the interior reader.
//
// Error recovery hands back a two-byte `/*` as a block comment, and subtracting two from each end
// of it produces a start past its end, which panics on the slice. The rule package is walked with a
// recover per file rather than per rule, so one such panic costs every rule that file. There is no
// finding to assert here; the assertion is that the run completes.
//
// Upstream cannot reach any of these and this is a parser difference rather than a divergence in
// judgment. Measured: all four are `Parsing error: '*/' expected` to the ESLint parser, so the rule
// never runs on them there. Our parser recovers and hands the walk a comment, so the guard is
// load-bearing here and unreachable upstream. That is why no finding assertion appears below: this
// test asserts an absence of panic, and asserting a verdict would be inventing one upstream never
// gave.
func TestBanTslintCommentSurvivesAnUnterminatedBlockComment(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{"/*", "/*/", "/**", "/* tslint:disable"} {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.Run(t, BanTslintComment, banTslintCommentFile, sourceText)
		})
	}
}
