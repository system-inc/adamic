package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const deleteVarFile = "/repository/source/Thing.ts"

func TestNoDeleteVarFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a bare name", "export function run() {\n    let value = 1;\n    delete value;\n}\n"},
		// Grouping is what a formatter can introduce, so a rule that missed it would report the
		// same code differently before and after formatting.
		{"a parenthesized name", "export function run() {\n    let value = 1;\n    delete (value);\n}\n"},
		{"a parameter", "export function run(value: number) {\n    delete value;\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoDeleteVar, deleteVarFile, testCase.sourceText),
				"unexpectedDeleteVar")
		})
	}
}

func TestNoDeleteVarStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The boundary is "bare name versus property access", and every legitimate use of delete
		// is on the other side of it.
		{"a dotted property", "declare const record: { key?: number };\nexport function run() {\n    delete record.key;\n}\n"},
		{"a computed property with a literal", "declare const record: Record<string, number>;\nexport function run() {\n    delete record['key'];\n}\n"},
		{"a computed property with a variable", "declare const record: Record<string, number>;\nexport function run(key: string) {\n    delete record[key];\n}\n"},
		{"a nested property", "declare const outer: { inner: { key?: number } };\nexport function run() {\n    delete outer.inner.key;\n}\n"},
		// A property access wrapped in parentheses is still a property access, so stripping the
		// grouping must not turn it into a bare name.
		{"a parenthesized property access", "declare const record: { key?: number };\nexport function run() {\n    delete (record.key);\n}\n"},
		{"no delete at all", "export function run() {\n    return 1;\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoDeleteVar, deleteVarFile, testCase.sourceText))
		})
	}
}
