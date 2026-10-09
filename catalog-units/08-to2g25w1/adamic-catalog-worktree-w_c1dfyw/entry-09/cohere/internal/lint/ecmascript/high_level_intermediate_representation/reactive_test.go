package high_level_intermediate_representation

import (
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/static_single_assignment"
)

// reactiveDeclarations is the part of `@types/react` these tests read.
//
// A local declaration file rather than the real package, for the reason
// `set_state_in_render_test.go` gives: the fixture program has no node_modules to resolve, and the
// three type names this pass keys on are the whole surface. The shapes are copied from
// `@types/react/index.d.ts`: `useState` returns `[S, Dispatch<SetStateAction<S>>]` at line 1689,
// `useRef` returns `RefObject<T>` at 1737, `useContext<T>(c): T` at 1682, and `useReducer` yields an
// `ActionDispatch` at 1708.
const reactiveDeclarations = `declare module 'react' {
  export type SetStateAction<S> = S | ((prev: S) => S);
  export type Dispatch<A> = (value: A) => void;
  export interface RefObject<T> { current: T }
  export interface Context<T> { Provider: unknown }
  export type ActionDispatch<A extends any[]> = (...args: A) => void;
  export function useState<S>(initial?: S): [S, Dispatch<SetStateAction<S>>];
  export function useRef<T>(initial: T): RefObject<T>;
  export function useContext<T>(context: Context<T>): T;
  export function useMemo<T>(compute: () => T, deps: unknown[]): T;
  export function useReducer<S, A>(reducer: (state: S, action: A) => S, initial: S): [S, ActionDispatch<[A]>];
}`

// reactiveNames returns the names of every value marked reactive, deduplicated and sorted.
//
// Sorted so an assertion does not depend on the order the walk marks values in. An order that
// varies between runs is not hypothetical in this tree: a rule shipped whose message named a
// different builtin between runs.
//
// Temporaries are excluded because they have no source name to assert against, and their count is
// an artifact of how an expression happened to be lowered rather than a fact about the program.
func reactiveNames(t *testing.T, source string) []string {
	t.Helper()

	found := map[string]bool{}
	probe := rule.Rule{
		Name:             "reactive-names-harness",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					if ctx.TypeChecker == nil {
						t.Fatal("the typed harness handed this probe a nil checker, so it would " +
							"pass vacuously; run it through RunTypedFiles")
					}
					forEachFunctionLike(node, func(functionNode *ast.Node) {
						function := Lower(functionNode, ctx.TypeChecker)
						if function == nil {
							return
						}
						Construct(function)
						InferReactive(function, ctx.TypeChecker)
						collectReactiveNames(function, found)
					})
				},
			}
		},
	}
	rule_testing.RunTypedFiles(t, probe, map[string]string{
		"react.d.ts":  reactiveDeclarations,
		"fixture.tsx": source,
	}, "fixture.tsx")

	names := make([]string, 0, len(found))
	for name := range found {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func collectReactiveNames(function *Function, found map[string]bool) {
	record := func(place Place) {
		if !place.Reactive {
			return
		}
		identifier := function.Identifiers[place.Identifier]
		if identifier != nil && identifier.Name != "" {
			found[identifier.Name] = true
		}
	}
	for _, block := range function.Blocks {
		for _, phi := range block.Phis {
			record(phi.Place)
		}
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			EachInstructionPlace(instruction, func(place Place, role PlaceRole) { record(place) })
		}
		EachTerminalPlace(block.Terminal, func(place Place, role PlaceRole) { record(place) })
	}
	for _, param := range function.Params {
		record(param)
	}
	for _, nested := range function.Functions {
		collectReactiveNames(nested, found)
	}
}

func hasName(names []string, want string) bool {
	for _, name := range names {
		if name == want {
			return true
		}
	}
	return false
}

func assertReactive(t *testing.T, names []string, want ...string) {
	t.Helper()
	for _, name := range want {
		if !hasName(names, name) {
			t.Errorf("expected %q to be reactive, got reactive set [%s]", name, strings.Join(names, " "))
		}
	}
}

func assertNotReactive(t *testing.T, names []string, want ...string) {
	t.Helper()
	for _, name := range want {
		if hasName(names, name) {
			t.Errorf("expected %q NOT to be reactive, got reactive set [%s]", name, strings.Join(names, " "))
		}
	}
}

// TestReactiveMarksParameters pins the root source: a component's props vary between renders.
func TestReactiveMarksParameters(t *testing.T) {
	t.Parallel()

	names := reactiveNames(t, `
export function Component(props: {title: string}) {
  const shown = props.title;
  return shown;
}
`)
	assertReactive(t, names, "props", "shown")
}

// TestReactivePropagatesThroughAliases pins that an alias chain of any length carries reactivity.
//
// This is the load/store propagation. Without it a rule asking about `c` would answer false while
// `c` is the same value as `props`, which is the class of defect the representation exists to end.
func TestReactivePropagatesThroughAliases(t *testing.T) {
	t.Parallel()

	names := reactiveNames(t, `
export function Component(props: {title: string}) {
  const a = props;
  const b = a;
  const c = b;
  return c;
}
`)
	assertReactive(t, names, "props", "a", "b", "c")
}

// TestReactiveLeavesConstantsAlone is the negative half of the two-sided key.
//
// A test asserting only that reactive things are reactive is equally satisfied by a pass that marks
// everything, so this asserts a value derived from nothing but literals stays false while a value
// derived from props in the same function is true.
func TestReactiveLeavesConstantsAlone(t *testing.T) {
	t.Parallel()

	names := reactiveNames(t, `
export function Component(props: {n: number}) {
  const constant = 1 + 2;
  const derived = props.n;
  return constant + derived;
}
`)
	assertReactive(t, names, "props", "derived")
	assertNotReactive(t, names, "constant")
}

// TestReactiveJoinsAtPhi pins the merge: reactive on ANY path means reactive after the join.
//
// This is the case that cannot be answered from the syntax tree without re-deriving control flow,
// and it is why this pass is in the representation rather than in a rule.
func TestReactiveJoinsAtPhi(t *testing.T) {
	t.Parallel()

	names := reactiveNames(t, `
export function Component(props: {c: boolean, n: number}) {
  let value = 0;
  if (props.c) {
    value = props.n;
  }
  return value;
}
`)
	assertReactive(t, names, "props", "value")
}

// phiResultIsReactive reports whether the phi DEFINING the named binding is itself reactive.
//
// `reactiveNames` cannot answer this and the difference is the whole reason this helper exists. That
// collector records every place naming a reactive value, and a phi's OPERANDS are places. So a phi
// whose operand is reactive puts the binding's name in the collected set whether or not the phi
// RESULT was ever marked, and an assertion built on it passes while the merge rule is disabled.
//
// Found by a surviving mutant. Disabling the phi operand join left every name-based assertion green
// because the operand `value$25` carries the name `value` on its own. The layer being asserted was
// "some value called value is reactive" when the property under test is "the MERGE is reactive".
func phiResultIsReactive(t *testing.T, source string, binding string) bool {
	t.Helper()

	found, sawPhi := false, false
	probe := rule.Rule{
		Name:             "reactive-phi-result-harness",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					forEachFunctionLike(node, func(functionNode *ast.Node) {
						function := Lower(functionNode, ctx.TypeChecker)
						if function == nil {
							return
						}
						Construct(function)
						InferReactive(function, ctx.TypeChecker)
						for _, block := range function.Blocks {
							for _, phi := range block.Phis {
								identifier := function.Identifiers[phi.Place.Identifier]
								if identifier == nil || identifier.Name != binding {
									continue
								}
								sawPhi = true
								if phi.Place.Reactive {
									found = true
								}
							}
						}
					})
				},
			}
		},
	}
	rule_testing.RunTypedFiles(t, probe, map[string]string{
		"react.d.ts":  reactiveDeclarations,
		"fixture.tsx": source,
	}, "fixture.tsx")

	if !sawPhi {
		t.Fatalf("no phi defines %q, so this assertion would prove nothing about the merge rule; "+
			"the fixture must contain a real join", binding)
	}
	return found
}

// TestReactivePhiResultIsMarkedByOperandJoin asserts the MERGE rather than the name.
//
// This is the fixture that kills the phi-operand-join mutant, and the second attempt at it. The
// first asserted through `reactiveNames` and stayed green under the mutation for the reason
// `phiResultIsReactive` documents: the operands carry the binding's name themselves.
//
// The branch test is a local constant, so the block is NOT reactive-controlled — checked directly,
// the controlled set is empty for this input — which means the operand join is the only rule that
// can mark this phi. That isolates the two phi rules from each other, which mutating them
// separately proved was necessary.
func TestReactivePhiResultIsMarkedByOperandJoin(t *testing.T) {
	t.Parallel()

	source := `
export function Component(props: {n: number}) {
  const flag = true;
  let value = 0;
  if (flag) {
    value = props.n;
  }
  return value;
}
`
	if !phiResultIsReactive(t, source, "value") {
		t.Error("the phi merging a reactive operand with a constant must itself be reactive; " +
			"a consumer reading the merge sees a value that cannot change between renders")
	}
}

// TestReactivePhiResultIsMarkedByReactiveControl is the same assertion for the other phi rule.
//
// Both operands are constants here, so the operand join cannot mark this phi and only the
// control-dominator rule can. Together with the test above, each phi rule now has a case the other
// cannot satisfy, which is what mutating them independently requires.
func TestReactivePhiResultIsMarkedByReactiveControl(t *testing.T) {
	t.Parallel()

	source := `
export function Component(props: {c: boolean}) {
  let value = 1;
  if (props.c) {
    value = 2;
  }
  return value;
}
`
	if !phiResultIsReactive(t, source, "value") {
		t.Error("a phi whose operands are constants but whose branch tests a reactive value must " +
			"be reactive; which constant it holds depends on something that varies per render")
	}
}

// TestReactiveJoinsReactiveOperandUnderConstantBranch separates the phi's two marking rules.
//
// This exists because a mutant disabling the phi OPERAND join survived the whole suite while a
// mutant disabling the phi CONTROL rule was caught. Reading rather than adding a fixture blind, the
// cause was subsumption: in every phi fixture written first, the branch test was `props.c`, so the
// joined value was reactive-controlled as well as having a reactive operand, and the control rule
// alone produced the same answer.
//
// The distinguishing input is a phi with a reactive OPERAND under a NON-reactive test. `flag` is a
// local constant, so the block is not reactive-controlled and only the operand join can mark
// `value`. Confirmed by re-scoring: with this case present the operand mutant is caught.
func TestReactiveJoinsReactiveOperandUnderConstantBranch(t *testing.T) {
	t.Parallel()

	names := reactiveNames(t, `
export function Component(props: {n: number}) {
  const flag = true;
  let value = 0;
  if (flag) {
    value = props.n;
  }
  return value;
}
`)
	assertReactive(t, names, "props", "value")
	assertNotReactive(t, names, "flag")
}

// TestReactiveMarksValueAssignedUnderReactiveBranch pins the control-dominator rule.
//
// Both operands are constants, so the operand rule alone leaves `value` non-reactive. It is
// reactive because WHICH constant it holds depends on a reactive test, which is what the
// post-dominator frontier computes. A pass missing this reports a stale dependency array as
// correct.
func TestReactiveMarksValueAssignedUnderReactiveBranch(t *testing.T) {
	t.Parallel()

	names := reactiveNames(t, `
export function Component(props: {c: boolean}) {
  let value = 1;
  if (props.c) {
    value = 2;
  }
  return value;
}
`)
	assertReactive(t, names, "props", "value")
}

// TestReactiveLeavesConstantBranchAlone is the control-rule's negative case.
//
// Identical shape to the test above with a non-reactive test, so a pass that marked every joined
// value regardless of the branch would pass that one and fail this. Without this pair the control
// rule is asserted in only one direction.
func TestReactiveLeavesConstantBranchAlone(t *testing.T) {
	t.Parallel()

	names := reactiveNames(t, `
export function Component() {
  const flag = true;
  let value = 1;
  if (flag) {
    value = 2;
  }
  return value;
}
`)
	assertNotReactive(t, names, "value", "flag")
}

// TestReactiveMarksValueUnderReactiveSwitchDiscriminant is the discriminant half of the same rule.
//
// Every case test is a constant here and only the switched value is reactive, so the case-test
// branch cannot answer this one. With the case-test fixture above, each half of the switch arm now
// has an input the other cannot satisfy, which is what scoring them independently requires.
func TestReactiveMarksValueUnderReactiveSwitchDiscriminant(t *testing.T) {
	t.Parallel()

	names := reactiveNames(t, `
export function Component(props: {n: number}) {
  let value = 1;
  switch (props.n) {
    case 0: {
      value = 2;
      break;
    }
    default: {
      value = 3;
    }
  }
  return value;
}
`)
	assertReactive(t, names, "props", "value")
}

// TestReactiveMarksValueUnderReactiveSwitchCase covers a switch whose CASE TEST is reactive.
//
// Upstream's control-dominator test asks three terminal kinds, and for a `switch` it asks both the
// discriminant AND every case test. The case-test half is a separate branch, and a mutant disabling
// it survived every other fixture here.
//
// The discriminant is a local constant and only the case test is reactive, which is what isolates
// the branch: a fixture switching on `props.kind` would be caught by the discriminant half and
// would leave this one unmeasured.
func TestReactiveMarksValueUnderReactiveSwitchCase(t *testing.T) {
	t.Parallel()

	names := reactiveNames(t, `
export function Component(props: {n: number}) {
  const key = 0;
  let value = 1;
  switch (key) {
    case props.n: {
      value = 2;
      break;
    }
    default: {
      value = 3;
    }
  }
  return value;
}
`)
	assertReactive(t, names, "props", "value")
}

// TestReactiveMarksHookResults pins the hook rule: a hook can read state or context, so its result
// varies between renders even when its arguments are constant.
//
// `useState(0)`, `useContext(Ctx)` and a custom hook all take constant or non-reactive arguments, so
// nothing but the hook rule can make any of them reactive.
func TestReactiveMarksHookResults(t *testing.T) {
	t.Parallel()

	names := reactiveNames(t, `
import {useState, useContext, Context} from 'react';
declare const Ctx: Context<{a: number}>;
declare function useCustomThing(): number;
export function Component() {
  const [count, setCount] = useState(0);
  const ctx = useContext(Ctx);
  const custom = useCustomThing();
  return count + ctx.a + custom;
}
`)
	assertReactive(t, names, "count", "ctx", "custom")
}

// TestReactiveMarksNamespacedHookResults covers the `React.useState(0)` form.
//
// This is a MethodCall rather than a CallExpression, so it reaches a different arm of the hook test,
// and a mutant deleting that arm survived every other fixture in this file. The namespace import is
// not an exotic spelling: it is what `import * as React from 'react'` produces, which is one of the
// two shapes real code is written in.
//
// Upstream tests the call's PROPERTY rather than its receiver, which is what makes
// `React.useState(0)` a hook call while a receiver named `useThing` would not be. Probed on our own
// lowering: the property's identifier carries an empty Name and a node whose text is `useState`,
// which is the same shape the plain-call arm sees.
func TestReactiveMarksNamespacedHookResults(t *testing.T) {
	t.Parallel()

	names := reactiveNames(t, `
import * as React from 'react';
export function Component() {
  const [count, setCount] = React.useState(0);
  return count;
}
`)
	assertReactive(t, names, "count")
	assertNotReactive(t, names, "setCount")
}

// TestReactiveExemptsStableHookValues is the stable side-map, and it is the highest-value negative
// test in this file.
//
// The setter is destructured out of a call the hook rule has just marked reactive, so without the
// exemption it is reactive too, and every dependency array containing a setter becomes wrong. The
// three names are the three the checker can actually see: Dispatch, ActionDispatch, RefObject.
//
// `count` is asserted reactive in the same input as the control: it comes out of the same call, so
// a pass that exempted the whole call rather than the stable ELEMENTS would pass the negative
// assertions and fail this one.
func TestReactiveExemptsStableHookValues(t *testing.T) {
	t.Parallel()

	names := reactiveNames(t, `
import {useState, useRef, useReducer} from 'react';
declare function reducer(s: number, a: {type: string}): number;
export function Component() {
  const [count, setCount] = useState(0);
  const myRef = useRef(null);
  const [reduced, dispatch] = useReducer(reducer, 0);
  return count + reduced;
}
`)
	assertReactive(t, names, "count", "reduced")
	assertNotReactive(t, names, "setCount", "myRef", "dispatch")
}

// TestReactiveExemptsAliasedSetter pins that stability survives an alias, matching upstream's
// StableSidemap propagation through LoadLocal and StoreLocal.
func TestReactiveExemptsAliasedSetter(t *testing.T) {
	t.Parallel()

	names := reactiveNames(t, `
import {useState} from 'react';
export function Component() {
  const [count, setCount] = useState(0);
  const aliased = setCount;
  return count;
}
`)
	assertReactive(t, names, "count")
	assertNotReactive(t, names, "setCount", "aliased")
}

// TestReactiveDistinguishesHookNameFromOrdinaryCall pins the hook NAME test in both directions.
//
// A call to `notAHook` with constant arguments is not a source; a call to `useThing` with the same
// arguments is. Asserting only the positive would be satisfied by a pass that treated every call as
// a source, which is the natural wrong implementation here.
func TestReactiveDistinguishesHookNameFromOrdinaryCall(t *testing.T) {
	t.Parallel()

	names := reactiveNames(t, `
declare function useThing(n: number): number;
declare function notAHook(n: number): number;
export function Component() {
  const hookValue = useThing(1);
  const plainValue = notAHook(1);
  return hookValue + plainValue;
}
`)
	assertReactive(t, names, "hookValue")
	assertNotReactive(t, names, "plainValue")
}

// TestReactiveRequiresCapitalAfterUse pins that `use` alone is not a hook name.
//
// Upstream's convention is `use` followed by a capital or a digit, which `classifyFunction` already
// applies. A function named `used` or `user` must not read as a hook. This is also the reason the
// `use` OPERATOR is not reachable by this pass, recorded as a divergence at `isHookCallee`.
func TestReactiveRequiresCapitalAfterUse(t *testing.T) {
	t.Parallel()

	names := reactiveNames(t, `
declare function user(n: number): number;
declare function useThing(n: number): number;
export function Component() {
  const notAHook = user(1);
  const isAHook = useThing(1);
  return notAHook + isAHook;
}
`)
	assertReactive(t, names, "isAHook")
	assertNotReactive(t, names, "notAHook")
}

// TestReactiveIsIdempotent pins the property that separates this pass from Construct.
//
// `Construct` is not idempotent and the cache exists because of it. This one is, because the flags
// are written once from a settled set rather than as a side effect of the traversal. A caller that
// runs it twice must get the same answer, and the failure if it did not would be silent.
func TestReactiveIsIdempotent(t *testing.T) {
	t.Parallel()

	source := `
export function Component(props: {c: boolean, n: number}) {
  let value = 0;
  if (props.c) { value = props.n; }
  return value;
}
`
	var first, second []string
	probe := rule.Rule{
		Name:             "reactive-idempotent-harness",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					forEachFunctionLike(node, func(functionNode *ast.Node) {
						function := Lower(functionNode, ctx.TypeChecker)
						if function == nil {
							return
						}
						Construct(function)

						InferReactive(function, ctx.TypeChecker)
						found := map[string]bool{}
						collectReactiveNames(function, found)
						for name := range found {
							first = append(first, name)
						}

						InferReactive(function, ctx.TypeChecker)
						found = map[string]bool{}
						collectReactiveNames(function, found)
						for name := range found {
							second = append(second, name)
						}
					})
				},
			}
		},
	}
	rule_testing.RunTypedFiles(t, probe, map[string]string{
		"react.d.ts":  reactiveDeclarations,
		"fixture.tsx": source,
	}, "fixture.tsx")

	sort.Strings(first)
	sort.Strings(second)
	if len(first) == 0 {
		t.Fatal("the first run marked nothing reactive, so this test would compare two empty " +
			"slices and prove nothing")
	}
	if strings.Join(first, " ") != strings.Join(second, " ") {
		t.Errorf("InferReactive is not idempotent: first run [%s], second run [%s]",
			strings.Join(first, " "), strings.Join(second, " "))
	}
}

// TestReactiveWithoutCheckerStillMarksParameters pins the documented degradation.
//
// With no checker the stable exemption and the hook test cannot be made, so the answer shrinks to
// parameters and what derives from them. It must SHRINK rather than become wrong, and it must not
// panic: `stableTypeName` returning early on a nil checker is the guard, and a rule handed a nil
// checker going vacuously green is a failure mode this tree has shipped three times.
func TestReactiveWithoutCheckerStillMarksParameters(t *testing.T) {
	t.Parallel()

	source := `
export function Component(props: {n: number}) {
  const derived = props.n;
  return derived;
}
`
	var names []string
	probe := rule.Rule{
		Name: "reactive-no-checker-harness",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					forEachFunctionLike(node, func(functionNode *ast.Node) {
						// Lowered with no checker deliberately, which is the state this asserts.
						function := Lower(functionNode, nil)
						if function == nil {
							return
						}
						Construct(function)
						InferReactive(function, nil)
						found := map[string]bool{}
						collectReactiveNames(function, found)
						for name := range found {
							names = append(names, name)
						}
					})
				},
			}
		},
	}
	rule_testing.Run(t, probe, "fixture.tsx", source)

	sort.Strings(names)
	if !hasName(names, "props") {
		t.Errorf("with no checker a parameter must still be reactive, got [%s]", strings.Join(names, " "))
	}
}

// TestReactiveConvergesInFewRounds pins the termination property.
//
// The fixpoint's value is a set keyed on IdentifierId alone, and the whole reason for that choice is
// that a comparison including spans or freshly-minted identifiers never settles. The failure mode is
// a HANG rather than a wrong answer, so it is worth an assertion that would catch a future edit
// making the value richer.
//
// Measured over 685 lowered functions of real TypeScript: max 4 rounds, with 508 settling in 2 and
// only 5 needing 4, against a bound of 10,000. A loop shape needs one extra round to carry a back
// edge, which is where 3 and 4 come from. An assertion of a small constant here is therefore
// generous rather than tight, and a regression that made this iterate would blow through it.
func TestReactiveConvergesInFewRounds(t *testing.T) {
	t.Parallel()

	source := `
export function Component(props: {a: boolean, b: number, items: number[]}) {
  let total = 0;
  for (const item of props.items) {
    if (props.a) {
      total = total + item;
    } else {
      total = total + props.b;
    }
  }
  let extra = 0;
  while (total > 10) {
    extra = extra + total;
    total = total - 1;
  }
  return total + extra;
}
`
	worst := 0
	probe := rule.Rule{
		Name:             "reactive-rounds-harness",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					forEachFunctionLike(node, func(functionNode *ast.Node) {
						function := Lower(functionNode, ctx.TypeChecker)
						if function == nil {
							return
						}
						Construct(function)
						state := &reactivity{
							function:    function,
							typeChecker: ctx.TypeChecker,
							reactive:    map[static_single_assignment.IdentifierId]bool{},
							stable:      map[static_single_assignment.IdentifierId]bool{},
						}
						state.run()
						if state.rounds > worst {
							worst = state.rounds
						}
					})
				},
			}
		},
	}
	rule_testing.RunTypedFiles(t, probe, map[string]string{
		"react.d.ts":  reactiveDeclarations,
		"fixture.tsx": source,
	}, "fixture.tsx")

	if worst == 0 {
		t.Fatal("no function was analysed, so this test measured nothing")
	}
	if worst > 10 {
		t.Errorf("the fixpoint took %d rounds over a function with two loops; the lattice value is "+
			"probably no longer a plain identifier set, which is the shape that terminates", worst)
	}
}

// TestReactiveGapsAreNamed pins the declared gaps.
//
// The list is asserted rather than described so that closing a gap is a visible event: a pass that
// learned to compute mutation-derived reactivity would have to change this test, which is where the
// under-approximation is documented in the API.
func TestReactiveGapsAreNamed(t *testing.T) {
	t.Parallel()

	gaps := ReactiveGaps()
	if len(gaps) != 2 {
		t.Fatalf("expected exactly the two declared gaps, got %d", len(gaps))
	}
	if gaps[0] != ReactiveGapMutation || gaps[1] != ReactiveGapAliasing {
		t.Errorf("the declared gaps changed; if a gap was closed, update the package comment and "+
			"the ReactiveGap constants together, got %v", gaps)
	}
}

// TestReactiveIsAnUnderApproximation pins the DIRECTION of the mutation gap.
//
// Upstream marks `sink` reactive here: `sink` is passed to a call alongside a reactive argument, so
// its mutable operand is marked. This pass cannot, because every Place.Effect is EffectUnknown.
//
// The test asserts the case is silent rather than asserting it reports, which is the honest
// expectation, and it names the input so that a later pass filling Place.Effect turns this test red
// and forces the package comment to be revisited. A gap asserted as silence is findable; a gap left
// undocumented is not.
func TestReactiveIsAnUnderApproximation(t *testing.T) {
	t.Parallel()

	names := reactiveNames(t, `
declare function mutate(target: number[], source: number): void;
export function Component(props: {n: number}) {
  const sink: number[] = [];
  mutate(sink, props.n);
  return sink;
}
`)
	assertReactive(t, names, "props")
	if hasName(names, "sink") {
		t.Errorf("`sink` became reactive, which means mutation-derived reactivity is now being " +
			"computed. That is upstream's answer and a real improvement, but ReactiveGapMutation " +
			"and the package comment now describe a gap that no longer exists; update both.")
	}
}
