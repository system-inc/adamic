package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// constructorReturnFile is where the fixtures pretend to live.
//
// A `.ts` extension because that is what this tree lints and what the ESLint side runs through
// `@typescript-eslint/parser`. The rule reads no file extension, so this only decides which parser
// the harness picks, and TypeScript is the one whose verdicts were measured for the added cases
// below.
const constructorReturnFile = "/repository/source/ConstructorReturn.ts"

// The corpus is ESLint's own, extracted rather than retyped.
//
// Every case marked upstream below is verbatim from
// `/tmp/lint-sources/eslint/tests/lib/rules/no-constructor-return.js`, pulled out by parsing that
// file with espree and walking to the `ruleTester.run` call rather than by reading it: 18 valid and
// 2 invalid, each invalid case naming one error. Each extracted string was then byte-verified by
// re-evaluating the original source literal and comparing, and no case in this corpus contains a
// backslash at all, so there was no escape for a tool to cook on the way here.
//
// Copied because a fixture a porter invents encodes the same belief as the port, and the case that
// catches a bug is the one nobody would think to write. Upstream's clean list is where that pays:
// three of its cases are a nested function, function expression and arrow inside a constructor, and
// each is a different way for a walk that forgets the function boundary to be wrong.
//
// The corpus is small and it is silent on almost everything TypeScript adds, so the added cases
// carry their own reasoning and every one of them was measured against the installed ESLint 10.8.1
// rule through the Linter API, under both the default parser and `@typescript-eslint/parser`.
func TestNoConstructorReturnFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		findings   int
	}{
		// upstream, invalid
		{"class C { constructor() { return '' } }", 1},
		{"class C { constructor(a) { if (!a) { return '' } else { a() } } }", 1},

		// added by this port, each measured against the installed rule
		// a string-literal key spelled `constructor` IS a constructor; upstream reports and our parser agrees
		{"class C { \"constructor\"() { return 1 } }", 1},
		// a class expression, which upstream's corpus never writes
		{"(class { constructor() { return 1 } })", 1},
		// a class expression in a declaration
		{"var C = class { constructor() { return 1 } };", 1},
		// a derived constructor, where the substitution hazard is worst
		{"class C extends B { constructor() { super(); return 1 } }", 1},
		// TypeScript only; @typescript-eslint/parser parses it and the rule reports
		{"class C { private constructor() { return 1 } }", 1},
		// an overload signature has no body, so only the implementation reports
		{"class C { constructor(); constructor() { return 1 } }", 1},
		// the INNER constructor reports and the outer contributes nothing; this is the case that separates a walk stopping at the first function-like ancestor from one that keeps going
		{"class C { constructor() { class D { constructor() { return 1 } } } }", 1},
		// two findings from one input, which a fixture asserting one per input would get wrong
		{"class C { constructor() { if (a) { return 1 } else { return 2 } } }", 2},
		// three returns, one of them bare, so two findings
		{"class C { constructor() { switch(x) { case 1: return a; case 2: return; default: return c } } }", 2},
		// reached through a try block
		{"class C { constructor() { try { return 1 } catch(e) {} } }", 1},
		// reached through a finally block
		{"class C { constructor() { try {} finally { return 1 } } }", 1},
		// reached through a do-while
		{"class C { constructor() { do { return 1 } while(x) } }", 1},
		// reached through a for-of
		{"class C { constructor() { for (const a of b) { return a } } }", 1},
		// reached through a while with a break beside it
		{"class C { constructor() { while(x) { if (y) break; else return 1 } } }", 1},
		// reached through a labeled block
		{"class C { constructor() { lbl: { return 1 } } }", 1},
		// returning `this` is the substitution case, and it is still reported
		{"class C { constructor() { return this } }", 1},
		// an explicit undefined is still an argument
		{"class C { constructor() { return undefined } }", 1},
		// so is a void expression
		{"class C { constructor() { return void 0 } }", 1},
		// a sibling method in the same class contributes nothing
		{"class C { constructor() { return 1 } method() { return 2 } }", 1},
		// illegal source both upstream parsers refuse; ours recovers and this port reports, see the rule's doc comment
		{"class C { async constructor() { return 1 } }", 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			wantIds := make([]string, testCase.findings)
			for index := range wantIds {
				wantIds[index] = "noConstructorReturn"
			}
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoConstructorReturn, constructorReturnFile, testCase.sourceText), wantIds...)
		})
	}
}

// The clean cases are the whole discrimination.
//
// They split into families and each family is a different way to be wrong. A rule that forgets the
// function boundary reports every nested function, function expression and arrow inside a
// constructor, which is three of upstream's cases. A rule that checks the statement rather than its
// argument reports every bare `return`, which is another three. A rule keying on the name
// `constructor` rather than on the node kind reports an object literal method and a computed key.
// And a rule translating `MethodDefinition` with `kind === "constructor"` straight into
// `IsConstructorDeclaration` reports `static constructor`, which our parser calls a constructor and
// upstream does not.
func TestNoConstructorReturnStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		// upstream, valid
		"function fn() { return }",
		"function fn(kumiko) { if (kumiko) { return kumiko } }",
		"const fn = function () { return }",
		"const fn = function () { if (kumiko) { return kumiko } }",
		"const fn = () => { return }",
		"const fn = () => { if (kumiko) { return kumiko } }",
		"return 'Kumiko Oumae'", // upstream sets globalReturn for this one; our parser accepts a top-level return without it
		"class C {  }",
		"class C { constructor() {} }",
		"class C { constructor() { let v } }",
		"class C { method() { return '' } }",
		"class C { get value() { return '' } }",
		"class C { constructor(a) { if (!a) { return } else { a() } } }",
		"class C { constructor() { function fn() { return true } } }",
		"class C { constructor() { this.fn = function () { return true } } }",
		"class C { constructor() { this.fn = () => { return true } } }",
		"class C { constructor() { return } }",
		"class C { constructor() { { return } } }",

		// added by this port, each measured against the installed rule
		// an object literal method named `constructor` is a plain method; our parser gives it KindMethodDeclaration and upstream is silent
		"({ constructor() { return 1 } })",
		// a property whose key is spelled `constructor`, silent upstream
		"({ constructor: function() { return 1 } })",
		// a static method named `constructor`; our parser calls it KindConstructor, so this is the false positive the modifier guard exists for
		"class C { static constructor() { return 1 } }",
		// static and computed, silent upstream
		"class C { static ['constructor']() { return 1 } }",
		// a computed key is not a constructor; our parser agrees for free and this pins that
		"class C { ['constructor']() { return 1 } }",
		// a generator named `constructor`; our parser gives it KindMethodDeclaration
		"class C { *constructor() { return 1 } }",
		// a class static block is not function-like here, so the walk answers nil
		"class C { static { return 1 } }",
		// a field initializer arrow is its own return target
		"class C { f = () => { return 1 }; constructor() { } }",
		// a concise arrow body is not a return statement at all
		"class C { constructor() { const f = () => 1; } }",
		// a getter nested in a constructor returns from itself
		"class C { constructor() { var o = { get a() { return 1 } }; } }",
		// a method of a class nested in a constructor is its own target
		"class C { constructor() { class D { method() { return 1 } } } }",
		// a bare return with a semicolon, the same discrimination as upstream's case without one
		"class C { constructor() { return; } }",
	}

	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoConstructorReturn, constructorReturnFile, sourceText))
		})
	}
}

// The span and the rendered message, which the id fixtures above cannot see.
//
// `ExpectFindings` asserts ids and count and nothing else, so a rule pointing at the wrong bytes
// passes the whole corpus while being wrong about the only thing a reader looks at. The expected
// text is taken from the installed rule's own reported columns rather than counted by hand: slicing
// each ESLint diagnostic's `[column, endColumn)` out of the input gives the whole ReturnStatement,
// including its semicolon where the source writes one and stopping at the expression where it does
// not. That is why `return a;` and `return c` differ below in the same input. A rule reporting the
// expression instead would drop the leading `return ` from every one of these.
//
// The message is asserted against a literal typed here rather than against the rule's own constant,
// because comparing a diagnostic to the constant it was built from is an equality that moves on both
// sides under mutation and cannot fail.
func TestNoConstructorReturnSpansTheWholeStatement(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSpans  []string
	}{
		{"class C { constructor() { return '' } }", []string{"return ''"}},
		{"class C { constructor() { if (a) { return 1 } else { return 2 } } }", []string{"return 1", "return 2"}},
		{"class C { constructor() { switch(x) { case 1: return a; case 2: return; default: return c } } }", []string{"return a;", "return c"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoConstructorReturn, constructorReturnFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d diagnostics, got %d",
					len(testCase.wantSpans), len(result.Diagnostics))
			}
			for index, wantSpan := range testCase.wantSpans {
				diagnostic := result.Diagnostics[index]
				gotSpan := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != wantSpan {
					t.Fatalf("finding %d spanned %q, wanted %q", index, gotSpan, wantSpan)
				}
			}
		})
	}
}

// The message identity, asserted against literals rather than against the rule's own constants.
//
// A rule reporting the right span under the wrong id is a real defect that every span assertion
// above passes, and an id assertion comparing against the constant the rule reports with is an
// equality whose two sides move together under mutation. Both are typed out here.
func TestNoConstructorReturnReportsItsOwnMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoConstructorReturn, constructorReturnFile,
		"class C { constructor() { return 1 } }")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted 1 diagnostic, got %d", len(result.Diagnostics))
	}
	if result.Diagnostics[0].Message.Id != "noConstructorReturn" {
		t.Fatalf("reported under %q", result.Diagnostics[0].Message.Id)
	}
	wantDescription := "This returns a value from a constructor, where it is either discarded or " +
		"worse. `new C()` evaluates to the new instance and throws the returned value away, " +
		"unless that value is an object, in which case the language substitutes it for the " +
		"instance and every field this constructor assigned is lost. Neither outcome is one a " +
		"reader expects from a `new`, so return nothing and assign to `this` instead."
	if result.Diagnostics[0].Message.Description != wantDescription {
		t.Fatalf("described as %q", result.Diagnostics[0].Message.Description)
	}
}
