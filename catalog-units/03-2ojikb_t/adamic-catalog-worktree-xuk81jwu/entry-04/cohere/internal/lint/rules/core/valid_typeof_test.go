package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const typeofFile = "/repository/source/Thing.ts"

func TestValidTypeofFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a misspelling", "declare const value: unknown;\nexport const Bad = typeof value === 'strng';\n"},
		// The two that are not typos but wrong beliefs about the language. An array reports as
		// 'object' and so does null, which is the oldest wart in JavaScript.
		{"array is not a typeof result", "declare const value: unknown;\nexport const Bad = typeof value === 'array';\n"},
		{"null is not a typeof result", "declare const value: unknown;\nexport const Bad = typeof value === 'null';\n"},
		{"typeof on the right", "declare const value: unknown;\nexport const Bad = 'array' === typeof value;\n"},
		{"inequality", "declare const value: unknown;\nexport const Bad = typeof value !== 'array';\n"},
		{"a template literal with no substitutions", "declare const value: unknown;\nexport const Bad = typeof value === `array`;\n"},
		{"parenthesized typeof", "declare const value: unknown;\nexport const Bad = (typeof value) === 'array';\n"},
		// A literal that is not a string at all, which no result of typeof can equal.
		{"the null literal", "declare const value: unknown;\nexport const Bad = typeof value === null;\n"},
		{"a number literal", "declare const value: unknown;\nexport const Bad = typeof value == 5;\n"},
		{"a boolean literal", "declare const value: unknown;\nexport const Bad = typeof value !== false;\n"},
		// The value undefined where the string was meant, upstream's own corpus row.
		{"the global undefined", "declare const value: unknown;\nexport const Bad = typeof value !== undefined;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, ValidTypeof, typeofFile, testCase.sourceText),
				"invalidTypeofValue")
		})
	}
}

func TestValidTypeofStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// All eight legal results. These are the decision boundary: the rule's entire job is
		// knowing exactly this set, and a rule that dropped one would report correct code.
		{"undefined", "declare const value: unknown;\nexport const Ok = typeof value === 'undefined';\n"},
		{"object", "declare const value: unknown;\nexport const Ok = typeof value === 'object';\n"},
		{"boolean", "declare const value: unknown;\nexport const Ok = typeof value === 'boolean';\n"},
		{"number", "declare const value: unknown;\nexport const Ok = typeof value === 'number';\n"},
		{"string", "declare const value: unknown;\nexport const Ok = typeof value === 'string';\n"},
		{"function", "declare const value: unknown;\nexport const Ok = typeof value === 'function';\n"},
		{"symbol", "declare const value: unknown;\nexport const Ok = typeof value === 'symbol';\n"},
		{"bigint", "declare const value: unknown;\nexport const Ok = typeof value === 'bigint';\n"},

		// Unanalyzable without the checker, so left alone rather than guessed at.
		{"typeof compared to typeof", "declare const a: unknown;\ndeclare const b: unknown;\nexport const Ok = typeof a === typeof b;\n"},
		{"typeof compared to a binding", "declare const value: unknown;\ndeclare const expected: string;\nexport const Ok = typeof value === expected;\n"},
		{"a template with a substitution", "declare const value: unknown;\ndeclare const rest: string;\nexport const Ok = typeof value === `str${rest}`;\n"},
		// A parameter named undefined is a binding like any other, and upstream reads it as one.
		{"a local named undefined", "export function check(value: unknown, undefined: string) { return typeof value === undefined; }\n"},
		// No typeof at all: a wrong-looking string on its own is not this rule's business.
		{"a plain string comparison", "declare const value: string;\nexport const Ok = value === 'array';\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, ValidTypeof, typeofFile, testCase.sourceText))
		})
	}
}

func TestValidTypeofRequireStringLiterals(t *testing.T) {
	t.Parallel()

	options := ValidTypeofOptions{RequireStringLiterals: true}
	fires := []struct {
		name       string
		sourceText string
		want       string
	}{
		{"a binding", "declare const value: unknown;\ndeclare const expected: string;\nexport const Bad = typeof value === expected;\n", "notString"},
		{"a binding on the left", "declare const value: unknown;\ndeclare const expected: string;\nexport const Bad = expected == typeof value;\n", "notString"},
		{"a template with a substitution", "declare const value: unknown;\ndeclare const rest: string;\nexport const Bad = typeof value === `str${rest}`;\n", "notString"},
		// The global undefined reads as notString under the option, as upstream switches its id.
		{"the global undefined", "declare const value: unknown;\nexport const Bad = typeof value === undefined;\n", "notString"},
		// A local undefined is no longer the global, so it falls to the option like any binding.
		{"a local named undefined", "export function check(value: unknown, undefined: string) { return typeof value === undefined; }\n", "notString"},
		// A literal is still judged against the eight, so a bad one stays invalidTypeofValue.
		{"a misspelled literal", "declare const value: unknown;\nexport const Bad = typeof value == 'invalid string';\n", "invalidTypeofValue"},
	}
	for _, testCase := range fires {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTypedWithOptions(t, ValidTypeof, typeofFile, testCase.sourceText, options), testCase.want)
		})
	}

	silent := []struct {
		name       string
		sourceText string
	}{
		{"a string literal", "declare const value: unknown;\nexport const Ok = typeof value === 'number';\n"},
		{"a template with no substitutions", "declare const value: unknown;\nexport const Ok = `object` === typeof value;\n"},
		{"typeof compared to typeof", "declare const a: unknown;\ndeclare const b: unknown;\nexport const Ok = typeof a === typeof b;\n"},
		{"typeof outside a comparison", "declare const value: unknown;\nexport const Ok = typeof value + 'thing';\n"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTypedWithOptions(t, ValidTypeof, typeofFile, testCase.sourceText, options))
		})
	}
}

// The global undefined carries upstream's one suggestion, quoting it, and nothing else carries any.
func TestValidTypeofSuggestsQuotingUndefined(t *testing.T) {
	t.Parallel()

	sourceText := "declare const value: unknown;\nexport const Bad = typeof value === undefined;\n"
	result := rule_testing.RunTyped(t, ValidTypeof, typeofFile, sourceText)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one diagnostic, got %d", len(result.Diagnostics))
	}
	suggestions := result.Diagnostics[0].Suggestions
	if len(suggestions) != 1 || len(suggestions[0].Fixes) != 1 || len(result.Diagnostics[0].Fixes) != 0 {
		t.Fatalf("wanted one suggestion carrying one fix and no automatic fix, got %d suggestions", len(suggestions))
	}
	if suggestions[0].Message.Id != "suggestString" {
		t.Fatalf("wanted suggestString, got %q", suggestions[0].Message.Id)
	}
	fix := suggestions[0].Fixes[0]
	rewritten := sourceText[:fix.Range.Pos()] + fix.Text + sourceText[fix.Range.End():]
	want := "declare const value: unknown;\nexport const Bad = typeof value === \"undefined\";\n"
	if rewritten != want {
		t.Fatalf("applying the suggestion gave %q, wanted %q", rewritten, want)
	}

	misspelled := rule_testing.RunTyped(t, ValidTypeof, typeofFile,
		"declare const value: unknown;\nexport const Bad = typeof value === 'strng';\n")
	if len(misspelled.Diagnostics) != 1 || len(misspelled.Diagnostics[0].Suggestions) != 0 {
		t.Fatalf("a misspelled literal should report once with no suggestion")
	}
}

// The decoder takes upstream's empty object and refuses a key upstream's schema does not declare.
func TestValidTypeofOptionsDecode(t *testing.T) {
	t.Parallel()

	decode := rule.DecodeOptionsInto[ValidTypeofOptions]()
	decoded, err := decode([]byte(`{}`))
	if err != nil {
		t.Fatalf("an empty object should decode: %v", err)
	}
	if decoded.(ValidTypeofOptions).RequireStringLiterals {
		t.Fatalf("requireStringLiterals should default to false")
	}
	decoded, err = decode([]byte(`{"requireStringLiterals": true}`))
	if err != nil || !decoded.(ValidTypeofOptions).RequireStringLiterals {
		t.Fatalf("requireStringLiterals should decode true, got %v, %v", decoded, err)
	}
	if _, err := decode([]byte(`{"requireStringLiteral": true}`)); err == nil {
		t.Fatalf("an unknown key should be refused")
	}
	if _, err := decode([]byte(`{"RequireStringLiterals": true}`)); err == nil {
		t.Fatalf("a key spelled in a different case should be refused")
	}
}
