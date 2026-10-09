package typescript

import (
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const awaitThenableFile = "/repository/source/Await.ts"

// disposeGlobals declares what lib ES2022 does not carry, and it is load-bearing rather than
// decorative.
//
// The harness tsconfig pins `lib: ["ES2022"]` and offers no way to raise it, so `Disposable`,
// `AsyncDisposable` and `Symbol.asyncDispose` do not resolve. Probed rather than assumed: under the
// default harness `declare const d: Disposable; await using x = d` produces a type whose flags
// report `any`, which is the error type, and `IsTypeAnyType` then skips the declarator. The case
// goes SILENT while upstream reports it.
//
// That is the dangerous direction. Three of upstream's failing cases and three of its passing ones
// mention these types, and dropping them into the fixture tables unchanged would have put three
// hollow entries in the clean list asserting the exact opposite of upstream, all six passing. The
// tell was a replay of the whole corpus rather than a reading of it.
//
// Declaring the interfaces locally is not enough either, and that was also measured: a local
// `interface AsyncDisposable` gave a `Symbol.asyncDispose` that is not the real well-known symbol,
// so `GetWellKnownSymbolPropertyOfType` missed it and a correct AsyncDisposable REPORTED. Merging
// onto the global `SymbolConstructor` is what makes the symbol unique and the lookup succeed, which
// is why this is a `declare global` block in a second file rather than a prelude in the first.
const disposeGlobals = `
declare global {
  interface SymbolConstructor {
    readonly asyncDispose: unique symbol;
    readonly dispose: unique symbol;
  }
  interface Disposable {
    [Symbol.dispose](): void;
  }
  interface AsyncDisposable {
    [Symbol.asyncDispose](): PromiseLike<void>;
  }
}
export {};
`

// TestAwaitThenableFires carries tsgolint's fifteen failing inputs that need no library beyond
// ES2022, verbatim.
//
// Each was additionally driven through `@typescript-eslint` 8.67.0 on a real program before it
// became a fixture, and both references produced the same verdict on all of them.
func TestAwaitThenableFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"await of a number literal", "await 0;", []string{"await"}},
		{"await of a string literal", "await 'value';", []string{"await"}},
		{"await of a conditional yielding string or number", "async () => await (Math.random() > 0.5 ? '' : 0);", []string{"await"}},
		{"await with no space before a parenthesized conditional", "async () => await(Math.random() > 0.5 ? '' : 0);", []string{"await"}},
		{"await of a class extending Array", "\nclass NonPromise extends Array {}\nawait new NonPromise();\n      ", []string{"await"}},
		{"await of a then method taking no callback", "\nasync function test() {\n  class IncorrectThenable {\n    then() {}\n  }\n  const thenable = new IncorrectThenable();\n\n  await thenable;\n}\n      ", []string{"await"}},
		{"await of an optional call returning void", "\ndeclare const callback: (() => void) | undefined;\nawait callback?.();\n      ", []string{"await"}},
		{"await of a nested optional call", "\ndeclare const obj: { a?: { b?: () => void } };\nawait obj.a?.b?.();\n      ", []string{"await"}},
		{"await of an optional chain through an optional object", "\ndeclare const obj: { a: { b: { c?: () => void } } } | undefined;\nawait obj?.a.b.c?.();\n      ", []string{"await"}},
		{"for await over a sync generator of numbers", "\nfunction* yieldNumbers() {\n  yield 1;\n  yield 2;\n  yield 3;\n}\nfor await (const value of yieldNumbers()) {\n  console.log(value);\n}\n      ", []string{"forAwaitOfNonAsyncIterable"}},
		{"for await over a sync generator of promises", "\nfunction* yieldNumberPromises() {\n  yield Promise.resolve(1);\n  yield Promise.resolve(2);\n  yield Promise.resolve(3);\n}\nfor await (const value of yieldNumberPromises()) {\n  console.log(value);\n}\n      ", []string{"forAwaitOfNonAsyncIterable"}},
		{"await using of an object with a sync dispose", "\nasync function foo() {\n  await using _ = {\n    async [Symbol.dispose]() {},\n  };\n}\n      ", []string{"awaitUsingOfNonAsyncDisposable"}},
		{"await of a generic constrained to number", "\nasync function wrapper<T extends number>(value: T) {\n  return await value;\n}\n      ", []string{"await"}},
		{"await of a method generic constrained to string", "\nclass C<T> {\n  async wrapper<T extends string>(value: T) {\n    return await value;\n  }\n}\n      ", []string{"await"}},
		{"await of a generic constrained to a numeric class parameter", "\nclass C<R extends number> {\n  async wrapper<T extends R>(value: T) {\n    return await value;\n  }\n}\n      ", []string{"await"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, AwaitThenable,
				awaitThenableFile, testCase.sourceText), testCase.wantIds...)
		})
	}
}

// TestAwaitThenableStaysSilent carries tsgolint's twenty seven passing inputs that need no library
// beyond ES2022, verbatim.
//
// The last eight are the ones worth naming. They vary how a generic is constrained, and they are
// the cases a rule written from the listener bodies alone would get wrong, because the decision
// lives in `NeedsToBeAwaited`: an unconstrained `T`, a `T extends unknown`, a `T extends any` and a
// `T extends R` where `R` is itself unconstrained all resolve to `May` rather than `Never` and stay
// silent, while `T extends number` reports and sits in the Fires table above. Nothing in the rule
// file shows that boundary.
func TestAwaitThenableStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"await of resolve and reject", "\nasync function test() {\n  await Promise.resolve('value');\n  await Promise.reject(new Error('message'));\n}\n    "},
		{"await of an immediately invoked async arrow", "\nasync function test() {\n  await (async () => true)();\n}\n    "},
		{"await of a function returning a promise", "\nasync function test() {\n  function returnsPromise() {\n    return Promise.resolve('value');\n  }\n  await returnsPromise();\n}\n    "},
		{"await of an async function call", "\nasync function test() {\n  async function returnsPromiseAsync() {}\n  await returnsPromiseAsync();\n}\n    "},
		{"await of an any value", "\nasync function test() {\n  let anyValue: any;\n  await anyValue;\n}\n    "},
		{"await of an unknown value", "\nasync function test() {\n  let unknownValue: unknown;\n  await unknownValue;\n}\n    "},
		{"await of a declared promise", "\nasync function test() {\n  const numberPromise: Promise<number>;\n  await numberPromise;\n}\n    "},
		{"await of subclasses of Promise", "\nasync function test() {\n  class Foo extends Promise<number> {}\n  const foo: Foo = Foo.resolve(2);\n  await foo;\n\n  class Bar extends Foo {}\n  const bar: Bar = Bar.resolve(2);\n  await bar;\n}\n    "},
		{"await of conditionals and an intersection with a promise", "\nasync function test() {\n  await (Math.random() > 0.5 ? numberPromise : 0);\n  await (Math.random() > 0.5 ? foo : 0);\n  await (Math.random() > 0.5 ? bar : 0);\n\n  const intersectionPromise: Promise<number> & number;\n  await intersectionPromise;\n}\n    "},
		{"await of a then method taking a callback", "\nasync function test() {\n  class Thenable {\n    then(callback: () => {}) {}\n  }\n  const thenable = new Thenable();\n\n  await thenable;\n}\n    "},
		{"await of a polyfilled promise constructor", "\n// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/promise-polyfill/index.d.ts\n// Type definitions for promise-polyfill 6.0\n// Project: https://github.com/taylorhakes/promise-polyfill\n// Definitions by: Steve Jenkins <https://github.com/skysteve>\n//                 Daniel Cassidy <https://github.com/djcsdy>\n// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped\n\ninterface PromisePolyfillConstructor extends PromiseConstructor {\n  _immediateFn?: (handler: (() => void) | string) => void;\n}\n\ndeclare const PromisePolyfill: PromisePolyfillConstructor;\n\nasync function test() {\n  const promise = new PromisePolyfill(() => {});\n\n  await promise;\n}\n    "},
		{"await of a bluebird promise", "\n// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/bluebird/index.d.ts\n// Type definitions for bluebird 3.5\n// Project: https://github.com/petkaantonov/bluebird\n// Definitions by: Leonard Hecker <https://github.com/lhecker>\n// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped\n// TypeScript Version: 2.8\n\n/*!\n * The code following this comment originates from:\n *   https://github.com/types/npm-bluebird\n *\n * Note for browser users: use bluebird-global typings instead of this one\n * if you want to use Bluebird via the global Promise symbol.\n *\n * Licensed under:\n *   The MIT License (MIT)\n *\n *   Copyright (c) 2016 unional\n *\n *   Permission is hereby granted, free of charge, to any person obtaining a copy\n *   of this software and associated documentation files (the \"Software\"), to deal\n *   in the Software without restriction, including without limitation the rights\n *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell\n *   copies of the Software, and to permit persons to whom the Software is\n *   furnished to do so, subject to the following conditions:\n *\n *   The above copyright notice and this permission notice shall be included in\n *   all copies or substantial portions of the Software.\n *\n *   THE SOFTWARE IS PROVIDED \"AS IS\", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR\n *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,\n *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE\n *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER\n *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,\n *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN\n *   THE SOFTWARE.\n */\n\ntype Constructor<E> = new (...args: any[]) => E;\ntype CatchFilter<E> = ((error: E) => boolean) | (object & E);\ntype IterableItem<R> = R extends Iterable<infer U> ? U : never;\ntype IterableOrNever<R> = Extract<R, Iterable<any>>;\ntype Resolvable<R> = R | PromiseLike<R>;\ntype IterateFunction<T, R> = (\n  item: T,\n  index: number,\n  arrayLength: number,\n) => Resolvable<R>;\n\ndeclare class Bluebird<R> implements PromiseLike<R> {\n  then<U>(\n    onFulfill?: (value: R) => Resolvable<U>,\n    onReject?: (error: any) => Resolvable<U>,\n  ): Bluebird<U>; // For simpler signature help.\n  then<TResult1 = R, TResult2 = never>(\n    onfulfilled?: ((value: R) => Resolvable<TResult1>) | null,\n    onrejected?: ((reason: any) => Resolvable<TResult2>) | null,\n  ): Bluebird<TResult1 | TResult2>;\n}\n\ndeclare const bluebird: Bluebird;\n\nasync function test() {\n  await bluebird;\n}\n    "},
		{"await of optional chains ending in a promise call", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  await obj1.a?.b?.c?.();\n  await obj2.a?.b?.c();\n  await obj3.a?.b.c?.();\n  await obj4.a.b.c?.();\n  await obj5.a?.().b?.c?.();\n  await obj6?.a.b.c?.();\n\n  await callback?.();\n};\n    "},
		{"for await over an async generator", "\nasync function* asyncYieldNumbers() {\n  yield 1;\n  yield 2;\n  yield 3;\n}\nfor await (const value of asyncYieldNumbers()) {\n  console.log(value);\n}\n      "},
		{"for await over an any value", "\ndeclare const anee: any;\nasync function forAwait() {\n  for await (const value of anee) {\n    console.log(value);\n  }\n}\n      "},
		{"for await over an async or sync iterable union", "\ndeclare const asyncIter: AsyncIterable<string> | Iterable<string>;\nfor await (const s of asyncIter) {\n}\n      "},
		{"using of an object with a sync dispose", "\nusing foo = {\n  [Symbol.dispose]() {},\n};\n\nexport {};\n      "},
		{"await using of an any value", "\nawait using foo = 3 as any;\n\nexport {};\n      "},
		{"using of an object with an async dispose", "\nusing foo = {\n  async [Symbol.dispose]() {},\n};\n\nexport {};\n      "},
		{"await of an unconstrained generic", "\nasync function wrapper<T>(value: T) {\n  return await value;\n}\n      "},
		{"await of a generic constrained to unknown", "\nasync function wrapper<T extends unknown>(value: T) {\n  return await value;\n}\n      "},
		{"await of a generic constrained to any", "\nasync function wrapper<T extends any>(value: T) {\n  return await value;\n}\n      "},
		{"await of a generic constrained to a promise", "\nasync function wrapper<T extends Promise<unknown>>(value: T) {\n  return await value;\n}\n      "},
		{"await of a generic constrained to number or promise", "\nasync function wrapper<T extends number | Promise<unknown>>(value: T) {\n  return await value;\n}\n      "},
		{"await of a method generic shadowing a class generic", "\nclass C<T> {\n  async wrapper<T>(value: T) {\n    return await value;\n  }\n}\n      "},
		{"await of a generic constrained to an unconstrained class parameter", "\nclass C<R> {\n  async wrapper<T extends R>(value: T) {\n    return await value;\n  }\n}\n      "},
		{"await of a generic constrained to a class parameter constrained to unknown", "\nclass C<R extends unknown> {\n  async wrapper<T extends R>(value: T) {\n    return await value;\n  }\n}\n      "},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, AwaitThenable,
				awaitThenableFile, testCase.sourceText))
		})
	}
}

// TestAwaitThenableDisposable carries the six upstream cases that need the disposable declarations,
// run against a two-file program that supplies them.
//
// Splitting these out is what keeps them honest. Run under the default harness these six all pass
// too, three by reporting nothing where nothing is expected and three by reporting nothing where a
// finding IS expected, so the failing half is the only reason the gap was visible at all. See the
// note on disposeGlobals.
func TestAwaitThenableDisposable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"await using of a Disposable", "\ndeclare const disposable: Disposable;\nasync function foo() {\n  await using d = disposable;\n}\n      ", []string{"awaitUsingOfNonAsyncDisposable"}},
		{"await using of five declarators mixing disposable kinds", "\ndeclare const disposable: Disposable;\ndeclare const asyncDisposable: AsyncDisposable;\nasync function foo() {\n  await using a = disposable,\n    b = asyncDisposable,\n    c = disposable,\n    d = asyncDisposable,\n    e = disposable;\n}\n      ", []string{"awaitUsingOfNonAsyncDisposable", "awaitUsingOfNonAsyncDisposable", "awaitUsingOfNonAsyncDisposable"}},
		{"await using of an any value then a Disposable", "\ndeclare const anee: any;\ndeclare const disposable: Disposable;\nasync function foo() {\n  await using a = anee,\n    b = disposable;\n}\n      ", []string{"awaitUsingOfNonAsyncDisposable"}},
		{"await using of an AsyncDisposable", "\ndeclare const d: AsyncDisposable;\n\nawait using foo = d;\n\nexport {};\n      ", nil},
		{"await using of a disposable union", "\ndeclare const maybeAsyncDisposable: Disposable | AsyncDisposable;\nasync function foo() {\n  await using _ = maybeAsyncDisposable;\n}\n      ", nil},
		{"for await using over async disposables", "\nasync function iterateUsing(arr: Array<AsyncDisposable>) {\n  for (await using foo of arr) {\n  }\n}\n      ", nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFiles(t, AwaitThenable, map[string]string{
				"Globals.ts": disposeGlobals,
				"Await.ts":   testCase.sourceText,
			}, "Await.ts")
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestAwaitThenableSpans asserts where each finding points and what text it renders.
//
// ExpectFindings compares message ids and a count and nothing else, so a rule pointing at the wrong
// node passes a complete fixture pair while being wrong. Three arms report three different shapes
// and each is asserted against a literal typed here rather than against the rule's own message
// constant, because a constant compared to itself moves under mutation and stays green.
//
// The await arm points at the WHOLE await expression rather than at its operand, which is worth
// pinning because reporting the operand is the natural way to write it and no id fixture could see
// the difference. The `for await` arm points at the loop HEAD, a computed range ending where the
// body begins, not at any single node. The `await using` arm points at the INITIALIZER rather than
// at the declaration, so five declarators produce findings at five different offsets.
func TestAwaitThenableSpans(t *testing.T) {
	t.Parallel()

	t.Run("the await arm spans the whole await expression", func(t *testing.T) {
		t.Parallel()
		source := "async function f() {\n  await 0;\n}"
		result := rule_testing.RunTyped(t, AwaitThenable, awaitThenableFile, source)
		if len(result.Diagnostics) != 1 {
			t.Fatalf("want one finding, got %d", len(result.Diagnostics))
		}
		finding := result.Diagnostics[0]
		reported := source[finding.Range.Pos():finding.Range.End()]
		if reported != "await 0" {
			t.Errorf("the finding points at %q, want the whole await expression", reported)
		}
		if finding.Message.Id != "await" {
			t.Errorf("message id is %q", finding.Message.Id)
		}
		if finding.Message.Description != "Unexpected `await` of a non-Promise (non-\"Thenable\") value." {
			t.Errorf("message text is %q", finding.Message.Description)
		}
	})

	t.Run("the for await arm spans the loop head", func(t *testing.T) {
		t.Parallel()
		source := "function* g() {\n  yield 1;\n}\nasync function f() {\n  for await (const value of g()) {\n    console.log(value);\n  }\n}"
		result := rule_testing.RunTyped(t, AwaitThenable, awaitThenableFile, source)
		if len(result.Diagnostics) != 1 {
			t.Fatalf("want one finding, got %d", len(result.Diagnostics))
		}
		finding := result.Diagnostics[0]
		// The range ends at the body's Pos(), which sits BEFORE the whitespace separating the head
		// from the brace, so the slice carries no trailing space. Upstream's own doc draws the
		// carets the same way, under `for await (const foo of bar)` and stopping there. This is
		// asserted rather than trimmed because the end offset is the half a rename cannot change
		// and a wrong `WithEnd` would silently swallow the loop body.
		reported := source[finding.Range.Pos():finding.Range.End()]
		if reported != "for await (const value of g())" {
			t.Errorf("the finding points at %q, want the loop head", reported)
		}
		if finding.Message.Description != "Unexpected `for await...of` of a value that is not async iterable." {
			t.Errorf("message text is %q", finding.Message.Description)
		}
	})

	t.Run("the await using arm points at each initializer", func(t *testing.T) {
		t.Parallel()
		source := "declare const disposable: Disposable;\ndeclare const asyncDisposable: AsyncDisposable;\nasync function foo() {\n  await using a = disposable,\n    b = asyncDisposable,\n    c = disposable;\n}"
		result := rule_testing.RunTypedFiles(t, AwaitThenable, map[string]string{
			"Globals.ts": disposeGlobals,
			"Await.ts":   source,
		}, "Await.ts")
		if len(result.Diagnostics) != 2 {
			t.Fatalf("want two findings, got %d", len(result.Diagnostics))
		}
		for index, finding := range result.Diagnostics {
			reported := source[finding.Range.Pos():finding.Range.End()]
			if reported != "disposable" {
				t.Errorf("finding %d points at %q, want the initializer", index, reported)
			}
			if finding.Message.Description != "Unexpected `await using` of a value that is not async disposable." {
				t.Errorf("message text is %q", finding.Message.Description)
			}
		}
		if result.Diagnostics[0].Range.Pos() == result.Diagnostics[1].Range.Pos() {
			t.Error("both findings point at the same offset, so the per-declarator span is not real")
		}
	})
}

// TestAwaitThenableSuggestions applies each suggested repair and compares the resulting source
// against the output upstream records for it.
//
// This is the assertion the brief calls non-optional and the harness does not supply: ExpectFixedSource
// applies FIXES, and every repair this rule offers is a SUGGESTION, which the engine never applies
// unattended. So the applier is hand-rolled below over a single suggestion's fix list.
//
// It is worth the twenty lines because a suggestion is where this rule can be wrong invisibly. A
// finding carrying a repair that deletes the wrong range satisfies its message id perfectly, and
// every id fixture above stays green over it. Upstream ships fourteen `removeAwait` outputs and two
// `convertToOrdinaryFor` outputs, each an exact before-and-after pair, and they pin that the removal
// takes the `await` KEYWORD and nothing else. Two of them, cases 2 and 3 here, differ only in whether
// a space follows the keyword, which is what makes them a real test of the range rather than of the
// intent.
func TestAwaitThenableSuggestions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantId     string
		wantOutput string
	}{
		{"removing await from a number literal", "await 0;", "removeAwait", " 0;"},
		{"removing await from a string literal", "await 'value';", "removeAwait", " 'value';"},
		{"removing await before a space and a parenthesis", "async () => await (Math.random() > 0.5 ? '' : 0);", "removeAwait", "async () =>  (Math.random() > 0.5 ? '' : 0);"},
		{"removing await directly before a parenthesis", "async () => await(Math.random() > 0.5 ? '' : 0);", "removeAwait", "async () => (Math.random() > 0.5 ? '' : 0);"},
		{"removing await from a constructor call", "\nclass NonPromise extends Array {}\nawait new NonPromise();\n      ", "removeAwait", "\nclass NonPromise extends Array {}\n new NonPromise();\n      "},
		{"removing await from a bad thenable", "\nasync function test() {\n  class IncorrectThenable {\n    then() {}\n  }\n  const thenable = new IncorrectThenable();\n\n  await thenable;\n}\n      ", "removeAwait", "\nasync function test() {\n  class IncorrectThenable {\n    then() {}\n  }\n  const thenable = new IncorrectThenable();\n\n   thenable;\n}\n      "},
		{"removing await from an optional call", "\ndeclare const callback: (() => void) | undefined;\nawait callback?.();\n      ", "removeAwait", "\ndeclare const callback: (() => void) | undefined;\n callback?.();\n      "},
		{"removing await from a nested optional call", "\ndeclare const obj: { a?: { b?: () => void } };\nawait obj.a?.b?.();\n      ", "removeAwait", "\ndeclare const obj: { a?: { b?: () => void } };\n obj.a?.b?.();\n      "},
		{"removing await from an optional chain", "\ndeclare const obj: { a: { b: { c?: () => void } } } | undefined;\nawait obj?.a.b.c?.();\n      ", "removeAwait", "\ndeclare const obj: { a: { b: { c?: () => void } } } | undefined;\n obj?.a.b.c?.();\n      "},
		{"converting a for await over a sync generator", "\nfunction* yieldNumbers() {\n  yield 1;\n  yield 2;\n  yield 3;\n}\nfor await (const value of yieldNumbers()) {\n  console.log(value);\n}\n      ", "convertToOrdinaryFor", "\nfunction* yieldNumbers() {\n  yield 1;\n  yield 2;\n  yield 3;\n}\nfor  (const value of yieldNumbers()) {\n  console.log(value);\n}\n      "},
		{"converting a for await over a generator of promises", "\nfunction* yieldNumberPromises() {\n  yield Promise.resolve(1);\n  yield Promise.resolve(2);\n  yield Promise.resolve(3);\n}\nfor await (const value of yieldNumberPromises()) {\n  console.log(value);\n}\n      ", "convertToOrdinaryFor", "\nfunction* yieldNumberPromises() {\n  yield Promise.resolve(1);\n  yield Promise.resolve(2);\n  yield Promise.resolve(3);\n}\nfor  (const value of yieldNumberPromises()) {\n  console.log(value);\n}\n      "},
		{"removing await from a using declaration", "\nasync function foo() {\n  await using _ = {\n    async [Symbol.dispose]() {},\n  };\n}\n      ", "removeAwait", "\nasync function foo() {\n   using _ = {\n    async [Symbol.dispose]() {},\n  };\n}\n      "},
		{"removing await from a constrained generic", "\nasync function wrapper<T extends number>(value: T) {\n  return await value;\n}\n      ", "removeAwait", "\nasync function wrapper<T extends number>(value: T) {\n  return  value;\n}\n      "},
		{"removing await from a method generic", "\nclass C<T> {\n  async wrapper<T extends string>(value: T) {\n    return await value;\n  }\n}\n      ", "removeAwait", "\nclass C<T> {\n  async wrapper<T extends string>(value: T) {\n    return  value;\n  }\n}\n      "},
		{"removing await from a nested constrained generic", "\nclass C<R extends number> {\n  async wrapper<T extends R>(value: T) {\n    return await value;\n  }\n}\n      ", "removeAwait", "\nclass C<R extends number> {\n  async wrapper<T extends R>(value: T) {\n    return  value;\n  }\n}\n      "},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, AwaitThenable, awaitThenableFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want one finding, got %d", len(result.Diagnostics))
			}
			suggestions := result.Diagnostics[0].Suggestions
			if len(suggestions) != 1 {
				t.Fatalf("want one suggestion, got %d", len(suggestions))
			}
			if suggestions[0].Message.Id != testCase.wantId {
				t.Errorf("suggestion id is %q, want %q", suggestions[0].Message.Id, testCase.wantId)
			}
			// The harness trims the fixture before building the program, so offsets are against the
			// trimmed text rather than against the literal written above.
			source := strings.TrimSpace(testCase.sourceText) + "\n"
			got := applySuggestion(t, source, suggestions[0])
			want := strings.TrimSpace(testCase.wantOutput) + "\n"
			if strings.TrimSpace(got) != strings.TrimSpace(want) {
				t.Errorf("applying the suggestion produced\n%q\nwant\n%q", got, want)
			}
		})
	}
}

// applySuggestion rewrites source with one suggestion's fixes, latest range first.
//
// Back to front so an earlier edit cannot move the offsets of a later one. Every suggestion this
// rule offers carries exactly one fix, so the ordering is defensive rather than exercised, and it is
// written this way because the alternative is correct only by accident.
func applySuggestion(t *testing.T, source string, suggestion rule.Suggestion) string {
	t.Helper()
	fixes := append([]rule.Fix(nil), suggestion.Fixes...)
	sort.Slice(fixes, func(i, j int) bool { return fixes[i].Range.Pos() > fixes[j].Range.Pos() })
	for _, fix := range fixes {
		start, end := fix.Range.Pos(), fix.Range.End()
		if start < 0 || end > len(source) || start > end {
			t.Fatalf("the suggestion proposes an out-of-range edit [%d,%d) over %d bytes", start, end, len(source))
		}
		source = source[:start] + fix.Text + source[end:]
	}
	return source
}

// TestAwaitThenableRequiresTheTypedHarness asserts the rule declares the checker and that the plain
// harness cannot prove it, so a later revert to rule_testing.Run fails loudly instead of going green.
//
// This rule is the silent kind rather than the panicking kind, which is the more dangerous of the
// two. Every listener dereferences ctx.TypeChecker, and under rule_testing.Run that field is nil, so a
// fixture set moved to the untyped harness would see every Fires case fail and every StaysSilent
// case pass VACUOUSLY, having proven nothing at all. The count below is what makes that visible.
//
// The nil guard the standing advice asks for now lives in this rule, because absorbing it off the
// adapter made the listeners ours to edit. It sits in Run rather than in each listener, so a nil
// checker declines the file once instead of being re-tested per node. The guard is unreachable
// through registration, since NeedsTypeChecker is declared; it covers the harness path, where a
// Context is built by hand. This test pins the declaration AND the decline, so losing either one
// fails loudly rather than going vacuously green.
func TestAwaitThenableRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !AwaitThenable.NeedsTypeChecker {
		t.Fatal("the rule stopped declaring NeedsTypeChecker, so every typed fixture would run against a nil checker")
	}
	if AwaitThenable.Name != "@typescript-eslint/await-thenable" {
		t.Errorf("the registered name is %q, and the inventory writes typescript/await-thenable", AwaitThenable.Name)
	}

	// A finding the typed harness produces and the untyped one cannot.
	source := "async function f() {\n  await 0;\n}"
	typed := rule_testing.RunTyped(t, AwaitThenable, awaitThenableFile, source)
	if len(typed.Diagnostics) != 1 {
		t.Fatalf("the typed harness found %d findings, want one", len(typed.Diagnostics))
	}

	// The decline the guard buys. A Context with no checker must yield no listeners rather than
	// dereferencing nil, and asserting it here is the only way the guard is covered at all: the
	// registration path always supplies a checker, so nothing else can reach this branch.
	if listeners := AwaitThenable.Run(rule.Context{}, nil); listeners != nil {
		t.Errorf("a nil checker produced %d listeners, want the rule to decline the file", len(listeners))
	}
}

// TestAwaitThenableUpstreamCorpus replays typescript-eslint 8.67.0's whole corpus, 122 cases, and
// asserts every finding's id and the exact text it covers.
//
// # Why this table exists, and what it caught
//
// The rule was absorbed from tsgolint and replayed tsgolint's 48 cases, which agreed. tsgolint
// predates the `invalidPromiseAggregatorInput` arm, so that replay could not see the arm was
// missing: cohere was silent on `await Promise.all([getCalendarContext(), getDailyPartialContext()])`
// at `modules/art/ArtGenerator.ts:210`, where the second element is a synchronous `string`, and
// ESLint reported it and was right. The doctrine is never worse than ESLint rule by rule, so the
// reference became the installed typescript-eslint and the reference corpus became its own.
//
// Built mechanically rather than retyped: upstream's test file was transpiled and loaded with
// `RuleTester` stubbed to collect every case (122 from one `run` call), then each case was replayed
// through the installed rule in its own program on the bytes this harness writes. The generated
// literals were then decoded with `go/parser` and compared with the replayed sources byte for byte.
// Before the aggregator arm existed, 25 of the 122 cases failed here, every one an aggregator case.
//
// The globals file supplies `Disposable` and `AsyncDisposable`, which lib ES2022 lacks; see
// disposeGlobals for why it is a second file.
func TestAwaitThenableUpstreamCorpus(t *testing.T) {
	t.Parallel()

	if len(awaitThenableCorpus) != 122 {
		t.Fatalf("the corpus table holds %d cases, want 122; a shorter table proves less while passing", len(awaitThenableCorpus))
	}

	for _, corpusCase := range awaitThenableCorpus {
		t.Run(strconv.Itoa(corpusCase.index), func(t *testing.T) {
			t.Parallel()

			result := rule_testing.RunTypedFiles(t, AwaitThenable, map[string]string{
				"Globals.ts": disposeGlobals,
				"Subject.ts": corpusCase.source,
			}, "Subject.ts")

			if len(result.Diagnostics) != len(corpusCase.findings) {
				t.Fatalf("got %d findings %v, want %d %v\n%s", len(result.Diagnostics), result.MessageIds(), len(corpusCase.findings), corpusCase.findings, corpusCase.source)
			}
			sort.SliceStable(result.Diagnostics, func(left, right int) bool {
				return result.Diagnostics[left].Range.Pos() < result.Diagnostics[right].Range.Pos()
			})
			for index, finding := range corpusCase.findings {
				diagnostic := result.Diagnostics[index]
				reported := corpusCase.source[diagnostic.Range.Pos():diagnostic.Range.End()]
				if diagnostic.Message.Id != finding.id || reported != finding.text {
					t.Errorf("finding %d is %s at %q, want %s at %q\n%s", index, diagnostic.Message.Id, reported, finding.id, finding.text, corpusCase.source)
				}
			}
		})
	}
}

// TestAwaitThenableAggregatorShapesTheCorpusDoesNotWrite pins five aggregator inputs upstream's
// corpus never writes, each verdict measured on the installed rule (8.67.0) rather than reasoned.
//
// A mutation sweep of the aggregator arm left three survivors with the corpus alone: the
// `PromiseConstructor` receiver check, the parenthesis unwrap on the argument, and the
// element-access spelling `Promise['all']`. Each case below kills one. The fourth is the ahra site
// that motivated the arm, `modules/art/ArtGenerator.ts:210`, reduced to its shape, and the fifth
// pins that a parenthesized element reports without its parentheses, as ESTree's range does.
//
// The tuple branch of awaitThenableValueTypesOfArrayLike also survives deletion, and that one is an
// equivalent mutant rather than a blind spot: a tuple is array-like, its number index type is the
// union of its members, and "some member part is never awaitable" asks the same question of either.
func TestAwaitThenableAggregatorShapesTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   []string
	}{
		{"an object that merely has an all method is not an aggregator",
			"async function run() {\n  const fake = { all(values: number[]): Promise<void> { return Promise.resolve(); } };\n  await fake.all([1, 2]);\n}",
			nil},
		{"a parenthesized array argument is still an array of elements",
			"async function run() {\n  await Promise.all(([1, Promise.resolve(2)]));\n}",
			[]string{"1"}},
		{"an element access names the aggregator too",
			"async function run() {\n  await Promise['all']([1, Promise.resolve(2)]);\n}",
			[]string{"1"}},
		{"the ahra shape: a synchronous call inside Promise.all",
			"function getDailyPartialContext(): string {\n  return 'today';\n}\nasync function getCalendarContext(): Promise<string> {\n  return 'calendar';\n}\nasync function run() {\n  const [calendar, daily] = await Promise.all([getCalendarContext(), getDailyPartialContext()]);\n  return calendar + daily;\n}",
			[]string{"getDailyPartialContext()"}},
		{"a parenthesized element reports without its parentheses",
			"async function run() {\n  await Promise.all([(1), Promise.resolve(2)]);\n}",
			[]string{"1"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := rule_testing.RunTyped(t, AwaitThenable, awaitThenableFile, testCase.source)
			if len(result.Diagnostics) != len(testCase.want) {
				t.Fatalf("got %d findings %v, want %d %v", len(result.Diagnostics), result.MessageIds(), len(testCase.want), testCase.want)
			}
			for index, want := range testCase.want {
				diagnostic := result.Diagnostics[index]
				reported := testCase.source[diagnostic.Range.Pos():diagnostic.Range.End()]
				if diagnostic.Message.Id != "invalidPromiseAggregatorInput" || reported != want {
					t.Errorf("finding %d is %s at %q, want invalidPromiseAggregatorInput at %q", index, diagnostic.Message.Id, reported, want)
				}
				if diagnostic.Message.Description != "Unexpected iterable of non-Promise (non-\"Thenable\") values passed to promise aggregator." {
					t.Errorf("message text is %q", diagnostic.Message.Description)
				}
			}
		})
	}
}
