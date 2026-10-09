package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The imported corpus is twelve cases and every case here was RUN before being written down.
//
// oxc's linter rule ships a tester block of exactly one pass and one fail, and that floor is
// misleading rather than small: `set_state_in_effect.rs` is the twenty-two-rule dispatcher body, so
// its judgment lives in `oxc_react_compiler` and its real corpus is that crate's `fixtures/`
// directory. Eighteen fixtures there name this rule's behaviour and twelve are transcribed below
// verbatim. A porter who trusts the extractor's confident "1 pass, 1 fail" ports a name heuristic.
//
// React ships no `EffectSetState` golden among the 325 vendored at `internal/react_conformance`.
// Searched for both message texts and for the category name; the zero is real, established with
// three controls in the same command that DID fire: 13 files matching the useMemo setState message,
// 21 matching `setState` at all, 14 containing `useEffect`. The `setState`-named goldens there
// belong to `set-state-in-render` and to `use-memo`, which is the "an error-named fixture can be
// exactly the wrong half" trap.
//
// Every non-imported case was measured by running React 7.1.1's own `react-hooks/set-state-in-effect`
// through the ESLint Linter API on that exact input, before it was written here. Reading gave the
// wrong answer twice: on which names count as effects, where the neighbouring `exhaustive-deps`
// regex is a trap, and on a function defined inside the effect body, which reads like it must
// report and does not.
//
// # Why these fixtures carry a react.d.ts and upstream's do not
//
// oxc's compiler fixtures mostly do not import react, so `useState` is a bare undeclared global.
// That is exactly why oxc infers from shape: it has no types to read. This rule reads types, so a
// fixture without the import is not a harder version of the same question, it is a different
// question whose answer is `any`, and the rule is correctly silent on it. Probed directly: with no
// import, `useState`, `useEffect` and `setState` all come back as `any` with no symbol at all.
//
// Rather than score near zero against a corpus written for a different mechanism, each case runs
// through `RunTypedFiles` against the stub below, whose declarations are copied from the real
// `@types/react`: `useState` at index.d.ts:1689, `Dispatch` at 1645, `useEffect` at 1785,
// `useLayoutEffect` at 1775, `useInsertionEffect` at 1915, `useRef` at 1737, `RefObject` at 154.
// The CASE TEXT is upstream's byte for byte; only the module specifier moves, from 'react' to
// "./react", and a script verified every remaining line against the upstream file rather than a
// reading of it.
//
// The trade is stated plainly because it is this rule's central design decision: a dependency on
// `@types/react` and typed props, instead of a builtin shape registry and a unification pass. Good
// for application code, weaker against a corpus that omits its imports.

// reactStub is a faithful subset of `@types/react`, not an invention.
//
// Only the declarations this rule's predicates read are present, and each is copied from the real
// file rather than approximated, because the whole rule rests on `Dispatch` being an ALIAS and
// `RefObject` being an INTERFACE. A stub that made either the other kind would test the opposite of
// what ships.
const reactStub = "export type SetStateAction<S> = S | ((prevState: S) => S);\nexport type Dispatch<A> = (value: A) => void;\nexport type Destructor = () => void;\nexport type EffectCallback = () => void | Destructor;\nexport type DependencyList = readonly unknown[];\nexport interface RefObject<T> { current: T }\nexport declare function useState<S>(initialState: S | (() => S)): [S, Dispatch<SetStateAction<S>>];\nexport declare function useEffect(effect: EffectCallback, deps?: DependencyList): void;\nexport declare function useLayoutEffect(effect: EffectCallback, deps?: DependencyList): void;\nexport declare function useInsertionEffect(effect: EffectCallback, deps?: DependencyList): void;\nexport declare function useEffectEvent<T extends Function>(callback: T): T;\nexport declare function useRef<T>(initialValue: T): RefObject<T>;\nexport declare function useCallback<T extends Function>(callback: T, deps: DependencyList): T;\nexport declare function useMemo<T>(factory: () => T, deps: DependencyList): T;\nexport type ActionDispatch<A extends unknown[]> = (...args: A) => void;\nexport declare function useReducer<S, A>(r: (s: S, a: A) => S, i: S): [S, ActionDispatch<[A]>];\n"

// otherModuleStub declares hooks that are NOT React's, for the custom-hook-name cases.
const otherModuleStub = "export declare function useMyEffect2(cb: () => void): void;\nexport declare function useEffective(cb: () => void): void;\n"

type setStateInEffectCase struct {
	name   string
	source string
}

// runSetStateInEffect runs one case against a program of the case plus the two stubs.
//
// `RunTypedFiles` rather than `RunTyped` because the whole rule is a type question, and the harness
// pins `types: []` in its tsconfig so a real `@types/react` on disk can never leak in and make a
// fixture pass for a reason this file did not state.
func runSetStateInEffect(t *testing.T, source string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, SetStateInEffect, map[string]string{
		"react.d.ts": reactStub,
		"other.d.ts": otherModuleStub,
		"a.tsx":      source,
	}, "a.tsx")
}

func TestSetStateInEffectFires(t *testing.T) {
	t.Parallel()

	// Written as a literal rather than referenced through the rule's own message constant, so the
	// assertion cannot move together with the code it guards.
	const setStateInEffect = "setStateInEffect"

	cases := []setStateInEffectCase{
		// oxc compiler fixture invalid-setState-in-useEffect.js, verbatim. React reports once at `setState`.
		{
			name:   "upstreamDirectSetStateInEffect",
			source: "import {useEffect, useState} from \"./react\";\n\nfunction Component() {\n  const [state, setState] = useState(0);\n  useEffect(() => {\n    setState(s => s + 1);\n  });\n  return state;\n}\n",
		},
		// invalid-setState-in-useEffect-transitive.js. Two levels of indirection outside the effect; this is
		// 	// the case that makes the map load-bearing.
		{
			name:   "upstreamTransitiveThroughTwoLocals",
			source: "import {useEffect, useState} from \"./react\";\n\nfunction Component() {\n  const [state, setState] = useState(0);\n  const f = () => {\n    setState(s => s + 1);\n  };\n  const g = () => {\n    f();\n  };\n  useEffect(() => {\n    g();\n  });\n  return state;\n}\n",
		},
		// invalid-setState-in-useEffect-namespace.js. `React.useEffect` needs no member special case: the
		// 	// checker resolves the property to the same declaration a named import resolves to.
		{
			name:   "upstreamNamespaceImport",
			source: "import * as React from \"./react\";\n\nfunction Component() {\n  const [state, setState] = React.useState(0);\n  React.useEffect(() => {\n    setState(s => s + 1);\n  });\n  return state;\n}\n",
		},
		// invalid-setState-in-useEffect-via-useEffectEvent.js. `useEffectEvent` propagates setter-ness to its
		// 	// result rather than reporting, so the finding lands at the effect that calls the result.
		{
			name:   "upstreamViaUseEffectEvent",
			source: "import {useEffect, useEffectEvent, useState} from \"./react\";\n\nfunction Component() {\n  const [state, setState] = useState(0);\n  const effectEvent = useEffectEvent(() => {\n    setState(true);\n  });\n  useEffect(() => {\n    effectEvent();\n  }, []);\n  return state;\n}\n",
		},
		// Measured: useLayoutEffect reports. Not in either corpus; written from running React on this input.
		{
			name:   "useLayoutEffectReports",
			source: "import {useLayoutEffect, useState} from \"./react\";\nfunction Component() {\n  const [state, setState] = useState(0);\n  useLayoutEffect(() => {\n    setState(1);\n  });\n  return state;\n}\n",
		},
		// Measured: useInsertionEffect reports. The third of upstream's three effect predicates, and the one
		// 	// a porter is most likely to forget.
		{
			name:   "useInsertionEffectReports",
			source: "import {useInsertionEffect, useState} from \"./react\";\nfunction Component() {\n  const [state, setState] = useState(0);\n  useInsertionEffect(() => {\n    setState(1);\n  });\n  return state;\n}\n",
		},
		// Measured: REPORTS. Upstream keys on `binding.imported`, so the local spelling is not consulted.
		// 	// This is the fixture that proves the rule resolves rather than reading the written name.
		{
			name:   "renamedImportStillReports",
			source: "import {useEffect as useFx, useState} from \"./react\";\nfunction Component() {\n  const [state, setState] = useState(0);\n  useFx(() => {\n    setState(1);\n  });\n  return state;\n}\n",
		},
		// Measured: REPORTS. The checker carries `Dispatch` through the assignment, so an alias chain needs
		// 	// no extra machinery here even though upstream needs its LoadLocal bookkeeping for it.
		{
			name:   "aliasedSetterReports",
			source: "import {useEffect, useState} from \"./react\";\nfunction Component() {\n  const [state, setState] = useState(0);\n  const aliased = setState;\n  useEffect(() => {\n    aliased(1);\n  });\n  return state;\n}\n",
		},
		// Measured: TWO findings. Paired with `threeCallsReportOnce` below, this pins the granularity as
		// 	// per-effect rather than per-call, which is a property no single-finding fixture can show.
		{
			name:   "twoEffectsReportTwice",
			source: "import {useEffect, useState} from \"./react\";\nfunction Component() {\n  const [state, setState] = useState(0);\n  useEffect(() => {\n    setState(1);\n  });\n  useEffect(() => {\n    setState(2);\n  });\n  return state;\n}\n",
		},
		// Measured: REPORTS. The effect's first argument is a plain identifier rather than an inline
		// 	// function, which is the arm of upstream's switch that reads `args[0]` as an Identifier.
		{
			name:   "effectBodyIsAHoistedFunction",
			source: "import {useEffect, useState} from \"./react\";\nfunction Component() {\n  const [state, setState] = useState(0);\n  const body = () => {\n    setState(1);\n  };\n  useEffect(body);\n  return state;\n}\n",
		},
		// Measured: REPORTS. A branch on something that is NOT a ref does not exempt, which is the control
		// 	// that gives the ref-controlled exemption its meaning.
		{
			name:   "setStateBehindABranchReports",
			source: "import {useEffect, useState} from \"./react\";\nfunction Component({flag}) {\n  const [state, setState] = useState(0);\n  useEffect(() => {\n    if (flag) {\n      setState(1);\n    }\n  });\n  return state;\n}\n",
		},
		// KNOWN DIVERGENCE in the same direction, found by probing rather than by any fixture.
		//
		// A setter laundered through a custom hook is SILENT upstream, measured on this exact input with a
		// firing control beside it. React's compiler loses its `BuiltInSetState` shape crossing the hook
		// boundary, because a custom hook's return type comes from its own unification pass rather than
		// from a declaration it can read.
		//
		// TypeScript's `Dispatch` alias survives the boundary intact, so we report. This is the clearest
		// instance of the rule's central trade: the checker is STRONGER than upstream's inference here,
		// and being stronger is still a divergence in a port. It is a true positive by the rule's stated
		// purpose, which is what makes it dangerous to leave unstated, since a reader comparing against
		// upstream would call it a bug.
		//
		// Pinned as reporting for the same reason as the case above.
		{
			name:   "setterLaunderedThroughACustomHookIsAKnownDivergence",
			source: "import {useEffect, useState} from \"./react\";\nfunction useThing() {\n  const [value, setValue] = useState(0);\n  return [value, setValue] as const;\n}\nfunction Component() {\n  const [value, setValue] = useThing();\n  useEffect(() => {\n    setValue(1);\n  });\n  return value;\n}\n",
		}}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runSetStateInEffect(t, testCase.source)
			want := []string{setStateInEffect}
			if testCase.name == "twoEffectsReportTwice" {
				want = []string{setStateInEffect, setStateInEffect}
			}
			rule_testing.ExpectFindings(t, result, want...)
		})
	}
}

func TestSetStateInEffectStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []setStateInEffectCase{
		// `valid-setState-in-useEffect-controlled-by-ref-value.js`, verbatim.
		//
		// This was the rule's one known false positive, pinned in the reporting list so that the day
		// the ref exemption's control half landed it would fail and say what to change. It did, twice:
		// `high_level_intermediate_representation.ControlDominators` supplied the frontier, and tainting a store's own lvalue supplied the
		// value half for `const previousX = previousXRef.current`.
		//
		// Moved here rather than deleted, because the fixture is upstream's and its passing is the
		// measurement that the two halves together reproduce upstream's behaviour on the shape they
		// were written for.
		{
			name:   "refControlledBranchIsAKnownDivergence",
			source: "import {useState, useRef, useEffect} from \"./react\";\n\nfunction Component({x, y}) {\n  const previousXRef = useRef(null);\n  const previousYRef = useRef(null);\n\n  const [data, setData] = useState(null);\n\n  useEffect(() => {\n    const previousX = previousXRef.current;\n    previousXRef.current = x;\n    const previousY = previousYRef.current;\n    previousYRef.current = y;\n    if (!areEqual(x, previousX) || !areEqual(y, previousY)) {\n      const data = load({x, y});\n      setData(data);\n    }\n  }, [x, y]);\n\n  return data;\n}\n\nfunction areEqual(a, b) {\n  return a === b;\n}\n\nfunction load({x, y}) {\n  return x * y;\n}\n",
		},
		// valid-setState-in-useEffect-listener.js. The setter is PASSED to setTimeout, never called in the
		// 	// body, which is the distinction the whole rule exists to draw.
		{
			name:   "upstreamListenerSetTimeout",
			source: "import {useEffect, useState} from \"./react\";\n\nfunction Component() {\n  const [state, setState] = useState(0);\n  useEffect(() => {\n    setTimeout(setState, 10);\n  });\n  return state;\n}\n",
		},
		// valid-setState-in-useEffect-listener-transitive.js.
		{
			name:   "upstreamListenerTransitive",
			source: "import {useEffect, useState} from \"./react\";\n\nfunction Component() {\n  const [state, setState] = useState(0);\n  useEffect(() => {\n    const f = () => {\n      setState();\n    };\n    setTimeout(() => f(), 10);\n  });\n  return state;\n}\n",
		},
		// valid-set-state-in-useEffect-from-ref.js. The value-taint half of the ref exemption.
		{
			name:   "upstreamRefDerived",
			source: "import {useState, useRef, useEffect} from \"./react\";\n\nfunction Tooltip() {\n  const ref = useRef(null);\n  const [tooltipHeight, setTooltipHeight] = useState(0);\n\n  useEffect(() => {\n    const {height} = ref.current.getBoundingClientRect();\n    setTooltipHeight(height);\n  }, []);\n\n  return tooltipHeight;\n}\n",
		},
		// valid-setState-in-effect-from-ref-arithmetic.js. Taint through a BinaryExpression.
		{
			name:   "upstreamRefArithmetic",
			source: "import {useState, useRef, useLayoutEffect} from \"./react\";\n\nfunction Component() {\n  const ref = useRef({size: 5});\n  const [computedSize, setComputedSize] = useState(0);\n\n  useLayoutEffect(() => {\n    setComputedSize(ref.current.size * 10);\n  }, []);\n\n  return computedSize;\n}\n",
		},
		// valid-setState-in-effect-from-ref-array-index.js. Taint through a ComputedLoad.
		{
			name:   "upstreamRefArrayIndex",
			source: "import {useState, useRef, useEffect} from \"./react\";\n\nfunction Component() {\n  const ref = useRef([1, 2, 3, 4, 5]);\n  const [value, setValue] = useState(0);\n\n  useEffect(() => {\n    const index = 2;\n    setValue(ref.current[index]);\n  }, []);\n\n  return value;\n}\n",
		},
		// valid-setState-in-effect-from-ref-function-call.js. Taint through a call's arguments.
		{
			name:   "upstreamRefFunctionCall",
			source: "import {useState, useRef, useEffect} from \"./react\";\n\nfunction Component() {\n  const ref = useRef(null);\n  const [width, setWidth] = useState(0);\n\n  useEffect(() => {\n    function getBoundingRect(ref) {\n      if (ref.current) {\n        return ref.current.getBoundingClientRect?.()?.width ?? 100;\n      }\n      return 100;\n    }\n\n    setWidth(getBoundingRect(ref));\n  }, []);\n\n  return width;\n}\n",
		},
		// valid-setState-in-useLayoutEffect-from-ref.js. The exemption is not specific to useEffect.
		{
			name:   "upstreamLayoutEffectFromRef",
			source: "import {useState, useRef, useLayoutEffect} from \"./react\";\n\nfunction Tooltip() {\n  const ref = useRef(null);\n  const [tooltipHeight, setTooltipHeight] = useState(0);\n\n  useLayoutEffect(() => {\n    const {height} = ref.current.getBoundingClientRect();\n    setTooltipHeight(height);\n  }, []);\n\n  return tooltipHeight;\n}\n",
		},
		// Measured: CLEAN, and this is the rule's most surprising behaviour. Moving that same declaration
		// 	// one line up, outside the effect, makes it report; see `upstreamTransitiveThroughTwoLocals`.
		// 	// `getSetStateCall` walks the effect's own blocks and never descends into a function declared
		// 	// inside it. A real hole rather than an exemption, reproduced rather than improved on.
		{
			name:   "nestedFunctionInsideEffectIsNotFollowed",
			source: "import {useEffect, useState} from \"./react\";\nfunction Component() {\n  const [state, setState] = useState(0);\n  useEffect(() => {\n    const f = () => {\n      setState(1);\n    };\n    f();\n  });\n  return state;\n}\n",
		},
		// Measured: CLEAN. The cleanup runs on unmount, not during the effect body, and it is a nested
		// 	// function so the same non-descent applies.
		{
			name:   "setterInCleanupIsClean",
			source: "import {useEffect, useState} from \"./react\";\nfunction Component() {\n  const [state, setState] = useState(0);\n  useEffect(() => {\n    return () => {\n      setState(1);\n    };\n  });\n  return state;\n}\n",
		},
		// Measured: CLEAN. `useReducer` returns `ActionDispatch`, a different alias, so the predicate
		// 	// correctly declines it. This is the fixture that proves the alias name is doing real work rather
		// 	// than matching anything function-shaped.
		{
			name:   "dispatchFromUseReducerIsNotASetter",
			source: "import {useEffect, useReducer} from \"./react\";\nfunction reducer(s: number, a: number) {\n  return s + a;\n}\nfunction Component() {\n  const [state, dispatch] = useReducer(reducer, 0);\n  useEffect(() => {\n    dispatch(1);\n  });\n  return state;\n}\n",
		},
		// Measured: CLEAN. `exhaustive-deps` computes isEffect as /Effect($|[^a-z])/, under which
		// 	// `useMyEffect2` WOULD be an effect. That regex is not consulted by this rule, and this fixture is
		// 	// what stops a porter importing it from the neighbouring rule.
		{
			name:   "customHookNamedLikeAnEffectIsNotOne",
			source: "import {useState} from \"./react\";\nimport {useMyEffect2} from \"./other\";\nfunction Component() {\n  const [state, setState] = useState(0);\n  useMyEffect2(() => {\n    setState(1);\n  });\n  return state;\n}\n",
		},
		// Measured: CLEAN. Only the FIRST argument is read, so a setter anywhere else is not a callback.
		{
			name:   "setterAsSecondArgumentIsClean",
			source: "import {useEffect, useState} from \"./react\";\nfunction Component() {\n  const [state, setState] = useState(0);\n  useEffect(() => {}, setState);\n  return state;\n}\n",
		},
		// A degenerate call with no arguments at all. Written from reading our own code rather than from
		// 	// the corpus: `firstIdentifierArgument` returns false on an empty slice, and without that guard
		// 	// this input would index out of range.
		{
			name:   "effectWithNoArgumentsIsClean",
			source: "import {useEffect} from \"./react\";\nfunction Component() {\n  useEffect();\n  return 1;\n}\n",
		}}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runSetStateInEffect(t, testCase.source))
		})
	}
}

// TestSetStateInEffectPointsAtTheSetterCall asserts WHERE the finding lands, which no message-id
// fixture can see.
//
// Upstream carries two spans: a primary on the setter reading "Avoid calling setState() directly
// within an effect", and a secondary on the effect callee reading "This is the containing effect".
// Our Diagnostic has one Range, so the primary is kept where upstream puts it. oxc's snapshot for
// its single fail case underlines `setState` at 6:5, eight characters, which is what this pins.
//
// The expectation is sliced from the source the HARNESS wrote rather than from the Go literal.
// `rule_testing.RunTypedFiles` writes `strings.TrimSpace(contents)+"\n"`, so a fixture carrying a
// leading newline sits one byte off from its literal, and an assertion built by slicing the literal
// reports a span shifted by one while the rule is correct. That is a documented trap in this
// harness and it has cost a porter a long hunt through the representation and the shim.
func TestSetStateInEffectPointsAtTheSetterCall(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   string
		wantText string
	}{
		{
			name:     "directCall",
			source:   "import {useEffect, useState} from \"./react\";\nfunction Component() {\n  const [state, setState] = useState(0);\n  useEffect(() => {\n    setState(1);\n  });\n  return state;\n}\n",
			wantText: "setState",
		},
		// An alias is reported at the name actually called, not at the original setter, which is
		// what makes the finding point at a line the reader has to change.
		{
			name:     "aliasedCall",
			source:   "import {useEffect, useState} from \"./react\";\nfunction Component() {\n  const [state, setState] = useState(0);\n  const aliased = setState;\n  useEffect(() => {\n    aliased(1);\n  });\n  return state;\n}\n",
			wantText: "aliased",
		},
		// The transitive case reports at the call INSIDE THE EFFECT, not at the setter inside the
		// helper. Measured against React 7.1.1: the finding underlines `f` at 8:5, seven lines
		// after the `setState` it eventually reaches.
		//
		// This expectation was written the other way round first, from reading upstream's
		// `getSetStateCall` and seeing it return the setter's callee. That reading was wrong about
		// WHICH invocation of that function produces the reported value, and only running React on
		// the input settled it. The rule was wrong in the same direction, so code and fixture
		// agreed with each other and disagreed with upstream, which is the failure a span
		// assertion exists to catch.
		{
			name:     "transitiveCall",
			source:   "import {useEffect, useState} from \"./react\";\nfunction Component() {\n  const [state, setState] = useState(0);\n  const f = () => {\n    setState(1);\n  };\n  useEffect(() => {\n    f();\n  });\n  return state;\n}\n",
			wantText: "f",
		},
		// A comment above the call is trivia, not the call. TanStack Query's utils.tsx writes
		// `// eslint-disable-next-line react-hooks/set-state-in-effect` there, and the finding used
		// to start at the comment, one line above the setter, so the suppression for the next line
		// missed it (#zx5xvtg item 8). A block comment on the same line is skipped the same way.
		{
			name:     "callUnderALineComment",
			source:   "import {useEffect, useState} from \"./react\";\nfunction Component() {\n  const [state, setState] = useState(0);\n  useEffect(() => {\n    // a note about the next line\n    setState(1);\n  });\n  return state;\n}\n",
			wantText: "setState",
		},
		{
			name:     "callAfterABlockComment",
			source:   "import {useEffect, useState} from \"./react\";\nfunction Component() {\n  const [state, setState] = useState(0);\n  useEffect(() => {\n    /* first */ setState(1);\n  });\n  return state;\n}\n",
			wantText: "setState",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runSetStateInEffect(t, testCase.source)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected exactly one finding, got %d", len(result.Diagnostics))
			}
			// The harness trims the fixture before writing it, so the expectation is sliced from
			// the same transformation rather than from the literal above.
			written := strings.TrimSpace(testCase.source) + "\n"
			span := result.Diagnostics[0].Range
			if span.Pos() < 0 || span.End() > len(written) || span.Pos() >= span.End() {
				t.Fatalf("finding range %d..%d is outside the %d byte file", span.Pos(), span.End(), len(written))
			}
			if got := written[span.Pos():span.End()]; got != testCase.wantText {
				t.Errorf("finding points at %q, want %q", got, testCase.wantText)
			}
		})
	}
}

// TestSetStateInEffectMessage asserts the rendered message exactly.
//
// `rule.Message` is `{Id, Description}` with no interpolation layer, so there is nothing to render
// and the two fields are compared directly. Asserted against literals typed here rather than
// against the rule's own constant: comparing a diagnostic to the constant it was reported with is
// equality that moves on both sides under mutation, which is how a message-text mutant survives a
// test that looks correct.
func TestSetStateInEffectMessage(t *testing.T) {
	t.Parallel()

	result := runSetStateInEffect(t, "import {useEffect, useState} from \"./react\";\nfunction Component() {\n  const [state, setState] = useState(0);\n  useEffect(() => {\n    setState(1);\n  });\n  return state;\n}\n")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected exactly one finding, got %d", len(result.Diagnostics))
	}
	if got := result.Diagnostics[0].Message.Id; got != "setStateInEffect" {
		t.Errorf("message id is %q, want %q", got, "setStateInEffect")
	}
	const wantDescription = "This effect calls a state setter synchronously in its body, so every time the " +
		"effect runs React starts another render, which runs the effect's dependencies again. " +
		"Effects are for synchronizing React with something outside it, such as the document " +
		"title, a subscription, or a network request, and a setter called straight from the body " +
		"is usually a value that could have been computed during render instead. Derive the value " +
		"while rendering, pass the initial value to `useState`, or set it from the event that " +
		"actually caused the change. Calling the setter from a callback the effect registers, " +
		"such as a subscription or a timer, is fine and is not reported."
	if got := result.Diagnostics[0].Message.Description; got != wantDescription {
		t.Errorf("message description is:\n%q\nwant:\n%q", got, wantDescription)
	}
}

// TestSetStateInEffectRequiresTheTypedHarness pins that this rule is useless without a checker.
//
// The port brief's measured warning: on this shim `GetSymbolAtLocation` and `GetTypeAtLocation`
// return nil on a nil checker rather than crashing, so a typed rule missing its guard buys a
// VACUOUS GREEN rather than an obvious panic, and every silent fixture passes for the wrong reason.
// This asserts the untyped harness produces nothing, so a later revert to `rule_testing.Run` fails
// loudly here instead of quietly everywhere.
func TestSetStateInEffectRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, SetStateInEffect, "a.tsx", "import {useEffect, useState} from \"./react\";\nfunction Component() {\n  const [state, setState] = useState(0);\n  useEffect(() => {\n    setState(1);\n  });\n  return state;\n}\n")
	rule_testing.ExpectClean(t, result)
}

// TestSetStateInEffectSpreadArgument pins the spread guard in `firstIdentifierArgument`.
//
// Upstream proceeds only when `args[0].kind === 'Identifier'`, which excludes a spread, and the
// spread guard here reproduces that. Measured against React on both inputs below: silent.
//
// This exists because a mutation deleting `|| args[0].Spread` SURVIVED the whole imported corpus,
// which writes no spread anywhere. Before writing it, the question asked was whether the branch is
// reachable at all or merely crash protection, since `Argument.Place` is a real Place even for a
// spread and nothing would panic. It is reachable and it is a real judgment: without the guard,
// `useEffect(...args)` reads the spread's own value as if it were the callback, so a spread of an
// array containing a setter-calling closure would be attributed to the effect.
func TestSetStateInEffectSpreadArgument(t *testing.T) {
	t.Parallel()

	// A bare spread of an unknown array. Upstream silent, and the guard is what makes us silent.
	rule_testing.ExpectClean(t, runSetStateInEffect(t, "import {useEffect, useState} from \"./react\";\nfunction Component({args}: {args: [() => void]}) {\n  const [state, setState] = useState(0);\n  useEffect(...args);\n  return state;\n}\n"))

	// A spread of an array literal holding a callback that DOES call a setter. This is the input
	// that separates the two versions: the callback is a real setter-calling closure, so without
	// the spread guard the first argument reads as one.
	rule_testing.ExpectClean(t, runSetStateInEffect(t, "import {useEffect, useState} from \"./react\";\nfunction Component() {\n  const [state, setState] = useState(0);\n  const args: [() => void] = [() => { setState(1); }];\n  useEffect(...args);\n  return state;\n}\n"))
}

// TestSetStateInEffectDescendsIntoNestedFunctions pins the walk over `Function.Functions`.
//
// A component or hook written INSIDE another function is lowered into its parent's arena rather
// than visited by `forEachCompiledFunction`, which stops at the outermost function-like node on
// each branch. Without the recursion at the end of `reportSetStateInEffects`, an effect written in
// such a nested function is never analysed at all, and a mutation deleting that loop survived the
// whole imported corpus, which contains no nested component.
//
// Measured against React 7.1.1 on both inputs: each reports once, at the `setS` inside the effect.
// This is the "reachable in the code, invisible to the corpus" case rather than an equivalence.
func TestSetStateInEffectDescendsIntoNestedFunctions(t *testing.T) {
	t.Parallel()

	const setStateInEffect = "setStateInEffect"

	// A nested component. The outer function is not a component itself, so nothing would look
	// inside it without the descent.
	rule_testing.ExpectFindings(t, runSetStateInEffect(t, "import {useEffect, useState} from \"./react\";\nfunction Outer() {\n  function Inner() {\n    const [state, setState] = useState(0);\n    useEffect(() => {\n      setState(1);\n    });\n    return state;\n  }\n  return Inner;\n}\n"), setStateInEffect)

	// A nested custom hook, which upstream admits through the same gate for a different reason.
	rule_testing.ExpectFindings(t, runSetStateInEffect(t, "import {useEffect, useState} from \"./react\";\nfunction Outer() {\n  function useInner() {\n    const [state, setState] = useState(0);\n    useEffect(() => {\n      setState(1);\n    });\n    return state;\n  }\n  return useInner;\n}\n"), setStateInEffect)
}

// TestSetStateInEffectDeclinesACallbackClosingOverNoSetter pins the `anyOperandIsSetter` gate.
//
// Upstream guards the descent into a nested function on an operand already being setter-like, so a
// callback that closes over nothing relevant is never scanned. A mutation deleting that guard
// survived the imported corpus, because every fail case there closes over a setter and every pass
// case either has no effect or no setter at all.
//
// The distinguishing input is an effect whose callback calls something that is NOT a setter. With
// the gate, the callback is never scanned; without it, the callback is scanned and declines on the
// type test instead. Both reach silence, so this fixture pins the OBSERVABLE behaviour rather than
// the path — and the path is what the gate is for, since scanning every callback in the tree is the
// cost the gate exists to avoid. Measured against React: silent.
func TestSetStateInEffectDeclinesACallbackClosingOverNoSetter(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, runSetStateInEffect(t, "import {useEffect} from \"./react\";\nfunction Component({onTick}: {onTick: () => void}) {\n  useEffect(() => {\n    onTick();\n  });\n  return null;\n}\n"))
}

// TestSetStateInEffectSeesThroughManualMemoization pins the `useCallback` / `useMemo` erasure.
//
// # What was missing, and how the gap was found
//
// This was found by differential run rather than by reading: `s l --linter both` over Kirk's tree
// had React reporting one finding this rule did not. Reduced, the difference was one wrapper:
//
//	const bump = useCallback(() => setState(1), []);  // React reports, this was silent
//	const bump = () => setState(1);                   // both report
//	useEffect(() => { bump(); });
//
// Both spellings were run through React 7.1.1 via the ESLint Linter API, not inferred. The
// mechanism is upstream's pipeline rather than its validator: `dropManualMemoization`
// (`Pipeline.ts:169`) rewrites `useCallback(fn, deps)` to a plain load of `fn` and
// `useMemo(fn, deps)` to `fn()`, and `validateNoSetStateInEffects` runs at line 50131 of the same
// bundle, long after. So upstream's validator never sees a memo call at all, and this rule --
// transcribed faithfully from that validator -- was faithful to a function whose input had not
// been prepared the same way.
//
// The fix is entirely in the input: `high_level_intermediate_representation.ForFunctionWithoutManualMemoization` runs the erasure
// before this rule reads the graph. The rule itself is unchanged, because it already carries
// setter-ness across `LoadLocal`, which is exactly what the `useCallback` rewrite produces.
//
// # Why the async cases are here
//
// The site that exposed this is a mount fetch whose setters all run after an `await`, so the
// obvious reading is that the await boundary is what upstream is tracking. It is not: `asyncAfterAwait`
// and `asyncBeforeAwait` below differ only in where the setter sits relative to the first await,
// React reports both, and so does this. The await is not consulted by either implementation. That
// is recorded because it is the wrong conclusion a reader is most likely to draw from the site,
// and because it is the one that would send someone rewriting application code to no effect.
func TestSetStateInEffectSeesThroughManualMemoization(t *testing.T) {
	t.Parallel()

	const setStateInEffect = "setStateInEffect"

	cases := []setStateInEffectCase{
		// The reduced form of the gap. React 7.1.1 reports at the `bump()` call site.
		{
			name:   "useCallbackWrappingASetter",
			source: "import {useCallback, useEffect, useState} from \"./react\";\n\nfunction Component() {\n  const [state, setState] = useState(0);\n  const bump = useCallback(() => {\n    setState(1);\n  }, []);\n  useEffect(() => {\n    bump();\n  }, [bump]);\n  return state;\n}\n",
		},
		// Every setter after the first await, which is the shape the real site has. React reports.
		{
			name:   "asyncAfterAwait",
			source: "import {useCallback, useEffect, useState} from \"./react\";\n\nfunction Component() {\n  const [state, setState] = useState(0);\n  const load = useCallback(async () => {\n    const value = await Promise.resolve(1);\n    setState(value);\n  }, []);\n  useEffect(() => {\n    void load();\n  }, [load]);\n  return state;\n}\n",
		},
		// The same with a setter before the await. Reports too, which is what shows the await is
		// not the discriminator either implementation uses.
		{
			name:   "asyncBeforeAwait",
			source: "import {useCallback, useEffect, useState} from \"./react\";\n\nfunction Component() {\n  const [state, setState] = useState(0);\n  const load = useCallback(async () => {\n    setState(0);\n    const value = await Promise.resolve(1);\n    setState(value);\n  }, []);\n  useEffect(() => {\n    void load();\n  }, [load]);\n  return state;\n}\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, runSetStateInEffect(t, testCase.source), setStateInEffect)
		})
	}
}

// TestSetStateInEffectReachesThroughAMemoizedCallbackThatReturnsAFunction closes the last
// divergence from React in this rule.
//
// # What this was, and what closed it
//
// `const bump = useMemo(() => () => setState(1), [])` called from an effect is reported by React
// 7.1.1 and was silent here through three separate versions of this comment, which is worth saying
// because each one was true when written.
//
// The erasure rewrites `useMemo(fn, deps)` to `fn()`, leaving a call whose result is the inner
// closure, and setter-ness would have to flow through a call's return value -- which neither
// upstream's validator nor this rule tracks. Upstream never faces that because
// `InlineImmediatelyInvokedFunctionExpressions` runs one line later and replaces the call with the
// closure itself. That pass landed here at `f9eba3f`.
//
// Landing it was not enough, and the measurement is the interesting part. Wiring the pass in as
// published changed nothing: 0 findings against 0 over Kirk's tree, identical node counts, this
// fixture still silent. The cause is `memoizedResults`, a deliberate exclusion in that file that
// declines to inline any callback `DropManualMemoization` marked -- and an inline `useMemo`
// callback is exactly what gets marked, so the one shape this test is about was the one shape the
// pass refused.
//
// The exclusion is right for the caller it was written for. `preserve-manual-memoization` runs with
// its dependency comparison off, and with it off, nine of its fixtures fire on the other two
// conditions because of the scope fusion inlining removes; inlining memo callbacks moves its
// goldens from 15 to 6. That cost is real and this rule does not get to impose it.
//
// It also does not have to. This rule reads no scope, compares no dependency, and has no golden, so
// it pays none of that cost and needs precisely what upstream does. The exclusion moved to the call
// site: `InlineImmediatelyInvokedFunctionExpressionsIncludingMemoCallbacks` lifts it, the guarded
// entry point is unchanged for the other caller, and the full `hir` suite is green either way.
//
// Verified against React 7.1.1 on the same input rather than assumed: both report, at the same line
// and the same column.
func TestSetStateInEffectReachesThroughAMemoizedCallbackThatReturnsAFunction(t *testing.T) {
	t.Parallel()

	const setStateInEffect = "setStateInEffect"

	// The case itself. Was the failing-shape fixture; now asserts the fix.
	rule_testing.ExpectFindings(t, runSetStateInEffect(t, "import {useEffect, useMemo, useState} from \"./react\";\n\nfunction Component() {\n  const [state, setState] = useState(0);\n  const bump = useMemo(() => () => {\n    setState(1);\n  }, []);\n  useEffect(() => {\n    bump();\n  }, [bump]);\n  return state;\n}\n"), setStateInEffect)

	// A `useMemo` returning a value rather than a function has no setter to reach, and inlining its
	// callback must not invent one. This is the case that would break first if the splice ever
	// spliced something it should not.
	rule_testing.ExpectClean(t, runSetStateInEffect(t, "import {useEffect, useMemo, useState} from \"./react\";\n\nfunction Component() {\n  const [state] = useState(0);\n  const doubled = useMemo(() => state * 2, [state]);\n  useEffect(() => {\n    console.info(doubled);\n  }, [doubled]);\n  return doubled;\n}\n"))

	// A programmer-written IIFE, which nothing marks and which the pass has always inlined. Silent
	// because there is no setter in it, not because it was declined.
	rule_testing.ExpectClean(t, runSetStateInEffect(t, "import {useEffect, useState} from \"./react\";\n\nfunction Component() {\n  const [state] = useState(0);\n  const value = (function() {\n    return 1;\n  })();\n  useEffect(() => {\n    console.info(value);\n  }, [value]);\n  return state + value;\n}\n"))
}

// TestSetStateInEffectDeclinesMemoizationItDoesNotReach records where the erasure still stops.
//
// The propagation is not general, which bounds what the erasure claims. A closure passed through an
// ordinary function or a custom hook stays silent in both implementations, measured on both: only
// `useMemo` and `useCallback` are erased, because only those two are what `dropManualMemoization`
// recognises.
func TestSetStateInEffectDeclinesMemoizationItDoesNotReach(t *testing.T) {
	t.Parallel()

	// A closure through an ordinary function. Silent in both, so the erasure is not "any call".
	rule_testing.ExpectClean(t, runSetStateInEffect(t, "import {useEffect, useState} from \"./react\";\n\nfunction identity<T>(callback: T): T {\n  return callback;\n}\n\nfunction Component() {\n  const [state, setState] = useState(0);\n  const bump = identity(() => {\n    setState(1);\n  });\n  useEffect(() => {\n    bump();\n  }, [bump]);\n  return state;\n}\n"))

	// A closure through a custom hook. Silent in both, for the same reason.
	rule_testing.ExpectClean(t, runSetStateInEffect(t, "import {useEffect, useState} from \"./react\";\n\nfunction useWrap<T>(callback: T): T {\n  return callback;\n}\n\nfunction Component() {\n  const [state, setState] = useState(0);\n  const bump = useWrap(() => {\n    setState(1);\n  });\n  useEffect(() => {\n    bump();\n  }, [bump]);\n  return state;\n}\n"))

	// A useCallback closing over no setter: the erasure does not make every memoized callback
	// suspect.
	rule_testing.ExpectClean(t, runSetStateInEffect(t, "import {useCallback, useEffect, useState} from \"./react\";\n\nfunction Component() {\n  const [state] = useState(0);\n  const log = useCallback(() => {\n    console.info(\"x\");\n  }, []);\n  useEffect(() => {\n    log();\n  }, [log]);\n  return state;\n}\n"))
}

// TestSetStateInEffectGatesTheSecondLoweringWithoutLosingFindings pins the fast path.
//
// `ForFunctionWithoutManualMemoization` skips its own lowering when the function's text names
// neither memo hook, which is what makes the erasure free: 12,986 of 13,171 functions in Kirk's
// tree take that path. The risk it introduces is one-directional and silent -- a function the gate
// wrongly calls memo-free is lowered from the shared cache, so the erasure never runs and the
// finding disappears with no error anywhere, which is the failure a fixture has to catch because
// nothing else can.
//
// Each case here reports, and each would go silent if the gate stopped seeing its spelling.
func TestSetStateInEffectGatesTheSecondLoweringWithoutLosingFindings(t *testing.T) {
	t.Parallel()

	const setStateInEffect = "setStateInEffect"

	// The namespace spelling: the span names `useCallback` only through a member access, and the
	// pass recognises it through its `react` sidemap rather than a bare identifier.
	rule_testing.ExpectFindings(t, runSetStateInEffect(t, "import * as React from \"./react\";\n\nfunction Component() {\n  const [state, setState] = React.useState(0);\n  const bump = React.useCallback(() => {\n    setState(1);\n  }, []);\n  React.useEffect(() => {\n    bump();\n  }, [bump]);\n  return state;\n}\n"), setStateInEffect)

	// Escaped spellings: each reaches the erasure through cooked identifier or string text while
	// the span never holds the name's letters in a row, so a substring-only gate handed this rule
	// the memo-intact graph and every one of these was silent where the plain spelling reports.
	for _, source := range []string{
		"import {use\\u0043allback, useEffect, useState} from \"./react\";\n\nfunction Component() {\n  const [state, setState] = useState(0);\n  const bump = use\\u0043allback(() => {\n    setState(1);\n  }, []);\n  useEffect(() => {\n    bump();\n  }, [bump]);\n  return state;\n}\n",
		"import {use\\u{43}allback, useEffect, useState} from \"./react\";\n\nfunction Component() {\n  const [state, setState] = useState(0);\n  const bump = use\\u{43}allback(() => {\n    setState(1);\n  }, []);\n  useEffect(() => {\n    bump();\n  }, [bump]);\n  return state;\n}\n",
		"import * as React from \"./react\";\n\nfunction Component() {\n  const [state, setState] = React.useState(0);\n  const bump = React.use\\u0043allback(() => {\n    setState(1);\n  }, []);\n  React.useEffect(() => {\n    bump();\n  }, [bump]);\n  return state;\n}\n",
		"import * as React from \"./react\";\n\nfunction Component() {\n  const [state, setState] = React.useState(0);\n  const bump = React['use\\x43allback'](() => {\n    setState(1);\n  }, []);\n  React.useEffect(() => {\n    bump();\n  }, [bump]);\n  return state;\n}\n",
	} {
		rule_testing.ExpectFindings(t, runSetStateInEffect(t, source), setStateInEffect)
	}

	// The control: no memo call at all, so this takes the gate's fast path and must still report
	// through the shared lowering. Without this, a gate that returned `nil` on its fast path would
	// look correct from the case above alone.
	rule_testing.ExpectFindings(t, runSetStateInEffect(t, "import {useEffect, useState} from \"./react\";\n\nfunction Component() {\n  const [state, setState] = useState(0);\n  const bump = () => {\n    setState(1);\n  };\n  useEffect(() => {\n    bump();\n  }, [bump]);\n  return state;\n}\n"), setStateInEffect)
}

// TestSetStateInEffectRenamedMemoImportIsAKnownDivergence records a gap the gate does not cause.
//
// `import {useCallback as useCached}` called as `useCached(...)` is reported by React 7.1.1 and is
// silent here, measured on both. It is recorded beside the gate fixtures because that is where a
// reader will look for it and wrongly blame the fast path: the gate is innocent, and the proof is
// that the ungated binary is silent on this input too.
//
// The cause is upstream's `Environment.getGlobalDeclaration`, which keys on `binding.imported` --
// the name a binding was declared as -- while this tree's lowering leaves `LoadGlobal.BindingKind`
// at `Global` for every global and never populates `Imported`, documented at
// `internal/utilities/hir/lower.go:58`. So the memo sidemap can only recognise a call spelled with the
// hook's own name. It is the same missing binding-kind information the rule's own header names as
// its other divergence, reached from the opposite direction, and it closes when that closes.
func TestSetStateInEffectRenamedMemoImportIsAKnownDivergence(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, runSetStateInEffect(t, "import {useCallback as useCached, useEffect, useState} from \"./react\";\n\nfunction Component() {\n  const [state, setState] = useState(0);\n  const bump = useCached(() => {\n    setState(1);\n  }, []);\n  useEffect(() => {\n    bump();\n  }, [bump]);\n  return state;\n}\n"))
}

// TestSetStateInEffectJudgesOnlyWhatReactCompiles pins the component-or-hook gate (#1pmwkmv).
//
// Upstream reaches its validator only for a function it compiles. This rule judged every function
// until the gate went in, and `notAComponent` below reported here while React is silent on it. A
// component declared inside a plain wrapper is still a unit, so the descent is pinned beside it.
func TestSetStateInEffectJudgesOnlyWhatReactCompiles(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, runSetStateInEffect(t, "import {useEffect, useState} from \"./react\";\n\nexport function notAComponent() {\n  const [state, setState] = useState(0);\n  useEffect(() => {\n    setState(1);\n  });\n  return state;\n}\n"))

	rule_testing.ExpectFindings(t, runSetStateInEffect(t, "import {useEffect, useState} from \"./react\";\n\nexport function wrapper() {\n  function Component() {\n    const [state, setState] = useState(0);\n    useEffect(() => {\n      setState(1);\n    });\n    return state;\n  }\n  return Component;\n}\n"), "setStateInEffect")
}
