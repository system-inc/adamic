package typescript

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const noUselessDefaultAssignmentFile = "/repository/source/Defaults.ts"

func noUselessDefaultAssignmentCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// noUselessDefaultAssignmentOptionsFor routes a case through the rule's own decoder.
//
// Building the options struct directly would leave the decoder untested, and the decoder is where
// upstream's long option name is transcribed. A typo there yields a struct that silently keeps
// the default, which no fixture built from a struct could see.
func noUselessDefaultAssignmentOptionsFor(t *testing.T, optionsJson string) any {
	t.Helper()
	if optionsJson == "" {
		return DefaultNoUselessDefaultAssignmentSettings()
	}
	decoded, err := DecodeNoUselessDefaultAssignmentOptions(json.RawMessage(optionsJson))
	if err != nil {
		t.Fatalf("decoding %s: %v", optionsJson, err)
	}
	return decoded
}

// TestNoUselessDefaultAssignmentStaysSilent is upstream's sixty two passing cases verbatim.
//
// Each was re-measured one file per program against the installed 8.67.0 build under THIS
// harness's compiler options. Three of them are written by upstream against a
// `noUncheckedIndexedAccess` project our harness cannot express; measured, all three stay clean
// under our options too, so the option is not what decides them and importing them costs nothing.
func TestNoUselessDefaultAssignmentStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"function Bar({ foo = '' }: { foo?: string }) {\n  return foo;\n}\n",
		"const { foo } = { foo: 'bar' };\n",
		"[1, 2, 3, undefined].map((a = 42) => a + 1);\n",
		"function test(a?: number) {\n  return a;\n}\n",
		"const obj: { a?: string } = {};\nconst { a = 'default' } = obj;\n",
		"function test(a: string | undefined = 'default') {\n  return a;\n}\n",
		"(a: string = 'default') => a;\n",
		"function test(a: string = 'default') {\n  return a;\n}\n",
		"class C {\n  public test(a: string = 'default') {\n    return a;\n  }\n}\n",
		"const obj: { a: string | undefined } = { a: undefined };\nconst { a = 'default' } = obj;\n",
		"function test(arr: number[] | undefined = []) {\n  return arr;\n}\n",
		"function Bar({ nested: { foo = '' } = {} }: { nested?: { foo?: string } }) {\n  return foo;\n}\n",
		"function test(a: any = 'default') {\n  return a;\n}\n",
		"function test(a: unknown = 'default') {\n  return a;\n}\n",
		"function test(a = 5) {\n  return a;\n}\n",
		"function createValidator(): () => void {\n  return (param = 5) => {};\n}\n",
		"function Bar({ foo = '' }: { foo: any }) {\n  return foo;\n}\n",
		"function Bar({ foo = '' }: { foo: unknown }) {\n  return foo;\n}\n",
		"function getValue(): undefined;\nfunction getValue(box: { value: string }): string;\nfunction getValue({ value = '' }: { value?: string } = {}): string | undefined {\n  return value;\n}\n",
		"function getValueObject({ value = '' }: Partial<{ value: string }>) {\n  return value;\n}\n",
		"const { value = 'default' } = someUnknownFunction();\n",
		"const [value = 'default'] = someUnknownFunction();\n",
		"for (const { value = 'default' } of []) {\n}\n",
		"for (const [value = 'default'] of []) {\n}\n",
		"declare const x: [[number | undefined]];\nconst [[a = 1]] = x;\n",
		"function foo(x: string = '') {}\n",
		"class C {\n  method(x: string = '') {}\n}\n",
		"const foo = (x: string = '') => {};\n",
		"const obj = { ab: { x: 1 } };\nconst {\n  ['a' + 'b']: { x = 1 },\n} = obj;\n",
		"const obj = { ab: 1 };\nconst { ['a' + 'b']: x = 1 } = obj;\n",
		"for ([[a = 1]] of []) {\n}\n",
		"declare const g: Array<string>;\nconst [foo = ''] = g;\n",
		"declare const g: Record<string, string>;\nconst { foo = '' } = g;\n",
		"declare const h: { [key: string]: string };\nconst { bar = '' } = h;\n",
		"declare const g: Array<string>;\nconst [foo = ''] = g;\n",
		"declare const g: Record<string, string>;\nconst { foo = '' } = g;\n",
		"declare const h: { [key: string]: string };\nconst { bar = '' } = h;\n",
		"type Merge = boolean | ((incoming: string[]) => void);\n\nconst policy: { merge: Merge } = {\n  merge: (incoming: string[] = []) => {\n    incoming;\n  },\n};\n",
		"const [a, b = ''] = 'somestr'.split('.');\n",
		"declare const params: string[];\nconst [c = '123'] = params;\n",
		"declare function useCallback<T>(callback: T);\nuseCallback((value: number[] = []) => {});\n",
		"declare const tuple: [string];\nconst [a, b = 'default'] = tuple;\n",
		"const run = (cb: (...args: unknown[]) => void) => cb();\nconst cb = (p: boolean = true) => null;\nrun(cb);\nrun((p: boolean = true) => null);\n",
		"const { a = 'default' } = Math.random() > 0.5 ? { a: 'Hello' } : {};\n",
		"const { a = 'default' } =\n  Math.random() > 0.5 ? (Math.random() > 0.5 ? { a: 'Hello' } : {}) : {};\n",
		"function findPosts({\n  category,\n  maxResults = 100,\n}: {\n  category: string;\n  maxResults?: number;\n}): Promise<string[]> {\n  return Promise.resolve([category, String(maxResults)]);\n}\n",
		"const { a = 'baz' } = cond ? {} : { a: 'bar' };\n",
		"const { a = 'baz' } = cond ? foo : { a: 'bar' };\n",
		"const { a = 'baz' } = foo && { a: 'bar' };\n",
		"const { a = 'baz' } = cond ? { a: 'foo', ...extra } : { a: 'bar' };\n",
		"const { a = 'baz' } = cond ? { ...foo } : { a: 'bar' };\n",
		"const key = Math.random() > 0.5 ? 'a' : 'b';\nconst { a = 'baz' } = cond ? { [key]: 'foo' } : { [key]: 'bar' };\n",
		"const { a = 'baz' } = cond ? foo && { a: 'bar' } : { a: 'baz' };\n",
		"const obj: unknown = { a: 'bar' };\nconst { a = 'baz' } = cond ? obj : { a: 'bar' };\n",
		"const sym = Symbol('a');\nconst { a = 'baz' } = cond ? { [sym]: 'foo' } : { [sym]: 'bar' };\n",
		"const { a = 'baz' } = cond ? { [`a${1}`]: 'foo' } : { a: 'bar' };\n",
		"class AbstractEntity {\n  public a: string | undefined;\n  public static fromJson<T extends { a: string }>(\n    this: new () => T,\n    { inner = { a: 'test' } }: { inner?: { a: string } },\n  ): T {\n    const entity = new this();\n    entity.a = inner?.a;\n    return entity;\n  }\n}\n",
		"type FetchFn<TParams> =\n  Partial<TParams> extends TParams\n    ? (params?: TParams) => void\n    : (params: TParams) => void;\n\nfunction createFetcher<TParams>() {\n  type Params = TParams;\n\n  const fn: FetchFn<TParams> = (\n    params: Partial<Params> = {} as Partial<Params>,\n  ) => {\n    console.log(params);\n  };\n\n  return fn;\n}\n",
		"interface Foos {\n  bar?: number;\n}\nconst foos: Foos[] = [];\nfoos.flatMap(({ bar = 42 }) => bar);\n",
		"function f(this: void, { bar = 42 }: { bar?: number }) {\n  return bar;\n}\n",
		"interface Fn {\n  (value: string): void;\n  (): void;\n}\nconst fn: Fn = (value = 'default') => {\n  return value;\n};\n",
		"interface Fn {\n  (value: string): void;\n  (value?: string): void;\n}\nconst fn: Fn = (value = 'default') => {\n  return value;\n};\n",
	}
	for index, sourceText := range cases {
		t.Run(noUselessDefaultAssignmentCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoUselessDefaultAssignment,
				noUselessDefaultAssignmentFile, sourceText,
				noUselessDefaultAssignmentOptionsFor(t, "")))
		})
	}
}

// TestNoUselessDefaultAssignmentFires is upstream's twenty three reporting cases verbatim.
//
// Every row asserts the span, the id, the whole rendered message, and the rewritten file. The
// rewrite carries the most weight: the repair deletes a range computed from the BOUND NAME's end
// rather than the default's start, so a fixer built the obvious way leaves `function foo(a = )`,
// which still satisfies every message id while producing source that does not parse. One case's
// repair also inserts a `?`, and one keeps a comment sitting between the name and the annotation.
//
// Three of these are written by upstream against a project WITHOUT strictNullChecks, where
// upstream emits an extra file-level `noStrictNullCheck` finding alongside the real one. Measured
// under our options, the real finding is identical and the file-level one is absent, which is
// what these rows assert. See the rule's doc comment for why the gate is a decline here.
func TestNoUselessDefaultAssignmentFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		optionsJson string
		wantSpan    string
		wantId      string
		wantMessage string
		wantFixed   string
	}{
		{
			sourceText:  "function Bar({ foo = '' }: { foo: string }) {\n  return foo;\n}\n",
			optionsJson: "",
			wantSpan:    "''",
			wantId:      "uselessDefaultAssignment",
			wantMessage: "Default value is useless because the property is not optional.",
			wantFixed:   "function Bar({ foo }: { foo: string }) {\n  return foo;\n}\n",
		},
		{
			sourceText:  "class C {\n  public method({ foo = '' }: { foo: string }) {\n    return foo;\n  }\n}\n",
			optionsJson: "",
			wantSpan:    "''",
			wantId:      "uselessDefaultAssignment",
			wantMessage: "Default value is useless because the property is not optional.",
			wantFixed:   "class C {\n  public method({ foo }: { foo: string }) {\n    return foo;\n  }\n}\n",
		},
		{
			sourceText:  "const { 'literal-key': literalKey = 'default' } = { 'literal-key': 'value' };\n",
			optionsJson: "",
			wantSpan:    "'default'",
			wantId:      "uselessDefaultAssignment",
			wantMessage: "Default value is useless because the property is not optional.",
			wantFixed:   "const { 'literal-key': literalKey } = { 'literal-key': 'value' };\n",
		},
		{
			sourceText:  "[1, 2, 3].map((a = 42) => a + 1);\n",
			optionsJson: "",
			wantSpan:    "42",
			wantId:      "uselessDefaultAssignment",
			wantMessage: "Default value is useless because the parameter is not optional.",
			wantFixed:   "[1, 2, 3].map((a) => a + 1);\n",
		},
		{
			sourceText:  "function getValue(): undefined;\nfunction getValue(box: { value: string }): string;\nfunction getValue({ value = '' }: { value: string } = {}): string | undefined {\n  return value;\n}\n",
			optionsJson: "",
			wantSpan:    "''",
			wantId:      "uselessDefaultAssignment",
			wantMessage: "Default value is useless because the property is not optional.",
			wantFixed:   "function getValue(): undefined;\nfunction getValue(box: { value: string }): string;\nfunction getValue({ value }: { value: string } = {}): string | undefined {\n  return value;\n}\n",
		},
		{
			sourceText:  "function getValue([value = '']: [string]) {\n  return value;\n}\n",
			optionsJson: "",
			wantSpan:    "''",
			wantId:      "uselessDefaultAssignment",
			wantMessage: "Default value is useless because the property is not optional.",
			wantFixed:   "function getValue([value]: [string]) {\n  return value;\n}\n",
		},
		{
			sourceText:  "declare const x: { hello: { world: string } };\n\nconst {\n  hello: { world = '' },\n} = x;\n",
			optionsJson: "",
			wantSpan:    "''",
			wantId:      "uselessDefaultAssignment",
			wantMessage: "Default value is useless because the property is not optional.",
			wantFixed:   "declare const x: { hello: { world: string } };\n\nconst {\n  hello: { world },\n} = x;\n",
		},
		{
			sourceText:  "declare const x: { hello: Array<{ world: string }> };\n\nconst {\n  hello: [{ world = '' }],\n} = x;\n",
			optionsJson: "",
			wantSpan:    "''",
			wantId:      "uselessDefaultAssignment",
			wantMessage: "Default value is useless because the property is not optional.",
			wantFixed:   "declare const x: { hello: Array<{ world: string }> };\n\nconst {\n  hello: [{ world }],\n} = x;\n",
		},
		{
			sourceText:  "interface B {\n  foo: (b: boolean | string) => void;\n}\n\nconst h: B = {\n  foo: (b = false) => {},\n};\n",
			optionsJson: "",
			wantSpan:    "false",
			wantId:      "uselessDefaultAssignment",
			wantMessage: "Default value is useless because the parameter is not optional.",
			wantFixed:   "interface B {\n  foo: (b: boolean | string) => void;\n}\n\nconst h: B = {\n  foo: (b) => {},\n};\n",
		},
		{
			sourceText:  "function foo(a = undefined) {}\n",
			optionsJson: "",
			wantSpan:    "undefined",
			wantId:      "uselessUndefined",
			wantMessage: "Default value is useless because it is undefined. Optional parameters are already undefined by default.",
			wantFixed:   "function foo(a) {}\n",
		},
		{
			sourceText:  "const { a = undefined } = {};\n",
			optionsJson: "",
			wantSpan:    "undefined",
			wantId:      "uselessUndefined",
			wantMessage: "Default value is useless because it is undefined. Optional propertys are already undefined by default.",
			wantFixed:   "const { a } = {};\n",
		},
		{
			sourceText:  "const [a = undefined] = [];\n",
			optionsJson: "",
			wantSpan:    "undefined",
			wantId:      "uselessUndefined",
			wantMessage: "Default value is useless because it is undefined. Optional propertys are already undefined by default.",
			wantFixed:   "const [a] = [];\n",
		},
		{
			sourceText:  "function foo({ a = undefined }) {}\n",
			optionsJson: "",
			wantSpan:    "undefined",
			wantId:      "uselessUndefined",
			wantMessage: "Default value is useless because it is undefined. Optional propertys are already undefined by default.",
			wantFixed:   "function foo({ a }) {}\n",
		},
		{
			sourceText:  "function myFunction(p1: string, p2: number | undefined = undefined) {\n  console.log(p1, p2);\n}\n",
			optionsJson: "",
			wantSpan:    "undefined",
			wantId:      "preferOptionalSyntax",
			wantMessage: "Using `= undefined` to make a parameter optional adds unnecessary runtime logic. Use the `?` optional syntax instead.",
			wantFixed:   "function myFunction(p1: string, p2?: number | undefined) {\n  console.log(p1, p2);\n}\n",
		},
		{
			sourceText:  "type SomeType = number | undefined;\nfunction f(\n  /* comment */ x /* comment 2 */ : /* comment 3 */ SomeType /* comment 4 */ = /* comment 5 */ undefined,\n) {}\n",
			optionsJson: "",
			wantSpan:    "undefined",
			wantId:      "preferOptionalSyntax",
			wantMessage: "Using `= undefined` to make a parameter optional adds unnecessary runtime logic. Use the `?` optional syntax instead.",
			wantFixed:   "type SomeType = number | undefined;\nfunction f(\n  /* comment */ x? /* comment 2 */ : /* comment 3 */ SomeType,\n) {}\n",
		},
		{
			sourceText:  "function Bar({ foo = '' }: { foo: string }) {\n  return foo;\n}\n",
			optionsJson: "",
			wantSpan:    "''",
			wantId:      "uselessDefaultAssignment",
			wantMessage: "Default value is useless because the property is not optional.",
			wantFixed:   "function Bar({ foo }: { foo: string }) {\n  return foo;\n}\n",
		},
		{
			sourceText:  "function foo(a = undefined) {}\n",
			optionsJson: "",
			wantSpan:    "undefined",
			wantId:      "uselessUndefined",
			wantMessage: "Default value is useless because it is undefined. Optional parameters are already undefined by default.",
			wantFixed:   "function foo(a) {}\n",
		},
		{
			sourceText:  "function Bar({ foo = '' }: { foo: string }) {\n  return foo;\n}\n",
			optionsJson: "{\"allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing\": true}",
			wantSpan:    "''",
			wantId:      "uselessDefaultAssignment",
			wantMessage: "Default value is useless because the property is not optional.",
			wantFixed:   "function Bar({ foo }: { foo: string }) {\n  return foo;\n}\n",
		},
		{
			sourceText:  "const { a = 'baz' } = Math.random() < 0.5 ? { a: 'foo' } : { a: 'bar' };\n",
			optionsJson: "",
			wantSpan:    "'baz'",
			wantId:      "uselessDefaultAssignment",
			wantMessage: "Default value is useless because the property is not optional.",
			wantFixed:   "const { a } = Math.random() < 0.5 ? { a: 'foo' } : { a: 'bar' };\n",
		},
		{
			sourceText:  "const { a = 'baz' } =\n  Math.random() < 0.5\n    ? { a: 'foo' }\n    : Math.random() > 0.2\n      ? { a: 'bar' }\n      : { a: 'qux' };\n",
			optionsJson: "",
			wantSpan:    "'baz'",
			wantId:      "uselessDefaultAssignment",
			wantMessage: "Default value is useless because the property is not optional.",
			wantFixed:   "const { a } =\n  Math.random() < 0.5\n    ? { a: 'foo' }\n    : Math.random() > 0.2\n      ? { a: 'bar' }\n      : { a: 'qux' };\n",
		},
		{
			sourceText:  "const { a = 'baz' } = cond ? { ['a']: 'foo' } : { ['a']: 'bar' };\n",
			optionsJson: "",
			wantSpan:    "'baz'",
			wantId:      "uselessDefaultAssignment",
			wantMessage: "Default value is useless because the property is not optional.",
			wantFixed:   "const { a } = cond ? { ['a']: 'foo' } : { ['a']: 'bar' };\n",
		},
		{
			sourceText:  "const { a = 'baz' } = cond ? { a() {} } : { a: 'bar' };\n",
			optionsJson: "",
			wantSpan:    "'baz'",
			wantId:      "uselessDefaultAssignment",
			wantMessage: "Default value is useless because the property is not optional.",
			wantFixed:   "const { a } = cond ? { a() {} } : { a: 'bar' };\n",
		},
		{
			sourceText:  "const { a = 'b' } = Math.random() < 0.5 ? { [`a`]: 'a' } : { a: 'b' };\n",
			optionsJson: "",
			wantSpan:    "'b'",
			wantId:      "uselessDefaultAssignment",
			wantMessage: "Default value is useless because the property is not optional.",
			wantFixed:   "const { a } = Math.random() < 0.5 ? { [`a`]: 'a' } : { a: 'b' };\n",
		},
	}
	for index, testCase := range cases {
		t.Run(noUselessDefaultAssignmentCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, NoUselessDefaultAssignment,
				noUselessDefaultAssignmentFile, testCase.sourceText,
				noUselessDefaultAssignmentOptionsFor(t, testCase.optionsJson))

			rule_testing.ExpectFindings(t, result, testCase.wantId)

			// The harness writes the fixture trimmed, so both the span slice and the expected
			// rewrite are against that text rather than the Go literal above.
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			diagnostic := result.Diagnostics[0]

			gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Fatalf("span: expected %q, got %q", testCase.wantSpan, gotSpan)
			}
			if diagnostic.Message.Id != testCase.wantId {
				t.Fatalf("id: expected %q, got %q", testCase.wantId, diagnostic.Message.Id)
			}
			if diagnostic.Message.Description != testCase.wantMessage {
				t.Fatalf("message: expected %q, got %q", testCase.wantMessage,
					diagnostic.Message.Description)
			}

			rule_testing.ExpectFixedSource(t, result, strings.TrimSpace(testCase.wantFixed)+"\n")
		})
	}
}
