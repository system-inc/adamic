package nexus

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const arrowFile = "/repository/source/Thing.tsx"

// A `.ts` file, where `<T>(value: T) =>` is a generic arrow rather than the opening of a JSX element.
const arrowPlainFile = "/repository/source/Thing.ts"

func TestConsistencyNoMultilineArrowFunctionFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"multi-line arrow with a block body",
			"const compute = (input) => {\n    return input * 2;\n};\n",
			[]string{"multilineArrow"},
		},
		{
			"react hook",
			"React.useEffect(() => {\n    run();\n}, []);\n",
			[]string{"reactHookArrow"},
		},
		{
			"react forwardRef",
			"const Thing = React.forwardRef((properties, reference) => {\n    return null;\n});\n",
			[]string{"reactHookArrow"},
		},
		{
			"addEventListener",
			"element.addEventListener('click', (event) => {\n    run(event);\n});\n",
			[]string{"addEventListenerArrow"},
		},
		{
			// A nested function binds its own `this`, so it says nothing about the outer arrow and
			// the outer arrow is still reportable.
			"this inside a nested function does not exempt the outer arrow",
			"const compute = () => {\n    return function () {\n        return this.value;\n    };\n};\n",
			[]string{"multilineArrow"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoMultilineArrowFunction, arrowFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

func TestConsistencyNoMultilineArrowFunctionStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"single-line implicit return", "const add = (a, b) => a + b;\n"},
		// The case that separates "has a block body" from "spans lines". An arrow has no prototype
		// and a function expression does, so converting this one changes what the value can do.
		{"single-line block body", "const noop = () => {};\n"},
		{"single-line block with a statement", "const run = () => { start(); };\n"},
		// A function declaration would rebind `this`, so the conversion would change meaning.
		{"arrow using this", "const compute = () => {\n    return this.value;\n};\n"},
		// A nested arrow inherits the outer `this`, so a `this` inside it is the outer arrow's.
		{"this inside a nested arrow", "const compute = () => {\n    return items.map(() => this.value);\n};\n"},
		{"a function declaration is already the target shape", "function compute(input) {\n    return input * 2;\n}\n"},
		{"a function expression", "const compute = function (input) {\n    return input * 2;\n};\n"},
		{"multi-line implicit return is still an implicit return", "const compute = (input) =>\n    input * 2;\n"},

		// An arrow indented inside a multi-line literal. Pos() sits before the leading newline and
		// indentation, so measuring the span from it puts the start on the previous line and this
		// genuinely single-line arrow reads as multi-line. This exact shape was the rule's one
		// false finding on the ahra tree, in a fixture list of non-constructible values where
		// converting it to a function expression would have given it a prototype and flipped the
		// assertion it exists to make.
		{"single-line arrow inside a multi-line array", "const values = [\n    SimpleClass,\n    () => {},\n    42,\n];\n"},
		{"single-line arrow as a multi-line call argument", "register(\n    'name',\n    () => {},\n);\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoMultilineArrowFunction, arrowFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// One arrow must never produce two findings. The general case has to skip what the call-expression
// case already claimed, and a fixture is the only thing that keeps that true.
func TestConsistencyNoMultilineArrowFunctionReportsEachArrowOnce(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantId     string
	}{
		{"react hook", "React.useEffect(() => {\n    run();\n}, []);\n", "reactHookArrow"},
		{"addEventListener", "element.addEventListener('click', (event) => {\n    run(event);\n});\n", "addEventListenerArrow"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoMultilineArrowFunction, arrowFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantId)
		})
	}
}

// The fix, as fix vectors. Every expected output below was produced by the installed TypeScript
// original (`ConsistencyNoMultilineArrowFunctionRule.ts`, driven through ESLint's `verifyAndFix`
// on 2026-10-01) and is byte-identical to it, except where a case says otherwise. The first four are
// the real ahra sites the rule review named.
func TestConsistencyNoMultilineArrowFunctionFixes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
		wantSource string
	}{
		{
			"MetaPollingCommandLineInterface.ts: a command's run property", arrowFile,
			"const commands = {\n    'poll-state': {\n        description: 'Show saved polling cursors',\n        run: () => {\n            MetaPolling.showState();\n        },\n    },\n};\n",
			"const commands = {\n    'poll-state': {\n        description: 'Show saved polling cursors',\n        run: function() {\n            MetaPolling.showState();\n        },\n    },\n};\n",
		},
		{
			"ReplicateCommandLineInterface.ts: an async arrow in a class field", arrowFile,
			"class ReplicateCommandLineInterface {\n    commands = {\n        run: {\n            run: async (commandArguments: string[]) => {\n                const flags = commandArgumentFlags(commandArguments, ['wait', 'json']);\n                await go(flags);\n            },\n        },\n    };\n}\n",
			"class ReplicateCommandLineInterface {\n    commands = {\n        run: {\n            run: async function(commandArguments: string[]) {\n                const flags = commandArgumentFlags(commandArguments, ['wait', 'json']);\n                await go(flags);\n            },\n        },\n    };\n}\n",
		},
		{
			"ClaudeUtilities.ts: a setTimeout callback", arrowFile,
			"const timer = setTimeout(() => {\n    abortController.abort();\n}, timeoutInMilliseconds);\n",
			"const timer = setTimeout(function() {\n    abortController.abort();\n}, timeoutInMilliseconds);\n",
		},
		{
			"CompositeConversationSource.test.ts: a parameter default", arrowFile,
			"const source = {\n    id,\n    readConversation: (reference, options = {}) => {\n        calls.push({ reference, options });\n        return [];\n    },\n};\n",
			"const source = {\n    id,\n    readConversation: function(reference, options = {}) {\n        calls.push({ reference, options });\n        return [];\n    },\n};\n",
		},
		{
			"a bare parameter gains parentheses", arrowFile,
			"const f = value => {\n    return value * 2;\n};\n",
			"const f = function(value) {\n    return value * 2;\n};\n",
		},
		{
			"type parameters move after the keyword", arrowPlainFile,
			"const f = <T>(value: T): T => {\n    return value;\n};\n",
			"const f = function<T>(value: T): T {\n    return value;\n};\n",
		},
		{
			"a .tsx type parameter list keeps its disambiguating comma", arrowFile,
			"const f = <T,>(value: T) => {\n    return value;\n};\n",
			"const f = function<T,>(value: T) {\n    return value;\n};\n",
		},
		{
			// The upstream regression: the first `=>` in the text belongs to the parameter's type.
			"an arrow inside a parameter type is not this arrow", arrowFile,
			"const f = (build: (batch: number) => void): void => {\n    build(1);\n};\n",
			"const f = function(build: (batch: number) => void): void {\n    build(1);\n};\n",
		},
		{
			"a comment inside the signature survives", arrowFile,
			"const f = (value /* the input */: number) => {\n    return value;\n};\n",
			"const f = function(value /* the input */: number) {\n    return value;\n};\n",
		},
		{
			"a hook's expression body becomes a block that returns it", arrowFile,
			"const g = React.useCallback((value: number) => value + 1, []);\n",
			"const g = React.useCallback(function(value: number) { return value + 1; }, []);\n",
		},
		{
			// ESTree has no parenthesized node, so the original's body text never had the parens.
			"a parenthesized object body loses its parentheses", arrowFile,
			"const g = React.useMemo(() => ({ a: 1 }), []);\n",
			"const g = React.useMemo(function() { return { a: 1 }; }, []);\n",
		},
		{
			"an addEventListener argument", arrowFile,
			"element.addEventListener('click', (event) => {\n    run(event);\n});\n",
			"element.addEventListener('click', function(event) {\n    run(event);\n});\n",
		},
		{
			// `import.meta` belongs to the module, so unlike `new.target` it does not block the fix.
			"import.meta survives the conversion", arrowFile,
			"const f = () => {\n    return import.meta.url;\n};\n",
			"const f = function() {\n    return import.meta.url;\n};\n",
		},
		{
			// DIVERGES from the original, which emits `async function async(value)`, a function
			// named async, because it strips `async ` by text and this one has no space.
			"async with no space before the parameters", arrowFile,
			"const f = async(value) => {\n    await value;\n};\n",
			"const f = async function(value) {\n    await value;\n};\n",
		},
		{
			// DIVERGES from the original, which emits `async function<T>(value(: T))`, a parse
			// error, because it slices the type parameters at an offset that still counts `async `.
			"an async arrow with type parameters", arrowPlainFile,
			"const f = async <T>(value: T) => {\n    await value;\n};\n",
			"const f = async function<T>(value: T) {\n    await value;\n};\n",
		},
		{
			// trpc's invalidateQueries.test.tsx and issue-4049, reduced (#kq9vtva). An arrow that
			// leads an expression statement became `function() { ... };`, which parses as a
			// declaration with no name (TS1003). Parenthesized, it stays an expression.
			"an arrow leading an expression statement", arrowPlainFile,
			"function narrows() {\n    () => {\n        const utils = useUtils();\n        return utils;\n    };\n}\n",
			"function narrows() {\n    (function() {\n        const utils = useUtils();\n        return utils;\n    });\n}\n",
		},
		{
			"an async arrow leading an expression statement", arrowPlainFile,
			"async (value) => {\n    await value;\n};\n",
			"(async function(value) {\n    await value;\n});\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoMultilineArrowFunction, testCase.fileName, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected one finding, got %v", result.MessageIds())
			}
			rule_testing.ExpectFixedSource(t, result, testCase.wantSource)
		})
	}
}

// Each of these is reported and not rewritten. An arrow inherits `this`, `arguments` and
// `new.target`, and a function binds its own, so the conversion would change what they read; `super`
// is not legal in a function expression at all. The original rewrites the first two on a hook or
// listener argument (verified: `React.useEffect(() => { this.run(); })` becomes a function reading
// its own `this`), which is the one place this fix declines what the original performs.
func TestConsistencyNoMultilineArrowFunctionDeclinesAFixThatRebinds(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantId     string
	}{
		{"this in a hook argument", "React.useEffect(() => {\n    this.run();\n}, []);\n", "reactHookArrow"},
		{"this in a listener argument", "element.addEventListener('click', () => {\n    this.run();\n});\n", "addEventListenerArrow"},
		{"arguments", "function outer() {\n    const f = () => {\n        return arguments[0];\n    };\n}\n", "multilineArrow"},
		{"arguments in a nested arrow", "function outer() {\n    const f = () => {\n        return [1].map(() => arguments[0]);\n    };\n}\n", "multilineArrow"},
		{"new.target", "function Outer() {\n    const f = () => {\n        return new.target;\n    };\n}\n", "multilineArrow"},
		{"super", "class B extends A {\n    method() {\n        const f = () => {\n            return super.method();\n        };\n    }\n}\n", "multilineArrow"},
		{"this in a parameter default", "React.useCallback((value = this.fallback) => {\n    run(value);\n}, []);\n", "reactHookArrow"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoMultilineArrowFunction, arrowFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantId)
			if len(result.Diagnostics[0].Fixes) != 0 {
				t.Fatalf("expected no fix, got %q", result.Diagnostics[0].Fixes[0].Text)
			}
		})
	}
}

// The other direction of the same walk: a binding that only looks inherited still converts. A
// nested function binds its own `arguments`, and a property named `arguments` is not the binding.
func TestConsistencyNoMultilineArrowFunctionFixesWhatOnlyLooksInherited(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantSource string
	}{
		{
			"arguments inside a nested function",
			"const f = () => {\n    return function () { return arguments[0]; };\n};\n",
			"const f = function() {\n    return function () { return arguments[0]; };\n};\n",
		},
		{
			"a property named arguments",
			"const f = () => {\n    return command.arguments;\n};\n",
			"const f = function() {\n    return command.arguments;\n};\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoMultilineArrowFunction, arrowFile, testCase.sourceText)
			rule_testing.ExpectFixedSource(t, result, testCase.wantSource)
		})
	}
}
