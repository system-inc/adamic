package typescript

import (
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const preferFindFile = "/repository/source/Searching.ts"

func preferFindCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

type preferFindFinding struct {
	wantSpan       string
	wantSuggestion string
}

// applyPreferFindSuggestion rewrites source with one suggestion's fixes, latest range first.
//
// The harness applies fixes and has no suggestion support, and this rule's entire repair surface is
// a suggestion, so without this the thing it offers would go unasserted. The ordering is
// load-bearing here rather than decorative: a ternary case carries THREE edits over three ranges in
// one expression, so applying them front to back would shift the two that follow.
func applyPreferFindSuggestion(t *testing.T, source string, suggestion rule.Suggestion) string {
	t.Helper()
	fixes := append([]rule.Fix(nil), suggestion.Fixes...)
	sort.Slice(fixes, func(first, second int) bool { return fixes[first].Range.Pos() > fixes[second].Range.Pos() })
	for _, fix := range fixes {
		start, end := fix.Range.Pos(), fix.Range.End()
		if start < 0 || end > len(source) || start > end {
			t.Fatalf("the suggestion proposes an out-of-range edit [%d,%d) over %d bytes", start, end, len(source))
		}
		source = source[:start] + fix.Text + source[end:]
	}
	return source
}

// TestPreferFindStaysSilentOnUpstreamPassCases is the imported clean corpus, verbatim.
//
// All nineteen of upstream's passing inputs, extracted by parsing the clone's test file with the
// TypeScript compiler. Every one was run through the installed 8.x build against a real program,
// which reported nothing and produced no parse error on any of them.
func TestPreferFindStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []string{
		"\ninterface JerkCode<T> {\n  filter(predicate: (item: T) => boolean): JerkCode<T>;\n}\n\ndeclare const jerkCode: JerkCode<string>;\n\njerkCode.filter(item => item === 'aha')[0];\n    ",
		"\ndeclare const arr: readonly string[];\narr.filter(item => item === 'aha')[1];\n    ",
		"\ndeclare const arr: string[];\narr.filter(item => item === 'aha').at(1);\n    ",
		"\ndeclare const notNecessarilyAnArray: unknown[] | undefined | null | string;\nnotNecessarilyAnArray?.filter(item => true)[0];\n    ",
		"[].filter(() => true)?.[0];",
		"[].filter(() => true)?.at?.(0);",
		"[].filter?.(() => true)[0];",
		"[1, 2, 3].filter(x => x > 0).at(-Infinity);",
		"\ndeclare const arr: string[];\ndeclare const cond: Parameters<Array<string>['filter']>[0];\nconst a = { arr };\na?.arr.filter(cond).at(1);\n    ",
		"['Just', 'a', 'filter'].filter(x => x.length > 4);",
		"['Just', 'a', 'find'].find(x => x.length > 4);",
		"undefined.filter(x => x)[0];",
		"null?.filter(x => x)[0];",
		"\ndeclare function foo(param: any): any;\nfoo(Symbol.for('foo'));\n    ",
		"\ndeclare const arr: string[];\nconst s = Symbol.for(\"Don't throw!\");\narr.filter(item => item === 'aha').at(s);\n    ",
		"[1, 2, 3].filter(x => x)[Symbol('0')];",
		"[1, 2, 3].filter(x => x)[Symbol.for('0')];",
		"(Math.random() < 0.5 ? [1, 2, 3].filter(x => true) : [1, 2, 3])[0];",
		"\n(Math.random() < 0.5\n  ? [1, 2, 3].find(x => true)\n  : [1, 2, 3].filter(x => true))[0];\n    ",
	}
	for index, sourceText := range cases {
		t.Run(preferFindCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, PreferFind,
				preferFindFile, sourceText))
		})
	}
}

// TestPreferFindFiresOnUpstreamFailCases is the imported reporting corpus, verbatim.
//
// All twenty-eight of upstream's failing inputs. Each span is measured from the installed build's
// reported range and each expected rewrite is that build's own suggestion replayed against the
// input; all twenty-eight were separately confirmed equal to the `output` fields the corpus
// records, so the running rule and the checked-in corpus agree.
//
// The harness writes each fixture as `strings.TrimSpace(source)+"\n"`, so both the span slices and
// the expected rewrites are against that trimmed text rather than the Go literal.
func TestPreferFindFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		wantFindings []preferFindFinding
	}{
		{
			sourceText: "\ndeclare const arr: string[];\narr.filter(item => item === 'aha')[0];\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "arr.filter(item => item === 'aha')[0]", wantSuggestion: "declare const arr: string[];\narr.find(item => item === 'aha');\n"},
			},
		},
		{
			sourceText: "\ndeclare const arr: Array<string>;\nconst zero = 0;\narr.filter(item => item === 'aha')[zero];\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "arr.filter(item => item === 'aha')[zero]", wantSuggestion: "declare const arr: Array<string>;\nconst zero = 0;\narr.find(item => item === 'aha');\n"},
			},
		},
		{
			sourceText: "\ndeclare const arr: Array<string>;\nconst zero = 0n;\narr.filter(item => item === 'aha')[zero];\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "arr.filter(item => item === 'aha')[zero]", wantSuggestion: "declare const arr: Array<string>;\nconst zero = 0n;\narr.find(item => item === 'aha');\n"},
			},
		},
		{
			sourceText: "\ndeclare const arr: Array<string>;\nconst zero = -0n;\narr.filter(item => item === 'aha')[zero];\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "arr.filter(item => item === 'aha')[zero]", wantSuggestion: "declare const arr: Array<string>;\nconst zero = -0n;\narr.find(item => item === 'aha');\n"},
			},
		},
		{
			sourceText: "\ndeclare const arr: readonly string[];\narr.filter(item => item === 'aha').at(0);\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "arr.filter(item => item === 'aha').at(0)", wantSuggestion: "declare const arr: readonly string[];\narr.find(item => item === 'aha');\n"},
			},
		},
		{
			sourceText: "\ndeclare const arr: ReadonlyArray<string>;\n(undefined, arr.filter(item => item === 'aha')).at(0);\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "(undefined, arr.filter(item => item === 'aha')).at(0)", wantSuggestion: "declare const arr: ReadonlyArray<string>;\n(undefined, arr.find(item => item === 'aha'));\n"},
			},
		},
		{
			sourceText: "\ndeclare const arr: string[];\nconst zero = 0;\narr.filter(item => item === 'aha').at(zero);\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "arr.filter(item => item === 'aha').at(zero)", wantSuggestion: "declare const arr: string[];\nconst zero = 0;\narr.find(item => item === 'aha');\n"},
			},
		},
		{
			sourceText: "\ndeclare const arr: string[];\narr.filter(item => item === 'aha')['0'];\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "arr.filter(item => item === 'aha')['0']", wantSuggestion: "declare const arr: string[];\narr.find(item => item === 'aha');\n"},
			},
		},
		{
			sourceText: "const two = [1, 2, 3].filter(item => item === 2)[0];",
			wantFindings: []preferFindFinding{
				{wantSpan: "[1, 2, 3].filter(item => item === 2)[0]", wantSuggestion: "const two = [1, 2, 3].find(item => item === 2);\n"},
			},
		},
		{
			sourceText: "const fltr = \"filter\"; (([] as unknown[]))[fltr] ((item) => { return item === 2 }  ) [ 0  ] ;",
			wantFindings: []preferFindFinding{
				{wantSpan: "(([] as unknown[]))[fltr] ((item) => { return item === 2 }  ) [ 0  ]", wantSuggestion: "const fltr = \"filter\"; (([] as unknown[]))[\"find\"] ((item) => { return item === 2 }  )  ;\n"},
			},
		},
		{
			sourceText: "(([] as unknown[]))?.[\"filter\"] ((item) => { return item === 2 }  ) [ 0  ] ;",
			wantFindings: []preferFindFinding{
				{wantSpan: "(([] as unknown[]))?.[\"filter\"] ((item) => { return item === 2 }  ) [ 0  ]", wantSuggestion: "(([] as unknown[]))?.[\"find\"] ((item) => { return item === 2 }  )  ;\n"},
			},
		},
		{
			sourceText: "\ndeclare const nullableArray: unknown[] | undefined | null;\nnullableArray?.filter(item => true)[0];\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "nullableArray?.filter(item => true)[0]", wantSuggestion: "declare const nullableArray: unknown[] | undefined | null;\nnullableArray?.find(item => true);\n"},
			},
		},
		{
			sourceText: "([]?.filter(f))[0];",
			wantFindings: []preferFindFinding{
				{wantSpan: "([]?.filter(f))[0]", wantSuggestion: "([]?.find(f));\n"},
			},
		},
		{
			sourceText: "\ndeclare const objectWithArrayProperty: { arr: unknown[] };\ndeclare function cond(x: unknown): boolean;\nconsole.log((1, 2, objectWithArrayProperty?.arr['filter'](cond)).at(0));\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "(1, 2, objectWithArrayProperty?.arr['filter'](cond)).at(0)", wantSuggestion: "declare const objectWithArrayProperty: { arr: unknown[] };\ndeclare function cond(x: unknown): boolean;\nconsole.log((1, 2, objectWithArrayProperty?.arr[\"find\"](cond)));\n"},
			},
		},
		{
			sourceText: "\n[1, 2, 3].filter(x => x > 0).at(NaN);\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "[1, 2, 3].filter(x => x > 0).at(NaN)", wantSuggestion: "[1, 2, 3].find(x => x > 0);\n"},
			},
		},
		{
			sourceText: "\nconst idxToLookUp = -0.12635678;\n[1, 2, 3].filter(x => x > 0).at(idxToLookUp);\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "[1, 2, 3].filter(x => x > 0).at(idxToLookUp)", wantSuggestion: "const idxToLookUp = -0.12635678;\n[1, 2, 3].find(x => x > 0);\n"},
			},
		},
		{
			sourceText: "\n[1, 2, 3].filter(x => x > 0)[`at`](0);\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "[1, 2, 3].filter(x => x > 0)[`at`](0)", wantSuggestion: "[1, 2, 3].find(x => x > 0);\n"},
			},
		},
		{
			sourceText: "\ndeclare const arr: string[];\ndeclare const cond: Parameters<Array<string>['filter']>[0];\nconst a = { arr };\na?.arr\n  .filter(cond) /* what a bad spot for a comment. Let's make sure\n  there's some yucky symbols too. [ . ?. <>   ' ' \\'] */\n  .at('0');\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "a?.arr\n  .filter(cond) /* what a bad spot for a comment. Let's make sure\n  there's some yucky symbols too. [ . ?. <>   ' ' \\'] */\n  .at('0')", wantSuggestion: "declare const arr: string[];\ndeclare const cond: Parameters<Array<string>['filter']>[0];\nconst a = { arr };\na?.arr\n  .find(cond) /* what a bad spot for a comment. Let's make sure\n  there's some yucky symbols too. [ . ?. <>   ' ' \\'] */\n  ;\n"},
			},
		},
		{
			sourceText: "\nconst imNotActuallyAnArray = [\n  [1, 2, 3],\n  [2, 3, 4],\n] as const;\nconst butIAm = [4, 5, 6];\nbutIAm.push(\n  // line comment!\n  ...imNotActuallyAnArray[/* comment */ 'filter' /* another comment */](\n    x => x[1] > 0,\n  ) /**/[`0`]!,\n);\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "imNotActuallyAnArray[/* comment */ 'filter' /* another comment */](\n    x => x[1] > 0,\n  ) /**/[`0`]", wantSuggestion: "const imNotActuallyAnArray = [\n  [1, 2, 3],\n  [2, 3, 4],\n] as const;\nconst butIAm = [4, 5, 6];\nbutIAm.push(\n  // line comment!\n  ...imNotActuallyAnArray[/* comment */ \"find\" /* another comment */](\n    x => x[1] > 0,\n  ) /**/!,\n);\n"},
			},
		},
		{
			sourceText: "\nfunction actingOnArray<T extends string[]>(values: T) {\n  return values.filter(filter => filter === 'filter')[\n    /* filter */ -0.0 /* filter */\n  ];\n}\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "values.filter(filter => filter === 'filter')[\n    /* filter */ -0.0 /* filter */\n  ]", wantSuggestion: "function actingOnArray<T extends string[]>(values: T) {\n  return values.find(filter => filter === 'filter');\n}\n"},
			},
		},
		{
			sourceText: "\nconst nestedSequenceAbomination =\n  (1,\n  2,\n  (1,\n  2,\n  3,\n  (1, 2, 3, 4),\n  (1, 2, 3, 4, 5, [1, 2, 3, 4, 5, 6].filter(x => x % 2 == 0)))['0']);\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "(1,\n  2,\n  3,\n  (1, 2, 3, 4),\n  (1, 2, 3, 4, 5, [1, 2, 3, 4, 5, 6].filter(x => x % 2 == 0)))['0']", wantSuggestion: "const nestedSequenceAbomination =\n  (1,\n  2,\n  (1,\n  2,\n  3,\n  (1, 2, 3, 4),\n  (1, 2, 3, 4, 5, [1, 2, 3, 4, 5, 6].find(x => x % 2 == 0))));\n"},
			},
		},
		{
			sourceText: "\ndeclare const arr: { a: 1 }[] & { b: 2 }[];\narr.filter(f, thisArg)[0];\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "arr.filter(f, thisArg)[0]", wantSuggestion: "declare const arr: { a: 1 }[] & { b: 2 }[];\narr.find(f, thisArg);\n"},
			},
		},
		{
			sourceText: "\ndeclare const arr: { a: 1 }[] & ({ b: 2 }[] | { c: 3 }[]);\narr.filter(f, thisArg)[0];\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "arr.filter(f, thisArg)[0]", wantSuggestion: "declare const arr: { a: 1 }[] & ({ b: 2 }[] | { c: 3 }[]);\narr.find(f, thisArg);\n"},
			},
		},
		{
			sourceText: "\n(Math.random() < 0.5\n  ? [1, 2, 3].filter(x => false)\n  : [1, 2, 3].filter(x => true))[0];\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "(Math.random() < 0.5\n  ? [1, 2, 3].filter(x => false)\n  : [1, 2, 3].filter(x => true))[0]", wantSuggestion: "(Math.random() < 0.5\n  ? [1, 2, 3].find(x => false)\n  : [1, 2, 3].find(x => true));\n"},
			},
		},
		{
			sourceText: "\nMath.random() < 0.5\n  ? [1, 2, 3].find(x => true)\n  : [1, 2, 3].filter(x => true)[0];\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "[1, 2, 3].filter(x => true)[0]", wantSuggestion: "Math.random() < 0.5\n  ? [1, 2, 3].find(x => true)\n  : [1, 2, 3].find(x => true);\n"},
			},
		},
		{
			sourceText: "\ndeclare const f: (arg0: unknown, arg1: number, arg2: Array<unknown>) => boolean,\n  g: (arg0: unknown) => boolean;\nconst nestedTernaries = (\n  Math.random() < 0.5\n    ? Math.random() < 0.5\n      ? [1, 2, 3].filter(f)\n      : []?.filter(x => 'shrug')\n    : [2, 3, 4]['filter'](g)\n).at(0.2);\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "(\n  Math.random() < 0.5\n    ? Math.random() < 0.5\n      ? [1, 2, 3].filter(f)\n      : []?.filter(x => 'shrug')\n    : [2, 3, 4]['filter'](g)\n).at(0.2)", wantSuggestion: "declare const f: (arg0: unknown, arg1: number, arg2: Array<unknown>) => boolean,\n  g: (arg0: unknown) => boolean;\nconst nestedTernaries = (\n  Math.random() < 0.5\n    ? Math.random() < 0.5\n      ? [1, 2, 3].find(f)\n      : []?.find(x => 'shrug')\n    : [2, 3, 4][\"find\"](g)\n);\n"},
			},
		},
		{
			sourceText: "\ndeclare const f: (arg0: unknown) => boolean, g: (arg0: unknown) => boolean;\nconst nestedTernariesWithSequenceExpression = (\n  Math.random() < 0.5\n    ? ('sequence',\n      'expression',\n      Math.random() < 0.5 ? [1, 2, 3].filter(f) : []?.filter(x => 'shrug'))\n    : [2, 3, 4]['filter'](g)\n).at(0.2);\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "(\n  Math.random() < 0.5\n    ? ('sequence',\n      'expression',\n      Math.random() < 0.5 ? [1, 2, 3].filter(f) : []?.filter(x => 'shrug'))\n    : [2, 3, 4]['filter'](g)\n).at(0.2)", wantSuggestion: "declare const f: (arg0: unknown) => boolean, g: (arg0: unknown) => boolean;\nconst nestedTernariesWithSequenceExpression = (\n  Math.random() < 0.5\n    ? ('sequence',\n      'expression',\n      Math.random() < 0.5 ? [1, 2, 3].find(f) : []?.find(x => 'shrug'))\n    : [2, 3, 4][\"find\"](g)\n);\n"},
			},
		},
		{
			sourceText: "\ndeclare const spreadArgs: [(x: unknown) => boolean];\n[1, 2, 3].filter(...spreadArgs).at(0);\n      ",
			wantFindings: []preferFindFinding{
				{wantSpan: "[1, 2, 3].filter(...spreadArgs).at(0)", wantSuggestion: "declare const spreadArgs: [(x: unknown) => boolean];\n[1, 2, 3].find(...spreadArgs);\n"},
			},
		},
	}
	for index, testCase := range cases {
		t.Run(preferFindCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferFind, preferFindFile, testCase.sourceText)

			wantIds := make([]string, len(testCase.wantFindings))
			for position := range wantIds {
				wantIds[position] = "preferFind"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			for position, want := range testCase.wantFindings {
				diagnostic := result.Diagnostics[position]

				gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != want.wantSpan {
					t.Fatalf("finding %d span: expected %q, got %q", position, want.wantSpan, gotSpan)
				}
				if diagnostic.Message.Description != "Prefer .find(...) instead of .filter(...)[0]." {
					t.Fatalf("finding %d message: got %q", position, diagnostic.Message.Description)
				}

				// The repair is offered, never applied. `filter(...)[0]` and `find(...)` produce
				// the same value and DIFFERENT types, so an unattended rewrite can break a build.
				if len(diagnostic.Fixes) != 0 {
					t.Fatalf("finding %d: the repair must be a suggestion, got %d fixes",
						position, len(diagnostic.Fixes))
				}
				if len(diagnostic.Suggestions) != 1 {
					t.Fatalf("finding %d: expected one suggestion, got %d", position, len(diagnostic.Suggestions))
				}
				suggestion := diagnostic.Suggestions[0]
				if suggestion.Message.Description != "Use .find(...) instead of .filter(...)[0]." {
					t.Fatalf("finding %d suggestion text: got %q", position, suggestion.Message.Description)
				}
				applied := applyPreferFindSuggestion(t, onDisk, suggestion)
				if applied != want.wantSuggestion {
					t.Fatalf("finding %d suggestion applied: expected %q, got %q",
						position, want.wantSuggestion, applied)
				}
			}
		})
	}
}

// TestPreferFindNeedsTheTypedHarness pins the checker declaration.
//
// Under the plain harness the checker is nil and every listener returns immediately, so the clean
// fixtures would pass having proven nothing.
func TestPreferFindNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !PreferFind.NeedsTypeChecker {
		t.Fatal("the rule resolves the filtered receiver's type, so it must declare NeedsTypeChecker")
	}

	rule_testing.ExpectClean(t, rule_testing.Run(t, PreferFind, preferFindFile,
		"declare const arr: string[];\narr.filter(item => item === 'a')[0];"))
}

// TestPreferFindStaysSilentOnShapesTheCorpusDoesNotWrite closes two mutation blind spots.
//
// Upstream's corpus writes no other single-argument array method after a filter, and no `at` call
// with the wrong number of arguments, so nothing in it can see either the member-name test or the
// argument-count test. Mutants removing each survived all forty-seven imported cases.
//
// The three method rows separate the name test: without it a filtered array followed by any
// one-argument call whose argument folds to zero would report. The two `at` rows separate the
// count test, which upstream writes as `arguments.length !== 1`.
//
// All five were measured clean on the installed 8.x build before being written here.
func TestPreferFindStaysSilentOnShapesTheCorpusDoesNotWrite(t *testing.T) {
	t.Parallel()

	cases := []string{
		"declare const arr: string[];\narr.filter(x => true).indexOf('a');",
		"declare const arr: string[];\narr.filter(x => true).includes('a');",
		"declare const arr: string[];\narr.filter(x => true).join('');",
		"declare const arr: string[];\narr.filter(x => true).at();",
		"declare const arr: string[];\narr.filter(x => true).at(0, 1);",
	}
	for index, sourceText := range cases {
		t.Run(preferFindCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, PreferFind, preferFindFile, sourceText))
		})
	}
}
