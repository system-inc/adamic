package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const componentDestructuringFile = "/repository/source/components/Button.tsx"

const componentDestructuringDeclarations = "import React from 'react';\n" +
	"declare function useImperativeHandle(reference: unknown, create: () => unknown): void;\n"

func TestReactComponentNoDestructuringFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantId     string
	}{
		{
			"a parameter destructure with no rest",
			"export function Button({ label }: { label: string }) {\n    return <button>{label}</button>;\n}\n",
			"noDestructuring",
		},
		{
			"an arrow component with no rest",
			"export const Button = ({ label }: { label: string }) => <button>{label}</button>;\n",
			"noDestructuring",
		},
		// Four different repairs, which is why they are four ids rather than one.
		{
			"a rest not ending in Properties",
			"export function Button({ label, ...rest }: any) {\n    return <button {...rest}>{label}</button>;\n}\n",
			"requirePropertiesSuffix",
		},
		// The boundary the check actually decides on. A name holding Properties as a substring
		// rather than a suffix is still wrong, and a `Contains` implementation passes every other
		// fixture in this file. Found by mutating the suffix test and watching the suite stay green.
		{
			"a rest with Properties in the middle of its name",
			"export function Button({ label, ...buttonPropertiesList }: any) {\n" +
				"    return <button {...buttonPropertiesList}>{label}</button>;\n}\n",
			"requirePropertiesSuffix",
		},
		{
			"a rest named for being a rest",
			"export function Button({ label, ...restProperties }: any) {\n" +
				"    return <button {...restProperties}>{label}</button>;\n}\n",
			"semanticSpreadName",
		},
		{
			"a rest named remainingProperties",
			"export function Button({ label, ...remainingProperties }: any) {\n" +
				"    return <button {...remainingProperties}>{label}</button>;\n}\n",
			"semanticSpreadName",
		},
		{
			"a rest that is never used",
			"export function Button({ label, ...buttonProperties }: any) {\n    return <button>{label}</button>;\n}\n",
			"spreadMustBeUsed",
		},
		// The same mistake written one line later, inside the body.
		{
			"destructuring from properties in the body",
			"export function Button(properties: { label: string }) {\n" +
				"    const { label } = properties;\n    return <button>{label}</button>;\n}\n",
			"noDestructuringFromPropsSource",
		},
		{
			"destructuring from props in the body",
			"export function Button(props: { label: string }) {\n" +
				"    const { label } = props;\n    return <button>{label}</button>;\n}\n",
			"noDestructuringFromPropsSource",
		},
		// A body destructure that does gather a rest is judged by the same rules as a parameter.
		{
			"a body destructure whose rest is badly named",
			"export function Button(properties: any) {\n    const { label, ...rest } = properties;\n" +
				"    return <button {...rest}>{label}</button>;\n}\n",
			"requirePropertiesSuffix",
		},
		// A ref destructure without an imperative handle gets no exemption.
		{
			"a ref destructure with no imperative handle",
			"export function Button({ ref, label }: any) {\n    return <button ref={ref}>{label}</button>;\n}\n",
			"noDestructuring",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, ReactComponentNoDestructuring, componentDestructuringFile,
				componentDestructuringDeclarations+testCase.sourceText), testCase.wantId)
		})
	}
}

func TestReactComponentNoDestructuringStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The two shapes the rule asks for.
		{
			"properties read directly",
			componentDestructuringFile,
			"export function Button(properties: { label: string }) {\n    return <button>{properties.label}</button>;\n}\n",
		},
		{
			"a well named rest that is spread",
			componentDestructuringFile,
			"export function Button({ label, ...buttonProperties }: any) {\n" +
				"    return <button {...buttonProperties}>{label}</button>;\n}\n",
		},
		// The imperative-handle exemption. A component taking ref out and exposing a handle is not
		// the leaf-DOM passthrough the rest convention was designed for.
		{
			"a ref destructure with an imperative handle",
			componentDestructuringFile,
			"export function Button({ ref, label }: any) {\n    useImperativeHandle(ref, () => ({}));\n" +
				"    return <button>{label}</button>;\n}\n",
		},
		{
			"a renamed ref property with an imperative handle",
			componentDestructuringFile,
			"export function Button({ ref: reference, label }: any) {\n" +
				"    useImperativeHandle(reference, () => ({}));\n    return <button>{label}</button>;\n}\n",
		},
		// The component gate. A lowercase name is not a component, whatever it renders.
		{
			"a lowercase function destructuring",
			componentDestructuringFile,
			"export function build({ label }: { label: string }) {\n    return <span>{label}</span>;\n}\n",
		},
		// A capitalized function that renders nothing is not a component, and holding it to a
		// component's convention would flag ordinary factories.
		{
			"a capitalized function with no JSX or hooks",
			componentDestructuringFile,
			"export function Build({ label }: { label: string }) {\n    return label.length;\n}\n",
		},
		// A body destructure from something that is not the properties object.
		{
			"destructuring from an unrelated object",
			componentDestructuringFile,
			"export function Button(properties: any) {\n    const source = properties.data;\n" +
				"    const { label } = source;\n    return <button>{label}</button>;\n}\n",
		},
		{
			"a body destructure outside any component",
			componentDestructuringFile,
			"declare const properties: any;\nconst { label } = properties;\nexport const value = label;\n",
		},
		// Outside a React file the rule declines before looking at anything.
		{
			"a component-shaped function in a ts file",
			"/repository/source/api/Thing.ts",
			"export function Button({ label }: { label: string }) {\n    return label;\n}\n",
		},
		{
			"a component with no destructuring at all",
			componentDestructuringFile,
			"export function Button() {\n    return <button />;\n}\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, ReactComponentNoDestructuring, testCase.fileName,
				componentDestructuringDeclarations+testCase.sourceText))
		})
	}
}
