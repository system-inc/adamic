package nexus

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const suffixFile = "/repository/source/Thing.ts"

func TestConsistencyRequireTypeSuffixFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantId     string
	}{
		{"bare type alias", "export type Order = { id: string };\n", "noTypeAliasSuffix"},
		{"bare interface", "export interface Order {\n    id: string;\n}\n", "noInterfaceSuffix"},
		{"bare const enum", "export const Operator = { Add: 'Add', Sub: 'Sub' } as const;\n", "noConstEnumSuffix"},
		// A suffix that is merely contained is not a suffix. "TypeName" does not end in "Type".
		{"suffix in the middle", "export type TypeName = { id: string };\n", "noTypeAliasSuffix"},
		// An interface may not use "Type", which the alias list allows. The two lists differ on
		// purpose, so a rule that shared one list would let this through.
		{"interface ending in Type", "export interface OrderType {\n    id: string;\n}\n", "noInterfaceSuffix"},
		{"single-property const enum", "export const Operator = { Add: 'Add' } as const;\n", "noConstEnumSuffix"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyRequireTypeSuffix, suffixFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantId)
		})
	}
}

func TestConsistencyRequireTypeSuffixStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"type alias with Type", "export type OrderType = { id: string };\n"},
		{"type alias with Properties", "export type OrderProperties = { id: string };\n"},
		{"type alias with Interface", "export type OrderInterface = { id: string };\n"},
		{"type alias with Options", "export type OrderOptions = { id: string };\n"},
		{"interface with Interface", "export interface OrderInterface {\n    id: string;\n}\n"},
		{"interface with Properties", "export interface OrderProperties {\n    id: string;\n}\n"},
		{"interface with Options", "export interface OrderOptions {\n    id: string;\n}\n"},
		{"const enum with Kind", "export const OperatorKind = { Add: 'Add' } as const;\n"},

		// These are the discriminating cases for the const-enum detector. Each is structurally
		// `{ ... } as const` and none is an enum, so a detector that keyed on the wrapper alone
		// would ask for a rename that would be wrong.
		{"lookup table with unrelated values", "export const ClassNames = { small: 'text-sm', large: 'text-lg' } as const;\n"},
		{"one entry that does not match", "export const Operator = { Add: 'Add', Sub: 'Subtract' } as const;\n"},
		{"computed keys are a map over an existing kind", "export const Labels = { [OperatorKind.Add]: 'Add' } as const;\n"},
		{"empty object", "export const Empty = {} as const;\n"},
		{"non-string values", "export const Sizes = { Small: 1, Large: 2 } as const;\n"},

		// A plain object without `as const` is not an enum shape at all.
		{"object without as const", "export const Operator = { Add: 'Add' };\n"},
		{"ordinary constant", "export const maximumRetries = 3;\n"},
		{"no declarations", "export function run() {\n    return 1;\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyRequireTypeSuffix, suffixFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}
