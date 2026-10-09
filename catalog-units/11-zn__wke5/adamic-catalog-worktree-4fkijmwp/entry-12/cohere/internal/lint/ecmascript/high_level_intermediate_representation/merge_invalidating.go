// Merging adjacent scopes that invalidate together: the tree-level merge.
//
// This is React's `mergeReactiveScopesThatInvalidateTogether`
// (`ReactiveScopes/MergeReactiveScopesThatInvalidateTogether.ts` at `react_conformance.UpstreamSha`).
//
// # Not the merge this package already has, and the distinction is load-bearing
//
// `merge_scopes.go` is `mergeOverlappingReactiveScopesHIR`, upstream pipeline position 394. It runs
// on the GRAPH and exists as a precondition: it unions scopes whose ranges overlap without nesting,
// so that every scope can become a block. Without it `assertValidBlockNesting` fails outright.
//
// This is pipeline position 482, runs on the TREE, and is an optimisation rather than a
// precondition. It fuses ADJACENT scopes whose invalidation conditions are identical and removes
// nested scopes that repeat their enclosing scope's dependency set. Both affect memo validation.
//
// Confusing the two is easy and the names encourage it. The first asks "do these ranges overlap",
// the second asks "would these two always recompute together".
//
// # Why the rule needs it, which is the reason this is ported at all
//
// `scope.merged` -- the set of scope ids a surviving scope absorbed -- is written in exactly one
// place in the whole compiler, this pass, and read by `validatePreservedManualMemoization` to fold
// absorbed ids into the scope set it tracks. Producer and consumer with no codegen between them.
//
// That is the opposite of `promoteUsedTemporaries`, which was declined: the only consumer of what
// that pass wrote was the renamer.
package high_level_intermediate_representation

import (
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/static_single_assignment"
)

// LastUsage records the last instruction at which each declaration is read.
//
// Upstream's `FindLastUsageVisitor`. Keyed by `DeclarationId` rather than `IdentifierId`, which is
// upstream's choice and carries its own TODO saying so: declaration-keying is what makes the output
// compatible, and it is deliberately imprecise where a scope defines a variable whose version is
// never read and is overwritten later. Do not "fix" this to `IdentifierId` -- upstream's note says
// that is correct only once the pass leaves single-assignment form behind, which it has not.
type LastUsage struct {
	byDeclaration map[static_single_assignment.DeclarationId]static_single_assignment.EvaluationOrder
}

// LastUsedAt returns the last order at which a declaration was read, and whether it was read at all.
//
// The second return is not decoration. Upstream indexes its map with a non-null assertion and would
// throw on a miss; a linter cannot, and the two callers treat a miss differently -- one skips the
// declaration, the other declines the merge.
func (l *LastUsage) LastUsedAt(declaration static_single_assignment.DeclarationId) (static_single_assignment.EvaluationOrder, bool) {
	if l == nil || l.byDeclaration == nil {
		return 0, false
	}
	order, found := l.byDeclaration[declaration]
	return order, found
}

// Len reports how many declarations were seen, for measurement.
func (l *LastUsage) Len() int {
	if l == nil {
		return 0
	}
	return len(l.byDeclaration)
}

// FindLastUsage records, for every declaration, the highest evaluation order at which it is read.
//
// Walks the tree rather than the graph, because the merge that consumes it operates on the tree and
// the two do not agree: a value block's instructions are nested inside a value here and are
// ordinary instructions there.
// The graph is taken alongside the tree because a `Place` here names an `IdentifierId` and this
// pass keys on `DeclarationId`; the mapping between them lives on the graph's identifier table.
func FindLastUsage(tree *ReactiveFunction, function *Function) *LastUsage {
	usage := &LastUsage{byDeclaration: map[static_single_assignment.DeclarationId]static_single_assignment.EvaluationOrder{}}
	if tree == nil || function == nil {
		return usage
	}

	VisitReactiveFunction(tree, ReactiveVisitor{
		Place: func(order static_single_assignment.EvaluationOrder, place Place, role PlaceRole) {
			// `declarationOf` answers zero for an identifier missing from the table, which groups
			// every such value together. That is the shelf helper's documented behaviour and it is
			// acceptable here for the same reason it is there: the consumers compare a declaration
			// against a scope's own declarations, which are written from real places.
			declaration := declarationOf(function, place.Identifier)
			if previous, seen := usage.byDeclaration[declaration]; !seen || order > previous {
				usage.byDeclaration[declaration] = order
			}
		},
	})
	return usage
}

// AreEqualDependencies reports whether two dependency sets name the same inputs.
//
// Upstream's `areEqualDependencies`. Set equality rather than slice equality: order is an artifact
// of collection, so two scopes depending on the same things in a different sequence must compare
// equal or the merge would depend on traversal order.
//
// Path comparison goes through `equalPaths`, which compares `Optional` as well as `Property`. That
// matters here for the same reason it matters there: `a?.b` and `a.b` are different dependencies,
// and collapsing them would merge two scopes that invalidate under different conditions.
func AreEqualDependencies(a, b []ReactiveScopeDependency) bool {
	if len(a) != len(b) {
		return false
	}
	for _, left := range a {
		found := false
		for _, right := range b {
			if left.Identifier == right.Identifier && equalPaths(left.Path, right.Path) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// AreLValuesLastUsedByScope reports whether every named value is read for the last time inside scope.
//
// Upstream's `areLValuesLastUsedByScope`. A value still read after the scope ends cannot have its
// scope merged forward, because the merged scope would extend past the point where that value must
// already be settled.
//
// A declaration with no recorded usage answers FALSE, which is the conservative direction and
// differs from upstream only in that upstream would throw. A value never read is not a value proven
// safe to merge; it is a value this pass has no information about.
func AreLValuesLastUsedByScope(scopeEnd static_single_assignment.EvaluationOrder, lvalues []static_single_assignment.DeclarationId,
	usage *LastUsage) bool {
	for _, lvalue := range lvalues {
		lastUsedAt, found := usage.LastUsedAt(lvalue)
		if !found || lastUsedAt >= scopeEnd {
			return false
		}
	}
	return true
}

// ScopeIsEligibleForMerging reports whether a scope may be fused with the ones after it.
//
// Upstream's `scopeIsEligibleForMerging`, and its comment is the whole reasoning: merging is only
// safe when the scope's output is guaranteed to change whenever its input changes. When the output
// may not change, keeping the scopes separate lets the later one compare its input and skip the
// update -- so merging there would make the program recompute more, not less.
//
// A scope with no dependencies is the special case and it is eligible: its output can never change,
// so there is nothing for a later scope to compare against and nothing lost by fusing.
//
// Otherwise the scope is eligible only if it declares a value of an always-invalidating type. Those
// are the values whose identity is fresh on every evaluation, so their scope really does invalidate
// whenever its inputs do.
func ScopeIsEligibleForMerging(function *Function, scope ScopeId, dependencies *ScopeDependencies,
	typeChecker *shimchecker.Checker) bool {
	if dependencies == nil {
		return false
	}
	if len(dependencies.DependenciesOf(scope)) == 0 {
		return true
	}
	for _, declaration := range dependencies.DeclarationsOf(scope) {
		if IsAlwaysInvalidatingType(function, declaration, typeChecker) {
			return true
		}
	}
	return false
}

// CanMergeScopes reports whether two adjacent scopes always invalidate together.
//
// Upstream's `canMergeScopes`, three conditions in order.
//
// # One: neither scope reassigns
//
// A reassignment makes a scope's output depend on control flow rather than on its dependencies
// alone, so two such scopes cannot be shown to invalidate together. Worth knowing before trusting
// corpus coverage here: only 2 of 1,576 corpus scopes carry a reassignment, so this arm is real but
// rare and needs its own fixture rather than the corpus to exercise it.
//
// # Two: identical dependencies
//
// Two scopes reading exactly the same inputs recompute under exactly the same conditions, so one
// scope does the work of both.
//
// # Three: the earlier scope's outputs are the later scope's inputs
//
// The arm with the judgment in it, and upstream's comment explains why it is not simply "the values
// flow from one to the other". A scope's output is not guaranteed to change when its input changes:
// `foo(x)` returning `x < 10` is unchanged as `x` goes from 0 to 1. Fusing on flow alone would make
// the later scope recompute whenever the earlier one's inputs moved, even when its own actual input
// did not.
//
// So the flow must additionally be through values whose type guarantees a fresh identity every
// evaluation, which is what `IsAlwaysInvalidatingType` answers. Upstream also requires those
// dependencies be rooted -- `path.length === 0` -- because a property read off an invalidating
// value is not itself guaranteed to change.
func CanMergeScopes(function *Function, current, next ScopeId, dependencies *ScopeDependencies,
	temporaries map[static_single_assignment.DeclarationId]static_single_assignment.DeclarationId, typeChecker *shimchecker.Checker) bool {
	if dependencies == nil {
		return false
	}
	if len(dependencies.ReassignmentsOf(current)) != 0 ||
		len(dependencies.ReassignmentsOf(next)) != 0 {
		return false
	}

	nextDependencies := dependencies.DependenciesOf(next)
	if AreEqualDependencies(dependencies.DependenciesOf(current), nextDependencies) {
		return true
	}

	// The earlier scope's declarations, read as if they were a dependency set, so the comparison
	// against the later scope's inputs is the same one condition two performs. Upstream builds the
	// same synthetic set inline, with `reactive: true` and an empty path.
	asDependencies := make([]ReactiveScopeDependency, 0, len(dependencies.DeclarationsOf(current)))
	for _, declaration := range dependencies.DeclarationsOf(current) {
		asDependencies = append(asDependencies, ReactiveScopeDependency{
			Identifier: declaration,
			Reactive:   true,
		})
	}
	if AreEqualDependencies(asDependencies, nextDependencies) {
		return true
	}

	if len(nextDependencies) == 0 {
		return false
	}
	for _, dependency := range nextDependencies {
		if len(dependency.Path) != 0 {
			return false
		}
		if !IsAlwaysInvalidatingType(function, dependency.Identifier, typeChecker) {
			return false
		}
		if !declaredByOrAliasedFrom(function, dependency.Identifier,
			dependencies.DeclarationsOf(current), temporaries) {
			return false
		}
	}
	return true
}

// declaredByOrAliasedFrom reports whether a dependency comes from one of the given declarations.
//
// Either directly, or through the temporaries map the driver threads along -- upstream resolves a
// `StoreLocal` chain so that a value copied into a temporary is still recognised as the earlier
// scope's output. Without that indirection an ordinary `const b = a` between two scopes would hide
// the flow and silently prevent a merge upstream performs.
func declaredByOrAliasedFrom(function *Function, dependency static_single_assignment.IdentifierId,
	declarations []static_single_assignment.IdentifierId, temporaries map[static_single_assignment.DeclarationId]static_single_assignment.DeclarationId) bool {
	target := declarationOf(function, dependency)
	aliased, hasAlias := temporaries[target]
	for _, declaration := range declarations {
		candidate := declarationOf(function, declaration)
		if candidate == target {
			return true
		}
		if hasAlias && candidate == aliased {
			return true
		}
	}
	return false
}

// mergeAllowedInstruction reports whether an instruction may sit between two scopes being merged.
//
// Upstream allows a fixed set between a merge candidate and the next scope, and resets on anything
// else. The reasoning is that these are all cheap re-computations with no identity of their own --
// re-running them costs nothing and cannot break a memoization boundary. Anything else might, so
// the merge stops there.
//
// # Upstream's ten kinds against this package's variant set, mapped rather than transcribed
//
// Upstream's list is `BinaryExpression`, `ComputedLoad`, `JSXText`, `LoadGlobal`, `LoadLocal`,
// `Primitive`, `PropertyLoad`, `TemplateLiteral`, `UnaryExpression` and `StoreLocal`. This tree
// splits some of those finer, so the mapping is stated per arm and anything unlisted resets --
// which is the safe direction, since a missed allowance costs a merge that upstream performs and a
// wrongly-added one costs correctness.
//
// `StoreLocal` is upstream's one conditional arm: it is allowed only when the stored value is
// itself a temporary the driver is tracking, so the driver handles it rather than this predicate.
func mergeAllowedInstruction(value InstructionValue) bool {
	switch value.(type) {
	case *BinaryExpression, *UnaryExpression:
		return true
	case *LoadLocal, *LoadGlobal, *LoadContext:
		return true
	case *PropertyLoad, *ComputedLoad:
		return true
	case *Primitive, *TemplateLiteral, *JsxText:
		return true
	default:
		return false
	}
}

// MergeReactiveScopesThatInvalidateTogether fuses adjacent scopes that always recompute together.
//
// Upstream's `mergeReactiveScopesThatInvalidateTogether`. Rewrites the tree in place and returns how
// many merges it performed, so a caller can tell a no-op from an unrun pass -- a distinction this
// package needs because several stages here legitimately find nothing on a given corpus.
//
// # The walk, which is a candidate span rather than a pairwise comparison
//
// Each block is scanned once holding at most one candidate: a scope, the index it started at, the
// index one past the last scope folded into it, and the declarations named since. Three things end
// a candidate -- a terminal, a pruned scope, and an instruction outside the allowlist -- because
// none of those can be re-run freely between two scopes that are about to become one.
//
// A candidate commits only when it spans more than one scope. I expected omitting that guard to
// produce self-references in `Merged` and said so before measuring; it does not. `rewrite` folds
// only the statements strictly after a candidate's own scope, which is an empty range for a span of
// one, so a single-scope run reaches the rewrite and contributes nothing.
//
// The guard is therefore a cost boundary rather than a correctness one, and the cost is real:
// measured over 100 corpus files, 556 scopes are eligible to open a candidate against 85 committed
// merges, so dropping it would send 471 single-scope runs through a block rebuild that changes
// nothing. Stated this way because the earlier framing would have sent someone hunting for a
// correctness bug that is not there.
//
// # Nested blocks first
//
// Upstream traverses nested blocks with the enclosing dependency set before scanning the current
// one. A nested scope with that same set is replaced by its body, separately from adjacent merges.
func MergeReactiveScopesThatInvalidateTogether(tree *ReactiveFunction, function *Function,
	dependencies *ScopeDependencies, typeChecker *shimchecker.Checker) MergeScopesResult {
	if tree == nil || function == nil || dependencies == nil {
		return MergeScopesResult{}
	}
	merger := scopeMerger{
		function:     function,
		dependencies: dependencies,
		typeChecker:  typeChecker,
		usage:        FindLastUsage(tree, function),
		temporaries:  map[static_single_assignment.DeclarationId]static_single_assignment.DeclarationId{},
	}
	tree.Body = merger.mergeBlock(tree.Body)
	return MergeScopesResult{
		NestedScopesRemoved:   merger.nestedScopesRemoved,
		Merges:                merger.merges,
		DeclarationPruneCalls: merger.declarationPruneCalls,
		DeclarationsPruned:    merger.declarationsPruned,
	}
}

// MergeScopesResult reports what the merge did, for measurement and for one invariant.
//
// `DeclarationPruneCalls` exists because `DeclarationsPruned` is zero on our corpus and a zero has
// two readings: the helper is correct and no declaration is prunable, or the helper stopped being
// called. The call count separates them, and a test asserts it equals `Merges` -- upstream calls
// `updateScopeDeclarations` exactly once per merge, so any refactor that drops the call or moves it
// behind a guard breaks that equality rather than passing silently.
type MergeScopesResult struct {
	NestedScopesRemoved int
	// Merges is how many scopes were absorbed into a survivor.
	Merges int
	// DeclarationPruneCalls is how many times the declaration pruning ran, which must equal Merges.
	DeclarationPruneCalls int
	// DeclarationsPruned is how many declarations it dropped. Zero on our corpus; see the type
	// comment for why the call count is reported beside it.
	DeclarationsPruned int
}

type scopeMerger struct {
	parentScope         *ScopeId
	nestedScopesRemoved int
	function            *Function
	dependencies        *ScopeDependencies
	typeChecker         *shimchecker.Checker
	usage               *LastUsage
	// temporaries resolves a StoreLocal chain, so a value copied into a temporary is still
	// recognised as the earlier scope's output. See `declaredByOrAliasedFrom`.
	temporaries map[static_single_assignment.DeclarationId]static_single_assignment.DeclarationId
	merges      int
	// declarationsPruned counts declarations dropped because the widened range left them dead.
	declarationsPruned int
	// declarationPruneCalls counts how many times the pruning ran, which is the control on the
	// zero above.
	declarationPruneCalls int
}

// mergeCandidate is a run of statements that may collapse into one scope.
type mergeCandidate struct {
	scope *ReactiveScopeBlock
	// from is the index of the candidate's own scope statement; to is one past the last folded in.
	from int
	to   int
	// lvalues are the declarations named since the candidate opened, which must all be last used
	// inside the scope being folded in or the merge is declined.
	lvalues []static_single_assignment.DeclarationId
}

func (m *scopeMerger) mergeBlock(block ReactiveBlock) ReactiveBlock {
	// Nested blocks first, so an inner merge is settled before the outer scan reads the statement
	// holding it.
	rebuilt := make(ReactiveBlock, 0, len(block))
	for _, statement := range block {
		switch shape := statement.(type) {
		case *ReactiveScopeBlock:
			parent := m.parentScope
			if !shape.Pruned {
				scope := shape.Scope
				m.parentScope = &scope
			}
			shape.Instructions = m.mergeBlock(shape.Instructions)
			m.parentScope = parent
			if !shape.Pruned && parent != nil && AreEqualDependencies(m.dependencies.DependenciesOf(*parent), m.dependencies.DependenciesOf(shape.Scope)) {
				m.nestedScopesRemoved++
				rebuilt = append(rebuilt, shape.Instructions...)
				continue
			}
		case *ReactiveTerminalStatement:
			m.mergeTerminalBlocks(shape)
		}
		rebuilt = append(rebuilt, statement)
	}
	block = rebuilt

	var candidate *mergeCandidate
	var committed []mergeCandidate

	// commit records a candidate if it spans more than one scope, then clears it.
	commit := func() {
		if candidate != nil && candidate.to > candidate.from+1 {
			committed = append(committed, *candidate)
		}
		candidate = nil
	}

	for index, statement := range block {
		switch shape := statement.(type) {
		case *ReactiveTerminalStatement:
			// Upstream does not merge across terminals.
			commit()

		case *ReactiveInstructionStatement:
			if shape.Instruction == nil {
				commit()
				continue
			}
			plain, isPlain := shape.Instruction.Value.(*ReactiveInstructionValue)
			if !isPlain || plain.Value == nil {
				commit()
				continue
			}
			if store, isStore := plain.Value.(*StoreLocal); isStore {
				if candidate == nil {
					continue
				}
				// Upstream allows a StoreLocal only while tracking it, recording the alias so a
				// later dependency on the stored name still resolves to the earlier scope's output.
				target := declarationOf(m.function, store.LValue.Identifier)
				source := declarationOf(m.function, store.Value.Identifier)
				if aliased, found := m.temporaries[source]; found {
					source = aliased
				}
				m.temporaries[target] = source
				candidate.lvalues = append(candidate.lvalues, target)
				continue
			}
			if !mergeAllowedInstruction(plain.Value) {
				commit()
				continue
			}
			if candidate != nil && shape.Instruction.LValue != nil {
				candidate.lvalues = append(candidate.lvalues,
					declarationOf(m.function, shape.Instruction.LValue.Identifier))
			}

		case *ReactiveScopeBlock:
			if shape.Pruned {
				// Upstream does not merge across pruned scopes.
				commit()
				continue
			}
			if candidate != nil &&
				CanMergeScopes(m.function, candidate.scope.Scope, shape.Scope, m.dependencies,
					m.temporaries, m.typeChecker) &&
				AreLValuesLastUsedByScope(shape.Range.End, candidate.lvalues, m.usage) {
				// Widen the survivor to cover the scope it absorbs, matching upstream's range
				// update, and reopen the lvalue set: everything named before this point is now
				// inside the merged scope.
				if shape.Range.End > candidate.scope.Range.End {
					candidate.scope.Range.End = shape.Range.End
				}
				candidate.to = index + 1
				candidate.lvalues = candidate.lvalues[:0]
				// Upstream's `updateScopeDeclarations`, called at exactly this point: the survivor's
				// range has just widened, so a value it declares but which is last read inside the
				// absorbed range is no longer an output of the merged scope.
				//
				// This was left unported when the merge landed, on the reasoning that a wider
				// declaration set is conservative. That holds only where declarations are read to
				// permit; `pruneAlwaysInvalidatingScopes` reads them to propagate, which prunes
				// further scopes downstream, so the debt came due when that pass was written.
				m.declarationPruneCalls++
				m.declarationsPruned += m.dependencies.PruneDeclarationsLastUsedBefore(
					candidate.scope.Scope, candidate.scope.Range.End, m.usage, m.function)
				if !ScopeIsEligibleForMerging(m.function, shape.Scope, m.dependencies,
					m.typeChecker) {
					commit()
				}
				continue
			}
			commit()
			if ScopeIsEligibleForMerging(m.function, shape.Scope, m.dependencies, m.typeChecker) {
				candidate = &mergeCandidate{scope: shape, from: index, to: index + 1}
			}
		}
	}
	commit()

	if len(committed) == 0 {
		return block
	}
	return m.rewrite(block, committed)
}

// rewrite splices each committed candidate's statements into its surviving scope.
func (m *scopeMerger) rewrite(block ReactiveBlock, committed []mergeCandidate) ReactiveBlock {
	rebuilt := make(ReactiveBlock, 0, len(block))
	index := 0
	for _, entry := range committed {
		if index < entry.from {
			rebuilt = append(rebuilt, block[index:entry.from]...)
			index = entry.from
		}
		survivor, isScope := block[entry.from].(*ReactiveScopeBlock)
		if !isScope {
			// Upstream raises an invariant here. A linter cannot, so the run is abandoned and the
			// statements are copied through unchanged rather than being silently dropped.
			continue
		}
		rebuilt = append(rebuilt, survivor)
		index++
		for index < entry.to {
			switch shape := block[index].(type) {
			case *ReactiveScopeBlock:
				survivor.Instructions = append(survivor.Instructions, shape.Instructions...)
				survivor.Merged = append(survivor.Merged, shape.Scope)
				m.merges++
			default:
				// Never taken on the corpus: 132 scopes fold and zero non-scope statements come
				// with them, so a mutation deleting this line changes no answer. It is kept because
				// the allowlist explicitly permits instructions between two merging scopes, and a
				// pass that silently dropped them the day one appeared would lose real work. The
				// zero is a fact about this corpus, not about the shape.
				survivor.Instructions = append(survivor.Instructions, block[index])
			}
			index++
		}
	}
	rebuilt = append(rebuilt, block[index:]...)
	return rebuilt
}

// mergeTerminalBlocks descends into every block a terminal contains.
//
// Each arm assigns the result back, for the reason `transformTerminalBlocks` records: a missing
// assignment leaves the terminal holding the pre-merge slice, and that is invisible to any test
// counting what the walk reached.
func (m *scopeMerger) mergeTerminalBlocks(statement *ReactiveTerminalStatement) {
	switch shape := statement.Terminal.(type) {
	case *ReactiveIf:
		shape.Consequent = m.mergeBlock(shape.Consequent)
		if shape.Alternate != nil {
			rewritten := m.mergeBlock(*shape.Alternate)
			shape.Alternate = &rewritten
		}
	case *ReactiveSwitch:
		for index := range shape.Cases {
			if shape.Cases[index].Block != nil {
				rewritten := m.mergeBlock(*shape.Cases[index].Block)
				shape.Cases[index].Block = &rewritten
			}
		}
	case *ReactiveFor:
		shape.Loop = m.mergeBlock(shape.Loop)
	case *ReactiveForOf:
		shape.Loop = m.mergeBlock(shape.Loop)
	case *ReactiveForIn:
		shape.Loop = m.mergeBlock(shape.Loop)
	case *ReactiveWhile:
		shape.Loop = m.mergeBlock(shape.Loop)
	case *ReactiveDoWhile:
		shape.Loop = m.mergeBlock(shape.Loop)
	case *ReactiveLabelTerminal:
		shape.Block = m.mergeBlock(shape.Block)
	case *ReactiveTry:
		shape.Block = m.mergeBlock(shape.Block)
		shape.Handler = m.mergeBlock(shape.Handler)
	}
}
