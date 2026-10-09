package typescript

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const consistentIndexedObjectStyleFile = "/repository/source/Indexing.ts"

func consistentIndexedObjectStyleCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// consistentIndexedObjectStyleFinding is one expected finding with every layer it can be wrong at.
//
// An empty wantSuggestion means the row must arrive as a fix or with no repair at all, which the
// body below distinguishes by looking at wantFixed. Keeping both fields is the point: this rule
// emits a fix, a suggestion, or nothing depending on comments and modifiers, and collapsing the
// three would let an unattended rewrite through where upstream only offers one.
type consistentIndexedObjectStyleFinding struct {
	wantId         string
	wantSpan       string
	wantMessage    string
	wantSuggestion string
}

func consistentIndexedObjectStyleOptionsFor(t *testing.T, optionsJson string) any {
	t.Helper()
	if optionsJson == "" {
		return DefaultConsistentIndexedObjectStyleSettings()
	}
	decoded, err := DecodeConsistentIndexedObjectStyleOptions(json.RawMessage(optionsJson))
	if err != nil {
		t.Fatalf("decoding %s: %v", optionsJson, err)
	}
	return decoded
}

// applyConsistentIndexedObjectStyleSuggestion rewrites source with one suggestion's fixes.
func applyConsistentIndexedObjectStyleSuggestion(t *testing.T, source string, suggestion rule.Suggestion) string {
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

// TestConsistentIndexedObjectStyleStaysSilentOnUpstreamPassCases is the imported clean corpus.
//
// All fifty-two of upstream's passing inputs, extracted by parsing the clone's test file with the
// TypeScript compiler. Every one was run through the installed 8.x build, which reported nothing
// and produced no parse error on any of them.
//
// Twenty-four of the fifty-two are circularity cases, which makes that test most of what this rule
// decides rather than an edge.
func TestConsistentIndexedObjectStyleStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		optionsJson string
	}{
		{sourceText: "type Foo = Record<string, any>;", optionsJson: ""},
		{sourceText: "interface Foo {}", optionsJson: ""},
		{sourceText: "\ninterface Foo {\n  bar: string;\n}\n    ", optionsJson: ""},
		{sourceText: "\ninterface Foo {\n  bar: string;\n  [key: string]: any;\n}\n    ", optionsJson: ""},
		{sourceText: "\ninterface Foo {\n  [key: string]: any;\n  bar: string;\n}\n    ", optionsJson: ""},
		{sourceText: "type Foo = { [key: string]: string | Foo };", optionsJson: ""},
		{sourceText: "type Foo = { [key: string]: Foo };", optionsJson: ""},
		{sourceText: "type Foo = { [key: string]: Foo } | Foo;", optionsJson: ""},
		{sourceText: "type Foo = { [key in string]: Foo };", optionsJson: ""},
		{sourceText: "\ninterface Foo {\n  [key: string]: Foo;\n}\n    ", optionsJson: ""},
		{sourceText: "\ninterface Foo<T> {\n  [key: string]: Foo<T>;\n}\n    ", optionsJson: ""},
		{sourceText: "\ninterface Foo<T> {\n  [key: string]: Foo<T> | string;\n}\n    ", optionsJson: ""},
		{sourceText: "\ninterface Foo {\n  [s: string]: Foo & {};\n}\n    ", optionsJson: ""},
		{sourceText: "\ninterface Foo {\n  [s: string]: Foo | string;\n}\n    ", optionsJson: ""},
		{sourceText: "\ninterface Foo<T> {\n  [s: string]: Foo extends T ? string : number;\n}\n    ", optionsJson: ""},
		{sourceText: "\ninterface Foo<T> {\n  [s: string]: T extends Foo ? string : number;\n}\n    ", optionsJson: ""},
		{sourceText: "\ninterface Foo<T> {\n  [s: string]: T extends true ? Foo : number;\n}\n    ", optionsJson: ""},
		{sourceText: "\ninterface Foo<T> {\n  [s: string]: T extends true ? string : Foo;\n}\n    ", optionsJson: ""},
		{sourceText: "\ninterface Foo {\n  [s: string]: Foo[number];\n}\n    ", optionsJson: ""},
		{sourceText: "\ninterface Foo {\n  [s: string]: {}[Foo];\n}\n    ", optionsJson: ""},
		{sourceText: "\ninterface Foo1 {\n  [key: string]: Foo2;\n}\n\ninterface Foo2 {\n  [key: string]: Foo1;\n}\n    ", optionsJson: ""},
		{sourceText: "\ninterface Foo1 {\n  [key: string]: Foo2;\n}\n\ninterface Foo2 {\n  [key: string]: Foo3;\n}\n\ninterface Foo3 {\n  [key: string]: Foo1;\n}\n    ", optionsJson: ""},
		{sourceText: "\ninterface Foo1 {\n  [key: string]: Foo2;\n}\n\ninterface Foo2 {\n  [key: string]: Foo3;\n}\n\ninterface Foo3 {\n  [key: string]: Record<string, Foo1>;\n}\n    ", optionsJson: ""},
		{sourceText: "\ntype Foo1 = {\n  [key: string]: Foo2;\n};\n\ntype Foo2 = {\n  [key: string]: Foo3;\n};\n\ntype Foo3 = {\n  [key: string]: Foo1;\n};\n    ", optionsJson: ""},
		{sourceText: "\ninterface Foo1 {\n  [key: string]: Foo2;\n}\n\ntype Foo2 = {\n  [key: string]: Foo3;\n};\n\ninterface Foo3 {\n  [key: string]: Foo1;\n}\n    ", optionsJson: ""},
		{sourceText: "\ntype Foo1 = {\n  [key: string]: Foo2;\n};\n\ninterface Foo2 {\n  [key: string]: Foo3;\n}\n\ninterface Foo3 {\n  [key: string]: Foo1;\n}\n    ", optionsJson: ""},
		{sourceText: "\ntype ExampleUnion = boolean | number;\n\ntype ExampleRoot = ExampleUnion | ExampleObject;\n\ninterface ExampleObject {\n  [key: string]: ExampleRoot;\n}\n    ", optionsJson: ""},
		{sourceText: "\ntype Bar<K extends string = never> = {\n  [k in K]: Bar;\n};\n    ", optionsJson: ""},
		{sourceText: "\ntype Bar<K extends string = never> = {\n  [k in K]: Foo;\n};\n\ntype Foo = Bar;\n    ", optionsJson: ""},
		{sourceText: "type Foo = {};", optionsJson: ""},
		{sourceText: "\ntype Foo = {\n  bar: string;\n  [key: string]: any;\n};\n    ", optionsJson: ""},
		{sourceText: "\ntype Foo = {\n  bar: string;\n};\n    ", optionsJson: ""},
		{sourceText: "\ntype Foo = {\n  [key: string]: any;\n  bar: string;\n};\n    ", optionsJson: ""},
		{sourceText: "\ntype Foo = Generic<{\n  [key: string]: any;\n  bar: string;\n}>;\n    ", optionsJson: ""},
		{sourceText: "function foo(arg: { [key: string]: any; bar: string }) {}", optionsJson: ""},
		{sourceText: "function foo(): { [key: string]: any; bar: string } {}", optionsJson: ""},
		{sourceText: "type Foo = { [key: string] };", optionsJson: ""},
		{sourceText: "type Foo = { [] };", optionsJson: ""},
		{sourceText: "\ninterface Foo {\n  [key: string];\n}\n    ", optionsJson: ""},
		{sourceText: "\ninterface Foo {\n  [];\n}\n    ", optionsJson: ""},
		{sourceText: "type Foo = Misc<string, unknown>;", optionsJson: "\"index-signature\""},
		{sourceText: "type Foo = Record;", optionsJson: "\"index-signature\""},
		{sourceText: "type Foo = Record<string>;", optionsJson: "\"index-signature\""},
		{sourceText: "type Foo = Record<string, number, unknown>;", optionsJson: "\"index-signature\""},
		{sourceText: "type Foo = { [key: string]: any };", optionsJson: "\"index-signature\""},
		{sourceText: "type Foo = Generic<{ [key: string]: any }>;", optionsJson: "\"index-signature\""},
		{sourceText: "function foo(arg: { [key: string]: any }) {}", optionsJson: "\"index-signature\""},
		{sourceText: "function foo(): { [key: string]: any } {}", optionsJson: "\"index-signature\""},
		{sourceText: "type T = A.B;", optionsJson: "\"index-signature\""},
		{sourceText: "type T = { [key in Foo]: key | number };", optionsJson: ""},
		{sourceText: "\nfunction foo(e: { readonly [key in PropertyKey]-?: key }) {}\n      ", optionsJson: ""},
		{sourceText: "\nfunction f(): {\n  // intentionally not using a Record to preserve optionals\n  [k in keyof ParseResult]: unknown;\n} {\n  return {};\n}\n      ", optionsJson: ""},
	}
	for index, testCase := range cases {
		t.Run(consistentIndexedObjectStyleCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, ConsistentIndexedObjectStyle,
				consistentIndexedObjectStyleFile, testCase.sourceText,
				consistentIndexedObjectStyleOptionsFor(t, testCase.optionsJson)))
		})
	}
}

// TestConsistentIndexedObjectStyleFiresOnUpstreamFailCases is the imported reporting corpus.
//
// All sixty of upstream's failing inputs across both message ids and both option values. Every span
// and rendered message was measured against the installed build, and every expected rewrite is
// upstream's own recorded output transformed the way the harness transforms the input.
//
// Six rows carry no fixed source at all. Those are the deliberate declines: an interface that
// extends something, an export-default subject, and a minus-readonly mapped type that has no Record
// spelling because there is no builtin Mutable. A rule that repaired any of them would satisfy every
// message assertion while writing source that does not compile.
func TestConsistentIndexedObjectStyleFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		optionsJson  string
		wantFixed    string
		wantFindings []consistentIndexedObjectStyleFinding
	}{
		{
			sourceText:  "\ninterface Foo {\n  [key: string]: any;\n}\n      ",
			optionsJson: "",
			wantFixed:   "type Foo = Record<string, any>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface Foo {\n  [key: string]: any;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ninterface Foo {\n  readonly [key: string]: any;\n}\n      ",
			optionsJson: "",
			wantFixed:   "type Foo = Readonly<Record<string, any>>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface Foo {\n  readonly [key: string]: any;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ninterface Foo<A> {\n  [key: string]: A;\n}\n      ",
			optionsJson: "",
			wantFixed:   "type Foo<A> = Record<string, A>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface Foo<A> {\n  [key: string]: A;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ninterface Foo<A = any> {\n  [key: string]: A;\n}\n      ",
			optionsJson: "",
			wantFixed:   "type Foo<A = any> = Record<string, A>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface Foo<A = any> {\n  [key: string]: A;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ninterface B extends A {\n  [index: number]: unknown;\n}\n      ",
			optionsJson: "",
			wantFixed:   "",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface B extends A {\n  [index: number]: unknown;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\nexport default interface Foo {\n  [key: string]: unknown;\n}\n      ",
			optionsJson: "",
			wantFixed:   "",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface Foo {\n  [key: string]: unknown;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ninterface Foo<A> {\n  readonly [key: string]: A;\n}\n      ",
			optionsJson: "",
			wantFixed:   "type Foo<A> = Readonly<Record<string, A>>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface Foo<A> {\n  readonly [key: string]: A;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ninterface Foo<A, B> {\n  [key: A]: B;\n}\n      ",
			optionsJson: "",
			wantFixed:   "type Foo<A, B> = Record<A, B>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface Foo<A, B> {\n  [key: A]: B;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ninterface Foo<A, B> {\n  readonly [key: A]: B;\n}\n      ",
			optionsJson: "",
			wantFixed:   "type Foo<A, B> = Readonly<Record<A, B>>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface Foo<A, B> {\n  readonly [key: A]: B;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "type Foo = { [key: string]: any };",
			optionsJson: "",
			wantFixed:   "type Foo = Record<string, any>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ [key: string]: any }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "type Foo = { readonly [key: string]: any };",
			optionsJson: "",
			wantFixed:   "type Foo = Readonly<Record<string, any>>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ readonly [key: string]: any }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "type Foo = Generic<{ [key: string]: any }>;",
			optionsJson: "",
			wantFixed:   "type Foo = Generic<Record<string, any>>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ [key: string]: any }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "type Foo = Generic<{ readonly [key: string]: any }>;",
			optionsJson: "",
			wantFixed:   "type Foo = Generic<Readonly<Record<string, any>>>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ readonly [key: string]: any }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "function foo(arg: { [key: string]: any }) {}",
			optionsJson: "",
			wantFixed:   "function foo(arg: Record<string, any>) {}\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ [key: string]: any }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "function foo(): { [key: string]: any } {}",
			optionsJson: "",
			wantFixed:   "function foo(): Record<string, any> {}\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ [key: string]: any }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "function foo(arg: { readonly [key: string]: any }) {}",
			optionsJson: "",
			wantFixed:   "function foo(arg: Readonly<Record<string, any>>) {}\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ readonly [key: string]: any }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "function foo(): { readonly [key: string]: any } {}",
			optionsJson: "",
			wantFixed:   "function foo(): Readonly<Record<string, any>> {}\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ readonly [key: string]: any }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "type Foo = Record<string, any>;",
			optionsJson: "\"index-signature\"",
			wantFixed:   "type Foo = { [key: string]: any };\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferIndexSignature", wantSpan: "Record<string, any>", wantMessage: "An index signature is preferred over a record.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "type Foo<T> = Record<string, T>;",
			optionsJson: "\"index-signature\"",
			wantFixed:   "type Foo<T> = { [key: string]: T };\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferIndexSignature", wantSpan: "Record<string, T>", wantMessage: "An index signature is preferred over a record.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "type Foo = { [k: string]: A.Foo };",
			optionsJson: "",
			wantFixed:   "type Foo = Record<string, A.Foo>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ [k: string]: A.Foo }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "type Foo = { [key: string]: AnotherFoo };",
			optionsJson: "",
			wantFixed:   "type Foo = Record<string, AnotherFoo>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ [key: string]: AnotherFoo }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "type Foo = { [key: string]: { [key: string]: Foo } };",
			optionsJson: "",
			wantFixed:   "type Foo = { [key: string]: Record<string, Foo> };\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ [key: string]: Foo }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "type Foo = { [key: string]: string } | Foo;",
			optionsJson: "",
			wantFixed:   "type Foo = Record<string, string> | Foo;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ [key: string]: string }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ninterface Foo<T> {\n  [k: string]: T;\n}\n      ",
			optionsJson: "",
			wantFixed:   "type Foo<T> = Record<string, T>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface Foo<T> {\n  [k: string]: T;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ninterface Foo {\n  [k: string]: A.Foo;\n}\n      ",
			optionsJson: "",
			wantFixed:   "type Foo = Record<string, A.Foo>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface Foo {\n  [k: string]: A.Foo;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ninterface Foo {\n  [k: string]: { [key: string]: Foo };\n}\n      ",
			optionsJson: "",
			wantFixed:   "interface Foo {\n  [k: string]: Record<string, Foo>;\n}\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ [key: string]: Foo }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ninterface Foo {\n  [key: string]: { foo: Foo };\n}\n      ",
			optionsJson: "",
			wantFixed:   "type Foo = Record<string, { foo: Foo }>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface Foo {\n  [key: string]: { foo: Foo };\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ninterface Foo {\n  [key: string]: Foo[];\n}\n      ",
			optionsJson: "",
			wantFixed:   "type Foo = Record<string, Foo[]>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface Foo {\n  [key: string]: Foo[];\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ninterface Foo {\n  [key: string]: () => Foo;\n}\n      ",
			optionsJson: "",
			wantFixed:   "type Foo = Record<string, () => Foo>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface Foo {\n  [key: string]: () => Foo;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ninterface Foo {\n  [s: string]: [Foo];\n}\n      ",
			optionsJson: "",
			wantFixed:   "type Foo = Record<string, [Foo]>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface Foo {\n  [s: string]: [Foo];\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ninterface Foo1 {\n  [key: string]: Foo2;\n}\n\ninterface Foo2 {\n  [key: string]: Foo3;\n}\n\ninterface Foo3 {\n  [key: string]: Foo2;\n}\n      ",
			optionsJson: "",
			wantFixed:   "type Foo1 = Record<string, Foo2>;\n\ninterface Foo2 {\n  [key: string]: Foo3;\n}\n\ninterface Foo3 {\n  [key: string]: Foo2;\n}\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface Foo1 {\n  [key: string]: Foo2;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ninterface Foo1 {\n  [key: string]: Record<string, Foo2>;\n}\n\ninterface Foo2 {\n  [key: string]: Foo3;\n}\n\ninterface Foo3 {\n  [key: string]: Foo2;\n}\n      ",
			optionsJson: "",
			wantFixed:   "type Foo1 = Record<string, Record<string, Foo2>>;\n\ninterface Foo2 {\n  [key: string]: Foo3;\n}\n\ninterface Foo3 {\n  [key: string]: Foo2;\n}\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface Foo1 {\n  [key: string]: Record<string, Foo2>;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ntype Foo1 = {\n  [key: string]: { foo2: Foo2 };\n};\n\ntype Foo2 = {\n  [key: string]: Foo3;\n};\n\ntype Foo3 = {\n  [key: string]: Record<string, Foo1>;\n};\n      ",
			optionsJson: "",
			wantFixed:   "type Foo1 = Record<string, { foo2: Foo2 }>;\n\ntype Foo2 = Record<string, Foo3>;\n\ntype Foo3 = Record<string, Record<string, Foo1>>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{\n  [key: string]: { foo2: Foo2 };\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
				{wantId: "preferRecord", wantSpan: "{\n  [key: string]: Foo3;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
				{wantId: "preferRecord", wantSpan: "{\n  [key: string]: Record<string, Foo1>;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ntype Foos<K extends string = never> = {\n  [k in K]: { foo: Foo };\n};\n\ntype Foo = Foos;\n      ",
			optionsJson: "",
			wantFixed:   "type Foos<K extends string = never> = Record<K, { foo: Foo }>;\n\ntype Foo = Foos;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{\n  [k in K]: { foo: Foo };\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ntype Foos<K extends string = never> = {\n  [k in K]: Foo[];\n};\n\ntype Foo = Foos;\n      ",
			optionsJson: "",
			wantFixed:   "type Foos<K extends string = never> = Record<K, Foo[]>;\n\ntype Foo = Foos;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{\n  [k in K]: Foo[];\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "type Foo = Generic<Record<string, any>>;",
			optionsJson: "\"index-signature\"",
			wantFixed:   "type Foo = Generic<{ [key: string]: any }>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferIndexSignature", wantSpan: "Record<string, any>", wantMessage: "An index signature is preferred over a record.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "type Foo = Record<string | number, any>;",
			optionsJson: "\"index-signature\"",
			wantFixed:   "",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferIndexSignature", wantSpan: "Record<string | number, any>", wantMessage: "An index signature is preferred over a record.", wantSuggestion: "type Foo = { [key: string | number]: any };\n"},
			},
		},
		{
			sourceText:  "type Foo = Record<Exclude<'a' | 'b' | 'c', 'a'>, any>;",
			optionsJson: "\"index-signature\"",
			wantFixed:   "",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferIndexSignature", wantSpan: "Record<Exclude<'a' | 'b' | 'c', 'a'>, any>", wantMessage: "An index signature is preferred over a record.", wantSuggestion: "type Foo = { [key: Exclude<'a' | 'b' | 'c', 'a'>]: any };\n"},
			},
		},
		{
			sourceText:  "type Foo = Record<number, any>;",
			optionsJson: "\"index-signature\"",
			wantFixed:   "type Foo = { [key: number]: any };\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferIndexSignature", wantSpan: "Record<number, any>", wantMessage: "An index signature is preferred over a record.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "type Foo = Record<symbol, any>;",
			optionsJson: "\"index-signature\"",
			wantFixed:   "type Foo = { [key: symbol]: any };\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferIndexSignature", wantSpan: "Record<symbol, any>", wantMessage: "An index signature is preferred over a record.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "function foo(arg: Record<string, any>) {}",
			optionsJson: "\"index-signature\"",
			wantFixed:   "function foo(arg: { [key: string]: any }) {}\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferIndexSignature", wantSpan: "Record<string, any>", wantMessage: "An index signature is preferred over a record.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "function foo(): Record<string, any> {}",
			optionsJson: "\"index-signature\"",
			wantFixed:   "function foo(): { [key: string]: any } {}\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferIndexSignature", wantSpan: "Record<string, any>", wantMessage: "An index signature is preferred over a record.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "type T = { readonly [key in string]: number };",
			optionsJson: "",
			wantFixed:   "type T = Readonly<Record<string, number>>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ readonly [key in string]: number }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "type T = { +readonly [key in string]: number };",
			optionsJson: "",
			wantFixed:   "type T = Readonly<Record<string, number>>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ +readonly [key in string]: number }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "type T = { -readonly [key in string]: number };",
			optionsJson: "",
			wantFixed:   "",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ -readonly [key in string]: number }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "type T = { [key in string]: number };",
			optionsJson: "",
			wantFixed:   "type T = Record<string, number>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ [key in string]: number }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\nfunction foo(e: { [key in PropertyKey]?: string }) {}\n      ",
			optionsJson: "",
			wantFixed:   "function foo(e: Partial<Record<PropertyKey, string>>) {}\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ [key in PropertyKey]?: string }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\nfunction foo(e: { [key in PropertyKey]+?: string }) {}\n      ",
			optionsJson: "",
			wantFixed:   "function foo(e: Partial<Record<PropertyKey, string>>) {}\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ [key in PropertyKey]+?: string }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\nfunction foo(e: { [key in PropertyKey]-?: string }) {}\n      ",
			optionsJson: "",
			wantFixed:   "function foo(e: Required<Record<PropertyKey, string>>) {}\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ [key in PropertyKey]-?: string }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\nfunction foo(e: { readonly [key in PropertyKey]-?: string }) {}\n      ",
			optionsJson: "",
			wantFixed:   "function foo(e: Readonly<Required<Record<PropertyKey, string>>>) {}\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ readonly [key in PropertyKey]-?: string }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ntype Options = [\n  { [Type in (typeof optionTesters)[number]['option']]?: boolean } & {\n    allow?: TypeOrValueSpecifier[];\n  },\n];\n      ",
			optionsJson: "",
			wantFixed:   "type Options = [\n  Partial<Record<(typeof optionTesters)[number]['option'], boolean>> & {\n    allow?: TypeOrValueSpecifier[];\n  },\n];\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ [Type in (typeof optionTesters)[number]['option']]?: boolean }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\nexport type MakeRequired<Base, Key extends keyof Base> = {\n  [K in Key]-?: NonNullable<Base[Key]>;\n} & Omit<Base, Key>;\n      ",
			optionsJson: "",
			wantFixed:   "export type MakeRequired<Base, Key extends keyof Base> = Required<Record<Key, NonNullable<Base[Key]>>> & Omit<Base, Key>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{\n  [K in Key]-?: NonNullable<Base[Key]>;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\nfunction f(): {\n  [k in (keyof ParseResult)]: unknown;\n} {\n  return {};\n}\n      ",
			optionsJson: "",
			wantFixed:   "function f(): Record<keyof ParseResult, unknown> {\n  return {};\n}\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{\n  [k in (keyof ParseResult)]: unknown;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ninterface Foo {\n  [key: string]: Bar;\n}\n\ninterface Bar {\n  [key: string];\n}\n      ",
			optionsJson: "",
			wantFixed:   "type Foo = Record<string, Bar>;\n\ninterface Bar {\n  [key: string];\n}\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface Foo {\n  [key: string]: Bar;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "\ntype Foo = {\n  [k in string];\n};\n      ",
			optionsJson: "",
			wantFixed:   "type Foo = Record<string, any>;\n",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{\n  [k in string];\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: ""},
			},
		},
		{
			sourceText:  "type Foo = { [key: string]: /* preserve me */ number };",
			optionsJson: "",
			wantFixed:   "",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{ [key: string]: /* preserve me */ number }", wantMessage: "A record is preferred over an index signature.", wantSuggestion: "type Foo = Record<string, number>;\n"},
			},
		},
		{
			sourceText:  "\ninterface Foo {\n  // preserve me\n  [key: string]: number;\n}\n      ",
			optionsJson: "",
			wantFixed:   "",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "interface Foo {\n  // preserve me\n  [key: string]: number;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: "type Foo = Record<string, number>;\n"},
			},
		},
		{
			sourceText:  "\ntype Foo = {\n  // preserve me\n  [Key in 'a' | 'b']: number;\n};\n      ",
			optionsJson: "",
			wantFixed:   "",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{\n  // preserve me\n  [Key in 'a' | 'b']: number;\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: "type Foo = Record<'a' | 'b', number>;\n"},
			},
		},
		{
			sourceText:  "\ntype Test = {\n  // 1\n  [Key in // 2\n    | 'a' // 3\n    | 'b' // 4\n    | 'c' // 5\n    | 'd' // 6\n  ]: string; // 7\n  // 8\n};\n      ",
			optionsJson: "",
			wantFixed:   "",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferRecord", wantSpan: "{\n  // 1\n  [Key in // 2\n    | 'a' // 3\n    | 'b' // 4\n    | 'c' // 5\n    | 'd' // 6\n  ]: string; // 7\n  // 8\n}", wantMessage: "A record is preferred over an index signature.", wantSuggestion: "type Test = Record<| 'a' // 3\n    | 'b' // 4\n    | 'c' // 5\n    | 'd', string>;\n"},
			},
		},
		{
			sourceText:  "type Foo = Record<string, /* preserve me */ number>;",
			optionsJson: "\"index-signature\"",
			wantFixed:   "",
			wantFindings: []consistentIndexedObjectStyleFinding{
				{wantId: "preferIndexSignature", wantSpan: "Record<string, /* preserve me */ number>", wantMessage: "An index signature is preferred over a record.", wantSuggestion: "type Foo = { [key: string]: number };\n"},
			},
		},
	}
	for index, testCase := range cases {
		t.Run(consistentIndexedObjectStyleCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, ConsistentIndexedObjectStyle,
				consistentIndexedObjectStyleFile, testCase.sourceText,
				consistentIndexedObjectStyleOptionsFor(t, testCase.optionsJson))

			wantIds := make([]string, len(testCase.wantFindings))
			for position, want := range testCase.wantFindings {
				wantIds[position] = want.wantId
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			for position, want := range testCase.wantFindings {
				diagnostic := result.Diagnostics[position]

				gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != want.wantSpan {
					t.Fatalf("finding %d span: expected %q, got %q", position, want.wantSpan, gotSpan)
				}
				if diagnostic.Message.Description != want.wantMessage {
					t.Fatalf("finding %d message: expected %q, got %q", position, want.wantMessage, diagnostic.Message.Description)
				}

				if want.wantSuggestion == "" && testCase.wantFixed == "" {
					// The row records a DECLINE: upstream reports it and deliberately offers no
					// repair, because there is no builtin Mutable and a minus-readonly mapped type
					// has no Record spelling. Without this assertion the decline is unchecked, and a
					// mutant that emitted a fix anyway survived the whole corpus: every other layer
					// is satisfied by a correct finding carrying an impossible rewrite.
					if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
						t.Fatalf("finding %d: expected no repair, got %d fixes and %d suggestions",
							position, len(diagnostic.Fixes), len(diagnostic.Suggestions))
					}
				}

				if want.wantSuggestion != "" {
					// A comment inside the replaced span downgrades the repair from a fix to a
					// suggestion, because the rewrite would drop it. Asserting the kind matters as
					// much as asserting the text.
					if len(diagnostic.Fixes) != 0 {
						t.Fatalf("finding %d: expected a suggestion, got %d fixes", position, len(diagnostic.Fixes))
					}
					if len(diagnostic.Suggestions) != 1 {
						t.Fatalf("finding %d: expected one suggestion, got %d", position, len(diagnostic.Suggestions))
					}
					applied := applyConsistentIndexedObjectStyleSuggestion(t, onDisk, diagnostic.Suggestions[0])
					if applied != want.wantSuggestion {
						t.Fatalf("finding %d suggestion applied: expected %q, got %q", position, want.wantSuggestion, applied)
					}
				}
			}

			if testCase.wantFixed != "" {
				rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
			}
		})
	}
}

// TestConsistentIndexedObjectStyleCarriesTheInterfaceModifiers pins what the repair must not delete.
//
// This shipped broken. The interface repair replaced the whole declaration starting at its first
// MODIFIER while building the alias from `type <name> = `, so `export interface X { [k: string]: V }`
// became `type X = Record<string, V>;` and six imports across the structure library stopped
// resolving. Upstream never had this defect only because estree hangs `export` on a wrapper node
// outside the replaced range; our declaration owns its modifiers.
//
// `declare` is carried too, which is a measured divergence: upstream's installed build rewrites
// `declare interface X` to `type X`, dropping it. That is harmless for a type alias, and carrying it
// is equally valid source, so the rewrite keeps every byte it did not need to change.
//
// The leading JSDoc and a comment between a modifier and the keyword sit outside the replaced
// range, so they survive and do not downgrade the fix to a suggestion.
func TestConsistentIndexedObjectStyleCarriesTheInterfaceModifiers(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSpan   string
		wantFixed  string
	}{
		{
			sourceText: "export interface Foo {\n  [key: string]: number;\n}\n",
			wantSpan:   "interface Foo {\n  [key: string]: number;\n}",
			wantFixed:  "export type Foo = Record<string, number>;\n",
		},
		{
			sourceText: "declare interface Foo {\n  [key: string]: number;\n}\n",
			wantSpan:   "interface Foo {\n  [key: string]: number;\n}",
			wantFixed:  "declare type Foo = Record<string, number>;\n",
		},
		{
			sourceText: "export declare interface Foo<T> {\n  [key: string]: T;\n}\n",
			wantSpan:   "interface Foo<T> {\n  [key: string]: T;\n}",
			wantFixed:  "export declare type Foo<T> = Record<string, T>;\n",
		},
		{
			sourceText: "export interface Foo<A = any, B extends string = string> {\n  readonly [key: string]: A | B;\n}\n",
			wantSpan:   "interface Foo<A = any, B extends string = string> {\n  readonly [key: string]: A | B;\n}",
			wantFixed:  "export type Foo<A = any, B extends string = string> = Readonly<Record<string, A | B>>;\n",
		},
		{
			sourceText: "/** The widths. */\nexport interface Foo {\n  [key: string]: number;\n}\n",
			wantSpan:   "interface Foo {\n  [key: string]: number;\n}",
			wantFixed:  "/** The widths. */\nexport type Foo = Record<string, number>;\n",
		},
		{
			sourceText: "export /* kept */ interface Foo {\n  [key: string]: number;\n}\n",
			wantSpan:   "interface Foo {\n  [key: string]: number;\n}",
			wantFixed:  "export /* kept */ type Foo = Record<string, number>;\n",
		},
		{
			sourceText: "namespace Outer {\n  export interface Foo {\n    [key: string]: number;\n  }\n}\n",
			wantSpan:   "interface Foo {\n    [key: string]: number;\n  }",
			wantFixed:  "namespace Outer {\n  export type Foo = Record<string, number>;\n}\n",
		},
	}

	for index, testCase := range cases {
		t.Run(consistentIndexedObjectStyleCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTypedWithOptions(t, ConsistentIndexedObjectStyle,
				consistentIndexedObjectStyleFile, testCase.sourceText,
				DefaultConsistentIndexedObjectStyleSettings())
			rule_testing.ExpectFindings(t, result, "preferRecord")

			diagnostic := result.Diagnostics[0]
			gotSpan := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Fatalf("span: expected %q, got %q", testCase.wantSpan, gotSpan)
			}
			if len(diagnostic.Suggestions) != 0 {
				t.Fatalf("expected an unattended fix, got %d suggestions", len(diagnostic.Suggestions))
			}
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

// TestConsistentIndexedObjectStyleNeedsTheTypedHarness pins the checker declaration.
//
// The circularity test resolves an identifier to its declaration, which is a checker call. Under
// the plain harness that answers nothing and every self-referencing case would report, so the
// failure direction here is over-reporting rather than silence.
func TestConsistentIndexedObjectStyleNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !ConsistentIndexedObjectStyle.NeedsTypeChecker {
		t.Fatal("the circularity test resolves identifiers, so the rule must declare NeedsTypeChecker")
	}
}

// TestDecodeConsistentIndexedObjectStyleOptions pins the decoder.
//
// This rule's option is a bare enum STRING rather than an object, so `rule.DecodeOptionsInto` has
// no struct to decode into and the decoder is hand-written for that reason rather than for the
// default-inversion reason its siblings in this package have.
//
// Absent and null are a bare severity and get the default. Every other value is refused, below.
func TestDecodeConsistentIndexedObjectStyleOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
		want string
	}{
		{name: "absent", raw: ``, want: "record"},
		{name: "null", raw: `null`, want: "record"},
		{name: "record", raw: `"record"`, want: "record"},
		{name: "indexSignature", raw: `"index-signature"`, want: "index-signature"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeConsistentIndexedObjectStyleOptions(json.RawMessage(testCase.raw))
			if err != nil {
				t.Fatalf("decoding %q: %v", testCase.raw, err)
			}
			options, isOptions := decoded.(ConsistentIndexedObjectStyleOptions)
			if !isOptions {
				t.Fatalf("the decoder returned %T rather than the options struct", decoded)
			}
			if options.Mode != testCase.want {
				t.Fatalf("mode: expected %q, got %q", testCase.want, options.Mode)
			}
		})
	}
}

// TestConsistentIndexedObjectStyleFallsBackToTheDefaultOnNilOptions pins the nil-options path.
//
// A rule configured as a bare "error" is handed nil options, and every fixture above reaches the
// rule through the decoder, so nothing there can see the fallback. The separating input is a Record
// type: silent under the default record mode, reporting under index-signature mode.
func TestConsistentIndexedObjectStyleFallsBackToTheDefaultOnNilOptions(t *testing.T) {
	t.Parallel()

	const record = "type T = Record<string, number>;"

	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, ConsistentIndexedObjectStyle,
		consistentIndexedObjectStyleFile, record, nil))

	// The control: the same input under the other mode must report, so the silence above is the
	// default being applied rather than the rule being inert.
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, ConsistentIndexedObjectStyle,
		consistentIndexedObjectStyleFile, record,
		ConsistentIndexedObjectStyleOptions{Mode: "index-signature"}), "preferIndexSignature")
}

// TestDecodeConsistentIndexedObjectStyleOptionsRefusesAValueUpstreamRefuses pins the other half
// (#p9s1131).
//
// Each of these used to decode to the default record mode, so a config naming one loaded clean and
// ran a mode nobody wrote. Upstream's schema is a string enum and refuses every one at load. The
// error names the value; the config layer prefixes the rule's name.
func TestDecodeConsistentIndexedObjectStyleOptionsRefusesAValueUpstreamRefuses(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		raw, named string
	}{
		{raw: `"index-signatures"`, named: "index-signatures"},
		{raw: `"Record"`, named: "Record"},
		{raw: `{"anything": 1}`, named: "anything"},
		{raw: `1`, named: "1"},
		{raw: `true`, named: "true"},
		{raw: `["record"]`, named: "record"},
	} {
		t.Run(testCase.raw, func(t *testing.T) {
			t.Parallel()
			_, err := DecodeConsistentIndexedObjectStyleOptions(json.RawMessage(testCase.raw))
			if err == nil {
				t.Fatalf("%s decoded, so the configured mode is silently not the one written", testCase.raw)
			}
			if !strings.Contains(err.Error(), testCase.named) {
				t.Fatalf("%s was refused without naming %q: %v", testCase.raw, testCase.named, err)
			}
		})
	}
}
