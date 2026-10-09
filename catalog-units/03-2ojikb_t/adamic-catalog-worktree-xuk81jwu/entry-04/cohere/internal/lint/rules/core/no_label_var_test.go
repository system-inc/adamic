package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// labelVarFile is where the fixtures pretend to live.
const labelVarFile = "/repository/source/LabelVar.ts"

// The corpus is ESLint's own, at `tests/lib/rules/no-label-var.js`, copied rather than rewritten:
// 2 clean cases and 3 reporting ones, which is the whole upstream file.
//
// Every reporting case carries line and column assertions, so the span is stated data and is
// asserted below rather than left to a message id. Every case string was verified byte against byte
// against the upstream file, and every verdict reproduced by driving the installed eslint at 10.8.1.
func TestNoLabelVarFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a label clashing with an outer var", "var x = foo; function bar() { x: for(;;) { break x; } }"},
		{"a label clashing with a local var", "function bar() { var x = foo; x: for(;;) { break x; } }"},
		{"a label clashing with a parameter", "function bar(x) { x: for(;;) { break x; } }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoLabelVar, labelVarFile, testCase.sourceText),
				"identifierClashWithLabel")
		})
	}
}

// Both upstream clean cases are near misses, and each fails a different way.
//
// The first has a variable `q` in a DIFFERENT function, so the two scope chains never meet: a rule
// searching the whole file for the name reports it. The second has a variable in the same scope
// whose name is not the label's, which is the ordinary case a rule that ignored the name would
// report.
func TestNoLabelVarStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a variable of the same name in a sibling function", "function bar() { q: for(;;) { break q; } } function foo () { var q = t; }"},
		{"a variable of a different name in the same scope", "function bar() { var x = foo; q: for(;;) { break q; } }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoLabelVar, labelVarFile, testCase.sourceText))
		})
	}
}

// Cases upstream does not write, each measured against the installed eslint at 10.8.1 with
// `sourceType: "script"` before being recorded here rather than reasoned about.
//
// The two globals are the reason this rule asks for VALUES rather than for variables. Upstream's
// `getVariableByName` walks the scope chain to the global scope, where every standard library name
// is a variable, so `Object:` and `undefined:` both report. That reads like a false positive and is
// upstream's actual behaviour, reproduced rather than improved on.
//
// The function and class rows are the same point from the other side: a binding that is not a `var`
// at all still clashes, because the clash is about the NAME being taken.
func TestNoLabelVarFiresOnCasesBeyondTheCorpus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a function declaration of the same name", "function x() {} x: for(;;) { break x; }"},
		{"a class declaration of the same name", "class x {} x: for(;;) { break x; }"},
		{"a let binding of the same name", "let x = 1; x: for(;;) { break x; }"},
		{"a const binding of the same name", "const x = 1; x: for(;;) { break x; }"},
		{"a global of the same name", "Object: for(;;) { break Object; }"},
		{"the undefined global", "undefined: for(;;) { break undefined; }"},
		// Hoisting: the var is declared after the label and still clashes, because a function-scoped
		// declaration is in scope throughout the function.
		{"a var hoisted from after the label", "function bar() { x: for(;;) { break x; } var x; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoLabelVar, labelVarFile, testCase.sourceText),
				"identifierClashWithLabel")
		})
	}
}

// The clean cases beyond the corpus are the block-scoping ones, and they are what separate "is this
// name in scope HERE" from "does this name appear anywhere".
//
// A `let` inside a block that has already closed is not in scope at the label. A `let` inside the
// label's own body is not in scope at the labeled statement either, because the label sits outside
// the block it labels. Both measured clean upstream, and a rule searching the file for the name
// reports both.
func TestNoLabelVarStaysSilentOnCasesBeyondTheCorpus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"no binding of that name at all", "x: for(;;) { break x; }"},
		{"a block-scoped binding that has already closed", "{ let x = 1; } x: for(;;) { break x; }"},
		{"a binding inside the label's own body", "x: for(;;) { let x = 1; break x; }"},
		{"a second label rather than a variable", "x: y: for(;;) { break x; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoLabelVar, labelVarFile, testCase.sourceText))
		})
	}
}

// Every reporting case upstream carries line and column assertions. These convert those columns
// into the slice the finding names, which is the whole labeled statement rather than the label.
//
// `var x = foo; function bar() { x: for(;;) { break x; } }` asserts columns 31 through 54, one-based
// and end-exclusive, which is the zero-based half-open range [30,53) and the text
// `x: for(;;) { break x; }`. Reporting the LABEL instead would be a defensible reading of the same
// rule with an identical message id, and no id fixture could tell the two apart.
func TestNoLabelVarSpansTheWholeLabeledStatement(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		wantReported string
	}{
		{"an outer var", "var x = foo; function bar() { x: for(;;) { break x; } }", "x: for(;;) { break x; }"},
		{"a local var", "function bar() { var x = foo; x: for(;;) { break x; } }", "x: for(;;) { break x; }"},
		{"a parameter", "function bar(x) { x: for(;;) { break x; } }", "x: for(;;) { break x; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoLabelVar, labelVarFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want one finding, got %d", len(result.Diagnostics))
			}
			source := result.SourceFile.Text()
			reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.wantReported {
				t.Errorf("the finding points at %q, want %q", reported, testCase.wantReported)
			}
		})
	}
}

// The rule declares NeedsTypeChecker, so the plain harness hands it a nil checker and the guard at
// the top of the listener makes it go completely silent. That is the more dangerous of the two
// failure modes, because every clean case would then pass vacuously.
func TestNoLabelVarRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.RunTyped(t, NoLabelVar, labelVarFile, "function bar(x) { x: for(;;) { break x; } }"),
		"identifierClashWithLabel")
	rule_testing.ExpectClean(t,
		rule_testing.Run(t, NoLabelVar, labelVarFile, "function bar(x) { x: for(;;) { break x; } }"))
}

// The message text is asserted against a literal typed here rather than against the rule's own
// constant, because comparing a diagnostic to the constant it was built from moves both sides
// together under mutation and asserts nothing.
func TestNoLabelVarMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoLabelVar, labelVarFile, "function bar(x) { x: for(;;) { break x; } }")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want one finding, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "identifierClashWithLabel" {
		t.Errorf("message id is %q, want %q", got, "identifierClashWithLabel")
	}
	if got := result.Diagnostics[0].Message.Description; !strings.HasPrefix(got, "This label has the same name as a variable in scope") {
		t.Errorf("message description starts %q, which is not the sentence this rule reports", got)
	}
}

// Block-scoped bindings inside a labeled BLOCK, which are what separate asking at the labeled
// statement from asking at its body.
//
// A mutation moving the scope query onto the label's own statement survived every fixture above,
// because the one case written for that distinction has a `for` statement as its body and both
// anchors agree there. Probed rather than guessed: the two anchors disagree only when the body
// introduces a scope the label is outside of, which a plain block and a `for` with a `let`
// initializer both do and a `for` with a body-level `let` does not.
//
// All five verdicts measured against the installed eslint at 10.8.1. The `var` case is the control
// and belongs with them: it REPORTS, because a function-scoped declaration is hoisted out of the
// block and is in scope at the label, so a rule that simply never looked inside would get the other
// four right for the wrong reason.
func TestNoLabelVarAsksAtTheLabelRatherThanInsideItsBody(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		sourceText  string
		wantFinding bool
	}{
		{"a let inside a labeled block", "x: { let x = 1; }", false},
		{"a function declaration inside a labeled block", "x: { function x() {} }", false},
		{"a class inside a labeled block", "x: { class x {} }", false},
		{"a let in a labeled for's initializer", "x: for (let x = 0;;) { break x; }", false},
		// The control: hoisted out of the block, so it IS in scope at the label.
		{"a var inside a labeled block", "x: { var x = 1; }", true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoLabelVar, labelVarFile, testCase.sourceText)
			if testCase.wantFinding {
				rule_testing.ExpectFindings(t, result, "identifierClashWithLabel")
				return
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}
