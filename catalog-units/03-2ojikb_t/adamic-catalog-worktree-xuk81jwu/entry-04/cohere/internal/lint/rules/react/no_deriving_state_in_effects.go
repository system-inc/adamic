package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/ecmascript/high_level_intermediate_representation"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/static_single_assignment"
)

var messageNoDerivingStateInEffects = rule.Message{
	Id: "noDerivingStateInEffects",
	Description: "This effect exists only to copy a value into state that was already computable " +
		"from the values it depends on. React renders once with the stale state, runs the effect, " +
		"then renders a second time with the new state, so the user can briefly see the old value " +
		"and every dependent effect runs twice. Compute the value during render instead, as a " +
		"plain constant beside the values it derives from. If the derivation is expensive, wrap " +
		"that constant in `useMemo` rather than storing it in state.",
}

// NoDerivingStateInEffects flags an effect whose only work is setting state from its own dependencies.
//
//	valid:   function C({a}) { const [v, setV] = useState(0); useEffect(() => { setV(a); log(); }, [a]); return v; }
//	valid:   function C({a, b}) { const [v, setV] = useState(0); useEffect(() => { setV(a); }, [a, b]); return v; }
//	invalid: function C({a, b}) { const [v, setV] = useState(''); useEffect(() => { setV(a + b); }, [a, b]); return v; }
//
// # The authority, and which of two passes it is
//
// React's own `validateNoDerivedComputationsInEffects`, raising `ErrorCategory.EffectDerivationsOfState`,
// at `Validation/ValidateNoDerivedComputationsInEffects.ts`. That file has a sibling named
// `ValidateNoDerivedComputationsInEffects_exp.ts`, 842 lines against the base pass's 229, and
// choosing between them is the first decision this port makes.
//
// The shipped ESLint plugin runs the BASE pass. Measured in
// `node_modules/eslint-plugin-react-hooks/cjs/eslint-plugin-react-hooks.development.js`: its
// `COMPILER_OPTIONS.environment` sets `validateNoDerivedComputationsInEffects: true` and never sets
// the `_exp` flag, whose schema default is false. The pipeline prefers `_exp` when both are on, so
// the flag that is absent is the one that decides. This rule is the base pass.
//
// That matters more than it sounds, because the CORPUS is the other way round. All 21 fixtures in
// `__tests__/fixtures/compiler/effect-derived-computations/` carry
// `@validateNoDerivedComputationsInEffects_exp`, so their names describe the experimental pass's
// verdict and not the shipped one. Exactly one fixture in the whole 2,688-fixture tree carries the
// base pragma: `error.invalid-derived-computation-in-effect.js`. Every one of the 21 was run
// against the shipped rule before being written into the test file, and five of them invert:
// three named `-no-error` report, and the one named `invalid-derived-computation-in-effect` is
// clean. Trusting the names would have produced a rule whose fixtures and code shared one wrong
// belief, which is the failure the brief warns about; the fixture file records each measured
// verdict beside its case.
//
// # What the pass decides
//
// Upstream gathers three maps over the enclosing function, then judges each effect:
//
//	LoadLocal          value -> the value it read, so a dep written as a local resolves to its source
//	ArrayExpression    value -> the literal, so a dependency array can be recovered from an identifier
//	FunctionExpression value -> the function, so an effect callback can be recovered from an identifier
//
// A call is a candidate when its callee is an effect hook AND it has exactly two arguments AND both
// are plain identifiers. Both halves are load-bearing: `useEffect(cb)` with no dependency array is
// never a candidate, because this rule's whole judgment is "the callback computes nothing the deps
// did not already supply", which has no meaning without deps.
//
// Then `validateEffect` requires, in order:
//
//   - Every value the callback closes over is either a state setter or one of the deps. Capturing
//     anything else means the effect reads something outside the derivation and is left alone.
//   - Every dep is actually captured. An unused dep means the effect is not purely a function of them.
//   - No back edge. A block whose predecessor has not been seen yet is a loop, and the pass returns
//     rather than analysing it.
//   - Every instruction is one of a small allowed set. Anything else, including a store, an object
//     literal, or a JSX element, returns.
//   - No terminal reads a derived value. A derived value used in a branch condition or a return
//     means the effect does more than set state.
//
// Within that, a taint map carries dep-ness forward through loads, property reads, binary
// expressions, template literals and calls. A setter call reports only when its argument's taint
// set has exactly as many distinct deps as the effect declared, so setting from ONE of two deps is
// silent and setting from both reports. A setter call whose argument carries no taint at all
// abandons the whole effect rather than skipping that one call.
//
// # Two predicates that upstream answers from a shape registry and this rule answers from types
//
// `isUseEffectHookType` and `isSetStateType` both read `Identifier.type.shapeId`, a builtin shape
// the compiler seeds when it sees an import from the literal module `react`. This tree has no such
// registry, and the sibling rule `set-state-in-effect` already established the replacement: ask the
// checker. `effectHookKind` resolves a callee to the name it was DECLARED as, which is what makes a
// renamed import report and a same-named local stay silent, and `isSetterTyped` reads the `Dispatch`
// type ALIAS that `@types/react` puts on a `useState` setter and on nothing else. Both helpers are
// this package's and are reused rather than rewritten, which is the whole reason these two rules
// were batched.
//
// The measured consequence of the difference is stated rather than smoothed over: upstream is
// silent on `import {useEffect} from './react'` because the specifier is not `react`, and this rule
// reports. That was measured in both directions rather than reasoned about, by running the shipped
// rule over the corpus twice, once with the original specifier and once rewritten. Four of the five
// reporting fixtures went silent under the rewrite. It is the same trade `set-state-in-effect`
// documents: a dependency on `@types/react` instead of a shape registry, better on application code
// and worse against a corpus that omits its imports.
//
// # Where the finding points
//
// At the setter's own reference inside the effect, which is `instr.value.callee.loc` upstream. Not
// at the effect, not at the dependency array, and not at the setter's binding in the `useState`
// destructure. Measured against React 7.1.1 on every reporting fixture and asserted by slicing the
// source with the finding's range, because a message id cannot see where a finding points and the
// three plausible anchors here all produce the same id.
//
// # What this rule deliberately does not do
//
// `enableTreatFunctionDepsAsConditional` and the `_exp` pass's data-flow tree are not ported. The
// `_exp` pass produces a much richer description naming the reactive sources and drawing a tree,
// and it is a different rule with a different verdict on a third of the corpus. When the shipped
// plugin switches to it this rule is a re-port rather than an adjustment, and the fixture file's
// recorded per-case verdicts are what will make that visible.
var NoDerivingStateInEffects = rule.Rule{
	Name:             "react-hooks/no-deriving-state-in-effects",
	NeedsTypeChecker: true,
	// Reads other files only through shape readers (rule.ExportNameIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,
	// Same predicate as `set-state-in-effect`: the setter is recognised by its type alias, so a hook
	// returning `any` takes every finding in the file with it. Declared so the tripwire can name it.
	ResolvesReactValueTypes: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			// The whole file at once, matching `set-state-in-effect` and `static-components`.
			// Lowering already descends into nested functions through the function arena, so
			// listening per function kind would lower every inner function twice.
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
					// Manual memoization is erased first, because upstream validates a graph where
					// `useMemo` and `useCallback` are already gone.
					lowered := high_level_intermediate_representation.ForFunctionWithoutManualMemoization(ctx, functionNode)
					if lowered == nil {
						return
					}
					analyzeDerivedComputationsSubject(ctx, lowered)
				})
			},
		}
	},
}

// analyzeDerivedComputationsSubject runs the validator over the functions React would have compiled.
//
// The same gate and the same descent as `analyzeSetStateInEffectSubject`, for the same reason: the
// effect is classified by its type, so only the component-or-hook test keeps this to the functions
// upstream validates, and it is what makes `hir.MayHoldComponentOrHook` exact for this rule.
func analyzeDerivedComputationsSubject(ctx rule.Context, function *high_level_intermediate_representation.Function) {
	if function == nil {
		return
	}
	if function.Node == nil || !utilsreact.IsComponentOrHookLike(function.Node) {
		for _, nested := range function.Functions {
			analyzeDerivedComputationsSubject(ctx, nested)
		}
		return
	}
	// A unit found nested in a function that is not one is analyzed as React Compiler compiles it,
	// on its own; see AsCompilationUnit.
	function = high_level_intermediate_representation.AsCompilationUnit(ctx, function,
		high_level_intermediate_representation.ForFunctionWithoutManualMemoization)
	reportDerivedComputationsInEffects(ctx, function)
}

// reportDerivedComputationsInEffects is upstream's `validateNoDerivedComputationsInEffects` over one
// compiled function: the gather loop, then a judgment per candidate effect.
func reportDerivedComputationsInEffects(ctx rule.Context, function *high_level_intermediate_representation.Function) {
	if function == nil {
		return
	}

	// Upstream's three maps, keyed the same way. Single-assignment form is what lets each be one
	// entry per value rather than per binding.
	candidateDependencies := map[static_single_assignment.IdentifierId]*high_level_intermediate_representation.ArrayExpression{}
	functions := map[static_single_assignment.IdentifierId]*high_level_intermediate_representation.FunctionExpression{}
	locals := map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId{}

	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			target := instruction.LValue.Identifier

			switch value := instruction.Value.(type) {
			case *high_level_intermediate_representation.LoadLocal:
				locals[target] = value.Place.Identifier
			case *high_level_intermediate_representation.ArrayExpression:
				candidateDependencies[target] = value
			case *high_level_intermediate_representation.FunctionExpression:
				functions[target] = value
			case *high_level_intermediate_representation.CallExpression:
				derivedEffectCandidate(ctx, function, value.Callee, value.Args, candidateDependencies, functions, locals)
			case *high_level_intermediate_representation.MethodCall:
				// `React.useEffect(...)`. The property Place is the callee, exactly as upstream
				// selects `instr.value.property` for a MethodCall.
				derivedEffectCandidate(ctx, function, value.Property, value.Args, candidateDependencies, functions, locals)
			}
		}
	}

	// A nested function is lowered into this function's arena but its own body is a separate graph
	// the walk above never entered. Same descent as `set-state-in-effect`.
	for _, nested := range function.Functions {
		reportDerivedComputationsInEffects(ctx, nested)
	}
}

// derivedEffectCandidate is the CallExpression and MethodCall arm of upstream's gather switch.
//
// The three conditions are upstream's and each is separately load-bearing. Two arguments, because a
// dependency array is what this rule reasons about. Both plain identifiers, because upstream reads
// them out of the maps above rather than off the call, so an inline function or an inline array
// literal is not a candidate at all. Verified against the shipped rule: `useEffect(() => setV(a),
// [a])` written inline is REPORTED, because lowering assigns the inline callback and the inline
// array to temporaries first, so both arguments are identifiers by the time the pass sees them.
// That is why this is not a syntactic test.
func derivedEffectCandidate(
	ctx rule.Context,
	function *high_level_intermediate_representation.Function,
	callee high_level_intermediate_representation.Place,
	args []high_level_intermediate_representation.Argument,
	candidateDependencies map[static_single_assignment.IdentifierId]*high_level_intermediate_representation.ArrayExpression,
	functions map[static_single_assignment.IdentifierId]*high_level_intermediate_representation.FunctionExpression,
	locals map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId,
) {
	if len(args) != 2 || args[0].Spread || args[1].Spread {
		return
	}
	if !isUseEffectExactly(ctx, function, callee) {
		return
	}

	effectFunction, hasFunction := functions[args[0].Place.Identifier]
	if !hasFunction {
		return
	}
	deps, hasDeps := candidateDependencies[args[1].Place.Identifier]
	// The `len == 0` half is upstream's `deps.elements.length !== 0` and it is SUBSUMED here,
	// deliberately kept rather than deleted. Its mutant survives the whole suite and no fixture can
	// catch it, so the equivalence is stated with the measurement rather than left as a blind spot.
	//
	// Why nothing can distinguish the two versions: with an empty dependency array `effectDeps` is
	// empty, so the taint map `values` is seeded with nothing, and `values` only ever gains entries
	// derived from existing ones. The single path to a finding requires `hasDeps` to be true for a
	// setter call's argument, which reads that map. It can never be true, so an empty-deps effect
	// reports nothing whether the guard is present or not, and the guard only decides how early it
	// stops.
	//
	// Measured rather than argued: nine empty-deps shapes were run against both versions, including
	// a setter reading state, a prop, a second state value, two setter calls, an empty body, and a
	// derived expression. Zero disagreements. React 7.1.1 is silent on all of them too.
	if !hasDeps || len(deps.Elements) == 0 {
		return
	}

	// `deps.elements.every(element => element.kind === 'Identifier')` upstream: a hole or a spread
	// in the dependency array abandons the effect rather than being skipped.
	dependencies := make([]static_single_assignment.IdentifierId, 0, len(deps.Elements))
	for _, element := range deps.Elements {
		// The `Hole` half is SUBSUMED and kept for fidelity, with the measurement recorded rather
		// than the branch deleted. A hole lowers to a real `<hole>` element, so this is reachable,
		// but dropping the test changes no verdict: the hole's Place carries no live identifier, so
		// it enters `dependencies` as a value nothing in the callback can capture, and the "every
		// dep is actually captured" check below declines the effect one step later.
		//
		// Both versions reach silence by different routes, which is why the mutant survives every
		// fixture. Measured on seven hole shapes against both versions, including a hole before,
		// between and after real deps, two holes, and a hole beside the two-dep derivation that
		// otherwise reports: zero disagreements. React 7.1.1 is silent on all of them.
		if element.Spread || element.Hole {
			return
		}
		// `locals.get(dep) ?? dep`: a dependency written as an identifier lowered to a LoadLocal, so
		// the taint has to be keyed on the value it READ rather than on the temporary that read it.
		identifier := element.Place.Identifier
		if origin, ok := locals[identifier]; ok {
			identifier = origin
		}
		dependencies = append(dependencies, identifier)
	}

	// A dependency that is a compile-time constant is one upstream never sees, and declining here
	// is how this rule reproduces a pass we do not have. See `dependencyIsFoldedConstant`.
	for _, dependency := range dependencies {
		if dependencyIsFoldedConstant(function, dependency) {
			return
		}
	}

	nested := setStateInEffectNestedFunction(function, effectFunction)
	if nested == nil {
		return
	}
	validateDerivedEffect(ctx, function, nested, effectFunction, dependencies)
}

// validateDerivedEffect is upstream's `validateEffect`: the capture check, then the taint walk.
//
// Every `return` here abandons the whole effect rather than skipping one instruction, and that is
// upstream's shape rather than an approximation of it. An effect that does one unrecognised thing
// is not a pure derivation, so partial credit would report on effects upstream leaves alone.
func validateDerivedEffect(
	ctx rule.Context,
	enclosing *high_level_intermediate_representation.Function,
	effectFunction *high_level_intermediate_representation.Function,
	effectValue *high_level_intermediate_representation.FunctionExpression,
	effectDeps []static_single_assignment.IdentifierId,
) {
	// The captures are places in the ENCLOSING function, which is what makes them comparable to the
	// dependency identifiers gathered there. `effectFunction.Context` holds the same values renamed
	// into the nested function's own identifier space, so comparing those to `effectDeps` would
	// compare two different numbering schemes and never match.
	captures := effectValue.Captures

	// Every capture is a setter or a dep. Anything else means the effect reads something the
	// dependency array did not declare, and upstream leaves it alone.
	for _, capture := range captures {
		if isSetterTyped(ctx, enclosing, capture) {
			continue
		}
		if containsIdentifier(effectDeps, capture.Identifier) {
			continue
		}
		return
	}

	// Every dep is actually captured. A declared dep the callback never reads means the effect is
	// not purely a function of its dependencies.
	for _, dep := range effectDeps {
		found := false
		for _, capture := range captures {
			if capture.Identifier == dep {
				found = true
				break
			}
		}
		if !found {
			return
		}
	}

	// The taint map is keyed in the NESTED function's identifier space, seeded from the captures.
	// Upstream seeds `values` directly from `effectDeps` because its context operands carry the same
	// identifier ids inside and out; ours are renamed by lowering, so the seed is translated through
	// the context list, which is index-aligned with the captures.
	values := map[static_single_assignment.IdentifierId][]static_single_assignment.IdentifierId{}
	for index, capture := range captures {
		if index >= len(effectFunction.Context) {
			break
		}
		if containsIdentifier(effectDeps, capture.Identifier) {
			inner := effectFunction.Context[index].Identifier
			values[inner] = []static_single_assignment.IdentifierId{capture.Identifier}
		}
	}
	// This guard is UNREACHABLE today and is kept deliberately, because what makes it unreachable
	// lives in another package.
	//
	// The seed can only come up short if `Captures` and `Context` stop corresponding index for
	// index, and `lower.go:459` states that correspondence as an invariant of the lowering:
	// "`FunctionExpression.Captures[i]` is the enclosing function's value for `nested.Context[i]`",
	// with `capturedSymbolsInOrder` built in first-encounter order precisely to hold it. Every dep
	// has already been proven captured by the loop above, so each one finds its inner identifier.
	// Probed on three shapes, two deps, three deps, and two setters over one dep: captures and
	// context were equal length every time.
	//
	// Its mutant survives the suite and that is expected rather than a fixture gap. It is kept
	// because a silent failure of that invariant would cost findings rather than add them, which is
	// the failure mode nothing in this file could detect: the count test below would compare
	// against a taint set that can never reach the required size, and the rule would just go quiet.
	// A cross-package invariant this rule cannot enforce is exactly the thing worth a cheap guard.
	if len(values) != len(effectDeps) {
		return
	}

	seenBlocks := map[static_single_assignment.BlockId]bool{}
	var setStateLocations []high_level_intermediate_representation.Place

	for _, block := range effectFunction.Blocks {
		// A predecessor not yet seen is a back edge, and upstream refuses to analyse a loop.
		// `Blocks` is in reverse postorder, so for an acyclic region every predecessor has already
		// been visited and this never fires.
		for _, predecessor := range block.Predecessors {
			if !seenBlocks[predecessor] {
				return
			}
		}

		for _, phi := range block.Phis {
			aggregate := map[static_single_assignment.IdentifierId]bool{}
			for _, predecessor := range high_level_intermediate_representation.PhiOperandsInOrder(phi) {
				operand := phi.Operands.At(predecessor)
				for _, dep := range values[operand.Identifier] {
					aggregate[dep] = true
				}
			}
			if len(aggregate) != 0 {
				values[phi.Place.Identifier] = identifiersOf(aggregate, effectDeps)
			}
		}

		for _, instructionId := range block.Instructions {
			instruction := effectFunction.Instructions[instructionId]
			if instruction == nil {
				continue
			}

			switch value := instruction.Value.(type) {
			// Upstream's inert set: a constant contributes no taint and is not a reason to give up.
			case *high_level_intermediate_representation.Primitive, *high_level_intermediate_representation.JsxText, *high_level_intermediate_representation.LoadGlobal:

			case *high_level_intermediate_representation.LoadLocal:
				if deps, ok := values[value.Place.Identifier]; ok {
					values[instruction.LValue.Identifier] = deps
				}

			// `LoadContext` is not in upstream's list and is here because our lowering produces it
			// where upstream's produces a plain load: a captured value read inside the callback is a
			// context read for us. Without this arm every dep read inside the effect loses its taint
			// at the first use and the rule can never report. Measured: removing it makes all five
			// reporting fixtures go silent.
			case *high_level_intermediate_representation.LoadContext:
				if deps, ok := values[value.Place.Identifier]; ok {
					values[instruction.LValue.Identifier] = deps
				}

			case *high_level_intermediate_representation.ComputedLoad, *high_level_intermediate_representation.PropertyLoad, *high_level_intermediate_representation.BinaryExpression, *high_level_intermediate_representation.TemplateLiteral,
				*high_level_intermediate_representation.CallExpression, *high_level_intermediate_representation.MethodCall:
				aggregate := map[static_single_assignment.IdentifierId]bool{}
				high_level_intermediate_representation.EachPlace(instruction.Value, func(place high_level_intermediate_representation.Place, _ high_level_intermediate_representation.PlaceRole) {
					for _, dep := range values[place.Identifier] {
						aggregate[dep] = true
					}
				})
				if len(aggregate) != 0 {
					values[instruction.LValue.Identifier] = identifiersOf(aggregate, effectDeps)
				}

				call, isCall := instruction.Value.(*high_level_intermediate_representation.CallExpression)
				// The `Spread` half is SUBSUMED, measured rather than assumed. `setV(...[x])`
				// does lower to a single spread argument, so the branch is reachable in the sense
				// that matters, but the effect never gets this far: building the spread's operand
				// requires an `ArrayExpression` instruction, which is not in the allowed set above
				// and so hits `default: return` first. Spreading a NAMED array is the same story
				// one step earlier, since the array is built in the enclosing function and the
				// callback then captures a value that is neither a setter nor a dep.
				//
				// Measured both ways on the spread shape: pristine 0, mutant 0. The argument-COUNT
				// half of this condition is load-bearing and its own mutant dies on
				// `setterCalledWithTwoArguments`.
				if !isCall || len(call.Args) != 1 || call.Args[0].Spread {
					break
				}
				if !isSetterTyped(ctx, effectFunction, call.Callee) {
					break
				}
				deps, hasDeps := values[call.Args[0].Place.Identifier]
				if hasDeps && distinctCount(deps) == len(effectDeps) {
					setStateLocations = append(setStateLocations, call.Callee)
					break
				}
				// A setter call whose argument carries no taint, or taint from only some of the
				// deps, abandons the WHOLE effect rather than skipping this call. That is
				// upstream's `return`, and it is why setting from one of two deps is silent even
				// when a second call in the same effect sets from both.
				return

			default:
				// Anything else, including every store, object literal and JSX element, means the
				// effect is doing more than deriving, so upstream gives up on it.
				return
			}
		}

		// A derived value read by a terminal is a branch condition or a return, which is again more
		// than a derivation.
		abandon := false
		high_level_intermediate_representation.EachTerminalPlace(block.Terminal, func(place high_level_intermediate_representation.Place, _ high_level_intermediate_representation.PlaceRole) {
			if _, ok := values[place.Identifier]; ok {
				abandon = true
			}
		})
		if abandon {
			return
		}

		seenBlocks[block.Id] = true
	}

	for _, location := range setStateLocations {
		reportDerivedSetter(ctx, effectFunction, location)
	}
}

// reportDerivedSetter points the finding at the setter's reference inside the effect.
//
// Same mechanism and same reasoning as `reportSetter` in `set_state_in_effect.go`: a Place carries
// the span of the REFERENCE while `Identifier.Node` carries the span of the binding, so reporting
// through the node points every finding at the `useState` destructure. The range is trimmed of
// leading trivia because a Place's range starts at the raw start of the reference, which includes
// the newline and indentation before it.
func reportDerivedSetter(ctx rule.Context, function *high_level_intermediate_representation.Function, setter high_level_intermediate_representation.Place) {
	if span := trimmedRange(ctx, setter.Range); span.Pos() < span.End() {
		ctx.ReportRange(span, messageNoDerivingStateInEffects)
		return
	}
	if node := setStateInEffectIdentifierNodeOf(function, setter.Identifier); node != nil {
		ctx.ReportNode(node, messageNoDerivingStateInEffects)
	}
}

// isUseEffectExactly answers whether a callee resolves to React's `useEffect` and nothing else.
//
// # Why this is not `effectHookKind`, which is the sibling helper sitting right there
//
// `set-state-in-effect` accepts three hooks, because its pass calls `isUseEffectHookType ||
// isUseLayoutEffectHookType || isUseInsertionEffectHookType` (`ValidateNoSetStateInEffects.ts:116`).
// This pass calls `isUseEffectHookType` alone (`ValidateNoDerivedComputationsInEffects.ts:69`) and
// imports no other predicate. Two rules in the same family, one hook apart.
//
// Reusing the neighbouring helper is exactly the "read siblings for shape, never for semantics"
// trap, and it was written that way first. No imported fixture could see it: the corpus writes
// `useLayoutEffect` into a derived effect nowhere at all. It was caught by running the shipped rule
// on an invented `useLayoutEffect` case, which is silent upstream and reported here, and that case
// is now a fixture.
//
// The resolution mechanism is the sibling's and is correct to share: read the type's symbol rather
// than `GetSymbolAtLocation`, so a renamed import reports under its DECLARED name and a same-named
// local does not. Only the accepted set differs.
// Kept in this file rather than lifted out of `effectHookKind`, and that is a deliberate scar
// rather than duplication left by accident. That function is `set_state_in_effect.go`'s and encodes
// its own accepted set in the same switch that does the resolution, so factoring the resolution out
// means editing a file another rule owns while it is being worked. The two copies are the same four
// lines of checker reads; when both rules are settled they want one shared
// `declaredHookName(ctx, function, place) string` in the package, with each rule keeping only its
// own accepted set. That is one commit, after, not a unilateral edit now.
func isUseEffectExactly(ctx rule.Context, function *high_level_intermediate_representation.Function, callee high_level_intermediate_representation.Place) bool {
	node := setStateInEffectIdentifierNodeOf(function, callee.Identifier)
	if node == nil || ctx.TypeChecker == nil {
		return false
	}
	// The TYPE's symbol, not `GetSymbolAtLocation`'s: for `import {useEffect as useFx}` the latter
	// returns the local name `useFx` while this returns the declared `useEffect`, which is the name
	// React's `binding.imported` asks for. Reading the local spelling silently declines every
	// renamed import, and a namespace member falls out of the same read with no extra case.
	calleeType := ctx.TypeChecker.GetTypeAtLocation(node)
	if calleeType == nil {
		return false
	}
	symbol := checker.Type_symbol(calleeType)
	return symbol != nil && symbol.Name == "useEffect"
}

// dependencyIsFoldedConstant answers whether a dependency is a value upstream's constant
// propagation would have erased before this validator ever ran.
//
// # Why this exists, which is a pass we do not have rather than a rule upstream states
//
// `constantPropagation` runs at `Entrypoint/Pipeline.ts:197` and this validator at line 271, so by
// the time upstream judges an effect, a `const lastName = 'Swift'` read inside the callback is a
// `Primitive` in the callback's own body rather than a capture. The dependency array still names
// `lastName`, so upstream's "every dep is actually captured" check fails and it abandons the effect.
// The silence is a side effect of the fold, not a rule anybody wrote down.
//
// Cohere has no constant propagation, so our lowering keeps the capture and every one of those
// checks passes. Three of upstream's clean fixtures reported without this. The lowered form was
// printed rather than guessed at: the enclosing function holds `$23 = Primitive Swift` then
// `$24 = StoreLocal const lastName$25 = $23`, and the callback holds `$15 = LoadContext lastName$7`
// where upstream would hold a primitive.
//
// # The boundary was measured rather than assumed
//
// Driving React 7.1.1 on single-variable edits of one shape:
//
//	const lastName = 'Swift'   listed as a dep    silent
//	const n = 42               listed as a dep    silent
//	const n = 40 + 2           listed as a dep    silent    (upstream folds the arithmetic too)
//	const o = {x: 1}           listed as a dep    REPORTS   (an object is not a constant)
//	const lastName = compute() listed as a dep    REPORTS
//	const lastName = 'Swift'   NOT listed as a dep REPORTS  (folded on both sides, so no mismatch)
//
// The last row is why this declines the whole effect rather than dropping the dep from the list: a
// constant that is not named in the array changes nothing on either side, and a constant that IS
// named is precisely the mismatch that makes upstream give up.
//
// # What is deliberately not reproduced
//
// Only a directly-assigned literal is recognised, not the full fold. `const n = 40 + 2` is silent
// upstream and REPORTS here, because our lowering leaves the addition as a `BinaryExpression` and
// nothing folds it. That is a stated divergence in the reporting direction on a shape nobody writes
// into a dependency array, and closing it means porting constant propagation, which is a pass for
// the intermediate representation rather than a helper for this rule. Recorded here so the next
// reader can see the limit was measured rather than assumed, and so it is findable when the fold
// lands: this function should be deleted then, not extended.
func dependencyIsFoldedConstant(function *high_level_intermediate_representation.Function, dependency static_single_assignment.IdentifierId) bool {
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			store, isStore := instruction.Value.(*high_level_intermediate_representation.StoreLocal)
			if !isStore || store.LValue.Identifier != dependency {
				continue
			}
			// The stored value's own defining instruction decides it. A `Primitive` is what
			// upstream folds; anything else it leaves alone.
			return definesPrimitive(function, store.Value.Identifier)
		}
	}
	return false
}

// definesPrimitive answers whether a value was produced by a literal.
func definesPrimitive(function *high_level_intermediate_representation.Function, identifier static_single_assignment.IdentifierId) bool {
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil || instruction.LValue.Identifier != identifier {
				continue
			}
			_, isPrimitive := instruction.Value.(*high_level_intermediate_representation.Primitive)
			return isPrimitive
		}
	}
	return false
}

// containsIdentifier is a linear scan because a dependency array is short. A map would cost a
// build per effect to answer at most a handful of questions.
func containsIdentifier(identifiers []static_single_assignment.IdentifierId, want static_single_assignment.IdentifierId) bool {
	for _, identifier := range identifiers {
		if identifier == want {
			return true
		}
	}
	return false
}

// identifiersOf flattens a taint set into a slice, ordered by the dependency array rather than by
// map iteration.
//
// The order is not observable through any finding, since the only question ever asked of the result
// is how many distinct entries it holds. It is fixed anyway because an unordered slice stored into a
// map that later feeds a phi aggregate makes the whole pass produce different intermediate values
// between runs, which is the class of bug `TestLowerIsDeterministic` exists to catch one layer down.
func identifiersOf(set map[static_single_assignment.IdentifierId]bool, order []static_single_assignment.IdentifierId) []static_single_assignment.IdentifierId {
	result := make([]static_single_assignment.IdentifierId, 0, len(set))
	for _, identifier := range order {
		if set[identifier] {
			result = append(result, identifier)
		}
	}
	return result
}

// distinctCount is upstream's `new Set(deps).size`, and here it is EQUIVALENT to `len`, which is
// stated rather than left as a surviving mutant.
//
// Replacing the body with `len(identifiers)` survives the whole suite. That is not a fixture gap:
// no slice reaching this function can hold a repeat, because every writer of the taint map produces
// a duplicate-free slice. There are exactly five, and they were enumerated rather than assumed:
//
//	the capture seed          a one-element literal
//	the phi aggregate         built through identifiersOf, which iterates a SET
//	the two load arms         copy an existing slice, so they inherit the property
//	the operand aggregate     built through identifiersOf, same as the phi
//
// `identifiersOf` walks `effectDeps` and emits each entry at most once, and `effectDeps` itself
// carries one entry per element of the dependency array. So a repeat would need the same identifier
// written into the dependency array twice, which upstream's own `new Set` is what handles, and which
// our taint aggregate has already collapsed one step earlier.
//
// The reachability was probed rather than argued: six shapes that could plausibly produce a repeat
// were run against both versions, including a ternary with identical arms, a logical operator over
// one dep, and `a + a` in a single expression. Zero disagreements. The `a + a` case is the
// interesting one, since it reports under both versions, proving the path IS reached and the count
// still agrees.
//
// Upstream's spelling is kept because it is upstream's, and because the property above is a
// consequence of how this file happens to build its slices rather than something the type system
// enforces. A future writer that appends without deduplicating would make the difference real, and
// this comment is what tells them the guard is load-bearing again at that point.
func distinctCount(identifiers []static_single_assignment.IdentifierId) int {
	seen := map[static_single_assignment.IdentifierId]bool{}
	for _, identifier := range identifiers {
		seen[identifier] = true
	}
	return len(seen)
}
