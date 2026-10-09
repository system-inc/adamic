package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const requireNamedExportFile = "/repository/source/components/Button.tsx"

func TestReactComponentRequireNamedExportFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		{
			"a component exported only as default",
			requireNamedExportFile,
			"function Button() {\n    return <div />;\n}\n\nexport default Button;\n",
		},
		{
			"a component declared and never exported",
			requireNamedExportFile,
			"function Button() {\n    return <div />;\n}\n\nconst unused = Button;\n",
		},
		{
			"an export default function declaration",
			requireNamedExportFile,
			"export default function Button() {\n    return <div />;\n}\n",
		},
		{
			"an arrow component exported only as default",
			requireNamedExportFile,
			"const Button = () => {\n    return <div />;\n};\n\nexport default Button;\n",
		},
		{
			"a forwardRef component with no named export",
			requireNamedExportFile,
			"const Button = React.forwardRef(function (properties, reference) {\n    return <div ref={reference} />;\n});\n\nexport default Button;\n",
		},
		// `export { Thing as default }` is a default export wearing named-export syntax. A guard
		// that matched any export specifier would read this as satisfying the rule.
		{
			"a component exported as default through a specifier",
			requireNamedExportFile,
			"function Button() {\n    return <div />;\n}\n\nexport { Button as default };\n",
		},
		// An alias to `default` is the shape that exposed a real bug: the specifier's Name() is
		// "default" and its PropertyName is the component, so matching the component against Name()
		// leaves it unmarked for the wrong reason and makes the `default` guard unreachable. This
		// case and the aliased clean case below are the pair that pins reading both names.
		{
			"a component aliased to default in a longer list",
			requireNamedExportFile,
			"function Button() {\n    return <div />;\n}\nconst other = 1;\n\nexport { Button as default, other };\n",
		},
		// Detected by its parameter name rather than by JSX. This is the arm that decides first and
		// the one that reaches shapes a JSX search cannot: the body here returns no JSX at all.
		{
			"a component detected by its properties parameter",
			requireNamedExportFile,
			"function Button(properties: { label: string }) {\n    return properties.label;\n}\n\nexport default Button;\n",
		},
		// The `props` spelling of the same arm.
		{
			"a component detected by its props parameter",
			requireNamedExportFile,
			"function Button(props: { label: string }) {\n    return props.label;\n}\n\nexport default Button;\n",
		},
		// A switch-dispatch component: JSX only inside switch arms, and a `properties` parameter.
		// This is the exact shape that reported a false positive on the live tree before the
		// predicate was corrected, so it is the fixture that pins the correction.
		{
			"a switch-dispatch component with no named export",
			requireNamedExportFile,
			"function Button(properties: { kind: string }) {\n    switch(properties.kind) {\n        case 'a': {\n            return <div />;\n        }\n    }\n    return null;\n}\n\nexport default Button;\n",
		},
		// A ternary return, which the original accepts when either branch is JSX.
		{
			"a component returning a jsx ternary",
			requireNamedExportFile,
			"function Button(flag: boolean) {\n    return flag ? <div /> : null;\n}\n\nexport default Button;\n",
		},
		// A .jsx file, since the path gate accepts both extensions and a suffix check written for
		// only one of them passes every .tsx fixture above.
		{
			"a jsx file",
			"/repository/source/components/Button.jsx",
			"function Button() {\n    return <div />;\n}\n\nexport default Button;\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.Run(t, ReactComponentRequireNamedExport, testCase.fileName, testCase.sourceText),
				"componentRequiresNamedExport")
		})
	}
}

func TestReactComponentRequireNamedExportStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The dominant shape in the tree: 1,116 files look like this and every one of them must
		// stay silent. If this rule is wrong in the eager direction, it is wrong here.
		{
			"a component exported on its declaration",
			requireNamedExportFile,
			"export function Button() {\n    return <div />;\n}\n",
		},
		{
			"an arrow component exported on its declaration",
			requireNamedExportFile,
			"export const Button = () => {\n    return <div />;\n};\n",
		},
		{
			"a component exported by a separate list",
			requireNamedExportFile,
			"function Button() {\n    return <div />;\n}\n\nexport { Button };\n",
		},
		{
			"a component exported both by name and as default",
			requireNamedExportFile,
			"export function Button() {\n    return <div />;\n}\n\nexport default Button;\n",
		},
		{
			"a forwardRef component exported on its declaration",
			requireNamedExportFile,
			"export const Button = React.forwardRef(function (properties, reference) {\n    return <div ref={reference} />;\n});\n",
		},
		// The path gate. Each of these declares an unexported component and would report if the
		// gate were wrong, so they sit exactly where the decision is made rather than passing
		// because there is nothing to find.
		{
			"a page file",
			"/repository/app/things/page.tsx",
			"function Page() {\n    return <div />;\n}\n\nexport default Page;\n",
		},
		{
			"a layout file",
			"/repository/app/things/layout.tsx",
			"function Layout() {\n    return <div />;\n}\n\nexport default Layout;\n",
		},
		{
			"an error file",
			"/repository/app/things/error.tsx",
			"function ErrorBoundary() {\n    return <div />;\n}\n\nexport default ErrorBoundary;\n",
		},
		{
			"a not-found file",
			"/repository/app/things/not-found.tsx",
			"function NotFound() {\n    return <div />;\n}\n\nexport default NotFound;\n",
		},
		{
			"a loading file",
			"/repository/app/things/loading.tsx",
			"function Loading() {\n    return <div />;\n}\n\nexport default Loading;\n",
		},
		{
			"a route file",
			"/repository/app/api/things/route.ts",
			"function Handler() {\n    return <div />;\n}\n\nexport default Handler;\n",
		},
		// A .ts file is not a React file. The crash guard runs every rule against .ts shapes, so a
		// rule gated on the wrong predicate would be silent there for the wrong reason.
		// A `.ts` file, which is not a React file. The component here is detected by its hook call
		// rather than by JSX, deliberately: JSX does not parse in a `.ts` file at all, so a fixture
		// relying on JSX would be silent whether or not the gate exists and would pin nothing.
		// A hook call parses fine in `.ts`, so this file reports the moment the gate is removed.
		{
			"a plain typescript file declaring a hook-using component",
			"/repository/source/Thing.ts",
			"function Button() {\n    const [open, setOpen] = useState(false);\n    return open;\n}\n\nexport default Button;\n",
		},
		// Files with nothing this rule considers a component. Each contains a default export and a
		// capitalized declaration, so a rule that keyed on either alone would report.
		{
			"a capitalized constant that is not a component",
			requireNamedExportFile,
			"const Colors = { primary: 'red' };\n\nexport default Colors;\n",
		},
		{
			"a lowercase function returning jsx",
			requireNamedExportFile,
			"function button() {\n    return <div />;\n}\n\nexport default button;\n",
		},
		{
			"a capitalized function with no jsx and no properties parameter",
			requireNamedExportFile,
			"function Helper() {\n    return 1 + 1;\n}\n\nexport default Helper;\n",
		},
		// A hook call is deliberately not a component signal here. The original's analysis uses a
		// predicate that reads parameter names and returned JSX and never looks at hook calls, so a
		// capitalized function calling one is not a component to the gate. Reaching for the
		// JSX-and-hook search instead would report this file, which the gate does not.
		{
			"a capitalized function calling a hook but returning no jsx",
			requireNamedExportFile,
			"function Helper() {\n    const [open, setOpen] = useState(false);\n    return open;\n}\n\nexport default Helper;\n",
		},
		// JSX returned from inside an if rather than at the top level of the body. The original
		// iterates only the body's own statements, so a nested return is not reached and this is
		// not a component to the gate. Widening the walk would find components it does not have.
		{
			"jsx returned from inside an if, with no properties parameter",
			requireNamedExportFile,
			"function Helper(flag: boolean) {\n    if(flag) {\n        return <div />;\n    }\n    return null;\n}\n\nexport default Helper;\n",
		},
		{
			"a file with no declarations at all",
			requireNamedExportFile,
			"export default 42;\n",
		},
		// An alias to a non-default name is still a named export: consumers have a name to import.
		// Matching only the exported name would leave the component unmarked and report this file.
		{
			"a component exported under an alias",
			requireNamedExportFile,
			"function Button() {\n    return <div />;\n}\n\nexport { Button as PrimaryButton };\n",
		},
		// A re-export names nothing this file declares, so it cannot satisfy the rule and must not
		// be read as if it did. This file declares no component, so it is silent for that reason.
		{
			"a re-export index file",
			"/repository/source/components/index.tsx",
			"export { Button } from './Button.tsx';\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, ReactComponentRequireNamedExport, testCase.fileName, testCase.sourceText))
		})
	}
}
