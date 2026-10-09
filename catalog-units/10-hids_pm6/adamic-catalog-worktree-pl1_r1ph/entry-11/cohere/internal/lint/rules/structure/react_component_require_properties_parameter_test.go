package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const propertiesParameterFile = "/repository/source/components/Button.tsx"

func TestReactComponentRequirePropertiesParameterFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{
			"a function component",
			"export function Button(props: { label: string }) {\n    return <button>{props.label}</button>;\n}\n",
		},
		{
			"an arrow component",
			"export const Button = (props: { label: string }) => <button>{props.label}</button>;\n",
		},
		{
			"a function expression component",
			"export const Button = function(props: { label: string }) {\n" +
				"    return <button>{props.label}</button>;\n};\n",
		},
		// Detected by hook calls rather than JSX, which is the other half of what makes a
		// component.
		{
			"a component that calls hooks and renders nothing",
			"import React from 'react';\n" +
				"export function Button(props: { label: string }) {\n" +
				"    React.useEffect(() => {}, []);\n    return props.label;\n}\n",
		},
		// A lowercase name is not judged by this rule at all, since it never checks the name: what
		// makes something a component here is that it renders.
		{
			"a lowercase function that renders",
			"export function renderButton(props: { label: string }) {\n" +
				"    return <button>{props.label}</button>;\n}\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, ReactComponentRequirePropertiesParameter,
				propertiesParameterFile, testCase.sourceText), "usePropertiesNotProps")
		})
	}
}

func TestReactComponentRequirePropertiesParameterStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The shape the rule asks for.
		{
			"a component named properties",
			propertiesParameterFile,
			"export function Button(properties: { label: string }) {\n" +
				"    return <button>{properties.label}</button>;\n}\n",
		},
		// A destructured parameter has no single name to judge, and is another rule's business.
		{
			"a destructured parameter",
			propertiesParameterFile,
			"export function Button({ label, ...buttonProperties }: any) {\n" +
				"    return <button {...buttonProperties}>{label}</button>;\n}\n",
		},
		// Any other name is the author's choice. The rule is about the abbreviation specifically,
		// not about enforcing one name.
		{
			"a differently named parameter",
			propertiesParameterFile,
			"export function Button(input: { label: string }) {\n    return <button>{input.label}</button>;\n}\n",
		},
		// A name holding props as a substring is not the abbreviation.
		{
			"a parameter named propsBag",
			propertiesParameterFile,
			"export function Button(propsBag: { label: string }) {\n" +
				"    return <button>{propsBag.label}</button>;\n}\n",
		},
		// The render gate. A function that renders nothing is not a component whatever its
		// parameter is called.
		{
			"a plain function taking props",
			propertiesParameterFile,
			"export function build(props: { label: string }) {\n    return props.label.length;\n}\n",
		},
		// Only the first parameter. A second one named props is not the properties object.
		{
			"props as a second parameter",
			propertiesParameterFile,
			"export function Button(properties: { label: string }, props: unknown) {\n" +
				"    return <button>{properties.label}</button>;\n}\n",
		},
		{
			"a component with no parameters",
			propertiesParameterFile,
			"export function Button() {\n    return <button />;\n}\n",
		},
		// Outside a React file the rule declines before looking at anything.
		//
		// The body has to render for this to measure the file gate. My first version returned a
		// plain string, which the render check stops first, so the case passed whether or not the
		// gate existed. Same shared-guard shape that silenced six fixtures on the previous rule.
		{
			"a component-shaped function in a ts file",
			"/repository/source/api/Thing.ts",
			"import React from 'react';\n" +
				"export function Button(props: { label: string }) {\n" +
				"    React.useEffect(() => {}, []);\n    return props.label;\n}\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, ReactComponentRequirePropertiesParameter,
				testCase.fileName, testCase.sourceText))
		})
	}
}

// The suggestion renames the declaration and stops there.
//
// Asserted against the resulting source, and the result deliberately does not compile: the uses
// inside the body are left alone, because a rename that guesses which `props` in a body are the
// parameter fails quietly and in the wrong place while one that stops at the declaration fails
// immediately and says exactly where.
//
// It is a suggestion rather than a fix precisely because of that. This shipped as a fix, which the
// engine applies unattended, so the whole argument for an intentional visible break rested on a
// human reading a diff nothing showed them. A fix preserves what the code means; this changes it by
// design, so the author has to be the one who accepts it.
//
// Asserted through the suggestion rather than through ExpectFixedSource, which reads only the
// automatic fixes and correctly refuses to check a rule that proposes none.
func TestReactComponentRequirePropertiesParameterSuggestsRenamingTheDeclaration(t *testing.T) {
	t.Parallel()

	const source = "export function Button(props: { label: string }) {\n" +
		"    return <button>{props.label}</button>;\n}\n"
	const wanted = "export function Button(properties: { label: string }) {\n" +
		"    return <button>{props.label}</button>;\n}\n"

	result := rule_testing.Run(t, ReactComponentRequirePropertiesParameter, propertiesParameterFile, source)
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}
	if len(result.Diagnostics[0].Fixes) != 0 {
		t.Fatalf("wanted no automatic fix, got %d", len(result.Diagnostics[0].Fixes))
	}

	suggestions := result.Diagnostics[0].Suggestions
	if len(suggestions) != 1 || len(suggestions[0].Fixes) != 1 {
		t.Fatalf("wanted one suggestion carrying one fix, got %d suggestions", len(suggestions))
	}

	fix := suggestions[0].Fixes[0]
	rewritten := source[:fix.Range.Pos()] + fix.Text + source[fix.Range.End():]
	if rewritten != wanted {
		t.Fatalf("applying the suggestion gave %q, wanted %q", rewritten, wanted)
	}
}
