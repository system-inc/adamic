package typescript

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const wrapperObjectFile = "/repository/source/Thing.ts"

// The fixtures below are seeded from oxc's own pass and fail vectors rather than invented, which is
// the discipline that found a shipped false positive in this package. A porter's fixtures encode the
// porter's beliefs; the upstream corpus is the only source of cases independent of them.

func TestNoWrapperObjectTypesFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"BigInt", "export let value: BigInt;\n"},
		{"Boolean", "export let value: Boolean;\n"},
		{"Number", "export let value: Number;\n"},
		{"Object", "export let value: Object;\n"},
		{"String", "export let value: String;\n"},
		{"Symbol", "export let value: Symbol;\n"},
		// Nested positions, all from oxc's fail vector. A listener keyed to declarations rather than
		// to type references reaches none of these.
		{"inside a property type", "export let value: { property: Number };\n"},
		{"in a type alias", "export type MyType = Number;\n"},
		{"in a tuple", "export type MyType = [Number];\n"},
		{"in an assertion", "export const Value = 0 as Number;\n"},
		// `implements` is a type position even though `extends` is not, which is the distinction this
		// rule exists to get right.
		{"a class implementing it", "export class MyClass implements Number {}\n"},
		{"an interface extending it", "export interface MyInterface extends Number {}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoWrapperObjectTypes, wrapperObjectFile, testCase.sourceText),
				"bannedWrapperObjectType")
		})
	}
}

// TestNoWrapperObjectTypesReportsEachWrapper pins that a type naming two wrappers reports twice.
//
// oxc lists `Number & String` in its fail vector without saying how many diagnostics it produces.
// One report for the whole annotation would leave the second name unfixed and the author would
// repair half the problem, so the count is asserted rather than assumed.
func TestNoWrapperObjectTypesReportsEachWrapper(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.Run(t, NoWrapperObjectTypes, wrapperObjectFile, "export type MyType = Number & String;\n"),
		"bannedWrapperObjectType", "bannedWrapperObjectType")
}

// TestNoWrapperObjectTypesFixesToThePrimitive proves what the fix actually writes.
//
// A fix is the one part of a rule that rewrites source, so a wrong range or wrong text changes
// something else silently. Each wrapper maps to its own primitive, and getting that mapping wrong
// would produce code that compiles and means something different.
func TestNoWrapperObjectTypesFixesToThePrimitive(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSource string
	}{
		{"Number", "export let value: Number;\n", "export let value: number;\n"},
		{"String", "export let value: String;\n", "export let value: string;\n"},
		{"BigInt", "export let value: BigInt;\n", "export let value: bigint;\n"},
		{"Boolean", "export let value: Boolean;\n", "export let value: boolean;\n"},
		{"Object", "export let value: Object;\n", "export let value: object;\n"},
		{"Symbol", "export let value: Symbol;\n", "export let value: symbol;\n"},
		// The fix replaces only the name, leaving the rest of the annotation intact.
		{"inside a property type", "export let value: { property: Number };\n", "export let value: { property: number };\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t,
				rule_testing.Run(t, NoWrapperObjectTypes, wrapperObjectFile, testCase.sourceText),
				testCase.wantSource)
		})
	}
}

// TestNoWrapperObjectTypesOffersNoFixForImplements pins the one place a fix is withheld.
//
// `class C implements number {}` does not compile, so replacing the name there would hand the author
// a repair that breaks the build. oxc makes the same distinction.
func TestNoWrapperObjectTypesOffersNoFixForImplements(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoWrapperObjectTypes, wrapperObjectFile, "export class MyClass implements Number {}\n")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
	}
	if len(result.Diagnostics[0].Fixes) != 0 {
		t.Fatalf("expected no fix on an implements clause, got %d", len(result.Diagnostics[0].Fixes))
	}
}

func TestNoWrapperObjectTypesStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"the primitive", "export let value: number;\n"},
		{"another primitive", "export let value: bigint;\n"},
		{"a name that merely contains it", "export let value: NumberLike;\n"},
		{"an unrelated type", "export let value: Other;\n"},
		{"a signature", "export let value: () => void;\n"},
		// The case that is easy to get backwards, and it is oxc's, not mine. A class extends a
		// *value*, and `Number` the value is a real constructor, so this is legal TypeScript that
		// means what it says. Only `implements` puts the name back in a type position.
		{"a class extending it", "export class MyClass extends Number {}\n"},
		// Wrapper names as values rather than types. The rule is about type positions.
		{"a variable named Number", "let Number;\nexport const Held = Number;\n"},
		{"a constructor call", "export const Made = String(1);\n"},
		// Shadowing, from oxc's pass vector. A file that declares its own `Number` has redefined the
		// name and the rule must not flag the author's own type.
		{"a shadowing type alias", "type Number = 0 | 1;\nexport let value: Number;\n"},
		{"a shadowing interface", "interface Number {\n    tag: string;\n}\nexport let value: Number;\n"},
		// An enum shadows too, and that was measured rather than assumed. An enum declares a value
		// and a TYPE of the same name, so the annotation below resolves to the enum: asked through
		// the checker, `Number` here resolves to a declaration in this file, while without the enum
		// it resolves into a library declaration file. The shadow walk counted only interface, type
		// alias and class until it was lifted onto the shelf, so this was a false positive on a name
		// the author owns.
		{"a shadowing enum", "enum Number {\n    A,\n}\nexport let value: Number;\n"},
		// Nested shadowing. A top-level scan misses this, which was a real false positive in this
		// package before oxc's corpus was read.
		{"a block-scoped shadow", "{\n    type Number = 0 | 1;\n    let value: Number;\n}\n"},
		// A near-miss name that is not one of the six.
		{"a similarly-cased non-wrapper", "export let value: Never;\n"},
		{"a qualified name", "export let value: Foo.Number;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoWrapperObjectTypes, wrapperObjectFile, testCase.sourceText))
		})
	}
}
