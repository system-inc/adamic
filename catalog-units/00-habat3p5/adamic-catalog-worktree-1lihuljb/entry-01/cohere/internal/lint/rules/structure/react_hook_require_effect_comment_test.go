package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const effectCommentFile = "/repository/source/components/Panel.tsx"

const effectCommentDeclarations = "import React from 'react';\ndeclare const route: string;\n"

func TestReactHookRequireEffectCommentFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{
			"no comment at all",
			"export function Panel() {\n    React.useEffect(() => {}, [route]);\n    return null;\n}\n",
		},
		{
			"a comment that does not make the claim",
			"export function Panel() {\n    // sync the title\n    React.useEffect(() => {}, [route]);\n    return null;\n}\n",
		},
		// The topmost comment of the run is what is checked. Here the nearest one makes the claim
		// and the one above it does not, so the run does not start with it.
		{
			"the claim on the second line of a run",
			"export function Panel() {\n    // some preamble\n    // Effect to sync the title\n" +
				"    React.useEffect(() => {}, [route]);\n    return null;\n}\n",
		},
		// A block comment ends a run, so a claim in one does not carry across a line comment below
		// it. Without that the walk would collapse the two and read the block as the run's start.
		{
			"a claiming block above a non-claiming line comment",
			"export function Panel() {\n    /* Effect to sync the title */\n    // then do the thing\n" +
				"    React.useEffect(() => {}, [route]);\n    return null;\n}\n",
		},
		{
			"a block comment that does not make the claim",
			"export function Panel() {\n    /* sync the title */\n    React.useEffect(() => {}, [route]);\n    return null;\n}\n",
		},
		// A blank line ends the run, so a claim above the gap does not attach to this effect.
		{
			"a claim separated by a blank line",
			"export function Panel() {\n    // Effect to sync the title\n\n    // unrelated\n" +
				"    React.useEffect(() => {}, [route]);\n    return null;\n}\n",
		},
		{
			"a second effect without its own comment",
			"export function Panel() {\n    // Effect to sync the title\n    React.useEffect(() => {}, [route]);\n" +
				"    React.useEffect(() => {}, []);\n    return null;\n}\n",
		},
		// The comment lookup happens on the enclosing statement, not the call. These two shapes are
		// where that matters: the call is not the statement's first token, so its own Pos sits
		// after the comment and a lookup there finds nothing.
		//
		// Both were clean cases in my first version, asserting the rule fires. It does, but for the
		// wrong reason under a mutant that looks up on the call: it finds no comment and reports,
		// which is the same answer. The firing half cannot measure this. The silent cases below
		// are what do.
		{
			"an effect assigned to a variable",
			"export function Panel() {\n    const unused = React.useEffect(() => {}, [route]);\n    return unused;\n}\n",
		},
		{
			"an effect inside a nested block",
			"export function Panel() {\n    if(route) {\n        React.useEffect(() => {}, [route]);\n    }\n    return null;\n}\n",
		},
		{
			"a comment whose claim is not at the start",
			"export function Panel() {\n    // This is an Effect to sync the title\n" +
				"    React.useEffect(() => {}, [route]);\n    return null;\n}\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, ReactHookRequireEffectComment, effectCommentFile,
				effectCommentDeclarations+testCase.sourceText), "missingEffectComment")
		})
	}
}

func TestReactHookRequireEffectCommentStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The shape the rule asks for.
		{
			"a line comment making the claim",
			effectCommentFile,
			"export function Panel() {\n    // Effect to sync the title with the route\n" +
				"    React.useEffect(() => {}, [route]);\n    return null;\n}\n",
		},
		// A multi-line explanation, where only the first line has to make the claim. Requiring
		// every line to would forbid a second sentence.
		{
			"a multi-line run whose first line makes the claim",
			effectCommentFile,
			"export function Panel() {\n    // Effect to sync the title with the route\n" +
				"    // Runs whenever the route changes.\n" +
				"    React.useEffect(() => {}, [route]);\n    return null;\n}\n",
		},
		{
			"a block comment making the claim",
			effectCommentFile,
			"export function Panel() {\n    /* Effect to sync the title */\n" +
				"    React.useEffect(() => {}, [route]);\n    return null;\n}\n",
		},
		// The JSDoc shape, where the leading asterisks are decoration rather than content.
		{
			"a JSDoc block making the claim",
			effectCommentFile,
			"export function Panel() {\n    /**\n     * Effect to sync the title with the route\n     */\n" +
				"    React.useEffect(() => {}, [route]);\n    return null;\n}\n",
		},
		{
			"two effects each with their own comment",
			effectCommentFile,
			"export function Panel() {\n    // Effect to sync the title\n    React.useEffect(() => {}, [route]);\n" +
				"    // Effect to log the route\n    React.useEffect(() => {}, []);\n    return null;\n}\n",
		},
		// Only React.useEffect. A bare useEffect is react-import-no-destructuring's finding, and
		// reporting it here too would give two findings for one mistake.
		{
			"a bare useEffect call",
			effectCommentFile,
			"declare function useEffect(effect: () => void, dependencies: unknown[]): void;\n" +
				"export function Panel() {\n    useEffect(() => {}, [route]);\n    return null;\n}\n",
		},
		{
			"a different React hook",
			effectCommentFile,
			"export function Panel() {\n    React.useMemo(() => 1, [route]);\n    return null;\n}\n",
		},
		{
			"useEffect on a receiver that is not React",
			effectCommentFile,
			"declare const Other: { useEffect(effect: () => void): void };\n" +
				"export function Panel() {\n    Other.useEffect(() => {});\n    return null;\n}\n",
		},
		// Outside a React file an effect is not a component's effect, and the rule declines the
		// file before scanning any comments.
		{
			"an effect in a plain ts file",
			"/repository/source/api/Thing.ts",
			"export function run() {\n    React.useEffect(() => {}, [route]);\n}\n",
		},
		// A commented effect that is not the statement's first token. Looking the comment up on
		// the call rather than the statement finds nothing here and reports, so this is the case
		// that measures the walk. Found by mutating the lookup and watching the suite stay green.
		{
			"a commented effect assigned to a variable",
			effectCommentFile,
			"export function Panel() {\n    // Effect to sync the title\n" +
				"    const unused = React.useEffect(() => {}, [route]);\n    return unused;\n}\n",
		},
		{
			"a commented effect inside a nested block",
			effectCommentFile,
			"export function Panel() {\n    if(route) {\n        // Effect to sync the title\n" +
				"        React.useEffect(() => {}, [route]);\n    }\n    return null;\n}\n",
		},
		{
			"a file with no effects",
			effectCommentFile,
			"export function Panel() { return null; }\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, ReactHookRequireEffectComment, testCase.fileName,
				effectCommentDeclarations+testCase.sourceText))
		})
	}
}
