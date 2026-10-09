package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
	"sort"
	"testing"
)

// ---------------------------------------------------------------------------
// The algorithm, pinned against React's own merge pass
// ---------------------------------------------------------------------------

// mergeCase builds a synthetic function with the given scopes and instructions.
//
// Synthetic rather than lowered, for the reason `TestScopeHullMatchesReact` records: the point is to
// pin the sweep against React's, and a lowered input cannot be made to produce an arbitrary scope
// table. Each instruction gets an lvalue and optionally a use, each carrying an identifier whose
// scope and mutable range are stated outright.
type mergeCase struct {
	name string
	// scopes are [start, end) in scope-id order starting at 1.
	scopes []mutation_aliasing.MutableRange
	// instructions are {order, lvalueScope, useScope}; a scope index of 0 means no scope.
	instructions [][3]int
	terminal     static_single_assignment.EvaluationOrder
	// terminalScope, when non-zero, gives the terminal a Return carrying a place in that scope.
	// Zero leaves the terminal an Unreachable with no operands.
	terminalScope int
	terminalRange mutation_aliasing.MutableRange
	// wantGroups maps each scope id to the id it should end up in.
	wantGroups map[ScopeId]ScopeId
	// wantRanges is the range each surviving scope should carry.
	wantRanges map[ScopeId]mutation_aliasing.MutableRange
}

func buildMergeCase(t *testing.T, testCase mergeCase) (*Function, *ReactiveScopes) {
	t.Helper()

	function := &Function{}
	scopes := &ReactiveScopes{
		byIdentifier: map[static_single_assignment.IdentifierId]ScopeId{},
		ranges:       map[ScopeId]mutation_aliasing.MutableRange{},
		members:      map[ScopeId][]static_single_assignment.IdentifierId{},
	}
	for index, scopeRange := range testCase.scopes {
		id := ScopeId(index + 1)
		scopes.ranges[id] = scopeRange
		scopes.order = append(scopes.order, id)
	}

	block := &BasicBlock{Id: 0, Kind: BlockKindBlock}
	function.Identifiers = append(function.Identifiers, nil)
	nextIdentifier := static_single_assignment.IdentifierId(1)

	place := func(scopeIndex int) Place {
		id := nextIdentifier
		nextIdentifier++
		function.Identifiers = append(function.Identifiers, &Identifier{Id: id})
		if scopeIndex != 0 {
			scope := ScopeId(scopeIndex)
			scopes.byIdentifier[id] = scope
			scopes.members[scope] = append(scopes.members[scope], id)
		}
		return Place{Identifier: id}
	}

	for _, spec := range testCase.instructions {
		order := static_single_assignment.EvaluationOrder(spec[0])
		lvalue := place(spec[1])
		use := place(spec[2])
		instruction := &Instruction{
			Id:     InstructionId(len(function.Instructions)),
			Order:  order,
			LValue: lvalue,
			Value:  &LoadLocal{Place: use},
		}
		function.Instructions = append(function.Instructions, instruction)
		block.Instructions = append(block.Instructions, instruction.Id)
	}
	if testCase.terminalScope != 0 {
		block.Terminal = &Return{Value: place(testCase.terminalScope), Order: testCase.terminal}
	} else {
		block.Terminal = &Unreachable{Order: testCase.terminal}
	}
	function.Blocks = append(function.Blocks, block)

	return function, scopes
}

// TestMergeMatchesReact pins the sweep against React's real implementation.
//
// These five cases were not invented. They were run through React 7.1.1's own
// `mergeOverlappingReactiveScopesHIR` -- reached by copying the development bundle, appending an
// export inside its IIFE, and linking the four peer dependencies that live only in the pnpm store --
// and the expectations below are the output that produced, transcribed.
//
// The cases exist to SEPARATE implementations rather than to confirm one, and `nestedButInterleaved`
// is the whole reason this table is here. Two scopes that nest perfectly, [1,9) and [3,5), still
// merge, because the outer one is READ at instruction 4 while the inner is open. No pairwise
// range-overlap rule can produce that answer: the ranges nest, so a comparison of the two intervals
// says nothing is wrong. A port written from the natural description of this pass -- "union any two
// scopes that overlap without nesting" -- passes `overlapNonNested`, `properlyNested`, `disjoint`
// and `sameStartSameEnd`, and fails exactly this one.
func TestMergeMatchesReact(t *testing.T) {
	t.Parallel()

	cases := []mergeCase{
		{
			// Two scopes overlapping without nesting. Scope 1 ends at 5 while scope 2 is open, so
			// scope 2 is above it on the stack and both are unioned into scope 1, widened to [1,8).
			name:         "overlapNonNested",
			scopes:       []mutation_aliasing.MutableRange{{Start: 1, End: 5}, {Start: 3, End: 8}},
			instructions: [][3]int{{1, 1, 0}, {3, 2, 0}, {4, 0, 1}},
			terminal:     9,
			wantGroups:   map[ScopeId]ScopeId{1: 1, 2: 1},
			wantRanges:   map[ScopeId]mutation_aliasing.MutableRange{1: {Start: 1, End: 8}},
		},
		{
			// Properly nested with no interleaving use: nothing merges and no range moves.
			name:         "properlyNested",
			scopes:       []mutation_aliasing.MutableRange{{Start: 1, End: 9}, {Start: 3, End: 5}},
			instructions: [][3]int{{1, 1, 0}, {3, 2, 0}},
			terminal:     10,
			wantGroups:   map[ScopeId]ScopeId{1: 1, 2: 2},
			wantRanges:   map[ScopeId]mutation_aliasing.MutableRange{1: {Start: 1, End: 9}, 2: {Start: 3, End: 5}},
		},
		{
			// The case that separates this pass from a pairwise overlap rule. The ranges nest, but
			// scope 1's value is read at 4 while scope 2 sits above it, so they union. React widens
			// nothing here, because scope 2's range is already inside scope 1's.
			name:         "nestedButInterleavedUse",
			scopes:       []mutation_aliasing.MutableRange{{Start: 1, End: 9}, {Start: 3, End: 5}},
			instructions: [][3]int{{1, 1, 0}, {3, 2, 0}, {4, 0, 1}},
			terminal:     10,
			wantGroups:   map[ScopeId]ScopeId{1: 1, 2: 1},
			wantRanges:   map[ScopeId]mutation_aliasing.MutableRange{1: {Start: 1, End: 9}},
		},
		{
			// Two scopes starting together and ending together cannot nest either way, so the start
			// branch unions them directly rather than waiting for an end or a use.
			name:         "sameStartSameEnd",
			scopes:       []mutation_aliasing.MutableRange{{Start: 2, End: 6}, {Start: 2, End: 6}},
			instructions: [][3]int{{2, 1, 2}},
			terminal:     7,
			wantGroups:   map[ScopeId]ScopeId{1: 1, 2: 1},
			wantRanges:   map[ScopeId]mutation_aliasing.MutableRange{1: {Start: 2, End: 6}},
		},
		{
			// Three scopes ending at the SAME position with different starts. The closing order is
			// start-descending, so scope 3 closes first, then 2, then 1 -- and each of them is
			// buried under the ones that started later, so all three union. This is the case that
			// exercises `sortByStartDescending` on a real multi-scope group: 11 end positions on
			// the corpus hold more than one scope, up to 5 of them.
			name:         "threeEndTogether",
			scopes:       []mutation_aliasing.MutableRange{{Start: 1, End: 9}, {Start: 3, End: 9}, {Start: 5, End: 9}},
			instructions: [][3]int{{1, 1, 0}, {3, 2, 0}, {5, 3, 0}, {6, 0, 1}},
			terminal:     10,
			wantGroups:   map[ScopeId]ScopeId{1: 1, 2: 1, 3: 1},
			wantRanges:   map[ScopeId]mutation_aliasing.MutableRange{1: {Start: 1, End: 9}},
		},
		{
			// Two scopes sharing an end with a third nested strictly inside both. Nothing merges:
			// the inner scope closes before the shared end is reached, so when 1 and 2 close they
			// are each on top of the stack in turn. React's own answer, and it is the negative
			// control for the case above -- a shared end alone does not force a union.
			name:         "sharedEndWithInterleave",
			scopes:       []mutation_aliasing.MutableRange{{Start: 1, End: 7}, {Start: 2, End: 7}, {Start: 3, End: 5}},
			instructions: [][3]int{{1, 1, 0}, {2, 2, 0}, {3, 3, 0}},
			terminal:     8,
			wantGroups:   map[ScopeId]ScopeId{1: 1, 2: 2, 3: 3},
			wantRanges:   map[ScopeId]mutation_aliasing.MutableRange{1: {Start: 1, End: 7}, 2: {Start: 2, End: 7}, 3: {Start: 3, End: 5}},
		},
		{
			// The case that makes the CLOSING ORDER load-bearing, minimised from the one function on
			// the corpus where reversing it changes an answer (two scopes ending at 292, starting at
			// 30 and 24). Both scopes end at 9. The later-started one is pushed second, so it sits
			// ON TOP. Closing start-descending pops it first, while it is on top, so nothing unions
			// and both survive. Closing in the other order pops the earlier-started one out from
			// underneath, which unions them spuriously.
			//
			// React's own answer is that they stay separate. Without this case a mutation reversing
			// `sortByStartDescending` SURVIVES the entire rest of this table, because every other
			// shared-end case here either unions under both orderings or holds one scope per
			// position.
			name:         "sharedEndReversedStack",
			scopes:       []mutation_aliasing.MutableRange{{Start: 5, End: 9}, {Start: 3, End: 9}},
			instructions: [][3]int{{3, 2, 0}, {5, 1, 0}},
			terminal:     10,
			wantGroups:   map[ScopeId]ScopeId{1: 1, 2: 2},
			wantRanges:   map[ScopeId]mutation_aliasing.MutableRange{1: {Start: 5, End: 9}, 2: {Start: 3, End: 9}},
		},
		{
			// The union arrives through the TERMINAL rather than through an instruction. Scope 1 is
			// open across the whole function and scope 2 opens inside it; the terminal at 6 reads
			// scope 1's value while scope 2 sits above it on the stack, so `visitPlace` on the
			// terminal operand is the only thing that can union them.
			//
			// Without this case a mutation deleting the terminal-place visit entirely SURVIVES. It
			// is reachable on the corpus -- 990 scoped terminal places over 400 files -- but inert
			// there, because every one of those unions is already produced by an instruction visit.
			// So this is a genuine blind spot rather than an unreachable branch, and only a
			// synthetic case can pin it.
			name:          "terminalOperandUnions",
			scopes:        []mutation_aliasing.MutableRange{{Start: 1, End: 9}, {Start: 3, End: 9}},
			instructions:  [][3]int{{1, 1, 0}, {3, 2, 0}},
			terminal:      6,
			terminalScope: 1,
			wantGroups:    map[ScopeId]ScopeId{1: 1, 2: 1},
			wantRanges:    map[ScopeId]mutation_aliasing.MutableRange{1: {Start: 1, End: 9}},
		},
		{
			// Disjoint scopes never interact: the first leaves the stack before the second is
			// pushed, so no union can see both.
			name:         "disjoint",
			scopes:       []mutation_aliasing.MutableRange{{Start: 1, End: 3}, {Start: 5, End: 8}},
			instructions: [][3]int{{1, 1, 0}, {5, 2, 0}},
			terminal:     9,
			wantGroups:   map[ScopeId]ScopeId{1: 1, 2: 2},
			wantRanges:   map[ScopeId]mutation_aliasing.MutableRange{1: {Start: 1, End: 3}, 2: {Start: 5, End: 8}},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function, scopes := buildMergeCase(t, testCase)
			merged := MergeOverlappingReactiveScopes(function, scopes)

			for scope, want := range testCase.wantGroups {
				if got := merged.GroupOf(scope); got != want {
					t.Errorf("scope %d merged into %d, want %d -- this is React's own answer",
						scope, got, want)
				}
			}
			for scope, want := range testCase.wantRanges {
				if got := merged.RangeOf(scope); got != want {
					t.Errorf("scope %d range = [%d,%d), want [%d,%d) -- React's own answer",
						scope, got.Start, got.End, want.Start, want.End)
				}
			}
			if got, want := merged.Len(), len(testCase.wantRanges); got != want {
				t.Errorf("%d scopes survived, want %d", got, want)
			}
		})
	}
}

// TestMergeIsNotAPairwiseOverlapRule is the negative control for the case above.
//
// It asserts the specific input on which the stack discipline and the natural pairwise description
// disagree, so a reader who simplifies this pass into "union scopes whose ranges overlap without
// nesting" cannot do it silently. The two ranges here NEST, so a pairwise rule finds nothing to do
// and leaves two scopes; React unions them because of the use at instruction 4.
func TestMergeIsNotAPairwiseOverlapRule(t *testing.T) {
	t.Parallel()

	function, scopes := buildMergeCase(t, mergeCase{
		scopes:       []mutation_aliasing.MutableRange{{Start: 1, End: 9}, {Start: 3, End: 5}},
		instructions: [][3]int{{1, 1, 0}, {3, 2, 0}, {4, 0, 1}},
		terminal:     10,
	})
	merged := MergeOverlappingReactiveScopes(function, scopes)

	if merged.Len() != 1 {
		t.Fatalf("got %d surviving scopes, want 1; the interleaving use did not union them, which "+
			"is what a pairwise range comparison would also conclude", merged.Len())
	}
	// And the ranges genuinely nest, which is what makes this a control rather than a duplicate.
	outer, inner := mutation_aliasing.MutableRange{Start: 1, End: 9}, mutation_aliasing.MutableRange{Start: 3, End: 5}
	if !(outer.Start <= inner.Start && inner.End <= outer.End) {
		t.Fatal("this case no longer nests, so it stopped separating the two spellings")
	}
}

// TestMergeSkipsDegenerateScopes pins upstream's guard, which is what makes the plain min/max safe.
//
// `collectScopeInfo` refuses a scope whose start equals its end, so an unset [0,0) scope never
// enters the sweep. Without that guard the plain minimum in `apply` would drag a merged group's
// start to zero, which is the failure `AssignReactiveScopes` uses an adopt-or-minimise hull to
// avoid. Both halves are asserted: the degenerate scope must survive untouched, and the real scopes
// around it must merge normally.
func TestMergeSkipsDegenerateScopes(t *testing.T) {
	t.Parallel()

	function, scopes := buildMergeCase(t, mergeCase{
		scopes:       []mutation_aliasing.MutableRange{{Start: 0, End: 0}, {Start: 1, End: 5}, {Start: 3, End: 8}},
		instructions: [][3]int{{1, 2, 1}, {3, 3, 0}, {4, 0, 2}},
		terminal:     9,
	})
	merged := MergeOverlappingReactiveScopes(function, scopes)

	if got := merged.GroupOf(1); got != 1 {
		t.Errorf("the unset scope merged into %d; it must never enter the sweep at all", got)
	}
	if got := merged.RangeOf(1); got != (mutation_aliasing.MutableRange{Start: 0, End: 0}) {
		t.Errorf("the unset scope's range moved to [%d,%d)", got.Start, got.End)
	}
	// The control: the two real scopes around it must still merge, or this test would pass on a
	// pass that simply does nothing.
	if merged.GroupOf(3) != merged.GroupOf(2) {
		t.Error("the two real scopes did not merge, so this test measured nothing")
	}
	if got := merged.RangeOf(merged.GroupOf(2)); got.Start == 0 {
		t.Errorf("a merged group took a zero start from the unset scope: [%d,%d)", got.Start, got.End)
	}
}

// TestMergeConservesMembership asserts no value is dropped or duplicated by a merge.
//
// A merge that loses members would still reach zero overlapping pairs, which is the failure this
// stage is uniquely positioned to commit, so membership is asserted separately from the ranges.
func TestMergeConservesMembership(t *testing.T) {
	t.Parallel()

	function, scopes := buildMergeCase(t, mergeCase{
		scopes:       []mutation_aliasing.MutableRange{{Start: 1, End: 5}, {Start: 3, End: 8}, {Start: 10, End: 12}},
		instructions: [][3]int{{1, 1, 0}, {3, 2, 0}, {4, 0, 1}, {10, 3, 0}},
		terminal:     13,
	})
	merged := MergeOverlappingReactiveScopes(function, scopes)

	before := map[static_single_assignment.IdentifierId]bool{}
	for _, scope := range scopes.Ids() {
		for _, member := range scopes.MembersOf(scope) {
			before[member] = true
		}
	}
	after := map[static_single_assignment.IdentifierId]bool{}
	for _, scope := range merged.Ids() {
		for _, member := range merged.MembersOf(scope) {
			if after[member] {
				t.Errorf("value %d appears in two surviving scopes", member)
			}
			after[member] = true
		}
	}
	if len(before) != len(after) {
		t.Errorf("membership changed: %d values before, %d after", len(before), len(after))
	}
	for member := range before {
		if !after[member] {
			t.Errorf("value %d was dropped by the merge", member)
		}
	}
}

// TestMergeIsASingleSweep pins the termination argument empirically.
//
// The claim in the doc comment is that this pass never revisits anything: one pass over the blocks,
// each instruction and each terminal handled exactly once. That is a stronger claim than "it
// terminates" and it is the one worth pinning, because the natural pairwise spelling of this pass
// DOES iterate and a reader carrying that model would add a loop.
func TestMergeIsASingleSweep(t *testing.T) {
	t.Parallel()

	testCase := mergeCase{
		scopes:       []mutation_aliasing.MutableRange{{Start: 1, End: 5}, {Start: 3, End: 8}},
		instructions: [][3]int{{1, 1, 0}, {3, 2, 0}, {4, 0, 1}, {5, 0, 2}},
		terminal:     9,
	}
	function, scopes := buildMergeCase(t, testCase)

	state := &mergeSweepState{rangeOf: scopes.RangeOf}
	state.starts, state.ends = collectScopeInfo(function, scopes)
	memberRanges := scopes.MemberRanges()
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			state.visitInstructionId(instruction.Order)
			EachInstructionPlace(instruction, func(place Place, role PlaceRole) {
				state.visitPlace(instruction.Order, place, scopes, memberRanges)
			})
		}
		state.visitInstructionId(TerminalOrder(block.Terminal))
	}

	// One visit per instruction plus one for the terminal, and not one more.
	if want := len(testCase.instructions) + 1; state.visits != want {
		t.Errorf("the sweep visited %d positions over %d instructions and one terminal; anything "+
			"above %d means something is revisiting", state.visits, len(testCase.instructions), want)
	}
	// And both queues must be fully drained, which is what makes the bound exact rather than lucky.
	if len(state.starts) != 0 || len(state.ends) != 0 {
		t.Errorf("the sweep left %d starts and %d ends unconsumed, so some scope never closed",
			len(state.starts), len(state.ends))
	}
}

// TestMergeGapsAreDeclared makes closing a gap a visible event rather than a silent improvement.
func TestMergeGapsAreDeclared(t *testing.T) {
	t.Parallel()

	gaps := MergeGaps()
	if len(gaps) != 2 {
		t.Fatalf("got %d gaps, want 2; a gap closing or opening should be a deliberate edit", len(gaps))
	}
	if gaps[0] != MergeGapPrimitiveOperandSkip || gaps[1] != MergeGapBlockScopeAlignment {
		t.Error("the declared gaps are not the two this pass documents")
	}
}

// TestMergeHandlesAnEmptyScopeTable pins the ordinary answer for a function with no entangled values.
func TestMergeHandlesAnEmptyScopeTable(t *testing.T) {
	t.Parallel()

	if got := MergeOverlappingReactiveScopes(nil, nil); got.Len() != 0 || got.Unions() != 0 {
		t.Error("a nil input produced a non-empty result")
	}
	// A scope this pass never saw must answer itself, which is what lets a caller key a map on
	// GroupOf without first asking whether the scope merged.
	var empty *MergedScopes
	if got := empty.GroupOf(7); got != 7 {
		t.Errorf("GroupOf on an empty result returned %d, want the scope itself", got)
	}
}

// ---------------------------------------------------------------------------
// The corpus: the measurements this stage exists to make
// ---------------------------------------------------------------------------

// mergeCorpusStats walks the corpus once and returns everything the measurements below read.
func mergeCorpusStats(t *testing.T, limit int) (functions, before, after, unions int,
	nonNestedBefore, nonNestedAfter int, widthOneMultiAfter, maxWidthBefore, maxWidthAfter int,
	nestingBefore, nestingAfter int) {
	t.Helper()

	forEachCorpusFunction(t, limit, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		functions++
		before += scopes.Len()

		beforeRanges := map[ScopeId]mutation_aliasing.MutableRange{}
		for _, scope := range scopes.Ids() {
			scopeRange := scopes.RangeOf(scope)
			beforeRanges[scope] = scopeRange
			if width := int(scopeRange.End - scopeRange.Start); width > maxWidthBefore {
				maxWidthBefore = width
			}
		}
		nonNestedBefore += countNonNestedPairs(beforeRanges)

		merged := MergeOverlappingReactiveScopes(function, scopes)
		after += merged.Len()
		unions += merged.Unions()

		afterRanges := map[ScopeId]mutation_aliasing.MutableRange{}
		for _, scope := range merged.Ids() {
			scopeRange := merged.RangeOf(scope)
			afterRanges[scope] = scopeRange
			width := int(scopeRange.End - scopeRange.Start)
			if width > maxWidthAfter {
				maxWidthAfter = width
			}
			if width == 1 && len(merged.MembersOf(scope)) > 1 {
				widthOneMultiAfter++
			}
		}
		nonNestedAfter += countNonNestedPairs(afterRanges)

		blocks := programBlockSubtrees(function)
		nestingBefore += blockNestingViolations(scopeNestingItems(beforeRanges), blocks)
		nestingAfter += blockNestingViolations(scopeNestingItems(afterRanges), blocks)
	})
	return
}

// countNonNestedPairs counts scope pairs that overlap without one containing the other.
func countNonNestedPairs(ranges map[ScopeId]mutation_aliasing.MutableRange) int {
	ids := make([]ScopeId, 0, len(ranges))
	for id := range ranges {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	count := 0
	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			first, second := ranges[ids[i]], ranges[ids[j]]
			if first.Start >= second.End || second.Start >= first.End {
				continue
			}
			nested := (first.Start <= second.Start && second.End <= first.End) ||
				(second.Start <= first.Start && first.End <= second.End)
			if !nested {
				count++
			}
		}
	}
	return count
}

// TestMergeDrivesNonNestedOverlapToZero is this stage's headline, pinned as a test.
//
// Over 400 corpus files and 685 outermost functions:
//
//	                                 before   after
//	scopes                            3,306   3,218
//	scope pairs overlapping without
//	  one nesting inside the other        48       0
//
// The zero is what `buildReactiveScopeTerminalsHIR` is blocked on for its own half, and it is a
// two-sided measurement rather than a bare zero: 48 before and 0 after, from the same walk, so a
// pass that silently declined every function would fail the before-count rather than passing the
// after-count vacuously.
//
// The thresholds below are looser than those numbers. This guards the SHAPE, not the corpus, which
// moves when the tree does.
func TestMergeDrivesNonNestedOverlapToZero(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)
	functions, before, after, unions, nonNestedBefore, nonNestedAfter,
		widthOneMultiAfter, maxWidthBefore, maxWidthAfter, _, _ := mergeCorpusStats(t, 400)

	t.Logf("functions=%d scopes %d->%d (unions=%d) nonNestedOverlap %d->%d maxWidth %d->%d",
		functions, before, after, unions, nonNestedBefore, nonNestedAfter,
		maxWidthBefore, maxWidthAfter)

	if nonNestedBefore == 0 {
		t.Fatal("no scopes overlapped without nesting before the merge, so the zero after it is " +
			"about the corpus rather than about this pass")
	}
	if nonNestedAfter != 0 {
		t.Errorf("%d scope pairs still overlap without nesting after merging; the block-nesting "+
			"assertion this pass exists to satisfy would still fail", nonNestedAfter)
	}

	// The over-merge control, which is the failure this stage is uniquely positioned to commit.
	// Reaching zero by merging everything into one scope per function would satisfy the assertion
	// above and destroy the output.
	if after < before/2 {
		t.Errorf("merging collapsed %d scopes into %d, which is more than half; that is the shape "+
			"of over-merging rather than of fixing overlaps", before, after)
	}
	if widthOneMultiAfter != 0 {
		t.Errorf("%d merged scopes hold several values in a single instruction, which is the shape "+
			"catastrophic over-merging produces", widthOneMultiAfter)
	}
	if maxWidthAfter > maxWidthBefore*2 {
		t.Errorf("the widest scope grew from %d to %d instructions, which is more widening than "+
			"fixing overlaps should require", maxWidthBefore, maxWidthAfter)
	}
}

// TestMergePreservesScopeWidthAndMembership re-runs Stage 4's cross-tabulation after merging.
//
// `AssignReactiveScopes` proved its scopes were real with a cross-tabulation of scope width against
// scope membership, whose empty cell -- no multi-member scope of width one -- is what rules out
// over-merging. This pass DELIBERATELY merges scopes, so that same table is the instrument that says
// whether it merged too many. Over the same 400 files:
//
//	                       width 1      width > 1              width 1      width > 1
//	one member               1,903             80                1,903             79
//	many members                 0          1,323                    0          1,236
//	                        (before)                             (after)
//
// The empty cell stays empty, the 1,903 narrow singletons are untouched, and the widest scope does
// not move at all: 955 instructions before and after. Every scope this pass merged came from the
// wide multi-member population, which is exactly the population that can overlap.
func TestMergePreservesScopeWidthAndMembership(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)
	var oneNarrowBefore, oneWideBefore, manyNarrowBefore, manyWideBefore int
	var oneNarrowAfter, oneWideAfter, manyNarrowAfter, manyWideAfter int

	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		for _, scope := range scopes.Ids() {
			scopeRange := scopes.RangeOf(scope)
			classify(int(scopeRange.End-scopeRange.Start), len(scopes.MembersOf(scope)),
				&oneNarrowBefore, &oneWideBefore, &manyNarrowBefore, &manyWideBefore)
		}
		merged := MergeOverlappingReactiveScopes(function, scopes)
		for _, scope := range merged.Ids() {
			scopeRange := merged.RangeOf(scope)
			classify(int(scopeRange.End-scopeRange.Start), len(merged.MembersOf(scope)),
				&oneNarrowAfter, &oneWideAfter, &manyNarrowAfter, &manyWideAfter)
		}
	})

	t.Logf("before  1member/width1=%d 1member/wide=%d many/width1=%d many/wide=%d",
		oneNarrowBefore, oneWideBefore, manyNarrowBefore, manyWideBefore)
	t.Logf("after   1member/width1=%d 1member/wide=%d many/width1=%d many/wide=%d",
		oneNarrowAfter, oneWideAfter, manyNarrowAfter, manyWideAfter)

	if manyNarrowBefore != 0 || manyNarrowAfter != 0 {
		t.Errorf("the empty cell filled: %d multi-member scopes of width one before and %d after",
			manyNarrowBefore, manyNarrowAfter)
	}
	// A ceiling rather than an equality, and the change from equality is recorded rather than
	// loosened silently.
	//
	// The equality held for as long as no narrow single-value scope overlapped anything, which was a
	// property of the corpus shape rather than of this pass. Upstream has no width guard at all --
	// `MergeOverlappingReactiveScopesHIR.ts` never consults a scope's width -- and `collectScopeInfo`
	// registers any scope whose start differs from its end, so a width-one scope is eligible to be
	// unioned exactly like any other.
	//
	// The frozen-capture rule in `ranges.go` changed the shape: values that used to be swallowed by a
	// wide hull now form their own narrow scopes, and 8 of them sit inside a hull that still exists,
	// so they merge back into it. Traced on all eight before this was relaxed: every one merges into
	// a scope that CONTAINS it, and the containing scope is byte-identical before and after. On
	// `pos=5163` scope 3 is `[15,382)` in both, holding 135 members before and 73 after -- the
	// frozen values left it, formed their own scopes, and rejoined. Nothing is grouped that was not
	// grouped before.
	//
	// The final grouping moves in the safe direction, which is what makes this a relaxation rather
	// than a retreat: `[40,44)` with 3 members becomes `[42,44)` with 1, and the corpus gains 92
	// groups. Narrower scopes with fewer members invalidate less, and `under` in the scope oracle
	// holds at 5 fixtures / 7 scopes across the change.
	//
	// The empty-cell assertion above is the one that catches over-merging, and it is untouched.
	if oneNarrowAfter > oneNarrowBefore {
		t.Errorf("the narrow single-value scopes GREW from %d to %d; merging cannot create a "+
			"scope, so this population can only shrink", oneNarrowBefore, oneNarrowAfter)
	}
	if manyWideBefore == 0 {
		t.Fatal("no wide multi-member scopes exist, so the table above measured nothing")
	}
}

func classify(width, members int, oneNarrow, oneWide, manyNarrow, manyWide *int) {
	switch {
	case members == 1 && width == 1:
		*oneNarrow++
	case members == 1:
		*oneWide++
	case width == 1:
		*manyNarrow++
	default:
		*manyWide++
	}
}

// TestMergeLeavesBlockScopeAlignmentUnclosed measures the gap this pass does NOT close.
//
// `assertValidBlockNesting` is the assertion `buildReactiveScopeTerminalsHIR` is blocked on, and it
// checks scopes against PROGRAM BLOCK SUBTREES as well as against each other. This pass closes the
// scope-against-scope half completely and leaves the other half open, so the number is recorded here
// rather than left as a claim in a comment, and the next stage can read it as a precondition.
//
//	                                 before merge   after merge
//	the full assertion                        226           166
//
// Attributed by kind after merging: 78 scopes crossing a block subtree, 87 block subtrees crossing a
// scope, and 1 block subtree crossing another block subtree. That last one is not a scope problem at
// all -- measured with the scope items removed entirely, it is still there -- so it belongs to the
// lowering rather than to any scope pass.
func TestMergeLeavesBlockScopeAlignmentUnclosed(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)
	_, _, _, _, _, _, _, _, _, nestingBefore, nestingAfter := mergeCorpusStats(t, 400)

	blocksAlone := 0
	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		blocksAlone += blockNestingViolations(nil, programBlockSubtrees(function))
	})

	t.Logf("assertValidBlockNesting violations: before=%d after=%d; among block subtrees alone=%d",
		nestingBefore, nestingAfter, blocksAlone)

	if nestingAfter >= nestingBefore {
		t.Errorf("merging did not reduce the block-nesting violations at all (%d -> %d), so this "+
			"pass is not doing its half", nestingBefore, nestingAfter)
	}
	if nestingAfter == 0 {
		t.Error("the full assertion now passes, which means MergeGapBlockScopeAlignment has " +
			"closed and this test and that gap should both be revisited")
	}
	if blocksAlone == 0 {
		t.Log("no block-subtree-against-block-subtree violations remain; the lowering note in " +
			"this pass's doc comment can be dropped")
	}
}

// TestMergeFunctionOperandSkipUpperBound bounds MergeGapPrimitiveOperandSkip.
//
// Upstream skips a primitive operand of a function expression or object method, and the IR carries
// no type, so the real gate cannot be asked. This measures the UPPER BOUND instead: skipping
// every use-operand of every function expression, which is strictly more than the gate would skip.
// The control is asserted in the same test, because a small delta from a probe that barely fires
// would not be a useful bound.
func TestMergeFunctionOperandSkipUpperBound(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)
	withoutGate, withGate, candidates := 0, 0, 0

	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		withoutGate += MergeOverlappingReactiveScopes(function, scopes).Unions()
		withGate += unionsSkippingFunctionOperands(function, scopes)
		for _, block := range function.Blocks {
			if block == nil {
				continue
			}
			for _, instructionId := range block.Instructions {
				instruction := function.Instructions[instructionId]
				if instruction == nil {
					continue
				}
				if _, isFunction := instruction.Value.(*FunctionExpression); isFunction {
					EachInstructionPlace(instruction, func(place Place, role PlaceRole) {
						if role == PlaceRoleUse {
							candidates++
						}
					})
				}
			}
		}
	})

	t.Logf("unions without the gate=%d, with it at its upper bound=%d; %d candidate operands",
		withoutGate, withGate, candidates)

	if candidates == 0 {
		t.Fatal("no function-expression operands exist on this corpus, so the zero below is about " +
			"the probe rather than about the gate")
	}
	// The gap's size, recorded as this assertion's own message asked once it stopped being zero.
	//
	// It was empty at 88 either way for as long as no function-expression operand named a value in a
	// scope that could overlap. The frozen-capture rule in `ranges.go` first exposed FileCarousel.
	// Propagating a frozen component parameter through Destructure then exposed InputSelect and
	// InputText. Each contributes one union, so the upper-bound probe now removes three: 118 without
	// the gate, 115 with it.
	//
	// Three is the UPPER bound rather than the real size. The probe skips every use-operand of every
	// `FunctionExpression`, which is strictly more than upstream's gate, since upstream skips only
	// the primitive ones. So the true divergence is at most three unions and may still be zero.
	//
	// The direction is unchanged and is the one already stated on `MergeGapPrimitiveOperandSkip`:
	// not skipping visits operands upstream skips, and visiting an operand can only ever ADD a
	// union, so this errs toward merging more than upstream rather than less. State freezing raises
	// the bound to seven, attributed to three input components in STATE_EFFECTS.md. Effect freezing
	// adds one AnimatedButton reference capture; EFFECT_HOOK_EFFECTS.md attributes the bound of eight.
	// Immutable alias-edge refinement removes the other seven, leaving AnimatedButton alone. The
	// full 135/127 to 139/138 union differential is recorded in IMMUTABLE_ALIAS_EDGES.md.
	const knownGateDivergence = 4
	if withoutGate-withGate != knownGateDivergence {
		t.Errorf("skipping every function-expression operand changed the union count from %d to "+
			"%d, a divergence of %d where %d was measured; MergeGapPrimitiveOperandSkip changed "+
			"size and the new one should be recorded", withoutGate, withGate,
			withoutGate-withGate, knownGateDivergence)
	}
}

// unionsSkippingFunctionOperands runs the sweep with every function-expression use-operand skipped.
func unionsSkippingFunctionOperands(function *Function, scopes *ReactiveScopes) int {
	if function == nil || scopes == nil || scopes.Len() == 0 {
		return 0
	}
	state := &mergeSweepState{rangeOf: scopes.RangeOf}
	state.starts, state.ends = collectScopeInfo(function, scopes)
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
			_, isFunction := instruction.Value.(*FunctionExpression)
			EachInstructionPlace(instruction, func(place Place, role PlaceRole) {
				if isFunction && role == PlaceRoleUse {
					return
				}
				state.visitPlace(instruction.Order, place, scopes, memberRanges)
			})
		}
		order := TerminalOrder(block.Terminal)
		state.visitInstructionId(order)
		EachTerminalPlace(block.Terminal, func(place Place, role PlaceRole) {
			state.visitPlace(order, place, scopes, memberRanges)
		})
	}
	return state.apply(scopes).Unions()
}

// TestMergeAssumesMonotoneEvaluationOrder pins an assumption this file makes about another's output.
//
// The stack discipline depends on evaluation order increasing along the block sequence: a scope that
// ends before the sweep reaches its end position would never leave `activeScopes` and would union
// with everything after it. `Function.Blocks` is documented as reverse postorder, which is what makes
// this hold, and it is asserted here rather than trusted because it is an assumption about a table
// `Construct` produces rather than about anything this file does.
func TestMergeAssumesMonotoneEvaluationOrder(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)
	functions, violations := 0, 0
	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		functions++
		var last static_single_assignment.EvaluationOrder
		for _, block := range function.Blocks {
			if block == nil {
				continue
			}
			for _, instructionId := range block.Instructions {
				instruction := function.Instructions[instructionId]
				if instruction == nil {
					continue
				}
				if instruction.Order < last {
					violations++
				}
				last = instruction.Order
			}
			if order := TerminalOrder(block.Terminal); order != 0 {
				if order < last {
					violations++
				}
				last = order
			}
		}
	})
	t.Logf("functions=%d non-monotone steps=%d", functions, violations)
	if functions == 0 {
		t.Fatal("no functions were walked, so this measured nothing")
	}
	if violations != 0 {
		t.Errorf("evaluation order steps backwards %d times across the block sequence; the sweep's "+
			"stack discipline is unsound on this graph", violations)
	}
}

// TestMergeIsDeterministic runs the pass twice over the corpus and compares.
//
// Two Go maps back this pass where upstream has insertion-ordered ones, so determinism is a property
// of `scopeSet.order` and of the id sort in `descendingByPosition` rather than something the
// structure provides. A non-deterministic merge produces a different representative for the same
// group on different runs, which is the shape that makes a cache non-reproducible without ever
// producing a wrong answer.
func TestMergeIsDeterministic(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)
	functions, mismatches, merges := 0, 0, 0
	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		functions++
		first := MergeOverlappingReactiveScopes(function, scopes)
		second := MergeOverlappingReactiveScopes(function, scopes)
		merges += first.Unions()
		if first.Len() != second.Len() || first.Unions() != second.Unions() {
			mismatches++
			return
		}
		for _, scope := range scopes.Ids() {
			if first.GroupOf(scope) != second.GroupOf(scope) {
				mismatches++
				return
			}
			if first.RangeOf(first.GroupOf(scope)) != second.RangeOf(second.GroupOf(scope)) {
				mismatches++
				return
			}
		}
	})
	t.Logf("functions=%d unions=%d mismatches=%d", functions, merges, mismatches)
	if merges == 0 {
		t.Fatal("nothing merged on the whole corpus, so running it twice compared two no-ops")
	}
	if mismatches != 0 {
		t.Errorf("%d functions merged differently on a second run", mismatches)
	}
}

// TestMergeSortComparatorReachability pins the two comparators' very different reachability.
//
// `sortByStartDescending` is load-bearing and `sharedEndReversedStack` covers it.
// `sortByEndDescending` is unreachable on this corpus because no two scopes start together, so a
// mutation reversing it survives correctly rather than through a fixture gap. That verdict expires
// the moment a pass able to mint coincident starts lands, and this test is what makes the expiry
// visible instead of silent.
func TestMergeSortComparatorReachability(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)
	startPositions, sharedStarts, endPositions, sharedEnds, widestEndGroup := 0, 0, 0, 0, 0

	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		starts, ends := collectScopeInfo(function, scopes)
		for _, group := range starts {
			startPositions++
			if len(group.scopes) > 1 {
				sharedStarts++
			}
		}
		for _, group := range ends {
			endPositions++
			if len(group.scopes) > 1 {
				sharedEnds++
			}
			if len(group.scopes) > widestEndGroup {
				widestEndGroup = len(group.scopes)
			}
		}
	})

	t.Logf("start positions=%d (%d hold several scopes); end positions=%d (%d hold several, "+
		"widest=%d)", startPositions, sharedStarts, endPositions, sharedEnds, widestEndGroup)

	if startPositions == 0 || endPositions == 0 {
		t.Fatal("no scope positions were collected, so this measured nothing")
	}
	if sharedEnds == 0 {
		t.Error("no end position holds several scopes, so sortByStartDescending is unreachable " +
			"too and sharedEndReversedStack has stopped covering anything")
	}
	if sharedStarts != 0 {
		t.Errorf("%d start positions now hold several scopes, so sortByEndDescending has become "+
			"reachable; its unreachability verdict in merge_scopes.go is void and a fixture "+
			"covering the stack-depth ordering is now required", sharedStarts)
	}
}

// TestMergeRegistersScopesFromTerminals pins why the terminal record in collectScopeInfo survives
// mutation, and asserts the mechanism directly rather than through the sweep's output.
//
// Deleting that line survives the fixture set because 0 of 3,306 corpus scopes are registered only
// through a terminal place. The synthetic half of this test asserts the line WORKS anyway, by
// building a scope whose sole appearance is a terminal operand and checking it reaches the tables.
func TestMergeRegistersScopesFromTerminals(t *testing.T) {
	t.Parallel()

	// The synthetic half: a scope that appears nowhere but the terminal must still be registered.
	function, scopes := buildMergeCase(t, mergeCase{
		scopes:        []mutation_aliasing.MutableRange{{Start: 1, End: 9}, {Start: 3, End: 12}},
		instructions:  [][3]int{{1, 1, 0}},
		terminal:      3,
		terminalScope: 2,
	})
	starts, ends := collectScopeInfo(function, scopes)

	registered := func(groups []scopesAtPosition, scope ScopeId) bool {
		for _, group := range groups {
			for _, candidate := range group.scopes {
				if candidate == scope {
					return true
				}
			}
		}
		return false
	}
	if !registered(starts, 2) || !registered(ends, 2) {
		t.Error("a scope whose only place is a terminal operand never reached the start or end " +
			"tables, so the sweep cannot see it at all")
	}
	// React's own answer for this input is that nothing merges: with one block the sweep never
	// reaches position 9, so scope 1 never closes and no union fires. Asserted so a change that
	// starts merging here is caught rather than read as an improvement.
	merged := MergeOverlappingReactiveScopes(function, scopes)
	if merged.Len() != 2 || merged.Unions() != 0 {
		t.Errorf("got %d scopes and %d unions, want 2 and 0; React leaves these separate",
			merged.Len(), merged.Unions())
	}

	// The corpus half: the measurement that explains the survivor.
	skipWithoutCorpus(t)
	onlyTerminal, total := 0, 0
	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		fromInstruction, fromTerminal := map[ScopeId]bool{}, map[ScopeId]bool{}
		for _, block := range function.Blocks {
			if block == nil {
				continue
			}
			for _, instructionId := range block.Instructions {
				instruction := function.Instructions[instructionId]
				if instruction == nil {
					continue
				}
				EachInstructionPlace(instruction, func(place Place, role PlaceRole) {
					if scope := scopes.ScopeOf(place.Identifier); scope != 0 {
						fromInstruction[scope] = true
					}
				})
			}
			EachTerminalPlace(block.Terminal, func(place Place, role PlaceRole) {
				if scope := scopes.ScopeOf(place.Identifier); scope != 0 {
					fromTerminal[scope] = true
				}
			})
		}
		total += scopes.Len()
		for scope := range fromTerminal {
			if !fromInstruction[scope] {
				onlyTerminal++
			}
		}
	})
	t.Logf("scopes=%d registered only through a terminal place=%d", total, onlyTerminal)
	if onlyTerminal != 0 {
		t.Errorf("%d scopes are now registered only through a terminal place, so deleting that "+
			"record would change an answer and the survivor verdict in merge_scopes.go is void",
			onlyTerminal)
	}
}
