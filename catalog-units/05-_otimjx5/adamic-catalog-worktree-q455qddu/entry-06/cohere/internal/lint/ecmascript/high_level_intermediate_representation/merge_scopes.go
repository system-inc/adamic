// Merging overlapping reactive scopes: making scope ranges nest, so a scope can become a block.
//
// This is React's `mergeOverlappingReactiveScopesHIR` (development bundle line 32468), run over this
// graph. oxc transcribes it at
// `oxc_react_compiler/src/react_compiler_reactive_scopes/merge_overlapping_reactive_scopes_hir.rs`;
// where the two disagree React wins, and each disagreement is recorded at the line that resolves it.
//
// # Why this pass exists, which is a precondition rather than an optimisation
//
// `AssignReactiveScopes` produces one scope per equivalence class with a range that is the hull of
// its members. Nothing in that construction makes two scopes NEST. Two classes can easily produce
// ranges like [5,12) and [9,20): they overlap, and neither contains the other.
//
// That is fatal downstream. A reactive scope eventually becomes a BLOCK in the output -- a
// `<scope>` terminal wrapping a range of instructions -- and blocks nest or they are disjoint.
// There is no way to emit two overlapping-but-not-nested spans as nested blocks. So upstream calls
// `assertValidBlockNesting` on the line immediately before `buildReactiveScopeTerminalsHIR`
// (lines 50213 and 50214), and that assertion is a hard `CompilerError.invariant` rather than a
// filter: `recursivelyTraverseItems` raises "Invalid nesting in program blocks or scopes" and stops.
//
// This pass is what makes that assertion pass, by unioning any two scopes that would violate it and
// widening the survivor to cover both.
//
// # The algorithm is a STACK DISCIPLINE, not a pairwise overlap fixpoint
//
// This is the single most important thing to know about this pass, and the natural mental model is
// wrong. "Union any two scopes whose ranges overlap without nesting, widen to the hull, repeat until
// nothing changes" is a plausible description, it reaches zero violations, and it is NOT what
// upstream does. Measured over 400 corpus files, both spellings run side by side:
//
//	                              unions   scopes after   non-nested scope pairs after
//	React's stack discipline          88          3,218                              0
//	pairwise overlap to fixpoint      41          3,265                              0
//
// Both reach zero, so the acceptance test alone cannot tell them apart, and a port written from the
// pairwise description would look correct while merging half as much. The difference is real: 47
// scopes that upstream merges stay separate under the pairwise rule.
//
// What upstream actually does is sweep the function in evaluation order maintaining a STACK of
// active scopes, `activeScopes`, ordered so that the scope ending LAST sits deepest. Three events
// drive it, and each one is a union:
//
//	a scope ENDS while it is not the top of the stack
//	    everything above it started later and ends later, so it and all of them are unioned
//	two scopes START at the same instruction with the SAME end
//	    they are unioned, because neither can nest inside the other
//	a place is USED while its own scope is not the top of the stack
//	    that scope and everything above it are unioned
//
// The third has no counterpart in the pairwise description at all, and it is where most of the extra
// unions come from. It fires on a value whose scope is still open being read while some other scope
// opened inside it and has not closed -- a genuine interleaving that a range comparison cannot see,
// because the two ranges may nest perfectly and still interleave through a use.
//
// # DIVERGENCE FROM React: the Primitive gate cannot fire, because the IR carries no type
//
// `getOverlappingReactiveScopes` skips an operand of a `FunctionExpression` or `ObjectMethod` whose
// `place.identifier.type.kind === 'Primitive'` (bundle line 32570). This IR carries no type, exactly
// as `disjoint.go` records for `mayAllocate`, so there is no primitive predicate to consult and the
// gate is unreachable rather than omitted.
//
// The direction of the error is stated rather than guessed. Not skipping means visiting operands
// upstream skips, and visiting an operand can only ever ADD a union, so this errs toward merging
// more than upstream rather than less.
//
// The size of that error is measured at its UPPER BOUND rather than estimated, because the gate's
// real condition cannot be evaluated here at all. Skipping EVERY use-operand of every
// `FunctionExpression` -- which is strictly more than the Primitive gate would skip, since the gate
// skips only the primitive ones -- changes the union count from 118 to 115. The real divergence is
// therefore at most three unions and may still be zero; without types, the probe cannot distinguish
// primitive captures from the non-primitive captures upstream still visits.
//
// A zero from a probe that never fires would be worthless, so the control is recorded beside it:
// 1,701 function-expression use-operands are visited and are candidates for the skip.
// None of the 116,476 identifiers carries a type, which is why the real gate cannot be asked.
// Recorded as `MergeGapPrimitiveOperandSkip` so a consumer can decline on it.
// `TestMergeFunctionOperandSkipUpperBound` pins the bound together with its control.
//
// # DIVERGENCE FROM React: the widening MINIMISES here, and upstream's own guard is why that is safe
//
// `AssignReactiveScopes` uses a guarded adopt-or-minimise for its hull, because zero is the unset
// sentinel and a plain minimum would drag a hull's start to zero. This pass uses a PLAIN min/max,
// because that is what upstream writes (`Math.min(groupScope.range.start, scope.range.start)`,
// bundle line 32474), and it is safe for a structural reason rather than by luck.
//
// `collectScopeInfo` only registers a scope when `scope.range.start !== scope.range.end`
// (bundle line 32488). A scope with an unset range is [0,0), so start equals end, so it is never
// recorded as a start or an end, never enters `activeScopes`, and can never be unioned. The plain
// minimum therefore never sees a zero.
//
// Both halves were checked rather than assumed, and the check is the one Stage 4 got wrong in the
// other direction. Running the arithmetic upstream writes:
//
//	group [4,6), member unset [0,0)  ->  [0,6)   a zero WOULD poison the start
//	group unset [0,0), member [4,6)  ->  [0,6)   and in this direction too
//	group [4,6), member [2,9)        ->  [2,9)   ordinary widening
//
// So the guard is load-bearing and not decorative. Measured over 400 corpus files: 0 scopes have
// `start == end`, so the branch declines nothing today, and 0 merged groups end up with a zero
// start. That zero is a fact about this corpus rather than about the algorithm, which is why the
// guard is reproduced rather than simplified away. `TestMergeSkipsDegenerateScopes` pins it.
//
// # What this pass does NOT close, measured, because the next stage is blocked on the difference
//
// `assertValidBlockNesting` does not check scopes against each other. It builds items from the
// scopes AND from every block that has a fallthrough -- a `ProgramBlockSubtree` spanning from the
// terminal to the first instruction of its fallthrough -- and asserts nesting over the COMBINED
// list (bundle lines 20954-20969). Over 400 corpus files:
//
//	                                 before merge   after merge
//	scope against scope                        40             0
//	the full assertion                        226           166
//
// So this pass closes its own half completely and leaves 166. Attributed by kind:
//
//	Scope inside ProgramBlockSubtree           78
//	ProgramBlockSubtree inside Scope           87
//	ProgramBlockSubtree against itself          1
//
// The first two are what `alignReactiveScopesToBlockScopes` exists to fix, and it runs immediately
// before this pass upstream (line 50201) for exactly that reason: it widens each scope out to the
// block boundaries it straddles, so a scope and a block subtree can no longer cross. It is now
// ported, in `align_scopes.go`, and running it IN FRONT of this pass takes the full assertion to 1.
// Running it behind instead reaches only 30, because widening creates overlaps that this pass is no
// longer there to close. `MergeGapBlockScopeAlignment` records what that means for a caller.
//
// The third is neither pass's to fix and the control says so: measured with the scope items removed
// entirely, the same violation is still there, so it is a property of this lowering's fallthrough
// structure rather than anything a scope pass produces. One occurrence over 400 files, recorded here
// so whoever ports alignment does not spend the afternoon hunting a scope bug.
//
// # Termination is a single sweep, and the bound is exact rather than empirical
//
// There is no fixpoint here, which is worth stating because the pairwise description this file warns
// about DOES iterate and a reader carrying that model will look for a loop bound. One pass over the
// blocks, in the order `Function.Blocks` already holds, visiting each instruction once and each
// terminal once. Both `scopeStarts` and `scopeEnds` are consumed by popping, never appended to, so
// the two inner `for` bodies run at most once per distinct instruction position over the whole
// sweep. `activeScopes` only grows when a start pops and only shrinks when an end pops, so its total
// churn is bounded by the same quantity.
//
// The work is therefore linear in instructions plus places, with a `sort` per pop over the scopes
// sharing one position. Nothing revisits anything, so there is no bound to prove empirically -- the
// structure forbids iteration rather than converging. `TestMergeIsASingleSweep` pins the visit count.
//
// The stack discipline itself depends on one property of this graph that upstream gets for free from
// its own construction: evaluation order must be monotone along the block sequence, or a scope could
// end before the sweep reaches it. Measured over all 685 corpus functions: 0 non-monotone steps.
// `TestMergeAssumesMonotoneEvaluationOrder` pins it, because it is an assumption this file makes
// about a table another file produces.
//
// # Determinism, which two Go maps would otherwise destroy
//
// Upstream's `DisjointSet` iterates a `Map` in insertion order and `collectScopeInfo` builds its
// tables from an insertion-ordered `Map` of `Set`s, so upstream is deterministic without trying.
// Both structures here are Go maps. `scopeSet` below therefore records insertion order explicitly,
// and the per-position scope lists are sorted by id before the range sorts run, so a tie in the
// range comparator resolves the same way on every run. Without that, two scopes ending at the same
// instruction would union in a random order and hand different representatives to different runs --
// the shape that makes a cache non-reproducible without ever producing a wrong answer.
// `TestMergeIsDeterministic` runs the pass twice over the corpus and compares.
package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
	"sort"
)

// MergeGap names a merge rule this pass cannot apply, for a caller that needs to know.
//
// Modelled on `ScopeGap`, `RangeGap`, `ReactiveGap` and `DisjointGap` deliberately, and for the same
// reason: a gap answerable through the API is one a consumer can decline on, while a gap living only
// in a comment is one the next reader inherits by accident.
type MergeGap uint8

const (
	// MergeGapPrimitiveOperandSkip is upstream's skip of a primitive operand of a function
	// expression or object method.
	//
	// This IR carries no type, so there is no primitive predicate to consult. Not skipping visits operands upstream skips, and a visit can only add a
	// union, so this errs toward merging MORE than upstream rather than less. Skipping every
	// function-expression operand, a strict superset of upstream's primitive-only skip, removes
	// three of 118 unions on the measured corpus. The real divergence is therefore at most three.
	MergeGapPrimitiveOperandSkip MergeGap = iota

	// MergeGapBlockScopeAlignment is `alignReactiveScopesToBlockScopes`, which upstream runs
	// immediately before this pass.
	//
	// It is now ported, as `AlignReactiveScopesToBlockScopes`, so this gap is about ORDER rather
	// than absence. Run this pass ALONE and a scope can still cross a program block boundary:
	// measured, it takes scope-against-scope violations to 0 and leaves 163 scope-against-block
	// ones. Run alignment in front of it -- `AlignThenMergeReactiveScopes` is the composed spelling
	// -- and the full assertion reaches 1, which is a block-against-block residue neither pass can
	// reach. A consumer that needs the full assertion to hold must call the composed spelling; this
	// pass on its own is still gapped.
	MergeGapBlockScopeAlignment
)

// MergeGaps are the merge rules MergeOverlappingReactiveScopes does not apply. See MergeGap.
//
// Returned as a value rather than documented alone so a test can assert on it, which makes closing a
// gap a visible event rather than a silent improvement.
func MergeGaps() []MergeGap {
	return []MergeGap{MergeGapPrimitiveOperandSkip, MergeGapBlockScopeAlignment}
}

// MergedScopes is what this pass produces: the surviving scopes and where the merged ones went.
//
// A scope that was not merged maps to itself and keeps its range. A scope that was merged maps to
// its group's representative and is no longer in `Ids`. This is a derived view rather than a
// rewrite of the input table, for the reason `MemberRanges` records: keeping both available costs
// one map and lets a consumer ask which of the two it wants.
//
// The zero value is an empty result and is ready to read.
type MergedScopes struct {
	ranges  map[ScopeId]mutation_aliasing.MutableRange
	group   map[ScopeId]ScopeId
	members map[ScopeId][]static_single_assignment.IdentifierId
	order   []ScopeId
	unions  int
}

// GroupOf returns the scope a scope was merged into, or the scope itself when it was not merged.
//
// Upstream's `joinedScopes.find(originalScope)` with its `?? originalScope` fallback folded in, so
// a caller does not re-derive it. A scope this pass never saw answers itself, which is the answer
// that lets a caller key a map on the result without first asking whether it merged.
func (m *MergedScopes) GroupOf(scope ScopeId) ScopeId {
	if m == nil || m.group == nil {
		return scope
	}
	if group, ok := m.group[scope]; ok {
		return group
	}
	return scope
}

// RangeOf returns a surviving scope's widened range, or the unset range for one that did not survive.
//
// Ask `GroupOf` first when holding a pre-merge id: a merged scope's range lives on its group.
func (m *MergedScopes) RangeOf(scope ScopeId) mutation_aliasing.MutableRange {
	if m == nil || m.ranges == nil {
		return mutation_aliasing.MutableRange{}
	}
	return m.ranges[scope]
}

// MembersOf returns every value belonging to a surviving scope, including the values that arrived
// through a merge, sorted.
//
// The slice is the table's own and must not be modified by a caller.
func (m *MergedScopes) MembersOf(scope ScopeId) []static_single_assignment.IdentifierId {
	if m == nil || m.members == nil {
		return nil
	}
	return m.members[scope]
}

// Ids returns the surviving scopes, in the input table's own order.
//
// Merged-away scopes are absent. Returned as a slice rather than by ranging a map so a caller gets
// the same sequence every run.
func (m *MergedScopes) Ids() []ScopeId {
	if m == nil {
		return nil
	}
	return m.order
}

// Len reports how many scopes survived, for measurement.
func (m *MergedScopes) Len() int {
	if m == nil {
		return 0
	}
	return len(m.order)
}

// Unions reports how many scopes were merged away, for measurement.
//
// This is `Len` subtracted from the input table's length, exposed directly because it is the number
// the counterfactual in this file's comment is stated in and a test comparing the two spellings
// needs it without recomputing.
func (m *MergedScopes) Unions() int {
	if m == nil {
		return 0
	}
	return m.unions
}

// scopeSet is a union-find over ScopeId, mirroring upstream's `DisjointSet` including its
// insertion-ordered iteration.
//
// `DisjointSet` in this package is keyed on IdentifierId and cannot be reused: this pass unions
// SCOPES rather than values, and the two id spaces are unrelated. Generics were the other option and
// were declined because `DisjointSet`'s doc comment is written entirely about values and mutation
// aliasing, and making it generic would leave that comment describing one of two instantiations.
//
// `order` exists because Go map iteration is randomised where upstream's `Map` is insertion-ordered,
// and upstream's `forEach` walks that order to assign group representatives. Without it the
// representative of a group would vary between runs.
type scopeSet struct {
	parent map[ScopeId]ScopeId
	order  []ScopeId
}

// union merges every scope in items into one class, transcribing upstream's `union`.
//
// The first item becomes the representative when none of the items is already in the set. An empty
// list is a no-op here; upstream raises an invariant on it, and the difference is deliberate for the
// reason `DisjointSet.Union` records: the call sites build the list conditionally.
func (d *scopeSet) union(items []ScopeId) {
	if len(items) == 0 {
		return
	}
	if d.parent == nil {
		d.parent = map[ScopeId]ScopeId{}
	}
	first := items[0]
	root, ok := d.find(first)
	if !ok {
		root = first
		d.parent[first] = first
		d.order = append(d.order, first)
	}
	for _, item := range items[1:] {
		itemParent, present := d.parent[item]
		if !present {
			d.parent[item] = root
			d.order = append(d.order, item)
			continue
		}
		if itemParent == root {
			continue
		}
		// Walk this item's chain to the root, repointing every link, exactly as upstream does. It
		// rewrites the whole path rather than only the item, which keeps the structure flat when two
		// established classes merge.
		current := item
		for itemParent != root {
			d.parent[current] = root
			current = itemParent
			itemParent = d.parent[current]
		}
	}
}

// find returns the representative of a scope's class, and whether the scope is in the set at all.
//
// Path compression on the way out, which is what keeps the recursion depth safe. The second return
// is upstream's `null`, kept separate from the id because zero is the absent ScopeId and conflating
// the two would make an absent scope read as a member of class zero.
func (d *scopeSet) find(item ScopeId) (ScopeId, bool) {
	if d.parent == nil {
		return 0, false
	}
	parent, ok := d.parent[item]
	if !ok {
		return 0, false
	}
	if parent == item {
		return item, true
	}
	root, ok := d.find(parent)
	if !ok {
		return 0, false
	}
	d.parent[item] = root
	return root, true
}

// scopesAtPosition is the set of scopes starting or ending at one evaluation position.
//
// Upstream keys a `Map` on the instruction id and holds a `Set` of scopes. The list here is sorted
// by scope id so that the range sorts below are stable under ties, which upstream gets from
// insertion order.
type scopesAtPosition struct {
	position static_single_assignment.EvaluationOrder
	scopes   []ScopeId
}

// mergeSweepState is the sweep's working state, matching upstream's `{joined, activeScopes}`.
type mergeSweepState struct {
	joined       scopeSet
	activeScopes []ScopeId
	starts       []scopesAtPosition
	ends         []scopesAtPosition
	rangeOf      func(ScopeId) mutation_aliasing.MutableRange

	// visits counts instruction positions handled, for the termination test only.
	visits int
}

// MergeOverlappingReactiveScopes unions the scopes of function whose ranges would not nest.
//
// Requires scopes: call `AssignReactiveScopes` first, or take its result from a caller that has one.
// A nil or empty scope table produces an empty result rather than an error, which is the ordinary
// answer for a function with no entangled values.
//
// The member ranges this reads are the REWRITTEN ones -- `ReactiveScopes.MemberRanges` -- because
// upstream reads `place.identifier.mutableRange` at a point in its pipeline where
// `inferReactiveScopeVariables` has already written the scope hull back onto every member. Reading
// the pre-entanglement ranges instead would make `isMutable` answer a different question and is the
// single easiest way to get this pass subtly wrong.
func MergeOverlappingReactiveScopes(function *Function, scopes *ReactiveScopes) *MergedScopes {
	result := &MergedScopes{}
	if function == nil || scopes == nil || scopes.Len() == 0 {
		return result
	}

	state := &mergeSweepState{rangeOf: scopes.RangeOf}
	state.starts, state.ends = collectScopeInfo(function, scopes)

	// Upstream reads `place.identifier.mutableRange`, which by this point in its pipeline is the
	// scope hull rather than the value's own range. See the doc comment above.
	memberRanges := scopes.MemberRanges()

	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			state.visitInstructionId(instruction.Order)
			// Upstream skips a primitive operand of a FunctionExpression or ObjectMethod here.
			// The IR carries no type, so that gate cannot be evaluated; see
			// `MergeGapPrimitiveOperandSkip`.
			EachInstructionPlace(instruction, func(place Place, role PlaceRole) {
				state.visitPlace(instruction.Order, place, scopes, memberRanges)
			})
		}
		order := TerminalOrder(block.Terminal)
		state.visitInstructionId(order)
		EachTerminalPlace(block.Terminal, func(place Place, role PlaceRole) {
			state.visitPlace(order, place, scopes, memberRanges)
		})
	}

	return state.apply(scopes)
}

// collectScopeInfo builds the start and end tables, transcribing upstream's `collectScopeInfo`.
//
// Both are returned sorted DESCENDING by position, because upstream reads them with `.at(-1)` and
// pops -- so the last element is the next position the sweep will reach. Written the same way rather
// than sorted ascending and read from the front, so that a reader diffing against the bundle finds
// the same shape.
//
// A scope is registered only when its range's start differs from its end. That is upstream's guard
// and it is what keeps an unset [0,0) scope out of the sweep entirely, which is what makes the plain
// min/max widening safe. See the doc comment above.
func collectScopeInfo(function *Function, scopes *ReactiveScopes) (starts, ends []scopesAtPosition) {
	startSets := map[static_single_assignment.EvaluationOrder]map[ScopeId]bool{}
	endSets := map[static_single_assignment.EvaluationOrder]map[ScopeId]bool{}

	record := func(place Place) {
		scope := scopes.ScopeOf(place.Identifier)
		if scope == 0 {
			return
		}
		scopeRange := scopes.RangeOf(scope)
		if scopeRange.Start == scopeRange.End {
			return
		}
		if startSets[scopeRange.Start] == nil {
			startSets[scopeRange.Start] = map[ScopeId]bool{}
		}
		startSets[scopeRange.Start][scope] = true
		if endSets[scopeRange.End] == nil {
			endSets[scopeRange.End] = map[ScopeId]bool{}
		}
		endSets[scopeRange.End][scope] = true
	}

	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, instructionId := range block.Instructions {
			if instruction := function.Instructions[instructionId]; instruction != nil {
				EachInstructionPlace(instruction, func(place Place, role PlaceRole) { record(place) })
			}
		}
		// Upstream records terminal operands here too (bundle line 32507). A mutation deleting this
		// line SURVIVES, and the reason is measured rather than assumed: over 400 corpus files, 0 of
		// 3,306 scopes are registered ONLY through a terminal place -- every one of them also
		// appears on an instruction operand or lvalue, which registers the same start and end.
		//
		// So the line is reachable and currently redundant rather than unreachable. It is kept
		// because it is upstream's and because the redundancy is a property of what this lowering
		// happens to emit rather than of the algorithm: a terminal whose operand is the sole
		// appearance of a value is well-formed, and the sweep would silently stop seeing that scope.
		// The verdict EXPIRES if lowering gains a terminal form that can hold a value no instruction
		// touches. `TestMergeRegistersScopesFromTerminals` pins the zero so the expiry is visible.
		EachTerminalPlace(block.Terminal, func(place Place, role PlaceRole) { record(place) })
	}

	return descendingByPosition(startSets), descendingByPosition(endSets)
}

// descendingByPosition flattens a position-keyed set table into a list sorted descending, with each
// position's scopes sorted by id.
//
// The inner sort is not upstream's -- upstream's `Set` is insertion-ordered and needs no sort. It is
// here so that the range comparators in the sweep, which are not total orders when two scopes share
// a start or an end, resolve ties the same way on every run.
func descendingByPosition(sets map[static_single_assignment.EvaluationOrder]map[ScopeId]bool) []scopesAtPosition {
	out := make([]scopesAtPosition, 0, len(sets))
	for position, set := range sets {
		list := make([]ScopeId, 0, len(set))
		for scope := range set {
			list = append(list, scope)
		}
		sort.Slice(list, func(i, j int) bool { return list[i] < list[j] })
		out = append(out, scopesAtPosition{position: position, scopes: list})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].position > out[j].position })
	return out
}

// visitInstructionId advances the sweep to one evaluation position, transcribing upstream's
// `visitInstructionId`.
//
// Ends are processed before starts, which is upstream's order and is load-bearing: a scope ending at
// exactly the position another starts must leave the stack before the new one is pushed, or the two
// would appear to be simultaneously active and union spuriously.
func (s *mergeSweepState) visitInstructionId(id static_single_assignment.EvaluationOrder) {
	s.visits++

	if len(s.ends) > 0 && s.ends[len(s.ends)-1].position <= id {
		top := s.ends[len(s.ends)-1]
		s.ends = s.ends[:len(s.ends)-1]

		// Sorted by start descending: the scope that started LATEST is closed first, so that when
		// several scopes end together the unions below see the stack in the order upstream does.
		closing := append([]ScopeId(nil), top.scopes...)
		sortByStartDescending(closing, s.rangeOf)

		for _, scope := range closing {
			index := indexOfScope(s.activeScopes, scope)
			if index == -1 {
				continue
			}
			// A scope ending while something sits above it on the stack: everything above started
			// later and has not ended, so it interleaves with this one rather than nesting inside
			// it. Union them all.
			//
			// The `index != len-1` test is upstream's and it is a COST OPTIMISATION rather than a
			// filter. When the scope is already on top the suffix is empty, so the call reduces to
			// `union([scope])`, which inserts the scope as its own representative and nothing else;
			// `apply` then skips it because `group == scope`. Both a neutralised and a
			// forced-always-on mutation of this guard SURVIVE, and that is the correct verdict
			// rather than a fixture gap.
			//
			// Proven rather than argued, because "equivalent" is the verdict most likely to be
			// wrong: the whole corpus was run with the guard removed entirely and compared group by
			// group and range by range against this spelling. 685 functions, 88 unions, 0
			// differences. Upstream's spelling is kept because it is upstream's and because the
			// no-op insertion it avoids would otherwise grow `joined.order` with entries that mean
			// nothing.
			if index != len(s.activeScopes)-1 {
				s.joined.union(append([]ScopeId{scope}, s.activeScopes[index+1:]...))
			}
			s.activeScopes = append(s.activeScopes[:index], s.activeScopes[index+1:]...)
		}
	}

	if len(s.starts) > 0 && s.starts[len(s.starts)-1].position <= id {
		top := s.starts[len(s.starts)-1]
		s.starts = s.starts[:len(s.starts)-1]

		// Sorted by end descending, so the scope ending LAST goes deepest on the stack. That is the
		// ordering every union above depends on: "everything above me" is then exactly "everything
		// that ends before me".
		opening := append([]ScopeId(nil), top.scopes...)
		sortByEndDescending(opening, s.rangeOf)
		s.activeScopes = append(s.activeScopes, opening...)

		// Two scopes starting together AND ending together cannot nest either way, so they are
		// unioned outright. Adjacent comparison suffices because the slice is sorted by end.
		//
		// A mutation inverting this comparison SURVIVES, and the reason is SUBSUMPTION rather than a
		// fixture gap. Two scopes sharing a start and an end also share the end, so the end branch
		// above reaches them at that position with one of them not on top of the stack and unions
		// them there. Measured on the synthetic case that exercises it: with this union removed
		// entirely, `sameStartSameEnd` still collapses to one scope with one union.
		//
		// The corpus half of that verdict has now EXPIRED, exactly as predicted. It fired 0 times
		// when it was written, because no two scopes shared both endpoints. `AlignReactiveScopesToBlockScopes`
		// landed as the producer this paragraph named and runs in front of this pass, and over the
		// aligned table 251 adjacent pairs share both a start and an end. So the branch is now
		// REACHABLE on the corpus and only the subsumption argument still explains the survival --
		// which is the stronger of the two reasons anyway, and it rests on the end branch's
		// ordering rather than on what the corpus happens to contain.
		//
		// Kept rather than deleted because it is upstream's and because the subsumption argument
		// rests on the end branch's ordering, which the same alignment pass could also change.
		for i := 1; i < len(opening); i++ {
			previous, current := opening[i-1], opening[i]
			if s.rangeOf(previous).End == s.rangeOf(current).End {
				s.joined.union([]ScopeId{previous, current})
			}
		}
	}
}

// visitPlace unions a place's scope with everything above it, transcribing upstream's `visitPlace`.
//
// This is the event with no counterpart in a pairwise-overlap model, and it is where most of the
// extra unions come from. A value read while its own scope is open but buried under a scope that
// opened later means the two genuinely interleave, even when their ranges nest perfectly.
func (s *mergeSweepState) visitPlace(id static_single_assignment.EvaluationOrder, place Place, scopes *ReactiveScopes,
	memberRanges *mutation_aliasing.MutableRanges) {
	placeScope := activeScopeOf(id, place, scopes)
	if placeScope == 0 {
		return
	}
	// Upstream's `isMutable(instr, place)`: the place's own recorded range must contain this
	// position. Reads the post-entanglement ranges, as upstream does at this point. See the doc
	// comment on `MergeOverlappingReactiveScopes`.
	if !memberRanges.Get(place.Identifier).Contains(id) {
		return
	}
	index := indexOfScope(s.activeScopes, placeScope)
	if index != -1 && index != len(s.activeScopes)-1 {
		s.joined.union(append([]ScopeId{placeScope}, s.activeScopes[index+1:]...))
	}
}

// activeScopeOf is upstream's `getPlaceScope`: the place's scope, but only while it is open.
//
// A place whose scope has already closed answers zero, which is upstream's `null`. Half-open, so
// the start is inside and the end is not, matching `MutableRange.Contains`.
func activeScopeOf(id static_single_assignment.EvaluationOrder, place Place, scopes *ReactiveScopes) ScopeId {
	scope := scopes.ScopeOf(place.Identifier)
	if scope == 0 {
		return 0
	}
	if !scopes.RangeOf(scope).Contains(id) {
		return 0
	}
	return scope
}

// indexOfScope returns a scope's position on the active stack, or -1.
//
// A linear scan rather than a map, because upstream uses `indexOf` and because the stack is short:
// it holds only the scopes open at one instant, and the slice-suffix unions above need positional
// order anyway.
func indexOfScope(stack []ScopeId, scope ScopeId) int {
	for index, active := range stack {
		if active == scope {
			return index
		}
	}
	return -1
}

// apply widens each group representative and builds the result, transcribing upstream's apply loop.
//
// The widening is a PLAIN min/max, which is upstream's spelling and is safe because `collectScopeInfo`
// refuses a scope whose start equals its end. See the doc comment above for the measurement.
func (s *mergeSweepState) apply(scopes *ReactiveScopes) *MergedScopes {
	result := &MergedScopes{
		ranges:  map[ScopeId]mutation_aliasing.MutableRange{},
		group:   map[ScopeId]ScopeId{},
		members: map[ScopeId][]static_single_assignment.IdentifierId{},
	}

	for _, scope := range scopes.Ids() {
		result.ranges[scope] = scopes.RangeOf(scope)
		result.group[scope] = scope
	}

	// Upstream's `joinedScopes.forEach`, walking insertion order so the representative is stable.
	for _, scope := range s.joined.order {
		group, ok := s.joined.find(scope)
		if !ok || group == scope {
			continue
		}
		result.group[scope] = group
		result.unions++

		widened := result.ranges[group]
		merged := scopes.RangeOf(scope)
		if merged.Start < widened.Start {
			widened.Start = merged.Start
		}
		if merged.End > widened.End {
			widened.End = merged.End
		}
		result.ranges[group] = widened
	}

	// A chain can put a scope's group behind another merge, so resolve to the final representative
	// before building membership. `find` has already compressed every path, so this is a lookup.
	for _, scope := range scopes.Ids() {
		if group, ok := s.joined.find(scope); ok {
			result.group[scope] = group
		}
	}

	for _, scope := range scopes.Ids() {
		group := result.group[scope]
		if group == scope {
			result.order = append(result.order, scope)
		} else {
			delete(result.ranges, scope)
		}
		result.members[group] = append(result.members[group], scopes.MembersOf(scope)...)
	}

	for _, group := range result.order {
		members := result.members[group]
		sort.Slice(members, func(i, j int) bool { return members[i] < members[j] })
		result.members[group] = members
	}

	return result
}

// sortByStartDescending and sortByEndDescending are the two orderings the sweep imposes on the
// scopes sharing one position. Extracted so the probes in the tests can reuse them rather than
// restating the comparators and accidentally measuring a different ordering.
//
// The two have DIFFERENT reachability on this corpus and it is worth stating, because a reader
// seeing one mutation caught and the other surviving would otherwise suspect the fixtures.
//
// `sortByStartDescending` is load-bearing. 11 end positions over 400 files hold more than one
// scope, up to 5 at once, and reversing this comparator changes a real answer in exactly one
// function: two scopes ending at 292, starting at 30 and 24, which stay separate under upstream's
// order and union spuriously under the reverse. `sharedEndReversedStack` is that case minimised, and
// it is the only fixture in the table that can see this comparator.
//
// `sortByEndDescending` is UNREACHABLE here, and this is a measurement rather than an argument:
// every one of the 3,306 start positions on the corpus holds exactly ONE scope, so the slice it
// sorts always has length one and no comparator can be called. A mutation reversing it therefore
// survives correctly. It is kept rather than deleted because it is upstream's and because it is what
// makes the stack ordering meaningful the moment two scopes can start together.
//
// That verdict has now EXPIRED, exactly as predicted, and this note is the re-measurement rather
// than an inherited zero. `AlignReactiveScopesToBlockScopes` landed as the producer this paragraph
// named, and it runs IN FRONT of this pass, so the table this sorts is the aligned one. Measured
// over the same corpus after alignment: 92 start positions hold several scopes, the widest holding
// 29, against 0 before. `sortByEndDescending` is therefore REACHABLE and load-bearing now.
//
// It also changes real answers, and the check that shows it is not the obvious one. Reversing the
// comparator leaves the union COUNT identical at 412, so a count assertion still reads as equivalent;
// the scope partition differs in 5 functions, which is only visible by comparing groups and ranges.
// `TestAlignVoidsTheMergesComparatorVerdicts` runs that comparison with the unaligned table as its
// control, where the two spellings still agree on every function.
func sortByStartDescending(scopes []ScopeId, rangeOf func(ScopeId) mutation_aliasing.MutableRange) {
	sort.SliceStable(scopes, func(i, j int) bool {
		return rangeOf(scopes[i]).Start > rangeOf(scopes[j]).Start
	})
}

func sortByEndDescending(scopes []ScopeId, rangeOf func(ScopeId) mutation_aliasing.MutableRange) {
	sort.SliceStable(scopes, func(i, j int) bool {
		return rangeOf(scopes[i]).End > rangeOf(scopes[j]).End
	})
}
