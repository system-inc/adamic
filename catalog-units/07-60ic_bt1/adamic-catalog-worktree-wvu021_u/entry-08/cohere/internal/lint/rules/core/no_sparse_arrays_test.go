package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const sparseArrayFile = "/repository/source/Thing.ts"

func TestNoSparseArraysFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a hole in the middle", "export const Values = [1, , 3];\n"},
		{"a leading hole", "export const Values = [, 2];\n"},
		{"two holes", "export const Values = [1, , , 4];\n"},
		// A hole followed by a trailing comma. This is the case that separates the two commas the
		// rule reports from the one it must not, and reading only the last element would miss it.
		{"a hole with a trailing comma", "export const Values = [1, , 3, ];\n"},
		{"a nested array's hole", "export const Values = [[1, , 3]];\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoSparseArrays, sparseArrayFile, testCase.sourceText),
				"unexpectedSparseArray")
		})
	}
}

func TestNoSparseArraysStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a dense array", "export const Values = [1, 2, 3];\n"},
		// The decision boundary. A trailing comma is one comma after the last element and produces
		// no hole; a hole is two commas with nothing between them. Every formatter in this stack
		// emits the trailing form, so a rule that reported it would fight the formatter over syntax
		// that means nothing.
		{"a trailing comma", "export const Values = [1, 2, 3, ];\n"},
		{"a trailing comma on one element", "export const Values = [1, ];\n"},
		{"an explicit undefined", "export const Values = [1, undefined, 3];\n"},
		{"an empty array", "export const Values = [];\n"},
		{"a spread", "declare const other: number[];\nexport const Values = [1, ...other, 3];\n"},
		{"no array at all", "export const Value = 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoSparseArrays, sparseArrayFile, testCase.sourceText))
		})
	}
}

// A hole in a destructuring target skips a value; it builds no array. ESLint never sees these:
// ESTree spells an assignment target as an ArrayPattern and the rule listens only to
// ArrayExpression, while TypeScript's tree spells both as an ArrayLiteralExpression.
func TestNoSparseArraysSkipsDestructuringTargets(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// excalidraw packages/element/src/image.ts:138, verbatim
		{"an assignment target", "declare const match: string[];\nlet width = '';\nlet height = '';\n[, width, height] = match;\n"},
		{"a target nested in a target", "declare const x: string[][];\nlet a = '';\n[[, a]] = x;\n"},
		{"a target under an object property", "declare const y: { p: string[] };\nlet b = '';\n({ p: [, b] } = y);\n"},
		{"a for-of head", "declare const z: string[][];\nlet c = '';\nfor([, c] of z) {}\n"},
		{"a for-in head's array", "declare const w: Record<string, string>;\nlet d = '';\nfor([, d] in w) {}\n"},
		{"a parenthesized target", "declare const v: string[];\nlet e = '';\n([, e]) = v;\n"},
		{"a target with a default", "declare const u: string[][];\nlet f = '';\n[[, f] = []] = u;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoSparseArrays, sparseArrayFile, testCase.sourceText))
		})
	}
}

// The controls beside the destructuring cases: an array literal that is a value, even one sitting
// next to a target, still builds an array, so its hole still reports.
func TestNoSparseArraysStillFiresBesideDestructuring(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a declaration's holes", "export const holes = [1, , 2];\n"},
		{"an array on the right side", "let a = 0;\nlet b = 0;\n[a, b] = [1, , 2];\n"},
		{"an argument", "declare function take(values: unknown[]): void;\ntake([1, , 2]);\n"},
		{"a destructuring default's value", "declare const u: number[][];\nlet a = 0;\n[[a] = [1, , 2]] = u;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoSparseArrays, sparseArrayFile, testCase.sourceText),
				"unexpectedSparseArray")
		})
	}
}

// One finding per array rather than one per hole, so a two-hole array reports once. Reported on the
// literal rather than the hole because the hole has no text of its own to point at.
func TestNoSparseArraysReportsOncePerArray(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoSparseArrays, sparseArrayFile, "export const Values = [1, , , 4];\n")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("want exactly one finding for a two-hole array, got %d", len(result.Diagnostics))
	}
}
