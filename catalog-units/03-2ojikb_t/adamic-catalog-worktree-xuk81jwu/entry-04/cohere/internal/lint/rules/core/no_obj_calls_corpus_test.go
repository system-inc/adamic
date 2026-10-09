package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * ESLint 10.8.1's no-obj-calls, run through the installed rule under @typescript-eslint/parser over
 * its test corpus and over the shapes the shelf's reference tracker has to follow (#jjfa7qb), with
 * every finding's id and span written here as ESLint answered.
 *
 * The corpus rows are the registry's corpus file, trimmed, less the 20 ESLint runs under configured
 * globals: several pin ecmaVersion 5, where Atomics and Intl are not globals yet, and the others
 * declare browser globals, a configuration cohere has no counterpart for. The edge rows add aliases
 * through assignments, defaults and object patterns, every pass-through expression, constant computed
 * keys, written globals, and the cycles the tracker must leave.
 *
 * Three edges are left out on purpose. `window.JSON()` and `self.Math()` with no DOM types: ESLint
 * without browser globals does not know `window`, and here an undeclared name is the runtime's global.
 * And `x += JSON`, which ESLint follows into x and the tracker does not, since x then holds a string.
 * Their own tests say what the rule does instead.
 */
func TestNoObjCallsAgreesWithESLint(t *testing.T) {
	t.Parallel()

	cases := []struct {
		origin string
		code   string
		want   []string
	}{
		{"corpus", "var x = Math;\n", []string{}},
		{"corpus", "var x = Math.random();\n", []string{}},
		{"corpus", "var x = Math.PI;\n", []string{}},
		{"corpus", "var x = foo.Math();\n", []string{}},
		{"corpus", "var x = new foo.Math();\n", []string{}},
		{"corpus", "var x = new Math.foo;\n", []string{}},
		{"corpus", "var x = new Math.foo();\n", []string{}},
		{"corpus", "JSON.parse(foo)\n", []string{}},
		{"corpus", "new JSON.parse\n", []string{}},
		{"corpus", "Reflect.get(foo, 'x')\n", []string{}},
		{"corpus", "new Reflect.foo(a, b)\n", []string{}},
		{"corpus", "Atomics.load(foo, 0)\n", []string{}},
		{"corpus", "new Atomics.foo()\n", []string{}},
		{"corpus", "new Intl.Segmenter()\n", []string{}},
		{"corpus", "Intl.foo()\n", []string{}},
		{"corpus", "Temporal.Now.instant()\n", []string{}},
		{"corpus", "new Temporal.Instant(0n)\n", []string{}},
		{"corpus", "globalThis.Math();\n", []string{"unexpectedCall globalThis.Math()"}},
		{"corpus", "var x = globalThis.Math();\n", []string{"unexpectedCall globalThis.Math()"}},
		{"corpus", "f(globalThis.Math());\n", []string{"unexpectedCall globalThis.Math()"}},
		{"corpus", "globalThis.Math().foo;\n", []string{"unexpectedCall globalThis.Math()"}},
		{"corpus", "var x = globalThis.JSON();\n", []string{"unexpectedCall globalThis.JSON()"}},
		{"corpus", "x = globalThis.JSON(str);\n", []string{"unexpectedCall globalThis.JSON(str)"}},
		{"corpus", "globalThis.Math( globalThis.JSON() );\n", []string{"unexpectedCall globalThis.Math( globalThis.JSON() )", "unexpectedCall globalThis.JSON()"}},
		{"corpus", "var x = globalThis.Reflect();\n", []string{"unexpectedCall globalThis.Reflect()"}},
		{"corpus", "/*globals Reflect: true*/ globalThis.Reflect();\n", []string{"unexpectedCall globalThis.Reflect()"}},
		{"corpus", "var x = globalThis.Atomics();\n", []string{"unexpectedCall globalThis.Atomics()"}},
		{"corpus", "var x = globalThis.Intl();\n", []string{"unexpectedCall globalThis.Intl()"}},
		{"corpus", "const x = globalThis.Temporal();\n", []string{"unexpectedCall globalThis.Temporal()"}},
		{"corpus", "/*globals Math: off*/ Math();\n", []string{"unexpectedCall Math()"}},
		{"corpus", "/*globals Math: off*/ new Math();\n", []string{"unexpectedCall new Math()"}},
		{"corpus", "Reflect();\n", []string{"unexpectedCall Reflect()"}},
		{"corpus", "Atomics();\n", []string{"unexpectedCall Atomics()"}},
		{"corpus", "new Reflect();\n", []string{"unexpectedCall new Reflect()"}},
		{"corpus", "new Atomics();\n", []string{"unexpectedCall new Atomics()"}},
		{"corpus", "Intl()\n", []string{"unexpectedCall Intl()"}},
		{"corpus", "new Intl()\n", []string{"unexpectedCall new Intl()"}},
		{"corpus", "Temporal();\n", []string{"unexpectedCall Temporal()"}},
		{"corpus", "new Temporal();\n", []string{"unexpectedCall new Temporal()"}},
		{"corpus", "var Math; Math();\n", []string{}},
		{"corpus", "var Math; new Math();\n", []string{}},
		{"corpus", "let JSON; JSON();\n", []string{}},
		{"corpus", "let JSON; new JSON();\n", []string{}},
		{"corpus", "if (foo) { const Reflect = 1; Reflect(); }\n", []string{}},
		{"corpus", "if (foo) { const Reflect = 1; new Reflect(); }\n", []string{}},
		{"corpus", "function foo(Math) { Math(); }\n", []string{}},
		{"corpus", "function foo(JSON) { new JSON(); }\n", []string{}},
		{"corpus", "function foo(Atomics) { Atomics(); }\n", []string{}},
		{"corpus", "function foo() { if (bar) { let Atomics; if (baz) { new Atomics(); } } }\n", []string{}},
		{"corpus", "function foo() { var JSON; JSON(); }\n", []string{}},
		{"corpus", "function foo(Intl) { Intl(); }\n", []string{}},
		{"corpus", "if (foo) { const Intl = 1; Intl(); }\n", []string{}},
		{"corpus", "if (foo) { const Intl = 1; new Intl(); }\n", []string{}},
		{"corpus", "if (foo) { const Temporal = 1; Temporal(); }\n", []string{}},
		{"corpus", "if (foo) { const Temporal = 1; new Temporal(); }\n", []string{}},
		{"corpus", "Math();\n", []string{"unexpectedCall Math()"}},
		{"corpus", "var x = Math();\n", []string{"unexpectedCall Math()"}},
		{"corpus", "f(Math());\n", []string{"unexpectedCall Math()"}},
		{"corpus", "Math().foo;\n", []string{"unexpectedCall Math()"}},
		{"corpus", "new Math;\n", []string{"unexpectedCall new Math"}},
		{"corpus", "new Math();\n", []string{"unexpectedCall new Math()"}},
		{"corpus", "new Math(foo);\n", []string{"unexpectedCall new Math(foo)"}},
		{"corpus", "new Math().foo;\n", []string{"unexpectedCall new Math()"}},
		{"corpus", "(new Math).foo();\n", []string{"unexpectedCall new Math"}},
		{"corpus", "var x = JSON();\n", []string{"unexpectedCall JSON()"}},
		{"corpus", "x = JSON(str);\n", []string{"unexpectedCall JSON(str)"}},
		{"corpus", "var x = new JSON();\n", []string{"unexpectedCall new JSON()"}},
		{"corpus", "Math( JSON() );\n", []string{"unexpectedCall Math( JSON() )", "unexpectedCall JSON()"}},
		{"corpus", "var x = Reflect();\n", []string{"unexpectedCall Reflect()"}},
		{"corpus", "var x = new Reflect();\n", []string{"unexpectedCall new Reflect()"}},
		{"corpus", "/*globals Reflect: true*/ Reflect();\n", []string{"unexpectedCall Reflect()"}},
		{"corpus", "/*globals Reflect: true*/ new Reflect();\n", []string{"unexpectedCall new Reflect()"}},
		{"corpus", "var x = Atomics();\n", []string{"unexpectedCall Atomics()"}},
		{"corpus", "var x = new Atomics();\n", []string{"unexpectedCall new Atomics()"}},
		{"corpus", "var x = Intl();\n", []string{"unexpectedCall Intl()"}},
		{"corpus", "var x = new Intl();\n", []string{"unexpectedCall new Intl()"}},
		{"corpus", "/*globals Intl: true*/ Intl();\n", []string{"unexpectedCall Intl()"}},
		{"corpus", "/*globals Intl: true*/ new Intl();\n", []string{"unexpectedCall new Intl()"}},
		{"corpus", "/* global Temporal */ Temporal();\n", []string{"unexpectedCall Temporal()"}},
		{"corpus", "/* global Temporal */ new Temporal();\n", []string{"unexpectedCall new Temporal()"}},
		{"corpus", "var x = new globalThis.Math();\n", []string{"unexpectedCall new globalThis.Math()"}},
		{"corpus", "new globalThis.Math().foo;\n", []string{"unexpectedCall new globalThis.Math()"}},
		{"corpus", "var x = new globalThis.Reflect;\n", []string{"unexpectedCall new globalThis.Reflect"}},
		{"corpus", "var x = new globalThis.Intl;\n", []string{"unexpectedCall new globalThis.Intl"}},
		{"corpus", "var foo = bar ? baz: JSON; foo();\n", []string{"unexpectedRefCall foo()"}},
		{"corpus", "var foo = bar ? baz: JSON; new foo();\n", []string{"unexpectedRefCall new foo()"}},
		{"corpus", "var foo = bar ? baz: globalThis.JSON; foo();\n", []string{"unexpectedRefCall foo()"}},
		{"corpus", "var foo = bar ? baz: globalThis.JSON; new foo();\n", []string{"unexpectedRefCall new foo()"}},
		{"corpus", "const foo = bar ? baz: globalThis.Temporal; new foo();\n", []string{"unexpectedRefCall new foo()"}},
		{"corpus", "var x = globalThis?.Reflect();\n", []string{"unexpectedCall globalThis?.Reflect()"}},
		{"corpus", "var x = (globalThis?.Reflect)();\n", []string{"unexpectedCall (globalThis?.Reflect)()"}},
		{"edge", "let f; f = JSON; f();\n", []string{"unexpectedRefCall f()"}},
		{"edge", "(f = JSON)();\n", []string{"unexpectedRefCall (f = JSON)()"}},
		{"edge", "function g(a = JSON) { a(); }\n", []string{"unexpectedRefCall a()"}},
		{"edge", "const { JSON: j } = globalThis; j();\n", []string{"unexpectedRefCall j()"}},
		{"edge", "const { j = JSON } = {}; j();\n", []string{"unexpectedRefCall j()"}},
		{"edge", "const [a = Math] = []; a();\n", []string{"unexpectedRefCall a()"}},
		{"edge", "const [a] = [Math]; a();\n", []string{}},
		{"edge", "JSON = 1; JSON();\n", []string{}},
		{"edge", "let x = JSON; x = 1; x();\n", []string{"unexpectedRefCall x()"}},
		{"edge", "globalThis[\"JSON\"]();\n", []string{"unexpectedCall globalThis[\"JSON\"]()"}},
		{"edge", "globalThis[`Math`]();\n", []string{"unexpectedCall globalThis[`Math`]()"}},
		{"edge", "globalThis[\"JS\" + \"ON\"]();\n", []string{"unexpectedRefCall globalThis[\"JS\" + \"ON\"]()"}},
		{"edge", "(a ? JSON : Math)();\n", []string{"unexpectedRefCall (a ? JSON : Math)()", "unexpectedRefCall (a ? JSON : Math)()"}},
		{"edge", "(0, JSON)();\n", []string{"unexpectedRefCall (0, JSON)()"}},
		{"edge", "(a || JSON)();\n", []string{"unexpectedRefCall (a || JSON)()"}},
		{"edge", "(JSON && a)();\n", []string{"unexpectedRefCall (JSON && a)()"}},
		{"edge", "(a ?? Reflect)();\n", []string{"unexpectedRefCall (a ?? Reflect)()"}},
		{"edge", "(JSON as any)();\n", []string{"unexpectedRefCall (JSON as any)()"}},
		{"edge", "JSON!();\n", []string{"unexpectedRefCall JSON!()"}},
		{"edge", "(<any>JSON)();\n", []string{"unexpectedRefCall (<any>JSON)()"}},
		{"edge", "(JSON satisfies object)();\n", []string{"unexpectedRefCall (JSON satisfies object)()"}},
		{"edge", "JSON?.();\n", []string{"unexpectedCall JSON?.()"}},
		{"edge", "new (globalThis.Intl)();\n", []string{"unexpectedCall new (globalThis.Intl)()"}},
		{"edge", "const g = globalThis; g.Math();\n", []string{"unexpectedCall g.Math()"}},
		{"edge", "function f(JSON) { JSON(); }\n", []string{}},
		{"edge", "x.JSON();\n", []string{}},
		{"edge", "Math.max();\n", []string{}},
		{"edge", "var foo = a ? b : globalThis; foo.JSON();\n", []string{"unexpectedCall foo.JSON()"}},
		{"edge", "let j; ({ JSON: j } = globalThis); j();\n", []string{"unexpectedRefCall j()"}},
		{"edge", "let f; [f] = [JSON]; f();\n", []string{}},
		{"edge", "const f = JSON; { const f = 1; f(); }\n", []string{}},
		{"edge", "let x; x ||= JSON; x();\n", []string{"unexpectedRefCall x()"}},
		{"edge", "var f = function () {}; f = Math; f();\n", []string{"unexpectedRefCall f()"}},
		{"edge", "const o = { j: JSON }; o.j();\n", []string{}},
		{"edge", "globalThis.globalThis.JSON();\n", []string{}},
		{"edge", "const { globalThis: g } = globalThis; g.JSON();\n", []string{}},
		{"edge", "const m = Math; new m();\n", []string{"unexpectedRefCall new m()"}},
		{"edge", "globalThis.JSON = 1; globalThis.JSON();\n", []string{"unexpectedCall globalThis.JSON()"}},
		{"edge", "globalThis = 1; globalThis.JSON();\n", []string{}},
		{"edge", "let Math2 = Math; Math2.abs(); Math2();\n", []string{"unexpectedRefCall Math2()"}},
		{"edge", "(globalThis?.Reflect)();\n", []string{"unexpectedCall (globalThis?.Reflect)()"}},
		{"edge", "globalThis?.Reflect();\n", []string{"unexpectedCall globalThis?.Reflect()"}},
		{"edge", "(globalThis.Reflect)?.();\n", []string{"unexpectedCall (globalThis.Reflect)?.()"}},
		{"edge", "foo(JSON)();\n", []string{}},
		{"edge", "Intl();\n", []string{"unexpectedCall Intl()"}},
		{"edge", "const { 'JSON': j } = globalThis; j();\n", []string{"unexpectedRefCall j()"}},
		{"edge", "const { ['JSON']: j } = globalThis; j();\n", []string{"unexpectedRefCall j()"}},
		{"edge", "const { JSON } = globalThis; JSON();\n", []string{"unexpectedCall JSON()"}},
		{"edge", "function f({ JSON: j } = globalThis) { j(); }\n", []string{"unexpectedRefCall j()"}},
		{"edge", "function f(a = b, c = a) {} const q = JSON; q``;\n", []string{}},
		{"edge", "JSON``;\n", []string{}},
		{"edge", "let a = JSON; let b = a; let c = b; b(); c();\n", []string{"unexpectedRefCall b()", "unexpectedRefCall c()"}},
		{"edge", "export const getConfig = getConfig; getConfig();\n", []string{}},
		{"edge", "let a = b, b = a; a();\n", []string{}},
		{"edge", "var x = (JSON as any) as any; x();\n", []string{"unexpectedRefCall x()"}},
		{"edge", "label: JSON();\n", []string{"unexpectedCall JSON()"}},
		{"edge", "class C { static m = JSON; } C.m();\n", []string{}},
		{"edge", "async function f() { (await JSON)(); }\n", []string{}},
		{"edge", "(void 0, JSON)();\n", []string{"unexpectedRefCall (void 0, JSON)()"}},
		{"edge", "const j = JSON; (j)();\n", []string{"unexpectedRefCall (j)()"}},
		{"edge", "var JSON2 = globalThis.JSON; JSON2();\n", []string{"unexpectedRefCall JSON2()"}},
		{"edge", "(JSON, a)();\n", []string{}},
		{"edge", "let a = JSON; (a = b)();\n", []string{}},
		{"edge", "let a = JSON; a = b; b();\n", []string{}},
		{"edge", "let j; ({ j = JSON } = {}); j();\n", []string{"unexpectedRefCall j()"}},
		{"edge", "foo = JSON; foo();\n", []string{}},
		{"edge", "const { ...JSON } = globalThis; JSON();\n", []string{}},
		{"edge", "let j; ({ JSON: j = x } = globalThis); j();\n", []string{"unexpectedRefCall j()"}},
		{"edge", "(JSON<string>)();\n", []string{"unexpectedRefCall (JSON<string>)()"}},
		{"edge", "let a = JSON; [a] = [1]; a();\n", []string{"unexpectedRefCall a()"}},
		{"edge", "let a = JSON; for (a of xs) {} a();\n", []string{"unexpectedRefCall a()"}},
		{"edge", "(JSON ? a : b)();\n", []string{}},
	}

	for index, testCase := range cases {
		result := rule_testing.RunTyped(t, NoObjCalls, "input.ts", testCase.code)
		source := result.SourceFile.Text()
		var got []string
		for _, diagnostic := range result.Diagnostics {
			got = append(got, diagnostic.Message.Id+" "+source[diagnostic.Range.Pos():diagnostic.Range.End()])
		}
		if strings.Join(got, "\n") != strings.Join(testCase.want, "\n") {
			t.Errorf("case %d (%s): %q\n got: %q\nwant: %q", index, testCase.origin, testCase.code, got, testCase.want)
		}
	}
}
