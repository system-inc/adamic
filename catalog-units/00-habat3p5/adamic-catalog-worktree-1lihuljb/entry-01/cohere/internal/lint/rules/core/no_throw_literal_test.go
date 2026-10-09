package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// throwLiteralFile is where the fixtures pretend to live.
const throwLiteralFile = "/repository/source/ThrowLiteral.ts"

// The corpus is upstream's, extracted from its tester rather than retyped.
//
// `/tmp/lint-sources/eslint/tests/lib/rules/no-throw-literal.js` carries 25 valid cases and 17
// invalid ones. Sixteen of the invalid cases name `object` and exactly one names `undef`, so they
// are split into two tests below rather than asserted with a shared id.
//
// The clean list is the specification of `couldBeError` and every entry is a distinct arm. Read
// against the reporting list it becomes a set of pairs, and the pairs are what make the arms
// load-bearing: `throw foo = new Error()` is clean while `throw foo += new Error()` reports,
// `throw 'literal' && new Error()` is clean while `throw foo && 'literal'` reports, and
// `throw foo ? new Error() : 'literal'` is clean while `throw foo ? 'x' : 'literal'` reports. Each
// pair differs in one operator or one branch.
//
// This rule reads the checker for one question only, whether a bare `undefined` is the global, so
// its fixtures use `RunTyped`. Handed the plain harness it guards and goes silent, which would make
// every clean case here pass for the wrong reason.
func TestNoThrowLiteralStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"throw new Error();"},
		{"throw new Error('error');"},
		{"throw Error('error');"},
		{"var e = new Error(); throw e;"},
		{"function foo(undefined) { throw undefined; }"},
		{"try {throw new Error();} catch (e) {throw e;};"},
		{"throw a;"},
		{"throw foo();"},
		{"throw new foo();"},
		{"throw foo.bar;"},
		{"throw foo[bar];"},
		{"class C { #field; foo() { throw foo.#field; } }"}, // lang={"ecmaVersion":2022}
		{"throw foo = new Error();"},
		{"throw foo.bar ||= 'literal'"},  // lang={"ecmaVersion":2021}
		{"throw foo[bar] ??= 'literal'"}, // lang={"ecmaVersion":2021}
		{"throw 1, 2, new Error();"},
		{"throw 'literal' && new Error();"},
		{"throw new Error() || 'literal';"},
		{"throw foo ? new Error() : 'literal';"},
		{"throw foo ? 'literal' : new Error();"},
		{"throw tag `${foo}`;"},                                     // lang={"ecmaVersion":6}
		{"function* foo() { var index = 0; throw yield index++; }"}, // lang={"ecmaVersion":6}
		{"async function foo() { throw await bar; }"},               // lang={"ecmaVersion":8}
		{"throw obj?.foo"},                                          // lang={"ecmaVersion":2020}
		{"throw obj?.foo()"},                                        // lang={"ecmaVersion":2020}
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTyped(t, NoThrowLiteral, throwLiteralFile, testCase.sourceText))
		})
	}
}

// The sixteen cases upstream reports as `object`.
func TestNoThrowLiteralFiresObject(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"throw 'error';"},
		{"throw 0;"},
		{"throw false;"},
		{"throw null;"},
		{"throw {};"},
		{"throw 'a' + 'b';"},
		{"var b = new Error(); throw 'a' + b;"},
		{"throw foo = 'error';"},
		{"throw foo += new Error();"},
		{"throw foo &= new Error();"},
		{"throw foo &&= 'literal'"}, // lang={"ecmaVersion":2021}
		{"throw new Error(), 1, 2, 3;"},
		{"throw 'literal' && 'not an Error';"},
		{"throw foo && 'literal'"},
		{"throw foo ? 'not an Error' : 'literal';"},
		{"throw `${err}`;"}, // lang={"ecmaVersion":6}
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoThrowLiteral, throwLiteralFile, testCase.sourceText), "object")
		})
	}
}

// The one case upstream reports as `undef`, which is a different judgment rather than a variant.
//
// `throw undefined` reaches the second test rather than the first, because an identifier always
// "could be an Error" as far as the syntax knows. A port collapsing the two messages into one
// passes an id assertion that names either, which is why they are separated here.
func TestNoThrowLiteralFiresUndef(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
	}{
		{"throw undefined;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, NoThrowLiteral, throwLiteralFile, testCase.sourceText), "undef")
		})
	}
}

// Shapes our parser produces that upstream's does not, and one upstream never wrote.
//
// `??` and `??=` are in upstream's implementation and appear nowhere in its corpus, so both arms
// ship untested by the imported cases. `??` yields either operand, exactly like `||`, so a port
// mapping it onto the `&&` arm by mistake would be silent on the first case and report the second.
//
// The last is not a valid program: `throw;` is a syntax error. Upstream's parser refuses the file
// and the rule never sees it, while ours recovers and hands back a throw with no argument, which a
// port dereferencing it would crash on. No `ExpectFindings` fixture can see a panic, so this is
// pinned by asserting the run completes at all.
func TestNoThrowLiteralHandlesShapesTheCorpusOmits(t *testing.T) {
	t.Parallel()

	t.Run("?? yields either operand, so an Error on the left is clean", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoThrowLiteral, throwLiteralFile,
			"declare const foo: unknown;\nthrow new Error() ?? 'literal';\n"))
	})

	t.Run("?? with literals on both sides reports", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoThrowLiteral, throwLiteralFile,
			"declare const foo: unknown;\nthrow 'a' ?? 'b';\n"), "object")
	})

	t.Run("??= yields either operand", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoThrowLiteral, throwLiteralFile,
			"declare let foo: unknown;\nthrow foo ??= 'literal';\n"))
	})

	t.Run("a throw with no argument does not crash", func(t *testing.T) {
		t.Parallel()
		// Reporting or not is beside the point; not panicking is the assertion. The parser recovers
		// from this and hands back a ThrowStatement whose Expression is nil.
		rule_testing.RunTyped(t, NoThrowLiteral, throwLiteralFile, "function f() { throw; }")
	})
}

// The reported span is the whole throw statement, which no id fixture above can see.
//
// Upstream passes `node`, the ThrowStatement, rather than its argument. A port anchoring on the
// argument satisfies every assertion above and points a reader past the keyword that makes the line
// worth reading.
func TestNoThrowLiteralReportsTheWholeStatement(t *testing.T) {
	t.Parallel()

	const sourceText = "throw 'error';"
	result := rule_testing.RunTyped(t, NoThrowLiteral, throwLiteralFile, sourceText)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
	}
	written := result.SourceFile.Text()
	reported := written[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "throw 'error';" {
		t.Fatalf("reported %q, wanted the whole throw statement", reported)
	}
}

// The shadow test on `undefined`, which the single imported `undef` case cannot exercise both ways.
//
// Upstream asks `isGlobalReference`, and its clean case `function foo(undefined) { throw undefined; }`
// is the other half. A port dropping the shadow test reports that case; a port dropping the
// `undefined` name test reports `throw a`. Both halves are asserted here, with the reporting form as
// the control.
func TestNoThrowLiteralUndefOnlyForTheGlobal(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoThrowLiteral, throwLiteralFile,
		"function foo(undefined) { throw undefined; }"))
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoThrowLiteral, throwLiteralFile,
		"function foo() { throw undefined; }"), "undef")
}

// The typed harness is required, so a later revert to `rule_testing.Run` fails loudly.
func TestNoThrowLiteralNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoThrowLiteral.NeedsTypeChecker {
		t.Fatal("this rule resolves `undefined` through the checker and must declare it")
	}
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoThrowLiteral, throwLiteralFile, "throw 'error';"))
}
