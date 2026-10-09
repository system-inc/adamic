// Checking that a developer's own memoization survived compilation.
//
// This is React's `validatePreservedManualMemoization`
// (`Validation/ValidatePreservedManualMemoization.ts` at `react_conformance.UpstreamSha`), and it is
// the rule the whole of phase 6 exists to enable.
//
// # What it reports
//
// A developer who wrote `useMemo` or `useCallback` made a claim: this value is stable across renders
// unless these dependencies change. The compiler is free to rewrite that, and normally does it
// better -- but if the rewrite ends up not memoizing the value, the developer's guarantee silently
// disappears. Code that relied on referential stability then breaks in a way nothing else catches.
//
// So the rule compares what was written against what survived, and reports where they disagree.
//
// # Two firing conditions, both about scopes rather than about dependencies
//
// At `StartMemoize`, an operand belonging to a scope that neither survived nor was pruned means the
// dependency may be mutated after the memo block, so the memoization cannot be trusted.
//
// At `FinishMemoize`, a memoized value belonging to a scope that did not survive means the value was
// memoized in source and is not in output. That is the headline case.
//
// Both read `identifier.scope`. Upstream keeps that on the identifier; this tree keeps it in the
// `ReactiveScopes` side table, which answers the same question -- 15,715 of 61,738 corpus
// identifiers carry one.
package high_level_intermediate_representation

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/static_single_assignment"
)

// PreserveManualMemoizationFinding is one disagreement between source and output.
type PreserveManualMemoizationFinding struct {
	// Identifier is the value whose memoization was lost.
	Identifier static_single_assignment.IdentifierId
	// Scope is the scope it belonged to, which did not survive.
	Scope ScopeId
	// Order is where in the function the finding sits, for reporting.
	Order static_single_assignment.EvaluationOrder
	// Kind separates the two conditions, which carry different messages upstream.
	Kind PreserveManualMemoizationKind
}

// PreserveManualMemoizationKind is which of the two conditions fired.
type PreserveManualMemoizationKind uint8

const (
	// PreserveManualMemoizationDependencyMutable is the `StartMemoize` condition: a dependency
	// belonging to a scope that neither survived nor was pruned, so it may be modified later.
	PreserveManualMemoizationDependencyMutable PreserveManualMemoizationKind = iota
	// PreserveManualMemoizationValueUnmemoized is the `FinishMemoize` condition: a value memoized in
	// source and not in the compilation output.
	PreserveManualMemoizationValueUnmemoized
)

func (kind PreserveManualMemoizationKind) String() string {
	switch kind {
	case PreserveManualMemoizationDependencyMutable:
		return "this dependency may be modified later"
	case PreserveManualMemoizationValueUnmemoized:
		return "could not preserve existing memoization"
	default:
		return "<unknown preserve-manual-memoization kind>"
	}
}

// ValidatePreservedManualMemoization reports where a developer's memoization did not survive.
//
// The scopes that survived are collected by the same walk that checks, which is upstream's shape and
// is load-bearing: a scope is only known to have survived once the walk reaches it, so a memo block
// is checked against the scopes closed before it rather than against every scope in the function.
func ValidatePreservedManualMemoization(tree *ReactiveFunction, function *Function,
	scopes *ReactiveScopes) []PreserveManualMemoizationFinding {
	if tree == nil || function == nil {
		return nil
	}
	return ValidatePreservedManualMemoizationWithDependencies(tree, function, scopes, nil)
}

// ValidatePreservedManualMemoizationWithDependencies adds the third firing condition.
//
// Separate entry point rather than a changed signature, matching `AssignReactiveScopesWithSets` and
// `PruneNonEscapingScopesWithScopes`: the inferred-versus-written comparison needs the collected
// dependency table and the temporaries map inside it, and neither is derivable from a tree.
//
// A caller that passes nil gets the two scope conditions and nothing else, which is what every
// caller had before this existed.
func ValidatePreservedManualMemoizationWithDependencies(tree *ReactiveFunction, function *Function,
	scopes *ReactiveScopes, dependencies *ScopeDependencies) []PreserveManualMemoizationFinding {
	return ValidatePreservedManualMemoizationWithPruned(tree, function, scopes, dependencies, nil)
}

// ValidatePreservedManualMemoizationWithPruned is the same validation, told which scopes the prune
// chain replaced rather than marked.
//
// `PruneNonEscapingScopes` replaces a scope block with its own instructions, so after it runs nothing
// in the tree records that the scope existed. The walk therefore sees neither a live scope block nor
// a pruned one, and the `DependencyMutable` condition -- which asks exactly "scoped, not survived,
// not pruned" -- reaches a fourth state upstream cannot: scoped, and absent.
//
// Upstream reads a `prunedScopes` set its own pass wrote, so it answers the three-part question with
// three states. Passing the set here does the same. A caller with nothing to pass keeps the previous
// behaviour, which is what the exported wrapper above preserves.
func ValidatePreservedManualMemoizationWithPruned(tree *ReactiveFunction, function *Function,
	scopes *ReactiveScopes, dependencies *ScopeDependencies,
	prunedByChain map[ScopeId]bool) []PreserveManualMemoizationFinding {
	if tree == nil || function == nil {
		return nil
	}
	validator := manualMemoValidator{
		function:         function,
		scopes:           scopes,
		liveScopes:       map[ScopeId]bool{},
		prunedScopes:     map[ScopeId]bool{},
		walkedScopes:     map[ScopeId]bool{},
		prunedByChain:    prunedByChain,
		openMemoBlocks:   map[int]bool{},
		dependencies:     dependencies,
		sourceDeps:       map[int][]ManualMemoDependency{},
		declsInMemoBlock: map[static_single_assignment.DeclarationId]bool{},
	}
	validator.walk(tree.Body)
	return validator.findings
}

type manualMemoValidator struct {
	function *Function
	scopes   *ReactiveScopes
	// liveScopes are scopes the walk has passed that survived, plus everything they absorbed.
	liveScopes map[ScopeId]bool
	// prunedScopes are scopes the walk has passed that were pruned.
	prunedScopes map[ScopeId]bool
	// prunedByChain are scopes the prune chain REPLACED, which leaves no block for the walk to see.
	prunedByChain map[ScopeId]bool
	// walkedScopes are scopes the walk entered at all, whether they survived or were pruned.
	//
	// Distinct from the other two rather than derivable from them: a scope absent from both is
	// either one the walk has not reached yet or one that is not in the tree at all, and only this
	// separates those.
	walkedScopes map[ScopeId]bool
	// openMemoBlocks are the `ManualMemoId`s of memo blocks opened and not yet closed.
	//
	// A set keyed by id rather than a boolean, because `ManualMemoId` exists precisely to make
	// pairing possible: memo calls nest -- a `useMemo` whose callback body holds another -- and once
	// the callbacks are inlined the markers do not form a simple stack in instruction order. A
	// boolean would let an inner `FinishMemoize` close an outer block, and the outer block's own
	// finish would then be dropped as unopened.
	openMemoBlocks map[int]bool
	// dependencies is the collected scope-dependency table, or nil when the caller did not supply
	// one. Nil disables the inferred-versus-written comparison and leaves the two scope conditions
	// unchanged, which is the state every caller was in before that comparison was wired.
	dependencies *ScopeDependencies
	// sourceDeps are the dependencies the developer wrote in the currently open memo block, keyed
	// by `ManualMemoId`. Upstream keeps one `manualMemoState` because it asserts memo blocks do not
	// nest; this keys by id for the same reason `openMemoBlocks` does.
	sourceDeps map[int][]ManualMemoDependency
	// declsInMemoBlock are declarations made inside an open memo block, by `DeclarationId`.
	//
	// Upstream's `manualMemoState.decls`. A scope dependency rooted at a value the memo block
	// itself declared is not a dependency the developer could have written, so it is skipped rather
	// than reported. Without this the comparison reports every intermediate value in the callback.
	declsInMemoBlock map[static_single_assignment.DeclarationId]bool
	findings         []PreserveManualMemoizationFinding
}

func (v *manualMemoValidator) walk(block ReactiveBlock) {
	for _, statement := range block {
		switch shape := statement.(type) {
		case *ReactiveInstructionStatement:
			v.visitInstruction(shape.Instruction)

		case *ReactiveScopeBlock:
			v.walkedScopes[shape.Scope] = true
			v.walk(shape.Instructions)
			// A pruned scope is recorded and NOT compared, which is upstream's split into two
			// methods: `visitScope` runs `validateInferredDep` over the scope's dependencies,
			// `visitPrunedScope` only adds the id to `prunedScopes`
			// (`ValidatePreservedManualMemoization.ts:413` and `:441`).
			//
			// One arm for both was a real defect rather than a shape difference. A pruned scope
			// still carries the dependency table entry it was assigned, so comparing it re-reports
			// whatever its surviving twin already agreed about. Measured on
			// `useCallback-alias-property-load-dep.ts`: two scopes both carry `propA.x` and `x`,
			// scope 2 is pruned, and upstream emits one guard naming exactly that pair.
			//
			// Recorded after the body, matching upstream: a scope is known to have survived only
			// once the walk has left it, so a memo block inside it is checked against the scopes
			// that closed before it rather than against its own enclosing scope.
			if shape.Pruned {
				v.prunedScopes[shape.Scope] = true
				continue
			}
			// The inferred-versus-written comparison, after the body and before the scope is
			// recorded live.
			v.compareInferredDependencies(shape)
			v.liveScopes[shape.Scope] = true
			for _, absorbed := range shape.Merged {
				// A scope absorbed by a merge survived under its survivor's identity, so the ids it
				// carried are live too. This is the only consumer of `Merged`, and the reason the
				// merge pass was owed.
				v.liveScopes[absorbed] = true
			}

		case *ReactiveTerminalStatement:
			v.walkTerminal(shape)
		}
	}
}

func (v *manualMemoValidator) visitInstruction(instruction *ReactiveInstruction) {
	if instruction == nil || instruction.Value == nil {
		return
	}
	// A sequence carries instructions inside a value; a memo marker can sit in one.
	if sequence, isSequence := instruction.Value.(*ReactiveSequenceValue); isSequence {
		for _, nested := range sequence.Instructions {
			v.visitInstruction(nested)
		}
	}
	plain, isPlain := instruction.Value.(*ReactiveInstructionValue)
	if !isPlain || plain.Value == nil {
		return
	}

	switch marker := plain.Value.(type) {
	case *StartMemoize:
		// Upstream carries two early returns here that this port deliberately does not.
		//
		// `value.deps == null` needs no code: upstream's null check exists because iterating null
		// throws in JavaScript, and ranging a nil slice below iterates zero times. Same behaviour,
		// no branch, and a branch added to mirror the source would be dead.
		//
		// `value.hasInvalidDeps` is unreachable rather than unported, which is a different verdict
		// and expires differently. Measured at the source: the flag is written in exactly one place,
		// `ValidateExhaustiveDependencies.ts:146`, inside `onFinishMemoize`, and only when
		// `env.configuration.validateExhaustiveMemoizationDependencies` is on and `validateDependencies`
		// already produced a diagnostic. It is a duplicate-error suppressor for a rule that has
		// already reported, not a false-positive guard.
		//
		// This tree does not run that pass at all -- `internal/rules/react/exhaustive_deps.go` is
		// the ESLint rule of the same subject, not the compiler validation -- so nothing here could
		// set the flag and a port of the guard would read as protection it does not provide.
		//
		// The verdict expires if that validation is ever ported: at that point this guard becomes
		// live and must land with it, or every program failing exhaustive-deps reports twice.
		v.openMemoBlocks[marker.ManualMemoId] = true
		if marker.Deps != nil && v.sourceDeps != nil {
			v.sourceDeps[marker.ManualMemoId] = marker.Deps
		}
		// A dependency belonging to a scope that neither survived nor was pruned may be mutated
		// after this point, so the memoization cannot be trusted.
		for _, dependency := range marker.Deps {
			if dependency.Root.IsGlobal {
				continue
			}
			v.check(dependency.Root.Place.Identifier, instruction.Order,
				PreserveManualMemoizationDependencyMutable)
		}

	case *FinishMemoize:
		if !v.openMemoBlocks[marker.ManualMemoId] {
			// Upstream returns early: a StartMemoize with invalid deps records no state, so its
			// FinishMemoize has nothing to close and validating it would report against a block
			// that was never opened.
			return
		}
		delete(v.openMemoBlocks, marker.ManualMemoId)
		delete(v.sourceDeps, marker.ManualMemoId)
		if marker.Pruned {
			// A pruned memo block was deliberately discarded, so there is nothing to preserve.
			return
		}
		v.check(marker.Value.Identifier, instruction.Order,
			PreserveManualMemoizationValueUnmemoized)

	default:
		// Every other instruction, for the decls set only.
		//
		// Upstream records a declaration into the open memo block at two sites: any named lvalue
		// (:394-397) and every lvalue of a `StoreLocal`, `StoreContext` or `Destructure` (:361-371).
		// Both reduce here to "what this instruction defines", because a value declared inside the
		// callback is not something the developer could have named in a dependency array.
		if len(v.openMemoBlocks) == 0 || v.declsInMemoBlock == nil {
			return
		}
		EachPlace(plain.Value, func(place Place, role PlaceRole) {
			if role == PlaceRoleDefine {
				v.declsInMemoBlock[declarationOf(v.function, place.Identifier)] = true
			}
		})
		if instruction.LValue != nil {
			v.declsInMemoBlock[declarationOf(v.function, instruction.LValue.Identifier)] = true
		}
	}
}

// check reports the identifier if it belongs to a scope that did not survive.
//
// Upstream's `isUnmemoized`: an identifier with no scope is fine -- that is its proxy for a
// primitive or global, which needs no memoization -- and one whose scope survived is fine. Only a
// scoped identifier whose scope is absent from the live set is a lost memoization.
//
// The `StartMemoize` condition additionally accepts a pruned scope, because a pruned scope was a
// deliberate decision rather than a lost one. The `FinishMemoize` condition does not.
func (v *manualMemoValidator) check(identifier static_single_assignment.IdentifierId, order static_single_assignment.EvaluationOrder,
	kind PreserveManualMemoizationKind) {
	if v.scopes == nil {
		return
	}
	scope := v.scopes.ScopeOf(identifier)
	if scope == 0 {
		return
	}
	if v.liveScopes[scope] {
		return
	}
	if kind == PreserveManualMemoizationDependencyMutable && v.prunedScopes[scope] {
		return
	}
	// # A scope the walk never entered is one this condition cannot judge
	//
	// Upstream asks the same three-part question -- scoped, not survived, not pruned -- but every
	// scope it asks about lives in the reactive tree it is walking, so "scoped but absent from the
	// tree" is not a state it can encounter and its behaviour there is unspecified rather than
	// matched.
	//
	// It is the dominant state here. Measured across the clean corpus, of 55 checks that get past
	// the scopeless return: 35 are on a scope the walk never entered, 20 on a live one, and zero on
	// a pruned one. Without this the rule reads every one of those 35 as "neither survived nor
	// pruned, so it may be modified later" and reports.
	//
	// `prune-nonescaping-useMemo.ts` is the shape. After `be5b49c` sweeps its body it keeps exactly
	// what upstream keeps -- the two markers, the callback, the call -- and upstream compiles it to
	// `function useFoo(x) {}` with no cache slots at all. Its own header calls reporting here
	// "technically a false positive". `x` still carries a scope in our table, from the call inside
	// the callback, and that scope has no block in the tree.
	//
	// Measured: false positives 31 to 25, golden unchanged at 25, and all three oracles byte
	// identical -- `matched` 77, `ours` 120, `under` 5 fixtures / 7 scopes.
	if kind == PreserveManualMemoizationDependencyMutable && !v.walkedScopes[scope] &&
		!v.prunedByChain[scope] {
		return
	}
	// # A component or hook parameter's scope is one it inherited, not one it earned
	//
	// `useMemo(() => [x.y.z], [x])` inside `useHook(x)` puts `x` and the callback in one class: the
	// invocation widens the callback's range across the memo call, and the capture widens `x` with
	// it. Measured, `x` takes range {2,7} and a scope spanning the whole memo block. Upstream's
	// compiled guard for that fixture is `$[0] !== x.y.z`, so its scope depends on the path and `x`
	// itself carries no scope at all.
	//
	// The condition above then reports, because a scope the memo marker sits inside is not yet
	// classified when this runs -- a scope is recorded only after the walk leaves it, which is
	// upstream's order too (`ValidatePreservedManualMemoization.ts:412`). Upstream reaches the same
	// intermediate state and stays silent, since its dependency has no scope to ask about.
	//
	// This is the escape upstream's own comment describes: "due to current limitations of mutable
	// range inference, there are edge cases in which we infer known-immutable values (e.g. props or
	// hook params) to have a mutable range and scope" (`CollectHoistablePropertyLoads.ts:105`). It
	// builds `knownImmutableIdentifiers` for exactly this, and consumes it only inside that file --
	// as did this tree, so the fact was known and never reached the place the rule reads.
	//
	// # Why here rather than in the partition or in scope assignment
	//
	// Both were built and measured, and both score identically to this. Filtering at scope
	// construction breaks `TestScopeMatchesTheDisjointSetPartition`, which requires every class to
	// map to a scope with the same members. Filtering at the partition satisfies that and then
	// breaks two more: the hoistable corpus distribution rises 574 deep to 692, which that test
	// calls the over-approximating direction, and the reactive tree loses instructions in five
	// functions and emits twenty-two blocks twice. The parameter's membership is load-bearing for
	// the tree even though it is wrong for this rule, so the exemption belongs where the rule reads
	// rather than where the scopes are made.
	//
	// Measured: false positives 25 to 15, ten fixtures, none newly firing, every oracle unchanged --
	// `matched` 77, `ours` 120, `under` 5 fixtures / 7 scopes -- and every structural invariant
	// intact.
	//
	// A mutation dropping the `kind` gate SURVIVES: the value condition checks a memoized value
	// rather than a written dependency, and no corpus fixture memoizes a bare parameter, so it never
	// asks about one. Kept narrow because the two conditions ask different questions and an
	// exemption argued from the dependency side should not silently answer the other.
	if kind == PreserveManualMemoizationDependencyMutable &&
		isKnownImmutableParameter(v.function, identifier) {
		return
	}
	v.findings = append(v.findings, PreserveManualMemoizationFinding{
		Identifier: identifier,
		Scope:      scope,
		Order:      order,
		Kind:       kind,
	})
}

// walkTerminal descends into every block a terminal contains.
func (v *manualMemoValidator) walkTerminal(statement *ReactiveTerminalStatement) {
	switch shape := statement.Terminal.(type) {
	case *ReactiveIf:
		v.walk(shape.Consequent)
		if shape.Alternate != nil {
			v.walk(*shape.Alternate)
		}
	case *ReactiveSwitch:
		for index := range shape.Cases {
			if shape.Cases[index].Block != nil {
				v.walk(*shape.Cases[index].Block)
			}
		}
	case *ReactiveFor:
		v.walk(shape.Loop)
	case *ReactiveForOf:
		v.walk(shape.Loop)
	case *ReactiveForIn:
		v.walk(shape.Loop)
	case *ReactiveWhile:
		v.walk(shape.Loop)
	case *ReactiveDoWhile:
		v.walk(shape.Loop)
	case *ReactiveLabelTerminal:
		v.walk(shape.Block)
	case *ReactiveTry:
		v.walk(shape.Block)
		v.walk(shape.Handler)
	}
}

// AnalyzePreservedManualMemoization runs the reactive-scope pipeline over one function and validates.
//
// Exported so the rule surface does not have to know the pass order. That order is upstream's and is
// load-bearing rather than incidental: the validation reads which scopes survived, which were pruned
// and which absorbed others in a merge, and each of those facts is written by a different pass. A
// caller running them in a different order gets a validation against a half-built world, which
// reports plausibly and is wrong.
//
// Expects `Lower` and `Construct` to have run already, since a caller that has a `*Function` has
// done both by definition.
func AnalyzePreservedManualMemoization(function *Function,
	typeChecker *shimchecker.Checker) []PreserveManualMemoizationFinding {
	if function == nil {
		return nil
	}
	OutlineFunctions(function)
	InferReactive(function, typeChecker)
	DropManualMemoization(function)

	if InlineImmediatelyInvokedFunctionExpressionsIncludingMemoCallbacks(function) > 0 {
		MergeConsecutiveBlocks(function)
	}
	EliminateDeadCode(function)

	ranges := InferMutableRanges(function)
	set := FindDisjointMutableValuesWithRanges(function, ranges)
	scopes := AssignReactiveScopesWithSets(function, ranges, set)
	aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
	identity := MergedScopeIdentity{Aligned: aligned, Merged: merged}
	BuildReactiveScopeTerminals(function, scopes, identity)

	// Upstream's position: immediately after the scope terminals exist and before dependencies are
	// collected. A scope inside a loop is pruned as policy rather than as a failure, and the
	// validator has to be able to tell those apart -- see `flatten_reactive_loops.go`.
	flattenedScopes := FlattenReactiveLoops(function)
	// Then the scopes a hook call sits in, upstream's next pass. A pruned hook scope is no parent to
	// the scopes nested in it, so a memo callback inside one keeps its own scope through the merge.
	for scope := range FlattenScopesWithHooksOrUse(function) {
		if flattenedScopes == nil {
			flattenedScopes = map[ScopeId]bool{}
		}
		flattenedScopes[scope] = true
	}

	dependencies := CollectScopeDependenciesWithHoistable(function, scopes, identity, ranges)

	tree, _ := BuildReactiveFunctionWithFlattenedScopes(function, flattenedScopes)
	if tree == nil {
		return nil
	}

	nonEscaping := PruneNonEscapingScopesWithScopes(tree, function, dependencies, scopes, typeChecker)
	PruneNonReactiveDependencies(tree, function, dependencies)
	PruneUnusedScopes(tree, dependencies)
	MergeReactiveScopesThatInvalidateTogether(tree, function, dependencies, typeChecker)
	PruneAlwaysInvalidatingScopes(tree, function, dependencies)

	// The third firing condition, on. This gate was closed and the reason it was closed is gone.
	//
	// It read: "the comparison itself is built and tested. Its INPUT is not ready:
	// `CollectScopeDependencies` truncates a path to its root wherever the hoistable set is empty,
	// and the hoistable analysis is `DependencyGapNullPropagation`, declined." That analysis landed.
	// `CollectHoistablePropertyLoads` populates the tree, the collector descends into nested
	// functions and into callbacks it can assume are invoked, and 595 dependencies now carry a path
	// against 2,092 that do not -- where the gate was written, 171 did against 1,435.
	//
	// Its own instruction was "turn this on when the hoistable analysis lands, and expect the
	// ungated numbers to be the ones that move." Both happened. The numbers it recorded --
	// golden 15 to 25 for clean 27 to 37 -- are stale by a dozen commits; measured now, the trade
	// is golden 15 to 19 for one additional false positive.
	//
	// That one is `useCallback-alias-property-load-dep.ts`, traced to a scope boundary rather than
	// to this comparison: `const x = propB.x.y` is declared inside our scope 1, so
	// `checkValidDependency` rejects `x` on the rule this tree shares with upstream verbatim
	// (`decl.order < scopeRange.Start`), while upstream's scope begins after that declaration and
	// names `x`. It belongs to the scope work rather than here.
	return ValidatePreservedManualMemoizationWithPruned(tree, function, scopes, dependencies,
		nonEscaping.PrunedScopes)
}

// ForEachFunctionLike calls visit for every outermost function-like node under root.
//
// Exported for the rule surface. Outermost only, which is the same denominator every measurement in
// this package uses: a nested function is reached through `Function.Functions` rather than by
// walking into it, because an IdentifierId names one value in this function and a different value in
// a nested one.
func ForEachFunctionLike(root *ast.Node, visit func(*ast.Node)) {
	if root == nil {
		return
	}
	root.ForEachChild(func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			visit(node)
			return false
		}
		ForEachFunctionLike(node, visit)
		return false
	})
}

// compareInferredDependencies is the third firing condition: what the compiler inferred as this
// scope's dependencies against what the developer wrote in the memo block enclosing it.
//
// Upstream's `validateInferredDep` (`ValidatePreservedManualMemoization.ts:228-300`), driven from
// `visitScope` (:417-433). It reports at most one finding per inferred dependency, and only when
// that dependency matches NO source dependency, keeping the strongest disagreement seen.
//
// # Why this runs on scope exit rather than at the marker
//
// The dependencies being compared are the SCOPE's, and a scope's dependency set is not complete
// until its body has been walked. The memo block's own markers sit inside that body, so the source
// side is available and the inferred side is not until this point. That is also why upstream places
// the loop before `this.scopes.add`: the comparison belongs to the scope being left, not to the
// scopes already closed.
//
// # The two early returns, both load-bearing
//
// A dependency whose root is a value the memo block itself declared is skipped. Those are the
// callback's own intermediates, which no developer could have written in a dependency array, and
// reporting them would flag every temporary in the body.
//
// A dependency that cannot be normalized is skipped rather than reported. Upstream raises an
// invariant there; see `NormalizeInferredDependency` for why declining is the honest answer in this
// tree.
func (v *manualMemoValidator) compareInferredDependencies(scope *ReactiveScopeBlock) {
	if scope == nil || v.dependencies == nil || len(v.sourceDeps) == 0 {
		return
	}
	// Upstream reads `state.manualMemoState`, a single open block, because it asserts memo blocks do
	// not nest. This walks whichever blocks are open; with the non-nesting invariant holding there
	// is exactly one, and if it ever does not hold this compares against each rather than silently
	// picking one.
	for id, sourceDependencies := range v.sourceDeps {
		if !v.openMemoBlocks[id] {
			continue
		}
		// An empty source array is not "nothing to compare against", it is the strictest thing a
		// developer can write. `useCallback(fn, [])` promises the value never changes, so every
		// inferred dependency contradicts it and upstream reports each one with "Inferred
		// dependency not present in source". Skipping the comparison here silenced exactly the
		// case the rule should report hardest.
		//
		// Measured: for a `[]` source array the `StartMemoize` carries a non-nil, zero-length
		// `Deps`, so the block does get an entry in `sourceDeps` and the loop below reaches it.
		// The matching check falls through with `matched` false for every inferred dependency,
		// which is the intended answer.
		for _, inferred := range v.dependencies.DependenciesOf(scope.Scope) {
			normalized, ok := v.dependencies.NormalizeInferredDependency(v.function, inferred)
			if !ok {
				continue
			}
			if !normalized.Root.IsGlobal &&
				v.declsInMemoBlock[declarationOf(v.function, normalized.Root.Place.Identifier)] {
				continue
			}
			matched := false
			for _, source := range sourceDependencies {
				if CompareManualMemoDependencies(normalized, source) == CompareDependencyOk {
					matched = true
					break
				}
			}
			if matched {
				continue
			}
			v.findings = append(v.findings, PreserveManualMemoizationFinding{
				Identifier: inferred.Identifier,
				Scope:      scope.Scope,
				Kind:       PreserveManualMemoizationValueUnmemoized,
			})
		}
	}
}
