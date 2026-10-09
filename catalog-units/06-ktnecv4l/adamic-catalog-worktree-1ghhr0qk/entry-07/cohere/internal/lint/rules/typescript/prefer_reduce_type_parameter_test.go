package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const preferReduceTypeParameterFile = "/repository/source/Reducing.ts"

func preferReduceTypeParameterCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// preferReduceTypeParameterModules supplies the one module upstream's corpus imports.
//
// Upstream's passing case writes `import { Reducable } from './class'` and relies on its own
// fixture project resolving it. An unresolved import would make the receiver `any`, which is
// silent for a different reason than the one the case is testing, so the module is supplied and
// the case tests what it was written to test: a user-defined `reduce` that is not an array.
var preferReduceTypeParameterModules = map[string]string{
	"/repository/source/class.ts": "export class Reducable {\n  reduce(callback: (argument: any) => any, argument: any): any {\n    return argument;\n  }\n}\n",
}

// TestPreferReduceTypeParameterStaysSilent is upstream's seventeen passing cases verbatim, plus
// the divergences this port measured against the installed build.
//
// Every one of these was measured ONE FILE PER PROGRAM against @typescript-eslint 8.67.0. That
// detail is load-bearing: run as a single program the corpus contradicts itself, because these
// files carry no import or export and so are global scripts whose `declare const tuple` merges
// across files. Under that merge upstream's valid case 10 reports, which reads exactly like a
// wrong case in upstream's own corpus and is an artifact of how it was measured.
func TestPreferReduceTypeParameterStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"new (class Mine {\n  reduce() {}\n})().reduce(() => {}, 1 as any);\n",
		"class Mine {\n  reduce() {}\n}\n\nnew Mine().reduce(() => {}, 1 as any);\n",
		"import { Reducable } from './class';\n\nnew Reducable().reduce(() => {}, 1 as any);\n",
		"[1, 2, 3]['reduce']((sum, num) => sum + num, 0);\n",
		"[1, 2, 3][null]((sum, num) => sum + num, 0);\n",
		"[1, 2, 3]?.[null]((sum, num) => sum + num, 0);\n",
		"[1, 2, 3].reduce((sum, num) => sum + num, 0);\n",
		"[1, 2, 3].reduce<number[]>((a, s) => a.concat(s * 2), []);\n",
		"[1, 2, 3]?.reduce<number[]>((a, s) => a.concat(s * 2), []);\n",
		"declare const tuple: [number, number, number];\ntuple.reduce<number[]>((a, s) => a.concat(s * 2), []);\n",
		"type Reducer = { reduce: (callback: (arg: any) => any, arg: any) => any };\ndeclare const tuple: [number, number, number] | Reducer;\ntuple.reduce(a => {\n  return a.concat(1);\n}, [] as number[]);\n",
		"type Reducer = { reduce: (callback: (arg: any) => any, arg: any) => any };\ndeclare const arrayOrReducer: number[] & Reducer;\narrayOrReducer.reduce(a => {\n  return a.concat(1);\n}, [] as number[]);\n",
		"['a', 'b'].reduce(\n  (accum, name) => ({\n    ...accum,\n    [name]: true,\n  }),\n  {} as Record<'a' | 'b', boolean>,\n);\n",
		"['a', 'b'].reduce(\n  (accum, name) => ({\n    ...accum,\n    [name]: true,\n  }),\n  { a: true, b: false, c: true } as Record<'a' | 'b', boolean>,\n);\n",
		"function f<T extends Record<string, boolean>>() {\n  ['a', 'b'].reduce(\n    (accum, name) => ({\n      ...accum,\n      [name]: true,\n    }),\n    {} as T,\n  );\n}\n",
		"function f<T>() {\n  ['a', 'b'].reduce(\n    (accum, name) => ({\n      ...accum,\n      [name]: true,\n    }),\n    {} as T,\n  );\n}\n",
		"['a', 'b'].reduce((accum, name) => `${accum} | hello ${name}!`);\n",

		// Measured divergences and boundary cases upstream does not write. Each is silent on the
		// installed build and silent here, and each pins a decision the corpus leaves untested.

		// `satisfies` checks without changing the type, so there is nothing unnecessary to strip.
		"declare const a: number[];\na.reduce((x, s) => x.concat(s), [] satisfies number[]);\n",

		// A spread argument is not a second argument the rule can read.
		"declare const a: number[];\na.reduce((x, s) => x.concat(s), ...([[] as number[]] as [number[]]));\n",

		// A computed key folded through a const binding. This one REPORTS upstream and is silent
		// here: it needs ESLint's scope-level constant folding, which this tree does not have. The
		// case is pinned as silent so the divergence is a fixture rather than a doc comment nobody
		// re-measures. Measured on the installed build, which rewrites it to `a[key]<number[]>(...)`.
		"declare const a: number[];\nconst reduceKey = 'reduce';\na[reduceKey]((x, s) => x.concat(s), [] as number[]);\n",

		// `as const` REPORTS upstream and is declined here. Upstream's own fixer writes
		// `reduce<const>`, which does not parse, and this rule's repair is a fix rather than a
		// suggestion, so it would be applied with nobody watching. Measured on the installed build.
		"declare const a: number[];\na.reduce((x, s) => x.concat(s), [] as const);\n",

		// One argument only, so upstream's `arguments.length < 2` test declines before the
		// assertion is ever read.
		"declare const a: number[];\na.reduce((x, s) => x + s);\n",

		// A member access that is not `reduce`, with an assertion in the same position.
		"declare const a: number[];\na.reduceRight((x, s) => x.concat(s), [] as number[]);\n",
	}
	for index, sourceText := range cases {
		t.Run(preferReduceTypeParameterCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedFiles(t, PreferReduceTypeParameter,
				preferReduceTypeParameterFilesFor(sourceText), preferReduceTypeParameterFile))
		})
	}
}

// preferReduceTypeParameterFilesFor puts the fixture beside the module the corpus imports.
func preferReduceTypeParameterFilesFor(sourceText string) map[string]string {
	files := map[string]string{preferReduceTypeParameterFile: sourceText}
	for name, contents := range preferReduceTypeParameterModules {
		files[name] = contents
	}
	return files
}

// TestPreferReduceTypeParameterFires is upstream's fourteen reporting cases verbatim.
//
// Every row asserts the span, the whole rendered message, and the rewritten file. The rewrite
// matters most: this rule assembles up to THREE edits per finding and the third is conditional on
// the call already carrying a type argument, so a message-id assertion cannot see a repair that
// strips the assertion and forgets the type parameter, which is source that still parses and no
// longer type-checks.
//
// The expected rewrites are upstream's own `output` fields, re-measured against the installed
// build over the TRIMMED source, because the harness writes each fixture as
// `strings.TrimSpace(source)+"\n"` and comparing against the untrimmed literal fails on the
// trailing bytes alone.
func TestPreferReduceTypeParameterFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		wantSpan    string
		wantMessage string
		wantFixed   string
	}{
		{
			sourceText:  "declare const arr: string[];\narr.reduce<string | undefined>(acc => acc, arr.shift() as string | undefined);\n",
			wantSpan:    "arr.shift() as string | undefined",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "declare const arr: string[];\narr.reduce<string | undefined>(acc => acc, arr.shift());\n",
		},
		{
			sourceText:  "[1, 2, 3].reduce((a, s) => a.concat(s * 2), [] as number[]);\n",
			wantSpan:    "[] as number[]",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "[1, 2, 3].reduce<number[]>((a, s) => a.concat(s * 2), []);\n",
		},
		{
			sourceText:  "[1, 2, 3].reduce((a, s) => a.concat(s * 2), <number[]>[]);\n",
			wantSpan:    "<number[]>[]",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "[1, 2, 3].reduce<number[]>((a, s) => a.concat(s * 2), []);\n",
		},
		{
			sourceText:  "[1, 2, 3]?.reduce((a, s) => a.concat(s * 2), [] as number[]);\n",
			wantSpan:    "[] as number[]",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "[1, 2, 3]?.reduce<number[]>((a, s) => a.concat(s * 2), []);\n",
		},
		{
			sourceText:  "[1, 2, 3]?.reduce((a, s) => a.concat(s * 2), <number[]>[]);\n",
			wantSpan:    "<number[]>[]",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "[1, 2, 3]?.reduce<number[]>((a, s) => a.concat(s * 2), []);\n",
		},
		{
			sourceText:  "const names = ['a', 'b', 'c'];\n\nnames.reduce(\n  (accum, name) => ({\n    ...accum,\n    [name]: true,\n  }),\n  {} as Record<string, boolean>,\n);\n",
			wantSpan:    "{} as Record<string, boolean>",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "const names = ['a', 'b', 'c'];\n\nnames.reduce<Record<string, boolean>>(\n  (accum, name) => ({\n    ...accum,\n    [name]: true,\n  }),\n  {},\n);\n",
		},
		{
			sourceText:  "['a', 'b'].reduce(\n  (accum, name) => ({\n    ...accum,\n    [name]: true,\n  }),\n  <Record<string, boolean>>{},\n);\n",
			wantSpan:    "<Record<string, boolean>>{}",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "['a', 'b'].reduce<Record<string, boolean>>(\n  (accum, name) => ({\n    ...accum,\n    [name]: true,\n  }),\n  {},\n);\n",
		},
		{
			sourceText:  "['a', 'b']['reduce'](\n  (accum, name) => ({\n    ...accum,\n    [name]: true,\n  }),\n  {} as Record<string, boolean>,\n);\n",
			wantSpan:    "{} as Record<string, boolean>",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "['a', 'b']['reduce']<Record<string, boolean>>(\n  (accum, name) => ({\n    ...accum,\n    [name]: true,\n  }),\n  {},\n);\n",
		},
		{
			sourceText:  "function f<T, U extends T[]>(a: U) {\n  return a.reduce(() => {}, {} as Record<string, boolean>);\n}\n",
			wantSpan:    "{} as Record<string, boolean>",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "function f<T, U extends T[]>(a: U) {\n  return a.reduce<Record<string, boolean>>(() => {}, {});\n}\n",
		},
		{
			sourceText:  "declare const tuple: [number, number, number];\ntuple.reduce((a, s) => a.concat(s * 2), [] as number[]);\n",
			wantSpan:    "[] as number[]",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "declare const tuple: [number, number, number];\ntuple.reduce<number[]>((a, s) => a.concat(s * 2), []);\n",
		},
		{
			sourceText:  "declare const tupleOrArray: [number, number, number] | number[];\ntupleOrArray.reduce((a, s) => a.concat(s * 2), [] as number[]);\n",
			wantSpan:    "[] as number[]",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "declare const tupleOrArray: [number, number, number] | number[];\ntupleOrArray.reduce<number[]>((a, s) => a.concat(s * 2), []);\n",
		},
		{
			sourceText:  "declare const tuple: [number, number, number] & number[];\ntuple.reduce((a, s) => a.concat(s * 2), [] as number[]);\n",
			wantSpan:    "[] as number[]",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "declare const tuple: [number, number, number] & number[];\ntuple.reduce<number[]>((a, s) => a.concat(s * 2), []);\n",
		},
		{
			sourceText:  "['a', 'b'].reduce(\n  (accum, name) => ({\n    ...accum,\n    [name]: true,\n  }),\n  {} as Record<string, boolean>,\n);\n",
			wantSpan:    "{} as Record<string, boolean>",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "['a', 'b'].reduce<Record<string, boolean>>(\n  (accum, name) => ({\n    ...accum,\n    [name]: true,\n  }),\n  {},\n);\n",
		},
		{
			sourceText:  "function f<T extends Record<string, boolean>>(t: T) {\n  ['a', 'b'].reduce(\n    (accum, name) => ({\n      ...accum,\n      [name]: true,\n    }),\n    t as Record<string, boolean | number>,\n  );\n}\n",
			wantSpan:    "t as Record<string, boolean | number>",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "function f<T extends Record<string, boolean>>(t: T) {\n  ['a', 'b'].reduce<Record<string, boolean | number>>(\n    (accum, name) => ({\n      ...accum,\n      [name]: true,\n    }),\n    t,\n  );\n}\n",
		},

		// Measured additions. Upstream's corpus writes no parenthesized form at all, and both of
		// these were measured against the installed build rather than reasoned about.
		{
			// A parenthesized callee. estree has no parenthesized node, so upstream's child
			// selector sees straight through it and the insert lands INSIDE the parentheses.
			sourceText:  "declare const a: number[];\n(a.reduce)((x, s) => x.concat(s), [] as number[]);\n",
			wantSpan:    "[] as number[]",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "declare const a: number[];\n(a.reduce<number[]>)((x, s) => x.concat(s), []);\n",
		},
		{
			// A parenthesized second argument. The finding anchors on the inner assertion and the
			// parentheses survive the repair.
			sourceText:  "declare const a: number[];\na.reduce((x, s) => x.concat(s), ([] as number[]));\n",
			wantSpan:    "[] as number[]",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "declare const a: number[];\na.reduce<number[]>((x, s) => x.concat(s), ([]));\n",
		},
		{
			// A third argument does not stop the rule: upstream tests `arguments.length < 2` and
			// never an upper bound.
			sourceText:  "declare const a: number[];\na.reduce((x, s) => x.concat(s), [] as number[], 1);\n",
			wantSpan:    "[] as number[]",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "declare const a: number[];\na.reduce<number[]>((x, s) => x.concat(s), [], 1);\n",
		},
		{
			// A readonly array is still an array to `isArrayType`.
			sourceText:  "declare const a: readonly number[];\na.reduce((x, s) => x.concat(s), [] as number[]);\n",
			wantSpan:    "[] as number[]",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "declare const a: readonly number[];\na.reduce<number[]>((x, s) => x.concat(s), []);\n",
		},
		{
			// An optional computed string key, which the corpus writes only in its non-reporting
			// `?.[null]` form.
			sourceText:  "declare const a: number[];\na?.['reduce']((x, s) => x.concat(s), [] as number[]);\n",
			wantSpan:    "[] as number[]",
			wantMessage: "Unnecessary assertion: Array#reduce accepts a type parameter for the default value.",
			wantFixed:   "declare const a: number[];\na?.['reduce']<number[]>((x, s) => x.concat(s), []);\n",
		},
	}
	for index, testCase := range cases {
		t.Run(preferReduceTypeParameterCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedFiles(t, PreferReduceTypeParameter,
				preferReduceTypeParameterFilesFor(testCase.sourceText), preferReduceTypeParameterFile)

			rule_testing.ExpectFindings(t, result, "preferTypeParameter")

			// The harness writes the fixture trimmed, so both the span slice and the expected
			// rewrite are against that text rather than the Go literal above.
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			diagnostic := result.Diagnostics[0]

			gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Fatalf("span: expected %q, got %q", testCase.wantSpan, gotSpan)
			}
			if diagnostic.Message.Description != testCase.wantMessage {
				t.Fatalf("message: expected %q, got %q", testCase.wantMessage,
					diagnostic.Message.Description)
			}
			if diagnostic.Message.Id != "preferTypeParameter" {
				t.Fatalf("message id: expected %q, got %q", "preferTypeParameter", diagnostic.Message.Id)
			}

			rule_testing.ExpectFixedSource(t, result, strings.TrimSpace(testCase.wantFixed)+"\n")
		})
	}
}
