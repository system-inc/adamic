package typescript

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// noImpliedEvalFile is the fixture name every case in this file runs under.
//
// A TypeScript extension rather than `.js`: a large part of the corpus writes `declare`,
// type annotations, `import type`, and module declarations, all of which would be parse
// errors under a JavaScript name rather than rule inputs.
const noImpliedEvalFile = "noImpliedEval.ts"

// TestNoImpliedEvalStaysSilent carries all forty eight of tsgolint's valid cases verbatim.
//
// Extracted from tsgolint's own `no_implied_eval_test.go` with Go's `go/ast` parser and emitted
// as Go literals by a generator, so no case was ever retyped and nothing could cook an escape on
// the way in. That mattered here: three of the invalid cases below are built upstream by
// concatenating backticks into a raw string (“ ` + "`" + ` “) to write an EMPTY TEMPLATE
// LITERAL, and a hand-copy produces either a triple backtick or a dropped pair. The parser
// reassembles them correctly and a byte comparison against the fetched source confirmed it.
//
// The set is far more pointed than its size suggests, and the distinctions it pins are ones a
// port written from the message text alone would miss:
//
//   - A receiver that is NOT one of `global`, `globalThis`, `window` never matches, so
//     `foo.setTimeout(null)` is silent even though the method name is eval-like.
//   - A handler that is any kind of FUNCTION is silent, and "function" is decided by the type
//     checker rather than by syntax: a declared `() => void`, an object method, a `.bind()`
//     call, and an element access resolving to a method all pass.
//   - A callee DECLARED IN THIS FILE is exempt, which is what makes the shadowing cases clean.
func TestNoImpliedEvalStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"upstream valid case 0", "foo.setImmediate(null);"},
		{"upstream valid case 1", "foo.setInterval(null);"},
		{"upstream valid case 2", "foo.execScript(null);"},
		{"upstream valid case 3", "foo.setTimeout(null);"},
		{"upstream valid case 4", "foo();"},
		{"upstream valid case 5", "(function () {})();"},
		{"upstream valid case 6", "setTimeout(() => {}, 0);"},
		{"upstream valid case 7", "window.setTimeout(() => {}, 0);"},
		{"upstream valid case 8", "window['setTimeout'](() => {}, 0);"},
		{"upstream valid case 9", "setInterval(() => {}, 0);"},
		{"upstream valid case 10", "window.setInterval(() => {}, 0);"},
		{"upstream valid case 11", "window['setInterval'](() => {}, 0);"},
		{"upstream valid case 12", "setImmediate(() => {});"},
		{"upstream valid case 13", "window.setImmediate(() => {});"},
		{"upstream valid case 14", "window['setImmediate'](() => {});"},
		{"upstream valid case 15", "execScript(() => {});"},
		{"upstream valid case 16", "window.execScript(() => {});"},
		{"upstream valid case 17", "window['execScript'](() => {});"},
		{"upstream valid case 18", "\nconst foo = () => {};\n\nsetTimeout(foo, 0);\nsetInterval(foo, 0);\nsetImmediate(foo);\nexecScript(foo);\n    "},
		{"upstream valid case 19", "\nconst foo = function () {};\n\nsetTimeout(foo, 0);\nsetInterval(foo, 0);\nsetImmediate(foo);\nexecScript(foo);\n    "},
		{"upstream valid case 20", "\nfunction foo() {}\n\nsetTimeout(foo, 0);\nsetInterval(foo, 0);\nsetImmediate(foo);\nexecScript(foo);\n    "},
		{"upstream valid case 21", "\nconst foo = {\n  fn: () => {},\n};\n\nsetTimeout(foo.fn, 0);\nsetInterval(foo.fn, 0);\nsetImmediate(foo.fn);\nexecScript(foo.fn);\n    "},
		{"upstream valid case 22", "\nconst foo = {\n  fn: function () {},\n};\n\nsetTimeout(foo.fn, 0);\nsetInterval(foo.fn, 0);\nsetImmediate(foo.fn);\nexecScript(foo.fn);\n    "},
		{"upstream valid case 23", "\nconst foo = {\n  fn: function foo() {},\n};\n\nsetTimeout(foo.fn, 0);\nsetInterval(foo.fn, 0);\nsetImmediate(foo.fn);\nexecScript(foo.fn);\n    "},
		{"upstream valid case 24", "\nconst foo = {\n  fn() {},\n};\n\nsetTimeout(foo.fn, 0);\nsetInterval(foo.fn, 0);\nsetImmediate(foo.fn);\nexecScript(foo.fn);\n    "},
		{"upstream valid case 25", "\nconst foo = {\n  fn: () => {},\n};\nconst fn = 'fn';\n\nsetTimeout(foo[fn], 0);\nsetInterval(foo[fn], 0);\nsetImmediate(foo[fn]);\nexecScript(foo[fn]);\n    "},
		{"upstream valid case 26", "\nconst foo = {\n  fn: () => {},\n};\n\nsetTimeout(foo['fn'], 0);\nsetInterval(foo['fn'], 0);\nsetImmediate(foo['fn']);\nexecScript(foo['fn']);\n    "},
		{"upstream valid case 27", "\nconst foo: () => void = () => {};\n\nsetTimeout(foo, 0);\nsetInterval(foo, 0);\nsetImmediate(foo);\nexecScript(foo);\n    "},
		{"upstream valid case 28", "\nconst foo: () => () => void = () => {\n  return () => {};\n};\n\nsetTimeout(foo(), 0);\nsetInterval(foo(), 0);\nsetImmediate(foo());\nexecScript(foo());\n    "},
		{"upstream valid case 29", "\nconst foo: () => () => void = () => () => {};\n\nsetTimeout(foo(), 0);\nsetInterval(foo(), 0);\nsetImmediate(foo());\nexecScript(foo());\n    "},
		{"upstream valid case 30", "\nconst foo = () => () => {};\n\nsetTimeout(foo(), 0);\nsetInterval(foo(), 0);\nsetImmediate(foo());\nexecScript(foo());\n    "},
		{"upstream valid case 31", "\nconst foo = function foo() {\n  return function foo() {};\n};\n\nsetTimeout(foo(), 0);\nsetInterval(foo(), 0);\nsetImmediate(foo());\nexecScript(foo());\n    "},
		{"upstream valid case 32", "\nconst foo = function () {\n  return function () {\n    return '';\n  };\n};\n\nsetTimeout(foo(), 0);\nsetInterval(foo(), 0);\nsetImmediate(foo());\nexecScript(foo());\n    "},
		{"upstream valid case 33", "\nconst foo: () => () => void = function foo() {\n  return function foo() {};\n};\n\nsetTimeout(foo(), 0);\nsetInterval(foo(), 0);\nsetImmediate(foo());\nexecScript(foo());\n    "},
		{"upstream valid case 34", "\nfunction foo() {\n  return function foo() {\n    return () => {};\n  };\n}\n\nsetTimeout(foo()(), 0);\nsetInterval(foo()(), 0);\nsetImmediate(foo()());\nexecScript(foo()());\n    "},
		{"upstream valid case 35", "\nclass Foo {\n  static fn = () => {};\n}\n\nsetTimeout(Foo.fn, 0);\nsetInterval(Foo.fn, 0);\nsetImmediate(Foo.fn);\nexecScript(Foo.fn);\n    "},
		{"upstream valid case 36", "\nclass Foo {\n  fn() {}\n}\n\nconst foo = new Foo();\n\nsetTimeout(foo.fn, 0);\nsetInterval(foo.fn, 0);\nsetImmediate(foo.fn);\nexecScript(foo.fn);\n    "},
		{"upstream valid case 37", "\nclass Foo {\n  fn() {}\n}\nconst foo = new Foo();\nconst fn = foo.fn;\n\nsetTimeout(fn.bind(null), 0);\nsetInterval(fn.bind(null), 0);\nsetImmediate(fn.bind(null));\nexecScript(fn.bind(null));\n    "},
		{"upstream valid case 38", "\nconst fn = (foo: () => void) => {\n  setTimeout(foo, 0);\n  setInterval(foo, 0);\n  setImmediate(foo);\n  execScript(foo);\n};\n    "},
		{"upstream valid case 40", "\nconst foo = (callback: Function) => {\n  setTimeout(callback, 0);\n};\n    "},
		{"upstream valid case 41", "\nconst foo = () => {};\nconst bar = () => {};\n\nsetTimeout(Math.radom() > 0.5 ? foo : bar, 0);\nsetTimeout(foo || bar, 500);\n    "},
		{"upstream valid case 42", "\nclass Foo {\n  func1() {}\n  func2(): void {\n    setTimeout(this.func1.bind(this), 1);\n  }\n}\n    "},
		{"upstream valid case 43", "\nclass Foo {\n  private a = {\n    b: {\n      c: function () {},\n    },\n  };\n  funcw(): void {\n    setTimeout(this.a.b.c.bind(this), 1);\n  }\n}\n    "},
		{"upstream valid case 44", "\nfunction setTimeout(input: string, value: number) {}\n\nsetTimeout('', 0);\n    "},
		{"upstream valid case 45", "\ndeclare module 'my-timers-promises' {\n  export function setTimeout(ms: number): void;\n}\n\nimport { setTimeout } from 'my-timers-promises';\n\nsetTimeout(1000);\n    "},
		{"upstream valid case 46", "\nfunction setTimeout() {}\n\n{\n  setTimeout(100);\n}\n    "},
		{"upstream valid case 47", "\nfunction setTimeout() {}\n\n{\n  setTimeout(\"alert('evil!')\");\n}\n    "},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoImpliedEval, noImpliedEvalFile, testCase.sourceText))
		})
	}
}

// TestNoImpliedEvalFires carries all twenty two of tsgolint's invalid cases verbatim.
//
// Seventy six findings over twenty two inputs, so most cases report more than once and the
// per-input counts come from upstream's own `Errors` slices rather than from any recovery
// heuristic: the extractor read them out of the same AST it read the code from, so code and
// expectation cannot drift apart.
//
// Two message ids appear, and the split is 70 `noImpliedEvalError` to 6 `noFunctionConstructor`.
// They are not alternatives: `new Function('x')` can only be the constructor arm, while the four
// eval-like callees only ever produce the other, and one case produces both.
func TestNoImpliedEvalFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		sourceText  string
		expectedIds []string
	}{
		{"upstream invalid case 0", "\nsetTimeout('x = 1', 0);\nsetInterval('x = 1', 0);\nsetImmediate('x = 1');\nexecScript('x = 1');\n      ", []string{"noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError"}},
		{"upstream invalid case 1", "\nsetTimeout(undefined, 0);\nsetInterval(undefined, 0);\nsetImmediate(undefined);\nexecScript(undefined);\n      ", []string{"noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError"}},
		{"upstream invalid case 2", "\nsetTimeout(1 + '' + (() => {}), 0);\nsetInterval(1 + '' + (() => {}), 0);\nsetImmediate(1 + '' + (() => {}));\nexecScript(1 + '' + (() => {}));\n      ", []string{"noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError"}},
		{"upstream invalid case 3", "\nconst foo = 'x = 1';\n\nsetTimeout(foo, 0);\nsetInterval(foo, 0);\nsetImmediate(foo);\nexecScript(foo);\n      ", []string{"noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError"}},
		{"upstream invalid case 4", "\nconst foo = function () {\n  return 'x + 1';\n};\n\nsetTimeout(foo(), 0);\nsetInterval(foo(), 0);\nsetImmediate(foo());\nexecScript(foo());\n      ", []string{"noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError"}},
		{"upstream invalid case 5", "\nconst foo = function () {\n  return () => 'x + 1';\n};\n\nsetTimeout(foo()(), 0);\nsetInterval(foo()(), 0);\nsetImmediate(foo()());\nexecScript(foo()());\n      ", []string{"noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError"}},
		{"upstream invalid case 6", "\nconst fn = function () {};\n\nsetTimeout(fn + '', 0);\nsetInterval(fn + '', 0);\nsetImmediate(fn + '');\nexecScript(fn + '');\n      ", []string{"noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError"}},
		{"upstream invalid case 7", "\nconst foo: string = 'x + 1';\n\nsetTimeout(foo, 0);\nsetInterval(foo, 0);\nsetImmediate(foo);\nexecScript(foo);\n      ", []string{"noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError"}},
		{"upstream invalid case 8", "\nconst foo = new String('x + 1');\n\nsetTimeout(foo, 0);\nsetInterval(foo, 0);\nsetImmediate(foo);\nexecScript(foo);\n      ", []string{"noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError"}},
		{"upstream invalid case 9", "\nconst foo = 'x + 1';\n\nsetTimeout(foo as any, 0);\nsetInterval(foo as any, 0);\nsetImmediate(foo as any);\nexecScript(foo as any);\n      ", []string{"noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError"}},
		{"upstream invalid case 10", "\nconst fn = (foo: string | any) => {\n  setTimeout(foo, 0);\n  setInterval(foo, 0);\n  setImmediate(foo);\n  execScript(foo);\n};\n      ", []string{"noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError"}},
		{"upstream invalid case 11", "\nconst foo = 'foo';\nconst bar = () => {};\n\nsetTimeout(Math.radom() > 0.5 ? foo : bar, 0);\n      ", []string{"noImpliedEvalError"}},
		{"upstream invalid case 12", "\nwindow.setTimeout(``, 0);\nwindow['setTimeout'](``, 0);\n\nwindow.setInterval(``, 0);\nwindow['setInterval'](``, 0);\n\nwindow.setImmediate(``);\nwindow['setImmediate'](``);\n\nwindow.execScript(``);\nwindow['execScript'](``);\n      ", []string{"noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError"}},
		{"upstream invalid case 13", "\nglobal.setTimeout(``, 0);\nglobal['setTimeout'](``, 0);\n\nglobal.setInterval(``, 0);\nglobal['setInterval'](``, 0);\n\nglobal.setImmediate(``);\nglobal['setImmediate'](``);\n\nglobal.execScript(``);\nglobal['execScript'](``);\n      ", []string{"noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError"}},
		{"upstream invalid case 14", "\nglobalThis.setTimeout(``, 0);\nglobalThis['setTimeout'](``, 0);\n\nglobalThis.setInterval(``, 0);\nglobalThis['setInterval'](``, 0);\n\nglobalThis.setImmediate(``);\nglobalThis['setImmediate'](``);\n\nglobalThis.execScript(``);\nglobalThis['execScript'](``);\n      ", []string{"noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError", "noImpliedEvalError"}},
		{"upstream invalid case 15", "\nconst foo: string | undefined = 'hello';\nconst bar = () => {};\n\nsetTimeout(foo || bar, 500);\n      ", []string{"noImpliedEvalError"}},
		{"upstream invalid case 16", "const fn = Function();", []string{"noFunctionConstructor"}},
		{"upstream invalid case 17", "const fn = new Function('a', 'b', 'return a + b');", []string{"noFunctionConstructor"}},
		{"upstream invalid case 18", "const fn = window.Function();", []string{"noFunctionConstructor"}},
		{"upstream invalid case 19", "const fn = new window.Function();", []string{"noFunctionConstructor"}},
		{"upstream invalid case 20", "const fn = window['Function']();", []string{"noFunctionConstructor"}},
		{"upstream invalid case 21", "const fn = new window['Function']();", []string{"noFunctionConstructor"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoImpliedEval, noImpliedEvalFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.expectedIds...)
		})
	}
}

// TestNoImpliedEvalExemptsAShadowingDeclarationFromAnotherFile is upstream's forty-eighth valid
// case, which the single-file harness cannot express.
//
// The case is `import { Function } from './class'; new Function('foo');`, and upstream's fixture
// directory supplies `class.ts` containing `export class Function {}` — a file whose comments
// name this rule as the reason it exists. Dropped into `RunTyped` unchanged the import resolves
// to nothing, `IsBuiltinSymbolLike` cannot rule the type out, and the case REPORTS: it would have
// sat in the clean list above asserting the exact opposite of upstream while passing.
//
// The control matters as much as the case. The same source with a class that does NOT shadow
// `Function` still reports, so the silence below is the shadowing doing work rather than the
// multi-file harness swallowing the rule.
func TestNoImpliedEvalExemptsAShadowingDeclarationFromAnotherFile(t *testing.T) {
	t.Parallel()

	const shadowed = "export class Function {}\n"
	const unrelated = "export class NotFunction {}\n"
	const subject = "\nimport { Function } from './class';\nnew Function('foo');\n    "

	t.Run("a locally declared Function shadows the constructor", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, rule_testing.RunTypedFiles(t, NoImpliedEval, map[string]string{
			"class.ts":        shadowed,
			noImpliedEvalFile: subject,
		}, noImpliedEvalFile))
	})

	t.Run("control: the real constructor still reports", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.RunTypedFiles(t, NoImpliedEval, map[string]string{
			"class.ts":        unrelated,
			noImpliedEvalFile: "new Function('foo');",
		}, noImpliedEvalFile), "noFunctionConstructor")
	})
}

// TestNoImpliedEvalRecognizesTheCalleeByNameRatherThanByResolution pins the decision the imported
// corpus cannot see, because upstream writes none of these shapes.
//
// Every case here was measured against the vendored rule before it was written down, and the
// silent ones each have a reporting neighbour so that no zero is a zero for a reason other than
// the one claimed.
func TestNoImpliedEvalRecognizesTheCalleeByNameRatherThanByResolution(t *testing.T) {
	t.Parallel()

	reporting := []struct{ name, sourceText string }{
		// The three receivers upstream hardcodes, plus the bare form.
		{"a bare global", "setTimeout('x=1', 0);"},
		{"the window receiver", "declare const window: any;\nwindow.setTimeout('x=1', 0);"},
		{"the globalThis receiver", "globalThis.setTimeout('x=1', 0);"},
		{"the global receiver", "declare const global: any;\nglobal.setTimeout('x=1', 0);"},
		// An element access matches, but only through a string LITERAL key.
		{"a string literal element key", "declare const window: any;\nwindow['setTimeout']('x=1', 0);"},
		// Optional chaining still produces an access node, so it matches.
		{"an optional chain", "declare const window: any;\nwindow?.setTimeout('x=1', 0);"},
		// The exemption is FILE-scoped rather than lexical: a declaration nested in a block
		// exempts nothing, because the symbol it creates is not the global one being called.
		{"a shadow confined to a nested block", "{\n  function setTimeout(x: string) {}\n}\nsetTimeout('x=1', 0);"},
	}
	for _, testCase := range reporting {
		t.Run("reports/"+testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoImpliedEval, noImpliedEvalFile, testCase.sourceText), "noImpliedEvalError")
		})
	}

	silent := []struct{ name, sourceText string }{
		// `self` is a genuine global alias for `window` and is simply not on upstream's list.
		{"the self receiver, which is not a global candidate", "declare const self: any;\nself.setTimeout('x=1', 0);"},
		{"an arbitrary receiver", "declare const foo: any;\nfoo.setTimeout('x=1', 0);"},
		// The key must be a string literal node; neither of these is one.
		{"a variable element key", "declare const window: any;\nconst k = 'setTimeout';\nwindow[k]('x=1', 0);"},
		{"a template literal element key", "declare const window: any;\nwindow[`setTimeout`]('x=1', 0);"},
		// The matched access's receiver must BE the candidate identifier, not contain it.
		{"one level of extra nesting under window", "declare const window: any;\nwindow.a.setTimeout('x=1', 0);"},
		// Resolution as the negative filter: a declaration anywhere in this file exempts.
		{"a local function declaration", "function setTimeout(x: string) {}\nsetTimeout('x=1');"},
		{"a local const", "const setTimeout = (x: string) => {};\nsetTimeout('x=1');"},
		{"a declare function in this file", "declare function setTimeout(h: any, t: number): void;\nsetTimeout('x=1', 0);"},
		{"an import declaration in this file", "import { setTimeout } from './timers';\nsetTimeout('x=1');"},
		// An alias carries the type but not the name, and names are all this rule reads.
		{"an alias to the global under a different name", "const st = setTimeout;\nst('x=1', 0);"},
	}
	for _, testCase := range silent {
		t.Run("silent/"+testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoImpliedEval, noImpliedEvalFile, testCase.sourceText))
		})
	}
}

// TestNoImpliedEvalTestsForANonFunctionRatherThanForAString pins the belief most likely to be got
// backwards by anyone reading the rule's name or its string-heavy corpus.
//
// Nothing in the rule asks whether a value is a string. `isFunction` asks the checker whether the
// handler is callable and reports everything that is not, so a number, `null`, `unknown` and
// `any` all report while a `.bind()` call and a `Function`-typed variable do not.
func TestNoImpliedEvalTestsForANonFunctionRatherThanForAString(t *testing.T) {
	t.Parallel()

	reporting := []struct{ name, sourceText string }{
		{"a number literal", "setTimeout(1, 0);"},
		{"null", "setTimeout(null, 0);"},
		{"a declared unknown", "declare const u: unknown;\nsetTimeout(u, 0);"},
		// The opposite of the usual any-skip hazard: an `any` has no call signature, so it
		// reports. This is why the fixture tsconfig's lib pinning cannot silence this rule.
		{"a declared any", "declare const a: any;\nsetTimeout(a, 0);"},
		{"a call returning a string", "declare function make(): string;\nsetTimeout(make(), 0);"},
		{"a String wrapper object", "setTimeout(new String('x=1'), 0);"},
		{"a template literal with substitutions", "declare const n: number;\nsetTimeout(`x=${n}`, 0);"},
		{"a tagged template", "declare function tag(s: TemplateStringsArray): string;\nsetTimeout(tag`x`, 0);"},
	}
	for _, testCase := range reporting {
		t.Run("reports/"+testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoImpliedEval, noImpliedEvalFile, testCase.sourceText), "noImpliedEvalError")
		})
	}

	silent := []struct{ name, sourceText string }{
		{"a bind call", "declare function f(): void;\nsetTimeout(f.bind(null), 0);"},
		// `isBind` keys on the NAME `bind` and never checks the receiver, so any method called
		// `bind` exempts regardless of what it returns. Upstream behavior, reproduced.
		{"a method merely NAMED bind returning a string", "declare const o: { bind(): string };\nsetTimeout(o.bind(), 0);"},
		{"a call returning a function", "declare function make(): () => void;\nsetTimeout(make(), 0);"},
		{"a Function-typed variable", "declare const f: Function;\nsetTimeout(f, 0);"},
		// No first argument means nothing to judge.
		{"a call with no arguments", "setTimeout();"},
	}
	for _, testCase := range silent {
		t.Run("silent/"+testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoImpliedEval, noImpliedEvalFile, testCase.sourceText))
		})
	}
}

// TestNoImpliedEvalDoesNotSkipParenthesesOnTheCallee pins the one axis on which the two upstream
// implementations disagree, and it is the axis the imported corpus is entirely silent about.
//
// `@typescript-eslint` reports every one of the callee cases below and tsgolint reports none of
// them, measured by running both on identical inputs rather than by reading either source. The
// cause is structural: ESTree has no parenthesized-expression node, so ESLint's rule never sees
// the parens, while typescript-go keeps them and tsgolint's identifier test declines.
//
// tsgolint wins because oxlint runs tsgolint, so these fixtures encode the SILENT answer
// deliberately. Without them a later reader could add a paren skip, believe it a free
// correctness improvement, and pass every imported fixture while introducing a difference the
// differential harness would see.
//
// The argument side falls the other way, which is why it is in the same test: nothing skips the
// parens there either, but the checker sees through them, so a parenthesized string still reports
// and a parenthesized arrow is still silent.
func TestNoImpliedEvalDoesNotSkipParenthesesOnTheCallee(t *testing.T) {
	t.Parallel()

	silent := []struct{ name, sourceText string }{
		{"a parenthesized bare callee", "(setTimeout)('x=1', 0);"},
		{"a doubly parenthesized bare callee", "((setTimeout))('x=1', 0);"},
		{"a parenthesized receiver", "declare const window: any;\n(window).setTimeout('x=1', 0);"},
		{"a parenthesized member callee", "declare const window: any;\n(window.setTimeout)('x=1', 0);"},
		{"a parenthesized constructor in a new expression", "new (Function)('a');"},
		{"a parenthesized constructor called directly", "(Function)('a');"},
	}
	for _, testCase := range silent {
		t.Run("silent/"+testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoImpliedEval, noImpliedEvalFile, testCase.sourceText))
		})
	}

	t.Run("control: the same calls without parentheses report", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoImpliedEval, noImpliedEvalFile,
			"declare const window: any;\nwindow.setTimeout('x=1', 0);"), "noImpliedEvalError")
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoImpliedEval, noImpliedEvalFile,
			"new Function('a');"), "noFunctionConstructor")
	})

	t.Run("the ARGUMENT is judged through its parentheses", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoImpliedEval, noImpliedEvalFile,
			"setTimeout(('x=1'), 0);"), "noImpliedEvalError")
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoImpliedEval, noImpliedEvalFile,
			"setTimeout((() => {}), 0);"))
	})
}

// TestNoImpliedEvalPointsAtTheRightNodeAndCarriesTheRightText asserts what ExpectFindings cannot.
//
// Message ids and counts are satisfied by a finding that points at the wrong node, and this rule
// reports at TWO different anchors depending on the arm: the HANDLER for an eval-like call, and
// the WHOLE CALL for the Function constructor. A port that reported the whole call in both cases
// would pass every id fixture above.
//
// The spans are also the only thing that can see the adapter's trivia trim. Reporting `node.Loc`
// rather than the token range carries leading whitespace and the preceding newline into the
// finding, which is invisible to every id assertion and reaches the user as a caret on the wrong
// line.
//
// The message text is asserted against literal strings typed here rather than against the rule's
// own message constants, because a comparison to the constant moves with the rule under mutation
// and therefore guards nothing.
func TestNoImpliedEvalPointsAtTheRightNodeAndCarriesTheRightText(t *testing.T) {
	t.Parallel()

	t.Run("the eval-like arm points at the HANDLER", func(t *testing.T) {
		t.Parallel()
		const sourceText = "declare const window: any;\nwindow.setTimeout('x = 1', 0);"
		result := rule_testing.RunTyped(t, NoImpliedEval, noImpliedEvalFile, sourceText)
		if len(result.Diagnostics) != 1 {
			t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
		}
		finding := result.Diagnostics[0]
		trimmed := strings.TrimSpace(sourceText)
		if reported := trimmed[finding.Range.Pos():finding.Range.End()]; reported != "'x = 1'" {
			t.Errorf("span: want %q, got %q", "'x = 1'", reported)
		}
		if finding.Message.Id != "noImpliedEvalError" {
			t.Errorf("id: want %q, got %q", "noImpliedEvalError", finding.Message.Id)
		}
		if finding.Message.Description != "Implied eval. Consider passing a function." {
			t.Errorf("description: got %q", finding.Message.Description)
		}
	})

	t.Run("the constructor arm points at the WHOLE call", func(t *testing.T) {
		t.Parallel()
		const sourceText = "new Function('a', 'return a');"
		result := rule_testing.RunTyped(t, NoImpliedEval, noImpliedEvalFile, sourceText)
		if len(result.Diagnostics) != 1 {
			t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
		}
		finding := result.Diagnostics[0]
		if reported := strings.TrimSpace(sourceText)[finding.Range.Pos():finding.Range.End()]; reported != "new Function('a', 'return a')" {
			t.Errorf("span: want %q, got %q", "new Function('a', 'return a')", reported)
		}
		if finding.Message.Id != "noFunctionConstructor" {
			t.Errorf("id: want %q, got %q", "noFunctionConstructor", finding.Message.Id)
		}
		if finding.Message.Description != "Implied eval. Do not use the Function constructor to create functions." {
			t.Errorf("description: got %q", finding.Message.Description)
		}
	})

	// An INDENTED call is the shape that exposes a raw `node.Loc` report: the node's position
	// begins at the end of the previous line, so an untrimmed range renders with a leading
	// newline and the indentation. Both anchors are checked because the adapter trims all three
	// of its node-report forms and a regression could restore any one of them.
	t.Run("leading trivia is not carried into either span", func(t *testing.T) {
		t.Parallel()
		const sourceText = "function outer() {\n    setTimeout('x = 1', 0);\n    new Function('a');\n}"
		result := rule_testing.RunTyped(t, NoImpliedEval, noImpliedEvalFile, sourceText)
		if len(result.Diagnostics) != 2 {
			t.Fatalf("want 2 findings, got %d", len(result.Diagnostics))
		}
		trimmed := strings.TrimSpace(sourceText)
		want := []string{"'x = 1'", "new Function('a')"}
		for index, expected := range want {
			if reported := trimmed[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]; reported != expected {
				t.Errorf("finding %d span: want %q, got %q", index, expected, reported)
			}
		}
	})
}

// TestNoImpliedEvalDeclaresItNeedsTheTypeChecker pins the declarations AND the nil guard.
//
// Every listener in this rule reads `ctx.TypeChecker`. While the rule was adapted, the nil guard the
// house style asks for could not be written, because the listener was upstream's; `upstream.Adapt`
// set NeedsTypeChecker unconditionally, which made the nil case unreachable through it. Absorbing
// the rule made the listener ours to edit and made that case REACHABLE, so both halves are asserted.
//
// The direction that makes this worth a test is not a crash. `GetTypeAtLocation` and
// `GetSymbolAtLocation` on this shim return nil rather than panicking, and `isFunction` treats a
// type it cannot call as NOT a function — so a checker-less run of an UNGUARDED listener would
// report every one of the corpus's valid `setTimeout(fn, 0)` cases while the same-file filter let
// each one through on a nil symbol. That is a vacuous RED rather than a vacuous green, and it is
// what the guard below prevents.
func TestNoImpliedEvalDeclaresItNeedsTheTypeChecker(t *testing.T) {
	t.Parallel()

	if !NoImpliedEval.NeedsTypeChecker {
		t.Error("NoImpliedEval must declare NeedsTypeChecker: every listener reads ctx.TypeChecker")
	}
	if NoImpliedEval.ProgramReads != rule.ReadsCompilerOptions|rule.ReadsDefaultLibrary {
		t.Errorf("NoImpliedEval must declare what IsBuiltinSymbolLike reads, the compiler options and the default library; declares %s", NoImpliedEval.ProgramReads)
	}
	if NoImpliedEval.Name != "@typescript-eslint/no-implied-eval" {
		t.Errorf("registered name: want %q, got %q", "no-implied-eval", NoImpliedEval.Name)
	}

	// The guard the absorption made possible. Driving both listeners with a checker-less Context must
	// return rather than resolve anything, and this is the only path that reaches that branch, since
	// registration always supplies a checker.
	typed := rule_testing.RunTyped(t, NoImpliedEval, noImpliedEvalFile, "setTimeout(() => {}, 0);\n")
	if len(typed.Diagnostics) != 0 {
		t.Fatalf("the typed harness found %d findings on a clean case, want none", len(typed.Diagnostics))
	}

	listeners := NoImpliedEval.Run(rule.Context{SourceFile: typed.SourceFile}, nil)
	for _, kind := range []ast.Kind{ast.KindCallExpression, ast.KindNewExpression} {
		listener, hasListener := listeners[kind]
		if !hasListener {
			t.Fatalf("the rule stopped listening on kind %v", kind)
		}
		for _, statement := range typed.SourceFile.Statements.Nodes {
			if statement.Kind != ast.KindExpressionStatement {
				continue
			}
			if expression := statement.AsExpressionStatement().Expression; expression.Kind == kind {
				listener(expression)
			}
		}
	}
}
