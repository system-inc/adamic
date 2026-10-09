package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const noRestrictedTypesFile = "/repository/source/Restricted.ts"

func noRestrictedTypesCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// noRestrictedTypesDecoded routes a fixture's options through the rule's own exported decoder
// rather than building the options struct directly.
//
// That is the whole point of doing it this way: the decoder is where the three wire shapes collapse
// into one struct and where `false` and `null` become "not banned", and a fixture handed a
// hand-built struct would leave every line of it untested. The options text below is the bare
// object rather than upstream's one-element array, because cohere's config layer unwraps the
// severity tuple before a decoder ever sees it.
func noRestrictedTypesDecoded(t *testing.T, optionsJson string) any {
	t.Helper()
	decoded, err := DecodeNoRestrictedTypesOptions([]byte(optionsJson))
	if err != nil {
		t.Fatalf("decoding %s: %v", optionsJson, err)
	}
	return decoded
}

// TestNoRestrictedTypesStaysSilent is upstream's eight passing cases verbatim, with their options.
//
// Each is a different reason to decline: nothing configured at all, a type whose name does not match
// the configured key, a non-empty literal where only the empty one is banned, and an entry written
// as `false` which un-bans rather than bans.
func TestNoRestrictedTypesStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		optionsJson string
	}{
		{
			sourceText:  "let f = Object();\n",
			optionsJson: "{}",
		},
		{
			sourceText:  "let f: { x: number; y: number } = { x: 1, y: 1 };\n",
			optionsJson: "{}",
		},
		{
			sourceText:  "let f = Object();\n",
			optionsJson: "{\"types\": {\"Object\": true}}",
		},
		{
			sourceText:  "let f = Object(false);\n",
			optionsJson: "{\"types\": {\"Object\": true}}",
		},
		{
			sourceText:  "let g = Object.create(null);\n",
			optionsJson: "{\"types\": {\"Object\": true}}",
		},
		{
			sourceText:  "let e: namespace.Object;\n",
			optionsJson: "{\"types\": {\"Object\": true}}",
		},
		{
			sourceText:  "let value: _.NS.Banned;\n",
			optionsJson: "{\"types\": {\"NS.Banned\": true}}",
		},
		{
			sourceText:  "let value: NS.Banned._;\n",
			optionsJson: "{\"types\": {\"NS.Banned\": true}}",
		},
		{
			// Measured, not imported. A NON-empty tuple and a non-empty type literal, which the
			// corpus never writes: every one of its tuple and literal cases is the empty spelling,
			// so a mutant dropping the size test survived all forty-two. Both are clean upstream.
			sourceText:  "let value: [string];\n",
			optionsJson: `{"types":{"[]":"no"}}`,
		},
		{
			sourceText:  "let value: { a: 1 };\n",
			optionsJson: `{"types":{"{}":"no"}}`,
		},
		{
			// The stronger half of the same pair: even NAMING the non-empty spelling does not
			// report, because the size test gates before the lookup happens at all. Measured
			// clean upstream, which is what makes the size test a real judgment rather than an
			// optimization over the lookup.
			sourceText:  "let value: [string];\n",
			optionsJson: `{"types":{"[string]":"no"}}`,
		},
		{
			sourceText:  "let value: { a: 1 };\n",
			optionsJson: `{"types":{"{a:1}":"no"}}`,
		},
		{
			// A banned entry written as `false`, which upstream's rule reads as not banned. Its
			// schema rejects the spelling before the rule sees it, so this is unreachable through
			// upstream's own configuration; cohere has no such schema layer, so a config here can
			// contain one and the branch decides what happens.
			sourceText:  "let value: Foo;\n",
			optionsJson: `{"types":{"Foo":false}}`,
		},
		{
			// The same for `null`, upstream's other not-banned spelling.
			sourceText:  "let value: Foo;\n",
			optionsJson: `{"types":{"Foo":null}}`,
		},
		{
			// A keyword that is not the banned one. This pins the keyword listener table itself:
			// with every listener registered unconditionally the lookup still declines, so the
			// registration test is a cost decision rather than a correctness one, and this row is
			// what shows the lookup half is load-bearing.
			sourceText:  "let value: string;\n",
			optionsJson: `{"types":{"number":"no"}}`,
		},
	}
	for index, testCase := range cases {
		t.Run(noRestrictedTypesCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoRestrictedTypes,
				noRestrictedTypesFile, testCase.sourceText,
				noRestrictedTypesDecoded(t, testCase.optionsJson)))
		})
	}
}

// noRestrictedTypesFinding is one expected finding with every layer it can be wrong at.
//
// There is one message id, so asserting it proves almost nothing. What varies is the rendered text,
// which carries both the stripped type name and whatever the configuration appended, and the span,
// which is the RAW source text rather than the stripped name. A generic reference reports twice,
// once on the name and once on the whole thing, and only the span separates those two.
type noRestrictedTypesFinding struct {
	wantSpan    string
	wantMessage string

	// wantNoFix marks a finding reported without a repair, which is every entry that configures no
	// `fixWith`.
	wantNoFix bool
}

// TestNoRestrictedTypesFires is upstream's thirty-four reporting cases verbatim, with their options,
// their repaired sources, and every finding's rendered text and span taken from the installed 8.67.0
// build.
func TestNoRestrictedTypesFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		optionsJson  string
		wantFindings []noRestrictedTypesFinding
		wantOutput   string
	}{
		{
			sourceText:  "let value: bigint;\n",
			optionsJson: "{\"types\": {\"bigint\": \"Use Ok instead.\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "bigint",
					wantMessage: "Don't use `bigint` as a type. Use Ok instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: bigint;\n",
		},
		{
			sourceText:  "let value: boolean;\n",
			optionsJson: "{\"types\": {\"boolean\": \"Use Ok instead.\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "boolean",
					wantMessage: "Don't use `boolean` as a type. Use Ok instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: boolean;\n",
		},
		{
			sourceText:  "let value: never;\n",
			optionsJson: "{\"types\": {\"never\": \"Use Ok instead.\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "never",
					wantMessage: "Don't use `never` as a type. Use Ok instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: never;\n",
		},
		{
			sourceText:  "let value: null;\n",
			optionsJson: "{\"types\": {\"null\": \"Use Ok instead.\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "null",
					wantMessage: "Don't use `null` as a type. Use Ok instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: null;\n",
		},
		{
			sourceText:  "let value: number;\n",
			optionsJson: "{\"types\": {\"number\": \"Use Ok instead.\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "number",
					wantMessage: "Don't use `number` as a type. Use Ok instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: number;\n",
		},
		{
			sourceText:  "let value: object;\n",
			optionsJson: "{\"types\": {\"object\": \"Use Ok instead.\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "object",
					wantMessage: "Don't use `object` as a type. Use Ok instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: object;\n",
		},
		{
			sourceText:  "let value: string;\n",
			optionsJson: "{\"types\": {\"string\": \"Use Ok instead.\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "string",
					wantMessage: "Don't use `string` as a type. Use Ok instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: string;\n",
		},
		{
			sourceText:  "let value: symbol;\n",
			optionsJson: "{\"types\": {\"symbol\": \"Use Ok instead.\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "symbol",
					wantMessage: "Don't use `symbol` as a type. Use Ok instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: symbol;\n",
		},
		{
			sourceText:  "let value: undefined;\n",
			optionsJson: "{\"types\": {\"undefined\": \"Use Ok instead.\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "undefined",
					wantMessage: "Don't use `undefined` as a type. Use Ok instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: undefined;\n",
		},
		{
			sourceText:  "let value: unknown;\n",
			optionsJson: "{\"types\": {\"unknown\": \"Use Ok instead.\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "unknown",
					wantMessage: "Don't use `unknown` as a type. Use Ok instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: unknown;\n",
		},
		{
			sourceText:  "let value: void;\n",
			optionsJson: "{\"types\": {\"void\": \"Use Ok instead.\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "void",
					wantMessage: "Don't use `void` as a type. Use Ok instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: void;\n",
		},
		{
			sourceText:  "let value: [];\n",
			optionsJson: "{\"types\": {\"[]\": \"Use unknown[] instead.\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "[]",
					wantMessage: "Don't use `[]` as a type. Use unknown[] instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: [];\n",
		},
		{
			sourceText:  "let value: [  ];\n",
			optionsJson: "{\"types\": {\"[]\": \"Use unknown[] instead.\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "[  ]",
					wantMessage: "Don't use `[]` as a type. Use unknown[] instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: [  ];\n",
		},
		{
			sourceText:  "let value: [[]];\n",
			optionsJson: "{\"types\": {\"[]\": \"Use unknown[] instead.\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "[]",
					wantMessage: "Don't use `[]` as a type. Use unknown[] instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: [[]];\n",
		},
		{
			sourceText:  "let value: Banned;\n",
			optionsJson: "{\"types\": {\"Banned\": true}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "Banned",
					wantMessage: "Don't use `Banned` as a type.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: Banned;\n",
		},
		{
			sourceText:  "let value: Banned;\n",
			optionsJson: "{\"types\": {\"Banned\": \"Use '{}' instead.\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "Banned",
					wantMessage: "Don't use `Banned` as a type. Use '{}' instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: Banned;\n",
		},
		{
			sourceText:  "let value: Banned[];\n",
			optionsJson: "{\"types\": {\"Banned\": \"Use '{}' instead.\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "Banned",
					wantMessage: "Don't use `Banned` as a type. Use '{}' instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: Banned[];\n",
		},
		{
			sourceText:  "let value: [Banned];\n",
			optionsJson: "{\"types\": {\"Banned\": \"Use '{}' instead.\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "Banned",
					wantMessage: "Don't use `Banned` as a type. Use '{}' instead.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: [Banned];\n",
		},
		{
			sourceText:  "let value: Banned;\n",
			optionsJson: "{\"types\": {\"Banned\": \"\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "Banned",
					wantMessage: "Don't use `Banned` as a type.",
					wantNoFix:   true,
				},
			},
			wantOutput: "let value: Banned;\n",
		},
		{
			sourceText:  "let b: { c: Banned };\n",
			optionsJson: "{\"types\": {\"Banned\": {\"fixWith\": \"Ok\", \"message\": \"Use Ok instead.\"}}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "Banned",
					wantMessage: "Don't use `Banned` as a type. Use Ok instead.",
				},
			},
			wantOutput: "let b: { c: Ok };\n",
		},
		{
			sourceText:  "1 as Banned;\n",
			optionsJson: "{\"types\": {\"Banned\": {\"fixWith\": \"Ok\", \"message\": \"Use Ok instead.\"}}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "Banned",
					wantMessage: "Don't use `Banned` as a type. Use Ok instead.",
				},
			},
			wantOutput: "1 as Ok;\n",
		},
		{
			sourceText:  "class Derived implements Banned {}\n",
			optionsJson: "{\"types\": {\"Banned\": {\"fixWith\": \"Ok\", \"message\": \"Use Ok instead.\"}}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "Banned",
					wantMessage: "Don't use `Banned` as a type. Use Ok instead.",
				},
			},
			wantOutput: "class Derived implements Ok {}\n",
		},
		{
			sourceText:  "class Derived implements Banned1, Banned2 {}\n",
			optionsJson: "{\"types\": {\"Banned1\": {\"fixWith\": \"Ok1\", \"message\": \"Use Ok1 instead.\"}, \"Banned2\": {\"fixWith\": \"Ok2\", \"message\": \"Use Ok2 instead.\"}}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "Banned1",
					wantMessage: "Don't use `Banned1` as a type. Use Ok1 instead.",
				},
				{
					wantSpan:    "Banned2",
					wantMessage: "Don't use `Banned2` as a type. Use Ok2 instead.",
				},
			},
			wantOutput: "class Derived implements Ok1, Ok2 {}\n",
		},
		{
			sourceText:  "class Derived implements Omit<Foo, 'a'> {}\n",
			optionsJson: "{\"types\": {\"Omit\": {\"fixWith\": \"Ok\", \"message\": \"Use Ok instead.\"}}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "Omit",
					wantMessage: "Don't use `Omit` as a type. Use Ok instead.",
				},
			},
			wantOutput: "class Derived implements Ok<Foo, 'a'> {}\n",
		},
		{
			sourceText:  "interface Derived extends Banned {}\n",
			optionsJson: "{\"types\": {\"Banned\": {\"fixWith\": \"Ok\", \"message\": \"Use Ok instead.\"}}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "Banned",
					wantMessage: "Don't use `Banned` as a type. Use Ok instead.",
				},
			},
			wantOutput: "interface Derived extends Ok {}\n",
		},
		{
			sourceText:  "interface Derived extends Omit<Foo, 'a'> {}\n",
			optionsJson: "{\"types\": {\"Omit\": {\"fixWith\": \"Ok\", \"message\": \"Use Ok instead.\"}}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "Omit",
					wantMessage: "Don't use `Omit` as a type. Use Ok instead.",
				},
			},
			wantOutput: "interface Derived extends Ok<Foo, 'a'> {}\n",
		},
		{
			sourceText:  "type Intersection = Banned & {};\n",
			optionsJson: "{\"types\": {\"Banned\": {\"fixWith\": \"Ok\", \"message\": \"Use Ok instead.\"}}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "Banned",
					wantMessage: "Don't use `Banned` as a type. Use Ok instead.",
				},
			},
			wantOutput: "type Intersection = Ok & {};\n",
		},
		{
			sourceText:  "type Union = Banned | {};\n",
			optionsJson: "{\"types\": {\"Banned\": {\"fixWith\": \"Ok\", \"message\": \"Use Ok instead.\"}}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "Banned",
					wantMessage: "Don't use `Banned` as a type. Use Ok instead.",
				},
			},
			wantOutput: "type Union = Ok | {};\n",
		},
		{
			sourceText:  "let value: NS.Banned;\n",
			optionsJson: "{\"types\": {\"NS.Banned\": {\"fixWith\": \"NS.Ok\", \"message\": \"Use NS.Ok instead.\"}}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "NS.Banned",
					wantMessage: "Don't use `NS.Banned` as a type. Use NS.Ok instead.",
				},
			},
			wantOutput: "let value: NS.Ok;\n",
		},
		{
			sourceText:  "let value: {} = {};\n",
			optionsJson: "{\"types\": {\"{}\": {\"fixWith\": \"object\", \"message\": \"Use object instead.\"}}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "{}",
					wantMessage: "Don't use `{}` as a type. Use object instead.",
				},
			},
			wantOutput: "let value: object = {};\n",
		},
		{
			sourceText:  "let value: NS.Banned;\n",
			optionsJson: "{\"types\": {\"  NS.Banned  \": {\"fixWith\": \"NS.Ok\", \"message\": \"Use NS.Ok instead.\"}}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "NS.Banned",
					wantMessage: "Don't use `NS.Banned` as a type. Use NS.Ok instead.",
				},
			},
			wantOutput: "let value: NS.Ok;\n",
		},
		{
			sourceText:  "let value: Type<   Banned   >;\n",
			optionsJson: "{\"types\": {\"       Banned      \": {\"fixWith\": \"Ok\", \"message\": \"Use Ok instead.\"}}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "Banned",
					wantMessage: "Don't use `Banned` as a type. Use Ok instead.",
				},
			},
			wantOutput: "let value: Type<   Ok   >;\n",
		},
		{
			sourceText:  "type Intersection = Banned<any>;\n",
			optionsJson: "{\"types\": {\"Banned<any>\": \"Don't use `any` as a type parameter to `Banned`\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "Banned<any>",
					wantMessage: "Don't use `Banned<any>` as a type. Don't use `any` as a type parameter to `Banned`",
					wantNoFix:   true,
				},
			},
			wantOutput: "type Intersection = Banned<any>;\n",
		},
		{
			sourceText:  "type Intersection = Banned<A,B>;\n",
			optionsJson: "{\"types\": {\"Banned<A, B>\": \"Don't pass `A, B` as parameters to `Banned`\"}}",
			wantFindings: []noRestrictedTypesFinding{
				{
					wantSpan:    "Banned<A,B>",
					wantMessage: "Don't use `Banned<A,B>` as a type. Don't pass `A, B` as parameters to `Banned`",
					wantNoFix:   true,
				},
			},
			wantOutput: "type Intersection = Banned<A,B>;\n",
		},
	}
	for index, testCase := range cases {
		t.Run(noRestrictedTypesCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoRestrictedTypes, noRestrictedTypesFile,
				testCase.sourceText, noRestrictedTypesDecoded(t, testCase.optionsJson))

			wantIds := make([]string, len(testCase.wantFindings))
			for position := range testCase.wantFindings {
				wantIds[position] = "bannedTypeMessage"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			anyFixExpected := false
			for position, want := range testCase.wantFindings {
				diagnostic := result.Diagnostics[position]

				gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != want.wantSpan {
					t.Fatalf("finding %d span: expected %q, got %q", position, want.wantSpan, gotSpan)
				}
				if diagnostic.Message.Description != want.wantMessage {
					t.Fatalf("finding %d message: expected %q, got %q", position, want.wantMessage,
						diagnostic.Message.Description)
				}

				wantFixes := 1
				if want.wantNoFix {
					wantFixes = 0
				} else {
					anyFixExpected = true
				}
				if len(diagnostic.Fixes) != wantFixes {
					t.Fatalf("finding %d fixes: expected %d, got %d", position, wantFixes,
						len(diagnostic.Fixes))
				}
			}

			if anyFixExpected {
				rule_testing.ExpectFixedSource(t, result, strings.TrimSpace(testCase.wantOutput)+"\n")
			} else if strings.TrimSpace(testCase.wantOutput)+"\n" != onDisk {
				t.Fatalf("a case expecting no repair must expect the source unchanged")
			}
		})
	}
}

// TestNoRestrictedTypesOffersSuggestions covers the half of the option surface upstream's corpus
// does not touch at all.
//
// The corpus asserts thirteen fixer outputs and ZERO suggestions, so every one of these rows is
// measured against the installed 8.67.0 build rather than imported. Two of them caught real defects
// in the first version of this rule, both invisible to the imported cases: a fix and suggestions are
// not exclusive and returning after the fix dropped the suggestions, and an empty `fixWith` is
// FALSY upstream so it produces no repair rather than one deleting the type.
func TestNoRestrictedTypesOffersSuggestions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		optionsJson     string
		wantMessage     string
		wantFixCount    int
		wantSuggestions []string
		wantOutput      string
	}{
		{
			name:         "suggest alone offers each replacement and proposes no fix",
			optionsJson:  `{"types":{"Foo":{"suggest":["Bar","Baz"]}}}`,
			wantMessage:  "Don't use `Foo` as a type.",
			wantFixCount: 0,
			wantSuggestions: []string{
				"Replace `Foo` with `Bar`.",
				"Replace `Foo` with `Baz`.",
			},
			wantOutput: "let value: Foo;\n",
		},
		{
			name:         "a custom message and a suggestion coexist",
			optionsJson:  `{"types":{"Foo":{"message":"no Foo","suggest":["Bar"]}}}`,
			wantMessage:  "Don't use `Foo` as a type. no Foo",
			wantFixCount: 0,
			wantSuggestions: []string{
				"Replace `Foo` with `Bar`.",
			},
			wantOutput: "let value: Foo;\n",
		},
		{
			// The row that caught the first defect. Upstream emits BOTH, so a reader who declines
			// the automatic repair still gets the offer.
			name:         "fixWith and suggest are not exclusive",
			optionsJson:  `{"types":{"Foo":{"fixWith":"Bar","suggest":["Baz"]}}}`,
			wantMessage:  "Don't use `Foo` as a type.",
			wantFixCount: 1,
			wantSuggestions: []string{
				"Replace `Foo` with `Baz`.",
			},
			wantOutput: "let value: Bar;\n",
		},
		{
			// The row that caught the second. Upstream tests the value's truthiness, so the empty
			// string means no repair rather than "replace with nothing".
			name:            "an empty fixWith proposes no repair",
			optionsJson:     `{"types":{"Foo":{"fixWith":""}}}`,
			wantMessage:     "Don't use `Foo` as a type.",
			wantFixCount:    0,
			wantSuggestions: nil,
			wantOutput:      "let value: Foo;\n",
		},
		{
			name:            "an empty object bans with the standard message and offers nothing",
			optionsJson:     `{"types":{"Foo":{}}}`,
			wantMessage:     "Don't use `Foo` as a type.",
			wantFixCount:    0,
			wantSuggestions: nil,
			wantOutput:      "let value: Foo;\n",
		},
	}

	const sourceText = "let value: Foo;\n"
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoRestrictedTypes, noRestrictedTypesFile,
				sourceText, noRestrictedTypesDecoded(t, testCase.optionsJson))
			rule_testing.ExpectFindings(t, result, "bannedTypeMessage")

			diagnostic := result.Diagnostics[0]
			if diagnostic.Message.Description != testCase.wantMessage {
				t.Fatalf("message: expected %q, got %q", testCase.wantMessage,
					diagnostic.Message.Description)
			}
			if len(diagnostic.Fixes) != testCase.wantFixCount {
				t.Fatalf("fixes: expected %d, got %d", testCase.wantFixCount, len(diagnostic.Fixes))
			}
			if len(diagnostic.Suggestions) != len(testCase.wantSuggestions) {
				t.Fatalf("suggestions: expected %d, got %d", len(testCase.wantSuggestions),
					len(diagnostic.Suggestions))
			}
			for position, wantDescription := range testCase.wantSuggestions {
				suggestion := diagnostic.Suggestions[position]
				if suggestion.Message.Id != "bannedTypeReplacement" {
					t.Fatalf("suggestion %d id: expected %q, got %q", position,
						"bannedTypeReplacement", suggestion.Message.Id)
				}
				if suggestion.Message.Description != wantDescription {
					t.Fatalf("suggestion %d: expected %q, got %q", position, wantDescription,
						suggestion.Message.Description)
				}
				// The harness applies fixes but not suggestions, so the replacement is checked by
				// applying it here. Comparing the fix's TEXT alone would pass a suggestion writing
				// the right string over the wrong span.
				if len(suggestion.Fixes) != 1 {
					t.Fatalf("suggestion %d: expected one fix, got %d", position, len(suggestion.Fixes))
				}
				fix := suggestion.Fixes[0]
				rewritten := sourceText[:fix.Range.Pos()] + fix.Text + sourceText[fix.Range.End():]
				wantRewritten := "let value: " +
					strings.TrimSuffix(strings.TrimPrefix(wantDescription, "Replace `Foo` with `"),
						"`.") + ";\n"
				if rewritten != wantRewritten {
					t.Fatalf("suggestion %d applied: expected %q, got %q", position, wantRewritten,
						rewritten)
				}
			}

			if testCase.wantFixCount > 0 {
				rule_testing.ExpectFixedSource(t, result, testCase.wantOutput)
			}
		})
	}
}

// TestNoRestrictedTypesSurvivesNilOptions covers the configuration this rule actually ships under.
//
// The live config enables it as a bare "error", which means the decoder is handed NOTHING. The
// brief's warning about this failure is specific: a rule whose decoder errors on empty input gets
// nil back, `options.(T)` on nil yields the zero value, and the rule registers on every file while
// matching nothing. That reads exactly like a correctly-declining rule, and every fixture routed
// through a populated decoder stays green over it.
//
// So this asserts the nil path directly, in three places it can break: the decoder handed nil, the
// decoder handed an empty slice, and the rule handed a nil options value. The last row is the
// control, since a rule that is inert for the right reason and one that is inert for the wrong
// reason are indistinguishable without something that fires.
func TestNoRestrictedTypesSurvivesNilOptions(t *testing.T) {
	t.Parallel()

	for _, raw := range [][]byte{nil, []byte("")} {
		decoded, err := DecodeNoRestrictedTypesOptions(raw)
		if err != nil {
			t.Fatalf("decoding %q: %v", raw, err)
		}
		settings, ok := decoded.(NoRestrictedTypesOptions)
		if !ok {
			t.Fatalf("decoding %q produced %T rather than the options struct", raw, decoded)
		}
		if settings.Types == nil {
			t.Fatalf("decoding %q left the ban map nil, so every lookup would panic or miss", raw)
		}
		if len(settings.Types) != 0 {
			t.Fatalf("decoding %q banned %d types, expected none", raw, len(settings.Types))
		}
	}

	// A nil options value is what the rule is handed when the config names it with no options at
	// all. It must decline rather than panic.
	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoRestrictedTypes,
		noRestrictedTypesFile, "let value: string;\n", nil))

	// The control: the same source with `string` banned has to report, or the row above proves
	// nothing about why the rule was silent.
	rule_testing.ExpectFindings(t, rule_testing.RunWithOptions(t, NoRestrictedTypes,
		noRestrictedTypesFile, "let value: string;\n",
		noRestrictedTypesDecoded(t, `{"types":{"string":"no"}}`)), "bannedTypeMessage")
}
