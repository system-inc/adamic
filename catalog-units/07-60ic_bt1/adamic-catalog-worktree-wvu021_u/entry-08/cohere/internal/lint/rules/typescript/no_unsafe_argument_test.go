package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const noUnsafeArgumentFile = "/repository/source/Arguments.ts"

func noUnsafeArgumentCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// TestNoUnsafeArgumentStaysSilent is upstream's twenty-three passing cases verbatim.
//
// Each was re-measured against the installed 8.67.0 build with the file written exactly as this
// harness writes it, trimmed with one trailing newline, so the two tools are reading the same
// bytes. All twenty-three are clean there, which is what makes them usable as evidence rather than
// as a label copied out of an array named `valid`.
//
// Several of them are the false positives a plausible port ships. A call to an undeclared function
// and a call on a non-function are both silent because the callee cannot resolve, and reproducing
// that costs nothing only because the signature lookup declines. Passing an `any` to a parameter
// typed `any` or `unknown` is silent because assignment INTO those is not unsafe, which lives in
// the shared assignment predicate rather than here. `acceptsMap(new Map())` is silent because of a
// special case inside that predicate for Map's empty constructor, which is typed `Map<any, any>`.
func TestNoUnsafeArgumentStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"doesNotExist(1 as any);\n",
		"const foo = 1;\nfoo(1 as any);\n",
		"declare function foo(arg: number): void;\nfoo(1, 1 as any, 2 as any);\n",
		"declare function foo(arg: number, arg2: string): void;\nfoo(1, 'a');\n",
		"declare function foo(arg: any): void;\nfoo(1 as any);\n",
		"declare function foo(arg: unknown): void;\nfoo(1 as any);\n",
		"declare function foo(...arg: number[]): void;\nfoo(1, 2, 3);\n",
		"declare function foo(...arg: any[]): void;\nfoo(1, 2, 3, 4 as any);\n",
		"declare function foo(arg: number, arg2: number): void;\nconst x = [1, 2] as const;\nfoo(...x);\n",
		"declare function foo(arg: any, arg2: number): void;\nconst x = [1 as any, 2] as const;\nfoo(...x);\n",
		"declare function foo(arg1: string, arg2: string): void;\nconst x: string[] = [];\nfoo(...x);\n",
		"function foo(arg1: number, arg2: number) {}\nfoo(...([1, 1, 1] as [number, number, number]));\n",
		"declare function foo(arg1: Set<string>, arg2: Map<string, string>): void;\n\nconst x = [new Map<string, string>()] as const;\nfoo(new Set<string>(), ...x);\n",
		"declare function foo(arg1: unknown, arg2: Set<unknown>, arg3: unknown[]): void;\nfoo(1 as any, new Set<any>(), [] as any[]);\n",
		"declare function foo(...params: [number, string, any]): void;\nfoo(1, 'a', 1 as any);\n",
		"declare function foo<E extends string[]>(...params: E): void;\n\nfoo('a', 'b', 1 as any);\n",
		"declare function toHaveBeenCalledWith<E extends any[]>(...params: E): void;\ntoHaveBeenCalledWith(1 as any);\n",
		"declare function acceptsMap(arg: Map<string, string>): void;\nacceptsMap(new Map());\n",
		"type T = [number, T[]];\ndeclare function foo(t: T): void;\ndeclare const t: T;\n\nfoo(t);\n",
		"type T = Array<T>;\ndeclare function foo<T>(t: T): T;\nconst t: T = [];\nfoo(t);\n",
		"function foo(templates: TemplateStringsArray) {}\nfoo``;\n",
		"function foo(templates: TemplateStringsArray, arg: any) {}\nfoo`${1 as any}`;\n",
		"declare function foo(...args: any): void;\nfoo(1 as any);\n",

		// Measured boundary cases upstream does not write, all four driven through the installed
		// 8.67.0 build before being written here.

		// A tuple spread whose target ends in a rest absorbs every argument after it, so the
		// trailing `any` gets no parameter to be compared against and is silent. These three exist
		// because a mutant that never sets the consumed flag survived the whole imported corpus:
		// upstream's own tuple cases all end the argument list at the spread, so nothing in them
		// can see the difference. Without the flag all three of these report.
		"declare function foo(a: string, b: string, c: string): void;\ndeclare const x: [string, ...string[]];\nfoo(...x, 1 as any);\n",
		"declare function foo(a: number, b: number, c: number): void;\ndeclare const x: [number, ...number[]];\nfoo(...x, 1 as any);\n",
		"declare function foo(a: string, b: string, c: number): void;\ndeclare const x: [string, ...string[]];\nfoo(...x, 1 as any, 2 as any);\n",
	}
	for index, sourceText := range cases {
		t.Run(noUnsafeArgumentCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUnsafeArgument,
				noUnsafeArgumentFile, sourceText))
		})
	}
}

// noUnsafeArgumentFinding is one expected finding with all three layers it can be wrong at.
//
// The rule has four message ids and every one of them interpolates at least one type name, so the
// id alone proves very little: a port that matched arguments against the wrong parameters would
// report the right count of the right ids while naming types nobody wrote. The rendered message is
// where the parameter walk becomes visible, and the span separates a finding on a spread from one
// on the expression inside it.
type noUnsafeArgumentFinding struct {
	wantSpan    string
	wantId      string
	wantMessage string
}

// TestNoUnsafeArgumentFires is upstream's twenty-two reporting cases verbatim, with every finding's
// id, rendered text, and span taken from the installed 8.67.0 build rather than from the corpus's
// own `errors` array. The corpus states ids and sometimes columns; it does not state the rendered
// text, and the rendered text is where the signature walk shows itself.
func TestNoUnsafeArgumentFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		wantFindings []noUnsafeArgumentFinding
	}{
		{
			sourceText: "declare function foo(...args: [string, string]): void;\n\ndeclare const spread: [string, ...string[]];\nfoo(...spread, 1 as any);\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "1 as any",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `string`.",
				},
			},
		},
		{
			sourceText: "declare function foo(...args: [string, ...string[]]): void;\n\nfoo('a', 'b', 1 as any);\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "1 as any",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `string`.",
				},
			},
		},
		{
			sourceText: "declare function foo(arg: number): void;\nfoo(1 as any);\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "1 as any",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `number`.",
				},
			},
		},
		{
			sourceText: "declare function foo(arg: number): void;\nfoo(error);\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "error",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type error typed assigned to a parameter of type `number`.",
				},
			},
		},
		{
			sourceText: "declare function foo(arg1: number, arg2: string): void;\nfoo(1, 1 as any);\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "1 as any",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `string`.",
				},
			},
		},
		{
			sourceText: "declare function foo(...arg: number[]): void;\nfoo(1, 2, 3, 1 as any);\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "1 as any",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `number`.",
				},
			},
		},
		{
			sourceText: "declare function foo(arg: string, ...arg: number[]): void;\nfoo(1 as any, 1 as any);\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "1 as any",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `string`.",
				},
				{
					wantSpan:    "1 as any",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `number`.",
				},
			},
		},
		{
			sourceText: "declare function foo(arg1: string, arg2: number): void;\n\nfoo(...(x as any));\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "...(x as any)",
					wantId:      "unsafeSpread",
					wantMessage: "Unsafe spread of an `any` type.",
				},
			},
		},
		{
			sourceText: "declare function foo(arg1: string, arg2: number): void;\n\nfoo(...(x as any[]));\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "...(x as any[])",
					wantId:      "unsafeArraySpread",
					wantMessage: "Unsafe spread of an `any[]` array type.",
				},
			},
		},
		{
			sourceText: "declare function foo(arg1: string, arg2: number): void;\n\ndeclare const errors: error[];\n\nfoo(...errors);\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "...errors",
					wantId:      "unsafeArraySpread",
					wantMessage: "Unsafe spread of an error array type.",
				},
			},
		},
		{
			sourceText: "declare function foo(arg1: string, arg2: number): void;\n\nconst x = ['a', 1 as any] as const;\nfoo(...x);\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "...x",
					wantId:      "unsafeTupleSpread",
					wantMessage: "Unsafe spread of a tuple type. The argument is of type `any` and is assigned to a parameter of type `number`.",
				},
			},
		},
		{
			sourceText: "declare function foo(arg1: string, arg2: number): void;\n\nconst x = ['a', error] as const;\nfoo(...x);\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "...x",
					wantId:      "unsafeTupleSpread",
					wantMessage: "Unsafe spread of a tuple type. The argument is error typed and is assigned to a parameter of type `number`.",
				},
			},
		},
		{
			sourceText: "declare function foo(arg1: string, arg2: number): void;\nfoo(...(['foo', 1, 2] as [string, any, number]));\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "...(['foo', 1, 2] as [string, any, number])",
					wantId:      "unsafeTupleSpread",
					wantMessage: "Unsafe spread of a tuple type. The argument is of type `any` and is assigned to a parameter of type `number`.",
				},
			},
		},
		{
			sourceText: "declare function foo(arg1: string, arg2: number, arg2: string): void;\n\nconst x = [1] as const;\nfoo('a', ...x, 1 as any);\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "1 as any",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `string`.",
				},
			},
		},
		{
			sourceText: "declare function foo(arg1: string, arg2: number, ...rest: string[]): void;\n\nconst x = [1, 2] as [number, ...number[]];\nfoo('a', ...x, 1 as any);\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "1 as any",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `string`.",
				},
			},
		},
		{
			sourceText: "declare function foo(arg1: Set<string>, arg2: Map<string, string>): void;\n\nconst x = [new Map<any, string>()] as const;\nfoo(new Set<any>(), ...x);\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "new Set<any>()",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `Set<any>` assigned to a parameter of type `Set<string>`.",
				},
				{
					wantSpan:    "...x",
					wantId:      "unsafeTupleSpread",
					wantMessage: "Unsafe spread of a tuple type. The argument is of type `Map<any, string>` and is assigned to a parameter of type `Map<string, string>`.",
				},
			},
		},
		{
			sourceText: "declare function foo(...params: [number, string, any]): void;\nfoo(1 as any, 'a' as any, 1 as any);\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "1 as any",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `number`.",
				},
				{
					wantSpan:    "'a' as any",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `string`.",
				},
			},
		},
		{
			sourceText: "declare function foo(param1: string, ...params: [number, string, any]): void;\nfoo('a', 1 as any, 'a' as any, 1 as any);\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "1 as any",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `number`.",
				},
				{
					wantSpan:    "'a' as any",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `string`.",
				},
			},
		},
		{
			sourceText: "type T = [number, T[]];\ndeclare function foo(t: T): void;\ndeclare const t: T;\nfoo(t as any);\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "t as any",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `T`.",
				},
			},
		},
		{
			sourceText: "function foo(\n  templates: TemplateStringsArray,\n  arg1: number,\n  arg2: any,\n  arg3: string,\n) {}\ndeclare const arg: any;\nfoo<number>`${arg}${arg}${arg}`;\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "arg",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `number`.",
				},
				{
					wantSpan:    "arg",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `string`.",
				},
			},
		},
		{
			sourceText: "function foo(templates: TemplateStringsArray, arg: number) {}\ndeclare const arg: any;\nfoo`${arg}`;\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "arg",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `number`.",
				},
			},
		},
		{
			sourceText: "type T = [number, T[]];\nfunction foo(templates: TemplateStringsArray, arg: T) {}\ndeclare const arg: any;\nfoo`${arg}`;\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "arg",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `T`.",
				},
			},
		},

		{
			// The other half of the pair above, and the reason it is not enough to assert the three
			// silences. A FIXED tuple carries no variable element, so nothing is consumed and the
			// trailing argument does get its parameter. A port that simply always consumed after a
			// tuple spread would pass all three silent rows and fail this one.
			sourceText: "declare function foo(a: number, b: string): void;\ndeclare const x: [number];\nfoo(...x, 1 as any);\n",
			wantFindings: []noUnsafeArgumentFinding{
				{
					wantSpan:    "1 as any",
					wantId:      "unsafeArgument",
					wantMessage: "Unsafe argument of type `any` assigned to a parameter of type `string`.",
				},
			},
		},
	}
	for index, testCase := range cases {
		t.Run(noUnsafeArgumentCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnsafeArgument, noUnsafeArgumentFile,
				testCase.sourceText)

			wantIds := make([]string, len(testCase.wantFindings))
			for position, want := range testCase.wantFindings {
				wantIds[position] = want.wantId
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			// The harness writes the fixture trimmed, so the span slices are against that text
			// rather than the Go literal above. The captured spans were taken from a file written
			// the same way, so the two agree by construction rather than by luck.
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
