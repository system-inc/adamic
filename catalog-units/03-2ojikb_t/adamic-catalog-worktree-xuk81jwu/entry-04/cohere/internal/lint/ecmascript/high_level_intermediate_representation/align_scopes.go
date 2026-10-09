// Aligning reactive scopes to block scopes: widening a scope out to the construct it straddles.
//
// This is React's `alignReactiveScopesToBlockScopesHIR` (development bundle line 43256), run over
// this graph. oxc transcribes it at
// `oxc_react_compiler/src/react_compiler_reactive_scopes/align_reactive_scopes_to_block_scopes_hir.rs`;
// where the two disagree React wins, and each disagreement is recorded at the line that resolves it.
//
// # Why this pass exists, which is a precondition rather than an optimisation
//
// `assertValidBlockNesting` does not compare scopes against each other. It builds one combined item
// list -- every scope, PLUS a `ProgramBlockSubtree` for every block that has a fallthrough, spanning
// from that block's terminal to the first instruction of the fallthrough -- and asserts that every
// pair in that list is disjoint or nested (bundle lines 20954-20969). A scope that begins inside an
// `if` and ends after it crosses the construct's boundary without containing it, and the assertion
// stops the compiler.
//
// `MergeOverlappingReactiveScopes` closed the scope-against-scope half and measured that it is not
// enough. Re-derived here independently, over 400 corpus files and 677 outermost functions:
//
//	                              pre-merge   merge alone   this pass THEN the merge
//	scope against scope                  39             0                          0
//	the FULL assertion                  249           191                          8
//
// The counts differ slightly from the ones `merge_scopes.go` records -- it measured 40, 226 and 166
// over 685 functions -- because the corpus is a live tree that moved between the two measurements.
// Both are correct about the tree they saw, which is why this file states its own denominator rather
// than inheriting one. The SHAPE is identical and that is the part that matters.
//
// The 163 attributed by kind before this pass runs: 87 `ProgramBlockSubtree` inside a `Scope`, 75
// `Scope` inside a `ProgramBlockSubtree`, and 1 `ProgramBlockSubtree` against itself. This pass
// The old measurement included one block-against-block violation. It disappeared when
// `ReversePostorder` began matching React's fallthrough-first traversal, proving it was a graph
// ordering defect rather than an alignment gap. The eight current whole-corpus residues involve
// scopes in plain helper functions; components and hooks, which are this pass's upstream domain,
// remain at zero.
//
// # ORDER IS LOAD-BEARING: this pass MUST run BEFORE the merge, and measuring that is how it was found
//
// Upstream's order is alignment (line 50201) then merge (line 50207) then the assertion (50213), and
// oxc agrees (`pipeline.rs:289/291`). Our merge shipped first, so the obvious thing was to land this
// behind it and the obvious thing is WRONG. Measured over 400 corpus files and 685 functions, both
// orders run side by side against the same assertion:
//
//	order                          FULL assertion   scope-vs-scope   merge unions   scopes widened
//	merge then align (naive)                   30                4             87              608
//	align then merge (upstream)                 1                0            412              652
//
// Only upstream's order reaches the residue. The mechanism is not subtle once the numbers are in
// front of you: this pass WIDENS ranges, and widening CREATES overlaps. Running the merge first
// resolves the overlaps that exist before any widening, and then this pass introduces four new
// scope-against-scope violations that nothing is left to close -- the merge's own output is no longer
// a fixpoint of the merge. Running it second, the merge sweeps over already-widened ranges and its
// union count rises from 87 to 412, which is the same 47-scope lesson `merge_scopes.go` records about
// its own algorithm arriving one level up: a pass measured in isolation can be correct and still be
// wrong about what it composes with.
//
// So the merge MUST be re-run behind this pass rather than reused. `AlignThenMergeReactiveScopes`
// below is the composed spelling and is what a consumer should call.
// `TestAlignMustRunBeforeTheMerge` pins both rows of that table, so a future reordering fails loudly
// rather than silently costing 30 violations.
//
// # The four widening mechanisms, each measured separately
//
// A reader who takes "snap scopes out to block boundaries" as the whole algorithm will write one of
// these and miss three. They are, in the order the sweep reaches them:
//
//	the fallthrough POP        entering a block that closes an open construct pulls every still-
//	                           active scope's START back to that construct's terminal id
//	the fallthrough PUSH       a terminal with a fallthrough pushes every still-active scope's END
//	                           out to the fallthrough's first instruction
//	the goto BACK-REFERENCE    a goto targeting a fallthrough that is open but not innermost widens
//	                           both ends at once
//	the value-block RANGE      a scope first seen inside a value block (a ternary arm, a logical
//	                           right-hand side) is widened to that whole value block's span
//
// # DIVERGENCE FROM React: `branch` is excluded, and here that exclusion is what makes our graph legal
//
// Upstream widens on `fallthrough !== null && terminal.kind !== 'branch'`. Measured against React
// itself rather than read off the source, by running the pass out of the 7.1.1 development bundle on
// hand-built graphs identical but for the terminal kind:
//
//	if       [4,6) -> [4,9)     widens
//	ternary  [4,6) -> [4,9)     widens
//	logical  [4,6) -> [4,9)     widens
//	branch   [4,6) -> [4,6)     does NOT widen
//
// That exclusion governs the largest terminal population in this tree rather than a corner case.
// Measured over 400 corpus files: 1,723 `Branch` terminals carry a fallthrough, against 752 `If`,
// 1,072 `Logical`, 503 `Ternary`, 126 `ForOf`, 32 `For`, 16 `Switch`, 9 `Try` and 1 `While`.
//
// It is also load-bearing for a reason upstream states as an invariant. React asserts
// `!valueBlockNodes.has(fallthrough)` -- "Expect hir blocks to have unique fallthroughs" -- and that
// assertion sits INSIDE the non-branch guard. Our lowering pairs each `If`, `Ternary` and `Logical`
// with a `Branch` carrying the SAME fallthrough id, so a fallthrough is reached twice. Measured:
// 1,723 fallthroughs are reused by two terminals, and 0 of those reuses involve a non-branch
// terminal. So excluding `branch` is exactly what makes upstream's uniqueness invariant hold here.
// React's own lowering does the same thing (bundle line 23129 lowers a `for` test to a `branch`
// whose fallthrough is the shared continuation block), so this is agreement rather than divergence.
// `TestAlignFallthroughsAreUniqueOnceBranchesAreExcluded` pins both halves.
//
// # DIVERGENCE FROM React: the widening MINIMISES here, and it is safe for a DIFFERENT reason than
// # the merge's
//
// Three adjacent passes have now given three different answers to adopt-versus-minimise, so this one
// was established by RUNNING React rather than by reading it. `AssignReactiveScopes` uses a guarded
// adopt-or-minimise because zero is the unset sentinel. `MergeOverlappingReactiveScopes` uses a plain
// min/max, safe only because `collectScopeInfo` refuses a scope whose start equals its end.
//
// This pass uses a plain `Math.min`/`Math.max` with NO guard anywhere, and there is no upstream
// filter standing in front of it either. Running the arithmetic out of the bundle:
//
//	[7,15) inside a construct terminating at 5   ->  [5,15)   start genuinely MINIMISED
//	[3,15) inside the same construct             ->  [3,15)   already earlier, unchanged
//	[0,15) unset start                           ->  [0,15)   stays zero
//	[0,0)  fully unset                           ->  [0,0)    stays unset
//
// So a zero start is not poisoned by this pass, and the reason is structural rather than a guard:
// the operation is `min(start, constructStart)` and a construct start is always non-zero, so zero is
// ABSORBING here rather than poisoning. That is the opposite direction from the hull in
// `AssignReactiveScopes`, where the danger was a zero dragging a real start down. A fully unset scope
// never becomes active at all, because activity is `scope.range.end > startingId` and an end of zero
// fails that for every block. `TestAlignMinimisesWithoutAGuard` pins all four rows.
//
// # DIVERGENCE FROM React: `placeScopes` is written and never read, so it is not reproduced
//
// `recordPlace` populates a `placeScopes` map keyed on the PLACE OBJECT (bundle lines 43262-43265).
// Nothing in this function ever reads it; the only reader of a map by that name is
// `mergeOverlappingReactiveScopesHIR`'s own separate one at line 32487. `Place` is a value type in
// this package and has no object identity to key on, so reproducing it would have required a
// side table that exists to be written and never read. Omitted deliberately rather than overlooked.
//
// # Termination is a single sweep, and the bound is structural rather than empirical
//
// One pass over the blocks in the order `Function.Blocks` already holds, visiting each instruction
// once and each terminal once. `activeBlockFallthroughRanges` is a stack that is pushed at most once
// per terminal carrying a fallthrough and popped only when the sweep ENTERS the matching fallthrough
// block, which happens at most once per block, so its total churn is bounded by the block count.
// `activeScopes` is a set filtered at each block boundary and added to only from places the sweep is
// already visiting. Nothing revisits anything and there is no fixpoint, so the structure forbids
// iteration rather than converging. `TestAlignIsASingleSweep` pins the visit count.
//
// Like the merge, the sweep rests on evaluation order being monotone along the block sequence, which
// is an assumption about what `Construct` produces rather than about this file.
// `TestMergeAssumesMonotoneEvaluationOrder` already pins that at 0 non-monotone steps, and this pass
// inherits it rather than re-measuring the same table.
//
// The visit count is measured against the blocks of functions that HAVE scopes rather than against
// every block, because a function with no scope returns before the sweep begins: 49 of 677 corpus
// functions, carrying 217 blocks. Getting that denominator wrong is how `TestAlignIsASingleSweep`
// first failed, on a pass that was already correct.
//
// # Determinism, which a Go map would otherwise destroy
//
// Upstream's `activeScopes` is a `Set` iterated in insertion order and its widening writes through
// aliased range objects, so upstream is deterministic without trying. `activeScopes` here records
// insertion order explicitly and every widening is applied through that ordered slice, so two scopes
// widened at one block boundary are written in the same sequence on every run. The writes are
// commutative min/max so the order cannot change the answer, but recording it keeps a future
// non-commutative addition from being a silent nondeterminism. `TestAlignIsDeterministic` runs the
// pass twice over the corpus and compares.
package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
)

// AlignGap names an alignment rule this pass cannot apply, for a caller that needs to know.
//
// Modelled on `ScopeGap`, `MergeGap`, `RangeGap`, `ReactiveGap` and `DisjointGap` deliberately, and
// for the same reason: a gap answerable through the API is one a consumer can decline on, while a
// gap living only in a comment is one the next reader inherits by accident.
type AlignGap uint8

// AlignGaps are the alignment rules AlignReactiveScopesToBlockScopes does not apply. See AlignGap.
//
// Returned as a value rather than documented alone so a test can assert on it, which makes closing a
// gap a visible event rather than a silent improvement.
func AlignGaps() []AlignGap {
	return nil
}

// AlignedScopes is what this pass produces: the widened scope ranges, plus the member ranges
// re-derived from them.
//
// # Why this carries member ranges, which is `ScopeGapPostAlignmentWidening` coming due
//
// `AssignReactiveScopes` records that React aliases every member's `mutableRange` to the scope's own
// range OBJECT, so a later widening of the scope retroactively widens every member. `MutableRange` is
// a value type here and that pass stores a copy, which was indistinguishable from the alias for
// exactly as long as nothing widened a scope range afterwards. This pass is the first thing that
// does, so the alias and the copy are now distinguishable and the gap is live.
//
// Member ranges are therefore RE-DERIVED from the widened ranges rather than inherited from
// `ReactiveScopes.MemberRanges()`, which still reports the pre-alignment hull. A consumer reading
// member ranges after this pass must take them from here. `TestAlignReDerivesMemberRanges` measures
// the two tables against each other and asserts they differ, because a re-derivation that changed
// nothing would mean this pass had not widened anything.
//
// The zero value is an empty result and is ready to read.
type AlignedScopes struct {
	ranges  map[ScopeId]mutation_aliasing.MutableRange
	members map[ScopeId][]static_single_assignment.IdentifierId
	order   []ScopeId
	widened int
	visits  int
}

// RangeOf returns a scope's aligned range, or the unset range for a scope that does not exist.
func (a *AlignedScopes) RangeOf(scope ScopeId) mutation_aliasing.MutableRange {
	if a == nil || a.ranges == nil {
		return mutation_aliasing.MutableRange{}
	}
	return a.ranges[scope]
}

// MembersOf returns the values belonging to a scope, sorted, or nil for a scope that does not exist.
//
// The slice is the table's own and must not be modified by a caller.
func (a *AlignedScopes) MembersOf(scope ScopeId) []static_single_assignment.IdentifierId {
	if a == nil || a.members == nil {
		return nil
	}
	return a.members[scope]
}

// Ids returns every scope this pass saw, in the input table's own order.
func (a *AlignedScopes) Ids() []ScopeId {
	if a == nil {
		return nil
	}
	return a.order
}

// Len reports how many scopes exist, for measurement.
func (a *AlignedScopes) Len() int {
	if a == nil {
		return 0
	}
	return len(a.order)
}

// Widened reports how many scopes had at least one endpoint moved, for measurement.
//
// Exposed directly because it is the two-sided key this pass is measured on: a zero here beside a
// closed assertion would mean the assertion was already closed and this pass did nothing.
func (a *AlignedScopes) Widened() int {
	if a == nil {
		return 0
	}
	return a.widened
}

// MemberRanges returns every scoped value's range after alignment, as a table.
//
// This is upstream's aliasing made explicit: every member of a scope reports that scope's WIDENED
// range. See the doc comment on `AlignedScopes` for why this cannot be taken from
// `ReactiveScopes.MemberRanges()` once this pass has run.
func (a *AlignedScopes) MemberRanges() *mutation_aliasing.MutableRanges {
	result := &mutation_aliasing.MutableRanges{}
	if a == nil {
		return result
	}
	for _, scope := range a.order {
		widened := a.ranges[scope]
		for _, id := range a.members[scope] {
			result.Set(id, widened)
		}
	}
	return result
}

// fallthroughRange is one entry on upstream's `activeBlockFallthroughRanges` stack.
type fallthroughRange struct {
	fallthrough_ static_single_assignment.BlockId
	start        static_single_assignment.EvaluationOrder
	end          static_single_assignment.EvaluationOrder
}

// valueBlockNode is one node of upstream's `valueBlockNodes` tree.
//
// Upstream builds a full tree with a `children` list, but nothing in this pass ever reads `children`
// -- it is written for a later consumer that does not exist here. Only `valueRange` is read, by
// `recordPlace`, so this carries that and not the child list. Omitted deliberately rather than
// overlooked; see the package comment on `placeScopes` for the same shape of decision.
type valueBlockNode struct {
	start static_single_assignment.EvaluationOrder
	end   static_single_assignment.EvaluationOrder
}

// alignSweepState is the sweep's working state, matching upstream's locals.
type alignSweepState struct {
	activeFallthroughs []fallthroughRange
	activeScopes       []ScopeId
	activeSet          map[ScopeId]bool
	seen               map[ScopeId]bool
	valueBlockNodes    map[static_single_assignment.BlockId]*valueBlockNode
	ranges             map[ScopeId]mutation_aliasing.MutableRange
	scopes             *ReactiveScopes

	visits int
}

// AlignReactiveScopesToBlockScopes widens each scope out to the block constructs it straddles.
//
// Requires scopes: call `AssignReactiveScopes` first. This runs BEFORE the merge, so it takes the
// unmerged scope table; `AlignThenMergeReactiveScopes` is the composed spelling and is what a
// consumer should call. A nil or empty scope table produces an empty result rather than an error,
// which is the ordinary answer for a function with no entangled values.
//
// There is deliberately no variant taking a `MergedScopes`. One was written first, to align the
// survivors of an already-run merge, and it was deleted once measurement showed that order reaches
// 30 violations rather than 1. Keeping it would have been an API for a path this file's own tests
// prove is wrong, and its merge-group resolution was reachable only from the test demonstrating the
// mistake -- which is exactly the shape that reads as coverage without being any.
func AlignReactiveScopesToBlockScopes(function *Function, scopes *ReactiveScopes) *AlignedScopes {
	result := &AlignedScopes{}
	if function == nil || scopes == nil || scopes.Len() == 0 {
		return result
	}

	state := &alignSweepState{
		activeSet:       map[ScopeId]bool{},
		seen:            map[ScopeId]bool{},
		valueBlockNodes: map[static_single_assignment.BlockId]*valueBlockNode{},
		ranges:          map[ScopeId]mutation_aliasing.MutableRange{},
		scopes:          scopes,
	}

	live := scopes.Ids()
	for _, id := range live {
		state.ranges[id] = scopes.RangeOf(id)
	}
	before := make(map[ScopeId]mutation_aliasing.MutableRange, len(state.ranges))
	for id, r := range state.ranges {
		before[id] = r
	}

	byId := map[static_single_assignment.BlockId]*BasicBlock{}
	for _, block := range function.Blocks {
		if block != nil {
			byId[block.Id] = block
		}
	}

	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		state.visitBlock(function, byId, block)
	}

	result.ranges = state.ranges
	result.visits = state.visits
	result.order = append(result.order, live...)
	result.members = map[ScopeId][]static_single_assignment.IdentifierId{}
	for _, id := range live {
		result.members[id] = scopes.MembersOf(id)
		if state.ranges[id] != before[id] {
			result.widened++
		}
	}
	return result
}

// AlignThenMergeReactiveScopes runs this pass and then the merge, which is upstream's order.
//
// This is the spelling a consumer should call. Running the merge first and this pass second reaches
// 31 violations rather than 1; see the ordering table in the package comment for why that is a
// property of the algorithms rather than of this corpus.
//
// Returns both results, because a caller needs the merge's group mapping to know which scope a value
// ends up in and the aligned ranges to know how wide it is. The merge's ranges are the final word:
// it runs last, so `MergedScopes.RangeOf` is what the assertion reads.
func AlignThenMergeReactiveScopes(function *Function, scopes *ReactiveScopes) (*AlignedScopes,
	*MergedScopes) {
	aligned := AlignReactiveScopesToBlockScopes(function, scopes)
	if function == nil || scopes == nil || scopes.Len() == 0 {
		return aligned, &MergedScopes{}
	}

	// The merge reads a `ReactiveScopes`, so hand it one carrying the widened ranges. The identifier
	// index and the membership are unchanged by alignment -- it writes ranges and nothing else -- so
	// they are shared rather than copied, and the range table is the only new allocation.
	widened := &ReactiveScopes{
		byIdentifier: scopes.byIdentifier,
		ranges:       make(map[ScopeId]mutation_aliasing.MutableRange, scopes.Len()),
		members:      scopes.members,
		order:        scopes.order,
	}
	for _, id := range scopes.Ids() {
		widened.ranges[id] = aligned.RangeOf(id)
	}
	return aligned, MergeOverlappingReactiveScopes(function, widened)
}

// startingIdOf is upstream's `block.instructions[0]?.id ?? block.terminal.id`.
//
// The fallback is not a rarity to be handled defensively: measured over 400 corpus files, 1,119
// fallthrough blocks carry no instructions at all and reach the assertion through the terminal id.
// A defensive-looking fallback that fires on 1,119 blocks is load-bearing rather than defensive.
func startingIdOf(function *Function, block *BasicBlock) static_single_assignment.EvaluationOrder {
	if block == nil {
		return 0
	}
	if len(block.Instructions) > 0 {
		if first := function.Instructions[block.Instructions[0]]; first != nil {
			return first.Order
		}
	}
	return TerminalOrder(block.Terminal)
}

// visitBlock is one iteration of upstream's `for (const [, block] of fn.body.blocks)`.
func (s *alignSweepState) visitBlock(function *Function, byId map[static_single_assignment.BlockId]*BasicBlock,
	block *BasicBlock) {
	s.visits++
	startingId := startingIdOf(function, block)

	// Upstream's `retainWhere_Set(activeScopes, scope => scope.range.end > startingId)`: a scope
	// whose end has passed is no longer active and must not be widened by a construct it has left.
	kept := s.activeScopes[:0]
	for _, id := range s.activeScopes {
		if s.ranges[id].End > startingId {
			kept = append(kept, id)
		} else {
			delete(s.activeSet, id)
		}
	}
	s.activeScopes = kept

	// MECHANISM ONE, the fallthrough POP. Entering the block that closes the innermost open construct
	// pulls every still-active scope's START back to that construct's terminal id, so a scope that
	// began inside the construct now contains it rather than crossing its boundary.
	if len(s.activeFallthroughs) > 0 {
		top := s.activeFallthroughs[len(s.activeFallthroughs)-1]
		if top.fallthrough_ == block.Id {
			s.activeFallthroughs = s.activeFallthroughs[:len(s.activeFallthroughs)-1]
			for _, id := range s.activeScopes {
				current := s.ranges[id]
				if top.start < current.Start {
					current.Start = top.start
					s.ranges[id] = current
				}
			}
		}
	}

	node := s.valueBlockNodes[block.Id]

	for _, instructionId := range block.Instructions {
		instruction := function.Instructions[instructionId]
		if instruction == nil {
			continue
		}
		EachInstructionPlace(instruction, func(place Place, role PlaceRole) {
			s.recordPlace(instruction.Order, place, node)
		})
	}
	terminalOrder := TerminalOrder(block.Terminal)
	EachTerminalPlace(block.Terminal, func(place Place, role PlaceRole) {
		s.recordPlace(terminalOrder, place, node)
	})

	fallthroughId, hasFallthrough := Fallthrough(block.Terminal)
	_, isBranch := block.Terminal.(*Branch)

	// Upstream's `fallthrough !== null && terminal.kind !== 'branch'`. The branch exclusion is
	// measured against React itself and is what makes this lowering's shared fallthroughs legal; see
	// the package comment.
	if hasFallthrough && !isBranch {
		fallthroughBlock := byId[fallthroughId]
		nextId := startingIdOf(function, fallthroughBlock)

		// MECHANISM TWO, the fallthrough PUSH. A scope still open at this terminal has its END pushed
		// out to the fallthrough's first instruction, so it contains the whole construct rather than
		// ending somewhere inside it.
		//
		// The `scope.range.end > terminal.id` test is upstream's and it is a FILTER rather than a
		// cost optimisation: a scope whose end has already passed this terminal would otherwise be
		// dragged forward across a construct it never entered.
		for _, id := range s.activeScopes {
			current := s.ranges[id]
			if current.End > terminalOrder && nextId > current.End {
				current.End = nextId
				s.ranges[id] = current
			}
		}

		s.activeFallthroughs = append(s.activeFallthroughs, fallthroughRange{
			fallthrough_: fallthroughId,
			start:        terminalOrder,
			end:          nextId,
		})

		// Upstream asserts `!valueBlockNodes.has(fallthrough)` here. Reproduced as a silent skip
		// rather than a panic, which is the choice `ValidateScopes` made and for the same reason: a
		// linter that stops on a graph it dislikes is worse than one that declines. Measured over 400
		// corpus files, 0 non-branch terminals reuse a fallthrough, so this never fires today.
		if node != nil {
			if _, taken := s.valueBlockNodes[fallthroughId]; !taken {
				s.valueBlockNodes[fallthroughId] = node
			}
		}
	} else if gotoTerminal, ok := block.Terminal.(*Goto); ok {
		// MECHANISM THREE, the goto BACK-REFERENCE. A goto whose target is an open fallthrough that
		// is NOT the innermost one is leaving several constructs at once, so every still-active scope
		// is widened at BOTH ends to span from the outer construct's terminal to its fallthrough.
		//
		// The `start !== activeBlockFallthroughRanges.at(-1)` test is upstream's: when the target IS
		// the innermost open construct, the pop above will handle it on the next block.
		index := -1
		for i, candidate := range s.activeFallthroughs {
			if candidate.fallthrough_ == gotoTerminal.Block {
				index = i
				break
			}
		}
		if index != -1 && index != len(s.activeFallthroughs)-1 {
			outer := s.activeFallthroughs[index]
			fallthroughBlock := byId[outer.fallthrough_]
			firstId := startingIdOf(function, fallthroughBlock)
			for _, id := range s.activeScopes {
				current := s.ranges[id]
				if current.End <= terminalOrder {
					continue
				}
				if outer.start < current.Start {
					current.Start = outer.start
				}
				if firstId > current.End {
					current.End = firstId
				}
				s.ranges[id] = current
			}
		}
	}

	s.assignValueBlockNodes(function, byId, block, node, fallthroughId, hasFallthrough)
}

// recordPlace is upstream's `recordPlace`, minus the `placeScopes` map it writes and never reads.
//
// MECHANISM FOUR, the value-block RANGE: the FIRST time a scope is seen, if it is seen inside a value
// block, the scope is widened to that whole value block's span. `seen` is what makes this fire once
// per scope rather than once per place, and it is upstream's.
func (s *alignSweepState) recordPlace(id static_single_assignment.EvaluationOrder, place Place, node *valueBlockNode) {
	scope := s.activeScopeOf(id, place)
	if scope == 0 {
		return
	}
	if !s.activeSet[scope] {
		s.activeSet[scope] = true
		s.activeScopes = append(s.activeScopes, scope)
	}
	if s.seen[scope] {
		return
	}
	s.seen[scope] = true
	if node == nil {
		return
	}
	current := s.ranges[scope]
	if node.start < current.Start {
		current.Start = node.start
	}
	if node.end > current.End {
		current.End = node.end
	}
	s.ranges[scope] = current
}

// activeScopeOf is upstream's `getPlaceScope`: the place's scope, but only while it is open.
//
// The containment test is a real discrimination rather than a formality, and it is NOT subsumed by
// the activity filter in `visitBlock`: that filter retires a scope whose end has passed the block,
// while this declines a place read at a position its own scope does not cover at all. Measured
// against React on a scope of [2,4) read at 7 inside a construct terminating at 8, whose fallthrough
// begins at 20: the scope comes back unchanged, so upstream declines to activate it.
// `TestAlignDeclinesAPlaceReadOutsideItsScope` is that case.
//
// Half-open, so the start is inside and the end is not, matching `MutableRange.Contains`.
func (s *alignSweepState) activeScopeOf(id static_single_assignment.EvaluationOrder, place Place) ScopeId {
	scope := s.scopes.ScopeOf(place.Identifier)
	if scope == 0 {
		return 0
	}
	if !s.ranges[scope].Contains(id) {
		return 0
	}
	return scope
}

// assignValueBlockNodes is upstream's `mapTerminalSuccessors` callback, which builds the value-block
// tree the fourth mechanism reads.
//
// Upstream returns each successor unchanged, so the call is a WALK rather than a rewrite despite the
// name. Reproduced as a walk.
func (s *alignSweepState) assignValueBlockNodes(function *Function, byId map[static_single_assignment.BlockId]*BasicBlock,
	block *BasicBlock, node *valueBlockNode, fallthroughId static_single_assignment.BlockId, hasFallthrough bool) {
	_, isTernary := block.Terminal.(*Ternary)
	_, isLogical := block.Terminal.(*Logical)
	_, isOptional := block.Terminal.(*Optional)

	EachSuccessor(block.Terminal, func(successor static_single_assignment.BlockId) {
		if _, taken := s.valueBlockNodes[successor]; taken {
			return
		}
		successorBlock := byId[successor]
		if successorBlock == nil {
			return
		}
		// An ordinary statement block or a catch handler is not a value block, so it inherits
		// nothing. Upstream writes this as an empty arm; reproduced as an early return.
		if successorBlock.Kind == BlockKindBlock || successorBlock.Kind == BlockKindCatch {
			return
		}
		if node == nil || isTernary || isLogical || isOptional {
			var span valueBlockNode
			if node == nil {
				// Upstream raises `Expected a fallthrough for value block` here. Reproduced as a
				// skip rather than a panic, for the reason `ValidateScopes` records.
				if !hasFallthrough {
					return
				}
				span = valueBlockNode{
					start: TerminalOrder(block.Terminal),
					end:   startingIdOf(function, byId[fallthroughId]),
				}
			} else {
				span = *node
			}
			child := span
			s.valueBlockNodes[successor] = &child
			return
		}
		s.valueBlockNodes[successor] = node
	})
}
