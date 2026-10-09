package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const isNaNFile = "/repository/source/Thing.ts"

func TestUseIsNaNFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"strict equality", "declare const value: number;\nexport const Bad = value === NaN;\n"},
		{"loose equality", "declare const value: number;\nexport const Bad = value == NaN;\n"},
		{"strict inequality", "declare const value: number;\nexport const Bad = value !== NaN;\n"},
		// NaN on the left rather than the right. A check reading only one operand misses half.
		{"NaN on the left", "declare const value: number;\nexport const Bad = NaN === value;\n"},
		// Relational operators are always false too, which is why they are in the operator set.
		{"less than", "declare const value: number;\nexport const Bad = value < NaN;\n"},
		{"greater or equal", "declare const value: number;\nexport const Bad = value >= NaN;\n"},
		// The same value reached through the constructor.
		{"Number.NaN", "declare const value: number;\nexport const Bad = value === Number.NaN;\n"},
		{"parenthesized", "declare const value: number;\nexport const Bad = value === (NaN);\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, UseIsNaN, isNaNFile, testCase.sourceText),
				"comparisonWithNaN")
		})
	}
}

// The switch half, which the live config enables (enforceForSwitchCase: true) and which the first
// version of this rule did not implement at all.
//
// Found by applying a finding from no-var to this rule: enumerate the shapes a construct takes
// before writing the listener. A NaN comparison reaches the tree as a binary expression and as a
// switch, and the binary listener alone is silent on the second.
func TestUseIsNaNFiresOnSwitches(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantId     string
	}{
		{
			"a case label of NaN",
			"declare const value: number;\nexport function run() {\n    switch(value) {\n        case NaN: return 1;\n    }\n    return 0;\n}\n",
			"caseWithNaN",
		},
		{
			"a case label of Number.NaN",
			"declare const value: number;\nexport function run() {\n    switch(value) {\n        case Number.NaN: return 1;\n    }\n    return 0;\n}\n",
			"caseWithNaN",
		},
		// Switching on NaN kills every case at once, so it reports once on the switch rather than
		// once per clause, and a switch with no cases at all still reports.
		{
			"switching on NaN",
			"declare const other: number;\nexport function run() {\n    switch(NaN) {\n        case other: return 1;\n    }\n    return 0;\n}\n",
			"switchOnNaN",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, UseIsNaN, isNaNFile, testCase.sourceText), testCase.wantId)
		})
	}
}

// The option is read rather than ignored, in both directions.
func TestUseIsNaNSwitchOption(t *testing.T) {
	t.Parallel()

	source := "declare const value: number;\nexport function run() {\n    switch(value) {\n        case NaN: return 1;\n    }\n    return 0;\n}\n"

	// Default is on, so a config that says nothing gets the rule rather than half of it.
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, UseIsNaN, isNaNFile, source), "caseWithNaN")

	disabled := false
	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, UseIsNaN, isNaNFile, source,
		UseIsNaNOptions{EnforceForSwitchCase: &disabled}))

	enabled := true
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, UseIsNaN, isNaNFile, source,
		UseIsNaNOptions{EnforceForSwitchCase: &enabled}), "caseWithNaN")
}

func TestUseIsNaNStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"the correct check", "declare const value: number;\nexport const Ok = Number.isNaN(value);\n"},
		// The decision boundary is "does an operand name NaN", and the way to get it wrong is to
		// match the letters. Each of these contains them and none is the value.
		{"a similarly named binding", "declare const isNaNLike: boolean;\nexport const Ok = isNaNLike === true;\n"},
		{"a property named NaN on something else", "declare const table: { NaN: number };\ndeclare const value: number;\nexport const Ok = value === table.NaN;\n"},
		{"the letters in a string", "declare const value: string;\nexport const Ok = value === 'NaN';\n"},
		// Not a comparison. Arithmetic with NaN is legal and produces NaN, which is not this
		// rule's business; only comparisons are always-constant.
		{"arithmetic rather than comparison", "declare const value: number;\nexport const Ok = value + NaN;\n"},
		{"a plain comparison", "declare const value: number;\nexport const Ok = value === 0;\n"},
		// A switch with no NaN anywhere. The boundary for the switch half is "does a label or the
		// discriminant name NaN", not "is this a switch".
		{"a switch without NaN", "declare const value: number;\nexport function run() {\n    switch(value) {\n        case 0: return 1;\n    }\n    return 0;\n}\n"},
		{"a default clause", "declare const value: number;\nexport function run() {\n    switch(value) {\n        default: return 1;\n    }\n}\n"},
		// A local named NaN or Number is not the global, and upstream is silent on each (#9g0v4j6).
		// This rule matched by spelling and reported them until it asked the checker, which is why
		// these run through the typed harness: without a checker the answer is the global.
		{"a local NaN", "declare const value: number;\nconst NaN = 1;\nexport const Ok = value === NaN;\n"},
		{"a parameter named NaN", "declare const value: number;\nexport function run(NaN: number) {\n    return value === NaN;\n}\n"},
		{"a parameter named Number", "declare const value: number;\nexport function run(Number: { NaN: number }) {\n    return value === Number.NaN;\n}\n"},
		{"a parameter named Number, read in brackets", "declare const value: number;\nexport function run(Number: Record<string, number>) {\n    return value === Number['NaN'];\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, UseIsNaN, isNaNFile, testCase.sourceText))
		})
	}
}
