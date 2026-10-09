package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const separateExportFile = "/repository/source/components/Button.tsx"

func TestReactComponentNoSeparateNamedExportFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{
			"a function component exported at the bottom",
			"function Button() {\n    return <button />;\n}\nexport { Button };\n",
		},
		{
			"an arrow component exported at the bottom",
			"const Button = () => <button />;\nexport { Button };\n",
		},
		// Reported once per statement rather than per specifier, since deleting the statement is
		// one edit however many names it holds.
		{
			"two components in one export list",
			"function Button() {\n    return <button />;\n}\nfunction Panel() {\n    return <div />;\n}\n" +
				"export { Button, Panel };\n",
		},
		{
			"a component among other exported names",
			"function Button() {\n    return <button />;\n}\nconst value = 1;\nexport { value, Button };\n",
		},
		// A component detected by its hook calls rather than by JSX.
		{
			"a component that calls hooks and returns nothing",
			"import React from 'react';\nfunction Button() {\n    React.useEffect(() => {}, []);\n    return null;\n}\n" +
				"export { Button };\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, ReactComponentNoSeparateNamedExport,
				separateExportFile, testCase.sourceText), "noSeparateNamedExport")
		})
	}
}

func TestReactComponentNoSeparateNamedExportStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The shape the rule asks for.
		{
			"an inline export",
			separateExportFile,
			"export function Button() {\n    return <button />;\n}\n",
		},
		// Every case below declares a component, deliberately.
		//
		// The rule returns no listeners at all when a file declares none, so a clean fixture
		// without one is silenced by that early return before reaching the check it was written
		// for. Six mutants survived the first sweep for exactly this reason: the guard subsumed
		// every branch the fixtures meant to exercise, and from a green suite that is
		// indistinguishable from the branches working.
		//
		// So each of these carries an unrelated component that is correctly exported inline,
		// which satisfies the guard and lets the case reach the branch it is actually about.

		// A re-export has somewhere else to be, so there is nothing to move the export onto. This
		// is the shape an index file uses and it is the one the tree actually contains.
		// The re-exported name must match a declared component for the exemption to be reached at
		// all. With an unrelated name the declared-component check silences it first, which is why
		// the second case here shadows the declaration's own name.
		{
			"a re-export from another module",
			separateExportFile,
			"export function Panel() {\n    return <div />;\n}\nexport { Button } from './Button';\n",
		},
		{
			"a re-export naming a component this file also declares",
			separateExportFile,
			"function Button() {\n    return <button />;\n}\nexport { Button } from './Other';\n",
		},
		{
			"several re-exports",
			separateExportFile,
			"export function Panel() {\n    return <div />;\n}\n" +
				"export { PaginationButton } from './PaginationButton';\n" +
				"export { PaginationItem } from './PaginationItem';\n",
		},
		// Only components. Export style for types and constants is a different question this rule
		// has no opinion on.
		{
			"a constant exported at the bottom",
			separateExportFile,
			"export function Panel() {\n    return <div />;\n}\nconst value = 1;\nexport { value };\n",
		},
		// A capitalized function that renders nothing is not a component, so exporting it from a
		// list is not this rule's business. Needs the same name in both places to reach the render
		// check rather than being stopped by the name check.
		{
			"a capitalized function that renders nothing",
			separateExportFile,
			"export function Panel() {\n    return <div />;\n}\nfunction Build() {\n    return 1;\n}\nexport { Build };\n",
		},
		// A lowercase name is not a component whatever it renders, in both the function and the
		// variable branch. The variable one is the case that measures the name check inside
		// componentNamesDeclaredIn: the function branch has its own.
		{
			"a lowercase function returning JSX",
			separateExportFile,
			"export function Panel() {\n    return <div />;\n}\nfunction build() {\n    return <div />;\n}\nexport { build };\n",
		},
		{
			"a lowercase arrow returning JSX",
			separateExportFile,
			"export function Panel() {\n    return <div />;\n}\nconst build = () => <div />;\nexport { build };\n",
		},
		{
			"a capitalized constant that is not a component",
			separateExportFile,
			"export function Panel() {\n    return <div />;\n}\nconst Colors = 'red';\nexport { Colors };\n",
		},
		{
			"a type exported at the bottom",
			separateExportFile,
			"export function Panel() {\n    return <div />;\n}\n" +
				"type ButtonProperties = { label: string };\nexport type { ButtonProperties };\n",
		},
		// A name the file does not declare cannot have its export moved onto a declaration here.
		{
			"a name declared elsewhere",
			separateExportFile,
			"export function Panel() {\n    return <div />;\n}\n" +
				"import { Button } from './Button';\nexport { Button };\n",
		},
		// A star re-export names nothing, so there is no specifier to match.
		{
			"a star re-export",
			separateExportFile,
			"export function Panel() {\n    return <div />;\n}\nexport * from './Other';\n",
		},
		// Outside a React file the rule declines before looking.
		{
			"a component-shaped export in a ts file",
			"/repository/source/api/Thing.ts",
			"import React from 'react';\nfunction Button() {\n    React.useEffect(() => {}, []);\n    return null;\n}\n" +
				"export { Button };\n",
		},
		{
			"a file with no export list",
			separateExportFile,
			"export function Button() {\n    return <button />;\n}\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, ReactComponentNoSeparateNamedExport,
				testCase.fileName, testCase.sourceText))
		})
	}
}

// The fix deletes the export statement and nothing else.
//
// Asserted against the resulting source rather than against the message id, using the assertion
// `@system_cohere` added after finding that an id-only fixture let a rewrite corrupt every import
// it touched while the suite stayed green. A fix is the one part of a rule that changes source, so
// the only honest test of it is what the source becomes.
//
// The blank line the deletion leaves behind is real and expected. `RemoveNode` takes the node's
// range, which ends at the semicolon rather than at the newline after it, so the line the statement
// occupied becomes empty. The formatter collapses it on the next pass.
//
// Worth writing down because I expected the clean text and this assertion is what told me
// otherwise, on the first fixable rule I have written since it existed. An id-only fixture would
// have agreed with me.
//
// This test used to assert the DEFECT. Its expected outputs were the source with the list deleted
// and no `export` anywhere, so every component the fixer touched was unexported and the suite
// stayed green, because the fixture recorded what the fixer did rather than what the file needed.
// The outputs below carry the export onto each declaration, which is the only rewrite that leaves
// every importer resolving.
func TestReactComponentNoSeparateNamedExportFixMovesTheExportOntoTheDeclaration(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantFixed  string
	}{
		{
			"a function component",
			"function Button() {\n    return <button />;\n}\nexport { Button };\n",
			"export function Button() {\n    return <button />;\n}\n\n",
		},
		// A list holding several names goes as one statement and one edit, which is what makes the
		// single report per statement the right shape.
		{
			"two components in one list",
			"function Button() {\n    return <button />;\n}\nfunction Panel() {\n    return <div />;\n}\n" +
				"export { Button, Panel };\n",
			"export function Button() {\n    return <button />;\n}\nexport function Panel() {\n    return <div />;\n}\n\n",
		},
		// The keyword lands on the token, below the JSDoc that documents the declaration.
		{
			"a documented component",
			"/** The button. */\nfunction Button() {\n    return <button />;\n}\nexport { Button };\n",
			"/** The button. */\nexport function Button() {\n    return <button />;\n}\n\n",
		},
		// In front of `async`, which is where `export` has to go.
		{
			"an async component",
			"async function Button() {\n    return <button />;\n}\nexport { Button };\n",
			"export async function Button() {\n    return <button />;\n}\n\n",
		},
		{
			"an arrow component",
			"const Button = () => <button />;\nexport { Button };\n",
			"export const Button = () => <button />;\n\n",
		},
		// A list written ABOVE the declaration it names, which hoisting makes legal.
		{
			"a list before its declaration",
			"export { Button };\nfunction Button() {\n    return <button />;\n}\n",
			"\nexport function Button() {\n    return <button />;\n}\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ReactComponentNoSeparateNamedExport, separateExportFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "noSeparateNamedExport")
			if len(result.Diagnostics[0].Fixes) != 1 {
				// One edit, not a deletion plus insertions. The engine applies a diagnostic's fixes
				// independently, and a deletion landing without its insertion is the original
				// defect.
				t.Fatalf("expected the repair as one edit, got %d", len(result.Diagnostics[0].Fixes))
			}
			rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
		})
	}
}

// TestReactComponentNoSeparateNamedExportDeclinesWhereTheExportCannotMove pins every decline.
//
// Each is reported, because the list is still the shape the rule exists to flag, and none carries a
// repair, because moving the export would change what the module offers. The first row is the one
// the deleting fixer got most wrong: it unexported `value`, a name the rule has no opinion on.
func TestReactComponentNoSeparateNamedExportDeclinesWhereTheExportCannotMove(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{
			"a component listed beside a name that is not one",
			"function Button() {\n    return <button />;\n}\nconst value = 1;\nexport { value, Button };\n",
		},
		{
			"a component listed beside an imported name",
			"import { other } from './other';\nfunction Button() {\n    return <button />;\n}\nexport { Button, other };\n",
		},
		{
			"a component declared alongside a sibling binding",
			"const Button = () => <button />, value = 1;\nexport { Button };\n",
		},
		{
			"a component declared through overloads",
			"function Button(): JSX.Element;\nfunction Button() {\n    return <button />;\n}\nexport { Button };\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ReactComponentNoSeparateNamedExport, separateExportFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "noSeparateNamedExport")
			diagnostic := result.Diagnostics[0]
			if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
				t.Fatalf("expected no repair, got %d fixes and %d suggestions",
					len(diagnostic.Fixes), len(diagnostic.Suggestions))
			}
		})
	}
}
