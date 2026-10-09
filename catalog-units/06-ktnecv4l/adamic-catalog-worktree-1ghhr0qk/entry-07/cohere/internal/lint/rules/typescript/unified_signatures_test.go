package typescript

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus is typescript-eslint's own, imported verbatim from
// packages/eslint-plugin/tests/rules/unified-signatures.test.ts: 50 valid rows and 62 invalid rows.
// Every source body was byte-checked against that file, and the check was itself controlled by
// corrupting one body and confirming it went red.
//
// Spans are asserted as the TEXT a finding covers rather than as line and column. The harness
// normalises a fixture to `strings.TrimSpace(contents)+"\n"` while the corpus writes its sources as
// indented template literals starting with a newline, so every upstream line number is one greater
// than ours. Text spans do not care: they are computed once from upstream's line/column against
// upstream's own text, and then compared against what our rule points at.
//
// # Four rows the installed oracle cannot see
//
// The clone is newer than the installed 8.67.0 build, which carries no `allParametersAreSame`
// message at all (`grep -c` over its dist returns 0). Four invalid rows assert it, and they are
// ported from the clone, which the brief names as the source. The other 58 were replayed through the
// installed build and agree with the corpus on every message id and every span, to the column.

// asTheHarnessWroteIt reproduces the fixture normalisation, so a span assertion indexes the same
// bytes the rule saw.
func asTheHarnessWroteIt(sourceText string) string {
	return strings.TrimSpace(sourceText) + "\n"
}

// unifiedSignaturesOptionsAsDecoded hands the rule what the config layer hands it: the options struct
// by value, or nil when nothing was configured. The table keeps pointers only so a row can say nil.
func unifiedSignaturesOptionsAsDecoded(options *UnifiedSignaturesOptions) any {
	if options == nil {
		return nil
	}
	return *options
}

func unifiedSignaturesFileFor(isJsx bool) string {
	if isJsx {
		return "file.tsx"
	}
	return "file.ts"
}

type unifiedSignaturesCase struct {
	sourceText string
	options    *UnifiedSignaturesOptions
	isJsx      bool
	wantIds    []string
	wantSpans  []string
	wantStarts []string
}

func TestUnifiedSignaturesValid(t *testing.T) {
	t.Parallel()

	cases := []unifiedSignaturesCase{
		{
			// Scope containment. Both members are named `f` and neither reports on its own, so this
			// only stays silent if the nested type literal's member is attributed to the LITERAL
			// rather than to the enclosing interface. A walk that descended into nested scopes while
			// collecting the outer one would group these two and report.
			sourceText: "interface I {\n  f(x: number): void;\n  a: { f(x: string): void };\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			// The overload key's computed-ness digit. A computed member never groups with a plain
			// one of the same name, because the rule cannot know a computed key resolves to that
			// name. Measured against the installed build, which is silent here and reports on the
			// two-computed case below.
			sourceText: "interface I {\n  [`a`](x: number): void;\n  a(x: string): void;\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "function g(): void;\nfunction g(a: number, b: number): void;\nfunction g(a?: number, b?: number): void {}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "function rest(...xs: number[]): void;\nfunction rest(xs: number[], y: string): void;\nfunction rest(...args: any[]) {}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "class C {\n  constructor();\n  constructor(a: number, b: number);\n  constructor(a?: number, b?: number) {}\n\n  a(): void;\n  a(a: number, b: number): void;\n  a(a?: number, b?: number): void {}\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "declare class Example {\n  privateMethod(a: number): void;\n  #privateMethod(a: number, b?: string): void;\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "declare class Example {\n  #privateMethod1(a: number): void;\n  #privateMethod2(a: number, b?: string): void;\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "interface I {\n  a2(): void;\n  a2(x: number, y: number): void;\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "interface I {\n  a4(): void;\n  a4(x: number): number;\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "interface I {\n  a5<T>(x: T): T;\n  a5(x: number): number;\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "interface I {\n  b2(x: string): void;\n  b2(...x: number[]): void;\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "interface I {\n  b3(...x: number[]): void;\n  b3(...x: string[]): void;\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "interface I {\n  c3(x: number): void;\n  c3(x?: string): void;\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "interface I {\n  d2(x: string, y: number): void;\n  d2(x: number, y: string): void;\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "declare class D {\n  static a();\n  a(x: number);\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "interface Generic<T> {\n  x(): void;\n  x(x: T[]): void;\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "interface I {\n  f(x1: number): void;\n  f(x1: boolean, x2?: number): void;\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "function f<T extends number>(x: T[]): void;\nfunction f<T extends string>(x: T): void;\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "declare function foo(n: number): number;\n\ndeclare module 'hello' {\n  function foo(n: number, s: string): number;\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "{\n  function block(): number;\n  function block(n: number): number;\n  function block(n?: number): number {\n    return 3;\n  }\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "export interface Foo {\n  bar(baz: string): number[];\n  bar(): string[];\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "declare module 'foo' {\n  export default function (foo: number): string[];\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "export default function (foo: number): string[];\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "function p(key: string): Promise<string | undefined>;\nfunction p(key: string, defaultValue: string): Promise<string>;\nfunction p(key: string, defaultValue?: string): Promise<string | undefined> {\n  const obj: Record<string, string> = {};\n  return obj[key] || defaultValue;\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "interface I {\n  p<T>(x: T): Promise<T>;\n  p(x: number): Promise<number>;\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "function rest(...xs: number[]): Promise<number[]>;\nfunction rest(xs: number[], y: string): Promise<string>;\nasync function rest(...args: any[], y?: string): Promise<number[] | string> {\n  return y || args;\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "declare class Foo {\n  get bar();\n  set bar(x: number);\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "interface Foo {\n  get bar();\n  set bar(x: number);\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "abstract class Foo {\n  abstract get bar();\n  abstract set bar(a: unknown);\n}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "function f(a: number): void;\nfunction f(b: string): void;\nfunction f(a: number | string): void {}\n",
			options:    &UnifiedSignaturesOptions{IgnoreDifferentlyNamedParameters: true},
			isJsx:      false,
		},
		{
			sourceText: "function f(m: number): void;\nfunction f(v: number, u: string): void;\nfunction f(v: number, u?: string): void {}\n",
			options:    &UnifiedSignaturesOptions{IgnoreDifferentlyNamedParameters: true},
			isJsx:      false,
		},
		{
			sourceText: "function f(v: boolean): number;\nfunction f(): string;\n",
			options:    &UnifiedSignaturesOptions{IgnoreDifferentlyNamedParameters: true},
			isJsx:      false,
		},
		{
			sourceText: "function f(v: boolean, u: boolean): number;\nfunction f(v: boolean): string;\n",
			options:    &UnifiedSignaturesOptions{IgnoreDifferentlyNamedParameters: true},
			isJsx:      false,
		},
		{
			sourceText: "function f(v: number, u?: string): void {}\nfunction f(v: number): void;\nfunction f(): string;\n",
			options:    &UnifiedSignaturesOptions{IgnoreDifferentlyNamedParameters: true},
			isJsx:      false,
		},
		{
			sourceText: "function f(a: boolean, ...c: number[]): void;\nfunction f(a: boolean, ...d: string[]): void;\nfunction f(a: boolean, ...c: (number | string)[]): void {}\n",
			options:    &UnifiedSignaturesOptions{IgnoreDifferentlyNamedParameters: true},
			isJsx:      false,
		},
		{
			sourceText: "class C {\n  constructor();\n  constructor(a: number, b: number);\n  constructor(c?: number, b?: number) {}\n\n  a(): void;\n  a(a: number, b: number): void;\n  a(a?: number, d?: number): void {}\n}\n",
			options:    &UnifiedSignaturesOptions{IgnoreDifferentlyNamedParameters: true},
			isJsx:      false,
		},
		{
			sourceText: "/** @deprecated */\ndeclare function f(x: number): unknown;\ndeclare function f(x: boolean): unknown;\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
		},
		{
			sourceText: "declare function f(x: number): unknown;\n/** @deprecated */\ndeclare function f(x: boolean): unknown;\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
		},
		{
			sourceText: "declare function f(x: number): unknown;\n/** @deprecated */ declare function f(x: boolean): unknown;\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
		},
		{
			sourceText: "declare function f(x: string): void;\n/**\n * @async\n */\ndeclare function f(x: boolean): void;\n/**\n * @deprecate\n */\ndeclare function f(x: number): void;\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
		},
		{
			sourceText: "/**\n * @deprecate\n */\ndeclare function f(x: string): void;\n/**\n * @async\n */\ndeclare function f(x: boolean): void;\ndeclare function f(x: number): void;\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
		},
		{
			sourceText: "/**\n * This signature does something.\n */\ndeclare function f(x: number): void;\n\n/**\n * This signature does something else.\n */\ndeclare function f(x: string): void;\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
		},
		{
			sourceText: "/** @deprecated */\nexport function f(x: number): unknown;\nexport function f(x: boolean): unknown;\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
		},
		{
			sourceText: "/**\n * This signature does something.\n */\n\n// some other comment\nexport function f(x: number): void;\n\n/**\n * This signature does something else.\n */\nexport function f(x: string): void;\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
		},
		{
			sourceText: "interface I {\n  /**\n   * This signature does something else.\n   */\n  f(x: number): void;\n  f(x: string): void;\n}\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
		},
		{
			sourceText: "/* @deprecated */\ndeclare function f(x: number): unknown;\ndeclare function f(x: boolean): unknown;\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
		},
		{
			sourceText: "/*\n * This signature does something.\n */\ndeclare function f(x: number): unknown;\ndeclare function f(x: boolean): unknown;\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
		},
		{
			sourceText: "/**\n * This signature does something.\n **/\ndeclare function f(x: number): unknown;\ndeclare function f(x: boolean): unknown;\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
		},
		{
			sourceText: "class C {\n  a(b: string): void;\n  /**\n   * @deprecate\n   */\n  a(b: number): void;\n}\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
		},
		{
			sourceText: "function f(): void;\nfunction f(this: {}): void;\nfunction f(this: void | {}): void {}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "function f(a: boolean): void;\nfunction f(this: {}, a: boolean): void;\nfunction f(this: void | {}, a: boolean): void {}\n",
			options:    nil,
			isJsx:      false,
		},
		{
			sourceText: "function f(this: void, a: boolean): void;\nfunction f(this: {}, a: boolean): void;\nfunction f(this: void | {}, a: boolean): void {}\n",
			options:    nil,
			isJsx:      false,
		},
	}
	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, UnifiedSignatures,
			unifiedSignaturesFileFor(testCase.isJsx), testCase.sourceText, unifiedSignaturesOptionsAsDecoded(testCase.options))
		rule_testing.ExpectClean(t, result)
	}
}

func TestUnifiedSignaturesInvalid(t *testing.T) {
	t.Parallel()

	cases := []unifiedSignaturesCase{
		{
			// Two reporting GROUPS in one scope, which is the only shape that can observe the order
			// findings come out in: every other fixture asserts a single finding. Groups are
			// iterated in first-seen order rather than over the map, and this row asserts source
			// order. It is worth knowing that this catches a map-order mutation only ~8% of the
			// time, because Go's randomisation happens to favour insertion order for two keys; the
			// determinism is enforced in the rule rather than relied on from here.
			sourceText: "interface I {\n  a(x: number): void;\n  a(x: string): void;\n  b(y: number): void;\n  b(y: string): void;\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference", "singleParameterDifference"},
			wantSpans:  []string{"x: string", "y: string"},
			wantStarts: []string{
				"These overloads can be combined into one signature taking `number | string`.",
				"These overloads can be combined into one signature taking `number | string`.",
			},
		},
		{
			// The control for the containment case in the valid table: the same two signatures
			// INSIDE one type literal DO group, so the rule is not simply ignoring nested members.
			sourceText: "interface I {\n  a: { f(x: number): void; f(x: string): void };\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: string"},
			wantStarts: []string{"These overloads can be combined into one signature taking `number | string`."},
		},
		{
			// The control for the row above: two computed members with the SAME key text do group,
			// so the digit is not simply suppressing every computed member.
			sourceText: "interface I {\n  [`a`](x: number): void;\n  [`a`](x: string): void;\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: string"},
			wantStarts: []string{"These overloads can be combined into one signature taking `number | string`."},
		},
		{
			sourceText: "declare function f(a: number): void;\ndeclare function f(a: number): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"allParametersAreSame"},
			wantSpans:  []string{"declare function f(a: number): void;"},
			wantStarts: []string{"These overloads can be combined into one signature with identical parameters."},
		},
		{
			sourceText: "declare function f(a: number): void;\ndeclare function f(a: number): void;\ndeclare function f(a: string): string;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"allParametersAreSame"},
			wantSpans:  []string{"declare function f(a: number): void;"},
			wantStarts: []string{"This overload and the one on line 1 can be combined into one signature with identical parameters."},
		},
		{
			sourceText: "function f(a: number): void;\nfunction f(a: number): void;\nfunction f(a: number): void {}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"allParametersAreSame"},
			wantSpans:  []string{"function f(a: number): void;"},
			wantStarts: []string{"These overloads can be combined into one signature with identical parameters."},
		},
		{
			sourceText: "class A {\n  f(a: number): void;\n  f(a: number): void;\n  f(a: number): void {}\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"allParametersAreSame"},
			wantSpans:  []string{"(a: number): void;"},
			wantStarts: []string{"These overloads can be combined into one signature with identical parameters."},
		},
		{
			sourceText: "function f(a: number): void;\nfunction f(b: string): void;\nfunction f(a: number | string): void {}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"b: string"},
			wantStarts: []string{"These overloads can be combined into one signature taking `number | string`."},
		},
		{
			sourceText: "function f(x: number): void;\nfunction f(x: string): void;\nfunction f(x: any): any {\n  return x;\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: string"},
			wantStarts: []string{"These overloads can be combined into one signature taking `number | string`."},
		},
		{
			sourceText: "function f(x: number): void;\nfunction f(x: string): void;\nfunction f(x: any): any {\n  return x;\n}\n",
			options:    &UnifiedSignaturesOptions{IgnoreDifferentlyNamedParameters: true},
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: string"},
			wantStarts: []string{"These overloads can be combined into one signature taking `number | string`."},
		},
		{
			sourceText: "function opt(xs?: number[]): void;\nfunction opt(xs: number[], y: string): void;\nfunction opt(...args: any[]) {}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"omittingSingleParameter"},
			wantSpans:  []string{"y: string"},
			wantStarts: []string{"These overloads can be combined into one signature with an optional parameter."},
		},
		{
			sourceText: "interface I {\n  a0(): void;\n  a0(x: string): string;\n  a0(x: number): void;\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"omittingSingleParameter"},
			wantSpans:  []string{"x: number"},
			wantStarts: []string{"This overload and the one on line 2 can be combined into one signature with an optional parameter."},
		},
		{
			sourceText: "interface I {\n  a0(): void;\n  a0(x: string): string;\n  a0(x: number): void;\n}\n",
			options:    &UnifiedSignaturesOptions{IgnoreDifferentlyNamedParameters: true},
			isJsx:      false,
			wantIds:    []string{"omittingSingleParameter"},
			wantSpans:  []string{"x: number"},
			wantStarts: []string{"This overload and the one on line 2 can be combined into one signature with an optional parameter."},
		},
		{
			sourceText: "interface I {\n  a1(): void;\n  a1(x: number): void;\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"omittingSingleParameter"},
			wantSpans:  []string{"x: number"},
			wantStarts: []string{"These overloads can be combined into one signature with an optional parameter."},
		},
		{
			sourceText: "interface I {\n  a3(): void;\n  a3(x: number, y?: number, ...z: number[]): void;\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"omittingRestParameter"},
			wantSpans:  []string{"...z: number[]"},
			wantStarts: []string{"These overloads can be combined into one signature with a rest parameter."},
		},
		{
			sourceText: "interface I {\n  b(): void;\n  b(...x: number[]): void;\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"omittingRestParameter"},
			wantSpans:  []string{"...x: number[]"},
			wantStarts: []string{"These overloads can be combined into one signature with a rest parameter."},
		},
		{
			sourceText: "interface I {\n  c(): void;\n  c(x?: number): void;\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"omittingSingleParameter"},
			wantSpans:  []string{"x?: number"},
			wantStarts: []string{"These overloads can be combined into one signature with an optional parameter."},
		},
		{
			sourceText: "interface I {\n  c2(x?: number): void;\n  c2(x?: string): void;\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x?: string"},
			wantStarts: []string{"These overloads can be combined into one signature taking `number | string`."},
		},
		{
			sourceText: "interface I {\n  d(x: number): void;\n  d(x: string): void;\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: string"},
			wantStarts: []string{"These overloads can be combined into one signature taking `number | string`."},
		},
		{
			sourceText: "type T = {\n  (): void;\n  (x: number): void;\n};\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"omittingSingleParameter"},
			wantSpans:  []string{"x: number"},
			wantStarts: []string{"These overloads can be combined into one signature with an optional parameter."},
		},
		{
			sourceText: "declare class Example {\n  #privateMethod(a: number): void;\n  #privateMethod(a: number, b?: string): void;\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"omittingSingleParameter"},
			wantSpans:  []string{"b?: string"},
			wantStarts: []string{"These overloads can be combined into one signature with an optional parameter."},
		},
		{
			sourceText: "declare class C {\n  constructor();\n  constructor(x: number);\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"omittingSingleParameter"},
			wantSpans:  []string{"x: number"},
			wantStarts: []string{"These overloads can be combined into one signature with an optional parameter."},
		},
		{
			sourceText: "interface I {\n  f(x: number);\n  f(x: string | boolean);\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: string | boolean"},
			wantStarts: []string{"These overloads can be combined into one signature taking `number | string | boolean`."},
		},
		{
			sourceText: "interface I {\n  f(x: number);\n  f(x: [string, boolean]);\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: [string, boolean]"},
			wantStarts: []string{"These overloads can be combined into one signature taking `number | [string, boolean]`."},
		},
		{
			sourceText: "interface Generic<T> {\n  y(x: T[]): void;\n  y(x: T): void;\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: T"},
			wantStarts: []string{"These overloads can be combined into one signature taking `T[] | T`."},
		},
		{
			sourceText: "function f<T>(x: T[]): void;\nfunction f<T>(x: T): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: T"},
			wantStarts: []string{"These overloads can be combined into one signature taking `T[] | T`."},
		},
		{
			sourceText: "function f<T extends number>(x: T[]): void;\nfunction f<T extends number>(x: T): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: T"},
			wantStarts: []string{"These overloads can be combined into one signature taking `T[] | T`."},
		},
		{
			sourceText: "abstract class Foo {\n  public abstract f(x: number): void;\n  public abstract f(x: string): void;\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: string"},
			wantStarts: []string{"These overloads can be combined into one signature taking `number | string`."},
		},
		{
			sourceText: "abstract class C {\n  a(b: string): void;\n  /**\n   * @deprecate\n   */\n  a(b: number): void;\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"b: number"},
			wantStarts: []string{"These overloads can be combined into one signature taking `string | number`."},
		},
		{
			sourceText: "interface Foo {\n  'f'(x: string): void;\n  'f'(x: number): void;\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: number"},
			wantStarts: []string{"These overloads can be combined into one signature taking `string | number`."},
		},
		{
			sourceText: "interface Foo {\n  new (x: string): Foo;\n  new (x: number): Foo;\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: number"},
			wantStarts: []string{"These overloads can be combined into one signature taking `string | number`."},
		},
		{
			sourceText: "enum Enum {\n  Func = 'function',\n}\n\ninterface IFoo {\n  [Enum.Func](x: string): void;\n  [Enum.Func](x: number): void;\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: number"},
			wantStarts: []string{"These overloads can be combined into one signature taking `string | number`."},
		},
		{
			sourceText: "export function foo(line: number): number;\nexport function foo(line: number, character?: number): number;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"omittingSingleParameter"},
			wantSpans:  []string{"character?: number"},
			wantStarts: []string{"These overloads can be combined into one signature with an optional parameter."},
		},
		{
			sourceText: "declare function foo(line: number): number;\nexport function foo(line: number, character?: number): number;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"omittingSingleParameter"},
			wantSpans:  []string{"character?: number"},
			wantStarts: []string{"These overloads can be combined into one signature with an optional parameter."},
		},
		{
			sourceText: "declare module 'foo' {\n  export default function (foo: number): string[];\n  export default function (foo: number, bar?: string): string[];\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"omittingSingleParameter"},
			wantSpans:  []string{"bar?: string"},
			wantStarts: []string{""},
		},
		{
			sourceText: "export default function (foo: number): string[];\nexport default function (foo: number, bar?: string): string[];\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"omittingSingleParameter"},
			wantSpans:  []string{"bar?: string"},
			wantStarts: []string{""},
		},
		{
			sourceText: "/**\n * @deprecate\n */\ndeclare function f(x: string): void;\ndeclare function f(x: number): void;\ndeclare function f(x: boolean): void;\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: boolean"},
			wantStarts: []string{"This overload and the one on line 5 can be combined into one signature taking `number | boolean`."},
		},
		{
			sourceText: "/**\n * @deprecate\n */\ndeclare function f(x: string): void;\n/**\n * @deprecate\n */\ndeclare function f(x: number): void;\ndeclare function f(x: boolean): void;\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: number"},
			wantStarts: []string{"This overload and the one on line 4 can be combined into one signature taking `string | number`."},
		},
		{
			sourceText: "declare function f(x: string): void;\n/**\n * @deprecate\n */\ndeclare function f(x: number): void;\n/**\n * @deprecate\n */\ndeclare function f(x: boolean): void;\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: boolean"},
			wantStarts: []string{"This overload and the one on line 5 can be combined into one signature taking `number | boolean`."},
		},
		{
			sourceText: "export function f(x: string): void;\n/**\n * @deprecate\n */\nexport function f(x: number): void;\n/**\n * @deprecate\n */\nexport function f(x: boolean): void;\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: boolean"},
			wantStarts: []string{"This overload and the one on line 5 can be combined into one signature taking `number | boolean`."},
		},
		{
			sourceText: "/**\n * This signature does something.\n */\n\n/**\n * This signature does something else.\n */\nfunction f(x: number): void;\n\n/**\n * This signature does something else.\n */\nfunction f(x: string): void;\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: string"},
			wantStarts: []string{"These overloads can be combined into one signature taking `number | string`."},
		},
		{
			sourceText: "interface I {\n  f(x: string): void;\n  /**\n   * @deprecate\n   */\n  f(x: number): void;\n  /**\n   * @deprecate\n   */\n  f(x: boolean): void;\n}\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: boolean"},
			wantStarts: []string{"This overload and the one on line 6 can be combined into one signature taking `number | boolean`."},
		},
		{
			sourceText: "// a line comment\ndeclare function f(x: number): unknown;\ndeclare function f(x: boolean): unknown;\n",
			options:    &UnifiedSignaturesOptions{IgnoreOverloadsWithDifferentJsDoc: true},
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: boolean"},
			wantStarts: []string{"These overloads can be combined into one signature taking `number | boolean`."},
		},
		{
			sourceText: "function f(this: {}, a: boolean): void;\nfunction f(this: {}, a: string): void;\nfunction f(this: {}, a: boolean | string): void {}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"a: string"},
			wantStarts: []string{""},
		},
		{
			sourceText: "function f(this: {}): void;\nfunction f(this: {}, a: string): void;\nfunction f(this: {}, a?: string): void {}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"omittingSingleParameter"},
			wantSpans:  []string{"a: string"},
			wantStarts: []string{""},
		},
		{
			sourceText: "function f(this: string): void;\nfunction f(this: number): void;\nfunction f(this: string | number): void {}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"this: number"},
			wantStarts: []string{"These overloads can be combined into one signature taking `string | number`."},
		},
		{
			sourceText: "function f(this: string, a: boolean): void;\nfunction f(this: number, a: boolean): void;\nfunction f(this: string | number, a: boolean): void {}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"this: number"},
			wantStarts: []string{"These overloads can be combined into one signature taking `string | number`."},
		},
		{
			sourceText: "function f(x: string | number): void;\nfunction f(x: number | boolean): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: number | boolean"},
			wantStarts: []string{"These overloads can be combined into one signature taking `string | number | boolean`."},
		},
		{
			sourceText: "function f(value): void;\nfunction f(value: string): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"value: string"},
			wantStarts: []string{"These overloads can be combined into one signature taking `string`."},
		},
		{
			sourceText: "function f(value: string): void;\nfunction f(value): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"value"},
			wantStarts: []string{"These overloads can be combined into one signature taking `string`."},
		},
		{
			sourceText: "type Alias = string;\nfunction f(x: Alias): void;\nfunction f(x: Alias | number): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: Alias | number"},
			wantStarts: []string{"These overloads can be combined into one signature taking `Alias | number`."},
		},
		{
			sourceText: "type Name = string;\nfunction f(x: Name): void;\nfunction f(x: string | Name): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: string | Name"},
			wantStarts: []string{"These overloads can be combined into one signature taking `Name | string`."},
		},
		{
			sourceText: "function f<T>(x: T): void;\nfunction f<T>(x: T | string): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: T | string"},
			wantStarts: []string{"These overloads can be combined into one signature taking `T | string`."},
		},
		{
			sourceText: "function f(x: 'a|b'): void;\nfunction f(x: 'a|b' | string): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: 'a|b' | string"},
			wantStarts: []string{"These overloads can be combined into one signature taking `'a|b' | string`."},
		},
		{
			sourceText: "function f(x: string /* first */ | number): void;\nfunction f(x: number | /* second */ string): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: number | /* second */ string"},
			wantStarts: []string{"These overloads can be combined into one signature taking `string | number`."},
		},
		{
			sourceText: "declare function fn(a: number): void;\ndeclare function fn(a: (/* before */ string /* after */)): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"a: (/* before */ string /* after */)"},
			wantStarts: []string{"These overloads can be combined into one signature taking `number | string`."},
		},
		{
			sourceText: "function f(x: string & { brand: true }): void;\nfunction f(x: number | string & /* brand */ { brand: true }): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: number | string & /* brand */ { brand: true }"},
			wantStarts: []string{"These overloads can be combined into one signature taking `string & { brand: true } | number | string & /* brand */ { brand: true }`."},
		},
		{
			sourceText: "function f(x: string & { brand: true }): void;\nfunction f(x: number | string & { brand: true }): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: number | string & { brand: true }"},
			wantStarts: []string{"These overloads can be combined into one signature taking `string & { brand: true } | number`."},
		},
		{
			sourceText: "function f<T, U>(x: () => void): void;\nfunction f<T, U>(x: T extends U ? string : number): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: T extends U ? string : number"},
			wantStarts: []string{"These overloads can be combined into one signature taking `(() => void) | (T extends U ? string : number)`."},
		},
		{
			sourceText: "interface Value {}\nfunction f(x: new () => Value): void;\nfunction f(x: string): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: string"},
			wantStarts: []string{"These overloads can be combined into one signature taking `(new () => Value) | string`."},
		},
		{
			sourceText: "interface I {\n  f(x: string | number): void;\n  f(x: number | boolean): void;\n  f(x: symbol): string;\n}\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: number | boolean"},
			wantStarts: []string{"This overload and the one on line 2 can be combined into one signature taking `string | number | boolean`."},
		},
		{
			sourceText: "function f(x: 'a' | 'b'): void;\nfunction f(x: 'b' | 'c'): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: 'b' | 'c'"},
			wantStarts: []string{"These overloads can be combined into one signature taking `'a' | 'b' | 'c'`."},
		},
		{
			sourceText: "function f(x: Array<string> | boolean): void;\nfunction f(x: Array<number> | boolean): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: Array<number> | boolean"},
			wantStarts: []string{"These overloads can be combined into one signature taking `Array<string> | boolean | Array<number>`."},
		},
		{
			sourceText: "function f(x: Promise | boolean): void;\nfunction f(x: Promise<string> | boolean): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: Promise<string> | boolean"},
			wantStarts: []string{"These overloads can be combined into one signature taking `Promise | boolean | Promise<string>`."},
		},
		{
			sourceText: "namespace Namespace {\n  export type Value = string;\n}\nfunction f(x: Namespace.Value): void;\nfunction f(x: Namespace.Value | string): void;\n",
			options:    nil,
			isJsx:      false,
			wantIds:    []string{"singleParameterDifference"},
			wantSpans:  []string{"x: Namespace.Value | string"},
			wantStarts: []string{"These overloads can be combined into one signature taking `Namespace.Value | string`."},
		},
	}
	for _, testCase := range cases {
		result := rule_testing.RunWithOptions(t, UnifiedSignatures,
			unifiedSignaturesFileFor(testCase.isJsx), testCase.sourceText, unifiedSignaturesOptionsAsDecoded(testCase.options))
		rule_testing.ExpectFindings(t, result, testCase.wantIds...)

		if len(result.Diagnostics) != len(testCase.wantSpans) {
			t.Fatalf("wanted %d findings, got %d, on %q",
				len(testCase.wantSpans), len(result.Diagnostics), testCase.sourceText)
		}
		written := asTheHarnessWroteIt(testCase.sourceText)
		for findingIndex, wantSpan := range testCase.wantSpans {
			reported := result.Diagnostics[findingIndex]
			gotSpan := written[reported.Range.Pos():reported.Range.End()]
			if gotSpan != wantSpan {
				t.Errorf("finding %d points at %q, wanted %q, on %q",
					findingIndex, gotSpan, wantSpan, testCase.sourceText)
			}
			// The WHOLE rendered message, not merely the id. Two halves carry real judgment: the
			// opening says which other overload this one joins, and for a union it names the type
			// the pair would become. Asserting only the prefix left the union text unchecked, which
			// a mutation sweep caught by deleting the parenthesising branch without a fixture
			// noticing.
			if want := testCase.wantStarts[findingIndex]; want != "" &&
				reported.Message.Description != want {
				t.Errorf("finding %d message is %q, wanted %q",
					findingIndex, reported.Message.Description, want)
			}
		}
	}
}

// TestUnifiedSignaturesOptionsArriveThroughTheDecoder starts from JSON and the registration's own
// Decode, which is the only path a configured option takes in a real run.
//
// Every other test here builds the options struct directly, so they all passed while the rule
// asserted a pointer and the decoder returned a value: both options read false for every
// configuration, and nothing in this file could see it. These rows pair each option's default with
// its configured value on the same source, so a rule that ignores the option fails one of the pair.
func TestUnifiedSignaturesOptionsArriveThroughTheDecoder(t *testing.T) {
	t.Parallel()

	var decode func(raw []byte) (any, error)
	for _, registration := range rule.Registered() {
		if registration.Rule.Name == UnifiedSignatures.Name {
			decode = registration.Decode
		}
	}
	if decode == nil {
		t.Fatalf("%s is not registered with a Decode", UnifiedSignatures.Name)
	}

	cases := []struct {
		name       string
		sourceText string
		option     string
	}{
		{
			name:       "ignoreDifferentlyNamedParameters",
			sourceText: "function f(a: number): void;\nfunction f(b: string): void;\nfunction f(a: number | string): void {}\n",
			option:     "ignoreDifferentlyNamedParameters",
		},
		{
			name:       "ignoreOverloadsWithDifferentJSDoc",
			sourceText: "/** @deprecated */\ndeclare function f(x: number): unknown;\ndeclare function f(x: boolean): unknown;\n",
			option:     "ignoreOverloadsWithDifferentJSDoc",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			for _, value := range []string{"false", "true"} {
				raw := `{"` + testCase.option + `": ` + value + `}`
				decoded, err := decode([]byte(raw))
				if err != nil {
					t.Fatalf("decoding %s: %v", raw, err)
				}
				result := rule_testing.RunWithOptions(t, UnifiedSignatures, "file.ts", testCase.sourceText, decoded)
				if value == "false" {
					// The baseline: without the option this source reports, so the true row below
					// can only go quiet because the option reached the rule.
					rule_testing.ExpectFindings(t, result, "singleParameterDifference")
				} else {
					rule_testing.ExpectClean(t, result)
				}
			}
		})
	}
}
