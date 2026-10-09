package typescript

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/corpus"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// noDeprecatedFinding is one expected report: which message, on what name, with what reason.
//
// The reason is asserted rather than only the id. 26 of upstream's cases carry a reason, and a rule
// that reported the right node with the wrong reason would pass an id-only assertion while telling
// every reader of the output something false.
type noDeprecatedFinding struct {
	messageId string
	name      string
	reason    string
}

// noDeprecatedCase is one row of the imported corpus.
type noDeprecatedCase struct {
	name       string
	sourceText string
	options    string

	// withDeprecatedFixture adds upstream's shared `deprecated.ts` beside the subject file. 51 cases
	// need it, and they are the ones that exercise the alias chain.
	withDeprecatedFixture bool

	// jsx compiles the subject as .tsx.
	jsx bool

	// environmentDependent marks a case whose answer depends on which @types packages are installed
	// rather than on the rule. Skipped with its reason named, never silently dropped.
	environmentDependent bool

	// withInstalledTypes links a real node_modules beside the fixture so an import resolves.
	//
	// Some cases take their deprecation FROM a @types package rather than from their own text, and
	// without the link they report nothing, which is indistinguishable from the rule declining them.
	// The link is enough for a NAMED import whose types the fixture reaches directly; it is not
	// enough for anything the harness's own tsconfig excludes, which writes `"types": []` after the
	// setup hook runs and so admits no ambient declarations at all. Those cases are marked
	// harnessBound instead.
	withInstalledTypes bool

	// harnessBound marks a case this test harness cannot express, as opposed to one the rule gets
	// wrong. Skipped with the reason named, and the arm it covers is proven separately.
	harnessBound string

	expected []noDeprecatedFinding
}

// noDeprecatedSharedFixture is upstream's `tests/fixtures/deprecated.ts`, verbatim.
//
// Copied rather than summarized because the three tag POSITIONS in it are the whole point: on the
// declaration, on an export specifier over an untagged declaration, and on one overload of several.
const noDeprecatedSharedFixture = "" +
	"/** @deprecated */\n" +
	"export class DeprecatedClass {\n" +
	"  /** @deprecated */\n" +
	"  foo: string = '';\n" +
	"}\n" +
	"/** @deprecated */\n" +
	"export const deprecatedVariable = 1;\n" +
	"/** @deprecated */\n" +
	"export function deprecatedFunction(): void {}\n" +
	"class NormalClass {}\n" +
	"const normalVariable = 1;\n" +
	"function normalFunction(): void;\n" +
	"function normalFunction(arg: string): void;\n" +
	"function normalFunction(arg?: string): void {}\n" +
	"function deprecatedFunctionWithOverloads(): void;\n" +
	"/** @deprecated */\n" +
	"function deprecatedFunctionWithOverloads(arg: string): void;\n" +
	"function deprecatedFunctionWithOverloads(arg?: string): void {}\n" +
	"export class ClassWithDeprecatedConstructor {\n" +
	"  constructor();\n" +
	"  /** @deprecated */\n" +
	"  constructor(arg: string);\n" +
	"  constructor(arg?: string) {}\n" +
	"}\n" +
	"export {\n" +
	"  /** @deprecated */\n" +
	"  NormalClass,\n" +
	"  /** @deprecated */\n" +
	"  normalVariable,\n" +
	"  /** @deprecated */\n" +
	"  normalFunction,\n" +
	"  deprecatedFunctionWithOverloads,\n" +
	"  /** @deprecated Reason */\n" +
	"  deprecatedFunctionWithOverloads as reexportedDeprecatedFunctionWithOverloads,\n" +
	"  /** @deprecated Reason */\n" +
	"  ClassWithDeprecatedConstructor as ReexportedClassWithDeprecatedConstructor,\n" +
	"};\n" +
	"\n" +
	"/** @deprecated Reason */\n" +
	"export type T = { a: string };\n" +
	"\n" +
	"export type U = { b: string };\n" +
	"\n" +
	"/** @deprecated */\n" +
	"export default {\n" +
	"  foo: 1,\n" +
	"};\n"

// runNoDeprecatedCases runs a block of rows.
func runNoDeprecatedCases(t *testing.T, cases []noDeprecatedCase) {
	t.Helper()
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.harnessBound != "" {
				t.Skip(testCase.harnessBound)
			}
			if testCase.environmentDependent {
				t.Skip("depends on the installed @types packages rather than on the rule; " +
					"upstream's declared expectation is recorded above")
			}

			subjectName := "/repository/source/Subject.ts"
			if testCase.jsx {
				subjectName = "/repository/source/Subject.tsx"
			}
			files := map[string]string{subjectName: testCase.sourceText}
			if testCase.withDeprecatedFixture {
				files["/repository/source/deprecated.ts"] = noDeprecatedSharedFixture
			}

			var decoded any
			if testCase.options != "" {
				parsed, err := DecodeNoDeprecatedOptions(json.RawMessage(testCase.options))
				if err != nil {
					t.Fatalf("the decoder refused %s: %v", testCase.options, err)
				}
				decoded = parsed
			}

			var result rule_testing.Result
			if testCase.withInstalledTypes {
				result = rule_testing.RunTypedFilesWithSetupAndOptions(
					t, NoDeprecated, files, subjectName, decoded, linkInstalledTypesForNoDeprecated(t))
			} else {
				result = rule_testing.RunTypedFilesWithOptions(t, NoDeprecated, files, subjectName, decoded)
			}

			if len(result.Diagnostics) != len(testCase.expected) {
				t.Fatalf("wanted %d findings, got %d: %s",
					len(testCase.expected), len(result.Diagnostics), describeNoDeprecated(result))
			}
			for index, want := range testCase.expected {
				got := result.Diagnostics[index]
				if got.Message.Id != want.messageId {
					t.Errorf("finding %d: wanted message %s, got %s",
						index, want.messageId, got.Message.Id)
				}
				// The name is interpolated into the description between backticks, and the reason
				// follows it. Asserting the rendered text rather than the pieces is what catches a
				// right-node-wrong-reason report.
				wantText := "`" + want.name + "` is deprecated."
				if want.reason != "" {
					wantText += " " + want.reason
				}
				if got.Message.Description != wantText {
					t.Errorf("finding %d: wanted %q, got %q", index, wantText, got.Message.Description)
				}
			}
		})
	}
}

// linkInstalledTypesForNoDeprecated links a real node_modules into the fixture directory.
//
// The three cases that need it take their deprecation from `@types/react` or `@types/node` rather
// than from their own text. Without the link the types do not resolve, the rule finds nothing, and
// the case reports zero findings, which is indistinguishable from the rule declining them. That is
// the failure mode this whole file is arranged against, so the link is what makes those three
// assertions mean anything.
//
// The packages are ahra's installed ones, so the case skips where the ahra corpus is not set, and the gate
// names it as not covered (#sycrdr6). Where it is set, a missing @types/react fails: the corpus is there and
// the case would otherwise read nothing.
func linkInstalledTypesForNoDeprecated(t *testing.T) func(string) {
	t.Helper()
	return func(directory string) {
		source := corpus.Ahra.Path(t, "node_modules")
		if _, err := os.Stat(filepath.Join(source, "@types", "react")); err != nil {
			t.Fatalf("this case reads its deprecation from @types/react, which the ahra corpus at %s does "+
				"not have installed: %v", source, err)
		}
		if err := os.Symlink(source, filepath.Join(directory, "node_modules")); err != nil {
			t.Fatalf("linking node_modules into the fixture: %v", err)
		}
	}
}

// describeNoDeprecated renders a result for a failure message.
func describeNoDeprecated(result rule_testing.Result) string {
	if len(result.Diagnostics) == 0 {
		return "(no findings)"
	}
	var builder strings.Builder
	for index, diagnostic := range result.Diagnostics {
		if index > 0 {
			builder.WriteString("; ")
		}
		builder.WriteString(diagnostic.Message.Id)
		builder.WriteString(": ")
		builder.WriteString(diagnostic.Message.Description)
	}
	return builder.String()
}

// TestNoDeprecatedAgainstUpstreamOrdinaryCases is the single-file part of upstream's corpus.
//
// Every expectation here was produced by replaying the case through the installed 8.67.0 build with
// a real type checker, one program per case so no two cases share an interned type, then checked
// against upstream's own declared expectations. 259 of the 264 agreed; the five that did not are
// marked environmentDependent where they appear and carry upstream's declaration instead.
func TestNoDeprecatedAgainstUpstreamOrdinaryCases(t *testing.T) {
	t.Parallel()

	cases := []noDeprecatedCase{
		{
			name:       "valid:0",
			sourceText: "/** @deprecated */ var a;",
			expected:   nil,
		},
		{
			name:       "valid:1",
			sourceText: "/** @deprecated */ var a = 1;",
			expected:   nil,
		},
		{
			name:       "valid:2",
			sourceText: "/** @deprecated */ let a;",
			expected:   nil,
		},
		{
			name:       "valid:3",
			sourceText: "/** @deprecated */ let a = 1;",
			expected:   nil,
		},
		{
			name:       "valid:4",
			sourceText: "/** @deprecated */ const a = 1;",
			expected:   nil,
		},
		{
			name:       "valid:5",
			sourceText: "/** @deprecated */ declare var a: number;",
			expected:   nil,
		},
		{
			name:       "valid:6",
			sourceText: "/** @deprecated */ declare let a: number;",
			expected:   nil,
		},
		{
			name:       "valid:7",
			sourceText: "/** @deprecated */ declare const a: number;",
			expected:   nil,
		},
		{
			name:       "valid:8",
			sourceText: "/** @deprecated */ export var a = 1;",
			expected:   nil,
		},
		{
			name:       "valid:9",
			sourceText: "/** @deprecated */ export let a = 1;",
			expected:   nil,
		},
		{
			name:       "valid:10",
			sourceText: "/** @deprecated */ export const a = 1;",
			expected:   nil,
		},
		{
			name:       "valid:11",
			sourceText: "const [/** @deprecated */ a] = [b];",
			expected:   nil,
		},
		{
			name:       "valid:12",
			sourceText: "const [/** @deprecated */ a] = b;",
			expected:   nil,
		},
		{
			name: "valid:13",
			sourceText: "\n" +
				"const a = {\n" +
				"  b: 1,\n" +
				"  /** @deprecated */ c: 2,\n" +
				"};\n" +
				"\n" +
				"a.b;\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:14",
			sourceText: "\n" +
				"const a = {\n" +
				"  b: 1,\n" +
				"  /** @deprecated */ c: 2,\n" +
				"};\n" +
				"\n" +
				"a['b'];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:15",
			sourceText: "\n" +
				"const a = {\n" +
				"  b: 1,\n" +
				"  /** @deprecated */ c: 2,\n" +
				"};\n" +
				"\n" +
				"a['b' + 'c'];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:16",
			sourceText: "\n" +
				"const a = {\n" +
				"  b: 1,\n" +
				"  /** @deprecated */ c: 2,\n" +
				"};\n" +
				"\n" +
				"const key = 'b';\n" +
				"\n" +
				"a[key];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:17",
			sourceText: "\n" +
				"const a = {\n" +
				"  b: 1,\n" +
				"  /** @deprecated */ c: 2,\n" +
				"};\n" +
				"\n" +
				"a?.b;\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:18",
			sourceText: "\n" +
				"const a = {\n" +
				"  b: 1,\n" +
				"  /** @deprecated */ c: 2,\n" +
				"};\n" +
				"\n" +
				"a?.['b'];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:19",
			sourceText: "\n" +
				"declare const a: {\n" +
				"  b: 1;\n" +
				"  /** @deprecated */ c: 2;\n" +
				"};\n" +
				"\n" +
				"a.b;\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:20",
			sourceText: "\n" +
				"declare const a: {\n" +
				"  b: 1;\n" +
				"  /** @deprecated */ c: 2;\n" +
				"};\n" +
				"\n" +
				"a['b'];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:21",
			sourceText: "\n" +
				"declare const a: {\n" +
				"  b: 1;\n" +
				"  /** @deprecated */ c: 2;\n" +
				"};\n" +
				"\n" +
				"a[`${'b'}`];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:22",
			sourceText: "\n" +
				"declare const a: {\n" +
				"  b: 1;\n" +
				"  /** @deprecated */ c: 2;\n" +
				"};\n" +
				"\n" +
				"const key = 'b';\n" +
				"\n" +
				"a[`${key}`];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:23",
			sourceText: "\n" +
				"declare const a: {\n" +
				"  /** @deprecated */ c: 1;\n" +
				"  cc: 2;\n" +
				"};\n" +
				"\n" +
				"const key = 'c';\n" +
				"\n" +
				"a[`${key + key}`];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:24",
			sourceText: "\n" +
				"declare const a: {\n" +
				"  /** @deprecated */ c: 1;\n" +
				"  cc: 2;\n" +
				"};\n" +
				"\n" +
				"const key = 'c';\n" +
				"\n" +
				"a[`${key}${key}`];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:25",
			sourceText: "\n" +
				"class A {\n" +
				"  b: 1;\n" +
				"  /** @deprecated */ c: 2;\n" +
				"}\n" +
				"\n" +
				"new A().b;\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:26",
			sourceText: "\n" +
				"class A {\n" +
				"  b: 1;\n" +
				"  /** @deprecated */ c: 2;\n" +
				"}\n" +
				"\n" +
				"new A()['b'];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:27",
			sourceText: "\n" +
				"class A {\n" +
				"  b: 1;\n" +
				"  /** @deprecated */ c: 2;\n" +
				"}\n" +
				"const key = 'b';\n" +
				"\n" +
				"new A()[b];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:28",
			sourceText: "\n" +
				"class A {\n" +
				"  c: 1;\n" +
				"}\n" +
				"class B {\n" +
				"  /** @deprecated */ c: 2;\n" +
				"}\n" +
				"\n" +
				"new A()['c'];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:29",
			sourceText: "\n" +
				"class A {\n" +
				"  b: () => {};\n" +
				"  /** @deprecated */ c: () => {};\n" +
				"}\n" +
				"\n" +
				"new A()['b']();\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:30",
			sourceText: "\n" +
				"class A {\n" +
				"  accessor b: 1;\n" +
				"  /** @deprecated */ accessor c: 2;\n" +
				"}\n" +
				"\n" +
				"new A().b;\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:31",
			sourceText: "\n" +
				"class A {\n" +
				"  accessor b: 1;\n" +
				"  /** @deprecated */ accessor c: 2;\n" +
				"}\n" +
				"\n" +
				"new A()['b'];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:32",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  static b: string;\n" +
				"  static c: string;\n" +
				"}\n" +
				"\n" +
				"A.c;\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:33",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  static b: string;\n" +
				"  static c: string;\n" +
				"}\n" +
				"\n" +
				"A['c'];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:34",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  static accessor b: string;\n" +
				"  static accessor c: string;\n" +
				"}\n" +
				"\n" +
				"A.c;\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:35",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  static accessor b: string;\n" +
				"  static accessor c: string;\n" +
				"}\n" +
				"\n" +
				"A['c'];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:36",
			sourceText: "\n" +
				"namespace A {\n" +
				"  /** @deprecated */\n" +
				"  export const b = '';\n" +
				"  export const c = '';\n" +
				"}\n" +
				"\n" +
				"A.c;\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:37",
			sourceText: "\n" +
				"namespace A {\n" +
				"  /** @deprecated */\n" +
				"  export const b = '';\n" +
				"  export const c = '';\n" +
				"}\n" +
				"\n" +
				"A['c'];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:38",
			sourceText: "\n" +
				"enum A {\n" +
				"  /** @deprecated */\n" +
				"  b = 'b',\n" +
				"  c = 'c',\n" +
				"}\n" +
				"\n" +
				"A.c;\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:39",
			sourceText: "\n" +
				"enum A {\n" +
				"  /** @deprecated */\n" +
				"  b = 'b',\n" +
				"  c = 'c',\n" +
				"}\n" +
				"\n" +
				"A['c'];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:40",
			sourceText: "\n" +
				"function a(value: 'b' | undefined): void;\n" +
				"/** @deprecated */\n" +
				"function a(value: 'c' | undefined): void;\n" +
				"function a(value: string | undefined): void {\n" +
				"  // ...\n" +
				"}\n" +
				"\n" +
				"a('b');\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:41",
			sourceText: "\n" +
				"function a(value: 'b' | undefined): void;\n" +
				"/** @deprecated */\n" +
				"function a(value: 'c' | undefined): void;\n" +
				"function a(value: string | undefined): void {\n" +
				"  // ...\n" +
				"}\n" +
				"\n" +
				"export default a('b');\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:42",
			sourceText: "\n" +
				"function notDeprecated(): object {\n" +
				"  return {};\n" +
				"}\n" +
				"\n" +
				"export default notDeprecated();\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:47",
			sourceText: "\n" +
				"class A {\n" +
				"  a(value: 'b'): void;\n" +
				"  /** @deprecated */\n" +
				"  a(value: 'c'): void;\n" +
				"}\n" +
				"declare const foo: A;\n" +
				"foo.a('b');\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:48",
			sourceText: "\n" +
				"const A = class {\n" +
				"  /** @deprecated */\n" +
				"  constructor();\n" +
				"  constructor(arg: string);\n" +
				"  constructor(arg?: string) {}\n" +
				"};\n" +
				"\n" +
				"new A('a');\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:49",
			sourceText: "\n" +
				"type A = {\n" +
				"  (value: 'b'): void;\n" +
				"  /** @deprecated */\n" +
				"  (value: 'c'): void;\n" +
				"};\n" +
				"declare const foo: A;\n" +
				"foo('b');\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:50",
			sourceText: "\n" +
				"declare const a: {\n" +
				"  new (value: 'b'): void;\n" +
				"  /** @deprecated */\n" +
				"  new (value: 'c'): void;\n" +
				"};\n" +
				"new a('b');\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:51",
			sourceText: "\n" +
				"namespace assert {\n" +
				"  export function fail(message?: string | Error): never;\n" +
				"  /** @deprecated since v10.0.0 - use fail([message]) or other assert functions instead. */\n" +
				"  export function fail(actual: unknown, expected: unknown): never;\n" +
				"}\n" +
				"\n" +
				"assert.fail('');\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:53",
			sourceText: "\n" +
				"declare module 'deprecations' {\n" +
				"  /** @deprecated */\n" +
				"  export const value = true;\n" +
				"}\n" +
				"\n" +
				"import { value } from 'deprecations';\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:54",
			sourceText: "\n" +
				"/** @deprecated Use ts directly. */\n" +
				"export * as ts from 'typescript';\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:55",
			sourceText: "\n" +
				"export {\n" +
				"  /** @deprecated Use ts directly. */\n" +
				"  default as ts,\n" +
				"} from 'typescript';\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:58",
			sourceText: "\n" +
				"namespace A {\n" +
				"  /** @deprecated */\n" +
				"  export type B = string;\n" +
				"  export type C = string;\n" +
				"  export type D = string;\n" +
				"}\n" +
				"\n" +
				"export type D = A.C | A.D;\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:59",
			sourceText: "\n" +
				"interface Props {\n" +
				"  anchor: 'foo';\n" +
				"}\n" +
				"declare const x: Props;\n" +
				"const { anchor = '' } = x;\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:60",
			sourceText: "\n" +
				"namespace Foo {}\n" +
				"\n" +
				"/**\n" +
				" * @deprecated\n" +
				" */\n" +
				"export import Bar = Foo;\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:62",
			sourceText: "\n" +
				"interface Props {\n" +
				"  anchor: 'foo';\n" +
				"}\n" +
				"declare const x: { bar: Props };\n" +
				"const {\n" +
				"  bar: { anchor = '' },\n" +
				"} = x;\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:63",
			sourceText: "\n" +
				"interface Props {\n" +
				"  anchor: 'foo';\n" +
				"}\n" +
				"declare const x: [item: Props];\n" +
				"const [{ anchor = 'bar' }] = x;\n" +
				"    ",
			expected: nil,
		},
		{
			name:       "valid:64",
			sourceText: "function fn(/** @deprecated */ foo = 4) {}",
			expected:   nil,
		},
		{
			name:       "valid:66",
			sourceText: "call();",
			expected:   nil,
		},
		{
			name: "valid:67",
			sourceText: "\n" +
				"class Foo implements Foo {\n" +
				"  get bar(): number {\n" +
				"    return 42;\n" +
				"  }\n" +
				"\n" +
				"  baz(): number {\n" +
				"    return this.bar;\n" +
				"  }\n" +
				"}\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:73",
			sourceText: "\n" +
				"export {\n" +
				"  /** @deprecated */\n" +
				"  foo,\n" +
				"};\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:75",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"declare class A {}\n" +
				"\n" +
				"new A();\n" +
				"      ",
			options:  "{\"allow\":[{\"from\":\"file\",\"name\":\"A\"}]}",
			expected: nil,
		},
		{
			name: "valid:76",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"const deprecatedValue = 45;\n" +
				"const bar = deprecatedValue;\n" +
				"      ",
			options:  "{\"allow\":[{\"from\":\"file\",\"name\":\"deprecatedValue\"}]}",
			expected: nil,
		},
		{
			name: "valid:77",
			sourceText: "\n" +
				"class MyClass {\n" +
				"  /** @deprecated */\n" +
				"  #privateProp = 42;\n" +
				"  value = this.#privateProp;\n" +
				"}\n" +
				"      ",
			options:  "{\"allow\":[{\"from\":\"file\",\"name\":\"privateProp\"}]}",
			expected: nil,
		},
		{
			name: "valid:78",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"const deprecatedValue = 45;\n" +
				"const bar = deprecatedValue;\n" +
				"      ",
			options:  "{\"allow\":[\"deprecatedValue\"]}",
			expected: nil,
		},
		{
			name: "valid:80",
			sourceText: "\n" +
				"const { exists } = import('fs');\n" +
				"exists('/foo');\n" +
				"      ",
			options:  "{\"allow\":[{\"from\":\"package\",\"name\":\"exists\",\"package\":\"fs\"}]}",
			expected: nil,
		},
		{
			name: "valid:81",
			sourceText: "\n" +
				"declare const test: string;\n" +
				"const bar = { test };\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:82",
			sourceText: "\n" +
				"const a = {\n" +
				"  /** @deprecated */\n" +
				"  b: 'string',\n" +
				"};\n" +
				"\n" +
				"const complex = Symbol() as any;\n" +
				"const c = a[complex];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:83",
			sourceText: "\n" +
				"const a = {\n" +
				"  b: 'string',\n" +
				"};\n" +
				"\n" +
				"const c = a['b'];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:84",
			sourceText: "\n" +
				"const a = {\n" +
				"  /** @deprecated */\n" +
				"  b: 'string',\n" +
				"};\n" +
				"\n" +
				"const key = {};\n" +
				"const c = a[key as any];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:85",
			sourceText: "\n" +
				"const a = {\n" +
				"  /** @deprecated */\n" +
				"  b: 'string',\n" +
				"};\n" +
				"\n" +
				"const key = Symbol();\n" +
				"const c = a[key as any];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:86",
			sourceText: "\n" +
				"const a = {\n" +
				"  /** @deprecated */\n" +
				"  b: 'string',\n" +
				"};\n" +
				"\n" +
				"const key = undefined;\n" +
				"const c = a[key as any];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:87",
			sourceText: "\n" +
				"const a = {\n" +
				"  /** @deprecated */\n" +
				"  b: 'string',\n" +
				"};\n" +
				"\n" +
				"const c = a['nonExistentProperty'];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:88",
			sourceText: "\n" +
				"const a = {\n" +
				"  /** @deprecated */\n" +
				"  b: 'string',\n" +
				"};\n" +
				"\n" +
				"function getKey() {\n" +
				"  return 'c';\n" +
				"}\n" +
				"\n" +
				"const c = a[getKey()];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:89",
			sourceText: "\n" +
				"const a = {\n" +
				"  /** @deprecated */\n" +
				"  b: 'string',\n" +
				"};\n" +
				"\n" +
				"const key = {};\n" +
				"const c = a[key];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:90",
			sourceText: "\n" +
				"const stringObj = new String('b');\n" +
				"const a = {\n" +
				"  /** @deprecated */\n" +
				"  b: 'string',\n" +
				"};\n" +
				"const c = a[stringObj];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:91",
			sourceText: "\n" +
				"const a = {\n" +
				"  /** @deprecated */\n" +
				"  b: 'string',\n" +
				"};\n" +
				"\n" +
				"const key = Symbol('key');\n" +
				"const c = a[key];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:92",
			sourceText: "\n" +
				"const a = {\n" +
				"  /** @deprecated */\n" +
				"  b: 'string',\n" +
				"};\n" +
				"\n" +
				"const key = null;\n" +
				"const c = a[key as any];\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:93",
			sourceText: "\n" +
				"interface Foo {\n" +
				"  /** @deprecated */\n" +
				"  deprecatedProperty: string;\n" +
				"  notDeprecatedProperty: string;\n" +
				"}\n" +
				"\n" +
				"declare const a: Foo;\n" +
				"const { notDeprecatedProperty: deprecatedProperty } = a;\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:94",
			sourceText: "\n" +
				"interface Foo {\n" +
				"  /** @deprecated */\n" +
				"  deprecatedProperty: string;\n" +
				"  notDeprecatedProperty: string;\n" +
				"}\n" +
				"\n" +
				"declare const a: { foo: Foo };\n" +
				"const {\n" +
				"  foo: { notDeprecatedProperty: deprecatedProperty },\n" +
				"} = a;\n" +
				"    ",
			expected: nil,
		},
		{
			name: "invalid:1",
			sourceText: "\n" +
				"/** @deprecated */ var a = undefined;\n" +
				"a;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:2",
			sourceText: "\n" +
				"/** @deprecated */ export var a = undefined;\n" +
				"a;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:3",
			sourceText: "\n" +
				"/** @deprecated */ let a = undefined;\n" +
				"a;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:4",
			sourceText: "\n" +
				"/** @deprecated */ export let a = undefined;\n" +
				"a;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:5",
			sourceText: "\n" +
				"/** @deprecated */ let aLongName = undefined;\n" +
				"aLongName;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "aLongName", reason: ""},
			},
		},
		{
			name: "invalid:6",
			sourceText: "\n" +
				"/** @deprecated */ const a = { b: 1 };\n" +
				"const c = a;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:7",
			sourceText: "\n" +
				"/** @deprecated Reason. */ const a = { b: 1 };\n" +
				"const c = a;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "a", reason: "Reason."},
			},
		},
		{
			name: "invalid:8",
			sourceText: "\n" +
				"/** @deprecated */ const a = { b: 1 };\n" +
				"const { c = a } = {};\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:9",
			sourceText: "\n" +
				"/** @deprecated */ const a = { b: 1 };\n" +
				"const [c = a] = [];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:10",
			sourceText: "\n" +
				"/** @deprecated */ const a = { b: 1 };\n" +
				"a;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:11",
			sourceText: "\n" +
				"/** @deprecated */ const a = 'foo';\n" +
				"import(`./path/${a}.js`);\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:12",
			sourceText: "\n" +
				"declare function log(...args: unknown): void;\n" +
				"\n" +
				"/** @deprecated */ const a = { b: 1 };\n" +
				"\n" +
				"log(a);\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:13",
			sourceText: "\n" +
				"/** @deprecated */ const a = { b: 1 };\n" +
				"a.b;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:14",
			sourceText: "\n" +
				"/** @deprecated */ const a = { b: 1 };\n" +
				"a['b'];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:15",
			sourceText: "\n" +
				"/** @deprecated */ const a = { b: 1 };\n" +
				"a?.b;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:16",
			sourceText: "\n" +
				"/** @deprecated */ const a = { b: 1 };\n" +
				"a?.['b'];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:17",
			sourceText: "\n" +
				"/** @deprecated */ const a = { b: { c: 1 } };\n" +
				"a.b.c;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:18",
			sourceText: "\n" +
				"/** @deprecated */ const a = { b: { c: 1 } };\n" +
				"a.b?.c;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:19",
			sourceText: "\n" +
				"/** @deprecated */ const a = { b: { c: 1 } };\n" +
				"a?.b?.c;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:20",
			sourceText: "\n" +
				"/** @deprecated */ const a = { b: { c: 1 } };\n" +
				"a?.['b']?.['c'];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:21",
			sourceText: "\n" +
				"const a = {\n" +
				"  /** @deprecated */ b: { c: 1 },\n" +
				"};\n" +
				"a.b.c;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:22",
			sourceText: "\n" +
				"const a = {\n" +
				"  /** @deprecated */ b: 1,\n" +
				"};\n" +
				"const key = 'b';\n" +
				"a[key];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:23",
			sourceText: "\n" +
				"const a = {\n" +
				"  /** @deprecated */ b: { c: 1 },\n" +
				"};\n" +
				"a['b']['c'];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:24",
			sourceText: "\n" +
				"declare const a: {\n" +
				"  /** @deprecated */ b: { c: 1 };\n" +
				"};\n" +
				"a.b.c;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:25",
			sourceText: "\n" +
				"/** @deprecated */ const a = { b: 1 };\n" +
				"const c = a.b;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:26",
			sourceText: "\n" +
				"/** @deprecated */ const a = { b: 1 };\n" +
				"const { c } = a.b;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:27",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"declare const test: string;\n" +
				"const myObj = {\n" +
				"  prop: test,\n" +
				"  deep: {\n" +
				"    prop: test,\n" +
				"  },\n" +
				"};\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "test", reason: ""},
				{messageId: "deprecated", name: "test", reason: ""},
			},
		},
		{
			name: "invalid:28",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"declare const test: string;\n" +
				"const bar = {\n" +
				"  test,\n" +
				"};\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "test", reason: ""},
			},
		},
		{
			name: "invalid:29",
			sourceText: "\n" +
				"/** @deprecated */ const a = { b: 1 };\n" +
				"const { c = 'd' } = a.b;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:30",
			sourceText: "\n" +
				"/** @deprecated */ const a = { b: 1 };\n" +
				"const { c: d } = a.b;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:31",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"declare const a: string[];\n" +
				"const [b] = [a];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:32",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"class A {}\n" +
				"\n" +
				"new A();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "A", reason: ""},
			},
		},
		{
			name: "invalid:33",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"export class A {}\n" +
				"\n" +
				"new A();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "A", reason: ""},
			},
		},
		{
			name: "invalid:34",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"const A = class {};\n" +
				"\n" +
				"new A();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "A", reason: ""},
			},
		},
		{
			name: "invalid:35",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"declare class A {}\n" +
				"\n" +
				"new A();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "A", reason: ""},
			},
		},
		{
			name: "invalid:36",
			sourceText: "\n" +
				"const A = class {\n" +
				"  /** @deprecated */\n" +
				"  constructor() {}\n" +
				"};\n" +
				"\n" +
				"new A();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "A", reason: ""},
			},
		},
		{
			name: "invalid:37",
			sourceText: "\n" +
				"const A = class {\n" +
				"  /** @deprecated */\n" +
				"  constructor();\n" +
				"  constructor(arg: string);\n" +
				"  constructor(arg?: string) {}\n" +
				"};\n" +
				"\n" +
				"new A();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "A", reason: ""},
			},
		},
		{
			name: "invalid:38",
			sourceText: "\n" +
				"declare const A: {\n" +
				"  /** @deprecated */\n" +
				"  new (): string;\n" +
				"};\n" +
				"\n" +
				"new A();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "A", reason: ""},
			},
		},
		{
			name: "invalid:39",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"declare class A {\n" +
				"  constructor();\n" +
				"}\n" +
				"\n" +
				"new A();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "A", reason: ""},
			},
		},
		{
			name: "invalid:40",
			sourceText: "\n" +
				"class A {\n" +
				"  /** @deprecated */\n" +
				"  b: string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"const { b } = a;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:41",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  b(): string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"a.b;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:42",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  b(): string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"const key = 'b';\n" +
				"\n" +
				"a[key];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:43",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  b(): string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"const key = 'b';\n" +
				"\n" +
				"a[key]();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:44",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  b(): string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"a.b();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:45",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  b: () => string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"a.b;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:46",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  b: () => string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"a['b'];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:47",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  b(): string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"const key = 'b';\n" +
				"\n" +
				"a[`${key}`];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:48",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  computed(): string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"const k1 = 'comp';\n" +
				"const k2 = 'uted';\n" +
				"\n" +
				"a[`${k1}${k2}`];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "computed", reason: ""},
			},
		},
		{
			name: "invalid:49",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  b(): string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"const c = `${a.b}`;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:50",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  b: () => string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"a.b();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:51",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  b: () => string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"a['b']();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:52",
			sourceText: "\n" +
				"interface A {\n" +
				"  /** @deprecated */\n" +
				"  b: () => string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"a.b();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:53",
			sourceText: "\n" +
				"interface A {\n" +
				"  /** @deprecated */\n" +
				"  b: () => string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"const key = 'b';\n" +
				"\n" +
				"a[key];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:54",
			sourceText: "\n" +
				"interface A {\n" +
				"  /** @deprecated */\n" +
				"  b: () => string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"const key = 'b';\n" +
				"\n" +
				"a[key]();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:55",
			sourceText: "\n" +
				"interface A {\n" +
				"  /** @deprecated */\n" +
				"  b: () => string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"a['b']();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:56",
			sourceText: "\n" +
				"class A {\n" +
				"  /** @deprecated */\n" +
				"  b(): string {\n" +
				"    return '';\n" +
				"  }\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"a.b();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:57",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated Use b(value). */\n" +
				"  b(): string;\n" +
				"  b(value: string): string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"a.b();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "b", reason: "Use b(value)."},
			},
		},
		{
			name: "invalid:58",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  static b: string;\n" +
				"}\n" +
				"\n" +
				"A.b;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:59",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  static b: string;\n" +
				"}\n" +
				"\n" +
				"A['b'];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:60",
			sourceText: "\n" +
				"declare const a: {\n" +
				"  /** @deprecated */\n" +
				"  b: string;\n" +
				"};\n" +
				"\n" +
				"a.b;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:61",
			sourceText: "\n" +
				"declare const a: {\n" +
				"  /** @deprecated */\n" +
				"  b: string;\n" +
				"};\n" +
				"\n" +
				"a['b'];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:62",
			sourceText: "\n" +
				"interface A {\n" +
				"  /** @deprecated */\n" +
				"  b: string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"a.b;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:63",
			sourceText: "\n" +
				"export interface A {\n" +
				"  /** @deprecated */\n" +
				"  b: string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"a.b;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:64",
			sourceText: "\n" +
				"interface A {\n" +
				"  /** @deprecated */\n" +
				"  b: string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"const { b } = a;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:65",
			sourceText: "\n" +
				"type A = {\n" +
				"  /** @deprecated */\n" +
				"  b: string;\n" +
				"};\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"const { b } = a;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:66",
			sourceText: "\n" +
				"export type A = {\n" +
				"  /** @deprecated */\n" +
				"  b: string;\n" +
				"};\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"const { b } = a;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:67",
			sourceText: "\n" +
				"type A = () => {\n" +
				"  /** @deprecated */\n" +
				"  b: string;\n" +
				"};\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"const { b } = a();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:68",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"type A = string[];\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"const [b] = a;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "A", reason: ""},
			},
		},
		{
			name: "invalid:69",
			sourceText: "\n" +
				"namespace A {\n" +
				"  /** @deprecated */\n" +
				"  export const b = '';\n" +
				"}\n" +
				"\n" +
				"A.b;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:70",
			sourceText: "\n" +
				"namespace A {\n" +
				"  /** @deprecated */\n" +
				"  export const b = '';\n" +
				"}\n" +
				"\n" +
				"A['b'];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:71",
			sourceText: "\n" +
				"export namespace A {\n" +
				"  /** @deprecated */\n" +
				"  export const b = '';\n" +
				"}\n" +
				"\n" +
				"A.b;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:72",
			sourceText: "\n" +
				"namespace A {\n" +
				"  /** @deprecated */\n" +
				"  export function b() {}\n" +
				"}\n" +
				"\n" +
				"A.b();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:73",
			sourceText: "\n" +
				"namespace A {\n" +
				"  /** @deprecated */\n" +
				"  export function b() {}\n" +
				"}\n" +
				"\n" +
				"A['b']();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:74",
			sourceText: "\n" +
				"namespace assert {\n" +
				"  export function fail(message?: string | Error): never;\n" +
				"  /** @deprecated since v10.0.0 - use fail([message]) or other assert functions instead. */\n" +
				"  export function fail(actual: unknown, expected: unknown): never;\n" +
				"}\n" +
				"\n" +
				"assert.fail({}, {});\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "fail", reason: "since v10.0.0 - use fail([message]) or other assert functions instead."},
			},
		},
		{
			name: "invalid:76",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"enum A {\n" +
				"  a,\n" +
				"}\n" +
				"\n" +
				"A.a;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "A", reason: ""},
			},
		},
		{
			name: "invalid:77",
			sourceText: "\n" +
				"enum A {\n" +
				"  /** @deprecated */\n" +
				"  a,\n" +
				"}\n" +
				"\n" +
				"A.a;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:78",
			sourceText: "\n" +
				"enum A {\n" +
				"  /** @deprecated */\n" +
				"  a,\n" +
				"}\n" +
				"\n" +
				"const key = 'a';\n" +
				"\n" +
				"A[key];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:79",
			sourceText: "\n" +
				"enum A {\n" +
				"  /** @deprecated */\n" +
				"  a,\n" +
				"}\n" +
				"\n" +
				"A['a'];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:80",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"function a() {}\n" +
				"\n" +
				"a();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:81",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"function a(): void;\n" +
				"function a() {}\n" +
				"\n" +
				"a();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:82",
			sourceText: "\n" +
				"function a(): void;\n" +
				"/** @deprecated */\n" +
				"function a(value: string): void;\n" +
				"function a(value?: string) {}\n" +
				"\n" +
				"a('');\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:83",
			sourceText: "\n" +
				"type A = {\n" +
				"  (value: 'b'): void;\n" +
				"  /** @deprecated */\n" +
				"  (value: 'c'): void;\n" +
				"};\n" +
				"declare const foo: A;\n" +
				"foo('c');\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "foo", reason: ""},
			},
		},
		{
			name: "invalid:84",
			sourceText: "\n" +
				"function a(\n" +
				"  /** @deprecated */\n" +
				"  b?: boolean,\n" +
				") {\n" +
				"  return b;\n" +
				"}\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:85",
			sourceText: "\n" +
				"export function isTypeFlagSet(\n" +
				"  type: ts.Type,\n" +
				"  flagsToCheck: ts.TypeFlags,\n" +
				"  /** @deprecated This param is not used and will be removed in the future. */\n" +
				"  isReceiver?: boolean,\n" +
				"): boolean {\n" +
				"  const flags = getTypeFlags(type);\n" +
				"\n" +
				"  if (isReceiver && flags & ANY_OR_UNKNOWN) {\n" +
				"    return true;\n" +
				"  }\n" +
				"\n" +
				"  return (flags & flagsToCheck) !== 0;\n" +
				"}\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "isReceiver", reason: "This param is not used and will be removed in the future."},
			},
		},
		{
			name: "invalid:86",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"declare function a(...args: unknown[]): string;\n" +
				"\n" +
				"a``;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:91",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"export type A = string;\n" +
				"export type B = string;\n" +
				"export type C = string;\n" +
				"\n" +
				"export type D = A | B | C;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "A", reason: ""},
			},
		},
		{
			name: "invalid:92",
			sourceText: "\n" +
				"namespace A {\n" +
				"  /** @deprecated */\n" +
				"  export type B = string;\n" +
				"  export type C = string;\n" +
				"  export type D = string;\n" +
				"}\n" +
				"\n" +
				"export type D = A.B | A.C | A.D;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "B", reason: ""},
			},
		},
		{
			name: "invalid:93",
			sourceText: "\n" +
				"interface Props {\n" +
				"  /** @deprecated */\n" +
				"  anchor: 'foo';\n" +
				"}\n" +
				"declare const x: Props;\n" +
				"const { anchor = '' } = x;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "anchor", reason: ""},
			},
		},
		{
			name: "invalid:94",
			sourceText: "\n" +
				"interface Props {\n" +
				"  /** @deprecated */\n" +
				"  anchor: 'foo';\n" +
				"}\n" +
				"declare const x: { bar: Props };\n" +
				"const {\n" +
				"  bar: { anchor = '' },\n" +
				"} = x;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "anchor", reason: ""},
			},
		},
		{
			name: "invalid:95",
			sourceText: "\n" +
				"interface Props {\n" +
				"  /** @deprecated */\n" +
				"  anchor: 'foo';\n" +
				"}\n" +
				"declare const x: [item: Props];\n" +
				"const [{ anchor = 'bar' }] = x;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "anchor", reason: ""},
			},
		},
		{
			name: "invalid:96",
			sourceText: "\n" +
				"interface Props {\n" +
				"  /** @deprecated */\n" +
				"  foo: Props;\n" +
				"}\n" +
				"declare const x: Props;\n" +
				"const { foo = x } = x;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "foo", reason: ""},
			},
		},
		{
			name: "invalid:135",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"interface Foo {}\n" +
				"\n" +
				"class Bar implements Foo {}\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "Foo", reason: ""},
			},
		},
		{
			name: "invalid:136",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"interface Foo {}\n" +
				"\n" +
				"export class Bar implements Foo {}\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "Foo", reason: ""},
			},
		},
		{
			name: "invalid:137",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"interface Foo {}\n" +
				"\n" +
				"interface Baz {}\n" +
				"\n" +
				"export class Bar implements Baz, Foo {}\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "Foo", reason: ""},
			},
		},
		{
			name: "invalid:138",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"class Foo {}\n" +
				"\n" +
				"export class Bar extends Foo {}\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "Foo", reason: ""},
			},
		},
		{
			name: "invalid:139",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"declare function decorator(constructor: Function);\n" +
				"\n" +
				"@decorator\n" +
				"export class Foo {}\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "decorator", reason: ""},
			},
		},
		{
			name: "invalid:140",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"function a(): object {\n" +
				"  return {};\n" +
				"}\n" +
				"\n" +
				"export default a();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "a", reason: ""},
			},
		},
		{
			name: "invalid:141",
			sourceText: "\n" +
				"class A {\n" +
				"  /** @deprecated */\n" +
				"  constructor() {}\n" +
				"}\n" +
				"\n" +
				"class B extends A {\n" +
				"  constructor() {\n" +
				"    /** should report but does not */\n" +
				"    super();\n" +
				"  }\n" +
				"}\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "super", reason: ""},
			},
		},
		{
			name: "invalid:142",
			sourceText: "\n" +
				"class A {\n" +
				"  /** @deprecated test reason*/\n" +
				"  constructor() {}\n" +
				"}\n" +
				"\n" +
				"class B extends A {\n" +
				"  constructor() {\n" +
				"    /** should report but does not */\n" +
				"    super();\n" +
				"  }\n" +
				"}\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "super", reason: "test reason"},
			},
		},
		{
			name: "invalid:147",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  accessor b: () => string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"a.b;\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:148",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  accessor b: () => string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"a['b'];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:149",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  accessor b: () => string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"a.b();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:150",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  accessor b: () => string;\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"a['b']();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:151",
			sourceText: "\n" +
				"class A {\n" +
				"  /** @deprecated */\n" +
				"  accessor b = (): string => {\n" +
				"    return '';\n" +
				"  };\n" +
				"}\n" +
				"\n" +
				"declare const a: A;\n" +
				"\n" +
				"a.b();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:152",
			sourceText: "\n" +
				"declare class A {\n" +
				"  /** @deprecated */\n" +
				"  static accessor b: () => string;\n" +
				"}\n" +
				"\n" +
				"A.b();\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:153",
			sourceText: "\n" +
				"class A {\n" +
				"  /** @deprecated */\n" +
				"  #b = () => {};\n" +
				"\n" +
				"  c() {\n" +
				"    this.#b();\n" +
				"  }\n" +
				"}\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "#b", reason: ""},
			},
		},
		{
			name: "invalid:154",
			sourceText: "\n" +
				"const a = {\n" +
				"  /** @deprecated */\n" +
				"  b: 'string',\n" +
				"};\n" +
				"\n" +
				"const c = a['b'];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:155",
			sourceText: "\n" +
				"const a = {\n" +
				"  /** @deprecated */\n" +
				"  b: 'string',\n" +
				"};\n" +
				"const x = 'b';\n" +
				"const c = a[x];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:156",
			sourceText: "\n" +
				"const a = {\n" +
				"  /** @deprecated */\n" +
				"  [2]: 'string',\n" +
				"};\n" +
				"const x = 'b';\n" +
				"const c = a[2];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "2", reason: ""},
			},
		},
		{
			name: "invalid:157",
			sourceText: "\n" +
				"const a = {\n" +
				"  /** @deprecated reason for deprecation */\n" +
				"  b: 'string',\n" +
				"};\n" +
				"\n" +
				"const key = 'b';\n" +
				"const stringKey = key as const;\n" +
				"const c = a[stringKey];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "b", reason: "reason for deprecation"},
			},
		},
		{
			name: "invalid:158",
			sourceText: "\n" +
				"enum Keys {\n" +
				"  B = 'b',\n" +
				"}\n" +
				"\n" +
				"const a = {\n" +
				"  /** @deprecated reason for deprecation */\n" +
				"  b: 'string',\n" +
				"};\n" +
				"\n" +
				"const key = Keys.B;\n" +
				"const c = a[key];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "b", reason: "reason for deprecation"},
			},
		},
		{
			name: "invalid:159",
			sourceText: "\n" +
				"declare const Keys: {\n" +
				"  a: 1;\n" +
				"};\n" +
				"\n" +
				"const a = {\n" +
				"  /** @deprecated reason for deprecation */\n" +
				"  [1]: 'string',\n" +
				"};\n" +
				"\n" +
				"const key = Keys.a;\n" +
				"const c = a[key];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "1", reason: "reason for deprecation"},
			},
		},
		{
			name: "invalid:160",
			sourceText: "\n" +
				"const a = {\n" +
				"  /** @deprecated */\n" +
				"  b: 'string',\n" +
				"};\n" +
				"\n" +
				"const key = `b`;\n" +
				"const c = a[key];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:161",
			sourceText: "\n" +
				"const stringObj = 'b';\n" +
				"const a = {\n" +
				"  /** @deprecated */\n" +
				"  b: 'string',\n" +
				"};\n" +
				"const c = a[stringObj];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:162",
			sourceText: "\n" +
				"declare function x(): 'b';\n" +
				"\n" +
				"const a = {\n" +
				"  /** @deprecated */\n" +
				"  b: 'string',\n" +
				"};\n" +
				"\n" +
				"const c = a[x()];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:163",
			sourceText: "\n" +
				"declare const x: { y: 'b' };\n" +
				"\n" +
				"const a = {\n" +
				"  /** @deprecated */\n" +
				"  b: 'string',\n" +
				"};\n" +
				"\n" +
				"const c = a[x.y];\n" +
				"      ",
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
	}
	runNoDeprecatedCases(t, cases)
}

// TestNoDeprecatedAcrossFiles is the part of the corpus that imports from another file.
//
// This is where the rule's difficulty lives and where the shim's alias accessors earn their place.
// The shared fixture writes tags in three positions that a one-hop answer confuses: on the
// declaration, on an export specifier that re-exports an UNTAGGED declaration, and on one overload
// of a function whose other overloads are clean. The middle position is invisible to a resolve,
// which jumps to the declaration and finds nothing, and is why deprecationInAliasChain walks.
func TestNoDeprecatedAcrossFiles(t *testing.T) {
	t.Parallel()

	cases := []noDeprecatedCase{
		{
			name: "valid:43",
			sourceText: "\n" +
				"import { deprecatedFunctionWithOverloads } from './deprecated';\n" +
				"\n" +
				"const foo = deprecatedFunctionWithOverloads();\n" +
				"    ",
			withDeprecatedFixture: true,
			expected:              nil,
		},
		{
			name: "valid:44",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const foo = imported.deprecatedFunctionWithOverloads();\n" +
				"    ",
			withDeprecatedFixture: true,
			expected:              nil,
		},
		{
			name: "valid:45",
			sourceText: "\n" +
				"import { ClassWithDeprecatedConstructor } from './deprecated';\n" +
				"\n" +
				"const foo = new ClassWithDeprecatedConstructor();\n" +
				"    ",
			withDeprecatedFixture: true,
			expected:              nil,
		},
		{
			name: "valid:46",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const foo = new imported.ClassWithDeprecatedConstructor();\n" +
				"    ",
			withDeprecatedFixture: true,
			expected:              nil,
		},
		{
			name: "valid:56",
			sourceText: "\n" +
				"export { deprecatedFunction as 'bur' } from './deprecated';\n" +
				"    ",
			withDeprecatedFixture: true,
			expected:              nil,
		},
		{
			name: "valid:57",
			sourceText: "\n" +
				"export { 'deprecatedFunction' } from './deprecated';\n" +
				"    ",
			withDeprecatedFixture: true,
			expected:              nil,
		},
		{
			name: "valid:61",
			sourceText: "\n" +
				"/**\n" +
				" * @deprecated\n" +
				" */\n" +
				"export import Bar = require('./deprecated');\n" +
				"    ",
			withDeprecatedFixture: true,
			expected:              nil,
		},
		{
			name: "valid:65",
			sourceText: "\n" +
				"async function fn() {\n" +
				"  const d = await import('./deprecated.js');\n" +
				"  d.default;\n" +
				"}\n" +
				"      ",
			withDeprecatedFixture: true,
			environmentDependent:  true,
			expected:              nil,
		},
		{
			name: "invalid:97",
			sourceText: "\n" +
				"import { DeprecatedClass } from './deprecated';\n" +
				"\n" +
				"const foo = new DeprecatedClass();\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "DeprecatedClass", reason: ""},
			},
		},
		{
			name: "invalid:98",
			sourceText: "\n" +
				"import { DeprecatedClass } from './deprecated';\n" +
				"\n" +
				"declare function inject(something: new () => unknown): void;\n" +
				"\n" +
				"inject(DeprecatedClass);\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "DeprecatedClass", reason: ""},
			},
		},
		{
			name: "invalid:99",
			sourceText: "\n" +
				"import { deprecatedVariable } from './deprecated';\n" +
				"\n" +
				"const foo = deprecatedVariable;\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "deprecatedVariable", reason: ""},
			},
		},
		{
			name: "invalid:100",
			sourceText: "\n" +
				"import { DeprecatedClass } from './deprecated';\n" +
				"\n" +
				"declare const x: DeprecatedClass;\n" +
				"\n" +
				"const { foo } = x;\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "DeprecatedClass", reason: ""},
				{messageId: "deprecated", name: "foo", reason: ""},
			},
		},
		{
			name: "invalid:101",
			sourceText: "\n" +
				"import { deprecatedFunction } from './deprecated';\n" +
				"\n" +
				"deprecatedFunction();\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "deprecatedFunction", reason: ""},
			},
		},
		{
			name: "invalid:102",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const foo = new imported.NormalClass();\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "NormalClass", reason: ""},
			},
		},
		{
			name: "invalid:103",
			sourceText: "\n" +
				"import { NormalClass } from './deprecated';\n" +
				"\n" +
				"const foo = new NormalClass();\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "NormalClass", reason: ""},
			},
		},
		{
			name: "invalid:104",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const foo = imported.NormalClass;\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "NormalClass", reason: ""},
			},
		},
		{
			name: "invalid:105",
			sourceText: "\n" +
				"import { NormalClass } from './deprecated';\n" +
				"\n" +
				"const foo = NormalClass;\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "NormalClass", reason: ""},
			},
		},
		{
			name: "invalid:106",
			sourceText: "\n" +
				"import { normalVariable } from './deprecated';\n" +
				"\n" +
				"const foo = normalVariable;\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "normalVariable", reason: ""},
			},
		},
		{
			name: "invalid:107",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const foo = imported.normalVariable;\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "normalVariable", reason: ""},
			},
		},
		{
			name: "invalid:108",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const { normalVariable } = imported;\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "normalVariable", reason: ""},
			},
		},
		{
			name: "invalid:109",
			sourceText: "\n" +
				"import { deprecatedVariable } from './deprecated';\n" +
				"\n" +
				"const test = {\n" +
				"  someField: deprecatedVariable,\n" +
				"};\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "deprecatedVariable", reason: ""},
			},
		},
		{
			name: "invalid:110",
			sourceText: "\n" +
				"import { normalFunction } from './deprecated';\n" +
				"\n" +
				"const foo = normalFunction;\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "normalFunction", reason: ""},
			},
		},
		{
			name: "invalid:111",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const foo = imported.normalFunction;\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "normalFunction", reason: ""},
			},
		},
		{
			name: "invalid:112",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const { normalFunction } = imported;\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "normalFunction", reason: ""},
			},
		},
		{
			name: "invalid:113",
			sourceText: "\n" +
				"import { normalFunction } from './deprecated';\n" +
				"\n" +
				"const foo = normalFunction();\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "normalFunction", reason: ""},
			},
		},
		{
			name: "invalid:114",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const foo = imported.normalFunction();\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "normalFunction", reason: ""},
			},
		},
		{
			name: "invalid:115",
			sourceText: "\n" +
				"import { deprecatedFunctionWithOverloads } from './deprecated';\n" +
				"\n" +
				"const foo = deprecatedFunctionWithOverloads('a');\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "deprecatedFunctionWithOverloads", reason: ""},
			},
		},
		{
			name: "invalid:116",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const foo = imported.deprecatedFunctionWithOverloads('a');\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "deprecatedFunctionWithOverloads", reason: ""},
			},
		},
		{
			name: "invalid:117",
			sourceText: "\n" +
				"import { reexportedDeprecatedFunctionWithOverloads } from './deprecated';\n" +
				"\n" +
				"const foo = reexportedDeprecatedFunctionWithOverloads;\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "reexportedDeprecatedFunctionWithOverloads", reason: "Reason"},
			},
		},
		{
			name: "invalid:118",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const foo = imported.reexportedDeprecatedFunctionWithOverloads;\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "reexportedDeprecatedFunctionWithOverloads", reason: "Reason"},
			},
		},
		{
			name: "invalid:119",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const { reexportedDeprecatedFunctionWithOverloads } = imported;\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "reexportedDeprecatedFunctionWithOverloads", reason: "Reason"},
			},
		},
		{
			name: "invalid:120",
			sourceText: "\n" +
				"import { reexportedDeprecatedFunctionWithOverloads } from './deprecated';\n" +
				"\n" +
				"const foo = reexportedDeprecatedFunctionWithOverloads();\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "reexportedDeprecatedFunctionWithOverloads", reason: "Reason"},
			},
		},
		{
			name: "invalid:121",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const foo = imported.reexportedDeprecatedFunctionWithOverloads();\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "reexportedDeprecatedFunctionWithOverloads", reason: "Reason"},
			},
		},
		{
			name: "invalid:122",
			sourceText: "\n" +
				"import { reexportedDeprecatedFunctionWithOverloads } from './deprecated';\n" +
				"\n" +
				"const foo = reexportedDeprecatedFunctionWithOverloads('a');\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "reexportedDeprecatedFunctionWithOverloads", reason: "Reason"},
			},
		},
		{
			name: "invalid:123",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const foo = imported.reexportedDeprecatedFunctionWithOverloads('a');\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "reexportedDeprecatedFunctionWithOverloads", reason: "Reason"},
			},
		},
		{
			name: "invalid:124",
			sourceText: "\n" +
				"import { ClassWithDeprecatedConstructor } from './deprecated';\n" +
				"\n" +
				"const foo = new ClassWithDeprecatedConstructor('a');\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "ClassWithDeprecatedConstructor", reason: ""},
			},
		},
		{
			name: "invalid:125",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const foo = new imported.ClassWithDeprecatedConstructor('a');\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "ClassWithDeprecatedConstructor", reason: ""},
			},
		},
		{
			name: "invalid:126",
			sourceText: "\n" +
				"import { ReexportedClassWithDeprecatedConstructor } from './deprecated';\n" +
				"\n" +
				"const foo = ReexportedClassWithDeprecatedConstructor;\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "ReexportedClassWithDeprecatedConstructor", reason: "Reason"},
			},
		},
		{
			name: "invalid:127",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const foo = imported.ReexportedClassWithDeprecatedConstructor;\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "ReexportedClassWithDeprecatedConstructor", reason: "Reason"},
			},
		},
		{
			name: "invalid:128",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const { ReexportedClassWithDeprecatedConstructor } = imported;\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "ReexportedClassWithDeprecatedConstructor", reason: "Reason"},
			},
		},
		{
			name: "invalid:129",
			sourceText: "\n" +
				"import { ReexportedClassWithDeprecatedConstructor } from './deprecated';\n" +
				"\n" +
				"const foo = ReexportedClassWithDeprecatedConstructor();\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "ReexportedClassWithDeprecatedConstructor", reason: "Reason"},
			},
		},
		{
			name: "invalid:130",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const foo = imported.ReexportedClassWithDeprecatedConstructor();\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "ReexportedClassWithDeprecatedConstructor", reason: "Reason"},
			},
		},
		{
			name: "invalid:131",
			sourceText: "\n" +
				"import { ReexportedClassWithDeprecatedConstructor } from './deprecated';\n" +
				"\n" +
				"const foo = ReexportedClassWithDeprecatedConstructor('a');\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "ReexportedClassWithDeprecatedConstructor", reason: "Reason"},
			},
		},
		{
			name: "invalid:132",
			sourceText: "\n" +
				"import * as imported from './deprecated';\n" +
				"\n" +
				"const foo = imported.ReexportedClassWithDeprecatedConstructor('a');\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "ReexportedClassWithDeprecatedConstructor", reason: "Reason"},
			},
		},
		{
			name: "invalid:133",
			sourceText: "\n" +
				"import imported from './deprecated';\n" +
				"\n" +
				"imported;\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "imported", reason: ""},
			},
		},
		{
			name: "invalid:134",
			sourceText: "\n" +
				"async function fn() {\n" +
				"  const d = await import('./deprecated.js');\n" +
				"  d.default.default;\n" +
				"}\n" +
				"      ",
			withDeprecatedFixture: true,
			environmentDependent:  true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "default", reason: ""},
			},
		},
		{
			name: "invalid:164",
			sourceText: "\n" +
				"import { deprecatedFunction } from './deprecated';\n" +
				"\n" +
				"export { deprecatedFunction };\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "deprecatedFunction", reason: ""},
			},
		},
		{
			name: "invalid:165",
			sourceText: "\n" +
				"export { deprecatedFunction } from './deprecated';\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "deprecatedFunction", reason: ""},
			},
		},
		{
			name: "invalid:166",
			sourceText: "\n" +
				"export type { T, U } from './deprecated';\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "T", reason: "Reason"},
			},
		},
		{
			name: "invalid:167",
			sourceText: "\n" +
				"export { default as foo } from './deprecated';\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "foo", reason: ""},
			},
		},
		{
			name: "invalid:168",
			sourceText: "\n" +
				"export { deprecatedFunction as bar } from './deprecated';\n" +
				"      ",
			withDeprecatedFixture: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "bar", reason: ""},
			},
		},
	}
	runNoDeprecatedCases(t, cases)
}

// TestNoDeprecatedInJsx covers the two JSX arms: a deprecated component and a deprecated prop.
//
// A prop's tag is on a property of the element name's CONTEXTUAL type rather than on anything the
// attribute itself resolves to, so this arm shares no code with the identifier arm and would be
// silent if only that one worked.
func TestNoDeprecatedInJsx(t *testing.T) {
	t.Parallel()

	cases := []noDeprecatedCase{
		{
			name: "valid:68",
			sourceText: "\n" +
				"declare namespace JSX {}\n" +
				"\n" +
				"<foo bar={1} />;\n" +
				"    ",
			jsx:      true,
			expected: nil,
		},
		{
			name: "valid:69",
			sourceText: "\n" +
				"declare namespace JSX {\n" +
				"  interface IntrinsicElements {\n" +
				"    foo: any;\n" +
				"  }\n" +
				"}\n" +
				"\n" +
				"<foo bar={1} />;\n" +
				"    ",
			jsx:      true,
			expected: nil,
		},
		{
			name: "valid:70",
			sourceText: "\n" +
				"declare namespace JSX {\n" +
				"  interface IntrinsicElements {\n" +
				"    foo: unknown;\n" +
				"  }\n" +
				"}\n" +
				"\n" +
				"<foo bar={1} />;\n" +
				"    ",
			jsx:      true,
			expected: nil,
		},
		{
			name: "valid:71",
			sourceText: "\n" +
				"declare namespace JSX {\n" +
				"  interface IntrinsicElements {\n" +
				"    foo: {\n" +
				"      bar: any;\n" +
				"    };\n" +
				"  }\n" +
				"}\n" +
				"<foo bar={1} />;\n" +
				"    ",
			jsx:      true,
			expected: nil,
		},
		{
			name: "valid:72",
			sourceText: "\n" +
				"declare namespace JSX {\n" +
				"  interface IntrinsicElements {\n" +
				"    foo: {\n" +
				"      bar: unknown;\n" +
				"    };\n" +
				"  }\n" +
				"}\n" +
				"<foo bar={1} />;\n" +
				"    ",
			jsx:      true,
			expected: nil,
		},
		{
			name: "valid:74",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"function A() {\n" +
				"  return <div />;\n" +
				"}\n" +
				"\n" +
				"const a = <A></A>;\n" +
				"      ",
			options:  "{\"allow\":[{\"from\":\"file\",\"name\":\"A\"}]}",
			jsx:      true,
			expected: nil,
		},
		{
			name: "invalid:0",
			sourceText: "\n" +
				"interface AProps {\n" +
				"  /** @deprecated */\n" +
				"  b: number | string;\n" +
				"}\n" +
				"\n" +
				"function A(props: AProps) {\n" +
				"  return <div />;\n" +
				"}\n" +
				"\n" +
				"const a = <A b=\"\" />;\n" +
				"      ",
			jsx: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "b", reason: ""},
			},
		},
		{
			name: "invalid:87",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"const A = () => <div />;\n" +
				"\n" +
				"const a = <A />;\n" +
				"      ",
			jsx: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "A", reason: ""},
			},
		},
		{
			name: "invalid:88",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"const A = () => <div />;\n" +
				"\n" +
				"const a = <A></A>;\n" +
				"      ",
			jsx: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "A", reason: ""},
			},
		},
		{
			name: "invalid:89",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"function A() {\n" +
				"  return <div />;\n" +
				"}\n" +
				"\n" +
				"const a = <A />;\n" +
				"      ",
			jsx: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "A", reason: ""},
			},
		},
		{
			name: "invalid:90",
			sourceText: "\n" +
				"/** @deprecated */\n" +
				"function A() {\n" +
				"  return <div />;\n" +
				"}\n" +
				"\n" +
				"const a = <A></A>;\n" +
				"      ",
			jsx: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "A", reason: ""},
			},
		},
		{
			name:         "invalid:143",
			sourceText:   "const a = <div aria-grabbed></div>;",
			jsx:          true,
			harnessBound: "the deprecation is on @types/react's JSX.IntrinsicElements, and the harness writes its own tsconfig with \"types\": [] after the setup hook runs, so no ambient declaration can reach the program; the intrinsic-attribute arm is proven in TestNoDeprecatedOnIntrinsicJsxAttributes instead",
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "aria-grabbed", reason: "in ARIA 1.1"},
			},
		},
		{
			name: "invalid:144",
			sourceText: "\n" +
				"declare namespace JSX {\n" +
				"  interface IntrinsicElements {\n" +
				"    'foo-bar:baz-bam': {\n" +
				"      name: string;\n" +
				"      /**\n" +
				"       * @deprecated\n" +
				"       */\n" +
				"      deprecatedProp: string;\n" +
				"    };\n" +
				"  }\n" +
				"}\n" +
				"\n" +
				"const componentDashed = <foo-bar:baz-bam name=\"e\" deprecatedProp=\"oh no\" />;\n" +
				"      ",
			jsx:                  true,
			environmentDependent: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "deprecatedProp", reason: ""},
			},
		},
		{
			name: "invalid:145",
			sourceText: "\n" +
				"import * as React from 'react';\n" +
				"\n" +
				"interface Props {\n" +
				"  /**\n" +
				"   * @deprecated\n" +
				"   */\n" +
				"  deprecatedProp: string;\n" +
				"}\n" +
				"\n" +
				"interface Tab {\n" +
				"  List: React.FC<Props>;\n" +
				"}\n" +
				"\n" +
				"const Tab: Tab = {\n" +
				"  List: () => <div>Hi</div>,\n" +
				"};\n" +
				"\n" +
				"const anotherExample = <Tab.List deprecatedProp=\"oh no\" />;\n" +
				"      ",
			jsx:                true,
			withInstalledTypes: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecated", name: "deprecatedProp", reason: ""},
			},
		},
	}
	runNoDeprecatedCases(t, cases)
}

// TestNoDeprecatedAgainstNodeTypes covers deprecation that comes from a @types package.
//
// Every case here is environment dependent: whether `fs.exists` or `assert.fail` is marked
// deprecated is a property of the installed @types/node, which is not the version upstream pins.
// They carry upstream's declared expectations and are skipped rather than asserted, because a
// machine with different types would fail them for a reason that says nothing about the rule.
func TestNoDeprecatedAgainstNodeTypes(t *testing.T) {
	t.Parallel()

	cases := []noDeprecatedCase{
		{
			name: "valid:52",
			sourceText: "\n" +
				"import assert from 'node:assert';\n" +
				"\n" +
				"assert.fail('');\n" +
				"    ",
			expected: nil,
		},
		{
			name: "valid:79",
			sourceText: "\n" +
				"import { exists } from 'fs';\n" +
				"exists('/foo');\n" +
				"      ",
			options:              "{\"allow\":[{\"from\":\"package\",\"name\":\"exists\",\"package\":\"fs\"}]}",
			environmentDependent: true,
			expected:             nil,
		},
		{
			name: "invalid:75",
			sourceText: "\n" +
				"import assert from 'node:assert';\n" +
				"\n" +
				"assert.fail({}, {});\n" +
				"      ",
			environmentDependent: true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "fail", reason: "since v10.0.0 - use fail([message]) or other assert functions instead."},
			},
		},
		{
			name: "invalid:146",
			sourceText: "\n" +
				"import { exists } from 'fs';\n" +
				"exists('/foo');\n" +
				"      ",
			options:      "{\"allow\":[{\"from\":\"package\",\"name\":\"exists\",\"package\":\"hoge\"}]}",
			harnessBound: "the deprecation is on @types/node's fs.exists, which the harness's \"types\": [] excludes; the allowlist arm this case exercises is proven by the other allow cases, and whether fs.exists is marked deprecated differs between @types/node versions anyway",
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "exists", reason: "Since v1.0.0 - Use {@link stat} or {@link access} instead."},
			},
		},
	}
	runNoDeprecatedCases(t, cases)
}

// TestNoDeprecatedOnIntrinsicJsxAttributes covers the arm upstream's corpus cannot reach here.
//
// Two of upstream's cases put the tag on an intrinsic element's attribute, `<div aria-grabbed>` and
// a custom `<foo-bar>`, and both need a JSX namespace this harness cannot supply: it writes its own
// tsconfig with `"types": []`, so `@types/react` never loads, and upstream's own fixture writes a
// bare `declare namespace JSX` that is not global inside a module.
//
// The arm is real and was silent until this was written, so it is proven here instead, with the
// namespace declared global in the fixture. The second row is the control that matters: a rule
// reaching the intrinsic attribute type but reporting unconditionally would pass the first row and
// fail only this one.
func TestNoDeprecatedOnIntrinsicJsxAttributes(t *testing.T) {
	t.Parallel()

	const intrinsicElements = "declare global {\n" +
		"  namespace JSX {\n" +
		"    interface IntrinsicElements {\n" +
		"      'foo-bar': {\n" +
		"        /** @deprecated in ARIA 1.1 */\n" +
		"        deprecatedProp: string;\n" +
		"        fine: string;\n" +
		"      };\n" +
		"    }\n" +
		"  }\n" +
		"}\n"

	runNoDeprecatedCases(t, []noDeprecatedCase{
		{
			name:       "a deprecated attribute on an intrinsic element",
			sourceText: intrinsicElements + "export const a = <foo-bar deprecatedProp=\"x\" />;\n",
			jsx:        true,
			expected: []noDeprecatedFinding{
				{messageId: "deprecatedWithReason", name: "deprecatedProp", reason: "in ARIA 1.1"},
			},
		},
		{
			name:       "an undeprecated attribute on the same element",
			sourceText: intrinsicElements + "export const a = <foo-bar fine=\"x\" />;\n",
			jsx:        true,
			expected:   nil,
		},
	})
}

// TestNoDeprecatedFiresAndStaysQuiet is the fixture pair, through the shared harness assertions.
//
// The corpus blocks above compare findings themselves, because they assert the rendered message and
// not only the id, and 26 of upstream's cases make that reason text part of the contract. This pair
// exists alongside them so the two claims a rule has to support, that it can fire and that it can
// stay silent, are made in the form the registry guard reads.
//
// The two rows are the same declaration used two ways, so a rule that reported on everything would
// fail the second and a rule that reported on nothing would fail the first. Neither can be passed by
// accident.
func TestNoDeprecatedFiresAndStaysQuiet(t *testing.T) {
	t.Parallel()

	const source = "/** @deprecated Use b instead. */\n" +
		"declare function a(): void;\n" +
		"declare function b(): void;\n"

	t.Run("fires on a use of the deprecated declaration", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTyped(t, NoDeprecated, "/repository/source/Subject.ts",
			source+"a();\n")
		rule_testing.ExpectFindings(t, result, "deprecatedWithReason")
	})

	t.Run("stays quiet on a use of the undeprecated one", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.RunTyped(t, NoDeprecated, "/repository/source/Subject.ts",
			source+"b();\n")
		rule_testing.ExpectClean(t, result)
	})

	t.Run("stays quiet on the declaration itself", func(t *testing.T) {
		t.Parallel()
		// The declaration site is not a use, so a rule keyed on the tag rather than on the
		// reference would report here and this is what catches it.
		result := rule_testing.RunTyped(t, NoDeprecated, "/repository/source/Subject.ts", source)
		rule_testing.ExpectClean(t, result)
	})

	t.Run("declines a file with no type checker", func(t *testing.T) {
		t.Parallel()
		// NeedsTypeChecker keeps this unreachable through registration, but the harness path and
		// the registry's crash corpus both build a Context by hand. The rule panicked on exactly
		// this before the guard in Run was written, so the decline is pinned here.
		result := rule_testing.Run(t, NoDeprecated, "/repository/source/Subject.ts", source+"a();\n")
		rule_testing.ExpectClean(t, result)
	})
}
