package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// A `.tsx` name, because most cases here write JSX in a component body the way React's own probe
// inputs do. Parsing those as plain TypeScript reads `<div>` as a type assertion, which would not
// change this rule's verdict but would stop the fixtures from being the source they claim to be.
const useMemoFile = "use_memo.tsx"

// The upstream corpus is six error-named fixtures and only two of them reach this rule as a lint
// rule.
//
// React ships `error.invalid-reassign-variable-in-usememo` and `error.useMemo-non-literal-depslist`,
// which report, and four more whose bodies are written inside a function named lowercase
// `component`: `error.invalid-ReactUseMemo-async-callback`, `error.invalid-useMemo-async-callback`,
// `error.invalid-useMemo-callback-args` and `error.useMemo-callback-generator`. Those four exercise
// the rule through React's compiler test harness, which compiles whatever it is handed, and NOT
// through the ESLint rule, which applies the component gate first. Run verbatim through React's own
// `react-hooks/use-memo`, all four are silent and only `rules-of-hooks` fires on them, saying the
// call sits in a function that is neither a component nor a hook.
//
// So four of the six error-named fixtures are the wrong half, which is the hazard the dispatch named
// and which cost `error-boundaries` real work in the other direction. The two that do report are
// transcribed here verbatim. The four that do not are represented by their own shapes rewritten
// under a capitalized name, because the shape each was written to test is real and worth pinning
// even though the file as shipped cannot pin it.
//
// Every case beyond those two was measured by running React's own `react-hooks/use-memo` through the
// ESLint Linter API on that exact input before it was written down, over roughly a hundred and ten
// probe inputs. Reading gave the wrong answer twice, on the import forms and on the rest parameter,
// and both are recorded at `UseMemo`.

type useMemoCase struct {
	name   string
	source string
	// ids are the message ids expected, one per finding, in the order the rule reports them.
	ids []string
	// reported is the source text the FIRST finding's range must cover. Asserted because
	// `ExpectFindings` sees ids and count and never where a finding points, and this rule has four
	// different anchors: the whole call, the callback, the first parameter, and an assignment
	// target.
	reported string
}

func TestUseMemoFires(t *testing.T) {
	t.Parallel()

	cases := []useMemoCase{
		{
			name:     "upstreamReassignVariableInUseMemo",
			source:   "function Component() {\n  let x;\n  const y = useMemo(() => {\n    let z;\n    x = [];\n    z = true;\n    return z;\n  }, []);\n  return [x, y];\n}\n",
			ids:      []string{"useMemoCallbackReassignsOuterVariable"},
			reported: "x",
		},
		{
			name:     "upstreamNonLiteralDepsList",
			source:   "import {useMemo} from 'react';\n\n// react-hooks-deps would error on this code (complex expression in depslist),\n// so Forget could bailout here\nfunction App({text, hasDeps}) {\n  const resolvedText = useMemo(\n    () => {\n      return text.toUpperCase();\n    },\n    hasDeps ? null : [text], // should be DCE'd\n  );\n  return resolvedText;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: App,\n  params: ['TodoAdd'],\n  isComponent: 'TodoAdd',\n};\n",
			ids:      []string{"useMemoDependencyListNotArrayLiteral"},
			reported: "hasDeps ? null : [text]",
		},
		{
			name:     "asyncArrowCallback",
			source:   "function Component(props) {\n  const x = useMemo(async () => props.a, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackAsyncOrGenerator"},
			reported: "async () => props.a",
		},
		{
			name:     "generatorFunctionCallback",
			source:   "function Component(props) {\n  const x = useMemo(function* () { yield 1; }, []);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackAsyncOrGenerator"},
			reported: "function* () { yield 1; }",
		},
		{
			name:     "callbackWithOneParameter",
			source:   "function Component(props) {\n  const x = useMemo((c) => c, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackHasParameters"},
			reported: "c",
		},
		{
			name:     "callbackWithTwoParameters",
			source:   "function Component(props) {\n  const x = useMemo((a, b) => a, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackHasParameters"},
			reported: "a",
		},
		{
			name:     "callbackWithDestructuredParameter",
			source:   "function Component(props) {\n  const x = useMemo(({a}) => a, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackHasParameters"},
			reported: "{a}",
		},
		{
			name:     "callbackWithDefaultParameter",
			source:   "function Component(props) {\n  const x = useMemo((a = 1) => a, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackHasParameters"},
			reported: "a = 1",
		},
		{
			name:     "callWithNoArguments",
			source:   "function Component(props) {\n  const x = useMemo();\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoMissingCallback"},
			reported: "useMemo()",
		},
		{
			name:     "firstArgumentIsMemberAccess",
			source:   "function Component(props) {\n  const x = useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackNotInline"},
			reported: "props.fn",
		},
		{
			name:     "firstArgumentIsNumericLiteral",
			source:   "function Component(props) {\n  const x = useMemo(42, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackNotInline"},
			reported: "42",
		},
		{
			name:     "firstArgumentIsLocalFunctionName",
			source:   "function Component(props) {\n  const cb = () => props.a;\n  const x = useMemo(cb, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackNotInline"},
			reported: "cb",
		},
		{
			name:     "dependencyListIsVariable",
			source:   "function Component(props) {\n  const d = [props.a];\n  const x = useMemo(() => 1, d);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoDependencyListNotArrayLiteral"},
			reported: "d",
		},
		{
			name:     "dependencyListIsNull",
			source:   "function Component(props) {\n  const x = useMemo(() => 1, null);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoDependencyListNotArrayLiteral"},
			reported: "null",
		},
		{
			name:     "dependencyListHasSpread",
			source:   "function Component(props) {\n  const x = useMemo(() => 1, [...props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoDependencyListNotArrayLiteral"},
			reported: "[...props.a]",
		},
		{
			name:     "dependencyListHasHole",
			source:   "function Component(props) {\n  const x = useMemo(() => 1, [, props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoDependencyListNotArrayLiteral"},
			reported: "[, props.a]",
		},
		{
			name:     "dependencyIsCallExpression",
			source:   "function Component(props) {\n  const x = useMemo(() => 1, [f()]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoDependencyNotSimple"},
			reported: "f()",
		},
		{
			name:     "dependencyIsBinaryExpression",
			source:   "function Component(props) {\n  const x = useMemo(() => props.a, [props.a + 1]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoDependencyNotSimple"},
			reported: "props.a + 1",
		},
		{
			name:     "dependencyIsComputedStringIndex",
			source:   "function Component(props) {\n  const x = useMemo(() => 1, [props['a']]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoDependencyNotSimple"},
			reported: "props['a']",
		},
		{
			name:     "dependencyIsNumericLiteral",
			source:   "function Component(props) {\n  const x = useMemo(() => 1, [1]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoDependencyNotSimple"},
			reported: "1",
		},
		{
			name:     "dependencyIsThisMember",
			source:   "function Component(props) {\n  const x = useMemo(() => 1, [this.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoDependencyNotSimple"},
			reported: "this.a",
		},
		{
			name:     "twoBadDependenciesReportTwice",
			source:   "function Component(props) {\n  const x = useMemo(() => 1, [f(), g()]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoDependencyNotSimple", "useMemoDependencyNotSimple"},
			reported: "f()",
		},
		{
			name:     "asyncCallbackWithParameterReportsTwice",
			source:   "function Component(props) {\n  const x = useMemo(async (c) => c, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackHasParameters", "useMemoCallbackAsyncOrGenerator"},
			reported: "c",
		},
		{
			name:     "notInlineWithBadDependencyReportsTwice",
			source:   "function Component(props) {\n  const x = useMemo(props.fn, [f()]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackNotInline", "useMemoDependencyNotSimple"},
			reported: "props.fn",
		},
		{
			name:     "twoCallsReportSeparately",
			source:   "function Component(props) {\n  const a = useMemo(props.f, [props.a]);\n  const b = useMemo(props.g, [props.b]);\n  return <div>{a}{b}</div>;\n}\n",
			ids:      []string{"useMemoCallbackNotInline", "useMemoCallbackNotInline"},
			reported: "props.f",
		},
		{
			name:     "compoundAssignmentToOuterVariable",
			source:   "function Component() {\n  let x = 0;\n  const y = useMemo(() => { x += 1; return 1; }, []);\n  return [x, y];\n}\n",
			ids:      []string{"useMemoCallbackReassignsOuterVariable"},
			reported: "x",
		},
		{
			name:     "assignmentToParameterOfComponent",
			source:   "function Component(p) {\n  const y = useMemo(() => { p = 1; return 1; }, []);\n  return [p, y];\n}\n",
			ids:      []string{"useMemoCallbackReassignsOuterVariable"},
			reported: "p",
		},
		{
			name:     "insideHookRatherThanComponent",
			source:   "function useThing(props) {\n  const x = useMemo(props.fn, [props.a]);\n  return x;\n}\n",
			ids:      []string{"useMemoCallbackNotInline"},
			reported: "props.fn",
		},
		{
			name:     "arrowFunctionComponent",
			source:   "const Component = (props) => {\n  const x = useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n};\n",
			ids:      []string{"useMemoCallbackNotInline"},
			reported: "props.fn",
		},
		{
			name:     "reactMemberCallOnGlobal",
			source:   "function Component(props) {\n  const x = React.useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackNotInline"},
			reported: "props.fn",
		},
		{
			name:     "namedImportFromReact",
			source:   "import {useMemo} from 'react';\nfunction Component(props) {\n  const x = useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackNotInline"},
			reported: "props.fn",
		},
		{
			name:     "renamedImportKeysOnImportedName",
			source:   "import {useMemo as um} from 'react';\nfunction Component(props) {\n  const x = um(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackNotInline"},
			reported: "props.fn",
		},
		{
			name:     "namespaceImportNamedReact",
			source:   "import * as React from 'react';\nfunction Component(props) {\n  const x = React.useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackNotInline"},
			reported: "props.fn",
		},
		{
			name:     "defaultImportNamedReact",
			source:   "import React from 'react';\nfunction Component(props) {\n  const x = React.useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackNotInline"},
			reported: "props.fn",
		},
		{
			// The receiver is keyed on the LOCAL name alone, with no module test, so a default
			// import named `React` reports whatever module it came from. Measured against React's
			// own rule on this exact source, and it is the input that killed a surviving mutant
			// whose verdict was right where the original rule was wrong. See the note at
			// `reactModuleLocalName`.
			name:     "defaultImportNamedReactFromAnotherModule",
			source:   "import React from 'preact/compat';\nfunction Component(props) {\n  const x = React.useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackNotInline"},
			reported: "props.fn",
		},
		{
			// The namespace form of the same, sharing upstream's `ImportDefault`/`ImportNamespace`
			// arm. Measured reporting.
			name:     "namespaceImportNamedReactFromAnotherModule",
			source:   "import * as React from 'not-react';\nfunction Component(props) {\n  const x = React.useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackNotInline"},
			reported: "props.fn",
		},
		{
			// A callback bound to a variable one line above still reports, which is worth pinning
			// because a sibling porting `void-use-memo` measured the OPPOSITE for their half of
			// this function: upstream's `functions` table is keyed on the single-assignment
			// temporary a FunctionExpression produced, so binding the callback to a variable
			// interposes a store and a load, the identifier differs, and their lookup misses.
			// These four diagnostics are not weakened that way, because `dropManualMemoization`
			// asks whether the argument's identifier is in `sidemap.functions`, and a variable
			// reference never is, which is exactly the reporting answer. Measured on React's own
			// rule.
			name:     "callbackBoundToAVariableStillReports",
			source:   "function Component(props) {\n  const cb = () => 1;\n  const x = useMemo(cb, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackNotInline"},
			reported: "cb",
		},
		{
			// A conditional expression producing a function is not an inline function, so it
			// reports and the span is the whole conditional. Measured.
			name:     "conditionalCallbackExpression",
			source:   "function Component(props) {\n  const x = useMemo(props.c ? () => 1 : () => 2, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackNotInline"},
			reported: "props.c ? () => 1 : () => 2",
		},
		{
			name:     "useCallbackWithNoArguments",
			source:   "function Component(props) {\n  const x = useCallback();\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoMissingCallback"},
			reported: "useCallback()",
		},
		{
			name:     "useCallbackNotInline",
			source:   "function Component(props) {\n  const x = useCallback(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackNotInline"},
			reported: "props.fn",
		},
		{
			name:     "useCallbackDependencyListNotLiteral",
			source:   "function Component(props) {\n  const x = useCallback(() => 1, props.d);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoDependencyListNotArrayLiteral"},
			reported: "props.d",
		},
		{
			name:     "componentWithRefSecondParameter",
			source:   "function Component(props, ref) {\n  const x = useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
			ids:      []string{"useMemoCallbackNotInline"},
			reported: "props.fn",
		},
		{
			name:     "componentNestedInsideAnotherFunction",
			source:   "function Outer() {\n  function Component(props) {\n    const x = useMemo(props.fn, [props.a]);\n    return <div>{x}</div>;\n  }\n  return Component;\n}\n",
			ids:      []string{"useMemoCallbackNotInline"},
			reported: "props.fn",
		},
		{
			// Upstream is silent here, and NOT because of anything this rule decides. Reassigning a
			// `const` is a semantic error, and React's own `react-hooks/syntax` rule claims the
			// input first with "Cannot reassign a `const` variable", so the compiler never lowers a
			// StoreContext for it and this category never sees it. Confirmed by running every
			// `react-hooks` rule over this exact source: two `syntax` findings appear and no
			// `use-memo` one.
			//
			// Pinned as REPORTING rather than as silent, which is the port brief's rule for a case
			// decided above the rule. `rule_testing` runs one rule against one file and consults no
			// other, so no layer here could reproduce the upstream silence, and bending this rule to
			// go quiet on a `const` target would break the ordinary `let` case it shares a code path
			// with. cohere surfaces the same conflict through the type checker rather than through a
			// lint rule, so a reader of real output sees both findings rather than this one alone.
			name:     "assignmentToOuterConstAlsoReports",
			source:   "function Component() {\n  const x = 1;\n  const y = useMemo(() => { x = 1; return 1; }, []);\n  return [x, y];\n}\n",
			ids:      []string{"useMemoCallbackReassignsOuterVariable"},
			reported: "x",
		},
		{
			name:     "restParameterElsewhereStillReports",
			source:   "function Component(props) {\n  const bad = useMemo(props.fn, [props.a]);\n  const helper = (...a) => a;\n  return <div>{bad}{helper(1)}</div>;\n}\n",
			ids:      []string{"useMemoCallbackNotInline"},
			reported: "props.fn",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, UseMemo, useMemoFile, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.ids...)
			if len(result.Diagnostics) == 0 {
				return
			}
			first := result.Diagnostics[0].Range
			got := testCase.source[first.Pos():first.End()]
			if got != testCase.reported {
				t.Errorf("finding covers %q, want %q", got, testCase.reported)
			}
		})
	}
}

func TestUseMemoStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []useMemoCase{
		{
			name:   "inlineArrowWithSimpleDeps",
			source: "function Component(props) {\n  const x = useMemo(() => props.a + 1, [props.a]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "inlineFunctionExpression",
			source: "function Component(props) {\n  const x = useMemo(function () { return props.a; }, [props.a]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "inlineNamedFunctionExpression",
			source: "function Component(props) {\n  const x = useMemo(function inner() { return props.a; }, [props.a]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "emptyDependencyList",
			source: "function Component(props) {\n  const x = useMemo(() => 1, []);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "noDependencyListAtAll",
			source: "function Component(props) {\n  const x = useMemo(() => props.a);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "memberChainDependency",
			source: "function Component(props) {\n  const x = useMemo(() => 1, [props.a.b.c]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "optionalChainDependency",
			source: "function Component(props) {\n  const x = useMemo(() => 1, [props?.a?.b]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "computedNumericIndexDependency",
			source: "function Component(props) {\n  const x = useMemo(() => 1, [props[0]]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "thirdArgumentIsIgnored",
			source: "function Component(props) {\n  const x = useMemo(() => 1, [props.a], extra);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "lowercaseFunctionIsNotCompiled",
			source: "function widget(props) {\n  const x = useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "lowercaseHookNameIsNotCompiled",
			source: "function usething(props) {\n  const x = useMemo(props.fn, [props.a]);\n  return x;\n}\n",
		},
		{
			name:   "moduleScopeIsNotCompiled",
			source: "const x = useMemo(fn, [a]);\n",
		},
		{
			name:   "functionWithNoJsxOrHookEvidence",
			source: "function Widget(props) {\n  const x = notMemo(props.fn, [props.a]);\n  return x;\n}\n",
		},
		{
			name:   "componentWithRestParameters",
			source: "function Component(...props) {\n  const x = useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "componentWithTwoNonRefParameters",
			source: "function Component(props, other) {\n  const x = useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "localBindingShadowsUseMemo",
			source: "function Component(props) {\n  const useMemo = (a, b) => a;\n  const x = useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			// A callee reached through a variable is silent, because upstream's `manualMemos`
			// sidemap is keyed on the temporary the LoadGlobal produced and a copy through a
			// variable is a different identifier. Measured, and it is the callee-side twin of
			// `callbackBoundToAVariableStillReports` above landing the other way.
			name:   "calleeReachedThroughAVariable",
			source: "function Component(props) {\n  const m = useMemo;\n  const x = m(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			// A callback whose only return is bare is CLEAN for these four diagnostics. Recorded
			// because React's own pass documentation at `docs/passes/40-validateUseMemo.md` says a
			// bare `return;` triggers an error, and a sibling measured that the documentation is
			// wrong: the lowering hardcodes an explicit return variant and the case is silent.
			// That claim is about the `void-use-memo` half rather than this one, and it is pinned
			// here as well so a reader following the citation into this rule finds the measurement
			// beside it.
			name:   "bareReturnInCallbackIsClean",
			source: "function Component(props) {\n  const x = useMemo(() => { return; }, [props.a]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			// A module-scope `const useMemo` shadows the global and silences these four, which is
			// worth pinning because it is where the two halves of `validateUseMemo` part company.
			// A sibling measured that a module-level shadow still reports for `void-use-memo`,
			// whose lookup keys on the LoadGlobal name alone. These four go through the
			// module-aware callee test instead, so the shadow wins. Measured on React's own rule.
			name:   "moduleScopeShadowSilencesTheseFour",
			source: "const useMemo = (a, b) => a;\nfunction Component(props) {\n  const x = useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "importFromNonReactModule",
			source: "import {useMemo} from 'not-react';\nfunction Component(props) {\n  const x = useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "importOtherNameAliasedToUseMemo",
			source: "import {somethingElse as useMemo} from 'react';\nfunction Component(props) {\n  const x = useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "defaultImportNotNamedReact",
			source: "import Rct from 'react';\nfunction Component(props) {\n  const x = Rct.useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "otherNamespaceMemberCall",
			source: "function Component(props) {\n  const x = Foo.useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "localObjectNamedReact",
			source: "function Component(props) {\n  const React = {useMemo: (a, b) => a};\n  const x = React.useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "innerVariableAssignmentOnly",
			source: "function Component() {\n  const y = useMemo(() => { let z; z = 1; return z; }, []);\n  return y;\n}\n",
		},
		{
			name:   "assignmentToTrueGlobal",
			source: "function Component() {\n  const y = useMemo(() => { globalThing = 1; return 1; }, []);\n  return y;\n}\n",
		},
		{
			name:   "propertyWriteOnOuterVariable",
			source: "function Component() {\n  let x = {};\n  const y = useMemo(() => { x.a = 1; return 1; }, []);\n  return [x, y];\n}\n",
		},
		{
			name:   "incrementOfOuterVariable",
			source: "function Component() {\n  let x = 0;\n  const y = useMemo(() => { x++; return 1; }, []);\n  return [x, y];\n}\n",
		},
		{
			name:   "assignmentFromNestedFunction",
			source: "function Component() {\n  let x = 0;\n  const y = useMemo(() => { const f = () => { x = 1; }; f(); return 1; }, []);\n  return [x, y];\n}\n",
		},
		{
			name:   "useCallbackReassignmentIsAllowed",
			source: "function Component() {\n  let x;\n  const y = useCallback(() => { x = 1; return 1; }, []);\n  return [x, y];\n}\n",
		},
		{
			name:   "useCallbackParametersAreAllowed",
			source: "function Component(props) {\n  const x = useCallback((c) => c, [props.a]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "useCallbackAsyncIsAllowed",
			source: "function Component(props) {\n  const x = useCallback(async () => 1, [props.a]);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "useEffectIsNotAManualMemoHook",
			source: "function Component(props) {\n  const x = useEffect(() => {}, props.d);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "spreadAsOnlyArgument",
			source: "function Component(props) {\n  const x = useMemo(...props.args);\n  return <div>{x}</div>;\n}\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, UseMemo, useMemoFile, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestUseMemoRequiresTheTypedHarness pins that this rule needs a real checker.
//
// The port brief measures that a rule declaring `NeedsTypeChecker` and then guarding on nil goes
// COMPLETELY SILENT under the plain harness, which makes every StaysSilent case pass vacuously and
// makes the whole suite look healthy while the rule decides nothing. That failure is invisible from
// outside, so it is pinned here rather than left to a reader to notice: an input that reports under
// `RunTyped` must be silent under `Run`, and if a later change makes the plain harness report, this
// test fails and says the guard moved.
func TestUseMemoRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	source := "function Component(props) {\n  const x = useMemo(props.fn, [props.a]);\n  return <div>{x}</div>;\n}\n"

	typed := rule_testing.RunTyped(t, UseMemo, useMemoFile, source)
	rule_testing.ExpectFindings(t, typed, "useMemoCallbackNotInline")

	untyped := rule_testing.Run(t, UseMemo, useMemoFile, source)
	rule_testing.ExpectClean(t, untyped)
}

// TestUseMemoMessages asserts the message identifiers and descriptions against literals typed here.
//
// The port brief's rule: never assert a finding against the rule's own message constant, because
// both sides move together under mutation and a message-text mutant survives a test written that
// way. These are literals, so a rename or a rewrite of either field fails here.
//
// A `rule.Message` is `{Id, Description}` with no interpolation layer, and this rule uses no format
// verbs, so there is no rendered text to guard beyond these two fields.
func TestUseMemoMessages(t *testing.T) {
	t.Parallel()

	cases := []struct {
		message         rule.Message
		id              string
		descriptionHead string
	}{
		{messageUseMemoMissingCallback, "useMemoMissingCallback",
			"This `useMemo` or `useCallback` call was given no arguments at all"},
		{messageUseMemoCallbackNotInline, "useMemoCallbackNotInline",
			"The first argument here is not a function written in place."},
		{messageUseMemoDependencyListNotArrayLiteral, "useMemoDependencyListNotArrayLiteral",
			"The dependency list is not an array literal written in place."},
		{messageUseMemoDependencyNotSimple, "useMemoDependencyNotSimple",
			"This dependency is not a plain value reference."},
		{messageUseMemoCallbackHasParameters, "useMemoCallbackHasParameters",
			"This memoization callback declares a parameter."},
		{messageUseMemoCallbackAsyncOrGenerator, "useMemoCallbackAsyncOrGenerator",
			"This memoization callback is an async function or a generator."},
		{messageUseMemoCallbackReassignsOuterVariable, "useMemoCallbackReassignsOuterVariable",
			"This memoization callback assigns to a variable declared outside of it."},
	}

	for _, testCase := range cases {
		t.Run(testCase.id, func(t *testing.T) {
			t.Parallel()
			if testCase.message.Id != testCase.id {
				t.Errorf("id is %q, want %q", testCase.message.Id, testCase.id)
			}
			if !strings.HasPrefix(testCase.message.Description, testCase.descriptionHead) {
				t.Errorf("description is %q, want it to start with %q",
					testCase.message.Description, testCase.descriptionHead)
			}
		})
	}
}

// TestUseMemoRuleName pins the registered name in its bare spelling.
//
// The port brief measures that a rename to the slash-namespaced form survives every guard, because
// `matchRegistered` tries an exact match against the inventory entry `react/use-memo` before it
// strips the namespace. So the parity guard reports the same count either way and cannot see the
// mistake. This is the assertion that can.
func TestUseMemoRuleName(t *testing.T) {
	t.Parallel()

	if UseMemo.Name != "react-hooks/use-memo" {
		t.Errorf("rule name is %q, want %q", UseMemo.Name, "use-memo")
	}
	if !UseMemo.NeedsTypeChecker {
		t.Error("rule must declare NeedsTypeChecker: the callee test resolves bindings")
	}
}
