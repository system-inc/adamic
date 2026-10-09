package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The imported corpus is two cases and the measured corpus is sixty five.
//
// oxc ships a linter rule for this one, unlike `gating` and `config`, and its tester block holds
// exactly one pass and one fail. Both are transcribed verbatim below. Its snapshot logs a single
// diagnostic for the fail case, spanning the callback, and React reports identically on the same
// bytes, so the two authorities agree completely on the imported floor and there was nothing to
// choose between them there.
//
// React ships no `VoidUseMemo` golden at all. The 325 fixtures vendored at
// `internal/react_conformance` were searched for the category name and for both message texts and
// contain neither, with a control search for "useMemo() callbacks may not accept parameters"
// finding `error.invalid-useMemo-callback-args.expect.md` in the same command, so the zero is a
// real absence rather than a bad pattern. The two error-named fixtures a porter would expect to
// be this rule's belong to `use-memo`, its sibling out of the same upstream function.
//
// Every other case here was measured by running React's own `react-hooks/void-use-memo` through
// the ESLint Linter API on that exact input before it was written down. Reading gave the wrong
// answer twice: on a bare `return;`, where React's own pass documentation contradicts React's own
// executable, and on a result assigned to a variable that is never read, which reads like a
// liveness question and is not one. Both are pinned below with the reasoning at the line.

type voidUseMemoCase struct {
	name   string
	source string
	want   []string
}

func TestVoidUseMemoFires(t *testing.T) {
	t.Parallel()

	// Message ids are written as literals rather than referenced through the rule's own message
	// constants, so an assertion cannot move together with the code it guards.
	const voidReturn = "useMemoCallbackReturnsNothing"
	const unusedResult = "useMemoResultUnused"

	cases := []voidUseMemoCase{
		// oxc's only fail case, transcribed from void_use_memo.rs. Its snapshot logs one
		// diagnostic spanning `() => {}`, and React reports identically on the same bytes.
		{
			name:   "upstreamOxcVoidCallback",
			source: "\nimport {useMemo} from 'react';\nfunction Component() {\n  const value = useMemo(() => {}, []);\n  return <div>{value}</div>;\n}\n",
			want:   []string{voidReturn},
		},
		// The canonical shape: a block body with no return at all.
		{
			name:   "voidCallbackWithSideEffect",
			source: "function Component() {\n  const x = useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
			want:   []string{voidReturn},
		},
		// A function expression is a callback too, not only an arrow.
		{
			name:   "voidCallbackFunctionExpression",
			source: "function Component() {\n  const x = useMemo(function () { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
			want:   []string{voidReturn},
		},
		// A name on the function expression changes nothing.
		{
			name:   "voidCallbackNamedFunctionExpression",
			source: "function Component() {\n  const x = useMemo(function compute() { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
			want:   []string{voidReturn},
		},
		// A throw is not a return, so there is no return terminal to find. Measured.
		{
			name:   "voidCallbackThrowOnly",
			source: "function Component() {\n  const x = useMemo(() => { throw new Error('x'); }, []);\n  return <div>{x}</div>;\n}\n",
			want:   []string{voidReturn},
		},
		// An async callback with no return reports HERE too, separately from the
		// `use-memo` rule's own async finding. Measured: this rule alone gives one finding.
		{
			name:   "voidCallbackAsync",
			source: "function Component() {\n  const x = useMemo(async () => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
			want:   []string{voidReturn},
		},
		// Same for a generator.
		{
			name:   "voidCallbackGenerator",
			source: "function Component() {\n  const x = useMemo(function* () { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
			want:   []string{voidReturn},
		},
		// Parameters are the `use-memo` rule's business; the missing return is still ours.
		{
			name:   "voidCallbackWithParameters",
			source: "function Component() {\n  const x = useMemo((a) => { foo(a); }, []);\n  return <div>{x}</div>;\n}\n",
			want:   []string{voidReturn},
		},
		// Parens around the callback are transparent. The span excludes them, which the
		// span test pins; measured on React at column 22 rather than 21.
		{
			name:   "voidCallbackParenthesized",
			source: "function Component() {\n  const x = useMemo((() => { foo(); }), []);\n  return <div>{x}</div>;\n}\n",
			want:   []string{voidReturn},
		},
		// Upstream tests `args.length === 0`, not `< 2`, so a missing dependency array
		// still reaches the callback check.
		{
			name:   "voidCallbackNoDependencyArgument",
			source: "function Component() {\n  const x = useMemo(() => { foo(); });\n  return <div>{x}</div>;\n}\n",
			want:   []string{voidReturn},
		},
		// `React.useMemo` is the second recognized spelling, via PropertyLoad on React.
		{
			name:   "voidCallbackReactNamespaced",
			source: "import * as React from 'react';\nfunction Component() {\n  const x = React.useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
			want:   []string{voidReturn},
		},
		// No import needed: the test is on the global binding name, not on resolution.
		{
			name:   "voidCallbackReactWithoutImport",
			source: "function Component() {\n  const x = React.useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
			want:   []string{voidReturn},
		},
		// An optional CALL still lowers to a CallExpression on the same callee. Measured.
		{
			name:   "voidCallbackOptionalCall",
			source: "function Component() {\n  const x = useMemo?.(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
			want:   []string{voidReturn},
		},
		// A hook is compiled on its name alone, with no parameter or return test.
		{
			name:   "voidCallbackInHook",
			source: "function useThing() {\n  const x = useMemo(() => { foo(); }, []);\n  return x;\n}\n",
			want:   []string{voidReturn},
		},
		// `use2Things` is a hook to React. The shelf's IsHookName rejects the digit, which
		// is why this package carries its own predicate. Measured as reporting.
		{
			name:   "voidCallbackInHookWithDigit",
			source: "function use2Things() {\n  const x = useMemo(() => { foo(); }, []);\n  return x;\n}\n",
			want:   []string{voidReturn},
		},
		// An arrow assigned to a capitalized declarator is a component.
		{
			name:   "voidCallbackInArrowComponent",
			source: "const Component = () => {\n  const x = useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n};\n",
			want:   []string{voidReturn},
		},
		// A component that creates no JSX still compiles, because the useMemo call is
		// itself a hook call and satisfies `callsHooksOrCreatesJsx`. Measured: reports.
		// This is the clause `error-boundaries` could omit and this rule cannot.
		{
			name:   "voidCallbackComponentWithoutJsx",
			source: "function Component() {\n  const x = useMemo(() => { foo(); }, []);\n  return x;\n}\n",
			want:   []string{voidReturn},
		},
		// Two parameters are valid when the second mentions `ref`.
		{
			name:   "voidCallbackWithRefSecondParameter",
			source: "function Component(props, ref) {\n  const x = useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
			want:   []string{voidReturn},
		},
		// A MODULE-level binding named useMemo still reports. LoadGlobal means `not bound
		// inside the compiled function`, and module scope is outside it. Measured, and it
		// is the opposite of what the name LoadGlobal suggests.
		{
			name:   "voidCallbackModuleScopeShadowStillReports",
			source: "const useMemo = (f) => f();\nfunction Component() {\n  const x = useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
			want:   []string{voidReturn},
		},
		// A shadow in a sibling block does not cover the call site.
		{
			name:   "voidCallbackBlockScopeShadowInComponent",
			source: "function Component() {\n  { const useMemo = (f) => f(); }\n  const x = useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
			want:   []string{voidReturn},
		},
		// A return inside a NESTED function does not count, because it cannot terminate the
		// callback. Added after a mutation removing the nested-function skip survived the whole
		// fixture set; both shapes were measured on React as reporting before either was written.
		{
			name:   "nestedArrowReturnDoesNotCount",
			source: "function Component() {\n  const x = useMemo(() => { const h = () => { return 1; }; h(); }, []);\n  return <div>{x}</div>;\n}\n",
			want:   []string{voidReturn},
		},
		{
			name:   "nestedFunctionDeclarationReturnDoesNotCount",
			source: "function Component() {\n  const x = useMemo(() => { function h() { return 1; } h(); }, []);\n  return <div>{x}</div>;\n}\n",
			want:   []string{voidReturn},
		},
		// One finding per offending call.
		{
			name:   "twoVoidCallbacksReportTwice",
			source: "function Component() {\n  const a = useMemo(() => { foo(); }, []);\n  const b = useMemo(() => { bar(); }, []);\n  return <div>{a}{b}</div>;\n}\n",
			want:   []string{voidReturn, voidReturn},
		},
		// The canonical unused shape: a call whose value nothing consumes.
		{
			name:   "resultUnusedBareStatement",
			source: "function Component() {\n  useMemo(() => { return 1; }, []);\n  return <div />;\n}\n",
			want:   []string{unusedResult},
		},
		// Parens do not consume a value. Measured as still reporting.
		{
			name:   "resultUnusedParenthesizedStatement",
			source: "function Component() {\n  (useMemo(() => { return 1; }, []));\n  return <div />;\n}\n",
			want:   []string{unusedResult},
		},
		// The LEFT operand of a comma is discarded, so it reports. The right operand does
		// not; that pair is the sharpest evidence the test is about value discard.
		{
			name:   "resultUnusedCommaLeftOperand",
			source: "function Component() {\n  (useMemo(() => { return 1; }, []), 0);\n  return <div />;\n}\n",
			want:   []string{unusedResult},
		},
		// Discard is per expression statement, not per path.
		{
			name:   "resultUnusedInsideIfBranch",
			source: "function Component({a}) {\n  if (a) { useMemo(() => { return 1; }, []); }\n  return <div />;\n}\n",
			want:   []string{unusedResult},
		},
		// One finding, not one per iteration. Measured.
		{
			name:   "resultUnusedInLoop",
			source: "function Component({items}) {\n  for (const i of items) { useMemo(() => { return i; }, [i]); }\n  return <div />;\n}\n",
			want:   []string{unusedResult},
		},
		// The unused span covers the whole member expression here, which the span test pins.
		{
			name:   "resultUnusedReactNamespaced",
			source: "import * as React from 'react';\nfunction Component() {\n  React.useMemo(() => { return 1; }, []);\n  return <div />;\n}\n",
			want:   []string{unusedResult},
		},
		// A parenthesized callee still resolves, and the span excludes the parens.
		{
			name:   "resultUnusedParenthesizedCallee",
			source: "function Component() {\n  (useMemo)(() => { return 1; }, []);\n  return <div />;\n}\n",
			want:   []string{unusedResult},
		},
		// One finding per unused call.
		{
			name:   "twoUnusedResultsReportTwice",
			source: "function Component() {\n  useMemo(() => { return 1; }, []);\n  useMemo(() => { return 2; }, []);\n  return <div />;\n}\n",
			want:   []string{unusedResult, unusedResult},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, VoidUseMemo, "component.tsx", testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.want...)
		})
	}
}

func TestVoidUseMemoStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []voidUseMemoCase{
		// oxc's only pass case. A concise arrow body lowers to an Implicit return.
		{
			name:   "upstreamOxcPass",
			source: "\nimport {useMemo} from 'react';\nfunction Component({a}) {\n  const x = useMemo(() => a + 1, [a]);\n  return <div>{x}</div>;\n}\n",
		},
		// The straightforward correct use.
		{
			name:   "explicitReturnValue",
			source: "function Component() {\n  const x = useMemo(() => { return 1; }, []);\n  return <div>{x}</div>;\n}\n",
		},
		// THE decisive case. A bare `return;` returns undefined and is still CLEAN, because
		// React's ReturnStatement lowering hardcodes returnVariant Explicit whether or not an
		// argument is present. React's own pass documentation states the opposite; the
		// executable is the authority and it was measured. See the note on the rule.
		{
			name:   "bareReturnStatementIsExplicit",
			source: "function Component() {\n  const x = useMemo(() => { return; }, []);\n  return <div>{x}</div>;\n}\n",
		},
		// A concise arrow body lowers to an Implicit return terminal, which counts.
		{
			name:   "conciseArrowBody",
			source: "function Component() {\n  const x = useMemo(() => 1, []);\n  return <div>{x}</div>;\n}\n",
		},
		// ANY return terminal suffices; there is no all-paths requirement. Measured clean
		// even though the callback returns undefined whenever `a` is falsy.
		{
			name:   "conditionalReturnOnOnePath",
			source: "function Component({a}) {\n  const x = useMemo(() => { if (a) { return 1; } }, [a]);\n  return <div>{x}</div>;\n}\n",
		},
		// One level of indirection makes the whole rule silent. Upstream keys its function
		// table on the temporary the FunctionExpression produced, and a `const` interposes a
		// store and a load whose identifiers differ, so the lookup misses. This is the case
		// that proves the rule only ever sees a syntactically inline callback.
		{
			name:   "callbackBoundThroughAVariable",
			source: "function Component() {\n  const cb = () => { foo(); };\n  const x = useMemo(cb, []);\n  return <div>{x}</div>;\n}\n",
		},
		// Same indirection on the callee side.
		{
			name:   "useMemoAliasedThroughAVariable",
			source: "function Component() {\n  const m = useMemo;\n  const x = m(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
		},
		// And on the namespace object.
		{
			name:   "reactAliasedThroughAVariable",
			source: "function Component() {\n  const R = React;\n  const x = R.useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
		},
		// No FunctionExpression means no callback to judge.
		{
			name:   "nonFunctionFirstArgument",
			source: "function Component() {\n  const x = useMemo(1, []);\n  return <div>{x}</div>;\n}\n",
		},
		// A spread in first position is declined explicitly by both implementations.
		{
			name:   "spreadArgument",
			source: "function Component({args}) {\n  const x = useMemo(...args);\n  return <div>{x}</div>;\n}\n",
		},
		// `args.length === 0` exits before anything else.
		{
			name:   "noArgumentsAtAll",
			source: "function Component() {\n  const x = useMemo();\n  return <div>{x}</div>;\n}\n",
		},
		// A binding INSIDE the compiled function is a local, not a global, so neither
		// diagnostic fires. Contrast the module-scope case, which reports.
		{
			name:   "functionScopeShadowOfUseMemo",
			source: "function Component() {\n  const useMemo = (f) => f;\n  useMemo(() => { foo(); }, []);\n  return <div />;\n}\n",
		},
		// A shadow in a BLOCK that encloses the call site covers it, which is finer-grained than
		// "does this function declare the name". Added after a mutation dropping blocks from the
		// scope walk survived; both shapes measured clean on React before either was written.
		{
			name:   "blockScopeShadowEnclosingTheCall",
			source: "function Component() {\n  if (true) {\n    const useMemo = (f) => f();\n    useMemo(() => { foo(); }, []);\n  }\n  return <div />;\n}\n",
		},
		{
			name:   "loopBodyShadowEnclosingTheCall",
			source: "function Component({items}) {\n  for (const i of items) {\n    const useMemo = (f) => f();\n    useMemo(() => { foo(); }, []);\n  }\n  return <div />;\n}\n",
		},
		// A FUNCTION DECLARATION shadows as well as a variable does. Added after a mutation
		// ignoring function declarations in the shadow test survived the whole set.
		{
			name:   "functionDeclarationShadowInComponent",
			source: "function Component() {\n  function useMemo(f) { return f(); }\n  const x = useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "functionDeclarationShadowInBlock",
			source: "function Component() {\n  {\n    function useMemo(f) { return f(); }\n    useMemo(() => { foo(); }, []);\n  }\n  return <div />;\n}\n",
		},
		// A parameter shadows the same way.
		{
			name:   "parameterShadowOfUseMemo",
			source: "function Component({useMemo}) {\n  const x = useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
		},
		// An enclosing FUNCTION's binding shadows, which is what draws the line at module
		// scope rather than at the component boundary.
		{
			name:   "outerFunctionScopeShadow",
			source: "function outer() {\n  const useMemo = (f) => f();\n  function Component() {\n    const x = useMemo(() => { foo(); }, []);\n    return <div>{x}</div>;\n  }\n  return Component;\n}\n",
		},
		// The React tracking is a global-binding test too.
		{
			name:   "localReactObject",
			source: "function Component() {\n  const React = {useMemo: (f) => f()};\n  const x = React.useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
		},
		// PropertyLoad matches a string property literal, not a computed access.
		{
			name:   "computedUseMemoProperty",
			source: "function Component() {\n  const x = React['useMemo'](() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
		},
		// Only a direct property of a global named React counts.
		{
			name:   "deeperMemberChain",
			source: "function Component() {\n  const x = A.React.useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
		},
		// An optional MEMBER access does not, unlike an optional call. Measured both ways.
		{
			name:   "optionalMemberAccess",
			source: "function Component() {\n  const x = React?.useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
		},
		// A DIFFERENT property of React is not useMemo. Added after a mutation dropping the
		// property-name test survived the whole fixture set: without this case nothing distinguished
		// `React.useMemo` from `React.anything`. Measured clean on React before it was written.
		{
			name:   "reactUseCallbackIsNotReactUseMemo",
			source: "import * as React from 'react';\nfunction Component() {\n  const x = React.useCallback(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
		},
		{
			name:   "reactMemoIsNotReactUseMemo",
			source: "function Component() {\n  const x = React.memo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
		},
		// The name test is exactly `useMemo`.
		{
			name:   "useCallbackIsNotUseMemo",
			source: "function Component() {\n  const x = useCallback(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
		},
		// The test is on the written name, not on what it resolves to, so an alias escapes.
		{
			name:   "aliasedImport",
			source: "import {useMemo as useM} from 'react';\nfunction Component() {\n  const x = useM(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
		},
		// Assigning to a never-read variable is CLEAN. `result is unused` is not liveness:
		// the store consumes the value. This is the case that decides the rule needs no
		// dataflow, and it would be the natural wrong guess.
		{
			name:   "resultAssignedButNeverRead",
			source: "function Component() {\n  const x = useMemo(() => { return 1; }, []);\n  return <div />;\n}\n",
		},
		// `void` emits an instruction taking the value as an operand, so it consumes it.
		// Semantically this DISCARDS the result and upstream is arguably wrong to be silent,
		// but it is silent and that is reproduced.
		{
			name:   "resultDiscardedByVoidOperator",
			source: "function Component() {\n  void useMemo(() => { return 1; }, []);\n  return <div />;\n}\n",
		},
		// The right operand becomes the sequence's value, so it is consumed. Pairs with the
		// left-operand case, which reports.
		{
			name:   "resultAsCommaRightOperand",
			source: "function Component() {\n  let y;\n  y = (0, useMemo(() => { return 1; }, []));\n  return <div />;\n}\n",
		},
		// Any consumer suffices.
		{
			name:   "resultUsedAsCallArgument",
			source: "function Component() {\n  foo(useMemo(() => { return 1; }, []));\n  return <div />;\n}\n",
		},
		// Including a member access whose own result is then discarded.
		{
			name:   "resultUsedByMemberAccess",
			source: "function Component() {\n  useMemo(() => { return 1; }, []).toString();\n  return <div />;\n}\n",
		},
		// The terminal operand sweep is what catches this one upstream.
		{
			name:   "resultReturnedDirectly",
			source: "function Component() {\n  return useMemo(() => { return 1; }, []);\n}\n",
		},
		// React compiles components and hooks only. A lowercase name is never compiled, and
		// this gate is most of what keeps the rule off ordinary code.
		{
			name:   "voidCallbackInPlainFunction",
			source: "function widget() {\n  const x = useMemo(() => { foo(); }, []);\n  return x;\n}\n",
		},
		// The component test is a leading ASCII capital.
		{
			name:   "underscorePrefixedName",
			source: "function _Private() {\n  const x = useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
		},
		// React's component regex is ASCII, so an accented capital is not a component.
		// The shelf's IsLikelyComponentName uses unicode.IsUpper and would answer true.
		{
			name:   "nonAsciiCapitalName",
			source: "function \u00c9omponent() {\n  const x = useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
		},
		// `use` must be followed by an uppercase letter or a digit.
		{
			name:   "lowercaseUseNameIsNotAHook",
			source: "function usething() {\n  const x = useMemo(() => { foo(); }, []);\n  return x;\n}\n",
		},
		// Measured clean. This package's isHookIdentifierName answers TRUE for bare `use`,
		// which is right for rules-of-hooks and wrong for this driver gate, so this rule
		// requires the uppercase-or-digit suffix. Pinned because the two disagree.
		{
			name:   "bareUseIsNotAHookHere",
			source: "function use() {\n  const x = useMemo(() => { foo(); }, []);\n  return x;\n}\n",
		},
		// The program traversal skips into a function only until it accepts one.
		{
			name:   "nestedComponentIsNeverOffered",
			source: "function outer() {\n  function Component() {\n    const x = useMemo(() => { foo(); }, []);\n    return <div>{x}</div>;\n  }\n  return Component;\n}\n",
		},
		// Class declarations are skipped outright.
		{
			name:   "classMethodIsNeverOffered",
			source: "class C {\n  Component() {\n    const x = useMemo(() => { foo(); }, []);\n    return <div>{x}</div>;\n  }\n}\n",
		},
		// No enclosing function means no compiled unit.
		{
			name:   "moduleScopeCall",
			source: "const x = useMemo(() => { foo(); }, []);\n",
		},
		// Three or more parameters is never a component.
		{
			name:   "threeParameters",
			source: "function Component(a, b, c) {\n  const x = useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
		},
		// A rest element in first position disqualifies.
		{
			name:   "restParameter",
			source: "function Component(...props) {\n  const x = useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
		},
		// A second parameter must mention `ref`.
		{
			name:   "secondParameterNotNamedRef",
			source: "function Component(props, other) {\n  const x = useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
		},
		// `returnsNonNode`: a function whose last-visited return hands back an object is
		// not a component.
		{
			name:   "returnsAnObject",
			source: "function Component() {\n  const x = useMemo(() => { foo(); }, []);\n  return {x};\n}\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, VoidUseMemo, "component.tsx", testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// Spans are asserted because `ExpectFindings` sees message ids and counts and nothing else.
//
// The two diagnostics point at different things, and both were byte-measured on React through the
// Linter API before this rule was written. The void-return finding points at the CALLBACK; the
// unused-result finding points at the CALLEE, which is a bare identifier for `useMemo(...)` and the
// whole member expression for `React.useMemo(...)`. oxc's snapshot independently confirms the
// callback span, logging column 25 for `() => {}` on its own fail case.
//
// Parentheses are transparent in both positions and the span excludes them, which the port brief
// makes a mandatory measurement rather than an assumption. Measured on React: the parenthesized
// callback reports at column 22 rather than 21, and the parenthesized callee at column 4 rather
// than 3.
func TestVoidUseMemoSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "voidReturnPointsAtArrowCallback",
			source: "function Component() {\n  const x = useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
			want:   "() => { foo(); }",
		},
		{
			name:   "voidReturnPointsAtFunctionExpression",
			source: "function Component() {\n  const x = useMemo(function compute() { foo(); }, []);\n  return <div>{x}</div>;\n}\n",
			want:   "function compute() { foo(); }",
		},
		{
			name:   "voidReturnExcludesParenthesesAroundCallback",
			source: "function Component() {\n  const x = useMemo((() => { foo(); }), []);\n  return <div>{x}</div>;\n}\n",
			want:   "() => { foo(); }",
		},
		{
			name:   "unusedResultPointsAtBareCallee",
			source: "function Component() {\n  useMemo(() => { return 1; }, []);\n  return <div />;\n}\n",
			want:   "useMemo",
		},
		{
			name:   "unusedResultPointsAtWholeMemberCallee",
			source: "import * as React from 'react';\nfunction Component() {\n  React.useMemo(() => { return 1; }, []);\n  return <div />;\n}\n",
			want:   "React.useMemo",
		},
		{
			name:   "unusedResultExcludesParenthesesAroundCallee",
			source: "function Component() {\n  (useMemo)(() => { return 1; }, []);\n  return <div />;\n}\n",
			want:   "useMemo",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, VoidUseMemo, "component.tsx", testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected exactly one finding, got %d", len(result.Diagnostics))
			}
			finding := result.Diagnostics[0]
			reported := testCase.source[finding.Range.Pos():finding.Range.End()]
			if reported != testCase.want {
				t.Errorf("finding points at %q, want %q", reported, testCase.want)
			}
		})
	}
}

// `rule.Message` is `{Id, Description}` with no interpolation layer, so there is no rendered text
// to guard and the assertion is on the two fields directly. Literals are typed here rather than
// compared against the rule's own constants, which would move together with the code.
func TestVoidUseMemoMessages(t *testing.T) {
	t.Parallel()

	if messageUseMemoCallbackReturnsNothing.Id != "useMemoCallbackReturnsNothing" {
		t.Errorf("void-return id is %q", messageUseMemoCallbackReturnsNothing.Id)
	}
	if messageUseMemoResultUnused.Id != "useMemoResultUnused" {
		t.Errorf("unused-result id is %q", messageUseMemoResultUnused.Id)
	}
	if !strings.Contains(messageUseMemoCallbackReturnsNothing.Description, "does not return a value") {
		t.Errorf("void-return description does not say what is wrong: %q", messageUseMemoCallbackReturnsNothing.Description)
	}
	if !strings.Contains(messageUseMemoResultUnused.Description, "nothing reads it") {
		t.Errorf("unused-result description does not say what is wrong: %q", messageUseMemoResultUnused.Description)
	}
}

// Both findings can arrive from one function, and their order is asserted rather than only their
// count.
//
// Upstream collects unused-result findings into a separate list and flushes it after the whole
// walk, so internally every void-return finding precedes every unused-result one. That order is
// NOT observable: measured on React, the two findings come back in source order in both
// arrangements, because the reporting layer sorts by position before handing them over. The
// collection order was asserted here first and the measurement corrected it, which is why the
// claim is now about source order.
func TestVoidUseMemoReportsBothKindsInSourceOrder(t *testing.T) {
	t.Parallel()

	unusedFirst := "function Component() {\n  useMemo(() => { return 1; }, []);\n  const x = useMemo(() => { foo(); }, []);\n  return <div>{x}</div>;\n}\n"
	result := rule_testing.Run(t, VoidUseMemo, "component.tsx", unusedFirst)
	rule_testing.ExpectFindings(t, result, "useMemoResultUnused", "useMemoCallbackReturnsNothing")

	voidFirst := "function Component() {\n  const x = useMemo(() => { foo(); }, []);\n  useMemo(() => { return 1; }, []);\n  return <div>{x}</div>;\n}\n"
	result = rule_testing.Run(t, VoidUseMemo, "component.tsx", voidFirst)
	rule_testing.ExpectFindings(t, result, "useMemoCallbackReturnsNothing", "useMemoResultUnused")
}
