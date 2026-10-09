package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const typeSuffixFile = "/repository/source/components/Button.tsx"

func TestReactComponentRequirePropertiesTypeSuffixFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"a component whose type is declared elsewhere",
			"import type { ButtonProps } from './Types';\n" +
				"export function Button(properties: ButtonProps) {\n    return <button>{properties.label}</button>;\n}\n",
			[]string{"useComponentPropertyTypeSuffix"},
		},
		{
			"an arrow component",
			"import type { ButtonProps } from './Types';\n" +
				"export const Button = (properties: ButtonProps) => <button>{properties.label}</button>;\n",
			[]string{"useComponentPropertyTypeSuffix"},
		},
		// Two findings for one wrong name, at the use and at the declaration. A reader fixing this
		// touches both, and seeing only one would leave the rename half done.
		{
			"an interface declared after the component",
			"export function Button(properties: ButtonInterface) {\n" +
				"    return <button>{properties.label}</button>;\n}\n" +
				"interface ButtonInterface { label: string }\n",
			[]string{"useComponentPropertyTypeSuffix", "useInterfaceSuffix"},
		},
		{
			"a type alias declared after the component",
			"export function Button(properties: ButtonProps) {\n" +
				"    return <button>{properties.label}</button>;\n}\n" +
				"type ButtonProps = { label: string };\n",
			[]string{"useComponentPropertyTypeSuffix", "useTypeAliasSuffix"},
		},
		// A name holding Properties as a substring rather than a suffix is still wrong.
		{
			"a type with Properties in the middle",
			"import type { ButtonPropertiesBag } from './Types';\n" +
				"export function Button(properties: ButtonPropertiesBag) {\n" +
				"    return <button>{properties.label}</button>;\n}\n",
			[]string{"useComponentPropertyTypeSuffix"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, ReactComponentRequirePropertiesTypeSuffix,
				typeSuffixFile, testCase.sourceText), testCase.wantIds...)
		})
	}
}

func TestReactComponentRequirePropertiesTypeSuffixStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The shape the rule asks for.
		{
			"a correctly suffixed type",
			typeSuffixFile,
			"interface ButtonProperties { label: string }\n" +
				"export function Button(properties: ButtonProperties) {\n" +
				"    return <button>{properties.label}</button>;\n}\n",
		},
		// The component-name gate. A lowercase function's parameter type is nobody's business.
		{
			"a lowercase function taking a badly named type",
			typeSuffixFile,
			"import type { ButtonProps } from './Types';\n" +
				"export function build(properties: ButtonProps) {\n    return <button />;\n}\n",
		},
		// An inline type has no name to judge, and a destructured parameter has no type reference.
		{
			"an inline object type",
			typeSuffixFile,
			"export function Button(properties: { label: string }) {\n" +
				"    return <button>{properties.label}</button>;\n}\n",
		},
		{
			"a destructured parameter",
			typeSuffixFile,
			"export function Button({ label }: { label: string }) {\n    return <button>{label}</button>;\n}\n",
		},
		// A destructured parameter with a badly named type reference. This is the shape the tree
		// actually has, in MenuItem.tsx, and the rule reported it until the identifier check was
		// added: a genuine violation of the convention that the gate does not report, which would
		// have become a second only-cohere divergence arrived at by accident rather than a ruling.
		//
		// Destructuring is `react-component-no-destructuring`'s business, and a destructured
		// parameter's type is often composed (`TypeA & { ... }`) rather than the single name this
		// rule renames.
		{
			"a destructured parameter with a badly named type",
			typeSuffixFile,
			"import type { MenuItemInterface } from './Types';\n" +
				"export function MenuItem({ label, ...buttonProperties }: MenuItemInterface) {\n" +
				"    return <button {...buttonProperties}>{label}</button>;\n}\n",
		},
		{
			"a component with no parameters",
			typeSuffixFile,
			"export function Button() {\n    return <button />;\n}\n",
		},
		// An interface not used as a component's property type is not this rule's business, even
		// when it is badly named.
		{
			"an unrelated interface",
			typeSuffixFile,
			"interface ButtonProps { label: string }\n" +
				"export function Button(properties: { label: string }) {\n" +
				"    return <button>{properties.label}</button>;\n}\n",
		},
		// Outside a React file the rule declines before looking.
		{
			"a component-shaped function in a ts file",
			"/repository/source/api/Thing.ts",
			"import React from 'react';\nimport type { ButtonProps } from './Types';\n" +
				"export function Button(properties: ButtonProps) {\n" +
				"    React.useEffect(() => {}, []);\n    return properties;\n}\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, ReactComponentRequirePropertiesTypeSuffix,
				testCase.fileName, testCase.sourceText))
		})
	}
}

// Declaration order used to decide whether the declaration itself was reported, and that was wrong.
//
// The previous version of this test pinned the order dependence deliberately, reasoning that the
// use site is reported either way and that fixing it costs a second walk of the file for a second
// finding about the same rename. Both halves of that were true. The conclusion was still wrong,
// and the reason is the fix rather than the finding.
//
// On the declaration-first ordering the rule emitted one finding carrying one fix, and that fix
// renamed the *usage*. So `CardProps` became `CardProperties` at the call site while the
// declaration kept its old name, and the result parses and does not compile:
//
//	dangling.ts(2,34): error TS2304: Cannot find name 'CardProperties'
//
// A fix engine's parse guard structurally cannot refuse that, because the output is syntactically
// valid. And declaration-first is the conventional ordering, so it was the common case.
//
// The trade was priced as "a second finding about the same rename" when what it actually bought was
// "the fix does not break the file". Recorded at length because the reasoning was careful and still
// reached the wrong answer: it weighed the cost of the walk against the value of the finding, and
// the finding was not the thing at stake.
//
// The fix derives the corrected name rather than appending to the wrong one.
func TestReactComponentRequirePropertiesTypeSuffixFixDerivesTheName(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFixedSource(t,
		rule_testing.Run(t, ReactComponentRequirePropertiesTypeSuffix, typeSuffixFile,
			"import type { ButtonProps } from './Types';\n"+
				"export function Button(properties: ButtonProps) {\n    return <button />;\n}\n"),
		"import type { ButtonProps } from './Types';\n"+
			"export function Button(properties: ButtonProperties) {\n    return <button />;\n}\n")

	// An Interface suffix is replaced rather than appended to, which is the case that separates a
	// derived rename from a blind one: appending gives ButtonInterfaceProperties.
	rule_testing.ExpectFixedSource(t,
		rule_testing.Run(t, ReactComponentRequirePropertiesTypeSuffix, typeSuffixFile,
			"import type { ButtonInterface } from './Types';\n"+
				"export function Button(properties: ButtonInterface) {\n    return <button />;\n}\n"),
		"import type { ButtonInterface } from './Types';\n"+
			"export function Button(properties: ButtonProperties) {\n    return <button />;\n}\n")

	// A name with no known suffix gets one appended.
	rule_testing.ExpectFixedSource(t,
		rule_testing.Run(t, ReactComponentRequirePropertiesTypeSuffix, typeSuffixFile,
			"import type { ButtonBag } from './Types';\n"+
				"export function Button(properties: ButtonBag) {\n    return <button />;\n}\n"),
		"import type { ButtonBag } from './Types';\n"+
			"export function Button(properties: ButtonBagProperties) {\n    return <button />;\n}\n")
}

// Both orderings, asserted separately, because either alone passes a broken rule.
//
// This is the fixture the original pair could not supply. The rule filled its map of
// component property types from the component listener and read it from the declaration listener,
// one walk in source order, so an interface declared above its component was visited first, saw an
// empty map, and was skipped.
//
// The failure was silent and it produced uncompilable source. The usage still got its fix, so
// `CardProps` became `CardProperties` at the call site while the declaration kept its old name, and
// the result parses. A fix engine's parse guard structurally cannot refuse that.
//
// Declaration-first is the conventional ordering, so this was the common case rather than an edge,
// and the existing fixtures passed because on that ordering the rule emits one finding rather than
// two. `ExpectFindings` asserts a count and cannot tell a missing second finding from correct
// behavior.
func TestReactComponentRequirePropertiesTypeSuffixIsOrderIndependent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{
			name: "the declaration comes first",
			sourceText: "interface CardProps { a: string }\n" +
				"export function Card(properties: CardProps) { return null; }\n",
		},
		{
			name: "the component comes first",
			sourceText: "export function Card(properties: CardProps) { return null; }\n" +
				"interface CardProps { a: string }\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ReactComponentRequirePropertiesTypeSuffix,
				typeSuffixFile, testCase.sourceText)

			// Both halves have to be renamed or the file stops compiling, so the count is the
			// assertion rather than a detail of it.
			if len(result.Diagnostics) != 2 {
				t.Fatalf("wanted two findings, one for the usage and one for the declaration, got %d",
					len(result.Diagnostics))
			}

			// And both have to carry a fix. A finding with no fix leaves the same broken half
			// behind while looking like the rule saw it.
			for index, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) != 1 {
					t.Fatalf("finding %d carries %d fixes, wanted one", index, len(diagnostic.Fixes))
				}
			}
		})
	}
}

// A rename onto a name the file already declares is refused, and the finding stays.
//
// Found on Kirk's real tree rather than in a corpus. `MenuItem.tsx` declares both spellings, which
// is what a codebase mid-migration looks like by definition, and the fix rewrote
//
//	export type MenuItemInterface = MenuItemProperties & { ... }
//
// into a type alias whose name is its own right-hand side:
//
//	TS2300: Duplicate identifier 'MenuItemProperties'
//	TS2456: Type alias 'MenuItemProperties' circularly references itself
//
// Both parse. Same class as the ordering defect this rule shipped an hour earlier: valid syntax,
// broken program, and a fix engine's refusal guard has nothing to catch.
//
// The rule renames toward a fixed convention, so a collision is not an exotic case. Any file
// holding both spellings has one, and holding both spellings is exactly what migrating looks like.
//
// The finding is kept and only the fix is withheld. The name still violates the convention, and
// which of the two declarations survives is a decision only the author can make.
func TestReactComponentRequirePropertiesTypeSuffixRefusesATakenName(t *testing.T) {
	t.Parallel()

	source := "export interface MenuItemProperties { a: string }\n" +
		"export type MenuItemInterface = MenuItemProperties;\n" +
		"export function MenuItem(properties: MenuItemInterface) { return null; }\n"

	result := rule_testing.Run(t, ReactComponentRequirePropertiesTypeSuffix, typeSuffixFile, source)
	if len(result.Diagnostics) == 0 {
		t.Fatal("wanted the convention violation still reported, got no findings")
	}
	for index, diagnostic := range result.Diagnostics {
		if len(diagnostic.Fixes) != 0 {
			t.Fatalf("finding %d proposes %d fixes onto a name the file already declares, wanted none",
				index, len(diagnostic.Fixes))
		}
	}
}

// The ordinary case still carries its fix, which is what keeps the guard from being a silencer.
//
// Without this, disabling every fix in the rule passes the test above, and a rule that proposes
// nothing is indistinguishable from one that correctly refuses one collision.
func TestReactComponentRequirePropertiesTypeSuffixStillFixesWhenTheNameIsFree(t *testing.T) {
	t.Parallel()

	source := "export interface MenuItemInterface { a: string }\n" +
		"export function MenuItem(properties: MenuItemInterface) { return null; }\n"

	result := rule_testing.Run(t, ReactComponentRequirePropertiesTypeSuffix, typeSuffixFile, source)
	if len(result.Diagnostics) != 2 {
		t.Fatalf("wanted two findings, got %d", len(result.Diagnostics))
	}
	for index, diagnostic := range result.Diagnostics {
		if len(diagnostic.Fixes) != 1 {
			t.Fatalf("finding %d carries %d fixes, wanted one", index, len(diagnostic.Fixes))
		}
	}
}
