package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const noRedundantTypeConstituentsFile = "/repository/source/Types.ts"

func noRedundantTypeConstituentsCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// TestNoRedundantTypeConstituentsStaysSilent is upstream's fifty-five passing cases verbatim.
//
// Each was re-measured against the installed 8.67.0 build with the file written exactly as this
// harness writes it, so these are evidence rather than a label copied out of an array named `valid`.
// All fifty-five are clean there.
//
// Several are the false positives a plausible port ships. `boolean | string` is clean because
// TypeScript stores `boolean` as `false | true` and splitting it would make the union look like it
// carries two boolean literals beside a string. `never` inside a return annotation is clean because
// it says something there. And `string | number` is clean in both directions: two primitives do not
// absorb each other at all.
func TestNoRedundantTypeConstituentsStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"type T = any;\ntype U = T;\n",
		"type T = never;\ntype U = T;\n",
		"type T = 1 | 2;\ntype U = T | 3;\ntype V = U;\n",
		"type T = () => never;\n",
		"type T = () => never | string;\n",
		"type B = never;\ntype T = () => B | string;\n",
		"type B = string;\ntype T = () => B | never;\n",
		"type T = () => string | never;\n",
		"type T = { (): string | never };\n",
		"function _(): string | never {\n  return '';\n}\n",
		"const _ = (): string | never => {\n  return '';\n};\n",
		"type B = string;\ntype T = { (): B | never };\n",
		"type T = { new (): string | never };\n",
		"type B = never;\ntype T = { new (): string | B };\n",
		"type B = unknown;\ntype T = B;\n",
		"type T = bigint;\n",
		"type B = bigint;\ntype T = B;\n",
		"type T = 1n | 2n;\n",
		"type B = 1n;\ntype T = B | 2n;\n",
		"type T = boolean;\n",
		"type B = boolean;\ntype T = B;\n",
		"type T = false | true;\n",
		"type B = false;\ntype T = B | true;\n",
		"type B = true;\ntype T = B | false;\n",
		"type T = number;\n",
		"type B = number;\ntype T = B;\n",
		"type T = 1 | 2;\n",
		"type B = 1;\ntype T = B | 2;\n",
		"type T = 1 | false;\n",
		"type B = 1;\ntype T = B | false;\n",
		"type T = string;\n",
		"type B = string;\ntype T = B;\n",
		"type T = 'a' | 'b';\n",
		"type B = 'b';\ntype T = 'a' | B;\n",
		"type B = 'a';\ntype T = B | 'b';\n",
		"type T = bigint | null;\n",
		"type B = bigint;\ntype T = B | null;\n",
		"type T = boolean | null;\n",
		"type B = boolean;\ntype T = B | null;\n",
		"type T = number | null;\n",
		"type B = number;\ntype T = B | null;\n",
		"type T = string | null;\n",
		"type B = string;\ntype T = B | null;\n",
		"type T = bigint & null;\n",
		"type B = bigint;\ntype T = B & null;\n",
		"type T = boolean & null;\n",
		"type B = boolean;\ntype T = B & null;\n",
		"type T = number & null;\n",
		"type B = number;\ntype T = B & null;\n",
		"type T = string & null;\n",
		"type B = string;\ntype T = B & null;\n",
		"type T = `${string}` & null;\n",
		"type B = `${string}`;\ntype T = B & null;\n",
		"type T = 'a' | 1 | 'b';\ntype U = T & string;\n",
		"declare function fn(): never | 'foo';\n",
	}
	for index, sourceText := range cases {
		t.Run(noRedundantTypeConstituentsCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoRedundantTypeConstituents,
				noRedundantTypeConstituentsFile, sourceText))
		})
	}
}

// noRedundantTypeConstituentsFinding is one expected finding with all three layers it can be wrong at.
//
// Five message ids describe five different judgments and every one interpolates. The id says which
// absorption fired; the rendered text names WHICH member was swallowed and by what, which is where a
// port that reversed the direction shows itself while satisfying every id; and the span says which
// member the reader is being sent to.
type noRedundantTypeConstituentsFinding struct {
	wantSpan    string
	wantId      string
	wantMessage string
}

// TestNoRedundantTypeConstituentsFires is upstream's forty-nine reporting cases verbatim, with every
// finding's id, rendered text and span taken from the installed 8.67.0 build rather than from the
// corpus's own `errors` array, which states ids but not the rendered text.
func TestNoRedundantTypeConstituentsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		wantFindings []noRedundantTypeConstituentsFinding
	}{
		{
			sourceText: "type T = number | any;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "any",
					wantId:      "overrides",
					wantMessage: "'any' overrides all other types in this union type.",
				},
			},
		},
		{
			sourceText: "type B = number;\ntype T = B | any;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "any",
					wantId:      "overrides",
					wantMessage: "'any' overrides all other types in this union type.",
				},
			},
		},
		{
			sourceText: "type T = any | number;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "any",
					wantId:      "overrides",
					wantMessage: "'any' overrides all other types in this union type.",
				},
			},
		},
		{
			sourceText: "type B = any;\ntype T = B | number;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "B",
					wantId:      "overrides",
					wantMessage: "'any' overrides all other types in this union type.",
				},
			},
		},
		{
			sourceText: "type T = number | never;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "never",
					wantId:      "overridden",
					wantMessage: "'never' is overridden by other types in this union type.",
				},
			},
		},
		{
			sourceText: "type B = number;\ntype T = B | never;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "never",
					wantId:      "overridden",
					wantMessage: "'never' is overridden by other types in this union type.",
				},
			},
		},
		{
			sourceText: "type B = never;\ntype T = B | number;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "B",
					wantId:      "overridden",
					wantMessage: "'never' is overridden by other types in this union type.",
				},
			},
		},
		{
			sourceText: "type T = never | number;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "never",
					wantId:      "overridden",
					wantMessage: "'never' is overridden by other types in this union type.",
				},
			},
		},
		{
			sourceText: "type T = number | unknown;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "unknown",
					wantId:      "overrides",
					wantMessage: "'unknown' overrides all other types in this union type.",
				},
			},
		},
		{
			sourceText: "type T = unknown | number;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "unknown",
					wantId:      "overrides",
					wantMessage: "'unknown' overrides all other types in this union type.",
				},
			},
		},
		{
			sourceText: "type ErrorTypes = NotKnown | 0;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "NotKnown",
					wantId:      "errorTypeOverrides",
					wantMessage: "'NotKnown' is an 'error' type that acts as 'any' and overrides all other types in this union type.",
				},
			},
		},
		{
			sourceText: "type T = number | 0;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "0",
					wantId:      "literalOverridden",
					wantMessage: "0 is overridden by number in this union type.",
				},
			},
		},
		{
			sourceText: "type T = number | (0 | 1);\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "0 | 1",
					wantId:      "literalOverridden",
					wantMessage: "0 | 1 is overridden by number in this union type.",
				},
			},
		},
		{
			sourceText: "type T = (0 | 0) | number;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "0 | 0",
					wantId:      "literalOverridden",
					wantMessage: "0 | 0 is overridden by number in this union type.",
				},
			},
		},
		{
			sourceText: "type B = 0 | 1;\ntype T = (2 | B) | number;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "2 | B",
					wantId:      "literalOverridden",
					wantMessage: "2 | 0 | 1 is overridden by number in this union type.",
				},
			},
		},
		{
			sourceText: "type T = (0 | (1 | 2)) | number;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "0 | (1 | 2)",
					wantId:      "literalOverridden",
					wantMessage: "0 | 1 | 2 is overridden by number in this union type.",
				},
			},
		},
		{
			sourceText: "type T = (0 | 1) | number;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "0 | 1",
					wantId:      "literalOverridden",
					wantMessage: "0 | 1 is overridden by number in this union type.",
				},
			},
		},
		{
			sourceText: "type T = (0 | (0 | 1)) | number;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "0 | (0 | 1)",
					wantId:      "literalOverridden",
					wantMessage: "0 | 0 | 1 is overridden by number in this union type.",
				},
			},
		},
		{
			sourceText: "type T = (2 | 'other' | 3) | number;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "2 | 'other' | 3",
					wantId:      "literalOverridden",
					wantMessage: "2 | 3 is overridden by number in this union type.",
				},
			},
		},
		{
			sourceText: "type T = '' | string;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "''",
					wantId:      "literalOverridden",
					wantMessage: "\"\" is overridden by string in this union type.",
				},
			},
		},
		{
			sourceText: "type B = 'b';\ntype T = B | string;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "B",
					wantId:      "literalOverridden",
					wantMessage: "\"b\" is overridden by string in this union type.",
				},
			},
		},
		{
			sourceText: "type T = `a${number}c` | string;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "`a${number}c`",
					wantId:      "literalOverridden",
					wantMessage: "template literal type is overridden by string in this union type.",
				},
			},
		},
		{
			sourceText: "type B = `a${number}c`;\ntype T = B | string;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "B",
					wantId:      "literalOverridden",
					wantMessage: "template literal type is overridden by string in this union type.",
				},
			},
		},
		{
			sourceText: "type T = `${number}` | string;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "`${number}`",
					wantId:      "literalOverridden",
					wantMessage: "template literal type is overridden by string in this union type.",
				},
			},
		},
		{
			sourceText: "type T = 0n | bigint;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "0n",
					wantId:      "literalOverridden",
					wantMessage: "0n is overridden by bigint in this union type.",
				},
			},
		},
		{
			sourceText: "type T = -1n | bigint;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "-1n",
					wantId:      "literalOverridden",
					wantMessage: "-1n is overridden by bigint in this union type.",
				},
			},
		},
		{
			sourceText: "type T = (-1n | 1n) | bigint;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "-1n | 1n",
					wantId:      "literalOverridden",
					wantMessage: "-1n | 1n is overridden by bigint in this union type.",
				},
			},
		},
		{
			sourceText: "type B = boolean;\ntype T = B | false;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "false",
					wantId:      "literalOverridden",
					wantMessage: "false is overridden by boolean in this union type.",
				},
			},
		},
		{
			sourceText: "type T = false | boolean;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "false",
					wantId:      "literalOverridden",
					wantMessage: "false is overridden by boolean in this union type.",
				},
			},
		},
		{
			sourceText: "type T = true | boolean;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "true",
					wantId:      "literalOverridden",
					wantMessage: "true is overridden by boolean in this union type.",
				},
			},
		},
		{
			sourceText: "type T = false & boolean;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "boolean",
					wantId:      "primitiveOverridden",
					wantMessage: "boolean is overridden by the false in this intersection type.",
				},
			},
		},
		{
			sourceText: "type B = false;\ntype T = B & boolean;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "boolean",
					wantId:      "primitiveOverridden",
					wantMessage: "boolean is overridden by the false in this intersection type.",
				},
			},
		},
		{
			sourceText: "type B = true;\ntype T = B & boolean;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "boolean",
					wantId:      "primitiveOverridden",
					wantMessage: "boolean is overridden by the true in this intersection type.",
				},
			},
		},
		{
			sourceText: "type T = true & boolean;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "boolean",
					wantId:      "primitiveOverridden",
					wantMessage: "boolean is overridden by the true in this intersection type.",
				},
			},
		},
		{
			sourceText: "type T = number & any;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "any",
					wantId:      "overrides",
					wantMessage: "'any' overrides all other types in this intersection type.",
				},
			},
		},
		{
			sourceText: "type T = any & number;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "any",
					wantId:      "overrides",
					wantMessage: "'any' overrides all other types in this intersection type.",
				},
			},
		},
		{
			sourceText: "type ErrorTypes = NotKnown & 0;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "NotKnown",
					wantId:      "errorTypeOverrides",
					wantMessage: "'NotKnown' is an 'error' type that acts as 'any' and overrides all other types in this intersection type.",
				},
			},
		},
		{
			sourceText: "type T = number & never;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "never",
					wantId:      "overrides",
					wantMessage: "'never' overrides all other types in this intersection type.",
				},
			},
		},
		{
			sourceText: "type B = never;\ntype T = B & number;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "B",
					wantId:      "overrides",
					wantMessage: "'never' overrides all other types in this intersection type.",
				},
			},
		},
		{
			sourceText: "type T = never & number;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "never",
					wantId:      "overrides",
					wantMessage: "'never' overrides all other types in this intersection type.",
				},
			},
		},
		{
			sourceText: "type T = number & unknown;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "unknown",
					wantId:      "overridden",
					wantMessage: "'unknown' is overridden by other types in this intersection type.",
				},
			},
		},
		{
			sourceText: "type T = unknown & number;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "unknown",
					wantId:      "overridden",
					wantMessage: "'unknown' is overridden by other types in this intersection type.",
				},
			},
		},
		{
			sourceText: "type T = number & 0;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "number",
					wantId:      "primitiveOverridden",
					wantMessage: "number is overridden by the 0 in this intersection type.",
				},
			},
		},
		{
			sourceText: "type T = '' & string;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "string",
					wantId:      "primitiveOverridden",
					wantMessage: "string is overridden by the \"\" in this intersection type.",
				},
			},
		},
		{
			sourceText: "type B = 0n;\ntype T = B & bigint;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "bigint",
					wantId:      "primitiveOverridden",
					wantMessage: "bigint is overridden by the 0n in this intersection type.",
				},
			},
		},
		{
			sourceText: "type T = 0n & bigint;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "bigint",
					wantId:      "primitiveOverridden",
					wantMessage: "bigint is overridden by the 0n in this intersection type.",
				},
			},
		},
		{
			sourceText: "type T = -1n & bigint;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "bigint",
					wantId:      "primitiveOverridden",
					wantMessage: "bigint is overridden by the -1n in this intersection type.",
				},
			},
		},
		{
			sourceText: "type T = 'a' | 'b';\ntype U = T & string;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "T",
					wantId:      "primitiveOverridden",
					wantMessage: "string is overridden by the \"a\" | \"b\" in this intersection type.",
				},
			},
		},
		{
			sourceText: "type S = 1 | 2;\ntype T = 'a' | 'b';\ntype U = S & T & string & number;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "S",
					wantId:      "primitiveOverridden",
					wantMessage: "number is overridden by the 1 | 2 in this intersection type.",
				},
				{
					wantSpan:    "T",
					wantId:      "primitiveOverridden",
					wantMessage: "string is overridden by the \"a\" | \"b\" in this intersection type.",
				},
			},
		},
		{
			// Measured, not imported. Two literals absorbed by one primitive, on DIFFERENT type
			// nodes, which is the only shape that can see the report ORDER. The corpus never writes
			// it: a mutant reversing the sort survived all one hundred and four imported cases.
			// Findings come out in source order, measured on the installed 8.67.0 build.
			sourceText: "type T = 'a' | 'b' | string;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "'a'",
					wantId:      "literalOverridden",
					wantMessage: `"a" is overridden by string in this union type.`,
				},
				{
					wantSpan:    "'b'",
					wantId:      "literalOverridden",
					wantMessage: `"b" is overridden by string in this union type.`,
				},
			},
		},
		{
			// The same with the primitive written BETWEEN the two literals, so source order and
			// discovery order differ.
			sourceText: "type T = 'a' | string | 'b';\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "'a'",
					wantId:      "literalOverridden",
					wantMessage: `"a" is overridden by string in this union type.`,
				},
				{
					wantSpan:    "'b'",
					wantId:      "literalOverridden",
					wantMessage: `"b" is overridden by string in this union type.`,
				},
			},
		},
		{
			// Two different primitives absorbing one literal each, which also pins that the
			// primitive iteration order does not leak into the report order.
			sourceText: "type T = 1 | 'a' | number | string;\n",
			wantFindings: []noRedundantTypeConstituentsFinding{
				{
					wantSpan:    "1",
					wantId:      "literalOverridden",
					wantMessage: "1 is overridden by number in this union type.",
				},
				{
					wantSpan:    "'a'",
					wantId:      "literalOverridden",
					wantMessage: `"a" is overridden by string in this union type.`,
				},
			},
		},
	}
	for index, testCase := range cases {
		t.Run(noRedundantTypeConstituentsCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoRedundantTypeConstituents,
				noRedundantTypeConstituentsFile, testCase.sourceText)

			wantIds := make([]string, len(testCase.wantFindings))
			for position, want := range testCase.wantFindings {
				wantIds[position] = want.wantId
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			// The harness writes the fixture trimmed, so the span slices are against that text
			// rather than the Go literal above. The captured spans came from a file written the
			// same way, so the two agree by construction rather than by luck.
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

// TestNoRedundantTypeConstituentsRequiresTheTypedHarness pins the nil-checker guard.
//
// A mutant removing that guard survives every other fixture here, and structurally so: the guard
// prevents a PANIC and no findings assertion can see one. It matters more than usual because the
// walk recovers per FILE rather than per rule, so one nil dereference costs every rule its verdict
// on that whole file rather than costing this rule one case.
//
// `NeedsTypeChecker` governs REGISTRATION only. The harness builds a Context by hand, which is what
// the untyped run below does, so the declaration is not the guard and the two have to be tested
// separately. The second half is the control: without something that fires, a rule which can never
// report at all satisfies the first half.
func TestNoRedundantTypeConstituentsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	const sourceText = "type T = string | any;\n"

	rule_testing.ExpectClean(t, rule_testing.Run(t, NoRedundantTypeConstituents,
		noRedundantTypeConstituentsFile, sourceText))

	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoRedundantTypeConstituents,
		noRedundantTypeConstituentsFile, sourceText), "overrides")
}
