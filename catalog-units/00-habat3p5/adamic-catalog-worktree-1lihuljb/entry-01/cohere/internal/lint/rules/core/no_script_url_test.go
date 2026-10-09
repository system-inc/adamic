package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// scriptUrlFile is where the fixtures pretend to live.
const scriptUrlFile = "/repository/source/ScriptUrl.ts"

// The corpus is ESLint's own, copied rather than rewritten.
//
// Every case below is verbatim from `eslint/tests/lib/rules/no-script-url.js`: 6 pass and 4 fail,
// each failing case naming exactly one `unexpectedScriptURL`. The cases were lifted by evaluating
// that test file with its `RuleTester` stubbed and serialising the spec object it was handed, so
// the strings here are the ones upstream runs rather than a retyping of them.
//
// It is a small corpus for a rule with two surfaces and a parent test, which is why the sections
// after it are longer than usual. Four failing cases cannot distinguish a direct-parent tagged
// template check from a walk up the ancestors, and cannot see the case fold at all beyond one
// capital letter.
func TestNoScriptUrlFires(t *testing.T) {
	t.Parallel()

	cases := []string{
		"var a = 'javascript:void(0);';",
		"var a = 'javascript:';",
		"var a = `javascript:`;",
		"var a = `JavaScript:`;",
	}

	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoScriptUrl, scriptUrlFile, sourceText), "unexpectedScriptURL")
		})
	}
}

// The clean cases are the whole discrimination, and each fails a different way.
//
// `'xjavascript:'` and its template twin are the anchor: the scheme is present and not at the
// start. A template written `${foo}javascript:` carries an expression, so its value is not
// statically knowable, and it is a different node kind here rather than a helper returning null.
// A template written foo`javaScript:` is tagged, so the tag decides what the text means.
func TestNoScriptUrlStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"var a = 'Hello World!';",
		"var a = 10;",
		"var url = 'xjavascript:'",
		"var url = `xjavascript:`",
		"var url = `${foo}javascript:`",
		"var a = foo`javaScript:`;",
	}

	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoScriptUrl, scriptUrlFile, sourceText))
		})
	}
}

// The cases below are ours. Each exists because reading our own code or driving the installed
// ESLint build raised a question the imported corpus does not answer.

// TestNoScriptUrlTaggedTemplateExemptionIsTheDirectParent is the sharpest case in this file.
//
// Upstream declines a template whose immediate `parent` is a `TaggedTemplateExpression`, and
// nothing wider. So a template nested inside a tagged template's interpolation is NOT exempt: its
// parent is the enclosing template, not the tagged expression. Measured against the installed
// ESLint build (version 10.8.1, driven through the Linter API), a template written
// foo`x${`javascript:`}y` reports at the inner template.
//
// The imported corpus has one tagged case and no nested one, so a port asking "is this anywhere
// under a tagged template" passes every fixture above while being silent here. That is the whole
// reason this test exists.
//
// The `String.raw` case is the same shape with a member-expression tag, confirming the check is on
// the parent's kind rather than on the tag being a plain identifier.
func TestNoScriptUrlTaggedTemplateExemptionIsTheDirectParent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFires  bool
	}{
		{"a directly tagged template", "var a = foo`javascript:`;", false},
		{"a member-expression tag", "var a = String.raw`javascript:`;", false},
		{"a template nested in a tagged template", "var a = foo`x${`javascript:`}y`;", true},
		{"an untagged template", "var a = `javascript:`;", true},
		{"a parenthesized untagged template", "var a = (`javascript:`);", true},
		{"a string inside a tagged template", "var a = foo`x${'javascript:'}y`;", true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoScriptUrl, scriptUrlFile, testCase.sourceText)
			if testCase.wantFires {
				rule_testing.ExpectFindings(t, result, "unexpectedScriptURL")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// TestNoScriptUrlComparisonIsAnchoredAndAsciiFolded covers both halves of upstream's
// `value.toLowerCase().indexOf("javascript:") === 0`, neither of which the corpus exercises past a
// single input.
//
// The anchor group pins that the scheme must begin the value. `' javascript:x'` is clean upstream;
// measured. So is a tab before it, and so is any prefix at all. A port using a search rather than a
// prefix test reports all of them.
//
// The fold group pins that the comparison ignores case across every letter rather than the one
// capital `S` the corpus happens to write. It also pins the boundary: the fold is ASCII, so the
// long s (U+017F) does NOT match, which is where Go's `strings.EqualFold` and JavaScript's
// `toLowerCase` disagree. Measured against the installed build, the long s spelling is clean and
// the plain one beside it reports. Writing this rule with `EqualFold` would ship a false positive
// that no imported fixture could see.
func TestNoScriptUrlComparisonIsAnchoredAndAsciiFolded(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFires  bool
	}{
		{"the plain spelling", "var a = 'javascript:x';", true},
		{"all capitals", "var a = 'JAVASCRIPT:alert(1)';", true},
		{"alternating case", "var a = 'JaVaScRiPt:x';", true},
		{"a capital at the start only", "var a = 'Javascript:x';", true},
		{"the scheme with nothing after the colon", "var a = 'javascript:';", true},
		{"a leading space", "var a = ' javascript:x';", false},
		{"a leading tab", "var a = '\tjavascript:x';", false},
		{"a letter before the scheme", "var a = 'xjavascript:';", false},
		{"the scheme one character short", "var a = 'javascript';", false},
		{"an empty string", "var a = '';", false},
		{"a different scheme", "var a = 'https://example.com';", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoScriptUrl, scriptUrlFile, testCase.sourceText)
			if testCase.wantFires {
				rule_testing.ExpectFindings(t, result, "unexpectedScriptURL")
			} else {
				rule_testing.ExpectClean(t, result)
			}
		})
	}
}

// TestNoScriptUrlDeclinesOtherLiteralKinds pins the surface.
//
// A regular expression is a `Literal` upstream with a truthy value, and it is declined by the
// `typeof value === "string"` test rather than by anything about regular expressions. Measured:
// `/javascript:/` is clean on the installed build. Here that falls out of the listener map, and
// this is what pins that the translation did not widen it.
func TestNoScriptUrlDeclinesOtherLiteralKinds(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a regular expression", "var a = /javascript:/;"},
		{"a number", "var a = 10;"},
		{"a template carrying an expression", "var a = `javascript:${x}`;"},
		{"a template with an expression before the scheme", "var url = `${foo}javascript:`;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoScriptUrl, scriptUrlFile, testCase.sourceText))
		})
	}
}

// TestNoScriptUrlReportsTheWholeLiteral asserts where every finding points, which ExpectFindings
// cannot see.
//
// The rule carries no repair, so nothing downstream would notice a finding anchored on the wrong
// node. Upstream reports the literal node itself, which its corpus does not record columns for;
// measured against the installed build, `var a = 'javascript:void(0);';` reports columns 9 through
// 30, which is the string including both quotes. The parenthesized template case pins that the
// span is the template rather than the parentheses; measured.
func TestNoScriptUrlReportsTheWholeLiteral(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantStart  int
		wantText   string
	}{
		{"var a = 'javascript:void(0);';", 8, "'javascript:void(0);'"},
		{"var a = `javascript:`;", 8, "`javascript:`"},
		{"var a = (`javascript:`);", 9, "`javascript:`"},
		{"foo('javascript:');", 4, "'javascript:'"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoScriptUrl, scriptUrlFile, testCase.sourceText)
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

// TestNoScriptUrlMessage asserts the reported id and that a description is present.
//
// The id literal is typed here rather than read from the rule's own constant, so a rename cannot
// move both sides at once and stay green.
func TestNoScriptUrlMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoScriptUrl, scriptUrlFile, "var a = 'javascript:void(0);';")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("got %d findings, want 1", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "unexpectedScriptURL" {
		t.Errorf("message id = %q, want %q", got, "unexpectedScriptURL")
	}
	if result.Diagnostics[0].Message.Description == "" {
		t.Error("the message carries no description, so the finding says nothing about why")
	}
}
