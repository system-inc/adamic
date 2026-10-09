package typescript

import (
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// noFloatingPromisesFile is the fixture name every case in this file runs under.
//
// A TypeScript extension rather than `.js`, because the corpus is full of `declare`, type
// annotations, generics and `import type`, all of which would be parse errors under a JavaScript
// name rather than rule inputs.
const noFloatingPromisesFile = "noFloatingPromises.ts"

// noFloatingPromisesSuggestion is one expected repair: the id it carries and the source it writes.
type noFloatingPromisesSuggestion struct {
	messageId string
	output    string
}

// runNoFloatingPromises routes every case through the typed harness, and through the OPTIONS path
// even when the case configures nothing.
//
// A case with no options passes nil rather than a zero struct, which is what a rule configured as
// bare `"error"` is actually handed in production. That is the shape the standing warning names:
// `options.(NoFloatingPromisesOptions)` on nil yields the zero value, and this rule's zero value is
// three nil pointers that the body would dereference. The `!ok` arm plus the three nil-defaulting
// assignments are what keep that from panicking, and routing the sixty-one unconfigured invalid
// cases through here is what puts that arm under test rather than leaving it to a mutation sweep.
//
// # Why the FILENAME is a parameter of this rule rather than a constant
//
// A `TypeOrValueSpecifier` with `From: file` and a `Path` matches by comparing the DECLARING
// file's absolute name against that path resolved from the program's current directory, in
// `typecheck`'s `typeDeclaredInFile`. Upstream's own runner writes every fixture as `file.ts` under
// `tests/fixtures`, so a `Path: "file.ts"` case is only clean because of what upstream named the
// file, and one upstream case relies on exactly that while three others rely on the same
// comparison FAILING for `tests/fixtures/file.ts`.
//
// Dropped in under one constant filename, upstream valid case 57 REPORTS, and it would have sat
// in the clean list asserting the opposite of upstream while looking like a rule defect. Measured
// on the one input, holding source and options fixed and varying only the name:
//
//	noFloatingPromises.ts     REPORTS floatingVoid    the path does not match
//	file.ts                   SILENT                  the path matches
//	tests/fixtures/file.ts    REPORTS floatingVoid    the path does not match
//
// The last row is the control, and it is the same verdict upstream's invalid cases 83 through 85
// assert with that exact path, which is why this is a harness-naming artifact rather than a
// divergence. Cases naming a path therefore run under it.
func runNoFloatingPromises(t *testing.T, sourceText string, options any) rule_testing.Result {
	t.Helper()
	fileName := noFloatingPromisesFile
	if parsed, ok := options.(NoFloatingPromisesOptions); ok {
		for _, specifier := range parsed.AllowForKnownSafeCalls {
			if specifier.Path != "" {
				fileName = specifier.Path
			}
		}
	}
	if options == nil {
		return rule_testing.RunTypedWithOptions(t, NoFloatingPromises, fileName, sourceText, nil)
	}
	return rule_testing.RunTypedWithOptions(t, NoFloatingPromises, fileName, sourceText, options)
}

// applyNoFloatingPromisesSuggestion rewrites source with one suggestion's fixes, latest range
// first.
//
// Back to front matters here rather than being defensive: the parenthesizing repairs carry TWO
// fixes, an insert before the statement and an insert after the expression, and applying them in
// source order would shift the second one's offset by the length of the first.
func applyNoFloatingPromisesSuggestion(t *testing.T, source string, suggestion rule.Suggestion) string {
	t.Helper()
	fixes := append([]rule.Fix(nil), suggestion.Fixes...)
	sort.SliceStable(fixes, func(i, j int) bool { return fixes[i].Range.Pos() > fixes[j].Range.Pos() })
	for _, fix := range fixes {
		start, end := fix.Range.Pos(), fix.Range.End()
		if start < 0 || end > len(source) || start > end {
			t.Fatalf("the suggestion proposes an out-of-range edit [%d,%d) over %d bytes", start, end, len(source))
		}
		source = source[:start] + fix.Text + source[end:]
	}
	return source
}

func TestNoFloatingPromisesStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    any
	}{
		{"upstream valid case 0", "\nasync function test() {\n  await Promise.resolve('value');\n  Promise.resolve('value').then(\n    () => {},\n    () => {},\n  );\n  Promise.resolve('value')\n    .then(() => {})\n    .catch(() => {});\n  Promise.resolve('value')\n    .then(() => {})\n    .catch(() => {})\n    .finally(() => {});\n  Promise.resolve('value').catch(() => {});\n  return Promise.resolve('value');\n}\n    ", nil},
		{"upstream valid case 1", "\nasync function test() {\n  void Promise.resolve('value');\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(true)}},
		{"upstream valid case 2", "\nasync function test() {\n  await Promise.reject(new Error('message'));\n  Promise.reject(new Error('message')).then(\n    () => {},\n    () => {},\n  );\n  Promise.reject(new Error('message'))\n    .then(() => {})\n    .catch(() => {});\n  Promise.reject(new Error('message'))\n    .then(() => {})\n    .catch(() => {})\n    .finally(() => {});\n  Promise.reject(new Error('message')).catch(() => {});\n  return Promise.reject(new Error('message'));\n}\n    ", nil},
		{"upstream valid case 3", "\nasync function test() {\n  await (async () => true)();\n  (async () => true)().then(\n    () => {},\n    () => {},\n  );\n  (async () => true)()\n    .then(() => {})\n    .catch(() => {});\n  (async () => true)()\n    .then(() => {})\n    .catch(() => {})\n    .finally(() => {});\n  (async () => true)().catch(() => {});\n  return (async () => true)();\n}\n    ", nil},
		{"upstream valid case 4", "\nasync function test() {\n  async function returnsPromise() {}\n  await returnsPromise();\n  returnsPromise().then(\n    () => {},\n    () => {},\n  );\n  returnsPromise()\n    .then(() => {})\n    .catch(() => {});\n  returnsPromise()\n    .then(() => {})\n    .catch(() => {})\n    .finally(() => {});\n  returnsPromise().catch(() => {});\n  return returnsPromise();\n}\n    ", nil},
		{"upstream valid case 5", "\nasync function test() {\n  const x = Promise.resolve();\n  const y = x.then(() => {});\n  y.catch(() => {});\n}\n    ", nil},
		{"upstream valid case 6", "\nasync function test() {\n  Math.random() > 0.5 ? Promise.resolve().catch(() => {}) : null;\n}\n    ", nil},
		{"upstream valid case 7", "\nasync function test() {\n  Promise.resolve().catch(() => {}), 123;\n  123,\n    Promise.resolve().then(\n      () => {},\n      () => {},\n    );\n  123,\n    Promise.resolve().then(\n      () => {},\n      () => {},\n    ),\n    123;\n}\n    ", nil},
		{"upstream valid case 8", "\nasync function test() {\n  void Promise.resolve().catch(() => {});\n}\n    ", nil},
		{"upstream valid case 9", "\nasync function test() {\n  Promise.resolve().catch(() => {}) ||\n    Promise.resolve().then(\n      () => {},\n      () => {},\n    );\n}\n    ", nil},
		{"upstream valid case 10", "\ndeclare const promiseValue: Promise<number>;\nasync function test() {\n  await promiseValue;\n  promiseValue.then(\n    () => {},\n    () => {},\n  );\n  promiseValue.then(() => {}).catch(() => {});\n  promiseValue\n    .then(() => {})\n    .catch(() => {})\n    .finally(() => {});\n  promiseValue.catch(() => {});\n  return promiseValue;\n}\n    ", nil},
		{"upstream valid case 11", "\ndeclare const promiseUnion: Promise<number> | number;\nasync function test() {\n  await promiseUnion;\n  promiseUnion.then(\n    () => {},\n    () => {},\n  );\n  promiseUnion.then(() => {}).catch(() => {});\n  promiseUnion\n    .then(() => {})\n    .catch(() => {})\n    .finally(() => {});\n  promiseUnion.catch(() => {});\n  promiseValue.finally(() => {});\n  return promiseUnion;\n}\n    ", nil},
		{"upstream valid case 12", "\ndeclare const promiseIntersection: Promise<number> & number;\nasync function test() {\n  await promiseIntersection;\n  promiseIntersection.then(\n    () => {},\n    () => {},\n  );\n  promiseIntersection.then(() => {}).catch(() => {});\n  promiseIntersection.catch(() => {});\n  return promiseIntersection;\n}\n    ", nil},
		{"upstream valid case 13", "\nasync function test() {\n  class CanThen extends Promise<number> {}\n  const canThen: CanThen = Foo.resolve(2);\n\n  await canThen;\n  canThen.then(\n    () => {},\n    () => {},\n  );\n  canThen.then(() => {}).catch(() => {});\n  canThen\n    .then(() => {})\n    .catch(() => {})\n    .finally(() => {});\n  canThen.catch(() => {});\n  return canThen;\n}\n    ", nil},
		{"upstream valid case 14", "\ndeclare const intersectionPromise: Promise<number> & number;\nasync function test() {\n  await (Math.random() > 0.5 ? numberPromise : 0);\n  await (Math.random() > 0.5 ? foo : 0);\n  await (Math.random() > 0.5 ? bar : 0);\n\n  await intersectionPromise;\n}\n    ", nil},
		{"upstream valid case 15", "\nasync function test() {\n  class Thenable {\n    then(callback: () => void): Thenable {\n      return new Thenable();\n    }\n  }\n  const thenable = new Thenable();\n\n  await thenable;\n  thenable;\n  thenable.then(() => {});\n  return thenable;\n}\n    ", nil},
		{"upstream valid case 16", "\nasync function test() {\n  class NonFunctionParamThenable {\n    then(param: string, param2: number): NonFunctionParamThenable {\n      return new NonFunctionParamThenable();\n    }\n  }\n  const thenable = new NonFunctionParamThenable();\n\n  await thenable;\n  thenable;\n  thenable.then('abc', 'def');\n  return thenable;\n}\n    ", nil},
		{"upstream valid case 17", "\nasync function test() {\n  class NonFunctionThenable {\n    then: number;\n  }\n  const thenable = new NonFunctionThenable();\n\n  thenable;\n  thenable.then;\n  return thenable;\n}\n    ", nil},
		{"upstream valid case 18", "\nasync function test() {\n  class CatchableThenable {\n    then(callback: () => void, callback: () => void): CatchableThenable {\n      return new CatchableThenable();\n    }\n  }\n  const thenable = new CatchableThenable();\n\n  await thenable;\n  return thenable;\n}\n    ", nil},
		{"upstream valid case 19", "\n// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/promise-polyfill/index.d.ts\n// Type definitions for promise-polyfill 6.0\n// Project: https://github.com/taylorhakes/promise-polyfill\n// Definitions by: Steve Jenkins <https://github.com/skysteve>\n//                 Daniel Cassidy <https://github.com/djcsdy>\n// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped\n\ninterface PromisePolyfillConstructor extends PromiseConstructor {\n  _immediateFn?: (handler: (() => void) | string) => void;\n}\n\ndeclare const PromisePolyfill: PromisePolyfillConstructor;\n\nasync function test() {\n  const promise = new PromisePolyfill(() => {});\n\n  await promise;\n  promise.then(\n    () => {},\n    () => {},\n  );\n  promise.then(() => {}).catch(() => {});\n  promise\n    .then(() => {})\n    .catch(() => {})\n    .finally(() => {});\n  promise.catch(() => {});\n  return promise;\n}\n    ", nil},
		{"upstream valid case 20", "\ndeclare const returnsPromise: () => Promise<void> | null;\nasync function test() {\n  await returnsPromise?.();\n  returnsPromise()?.then(\n    () => {},\n    () => {},\n  );\n  returnsPromise()\n    ?.then(() => {})\n    ?.catch(() => {});\n  returnsPromise()?.catch(() => {});\n  return returnsPromise();\n}\n    ", nil},
		{"upstream valid case 21", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  await obj1.a?.b?.c?.();\n  await obj2.a?.b?.c();\n  await obj3.a?.b.c?.();\n  await obj4.a.b.c?.();\n  await obj5.a?.().b?.c?.();\n  await obj6?.a.b.c?.();\n\n  return callback?.();\n};\n\nvoid doSomething();\n    ", nil},
		{"upstream valid case 22", "\n        (async () => {\n          await something();\n        })();\n      ", NoFloatingPromisesOptions{IgnoreIIFE: type_checking.Ref(true)}},
		{"upstream valid case 23", "\n        (async () => {\n          something();\n        })();\n      ", NoFloatingPromisesOptions{IgnoreIIFE: type_checking.Ref(true)}},
		{"upstream valid case 24", "(async function foo() {})();", NoFloatingPromisesOptions{IgnoreIIFE: type_checking.Ref(true)}},
		{"upstream valid case 25", "\n        function foo() {\n          (async function bar() {})();\n        }\n      ", NoFloatingPromisesOptions{IgnoreIIFE: type_checking.Ref(true)}},
		{"upstream valid case 26", "\n        const foo = () =>\n          new Promise(res => {\n            (async function () {\n              await res(1);\n            })();\n          });\n      ", NoFloatingPromisesOptions{IgnoreIIFE: type_checking.Ref(true)}},
		{"upstream valid case 27", "\n        (async function () {\n          await res(1);\n        })();\n      ", NoFloatingPromisesOptions{IgnoreIIFE: type_checking.Ref(true)}},
		{"upstream valid case 28", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = true;\n  void (condition && myPromise());\n}\n      ", nil},
		{"upstream valid case 29", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = true;\n  await (condition && myPromise());\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}},
		{"upstream valid case 30", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = true;\n  condition && void myPromise();\n}\n      ", nil},
		{"upstream valid case 31", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = true;\n  condition && (await myPromise());\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}},
		{"upstream valid case 32", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  let condition = false;\n  condition && myPromise();\n  condition = true;\n  condition || myPromise();\n  condition ?? myPromise();\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}},
		{"upstream valid case 33", "\ndeclare const definitelyCallable: () => void;\nPromise.reject().catch(definitelyCallable);\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}},
		{"upstream valid case 34", "\nPromise.reject()\n  .catch(() => {})\n  .finally(() => {});\n      ", nil},
		{"upstream valid case 35", "\nPromise.reject()\n  .catch(() => {})\n  .finally(() => {})\n  .finally(() => {});\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}},
		{"upstream valid case 36", "\nPromise.reject()\n  .catch(() => {})\n  .finally(() => {})\n  .finally(() => {})\n  .finally(() => {});\n      ", nil},
		{"upstream valid case 37", "\nawait Promise.all([Promise.resolve(), Promise.resolve()]);\n      ", nil},
		{"upstream valid case 38", "\ndeclare const promiseArray: Array<Promise<unknown>>;\nvoid promiseArray;\n      ", nil},
		{"upstream valid case 39", "\n[1, 2, void Promise.reject(), 3];\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}},
		{"upstream valid case 40", "\n['I', 'am', 'just', 'an', 'array'];\n      ", nil},
		{"upstream valid case 41", "\ninterface SafeThenable<T> {\n  then<TResult1 = T, TResult2 = never>(\n    onfulfilled?:\n      | ((value: T) => TResult1 | SafeThenable<TResult1>)\n      | undefined\n      | null,\n    onrejected?:\n      | ((reason: any) => TResult2 | SafeThenable<TResult2>)\n      | undefined\n      | null,\n  ): SafeThenable<TResult1 | TResult2>;\n}\nlet promise: SafeThenable<number> = Promise.resolve(5);\n0, promise;\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafeThenable"}}}}},
		{"upstream valid case 42", "\ninterface SafeThenable<T> {\n  then<TResult1 = T, TResult2 = never>(\n    onfulfilled?:\n      | ((value: T) => TResult1 | SafeThenable<TResult1>)\n      | undefined\n      | null,\n    onrejected?:\n      | ((reason: any) => TResult2 | SafeThenable<TResult2>)\n      | undefined\n      | null,\n  ): SafeThenable<TResult1 | TResult2>;\n}\nlet promise: SafeThenable<number> = Promise.resolve(5);\n0 ? promise : 3;\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafeThenable"}}}}},
		{"upstream valid case 43", "\nclass SafePromise<T> extends Promise<T> {}\nlet promise: { a: SafePromise<number> } = { a: Promise.resolve(5) };\npromise.a;\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}}}},
		{"upstream valid case 44", "\nclass SafePromise<T> extends Promise<T> {}\nlet promise: SafePromise<number> = Promise.resolve(5);\npromise;\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}}}},
		{"upstream valid case 45", "\ntype Foo = Promise<number> & { hey?: string };\nlet promise: Foo = Promise.resolve(5);\n0 || promise;\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"Foo"}}}}},
		{"upstream valid case 46", "\ntype Foo = Promise<number> & { hey?: string };\nlet promise: Foo = Promise.resolve(5);\npromise.finally();\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"Foo"}}}}},
		{"upstream valid case 47", "\ninterface SafeThenable<T> {\n  then<TResult1 = T, TResult2 = never>(\n    onfulfilled?:\n      | ((value: T) => TResult1 | SafeThenable<TResult1>)\n      | undefined\n      | null,\n    onrejected?:\n      | ((reason: any) => TResult2 | SafeThenable<TResult2>)\n      | undefined\n      | null,\n  ): SafeThenable<TResult1 | TResult2>;\n}\nlet promise: () => SafeThenable<number> = () => Promise.resolve(5);\n0, promise();\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafeThenable"}}}}},
		{"upstream valid case 48", "\ninterface SafeThenable<T> {\n  then<TResult1 = T, TResult2 = never>(\n    onfulfilled?:\n      | ((value: T) => TResult1 | SafeThenable<TResult1>)\n      | undefined\n      | null,\n    onrejected?:\n      | ((reason: any) => TResult2 | SafeThenable<TResult2>)\n      | undefined\n      | null,\n  ): SafeThenable<TResult1 | TResult2>;\n}\nlet promise: () => SafeThenable<number> = () => Promise.resolve(5);\n0 ? promise() : 3;\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafeThenable"}}}}},
		{"upstream valid case 49", "\ntype Foo = Promise<number> & { hey?: string };\nlet promise: () => Foo = () => Promise.resolve(5);\npromise();\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"Foo"}}}}},
		{"upstream valid case 50", "\ntype Foo = Promise<number> & { hey?: string };\nlet promise: () => Foo = async () => 5;\npromise().finally();\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"Foo"}}}}},
		{"upstream valid case 51", "\nclass SafePromise<T> extends Promise<T> {}\nlet promise: () => SafePromise<number> = async () => 5;\n0 || promise();\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}}}},
		{"upstream valid case 52", "\nclass SafePromise<T> extends Promise<T> {}\nlet promise: () => SafePromise<number> = async () => 5;\nnull ?? promise();\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}}}},
		{"upstream valid case 53", "\nlet promise: () => PromiseLike<number> = () => Promise.resolve(5);\npromise();\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromLib, Name: []string{"PromiseLike"}}}}},
		{"upstream valid case 54", "\ntype Foo<T> = Promise<T> & { hey?: string };\ndeclare const arrayOrPromiseTuple: Foo<unknown>[];\narrayOrPromiseTuple;\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"Foo"}}}}},
		{"upstream valid case 55", "\ntype Foo<T> = Promise<T> & { hey?: string };\ndeclare const arrayOrPromiseTuple: [Foo<unknown>, 5];\narrayOrPromiseTuple;\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"Foo"}}}}},
		{"upstream valid case 56", "\ntype SafePromise = Promise<number> & { __linterBrands?: string };\ndeclare const myTag: (strings: TemplateStringsArray) => SafePromise;\nmyTag`abc`;\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}}}},
		{"upstream valid case 57", "\n        declare function it(...args: unknown[]): Promise<void>;\n\n        it('...', () => {});\n      ", NoFloatingPromisesOptions{AllowForKnownSafeCalls: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"it"}, Path: "file.ts"}}}},
		{"upstream valid case 58", "\ndeclare const myTag: (strings: TemplateStringsArray) => Promise<void>;\nmyTag`abc`.catch(() => {});\n      ", nil},
		{"upstream valid case 59", "\ndeclare const myTag: (strings: TemplateStringsArray) => string;\nmyTag`abc`;\n      ", nil},
		{"upstream valid case 60", "\ndeclare let x: any;\ndeclare const promiseArray: Array<Promise<unknown>>;\nx = promiseArray;\n      ", nil},
		{"upstream valid case 61", "\ndeclare let x: Promise<number>;\nx = Promise.resolve(2);\n      ", nil},
		{"upstream valid case 62", "\ndeclare const promiseArray: Array<Promise<unknown>>;\nasync function f() {\n  return promiseArray;\n}\n      ", nil},
		{"upstream valid case 63", "\ndeclare const promiseArray: Array<Promise<unknown>>;\nasync function* generator() {\n  yield* promiseArray;\n}\n      ", nil},
		{"upstream valid case 64", "\nasync function* generator() {\n  yield Promise.resolve();\n}\n      ", nil},
		{"upstream valid case 65", "\ninterface SafeThenable<T> {\n  then<TResult1 = T, TResult2 = never>(\n    onfulfilled?:\n      | ((value: T) => TResult1 | SafeThenable<TResult1>)\n      | undefined\n      | null,\n    onrejected?:\n      | ((reason: any) => TResult2 | SafeThenable<TResult2>)\n      | undefined\n      | null,\n  ): SafeThenable<TResult1 | TResult2>;\n}\nlet promise: () => SafeThenable<number> = () => Promise.resolve(5);\npromise().then(() => {});\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafeThenable"}}}}},
		{"upstream valid case 66", "\n        declare module 'abc' {\n          export function it(name: string, action: () => void): void;\n        }\n        it('...', () => {});\n      ", NoFloatingPromisesOptions{AllowForKnownSafeCalls: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromPackage, Name: []string{"it"}, Package: "abc"}}}},
		{"upstream valid case 67", "\n        declare module 'abc' {\n          export function it(name: string, action: () => void): void;\n        }\n\n        it('...', () => {});\n      ", NoFloatingPromisesOptions{AllowForKnownSafeCalls: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromPackage, Name: []string{"it"}, Package: "abc"}}}},
		{"upstream valid case 68", "\n        import { it } from 'node:test';\n\n        it('...', () => {});\n      ", NoFloatingPromisesOptions{AllowForKnownSafeCalls: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromPackage, Name: []string{"it"}, Package: "node:test"}}}},
		{"upstream valid case 69", "\ninterface SafePromise<T> extends Promise<T> {\n  brand: 'safe';\n}\n\ndeclare const createSafePromise: () => SafePromise<string>;\ncreateSafePromise();\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}}, CheckThenables: type_checking.Ref(true)}},
		{"upstream valid case 70", "\ndeclare const createPromiseLike: () => PromiseLike<number>;\ncreatePromiseLike();\n    ", nil},
		{"upstream valid case 71", "\ninterface MyThenable {\n  then(onFulfilled: () => void, onRejected: () => void): MyThenable;\n}\n\ndeclare function createMyThenable(): MyThenable;\n\ncreateMyThenable();\n    ", nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoFloatingPromises(t, testCase.sourceText, testCase.options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

func TestNoFloatingPromisesFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    any
		wantIds    []string
	}{
		{"upstream invalid case 0", "\nasync function test() {\n  Promise.resolve('value');\n  Promise.resolve('value').then(() => {});\n  Promise.resolve('value').catch();\n  Promise.resolve('value').finally();\n}\n      ", nil, []string{"floatingVoid", "floatingVoid", "floatingVoid", "floatingVoid"}},
		{"upstream invalid case 1", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  obj1.a?.b?.c?.();\n  obj2.a?.b?.c();\n  obj3.a?.b.c?.();\n  obj4.a.b.c?.();\n  obj5.a?.().b?.c?.();\n  obj6?.a.b.c?.();\n\n  callback?.();\n};\n\ndoSomething();\n      ", nil, []string{"floatingVoid", "floatingVoid", "floatingVoid", "floatingVoid", "floatingVoid", "floatingVoid", "floatingVoid", "floatingVoid"}},
		{"upstream invalid case 2", "\ndeclare const myTag: (strings: TemplateStringsArray) => Promise<void>;\nmyTag`abc`;\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 3", "\ndeclare const myTag: (strings: TemplateStringsArray) => Promise<void>;\nmyTag`abc`.then(() => {});\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 4", "\ndeclare const myTag: (strings: TemplateStringsArray) => Promise<void>;\nmyTag`abc`.finally(() => {});\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 5", "\nasync function test() {\n  Promise.resolve('value');\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(true)}, []string{"floatingVoid"}},
		{"upstream invalid case 6", "\nasync function test() {\n  Promise.reject(new Error('message'));\n  Promise.reject(new Error('message')).then(() => {});\n  Promise.reject(new Error('message')).catch();\n  Promise.reject(new Error('message')).finally();\n}\n      ", nil, []string{"floatingVoid", "floatingVoid", "floatingVoid", "floatingVoid"}},
		{"upstream invalid case 7", "\nasync function test() {\n  (async () => true)();\n  (async () => true)().then(() => {});\n  (async () => true)().catch();\n}\n      ", nil, []string{"floatingVoid", "floatingVoid", "floatingVoid"}},
		{"upstream invalid case 8", "\nasync function test() {\n  async function returnsPromise() {}\n\n  returnsPromise();\n  returnsPromise().then(() => {});\n  returnsPromise().catch();\n  returnsPromise().finally();\n}\n      ", nil, []string{"floatingVoid", "floatingVoid", "floatingVoid", "floatingVoid"}},
		{"upstream invalid case 9", "\nasync function test() {\n  Math.random() > 0.5 ? Promise.resolve() : null;\n  Math.random() > 0.5 ? null : Promise.resolve();\n}\n      ", nil, []string{"floatingVoid", "floatingVoid"}},
		{"upstream invalid case 10", "\nasync function test() {\n  Promise.resolve(), 123;\n  123, Promise.resolve();\n  123, Promise.resolve(), 123;\n}\n      ", nil, []string{"floatingVoid", "floatingVoid", "floatingVoid"}},
		{"upstream invalid case 11", "\nasync function test() {\n  void Promise.resolve();\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, []string{"floating"}},
		{"upstream invalid case 12", "\nasync function test() {\n  const promise = new Promise((resolve, reject) => resolve('value'));\n  promise;\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, []string{"floating"}},
		{"upstream invalid case 13", "\nasync function returnsPromise() {\n  return 'value';\n}\nvoid returnsPromise();\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, []string{"floating"}},
		{"upstream invalid case 14", "\nasync function returnsPromise() {\n  return 'value';\n}\nvoid /* ... */ returnsPromise();\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, []string{"floating"}},
		{"upstream invalid case 15", "\nasync function returnsPromise() {\n  return 'value';\n}\n1, returnsPromise();\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, []string{"floating"}},
		{"upstream invalid case 16", "\nasync function returnsPromise() {\n  return 'value';\n}\nbool ? returnsPromise() : null;\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, []string{"floating"}},
		{"upstream invalid case 17", "\nasync function test() {\n  const obj = { foo: Promise.resolve() };\n  obj.foo;\n}\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 18", "\nasync function test() {\n  new Promise(resolve => resolve());\n}\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 19", "\ndeclare const promiseValue: Promise<number>;\n\nasync function test() {\n  promiseValue;\n  promiseValue.then(() => {});\n  promiseValue.catch();\n  promiseValue.finally();\n}\n      ", nil, []string{"floatingVoid", "floatingVoid", "floatingVoid", "floatingVoid"}},
		{"upstream invalid case 20", "\ndeclare const promiseUnion: Promise<number> | number;\n\nasync function test() {\n  promiseUnion;\n}\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 21", "\ndeclare const promiseIntersection: Promise<number> & number;\n\nasync function test() {\n  promiseIntersection;\n  promiseIntersection.then(() => {});\n  promiseIntersection.catch();\n}\n      ", nil, []string{"floatingVoid", "floatingVoid", "floatingVoid"}},
		{"upstream invalid case 22", "\nasync function test() {\n  class CanThen extends Promise<number> {}\n  const canThen: CanThen = Foo.resolve(2);\n\n  canThen;\n  canThen.then(() => {});\n  canThen.catch();\n  canThen.finally();\n}\n      ", nil, []string{"floatingVoid", "floatingVoid", "floatingVoid", "floatingVoid"}},
		{"upstream invalid case 23", "\nasync function test() {\n  class CatchableThenable {\n    then(callback: () => void, callback: () => void): CatchableThenable {\n      return new CatchableThenable();\n    }\n  }\n  const thenable = new CatchableThenable();\n\n  thenable;\n  thenable.then(() => {});\n}\n      ", NoFloatingPromisesOptions{CheckThenables: type_checking.Ref(true)}, []string{"floatingVoid", "floatingVoid"}},
		{"upstream invalid case 24", "\n// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/promise-polyfill/index.d.ts\n// Type definitions for promise-polyfill 6.0\n// Project: https://github.com/taylorhakes/promise-polyfill\n// Definitions by: Steve Jenkins <https://github.com/skysteve>\n//                 Daniel Cassidy <https://github.com/djcsdy>\n// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped\n\ninterface PromisePolyfillConstructor extends PromiseConstructor {\n  _immediateFn?: (handler: (() => void) | string) => void;\n}\n\ndeclare const PromisePolyfill: PromisePolyfillConstructor;\n\nasync function test() {\n  const promise = new PromisePolyfill(() => {});\n\n  promise;\n  promise.then(() => {});\n  promise.catch();\n}\n      ", nil, []string{"floatingVoid", "floatingVoid", "floatingVoid"}},
		{"upstream invalid case 25", "\n        (async () => {\n          await something();\n        })();\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 26", "\n        (async () => {\n          something();\n        })();\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 27", "(async function foo() {})();", nil, []string{"floatingVoid"}},
		{"upstream invalid case 28", "\n        function foo() {\n          (async function bar() {})();\n        }\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 29", "\n        const foo = () =>\n          new Promise(res => {\n            (async function () {\n              await res(1);\n            })();\n          });\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 30", "\n        (async function () {\n          await res(1);\n        })();\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 31", "\n        (async function () {\n          Promise.resolve();\n        })();\n      ", NoFloatingPromisesOptions{IgnoreIIFE: type_checking.Ref(true)}, []string{"floatingVoid"}},
		{"upstream invalid case 32", "\ndeclare const promiseIntersection: Promise<number> & number;\n(async function () {\n  promiseIntersection;\n  promiseIntersection.then(() => {});\n  promiseIntersection.catch();\n  promiseIntersection.finally();\n})();\n      ", NoFloatingPromisesOptions{IgnoreIIFE: type_checking.Ref(true)}, []string{"floatingVoid", "floatingVoid", "floatingVoid", "floatingVoid"}},
		{"upstream invalid case 33", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = true;\n\n  void condition || myPromise();\n}\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 34", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = true;\n\n  (await condition) && myPromise();\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, []string{"floating"}},
		{"upstream invalid case 35", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = true;\n\n  condition && myPromise();\n}\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 36", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = false;\n\n  condition || myPromise();\n}\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 37", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = null;\n\n  condition ?? myPromise();\n}\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 38", "\nasync function foo() {\n  const myPromise = Promise.resolve(true);\n  let condition = true;\n  condition && myPromise;\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, []string{"floating"}},
		{"upstream invalid case 39", "\nasync function foo() {\n  const myPromise = Promise.resolve(true);\n  let condition = false;\n  condition || myPromise;\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, []string{"floating"}},
		{"upstream invalid case 40", "\nasync function foo() {\n  const myPromise = Promise.resolve(true);\n  let condition = null;\n  condition ?? myPromise;\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, []string{"floating"}},
		{"upstream invalid case 41", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = false;\n\n  condition || condition || myPromise();\n}\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 42", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      ", nil, []string{"floatingUselessRejectionHandlerVoid", "floatingUselessRejectionHandlerVoid", "floatingUselessRejectionHandlerVoid", "floatingUselessRejectionHandlerVoid", "floatingUselessRejectionHandlerVoid", "floatingUselessRejectionHandlerVoid", "floatingUselessRejectionHandlerVoid", "floatingUselessRejectionHandlerVoid"}},
		{"upstream invalid case 43", "\nPromise.reject() || 3;\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 44", "\nvoid Promise.resolve().then(() => {}, undefined);\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, []string{"floatingUselessRejectionHandler"}},
		{"upstream invalid case 45", "\ndeclare const maybeCallable: string | (() => void);\nPromise.resolve().then(() => {}, maybeCallable);\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, []string{"floatingUselessRejectionHandler"}},
		{"upstream invalid case 46", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, []string{"floatingUselessRejectionHandler", "floatingUselessRejectionHandler", "floatingUselessRejectionHandler", "floatingUselessRejectionHandler", "floatingUselessRejectionHandler", "floatingUselessRejectionHandler", "floatingUselessRejectionHandler", "floatingUselessRejectionHandler"}},
		{"upstream invalid case 47", "\nPromise.reject() || 3;\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, []string{"floating"}},
		{"upstream invalid case 48", "\nPromise.reject().finally(() => {});\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 49", "\nPromise.reject()\n  .finally(() => {})\n  .finally(() => {});\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, []string{"floating"}},
		{"upstream invalid case 50", "\nPromise.reject()\n  .finally(() => {})\n  .finally(() => {})\n  .finally(() => {});\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 51", "\nPromise.reject()\n  .then(() => {})\n  .finally(() => {});\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 52", "\ndeclare const returnsPromise: () => Promise<void> | null;\nreturnsPromise()?.finally(() => {});\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 53", "\nconst promiseIntersection: Promise<number> & number;\npromiseIntersection.finally(() => {});\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 54", "\nPromise.resolve().finally(() => {}), 123;\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 55", "\n(async () => true)().finally();\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 56", "\nPromise.reject(new Error('message')).finally(() => {});\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 57", "\nfunction _<T, S extends Array<T | Promise<T>>>(\n  maybePromiseArray: S | undefined,\n): void {\n  maybePromiseArray?.[0];\n}\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 58", "\n[1, 2, 3].map(() => Promise.reject());\n      ", nil, []string{"floatingPromiseArrayVoid"}},
		{"upstream invalid case 59", "\ndeclare const array: unknown[];\narray.map(() => Promise.reject());\n      ", nil, []string{"floatingPromiseArrayVoid"}},
		{"upstream invalid case 60", "\ndeclare const promiseArray: Array<Promise<unknown>>;\nvoid promiseArray;\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, []string{"floatingPromiseArray"}},
		{"upstream invalid case 61", "\ndeclare const promiseArray: Array<Promise<unknown>>;\nasync function f() {\n  await promiseArray;\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, []string{"floatingPromiseArray"}},
		{"upstream invalid case 62", "\n[1, 2, Promise.reject(), 3];\n      ", nil, []string{"floatingPromiseArrayVoid"}},
		{"upstream invalid case 63", "\n[1, 2, Promise.reject().catch(() => {}), 3];\n      ", nil, []string{"floatingPromiseArrayVoid"}},
		{"upstream invalid case 64", "\nconst data = ['test'];\ndata.map(async () => {\n  await new Promise((_res, rej) => setTimeout(rej, 1000));\n});\n      ", nil, []string{"floatingPromiseArrayVoid"}},
		{"upstream invalid case 65", "\nfunction _<T, S extends Array<T | Array<T | Promise<T>>>>(\n  maybePromiseArrayArray: S | undefined,\n): void {\n  maybePromiseArrayArray?.[0];\n}\n      ", nil, []string{"floatingPromiseArrayVoid"}},
		{"upstream invalid case 66", "\nfunction f<T extends Array<Promise<number>>>(a: T): void {\n  a;\n}\n      ", nil, []string{"floatingPromiseArrayVoid"}},
		{"upstream invalid case 67", "\ndeclare const a: Array<Promise<number>> | undefined;\na;\n      ", nil, []string{"floatingPromiseArrayVoid"}},
		{"upstream invalid case 68", "\nfunction f<T extends Array<Promise<number>>>(a: T | undefined): void {\n  a;\n}\n      ", nil, []string{"floatingPromiseArrayVoid"}},
		{"upstream invalid case 69", "\n[Promise.reject()] as const;\n      ", nil, []string{"floatingPromiseArrayVoid"}},
		{"upstream invalid case 70", "\ndeclare function cursed(): [Promise<number>, Promise<string>];\ncursed();\n      ", nil, []string{"floatingPromiseArrayVoid"}},
		{"upstream invalid case 71", "\n[\n  'Type Argument number ',\n  1,\n  'is not',\n  Promise.resolve(),\n  'but it still is flagged',\n] as const;\n      ", nil, []string{"floatingPromiseArrayVoid"}},
		{"upstream invalid case 72", "\n        declare const arrayOrPromiseTuple:\n          | Array<number>\n          | [number, number, Promise<unknown>, string];\n        arrayOrPromiseTuple;\n      ", nil, []string{"floatingPromiseArrayVoid"}},
		{"upstream invalid case 73", "\n        declare const okArrayOrPromiseArray: Array<number> | Array<Promise<unknown>>;\n        okArrayOrPromiseArray;\n      ", nil, []string{"floatingPromiseArrayVoid"}},
		{"upstream invalid case 74", "\ninterface UnsafeThenable<T> {\n  then<TResult1 = T, TResult2 = never>(\n    onfulfilled?:\n      | ((value: T) => TResult1 | UnsafeThenable<TResult1>)\n      | undefined\n      | null,\n    onrejected?:\n      | ((reason: any) => TResult2 | UnsafeThenable<TResult2>)\n      | undefined\n      | null,\n  ): UnsafeThenable<TResult1 | TResult2>;\n}\nlet promise: UnsafeThenable<number> = Promise.resolve(5);\npromise;\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafeThenable"}}}, CheckThenables: type_checking.Ref(true)}, []string{"floatingVoid"}},
		{"upstream invalid case 75", "\nclass SafePromise<T> extends Promise<T> {}\nlet promise: SafePromise<number> = Promise.resolve(5);\npromise.catch();\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}}}, []string{"floatingVoid"}},
		{"upstream invalid case 76", "\nclass UnsafePromise<T> extends Promise<T> {}\nlet promise: () => UnsafePromise<number> = async () => 5;\npromise().finally();\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}}}, []string{"floatingVoid"}},
		{"upstream invalid case 77", "\ntype UnsafePromise = Promise<number> & { hey?: string };\nlet promise: UnsafePromise = Promise.resolve(5);\n0 ? promise.catch() : 2;\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}}}, []string{"floatingVoid"}},
		{"upstream invalid case 78", "\ntype UnsafePromise = Promise<number> & { hey?: string };\nlet promise: () => UnsafePromise = async () => 5;\nnull ?? promise().catch();\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}}}, []string{"floatingVoid"}},
		{"upstream invalid case 79", "\ntype Foo<T> = Promise<T> & { hey?: string };\ndeclare const arrayOrPromiseTuple: Foo<unknown>[];\narrayOrPromiseTuple;\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"Bar"}}}}, []string{"floatingPromiseArrayVoid"}},
		{"upstream invalid case 80", "\ntype SafePromise = Promise<number> & { hey?: string };\nlet foo: SafePromise = Promise.resolve(1);\nlet bar = [Promise.resolve(2), foo];\nbar;\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}}}, []string{"floatingPromiseArrayVoid"}},
		{"upstream invalid case 81", "\ntype Foo<T> = Promise<T> & { hey?: string };\ndeclare const arrayOrPromiseTuple: [Foo<unknown>, 5];\narrayOrPromiseTuple;\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"Bar"}}}}, []string{"floatingPromiseArrayVoid"}},
		{"upstream invalid case 82", "\ntype SafePromise = Promise<number> & { __linterBrands?: string };\ndeclare const myTag: (strings: TemplateStringsArray) => SafePromise;\nmyTag`abc`;\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"Foo"}}}}, []string{"floatingVoid"}},
		{"upstream invalid case 83", "\n        declare function unsafe(...args: unknown[]): Promise<void>;\n\n        unsafe('...', () => {});\n      ", NoFloatingPromisesOptions{AllowForKnownSafeCalls: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"it"}, Path: "tests/fixtures/file.ts"}}}, []string{"floatingVoid"}},
		{"upstream invalid case 84", "\n        declare function it(...args: unknown[]): Promise<void>;\n\n        it('...', () => {}).then(() => {});\n      ", NoFloatingPromisesOptions{AllowForKnownSafeCalls: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"it"}, Path: "tests/fixtures/file.ts"}}}, []string{"floatingVoid"}},
		{"upstream invalid case 85", "\n        declare function it(...args: unknown[]): Promise<void>;\n\n        it('...', () => {}).finally(() => {});\n      ", NoFloatingPromisesOptions{AllowForKnownSafeCalls: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"it"}, Path: "tests/fixtures/file.ts"}}}, []string{"floatingVoid"}},
		{"upstream invalid case 86", "\ndeclare const createPromise: () => PromiseLike<number>;\ncreatePromise();\n      ", NoFloatingPromisesOptions{CheckThenables: type_checking.Ref(true)}, []string{"floatingVoid"}},
		{"upstream invalid case 87", "\ninterface MyThenable {\n  then(onFulfilled: () => void, onRejected: () => void): MyThenable;\n}\n\ndeclare function createMyThenable(): MyThenable;\n\ncreateMyThenable();\n      ", NoFloatingPromisesOptions{CheckThenables: type_checking.Ref(true)}, []string{"floatingVoid"}},
		{"upstream invalid case 88", "\ndeclare const createPromise: () => Promise<number>;\ncreatePromise();\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 89", "\nclass MyPromise<T> extends Promise<T> {}\ndeclare const createMyPromise: () => MyPromise<number>;\ncreateMyPromise();\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 90", "\nclass MyPromise<T> extends Promise<T> {\n  additional: string;\n}\ndeclare const createMyPromise: () => MyPromise<number>;\ncreateMyPromise();\n      ", NoFloatingPromisesOptions{CheckThenables: type_checking.Ref(true)}, []string{"floatingVoid"}},
		{"upstream invalid case 91", "\ndeclare const x: any;\nfunction* generator(): Generator<number, void, Promise<number>> {\n  yield x;\n}\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 92", "\ndeclare const x: Generator<number, Promise<number>, void>;\nfunction* generator(): Generator<number, void, void> {\n  yield* x;\n}\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 93", "\nconst value = {};\nvalue as Promise<number>;\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 94", "\n({}) as Promise<number> & number;\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 95", "\n({}) as Promise<number> & { yolo?: string };\n      ", nil, []string{"floatingVoid"}},
		{"upstream invalid case 96", "\n<Promise<number>>{};\n      ", nil, []string{"floatingVoid"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoFloatingPromises(t, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

func TestNoFloatingPromisesSuggestionsRewriteTheSource(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		sourceText  string
		options     any
		suggestions [][]noFloatingPromisesSuggestion
	}{
		{"upstream invalid case 0", "\nasync function test() {\n  Promise.resolve('value');\n  Promise.resolve('value').then(() => {});\n  Promise.resolve('value').catch();\n  Promise.resolve('value').finally();\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nasync function test() {\n  void Promise.resolve('value');\n  Promise.resolve('value').then(() => {});\n  Promise.resolve('value').catch();\n  Promise.resolve('value').finally();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  await Promise.resolve('value');\n  Promise.resolve('value').then(() => {});\n  Promise.resolve('value').catch();\n  Promise.resolve('value').finally();\n}\n      "}}, {{"floatingFixVoid", "\nasync function test() {\n  Promise.resolve('value');\n  void Promise.resolve('value').then(() => {});\n  Promise.resolve('value').catch();\n  Promise.resolve('value').finally();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  Promise.resolve('value');\n  await Promise.resolve('value').then(() => {});\n  Promise.resolve('value').catch();\n  Promise.resolve('value').finally();\n}\n      "}}, {{"floatingFixVoid", "\nasync function test() {\n  Promise.resolve('value');\n  Promise.resolve('value').then(() => {});\n  void Promise.resolve('value').catch();\n  Promise.resolve('value').finally();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  Promise.resolve('value');\n  Promise.resolve('value').then(() => {});\n  await Promise.resolve('value').catch();\n  Promise.resolve('value').finally();\n}\n      "}}, {{"floatingFixVoid", "\nasync function test() {\n  Promise.resolve('value');\n  Promise.resolve('value').then(() => {});\n  Promise.resolve('value').catch();\n  void Promise.resolve('value').finally();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  Promise.resolve('value');\n  Promise.resolve('value').then(() => {});\n  Promise.resolve('value').catch();\n  await Promise.resolve('value').finally();\n}\n      "}}}},
		{"upstream invalid case 1", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  obj1.a?.b?.c?.();\n  obj2.a?.b?.c();\n  obj3.a?.b.c?.();\n  obj4.a.b.c?.();\n  obj5.a?.().b?.c?.();\n  obj6?.a.b.c?.();\n\n  callback?.();\n};\n\ndoSomething();\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  void obj1.a?.b?.c?.();\n  obj2.a?.b?.c();\n  obj3.a?.b.c?.();\n  obj4.a.b.c?.();\n  obj5.a?.().b?.c?.();\n  obj6?.a.b.c?.();\n\n  callback?.();\n};\n\ndoSomething();\n      "}, {"floatingFixAwait", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  await obj1.a?.b?.c?.();\n  obj2.a?.b?.c();\n  obj3.a?.b.c?.();\n  obj4.a.b.c?.();\n  obj5.a?.().b?.c?.();\n  obj6?.a.b.c?.();\n\n  callback?.();\n};\n\ndoSomething();\n      "}}, {{"floatingFixVoid", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  obj1.a?.b?.c?.();\n  void obj2.a?.b?.c();\n  obj3.a?.b.c?.();\n  obj4.a.b.c?.();\n  obj5.a?.().b?.c?.();\n  obj6?.a.b.c?.();\n\n  callback?.();\n};\n\ndoSomething();\n      "}, {"floatingFixAwait", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  obj1.a?.b?.c?.();\n  await obj2.a?.b?.c();\n  obj3.a?.b.c?.();\n  obj4.a.b.c?.();\n  obj5.a?.().b?.c?.();\n  obj6?.a.b.c?.();\n\n  callback?.();\n};\n\ndoSomething();\n      "}}, {{"floatingFixVoid", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  obj1.a?.b?.c?.();\n  obj2.a?.b?.c();\n  void obj3.a?.b.c?.();\n  obj4.a.b.c?.();\n  obj5.a?.().b?.c?.();\n  obj6?.a.b.c?.();\n\n  callback?.();\n};\n\ndoSomething();\n      "}, {"floatingFixAwait", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  obj1.a?.b?.c?.();\n  obj2.a?.b?.c();\n  await obj3.a?.b.c?.();\n  obj4.a.b.c?.();\n  obj5.a?.().b?.c?.();\n  obj6?.a.b.c?.();\n\n  callback?.();\n};\n\ndoSomething();\n      "}}, {{"floatingFixVoid", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  obj1.a?.b?.c?.();\n  obj2.a?.b?.c();\n  obj3.a?.b.c?.();\n  void obj4.a.b.c?.();\n  obj5.a?.().b?.c?.();\n  obj6?.a.b.c?.();\n\n  callback?.();\n};\n\ndoSomething();\n      "}, {"floatingFixAwait", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  obj1.a?.b?.c?.();\n  obj2.a?.b?.c();\n  obj3.a?.b.c?.();\n  await obj4.a.b.c?.();\n  obj5.a?.().b?.c?.();\n  obj6?.a.b.c?.();\n\n  callback?.();\n};\n\ndoSomething();\n      "}}, {{"floatingFixVoid", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  obj1.a?.b?.c?.();\n  obj2.a?.b?.c();\n  obj3.a?.b.c?.();\n  obj4.a.b.c?.();\n  void obj5.a?.().b?.c?.();\n  obj6?.a.b.c?.();\n\n  callback?.();\n};\n\ndoSomething();\n      "}, {"floatingFixAwait", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  obj1.a?.b?.c?.();\n  obj2.a?.b?.c();\n  obj3.a?.b.c?.();\n  obj4.a.b.c?.();\n  await obj5.a?.().b?.c?.();\n  obj6?.a.b.c?.();\n\n  callback?.();\n};\n\ndoSomething();\n      "}}, {{"floatingFixVoid", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  obj1.a?.b?.c?.();\n  obj2.a?.b?.c();\n  obj3.a?.b.c?.();\n  obj4.a.b.c?.();\n  obj5.a?.().b?.c?.();\n  void obj6?.a.b.c?.();\n\n  callback?.();\n};\n\ndoSomething();\n      "}, {"floatingFixAwait", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  obj1.a?.b?.c?.();\n  obj2.a?.b?.c();\n  obj3.a?.b.c?.();\n  obj4.a.b.c?.();\n  obj5.a?.().b?.c?.();\n  await obj6?.a.b.c?.();\n\n  callback?.();\n};\n\ndoSomething();\n      "}}, {{"floatingFixVoid", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  obj1.a?.b?.c?.();\n  obj2.a?.b?.c();\n  obj3.a?.b.c?.();\n  obj4.a.b.c?.();\n  obj5.a?.().b?.c?.();\n  obj6?.a.b.c?.();\n\n  void callback?.();\n};\n\ndoSomething();\n      "}, {"floatingFixAwait", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  obj1.a?.b?.c?.();\n  obj2.a?.b?.c();\n  obj3.a?.b.c?.();\n  obj4.a.b.c?.();\n  obj5.a?.().b?.c?.();\n  obj6?.a.b.c?.();\n\n  await callback?.();\n};\n\ndoSomething();\n      "}}, {{"floatingFixVoid", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  obj1.a?.b?.c?.();\n  obj2.a?.b?.c();\n  obj3.a?.b.c?.();\n  obj4.a.b.c?.();\n  obj5.a?.().b?.c?.();\n  obj6?.a.b.c?.();\n\n  callback?.();\n};\n\nvoid doSomething();\n      "}, {"floatingFixAwait", "\nconst doSomething = async (\n  obj1: { a?: { b?: { c?: () => Promise<void> } } },\n  obj2: { a?: { b?: { c: () => Promise<void> } } },\n  obj3: { a?: { b: { c?: () => Promise<void> } } },\n  obj4: { a: { b: { c?: () => Promise<void> } } },\n  obj5: { a?: () => { b?: { c?: () => Promise<void> } } },\n  obj6?: { a: { b: { c?: () => Promise<void> } } },\n  callback?: () => Promise<void>,\n): Promise<void> => {\n  obj1.a?.b?.c?.();\n  obj2.a?.b?.c();\n  obj3.a?.b.c?.();\n  obj4.a.b.c?.();\n  obj5.a?.().b?.c?.();\n  obj6?.a.b.c?.();\n\n  callback?.();\n};\n\nawait doSomething();\n      "}}}},
		{"upstream invalid case 2", "\ndeclare const myTag: (strings: TemplateStringsArray) => Promise<void>;\nmyTag`abc`;\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\ndeclare const myTag: (strings: TemplateStringsArray) => Promise<void>;\nvoid myTag`abc`;\n      "}, {"floatingFixAwait", "\ndeclare const myTag: (strings: TemplateStringsArray) => Promise<void>;\nawait myTag`abc`;\n      "}}}},
		{"upstream invalid case 3", "\ndeclare const myTag: (strings: TemplateStringsArray) => Promise<void>;\nmyTag`abc`.then(() => {});\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\ndeclare const myTag: (strings: TemplateStringsArray) => Promise<void>;\nvoid myTag`abc`.then(() => {});\n      "}, {"floatingFixAwait", "\ndeclare const myTag: (strings: TemplateStringsArray) => Promise<void>;\nawait myTag`abc`.then(() => {});\n      "}}}},
		{"upstream invalid case 4", "\ndeclare const myTag: (strings: TemplateStringsArray) => Promise<void>;\nmyTag`abc`.finally(() => {});\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\ndeclare const myTag: (strings: TemplateStringsArray) => Promise<void>;\nvoid myTag`abc`.finally(() => {});\n      "}, {"floatingFixAwait", "\ndeclare const myTag: (strings: TemplateStringsArray) => Promise<void>;\nawait myTag`abc`.finally(() => {});\n      "}}}},
		{"upstream invalid case 5", "\nasync function test() {\n  Promise.resolve('value');\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(true)}, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nasync function test() {\n  void Promise.resolve('value');\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  await Promise.resolve('value');\n}\n      "}}}},
		{"upstream invalid case 6", "\nasync function test() {\n  Promise.reject(new Error('message'));\n  Promise.reject(new Error('message')).then(() => {});\n  Promise.reject(new Error('message')).catch();\n  Promise.reject(new Error('message')).finally();\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nasync function test() {\n  void Promise.reject(new Error('message'));\n  Promise.reject(new Error('message')).then(() => {});\n  Promise.reject(new Error('message')).catch();\n  Promise.reject(new Error('message')).finally();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  await Promise.reject(new Error('message'));\n  Promise.reject(new Error('message')).then(() => {});\n  Promise.reject(new Error('message')).catch();\n  Promise.reject(new Error('message')).finally();\n}\n      "}}, {{"floatingFixVoid", "\nasync function test() {\n  Promise.reject(new Error('message'));\n  void Promise.reject(new Error('message')).then(() => {});\n  Promise.reject(new Error('message')).catch();\n  Promise.reject(new Error('message')).finally();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  Promise.reject(new Error('message'));\n  await Promise.reject(new Error('message')).then(() => {});\n  Promise.reject(new Error('message')).catch();\n  Promise.reject(new Error('message')).finally();\n}\n      "}}, {{"floatingFixVoid", "\nasync function test() {\n  Promise.reject(new Error('message'));\n  Promise.reject(new Error('message')).then(() => {});\n  void Promise.reject(new Error('message')).catch();\n  Promise.reject(new Error('message')).finally();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  Promise.reject(new Error('message'));\n  Promise.reject(new Error('message')).then(() => {});\n  await Promise.reject(new Error('message')).catch();\n  Promise.reject(new Error('message')).finally();\n}\n      "}}, {{"floatingFixVoid", "\nasync function test() {\n  Promise.reject(new Error('message'));\n  Promise.reject(new Error('message')).then(() => {});\n  Promise.reject(new Error('message')).catch();\n  void Promise.reject(new Error('message')).finally();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  Promise.reject(new Error('message'));\n  Promise.reject(new Error('message')).then(() => {});\n  Promise.reject(new Error('message')).catch();\n  await Promise.reject(new Error('message')).finally();\n}\n      "}}}},
		{"upstream invalid case 7", "\nasync function test() {\n  (async () => true)();\n  (async () => true)().then(() => {});\n  (async () => true)().catch();\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nasync function test() {\n  void (async () => true)();\n  (async () => true)().then(() => {});\n  (async () => true)().catch();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  await (async () => true)();\n  (async () => true)().then(() => {});\n  (async () => true)().catch();\n}\n      "}}, {{"floatingFixVoid", "\nasync function test() {\n  (async () => true)();\n  void (async () => true)().then(() => {});\n  (async () => true)().catch();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  (async () => true)();\n  await (async () => true)().then(() => {});\n  (async () => true)().catch();\n}\n      "}}, {{"floatingFixVoid", "\nasync function test() {\n  (async () => true)();\n  (async () => true)().then(() => {});\n  void (async () => true)().catch();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  (async () => true)();\n  (async () => true)().then(() => {});\n  await (async () => true)().catch();\n}\n      "}}}},
		{"upstream invalid case 8", "\nasync function test() {\n  async function returnsPromise() {}\n\n  returnsPromise();\n  returnsPromise().then(() => {});\n  returnsPromise().catch();\n  returnsPromise().finally();\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nasync function test() {\n  async function returnsPromise() {}\n\n  void returnsPromise();\n  returnsPromise().then(() => {});\n  returnsPromise().catch();\n  returnsPromise().finally();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  async function returnsPromise() {}\n\n  await returnsPromise();\n  returnsPromise().then(() => {});\n  returnsPromise().catch();\n  returnsPromise().finally();\n}\n      "}}, {{"floatingFixVoid", "\nasync function test() {\n  async function returnsPromise() {}\n\n  returnsPromise();\n  void returnsPromise().then(() => {});\n  returnsPromise().catch();\n  returnsPromise().finally();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  async function returnsPromise() {}\n\n  returnsPromise();\n  await returnsPromise().then(() => {});\n  returnsPromise().catch();\n  returnsPromise().finally();\n}\n      "}}, {{"floatingFixVoid", "\nasync function test() {\n  async function returnsPromise() {}\n\n  returnsPromise();\n  returnsPromise().then(() => {});\n  void returnsPromise().catch();\n  returnsPromise().finally();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  async function returnsPromise() {}\n\n  returnsPromise();\n  returnsPromise().then(() => {});\n  await returnsPromise().catch();\n  returnsPromise().finally();\n}\n      "}}, {{"floatingFixVoid", "\nasync function test() {\n  async function returnsPromise() {}\n\n  returnsPromise();\n  returnsPromise().then(() => {});\n  returnsPromise().catch();\n  void returnsPromise().finally();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  async function returnsPromise() {}\n\n  returnsPromise();\n  returnsPromise().then(() => {});\n  returnsPromise().catch();\n  await returnsPromise().finally();\n}\n      "}}}},
		{"upstream invalid case 9", "\nasync function test() {\n  Math.random() > 0.5 ? Promise.resolve() : null;\n  Math.random() > 0.5 ? null : Promise.resolve();\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nasync function test() {\n  void (Math.random() > 0.5 ? Promise.resolve() : null);\n  Math.random() > 0.5 ? null : Promise.resolve();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  await (Math.random() > 0.5 ? Promise.resolve() : null);\n  Math.random() > 0.5 ? null : Promise.resolve();\n}\n      "}}, {{"floatingFixVoid", "\nasync function test() {\n  Math.random() > 0.5 ? Promise.resolve() : null;\n  void (Math.random() > 0.5 ? null : Promise.resolve());\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  Math.random() > 0.5 ? Promise.resolve() : null;\n  await (Math.random() > 0.5 ? null : Promise.resolve());\n}\n      "}}}},
		{"upstream invalid case 10", "\nasync function test() {\n  Promise.resolve(), 123;\n  123, Promise.resolve();\n  123, Promise.resolve(), 123;\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nasync function test() {\n  void (Promise.resolve(), 123);\n  123, Promise.resolve();\n  123, Promise.resolve(), 123;\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  await (Promise.resolve(), 123);\n  123, Promise.resolve();\n  123, Promise.resolve(), 123;\n}\n      "}}, {{"floatingFixVoid", "\nasync function test() {\n  Promise.resolve(), 123;\n  void (123, Promise.resolve());\n  123, Promise.resolve(), 123;\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  Promise.resolve(), 123;\n  await (123, Promise.resolve());\n  123, Promise.resolve(), 123;\n}\n      "}}, {{"floatingFixVoid", "\nasync function test() {\n  Promise.resolve(), 123;\n  123, Promise.resolve();\n  void (123, Promise.resolve(), 123);\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  Promise.resolve(), 123;\n  123, Promise.resolve();\n  await (123, Promise.resolve(), 123);\n}\n      "}}}},
		{"upstream invalid case 11", "\nasync function test() {\n  void Promise.resolve();\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, [][]noFloatingPromisesSuggestion{{{"floatingFixAwait", "\nasync function test() {\n  await Promise.resolve();\n}\n      "}}}},
		{"upstream invalid case 12", "\nasync function test() {\n  const promise = new Promise((resolve, reject) => resolve('value'));\n  promise;\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, [][]noFloatingPromisesSuggestion{{{"floatingFixAwait", "\nasync function test() {\n  const promise = new Promise((resolve, reject) => resolve('value'));\n  await promise;\n}\n      "}}}},
		{"upstream invalid case 13", "\nasync function returnsPromise() {\n  return 'value';\n}\nvoid returnsPromise();\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, [][]noFloatingPromisesSuggestion{{{"floatingFixAwait", "\nasync function returnsPromise() {\n  return 'value';\n}\nawait returnsPromise();\n      "}}}},
		{"upstream invalid case 14", "\nasync function returnsPromise() {\n  return 'value';\n}\nvoid /* ... */ returnsPromise();\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, [][]noFloatingPromisesSuggestion{{{"floatingFixAwait", "\nasync function returnsPromise() {\n  return 'value';\n}\nawait /* ... */ returnsPromise();\n      "}}}},
		{"upstream invalid case 15", "\nasync function returnsPromise() {\n  return 'value';\n}\n1, returnsPromise();\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, [][]noFloatingPromisesSuggestion{{{"floatingFixAwait", "\nasync function returnsPromise() {\n  return 'value';\n}\nawait (1, returnsPromise());\n      "}}}},
		{"upstream invalid case 16", "\nasync function returnsPromise() {\n  return 'value';\n}\nbool ? returnsPromise() : null;\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, [][]noFloatingPromisesSuggestion{{{"floatingFixAwait", "\nasync function returnsPromise() {\n  return 'value';\n}\nawait (bool ? returnsPromise() : null);\n      "}}}},
		{"upstream invalid case 17", "\nasync function test() {\n  const obj = { foo: Promise.resolve() };\n  obj.foo;\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nasync function test() {\n  const obj = { foo: Promise.resolve() };\n  void obj.foo;\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  const obj = { foo: Promise.resolve() };\n  await obj.foo;\n}\n      "}}}},
		{"upstream invalid case 18", "\nasync function test() {\n  new Promise(resolve => resolve());\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nasync function test() {\n  void new Promise(resolve => resolve());\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  await new Promise(resolve => resolve());\n}\n      "}}}},
		{"upstream invalid case 19", "\ndeclare const promiseValue: Promise<number>;\n\nasync function test() {\n  promiseValue;\n  promiseValue.then(() => {});\n  promiseValue.catch();\n  promiseValue.finally();\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\ndeclare const promiseValue: Promise<number>;\n\nasync function test() {\n  void promiseValue;\n  promiseValue.then(() => {});\n  promiseValue.catch();\n  promiseValue.finally();\n}\n      "}, {"floatingFixAwait", "\ndeclare const promiseValue: Promise<number>;\n\nasync function test() {\n  await promiseValue;\n  promiseValue.then(() => {});\n  promiseValue.catch();\n  promiseValue.finally();\n}\n      "}}, {{"floatingFixVoid", "\ndeclare const promiseValue: Promise<number>;\n\nasync function test() {\n  promiseValue;\n  void promiseValue.then(() => {});\n  promiseValue.catch();\n  promiseValue.finally();\n}\n      "}, {"floatingFixAwait", "\ndeclare const promiseValue: Promise<number>;\n\nasync function test() {\n  promiseValue;\n  await promiseValue.then(() => {});\n  promiseValue.catch();\n  promiseValue.finally();\n}\n      "}}, {{"floatingFixVoid", "\ndeclare const promiseValue: Promise<number>;\n\nasync function test() {\n  promiseValue;\n  promiseValue.then(() => {});\n  void promiseValue.catch();\n  promiseValue.finally();\n}\n      "}, {"floatingFixAwait", "\ndeclare const promiseValue: Promise<number>;\n\nasync function test() {\n  promiseValue;\n  promiseValue.then(() => {});\n  await promiseValue.catch();\n  promiseValue.finally();\n}\n      "}}, {{"floatingFixVoid", "\ndeclare const promiseValue: Promise<number>;\n\nasync function test() {\n  promiseValue;\n  promiseValue.then(() => {});\n  promiseValue.catch();\n  void promiseValue.finally();\n}\n      "}, {"floatingFixAwait", "\ndeclare const promiseValue: Promise<number>;\n\nasync function test() {\n  promiseValue;\n  promiseValue.then(() => {});\n  promiseValue.catch();\n  await promiseValue.finally();\n}\n      "}}}},
		{"upstream invalid case 20", "\ndeclare const promiseUnion: Promise<number> | number;\n\nasync function test() {\n  promiseUnion;\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\ndeclare const promiseUnion: Promise<number> | number;\n\nasync function test() {\n  void promiseUnion;\n}\n      "}, {"floatingFixAwait", "\ndeclare const promiseUnion: Promise<number> | number;\n\nasync function test() {\n  await promiseUnion;\n}\n      "}}}},
		{"upstream invalid case 21", "\ndeclare const promiseIntersection: Promise<number> & number;\n\nasync function test() {\n  promiseIntersection;\n  promiseIntersection.then(() => {});\n  promiseIntersection.catch();\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\ndeclare const promiseIntersection: Promise<number> & number;\n\nasync function test() {\n  void promiseIntersection;\n  promiseIntersection.then(() => {});\n  promiseIntersection.catch();\n}\n      "}, {"floatingFixAwait", "\ndeclare const promiseIntersection: Promise<number> & number;\n\nasync function test() {\n  await promiseIntersection;\n  promiseIntersection.then(() => {});\n  promiseIntersection.catch();\n}\n      "}}, {{"floatingFixVoid", "\ndeclare const promiseIntersection: Promise<number> & number;\n\nasync function test() {\n  promiseIntersection;\n  void promiseIntersection.then(() => {});\n  promiseIntersection.catch();\n}\n      "}, {"floatingFixAwait", "\ndeclare const promiseIntersection: Promise<number> & number;\n\nasync function test() {\n  promiseIntersection;\n  await promiseIntersection.then(() => {});\n  promiseIntersection.catch();\n}\n      "}}, {{"floatingFixVoid", "\ndeclare const promiseIntersection: Promise<number> & number;\n\nasync function test() {\n  promiseIntersection;\n  promiseIntersection.then(() => {});\n  void promiseIntersection.catch();\n}\n      "}, {"floatingFixAwait", "\ndeclare const promiseIntersection: Promise<number> & number;\n\nasync function test() {\n  promiseIntersection;\n  promiseIntersection.then(() => {});\n  await promiseIntersection.catch();\n}\n      "}}}},
		{"upstream invalid case 22", "\nasync function test() {\n  class CanThen extends Promise<number> {}\n  const canThen: CanThen = Foo.resolve(2);\n\n  canThen;\n  canThen.then(() => {});\n  canThen.catch();\n  canThen.finally();\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nasync function test() {\n  class CanThen extends Promise<number> {}\n  const canThen: CanThen = Foo.resolve(2);\n\n  void canThen;\n  canThen.then(() => {});\n  canThen.catch();\n  canThen.finally();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  class CanThen extends Promise<number> {}\n  const canThen: CanThen = Foo.resolve(2);\n\n  await canThen;\n  canThen.then(() => {});\n  canThen.catch();\n  canThen.finally();\n}\n      "}}, {{"floatingFixVoid", "\nasync function test() {\n  class CanThen extends Promise<number> {}\n  const canThen: CanThen = Foo.resolve(2);\n\n  canThen;\n  void canThen.then(() => {});\n  canThen.catch();\n  canThen.finally();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  class CanThen extends Promise<number> {}\n  const canThen: CanThen = Foo.resolve(2);\n\n  canThen;\n  await canThen.then(() => {});\n  canThen.catch();\n  canThen.finally();\n}\n      "}}, {{"floatingFixVoid", "\nasync function test() {\n  class CanThen extends Promise<number> {}\n  const canThen: CanThen = Foo.resolve(2);\n\n  canThen;\n  canThen.then(() => {});\n  void canThen.catch();\n  canThen.finally();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  class CanThen extends Promise<number> {}\n  const canThen: CanThen = Foo.resolve(2);\n\n  canThen;\n  canThen.then(() => {});\n  await canThen.catch();\n  canThen.finally();\n}\n      "}}, {{"floatingFixVoid", "\nasync function test() {\n  class CanThen extends Promise<number> {}\n  const canThen: CanThen = Foo.resolve(2);\n\n  canThen;\n  canThen.then(() => {});\n  canThen.catch();\n  void canThen.finally();\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  class CanThen extends Promise<number> {}\n  const canThen: CanThen = Foo.resolve(2);\n\n  canThen;\n  canThen.then(() => {});\n  canThen.catch();\n  await canThen.finally();\n}\n      "}}}},
		{"upstream invalid case 23", "\nasync function test() {\n  class CatchableThenable {\n    then(callback: () => void, callback: () => void): CatchableThenable {\n      return new CatchableThenable();\n    }\n  }\n  const thenable = new CatchableThenable();\n\n  thenable;\n  thenable.then(() => {});\n}\n      ", NoFloatingPromisesOptions{CheckThenables: type_checking.Ref(true)}, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nasync function test() {\n  class CatchableThenable {\n    then(callback: () => void, callback: () => void): CatchableThenable {\n      return new CatchableThenable();\n    }\n  }\n  const thenable = new CatchableThenable();\n\n  void thenable;\n  thenable.then(() => {});\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  class CatchableThenable {\n    then(callback: () => void, callback: () => void): CatchableThenable {\n      return new CatchableThenable();\n    }\n  }\n  const thenable = new CatchableThenable();\n\n  await thenable;\n  thenable.then(() => {});\n}\n      "}}, {{"floatingFixVoid", "\nasync function test() {\n  class CatchableThenable {\n    then(callback: () => void, callback: () => void): CatchableThenable {\n      return new CatchableThenable();\n    }\n  }\n  const thenable = new CatchableThenable();\n\n  thenable;\n  void thenable.then(() => {});\n}\n      "}, {"floatingFixAwait", "\nasync function test() {\n  class CatchableThenable {\n    then(callback: () => void, callback: () => void): CatchableThenable {\n      return new CatchableThenable();\n    }\n  }\n  const thenable = new CatchableThenable();\n\n  thenable;\n  await thenable.then(() => {});\n}\n      "}}}},
		{"upstream invalid case 24", "\n// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/promise-polyfill/index.d.ts\n// Type definitions for promise-polyfill 6.0\n// Project: https://github.com/taylorhakes/promise-polyfill\n// Definitions by: Steve Jenkins <https://github.com/skysteve>\n//                 Daniel Cassidy <https://github.com/djcsdy>\n// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped\n\ninterface PromisePolyfillConstructor extends PromiseConstructor {\n  _immediateFn?: (handler: (() => void) | string) => void;\n}\n\ndeclare const PromisePolyfill: PromisePolyfillConstructor;\n\nasync function test() {\n  const promise = new PromisePolyfill(() => {});\n\n  promise;\n  promise.then(() => {});\n  promise.catch();\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\n// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/promise-polyfill/index.d.ts\n// Type definitions for promise-polyfill 6.0\n// Project: https://github.com/taylorhakes/promise-polyfill\n// Definitions by: Steve Jenkins <https://github.com/skysteve>\n//                 Daniel Cassidy <https://github.com/djcsdy>\n// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped\n\ninterface PromisePolyfillConstructor extends PromiseConstructor {\n  _immediateFn?: (handler: (() => void) | string) => void;\n}\n\ndeclare const PromisePolyfill: PromisePolyfillConstructor;\n\nasync function test() {\n  const promise = new PromisePolyfill(() => {});\n\n  void promise;\n  promise.then(() => {});\n  promise.catch();\n}\n      "}, {"floatingFixAwait", "\n// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/promise-polyfill/index.d.ts\n// Type definitions for promise-polyfill 6.0\n// Project: https://github.com/taylorhakes/promise-polyfill\n// Definitions by: Steve Jenkins <https://github.com/skysteve>\n//                 Daniel Cassidy <https://github.com/djcsdy>\n// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped\n\ninterface PromisePolyfillConstructor extends PromiseConstructor {\n  _immediateFn?: (handler: (() => void) | string) => void;\n}\n\ndeclare const PromisePolyfill: PromisePolyfillConstructor;\n\nasync function test() {\n  const promise = new PromisePolyfill(() => {});\n\n  await promise;\n  promise.then(() => {});\n  promise.catch();\n}\n      "}}, {{"floatingFixVoid", "\n// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/promise-polyfill/index.d.ts\n// Type definitions for promise-polyfill 6.0\n// Project: https://github.com/taylorhakes/promise-polyfill\n// Definitions by: Steve Jenkins <https://github.com/skysteve>\n//                 Daniel Cassidy <https://github.com/djcsdy>\n// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped\n\ninterface PromisePolyfillConstructor extends PromiseConstructor {\n  _immediateFn?: (handler: (() => void) | string) => void;\n}\n\ndeclare const PromisePolyfill: PromisePolyfillConstructor;\n\nasync function test() {\n  const promise = new PromisePolyfill(() => {});\n\n  promise;\n  void promise.then(() => {});\n  promise.catch();\n}\n      "}, {"floatingFixAwait", "\n// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/promise-polyfill/index.d.ts\n// Type definitions for promise-polyfill 6.0\n// Project: https://github.com/taylorhakes/promise-polyfill\n// Definitions by: Steve Jenkins <https://github.com/skysteve>\n//                 Daniel Cassidy <https://github.com/djcsdy>\n// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped\n\ninterface PromisePolyfillConstructor extends PromiseConstructor {\n  _immediateFn?: (handler: (() => void) | string) => void;\n}\n\ndeclare const PromisePolyfill: PromisePolyfillConstructor;\n\nasync function test() {\n  const promise = new PromisePolyfill(() => {});\n\n  promise;\n  await promise.then(() => {});\n  promise.catch();\n}\n      "}}, {{"floatingFixVoid", "\n// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/promise-polyfill/index.d.ts\n// Type definitions for promise-polyfill 6.0\n// Project: https://github.com/taylorhakes/promise-polyfill\n// Definitions by: Steve Jenkins <https://github.com/skysteve>\n//                 Daniel Cassidy <https://github.com/djcsdy>\n// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped\n\ninterface PromisePolyfillConstructor extends PromiseConstructor {\n  _immediateFn?: (handler: (() => void) | string) => void;\n}\n\ndeclare const PromisePolyfill: PromisePolyfillConstructor;\n\nasync function test() {\n  const promise = new PromisePolyfill(() => {});\n\n  promise;\n  promise.then(() => {});\n  void promise.catch();\n}\n      "}, {"floatingFixAwait", "\n// https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/promise-polyfill/index.d.ts\n// Type definitions for promise-polyfill 6.0\n// Project: https://github.com/taylorhakes/promise-polyfill\n// Definitions by: Steve Jenkins <https://github.com/skysteve>\n//                 Daniel Cassidy <https://github.com/djcsdy>\n// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped\n\ninterface PromisePolyfillConstructor extends PromiseConstructor {\n  _immediateFn?: (handler: (() => void) | string) => void;\n}\n\ndeclare const PromisePolyfill: PromisePolyfillConstructor;\n\nasync function test() {\n  const promise = new PromisePolyfill(() => {});\n\n  promise;\n  promise.then(() => {});\n  await promise.catch();\n}\n      "}}}},
		{"upstream invalid case 25", "\n        (async () => {\n          await something();\n        })();\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\n        void (async () => {\n          await something();\n        })();\n      "}, {"floatingFixAwait", "\n        await (async () => {\n          await something();\n        })();\n      "}}}},
		{"upstream invalid case 26", "\n        (async () => {\n          something();\n        })();\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\n        void (async () => {\n          something();\n        })();\n      "}, {"floatingFixAwait", "\n        await (async () => {\n          something();\n        })();\n      "}}}},
		{"upstream invalid case 27", "(async function foo() {})();", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "void (async function foo() {})();"}, {"floatingFixAwait", "await (async function foo() {})();"}}}},
		{"upstream invalid case 28", "\n        function foo() {\n          (async function bar() {})();\n        }\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\n        function foo() {\n          void (async function bar() {})();\n        }\n      "}, {"floatingFixAwait", "\n        function foo() {\n          await (async function bar() {})();\n        }\n      "}}}},
		{"upstream invalid case 29", "\n        const foo = () =>\n          new Promise(res => {\n            (async function () {\n              await res(1);\n            })();\n          });\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\n        const foo = () =>\n          new Promise(res => {\n            void (async function () {\n              await res(1);\n            })();\n          });\n      "}, {"floatingFixAwait", "\n        const foo = () =>\n          new Promise(res => {\n            await (async function () {\n              await res(1);\n            })();\n          });\n      "}}}},
		{"upstream invalid case 30", "\n        (async function () {\n          await res(1);\n        })();\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\n        void (async function () {\n          await res(1);\n        })();\n      "}, {"floatingFixAwait", "\n        await (async function () {\n          await res(1);\n        })();\n      "}}}},
		{"upstream invalid case 31", "\n        (async function () {\n          Promise.resolve();\n        })();\n      ", NoFloatingPromisesOptions{IgnoreIIFE: type_checking.Ref(true)}, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\n        (async function () {\n          void Promise.resolve();\n        })();\n      "}, {"floatingFixAwait", "\n        (async function () {\n          await Promise.resolve();\n        })();\n      "}}}},
		{"upstream invalid case 32", "\ndeclare const promiseIntersection: Promise<number> & number;\n(async function () {\n  promiseIntersection;\n  promiseIntersection.then(() => {});\n  promiseIntersection.catch();\n  promiseIntersection.finally();\n})();\n      ", NoFloatingPromisesOptions{IgnoreIIFE: type_checking.Ref(true)}, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\ndeclare const promiseIntersection: Promise<number> & number;\n(async function () {\n  void promiseIntersection;\n  promiseIntersection.then(() => {});\n  promiseIntersection.catch();\n  promiseIntersection.finally();\n})();\n      "}, {"floatingFixAwait", "\ndeclare const promiseIntersection: Promise<number> & number;\n(async function () {\n  await promiseIntersection;\n  promiseIntersection.then(() => {});\n  promiseIntersection.catch();\n  promiseIntersection.finally();\n})();\n      "}}, {{"floatingFixVoid", "\ndeclare const promiseIntersection: Promise<number> & number;\n(async function () {\n  promiseIntersection;\n  void promiseIntersection.then(() => {});\n  promiseIntersection.catch();\n  promiseIntersection.finally();\n})();\n      "}, {"floatingFixAwait", "\ndeclare const promiseIntersection: Promise<number> & number;\n(async function () {\n  promiseIntersection;\n  await promiseIntersection.then(() => {});\n  promiseIntersection.catch();\n  promiseIntersection.finally();\n})();\n      "}}, {{"floatingFixVoid", "\ndeclare const promiseIntersection: Promise<number> & number;\n(async function () {\n  promiseIntersection;\n  promiseIntersection.then(() => {});\n  void promiseIntersection.catch();\n  promiseIntersection.finally();\n})();\n      "}, {"floatingFixAwait", "\ndeclare const promiseIntersection: Promise<number> & number;\n(async function () {\n  promiseIntersection;\n  promiseIntersection.then(() => {});\n  await promiseIntersection.catch();\n  promiseIntersection.finally();\n})();\n      "}}, {{"floatingFixVoid", "\ndeclare const promiseIntersection: Promise<number> & number;\n(async function () {\n  promiseIntersection;\n  promiseIntersection.then(() => {});\n  promiseIntersection.catch();\n  void promiseIntersection.finally();\n})();\n      "}, {"floatingFixAwait", "\ndeclare const promiseIntersection: Promise<number> & number;\n(async function () {\n  promiseIntersection;\n  promiseIntersection.then(() => {});\n  promiseIntersection.catch();\n  await promiseIntersection.finally();\n})();\n      "}}}},
		{"upstream invalid case 33", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = true;\n\n  void condition || myPromise();\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = true;\n\n  void (void condition || myPromise());\n}\n      "}, {"floatingFixAwait", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = true;\n\n  await (void condition || myPromise());\n}\n      "}}}},
		{"upstream invalid case 34", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = true;\n\n  (await condition) && myPromise();\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, [][]noFloatingPromisesSuggestion{{{"floatingFixAwait", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = true;\n\n  await ((await condition) && myPromise());\n}\n      "}}}},
		{"upstream invalid case 35", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = true;\n\n  condition && myPromise();\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = true;\n\n  void (condition && myPromise());\n}\n      "}, {"floatingFixAwait", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = true;\n\n  await (condition && myPromise());\n}\n      "}}}},
		{"upstream invalid case 36", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = false;\n\n  condition || myPromise();\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = false;\n\n  void (condition || myPromise());\n}\n      "}, {"floatingFixAwait", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = false;\n\n  await (condition || myPromise());\n}\n      "}}}},
		{"upstream invalid case 37", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = null;\n\n  condition ?? myPromise();\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = null;\n\n  void (condition ?? myPromise());\n}\n      "}, {"floatingFixAwait", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = null;\n\n  await (condition ?? myPromise());\n}\n      "}}}},
		{"upstream invalid case 38", "\nasync function foo() {\n  const myPromise = Promise.resolve(true);\n  let condition = true;\n  condition && myPromise;\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, [][]noFloatingPromisesSuggestion{{{"floatingFixAwait", "\nasync function foo() {\n  const myPromise = Promise.resolve(true);\n  let condition = true;\n  await (condition && myPromise);\n}\n      "}}}},
		{"upstream invalid case 39", "\nasync function foo() {\n  const myPromise = Promise.resolve(true);\n  let condition = false;\n  condition || myPromise;\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, [][]noFloatingPromisesSuggestion{{{"floatingFixAwait", "\nasync function foo() {\n  const myPromise = Promise.resolve(true);\n  let condition = false;\n  await (condition || myPromise);\n}\n      "}}}},
		{"upstream invalid case 40", "\nasync function foo() {\n  const myPromise = Promise.resolve(true);\n  let condition = null;\n  condition ?? myPromise;\n}\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, [][]noFloatingPromisesSuggestion{{{"floatingFixAwait", "\nasync function foo() {\n  const myPromise = Promise.resolve(true);\n  let condition = null;\n  await (condition ?? myPromise);\n}\n      "}}}},
		{"upstream invalid case 41", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = false;\n\n  condition || condition || myPromise();\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = false;\n\n  void (condition || condition || myPromise());\n}\n      "}, {"floatingFixAwait", "\nasync function foo() {\n  const myPromise = async () => void 0;\n  const condition = false;\n\n  await (condition || condition || myPromise());\n}\n      "}}}},
		{"upstream invalid case 42", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nvoid Promise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}, {"floatingFixAwait", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nawait Promise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}}, {{"floatingFixVoid", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nvoid Promise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}, {"floatingFixAwait", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nawait Promise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}}, {{"floatingFixVoid", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nvoid Promise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}, {"floatingFixAwait", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nawait Promise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}}, {{"floatingFixVoid", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nvoid Promise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}, {"floatingFixAwait", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nawait Promise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}}, {{"floatingFixVoid", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nvoid Promise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}, {"floatingFixAwait", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nawait Promise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}}, {{"floatingFixVoid", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nvoid Promise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}, {"floatingFixAwait", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nawait Promise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}}, {{"floatingFixVoid", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nvoid Promise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}, {"floatingFixAwait", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nawait Promise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}}, {{"floatingFixVoid", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nvoid Promise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}, {"floatingFixAwait", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nawait Promise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}}}},
		{"upstream invalid case 43", "\nPromise.reject() || 3;\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nvoid (Promise.reject() || 3);\n      "}, {"floatingFixAwait", "\nawait (Promise.reject() || 3);\n      "}}}},
		{"upstream invalid case 44", "\nvoid Promise.resolve().then(() => {}, undefined);\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, [][]noFloatingPromisesSuggestion{{{"floatingFixAwait", "\nawait Promise.resolve().then(() => {}, undefined);\n      "}}}},
		{"upstream invalid case 45", "\ndeclare const maybeCallable: string | (() => void);\nPromise.resolve().then(() => {}, maybeCallable);\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, [][]noFloatingPromisesSuggestion{{{"floatingFixAwait", "\ndeclare const maybeCallable: string | (() => void);\nawait Promise.resolve().then(() => {}, maybeCallable);\n      "}}}},
		{"upstream invalid case 46", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, [][]noFloatingPromisesSuggestion{{{"floatingFixAwait", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nawait Promise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}}, {{"floatingFixAwait", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nawait Promise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}}, {{"floatingFixAwait", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nawait Promise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}}, {{"floatingFixAwait", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nawait Promise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}}, {{"floatingFixAwait", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nawait Promise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}}, {{"floatingFixAwait", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nawait Promise.resolve().catch(null);\nPromise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}}, {{"floatingFixAwait", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nawait Promise.resolve().catch(3);\nPromise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}}, {{"floatingFixAwait", "\ndeclare const maybeCallable: string | (() => void);\ndeclare const definitelyCallable: () => void;\nPromise.resolve().then(() => {}, undefined);\nPromise.resolve().then(() => {}, null);\nPromise.resolve().then(() => {}, 3);\nPromise.resolve().then(() => {}, maybeCallable);\nPromise.resolve().then(() => {}, definitelyCallable);\n\nPromise.resolve().catch(undefined);\nPromise.resolve().catch(null);\nPromise.resolve().catch(3);\nawait Promise.resolve().catch(maybeCallable);\nPromise.resolve().catch(definitelyCallable);\n      "}}}},
		{"upstream invalid case 47", "\nPromise.reject() || 3;\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, [][]noFloatingPromisesSuggestion{{{"floatingFixAwait", "\nawait (Promise.reject() || 3);\n      "}}}},
		{"upstream invalid case 48", "\nPromise.reject().finally(() => {});\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nvoid Promise.reject().finally(() => {});\n      "}, {"floatingFixAwait", "\nawait Promise.reject().finally(() => {});\n      "}}}},
		{"upstream invalid case 49", "\nPromise.reject()\n  .finally(() => {})\n  .finally(() => {});\n      ", NoFloatingPromisesOptions{IgnoreVoid: type_checking.Ref(false)}, [][]noFloatingPromisesSuggestion{{{"floatingFixAwait", "\nawait Promise.reject()\n  .finally(() => {})\n  .finally(() => {});\n      "}}}},
		{"upstream invalid case 50", "\nPromise.reject()\n  .finally(() => {})\n  .finally(() => {})\n  .finally(() => {});\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nvoid Promise.reject()\n  .finally(() => {})\n  .finally(() => {})\n  .finally(() => {});\n      "}, {"floatingFixAwait", "\nawait Promise.reject()\n  .finally(() => {})\n  .finally(() => {})\n  .finally(() => {});\n      "}}}},
		{"upstream invalid case 51", "\nPromise.reject()\n  .then(() => {})\n  .finally(() => {});\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nvoid Promise.reject()\n  .then(() => {})\n  .finally(() => {});\n      "}, {"floatingFixAwait", "\nawait Promise.reject()\n  .then(() => {})\n  .finally(() => {});\n      "}}}},
		{"upstream invalid case 52", "\ndeclare const returnsPromise: () => Promise<void> | null;\nreturnsPromise()?.finally(() => {});\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\ndeclare const returnsPromise: () => Promise<void> | null;\nvoid returnsPromise()?.finally(() => {});\n      "}, {"floatingFixAwait", "\ndeclare const returnsPromise: () => Promise<void> | null;\nawait returnsPromise()?.finally(() => {});\n      "}}}},
		{"upstream invalid case 53", "\nconst promiseIntersection: Promise<number> & number;\npromiseIntersection.finally(() => {});\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nconst promiseIntersection: Promise<number> & number;\nvoid promiseIntersection.finally(() => {});\n      "}, {"floatingFixAwait", "\nconst promiseIntersection: Promise<number> & number;\nawait promiseIntersection.finally(() => {});\n      "}}}},
		{"upstream invalid case 54", "\nPromise.resolve().finally(() => {}), 123;\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nvoid (Promise.resolve().finally(() => {}), 123);\n      "}, {"floatingFixAwait", "\nawait (Promise.resolve().finally(() => {}), 123);\n      "}}}},
		{"upstream invalid case 55", "\n(async () => true)().finally();\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nvoid (async () => true)().finally();\n      "}, {"floatingFixAwait", "\nawait (async () => true)().finally();\n      "}}}},
		{"upstream invalid case 56", "\nPromise.reject(new Error('message')).finally(() => {});\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nvoid Promise.reject(new Error('message')).finally(() => {});\n      "}, {"floatingFixAwait", "\nawait Promise.reject(new Error('message')).finally(() => {});\n      "}}}},
		{"upstream invalid case 57", "\nfunction _<T, S extends Array<T | Promise<T>>>(\n  maybePromiseArray: S | undefined,\n): void {\n  maybePromiseArray?.[0];\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nfunction _<T, S extends Array<T | Promise<T>>>(\n  maybePromiseArray: S | undefined,\n): void {\n  void maybePromiseArray?.[0];\n}\n      "}, {"floatingFixAwait", "\nfunction _<T, S extends Array<T | Promise<T>>>(\n  maybePromiseArray: S | undefined,\n): void {\n  await maybePromiseArray?.[0];\n}\n      "}}}},
		{"upstream invalid case 74", "\ninterface UnsafeThenable<T> {\n  then<TResult1 = T, TResult2 = never>(\n    onfulfilled?:\n      | ((value: T) => TResult1 | UnsafeThenable<TResult1>)\n      | undefined\n      | null,\n    onrejected?:\n      | ((reason: any) => TResult2 | UnsafeThenable<TResult2>)\n      | undefined\n      | null,\n  ): UnsafeThenable<TResult1 | TResult2>;\n}\nlet promise: UnsafeThenable<number> = Promise.resolve(5);\npromise;\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafeThenable"}}}, CheckThenables: type_checking.Ref(true)}, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\ninterface UnsafeThenable<T> {\n  then<TResult1 = T, TResult2 = never>(\n    onfulfilled?:\n      | ((value: T) => TResult1 | UnsafeThenable<TResult1>)\n      | undefined\n      | null,\n    onrejected?:\n      | ((reason: any) => TResult2 | UnsafeThenable<TResult2>)\n      | undefined\n      | null,\n  ): UnsafeThenable<TResult1 | TResult2>;\n}\nlet promise: UnsafeThenable<number> = Promise.resolve(5);\nvoid promise;\n      "}, {"floatingFixAwait", "\ninterface UnsafeThenable<T> {\n  then<TResult1 = T, TResult2 = never>(\n    onfulfilled?:\n      | ((value: T) => TResult1 | UnsafeThenable<TResult1>)\n      | undefined\n      | null,\n    onrejected?:\n      | ((reason: any) => TResult2 | UnsafeThenable<TResult2>)\n      | undefined\n      | null,\n  ): UnsafeThenable<TResult1 | TResult2>;\n}\nlet promise: UnsafeThenable<number> = Promise.resolve(5);\nawait promise;\n      "}}}},
		{"upstream invalid case 75", "\nclass SafePromise<T> extends Promise<T> {}\nlet promise: SafePromise<number> = Promise.resolve(5);\npromise.catch();\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}}}, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nclass SafePromise<T> extends Promise<T> {}\nlet promise: SafePromise<number> = Promise.resolve(5);\nvoid promise.catch();\n      "}, {"floatingFixAwait", "\nclass SafePromise<T> extends Promise<T> {}\nlet promise: SafePromise<number> = Promise.resolve(5);\nawait promise.catch();\n      "}}}},
		{"upstream invalid case 76", "\nclass UnsafePromise<T> extends Promise<T> {}\nlet promise: () => UnsafePromise<number> = async () => 5;\npromise().finally();\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}}}, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nclass UnsafePromise<T> extends Promise<T> {}\nlet promise: () => UnsafePromise<number> = async () => 5;\nvoid promise().finally();\n      "}, {"floatingFixAwait", "\nclass UnsafePromise<T> extends Promise<T> {}\nlet promise: () => UnsafePromise<number> = async () => 5;\nawait promise().finally();\n      "}}}},
		{"upstream invalid case 77", "\ntype UnsafePromise = Promise<number> & { hey?: string };\nlet promise: UnsafePromise = Promise.resolve(5);\n0 ? promise.catch() : 2;\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}}}, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\ntype UnsafePromise = Promise<number> & { hey?: string };\nlet promise: UnsafePromise = Promise.resolve(5);\nvoid (0 ? promise.catch() : 2);\n      "}, {"floatingFixAwait", "\ntype UnsafePromise = Promise<number> & { hey?: string };\nlet promise: UnsafePromise = Promise.resolve(5);\nawait (0 ? promise.catch() : 2);\n      "}}}},
		{"upstream invalid case 78", "\ntype UnsafePromise = Promise<number> & { hey?: string };\nlet promise: () => UnsafePromise = async () => 5;\nnull ?? promise().catch();\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"SafePromise"}}}}, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\ntype UnsafePromise = Promise<number> & { hey?: string };\nlet promise: () => UnsafePromise = async () => 5;\nvoid (null ?? promise().catch());\n      "}, {"floatingFixAwait", "\ntype UnsafePromise = Promise<number> & { hey?: string };\nlet promise: () => UnsafePromise = async () => 5;\nawait (null ?? promise().catch());\n      "}}}},
		{"upstream invalid case 82", "\ntype SafePromise = Promise<number> & { __linterBrands?: string };\ndeclare const myTag: (strings: TemplateStringsArray) => SafePromise;\nmyTag`abc`;\n      ", NoFloatingPromisesOptions{AllowForKnownSafePromises: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"Foo"}}}}, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\ntype SafePromise = Promise<number> & { __linterBrands?: string };\ndeclare const myTag: (strings: TemplateStringsArray) => SafePromise;\nvoid myTag`abc`;\n      "}, {"floatingFixAwait", "\ntype SafePromise = Promise<number> & { __linterBrands?: string };\ndeclare const myTag: (strings: TemplateStringsArray) => SafePromise;\nawait myTag`abc`;\n      "}}}},
		{"upstream invalid case 83", "\n        declare function unsafe(...args: unknown[]): Promise<void>;\n\n        unsafe('...', () => {});\n      ", NoFloatingPromisesOptions{AllowForKnownSafeCalls: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"it"}, Path: "tests/fixtures/file.ts"}}}, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\n        declare function unsafe(...args: unknown[]): Promise<void>;\n\n        void unsafe('...', () => {});\n      "}, {"floatingFixAwait", "\n        declare function unsafe(...args: unknown[]): Promise<void>;\n\n        await unsafe('...', () => {});\n      "}}}},
		{"upstream invalid case 84", "\n        declare function it(...args: unknown[]): Promise<void>;\n\n        it('...', () => {}).then(() => {});\n      ", NoFloatingPromisesOptions{AllowForKnownSafeCalls: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"it"}, Path: "tests/fixtures/file.ts"}}}, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\n        declare function it(...args: unknown[]): Promise<void>;\n\n        void it('...', () => {}).then(() => {});\n      "}, {"floatingFixAwait", "\n        declare function it(...args: unknown[]): Promise<void>;\n\n        await it('...', () => {}).then(() => {});\n      "}}}},
		{"upstream invalid case 85", "\n        declare function it(...args: unknown[]): Promise<void>;\n\n        it('...', () => {}).finally(() => {});\n      ", NoFloatingPromisesOptions{AllowForKnownSafeCalls: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"it"}, Path: "tests/fixtures/file.ts"}}}, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\n        declare function it(...args: unknown[]): Promise<void>;\n\n        void it('...', () => {}).finally(() => {});\n      "}, {"floatingFixAwait", "\n        declare function it(...args: unknown[]): Promise<void>;\n\n        await it('...', () => {}).finally(() => {});\n      "}}}},
		{"upstream invalid case 86", "\ndeclare const createPromise: () => PromiseLike<number>;\ncreatePromise();\n      ", NoFloatingPromisesOptions{CheckThenables: type_checking.Ref(true)}, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\ndeclare const createPromise: () => PromiseLike<number>;\nvoid createPromise();\n      "}, {"floatingFixAwait", "\ndeclare const createPromise: () => PromiseLike<number>;\nawait createPromise();\n      "}}}},
		{"upstream invalid case 87", "\ninterface MyThenable {\n  then(onFulfilled: () => void, onRejected: () => void): MyThenable;\n}\n\ndeclare function createMyThenable(): MyThenable;\n\ncreateMyThenable();\n      ", NoFloatingPromisesOptions{CheckThenables: type_checking.Ref(true)}, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\ninterface MyThenable {\n  then(onFulfilled: () => void, onRejected: () => void): MyThenable;\n}\n\ndeclare function createMyThenable(): MyThenable;\n\nvoid createMyThenable();\n      "}, {"floatingFixAwait", "\ninterface MyThenable {\n  then(onFulfilled: () => void, onRejected: () => void): MyThenable;\n}\n\ndeclare function createMyThenable(): MyThenable;\n\nawait createMyThenable();\n      "}}}},
		{"upstream invalid case 88", "\ndeclare const createPromise: () => Promise<number>;\ncreatePromise();\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\ndeclare const createPromise: () => Promise<number>;\nvoid createPromise();\n      "}, {"floatingFixAwait", "\ndeclare const createPromise: () => Promise<number>;\nawait createPromise();\n      "}}}},
		{"upstream invalid case 89", "\nclass MyPromise<T> extends Promise<T> {}\ndeclare const createMyPromise: () => MyPromise<number>;\ncreateMyPromise();\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nclass MyPromise<T> extends Promise<T> {}\ndeclare const createMyPromise: () => MyPromise<number>;\nvoid createMyPromise();\n      "}, {"floatingFixAwait", "\nclass MyPromise<T> extends Promise<T> {}\ndeclare const createMyPromise: () => MyPromise<number>;\nawait createMyPromise();\n      "}}}},
		{"upstream invalid case 90", "\nclass MyPromise<T> extends Promise<T> {\n  additional: string;\n}\ndeclare const createMyPromise: () => MyPromise<number>;\ncreateMyPromise();\n      ", NoFloatingPromisesOptions{CheckThenables: type_checking.Ref(true)}, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nclass MyPromise<T> extends Promise<T> {\n  additional: string;\n}\ndeclare const createMyPromise: () => MyPromise<number>;\nvoid createMyPromise();\n      "}, {"floatingFixAwait", "\nclass MyPromise<T> extends Promise<T> {\n  additional: string;\n}\ndeclare const createMyPromise: () => MyPromise<number>;\nawait createMyPromise();\n      "}}}},
		{"upstream invalid case 91", "\ndeclare const x: any;\nfunction* generator(): Generator<number, void, Promise<number>> {\n  yield x;\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\ndeclare const x: any;\nfunction* generator(): Generator<number, void, Promise<number>> {\n  void (yield x);\n}\n      "}, {"floatingFixAwait", "\ndeclare const x: any;\nfunction* generator(): Generator<number, void, Promise<number>> {\n  await (yield x);\n}\n      "}}}},
		{"upstream invalid case 92", "\ndeclare const x: Generator<number, Promise<number>, void>;\nfunction* generator(): Generator<number, void, void> {\n  yield* x;\n}\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\ndeclare const x: Generator<number, Promise<number>, void>;\nfunction* generator(): Generator<number, void, void> {\n  void (yield* x);\n}\n      "}, {"floatingFixAwait", "\ndeclare const x: Generator<number, Promise<number>, void>;\nfunction* generator(): Generator<number, void, void> {\n  await (yield* x);\n}\n      "}}}},
		{"upstream invalid case 93", "\nconst value = {};\nvalue as Promise<number>;\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nconst value = {};\nvoid (value as Promise<number>);\n      "}, {"floatingFixAwait", "\nconst value = {};\nawait (value as Promise<number>);\n      "}}}},
		{"upstream invalid case 94", "\n({}) as Promise<number> & number;\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nvoid (({}) as Promise<number> & number);\n      "}, {"floatingFixAwait", "\nawait (({}) as Promise<number> & number);\n      "}}}},
		{"upstream invalid case 95", "\n({}) as Promise<number> & { yolo?: string };\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nvoid (({}) as Promise<number> & { yolo?: string });\n      "}, {"floatingFixAwait", "\nawait (({}) as Promise<number> & { yolo?: string });\n      "}}}},
		{"upstream invalid case 96", "\n<Promise<number>>{};\n      ", nil, [][]noFloatingPromisesSuggestion{{{"floatingFixVoid", "\nvoid (<Promise<number>>{});\n      "}, {"floatingFixAwait", "\nawait (<Promise<number>>{});\n      "}}}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoFloatingPromises(t, testCase.sourceText, testCase.options)
			if len(result.Diagnostics) != len(testCase.suggestions) {
				t.Fatalf("want %d findings, got %d", len(testCase.suggestions), len(result.Diagnostics))
			}
			// The harness trims the fixture before building the program, so every offset the rule
			// computed is against the trimmed text rather than against the literal above. The expected
			// output is trimmed exactly the same way rather than with a two-sided trim, because a
			// leading space this repair leaves behind is meaningful and a symmetric trim eats it.
			source := strings.TrimSpace(testCase.sourceText)
			for diagnosticIndex, wantSuggestions := range testCase.suggestions {
				gotSuggestions := result.Diagnostics[diagnosticIndex].Suggestions
				if len(gotSuggestions) != len(wantSuggestions) {
					t.Fatalf("finding %d offers %d suggestions, want %d", diagnosticIndex, len(gotSuggestions), len(wantSuggestions))
				}
				for suggestionIndex, want := range wantSuggestions {
					got := gotSuggestions[suggestionIndex]
					if got.Message.Id != want.messageId {
						t.Errorf("finding %d suggestion %d has id %q, want %q", diagnosticIndex, suggestionIndex, got.Message.Id, want.messageId)
					}
					applied := applyNoFloatingPromisesSuggestion(t, source, got)
					if applied != strings.TrimSpace(want.output) {
						t.Errorf("finding %d suggestion %d rewrote the source to\n%q\nwant\n%q", diagnosticIndex, suggestionIndex, applied, strings.TrimSpace(want.output))
					}
				}
			}
		})
	}
}

// TestNoFloatingPromisesMeasuredParentheses pins the paren axis, which the imported corpus writes
// nowhere and which falls THREE different ways in this one rule.
//
// The listener runs `ast.SkipParentheses` over the statement's expression, so parens around the
// WHOLE expression are transparent. They are not skipped again once the descent has started, so a
// parenthesized CALLEE is opaque: `ast.IsAccessExpression(callee)` declines and the `.catch` /
// `.then` / `.finally` handling never runs. Measured, holding everything else fixed:
//
//	(p.catch(() => {}));       SILENT    parens around the statement expression, skipped
//	(p).catch(() => {});       SILENT    parens around the RECEIVER, the checker sees through
//	(p.catch)(() => {});       REPORTS   parens around the CALLEE, the access test declines
//
// The first of those is the case that separates the skip from its absence. Dropping the skip leaves
// `(p.catch(() => {}))` reporting, and every one of the 169 imported cases stays green, because
// none of them writes a parenthesis. It was measured against BOTH references before being written
// here: `@typescript-eslint` is silent on it too, so this is shared behavior rather than a choice.
//
// The third row is a real divergence and it is recorded in the rule's doc comment: `@typescript-
// eslint` is SILENT on `(p.catch)(() => {})` because ESTree has no parenthesized-expression node,
// and tsgolint reports because typescript-go keeps one. tsgolint wins, so the report is deliberate.
func TestNoFloatingPromisesMeasuredParentheses(t *testing.T) {
	t.Parallel()

	prelude := "declare const p: Promise<number>;\n"
	cases := []struct {
		name    string
		body    string
		wantIds []string
	}{
		{"parens around a handled call are transparent", "(p.catch(() => {}));", nil},
		{"parens around a handled two-argument then are transparent", "(p.then(() => {}, () => {}));", nil},
		{"doubled parens around a handled call are transparent", "((p.catch(() => {})));", nil},
		{"parens around the receiver stay handled", "(p).catch(() => {});", nil},
		{"parens around a bare promise still report", "(p);", []string{"floatingVoid"}},
		{"parens around the callee make the handler invisible", "(p.catch)(() => {});", []string{"floatingVoid"}},
		{"parens around a then callee make the handler invisible", "(p.then)(() => {}, () => {});", []string{"floatingVoid"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoFloatingPromises(t, prelude+"async function wrapper() {\n  "+testCase.body+"\n}\n", nil)
			if testCase.wantIds == nil {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoFloatingPromisesReportsOncePerStatement pins that an array of floating promises is ONE
// finding against the whole statement rather than one per element.
//
// The anchor is `KindExpressionStatement`, so per-element reporting is not a close call the corpus
// happens not to cover, it is structurally impossible. The corpus's array cases are all single
// statements, so nothing in it separates "reports once because it groups" from "reports once
// because there is one statement". A TRIPLE does, and every position is read rather than the count
// alone.
func TestNoFloatingPromisesReportsOncePerStatement(t *testing.T) {
	t.Parallel()

	triple := "declare const a: Promise<number>;\ndeclare const b: Promise<number>;\ndeclare const c: Promise<number>;\n"

	grouped := runNoFloatingPromises(t, triple+"[a, b, c];\n", nil)
	rule_testing.ExpectFindings(t, grouped, "floatingPromiseArrayVoid")
	if got := len(grouped.Diagnostics[0].Suggestions); got != 0 {
		t.Errorf("the promise-array finding offers %d suggestions, want none", got)
	}
	if got := sourceOf(t, triple+"[a, b, c];\n", grouped, 0); got != "[a, b, c];" {
		t.Errorf("the promise-array finding points at %q, want the whole statement", got)
	}

	separate := runNoFloatingPromises(t, triple+"a;\nb;\nc;\n", nil)
	rule_testing.ExpectFindings(t, separate, "floatingVoid", "floatingVoid", "floatingVoid")
	for index, want := range []string{"a;", "b;", "c;"} {
		if got := sourceOf(t, triple+"a;\nb;\nc;\n", separate, index); got != want {
			t.Errorf("finding %d points at %q, want %q", index, got, want)
		}
	}

	// Promise.all of a triple is ONE floating CALL, not a promise array: the argument is handled by
	// Promise.all and the call's own result is what floats.
	aggregated := runNoFloatingPromises(t, "Promise.all([Promise.resolve(1), Promise.resolve(2), Promise.resolve(3)]);\n", nil)
	rule_testing.ExpectFindings(t, aggregated, "floatingVoid")
}

// sourceOf slices the trimmed source with a finding's own range, which is the only assertion that
// can see WHERE a finding points.
func sourceOf(t *testing.T, sourceText string, result rule_testing.Result, index int) string {
	t.Helper()
	trimmed := strings.TrimSpace(sourceText)
	diagnostic := result.Diagnostics[index]
	start, end := diagnostic.Range.Pos(), diagnostic.Range.End()
	if start < 0 || end > len(trimmed) || start > end {
		t.Fatalf("finding %d has range [%d,%d) over %d bytes", index, start, end, len(trimmed))
	}
	return trimmed[start:end]
}

// TestNoFloatingPromisesMessagesReadAsUpstreamWrote asserts the rendered text of all eight
// messages against literals typed here rather than against the rule's own constants.
//
// Comparing a finding to the constant it was built from is equality that moves under mutation:
// both sides change together and the assertion stays green. These strings are typed from
// upstream's source instead, so a reworded message fails.
func TestNoFloatingPromisesMessagesReadAsUpstreamWrote(t *testing.T) {
	t.Parallel()

	wantById := map[string]string{
		"floating":                            "Promises must be awaited, end with a call to .catch, or end with a call to .then with a rejection handler.",
		"floatingVoid":                        "Promises must be awaited, end with a call to .catch, end with a call to .then with a rejection handler or be explicitly marked as ignored with the `void` operator.",
		"floatingUselessRejectionHandler":     "Promises must be awaited, end with a call to .catch, or end with a call to .then with a rejection handler. A rejection handler that is not a function will be ignored.",
		"floatingUselessRejectionHandlerVoid": "Promises must be awaited, end with a call to .catch, end with a call to .then with a rejection handler or be explicitly marked as ignored with the `void` operator. A rejection handler that is not a function will be ignored.",
		"floatingPromiseArray":                "An array of Promises may be unintentional. Consider handling the promises' fulfillment or rejection with Promise.all or similar.",
		"floatingPromiseArrayVoid":            "An array of Promises may be unintentional. Consider handling the promises' fulfillment or rejection with Promise.all or similar, or explicitly marking the expression as ignored with the `void` operator.",
		"floatingFixAwait":                    "Add await operator.",
		"floatingFixVoid":                     "Add void operator to ignore.",
	}
	got := []rule.Message{
		buildFloatingMessage(),
		buildFloatingVoidMessage(),
		buildFloatingUselessRejectionHandlerMessage(),
		buildFloatingUselessRejectionHandlerVoidMessage(),
		buildFloatingPromiseArrayMessage(),
		buildFloatingPromiseArrayVoidMessage(),
		buildFloatingFixAwaitMessage(),
		buildFloatingFixVoidMessage(),
	}
	if len(got) != len(wantById) {
		t.Fatalf("the rule builds %d messages, want %d", len(got), len(wantById))
	}
	for _, message := range got {
		want, ok := wantById[message.Id]
		if !ok {
			t.Errorf("the rule reports an id upstream does not have: %q", message.Id)
			continue
		}
		if message.Description != want {
			t.Errorf("message %q reads\n%q\nwant\n%q", message.Id, message.Description, want)
		}
	}
}

// TestNoFloatingPromisesRequiresTheTypedHarness pins the checker declaration and the nil guard
// together, so a later revert fails loudly rather than going green.
//
// This rule is the SILENT kind rather than the panicking kind. Every type question it asks routes
// through `ctx.TypeChecker`, and on this shim `GetTypeAtLocation` returns nil rather than crashing,
// so a rule whose guard was dropped would not blow up: `isPromiseLike` would answer false for
// everything and the whole corpus would go quiet. Ninety-seven reporting fixtures would fail in a
// way that reads like a rule defect, and seventy-two clean ones would pass VACUOUSLY. The count
// below is what makes the difference visible.
//
// The guard could not be written while this rule lived behind the tsgolint adapter, because the
// adapter set NeedsTypeChecker on everything it wrapped and the listener was upstream's. Absorbing
// the rule made the nil case reachable AND made the listener ours, so the guard is written here.
func TestNoFloatingPromisesRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoFloatingPromises.NeedsTypeChecker {
		t.Fatal("the rule stopped declaring NeedsTypeChecker, so every typed fixture would run against a nil checker")
	}
	if NoFloatingPromises.ProgramReads&rule.ReadsDefaultLibrary == 0 {
		t.Fatal("the rule stopped declaring ReadsDefaultLibrary, and IsPromiseLike asks whether a declaration is the default library's")
	}

	source := "declare const p: Promise<number>;\np;\n"
	typed := runNoFloatingPromises(t, source, nil)
	rule_testing.ExpectFindings(t, typed, "floatingVoid")

	untyped := rule_testing.Run(t, NoFloatingPromises, noFloatingPromisesFile, source)
	rule_testing.ExpectClean(t, untyped)
}

// TestNoFloatingPromisesThenableNeedsTwoFunctionParameters pins upstream's deliberate narrowing of
// the thenable test, which the imported corpus does not separate from a plain "has a then".
//
// Upstream's own comment says it: this is `tsutils.isThenable()` modified "to only consider
// thenables which can be rejected/caught via a second parameter". A `then` taking ONE function is
// not something this rule can tell you to handle, because there is nowhere to put the rejection
// handler, so it is left alone. Both parameters must be callable; a second parameter that is not a
// function does not count either.
//
// Loosening the requirement to one parameter leaves every one of the 169 imported cases green,
// because the corpus's `checkThenables` cases all declare a two-function `then`. Each row below was
// measured against `@typescript-eslint` before being written here, and both references agree on
// all five.
func TestNoFloatingPromisesThenableNeedsTwoFunctionParameters(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{"one function parameter is not rejectable", "declare const t: { then(onOk: () => void): void };\nt;\n", nil},
		{"no parameters at all", "declare const t: { then(): void };\nt;\n", nil},
		{"a second parameter that is not a function", "declare const t: { then(onOk: () => void, onErr: number): void };\nt;\n", nil},
		{"a first parameter that is not a function", "declare const t: { then(onOk: number, onErr: () => void): void };\nt;\n", nil},
		{"two function parameters", "declare const t: { then(onOk: () => void, onErr: () => void): void };\nt;\n", []string{"floatingVoid"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runNoFloatingPromises(t, testCase.sourceText,
				NoFloatingPromisesOptions{CheckThenables: type_checking.Ref(true)})
			if testCase.wantIds == nil {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoFloatingPromisesResolvesAStaticMemberName pins the one place this rule genuinely diverges
// from `@typescript-eslint`, and pins that the resolution reaches ACROSS FILES.
//
// tsgolint uses `checker.Checker_getAccessedPropertyName`, which resolves a computed member key
// through the type checker. `@typescript-eslint` uses `getStaticMemberAccessValue`, which reads
// only literal and simple-constant keys. Upstream's source marks the difference with a
// `TODO(port)` calling its own version "an enhancement" and noting it lacks tests, so this is a
// known, deliberate divergence rather than an accident of the port. Measured on one input:
//
//	enum Methods { Catch = 'catch' }
//	p[Methods.Catch](() => {});          tsgolint SILENT    @typescript-eslint REPORTS
//
// tsgolint wins, because oxlint runs tsgolint and the differential harness compares against
// oxlint, so reproducing the plugin's answer would be a difference the harness could see.
//
// The cross-file case is why this rule's findings are keyed on the type fingerprint rather than the
// linted file alone. `type_checking.IsPromiseLike` asks about the default library on the core path of
// every promise test, and enum resolution reaches the declaring module.
func TestNoFloatingPromisesResolvesAStaticMemberName(t *testing.T) {
	t.Parallel()

	sameFile := "enum Methods {\n  Catch = 'catch',\n}\ndeclare const p: Promise<number>;\np[Methods.Catch](() => {});\n"
	rule_testing.ExpectClean(t, runNoFloatingPromises(t, sameFile, nil))

	crossFile := rule_testing.RunTypedFilesWithOptions(t, NoFloatingPromises, map[string]string{
		"methods.ts":           "export enum Methods {\n  Catch = 'catch',\n}\n",
		noFloatingPromisesFile: "import { Methods } from './methods';\ndeclare const p: Promise<number>;\np[Methods.Catch](() => {});\n",
	}, noFloatingPromisesFile, nil)
	rule_testing.ExpectClean(t, crossFile)

	// The control: the same shape with a key that resolves to something else still reports, so the
	// two clean verdicts above are resolution rather than a blanket exemption for computed keys.
	other := "enum Methods {\n  Finally = 'finally',\n}\ndeclare const p: Promise<number>;\np[Methods.Finally](() => {});\n"
	rule_testing.ExpectFindings(t, runNoFloatingPromises(t, other, nil), "floatingVoid")
}

// TestDecodeNoFloatingPromisesOptions routes configuration through the rule's own exported decoder,
// which is where the two translations with no upstream counterpart live.
//
// The brief's standing warning applies squarely here: handing `RunWithOptions` a struct built in Go
// would leave both translations untested. The heterogeneous allowlist splitting into two typed
// fields, and `from` moving from a string to an integer enum whose zero value is `file`, are both
// invisible from a struct literal.
//
// The `ignoreVoid` rows are the ones that matter most. Its default is TRUE, so a decoder binding it
// to a plain bool would turn every unconfigured run into an explicit `ignoreVoid: false`, inverting
// almost the whole corpus. It stays a pointer end to end so that absent and false stay different.
func TestDecodeNoFloatingPromisesOptions(t *testing.T) {
	t.Parallel()

	decode := func(t *testing.T, raw string) NoFloatingPromisesOptions {
		t.Helper()
		decoded, err := DecodeNoFloatingPromisesOptions([]byte(raw))
		if err != nil {
			t.Fatalf("decoding %s: %v", raw, err)
		}
		options, ok := decoded.(NoFloatingPromisesOptions)
		if !ok {
			t.Fatalf("the decoder returned %T, want NoFloatingPromisesOptions", decoded)
		}
		return options
	}

	t.Run("an empty object leaves every pointer nil so the rule's own defaults apply", func(t *testing.T) {
		t.Parallel()
		options := decode(t, `{}`)
		if options.IgnoreVoid != nil || options.IgnoreIIFE != nil || options.CheckThenables != nil {
			t.Errorf("an empty object should leave the three booleans nil, got %v %v %v",
				options.IgnoreVoid, options.IgnoreIIFE, options.CheckThenables)
		}
	})

	t.Run("an explicit false is not the same as absent", func(t *testing.T) {
		t.Parallel()
		options := decode(t, `{"ignoreVoid": false}`)
		if options.IgnoreVoid == nil {
			t.Fatal("an explicit ignoreVoid should decode to a non-nil pointer")
		}
		if *options.IgnoreVoid {
			t.Error("ignoreVoid decoded to true, want false")
		}
	})

	t.Run("a bare string specifier lands in the inline list", func(t *testing.T) {
		t.Parallel()
		options := decode(t, `{"allowForKnownSafePromises": ["SafePromise"]}`)
		if len(options.AllowForKnownSafePromises) != 0 {
			t.Errorf("a bare string should not become a structured specifier, got %v", options.AllowForKnownSafePromises)
		}
		if len(options.AllowForKnownSafePromisesInline) != 1 || options.AllowForKnownSafePromisesInline[0] != "SafePromise" {
			t.Errorf("the inline list reads %v, want [SafePromise]", options.AllowForKnownSafePromisesInline)
		}
	})

	t.Run("from is translated from its string spelling to the enum", func(t *testing.T) {
		t.Parallel()
		options := decode(t, `{"allowForKnownSafePromises":[{"from":"lib","name":"PromiseLike"}]}`)
		if len(options.AllowForKnownSafePromises) != 1 {
			t.Fatalf("want one specifier, got %d", len(options.AllowForKnownSafePromises))
		}
		specifier := options.AllowForKnownSafePromises[0]
		if specifier.From != type_checking.TypeOrValueSpecifierFromLib {
			t.Errorf("from decoded to %v, want the lib enum", specifier.From)
		}
		// A bare string name widens to a one-element slice rather than erroring.
		if len(specifier.Name) != 1 || specifier.Name[0] != "PromiseLike" {
			t.Errorf("name decoded to %v, want [PromiseLike]", specifier.Name)
		}
	})

	t.Run("an unrecognized from is dropped rather than silently meaning file", func(t *testing.T) {
		t.Parallel()
		// `file` is the zero value of the enum, so a decoder that merely failed to set From would
		// turn a typo into a file specifier. This asserts the specifier is skipped instead.
		options := decode(t, `{"allowForKnownSafeCalls":[{"from":"module","name":"it"}]}`)
		if len(options.AllowForKnownSafeCalls) != 0 {
			t.Errorf("an unrecognized from should be dropped, got %v", options.AllowForKnownSafeCalls)
		}
	})

	t.Run("both wire forms coexist in one array", func(t *testing.T) {
		t.Parallel()
		options := decode(t, `{"allowForKnownSafeCalls":["it",{"from":"package","name":["it","test"],"package":"node:test"}]}`)
		if len(options.AllowForKnownSafeCallsInline) != 1 || options.AllowForKnownSafeCallsInline[0] != "it" {
			t.Errorf("the inline list reads %v, want [it]", options.AllowForKnownSafeCallsInline)
		}
		if len(options.AllowForKnownSafeCalls) != 1 {
			t.Fatalf("want one structured specifier, got %d", len(options.AllowForKnownSafeCalls))
		}
		specifier := options.AllowForKnownSafeCalls[0]
		if specifier.From != type_checking.TypeOrValueSpecifierFromPackage || specifier.Package != "node:test" {
			t.Errorf("the structured specifier reads %+v, want a node:test package specifier", specifier)
		}
		if len(specifier.Name) != 2 {
			t.Errorf("name decoded to %v, want two entries", specifier.Name)
		}
	})

	t.Run("a decoded allowlist actually reaches the rule", func(t *testing.T) {
		t.Parallel()
		// The decoder can be right and the wiring still wrong, so one end-to-end row runs a real
		// input through the decoded options rather than through a struct built here.
		source := "type SafePromise = Promise<number> & { __linterBrands?: string };\ndeclare const p: SafePromise;\np;\n"
		rule_testing.ExpectFindings(t, runNoFloatingPromises(t, source, nil), "floatingVoid")

		decoded, err := DecodeNoFloatingPromisesOptions([]byte(`{"allowForKnownSafePromises":[{"from":"file","name":"SafePromise"}]}`))
		if err != nil {
			t.Fatalf("decoding: %v", err)
		}
		rule_testing.ExpectClean(t, runNoFloatingPromises(t, source, decoded))
	})
}

// TestNoFloatingPromisesResolvesThroughANamespaceImport pins that an allowlisted call is recognised
// when it arrives through a namespace import rather than a named one.
//
// `import * as Runner from './runner'` then `Runner.describe(...)` is the same function as
// `import { describe }` then `describe(...)`, and upstream exempts both. Written after an
// `allowForKnownSafeCalls` experiment on the ahra tree cleared 144 of this rule's 150 findings
// against ESLint's 150, which made namespace imports look like the boundary. They are not: this
// test passes, and the six survivors are all `describe` reached through either import form. The
// test is kept because the property is worth holding and nothing else held it.
//
// Written against the multi-file harness deliberately. A single-file `declare module` does not
// resolve its own imports here, so both forms report nothing at all and every assertion passes
// without testing anything. That is measured, not assumed: a local object with a promise-returning
// method reports `floatingVoid`, and the same call through a declared module reports nothing in
// either import form.
func TestNoFloatingPromisesResolvesThroughANamespaceImport(t *testing.T) {
	t.Parallel()

	const runnerModule = "export function describe(name: string, action: () => void): Promise<void> {\n" +
		"  return Promise.resolve();\n" +
		"}\n" +
		"export function teardown(): Promise<void> {\n" +
		"  return Promise.resolve();\n" +
		"}\n"

	options := NoFloatingPromisesOptions{
		AllowForKnownSafeCalls: []type_checking.TypeOrValueSpecifier{
			{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"describe"}, Path: "runner.ts"},
		},
	}

	runWith := func(t *testing.T, entry string) rule_testing.Result {
		t.Helper()
		return rule_testing.RunTypedFilesWithOptions(t, NoFloatingPromises, map[string]string{
			"runner.ts":            runnerModule,
			noFloatingPromisesFile: entry,
		}, noFloatingPromisesFile, options)
	}

	// The control comes first, because it is what proves the other two mean anything: same module,
	// same namespace, a name the allowlist does not carry.
	rule_testing.ExpectFindings(t,
		runWith(t, "import * as Runner from './runner';\nRunner.teardown();\n"), "floatingVoid")

	rule_testing.ExpectClean(t,
		runWith(t, "import { describe } from './runner';\ndescribe('...', () => {});\n"))

	rule_testing.ExpectClean(t,
		runWith(t, "import * as Runner from './runner';\nRunner.describe('...', () => {});\n"))
}
