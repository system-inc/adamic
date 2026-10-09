package typescript

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const emptyObjectFile = "/repository/source/Thing.ts"

// Upstream's own rows, all of them, replay in no_empty_object_type_corpus_test.go, suggestions and the
// shapes a rewrite keeps or drops included. These were the oxc port's vectors, kept for the shapes the
// replay does not write: values that look like `{}` and are not types.

func TestNoEmptyObjectTypeFiresOnInterfaces(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		id         string
	}{
		{"an empty interface", "export interface Base {}\n", "noEmptyInterface"},
		// Extending exactly one name is the supertype under a second name, flagged under the default
		// `allowInterfaces: never` with its own message.
		{"extending a single name", "interface Base {\n    name: string;\n}\nexport interface Derived extends Base {}\n", "noEmptyInterfaceWithSuper"},
		{"extending a generic", "export interface Base extends Array<number> {}\n", "noEmptyInterfaceWithSuper"},
		{"a generic extending a generic", "interface Derived<T> {\n    value: T;\n}\nexport interface Base<T> extends Derived<T> {}\n", "noEmptyInterfaceWithSuper"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoEmptyObjectType, emptyObjectFile, testCase.sourceText),
				testCase.id)
		})
	}
}

func TestNoEmptyObjectTypeFiresOnObjectTypes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a type alias", "export type Base = {};\n"},
		{"an annotation", "export let value: {};\n"},
		{"inside a union", "export type MyUnion<T> = T | {};\n"},
		{"unioned with null", "export type Base = {} | null;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoEmptyObjectType, emptyObjectFile, testCase.sourceText),
				"noEmptyObject")
		})
	}
}

// TestNoEmptyObjectTypeOffersTwoSuggestions pins that the repair is a choice rather than a fix.
//
// `object` and `unknown` mean different things and only the author knows which was meant, so a rule
// that applied one unattended would be changing the type rather than repairing a spelling. Asserting
// the count is what keeps a later change from quietly collapsing them into one automatic fix.
func TestNoEmptyObjectTypeOffersTwoSuggestions(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoEmptyObjectType, emptyObjectFile, "export let value: {};\n")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
	}
	if got := len(result.Diagnostics[0].Suggestions); got != 2 {
		t.Fatalf("expected two suggestions, got %d", got)
	}
	if got := len(result.Diagnostics[0].Fixes); got != 0 {
		t.Fatalf("expected no automatic fix, got %d", got)
	}

	// What each suggestion would write, which counting them cannot see.
	//
	// A mutation making both suggestions produce `unknown` compiled and changed nothing here, because
	// this test asserted how many were offered and never what they said. The two are not
	// interchangeable: `object` means any non-primitive and `unknown` means any value at all, so a
	// reader choosing between them is choosing between different types, and a suggestion offering
	// the same one twice is a menu with one item wearing two labels.
	//
	// Asserted by applying each fix to the source rather than by comparing its text, so the span is
	// pinned at the same time: a suggestion writing the right word over the wrong range is the other
	// half of this defect and reads identically from the text alone.
	const source = "export let value: {};\n"
	wanted := []string{"export let value: object;\n", "export let value: unknown;\n"}
	for index, suggestion := range result.Diagnostics[0].Suggestions {
		if len(suggestion.Fixes) != 1 {
			t.Fatalf("suggestion %d carries %d fixes, wanted one", index, len(suggestion.Fixes))
		}
		fix := suggestion.Fixes[0]
		rewritten := source[:fix.Range.Pos()] + fix.Text + source[fix.Range.End():]
		if rewritten != wanted[index] {
			t.Fatalf("applying suggestion %d gave %q, wanted %q", index, rewritten, wanted[index])
		}
	}
}

func TestNoEmptyObjectTypeStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an interface with members", "export interface Base {\n    name: string;\n}\n"},
		// The case the whole rule turns on: extending multiple interfaces is how a reader expresses an
		// intersection of named types where a union would be wrong. Two or more is allowed; exactly one
		// is not.
		{"extending two names", "interface Base {\n    name: string;\n}\ninterface Derived {\n    age: number;\n}\nexport interface Both extends Base, Derived {}\n"},
		{"extending three names", "interface A {\n    a: string;\n}\ninterface B {\n    b: string;\n}\ninterface C {\n    c: string;\n}\nexport interface All extends A, B, C {}\n"},
		// `Base & {}` inside an intersection is exempt: the empty half is doing real work there.
		{"inside an intersection", "interface Base {\n    name: string;\n}\nexport type Expanded = Base & {};\n"},
		// A non-empty object type is not this rule's concern in either position.
		{"a populated type literal", "export type Base = { name: string };\n"},
		{"a populated annotation", "export let value: { name: string };\n"},
		// An empty *object literal* is a value, not a type. A rule keyed to braces rather than to
		// type nodes would fire on every one of these.
		{"an empty object value", "export const Value = {};\n"},
		{"an empty object argument", "export const Result = JSON.stringify({});\n"},
		{"an empty destructuring default", "export function run({ a } = { a: 1 }) {\n    return a;\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoEmptyObjectType, emptyObjectFile, testCase.sourceText))
		})
	}
}
