package typescript

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const preferPromiseRejectErrorsFile = "/repository/source/Rejections.ts"

// The tables below are typescript-eslint's own corpus for this rule, extracted from
// tests/rules/prefer-promise-reject-errors.test.ts by parsing it with the TypeScript compiler
// API. Nothing was retyped, and every string was byte-verified against the literals in that
// file before this table was written.
//
// Cases carrying options run through DecodePreferPromiseRejectErrorsOptions rather than
// building the options struct directly, so the decoder is under test too. Cases importing the
// ambient 'errors' module need a second file in the program and run through RunTypedFiles.

func TestPreferPromiseRejectErrorsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"Promise.resolve(5);",
		"Promise.reject(new Error());",
		"Promise.reject(new TypeError());",
		"Promise.reject(new Error('foo'));",
		"\nclass CustomError extends Error {}\nPromise.reject(new CustomError());\n    ",
		"\ndeclare const foo: () => { err: SyntaxError };\nPromise.reject(foo().err);\n    ",
		"\ndeclare const foo: () => Promise<Error>;\nPromise.reject(await foo());\n    ",
		"Promise.reject((foo = new Error()));",
		"\nconst foo = Promise;\nfoo.reject(new Error());\n    ",
		"Promise['reject'](new Error());",
		"Promise.reject(true && new Error());",
		"\nconst foo = false;\nPromise.reject(false || new Error());\n    ",
		"\ndeclare const foo: Readonly<Error>;\nPromise.reject(foo);\n    ",
		"\ndeclare const foo: Readonly<Error> | Readonly<TypeError>;\nPromise.reject(foo);\n    ",
		"\ndeclare const foo: Readonly<Error> & Readonly<TypeError>;\nPromise.reject(foo);\n    ",
		"\ndeclare const foo: Readonly<Error> & { foo: 'bar' };\nPromise.reject(foo);\n    ",
		"\ndeclare const foo: Readonly<Error & { bar: 'foo' }> & { foo: 'bar' };\nPromise.reject(foo);\n    ",
		"\ndeclare const foo: Readonly<Readonly<Error>>;\nPromise.reject(foo);\n    ",
		"\ndeclare const foo: Readonly<Readonly<Readonly<Error>>>;\nPromise.reject(foo);\n    ",
		"\ndeclare const foo: Readonly<\n  Readonly<Readonly<Error & { bar: 'foo' }> & { foo: 'bar' }> & {\n    fooBar: 'barFoo';\n  }\n> & { barFoo: 'fooBar' };\nPromise.reject(foo);\n    ",
		"\ndeclare const foo:\n  Readonly<Readonly<Error> | Readonly<TypeError & string>> | Readonly<Error>;\nPromise.reject(foo);\n    ",
		"\ntype Wrapper<T> = { foo: Readonly<T>[] };\ndeclare const foo: Wrapper<Error>['foo'][5];\nPromise.reject(foo);\n    ",
		"\ndeclare const foo: Error[];\nPromise.reject(foo[5]);\n    ",
		"\ndeclare const foo: ReadonlyArray<Error>;\nPromise.reject(foo[5]);\n    ",
		"\ndeclare const foo: [Error];\nPromise.reject(foo[0]);\n    ",
		"\nnew Promise(function (resolve, reject) {\n  resolve(5);\n});\n    ",
		"\nnew Promise(function (resolve, reject) {\n  reject(new Error());\n});\n    ",
		"\nnew Promise((resolve, reject) => {\n  reject(new Error());\n});\n    ",
		"new Promise((resolve, reject) => reject(new Error()));",
		"new Promise((yes, no) => no(new Error()));",
		"new Promise();",
		"new Promise(5);",
		"new Promise((resolve, { apply }) => {});",
		"new Promise((resolve, reject) => {});",
		"new Promise((resolve, reject) => reject);",
		"\nclass CustomError extends Error {}\nnew Promise(function (resolve, reject) {\n  reject(new CustomError());\n});\n    ",
		"\ndeclare const foo: () => { err: SyntaxError };\nnew Promise(function (resolve, reject) {\n  reject(foo().err);\n});\n    ",
		"new Promise((resolve, reject) => reject((foo = new Error())));",
		"\nnew Foo((resolve, reject) => reject(5));\n    ",
		"\nclass Foo {\n  constructor(\n    executor: (resolve: () => void, reject: (reason?: any) => void) => void,\n  ): Promise<any> {}\n}\nnew Foo((resolve, reject) => reject(5));\n    ",
		"\nnew Promise((resolve, reject) => {\n  return function (reject) {\n    reject(5);\n  };\n});\n    ",
		"new Promise((resolve, reject) => resolve(5, reject));",
		"\nclass C {\n  #error: Error;\n  foo() {\n    Promise.reject(this.#error);\n  }\n}\n    ",
		"\nconst foo = Promise;\nnew foo((resolve, reject) => reject(new Error()));\n    ",
		"\ndeclare const foo: Readonly<Error>;\nnew Promise((resolve, reject) => reject(foo));\n    ",
		"\ndeclare const foo: Readonly<Error> | Readonly<TypeError>;\nnew Promise((resolve, reject) => reject(foo));\n    ",
		"\ndeclare const foo: Readonly<Error> & Readonly<TypeError>;\nnew Promise((resolve, reject) => reject(foo));\n    ",
		"\ndeclare const foo: Readonly<Error> & { foo: 'bar' };\nnew Promise((resolve, reject) => reject(foo));\n    ",
		"\ndeclare const foo: Readonly<Error & { bar: 'foo' }> & { foo: 'bar' };\nnew Promise((resolve, reject) => reject(foo));\n    ",
		"\ndeclare const foo: Readonly<Readonly<Error>>;\nnew Promise((resolve, reject) => reject(foo));\n    ",
		"\ndeclare const foo: Readonly<Readonly<Readonly<Error>>>;\nnew Promise((resolve, reject) => reject(foo));\n    ",
		"\ndeclare const foo: Readonly<\n  Readonly<Readonly<Error & { bar: 'foo' }> & { foo: 'bar' }> & {\n    fooBar: 'barFoo';\n  }\n> & { barFoo: 'fooBar' };\nnew Promise((resolve, reject) => reject(foo));\n    ",
		"\ndeclare const foo:\n  Readonly<Readonly<Error> | Readonly<TypeError & string>> | Readonly<Error>;\nnew Promise((resolve, reject) => reject(foo));\n    ",
		"\ntype Wrapper<T> = { foo: Readonly<T>[] };\ndeclare const foo: Wrapper<Error>['foo'][5];\nnew Promise((resolve, reject) => reject(foo));\n    ",
		"\ndeclare const foo: Error[];\nnew Promise((resolve, reject) => reject(foo[5]));\n    ",
		"\ndeclare const foo: ReadonlyArray<Error>;\nnew Promise((resolve, reject) => reject(foo[5]));\n    ",
		"\ndeclare const foo: [Error];\nnew Promise((resolve, reject) => reject(foo[0]));\n    ",
		"\nclass Foo extends Promise<number> {}\nFoo.reject(new Error());\n    ",
		"\nclass Foo extends Promise<number> {}\nnew Foo((resolve, reject) => reject(new Error()));\n    ",
		"\ndeclare const someRandomCall: {\n  reject(arg: any): void;\n};\nsomeRandomCall.reject(5);\n    ",
		"\ndeclare const foo: PromiseConstructor;\nfoo.reject(new Error());\n    ",
		"console[Symbol.iterator]();",
		"\nclass A {\n  a = [];\n  [Symbol.iterator]() {\n    return this.a[Symbol.iterator]();\n  }\n}\n    ",
		"\ndeclare const foo: PromiseConstructor;\nfunction fun<T extends Error>(t: T): void {\n  foo.reject(t);\n}\n    ",
	}
	for index, sourceText := range cases {
		t.Run(fmt.Sprintf("upstream valid %d", index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, PreferPromiseRejectErrors, preferPromiseRejectErrorsFile, sourceText))
		})
	}
}

func TestPreferPromiseRejectErrorsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		findings   int
	}{
		{"Promise.reject(5);", 1},
		{"Promise.reject('foo');", 1},
		{"Promise.reject(`foo`);", 1},
		{"Promise.reject('foo', somethingElse);", 1},
		{"Promise.reject(false);", 1},
		{"Promise.reject(void `foo`);", 1},
		{"Promise.reject();", 1},
		{"Promise.reject(undefined);", 1},
		{"Promise.reject(null);", 1},
		{"Promise.reject({ foo: 1 });", 1},
		{"Promise.reject([1, 2, 3]);", 1},
		{"\ndeclare const foo: Error | undefined;\nPromise.reject(foo);\n      ", 1},
		{"\ndeclare const foo: () => Promise<string>;\nPromise.reject(await foo());\n      ", 1},
		{"\ndeclare const foo: boolean;\nPromise.reject(foo && new Error());\n      ", 1},
		{"\nconst foo = Promise;\nfoo.reject();\n      ", 1},
		{"Promise.reject?.(5);", 1},
		{"Promise?.reject(5);", 1},
		{"Promise?.reject?.(5);", 1},
		{"(Promise?.reject)(5);", 1},
		{"(Promise?.reject)?.(5);", 1},
		{"Promise['reject'](5);", 1},
		{"Promise.reject((foo += new Error()));", 1},
		{"Promise.reject((foo -= new Error()));", 1},
		{"Promise.reject((foo **= new Error()));", 1},
		{"Promise.reject((foo <<= new Error()));", 1},
		{"Promise.reject((foo |= new Error()));", 1},
		{"Promise.reject((foo &= new Error()));", 1},
		{"\ndeclare const foo: never;\nPromise.reject(foo);\n      ", 1},
		{"\ndeclare const foo: unknown;\nPromise.reject(foo);\n      ", 1},
		{"\ntype FakeReadonly<T> = { 'fake readonly': T };\ndeclare const foo: FakeReadonly<Error>;\nPromise.reject(foo);\n      ", 1},
		{"\ndeclare const foo: Readonly<'error'>;\nPromise.reject(foo);\n      ", 1},
		{"\ndeclare const foo: Readonly<Error | 'error'>;\nPromise.reject(foo);\n      ", 1},
		{"\ndeclare const foo: Readonly<Error> | 'error';\nPromise.reject(foo);\n      ", 1},
		{"\ndeclare const foo: Readonly<Error> | Readonly<TypeError> | Readonly<'error'>;\nPromise.reject(foo);\n      ", 1},
		{"\ndeclare const foo: Readonly<Readonly<'error'>>;\nPromise.reject(foo);\n      ", 1},
		{"\ndeclare const foo: Readonly<Readonly<Readonly<Error> | 'error'>>;\nPromise.reject(foo);\n      ", 1},
		{"\ndeclare const foo: Readonly<Readonly<Readonly<Error> & TypeError>> | 'error';\nPromise.reject(foo);\n      ", 1},
		{"\ndeclare const foo: Readonly<Readonly<Error>> | Readonly<TypeError> | 'error';\nPromise.reject(foo);\n      ", 1},
		{"\ntype Wrapper<T> = { foo: Readonly<T>[] };\ndeclare const foo: Wrapper<Error | 'error'>['foo'][5];\nPromise.reject(foo);\n      ", 1},
		{"\ndeclare const foo: Error[];\nPromise.reject(foo);\n      ", 1},
		{"\ndeclare const foo: ReadonlyArray<Error>;\nPromise.reject(foo);\n      ", 1},
		{"\ndeclare const foo: [Error];\nPromise.reject(foo);\n      ", 1},
		{"\nnew Promise(function (resolve, reject) {\n  reject();\n});\n      ", 1},
		{"\nnew Promise(function (resolve, reject) {\n  reject(5);\n});\n      ", 1},
		{"\nnew Promise((resolve, reject) => {\n  reject();\n});\n      ", 1},
		{"new Promise((resolve, reject) => reject(5));", 1},
		{"\nnew Promise((resolve, reject) => {\n  fs.readFile('foo.txt', (err, file) => {\n    if (err) reject('File not found');\n    else resolve(file);\n  });\n});\n      ", 1},
		{"new Promise((yes, no) => no(5));", 1},
		{"new Promise(({ foo, bar, baz }, reject) => reject(5));", 1},
		{"\nnew Promise(function (reject, reject) {\n  reject(5);\n});\n      ", 0}, // upstream: 1, see the duplicate-parameter note in the rule
		{"\nnew Promise(function (foo, arguments) {\n  arguments(5);\n});\n      ", 1},
		{"new Promise((foo, arguments) => arguments(5));", 1},
		{"\nnew Promise(function ({}, reject) {\n  reject(5);\n});\n      ", 1},
		{"new Promise(({}, reject) => reject(5));", 1},
		{"new Promise((resolve, reject, somethingElse = reject(5)) => {});", 1},
		{"\ndeclare const foo: {\n  bar: PromiseConstructor;\n};\nnew foo.bar((resolve, reject) => reject(5));\n      ", 1},
		{"\ndeclare const foo: {\n  bar: PromiseConstructor;\n};\nnew (foo?.bar)((resolve, reject) => reject(5));\n      ", 1},
		{"\nconst foo = Promise;\nnew foo((resolve, reject) => reject(5));\n      ", 1},
		{"\ndeclare const foo: never;\nnew Promise((resolve, reject) => reject(foo));\n      ", 1},
		{"\ndeclare const foo: unknown;\nnew Promise((resolve, reject) => reject(foo));\n      ", 1},
		{"\ntype FakeReadonly<T> = { 'fake readonly': T };\ndeclare const foo: FakeReadonly<Error>;\nnew Promise((resolve, reject) => reject(foo));\n      ", 1},
		{"\ndeclare const foo: Readonly<'error'>;\nnew Promise((resolve, reject) => reject(foo));\n      ", 1},
		{"\ndeclare const foo: Readonly<Error | 'error'>;\nnew Promise((resolve, reject) => reject(foo));\n      ", 1},
		{"\ndeclare const foo: Readonly<Error> | 'error';\nnew Promise((resolve, reject) => reject(foo));\n      ", 1},
		{"\ndeclare const foo: Readonly<Error> | Readonly<TypeError> | Readonly<'error'>;\nnew Promise((resolve, reject) => reject(foo));\n      ", 1},
		{"\ndeclare const foo: Readonly<Readonly<'error'>>;\nnew Promise((resolve, reject) => reject(foo));\n      ", 1},
		{"\ndeclare const foo: Readonly<Readonly<Readonly<Error> | 'error'>>;\nnew Promise((resolve, reject) => reject(foo));\n      ", 1},
		{"\ndeclare const foo: Readonly<Readonly<Readonly<Error> & TypeError>> | 'error';\nnew Promise((resolve, reject) => reject(foo));\n      ", 1},
		{"\ndeclare const foo: Readonly<Readonly<Error>> | Readonly<TypeError> | 'error';\nnew Promise((resolve, reject) => reject(foo));\n      ", 1},
		{"\ntype Wrapper<T> = { foo: Readonly<T>[] };\ndeclare const foo: Wrapper<Error | 'error'>['foo'][5];\nnew Promise((resolve, reject) => reject(foo));\n      ", 1},
		{"\ndeclare const foo: Error[];\nnew Promise((resolve, reject) => reject(foo));\n      ", 1},
		{"\ndeclare const foo: ReadonlyArray<Error>;\nnew Promise((resolve, reject) => reject(foo));\n      ", 1},
		{"\ndeclare const foo: [Error];\nnew Promise((resolve, reject) => reject(foo));\n      ", 1},
		{"\nclass Foo extends Promise<number> {}\nFoo.reject(5);\n      ", 1},
		{"\ndeclare const foo: PromiseConstructor & string;\nfoo.reject(5);\n      ", 1},
		{"\nclass Foo extends Promise<number> {}\nclass Bar extends Foo {}\nBar.reject(5);\n      ", 1},
		{"\ndeclare const foo: PromiseConstructor;\nfunction fun<T extends number>(t: T): void {\n  foo.reject(t);\n}\n      ", 1},
		{"\ndeclare const someUnknownValue: unknown;\nPromise.reject(someUnknownValue);\n      ", 1},
		{"\ndeclare const someAnyValue: any;\nPromise.reject(someAnyValue);\n      ", 1},
		{"\nclass CustomRejection {}\nPromise.reject(new CustomRejection());\n      ", 1},
		{"\nPromise.reject(new Date());\n      ", 1},
	}
	for index, testCase := range cases {
		t.Run(fmt.Sprintf("upstream invalid %d", index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferPromiseRejectErrors, preferPromiseRejectErrorsFile, testCase.sourceText)
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "rejectAnError"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestPreferPromiseRejectErrorsWithOptions routes every option case through the rule's own
// exported decoder rather than constructing the options struct, which is what puts the
// heterogeneous `allow` array and the string-to-enum mapping of `from` under test. Building
// the struct here would leave both untested, and they are the two pieces with no upstream
// counterpart to compare against.
func TestPreferPromiseRejectErrorsWithOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		optionsJson string
		sourceText  string
		findings    int
	}{
		{"{\"allowEmptyReject\":true}", "Promise.reject();", 0},
		{"{\"allowThrowingAny\":true,\"allowThrowingUnknown\":false}", "\ndeclare const someAnyValue: any;\nPromise.reject(someAnyValue);\n      ", 0},
		{"{\"allowThrowingAny\":false,\"allowThrowingUnknown\":true}", "\ndeclare const someUnknownValue: unknown;\nPromise.reject(someUnknownValue);\n      ", 0},
		{"{\"allowEmptyReject\":true}", "\nnew Promise(function (resolve, reject) {\n  reject();\n});\n      ", 0},
		{"{\"allowThrowingAny\":true,\"allowThrowingUnknown\":true}", "\ndeclare const someAnyValue: any;\nPromise.reject(someAnyValue);\n      ", 0},
		{"{\"allowThrowingAny\":true,\"allowThrowingUnknown\":true}", "\ndeclare const someUnknownValue: unknown;\nPromise.reject(someUnknownValue);\n      ", 0},
		{"{\"allow\":[{\"from\":\"file\",\"name\":\"CustomRejection\"}]}", "\nclass CustomRejection {}\nPromise.reject(new CustomRejection());\n      ", 0},
		{"{\"allow\":[\"CustomRejection\"]}", "\nclass CustomRejection {}\nPromise.reject(new CustomRejection());\n      ", 0},
		{"{\"allow\":[{\"from\":\"lib\",\"name\":\"Date\"}]}", "\nPromise.reject(new Date());\n      ", 0},
		{"{\"allow\":[{\"from\":\"lib\",\"name\":\"Date\"}]}", "\nnew Promise((resolve, reject) => reject(new Date()));\n      ", 0},
		{"{\"allowEmptyReject\":true}", "Promise.reject(undefined);", 1},
		{"{\"allowThrowingAny\":false,\"allowThrowingUnknown\":true}", "\ndeclare const someAnyValue: any;\nPromise.reject(someAnyValue);\n      ", 1},
		{"{\"allowThrowingAny\":true,\"allowThrowingUnknown\":false}", "\ndeclare const someUnknownValue: unknown;\nPromise.reject(someUnknownValue);\n      ", 1},
	}
	for index, testCase := range cases {
		t.Run(fmt.Sprintf("upstream option case %d", index), func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodePreferPromiseRejectErrorsOptions(json.RawMessage(testCase.optionsJson))
			if err != nil {
				t.Fatalf("decoding %s: %v", testCase.optionsJson, err)
			}
			result := rule_testing.RunTypedWithOptions(t, PreferPromiseRejectErrors, preferPromiseRejectErrorsFile, testCase.sourceText, decoded)
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "rejectAnError"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// preferPromiseRejectErrorsErrorsModule is upstream's own tests/fixtures/errors.ts, copied
// verbatim. Without it in the program the cases below would test an unresolved import rather
// than the `from: package` specifier they exist to cover.
const preferPromiseRejectErrorsErrorsModule = "// @ts-ignore\ndeclare module 'errors' {\n  class ErrorLike {}\n\n  export function createError(): ErrorLike;\n}\n"

// The two `from: package` rows below are recorded as REPORTING, and upstream has them clean.
// That is a fact about this harness rather than about the rule, and weakening the rule to
// make them green would turn a harness limitation into a rule limitation.
//
// ruletest pins `moduleDetection: "force"` (internal/rule_testing/program.go:27), which makes an
// ambient `declare module 'errors'` unresolvable. Measured with two controls: the same input
// with no options reports (so the rule sees the call), and the same input with
// `allowThrowingAny: true` goes CLEAN, which is only possible if the imported type resolved to
// `any`. An `any` carries no symbol, so no specifier can match it and the allowlist cannot
// fire. The rows are pinned here so that a future harness change makes this test fail loudly
// rather than leaving the divergence unrecorded.
func TestPreferPromiseRejectErrorsAcrossFiles(t *testing.T) {
	t.Parallel()

	cases := []struct {
		optionsJson string
		sourceText  string
		findings    int
	}{
		{"{\"allow\":[{\"from\":\"package\",\"name\":\"ErrorLike\",\"package\":\"errors\"}]}", "\nimport { createError } from 'errors';\nPromise.reject(createError());\n      ", 1},                           // upstream: 0, see the harness note above
		{"{\"allow\":[{\"from\":\"package\",\"name\":\"ErrorLike\",\"package\":\"errors\"}]}", "\nimport { createError } from 'errors';\nnew Promise((resolve, reject) => reject(createError()));\n      ", 1}, // upstream: 0, see the harness note above
		{"", "\nimport { createError } from 'errors';\nPromise.reject(createError());\n      ", 1},
		{"", "\nimport { createError } from 'errors';\nnew Promise((resolve, reject) => reject(createError()));\n      ", 1},
	}
	for index, testCase := range cases {
		t.Run(fmt.Sprintf("upstream module case %d", index), func(t *testing.T) {
			t.Parallel()
			files := map[string]string{
				"/repository/source/Rejections.ts": testCase.sourceText,
				"/repository/source/errors.ts":     preferPromiseRejectErrorsErrorsModule,
			}
			var decoded any
			if testCase.optionsJson != "" {
				value, err := DecodePreferPromiseRejectErrorsOptions(json.RawMessage(testCase.optionsJson))
				if err != nil {
					t.Fatalf("decoding %s: %v", testCase.optionsJson, err)
				}
				decoded = value
			}
			result := rule_testing.RunTypedFilesWithOptions(t, PreferPromiseRejectErrors, files, "/repository/source/Rejections.ts", decoded)
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "rejectAnError"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestPreferPromiseRejectErrorsSpans pins WHERE the finding points, which ExpectFindings
// cannot see. The finding covers the whole call expression rather than the argument, which is
// what makes the no-argument case reportable at all.
//
// The expected text is read from the RAW case at upstream's own line and column, then compared
// against the source the harness actually wrote, which is the trimmed text. Deriving it from
// the trimmed source instead lands one line early and reads exactly like an off-by-one in the
// rule.
func TestPreferPromiseRejectErrorsSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantText   string
	}{
		{"Promise.reject(5);", "Promise.reject(5)"},
		{"Promise.reject('foo');", "Promise.reject('foo')"},
		{"Promise.reject(`foo`);", "Promise.reject(`foo`)"},
		{"Promise.reject('foo', somethingElse);", "Promise.reject('foo', somethingElse)"},
		{"Promise.reject(false);", "Promise.reject(false)"},
		{"Promise.reject(void `foo`);", "Promise.reject(void `foo`)"},
		{"Promise.reject();", "Promise.reject()"},
		{"Promise.reject(undefined);", "Promise.reject(undefined)"},
		{"Promise.reject(null);", "Promise.reject(null)"},
		{"Promise.reject({ foo: 1 });", "Promise.reject({ foo: 1 })"},
		{"Promise.reject([1, 2, 3]);", "Promise.reject([1, 2, 3])"},
		{"\ndeclare const foo: Error | undefined;\nPromise.reject(foo);\n      ", "Promise.reject(foo)"},
		{"\ndeclare const foo: () => Promise<string>;\nPromise.reject(await foo());\n      ", "Promise.reject(await foo())"},
		{"\ndeclare const foo: boolean;\nPromise.reject(foo && new Error());\n      ", "Promise.reject(foo && new Error())"},
		{"\nconst foo = Promise;\nfoo.reject();\n      ", "foo.reject()"},
		{"Promise.reject?.(5);", "Promise.reject?.(5)"},
		{"Promise?.reject(5);", "Promise?.reject(5)"},
		{"Promise?.reject?.(5);", "Promise?.reject?.(5)"},
		{"(Promise?.reject)(5);", "(Promise?.reject)(5)"},
		{"(Promise?.reject)?.(5);", "(Promise?.reject)?.(5)"},
		{"Promise['reject'](5);", "Promise['reject'](5)"},
		{"Promise.reject((foo += new Error()));", "Promise.reject((foo += new Error()))"},
		{"Promise.reject((foo -= new Error()));", "Promise.reject((foo -= new Error()))"},
		{"Promise.reject((foo **= new Error()));", "Promise.reject((foo **= new Error()))"},
		{"Promise.reject((foo <<= new Error()));", "Promise.reject((foo <<= new Error()))"},
		{"Promise.reject((foo |= new Error()));", "Promise.reject((foo |= new Error()))"},
		{"Promise.reject((foo &= new Error()));", "Promise.reject((foo &= new Error()))"},
		{"\nnew Promise(function (resolve, reject) {\n  reject();\n});\n      ", "reject()"},
		{"\nnew Promise(function (resolve, reject) {\n  reject(5);\n});\n      ", "reject(5)"},
		{"\nnew Promise((resolve, reject) => {\n  fs.readFile('foo.txt', (err, file) => {\n    if (err) reject('File not found');\n    else resolve(file);\n  });\n});\n      ", "reject('File not found')"},
		{"new Promise((yes, no) => no(5));", "no(5)"},
		{"\nnew Promise(function (foo, arguments) {\n  arguments(5);\n});\n      ", "arguments(5)"},
		{"\ndeclare const foo: never;\nnew Promise((resolve, reject) => reject(foo));\n      ", "reject(foo)"},
		{"\nclass Foo extends Promise<number> {}\nFoo.reject(5);\n      ", "Foo.reject(5)"},
		{"\ndeclare const foo: PromiseConstructor & string;\nfoo.reject(5);\n      ", "foo.reject(5)"},
		{"\nclass Foo extends Promise<number> {}\nclass Bar extends Foo {}\nBar.reject(5);\n      ", "Bar.reject(5)"},
		{"\ndeclare const foo: PromiseConstructor;\nfunction fun<T extends number>(t: T): void {\n  foo.reject(t);\n}\n      ", "foo.reject(t)"},
	}
	for index, testCase := range cases {
		t.Run(fmt.Sprintf("span %d", index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferPromiseRejectErrors, preferPromiseRejectErrorsFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
			}
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			finding := result.Diagnostics[0]
			got := onDisk[finding.Range.Pos():finding.Range.End()]
			if got != testCase.wantText {
				t.Fatalf("finding points at %q, want %q", got, testCase.wantText)
			}
		})
	}
}

// TestPreferPromiseRejectErrorsExecutorReferenceScan pins the two properties of the reference scan
// that rule out both shortcuts a reader would reach for, and the shadowing row that rules out a
// third. None of these is covered by an imported case in a way that could distinguish them.
//
// Every row was measured against the installed rule before it was written here.
func TestPreferPromiseRejectErrorsExecutorReferenceScan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   int
	}{
		// Not "calls directly in the executor body": a closure counts.
		{
			"reject called from a nested arrow",
			"new Promise((resolve, reject) => { const f = () => reject(5); f(); });\n",
			1,
		},
		// Not "any mention of reject": the reference has to be the callee.
		{
			"reject passed as a value is not a rejection",
			"declare function take(f: unknown): void;\nnew Promise((resolve, reject) => { take(reject); });\n",
			0,
		},
		// Not a name match: an inner binding with the same name is a different variable.
		{
			"a shadowing inner parameter is a different binding",
			"new Promise((resolve, reject) => {\n  const inner = (reject: (r: unknown) => void) => reject(5);\n  inner(() => {});\n});\n",
			0,
		},
		// The control for the shadowing row: the same shape without the shadow does report.
		{
			"the same shape without the shadow (control)",
			"new Promise((resolve, reject) => {\n  const inner = () => reject(5);\n  inner();\n});\n",
			1,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferPromiseRejectErrors, preferPromiseRejectErrorsFile, testCase.sourceText)
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "rejectAnError"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestPreferPromiseRejectErrorsExecutorShape pins the four conditions the executor arm requires.
// The upstream source states them as a conjunction, which says nothing about which conjunct is
// doing the work, so each row here varies exactly one of them against a reporting control.
func TestPreferPromiseRejectErrorsExecutorShape(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   int
	}{
		{"arrow executor (control)", "new Promise((resolve, reject) => reject(5));\n", 1},
		{"function expression executor", "new Promise(function (resolve, reject) { reject(5); });\n", 1},
		{"only one parameter, so no reject binding", "new Promise(resolve => resolve(5));\n", 0},
		{"destructured second parameter", "new Promise((resolve, { reject }: any) => reject(5));\n", 0},
		{"rest second parameter", "new Promise((resolve, ...rest: any[]) => rest[0](5));\n", 0},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferPromiseRejectErrors, preferPromiseRejectErrorsFile, testCase.sourceText)
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "rejectAnError"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestPreferPromiseRejectErrorsStaticReceiver pins that the RECEIVER's type decides, not the
// property name. The object-literal row is the one that would report if the name alone were
// enough, and it is clean upstream.
func TestPreferPromiseRejectErrorsStaticReceiver(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		findings   int
	}{
		{"Promise.reject (control)", "Promise.reject(5);\n", 1},
		{"computed literal key", "Promise['reject'](5);\n", 1},
		{"optional chain", "Promise?.reject(5);\n", 1},
		{"a Promise subclass", "class MyPromise extends Promise<void> {}\nMyPromise.reject(5);\n", 1},
		{"a plain object with a reject method", "const o = { reject(x: unknown) {} };\no.reject(5);\n", 0},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferPromiseRejectErrors, preferPromiseRejectErrorsFile, testCase.sourceText)
			expected := make([]string, testCase.findings)
			for position := range expected {
				expected[position] = "rejectAnError"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

// TestPreferPromiseRejectErrorsDefaultsAreOff is the fixture that BYPASSES the decoder, which is the
// only way to see the nil-options path a rule configured as bare "error" actually takes.
//
// Every default here is false, so the zero value is already correct and the rule needs no defaulting
// code. That is the opposite of only-throw-error, whose three same-looking booleans all default to
// TRUE and which therefore has to carry pointers. Copying that neighbour's shape would make an
// unconfigured rule silently permissive here, and no fixture reaching the rule through the decoder
// could see it.
func TestPreferPromiseRejectErrorsDefaultsAreOff(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"any is not allowed by default", "declare const a: any;\nPromise.reject(a);\n"},
		{"unknown is not allowed by default", "declare const u: unknown;\nPromise.reject(u);\n"},
		{"an empty reject is not allowed by default", "Promise.reject();\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// nil options, exactly as a rule configured as bare "error" receives.
			result := rule_testing.RunTypedWithOptions(t, PreferPromiseRejectErrors, preferPromiseRejectErrorsFile, testCase.sourceText, nil)
			rule_testing.ExpectFindings(t, result, "rejectAnError")
		})
	}
}

// TestPreferPromiseRejectErrorsSpreadArgumentReports pins that a spread is type-checked like any
// other argument rather than skipped. Upstream reads arguments.at(0) with no kind test, so this
// reports; a port that guarded on the kind would go silent and no imported case would notice.
func TestPreferPromiseRejectErrorsSpreadArgumentReports(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, PreferPromiseRejectErrors, preferPromiseRejectErrorsFile,
		"declare const args: [unknown];\nPromise.reject(...args);\n")
	rule_testing.ExpectFindings(t, result, "rejectAnError")
}

// TestPreferPromiseRejectErrorsDuplicateParameterNamesIsADivergence pins a case upstream REPORTS and
// this port does not, so the difference is recorded rather than silent.
//
// ESLint's scope manager merges same-named parameters into one variable; our checker gives each its
// own symbol and the call site resolves to the first. Measured by printing the symbol pointers at
// both parameters and the call site. Reproducing upstream would mean name matching, which would cost
// the shadowing case above, and duplicate parameter names are a TypeScript error anyway.
//
// If the checker ever starts merging them, this test fails and the divergence can be removed.
func TestPreferPromiseRejectErrorsDuplicateParameterNamesIsADivergence(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, PreferPromiseRejectErrors, preferPromiseRejectErrorsFile,
		"new Promise(function (reject, reject) {\n  reject(5);\n});\n")
	rule_testing.ExpectClean(t, result)

	// The control: distinct names, same shape, reports. Without it the silence above could be
	// explained by the function-expression executor not being handled at all.
	control := rule_testing.RunTyped(t, PreferPromiseRejectErrors, preferPromiseRejectErrorsFile,
		"new Promise(function (resolve, reject) {\n  reject(5);\n});\n")
	rule_testing.ExpectFindings(t, control, "rejectAnError")
}

// TestPreferPromiseRejectErrorsMessage asserts the reported id and description against literals
// typed here rather than against the rule's own constant, which would move with any mutation to it.
func TestPreferPromiseRejectErrorsMessage(t *testing.T) {
	t.Parallel()

	message := buildPreferPromiseRejectErrorsMessage()
	if message.Id != "rejectAnError" {
		t.Fatalf("message id is %q, want %q", message.Id, "rejectAnError")
	}
	if !strings.Contains(message.Description, "not an Error") {
		t.Fatalf("description does not say what is wrong: %q", message.Description)
	}
}

// TestPreferPromiseRejectErrorsRequiresTheTypedHarness pins that this rule declares the checker and
// guards on it in both listeners. Run through the plain harness the checker is nil, and the shim
// returns nil rather than panicking, so a missing guard buys a vacuous green rather than a crash.
func TestPreferPromiseRejectErrorsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !PreferPromiseRejectErrors.NeedsTypeChecker {
		t.Fatal("rule must declare NeedsTypeChecker")
	}
	for _, source := range []string{"Promise.reject(5);\n", "new Promise((resolve, reject) => reject(5));\n"} {
		typed := rule_testing.RunTyped(t, PreferPromiseRejectErrors, preferPromiseRejectErrorsFile, source)
		if len(typed.Diagnostics) != 1 {
			t.Fatalf("typed harness on %q: want 1 finding, got %d", source, len(typed.Diagnostics))
		}
		untyped := rule_testing.Run(t, PreferPromiseRejectErrors, preferPromiseRejectErrorsFile, source)
		if len(untyped.Diagnostics) != 0 {
			t.Fatalf("untyped harness on %q: want 0 from the nil-checker guard, got %d", source, len(untyped.Diagnostics))
		}
	}
}
