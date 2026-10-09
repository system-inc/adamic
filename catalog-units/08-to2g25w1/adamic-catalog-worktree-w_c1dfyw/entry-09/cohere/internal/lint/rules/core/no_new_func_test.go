package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// newFuncFile is where the fixtures pretend to live.
const newFuncFile = "/repository/source/NewFunc.ts"

// The corpus is ESLint's own, extracted rather than retyped.
//
// Every case is verbatim from `eslint/tests/lib/rules/no-new-func.js` at eslint 10.8.1: 13 valid and
// 10 invalid, one `noFunctionConstructor` each. Extracted by executing that file against a stub
// RuleTester and rendered through a serializer, so no escape was typed on the way here. All 23 were
// additionally driven through the installed build and agreed with the file, spans included.
//
// This rule reads the checker, so the fixtures run through `rule_testing.RunTyped`. The plain harness
// hands the rule a nil checker, and because this rule guards on that it would go completely silent:
// every silent case would pass vacuously and every reporting case would fail in a way that reads
// like a rule defect. `TestNoNewFuncNeedsTheTypedHarness` at the bottom pins that.
func TestNoNewFuncFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"new with the global", "var a = new Function(\"b\", \"c\", \"return b+c\");"},
		{"a plain call of the global", "var a = Function(\"b\", \"c\", \"return b+c\");"},
		{"call on the global", "var a = Function.call(null, \"b\", \"c\", \"return b+c\");"},
		{"apply on the global", "var a = Function.apply(null, [\"b\", \"c\", \"return b+c\"]);"},
		{"bind on the global, then invoked", "var a = Function.bind(null, \"b\", \"c\", \"return b+c\")();"},
		{"bind on the global, uncalled", "var a = Function.bind(null, \"b\", \"c\", \"return b+c\");"},
		{"a subscripted call on the global", "var a = Function[\"call\"](null, \"b\", \"c\", \"return b+c\");"},
		{"a parenthesized optional call on the global", "var a = (Function?.call)(null, \"b\", \"c\", \"return b+c\");"},
		{"a class shadowing only an inner scope", "const fn = () => { class Function {} }; new Function('', '')"},
		{"a function declaration shadowing only an inner scope", "var fn = function () { function Function() {} }; Function('', '')"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoNewFunc, newFuncFile, testCase.sourceText),
				"noFunctionConstructor")
		})
	}
}

// The clean cases carry the whole shadow discrimination.
//
// Six of the thirteen are a local binding named `Function`, in four different declaration forms, and
// no bounded structural walk separates them from the two reporting cases where the shadow sits in an
// inner scope. The rest pin that a mention of the global is not an invocation of it: passed as an
// argument, used as a computed key, or reached through a method that does not construct.
func TestNoNewFuncStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a differently-named constructor", "var a = new _function(\"b\", \"c\", \"return b+c\");"},
		{"a differently-named call", "var a = _function(\"b\", \"c\", \"return b+c\");"},
		{"a class shadowing the global", "class Function {}; new Function()"},
		{"a class shadowing inside an arrow body", "const fn = () => { class Function {}; new Function() }"},
		{"a function declaration shadowing the global", "function Function() {}; Function()"},
		{"a function declaration shadowing inside a function body", "var fn = function () { function Function() {}; Function() }"},
		{"a named function expression shadowing its own body", "var x = function Function() { Function(); }"},
		{"the global passed as an argument", "call(Function)"},
		{"the global passed to another constructor", "new Class(Function)"},
		{"the global used as a computed key", "foo[Function]()"},
		{"a method reference passed as an argument", "foo(Function.bind)"},
		{"a method that does not construct", "Function.toString()"},
		{"a computed method name that is not static", "Function[call]()"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoNewFunc, newFuncFile, testCase.sourceText))
		})
	}
}

// Where the finding points, which the message-id fixtures above cannot see.
//
// The indirect form is the interesting one and it is measured rather than inferred: on eslint
// 10.8.1, `Function.bind(null, "b", "c", "return b+c")()` reports columns 9 through 52, which ends
// at the bind call's own closing paren rather than at the trailing `()`. A port anchoring on the
// outermost call passes every message-id fixture in this file and is wrong on that row.
//
// `RunTyped` writes `strings.TrimSpace(contents)+"\n"` to disk, so a fixture with leading whitespace
// is one byte offset from the literal here. These have none, and the assertion slices the trimmed
// text rather than the literal so that a later edit adding a blank line fails loudly instead of
// silently sliding.
func TestNoNewFuncReportsTheInvokingExpression(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantPos    int
		wantEnd    int
	}{
		{"var a = new Function(\"b\", \"c\", \"return b+c\");", 8, 44},
		{"var a = Function(\"b\", \"c\", \"return b+c\");", 8, 40},
		{"var a = Function.bind(null, \"b\", \"c\", \"return b+c\")();", 8, 51},
		{"var a = Function.bind(null, \"b\", \"c\", \"return b+c\");", 8, 51},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoNewFunc, newFuncFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
			}
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			got := result.Diagnostics[0].Range
			if got.Pos() != testCase.wantPos || got.End() != testCase.wantEnd {
				t.Fatalf("reported [%d:%d) which is %q, wanted [%d:%d) which is %q",
					got.Pos(), got.End(), onDisk[got.Pos():got.End()],
					testCase.wantPos, testCase.wantEnd,
					onDisk[testCase.wantPos:testCase.wantEnd])
			}
		})
	}
}

// Reporting shapes the corpus does not write, each measured upstream first.
//
// The parenthesized receiver is the one the corpus nearly reaches: it writes
// `(Function?.call)(null, ...)` but never a plain `(Function).call()`, and our tree keeps the
// parenthesis node that ESTree discards, so both need the skip.
func TestNoNewFuncFiresOnShapesTheCorpusOmits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an optional call of the global", "Function?.();"},
		{"an optional method call on the global", "Function?.call();"},
		{"a parenthesized receiver", "(Function).call();"},
		{"new with no argument list", "new Function;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoNewFunc, newFuncFile, testCase.sourceText),
				"noFunctionConstructor")
		})
	}
}

// Silent shapes the corpus does not write, measured upstream first.
//
// `Function.bind;` is the sharper of the two: the corpus asserts that `Function.bind(...)` reports
// even when nothing calls the result, so a reader could reasonably conclude the rule anchors on the
// member access. It does not, and this is the case that separates them.
func TestNoNewFuncDeclinesShapesTheCorpusOmits(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an uncalled method reference", "Function.bind;"},
		{"a method that does not construct", "Function.notAMethod();"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoNewFunc, newFuncFile, testCase.sourceText))
		})
	}
}

// The typed harness is required, and a revert to the plain one must fail loudly.
//
// This rule guards on a nil checker, so under `rule_testing.Run` it goes completely silent rather than
// panicking. Silence is the more dangerous failure: every StaysSilent case above would pass
// vacuously. This asserts the difference directly on an input the rule certainly reports.
func TestNoNewFuncNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	const source = "var a = new Function(\"b\", \"c\", \"return b+c\");"

	rule_testing.ExpectFindings(t,
		rule_testing.RunTyped(t, NoNewFunc, newFuncFile, source), "noFunctionConstructor")

	if got := len(rule_testing.Run(t, NoNewFunc, newFuncFile, source).Diagnostics); got != 0 {
		t.Fatalf("the untyped harness produced %d diagnostics, so the nil-checker guard has moved "+
			"and this test no longer measures what it claims", got)
	}
}

// The message, asserted against literals typed here rather than the rule's own constants.
func TestNoNewFuncReportsWhyItMatters(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, NoNewFunc, newFuncFile,
		"var a = new Function(\"b\", \"c\", \"return b+c\");")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "noFunctionConstructor" {
		t.Fatalf("message id was %q", got)
	}
	const wantPrefix = "This builds a function by compiling a string at runtime"
	if got := result.Diagnostics[0].Message.Description; !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("description was %q", got)
	}
}

// The shadow discrimination on the INDIRECT arm, which the imported corpus cannot see.
//
// Upstream's corpus shadows `Function` in six cases and reaches it through `call`/`bind`/`apply` in
// four, and no case does both. So a port that checks the global on the direct arm and forgets it on
// the indirect one passes all 23 imported cases. A mutation dropping exactly that check survived
// them, which is how this test came to exist.
//
// All four rows measured clean on eslint 10.8.1 before being written here.
func TestNoNewFuncDeclinesAShadowReachedThroughAMethod(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a class shadow reached through call", "class Function {}; Function.call(null, \"return 1\");"},
		{"a class shadow reached through bind", "class Function {}; Function.bind(null, \"return 1\");"},
		{"a function-declaration shadow reached through apply", "function Function() {}; Function.apply(null, []);"},
		{"a class shadow reached through a subscripted call", "class Function {}; Function[\"call\"](null);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoNewFunc, newFuncFile, testCase.sourceText))
		})
	}
}

// The accept set on the method name, which the imported corpus cannot see.
//
// The corpus writes the method as a bare identifier and once as a string subscript, and both are
// accepted by the narrower `property.Textual` set as well, so a mutation narrowing the set survived
// all 27 cases. Upstream asks for the static property name, which resolves a no-substitution
// template too, and eslint 10.8.1 reports both template rows below.
//
// The numeric row is the other side of the same set: `Function[0]` names no method and is clean on
// the same build, which is what makes the wider set safe rather than merely wider.
func TestNoNewFuncReadsATemplateMethodName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a template-literal method name", "Function[`call`](null, \"return 1\");"},
		{"a template-literal bind", "Function[`bind`](null);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoNewFunc, newFuncFile, testCase.sourceText),
				"noFunctionConstructor")
		})
	}
}

func TestNoNewFuncDeclinesANumericMethodSubscript(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a numeric subscript, which names no method", "Function[0](null);"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoNewFunc, newFuncFile, testCase.sourceText))
		})
	}
}

// The receiver kind guard is CRASH protection, not a behavioral filter, and no findings fixture can
// see that.
//
// `Node.Text()` panics with "Unhandled case in Node.Text" on a `KindPropertyAccessExpression`, so
// without the kind test `foo.Function.call(null)` takes the whole linter down rather than producing
// a wrong finding. A mutation removing the guard therefore survives every `ExpectFindings` and
// `ExpectClean` case in this file while being a crash, which is why these rows exist and why they
// are described as what they PREVENT rather than what they decide.
//
// All four are clean on eslint 10.8.1: upstream reaches the global through its scope analysis rather
// than by reading a receiver's text, so a qualified name never enters its loop at all.
func TestNoNewFuncSurvivesANonIdentifierReceiver(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a qualified receiver", "declare const foo: any;\nfoo.Function.call(null);\n"},
		{"a comma expression as receiver", "declare const a: any;\n(a, Function).call(null);\n"},
		{"a nested property receiver", "Function.Function.call(null);\n"},
		{"a call result as receiver", "declare const a: any;\na().Function.call(null);\n"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Reaching this assertion at all is most of the test: without the guard the call above
			// panics inside the walk and the failure is a stack trace rather than a diagnostic count.
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoNewFunc, newFuncFile, testCase.sourceText))
		})
	}
}
