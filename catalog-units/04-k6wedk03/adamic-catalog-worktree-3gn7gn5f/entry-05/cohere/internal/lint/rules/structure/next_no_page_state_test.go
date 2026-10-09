package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const pageStateFile = "/repository/app/account/page.tsx"

const pageStateDeclarations = "import React from 'react';\n" +
	"declare function useState(initial?: unknown): [unknown, (value: unknown) => void];\n" +
	"declare function useReducer(reducer: unknown, initial?: unknown): [unknown, unknown];\n"

func TestNextNoPageStateFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{
			"React.useState in a page",
			"export default function AccountPageRoute() {\n    const state = React.useState(1);\n    return state;\n}\n",
		},
		{
			"React.useReducer in a page",
			"export default function AccountPageRoute() {\n" +
				"    const state = React.useReducer(() => 1, 1);\n    return state;\n}\n",
		},
		// The bare form is checked too. This codebase forbids the import that produces it, but a
		// page violating that rule would otherwise slip past this one as well, and two rules
		// failing together is when a page most needs the finding.
		{
			"a bare useState in a page",
			"export default function AccountPageRoute() {\n    const state = useState(1);\n    return state;\n}\n",
		},
		{
			"a bare useReducer in a page",
			"export default function AccountPageRoute() {\n" +
				"    const state = useReducer(() => 1, 1);\n    return state;\n}\n",
		},
		// A nested component in the page file has the same problem: the whole module is remounted.
		{
			"state in a component declared inside the page file",
			"function Panel() {\n    return React.useState(1);\n}\n" +
				"export default function AccountPageRoute() {\n    return Panel();\n}\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NextNoPageState, pageStateFile,
				pageStateDeclarations+testCase.sourceText), "pageStateRemounts")
		})
	}
}

func TestNextNoPageStateStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The shape the rule asks for: a page that holds no state of its own.
		{
			"a page with no state",
			pageStateFile,
			"export default function AccountPageRoute() {\n    return null;\n}\n",
		},
		// Hooks that do not hold state across a remount in the way this rule is about.
		{
			"other React hooks in a page",
			pageStateFile,
			"export default function AccountPageRoute() {\n" +
				"    React.useEffect(() => {}, []);\n    return React.useMemo(() => 1, []);\n}\n",
		},
		// useRef is deliberately absent from the set: a ref is a mutable box rather than state,
		// and flagging it would report every measured element and timer handle on the page.
		{
			"useRef in a page",
			pageStateFile,
			"export default function AccountPageRoute() {\n    const box = React.useRef(null);\n    return box;\n}\n",
		},
		// The file gate. A layout is not remounted, which is where the rule tells people to put
		// their state, so it must not be flagged for holding it.
		{
			"state in a layout",
			"/repository/app/account/layout.tsx",
			"export default function AccountLayout() {\n    const state = React.useState(1);\n    return state;\n}\n",
		},
		{
			"state in an ordinary component",
			"/repository/source/components/Panel.tsx",
			"export function Panel() {\n    const state = React.useState(1);\n    return state;\n}\n",
		},
		{
			"state in a page.ts rather than page.tsx",
			"/repository/app/account/page.ts",
			"export function run() {\n    return React.useState(1);\n}\n",
		},
		// The receiver test. Somebody's own accessor named useState has nothing to do with the
		// App Router.
		{
			"useState on a receiver that is not React",
			pageStateFile,
			"declare const Store: { useState(initial: unknown): unknown };\n" +
				"export default function AccountPageRoute() {\n    return Store.useState(1);\n}\n",
		},
		// A computed access is not matched, following the original. Nobody writes a hook call this
		// way, and matching it would mean deciding what a computed key resolves to. Pinned so the
		// exemption is a decision rather than an accident.
		{
			"a computed React member call",
			pageStateFile,
			"export default function AccountPageRoute() {\n    return React['useState'](1);\n}\n",
		},
		// A reference that is never called does not create state.
		{
			"a reference to useState that is not called",
			pageStateFile,
			"export default function AccountPageRoute() {\n    return React.useState;\n}\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NextNoPageState, testCase.fileName,
				pageStateDeclarations+testCase.sourceText))
		})
	}
}
