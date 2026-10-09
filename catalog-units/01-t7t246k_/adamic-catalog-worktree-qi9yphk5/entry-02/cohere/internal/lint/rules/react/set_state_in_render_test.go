package react

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// The corpus for this rule is React's own, not oxc's.
//
// oxc's linter rule file carries a single pass case and a single fail case, because its body is
// `run_react_compiler_rule(ctx, ErrorCategory::RenderSetState)` and the judgment lives in the
// compiler crate. React ships ten fixtures whose golden expects `Cannot call setState during
// render`, vendored at `internal/react_conformance/testdata/fixtures`, and every source string below
// is one of them byte for byte. They were copied programmatically and are byte-verified against
// their vendored originals by TestFixturesMatchTheVendoredCorpusByte.
//
// # Four fixtures whose name is not their verdict
//
// Fourteen vendored fixtures match `set.?state` by name. Only ten expect this rule's diagnostic.
// Three expect `Calling setState from useMemo`, which is this validator's OTHER arm, and
// `error.invalid-hoisting-setstate` expects `Cannot access variable before it is declared`, which
// is a different pass entirely. Selecting by filename would have imported a corpus a quarter of
// which asserts something this rule never emits.
//
// # The missing import, which is the whole story of this rule's fixtures
//
// Nine of the ten do not import React. Under the compiler harness that does not matter, because
// React seeds a global table naming `useState` and infers the setter's type by unification from
// there. This port asks the type checker instead, so an unimported `useState` is an undeclared
// global that resolves to nothing and the entire file is `any`.
//
// That is not recoverable and it is not a defect. Measured on the exact fixture body: as written,
// `setX` has a nil type alias and type `any`; with `import {useState} from 'react';` prepended and
// nothing else changed, the same identifier has alias `Dispatch` and type
// `Dispatch<SetStateAction<number>>`. One line of input decides it.
//
// So each fixture is recorded TWICE, and both halves are real assertions rather than one being a
// workaround:
//
//   - `TestSetStateInRenderStaysSilentOnTheCorpusAsWritten` runs the body verbatim and asserts
//     silence. That is the honest verdict for a file where the setter has no type, and it is a
//     genuine guard: a later change that starts guessing from the NAME `setX` rather than from the
//     type would turn these green cases red, which is exactly the divergence worth catching.
//   - `TestSetStateInRenderFires` prepends one import line and asserts the finding. The body is
//     untouched; the prepend is stated in the helper rather than baked into the constant, so the
//     verbatim text stays comparable to upstream.
//
// The pass-rate on the corpus as written is therefore the wrong instrument for this rule, and
// saying so is part of the port rather than an excuse: the trade is a dependency on `@types/react`
// and typed code in exchange for not writing 1,450 lines of unification, and it is a good trade for
// Kirk's tree and a bad one for a conformance score.
//
// # Why a local declaration file rather than the real @types/react
//
// `rule_testing`'s tsconfig sets `types: []` and points at a temp directory, deliberately, so that a
// fixture cannot pick up whatever happens to be installed near the test. There is no `node_modules`
// to resolve. `reactStateDeclarations` below is the minimum of `@types/react` this rule reads, and
// the two declarations that matter are copied in the shape upstream writes them: `Dispatch` is a
// TYPE ALIAS and `RefObject` is an INTERFACE. That difference is the rule's whole predicate, so a
// stand-in that made both aliases would test a checker behaviour that does not exist.
const (
	unconditionalSetState     = "// @validateNoSetStateInRender\nfunction Component(props) {\n  const [x, setX] = useState(0);\n  const aliased = setX;\n\n  setX(1);\n  aliased(2);\n\n  return x;\n}\n"
	unboundState              = "function Component(props) {\n  // Intentionally don't bind state, this repros a bug where we didn't\n  // infer the type of destructured properties after a hole in the array\n  let [, setState] = useState();\n  setState(1);\n  return props.foo;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: ['TodoAdd'],\n  isComponent: 'TodoAdd',\n};\n"
	keyedState                = "// @validateNoSetStateInRender @enableUseKeyedState\nimport {useState} from 'react';\n\nfunction Component() {\n  const [total, setTotal] = useState(0);\n  setTotal(42);\n  return total;\n}\n\nexport const FIXTURE_ENTRYPOINT = {\n  fn: Component,\n  params: [],\n  isComponent: true,\n};\n"
	afterLoop                 = "// @validateNoSetStateInRender\nfunction Component(props) {\n  const [state, setState] = useState(false);\n  for (const _ of props) {\n  }\n  setState(true);\n  return state;\n}\n"
	afterLoopBreak            = "// @validateNoSetStateInRender\nfunction Component(props) {\n  const [state, setState] = useState(false);\n  for (const _ of props) {\n    if (props.cond) {\n      break;\n    } else {\n      continue;\n    }\n  }\n  setState(true);\n  return state;\n}\n"
	loopThrow                 = "// @validateNoSetStateInRender\nfunction Component(props) {\n  const [state, setState] = useState(false);\n  for (const _ of props) {\n    if (props.cond) {\n      break;\n    } else {\n      throw new Error('bye!');\n    }\n  }\n  setState(true);\n  return state;\n}\n"
	lambda                    = "// @validateNoSetStateInRender\nfunction Component(props) {\n  const [x, setX] = useState(0);\n\n  const foo = () => {\n    setX(1);\n  };\n  foo();\n\n  return [x];\n}\n"
	nestedFunctionExpressions = "// @validateNoSetStateInRender\nfunction Component(props) {\n  const [x, setX] = useState(0);\n\n  const foo = () => {\n    setX(1);\n  };\n\n  const bar = () => {\n    foo();\n  };\n\n  const baz = () => {\n    bar();\n  };\n  baz();\n\n  return [x];\n}\n"
	hookReturn                = "// @validateNoSetStateInRender @enableTreatSetIdentifiersAsStateSetters\nfunction Component() {\n  const [state, setState] = useCustomState(0);\n  const aliased = setState;\n\n  setState(1);\n  aliased(2);\n\n  return state;\n}\n\nfunction useCustomState(init) {\n  return useState(init);\n}\n"
	propInRender              = "// @validateNoSetStateInRender @enableTreatSetIdentifiersAsStateSetters\nfunction Component({setX}) {\n  const aliased = setX;\n\n  setX(1);\n  aliased(2);\n\n  return x;\n}\n"
)

// reactStateDeclarations is the part of `@types/react` this rule reads.
//
// `Dispatch` is declared as a type alias and `RefObject` as an interface, matching
// `@types/react/index.d.ts` at lines 1645 and 154 respectively. The rule keys on the type's ALIAS,
// which is populated for the first and nil for the second, so writing `RefObject` as an alias here
// would make a passing test out of a predicate that cannot work on real code.
//
// `useReducer` is present so the negative control for `ActionDispatch` has something to resolve.
const reactStateDeclarations = `declare module 'react' {
  export type SetStateAction<S> = S | ((prev: S) => S);
  export type Dispatch<A> = (value: A) => void;
  export interface RefObject<T> { current: T }
  export function useState<S>(initial?: S): [S, Dispatch<SetStateAction<S>>];
  export function useRef<T>(initial: T): RefObject<T>;
  export type ActionDispatch<A extends any[]> = (...args: A) => void;
  export function useReducer<S, A>(reducer: (state: S, action: A) => S, initial: S): [S, ActionDispatch<[A]>];
  export function useMemo<T>(compute: () => T, deps: unknown[]): T;
  export function useCallback<T>(callback: T, deps: unknown[]): T;
  export function useEffect(effect: () => void, deps?: unknown[]): void;
}`

// reactImport is the one line prepended to a verbatim fixture so its setter has a type.
//
// Kept as its own constant so that every use of it is greppable and so the fixture constants stay
// byte-comparable against the vendored originals.
const reactImport = "import {useState} from 'react';\n"

// runSetStateFixture runs one source against the rule with the React declarations available.
func runSetStateFixture(t *testing.T, source string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, SetStateInRender, map[string]string{
		"react.d.ts":  reactStateDeclarations,
		"fixture.tsx": source,
	}, "fixture.tsx")
}

// TestSetStateInRenderFires runs each vendored fixture with one import line prepended and asserts
// the finding count React's own rule produces on the same input.
//
// Every `want` here was established by running `eslint-plugin-react-hooks` version 7.1.1 through the
// ESLint Linter API on the exact string being asserted, not by reading the golden. That matters
// because the goldens come from the compiler harness, which applies pragmas the lint rule has no
// option surface for: two of the ten expect an error that the shipped rule does not produce.
func TestSetStateInRenderFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		// prependImport is false only for the one fixture that already imports React. Prepending to
		// that one produces `Identifier 'useState' has already been declared`, which was measured
		// rather than guessed.
		prependImport bool
		source        string
		want          []string
	}{
		{
			// The canonical case, and the one that shows an alias is followed: `setX(1)` reports at
			// `setX` and `aliased(2)` reports at `aliased`, four characters and seven.
			name:          "a setter and its alias, both called unconditionally",
			prependImport: true,
			source:        unconditionalSetState,
			want:          []string{"setStateInRender", "setStateInRender"},
		},
		{
			// A destructuring hole. `let [, setState] = useState()` gives element one a type even
			// though element zero is skipped, which upstream's own comment says was once a bug in
			// their inference and is free here.
			name:          "a setter destructured after a hole",
			prependImport: true,
			source:        unboundState,
			want:          []string{"setStateInRender"},
		},
		{
			// Already imports React, so it is run untouched.
			name:          "a setter called unconditionally, with the import already present",
			prependImport: false,
			source:        keyedState,
			want:          []string{"setStateInRender"},
		},
		{
			// The three loop shapes. All three report, and they are the reason this rule needs post
			// dominance: the call after the loop is syntactically identical in each and the graphs
			// differ.
			name:          "a setter after a loop that always terminates",
			prependImport: true,
			source:        afterLoop,
			want:          []string{"setStateInRender"},
		},
		{
			name:          "a setter after a loop whose arms break and continue",
			prependImport: true,
			source:        afterLoopBreak,
			want:          []string{"setStateInRender"},
		},
		{
			// The case that fixes the exit set. A throwing arm is NOT an exit, so the break is the
			// only way out and the tail is unconditional. Counting throws as exits makes this clean
			// and disagrees with React.
			name:          "a setter after a loop whose other arm throws",
			prependImport: true,
			source:        loopThrow,
			want:          []string{"setStateInRender"},
		},
		{
			// The finding lands on `foo`, not on `setX`. A function proven to set state
			// unconditionally makes its own call sites report.
			name:          "a setter called through a lambda that is then called",
			prependImport: true,
			source:        lambda,
			want:          []string{"setStateInRender"},
		},
		{
			// Three levels of wrapper, and the finding lands on `baz`.
			name:          "a setter called through three nested function expressions",
			prependImport: true,
			source:        nestedFunctionExpressions,
			want:          []string{"setStateInRender"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := testCase.source
			if testCase.prependImport {
				source = reactImport + source
			}
			result := runSetStateFixture(t, source)
			rule_testing.ExpectFindings(t, result, testCase.want...)
		})
	}
}

// TestSetStateInRenderStaysSilentOnFixturesTheShippedRuleDoesNotReport records the two vendored
// fixtures whose golden expects this diagnostic and whose shipped rule does not produce it.
//
// Both are gated on `@enableTreatSetIdentifiersAsStateSetters`, a compiler pragma that treats any
// callee whose NAME starts with `set` as a setter. It defaults to false, the ESLint rule exposes no
// option to turn it on, and running both fixtures through the shipped rule produces nothing.
//
// They are kept as silent cases rather than deleted, because the alternative readings are both
// wrong. Deleting them loses the record that the goldens and the shipped rule disagree here.
// Asserting a finding would mean implementing a name heuristic that upstream turns off by default,
// and that heuristic is the thing this port replaced with a type test: `function Component({setX})`
// takes an untyped prop, so no amount of checking recovers a setter from it. This is the honest
// answer for both implementations and for different reasons.
func TestSetStateInRenderStaysSilentOnFixturesTheShippedRuleDoesNotReport(t *testing.T) {
	t.Parallel()

	// `function Component({setX})` with no type annotation. The prop is `any`, and a setter arriving
	// as an untyped prop is unrecoverable by any type-directed analysis. Upstream only catches it by
	// matching the NAME, under a pragma that defaults to off and that the ESLint rule exposes no
	// option for. Measured: the shipped rule produces nothing on this input, with or without the
	// import, so we and React agree here and only the compiler golden differs.
	rule_testing.ExpectClean(t, runSetStateFixture(t, reactImport+propInRender))
}

// TestSetStateInRenderReportsWhereTheCheckerBeatsReactsInference records a DIVERGENCE, in the
// direction of finding a real defect React's shipped rule misses.
//
// The vendored golden for this fixture expects two findings. React's shipped ESLint rule produces
// ZERO, measured both as written and with the import prepended, because its own inference does not
// follow a custom hook's return type without `@enableTreatSetIdentifiersAsStateSetters` — a compiler
// pragma with no ESLint option surface. The golden was produced by the compiler harness with that
// pragma on.
//
// This port produces TWO, matching the golden, and it does so with no pragma and no name heuristic:
// `useCustomState` is declared in the same file, so once `useState` resolves the checker infers the
// hook's return type and `Dispatch` propagates to its caller. That is the capability the task's
// research pass measured and the reason this rule dropped a tier.
//
// It is recorded as its own test rather than folded into the Fires table because it is the one
// input where this implementation and the reference implementation disagree, and a future reader
// comparing the two needs to find the disagreement rather than infer it from a count. The
// differential against oxlint will show a difference here and it is the right difference: the code
// really does set state unconditionally during render.
//
// Both calls report, `setState` and its alias `aliased`, for the same reason as the canonical case.
func TestSetStateInRenderReportsWhereTheCheckerBeatsReactsInference(t *testing.T) {
	t.Parallel()

	result := runSetStateFixture(t, reactImport+hookReturn)
	rule_testing.ExpectFindings(t, result, "setStateInRender", "setStateInRender")
}

// TestSetStateInRenderStaysSilentOnTheCorpusAsWritten runs every vendored fixture byte for byte,
// with nothing prepended, and asserts silence.
//
// This is a real guard rather than a concession. Nine of the ten do not import React, so the setter
// has no type and this rule correctly declines. A later change that started guessing from the name
// `setX` — which is exactly what upstream's optional pragma does — would turn these red, and that
// is the divergence most worth catching, because it would silently widen the rule on every file in
// the tree that happens to name a callback `setSomething`.
func TestSetStateInRenderStaysSilentOnTheCorpusAsWritten(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{name: "unconditional set state", source: unconditionalSetState},
		{name: "unbound state", source: unboundState},
		{name: "after loop", source: afterLoop},
		{name: "after loop break", source: afterLoopBreak},
		{name: "loop throw", source: loopThrow},
		{name: "lambda", source: lambda},
		{name: "nested function expressions", source: nestedFunctionExpressions},
		{name: "hook return", source: hookReturn},
		{name: "prop in render", source: propInRender},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runSetStateFixture(t, testCase.source))
		})
	}
}

// TestSetStateInRenderStaysSilent holds the clean cases, each one measured on React's own rule
// rather than reasoned about.
//
// Upstream's linter corpus contributes exactly one of these, the `onClick` handler, because oxc's
// rule file is a dispatcher with a single pass case. The rest come from probing the shipped rule on
// inputs that separate this rule's decisions from the ones a reader would assume it makes.
func TestSetStateInRenderStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		{
			// oxc's only pass case, verbatim apart from the leading newline its tester adds.
			name: "a setter called from an event handler",
			source: "import {useState} from 'react';\n" +
				"function Component(props) {\n" +
				"  const [state, setState] = useState(0);\n" +
				"  return <div onClick={() => setState(state + 1)}>{state}</div>;\n" +
				"}\n",
		},
		{
			// The call is conditional, so it costs a double render rather than looping. Upstream
			// deliberately says nothing, and flagging it would report most legitimate
			// derived-state-from-props code.
			name: "a setter called inside an if",
			source: reactImport + "function Component(props: {c: boolean}) {\n" +
				"  const [x, setX] = useState(0);\n" +
				"  if (props.c) { setX(1); }\n" +
				"  return x;\n" +
				"}\n",
		},
		{
			// The distinguishing case for post dominance versus plain dominance. The entry block
			// DOMINATES the block holding the call, so a forward-dominance test reports here and
			// React does not. Measured clean on React's rule.
			name: "a setter called after an early return",
			source: reactImport + "function Component(props: {c: boolean}) {\n" +
				"  const [x, setX] = useState(0);\n" +
				"  if (props.c) { return null; }\n" +
				"  setX(1);\n" +
				"  return x;\n" +
				"}\n",
		},
		{
			// A try body may not complete, so the call is not on every path.
			name: "a setter called inside a try",
			source: reactImport + "function Component() {\n" +
				"  const [x, setX] = useState(0);\n" +
				"  try { setX(1); } catch (e) {}\n" +
				"  return x;\n" +
				"}\n",
		},
		{
			// A loop body is conditional, however the loop is written.
			name: "a setter called inside a loop body",
			source: reactImport + "function Component(props: number[]) {\n" +
				"  const [x, setX] = useState(0);\n" +
				"  for (const p of props) { setX(1); }\n" +
				"  return x;\n" +
				"}\n",
		},
		{
			// An optional call is silent upstream. Reproduced deliberately; see the rule comment.
			name: "a setter called with an optional call",
			source: reactImport + "function Component() {\n" +
				"  const [x, setX] = useState(0);\n" +
				"  setX?.(1);\n" +
				"  return x;\n" +
				"}\n",
		},
		{
			// Reached through a property rather than a local, which upstream never tracks.
			name: "a setter reached through a tuple index",
			source: reactImport + "function Component() {\n" +
				"  const s = useState(0);\n" +
				"  s[1](1);\n" +
				"  return s[0];\n" +
				"}\n",
		},
		{
			// Same, through an object property.
			name: "a setter reached through an object property",
			source: reactImport + "function Component() {\n" +
				"  const [x, setX] = useState(0);\n" +
				"  const o = {go: setX};\n" +
				"  o.go(1);\n" +
				"  return x;\n" +
				"}\n",
		},
		{
			// The setter is a value here, never called.
			name: "a setter returned rather than called",
			source: reactImport + "function Component() {\n" +
				"  const [x, setX] = useState(0);\n" +
				"  return setX;\n" +
				"}\n",
		},
		{
			// `useReducer` gives `ActionDispatch`, a different alias. This is the negative control
			// for the predicate: without it, a widening of the alias test to any dispatch-shaped
			// type would go unnoticed.
			name: "a reducer dispatch called during render",
			source: "import {useReducer} from 'react';\n" +
				"function Component() {\n" +
				"  const [n, dispatch] = useReducer((s: number, a: number) => s, 0);\n" +
				"  dispatch(1);\n" +
				"  return n;\n" +
				"}\n",
		},
		{
			// A structurally identical function with no alias. `Dispatch<A>` IS `(value: A) => void`
			// structurally, so this is what proves the predicate is nominal rather than structural.
			// Without this case a change from the alias to a structural signature test would pass
			// every other fixture here.
			name: "a plain one-argument function called during render",
			source: "function Component() {\n" +
				"  const plain: (value: number) => void = (v) => {};\n" +
				"  plain(1);\n" +
				"  return null;\n" +
				"}\n",
		},
		{
			// A name starting with `set` that is not a setter. This is precisely what upstream's
			// optional pragma would report and what the default does not.
			name: "a function named like a setter that is not one",
			source: "function Component() {\n" +
				"  const setThing = (v: number) => {};\n" +
				"  setThing(1);\n" +
				"  return null;\n" +
				"}\n",
		},
		{
			// Not a component and not a hook, so it is not a compilation unit and nothing in it is
			// analysed.
			name: "a setter called in a lowercase function",
			source: reactImport + "function helper() {\n" +
				"  const [x, setX] = useState(0);\n" +
				"  setX(1);\n" +
				"  return x;\n" +
				"}\n",
		},
		{
			// Three parameters is never a component shape.
			name: "a setter called in a capitalized function taking three parameters",
			source: reactImport + "function Component(a: number, b: number, c: number) {\n" +
				"  const [x, setX] = useState(0);\n" +
				"  setX(1);\n" +
				"  return x;\n" +
				"}\n",
		},
		{
			// A second parameter must name a ref. Measured: `(props, ref)` reports and
			// `(props, other)` does not.
			name: "a setter called in a component whose second parameter is not a ref",
			source: reactImport + "function Component(props: {a: number}, other: number) {\n" +
				"  const [x, setX] = useState(0);\n" +
				"  setX(1);\n" +
				"  return x;\n" +
				"}\n",
		},
		{
			// A rest parameter takes the function out of the gate entirely.
			name: "a setter called in a component with a rest parameter",
			source: reactImport + "function Component(...props: number[]) {\n" +
				"  const [x, setX] = useState(0);\n" +
				"  setX(1);\n" +
				"  return x;\n" +
				"}\n",
		},
		{
			// A `useCallback` callback is never invoked during render, so its body does not run
			// there. This is the one clean case most likely to look like a miss: `useMemo` with the
			// same body reports. See the rule comment for the mechanism.
			name: "a setter called inside a useCallback callback",
			source: "import {useState, useCallback} from 'react';\n" +
				"function Component() {\n" +
				"  const [x, setX] = useState(0);\n" +
				"  const cb = useCallback(() => { setX(1); }, []);\n" +
				"  return cb;\n" +
				"}\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runSetStateFixture(t, testCase.source))
		})
	}
}

// TestSetStateInRenderFiresInsideAHook records that the subject is any compiled function, not only
// a component.
//
// The rule's NAME says render and a hook is not render, but React compiles hooks too and reports
// there. Measured on React's rule: this exact input reports. Kept as its own test because it is the
// clearest case of the name being narrower than the behaviour.
func TestSetStateInRenderFiresInsideAHook(t *testing.T) {
	t.Parallel()

	result := runSetStateFixture(t, reactImport+
		"export function useThing() {\n"+
		"  const [x, setX] = useState(0);\n"+
		"  setX(1);\n"+
		"  return x;\n"+
		"}\n")
	rule_testing.ExpectFindings(t, result, "setStateInRender")
}

// TestSetStateInRenderFollowsAnAliasChain checks that taint travels any distance.
//
// Measured on React: `const a = setX; const b = a; b(1);` reports at `b`. This is the property
// single-assignment form supplies for free and the reason the rule reads `LoadLocal` and
// `StoreLocal` rather than matching syntax.
func TestSetStateInRenderFollowsAnAliasChain(t *testing.T) {
	t.Parallel()

	result := runSetStateFixture(t, reactImport+
		"function Component() {\n"+
		"  const [x, setX] = useState(0);\n"+
		"  const a = setX;\n"+
		"  const b = a;\n"+
		"  b(1);\n"+
		"  return x;\n"+
		"}\n")
	rule_testing.ExpectFindings(t, result, "setStateInRender")
}

// TestSetStateInRenderFollowsACustomHook checks the capability the checker supplies and upstream
// needs a pragma and a name heuristic to approximate.
//
// With `useCustomState` actually DECLARED, its inferred return type carries `Dispatch` through, so
// the setter is identified with no seed data at all. The vendored fixture for this is silent only
// because it leaves the hook undeclared.
func TestSetStateInRenderFollowsACustomHook(t *testing.T) {
	t.Parallel()

	result := runSetStateFixture(t, reactImport+
		"function useCustomState(init: number) {\n"+
		"  return useState(init);\n"+
		"}\n"+
		"export function Component() {\n"+
		"  const [state, setState] = useCustomState(0);\n"+
		"  setState(1);\n"+
		"  return state;\n"+
		"}\n")
	rule_testing.ExpectFindings(t, result, "setStateInRender")
}

// TestSetStateInRenderReportsInsideUseMemo covers the validator's second arm.
//
// The `useMemo` test comes BEFORE the unconditional test upstream, so a call that is conditional
// within the callback still reports. Both cases measured on React's rule, which produces `Calling
// setState from useMemo may trigger an infinite loop` for each.
func TestSetStateInRenderReportsInsideUseMemo(t *testing.T) {
	t.Parallel()

	t.Run("an unconditional call inside a useMemo callback", func(t *testing.T) {
		t.Parallel()
		result := runSetStateFixture(t, "import {useState, useMemo} from 'react';\n"+
			"function Component() {\n"+
			"  const [x, setX] = useState(0);\n"+
			"  const y = useMemo(() => { setX(1); return 1; }, []);\n"+
			"  return y;\n"+
			"}\n")
		rule_testing.ExpectFindings(t, result, "setStateInUseMemo")
	})

	t.Run("a conditional call inside a useMemo callback still reports", func(t *testing.T) {
		t.Parallel()
		result := runSetStateFixture(t, "import {useState, useMemo} from 'react';\n"+
			"function Component(props: {c: boolean}) {\n"+
			"  const [x, setX] = useState(0);\n"+
			"  const y = useMemo(() => { if (props.c) { setX(1); } return 1; }, []);\n"+
			"  return y;\n"+
			"}\n")
		rule_testing.ExpectFindings(t, result, "setStateInUseMemo")
	})
}

// TestSetStateInRenderBailsOutPerFunctionRatherThanPerFile pins the scope of the parameter-shape
// decline.
//
// A rest parameter takes its own function out of the gate. It would be easy to write that as a
// whole-file bailout by accident, and no fixture with one component in it could tell the difference.
// Measured on React: the second component in this file reports while the first is silent.
func TestSetStateInRenderBailsOutPerFunctionRatherThanPerFile(t *testing.T) {
	t.Parallel()

	result := runSetStateFixture(t, reactImport+
		"export function Bailout(...props: number[]) {\n"+
		"  const [x, setX] = useState(0);\n"+
		"  setX(1);\n"+
		"  return x;\n"+
		"}\n"+
		"export function Control() {\n"+
		"  const [y, setY] = useState(0);\n"+
		"  setY(1);\n"+
		"  return y;\n"+
		"}\n")
	rule_testing.ExpectFindings(t, result, "setStateInRender")
}

// TestSetStateInRenderPointsAtTheCallee asserts WHERE the finding lands, which no message-id
// assertion can see.
//
// Upstream's span is `callee.span`: React's own golden underlines four characters for `setX(1)` and
// seven for `aliased(2)`, the name being called rather than the whole call expression. A rule
// pointing at the call, at the statement, or at the argument satisfies every count assertion above.
//
// The source is sliced the way the HARNESS wrote it rather than the way the literal reads.
// `rule_testing.RunTyped` writes `strings.TrimSpace(contents)+"\n"`, so a fixture carrying leading
// whitespace is one byte offset on disk from the Go string, and slicing the literal reports a span
// shifted by one. That is transformed here rather than worked around.
func TestSetStateInRenderPointsAtTheCallee(t *testing.T) {
	t.Parallel()

	source := reactImport +
		"function Component() {\n" +
		"  const [x, setX] = useState(0);\n" +
		"  const aliased = setX;\n" +
		"  setX(1);\n" +
		"  aliased(2);\n" +
		"  return x;\n" +
		"}\n"
	onDisk := strings.TrimSpace(source) + "\n"

	result := runSetStateFixture(t, source)
	if len(result.Diagnostics) != 2 {
		t.Fatalf("got %d findings, want 2", len(result.Diagnostics))
	}

	want := []string{"setX", "aliased"}
	for index, diagnostic := range result.Diagnostics {
		span := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
		if span != want[index] {
			t.Errorf("finding %d points at %q, want %q", index, span, want[index])
		}
	}
}

// TestSetStateInRenderMessagesSayWhy asserts the message text directly.
//
// `rule.Message` is `{Id, Description}` with no interpolation, so there is nothing to render and
// nothing a format string can corrupt. The assertion is equality against a literal typed here rather
// than a comparison to the rule's own constant, because a comparison to the constant moves with it
// under mutation and passes while the message is wrong.
func TestSetStateInRenderMessagesSayWhy(t *testing.T) {
	t.Parallel()

	if setStateInRenderMessage.Id != "setStateInRender" {
		t.Errorf("render message id = %q, want %q", setStateInRenderMessage.Id, "setStateInRender")
	}
	if setStateInUseMemoMessage.Id != "setStateInUseMemo" {
		t.Errorf("useMemo message id = %q, want %q", setStateInUseMemoMessage.Id, "setStateInUseMemo")
	}
	if !strings.Contains(setStateInRenderMessage.Description, "Too many re-renders") {
		t.Errorf("the render message should name the error React actually throws, got %q",
			setStateInRenderMessage.Description)
	}
	if !strings.Contains(setStateInUseMemoMessage.Description, "useMemo") {
		t.Errorf("the useMemo message should name useMemo, got %q", setStateInUseMemoMessage.Description)
	}
	if setStateInRenderMessage.Description == setStateInUseMemoMessage.Description {
		t.Error("the two messages must not be identical; they describe different defects")
	}
}

// TestSetStateInRenderRequiresTheTypedHarness proves the rule is inert without a checker.
//
// A typed rule handed a nil checker goes silent rather than crashing, because `GetSymbolAtLocation`
// and friends return nil rather than panicking. Silence is the more dangerous failure: every
// StaysSilent case passes vacuously and only the Fires cases notice. This asserts the declaration is
// present so a later revert fails loudly here rather than by turning the whole suite green.
func TestSetStateInRenderRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !SetStateInRender.NeedsTypeChecker {
		t.Fatal("the rule reads the type checker and must declare NeedsTypeChecker")
	}
	// The untyped harness hands the rule a nil checker. The rule must decline rather than panic.
	result := rule_testing.Run(t, SetStateInRender, "fixture.tsx",
		"function Component() {\n  const [x, setX] = useState(0);\n  setX(1);\n  return x;\n}\n")
	rule_testing.ExpectClean(t, result)
}

// TestSetStateInRenderIsRegistered checks the rule reached the catalog under the name the inventory
// writes.
func TestSetStateInRenderIsRegistered(t *testing.T) {
	t.Parallel()

	if SetStateInRender.Name != "react-hooks/set-state-in-render" {
		t.Errorf("rule name = %q, want %q", SetStateInRender.Name, "set-state-in-render")
	}
	found := false
	for _, registration := range rule.Registered() {
		if registration.Rule.Name == SetStateInRender.Name {
			found = true
			break
		}
	}
	if !found {
		t.Error("the rule is not registered, so it lints nothing")
	}
}

// TestFixturesMatchTheVendoredCorpusByte compares every fixture constant against the file it was
// copied from, byte for byte.
//
// Reading them is not enough and this has cost three porters. The failure is not transcription by
// hand, which is obvious when wrong; it is a TOOL silently cooking an escape on the way in, so that
// `  ` becomes real spaces or `\\1` becomes `\1`, and the fixture then sits in the suite
// asserting the opposite of upstream while compiling and going green.
//
// These constants were generated by reading each file and emitting it through a JSON encoder, so no
// escape was ever typed. This test is what proves that held.
func TestFixturesMatchTheVendoredCorpusByte(t *testing.T) {
	t.Parallel()

	cases := []struct {
		constant string
		fileName string
	}{
		{unconditionalSetState, "error.invalid-unconditional-set-state-in-render.js"},
		{unboundState, "error.invalid-setState-in-render-unbound-state.js"},
		{keyedState, "error.invalid-setstate-unconditional-with-keyed-state.js"},
		{afterLoop, "error.unconditional-set-state-in-render-after-loop.js"},
		{afterLoopBreak, "error.unconditional-set-state-in-render-after-loop-break.js"},
		{loopThrow, "error.unconditional-set-state-in-render-with-loop-throw.js"},
		{lambda, "error.unconditional-set-state-lambda.js"},
		{nestedFunctionExpressions, "error.unconditional-set-state-nested-function-expressions.js"},
		{hookReturn, "error.invalid-unconditional-set-state-hook-return-in-render.js"},
		{propInRender, "error.invalid-unconditional-set-state-prop-in-render.js"},
	}

	directory := filepath.Join("conformance", "testdata", "fixtures")
	for _, testCase := range cases {
		t.Run(testCase.fileName, func(t *testing.T) {
			t.Parallel()
			vendored, err := os.ReadFile(filepath.Join(directory, testCase.fileName))
			if err != nil {
				t.Fatalf("reading the vendored fixture: %v", err)
			}
			if string(vendored) != testCase.constant {
				t.Errorf("the constant and the vendored file differ\n constant: %q\n vendored: %q",
					testCase.constant, string(vendored))
			}
		})
	}
	if len(cases) != 10 {
		t.Errorf("checked %d fixtures, want the 10 whose golden expects this rule's diagnostic", len(cases))
	}
}

// TestSetStateInRenderFindsATransitiveSetterCapturedAfterAnother covers the positional capture
// translation, and the input that covers it is not the obvious one.
//
// Written for a mutant that rewrote `nested.Context[index]` to `nested.Context[0]`. The first
// fixture written for it captured a plain value and then a setter, which looked like the
// distinguishing shape and was not: with the mutation that input STILL reports, because a directly
// captured setter is recognized by a second, independent route. Inside the nested function the call
// to `setX` has a real syntax node, so `isStateSetter` asks the checker about it and answers true
// without consulting the translated set at all. The set is a redundant guard for that case.
//
// The set is the ONLY route for a TRANSITIVE capture: a value that is a function proven to set
// state carries no type saying so, so nothing but the translated set can recognize it. Making that
// value the SECOND capture is what separates the two versions, and the separation is total —
// pristine reports one finding, the mutant reports none.
//
// Recorded at this length because the first hypothesis was right about the mechanism and wrong
// about the input, which the port brief names as its own failure mode: a fixture written for a
// survivor is a hypothesis, and the re-score is what tests it.
//
// Measured on React, which reports once at `bar`.
func TestSetStateInRenderFindsATransitiveSetterCapturedAfterAnother(t *testing.T) {
	t.Parallel()

	result := runSetStateFixture(t, reactImport+
		"function Component(props: {a: number}) {\n"+
		"  const [x, setX] = useState(0);\n"+
		"  const other = props.a;\n"+
		"  const foo = () => { setX(1); };\n"+
		"  const bar = () => { const t = other; foo(); return t; };\n"+
		"  bar();\n"+
		"  return x;\n"+
		"}\n")
	rule_testing.ExpectFindings(t, result, "setStateInRender")
}

// TestSetStateInRenderFindsADirectlyCapturedSetterAfterAnother is the case above with the wrapper
// removed, kept because it is the shape a reader expects and because it documents the redundancy.
//
// It reports for a different reason than the transitive one: the checker recognizes `setX` at the
// call site inside the lambda without any capture bookkeeping. Two independent routes reach the same
// verdict here, which is why this input cannot see the index mutation.
func TestSetStateInRenderFindsADirectlyCapturedSetterAfterAnother(t *testing.T) {
	t.Parallel()

	result := runSetStateFixture(t, reactImport+
		"function Component(props: {a: number}) {\n"+
		"  const [x, setX] = useState(0);\n"+
		"  const other = props.a;\n"+
		"  const foo = () => { const t = other; setX(1); return t; };\n"+
		"  foo();\n"+
		"  return x;\n"+
		"}\n")
	rule_testing.ExpectFindings(t, result, "setStateInRender")
}

// TestSetStateInRenderFindsASetterDeclaredInsideAUseMemo covers a defect a MUTATION found and no
// fixture could.
//
// A mutant removing the "does this nested function capture a setter" guard survived the whole suite.
// Hunting for an input that separated the two versions produced this one, and the separating input
// turned out to be a case React REPORTS and this rule did not: the mutant was right and the original
// was wrong. That is the port brief's first survivor category, and it is the reason it says to
// consider it first rather than last.
//
// The mechanism, once seen, is the same one that makes `useMemo` special everywhere else in this
// rule. Upstream replaces a `useMemo` callback with a direct call to itself, so the callback's body
// is inlined into the render graph; a setter declared inside it is an ordinary value of the
// enclosing function and reaches the validator with no capture involved. Every fixture in this file
// reaches its setter through a capture, so none of them could see the guard being wrong.
//
// The three surrounding verdicts were measured at the same time and pin the boundary:
//   - the same body inside `useCallback` is CLEAN, because that callback is not invoked in render
//   - the same body with the call made conditional still REPORTS, because the memoized arm is
//     tested before the unconditional one
//   - a nested plain arrow declaring its own setter is CLEAN, because it is not inlined
func TestSetStateInRenderFindsASetterDeclaredInsideAUseMemo(t *testing.T) {
	t.Parallel()

	t.Run("declared and called inside a useMemo callback", func(t *testing.T) {
		t.Parallel()
		result := runSetStateFixture(t, "import {useState, useMemo} from 'react';\n"+
			"function Component() {\n"+
			"  const y = useMemo(() => { const [a, setA] = useState(1); setA(2); return a; }, []);\n"+
			"  return y;\n"+
			"}\n")
		rule_testing.ExpectFindings(t, result, "setStateInUseMemo")
	})

	t.Run("conditional inside a useMemo callback still reports", func(t *testing.T) {
		t.Parallel()
		result := runSetStateFixture(t, "import {useState, useMemo} from 'react';\n"+
			"function Component(props: {c: boolean}) {\n"+
			"  const y = useMemo(() => { const [a, setA] = useState(1); if (props.c) { setA(2); } return a; }, []);\n"+
			"  return y;\n"+
			"}\n")
		rule_testing.ExpectFindings(t, result, "setStateInUseMemo")
	})

	t.Run("the same body inside a useCallback is clean", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, runSetStateFixture(t, "import {useState, useCallback} from 'react';\n"+
			"function Component() {\n"+
			"  const f = useCallback(() => { const [a, setA] = useState(1); setA(2); }, []);\n"+
			"  return f;\n"+
			"}\n"))
	})

	t.Run("a plain nested arrow declaring its own setter is clean", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectClean(t, runSetStateFixture(t, reactImport+
			"function Component() {\n"+
			"  const f = () => { const [a, setA] = useState(0); setA(1); };\n"+
			"  f();\n"+
			"  return 1;\n"+
			"}\n"))
	})
}
