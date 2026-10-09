package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// symbolDescriptionFile is where the fixtures pretend to live.
const symbolDescriptionFile = "/repository/source/SymbolDescription.ts"

// The corpus is ESLint's own, at `tests/lib/rules/symbol-description.js`, copied rather than
// rewritten: 6 clean cases and 2 reporting ones, which is the whole upstream file.
//
// Every case string was verified byte against byte against the upstream file by script, and every
// verdict was reproduced by driving the installed eslint at 10.8.1 before being written here.
func TestSymbolDescriptionFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a bare call", "Symbol();"},
		{"a bare call before an assignment to the name", "Symbol(); Symbol = function () {};"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, SymbolDescription, symbolDescriptionFile, testCase.sourceText), "expected")
		})
	}
}

// Four of upstream's six clean cases are shadowing, which is what makes this a scope question rather
// than a call-shape one. A rule matching the name `Symbol` textually reports all four.
//
// The second case is worth naming separately: `Symbol(foo)` passes an argument whose value is not
// knowable, and upstream counts arguments rather than inspecting them. So a description that is
// empty at run time is not this rule's business, which is a decision rather than a gap.
func TestSymbolDescriptionStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a call with a literal description", `Symbol("Foo");`},
		{"a call with a variable description", `var foo = "foo"; Symbol(foo);`},
		{"a local function shadowing the name", "var Symbol = function () {}; Symbol();"},
		{"a local shadow declared after the call", "Symbol(); var Symbol = function () {};"},
		{"a shadow inside a function", "function bar() { var Symbol = function () {}; Symbol(); }"},
		{"a parameter shadowing the name", "function bar(Symbol) { Symbol(); }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, SymbolDescription, symbolDescriptionFile, testCase.sourceText))
		})
	}
}

// Cases upstream does not write, each measured against the installed eslint at 10.8.1 with
// `ecmaVersion: 6, sourceType: "script"` before being recorded here rather than reasoned about.
//
// The parenthesized callee is the one that needs the skip: espree gives parentheses no node, so
// upstream's `isCallee` already sees through them, and it reports `(Symbol)()` at columns 1 to 11.
func TestSymbolDescriptionFiresOnCasesBeyondTheCorpus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantCount  int
	}{
		{"a parenthesized callee", "(Symbol)();", 1},
		{"a bare call as an initializer", "let x = Symbol();", 1},
		{"a bare call as an argument", "foo(Symbol());", 1},
		{"two bare calls report twice", "Symbol(); Symbol();", 2},
		// A block-scoped shadow does not reach the outer call, so the outer `Symbol` is still the
		// global one and the call still reports.
		{"a block-scoped shadow that does not reach the call", "{ let Symbol = 1; } Symbol();", 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			wantIds := make([]string, testCase.wantCount)
			for index := range wantIds {
				wantIds[index] = "expected"
			}
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, SymbolDescription, symbolDescriptionFile, testCase.sourceText), wantIds...)
		})
	}
}

// The clean cases beyond the corpus pin three decisions upstream makes that read as oversights and
// are not, all measured against the installed build.
//
// `new Symbol()` is clean because upstream's `isCallee` is false for a `new` expression, even though
// that call also has no description and also throws. `Symbol(undefined)` and `Symbol(...args)` are
// clean because the check is `arguments.length === 0` and neither has zero arguments, so a spread
// that expands to nothing is not seen. And `globalThis.Symbol()` is clean because the identifier
// there is a property rather than the callee.
func TestSymbolDescriptionStaysSilentOnCasesBeyondTheCorpus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a new expression", "new Symbol();"},
		{"an explicit undefined argument", "Symbol(undefined);"},
		{"a spread argument", "Symbol(...args);"},
		{"a member call on the namespace", "Symbol.for();"},
		{"a property access rather than a call", "Symbol.iterator;"},
		{"a reference through globalThis", "globalThis.Symbol();"},
		{"a call to a method named call", "Symbol.call(null);"},
		{"a class shadowing the name", "class Symbol {}; Symbol();"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, SymbolDescription, symbolDescriptionFile, testCase.sourceText))
		})
	}
}

// Upstream's corpus states no columns for this rule, so the span was taken from the installed build
// rather than from the corpus. Measured: `Symbol();` reports columns 1 to 9 and `(Symbol)();`
// reports columns 1 to 11, which are the whole call expression in both cases including the
// parentheses around the callee.
func TestSymbolDescriptionSpansTheCall(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		sourceText   string
		wantReported string
	}{
		{"a bare call", "Symbol();", "Symbol()"},
		{"a parenthesized callee", "(Symbol)();", "(Symbol)()"},
		{"a call as an initializer", "let x = Symbol();", "Symbol()"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, SymbolDescription, symbolDescriptionFile, testCase.sourceText)
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
// failure modes, because every StaysSilent case would pass vacuously.
//
// This asserts the typed harness is required, so a later revert to `rule_testing.Run` fails loudly
// here rather than turning the whole clean set into a vacuous green.
func TestSymbolDescriptionRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.RunTyped(t, SymbolDescription, symbolDescriptionFile, "Symbol();"), "expected")
	rule_testing.ExpectClean(t,
		rule_testing.Run(t, SymbolDescription, symbolDescriptionFile, "Symbol();"))
}

// The message text is asserted against a literal typed here rather than against the rule's own
// constant, because comparing a diagnostic to the constant it was built from moves both sides
// together under mutation and asserts nothing.
func TestSymbolDescriptionMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, SymbolDescription, symbolDescriptionFile, "Symbol();")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want one finding, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "expected" {
		t.Errorf("message id is %q, want %q", got, "expected")
	}
	if got := result.Diagnostics[0].Message.Description; !strings.HasPrefix(got, "This creates a symbol with no description") {
		t.Errorf("message description starts %q, which is not the sentence this rule reports", got)
	}
}

// Other globals called with no arguments, which are the inputs that separate the name test from the
// global-resolution test.
//
// A mutation dropping `callee.Text() != "Symbol"` survived every other fixture in this file, because
// nothing above reached the rule with a global that is not `Symbol`. Each of these resolves to the
// standard library exactly as `Symbol` does and each is clean upstream, measured.
func TestSymbolDescriptionStaysSilentOnOtherGlobals(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"isNaN", "isNaN();"},
		{"parseInt", "parseInt();"},
		{"Number", "Number();"},
		{"String", "String();"},
		{"Boolean", "Boolean();"},
		{"Array", "Array();"},
		{"Object", "Object();"},
		{"Date", "Date();"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, SymbolDescription, symbolDescriptionFile, testCase.sourceText))
		})
	}
}
