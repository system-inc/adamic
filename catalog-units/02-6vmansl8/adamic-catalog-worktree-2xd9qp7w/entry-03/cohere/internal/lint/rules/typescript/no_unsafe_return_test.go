package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const noUnsafeReturnFile = "/repository/source/Returns.ts"

func noUnsafeReturnCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// TestNoUnsafeReturnStaysSilent is upstream's thirty passing cases verbatim.
//
// Upstream runs this rule's whole corpus under `tsconfig.noImplicitThis.json` and our harness pins
// `strict: true`, so every case was measured against the installed 8.67.0 build under BOTH settings
// rather than assumed to transfer. All thirty are clean under both, so nothing in this list depends
// on which one is in force. The single case whose verdict does move is in the firing test, at the id
// our configuration produces.
//
// Several of these are the false positives a plausible port ships. Returning `any` where `unknown`
// was declared is SAFE, and so is `any[]` into `unknown[]` and `Promise<any>` into a function whose
// awaited return is `unknown`, because `unknown` forces the caller to narrow. An explicit `any`
// return annotation is a decision the author made and is respected. And a `Promise<any>` returned
// from a synchronous function is not this rule's finding.
func TestNoUnsafeReturnStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"function foo() {\n  return;\n}\n",
		"function foo() {\n  return 1;\n}\n",
		"function foo() {\n  return '';\n}\n",
		"function foo() {\n  return true;\n}\n",
		"function foo() {\n  return [];\n}\n",
		"function foo(): any {\n  return {} as any;\n}\n",
		"declare function foo(arg: () => any): void;\nfoo((): any => 'foo' as any);\n",
		"declare function foo(arg: null | (() => any)): void;\nfoo((): any => 'foo' as any);\n",
		"function foo(): any[] {\n  return [] as any[];\n}\n",
		"function foo(): Set<any> {\n  return new Set<any>();\n}\n",
		"async function foo(): Promise<any> {\n  return Promise.resolve({} as any);\n}\n",
		"async function foo(): Promise<any> {\n  return {} as any;\n}\n",
		"function foo(): object {\n  return Promise.resolve({} as any);\n}\n",
		"function foo(): ReadonlySet<number> {\n  return new Set<any>();\n}\n",
		"function foo(): Set<number> {\n  return new Set([1]);\n}\n",
		"type Foo<T = number> = { prop: T };\nfunction foo(): Foo {\n  return { prop: 1 } as Foo<number>;\n}\n",
		"type Foo = { prop: any };\nfunction foo(): Foo {\n  return { prop: '' } as Foo;\n}\n",
		"function fn<T extends any>(x: T) {\n  return x;\n}\n",
		"function fn<T extends any>(x: T): unknown {\n  return x as any;\n}\n",
		"function fn<T extends any>(x: T): unknown[] {\n  return x as any[];\n}\n",
		"function fn<T extends any>(x: T): Set<unknown> {\n  return x as Set<any>;\n}\n",
		"async function fn<T extends any>(x: T): Promise<unknown> {\n  return x as any;\n}\n",
		"function fn<T extends any>(x: T): Promise<unknown> {\n  return Promise.resolve(x as any);\n}\n",
		"type Wrapper<T> = { inner: T };\ntype Extractor<D extends Wrapper<any>> = D extends Wrapper<infer V> ? V : never;\nconst fn =\n  <D extends Wrapper<any>>(foo: Extractor<D>) =>\n  () =>\n    foo;\n",
		"function test(): Map<string, string> {\n  return new Map();\n}\n",
		"function foo(): any {\n  return [] as any[];\n}\n",
		"function foo(): unknown {\n  return [] as any[];\n}\n",
		"declare const value: Promise<any>;\nfunction foo() {\n  return value;\n}\n",
		"const foo: (() => void) | undefined = () => 1;\n",
		"class Foo {\n  public foo(): this {\n    return this;\n  }\n\n  protected then(resolve: () => void): void {\n    resolve();\n  }\n}\n",

		// Measured boundary cases upstream does not write, all driven through the installed 8.67.0
		// build before being written here.

		// `unknown` absorbing `any` where the function has NO return annotation of its own and gets
		// its return type contextually. These exist because a mutant removing that branch survived
		// the whole imported corpus: every one of upstream's `unknown` cases writes an explicit
		// annotation, and the annotation branch above keeps those clean on its own, so nothing
		// imported can see the second branch at all. Without it all three of these report.
		"const foo: () => unknown = () => 1 as any;\n",
		"declare function wrap(f: () => unknown): void;\nwrap(() => 1 as any);\n",
		"type F = () => unknown;\nconst f: F = () => JSON.parse('');\n",

		// A UNION-typed receiver in the unsafe-assignment branch. Upstream uses two different
		// signature accessors in this rule, a union-flattening one for the excuse loops and the
		// checker's own for this branch, and a port using the flattening one everywhere reports
		// both of these. Measured clean upstream; the non-union control that DOES report is in the
		// firing test.
		"declare function foo(arg: null | (() => Set<string>)): void;\nfoo(() => new Set<any>());\n",
		"const foo: (() => Set<string>) | null = () => new Set<any>();\n",
	}
	for index, sourceText := range cases {
		t.Run(noUnsafeReturnCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUnsafeReturn,
				noUnsafeReturnFile, sourceText))
		})
	}
}

// noUnsafeReturnFinding is one expected finding with all three layers it can be wrong at.
//
// Three message ids describe three different defects, and two of them interpolate. The plain
// message names which of four things the value actually was, and the assignment message names both
// types by their printed names, which is where a port comparing the wrong pair of types shows
// itself. The span separates a finding on a return statement from one on a concise arrow body.
type noUnsafeReturnFinding struct {
	wantSpan    string
	wantId      string
	wantMessage string
}

// TestNoUnsafeReturnFires is upstream's thirty-two reporting cases verbatim, at the verdicts this
// harness can observe.
//
// Every row's id, rendered text, and span come from the installed 8.67.0 build run under OUR
// compiler options rather than upstream's. Thirty-one are identical under both settings. The
// thirty-second is the `return this` case, which reports twice either way while its id moves from
// `unsafeReturnThis` to `unsafeReturn`; it is pinned at ours and the other is recorded in the rule's
// doc comment so the branch is not read as unported.
func TestNoUnsafeReturnFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		wantFindings []noUnsafeReturnFinding
	}{
		{
			sourceText: "function foo() {\n  return 1 as any;\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return 1 as any;",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `any`.",
				},
			},
		},
		{
			sourceText: "function foo() {\n  return Object.create(null);\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return Object.create(null);",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `any`.",
				},
			},
		},
		{
			sourceText: "const foo = () => {\n  return 1 as any;\n};\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return 1 as any;",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `any`.",
				},
			},
		},
		{
			sourceText: "const foo = () => Object.create(null);\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "Object.create(null)",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `any`.",
				},
			},
		},
		{
			sourceText: "function foo() {\n  return [] as any[];\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return [] as any[];",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `any[]`.",
				},
			},
		},
		{
			sourceText: "function foo() {\n  return [] as Array<any>;\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return [] as Array<any>;",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `any[]`.",
				},
			},
		},
		{
			sourceText: "function foo() {\n  return [] as readonly any[];\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return [] as readonly any[];",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `any[]`.",
				},
			},
		},
		{
			sourceText: "function foo() {\n  return [] as Readonly<any[]>;\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return [] as Readonly<any[]>;",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `any[]`.",
				},
			},
		},
		{
			sourceText: "const foo = () => {\n  return [] as any[];\n};\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return [] as any[];",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `any[]`.",
				},
			},
		},
		{
			sourceText: "const foo = () => [] as any[];\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "[] as any[]",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `any[]`.",
				},
			},
		},
		{
			sourceText: "function foo(): Set<string> {\n  return new Set<any>();\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return new Set<any>();",
					wantId:      "unsafeReturnAssignment",
					wantMessage: "Unsafe return of type `Set<any>` from function with return type `Set<string>`.",
				},
			},
		},
		{
			sourceText: "function foo(): Map<string, string> {\n  return new Map<string, any>();\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return new Map<string, any>();",
					wantId:      "unsafeReturnAssignment",
					wantMessage: "Unsafe return of type `Map<string, any>` from function with return type `Map<string, string>`.",
				},
			},
		},
		{
			sourceText: "function foo(): Set<string[]> {\n  return new Set<any[]>();\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return new Set<any[]>();",
					wantId:      "unsafeReturnAssignment",
					wantMessage: "Unsafe return of type `Set<any[]>` from function with return type `Set<string[]>`.",
				},
			},
		},
		{
			sourceText: "function foo(): Set<Set<Set<string>>> {\n  return new Set<Set<Set<any>>>();\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return new Set<Set<Set<any>>>();",
					wantId:      "unsafeReturnAssignment",
					wantMessage: "Unsafe return of type `Set<Set<Set<any>>>` from function with return type `Set<Set<Set<string>>>`.",
				},
			},
		},
		{
			sourceText: "type Fn = () => Set<string>;\nconst foo1: Fn = () => new Set<any>();\nconst foo2: Fn = function test() {\n  return new Set<any>();\n};\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "new Set<any>()",
					wantId:      "unsafeReturnAssignment",
					wantMessage: "Unsafe return of type `Set<any>` from function with return type `Set<string>`.",
				},
				{
					wantSpan:    "return new Set<any>();",
					wantId:      "unsafeReturnAssignment",
					wantMessage: "Unsafe return of type `Set<any>` from function with return type `Set<string>`.",
				},
			},
		},
		{
			sourceText: "type Fn = () => Set<string>;\nfunction receiver(arg: Fn) {}\nreceiver(() => new Set<any>());\nreceiver(function test() {\n  return new Set<any>();\n});\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "new Set<any>()",
					wantId:      "unsafeReturnAssignment",
					wantMessage: "Unsafe return of type `Set<any>` from function with return type `Set<string>`.",
				},
				{
					wantSpan:    "return new Set<any>();",
					wantId:      "unsafeReturnAssignment",
					wantMessage: "Unsafe return of type `Set<any>` from function with return type `Set<string>`.",
				},
			},
		},
		{
			sourceText: "function foo() {\n  return this;\n}\n\nfunction bar() {\n  return () => this;\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return this;",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `any`.",
				},
				{
					wantSpan:    "this",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `any`.",
				},
			},
		},
		{
			sourceText: "declare function foo(arg: null | (() => any)): void;\nfoo(() => 'foo' as any);\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "'foo' as any",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `any`.",
				},
			},
		},
		{
			sourceText: "let value: NotKnown;\n\nfunction example() {\n  return value;\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return value;",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type error.",
				},
			},
		},
		{
			sourceText: "declare const value: any;\nasync function foo() {\n  return value;\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return value;",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `any`.",
				},
			},
		},
		{
			sourceText: "declare const value: Promise<any>;\nasync function foo(): Promise<number> {\n  return value;\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return value;",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `Promise<any>`.",
				},
			},
		},
		{
			sourceText: "async function foo(arg: number) {\n  return arg as Promise<any>;\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return arg as Promise<any>;",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `Promise<any>`.",
				},
			},
		},
		{
			sourceText: "function foo(): Promise<any> {\n  return {} as any;\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return {} as any;",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `any`.",
				},
			},
		},
		{
			sourceText: "function foo(): Promise<object> {\n  return {} as any;\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return {} as any;",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `any`.",
				},
			},
		},
		{
			sourceText: "async function foo(): Promise<object> {\n  return Promise.resolve<any>({});\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return Promise.resolve<any>({});",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `Promise<any>`.",
				},
			},
		},
		{
			sourceText: "async function foo(): Promise<object> {\n  return Promise.resolve<Promise<Promise<any>>>({} as Promise<any>);\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return Promise.resolve<Promise<Promise<any>>>({} as Promise<any>);",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `Promise<any>`.",
				},
			},
		},
		{
			sourceText: "async function foo(): Promise<object> {\n  return {} as Promise<Promise<Promise<Promise<any>>>>;\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return {} as Promise<Promise<Promise<Promise<any>>>>;",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `Promise<any>`.",
				},
			},
		},
		{
			sourceText: "async function foo() {\n  return {} as Promise<Promise<Promise<Promise<any>>>>;\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return {} as Promise<Promise<Promise<Promise<any>>>>;",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `Promise<any>`.",
				},
			},
		},
		{
			sourceText: "async function foo() {\n  return {} as Promise<any> | Promise<object>;\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return {} as Promise<any> | Promise<object>;",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `Promise<any>`.",
				},
			},
		},
		{
			sourceText: "async function foo() {\n  return {} as Promise<any | object>;\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return {} as Promise<any | object>;",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `Promise<any>`.",
				},
			},
		},
		{
			sourceText: "async function foo() {\n  return {} as Promise<any> & { __brand: 'any' };\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return {} as Promise<any> & { __brand: 'any' };",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `Promise<any>`.",
				},
			},
		},
		{
			sourceText: "interface Alias<T> extends Promise<any> {\n  foo: 'bar';\n}\n\ndeclare const value: Alias<number>;\nasync function foo() {\n  return value;\n}\n",
			wantFindings: []noUnsafeReturnFinding{
				{
					wantSpan:    "return value;",
					wantId:      "unsafeReturn",
					wantMessage: "Unsafe return of a value of type `Promise<any>`.",
				},
			},
		},
	}
	for index, testCase := range cases {
		t.Run(noUnsafeReturnCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnsafeReturn, noUnsafeReturnFile,
				testCase.sourceText)

			wantIds := make([]string, len(testCase.wantFindings))
			for position, want := range testCase.wantFindings {
				wantIds[position] = want.wantId
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			// The harness writes the fixture trimmed, so the span slices are against that text
			// rather than the Go literal above. The captured spans came from a file written the
			// same way, so the two agree by construction.
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			for position, want := range testCase.wantFindings {
				diagnostic := result.Diagnostics[position]

				gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != want.wantSpan {
					t.Fatalf("finding %d span: expected %q, got %q", position, want.wantSpan, gotSpan)
				}
				if diagnostic.Message.Id != want.wantId {
					t.Fatalf("finding %d id: expected %q, got %q", position, want.wantId,
						diagnostic.Message.Id)
				}
				if diagnostic.Message.Description != want.wantMessage {
					t.Fatalf("finding %d message: expected %q, got %q", position, want.wantMessage,
						diagnostic.Message.Description)
				}
			}
		})
	}
}
