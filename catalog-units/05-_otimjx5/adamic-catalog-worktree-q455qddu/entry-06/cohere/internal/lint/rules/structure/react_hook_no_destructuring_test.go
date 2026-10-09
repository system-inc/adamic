package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const hookDestructuringFile = "/repository/source/hooks/useThing.ts"

const hookDestructuringDeclarations = "declare function useQuery(): { data: unknown; error: unknown };\n" +
	"declare function useState(): [unknown, (value: unknown) => void];\n" +
	"declare function subscribe(handler: () => void): void;\n"

func TestReactHookNoDestructuringFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{
			"a function declaration hook",
			"export function useThing() {\n    const { data } = useQuery();\n    return data;\n}\n",
		},
		{
			"an arrow hook assigned to a variable",
			"export const useThing = () => {\n    const { data } = useQuery();\n    return data;\n};\n",
		},
		{
			"a function expression hook assigned to a variable",
			"export const useThing = function() {\n    const { data } = useQuery();\n    return data;\n};\n",
		},
		{
			"a named function expression hook",
			"export const alias = function useThing() {\n    const { data } = useQuery();\n    return data;\n};\n",
		},
		// A hook's callbacks are part of the hook, so a destructure inside one still counts. The
		// nearest enclosing function here is an anonymous callback with no name to test, which is
		// why the walk continues past it rather than stopping.
		{
			"inside a callback inside a hook",
			"export function useThing() {\n    subscribe(function() {\n        const { data } = useQuery();\n" +
				"        void data;\n    });\n    return 1;\n}\n",
		},
		{
			"inside an arrow callback inside a hook",
			"export function useThing() {\n    subscribe(() => {\n        const { data } = useQuery();\n" +
				"        void data;\n    });\n    return 1;\n}\n",
		},
		{
			"a nested destructure",
			"export function useThing() {\n    const { data: { inner } } = useQuery() as any;\n    return inner;\n}\n",
		},
		{
			"a destructure inside a hook that is a method",
			"export const container = { useThing() {\n    const { data } = useQuery();\n    return data;\n} };\n",
		},
		// A function assigned to an object property takes that property's name. The shorthand
		// method above is a different node kind and reaches a different arm, so it does not measure
		// this one. Found by removing the property-assignment arm and watching the suite stay green.
		{
			"a function expression assigned to a hook-named property",
			"export const container = { useThing: function() {\n    const { data } = useQuery();\n    return data;\n} };\n",
		},
		{
			"an arrow assigned to a hook-named property",
			"export const container = { useThing: () => {\n    const { data } = useQuery();\n    return data;\n} };\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, ReactHookNoDestructuring, hookDestructuringFile,
				hookDestructuringDeclarations+testCase.sourceText), "noDestructuringInHook")
		})
	}
}

func TestReactHookNoDestructuringStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The shape the rule asks for.
		{
			"a stored value read by property",
			"export function useThing() {\n    const result = useQuery();\n    return result.data;\n}\n",
		},
		// Array destructuring is exempt, and this is the convention rather than an oversight: there
		// are no property names to preserve, so there is nothing to protect.
		{
			"array destructuring in a hook",
			"export function useThing() {\n    const [value, setValue] = useState();\n    return { value, setValue };\n}\n",
		},
		// The hook name gate. Identical body, a name that is not a hook.
		{
			"destructuring in a plain function",
			"export function buildThing() {\n    const { data } = useQuery();\n    return data;\n}\n",
		},
		{
			"destructuring in a component",
			"export function Thing() {\n    const { data } = useQuery();\n    return data;\n}\n",
		},
		{
			"destructuring at the top level of a file",
			"const { data } = useQuery();\nexport const value = data;\n",
		},
		// `use` followed by a lowercase letter is not a hook name, which is React's own test.
		{
			"a function named used",
			"export function used() {\n    const { data } = useQuery();\n    return data;\n}\n",
		},
		{
			"a function named use",
			"export function use() {\n    const { data } = useQuery();\n    return data;\n}\n",
		},
		// An anonymous callback outside any hook inherits no name from what encloses it.
		{
			"a callback inside a plain function",
			"export function buildThing() {\n    subscribe(() => {\n        const { data } = useQuery();\n" +
				"        void data;\n    });\n    return 1;\n}\n",
		},
		// A for-of head has no initializer, so it is not reported. That is the original's behavior
		// rather than a gap in ours: its check reads a VariableDeclarator's `init`, which ESTree
		// leaves null here. Pinned as a fixture so the exemption is a decision rather than an
		// accident, and so a future change to the initializer guard fails loudly.
		{
			"a destructuring for-of head inside a hook",
			"export function useThing() {\n    for(const { data } of []) { void data; }\n    return 1;\n}\n",
		},
		{
			"a hook with no destructuring at all",
			"export function useThing() {\n    return useQuery().data;\n}\n",
		},
		{
			"a file with no hooks",
			"export const value = 1;\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, ReactHookNoDestructuring, hookDestructuringFile,
				hookDestructuringDeclarations+testCase.sourceText))
		})
	}
}
