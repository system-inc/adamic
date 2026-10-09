package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const displayNameFile = "/repository/source/components/Field.tsx"

func TestReactComponentNoDisplayNameFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantCount  int
	}{
		{
			"a function component assigning its own name",
			"export function Field() { return <div />; }\nField.displayName = 'Field';\n",
			1,
		},
		{
			"an arrow component assigning its name",
			"export const Field = () => <div />;\nField.displayName = 'Field';\n",
			1,
		},
		// The detector's parameter arm: a capitalized function taking `properties` is a component
		// whatever its body returns.
		{
			"a component known by its properties parameter",
			"export function Field(properties: { label: string }) { return null; }\nField.displayName = 'Field';\n",
			1,
		},
		{
			"a class component",
			"import React from 'react';\nexport class Field extends React.Component { render() { return <div />; } }\n" +
				"Field.displayName = 'Field';\n",
			1,
		},
		// A function declaration hoists, so an assignment above it names the component all the same.
		// The set is gathered from the whole file for this case, not built up during the walk.
		{
			"an assignment above the function it names",
			"Field.displayName = 'Field';\nexport function Field() { return <div />; }\n",
			1,
		},
		{
			"two assignments report twice",
			"export function A() { return <div />; }\nexport function B() { return <div />; }\n" +
				"A.displayName = 'A';\nB.displayName = 'B';\n",
			2,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ReactComponentNoDisplayName, displayNameFile, testCase.sourceText)
			ids := result.MessageIds()
			if len(ids) != testCase.wantCount {
				t.Fatalf("expected %d findings, got %d: %v", testCase.wantCount, len(ids), ids)
			}
		})
	}
}

func TestReactComponentNoDisplayNameStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The shape the rule asks for: no assignment at all, because the name is already there.
		{
			"a component that does not assign displayName",
			"export function Field() { return null; }\n",
		},
		// The exemption, in each of the four wrapper spellings. A component built by one of these
		// has no name of its own, so the assignment is the only way to label it.
		{
			"a bare memo wrapper",
			"declare function memo(component: unknown): unknown;\n" +
				"export const Field = memo(function () { return null; });\nField.displayName = 'Field';\n",
		},
		{
			"a React.memo wrapper",
			"import React from 'react';\n" +
				"export const Field = React.memo(function () { return null; });\nField.displayName = 'Field';\n",
		},
		{
			"a bare forwardRef wrapper",
			"declare function forwardRef(component: unknown): unknown;\n" +
				"export const Field = forwardRef(function () { return null; });\nField.displayName = 'Field';\n",
		},
		{
			"a React.forwardRef wrapper",
			"import React from 'react';\n" +
				"export const Field = React.forwardRef(function () { return null; });\nField.displayName = 'Field';\n",
		},
		// A compound assignment is not establishing a name, and reporting it would be a claim about
		// string arithmetic rather than about naming.
		{
			"a compound assignment",
			"declare const Field: { displayName: string };\nField.displayName += 'x';\n",
		},
		// A different property entirely.
		{
			"an assignment to a property that is not displayName",
			"export function Field() { return null; }\nField.propertyTypes = {};\n",
		},
		// Reading displayName is not assigning it.
		{
			"reading displayName",
			"declare const Field: { displayName: string };\nexport const name = Field.displayName;\n",
		},
		// A domain field. These are api-phi-health's 17 former findings, all in plain services, and the
		// reason the rule asks whether the target is a component at all (#qpms5xz).
		{
			"a domain object's displayName field",
			"declare const event: { displayName: string };\ndeclare const organization: { displayName: string };\n" +
				"event.displayName = organization.displayName;\n",
		},
		// A capital letter alone does not make a component: the binding has to be one.
		{
			"a capitalized object that is not a component",
			"export const Organization = { displayName: 'Phi' };\nOrganization.displayName = 'Phi, Inc.';\n",
		},
		{
			"a lowercase function that returns JSX",
			"export function field() { return <div />; }\nfield.displayName = 'field';\n",
		},
		// A component this file does not declare is not seen, so these miss rather than guess.
		{
			"a binding initialized from an unrelated call",
			"declare function build(): unknown;\nexport const Field = build();\nField.displayName = 'Field';\n",
		},
		{
			"an assignment to a member of an object",
			"declare const components: { Field: { displayName?: string } };\ncomponents.Field.displayName = 'Field';\n",
		},
		// `Other.memo` is not React's wrapper, so its result is neither exempt nor known to be a
		// component, and the binding is as opaque as any other factory's.
		{
			"memo reached off a receiver that is not React",
			"declare const Other: { memo(component: unknown): unknown };\n" +
				"export const Field = Other.memo(function () { return <div />; });\nField.displayName = 'Field';\n",
		},
		{
			"an imported component",
			"import { Field } from './Field';\nField.displayName = 'Field';\n",
		},
		{
			"a file with no assignments",
			"export const value = 1;\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, ReactComponentNoDisplayName, displayNameFile, testCase.sourceText))
		})
	}
}

// Declaration order decides the exemption, and this pins which way.
//
// The rule collects wrapper-initialized bindings during the walk and reads that set at the
// assignment, so a binding declared after the assignment does not exempt it. Worth a test of its
// own rather than a comment, because it is the one behavior that depends on the traversal order
// rather than on any single node, and a change to how listeners are dispatched would break it
// silently.
//
// The limit is the right one: an assignment above `const Field = memo(...)` is a temporal dead
// zone error at runtime, so exempting it would excuse code that cannot run.
func TestReactComponentNoDisplayNameDependsOnDeclarationOrder(t *testing.T) {
	t.Parallel()

	declaredFirst := "declare function memo(component: unknown): unknown;\n" +
		"export const Field = memo(function () { return null; });\nField.displayName = 'Field';\n"
	rule_testing.ExpectClean(t, rule_testing.Run(t, ReactComponentNoDisplayName, displayNameFile, declaredFirst))

	assignedFirst := "declare function memo(component: unknown): unknown;\n" +
		"Field.displayName = 'Field';\nexport const Field = memo(function () { return null; });\n"
	rule_testing.ExpectFindings(t, rule_testing.Run(t, ReactComponentNoDisplayName, displayNameFile, assignedFirst),
		"noDisplayNameAssignment")
}
