package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/ecmascript/high_level_intermediate_representation"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/static_single_assignment"
)

var setStateInRenderMessage = rule.Message{
	Id: "setStateInRender",
	Description: "This state setter is called while the component is rendering, on every render " +
		"rather than in response to an event. Setting state schedules another render, and that " +
		"render calls the setter again, so this loops until React gives up and throws " +
		"`Too many re-renders`. To reset state when a prop changes, keep the previous prop in " +
		"state and compare it during render. To derive a value from state or props, compute it " +
		"during render instead of storing it.",
}

var setStateInUseMemoMessage = rule.Message{
	Id: "setStateInUseMemo",
	Description: "This state setter is called inside a `useMemo` callback. That callback runs " +
		"during render, so setting state there schedules another render whose `useMemo` runs " +
		"again, which can loop forever. A `useMemo` is for computing a value from its " +
		"dependencies and must not have side effects; move the state update into an event " +
		"handler or an effect.",
}

// SetStateInRender flags a `useState` setter called unconditionally during render.
//
//	valid:   function Component() { const [x, setX] = useState(0); return <button onClick={() => setX(1)}>{x}</button>; }
//	valid:   function Component(props) { const [x, setX] = useState(0); if (props.c) { setX(1); } return x; }
//	valid:   function Component(props) { const [x, setX] = useState(0); if (props.c) { return null; } setX(1); return x; }
//	valid:   function Component() { const [x, setX] = useState(0); try { setX(1); } catch (e) {} return x; }
//	invalid: function Component() { const [x, setX] = useState(0); setX(1); return x; }
//	invalid: function Component() { const [x, setX] = useState(0); const a = setX; a(1); return x; }
//	invalid: function Component(props) { const [x, setX] = useState(0); for (const _ of props) {} setX(1); return x; }
//
// Rendering is supposed to be a pure function of props and state. A setter called while that
// function runs schedules a second render, whose body calls the setter again, and React aborts the
// component with `Too many re-renders` after about fifty rounds. When the call is conditional the
// loop usually terminates, and upstream deliberately says nothing about that case: it costs a
// double render rather than hanging, and flagging it would report most legitimate
// derived-state-from-props code.
//
// # What this rule actually discriminates on, which is not what its name says
//
// The name reads as "a setter, during render", and both halves are narrower than the truth.
//
// It is not only render. A **hook** reports too, measured on React's own rule:
// `function useThing() { const [x, setX] = useState(0); setX(1); return x; }` reports. The subject
// is any function React would compile, which is components and hooks alike.
//
// And it is not the call's syntax. What decides it is the **type** of the value being called, and
// then whether the block holding the call runs on every path. Three inputs that look identical at
// the call site fall three different ways, all measured:
//
//	setX(1);                    reports    setX is a Dispatch
//	const a = setX; a(1);       reports    the alias carries the type
//	s[1](1);                    clean      reached through a property, never tracked
//
// # How a setter is identified here, and why no React knowledge is compiled in
//
// `@types/react` declares `useState<S>(): [S, Dispatch<SetStateAction<S>>]`. `Dispatch` is a **type
// alias** and the checker records it on the type, so the setter element of the tuple carries the
// alias `Dispatch` and the value element carries none. That single fact replaces upstream's entire
// mechanism, which is a seeded table naming `useState` plus 1,450 lines of unification to propagate
// a `BuiltInSetState` shape id out of it.
//
// The **alias** is the field, and the symbol is not. Probed on our own checker: a setter's type
// symbol is an anonymous type-literal whose `Name` is a non-printable internal marker, so a
// predicate keyed on the symbol name compares against binary noise. `useRef` is the mirror image,
// and the reason this is stated rather than left implicit: `RefObject` is an **interface**, so it
// lands in the type's `symbol` and its alias is correctly nil. A sibling rule reading refs must use
// the opposite field, and a predicate written for one of them silently reports nothing for the other.
//
// Three consequences, all probed rather than assumed:
//
//   - **A custom hook wrapping `useState` is followed for free.** `useCustomState` returning
//     `useState(init)` gives its caller a `Dispatch`, because the checker infers the return type
//     through the wrapper. Upstream needs `enableTreatSetIdentifiersAsStateSetters` and a name
//     heuristic to approximate this.
//   - **`useReducer` does not false-positive.** `dispatch` carries `ActionDispatch`, a different
//     alias.
//   - **A structurally identical function does not false-positive.** `const plain: (value: number)
//     => void` has a nil alias while `const declared: Dispatch<number>` has the alias, so the test
//     is nominal rather than structural. That is the negative control for this predicate, and it is
//     the reason the predicate is safe to apply to every call in the tree.
//
// The trade is stated plainly because it decides how this rule should be judged. This rule depends
// on `@types/react` resolving and on the setter having a real type. That is right for application
// code and wrong for a corpus: **nine of upstream's ten fixtures for this rule do not import
// React**, so `useState` is a bare undeclared global that resolves to nothing and the whole file is
// `any`. A pass-rate measured on those fixtures as written would be near zero while the rule is
// correct on everything Kirk actually writes. See the test file for how that is handled.
//
// # Why this needs the intermediate representation and the post-dominator tree
//
// Two of the three questions this rule asks cannot be answered from the syntax tree.
//
// **Does this call reach a setter** needs alias chains followed to any depth. `const a = setX;
// const b = a; b(1);` reports at `b`, and single-assignment form gives that for free through
// `LoadLocal` and `StoreLocal` propagation, which is what upstream does over the same shape.
//
// **Does this call happen on every path** is post-dominance and nothing else. The three loop
// fixtures are syntactically indistinguishable and fall differently:
//
//	for (const _ of props) {}                                    setState after: REPORTS
//	for (const _ of props) { if (c) break; else continue; }       setState after: REPORTS
//	for (const _ of props) { if (c) break; else throw new Error(); } setState after: REPORTS
//	for (const p of props) { setState(true); }                    setState inside: CLEAN
//
// and an early return before the call makes it clean while leaving the syntax around the call
// unchanged. `high_level_intermediate_representation.UnconditionalBlocks` is the port of upstream's `computeUnconditionalBlocks`; its
// file explains why forward dominance, which this tree already had in two places, answers the wrong
// question and would have reported the early-return case.
//
// This is therefore the opposite conclusion to `error-boundaries`, which read the same graph
// machinery and used none of it. That rule's question was "is B written inside A", which is about
// text. This one's is "must control reach B", which is about paths, and no parent walk answers it.
//
// # Two message ids from one validator
//
// Upstream's `validate_no_set_state_in_render` emits two diagnostics, and the branch order matters:
// the `useMemo` arm is checked BEFORE the unconditional test, so a setter called inside a `useMemo`
// reports even when the call is conditional within that callback. Measured, `useMemo(() => { if
// (props.c) { setX(1); } return 1; }, [])` reports.
//
// **`useCallback` does not report and `useMemo` does**, which is not a distinction anyone would
// guess and is not a judgment about memoization at all. Upstream replaces a `useMemo` callback with
// a **direct call** to it, so the callback's body is inlined into the render graph and its setState
// becomes a real instruction; a `useCallback` callback is replaced with a **LoadLocal**, so the
// function is only handed around and its body never runs during render. Read at
// `getManualMemoizationReplacement` in the shipped plugin, and confirmed by running both.
//
// Because our lowering never emits `StartMemoize` (that marker is inserted by a compiler pass this
// tree does not run, as `lower.go` states), the memoized region is recognized syntactically instead,
// the same way `use_memo.go` and `void_use_memo.go` already recognize one. Fidelity is owed to what
// the rule decides, not to how upstream obtains it.
//
// # Reproduced holes, each measured on React's own rule rather than inferred
//
//   - **An optional call is silent.** `setX?.(1)` reports nothing upstream. Reproduced by declining
//     `CallExpression.Optional`, rather than left as an accident of our lowering.
//   - **A setter reached through a property is silent.** `const s = useState(0); s[1](1);` and
//     `const o = {go: setX}; o.go(1);` both report nothing, because upstream tracks only
//     `LoadLocal` and `StoreLocal`. Our `MethodCall` and `ComputedLoad` are deliberately not
//     consulted.
//   - **A rest parameter silences its function.** `function Component(...props)` is not a component
//     to React's parameter test, so nothing in it reports. Checked with a control in the same file:
//     a second component beside it still reports, so this is a per-function decline rather than a
//     whole-file bailout.
var SetStateInRender = rule.Rule{
	Name:             "react-hooks/set-state-in-render",
	NeedsTypeChecker: true,
	// Reads other files only through shape readers (rule.ExportNameIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,
	// The setter is identified by the type's alias, so a `useState` resolving to `any` carries
	// no alias and this rule declines the whole file in silence. `structure/react-hook-no-any-type`
	// reads this flag to name that cost at the site where the types went bad.
	ResolvesReactValueTypes: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			// The whole file at once, for the reason `static_components.go` gives: lowering already
			// descends into nested functions through the function arena, so listening per function
			// kind would lower every inner function twice, the second time without the bindings its
			// parent supplies.
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
					// Shared with the other rules that lower this same function; see high_level_intermediate_representation.ForFunction.
					// Construct runs inside the cached computation, because it mutates in place and is
					// not idempotent.
					lowered := high_level_intermediate_representation.ForFunction(ctx, functionNode)
					if lowered == nil {
						return
					}
					analyzeSetStateSubject(ctx, lowered)
				})
			},
		}
	},
}

// analyzeSetStateSubject runs the validator over the functions React would have compiled.
//
// The gate is the react shelf's `IsComponentOrHookLike`, taken whole rather than written
// again. Two porters have paid for that gate already, seventeen probe rounds and seven false
// positives, and a second copy of it in this package would be a second thing to keep in agreement
// with React for no benefit. It is the right gate here rather than a convenient one:
// `static_components.go` asks only for JSX because its finding requires a JSX tag, but this rule
// fires inside a hook with no JSX at all, so it needs the hook half that rule argues is unreachable
// for itself.
//
// A function the gate declines is still descended into, because a component nested inside a plain
// wrapper is a unit even when the wrapper is not.
func analyzeSetStateSubject(ctx rule.Context, function *high_level_intermediate_representation.Function) {
	if function == nil {
		return
	}
	if function.Node == nil || !utilsreact.IsComponentOrHookLike(function.Node) {
		for _, nested := range function.Functions {
			analyzeSetStateSubject(ctx, nested)
		}
		return
	}
	// A unit found nested in a function that is not one is analyzed as React Compiler compiles it,
	// on its own; see AsCompilationUnit.
	function = high_level_intermediate_representation.AsCompilationUnit(ctx, function,
		high_level_intermediate_representation.ForFunction)
	reportSetStateInRender(ctx, function, function.Node, true, map[static_single_assignment.IdentifierId]bool{})
}

// reportSetStateInRender is upstream's `validate_impl` over one lowered function.
//
// `unconditionalSetStateFunctions` is upstream's set of the same name, threaded through the
// recursion rather than rebuilt, so a function proven to set state unconditionally makes its own
// call sites report. It is what turns
//
//	const foo = () => { setX(1); }; foo();
//
// into a finding at `foo` rather than at `setX`, and chains through any depth of wrapper: the
// nested-function fixture reports at `baz`, three levels out.
//
// The return value says whether this function itself reported, which is the signal the caller uses
// to decide whether the value holding it joins the set.
//
// `emit` separates reporting from detecting, and the separation is upstream's rather than an
// optimization. `validate_impl` calls itself, binds the result to `inner_errors`, and then uses
// only `!inner_errors.is_empty()`: the inner findings are DISCARDED and the finding that survives
// is the one at the outer call site. Measured on React, `const foo = () => { setX(1); }; foo();`
// produces exactly one finding, pointing at `foo` and never at `setX`.
//
// Transcribing the recursion without this reports twice, which is what the lambda fixture caught.
// The reason upstream does it is worth stating: the inner function is not itself render, so the
// call inside it is not a call during render; what makes it a defect is that the OUTER call invokes
// it unconditionally while rendering, and that is where a reader has to make the edit.
//
// `compiledUnit` stays pinned to the OUTERMOST component or hook across the recursion rather than
// following `function`. It is only used as the outer boundary for shadow resolution on a `useMemo`
// callee, and that boundary is upstream's compiled unit: a `useMemo` binding at module scope does
// not shadow, while one inside the component does. Passing the nested function instead would stop
// the scope walk early and make a binding in the component invisible to a callback inside it.
func reportSetStateInRender(
	ctx rule.Context,
	function *high_level_intermediate_representation.Function,
	compiledUnit *ast.Node,
	emit bool,
	unconditionalSetStateFunctions map[static_single_assignment.IdentifierId]bool,
) bool {
	unconditional := high_level_intermediate_representation.UnconditionalBlocks(function)
	reported := false

	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}

			switch value := instruction.Value.(type) {
			// Reading and writing carry the "this value calls a setter unconditionally" property
			// along, which is what makes an alias chain of any length behave like its head. The
			// setter's own type is followed by the checker instead, so these two arms exist only
			// for values that are not setters themselves but call one.
			case *high_level_intermediate_representation.LoadLocal:
				if unconditionalSetStateFunctions[value.Place.Identifier] {
					unconditionalSetStateFunctions[instruction.LValue.Identifier] = true
				}
			// Reading a CAPTURED binding carries the property too. Upstream has no counterpart to
			// this arm, because its captures are ordinary values in one shared arena and a read of
			// one is a `LoadLocal`. Here a capture is its own instruction, so a nested function
			// calling something it closed over reaches the setter through `LoadContext` and nothing
			// else. Without this arm the three-level fixture is silent: inside `bar`, the call is
			// `Call $4()` where `$4 = LoadContext foo$1`.
			case *high_level_intermediate_representation.LoadContext:
				if unconditionalSetStateFunctions[value.Place.Identifier] {
					unconditionalSetStateFunctions[instruction.LValue.Identifier] = true
				}
			case *high_level_intermediate_representation.StoreLocal:
				if unconditionalSetStateFunctions[value.Value.Identifier] {
					unconditionalSetStateFunctions[value.LValue.Identifier] = true
					unconditionalSetStateFunctions[instruction.LValue.Identifier] = true
				}

			// A nested function is analysed only when it closes over something that could be a
			// setter, which is upstream's own `has_set_state_operand` guard. Without it every
			// callback in every component would be walked, and with it the walk is bounded by what
			// the function can actually reach.
			case *high_level_intermediate_representation.FunctionExpression:
				nested := nestedFunction(function, value)
				if nested == nil {
					continue
				}
				// The nested function names its captures with its OWN identifier ids, so the set
				// built in this function's namespace means nothing inside it. Translate across the
				// boundary rather than passing the map down; see setterCapturesOf.
				nestedSetters, capturesOne := setterCapturesOf(ctx, function, value, nested, unconditionalSetStateFunctions)

				// A `useMemo` callback is analysed whatever it captures, and this is not an
				// optimization waiver. Upstream inlines that callback into the render graph before
				// validating, so a setter DECLARED INSIDE it is an ordinary value of the enclosing
				// function and needs no capture to be seen. Measured on React:
				//
				//	useMemo(() => { const [a, setA] = useState(1); setA(2); return a; }, [])
				//
				// reports, and it captures nothing at all. `useCallback` with the identical body is
				// clean, because its callback is never invoked during render.
				//
				// This was found by a mutation, not by a fixture. Removing the capture guard made
				// the rule report that input, and the input turned out to be one React reports and
				// this rule did not: the mutant was right and the original was wrong. Every fixture
				// stayed green either way, because every one of them reaches its setter through a
				// capture.
				emitFromNested := isUseMemoCallbackOf(nested.Node, compiledUnit)
				if !capturesOne && !emitFromNested {
					continue
				}
				// A `useMemo` callback is the one nested function whose own findings survive, and
				// it is not an exception to upstream's rule so much as a consequence of upstream
				// never recursing into it at all: the callback is REPLACED by a direct call to
				// itself before validation, so its body is inlined into the render graph and its
				// setter call is an ordinary instruction of the enclosing function. There is no
				// inner walk to discard. Reproduced here by emitting from the recursion for this
				// one shape, which reaches the same verdict without an inlining pass.
				//
				// A `useCallback` callback is replaced by a LoadLocal instead, so it is never
				// invoked during render and reports nothing. That asymmetry is measured, not
				// guessed; see the rule comment.
				if reportSetStateInRender(ctx, nested, compiledUnit, emitFromNested, nestedSetters) {
					unconditionalSetStateFunctions[instruction.LValue.Identifier] = true
				}

			case *high_level_intermediate_representation.CallExpression:
				// An optional call is silent upstream. See the rule comment.
				if value.Optional {
					continue
				}
				if !isStateSetter(ctx, function, value.Callee.Identifier) &&
					!unconditionalSetStateFunctions[value.Callee.Identifier] {
					continue
				}
				// The memoized test comes first, matching upstream's branch order, so a call inside
				// a `useMemo` reports even where it is conditional within that callback.
				if isUseMemoCallbackOf(function.Node, compiledUnit) {
					if emit {
						reportSetterCall(ctx, function, value.Callee.Identifier, setStateInUseMemoMessage)
					}
					reported = true
					continue
				}
				if !unconditional[block.Id] {
					continue
				}
				if emit {
					reportSetterCall(ctx, function, value.Callee.Identifier, setStateInRenderMessage)
				}
				reported = true
			}
		}
	}

	return reported
}

// nestedFunction resolves a FunctionExpression to the function it names.
//
// The index is bounds-checked rather than trusted: `Functions` is populated by lowering and a
// malformed or partially lowered function is exactly the input a linter is handed.
func nestedFunction(function *high_level_intermediate_representation.Function, value *high_level_intermediate_representation.FunctionExpression) *high_level_intermediate_representation.Function {
	if value == nil || int(value.Function) >= len(function.Functions) {
		return nil
	}
	return function.Functions[value.Function]
}

// setterCapturesOf translates the known-setter set across a function boundary.
//
// It returns the set expressed in the NESTED function's identifier namespace, plus whether that set
// is non-empty, which is upstream's `has_set_state_operand` guard.
//
// # Two identifier namespaces, and why passing the map down is wrong
//
// Upstream shares one arena across every function in a compilation, so an `IdentifierId` means the
// same value everywhere and `unconditional_set_state_functions` can simply be threaded through the
// recursion. This IR gives each `Function` its OWN tables, which the package comment names as its
// main deliberate divergence, so the same integer denotes different values in a parent and a child.
// Passing the parent's map into the child asks about ids that happen to collide, which is not a
// milder version of the right question but an unrelated one.
//
// The correspondence that makes the translation exact is positional: `FunctionExpression.Captures`
// and `nested.Context` are the two ends of one edge list, in the same order. Probed on a function
// capturing two values, the parent side reads `[setX, props]` and the nested side reads
// `[setX, props]`, index for index.
//
// # Why the parent side is where the type question is asked
//
// A nested `Context` place has a nil `Identifier.Node`, because a capture is a value the callee
// receives rather than a syntactic occurrence inside it. There is nothing to hand the checker, so a
// setter read from that side answers false. The same capture read from `Captures`, which indexes the
// enclosing function's table, resolves to alias `Dispatch`. Both fixtures with a nested function
// went silent on the first version of this, and the symptom looked like an unfixed capture gap in
// lowering rather than like the wrong table being read.
//
// # The chain, which is what the three-level fixture tests
//
// `const foo = () => setX(1); const bar = () => foo(); const baz = () => bar(); baz();` reports once,
// at `baz`. Each level is admitted because it captures something the level below proved: `foo`
// captures a real setter, so analysing it marks `foo`'s value; `bar` captures `foo`, which is now in
// the set, so `bar` is analysed and marked; and so on. That is why the parent-side lookup consults
// `unconditionalSetStateFunctions` as well as the type, and why the walk must reach the definitions
// in evaluation order, which `Function.Blocks` being in reverse postorder already guarantees.
func setterCapturesOf(
	ctx rule.Context,
	enclosing *high_level_intermediate_representation.Function,
	value *high_level_intermediate_representation.FunctionExpression,
	nested *high_level_intermediate_representation.Function,
	unconditionalSetStateFunctions map[static_single_assignment.IdentifierId]bool,
) (map[static_single_assignment.IdentifierId]bool, bool) {
	translated := map[static_single_assignment.IdentifierId]bool{}
	for index, captured := range value.Captures {
		isSetter := unconditionalSetStateFunctions[captured.Identifier] ||
			isStateSetter(ctx, enclosing, captured.Identifier)
		if !isSetter {
			continue
		}
		// The nested side may be shorter than the parent side if lowering ever stops emitting them
		// in lockstep, so the index is checked rather than trusted.
		if index >= len(nested.Context) {
			continue
		}
		translated[nested.Context[index].Identifier] = true
	}
	return translated, len(translated) > 0
}

// isStateSetter reports whether a value's type is a React state setter.
//
// The predicate is the type's ALIAS, for the reasons in the rule comment: `Dispatch` is a type
// alias so it lands there, while a ref's `RefObject` is an interface and lands in the symbol
// instead. Reading the wrong field reports a genuine `useRef` as neither.
//
// `ActionDispatch`, which `useReducer` produces, is deliberately not accepted. Upstream gives it a
// different built-in shape and its own diagnostics, and a dispatch called during render is not the
// same defect.
func isStateSetter(ctx rule.Context, function *high_level_intermediate_representation.Function, id static_single_assignment.IdentifierId) bool {
	if ctx.TypeChecker == nil {
		return false
	}
	node := identifierNodeOf(function, id)
	if node == nil {
		return false
	}
	valueType := ctx.TypeChecker.GetTypeAtLocation(node)
	if valueType == nil {
		return false
	}
	alias := shimchecker.Type_alias(valueType)
	if alias == nil {
		return false
	}
	symbol := alias.Symbol()
	return symbol != nil && symbol.Name == "Dispatch"
}

// identifierNodeOf returns the syntax a value came from, or nil for a pure temporary.
//
// This is the seam `Identifier.Node` documents itself as existing for: the node goes to the checker
// rather than to a local inference pass.
func identifierNodeOf(function *high_level_intermediate_representation.Function, id static_single_assignment.IdentifierId) *ast.Node {
	if function == nil || int(id) >= len(function.Identifiers) {
		return nil
	}
	identifier := function.Identifiers[id]
	if identifier == nil {
		return nil
	}
	return identifier.Node
}

// reportSetterCall points the finding at the callee identifier.
//
// Upstream's span is `callee.span`, which is the name being called and not the whole call: the
// golden for `setX(1)` underlines four characters and the one for `aliased(2)` underlines seven.
// Reporting through `ctx.ReportNode` routes via `rule.TokenRange`, which trims leading trivia at
// the harness.
func reportSetterCall(ctx rule.Context, function *high_level_intermediate_representation.Function, id static_single_assignment.IdentifierId, message rule.Message) {
	if node := identifierNodeOf(function, id); node != nil {
		ctx.ReportNode(node, message)
		return
	}
	if int(id) < len(function.Identifiers) {
		if identifier := function.Identifiers[id]; identifier != nil && identifier.Node != nil {
			ctx.ReportNode(identifier.Node, message)
		}
	}
}

// isUseMemoCallbackOf reports whether a function node is the callback argument of a `useMemo` call.
//
// Recognized syntactically, because our lowering never emits `StartMemoize`: that marker is
// inserted by a compiler pass this tree does not run, which `lower.go` states as a design decision
// rather than a gap. `use_memo.go` and `void_use_memo.go` already identify a `useMemo` this way, so
// this is the established shape here rather than a new one.
//
// `useCallback` is deliberately excluded and it is the one line in this rule most likely to be
// "corrected" by a later reader. Its callback is not invoked during render, so its body never runs
// there; see the rule comment for upstream's mechanism.
//
// The callee test is `void_use_memo.go`'s `isUseMemoCallee` rather than a second copy. That helper
// already resolves shadowing, skips parentheses, requires the receiver of a member form to be a
// global `React`, and declines an optional member access while allowing an optional call, each of
// those measured against React. A local version written here handled none of it and would have
// accepted `Other.useMemo(...)`, which upstream declines.
func isUseMemoCallbackOf(node *ast.Node, enclosing *ast.Node) bool {
	if node == nil {
		return false
	}
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	call := parent.AsCallExpression()
	if call == nil || len(call.Arguments.Nodes) == 0 || call.Arguments.Nodes[0] != node {
		return false
	}
	return isUseMemoCallee(call.Expression, enclosing)
}
