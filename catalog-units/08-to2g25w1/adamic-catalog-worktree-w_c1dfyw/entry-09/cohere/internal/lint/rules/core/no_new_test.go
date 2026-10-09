package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// newFile is where the fixtures pretend to live.
const newFile = "/repository/source/New.ts"

// The corpus is ESLint's own and it is three cases, which is why most of this file is measured.
//
// `eslint/tests/lib/rules/no-new.js` carries 2 valid and 1 invalid. That is a floor and a very low
// one: it writes no parenthesized form, no nesting, no statement position other than the top level,
// and nothing that separates "the statement's expression is a new" from "a new appears in the
// statement". Every additional case below was driven through the installed eslint 10.8.1 build
// first, and the verdict recorded here is the one that build gave.
func TestNoNewFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a bare new statement", "new Date()"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoNew, newFile, testCase.sourceText), "noNewStatement")
		})
	}
}

func TestNoNewStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a new bound to a variable", "var a = new Date()"},
		{"a new compared inside a condition", "var a; if (a === new Date()) { a = false; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoNew, newFile, testCase.sourceText))
		})
	}
}

// Reporting shapes the three-case corpus does not reach, each measured upstream first.
//
// The parenthesized rows are the reason this test exists. Upstream's selector demands a direct
// child of the statement and gets one however many parentheses are written, because ESTree discards
// them; ours keeps them, so a literal port of the direct-child reading is silent on all three
// parenthesized rows while eslint reports every one.
//
// The nesting row pins that `new new Foo();` is one finding rather than two: only the outer `new` is
// the statement's expression.
func TestNoNewFiresOnShapesTheCorpusOmits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"no parentheses and no semicolon", "new Date"},
		{"no parentheses, with a semicolon", "new Date;"},
		{"one pair of parentheses", "(new Date());"},
		{"parentheses with no call", "(new Date);"},
		{"two pairs of parentheses", "((new Date()));"},
		{"a qualified constructor", "new foo.Bar();"},
		{"a computed callee", "new (foo())();"},
		{"nested new reports once", "new new Foo();"},
		{"the body of an if", "if (a) new Date();"},
		{"the body of a for", "for(;;) new Date();"},
		{"inside a function body", "function f(){ new Date(); }"},
		{"a labelled statement", "label: new Foo();"},
		{"the body of a do-while", "do new Foo(); while(a);"},
		{"inside a constructor", "class C { constructor(){ new Foo(); } }"},
		{"a switch case", "switch(a){ case 1: new Foo(); }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoNew, newFile, testCase.sourceText), "noNewStatement")
		})
	}
}

// Silent shapes the corpus does not reach, each measured upstream first.
//
// The comma row is the one worth naming: `new Date(), new Foo();` reports nothing on eslint 10.8.1,
// because the statement's expression is the comma rather than either `new`. That reads like a gap in
// the rule and it is upstream's, reproduced here rather than improved on.
func TestNoNewDeclinesShapesTheCorpusOmits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a new whose result is called through", "new Date().getTime();"},
		{"a new under void", "void new Date();"},
		{"a new inside an operator expression", "new Date() + 1;"},
		{"a comma expression, which upstream also declines", "new Date(), new Foo();"},
		{"a new after a comma", "a, new Foo();"},
		{"a new as a condition", "new Foo() ? 1 : 2;"},
		{"a new returned from a concise arrow", "() => new Foo();"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoNew, newFile, testCase.sourceText))
		})
	}
}

// The span is the statement, which the message-id fixtures above cannot see.
//
// Upstream reports `node.parent`, so the trailing semicolon is inside the finding and a
// parenthesized form is covered from the opening paren. Both are measured: `new Date;` is columns 1
// through 10 on eslint 10.8.1, one past the semicolon, and `((new Date()));` is 1 through 16.
//
// A port anchored on the `new` expression instead passes every message-id fixture in this file and
// is wrong on every row here.
func TestNoNewReportsTheWholeStatement(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantPos    int
		wantEnd    int
	}{
		{"new Date()", 0, 10},
		{"new Date;", 0, 9},
		{"(new Date());", 0, 13},
		{"((new Date()));", 0, 15},
		{"if (a) new Date();", 7, 18},
		{"new new Foo();", 0, 14},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoNew, newFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			got := result.Diagnostics[0].Range
			if got.Pos() != testCase.wantPos || got.End() != testCase.wantEnd {
				t.Fatalf("reported [%d:%d) which is %q, wanted [%d:%d) which is %q",
					got.Pos(), got.End(),
					testCase.sourceText[got.Pos():got.End()],
					testCase.wantPos, testCase.wantEnd,
					testCase.sourceText[testCase.wantPos:testCase.wantEnd])
			}
		})
	}
}

// The message, asserted against literals typed here rather than the rule's own constants.
func TestNoNewReportsWhyItMatters(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoNew, newFile, "new Date()")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "noNewStatement" {
		t.Fatalf("message id was %q", got)
	}
	const wantPrefix = "This constructs an object and then throws it away"
	if got := result.Diagnostics[0].Message.Description; !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("description was %q", got)
	}
}
