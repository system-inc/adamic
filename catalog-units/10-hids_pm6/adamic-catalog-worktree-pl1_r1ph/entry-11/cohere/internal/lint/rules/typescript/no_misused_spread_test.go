package typescript

import (
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus for no-misused-spread, taken verbatim from upstream's own tester.
//
// Extracted by PARSING the upstream test file with the TypeScript compiler and serialising each
// string through a JSON encoder, so nothing was retyped or passed through a shell, then byte-compared
// against the source with a control asserted absent: all 131 cases match, one of them under the
// backtick escaping a template literal requires.
//
// The whole corpus was driven through the INSTALLED rule at 8.67.0 on a real type graph, honouring
// each case's own options and giving the nine JSX cases a .tsx filename. It agreed with upstream's
// recorded expectations on all 131 inputs and reproduced all 21 recorded suggestion outputs, so the
// oracle used to settle the questions below is pinned against the corpus rather than trusted.
//
// One upstream case is absent below and it is a HARNESS fact rather than a rule fact. See
// TestNoMisusedSpreadCaseThatCannotBeExpressedHere.
const (
	noMisusedSpreadFile    = "file.ts"
	noMisusedSpreadJsxFile = "file.tsx"
)

// TestNoMisusedSpreadStaysSilent is upstream's valid list, each under its own options.
func TestNoMisusedSpreadStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
		options    any
	}{
		{"upstream valid 0", "file.ts", "const a = [...[1, 2, 3]];", nil},
		{"upstream valid 1", "file.ts", "const a = [...([1, 2, 3] as const)];", nil},
		{"upstream valid 2", "file.ts", "\ndeclare const data: any;\nconst a = [...data];\n    ", nil},
		{"upstream valid 3", "file.ts", "\ndeclare const data: unknown;\nconst a = [...data];\n    ", nil},
		{"upstream valid 4", "file.ts", "\nconst a = [1, 2, 3];\nconst b = [...a];\n    ", nil},
		{"upstream valid 5", "file.ts", "\nconst a = [1, 2, 3] as const;\nconst b = [...a];\n    ", nil},
		{"upstream valid 6", "file.ts", "\ndeclare function getArray(): number[];\nconst a = [...getArray()];\n    ", nil},
		{"upstream valid 7", "file.ts", "\ndeclare function getTuple(): readonly number[];\nconst a = [...getTuple()];\n    ", nil},
		{"upstream valid 8", "file.ts", "\nconst iterator = {\n  *[Symbol.iterator]() {\n    yield 1;\n    yield 2;\n    yield 3;\n  },\n};\n\nconst a = [...iterator];\n    ", nil},
		{"upstream valid 9", "file.ts", "\ndeclare const data: Iterable<number> | number[];\n\nconst a = [...data];\n    ", nil},
		{"upstream valid 10", "file.ts", "\ndeclare const data: Iterable<number> & number[];\n\nconst a = [...data];\n    ", nil},
		{"upstream valid 11", "file.ts", "\ndeclare function getIterable(): Iterable<number>;\n\nconst a = [...getIterable()];\n    ", nil},
		{"upstream valid 12", "file.ts", "\ndeclare const data: Uint8Array;\n\nconst a = [...data];\n    ", nil},
		{"upstream valid 13", "file.ts", "\ndeclare const data: TypedArray;\n\nconst a = [...data];\n    ", nil},
		{"upstream valid 14", "file.ts", "const o = { ...{ a: 1, b: 2 } };", nil},
		{"upstream valid 15", "file.ts", "const o = { ...({ a: 1, b: 2 } as const) };", nil},
		{"upstream valid 16", "file.ts", "\ndeclare const obj: any;\n\nconst o = { ...obj };\n    ", nil},
		{"upstream valid 17", "file.ts", "\ndeclare const obj: { a: number; b: number } | any;\n\nconst o = { ...obj };\n    ", nil},
		{"upstream valid 18", "file.ts", "\ndeclare const obj: { a: number; b: number } & any;\n\nconst o = { ...obj };\n    ", nil},
		{"upstream valid 19", "file.ts", "\nconst obj = { a: 1, b: 2 };\nconst o = { ...obj };\n    ", nil},
		{"upstream valid 20", "file.ts", "\ndeclare const obj: { a: number; b: number };\nconst o = { ...obj };\n    ", nil},
		{"upstream valid 21", "file.ts", "\ndeclare function getObject(): { a: number; b: number };\nconst o = { ...getObject() };\n    ", nil},
		{"upstream valid 22", "file.ts", "\nfunction f() {}\n\nf.prop = 1;\n\nconst o = { ...f };\n    ", nil},
		{"upstream valid 23", "file.ts", "\nconst f = () => {};\n\nf.prop = 1;\n\nconst o = { ...f };\n    ", nil},
		{"upstream valid 24", "file.ts", "\nfunction* generator() {}\n\ngenerator.prop = 1;\n\nconst o = { ...generator };\n    ", nil},
		{"upstream valid 25", "file.ts", "\ndeclare const promiseLike: PromiseLike<number>;\n\nconst o = { ...promiseLike };\n    ", nil},
		{"upstream valid 26", "file.tsx", "\nconst obj = { a: 1, b: 2 };\nconst o = <div {...x} />;\n      ", nil},
		{"upstream valid 27", "file.tsx", "\ndeclare const obj: { a: number; b: number } | any;\nconst o = <div {...x} />;\n      ", nil},
		{"upstream valid 28", "file.ts", "\nconst promise = new Promise(() => {});\nconst o = { ...promise };\n      ", NoMisusedSpreadOptions{AllowInline: []string{"Promise"}}},
		{"upstream valid 29", "file.ts", "\ninterface A {}\n\ndeclare const a: A;\n\nconst o = { ...a };\n    ", nil},
		{"upstream valid 30", "file.ts", "\nconst o = { ...'test' };\n    ", nil},
		{"upstream valid 31", "file.ts", "\nconst str: string = 'test';\nconst a = [...str];\n      ", NoMisusedSpreadOptions{AllowInline: []string{"string"}}},
		{"upstream valid 32", "file.ts", "\nfunction f() {}\n\nconst a = { ...f };\n      ", NoMisusedSpreadOptions{AllowInline: []string{"f"}}},
		{"upstream valid 33", "file.ts", "\ndeclare const iterator: Iterable<string>;\n\nconst a = { ...iterator };\n      ", NoMisusedSpreadOptions{Allow: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromLib, Name: []string{"Iterable"}}}}},
		{"upstream valid 34", "file.ts", "\ntype BrandedString = string & { __brand: 'safe' };\n\ndeclare const brandedString: BrandedString;\n\nconst spreadBrandedString = [...brandedString];\n      ", NoMisusedSpreadOptions{Allow: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"BrandedString"}}}}},
		{"upstream valid 35", "file.ts", "\ntype CustomIterable = {\n  [Symbol.iterator]: () => Generator<string>;\n};\n\ndeclare const iterator: CustomIterable;\n\nconst a = { ...iterator };\n      ", NoMisusedSpreadOptions{AllowInline: []string{"CustomIterable"}}},
		{"upstream valid 36", "file.ts", "\ntype CustomIterable = {\n  [Symbol.iterator]: () => string;\n};\n\ndeclare const iterator: CustomIterable;\n\nconst a = { ...iterator };\n      ", NoMisusedSpreadOptions{Allow: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"CustomIterable"}}}}},
		{"upstream valid 37", "file.ts", "\ndeclare module 'module' {\n  export type CustomIterable = {\n    [Symbol.iterator]: () => string;\n  };\n}\n\nimport { CustomIterable } from 'module';\n\ndeclare const iterator: CustomIterable;\n\nconst a = { ...iterator };\n      ", NoMisusedSpreadOptions{Allow: []type_checking.TypeOrValueSpecifier{{From: type_checking.TypeOrValueSpecifierFromPackage, Name: []string{"CustomIterable"}, Package: "module"}}}},
		{"upstream valid 38", "file.ts", "\nclass A {\n  a = 1;\n}\n\nconst a = new A();\n\nconst o = { ...a };\n      ", NoMisusedSpreadOptions{AllowInline: []string{"A"}}},
		{"upstream valid 39", "file.ts", "\nconst a = {\n  ...class A {\n    static value = 1;\n  },\n};\n      ", NoMisusedSpreadOptions{AllowInline: []string{"A"}}}}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMisusedSpread, testCase.fileName, testCase.sourceText, testCase.options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoMisusedSpreadFires is upstream's invalid list, with the ids it records per case.
//
// Upstream invalid 0 through 13 are absent: they are the string branch, which cohere drops on
// purpose, and they sit in TestNoMisusedSpreadLeavesStringSpreadAlone as must-stay-silent rows.
//
// Seven of the rule's nine message ids appear here; the other two are the suggestion ids, which no
// finding carries and which the suggestion test below asserts instead.
//
// The ordering of the object cascade is what most of these rows really pin. A Map is iterable, an
// array is iterable, and a class instance can be, so several of these inputs satisfy two arms at
// once and the id recorded is the one the earlier arm produces.
func TestNoMisusedSpreadFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
		options    any
		wantIds    []string
	}{
		{"upstream invalid 14", "file.ts", "const o = { ...[1, 2, 3] };", nil, []string{"noArraySpreadInObject"}},
		{"upstream invalid 15", "file.ts", "\nconst arr = [1, 2, 3];\nconst o = { ...arr };\n      ", nil, []string{"noArraySpreadInObject"}},
		{"upstream invalid 16", "file.ts", "\nconst arr = [1, 2, 3] as const;\nconst o = { ...arr };\n      ", nil, []string{"noArraySpreadInObject"}},
		{"upstream invalid 17", "file.ts", "\ndeclare const arr: number[];\nconst o = { ...arr };\n      ", nil, []string{"noArraySpreadInObject"}},
		{"upstream invalid 18", "file.ts", "\ndeclare const arr: readonly number[];\nconst o = { ...arr };\n      ", nil, []string{"noArraySpreadInObject"}},
		{"upstream invalid 19", "file.ts", "\ndeclare const arr: number[] | string[];\nconst o = { ...arr };\n      ", nil, []string{"noArraySpreadInObject"}},
		{"upstream invalid 20", "file.ts", "\ndeclare const arr: number[] & string[];\nconst o = { ...arr };\n      ", nil, []string{"noArraySpreadInObject"}},
		{"upstream invalid 21", "file.ts", "\ndeclare function getArray(): number[];\nconst o = { ...getArray() };\n      ", nil, []string{"noArraySpreadInObject"}},
		{"upstream invalid 22", "file.ts", "\ndeclare function getArray(): readonly number[];\nconst o = { ...getArray() };\n      ", nil, []string{"noArraySpreadInObject"}},
		{"upstream invalid 23", "file.ts", "const o = { ...new Set([1, 2, 3]) };", nil, []string{"noIterableSpreadInObject"}},
		{"upstream invalid 24", "file.ts", "\nconst set = new Set([1, 2, 3]);\nconst o = { ...set };\n      ", nil, []string{"noIterableSpreadInObject"}},
		{"upstream invalid 25", "file.ts", "\ndeclare const set: Set<number>;\nconst o = { ...set };\n      ", nil, []string{"noIterableSpreadInObject"}},
		{"upstream invalid 26", "file.ts", "\ndeclare const set: WeakSet<object>;\nconst o = { ...set };\n      ", nil, []string{"noClassInstanceSpreadInObject"}},
		{"upstream invalid 27", "file.ts", "\ndeclare const set: ReadonlySet<number>;\nconst o = { ...set };\n      ", nil, []string{"noIterableSpreadInObject"}},
		{"upstream invalid 28", "file.ts", "\ndeclare const set: Set<number> | { a: number };\nconst o = { ...set };\n      ", nil, []string{"noIterableSpreadInObject"}},
		{"upstream invalid 29", "file.ts", "\ndeclare function getSet(): Set<number>;\nconst o = { ...getSet() };\n      ", nil, []string{"noIterableSpreadInObject"}},
		{"upstream invalid 30", "file.ts", "\nconst o = {\n  ...new Map([\n    ['test-1', 1],\n    ['test-2', 2],\n  ]),\n};\n      ", nil, []string{"noMapSpreadInObject"}},
		{"upstream invalid 31", "file.ts", "\nconst map = new Map([\n  ['test-1', 1],\n  ['test-2', 2],\n]);\n\nconst o = { ...map };\n      ", nil, []string{"noMapSpreadInObject"}},
		{"upstream invalid 32", "file.ts", "\ndeclare const map: Map<string, number>;\nconst o = { ...map };\n      ", nil, []string{"noMapSpreadInObject"}},
		{"upstream invalid 33", "file.ts", "\n        declare const map: Map<string, number>;\n        const o = { ...(map) };\n      ", nil, []string{"noMapSpreadInObject"}},
		{"upstream invalid 34", "file.ts", "\ndeclare const map: Map<string, number>;\nconst o = { ...(map, map) };\n      ", nil, []string{"noMapSpreadInObject"}},
		{"upstream invalid 35", "file.ts", "\ndeclare const map: Map<string, number>;\nconst others = { a: 1 };\nconst o = { ...map, ...others };\n      ", nil, []string{"noMapSpreadInObject"}},
		{"upstream invalid 36", "file.ts", "\ndeclare const map: Map<string, number>;\nconst o = { other: 1, ...map };\n      ", nil, []string{"noMapSpreadInObject"}},
		{"upstream invalid 37", "file.ts", "\ndeclare const map: ReadonlyMap<string, number>;\nconst o = { ...map };\n      ", nil, []string{"noMapSpreadInObject"}},
		{"upstream invalid 38", "file.ts", "\ndeclare const map: WeakMap<{ a: number }, string>;\nconst o = { ...map };\n      ", nil, []string{"noMapSpreadInObject"}},
		{"upstream invalid 39", "file.ts", "\ndeclare const map: Map<string, number> | { a: number };\nconst o = { ...map };\n      ", nil, []string{"noMapSpreadInObject"}},
		{"upstream invalid 40", "file.ts", "\ndeclare function getMap(): Map<string, number>;\nconst o = { ...getMap() };\n      ", nil, []string{"noMapSpreadInObject"}},
		{"upstream invalid 41", "file.ts", "\ndeclare const a: Map<boolean, string> & Set<number>;\nconst o = { ...a };\n      ", nil, []string{"noMapSpreadInObject"}},
		{"upstream invalid 42", "file.ts", "\nconst ref = new WeakRef({ a: 1 });\nconst o = { ...ref };\n      ", nil, []string{"noClassInstanceSpreadInObject"}},
		{"upstream invalid 43", "file.ts", "\nconst promise = new Promise(() => {});\nconst o = { ...promise };\n      ", nil, []string{"noPromiseSpreadInObject"}},
		{"upstream invalid 44", "file.ts", "\ndeclare const promise: Promise<{ a: 1 }>;\nasync function foo() {\n  return { ...(promise || {}) };\n}\n      ", nil, []string{"noPromiseSpreadInObject"}},
		{"upstream invalid 45", "file.ts", "\ndeclare const promise: Promise<any>;\nasync function foo() {\n  return { ...(Math.random() < 0.5 ? promise : {}) };\n}\n      ", nil, []string{"noPromiseSpreadInObject"}},
		{"upstream invalid 46", "file.ts", "\nfunction withPromise<P extends Promise<void>>(promise: P) {\n  return { ...promise };\n}\n      ", nil, []string{"noPromiseSpreadInObject"}},
		{"upstream invalid 47", "file.ts", "\ndeclare const maybePromise: Promise<number> | { a: number };\nconst o = { ...maybePromise };\n      ", nil, []string{"noPromiseSpreadInObject"}},
		{"upstream invalid 48", "file.ts", "\ndeclare const promise: Promise<number> & { a: number };\nconst o = { ...promise };\n      ", nil, []string{"noPromiseSpreadInObject"}},
		{"upstream invalid 49", "file.ts", "\ndeclare function getPromise(): Promise<number>;\nconst o = { ...getPromise() };\n      ", nil, []string{"noPromiseSpreadInObject"}},
		{"upstream invalid 50", "file.ts", "\ndeclare function getPromise<T extends Promise<number>>(arg: T): T;\nconst o = { ...getPromise() };\n      ", nil, []string{"noPromiseSpreadInObject"}},
		{"upstream invalid 51", "file.ts", "\nfunction f() {}\n\nconst o = { ...f };\n      ", nil, []string{"noFunctionSpreadInObject"}},
		{"upstream invalid 52", "file.ts", "\ninterface FunctionWithProps {\n  (): string;\n  prop: boolean;\n}\n\ntype FunctionWithoutProps = () => string;\n\ndeclare const obj: FunctionWithProps | FunctionWithoutProps | object;\n\nconst o = { ...obj };\n      ", nil, []string{"noFunctionSpreadInObject"}},
		{"upstream invalid 53", "file.ts", "\nconst f = () => {};\n\nconst o = { ...f };\n      ", nil, []string{"noFunctionSpreadInObject"}},
		{"upstream invalid 54", "file.ts", "\ndeclare function f(): void;\n\nconst o = { ...f };\n      ", nil, []string{"noFunctionSpreadInObject"}},
		{"upstream invalid 55", "file.ts", "\ndeclare function getFunction(): () => void;\n\nconst o = { ...getFunction() };\n      ", nil, []string{"noFunctionSpreadInObject"}},
		{"upstream invalid 56", "file.ts", "\ndeclare const f: () => void;\n\nconst o = { ...f };\n      ", nil, []string{"noFunctionSpreadInObject"}},
		{"upstream invalid 57", "file.ts", "\ndeclare const f: () => void | { a: number };\n\nconst o = { ...f };\n      ", nil, []string{"noFunctionSpreadInObject"}},
		{"upstream invalid 58", "file.ts", "\nfunction* generator() {}\n\nconst o = { ...generator };\n      ", nil, []string{"noFunctionSpreadInObject"}},
		{"upstream invalid 59", "file.ts", "\nconst iterator = {\n  *[Symbol.iterator]() {\n    yield 'test';\n  },\n};\n\nconst o = { ...iterator };\n      ", nil, []string{"noIterableSpreadInObject"}},
		{"upstream invalid 60", "file.ts", "\ntype CustomIterable = {\n  [Symbol.iterator]: () => Generator<string>;\n};\n\nconst iterator: CustomIterable = {\n  *[Symbol.iterator]() {\n    yield 'test';\n  },\n};\n\nconst a = { ...iterator };\n      ", NoMisusedSpreadOptions{AllowInline: []string{"AnotherIterable"}}, []string{"noIterableSpreadInObject"}},
		{"upstream invalid 62", "file.ts", "\ndeclare const iterator: Iterable<string>;\n\nconst o = { ...iterator };\n      ", nil, []string{"noIterableSpreadInObject"}},
		{"upstream invalid 63", "file.ts", "\ndeclare const iterator: Iterable<string> | { a: number };\n\nconst o = { ...iterator };\n      ", nil, []string{"noIterableSpreadInObject"}},
		{"upstream invalid 64", "file.ts", "\ndeclare function getIterable(): Iterable<string>;\n\nconst o = { ...getIterable() };\n      ", nil, []string{"noIterableSpreadInObject"}},
		{"upstream invalid 65", "file.ts", "\nclass A {\n  [Symbol.iterator]() {\n    return {\n      next() {\n        return { done: true, value: undefined };\n      },\n    };\n  }\n}\n\nconst a = { ...new A() };\n      ", nil, []string{"noIterableSpreadInObject"}},
		{"upstream invalid 66", "file.ts", "\nconst o = { ...new Date() };\n      ", nil, []string{"noClassInstanceSpreadInObject"}},
		{"upstream invalid 67", "file.ts", "\ndeclare class HTMLElementLike {}\ndeclare const element: HTMLElementLike;\nconst o = { ...element };\n      ", nil, []string{"noClassInstanceSpreadInObject"}},
		{"upstream invalid 68", "file.ts", "\ndeclare const regex: RegExp;\nconst o = { ...regex };\n      ", nil, []string{"noClassInstanceSpreadInObject"}},
		{"upstream invalid 77", "file.ts", "\nclass A {}\n\nconst o = { ...A };\n      ", nil, []string{"noClassDeclarationSpreadInObject"}},
		{"upstream invalid 78", "file.ts", "\nconst A = class {};\n\nconst o = { ...A };\n      ", nil, []string{"noClassDeclarationSpreadInObject"}},
		{"upstream invalid 79", "file.ts", "\nclass Declaration {\n  declaration?: boolean;\n}\nconst Expression = class {\n  expression?: boolean;\n};\n\ndeclare const either: typeof Declaration | typeof Expression;\n\nconst o = { ...either };\n      ", nil, []string{"noClassDeclarationSpreadInObject"}},
		{"upstream invalid 80", "file.ts", "\nconst A = Set<number>;\n\nconst o = { ...A };\n      ", nil, []string{"noClassDeclarationSpreadInObject"}},
		{"upstream invalid 81", "file.ts", "\nconst a = {\n  ...class A {\n    static value = 1;\n    nonStatic = 2;\n  },\n};\n      ", nil, []string{"noClassDeclarationSpreadInObject"}},
		{"upstream invalid 82", "file.ts", "\n        const a = { ...(class A { static value = 1 }) }\n      ", nil, []string{"noClassDeclarationSpreadInObject"}},
		{"upstream invalid 84", "file.tsx", "\nconst o = <div {...[1, 2, 3]} />;\n      ", nil, []string{"noArraySpreadInObject"}},
		{"upstream invalid 85", "file.tsx", "\nclass A {}\n\nconst o = <div {...A} />;\n      ", nil, []string{"noClassDeclarationSpreadInObject"}},
		{"upstream invalid 86", "file.tsx", "\nconst o = <div {...new Date()} />;\n      ", nil, []string{"noClassInstanceSpreadInObject"}},
		{"upstream invalid 87", "file.tsx", "\nfunction f() {}\n\nconst o = <div {...f} />;\n      ", nil, []string{"noFunctionSpreadInObject"}},
		{"upstream invalid 88", "file.tsx", "\nconst o = <div {...new Set([1, 2, 3])} />;\n      ", nil, []string{"noIterableSpreadInObject"}},
		{"upstream invalid 89", "file.tsx", "\ndeclare const map: Map<string, number>;\n\nconst o = <div {...map} />;\n      ", nil, []string{"noMapSpreadInObject"}},
		{"upstream invalid 90", "file.tsx", "\nconst promise = new Promise(() => {});\n\nconst o = <div {...promise} />;\n      ", nil, []string{"noPromiseSpreadInObject"}}}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMisusedSpread, testCase.fileName, testCase.sourceText, testCase.options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestNoMisusedSpreadLeavesStringSpreadAlone pins cohere's one deliberate divergence from upstream:
// the string branch is gone, so a string spread into an array, a call or a constructor is silent.
//
// Upstream's `noStringSpread` asks for `Intl.Segmenter` because `[...text]` yields code points
// rather than graphemes. That is a preference about intent the rule cannot see. Spread is the
// correct way to walk a string by code point (it is what the `.split("")` the message lumps it with
// gets wrong), and every string spread in ahra wants code points on purpose: ISO letters mapped to
// regional indicators, a Latin-1 filter, quote-mark scanning. Each of those findings was
// intent-false, and the only repair upstream offers is `Array.from(text)`, the same iteration in a
// different hat, which launders the finding rather than fixing anything. Kirk's ruling, 2026-10-01.
//
// The rows are upstream's invalid 0 through 13 verbatim, then the four ahra sites the gate reports
// (Countries.ts, PngTextMetadata.ts, two in PensieveDailies.ts), then the constructor shape the old
// arm deliberately skipped. The last group is the control: the object cascade in the same harness
// still reports, so silence above is the dropped branch and not a rule that stopped running.
func TestNoMisusedSpreadLeavesStringSpreadAlone(t *testing.T) {
	t.Parallel()

	silent := []struct {
		name       string
		sourceText string
	}{
		{"upstream invalid 0", "const a = [...'test'];"},
		{"upstream invalid 1", "\nfunction withText<Text extends string>(text: Text) {\n  return [...text];\n}\n      "},
		{"upstream invalid 2", "\nconst test = 'hello';\nconst a = [...test];\n      "},
		{"upstream invalid 3", "\nconst test = `he${'ll'}o`;\nconst a = [...test];\n      "},
		{"upstream invalid 4", "\ndeclare const test: string;\nconst a = [...test];\n      "},
		{"upstream invalid 5", "\ndeclare const test: string | number[];\nconst a = [...test];\n      "},
		{"upstream invalid 6", "\ndeclare const test: string & { __brand: 'test' };\nconst a = [...test];\n      "},
		{"upstream invalid 7", "\ndeclare const test: number | (boolean | (string & { __brand: true }));\nconst a = [...test];\n      "},
		{"upstream invalid 8", "\ndeclare function getString(): string;\nconst a = [...getString()];\n      "},
		{"upstream invalid 9", "\ndeclare function textIdentity(...args: string[]);\n\ndeclare const text: string;\n\ntextIdentity(...text);\n      "},
		{"upstream invalid 10", "\ndeclare function textIdentity(...args: string[]);\n\ndeclare const text: string;\n\ntextIdentity(...text, 'and', ...text);\n      "},
		{"upstream invalid 11", "\ndeclare function textIdentity(...args: string[]);\n\nfunction withText<Text extends string>(text: Text) {\n  textIdentity(...text);\n}\n      "},
		{"upstream invalid 12", "\ndeclare function getString<T extends string>(): T;\nconst a = [...getString()];\n      "},
		{"upstream invalid 13", "\ndeclare function getString(): string & { __brand: 'test' };\nconst a = [...getString()];\n      "},
		{"ahra Countries.ts:13, ISO letters to regional indicators", "export function getCountryEmojiByCountryCode(countryCode: string): string {\n    return [...countryCode.toUpperCase()]\n        .map(function(character) {\n            return String.fromCodePoint(0x1f1e6 + character.charCodeAt(0) - 0x41);\n        })\n        .join('');\n}\n"},
		{"ahra PngTextMetadata.ts:64, a Latin-1 filter by code point", "function toLatin1Safe(value: string): string {\n    return [...value].filter((character) => character.charCodeAt(0) <= 0xff).join('');\n}\n"},
		{"ahra PensieveDailies.ts:288, quote marks scanned by code point", "declare const quoteMarkCharacters: Set<string>;\nfunction leadingQuoteCount(anchor: string): number {\n    const characters = [...anchor.trim()];\n    let count = 0;\n    while(count < characters.length && quoteMarkCharacters.has(characters[count]!)) count++;\n    return count;\n}\n"},
		{"ahra PensieveDailies.ts:326, ignored characters filtered by code point", "declare const anchorIgnoredCharacters: Set<string>;\ndeclare function normalizeMomentText(text: string): string;\nfunction normalize(text: string): string {\n    return normalizeMomentText([...text].filter((character) => !anchorIgnoredCharacters.has(character)).join(''));\n}\n"},
		{"a string spread into a constructor", "class C {\n  constructor(...a: string[]) {}\n}\ndeclare const s: string;\nconst x = new C(...s);"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoMisusedSpread, noMisusedSpreadFile, testCase.sourceText, nil))
		})
	}

	t.Run("control: an array spread into an object in the same harness still reports", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, NoMisusedSpread, noMisusedSpreadFile,
			"declare const s: string;\nconst x = [...s];\nconst o = { ...[...s] };", nil), "noArraySpreadInObject")
	})
}

// TestNoMisusedSpreadSuggestions applies each suggested repair and compares the whole rewritten file
// against the output upstream records for it.
//
// Every repair this rule offers is a SUGGESTION, which the engine never applies unattended, so
// ExpectFixedSource cannot be used and the applier is the one await-thenable already needed.
//
// These 21 rows are where this rule can be wrong invisibly, and they cover both fixers. The await
// rows split on precedence: an argument that binds tighter than `await` takes the keyword alone,
// while `(promise || {})` has to be wrapped or the await would attach to the wrong operand. The Map
// rows split on whether the spread is the object's only property, and two of them turn on
// parentheses in opposite directions, since `{ ...(map) }` must lose them and `{ ...(map, map) }`
// must keep them.
func TestNoMisusedSpreadSuggestions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
		options    any
		wantId     string
		wantOutput string
	}{
		{"upstream invalid 30", "file.ts", "\nconst o = {\n  ...new Map([\n    ['test-1', 1],\n    ['test-2', 2],\n  ]),\n};\n      ", nil, "replaceMapSpreadInObject", "\nconst o = Object.fromEntries(new Map([\n    ['test-1', 1],\n    ['test-2', 2],\n  ]));\n      "},
		{"upstream invalid 31", "file.ts", "\nconst map = new Map([\n  ['test-1', 1],\n  ['test-2', 2],\n]);\n\nconst o = { ...map };\n      ", nil, "replaceMapSpreadInObject", "\nconst map = new Map([\n  ['test-1', 1],\n  ['test-2', 2],\n]);\n\nconst o = Object.fromEntries(map);\n      "},
		{"upstream invalid 32", "file.ts", "\ndeclare const map: Map<string, number>;\nconst o = { ...map };\n      ", nil, "replaceMapSpreadInObject", "\ndeclare const map: Map<string, number>;\nconst o = Object.fromEntries(map);\n      "},
		{"upstream invalid 33", "file.ts", "\n        declare const map: Map<string, number>;\n        const o = { ...(map) };\n      ", nil, "replaceMapSpreadInObject", "\n        declare const map: Map<string, number>;\n        const o = Object.fromEntries(map);\n      "},
		{"upstream invalid 34", "file.ts", "\ndeclare const map: Map<string, number>;\nconst o = { ...(map, map) };\n      ", nil, "replaceMapSpreadInObject", "\ndeclare const map: Map<string, number>;\nconst o = Object.fromEntries((map, map));\n      "},
		{"upstream invalid 35", "file.ts", "\ndeclare const map: Map<string, number>;\nconst others = { a: 1 };\nconst o = { ...map, ...others };\n      ", nil, "replaceMapSpreadInObject", "\ndeclare const map: Map<string, number>;\nconst others = { a: 1 };\nconst o = { ...Object.fromEntries(map), ...others };\n      "},
		{"upstream invalid 36", "file.ts", "\ndeclare const map: Map<string, number>;\nconst o = { other: 1, ...map };\n      ", nil, "replaceMapSpreadInObject", "\ndeclare const map: Map<string, number>;\nconst o = { other: 1, ...Object.fromEntries(map) };\n      "},
		{"upstream invalid 37", "file.ts", "\ndeclare const map: ReadonlyMap<string, number>;\nconst o = { ...map };\n      ", nil, "replaceMapSpreadInObject", "\ndeclare const map: ReadonlyMap<string, number>;\nconst o = Object.fromEntries(map);\n      "},
		{"upstream invalid 38", "file.ts", "\ndeclare const map: WeakMap<{ a: number }, string>;\nconst o = { ...map };\n      ", nil, "replaceMapSpreadInObject", "\ndeclare const map: WeakMap<{ a: number }, string>;\nconst o = Object.fromEntries(map);\n      "},
		{"upstream invalid 40", "file.ts", "\ndeclare function getMap(): Map<string, number>;\nconst o = { ...getMap() };\n      ", nil, "replaceMapSpreadInObject", "\ndeclare function getMap(): Map<string, number>;\nconst o = Object.fromEntries(getMap());\n      "},
		{"upstream invalid 41", "file.ts", "\ndeclare const a: Map<boolean, string> & Set<number>;\nconst o = { ...a };\n      ", nil, "replaceMapSpreadInObject", "\ndeclare const a: Map<boolean, string> & Set<number>;\nconst o = Object.fromEntries(a);\n      "},
		{"upstream invalid 43", "file.ts", "\nconst promise = new Promise(() => {});\nconst o = { ...promise };\n      ", nil, "addAwait", "\nconst promise = new Promise(() => {});\nconst o = { ...await promise };\n      "},
		{"upstream invalid 44", "file.ts", "\ndeclare const promise: Promise<{ a: 1 }>;\nasync function foo() {\n  return { ...(promise || {}) };\n}\n      ", nil, "addAwait", "\ndeclare const promise: Promise<{ a: 1 }>;\nasync function foo() {\n  return { ...(await (promise || {})) };\n}\n      "},
		{"upstream invalid 45", "file.ts", "\ndeclare const promise: Promise<any>;\nasync function foo() {\n  return { ...(Math.random() < 0.5 ? promise : {}) };\n}\n      ", nil, "addAwait", "\ndeclare const promise: Promise<any>;\nasync function foo() {\n  return { ...(await (Math.random() < 0.5 ? promise : {})) };\n}\n      "},
		{"upstream invalid 46", "file.ts", "\nfunction withPromise<P extends Promise<void>>(promise: P) {\n  return { ...promise };\n}\n      ", nil, "addAwait", "\nfunction withPromise<P extends Promise<void>>(promise: P) {\n  return { ...await promise };\n}\n      "},
		{"upstream invalid 47", "file.ts", "\ndeclare const maybePromise: Promise<number> | { a: number };\nconst o = { ...maybePromise };\n      ", nil, "addAwait", "\ndeclare const maybePromise: Promise<number> | { a: number };\nconst o = { ...await maybePromise };\n      "},
		{"upstream invalid 48", "file.ts", "\ndeclare const promise: Promise<number> & { a: number };\nconst o = { ...promise };\n      ", nil, "addAwait", "\ndeclare const promise: Promise<number> & { a: number };\nconst o = { ...await promise };\n      "},
		{"upstream invalid 49", "file.ts", "\ndeclare function getPromise(): Promise<number>;\nconst o = { ...getPromise() };\n      ", nil, "addAwait", "\ndeclare function getPromise(): Promise<number>;\nconst o = { ...await getPromise() };\n      "},
		{"upstream invalid 50", "file.ts", "\ndeclare function getPromise<T extends Promise<number>>(arg: T): T;\nconst o = { ...getPromise() };\n      ", nil, "addAwait", "\ndeclare function getPromise<T extends Promise<number>>(arg: T): T;\nconst o = { ...await getPromise() };\n      "},
		{"upstream invalid 89", "file.tsx", "\ndeclare const map: Map<string, number>;\n\nconst o = <div {...map} />;\n      ", nil, "replaceMapSpreadInObject", "\ndeclare const map: Map<string, number>;\n\nconst o = <div {...Object.fromEntries(map)} />;\n      "},
		{"upstream invalid 90", "file.tsx", "\nconst promise = new Promise(() => {});\n\nconst o = <div {...promise} />;\n      ", nil, "addAwait", "\nconst promise = new Promise(() => {});\n\nconst o = <div {...await promise} />;\n      "}}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMisusedSpread, testCase.fileName, testCase.sourceText, testCase.options)
			if len(result.Diagnostics) == 0 {
				t.Fatal("want at least one finding, got none")
			}
			// The harness trims the fixture before building the program, so the offsets a fix
			// carries are against the trimmed text rather than against the literal written above.
			source := strings.TrimSpace(testCase.sourceText) + "\n"
			want := strings.TrimSpace(testCase.wantOutput) + "\n"

			matchedTheId := false
			for _, diagnostic := range result.Diagnostics {
				for _, candidate := range diagnostic.Suggestions {
					if candidate.Message.Id != testCase.wantId {
						continue
					}
					matchedTheId = true
					if applySuggestion(t, source, candidate) == want {
						return
					}
				}
			}
			if !matchedTheId {
				t.Fatalf("no suggestion with id %q among the findings", testCase.wantId)
			}
			t.Errorf("no suggestion with id %q produced\n%q", testCase.wantId, want)
		})
	}
}

// TestNoMisusedSpreadSpans asserts WHERE each finding points, which no id assertion can see.
//
// The span is the whole spread, including the three dots, rather than the argument. The JSX rows are
// the sharper half: there the span includes the surrounding BRACES, so `{...A}` and not `...A` and
// not `A`, which pins that a JSX spread attribute is reported as its own node rather than being
// routed through the object-literal shape it otherwise behaves like.
//
// One row per message id, plus every JSX shape, all taken from upstream's recorded columns.
func TestNoMisusedSpreadSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
		wantSpans  []string
	}{
		{"upstream invalid 14", "file.ts", "const o = { ...[1, 2, 3] };", []string{"...[1, 2, 3]"}},
		{"upstream invalid 23", "file.ts", "const o = { ...new Set([1, 2, 3]) };", []string{"...new Set([1, 2, 3])"}},
		{"upstream invalid 26", "file.ts", "\ndeclare const set: WeakSet<object>;\nconst o = { ...set };\n      ", []string{"...set"}},
		{"upstream invalid 30", "file.ts", "\nconst o = {\n  ...new Map([\n    ['test-1', 1],\n    ['test-2', 2],\n  ]),\n};\n      ", []string{"...new Map([\n    ['test-1', 1],\n    ['test-2', 2],\n  ])"}},
		{"upstream invalid 43", "file.ts", "\nconst promise = new Promise(() => {});\nconst o = { ...promise };\n      ", []string{"...promise"}},
		{"upstream invalid 51", "file.ts", "\nfunction f() {}\n\nconst o = { ...f };\n      ", []string{"...f"}},
		{"upstream invalid 77", "file.ts", "\nclass A {}\n\nconst o = { ...A };\n      ", []string{"...A"}},
		{"upstream invalid 84", "file.tsx", "\nconst o = <div {...[1, 2, 3]} />;\n      ", []string{"{...[1, 2, 3]}"}},
		{"upstream invalid 85", "file.tsx", "\nclass A {}\n\nconst o = <div {...A} />;\n      ", []string{"{...A}"}},
		{"upstream invalid 86", "file.tsx", "\nconst o = <div {...new Date()} />;\n      ", []string{"{...new Date()}"}},
		{"upstream invalid 87", "file.tsx", "\nfunction f() {}\n\nconst o = <div {...f} />;\n      ", []string{"{...f}"}},
		{"upstream invalid 88", "file.tsx", "\nconst o = <div {...new Set([1, 2, 3])} />;\n      ", []string{"{...new Set([1, 2, 3])}"}},
		{"upstream invalid 89", "file.tsx", "\ndeclare const map: Map<string, number>;\n\nconst o = <div {...map} />;\n      ", []string{"{...map}"}},
		{"upstream invalid 90", "file.tsx", "\nconst promise = new Promise(() => {});\n\nconst o = <div {...promise} />;\n      ", []string{"{...promise}"}}}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMisusedSpread, testCase.fileName, testCase.sourceText, nil)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("want %d findings, got %d", len(testCase.wantSpans), len(result.Diagnostics))
			}
			// The harness trims the fixture, so slice the text it actually wrote.
			source := strings.TrimSpace(testCase.sourceText) + "\n"
			for index, want := range testCase.wantSpans {
				reported := source[result.Diagnostics[index].Range.Pos():result.Diagnostics[index].Range.End()]
				if reported != want {
					t.Errorf("finding %d points at %q, want %q", index, reported, want)
				}
			}
		})
	}
}

// TestNoMisusedSpreadCaseThatCannotBeExpressedHere records the one upstream invalid case absent from
// the imported list above, and why its absence is a fact about the HARNESS rather than about the
// rule.
//
// Upstream's invalid case 61 spreads a type imported from an ambient `declare module 'module'`, and
// its allow list uses a `from: package` specifier naming that module. Reproducing it needs `module`
// to be a resolvable package name, which means `@types/node` on the program.
//
// `rule_testing` pins `types: []` deliberately, so that a fixture cannot pick up whatever happens to be
// installed near the temp directory and pass or fail by machine. With no node types the input is
// TypeScript error 2664, "Invalid module name in augmentation", and the rule is silent.
//
// This was measured rather than assumed, and the FIRST hypothesis was wrong, which is why the
// evidence is recorded rather than the conclusion. The obvious suspect was `moduleDetection: "force"`,
// which a sibling rule hit on the same shape. Rebuilding the oracle without it changed nothing:
// still silent. Printing the FULL diagnostic list rather than grepping for the rule's own findings
// showed error 2664 and 2591, and adding `types: ["node"]` made the same input report
// `noIterableSpreadInObject` exactly as upstream records.
//
// So the case is left out of the imported list rather than being weakened into passing, and the
// half of it that CAN be expressed is pinned below: the same iterable shape without the ambient
// module reports, which is the control saying the arm works and only the module resolution is
// missing.
func TestNoMisusedSpreadCaseThatCannotBeExpressedHere(t *testing.T) {
	t.Parallel()

	t.Run("the same iterable shape without an ambient module reports", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTypedWithOptions(t, NoMisusedSpread, noMisusedSpreadFile,
			"type CustomIterable = {\n  [Symbol.iterator]: () => string;\n};\ndeclare const iterator: CustomIterable;\nconst a = { ...iterator };", nil)
		rule_testing.ExpectFindings(t, result, "noIterableSpreadInObject")
	})

	t.Run("an allow specifier naming that type silences it", func(t *testing.T) {
		t.Parallel()
		// The other half of what case 61 was testing, expressed with a `from: file` specifier, which
		// needs no package resolution.
		result := rule_testing.RunTypedWithOptions(t, NoMisusedSpread, noMisusedSpreadFile,
			"type CustomIterable = {\n  [Symbol.iterator]: () => string;\n};\ndeclare const iterator: CustomIterable;\nconst a = { ...iterator };",
			NoMisusedSpreadOptions{Allow: []type_checking.TypeOrValueSpecifier{
				{From: type_checking.TypeOrValueSpecifierFromFile, Name: []string{"CustomIterable"}},
			}})
		rule_testing.ExpectClean(t, result)
	})
}

// TestNoMisusedSpreadCascadeOrdering pins the order of the object cascade, which the imported corpus
// exercises but never isolates.
//
// Several of these types satisfy more than one arm at once, and the message they get is decided by
// which arm is asked first. A Map is iterable AND a class instance; an array is iterable; a Set is
// iterable and a class instance. Reversing any adjacent pair changes a real input's message while
// leaving the finding count alone, which every id fixture above would still accept for the wrong
// reason if it were checking only that something reported.
func TestNoMisusedSpreadCascadeOrdering(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantId     string
	}{
		{
			"a Map is iterable and a class instance, and Map wins",
			"declare const map: Map<string, number>;\nconst o = { ...map };",
			"noMapSpreadInObject",
		},
		{
			"an array is iterable, and array wins",
			"declare const arr: number[];\nconst o = { ...arr };",
			"noArraySpreadInObject",
		},
		{
			"a Set is iterable and a class instance, and iterable wins",
			"declare const set: Set<number>;\nconst o = { ...set };",
			"noIterableSpreadInObject",
		},
		{
			"a promise is a class instance, and promise wins",
			"declare const promise: Promise<number>;\nconst o = { ...promise };",
			"noPromiseSpreadInObject",
		},
		{
			"a non-iterable class instance falls through to the instance arm",
			"class A {\n  z = 1;\n  m() {}\n}\ndeclare const a: A;\nconst o = { ...a };",
			"noClassInstanceSpreadInObject",
		},
		{
			"the class itself is a declaration rather than an instance",
			"class A {\n  static z = 1;\n}\nconst o = { ...A };",
			"noClassDeclarationSpreadInObject",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMisusedSpread, noMisusedSpreadFile, testCase.sourceText, nil)
			rule_testing.ExpectFindings(t, result, testCase.wantId)
		})
	}
}

// TestNoMisusedSpreadRequiresTheTypedHarness asserts the rule declares the checker and that the
// plain harness cannot prove it, so a later revert to rule_testing.Run fails loudly.
func TestNoMisusedSpreadRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoMisusedSpread.NeedsTypeChecker {
		t.Fatal("the rule stopped declaring NeedsTypeChecker, so every typed fixture would run against a nil checker")
	}

	const source = "const o = { ...[1, 2, 3] };"

	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoMisusedSpread, noMisusedSpreadFile, source), "noArraySpreadInObject")
	rule_testing.ExpectClean(t, rule_testing.Run(t, NoMisusedSpread, noMisusedSpreadFile, source))
}

// TestNoMisusedSpreadSurvivesMalformedSpreads runs the rule over spread shapes where a node it
// reaches for is absent or unusual.
//
// A panic costs every rule its verdict on the whole file. This rule reaches for a spread's
// expression, its parent, and an object literal's property list, and reads the source text between
// two computed offsets when building the Map suggestion, so a malformed parse has several ways in.
func TestNoMisusedSpreadSurvivesMalformedSpreads(t *testing.T) {
	t.Parallel()

	sources := []string{
		"const a = [...];\n",
		"const a = { ... };\n",
		"const a = f(...);\n",
		"const a = [... ,1];\n",
		"const a = { ...,  b: 1 };\n",
		"const a = new Set([...'x']);\n",
		"declare const m: Map<string, number>;\nconst a = { ...(m) };\n",
		"declare const m: Map<string, number>;\nconst a = { ...((m)) };\n",
		"declare const m: Map<string, number>;\nconst a = { ...(m, m) };\n",
		"declare const m: Map<string, number>;\nconst a = { ...m, ...m };\n",
		"declare const p: Promise<void>;\nconst a = { ...(p) };\n",
		"declare const p: Promise<void>;\nconst a = { ...(p as Promise<void>) };\n",
		"const a = { ...undefined };\n",
		"const a = { ...null };\n",
		"declare const x: never;\nconst a = { ...x };\n",
		"declare const x: any;\nconst a = { ...x };\n",
		"declare const x: unknown;\nconst a = { ...x };\n",
		"function f<T>(t: T) {\n  return { ...t };\n}\n",
	}

	for index, source := range sources {
		t.Run(fmt.Sprintf("shape-%d", index), func(t *testing.T) {
			t.Parallel()
			// A panic fails the test. Findings are deliberately unasserted.
			rule_testing.RunTypedWithOptions(t, NoMisusedSpread, noMisusedSpreadFile, source, nil)
		})
	}
}

// TestNoMisusedSpreadDecoder routes configuration through the rule's own exported decoder.
//
// Unlike restrict-plus-operands there is no default to invert here, since an empty allow list is
// both the default and the zero value. What the decoder still owns is the wire format: `allow` mixes
// bare strings with specifier objects, and `name` is itself either a string or an array of them, so
// a fixture built from the struct directly leaves all of that untested.
func TestNoMisusedSpreadDecoder(t *testing.T) {
	t.Parallel()

	decode := func(t *testing.T, raw string) NoMisusedSpreadOptions {
		t.Helper()
		decoded, err := DecodeNoMisusedSpreadOptions([]byte(raw))
		if err != nil {
			t.Fatalf("decoding %s: %v", raw, err)
		}
		options, _ := decoded.(NoMisusedSpreadOptions)
		return options
	}

	t.Run("an empty object allows nothing", func(t *testing.T) {
		t.Parallel()
		got := decode(t, `{}`)
		if len(got.Allow) != 0 || len(got.AllowInline) != 0 {
			t.Errorf("got %+v, want an empty allow list", got)
		}
	})

	t.Run("a bare string becomes an inline specifier", func(t *testing.T) {
		t.Parallel()
		got := decode(t, `{"allow": ["Map"]}`)
		if len(got.AllowInline) != 1 || got.AllowInline[0] != "Map" {
			t.Errorf("got %+v, want one inline specifier naming Map", got)
		}
	})

	t.Run("a specifier object with a string name becomes a one-element name list", func(t *testing.T) {
		t.Parallel()
		got := decode(t, `{"allow": [{"from": "lib", "name": "Map"}]}`)
		if len(got.Allow) != 1 || len(got.Allow[0].Name) != 1 || got.Allow[0].Name[0] != "Map" {
			t.Fatalf("got %+v, want one specifier naming Map", got)
		}
		if got.Allow[0].From != type_checking.TypeOrValueSpecifierFromLib {
			t.Errorf("the from field decoded as %v, want lib", got.Allow[0].From)
		}
	})

	t.Run("the decoded allow list reaches the rule and changes its verdict", func(t *testing.T) {
		t.Parallel()
		// Byte-identical source, opposite verdicts, separated only by what came off the wire. The
		// subject is upstream valid 38's class instance, which an inline specifier names by type, with
		// a method added: a fields-only class is exempt without any allow list, so it could not show
		// the list changing anything (see TestNoMisusedSpreadExemptsAClassWhoseCopyIsComplete).
		const source = "class A {\n  a = 1;\n  m() {}\n}\nconst a = new A();\nconst o = { ...a };"

		rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, NoMisusedSpread, noMisusedSpreadFile, source,
			decode(t, `{}`)), "noClassInstanceSpreadInObject")
		rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoMisusedSpread, noMisusedSpreadFile, source,
			decode(t, `{"allow": ["A"]}`)))
	})
}

// TestNoMisusedSpreadMapSuggestionKeepsExistingParentheses pins a repair the imported corpus cannot
// see, and which this port originally got wrong.
//
// When the spread has SIBLING properties the Map suggestion rewrites the argument in place, and
// upstream replaces the node its own tree hands it, which is the unparenthesized expression. So
// parentheses already in the source survive: `{ ...(map), other: 1 }` becomes
// `{ ...(Object.fromEntries(map)), other: 1 }`, keeping a pair that is now redundant.
//
// This port first replaced the whole argument including those parentheses, producing the tidier
// `{ ...Object.fromEntries(map), other: 1 }`. All 151 imported rows stayed green over that, because
// the corpus writes a parenthesized argument ONLY in the sole-property shape, where the whole object
// literal is replaced and this branch never runs. Found by mutation, then settled against the
// installed rule rather than by reading, which is what showed the tidier answer was the wrong one.
//
// The last row is the control that keeps the two decisions apart: a comma expression is weak
// precedence, so it gets parentheses ADDED by this rule regardless, and it would still pass if the
// keep-existing-parentheses behaviour were lost.
func TestNoMisusedSpreadMapSuggestionKeepsExistingParentheses(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantOutput string
	}{
		{
			"a parenthesized argument before a sibling keeps its parentheses",
			"declare const m: Map<string, number>;\nconst o = { ...(m), other: 1 };",
			"declare const m: Map<string, number>;\nconst o = { ...(Object.fromEntries(m)), other: 1 };",
		},
		{
			"a parenthesized argument after a sibling keeps its parentheses",
			"declare const m: Map<string, number>;\nconst o = { other: 1, ...(m) };",
			"declare const m: Map<string, number>;\nconst o = { other: 1, ...(Object.fromEntries(m)) };",
		},
		{
			"both pairs survive a doubly parenthesized argument",
			"declare const m: Map<string, number>;\nconst o = { ...((m)), other: 1 };",
			"declare const m: Map<string, number>;\nconst o = { ...((Object.fromEntries(m))), other: 1 };",
		},
		{
			"a weak-precedence argument gets parentheses added as well as kept",
			"declare const m: Map<string, number>;\nconst o = { ...(m, m), other: 1 };",
			"declare const m: Map<string, number>;\nconst o = { ...(Object.fromEntries((m, m))), other: 1 };",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMisusedSpread, noMisusedSpreadFile, testCase.sourceText, nil)
			rule_testing.ExpectFindings(t, result, "noMapSpreadInObject")

			source := strings.TrimSpace(testCase.sourceText) + "\n"
			want := strings.TrimSpace(testCase.wantOutput) + "\n"
			if got := applySuggestion(t, source, result.Diagnostics[0].Suggestions[0]); got != want {
				t.Errorf("the repair produced\n%q\nwant\n%q", got, want)
			}
		})
	}
}

// TestNoMisusedSpreadMergedClassDeclarations pins that a class merged with an interface is still
// recognised as a class instance in either source ordering.
//
// Upstream reports both, measured against the installed rule. The rows matter less for the loop over
// declarations, which is inert here for reasons recorded at the line, than for the guard above it:
// a merged symbol is exactly the shape where a rule asking the wrong declaration goes silent, and
// these pin that it does not.
func TestNoMisusedSpreadMergedClassDeclarations(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{
			"an interface declared before the class it merges with",
			"interface A {\n  z: number;\n}\nclass A {\n  z = 1;\n}\ndeclare const a: A;\nconst o = { ...a };",
		},
		{
			"an interface declared after the class it merges with",
			"class A {\n  z = 1;\n}\ninterface A {\n  y: number;\n}\ndeclare const a: A;\nconst o = { ...a };",
		},
		{
			"two interfaces before the class",
			"interface B {\n  z: number;\n}\ninterface B {\n  y: number;\n}\nclass B {\n  z = 1;\n  y = 2;\n}\ndeclare const b: B;\nconst o = { ...b };",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMisusedSpread, noMisusedSpreadFile, testCase.sourceText, nil)
			rule_testing.ExpectFindings(t, result, "noClassInstanceSpreadInObject")
		})
	}
}

// TestNoMisusedSpreadExemptsAClassWhoseCopyIsComplete pins cohere's second deliberate divergence from
// upstream: a spread of a class instance is silent when the copy loses nothing, because the class and
// every base declare no method, accessor or #private member, are written in our own source, and carry
// only decorators our own source declares. @system_cohere's ruling, 2026-10-03, from api's decorated
// GraphQL and Serializable data classes (#ynneze5). See noMisusedSpreadCopiesCompletely.
//
// The first group is upstream's invalid 69 through 76 and 83 verbatim, every one a fields-only class,
// which upstream reports and cohere does not. The rest are the clauses, each with the case it exempts
// and the case it must still report, so a mutant dropping any one clause fails here.
func TestNoMisusedSpreadExemptsAClassWhoseCopyIsComplete(t *testing.T) {
	t.Parallel()

	silent := []struct {
		name       string
		sourceText string
	}{
		{"upstream invalid 69", "\nclass A {\n  a = 1;\n  public b = 2;\n  private c = 3;\n  protected d = 4;\n  static e = 5;\n}\n\nconst o = { ...new A() };\n      "},
		{"upstream invalid 70", "\nclass A {\n  a = 1;\n}\n\nconst a = new A();\n\nconst o = { ...a };\n      "},
		{"upstream invalid 71", "\nclass A {\n  a = 1;\n}\n\ndeclare const a: A;\n\nconst o = { ...a };\n      "},
		{"upstream invalid 72", "\nclass A {\n  a = 1;\n}\n\ndeclare function getA(): A;\n\nconst o = { ...getA() };\n      "},
		{"upstream invalid 73", "\nclass A {\n  a = 1;\n}\n\ndeclare function getA<T extends A>(arg: T): T;\n\nconst o = { ...getA() };\n      "},
		{"upstream invalid 74", "\nclass A {\n  a = 1;\n}\n\nclass B extends A {}\n\nconst o = { ...new B() };\n      "},
		{"upstream invalid 75", "\nclass A {\n  a = 1;\n}\n\ndeclare const a: A | { b: string };\n\nconst o = { ...a };\n      "},
		{"upstream invalid 76", "\nclass A {\n  a = 1;\n}\n\ndeclare const a: A & { b: string };\n\nconst o = { ...a };\n      "},
		{"upstream invalid 83", "\n        const a = { ...new (class A { static value = 1; })() };\n      "},
		{"a generic abstract base, as Base's PaginationResult", "abstract class Page<T> {\n  abstract items: T[];\n  total = 0;\n}\nclass Users extends Page<string> {\n  items: string[] = [];\n}\ndeclare const page: Users;\nconst o = { ...page, more: true };"},
		{"a static method lives on the constructor, which an instance never carried", "class A {\n  a = 1;\n  static make(): A { return new A(); }\n}\nconst o = { ...new A() };"},
		{"constructor parameter properties are own properties", "class A {\n  constructor(public a: number, readonly b: string) {}\n}\nconst o = { ...new A(1, 'b') };"},
		{"decorators our own source declares, on the class, a field and a parameter", "function Entity(): ClassDecorator { return () => {}; }\nfunction Field(): PropertyDecorator { return () => {}; }\nfunction Inject(): ParameterDecorator { return () => {}; }\n@Entity()\nclass A {\n  @Field() a = 1;\n  constructor(@Inject() public b: number) {}\n}\ndeclare const a: A;\nconst o = { ...a };"},
		{"every class in a union is exempt", "class A {\n  a = 1;\n}\nclass B {\n  b = 2;\n}\ndeclare const value: A | B;\nconst o = { ...value };"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoMisusedSpread, noMisusedSpreadFile, testCase.sourceText, nil))
		})
	}

	reported := []struct {
		name       string
		sourceText string
		wantId     string
	}{
		{"a method", "class A {\n  a = 1;\n  m() {}\n}\nconst o = { ...new A() };", "noClassInstanceSpreadInObject"},
		{"a getter", "class A {\n  a = 1;\n  get double() { return this.a * 2; }\n}\nconst o = { ...new A() };", "noClassInstanceSpreadInObject"},
		{"a setter", "class A {\n  a = 1;\n  set value(next: number) { this.a = next; }\n}\nconst o = { ...new A() };", "noClassInstanceSpreadInObject"},
		{"an auto-accessor", "class A {\n  accessor a = 1;\n}\nconst o = { ...new A() };", "noClassInstanceSpreadInObject"},
		{"a #private field, which a spread does not copy", "class A {\n  a = 1;\n  #secret = 2;\n}\nconst o = { ...new A() };", "noClassInstanceSpreadInObject"},
		{"a method on a base class", "class Base {\n  m() {}\n}\nclass A extends Base {\n  a = 1;\n}\nconst o = { ...new A() };", "noClassInstanceSpreadInObject"},
		{"a base from a mixin, which has no class declaration to read", "type Constructor = new (...args: any[]) => object;\nfunction Tagged<T extends Constructor>(Base: T) { return class extends Base { tag = 'x'; }; }\nclass Plain {}\nclass A extends Tagged(Plain) {\n  a = 1;\n}\nconst o = { ...new A() };", "noClassInstanceSpreadInObject"},
		{"an ambient class, whose prototype is not ours to read (upstream invalid 67)", "declare class A {\n  a: number;\n}\ndeclare const a: A;\nconst o = { ...a };", "noClassInstanceSpreadInObject"},
		{"an interface merged into the class, which may describe methods a mixin installs", "interface A {\n  m(): void;\n}\nclass A {\n  a = 1;\n}\ndeclare const a: A;\nconst o = { ...a };", "noClassInstanceSpreadInObject"},
		{"an ambient decorator, with no implementation in our source", "declare function observable(target: object, key: string): void;\nclass A {\n  @observable a = 1;\n}\nconst o = { ...new A() };", "noClassInstanceSpreadInObject"},
		{"an ambient decorator on the class alone", "declare function sealed(target: Function): void;\n@sealed\nclass A {\n  a = 1;\n}\nconst o = { ...new A() };", "noClassInstanceSpreadInObject"},
		{"an ambient decorator on a constructor parameter alone", "declare function inject(target: object, key: string | symbol | undefined, index: number): void;\nclass A {\n  constructor(@inject public a: number) {}\n}\nconst o = { ...new A(1) };", "noClassInstanceSpreadInObject"},
		{"one class in a union that is not exempt", "class A {\n  a = 1;\n}\nclass B {\n  b = 2;\n  m() {}\n}\ndeclare const value: A | B;\nconst o = { ...value };", "noClassInstanceSpreadInObject"},
		{"the class itself still reports as a declaration", "class A {\n  static a = 1;\n}\nconst o = { ...A };", "noClassDeclarationSpreadInObject"},
	}
	for _, testCase := range reported {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoMisusedSpread, noMisusedSpreadFile, testCase.sourceText, nil)
			rule_testing.ExpectFindings(t, result, testCase.wantId)
		})
	}

	// Decorators from other files: the package and declaration-file cases a single-file harness cannot
	// write, against an own-source decorator in a second file, which is exactly api's shape (Base's
	// GraphQlField imported into an entity).
	subject := "import { Field } from './Field';\nclass A {\n  @Field() a = 1;\n}\nconst o = { ...new A() };"
	files := []struct {
		name      string
		decorator map[string]string
		findings  int
	}{
		{"a decorator from our own source in another file", map[string]string{
			"Field.ts": "export function Field(): PropertyDecorator { return () => {}; }\n",
		}, 0},
		{"a decorator from a declaration file", map[string]string{
			"Field.d.ts": "export declare function Field(): PropertyDecorator;\n",
		}, 1},
		{"a decorator from a package", map[string]string{
			"Field.ts":                   "export { Field } from './node_modules/mobx/index';\n",
			"node_modules/mobx/index.ts": "export function Field(): PropertyDecorator { return () => {}; }\n",
		}, 1},
	}
	for _, testCase := range files {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			sources := map[string]string{"Subject.ts": subject}
			for name, text := range testCase.decorator {
				sources[name] = text
			}
			result := rule_testing.RunTypedFiles(t, NoMisusedSpread, sources, "Subject.ts")
			if len(result.Diagnostics) != testCase.findings {
				t.Errorf("got %d findings, want %d: %+v", len(result.Diagnostics), testCase.findings, result.Diagnostics)
			}
		})
	}
}
