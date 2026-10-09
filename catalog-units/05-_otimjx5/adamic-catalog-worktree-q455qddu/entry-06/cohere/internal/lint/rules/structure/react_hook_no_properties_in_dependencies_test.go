package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const propertiesInDependenciesFile = "/repository/source/components/Thing.tsx"

func TestReactHookNoPropertiesInDependenciesFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		count      int
	}{
		// One case per hook, because the dependency argument is read by position and a rule that
		// hardcoded one index would pass every fixture written for that hook alone.
		{
			"useEffect",
			"function Thing(properties: { id: string }) {\n    React.useEffect(() => {\n        run(properties.id);\n    }, [properties]);\n    return null;\n}\n",
			1,
		},
		{
			"useCallback",
			"function Thing(properties: { id: string }) {\n    const handler = React.useCallback(() => {\n        run(properties.id);\n    }, [properties]);\n    return handler;\n}\n",
			1,
		},
		{
			"useMemo",
			"function Thing(properties: { id: string }) {\n    const value = React.useMemo(() => properties.id, [properties]);\n    return value;\n}\n",
			1,
		},
		{
			"useLayoutEffect",
			"function Thing(properties: { id: string }) {\n    React.useLayoutEffect(() => {\n        run(properties.id);\n    }, [properties]);\n    return null;\n}\n",
			1,
		},
		// The third-argument hook. Its array is in a different position, so a rule reading the
		// second argument finds the factory function there, sees no array, and stays silent. Every
		// other fixture above passes in that broken state.
		{
			"useImperativeHandle",
			"function Thing(properties: { id: string }) {\n    React.useImperativeHandle(properties.reference, () => ({}), [properties]);\n    return null;\n}\n",
			1,
		},
		// Position within the array must not matter.
		{
			"properties alongside other dependencies",
			"function Thing(properties: { id: string }) {\n    React.useEffect(() => {\n        run(properties.id);\n    }, [other, properties, another]);\n    return null;\n}\n",
			1,
		},
		// Two occurrences report twice, matching the original, which reports per element.
		{
			"properties listed twice",
			"function Thing(properties: { id: string }) {\n    React.useEffect(() => {\n        run(properties.id);\n    }, [properties, properties]);\n    return null;\n}\n",
			2,
		},
		// The enclosing-function walk. The nearest function is the anonymous arrow passed to the
		// hook, which has no name, so a walk that stopped at the nearest would find nothing here.
		{
			"inside a nested callback within a component",
			"function Thing(properties: { id: string }) {\n    return items.map(() => {\n        React.useEffect(() => {\n            run(properties.id);\n        }, [properties]);\n        return null;\n    });\n}\n",
			1,
		},
		// A custom hook nested inside a component, which is the one shape that separates "check
		// every enclosing function" from "check the nearest one".
		//
		// The chain here is [useInner, Thing]: the nearest named function is the hook, so a walk
		// that decided on the first name it found would answer "hook, not a component" and stay
		// silent. The original loops until it finds a component, so this reports.
		//
		// The mirror case (a component inside a hook) does not distinguish them, which is worth
		// recording because it was written first and looked like it did: there the chain is
		// [Inner, useOuter] and the component is already nearest, so both walks agree. Mutation
		// caught that the fixture measured nothing, and the direction of nesting is the whole
		// difference.
		{
			"a custom hook nested inside a component",
			"function Thing(properties: { id: string }) {\n    function useInner() {\n        React.useEffect(() => {\n            run(properties.id);\n        }, [properties]);\n    }\n    return useInner;\n}\n",
			1,
		},
		// An arrow-bodied component takes its name from the variable it is assigned to.
		{
			"inside an arrow component",
			"const Thing = (properties: { id: string }) => {\n    React.useEffect(() => {\n        run(properties.id);\n    }, [properties]);\n    return null;\n};\n",
			1,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ReactHookNoPropertiesInDependencies, propertiesInDependenciesFile, testCase.sourceText)
			expected := make([]string, testCase.count)
			for index := range expected {
				expected[index] = "extractPropertiesFirst"
			}
			rule_testing.ExpectFindings(t, result, expected...)
		})
	}
}

func TestReactHookNoPropertiesInDependenciesStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The shape the rule is asking authors to write. This is the dominant correct form and the
		// one a rule matching the token `properties` anywhere in the array would wrongly report.
		{
			"an extracted property variable",
			propertiesInDependenciesFile,
			"function Thing(properties: { id: string }) {\n    const propertiesId = properties.id;\n    React.useEffect(() => {\n        run(propertiesId);\n    }, [propertiesId]);\n    return null;\n}\n",
		},
		// A member access in the array is not the whole object, so it is not this rule's concern.
		{
			"a member access in the dependency array",
			propertiesInDependenciesFile,
			"function Thing(properties: { id: string }) {\n    React.useEffect(() => {\n        run(properties.id);\n    }, [properties.id]);\n    return null;\n}\n",
		},
		// `properties` used inside the callback body is the entire point of the component. Only the
		// dependency array is the subject, so a rule scanning the call subtree reports this.
		{
			"properties used only in the callback body",
			propertiesInDependenciesFile,
			"function Thing(properties: { id: string }) {\n    React.useEffect(() => {\n        run(properties.id, properties.name);\n    }, [properties.id]);\n    return null;\n}\n",
		},
		// The `React.` requirement. A bare hook call is invisible to the original, so reproducing
		// that narrowing is parity rather than a gap to fix. Measured: this codebase has 0 bare
		// hook calls, so the narrowing costs nothing while remaining a real behavioral difference.
		{
			"a bare useEffect call",
			propertiesInDependenciesFile,
			"function Thing(properties: { id: string }) {\n    useEffect(() => {\n        run(properties.id);\n    }, [properties]);\n    return null;\n}\n",
		},
		{
			"a hook on some other namespace",
			propertiesInDependenciesFile,
			"function Thing(properties: { id: string }) {\n    Other.useEffect(() => {\n        run(properties.id);\n    }, [properties]);\n    return null;\n}\n",
		},
		// A React method that is not a dependency-taking hook. Its second argument being an array
		// containing `properties` must not report.
		{
			"a react method with no dependency array",
			propertiesInDependenciesFile,
			"function Thing(properties: { id: string }) {\n    React.createElement('div', [properties]);\n    return null;\n}\n",
		},
		// The component-containment test, both halves. A custom hook is not a component, and the
		// hook exclusion is what stops `useThing` passing the capital test on its third letter.
		{
			"inside a custom hook rather than a component",
			propertiesInDependenciesFile,
			"function useThing(properties: { id: string }) {\n    React.useEffect(() => {\n        run(properties.id);\n    }, [properties]);\n    return null;\n}\n",
		},
		// A custom hook nested inside a component is the mirror of the firing case above, and it
		// pins the direction: the walk keeps going past the hook and finds the component, so this
		// reports rather than staying silent. Kept on the silent side would be wrong, and writing
		// it here first is how that was checked.
		//
		// The remaining case for the walk is a helper nested in a helper, where nothing in the
		// chain is a component and no amount of walking finds one.
		{
			"a helper nested inside another helper",
			propertiesInDependenciesFile,
			"function buildOuter(properties: { id: string }) {\n    function buildInner() {\n        React.useEffect(() => {\n            run(properties.id);\n        }, [properties]);\n    }\n    return buildInner;\n}\n",
		},
		{
			"inside a lowercase helper function",
			propertiesInDependenciesFile,
			"function buildThing(properties: { id: string }) {\n    React.useEffect(() => {\n        run(properties.id);\n    }, [properties]);\n    return null;\n}\n",
		},
		{
			"at module scope with no enclosing function",
			propertiesInDependenciesFile,
			"React.useEffect(() => {\n    run(properties.id);\n}, [properties]);\n",
		},
		// A dependency argument that is not an array literal. There is nothing to inspect and
		// guessing what the variable holds is not this rule's business.
		{
			"a dependency variable rather than a literal",
			propertiesInDependenciesFile,
			"function Thing(properties: { id: string }) {\n    React.useEffect(() => {\n        run(properties.id);\n    }, dependencies);\n    return null;\n}\n",
		},
		{
			"a hook called with no dependency argument",
			propertiesInDependenciesFile,
			"function Thing(properties: { id: string }) {\n    React.useEffect(() => {\n        run(properties.id);\n    });\n    return null;\n}\n",
		},
		// useImperativeHandle's second argument is its factory, not its dependencies. An array
		// literal appearing there must not be read as a dependency array.
		{
			"useImperativeHandle with an array in the factory position",
			propertiesInDependenciesFile,
			"function Thing(properties: { id: string }) {\n    React.useImperativeHandle(properties.reference, [properties]);\n    return null;\n}\n",
		},
		// A different identifier that merely contains the word.
		{
			"an identifier that is not exactly properties",
			propertiesInDependenciesFile,
			"function Thing(properties: { id: string }) {\n    React.useEffect(() => {\n        run(properties.id);\n    }, [componentProperties]);\n    return null;\n}\n",
		},
		// The file gate. This is a .ts file, so the rule declines it entirely.
		{
			"a plain typescript file",
			"/repository/source/Thing.ts",
			"function Thing(properties: { id: string }) {\n    React.useEffect(() => {\n        run(properties.id);\n    }, [properties]);\n    return null;\n}\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.Run(t, ReactHookNoPropertiesInDependencies, testCase.fileName, testCase.sourceText))
		})
	}
}

// TestReactHookNoPropertiesInDependenciesSkipsParenthesizedReceiver pins a shape the rule declined
// before it adopted the shared namespaced-member predicate.
//
// `(React).useEffect(...)` is the same call as `React.useEffect(...)`, and the hand-rolled receiver
// test read the parenthesis as a non-identifier and declined. Nobody writes it deliberately, which
// is exactly why no fixture covered it and the miss was invisible.
func TestReactHookNoPropertiesInDependenciesSkipsParenthesizedReceiver(t *testing.T) {
	t.Parallel()

	const sourceText = "function Thing(properties: { id: string }) {\n" +
		"    (React).useEffect(() => { run(properties.id); }, [properties]);\n" +
		"    return <div />;\n}\n"
	result := rule_testing.Run(t, ReactHookNoPropertiesInDependencies, propertiesInDependenciesFile, sourceText)
	if len(result.Diagnostics) != 1 {
		t.Errorf("a parenthesized React receiver produced %d findings, want 1",
			len(result.Diagnostics))
	}
}
