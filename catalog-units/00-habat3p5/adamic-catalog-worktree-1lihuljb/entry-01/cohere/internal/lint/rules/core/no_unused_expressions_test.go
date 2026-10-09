package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// unusedExpressionsFile is where the fixtures pretend to live.
//
// A `.tsx` extension rather than `.ts`, because a third of the corpus is JSX and the rest is
// TypeScript. oxc runs this whole corpus through one `.tsx` parse for the same reason, and a `.ts`
// file would read `<div />` as a type assertion rather than an element, turning five cases into
// something upstream never wrote.
const unusedExpressionsFile = "/repository/source/UnusedExpressions.tsx"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_unused_expressions.rs`,
// which merges ESLint's own corpus with typescript-eslint's extension corpus into one vector: 56
// pass and 60 fail. The snapshot records 61 diagnostics against those 60 fail inputs, so one finding
// per input is wrong for exactly one case, and that case is named at TestNoUnusedExpressionsReports
// TwiceInAClassStaticBlock below rather than averaged away.
//
// Copied because a fixture a porter invents encodes the same belief as the port, and the case that
// catches a bug is the one nobody would think to write. Two of these earned their keep during this
// port: `f(); g()` and `a?.b()?.c;`.

// TestNoUnusedExpressionsFires runs every fail case upstream lists under default options.
func TestNoUnusedExpressionsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a bare number", "0"},
		{"a bare identifier", "a"},
		{"a sequence whose tail is a literal", "f(), 0"},
		{"a literal inside a block", "{0}"},
		{"an array literal", "[]"},
		{"a short circuit onto a call, unallowed", "a && b();"},
		{"a short circuit onto a literal", "a() || false"},
		{"a short circuit onto an assignment, unallowed", "a || (b = c)"},
		{"a ternary onto a short circuit, unallowed", "a ? b() || (c = d) : e"},
		{"an untagged template", "`untagged template literal`"},
		{"a tagged template, unallowed", "tag`tagged template literal`"},
		{"a member access", "foo.bar;"},
		{"a logical negation", "!a"},
		{"a unary plus", "+a"},
		{"a string after a call is no longer a prologue", `"directive one"; f(); "directive two";`},
		{"the same inside a function body", `function foo() {"directive one"; f(); "directive two"; }`},
		{"a string inside an if block, which has no prologue", `if (0) { "not a directive"; f(); }`},
		{"a string after a declaration in a function body", `function foo() { var foo = true; "use strict"; }`},
		{"the same inside an arrow body", `var foo = () => { var foo = true; "use strict"; }`},
		{"an optional member access", "obj?.foo"},
		{"an optional access continued", "obj?.foo.bar"},
		{"an optional call whose result is then read", "obj?.foo().bar"},
		{"a class static block has no directive prologue", "class C { static { 'use strict'; } }"},
		// The typescript-eslint half of upstream's vector. Written with the leading newline and the
		// indentation upstream wrote them with, because trimming them is a rewrite and the point of
		// copying is that nothing here is my judgment.
		{"a literal as an if consequent", "\n            if (0) 0;\n                  "},
		{"a sequence whose tail is an object", "\n            f(0), {};\n                  "},
		{"a sequence whose head is an identifier", "\n            a, b();\n                  "},
		{
			"a short circuit onto a named function expression",
			"\n            a() &&\n              function namedFunctionInExpressionContext() {\n                f();\n              };\n                  ",
		},
		{"an optional access", "\n            a?.b;\n                  "},
		{"a parenthesised optional access, then read", "\n            (a?.b).c;\n                  "},
		{"an optional element access", "\n            a?.['b'];\n                  "},
		{"a parenthesised optional element access, then read", "\n            (a?.['b']).c;\n                  "},
		{"an optional call whose result is read optionally", "\n            a?.b()?.c;\n                  "},
		{"a parenthesised optional call, then read", "\n            (a?.b()).c;\n                  "},
		{"a chained optional element access", "\n            one[2]?.[3][4];\n                  "},
		{"a chained optional member access", "\n            one.two?.three.four;\n                  "},
		{
			"a string after a declaration in a module block",
			"\n            module Foo {\n              const foo = true;\n              'use strict';\n            }\n                  ",
		},
		{
			"a string after declarations in a namespace",
			"\n            namespace Foo {\n              export class Foo {}\n              export class Bar {}\n\n              'use strict';\n            }\n                  ",
		},
		{
			"a string after a declaration in a function",
			"\n            function foo() {\n              const foo = true;\n\n              'use strict';\n            }\n                  ",
		},
		{
			"a parenthesised string is not a directive",
			"function foo() {\n              const foo = true;\n              ('use strict');\n            }",
		},
		{
			"a type instantiation of a class",
			"\n            class Foo<T> {}\n            Foo<string>;\n                  ",
		},
		{"a bare type instantiation", "Map<string, string>;"},
		{
			"a declared identifier",
			"\n            declare const foo: number | undefined;\n            foo;\n                  ",
		},
		{
			"an as-expression over an identifier",
			"\n            declare const foo: number | undefined;\n            foo as any;\n                  ",
		},
		{
			"a non-null assertion over an identifier",
			"\n            declare const foo: number | undefined;\n            foo!;\n                  ",
		},
		{"an arrow with a block body holding a bare expression", "const _func = (value: number) => { value + 1; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoUnusedExpressions, unusedExpressionsFile, testCase.sourceText),
				"unusedExpression")
		})
	}
}

// TestNoUnusedExpressionsStaysSilent runs every pass case upstream lists under default options.
//
// These are the cases that catch a port, not the fail cases. `f(); g()` is the one that separates a
// rule walking every expression from a rule watching expression statements: `g()` is used, and a
// rule reading the sequence wrongly reports it. `a?.['b']?.c()` and `one[2]?.[3][4]?.()` are the
// optional-call question, which is the whole typescript-eslint divergence and which TypeScript's AST
// answers for free because it has no ChainExpression node.
func TestNoUnusedExpressionsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a function declaration is not an expression statement", "function f(){}"},
		{"an assignment", "a = b"},
		{"a construction", "new a"},
		{"an empty block", "{}"},
		{"two calls", "f(); g()"},
		{"an increment", "i++"},
		{"a call", "a()"},
		{"a delete", "delete foo.bar"},
		{"a void", "void new C"},
		{"a directive at the top of a file", `"use strict";`},
		{"two directives then a call", `"directive one"; "directive two"; f();`},
		{"a directive at the top of a function body", `function foo() {"use strict"; return true; }`},
		{"a directive at the top of an arrow body", `var foo = () => {"use strict"; return true; }`},
		{"two directives in a function body", `function foo() {"directive one"; "directive two"; f(); }`},
		{"a string used as an initialiser", `function foo() { var foo = "use strict"; return true; }`},
		{"a yield", "function* foo(){ yield 0; }"},
		{"an await of a literal", "async function foo() { await 5; }"},
		{"an await of a member access", "async function foo() { await foo.bar; }"},
		{"a dynamic import", `import("foo")`},
		{"an optional call", `func?.("foo")`},
		{"an optional member call", `obj?.foo("bar")`},
		{"a self-closing element with no option", "<div />"},
		{"a fragment with no option", "<></>"},
		{"an element used as an initialiser", "var partial = <div />"},
		// The typescript-eslint half.
		{"an optional call as a statement", "\n                  test.age?.toLocaleString();\n                "},
		{"a parenthesised optional access used as an initialiser", "\n                  let a = (a?.b).c;\n                "},
		{"an optional element access used as an initialiser", "\n                  let b = a?.['b'];\n                "},
		{"a chained optional access used as an initialiser", "\n                  let c = one[2]?.[3][4];\n                "},
		{"a chained optional call as a statement", "\n                  one[2]?.[3][4]?.();\n                "},
		{"an optional element access then an optional call", "\n                  a?.['b']?.c();\n                "},
		{"a directive at the top of a module block", "\n                  module Foo {\n                    'use strict';\n                  }\n                "},
		{
			"a directive at the top of a namespace",
			"\n                  namespace Foo {\n                    'use strict';\n\n                    export class Foo {}\n                    export class Bar {}\n                  }\n                ",
		},
		{
			"a directive at the top of a function",
			"\n                  function foo() {\n                    'use strict';\n\n                    return null;\n                  }\n                ",
		},
		{"a dynamic import as a statement", "\n                  import('./foo');\n                "},
		{"a dynamic import then a call", "\n                  import('./foo').then(() => {});\n                "},
		{
			"a construction with type arguments",
			"\n                  class Foo<T> {}\n                  new Foo<string>();\n                ",
		},
		{"an arrow with an expression body is no statement at all", "const _func = (value: number) => value + 1;"},
		{
			"a satisfies expression in a switch default",
			"\n            type FooBarBaz = 'foo' | 'bar' | 'baz';\n            export function satisfiesTest(c: FooBarBaz): string {\n                switch(c) {\n                    case 'foo':\n                        return 'foo';\n                    case 'bar':\n                        return 'bar';\n                    default:\n                        c satisfies never;\n                        return '';\n                }\n            }\n                ",
		},
		{"a bare satisfies expression", "value satisfies number;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoUnusedExpressions, unusedExpressionsFile, testCase.sourceText))
		})
	}
}

// TestNoUnusedExpressionsUnderOptions runs every case upstream configures, on both verdicts.
//
// Kept in one table rather than split across the two above, because the interesting property of
// these cases is that the same source flips verdict with the option: `a ? b() : c()` passes under
// allowTernary and fails under allowShortCircuit, and reading them apart loses that.
func TestNoUnusedExpressionsUnderOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    NoUnusedExpressionsOptions
		findings   int
	}{
		// allowShortCircuit
		{"a short circuit onto a call", "a && a()", NoUnusedExpressionsOptions{AllowShortCircuit: true}, 0},
		{"a short circuit onto an assignment", "a() || (b = c)", NoUnusedExpressionsOptions{AllowShortCircuit: true}, 0},
		{
			"a short circuit onto an await",
			"async function foo() { bar && await baz; }",
			NoUnusedExpressionsOptions{AllowShortCircuit: true}, 0,
		},
		{"a short circuit onto an optional call", "foo && foo?.();", NoUnusedExpressionsOptions{AllowShortCircuit: true}, 0},
		{"a short circuit onto a dynamic import", "foo && import('./foo');", NoUnusedExpressionsOptions{AllowShortCircuit: true}, 0},
		{"a short circuit onto an identifier", "a || b", NoUnusedExpressionsOptions{AllowShortCircuit: true}, 1},
		{"a short circuit onto an identifier, left used", "a() && b", NoUnusedExpressionsOptions{AllowShortCircuit: true}, 1},
		{"a short circuit onto an optional access", "foo && foo?.bar;", NoUnusedExpressionsOptions{AllowShortCircuit: true}, 1},
		{"a ternary under the short-circuit option only", "a ? b() : c()", NoUnusedExpressionsOptions{AllowShortCircuit: true}, 1},

		// allowTernary
		{"a ternary onto two calls", "a ? b() : c()", NoUnusedExpressionsOptions{AllowTernary: true}, 0},
		{
			"a ternary onto two awaits",
			"async function foo() { foo ? await bar : await baz; }",
			NoUnusedExpressionsOptions{AllowTernary: true}, 0,
		},
		{
			"a ternary onto two dynamic imports",
			"foo ? import('./foo') : import('./bar');",
			NoUnusedExpressionsOptions{AllowTernary: true}, 0,
		},
		{"a ternary with a literal alternate", "a ? b : 0", NoUnusedExpressionsOptions{AllowTernary: true}, 1},
		{"a ternary with an identifier consequent", "a ? b : c()", NoUnusedExpressionsOptions{AllowTernary: true}, 1},
		{"a ternary onto an optional access", "foo ? foo?.bar : bar.baz;", NoUnusedExpressionsOptions{AllowTernary: true}, 1},
		{"a short circuit under the ternary option only", "a && b()", NoUnusedExpressionsOptions{AllowTernary: true}, 1},

		// both
		{
			"a ternary onto a short circuit onto an assignment",
			"a ? b() || (c = d) : e()",
			NoUnusedExpressionsOptions{AllowShortCircuit: true, AllowTernary: true}, 0,
		},

		// allowTaggedTemplates
		{"a tagged template", "tag`tagged template literal`", NoUnusedExpressionsOptions{AllowTaggedTemplates: true}, 0},
		{
			"a call is unaffected by the tagged-template option",
			"shouldNotBeAffectedByAllowTemplateTagsOption()",
			NoUnusedExpressionsOptions{AllowTaggedTemplates: true}, 0,
		},
		{"an untagged template is unaffected", "`untagged template literal`", NoUnusedExpressionsOptions{AllowTaggedTemplates: true}, 1},
		{"an untagged template with the option off", "`untagged template literal`", NoUnusedExpressionsOptions{AllowTaggedTemplates: false}, 1},
		{"a tagged template with the option off", "tag`tagged template literal`", NoUnusedExpressionsOptions{AllowTaggedTemplates: false}, 1},

		// enforceForJSX
		{"an element used as an initialiser is still not a statement", "var partial = <div />", NoUnusedExpressionsOptions{EnforceForJSX: true}, 0},
		{"a fragment used as an initialiser", "var partial = <></>", NoUnusedExpressionsOptions{EnforceForJSX: true}, 0},
		{"a self-closing element as a statement", "<div />", NoUnusedExpressionsOptions{EnforceForJSX: true}, 1},
		{"a fragment as a statement", "<></>", NoUnusedExpressionsOptions{EnforceForJSX: true}, 1},

		// ignoreDirectives. Every one of these already passes without the option, because a real
		// directive prologue is exempt unconditionally in both upstreams. They are here verbatim
		// anyway: upstream wrote them, and a port under which they changed verdict would be wrong.
		{"a directive under the option", `"use strict";`, NoUnusedExpressionsOptions{IgnoreDirectives: true}, 0},
		{"two directives then a call, under the option", `"directive one"; "directive two"; f();`, NoUnusedExpressionsOptions{IgnoreDirectives: true}, 0},
		{"a function-body directive under the option", `function foo() {"use strict"; return true; }`, NoUnusedExpressionsOptions{IgnoreDirectives: true}, 0},
		{"two function-body directives under the option", `function foo() {"directive one"; "directive two"; f(); }`, NoUnusedExpressionsOptions{IgnoreDirectives: true}, 0},
		{"an identifier is not a directive", "foo;", NoUnusedExpressionsOptions{IgnoreDirectives: true}, 1},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoUnusedExpressions, unusedExpressionsFile,
				testCase.sourceText, testCase.options)
			expected := make([]string, testCase.findings)
			for index := range expected {
				expected[index] = "unusedExpression"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestNoUnusedExpressionsReportsTwiceInAClassStaticBlock is the corpus discrepancy, pinned.
//
// The extractor reports 61 diagnostics against 60 fail inputs, and this is the one input that
// carries two. It matters beyond arithmetic: a class static block has no directive prologue, so
// *both* strings report, where the same two lines at the top of a function body would report
// neither. A port whose directive carve-out asked only "is this a string statement" would report
// zero here and pass every other fail case in the corpus.
func TestNoUnusedExpressionsReportsTwiceInAClassStaticBlock(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoUnusedExpressions, unusedExpressionsFile,
			"class C { static {\n            'foo'\n            'bar'\n             } }"),
		"unusedExpression", "unusedExpression")
}

// TestNoUnusedExpressionsPointsAtTheStatement asserts where every finding lands.
//
// ExpectFindings checks message ids and count and nothing else, so a rule reporting the right
// judgment at the wrong span passes the whole table above. That is not hypothetical here: the
// obvious span is the node's own Pos(), and a TypeScript node's Pos() includes its leading trivia,
// so `"directive one"; f(); "directive two";` would report a span beginning at the space before
// `"directive two"`. Every span below is upstream's, read off the snapshot rather than off this
// rule's output.
//
// The statement rather than the expression is the reported node, which is why the trailing
// semicolon is inside every span that has one.
func TestNoUnusedExpressionsPointsAtTheStatement(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		reported   []string
	}{
		{"a whole statement including its semicolon", "foo.bar;", []string{"foo.bar;"}},
		{"a statement with no semicolon", "0", []string{"0"}},
		{"the literal inside a block, not the block", "{0}", []string{"0"}},
		{"the whole sequence", "f(), 0", []string{"f(), 0"}},
		{
			"the trailing string only, with no leading space",
			`"directive one"; f(); "directive two";`,
			[]string{`"directive two";`},
		},
		{
			"the string inside the if block only",
			`if (0) { "not a directive"; f(); }`,
			[]string{`"not a directive";`},
		},
		{
			"the string after the declaration only",
			`function foo() { var foo = true; "use strict"; }`,
			[]string{`"use strict";`},
		},
		{
			"both strings in a static block, each on its own",
			"class C { static {\n            'foo'\n            'bar'\n             } }",
			[]string{"'foo'", "'bar'"},
		},
		{
			"the parenthesised string including its parentheses",
			"function foo() {\n              const foo = true;\n              ('use strict');\n            }",
			[]string{"('use strict');"},
		},
		{"the whole optional chain", "obj?.foo().bar", []string{"obj?.foo().bar"}},
		{"the whole type instantiation", "Map<string, string>;", []string{"Map<string, string>;"}},
		{"the consequent of the if, not the if", "if (0) 0;", []string{"0;"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedExpressions, unusedExpressionsFile, testCase.sourceText)
			if len(result.Diagnostics) != len(testCase.reported) {
				t.Fatalf("wanted %d diagnostics, got %d", len(testCase.reported), len(result.Diagnostics))
			}
			for index, want := range testCase.reported {
				found := result.Diagnostics[index]
				got := testCase.sourceText[found.Range.Pos():found.Range.End()]
				if got != want {
					t.Errorf("diagnostic %d reported %q, wanted %q", index, got, want)
				}
			}
		})
	}
}

// Cases written from reading our code and the two upstreams, rather than copied.
//
// The imported corpus is a floor. Each case below covers a discrimination this port has to make
// that no upstream case reaches, and each says why it exists.

// TestNoUnusedExpressionsSeparatesBinaryOperators covers the kind collision the corpus hides.
//
// oxc and ESLint both get four distinct node types here: AssignmentExpression, SequenceExpression,
// LogicalExpression, BinaryExpression. TypeScript gives all four the single kind
// KindBinaryExpression and puts the distinction in the operator token, so a port that switched on
// kind alone would treat `a = b` exactly like `a + b` and report every assignment in the tree.
//
// Upstream's corpus does cover `a = b` and `a + b` separately, so that much is caught. What it does
// not cover is the compound and logical assignment operators, which are assignments spelled with a
// logical operator and which a port discriminating by "is the operator token logical" would report.
func TestNoUnusedExpressionsSeparatesBinaryOperators(t *testing.T) {
	t.Parallel()

	silent := []string{
		"a += b;",
		"a &&= b;",
		"a ||= b;",
		"a ??= b;",
		"a **= b;",
		"a >>>= b;",
	}
	for _, sourceText := range silent {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoUnusedExpressions, unusedExpressionsFile, sourceText))
		})
	}

	fires := []string{
		"a ?? b;",
		"a ** b;",
		"a >>> b;",
		"a in b;",
		"a instanceof b;",
	}
	for _, sourceText := range fires {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoUnusedExpressions, unusedExpressionsFile, sourceText),
				"unusedExpression")
		})
	}
}

// TestNoUnusedExpressionsTreatsNullishAsShortCircuit pins `??` under allowShortCircuit.
//
// ESLint's LogicalExpression handler covers `&&`, `||` and `??` alike, and oxc's LogicalExpression
// arm does the same. So `a ?? b()` is allowed under the option for the same reason `a || b()` is.
// No upstream case configures the option over `??`, so the behaviour is inherited rather than
// tested, and a port using a helper that answered only for `&&` and `||` would differ here with
// nothing to say so.
func TestNoUnusedExpressionsTreatsNullishAsShortCircuit(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoUnusedExpressions, unusedExpressionsFile,
		"a ?? b();", NoUnusedExpressionsOptions{AllowShortCircuit: true}))
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoUnusedExpressions, unusedExpressionsFile,
		"a ?? b;", NoUnusedExpressionsOptions{AllowShortCircuit: true}), "unusedExpression")
}

// TestNoUnusedExpressionsRecursesThroughAllowedForms covers nesting the corpus stops one level short
// of.
//
// Upstream configures `a ? b() || (c = d) : e()` under both options, which is a ternary holding a
// short circuit. It never nests a short circuit inside a short circuit, nor a ternary inside a
// ternary, and both recursions run through the same call in this port. A port that checked the
// operand's kind instead of recursing would pass upstream's case and fail these.
func TestNoUnusedExpressionsRecursesThroughAllowedForms(t *testing.T) {
	t.Parallel()

	both := NoUnusedExpressionsOptions{AllowShortCircuit: true, AllowTernary: true}

	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoUnusedExpressions, unusedExpressionsFile,
		"a && b && c();", NoUnusedExpressionsOptions{AllowShortCircuit: true}))
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoUnusedExpressions, unusedExpressionsFile,
		"a && b && c;", NoUnusedExpressionsOptions{AllowShortCircuit: true}), "unusedExpression")

	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoUnusedExpressions, unusedExpressionsFile,
		"a ? b ? c() : d() : e();", NoUnusedExpressionsOptions{AllowTernary: true}))
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoUnusedExpressions, unusedExpressionsFile,
		"a ? b ? c() : d : e();", NoUnusedExpressionsOptions{AllowTernary: true}), "unusedExpression")

	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoUnusedExpressions, unusedExpressionsFile,
		"a ? b && c() : d();", both))
}

// TestNoUnusedExpressionsChecksBothTernaryBranches pins that one bad branch is enough.
//
// ESLint returns `isDisallowed(consequent) || isDisallowed(alternate)` and oxc returns the same
// disjunction with the operands written the other way round. Upstream's corpus covers a bad
// alternate (`a ? b : 0`) and a bad consequent (`a ? b : c()`), so both arms are reached, but only
// under allowTernary alone. A port that recursed into one branch and took the other on trust would
// need both of these to notice, and they are cheap to state directly.
func TestNoUnusedExpressionsChecksBothTernaryBranches(t *testing.T) {
	t.Parallel()

	options := NoUnusedExpressionsOptions{AllowTernary: true}
	for _, sourceText := range []string{"a ? b() : c;", "a ? b : c();"} {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunWithOptions(t, NoUnusedExpressions, unusedExpressionsFile, sourceText, options),
				"unusedExpression")
		})
	}
}

// TestNoUnusedExpressionsUnwrapsTypeOnlyWrappers covers the TypeScript wrappers by kind.
//
// Upstream covers `foo as any` and `foo!` over an identifier, both of which report, and `value
// satisfies number`, which does not. What it does not cover is the same wrappers over a *call*,
// which is the direction that tells a recursion apart from a blanket verdict: `(a()) as any` must
// stay silent because the call inside is used, and a port hard-coding "an as-expression is unused"
// would report it. `<any>foo` is the older assertion spelling, which has its own node kind and
// which upstream reaches only through ESLint's TSTypeAssertion arm with no case exercising it.
func TestNoUnusedExpressionsUnwrapsTypeOnlyWrappers(t *testing.T) {
	t.Parallel()

	silent := []string{
		"a() as any;",
		"a()!;",
		"(a() as any)!;",
		"a() satisfies unknown;",
		"(a());",
	}
	for _, sourceText := range silent {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoUnusedExpressions, unusedExpressionsFile, sourceText))
		})
	}

	fires := []string{
		"(a)!;",
		"(a as any)!;",
		"(a);",
		"((a));",
	}
	for _, sourceText := range fires {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoUnusedExpressions, unusedExpressionsFile, sourceText),
				"unusedExpression")
		})
	}
}

// TestNoUnusedExpressionsDivergesFromOxcOnSatisfies states a divergence rather than hiding it.
//
// oxc lists `Expression::TSSatisfiesExpression` among the forms that are always *used*, alongside
// calls and assignments, so `value satisfies never` passes there whatever it wraps. ESLint has no
// `TSSatisfiesExpression` arm at all, so it falls through to `alwaysFalse` and reaches the same
// verdict by a different road. Both agree on every case upstream wrote.
//
// This port follows both: a satisfies expression is used, full stop, without recursing into its
// operand. The divergence that would exist if it recursed is `0 satisfies number;`, which reports
// under a recursing port and passes under both upstreams. Pinned here so a later reader changing
// this to recurse finds out immediately.
func TestNoUnusedExpressionsDivergesFromOxcOnSatisfies(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t,
		rule_testing.Run(t, NoUnusedExpressions, unusedExpressionsFile, "0 satisfies number;"))
}

// TestNoUnusedExpressionsCarvesOutOnlyRealDirectivePrologues is the carve-out, stated directly.
//
// This is the one piece of the rule that has no counterpart in oxc's rule file, because oxc's parser
// hoists a directive prologue into a separate `directives` vector on the enclosing body, so a real
// directive never reaches oxc's ExpressionStatement listener and its `ignore_directives` option is
// unreachable code. TypeScript's parser leaves directives as ordinary expression statements, so the
// exemption has to be written here or every `'use client'` at the top of a component file reports.
//
// The shape is ESLint's isTopLevelExpressionStatement: the statement's container must be a source
// file, a module block, or a block belonging to something function-like, and the statement must be
// in the unbroken leading run of string statements.
//
// `ast.IsPrologueDirective` is deliberately not used, having been probed rather than trusted: it is
// `Kind == ExpressionStatement && Expression().Kind == StringLiteral` and nothing more, so it
// answers true for every string statement anywhere, including the class-static-block cases upstream
// reports and the after-a-declaration cases upstream reports. Building on it would have silenced
// eight fail cases.
func TestNoUnusedExpressionsCarvesOutOnlyRealDirectivePrologues(t *testing.T) {
	t.Parallel()

	silent := []string{
		// A prologue at the top of each container that has one.
		"'use client';\nexport const a = 1;",
		"function f() { 'use strict'; return 1; }",
		"const f = function () { 'use strict'; return 1; };",
		"const f = () => { 'use strict'; return 1; };",
		"class C { m() { 'use strict'; return 1; } }",
		"class C { get m() { 'use strict'; return 1; } }",
		"class C { constructor() { 'use strict'; } }",
		"namespace N { 'use strict'; export const a = 1; }",
		// A run of several, all still prologue.
		"'a';\n'b';\n'c';\nf();",
	}
	for _, sourceText := range silent {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoUnusedExpressions, unusedExpressionsFile, sourceText))
		})
	}

	fires := []string{
		// A block with no function attached has no prologue, even at its very top.
		"{ 'use strict'; }",
		"if (a) { 'use strict'; }",
		"for (;;) { 'use strict'; }",
		"while (a) { 'use strict'; }",
		"try { 'use strict'; } catch {}",
		"label: { 'use strict'; }",
		"switch (a) { case 1: 'use strict'; }",
		// A class static block is a block whose parent is not function-like, which is the
		// distinction that makes upstream's two static-block cases report.
		"class C { static { 'use strict'; } }",
		// A template literal is not a string literal, so it is never a directive. Runtime agrees:
		// `` `use strict` `` at the top of a function does not enable strict mode.
		"function f() { `use strict`; return 1; }",
		// A number is not a directive either, and it breaks the run for what follows.
		"0;\n'use strict';",
	}
	for _, sourceText := range fires {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUnusedExpressions, unusedExpressionsFile, sourceText)
			if len(result.Diagnostics) == 0 {
				t.Fatalf("wanted at least one diagnostic for %q, got none", sourceText)
			}
		})
	}
}

// TestNoUnusedExpressionsIgnoreDirectivesIsInertHere records an option that changes nothing.
//
// ESLint has two directive checks. `astUtils.isDirective` reads the parser's own `.directive`
// property and is applied unconditionally; the rule's private `isDirective` is a structural
// re-derivation applied only under `ignoreDirectives`. In espree and typescript-eslint the two agree
// on every input, because the parser sets `.directive` on exactly the leading run of string
// statements in a Program, a function body, or a TSModuleBlock, which is what the structural version
// recomputes. So the option is a no-op upstream, and the whole of upstream's `ignoreDirectives`
// corpus is cases that already pass without it.
//
// This port has one structural check rather than two, which makes the option genuinely inert, and
// that is a deliberate simplification rather than an oversight. This test is what would fail if
// somebody later wired the option to something, and it runs every string-statement case in the file
// through both settings to show they agree.
func TestNoUnusedExpressionsIgnoreDirectivesIsInertHere(t *testing.T) {
	t.Parallel()

	sources := []string{
		`"use strict";`,
		`"directive one"; f(); "directive two";`,
		`function foo() { var foo = true; "use strict"; }`,
		"class C { static { 'use strict'; } }",
		"namespace N { 'use strict'; }",
		"{ 'use strict'; }",
		"foo;",
		"0;",
	}
	for _, sourceText := range sources {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			off := rule_testing.RunWithOptions(t, NoUnusedExpressions, unusedExpressionsFile,
				sourceText, NoUnusedExpressionsOptions{IgnoreDirectives: false})
			on := rule_testing.RunWithOptions(t, NoUnusedExpressions, unusedExpressionsFile,
				sourceText, NoUnusedExpressionsOptions{IgnoreDirectives: true})
			if len(off.Diagnostics) != len(on.Diagnostics) {
				t.Errorf("ignoreDirectives changed the verdict on %q: %d without, %d with",
					sourceText, len(off.Diagnostics), len(on.Diagnostics))
			}
		})
	}
}

// TestNoUnusedExpressionsSeparatesUnaryFormsByKind covers the unary split TypeScript makes.
//
// ESLint has one UnaryExpression arm reading `node.operator`, and oxc has one UnaryExpression arm
// reading `unary_expression.operator`, both excusing exactly `void` and `delete`. TypeScript splits
// that single node four ways: `void x` and `delete x` and `typeof x` each get their own kind, and
// only `!x`, `+x`, `-x`, `~x` remain on KindPrefixUnaryExpression, which also carries `++x` and
// `--x` that upstream calls an update rather than a unary.
//
// So the one upstream arm becomes four arms here, and three of the four splits are untested by the
// corpus: it covers `!a` and `+a` firing and `void new C` and `delete foo.bar` staying silent, but
// never `typeof`, and never a prefix increment. A mutation flipping the typeof verdict survived the
// whole corpus before these cases existed.
func TestNoUnusedExpressionsSeparatesUnaryFormsByKind(t *testing.T) {
	t.Parallel()

	fires := []string{"typeof a;", "!a;", "+a;", "-a;", "~a;"}
	for _, sourceText := range fires {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, NoUnusedExpressions, unusedExpressionsFile, sourceText),
				"unusedExpression")
		})
	}

	silent := []string{"void 0;", "delete a.b;", "++a;", "--a;", "a++;", "a--;"}
	for _, sourceText := range silent {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, NoUnusedExpressions, unusedExpressionsFile, sourceText))
		})
	}
}

// TestNoUnusedExpressionsLeavesUnrecognisedFormsAlone pins the fallthrough arm's verdict.
//
// Both upstreams default an unknown node to *allowed*: ESLint's `Checker.isDisallowed` falls back to
// `alwaysFalse`, and oxc's match has no catch-all because its Expression enum is closed. So a form
// this port has never been taught goes unreported rather than reported on a guess, which is the
// right direction for a correctness rule: a false negative is a rule that has not caught up, and a
// false positive is a rule the tree has to be edited around.
//
// That arm is reachable, which was measured rather than assumed. A bare `#x;` parses to a statement
// whose expression is a lone KindPrivateIdentifier, a kind no arm above names, and it was the only
// form out of roughly seventy spellings probed that reached the fallthrough. Without this case a
// mutation flipping the default to `return true` survives the entire corpus, which is how it was
// found: the mutant compiled, changed bytes, and nothing noticed.
func TestNoUnusedExpressionsLeavesUnrecognisedFormsAlone(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t,
		rule_testing.Run(t, NoUnusedExpressions, unusedExpressionsFile, "#x;"))
}

// TestNoUnusedExpressionsReadsOptionsFromJSON pins the wire names.
//
// The struct fields are Go-cased and the config is JSON, so a missing or misspelled tag reads as the
// option being absent, which is silent: the rule runs with defaults and every fires case still
// fires. The five names below are ESLint's meta.schema verbatim, which is the authoritative surface
// and which the inventory got wrong for this rule, recording "options": "no" for a rule with five.
func TestNoUnusedExpressionsReadsOptionsFromJSON(t *testing.T) {
	t.Parallel()

	decoded, err := rule.DecodeOptionsInto[NoUnusedExpressionsOptions]()([]byte(
		`{"allowShortCircuit":true,"allowTernary":true,"allowTaggedTemplates":true,` +
			`"enforceForJSX":true,"ignoreDirectives":true}`))
	if err != nil {
		t.Fatalf("decoding the full option object failed: %v", err)
	}
	options, ok := decoded.(NoUnusedExpressionsOptions)
	if !ok {
		t.Fatalf("decoded to %T rather than NoUnusedExpressionsOptions", decoded)
	}
	if !options.AllowShortCircuit || !options.AllowTernary || !options.AllowTaggedTemplates ||
		!options.EnforceForJSX || !options.IgnoreDirectives {
		t.Errorf("some option did not survive the round trip: %+v", options)
	}
}
