package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// multiStrFile is where the fixtures pretend to live.
const multiStrFile = "/repository/source/MultiStr.ts"

// multiStrJsxFile is the same, for the cases that have to parse as JSX.
//
// The extension decides the script kind in the harness, so a JSX case named `.ts` parses as
// TypeScript and quietly contains no JSX node at all, which would make a JSX-gated assertion pass
// for the wrong reason.
const multiStrJsxFile = "/repository/source/MultiStr.tsx"

// The corpus is ESLint's own, copied rather than rewritten.
//
// Every case below is verbatim from `eslint/tests/lib/rules/no-multi-str.js`: 2 pass and 5 fail,
// each failing case naming exactly one `multilineString`. The cases were lifted by evaluating that
// test file with its `RuleTester` stubbed and serialising the spec object it was handed, so the
// escapes here are the bytes upstream runs rather than a retyping of them. That mattered for three
// of these five: they carry a real carriage return, a real U+2028, and a real U+2029 inside the
// source string, and any path that retyped them would have produced an escape sequence instead and
// tested nothing.
//
// The two clean cases are split out below, because one of them has to parse as JSX.
func TestNoMultiStrFires(t *testing.T) {
	t.Parallel()

	cases := []string{
		"var x = 'Line 1 \\\n Line 2'",
		"test('Line 1 \\\n Line 2');",
		"'foo\\\rbar';",
		"'foo\\\u2028bar';",
		"'foo\\\u2029ar';",
	}

	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoMultiStr, multiStrFile, sourceText), "multilineString")
		})
	}
}

// The two clean cases upstream ships, and they are doing different jobs.
//
// The first is a string with no line break at all, which is the ordinary case. The second is the
// JSX one, and it is weaker evidence than it looks: its newlines are in JSX text rather than in a
// string literal, so no string literal in it has a line break and it would stay clean even with the
// JSX exemption removed. That is why the exemption gets its own test below rather than resting on
// this case.
func TestNoMultiStrStaysSilent(t *testing.T) {
	t.Parallel()

	t.Run("a string on one line", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t,
			rule_testing.Run(t, NoMultiStr, multiStrFile, "var a = 'Line 1 Line 2';"))
	})

	t.Run("a JSX element whose text spans lines", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t,
			rule_testing.Run(t, NoMultiStr, multiStrJsxFile, "var a = <div>\n<h1>Wat</h1>\n</div>;"))
	})
}

// The cases below are ours. Each exists because reading our own code or driving the installed
// ESLint build raised a question the imported corpus does not answer.

// TestNoMultiStrJsxExemption is the case the imported corpus cannot make.
//
// Upstream declines when the literal's parent type begins `JSX`, and its only JSX fixture cannot
// see that branch at all, because the newlines in it are JSX text rather than a string literal. So
// the exemption is untested upstream and would ship broken either way.
//
// Enumerating every JSX position a string literal can occupy in our parser gives exactly two
// parents: `KindJsxAttribute` for a bare attribute value, and `KindJsxExpression` for a string
// inside braces, whether those braces are an attribute value or element content. Driving the
// installed ESLint build (version 10.8.1, through the Linter API) over the same four shapes shows
// both exempt.
//
// The spread case is the boundary and it is the reason this is not simply "anything inside JSX":
// a string in an object literal spread into an element has a `PropertyAssignment` parent, which is
// not a JSX kind, and it reports. Measured on the installed build.
//
// Without the exemption this rule would report every multiline attribute value in every component
// file in the tree, which is a false-positive class no imported fixture can see.
func TestNoMultiStrJsxExemption(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFires  bool
	}{
		{"a multiline attribute value", "var a = <div attr='Line 1 \\\n Line 2' />;", false},
		{"a multiline string in attribute braces", "var a = <div attr={'Line 1 \\\n Line 2'} />;", false},
		{"a multiline string in element braces", "var a = <div>{'Line 1 \\\n Line 2'}</div>;", false},
		{"a multiline string in a spread object", "var a = <div {...{a: 'Line 1 \\\n Line 2'}} />;", true},
		// A string spread DIRECTLY as attributes, whose parent is the spread attribute itself rather
		// than an object literal inside it. Upstream is silent; measured on the installed build.
		// This case was found by a surviving mutant: narrowing the kind enumeration to the two
		// parents that everyday well-formed JSX produces passed every other fixture here, and
		// enumerating parents over malformed and unusual source is what turned it up.
		{"a multiline string spread as attributes", "var a = <div {...'Line 1 \\\n Line 2'} />;", false},
		{"a multiline string outside JSX in a JSX file", "var a = 'Line 1 \\\n Line 2';", true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoMultiStr, multiStrJsxFile, testCase.sourceText)
			if testCase.wantFires {
				rule_testing.ExpectFindings(t, result, "multilineString")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// TestNoMultiStrDistinguishesEscapesFromContinuations pins that the subject is the raw text.
//
// A `\n` escape and a real line continuation are nearly the same value and completely different
// source. Upstream tests `node.raw`, so the escape is clean and the continuation reports; measured
// against the installed build. A port reading the cooked value reports the escape, which is the
// ordinary correct way to write a newline, and would fire on a large fraction of every real tree.
//
// The last case is the sharpest: a string whose VALUE contains a newline because of an escape,
// sitting on one source line, is clean, while a string whose value contains NO newline because the
// continuation swallows it does report. The rule is about the source, not the string.
func TestNoMultiStrDistinguishesEscapesFromContinuations(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFires  bool
	}{
		{"an escape sequence", "var a = 'Line 1 \\n Line 2';", false},
		{"a line continuation", "var a = 'Line 1 \\\n Line 2';", true},
		{"an escape and no break", "var a = 'a\\nb\\nc';", false},
		{"a continuation with nothing after it", "var a = 'a\\\n';", true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoMultiStr, multiStrFile, testCase.sourceText)
			if testCase.wantFires {
				rule_testing.ExpectFindings(t, result, "multilineString")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// TestNoMultiStrDeclinesOtherLiteralKinds pins the surface.
//
// A template literal spanning lines is the repair this rule points at, not the defect, and it is a
// different node in ESTree that never reaches upstream's handler. Measured: a multiline template is
// clean on the installed build. Here that falls out of the listener map, and a port adding the
// template kind to it would report on every multiline template in the tree.
func TestNoMultiStrDeclinesOtherLiteralKinds(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a template spanning lines", "var a = `Line 1\nLine 2`;"},
		{"a tagged template spanning lines", "var a = tag`Line 1\nLine 2`;"},
		{"a regular expression", "var a = /foo/;"},
		{"a number", "var a = 12;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoMultiStr, multiStrFile, testCase.sourceText))
		})
	}
}

// TestNoMultiStrReportsTheWholeLiteral asserts where every finding points, which ExpectFindings
// cannot see.
//
// The rule carries no repair, so nothing downstream would notice a finding anchored on the wrong
// node. Upstream reports the literal node; measured against the installed build,
// `var x = 'Line 1 \` continued reports from column 9 on line 1 to column 9 on line 2, which is the
// string including both quotes and spanning the break.
//
// Slicing the source with the finding's own range is what makes this an assertion about the span
// rather than about the line number, and a span that stopped at the line break would slice to
// something with no closing quote.
func TestNoMultiStrReportsTheWholeLiteral(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantStart  int
		wantText   string
	}{
		{"a declaration", "var x = 'Line 1 \\\n Line 2'", 8, "'Line 1 \\\n Line 2'"},
		{"a call argument", "test('Line 1 \\\n Line 2');", 5, "'Line 1 \\\n Line 2'"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoMultiStr, multiStrFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if result.Diagnostics[0].Range.Pos() != testCase.wantStart || reported != testCase.wantText {
				t.Errorf("finding at [%d,%d) = %q, want start %d and %q",
					result.Diagnostics[0].Range.Pos(), result.Diagnostics[0].Range.End(),
					reported, testCase.wantStart, testCase.wantText)
			}
		})
	}
}

// TestNoMultiStrMessage asserts the reported id and that a description is present.
//
// The id literal is typed here rather than read from the rule's own constant, so a rename cannot
// move both sides at once and stay green.
func TestNoMultiStrMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoMultiStr, multiStrFile, "var x = 'Line 1 \\\n Line 2'")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "multilineString" {
		t.Errorf("message id = %q, want %q", got, "multilineString")
	}
	if result.Diagnostics[0].Message.Description == "" {
		t.Error("the message carries no description, so the finding says nothing about why")
	}
}
