package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/high_level_intermediate_representation"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/static_single_assignment"
)

var messageSetStateInEffect = rule.Message{
	Id: "setStateInEffect",
	Description: "This effect calls a state setter synchronously in its body, so every time the " +
		"effect runs React starts another render, which runs the effect's dependencies again. " +
		"Effects are for synchronizing React with something outside it, such as the document " +
		"title, a subscription, or a network request, and a setter called straight from the body " +
		"is usually a value that could have been computed during render instead. Derive the value " +
		"while rendering, pass the initial value to `useState`, or set it from the event that " +
		"actually caused the change. Calling the setter from a callback the effect registers, " +
		"such as a subscription or a timer, is fine and is not reported.",
}

// SetStateInEffect flags a state setter called synchronously in the body of an effect.
//
//	valid:   function C() { const [s, setS] = useState(0); useEffect(() => { setTimeout(setS, 10); }); return s; }
//	valid:   function C() { const r = useRef(0); const [s, setS] = useState(0); useEffect(() => { setS(r.current + 1); }, []); return s; }
//	valid:   function C() { const [s, setS] = useState(0); useEffect(() => { const f = () => setS(1); f(); }); return s; }
//	invalid: function C() { const [s, setS] = useState(0); useEffect(() => { setS(s => s + 1); }); return s; }
//	invalid: function C() { const [s, setS] = useState(0); const f = () => setS(1); useEffect(() => { f(); }); return s; }
//	invalid: function C() { const [s, setS] = useState(0); React.useEffect(() => { setS(1); }); return s; }
//
// Calling a setter synchronously inside an effect body schedules a second render immediately after
// the first commits. The user sees the intermediate state paint, and if the effect's dependencies
// include anything the setter changes, the pair loops.
//
// # Where this comes from
//
// React's own `validateNoSetStateInEffects`, which is `ErrorCategory.EffectSetState`. oxc's linter
// rule at `crates/oxc_linter/src/rules/react/set_state_in_effect.rs` is the twenty-two-rule
// dispatcher body `run_react_compiler_rule(ctx, ErrorCategory::EffectSetState)`, so the judgment
// lives in the compiler crate on that side too.
//
//	react 7.1.1   node_modules/eslint-plugin-react-hooks/cjs/eslint-plugin-react-hooks.development.js
//	              `validateNoSetStateInEffects` at 46327, `getSetStateCall` at 46426
//	oxc           crates/oxc_react_compiler/src/react_compiler_validation/validate_no_set_state_in_effects.rs
//
// React is the authority here per the port brief's table. Every behavioural claim in this comment
// was established by RUNNING React 7.1.1 through the ESLint Linter API on the exact input named,
// not by reading, and reading gave the wrong answer twice: on which names count as effects, and on
// a function defined inside the effect body.
//
// # What upstream actually treats as an effect callback, measured rather than assumed
//
// This is the half of the rule its sibling `set-state-in-render` does not face, so it is stated
// precisely. Upstream's test is three type predicates, `isUseEffectHookType`,
// `isUseLayoutEffectHookType` and `isUseInsertionEffectHookType`, each comparing the callee's
// inferred type against a `shapeId` from React's builtin registry. A name reaches that registry
// through `Environment.getGlobalDeclaration`, which keys the lookup on the name a binding was
// DECLARED as: `binding.imported` for a named import, `binding.name` for a global or a namespace
// member. So the decision is: does this callee resolve to a declaration named exactly `useEffect`,
// `useLayoutEffect`, or `useInsertionEffect`.
//
// Four consequences, each measured on the input named rather than inferred from that sentence:
//
//   - **A local rename still reports.** `import {useEffect as useFx} from 'react'; useFx(...)`
//     reports, because `binding.imported` is `useEffect`. The local spelling is not consulted.
//   - **The reverse does not.** `import {somethingElse as useEffect} from 'react'` is SILENT,
//     which is the same rule read the other way and is the case that proves the local name is not
//     what is being asked.
//   - **A namespace member is the same question.** `React.useEffect(...)` reports, and needs no
//     syntactic special case here, because the checker resolves the property access to the same
//     `useEffect` declaration a named import resolves to.
//   - **A custom hook is NOT an effect, however it is spelled.** `useMyEffect2` does not report and
//     neither does `useEffective`. Both were run. This matters because a related rule,
//     `exhaustive-deps`, computes `isEffect` as the regex `/Effect($|[^a-z])/` at bundle line 940,
//     under which `useMyEffect2` WOULD be an effect and `useEffective` would not. That regex
//     belongs to that rule and is not consulted by this one; porting it here would report a class
//     of custom hooks upstream is silent on, and no fixture in either corpus contains either name.
//
// A locally declared `function useEffect` is silent, and an import of `useEffect` from a non-React
// module is silent too, both measured. Those two are NOT reproduced here and the divergence is
// deliberate; see "Where this diverges" below.
//
// # How a setter is identified, and why no React seed data is needed
//
// `@types/react` declares `useState<S>(): [S, Dispatch<SetStateAction<S>>]`, so the setter element
// of the returned tuple carries the type ALIAS `Dispatch` while the value element does not. That is
// the whole predicate. The checker is asked for the callee's type and the alias symbol's name is
// compared against `Dispatch`.
//
// **The alias is the correct field and the symbol is not**, which is worth stating because reaching
// for the symbol is the natural first move and it fails silently. Probed on a real program: a
// setter's `Type_symbol` is the anonymous type literal behind the alias, so keying on it reports a
// genuine setter as not-a-setter. The mirror-image trap belongs to the sibling rule `refs`, where
// `RefObject` is an interface and therefore lives in the SYMBOL with a correctly nil alias. One
// predicate cannot serve both, and an agent that wrote one for both reported a genuine `useRef` as
// neither.
//
// Two properties this buys with no extra code, both probed rather than assumed:
//
//   - **An alias chain is free.** `const aliased = setState; aliased(1)` still reads `Dispatch`,
//     because the checker propagates the type through the assignment. Upstream needs its
//     `LoadLocal`/`StoreLocal` bookkeeping to follow that; here it falls out of resolution. The
//     bookkeeping is still transcribed below, because it is also what carries a setter through a
//     function boundary, which the type alone does not do.
//   - **`useReducer` does not collide.** Its `dispatch` carries `ActionDispatch`, a different
//     alias, so it is correctly not a setter. Measured: a `dispatch` called in an effect is silent
//     upstream too.
//
// # Why this needs the intermediate representation
//
// The transitive case is the reason, and it is upstream's own shape. `const f = () => setS(1);`
// declared beside the effect and called as `f()` inside it REPORTS, at two levels of indirection as
// well. Answering that from syntax means tracking which locals hold a setter, following them
// through assignments, and then asking whether a value flowing into `useEffect`'s first argument
// transitively reaches one. That is `setStateFunctions`, upstream's map, and it is a dataflow
// question over values rather than a property of any one call site.
//
// The ref exemption below needs the graph for a second, independent reason.
//
// # The ref exemption, and the one place this deliberately stops short
//
// `enableAllowSetStateFromRefsInEffects` defaults to TRUE in the shipped rule (bundle line 31649),
// so the exemption is live rather than experimental, and it decides six of the twelve imported
// cases. It has two halves and both are reachable on their own, measured with the other held out:
//
//	value taint         useEffect(() => { setS(r.current + 1); }, [])   silent
//	control dominance   useEffect(() => { if (r.current) setS(1); }, []) silent
//	neither             useEffect(() => { const x = r.current; setS(1); }, []) REPORTS
//
// The third line is the control that makes the first two mean something: merely mentioning a ref
// in the effect does not exempt it.
//
// **Both halves are implemented here.** The value-taint half tracks a `refDerivedValues` set through
// the effect's blocks; the control-dominance half asks `high_level_intermediate_representation.ControlDominators` whether the block
// holding the setter is control-dependent on a branch whose test is ref-derived.
//
// The control half was missing for a while and the file said so, because
// `internal/utilities/hir/postdominator.go` had the tree and kept it unexported: it exposed
// `UnconditionalBlocks`, which answers "does this block lie on every path to a return", the question
// `set-state-in-render` asks. The frontier is a different question over the same tree and
// additionally needs the branch test's Place, so `ControlDominators` sits beside it on the shelf
// rather than inside this rule.
//
// Measured on the tree this gates: the control half moved 20 findings to 15, silencing the
// ref-sentinel shape `if (previousReference.current !== value) setValue(value)`. Tainting a store's
// own lvalue took it to 10, silencing the measure idiom, `const element = reference.current` and
// then a read of `element.scrollHeight`, which cannot be computed during render at all.
//
// # Behaviours that look like defects and are upstream's
//
//   - **A function defined INSIDE the effect is never followed.** `useEffect(() => { const f = ()
//     => setS(1); f(); })` is SILENT, while moving that same declaration one line up, outside the
//     effect, REPORTS. Measured on both inputs. The mechanism is visible in the algorithm:
//     `getSetStateCall` walks the effect function's own blocks, and a function declared inside it
//     is a nested `FunctionExpression` whose body those blocks do not contain. The map it consults
//     was populated in the ENCLOSING function, so only setters that crossed the boundary from
//     outside are known. This is a real hole rather than an exemption, and it is reproduced.
//   - **A `useCallback` or `useMemo` wrapper is invisible to both sides**, because upstream deletes
//     it before this validator runs and this rule now reads a graph prepared the same way: the
//     erasure, then the inlining pass that follows it one line later upstream. See
//     `high_level_intermediate_representation.ForFunctionWithoutManualMemoization`; the fixtures are
//     `TestSetStateInEffectSeesThroughManualMemoization` and
//     `TestSetStateInEffectReachesThroughAMemoizedCallbackThatReturnsAFunction`.
//   - **A setter reached through a callback is silent, by the same mechanism running the other
//     way.** `setTimeout(setS, 10)` and `subscribe(() => setS(1))` are both clean, which is the
//     rule's actual purpose: the setter is not called during the effect body.
//   - **One finding per effect, not per call.** Three `setS` calls in one effect report ONCE,
//     because `getSetStateCall` returns the FIRST setter it finds and stops. Two separate effects
//     report twice. Both measured.
//   - **The dependency array is not consulted at all.** `useEffect(cb)` and `useEffect(cb, [])`
//     behave identically, and a setter passed as the SECOND argument is silent.
//
// # Where this diverges, and why
//
// Upstream distinguishes an import of `useEffect` from React from a same-named local declaration or
// an import from an unrelated module, and is silent on the latter two. That distinction lives in
// `Environment.getGlobalDeclaration`'s switch over binding kinds. This rule cannot make it: our
// lowering leaves `LoadGlobal.BindingKind` at `Global` for every global, never populating `Source`
// or `Imported`, documented at `internal/utilities/hir/lower.go:58` and confirmed still true here.
//
// The divergence is therefore that a function named exactly `useEffect`, `useLayoutEffect` or
// `useInsertionEffect` that is NOT React's would be treated as an effect. Its practical reach is
// bounded by the other half of the rule: a finding still requires a genuine `Dispatch`-typed setter
// from a real `useState`, so the false-positive surface is a file that declares its own `useEffect`
// AND calls a real React setter synchronously inside it. No instance appeared in the dry run
// described below, which is a weaker statement than "none exists" and is deliberately left weaker.
//
// # What the dry run actually found, which is the honest summary of this port
//
// Run over Kirk's tree with `--no-fix --lint --timing`: 20 findings over 3,407 files, and every one
// was read rather than counted. It does not explode.
//
// That figure predates both the manual-memoization erasure and the ref work that landed beside it.
// Re-measured after the erasure: 2 findings over 3,482 files, both `useCallback`-wrapped async
// fetches called from an effect, both confirmed against React 7.1.1. One of the two is reported by
// React as well; the other is not, and the difference is the ref exemption rather than this change.
//
// The cost figure below is also from that first run and has moved since, in the rule's favour: the
// same tree now attributes about 40ms to this rule, and `refs` is the expensive one at roughly
// 1.5s. The per-rule numbers in this paragraph came from a `--timing` run, which times every
// listener call and is slower than a real one, so they compare rules to each other and are not the
// tool's speed. What has not changed is the reason they are what they are, which is why the
// paragraph stays.
//
// **It costs 888ms, which is 10% of all rule time and the most expensive rule in the run.** That is
// stated plainly rather than buried, but it is a property of the approach rather than of this rule:
// its two nearest neighbours are `set-state-in-render` at 571ms and `static-components` at 532ms,
// and those three are exactly the three rules built on the intermediate representation. Every
// syntactic react rule in the same run is between 1ms and 20ms. Lowering every function in the file
// is the cost, it is paid once per rule today because each lowers independently, and the obvious
// saving is a per-file cache of the lowered form shared across the three. This rule sits above the
// other two because it also asks the checker a type question per call site.
//
// **Roughly three quarters of them are the post-dominator gap above, not new information.** Each
// was reproduced as a minimal shape and run against React 7.1.1, and the ref-controlled ones are
// silent upstream: a measurement effect reading `ref.current` and setting a derived flag, a panel
// that opens when a `previousXReference` transitions, a focus flag set behind a DOM identity check.
// Those are false positives by upstream's standard and they are the cost of shipping the value half
// of the ref exemption without the control half.
//
// **A real minority are true positives that React reports too**, verified the same way: an
// unguarded functional `setNodes` in an effect keyed on a prop, at
// `OsKingdomGraphCanvas.tsx:242`, is the clearest, and React reports the reproduced shape.
//
// A third group is the custom-hook divergence: a setter reached through a hook or a context object
// rather than a direct `useState`, which the checker still types as `Dispatch` and upstream has
// lost the shape of. `DropArea.tsx` is that shape, measured silent upstream.
//
// The proportion is recorded because it is the fact a reader most needs and the one most likely to
// go stale: **this rule is materially noisier than upstream until post-dominators land**, and the
// fixture `refControlledBranchIsAKnownDivergence` is the thing that will fail when they do.
//
// Recorded rather than smoothed over, and a differential run against a tool implementing upstream's
// binding-kind switch or its control dominators will show differences here for those reasons.
var SetStateInEffect = rule.Rule{
	Name:             "react-hooks/set-state-in-effect",
	NeedsTypeChecker: true,
	// Reads other files only through shape readers (rule.ExportNameIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,
	// Same predicate as `set-state-in-render`: the setter is the type's alias, so an `any` hook
	// return takes every finding in the file with it. Declared so the tripwire can name it.
	ResolvesReactValueTypes: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			// The whole file at once, matching `static-components`. Lowering already descends into
			// nested functions through the function arena, so listening per function kind would
			// lower every inner function twice, once as a child and once standalone without the
			// bindings its parent supplies.
			ast.KindSourceFile: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				// A file that cannot hold a component or a hook is not lowered at all; see
				// high_level_intermediate_representation.MayHoldComponentOrHook for why that is exact.
				if !high_level_intermediate_representation.MayHoldComponentOrHook(ctx) {
					return
				}
				forEachCompiledFunction(node, func(functionNode *ast.Node) {
					// Manual memoization is erased first, because upstream validates a graph
					// where `useMemo` and `useCallback` are already gone; see
					// high_level_intermediate_representation.ForFunctionWithoutManualMemoization for why that is a separate cache
					// entry rather than a step in the shared one.
					lowered := high_level_intermediate_representation.ForFunctionWithoutManualMemoization(ctx, functionNode)
					if lowered == nil {
						return
					}
					analyzeSetStateInEffectSubject(ctx, lowered)
				})
			},
		}
	},
}

// analyzeSetStateInEffectSubject runs the validator over the functions React would have compiled.
//
// Upstream reaches `validateNoSetStateInEffects` only for a function it compiles, so the gate is the
// react shelf's `IsComponentOrHookLike`, the same one purity and set-state-in-render take. A declined
// function is descended into, because a component nested inside a plain wrapper is still a unit.
//
// This rule judged every function before, on the reasoning that a finding already needs a real
// setter and a real effect hook, which looked narrower than "is a component". It is not narrower in
// the way that matters: the effect hook is classified by its type, so a value typed `typeof
// useEffect` reaches it from a function that is no component and spells no hook name, and upstream
// is silent there, measured on `function notAComponent() { const [s, setS] = useState(0);
// useEffect(() => setS(1)); }`. Gating also makes `hir.MayHoldComponentOrHook` exact for this rule,
// so a file that cannot hold a component is not lowered at all (#1pmwkmv).
func analyzeSetStateInEffectSubject(ctx rule.Context, function *high_level_intermediate_representation.Function) {
	if function == nil {
		return
	}
	if function.Node == nil || !utilsreact.IsComponentOrHookLike(function.Node) {
		for _, nested := range function.Functions {
			analyzeSetStateInEffectSubject(ctx, nested)
		}
		return
	}
	// A unit found nested in a function that is not one is analyzed as React Compiler compiles it,
	// on its own; see AsCompilationUnit.
	function = high_level_intermediate_representation.AsCompilationUnit(ctx, function,
		high_level_intermediate_representation.ForFunctionWithoutManualMemoization)
	reportSetStateInEffects(ctx, function)
}

// reportSetStateInEffects is upstream's `validateNoSetStateInEffects` over one compiled function.
func reportSetStateInEffects(ctx rule.Context, function *high_level_intermediate_representation.Function) {
	if function == nil {
		return
	}

	// Value to the Place that made it a setter. Upstream's `setStateFunctions`, keyed the same way.
	// Single-assignment form is what lets this be one entry per value rather than per binding.
	setStateFunctions := map[static_single_assignment.IdentifierId]high_level_intermediate_representation.Place{}

	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			target := instruction.LValue.Identifier

			switch value := instruction.Value.(type) {
			// Reads and writes carry setter-ness along, which is what makes an alias chain of any
			// length behave like the setter at its head. Transcribed from upstream even though the
			// checker already follows a direct alias, because this is also what carries a setter
			// into the map that the nested-function scan below consults.
			case *high_level_intermediate_representation.LoadLocal:
				if origin, ok := setStateFunctions[value.Place.Identifier]; ok {
					setStateFunctions[target] = origin
				}
			// Same capture path as in `findSetStateCall`, for an effect written inside a nested
			// function that closes over a setter from further out.
			//
			// Only the MAP branch is here. A type check was written beside it and removed: the
			// mutation deleting it survived the whole suite, and the reason is subsumption rather
			// than a fixture gap. Every value this arm could newly identify by type is a value that
			// gets asked the same type question again at the point it is CALLED, by
			// `isSetterPlace`, whose own type branch is caught by 22 lines. A captured setter that
			// is never called cannot produce a finding by any route, so no input can distinguish
			// the two versions.
			case *high_level_intermediate_representation.LoadContext:
				if origin, ok := setStateFunctions[value.Place.Identifier]; ok {
					setStateFunctions[target] = origin
				}
			case *high_level_intermediate_representation.StoreLocal:
				if origin, ok := setStateFunctions[value.Value.Identifier]; ok {
					setStateFunctions[value.LValue.Identifier] = origin
					setStateFunctions[target] = origin
				}

			// A function closing over a setter may itself BE a setter call, which is the transitive
			// case: `const f = () => setS(1)` makes `f` behave as a setter for this rule's purpose.
			// Upstream gates the descent on an operand already being a setter, so a function that
			// closes over nothing relevant is never scanned.
			//
			// # The gate is a COST guard, not a behavioural one, and its mutant survives
			//
			// Deleting it survives the whole suite, and the reason is the brief's "both branches
			// reach the same verdict by different routes" rather than a fixture gap. Without the
			// gate, a callback closing over no setter is scanned anyway and then declines inside
			// `findSetStateCall` on the same type test; with it, the scan never starts. Silence
			// either way, so no assertion on findings can separate them.
			//
			// What the gate actually buys is that every function expression in the file is not
			// lowered-and-walked on the chance it calls a setter, which is a real cost on a tree
			// where most callbacks are event handlers. It is upstream's condition, it is kept for
			// fidelity, and the fixture that would "cover" it is in the silent test asserting the
			// observable behaviour both versions share.
			case *high_level_intermediate_representation.FunctionExpression:
				if !anyOperandIsSetter(ctx, function, value, setStateFunctions) {
					continue
				}
				nested := setStateInEffectNestedFunction(function, value)
				if nested == nil {
					continue
				}
				// The Place stored is the one found INSIDE the nested function, and it is what
				// upstream stores too. It is only ever reported when this value is itself the
				// callback handed to an effect; when this value is instead CALLED from a later
				// effect body, that body's own call site supplies the span. See the note on where
				// findings point in `findSetStateCall`.
				if callee, found := findSetStateCall(ctx, nested, translateAcrossCaptures(value, nested, setStateFunctions)); found {
					setStateFunctions[target] = callee
				}

			case *high_level_intermediate_representation.CallExpression:
				handleEffectCall(ctx, function, target, value.Callee, value.Args, setStateFunctions)
			case *high_level_intermediate_representation.MethodCall:
				// `React.useEffect(...)`. The property Place is the callee, exactly as upstream
				// selects `instr.value.property` for a MethodCall.
				handleEffectCall(ctx, function, target, value.Property, value.Args, setStateFunctions)
			}
		}
	}

	// A nested function is lowered into this function's arena, but its own body is a separate graph
	// that the walk above never entered. Upstream reaches it because the compiler lowers each
	// component or hook and this validator runs per lowered function; here the arena is walked so an
	// effect written inside a nested component is still seen.
	for _, nested := range function.Functions {
		reportSetStateInEffects(ctx, nested)
	}
}

// handleEffectCall is the CallExpression and MethodCall arm of upstream's switch.
//
// Two hooks are recognized and they do opposite things. `useEffectEvent` PROPAGATES setter-ness to
// its result, so a setter reached through one is still reported when that result is called inside
// an effect. The three effect hooks REPORT. Both read only the first argument, and upstream reads
// it only when it is a plain identifier rather than an inline function, which is what makes
// `useEffect(() => {...})` reach the report through the FunctionExpression arm above rather than
// here.
func handleEffectCall(
	ctx rule.Context,
	function *high_level_intermediate_representation.Function,
	target static_single_assignment.IdentifierId,
	callee high_level_intermediate_representation.Place,
	args []high_level_intermediate_representation.Argument,
	setStateFunctions map[static_single_assignment.IdentifierId]high_level_intermediate_representation.Place,
) {
	firstArgument, hasFirst := firstIdentifierArgument(args)
	if !hasFirst {
		return
	}
	setter, isSetter := setStateFunctions[firstArgument]
	if !isSetter {
		return
	}

	switch effectHookKind(ctx, function, callee) {
	case effectHookUseEffectEvent:
		setStateFunctions[target] = setter
	case effectHookEffect:
		reportSetter(ctx, function, setter)
	}
}

// findSetStateCall is upstream's `getSetStateCall`, minus the control-dominance half.
//
// It scans one function's own blocks for a call whose callee is a setter, and returns the FIRST
// one. Returning the first rather than collecting all of them is why three setter calls in one
// effect produce one finding, which was measured rather than inferred.
//
// The ref exemption's value-taint half lives here: a value derived from a ref taints forward, and a
// setter call whose first argument is ref-derived is skipped. `refDerivedValues` is upstream's set
// of the same name.
func findSetStateCall(
	ctx rule.Context,
	function *high_level_intermediate_representation.Function,
	setStateFunctions map[static_single_assignment.IdentifierId]high_level_intermediate_representation.Place,
) (high_level_intermediate_representation.Place, bool) {
	if function == nil {
		return high_level_intermediate_representation.Place{}, false
	}

	refDerived := map[static_single_assignment.IdentifierId]bool{}
	local := map[static_single_assignment.IdentifierId]high_level_intermediate_representation.Place{}
	for id, place := range setStateFunctions {
		local[id] = place
	}

	isDerivedFromRef := func(place high_level_intermediate_representation.Place) bool {
		return refDerived[place.Identifier] || isRefTyped(ctx, function, place)
	}

	// The control-dominance half of the ref exemption, upstream's `isRefControlledBlock`.
	//
	// Built here rather than per call so its per-block cache is shared, and built over a closure
	// that reads `refDerived` live: the taint map is still filling as the blocks below are walked,
	// and a branch on `reference.current` taints through the same loop that later reaches the
	// setter. The predicate is only ever CALLED at a setter, by which point every instruction above
	// it in the walk has been seen.
	isRefControlledBlock := high_level_intermediate_representation.ControlDominators(function, isDerivedFromRef)

	for _, block := range function.Blocks {
		// A phi joining a ref-derived operand is ref-derived. Upstream does the same, and it is
		// what carries the exemption across an `if` that assigns from a ref on one path.
		for _, phi := range block.Phis {
			for _, entry := range phi.Operands {
				if refDerived[entry.Place.Identifier] {
					refDerived[phi.Place.Identifier] = true
					break
				}
			}
		}

		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			target := instruction.LValue.Identifier

			// `ref.current` is the canonical ref read. Upstream special-cases exactly this shape,
			// keying on the property name and the object being ref-typed.
			if load, isLoad := instruction.Value.(*high_level_intermediate_representation.PropertyLoad); isLoad {
				if load.Property == "current" && isRefTyped(ctx, function, load.Object) {
					refDerived[target] = true
				}
			}

			// Any instruction reading a ref-derived operand produces a ref-derived value, which is
			// what makes `r.current + 1` and `arr[r.current]` exempt.
			if instructionReadsRefDerived(instruction, isDerivedFromRef) {
				refDerived[target] = true

				// A store binds through its OWN lvalue as well as the instruction's, and both have
				// to be tainted. Upstream reaches them together through `eachInstructionLValue`,
				// which yields `instr.lvalue` and then the value's own lvalue; marking only the
				// first loses the binding a reader actually named.
				//
				// The gap was measured rather than reasoned about: `setV(r.current.scrollHeight)`
				// was exempt while `const e = r.current; setV(e.scrollHeight)` reported, which is
				// the same code with a name on the intermediate. Four sites in the tree this gates
				// are the second spelling, all of them the measure idiom, and `scrollHeight` cannot
				// be read during render at all.
				//
				// The setter map two hundred lines up already does this for its own `StoreLocal`
				// arm, which is what makes the omission here a slip rather than a decision.
				if store, isStore := instruction.Value.(*high_level_intermediate_representation.StoreLocal); isStore {
					refDerived[store.LValue.Identifier] = true
				}

				// A destructure binds through a PATTERN rather than through the instruction's
				// LValue, so tainting only the LValue loses every binding it introduced. That gap
				// was not hypothetical: `const {height} = ref.current.getBoundingClientRect()` is
				// upstream's own `valid-set-state-in-useEffect-from-ref` fixture, and it was the
				// first thing this rule got wrong, reporting two clean cases while the three
				// non-destructuring ref fixtures beside them passed. Upstream reaches these places
				// through `eachInstructionLValue`, which walks the pattern.
				// `LValue` only. `Destructure` carries a second `Pattern` field, and walking both
				// was written here first because a struct with two pattern-shaped fields reads as
				// two different patterns.
				//
				// It is not. All four sites that construct a `Destructure` in lowering assign the
				// SAME `pattern` local to both fields (`lower.go:498,640,897` and
				// `lower_expression.go:432`), so the second walk can never mark anything the first
				// did not. A mutation deleting either call survived the whole suite while deleting
				// both was caught by six lines, which is the signature of a subsumed branch rather
				// than a fixture blind spot, and the construction sites confirm it mechanically
				// rather than by argument.
				//
				// Removed rather than kept with a fixture, because a test written for a branch that
				// no input can reach asserts nothing.
				if destructure, isDestructure := instruction.Value.(*high_level_intermediate_representation.Destructure); isDestructure {
					markPatternRefDerived(destructure.LValue, refDerived)
				}
			}

			switch value := instruction.Value.(type) {
			case *high_level_intermediate_representation.LoadLocal:
				if origin, ok := local[value.Place.Identifier]; ok {
					local[target] = origin
				}

			// A setter reaches a nested function as a CAPTURE, so inside that function it is read
			// with LoadContext rather than LoadLocal, and the value it produces is a fresh
			// temporary with no syntax node of its own.
			//
			// This arm is the reason three fixtures failed at once on the first run, including two
			// transcribed verbatim from upstream, and the failure did not look like one cause. The
			// captured place keeps the identifier the enclosing function knew it by, so the type
			// question can be asked of that place while the CALL is attributed to the temporary.
			// Upstream has no equivalent arm because its `setStateFunctions` map is threaded into
			// the nested scan by the caller rather than rebuilt from the graph.
			case *high_level_intermediate_representation.LoadContext:
				// Keyed by the temporary this defines, VALUED by the captured place, because that
				// place carries the span of the name as written at THIS call site. Upstream returns
				// `callee`, the Place inside the function doing the calling, and reporting the
				// original setter instead points the finding at the `useState` line rather than at
				// the line the reader has to change.
				//
				// Measured against React on `const aliased = setState; useEffect(() => aliased(1))`:
				// the finding underlines `aliased` at 6:5, not `setState`. A message-id fixture
				// cannot see the difference, and this rule had it backwards until the span
				// assertion was written.
				// As in the outer pass, the type check that stood beside this was removed as
				// subsumed by `isSetterPlace` at the call site: deleting it survived every fixture
				// while deleting `isSetterPlace`'s own type branch was caught by 22 lines.
				//
				// What remains is load-bearing for a different reason, and its mutant IS caught by
				// eight lines: the entry it writes is keyed by the temporary but VALUED by the
				// captured place, which is what puts the finding on the name as written here rather
				// than on the `useState` line.
				if _, ok := local[value.Place.Identifier]; ok {
					local[target] = value.Place
				}

			case *high_level_intermediate_representation.StoreLocal:
				if origin, ok := local[value.Value.Identifier]; ok {
					local[value.LValue.Identifier] = origin
					local[target] = origin
				}
			case *high_level_intermediate_representation.CallExpression:
				if !isSetterPlace(ctx, function, value.Callee, local) {
					continue
				}
				// The value-taint half of the ref exemption: a setter fed a ref-derived value is
				// synchronizing React with something outside it, which is what effects are for.
				if argument, ok := firstIdentifierArgument(value.Args); ok && refDerived[argument] {
					return high_level_intermediate_representation.Place{}, false
				}
				// The control-dominance half: a setter that only runs when a branch on a ref says
				// so is the same synchronization written the other way. `if (previous.current !==
				// value) setValue(value)` reaches here with an argument that is not ref-derived,
				// and upstream is silent on it because the ref decides whether the call happens.
				if isRefControlledBlock(block.Id) {
					return high_level_intermediate_representation.Place{}, false
				}
				return value.Callee, true
			}
		}
	}
	return high_level_intermediate_representation.Place{}, false
}

// instructionReadsRefDerived reports whether any operand of this instruction is ref-derived.
//
// Upstream reaches this through `eachInstructionValueOperand`, a generic operand iterator over its
// whole instruction set. This handles the shapes a ref value can actually flow through in the
// corpus and in the probes, and declines the rest rather than pretending to be exhaustive. Naming
// the covered set explicitly is deliberate: a silent default that answered "no operands" for an
// unhandled instruction would drop the exemption and produce a false positive, which is the
// direction that costs a user rather than the direction that costs a finding.
func instructionReadsRefDerived(instruction *high_level_intermediate_representation.Instruction, isDerived func(high_level_intermediate_representation.Place) bool) bool {
	switch value := instruction.Value.(type) {
	case *high_level_intermediate_representation.LoadLocal:
		return isDerived(value.Place)
	case *high_level_intermediate_representation.StoreLocal:
		return isDerived(value.Value)
	case *high_level_intermediate_representation.PropertyLoad:
		return isDerived(value.Object)
	case *high_level_intermediate_representation.Destructure:
		return isDerived(value.Value)
	case *high_level_intermediate_representation.ComputedLoad:
		return isDerived(value.Object) || isDerived(value.Property)
	case *high_level_intermediate_representation.BinaryExpression:
		return isDerived(value.Left) || isDerived(value.Right)
	case *high_level_intermediate_representation.UnaryExpression:
		return isDerived(value.Value)
	case *high_level_intermediate_representation.CallExpression:
		if isDerived(value.Callee) {
			return true
		}
		for _, argument := range value.Args {
			if isDerived(argument.Place) {
				return true
			}
		}
	case *high_level_intermediate_representation.MethodCall:
		if isDerived(value.Receiver) || isDerived(value.Property) {
			return true
		}
		for _, argument := range value.Args {
			if isDerived(argument.Place) {
				return true
			}
		}
	}
	return false
}

// translateAcrossCaptures rewrites a setter map from one function's value space into a nested
// function's, using the capture edge.
//
// **A value's identifier does not survive a function boundary**, and this is the trap that made two
// upstream fixtures fail with a completely plausible-looking implementation. `f` in the enclosing
// function is `f$22`; the same binding inside the closure that captured it is `f$1`. Handing the
// enclosing map straight to a scan of the nested body looks correct, compiles, and matches on
// nothing, because every key names a value the nested function's table has never heard of.
//
// The edge is documented at `internal/utilities/hir/lower.go:315`: `Captures[i]` and `nested.Context[i]`
// name the same source binding seen from the two sides, in the same order, and the order is
// guaranteed rather than incidental. So the translation is an index-wise walk of those two slices.
//
// Upstream never needs this because its `setStateFunctions` map is threaded down by the caller
// while it lowers, so both sides are in one identifier space. Ours is rebuilt from the graph, which
// is the cost of reading a representation rather than building one.
//
// The lengths are compared rather than assumed equal: a partially lowered function is exactly the
// input a linter is handed, and a mismatched pair would otherwise index out of range.
func translateAcrossCaptures(
	value *high_level_intermediate_representation.FunctionExpression,
	nested *high_level_intermediate_representation.Function,
	outer map[static_single_assignment.IdentifierId]high_level_intermediate_representation.Place,
) map[static_single_assignment.IdentifierId]high_level_intermediate_representation.Place {
	inner := map[static_single_assignment.IdentifierId]high_level_intermediate_representation.Place{}
	if value == nil || nested == nil {
		return inner
	}
	limit := min(len(value.Captures), len(nested.Context))
	for i := range limit {
		if origin, known := outer[value.Captures[i].Identifier]; known {
			// Keyed by the INNER identifier, valued by the OUTER place, so the finding still points
			// at a span in a file rather than at a value in a closure's private table.
			inner[nested.Context[i].Identifier] = origin
		}
	}
	return inner
}

// markPatternRefDerived taints every binding a destructuring pattern introduces.
//
// Upstream walks the same shapes through `eachInstructionLValue`. Rest and default places are
// included: `const {height = fallback} = ref.current...` binds `height` from a ref on the path where
// the property exists, and a pass that skipped the default would exempt the one path and not the
// other, which is a distinction upstream does not draw either.
func markPatternRefDerived(pattern high_level_intermediate_representation.Pattern, refDerived map[static_single_assignment.IdentifierId]bool) {
	switch shape := pattern.(type) {
	case *high_level_intermediate_representation.PlacePattern:
		refDerived[shape.Place.Identifier] = true
	case *high_level_intermediate_representation.ObjectPattern:
		for _, property := range shape.Properties {
			if property.Value != nil {
				markPatternRefDerived(property.Value, refDerived)
			}
			if property.Default != nil {
				refDerived[property.Default.Identifier] = true
			}
		}
		if shape.Rest != nil {
			refDerived[shape.Rest.Identifier] = true
		}
	case *high_level_intermediate_representation.ArrayPattern:
		for _, element := range shape.Elements {
			// A hole from `[a, , b]` has a nil Value and binds nothing.
			if element.Value != nil {
				markPatternRefDerived(element.Value, refDerived)
			}
			if element.Default != nil {
				refDerived[element.Default.Identifier] = true
			}
		}
		if shape.Rest != nil {
			refDerived[shape.Rest.Identifier] = true
		}
	}
}

// firstIdentifierArgument returns the first argument's value when it is a plain, non-spread one.
//
// Upstream reads `args[0]` and proceeds only when `arg.kind === 'Identifier'`, which excludes a
// spread. A spread is excluded here for the same reason and it is not an approximation: `useEffect(
// ...rest)` cannot be attributed to a known value.
//
// # The spread half of this condition is EQUIVALENT, and it is kept anyway
//
// A mutation deleting `|| args[0].Spread` survives every fixture, and it survives because no input
// can distinguish the two versions rather than because the corpus is thin. The one-sentence reason:
// a spread's operand is the ITERABLE being spread, never the callback itself, so it can never be a
// value `setStateFunctions` holds, and consulting that map is the only thing the guard's absence
// would newly do.
//
// That was probed rather than argued, because the brief is right that a confident equivalence
// argument reads exactly like a correct one. Two fixtures were written for it first and both
// stayed green under the mutant, which is the signal that the hypothesis rather than the coverage
// was wrong. The strongest candidate inputs — `useEffect(...[f])` and `useEffect(...f)` where `f`
// IS registered in the map — were then run against the pristine and mutated rule side by side and
// produced byte-identical output, zero findings from both.
//
// It is kept because it is upstream's condition and this is a port: fidelity to the decision is the
// standard, and a reader comparing the two implementations should find the same test in both. Both
// spread fixtures are also kept, in `TestSetStateInEffectSpreadArgument`, since they assert real
// upstream behaviour even though they cannot see this particular branch.
func firstIdentifierArgument(args []high_level_intermediate_representation.Argument) (static_single_assignment.IdentifierId, bool) {
	if len(args) == 0 || args[0].Spread {
		return 0, false
	}
	return args[0].Place.Identifier, true
}

// anyOperandIsSetter reports whether a function expression closes over a known or typed setter.
//
// Upstream's guard before descending into a nested function. `Captures` is what makes this work and
// it only started being populated one commit before this rule was written; before that it was
// always empty and this branch could never fire.
func anyOperandIsSetter(
	ctx rule.Context,
	function *high_level_intermediate_representation.Function,
	value *high_level_intermediate_representation.FunctionExpression,
	setStateFunctions map[static_single_assignment.IdentifierId]high_level_intermediate_representation.Place,
) bool {
	for _, capture := range value.Captures {
		if _, known := setStateFunctions[capture.Identifier]; known {
			return true
		}
		if isSetterTyped(ctx, function, capture) {
			return true
		}
	}
	// A nested function's own captured context is the other way a setter crosses the boundary,
	// which is the two-level case `g` calls `f` calls `setS`.
	nested := setStateInEffectNestedFunction(function, value)
	if nested == nil {
		return false
	}
	for _, contextPlace := range nested.Context {
		if _, known := setStateFunctions[contextPlace.Identifier]; known {
			return true
		}
		if isSetterTyped(ctx, function, contextPlace) {
			return true
		}
	}
	return false
}

// isSetterPlace reports whether a callee is a setter, by type or by the propagation map.
func isSetterPlace(
	ctx rule.Context,
	function *high_level_intermediate_representation.Function,
	place high_level_intermediate_representation.Place,
	known map[static_single_assignment.IdentifierId]high_level_intermediate_representation.Place,
) bool {
	if _, ok := known[place.Identifier]; ok {
		return true
	}
	return isSetterTyped(ctx, function, place)
}

// setStateTypeAliasName is the alias `@types/react` puts on a `useState` setter.
//
// `useState<S>(): [S, Dispatch<SetStateAction<S>>]`, so the setter element carries `Dispatch` and
// the value element does not. Written as the alias name rather than asked of any React-specific
// seed table, which is the whole reason this rule needs no builtin registry.
//
// `useReducer` returns `ActionDispatch`, a distinct alias, so it does not match. That is upstream's
// answer too: a `dispatch` called in an effect is silent.
const setStateTypeAliasName = "Dispatch"

// refTypeSymbolName is the type `@types/react` puts on a ref.
//
// This reads the SYMBOL where the setter reads the ALIAS, and the asymmetry is real rather than an
// inconsistency to tidy up. `RefObject` is an interface, so it has no alias and its name lives in
// the symbol; `Dispatch` is a type alias, so its name lives in the alias and its symbol is the
// anonymous literal behind it. Using one field for both reports a genuine ref as not-a-ref, which
// is a defect an agent shipped on this exact pair before it was measured.
const refTypeSymbolName = "RefObject"

// isSetterTyped asks the checker whether a value is a `useState` setter.
func isSetterTyped(ctx rule.Context, function *high_level_intermediate_representation.Function, place high_level_intermediate_representation.Place) bool {
	return typeAliasNameOf(ctx, function, place) == setStateTypeAliasName
}

// isRefTyped asks the checker whether a value is a ref object.
func isRefTyped(ctx rule.Context, function *high_level_intermediate_representation.Function, place high_level_intermediate_representation.Place) bool {
	node := setStateInEffectIdentifierNodeOf(function, place.Identifier)
	if node == nil || ctx.TypeChecker == nil {
		return false
	}
	valueType := ctx.TypeChecker.GetTypeAtLocation(node)
	if valueType == nil {
		return false
	}
	symbol := checker.Type_symbol(valueType)
	return symbol != nil && symbol.Name == refTypeSymbolName
}

// typeAliasNameOf returns the name of the type alias on a value, or the empty string.
func typeAliasNameOf(ctx rule.Context, function *high_level_intermediate_representation.Function, place high_level_intermediate_representation.Place) string {
	node := setStateInEffectIdentifierNodeOf(function, place.Identifier)
	if node == nil || ctx.TypeChecker == nil {
		return ""
	}
	valueType := ctx.TypeChecker.GetTypeAtLocation(node)
	if valueType == nil {
		return ""
	}
	alias := checker.Type_alias(valueType)
	if alias == nil {
		return ""
	}
	symbol := alias.Symbol()
	if symbol == nil {
		return ""
	}
	return symbol.Name
}

// effectHookKindOf classifies a callee as an effect hook, `useEffectEvent`, or neither.
type effectHookKindOf int

const (
	effectHookNone effectHookKindOf = iota
	effectHookEffect
	effectHookUseEffectEvent
)

// effectHookKind resolves a callee to the name it was DECLARED as and classifies it.
//
// Resolving through the checker rather than reading the written spelling is what makes
// `import {useEffect as useFx}` report and `import {somethingElse as useEffect}` stay silent, which
// is upstream's `binding.imported` rule. It is also what makes `React.useEffect` need no separate
// member-expression case: the property access resolves to the same declaration.
func effectHookKind(ctx rule.Context, function *high_level_intermediate_representation.Function, callee high_level_intermediate_representation.Place) effectHookKindOf {
	node := setStateInEffectIdentifierNodeOf(function, callee.Identifier)
	if node == nil || ctx.TypeChecker == nil {
		return effectHookNone
	}

	// The TYPE's symbol, not `GetSymbolAtLocation`'s. The distinction is the whole point of this
	// function and it is easy to get backwards: for `import {useEffect as useFx}`,
	// `GetSymbolAtLocation` returns the LOCAL name `useFx`, while the type's symbol returns the
	// DECLARED name `useEffect`. React asks `binding.imported`, which is the declared one, so
	// reading the local name silently declines every renamed import.
	//
	// This rule ran once with the wrong accessor and the failure was instructive: the renamed
	// fixture failed while the direct one passed, from byte-identical lowered graphs. Nothing in
	// the intermediate representation distinguished them, which is what sent this back to the
	// checker rather than to the lowering.
	//
	// A namespace member falls out of the same read: `React.useEffect` resolves to the same
	// declaration as the named import, so this rule needs no member-expression case anywhere.
	calleeType := ctx.TypeChecker.GetTypeAtLocation(node)
	if calleeType == nil {
		return effectHookNone
	}
	symbol := checker.Type_symbol(calleeType)
	if symbol == nil {
		return effectHookNone
	}
	switch symbol.Name {
	case "useEffect", "useLayoutEffect", "useInsertionEffect":
		return effectHookEffect
	case "useEffectEvent":
		return effectHookUseEffectEvent
	}
	return effectHookNone
}

// The two helpers below are prefixed with this rule's name, and that is a deliberate scar rather
// than a style choice.
//
// `set_state_in_render.go`, this package's sibling out of the same upstream family, was written in
// the same hour and independently reached for `nestedFunction` and `identifierNodeOf` with the same
// meanings. The package stopped compiling with a redeclaration error naming both files, and then
// both authors removed their copy at once on the assumption the other's would stand, which turned
// one collision into an undefined-symbol error pointing at nobody.
//
// Prefixing makes the next simultaneous port a non-event. Two agents cannot see each other's
// uncommitted files, so an unprefixed helper name in a shared package is a collision waiting on
// timing rather than on judgment.
//
// **Both of these should be lifted to `internal/utilities/hir` later, and neither is lifted now.** They
// are genuinely general: "resolve the nested function this expression creates" and "get the syntax
// node behind a value" are questions about the representation, not about either rule, and the house
// rule is that two helpers doing almost the same thing is worse than one slightly the wrong shape.
// Lifting was declined here only because doing it unilaterally while a sibling holds the other copy
// mid-edit is how the collision above happened in the first place. It wants one commit that moves
// both copies and gives them their own tests, after both rules have landed.

// setStateInEffectNestedFunction resolves a FunctionExpression to the function it names.
//
// The index is bounds-checked rather than trusted: `Functions` is populated by lowering, and a
// malformed or partially lowered function is exactly the input a linter is handed.
func setStateInEffectNestedFunction(function *high_level_intermediate_representation.Function, value *high_level_intermediate_representation.FunctionExpression) *high_level_intermediate_representation.Function {
	if function == nil || value == nil || int(value.Function) >= len(function.Functions) {
		return nil
	}
	return function.Functions[value.Function]
}

// setStateInEffectIdentifierNodeOf returns the syntax a value came from, or nil for a temporary.
//
// This is the seam `Identifier.Node` documents itself as existing for: the node goes to the checker
// rather than to a local inference pass, which is why this rule needs no type inference.
func setStateInEffectIdentifierNodeOf(function *high_level_intermediate_representation.Function, id static_single_assignment.IdentifierId) *ast.Node {
	if function == nil || int(id) >= len(function.Identifiers) {
		return nil
	}
	identifier := function.Identifiers[id]
	if identifier == nil {
		return nil
	}
	return identifier.Node
}

// reportSetter points the finding at the setter call, which is where upstream points it.
//
// Upstream carries two spans, a primary on the setter reading "Avoid calling setState() directly
// within an effect" and a secondary on the effect callee reading "This is the containing effect".
// Our Diagnostic has one Range, so the primary is kept where upstream puts it and the containing
// effect is not named. That is a deliberate shape difference rather than dropped information: the
// setter is the line a reader has to change.
func reportSetter(ctx rule.Context, function *high_level_intermediate_representation.Function, setter high_level_intermediate_representation.Place) {
	// `Place.Range` rather than `Identifier.Node`, and the distinction is the whole correctness of
	// this function.
	//
	// A Place is a REFERENCE to a value at one point in the program and carries the span of that
	// reference. `Identifier.Node` is the node that INTRODUCED the value, which for a setter is its
	// binding in the `useState` destructure. Reporting through the node therefore points every
	// finding at the declaration line no matter which call produced it, so
	// `const aliased = setState` reported at `setState` where React reports at `aliased`, and the
	// transitive case reported inside the helper where React reports at the call in the effect.
	//
	// Both were invisible to every message-id fixture and both were caught only by slicing the
	// source with the finding's own range. Measured against React 7.1.1 on each input before being
	// changed, because the first guess about where the transitive case should point was also wrong.
	//
	// The node is kept as a fallback for a Place with no range, which is what a synthesized value
	// carries.
	// The range is trimmed of leading trivia here rather than at the harness. `ctx.ReportNode`
	// routes through `rule.TokenRange` and strips it for free, but that path is the one this
	// function cannot use, so the same job is done explicitly: a Place's range begins at the raw
	// start of the reference, which includes the newline and indentation before it, and an
	// untrimmed span reports `"\n    setState"` where upstream underlines `"setState"`.
	if span := trimmedRange(ctx, setter.Range); span.Pos() < span.End() {
		ctx.ReportRange(span, messageSetStateInEffect)
		return
	}
	if node := setStateInEffectIdentifierNodeOf(function, setter.Identifier); node != nil {
		ctx.ReportNode(node, messageSetStateInEffect)
	}
}

// trimmedRange advances a range past its leading trivia, comments included, matching what
// `rule.TokenRange` does for a node.
//
// It skipped whitespace alone, and a Place's range begins at the raw start of the reference, so a
// setter call with a comment above it reported from the comment. On TanStack Query's own utils.tsx
// that comment was `// eslint-disable-next-line react-hooks/set-state-in-effect`: the finding moved
// onto the comment's line, the suppression naming the next line no longer covered it, and cohere
// reported what the author had suppressed and React does not (#zx5xvtg item 8).
func trimmedRange(ctx rule.Context, span core.TextRange) core.TextRange {
	if ctx.SourceFile == nil {
		return span
	}
	sourceText := ctx.SourceFile.Text()
	start, end := span.Pos(), span.End()
	if start < 0 || end > len(sourceText) || start >= end {
		return span
	}
	if trimmed := scanner.SkipTrivia(sourceText, start); trimmed < end {
		return core.NewTextRange(trimmed, end)
	}
	return span
}
