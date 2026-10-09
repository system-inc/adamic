package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// divRegexFile is where the fixtures pretend to live.
const divRegexFile = "/repository/source/DivRegex.ts"

// The corpus is ESLint's own, copied rather than rewritten.
//
// Every case below is verbatim from `eslint/tests/lib/rules/no-div-regex.js`, lifted by evaluating
// that test file with its `RuleTester` stubbed and serialising the spec object it was handed. It is
// a very small corpus: 2 pass and 1 fail, with a single fix vector.
//
// That thinness is why the sections after it are longer than the import. One failing case cannot
// distinguish a positional test from a pattern parse, cannot see the flags, and cannot see a fix
// that rewrites the wrong byte, because with a single equals sign in the input every wrong span
// that lands inside the literal produces different source but there is only one input to notice it
// on.
func TestNoDivRegexFires(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoDivRegex, divRegexFile, "var f = function() { return /=foo/; };"),
		"unexpected")
}

// The two clean cases upstream ships, and they fail the test different ways.
//
// `/foo/ig` simply does not begin with an equals sign. `/\=foo/` does contain one at the start of
// the pattern, but the character at index 1 of the token is the backslash, so the positional test
// fails. That second case is the whole reason the rule needs no escape handling of its own.
func TestNoDivRegexStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"var f = function() { return /foo/ig.test('bar'); };",
		`var f = function() { return /\=foo/; };`,
	}

	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoDivRegex, divRegexFile, sourceText))
		})
	}
}

// The single fix vector upstream ships, asserted by applying the repair and comparing the whole
// rewritten file rather than by comparing the fix's text.
//
// The distinction matters more here than the one vector suggests: a fix writing `[=]` over the
// wrong byte produces a different file and an identical fix text, so a text comparison cannot tell
// the two apart. The cases below this one are where that gets exercised properly.
func TestNoDivRegexFixes(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFixedSource(t,
		rule_testing.Run(t, NoDivRegex, divRegexFile, "var f = function() { return /=foo/; };"),
		"var f = function() { return /[=]foo/; };")
}

// The cases below are ours. Each exists because reading our own code or driving the installed
// ESLint build raised a question the imported corpus does not answer.

// TestNoDivRegexIsPositionalNotSemantic is the distinction the corpus cannot make.
//
// Upstream asks whether the token's character at index 1 is an equals sign, and nothing else. It is
// not asking what the pattern matches, and the two questions disagree on real inputs. Every verdict
// below was measured against the installed ESLint build (version 10.8.1, driven through the Linter
// API) rather than derived from the source:
//
//	/[=]foo/  clean    matches an equals sign, and is this rule's own repair
//	/(=)/     clean    matches an equals sign, in a group
//	/==/      reports  index 1 is an equals sign
//	/=/       reports  the whole pattern
//
// A port that parsed the pattern and asked whether the first element can match `=` would report the
// first two, one of which is the output of its own fixer. That is a rule that never converges.
func TestNoDivRegexIsPositionalNotSemantic(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFires  bool
	}{
		{"an equals sign in a character class", "var a = /[=]foo/;", false},
		{"an equals sign in a group", "var a = /(=)/;", false},
		{"an escaped equals sign", `var a = /\=foo/;`, false},
		{"two equals signs", "var a = /==/;", true},
		{"nothing but an equals sign", "var a = /=/;", true},
		{"an equals sign with flags", "var a = /=foo/g;", true},
		{"an equals sign not at the start", "var a = /f=oo/;", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDivRegex, divRegexFile, testCase.sourceText)
			if testCase.wantFires {
				rule_testing.ExpectFindings(t, result, "unexpected")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// TestNoDivRegexDeclinesTheConstructor pins the surface.
//
// Upstream requires the node's first token to be of type `RegularExpression`, so a constructor call
// never qualifies: its first token is `new` or the callee name. Measured, `new RegExp('=foo')` is
// clean. That is a real difference from `no-regex-spaces`, which handles both spellings, and it is
// upstream's difference rather than something dropped here.
func TestNoDivRegexDeclinesTheConstructor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"the constructor form", "var a = new RegExp('=foo');"},
		{"the call form", "var a = RegExp('=foo');"},
		{"a plain string", "var a = '=foo';"},
		{"a template", "var a = `=foo`;"},
		{"a division that is not a regular expression", "var a = b /= c;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoDivRegex, divRegexFile, testCase.sourceText))
		})
	}
}

// TestNoDivRegexReportsTheWholeLiteral asserts where the finding points, which neither
// ExpectFindings nor ExpectFixedSource can see.
//
// The finding and the fix have deliberately different spans here: upstream reports the whole
// literal, flags included, and repairs one byte inside it. Its corpus states columns 29 through 35
// for `/=foo/`, which is the six characters of the literal.
//
// The flags case is the one that separates "the literal" from "the pattern": `/=foo/g` reports a
// span of seven, measured on the installed build, so the finding covers the trailing flag. A port
// reporting the pattern alone would pass every message-id fixture and every fix assertion in this
// file while pointing one character short.
//
// The parenthesized case pins that the span is the literal rather than the parentheses; measured.
func TestNoDivRegexReportsTheWholeLiteral(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantStart  int
		wantText   string
	}{
		{"var f = function() { return /=foo/; };", 28, "/=foo/"},
		{"var a = /=foo/g;", 8, "/=foo/g"},
		{"var a = /=/;", 8, "/=/"},
		{"var a = (/=foo/);", 9, "/=foo/"},
		{"var a = /=foo/gimsuy;", 8, "/=foo/gimsuy"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDivRegex, divRegexFile, testCase.sourceText)
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

// TestNoDivRegexRepairsExactlyTheEqualsSign is the fix assertion the single imported vector cannot
// make.
//
// A fix is applied unattended, so a repair landing on the wrong byte silently rewrites working
// code. With one vector and one equals sign in the input, several wrong spans produce plausible
// output; these separate them.
//
// The flags case is the sharpest: a repair anchored on the literal's END rather than its start
// would land in or past the flags. The `/=/` case pins that the repair does not eat the closing
// slash. And the second and third cases have a second equals sign later in the pattern, which a
// repair searching for an equals sign rather than taking index 1 would find instead.
//
// Every expected output below parses as a regular expression matching the same input as its
// original, which is the property that makes this a fix rather than a suggestion.
func TestNoDivRegexRepairsExactlyTheEqualsSign(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSource string
	}{
		{"var a = /=foo/;", "var a = /[=]foo/;"},
		{"var a = /=f=oo/;", "var a = /[=]f=oo/;"},
		{"var a = /==/;", "var a = /[=]=/;"},
		{"var a = /=/;", "var a = /[=]/;"},
		{"var a = /=foo/g;", "var a = /[=]foo/g;"},
		{"var a = /=foo/gimsuy;", "var a = /[=]foo/gimsuy;"},
		{"var a = (/=foo/);", "var a = (/[=]foo/);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t,
				rule_testing.Run(t, NoDivRegex, divRegexFile, testCase.sourceText), testCase.wantSource)
		})
	}
}

// TestNoDivRegexRepairIsIdempotent pins that the fixer's own output is clean.
//
// This is the property a semantic port would break: `/[=]foo/` is what the repair writes, and if
// the rule reported it the edit engine would rewrite it again on the next pass. The positional test
// makes this fall out rather than needing a guard, and this pins it so a later rewrite toward a
// pattern parse fails loudly instead of producing a rule that never settles.
func TestNoDivRegexRepairIsIdempotent(t *testing.T) {
	t.Parallel()

	repaired := "var a = /[=]foo/;"
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoDivRegex, divRegexFile, repaired))
}

// TestNoDivRegexMessage asserts the reported id and that a description is present.
//
// The id literal is typed here rather than read from the rule's own constant, so a rename cannot
// move both sides at once and stay green.
func TestNoDivRegexMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoDivRegex, divRegexFile, "var a = /=foo/;")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "unexpected" {
		t.Errorf("message id = %q, want %q", got, "unexpected")
	}
	if result.Diagnostics[0].Message.Description == "" {
		t.Error("the message carries no description, so the finding says nothing about why")
	}
	if len(result.Diagnostics[0].Fixes) != 1 {
		t.Errorf("got %d fixes, want 1", len(result.Diagnostics[0].Fixes))
	}
}

// TestNoDivRegexSurvivesATruncatedLiteral pins the length guard, which no findings assertion can
// see because what it prevents is a panic rather than a wrong verdict.
//
// A mutation deleting the guard survived every other fixture in this file, which is the shape the
// brief calls crash protection: the sweep reports SURVIVED identically whether the guard matters or
// not, because a panic is not a finding.
//
// It matters. The grammar cannot produce a regular expression literal shorter than two characters,
// but the parser recovers from source that is not well formed, and error recovery hands back
// exactly that. Probing eleven malformed shapes found six that produce a one-byte literal holding
// just the opening slash, of which the four below are the most ordinary: a truncated file, an
// unterminated pattern, a stray slash in a bracket, and one inside a condition. Indexing byte 1 of
// any of them takes the whole linter down on a file that was already failing to parse.
//
// The assertion is that the rule returns at all. Each input is also expected to be clean, since a
// literal with no second character cannot begin with an equals sign.
func TestNoDivRegexSurvivesATruncatedLiteral(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an unterminated pattern", "var a = /;"},
		{"a truncated file", "var a = /"},
		{"a stray slash in a bracket", "var a = [/];"},
		{"a stray slash in a condition", "if (/) {}"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// A panic here is the defect under test, so it is allowed to fail the test rather than
			// being recovered: an unrecovered panic in a rule takes the whole run down in
			// production, and the test should say so in the same voice.
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoDivRegex, divRegexFile, testCase.sourceText))
		})
	}
}

// TestNoDivRegexOnAnUnterminatedLiteralIsOurs records a divergence that comes from the parser
// rather than from the rule, so the next reader does not helpfully correct it.
//
// `var a = /=` is a two-byte regular expression literal our parser recovers into a real node, and
// the character at index 1 is an equals sign, so this rule reports and offers its repair. ESLint
// cannot express the input at all: driving the installed build on it returns a fatal
// "Unterminated regular expression" parse error and the rule never runs.
//
// So this is not a behavioural disagreement about the rule, it is a file that only one of the two
// parsers admits. Reporting is the consistent answer, since our tree does hand the rule the node,
// and going silent would need a special case whose only justification is matching a tool that never
// reached the code. Recorded here rather than argued at the rule, because it is a fact about the
// harness boundary.
//
// The intuitive reading, which is the one most likely to get written in later: that a rule should
// stay quiet on source that does not parse. That is a defensible position and it is not this
// rule's, because the file already carries a syntax error from the compiler and a second complaint
// pointing at the same two characters is what the sibling regex rules also decline to suppress.
func TestNoDivRegexOnAnUnterminatedLiteralIsOurs(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoDivRegex, divRegexFile, "var a = /="), "unexpected")
}
