package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
	"sort"
	"testing"
)

// tagged is a nesting item plus which list it came from, so a violation can be attributed to a kind
// rather than only counted.
type tagged struct {
	item    nestingItem
	isScope bool
}

// scopeRangeMap, mergedRangeMap and alignedRangeMap adapt the three result shapes to the one input
// `scopeNestingItems` takes, so the assertion is computed the same way for all three.
func scopeRangeMap(scopes *ReactiveScopes) map[ScopeId]mutation_aliasing.MutableRange {
	out := map[ScopeId]mutation_aliasing.MutableRange{}
	for _, id := range scopes.Ids() {
		out[id] = scopes.RangeOf(id)
	}
	return out
}

func mergedRangeMap(merged *MergedScopes) map[ScopeId]mutation_aliasing.MutableRange {
	out := map[ScopeId]mutation_aliasing.MutableRange{}
	for _, id := range merged.Ids() {
		out[id] = merged.RangeOf(id)
	}
	return out
}

func alignedRangeMap(aligned *AlignedScopes) map[ScopeId]mutation_aliasing.MutableRange {
	out := map[ScopeId]mutation_aliasing.MutableRange{}
	for _, id := range aligned.Ids() {
		out[id] = aligned.RangeOf(id)
	}
	return out
}

// violationsByKind repeats blockNestingViolations and attributes each violation to the pair of kinds
// that produced it.
//
// Necessary rather than cosmetic: the headline number this pass is measured on is a total, and a
// total cannot say whether a residue is the one this pass structurally cannot reach. The block
// against block count is that residue, and it is the control that proves this pass did not simply
// delete scopes to reach its target.
func violationsByKind(scopes, blocks []nestingItem) (scopeInBlock, blockInScope, blockInBlock int) {
	items := make([]tagged, 0, len(scopes)+len(blocks))
	for _, s := range scopes {
		items = append(items, tagged{s, true})
	}
	for _, b := range blocks {
		items = append(items, tagged{b, false})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].item.start != items[j].item.start {
			return items[i].item.start < items[j].item.start
		}
		return items[i].item.end > items[j].item.end
	})

	var active []tagged
	for _, current := range items {
		for i := len(active) - 1; i >= 0; i-- {
			parent := active[i]
			disjoint := current.item.start >= parent.item.end
			nested := current.item.end <= parent.item.end
			if !disjoint && !nested {
				switch {
				case current.isScope && !parent.isScope:
					scopeInBlock++
				case !current.isScope && parent.isScope:
					blockInScope++
				case !current.isScope && !parent.isScope:
					blockInBlock++
				}
			}
			if disjoint {
				active = active[:i]
			} else {
				break
			}
		}
		active = append(active, current)
	}
	return
}

// widenedTableFrom rebuilds a ReactiveScopes carrying the aligned ranges, which is what the merge
// reads when it runs behind this pass.
func widenedTableFrom(scopes *ReactiveScopes, aligned *AlignedScopes) *ReactiveScopes {
	widened := &ReactiveScopes{
		byIdentifier: scopes.byIdentifier,
		ranges:       make(map[ScopeId]mutation_aliasing.MutableRange, scopes.Len()),
		members:      scopes.members,
		order:        scopes.order,
	}
	for _, id := range scopes.Ids() {
		widened.ranges[id] = aligned.RangeOf(id)
	}
	return widened
}

// TestAlignClosesTheFullBlockNestingAssertion is the headline, and it asserts the assertion upstream
// actually runs rather than the scope-against-scope half of it.
//
// `assertValidBlockNesting` traverses scopes AND block subtrees together. Measuring only scopes
// against scopes reports zero for a graph that still stops upstream's compiler, which is why the
// merge shipped with 166 outstanding and this pass exists.
//
// The block-against-block count is carried alongside as the control. It is the one violation this
// pass structurally cannot reach, so it staying at 1 is what proves the drop from 166 came from
// widening scopes rather than from losing them.
func TestAlignClosesTheFullBlockNestingAssertion(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)
	functions, scopesBefore, scopesAfter := 0, 0, 0
	fullPre, fullMergeOnly, fullAligned := 0, 0, 0
	scopePre, scopeMergeOnly, scopeAligned := 0, 0, 0
	blockOnly, widened := 0, 0
	mergeOnlyScopeInBlock, mergeOnlyBlockInScope, mergeOnlyBlockInBlock := 0, 0, 0
	componentFunctions, componentScopeInBlock, componentBlockInScope := 0, 0, 0
	finalScopeInBlock, finalBlockInScope, finalBlockInBlock := 0, 0, 0

	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		functions++
		scopesBefore += scopes.Len()
		blocks := programBlockSubtrees(function)

		pre := scopeNestingItems(scopeRangeMap(scopes))
		fullPre += blockNestingViolations(pre, blocks)
		scopePre += blockNestingViolations(pre, nil)

		mergeOnly := MergeOverlappingReactiveScopes(function, scopes)
		mergeItems := scopeNestingItems(mergedRangeMap(mergeOnly))
		fullMergeOnly += blockNestingViolations(mergeItems, blocks)
		scopeMergeOnly += blockNestingViolations(mergeItems, nil)
		s, b, bb := violationsByKind(mergeItems, blocks)
		mergeOnlyScopeInBlock += s
		mergeOnlyBlockInScope += b
		mergeOnlyBlockInBlock += bb

		// The control for the residue: the same assertion with the scope items removed entirely.
		blockOnly += blockNestingViolations(nil, blocks)

		aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
		widened += aligned.Widened()
		scopesAfter += merged.Len()
		finalItems := scopeNestingItems(mergedRangeMap(merged))
		fullAligned += blockNestingViolations(finalItems, blocks)
		scopeAligned += blockNestingViolations(finalItems, nil)
		fs, fb, fbb := violationsByKind(finalItems, blocks)
		if function.Kind == FunctionKindComponent || function.Kind == FunctionKindHook {
			componentFunctions++
			componentScopeInBlock += fs
			componentBlockInScope += fb
		}
		finalScopeInBlock += fs
		finalBlockInScope += fb
		finalBlockInBlock += fbb
	})

	t.Logf("functions=%d scopes %d -> %d, scopes widened by alignment=%d",
		functions, scopesBefore, scopesAfter, widened)
	t.Logf("scope-vs-scope only   pre=%d  merge alone=%d  align+merge=%d",
		scopePre, scopeMergeOnly, scopeAligned)
	t.Logf("FULL assertion        pre=%d merge alone=%d  align+merge=%d",
		fullPre, fullMergeOnly, fullAligned)
	t.Logf("block-items-only control = %d", blockOnly)
	t.Logf("  merge alone: ScopeInBlock=%d BlockInScope=%d BlockInBlock=%d",
		mergeOnlyScopeInBlock, mergeOnlyBlockInScope, mergeOnlyBlockInBlock)
	t.Logf("  align+merge: ScopeInBlock=%d BlockInScope=%d BlockInBlock=%d",
		finalScopeInBlock, finalBlockInScope, finalBlockInBlock)

	if functions == 0 || scopesBefore == 0 {
		t.Fatal("no functions or no scopes were seen, so this measured nothing")
	}
	if fullMergeOnly <= fullAligned {
		t.Errorf("the merge alone leaves %d violations and align+merge leaves %d, so this pass "+
			"closed nothing", fullMergeOnly, fullAligned)
	}
	// The scope-involving part of the residue is asserted over components and hooks below, for the
	// reason recorded there. Here the whole-corpus residue is bounded rather than pinned: subtracting
	// the scope-involving count leaves exactly the violations this pass structurally cannot reach,
	// and that identity holds whichever population the scope violations come from.
	if fullAligned-finalScopeInBlock-finalBlockInScope != blockOnly {
		t.Errorf("the full assertion leaves %d violations of which %d involve a scope, against a "+
			"block-items-only control of %d; the non-scope residue should equal the control because "+
			"those are the violations this pass structurally cannot reach",
			fullAligned, finalScopeInBlock+finalBlockInScope, blockOnly)
	}
	// # The population is every function in the corpus, and upstream's is not
	//
	// This walks all 677 corpus functions. React's compiler only ever runs on components and hooks,
	// so a violation in a plain helper is a shape upstream never encounters and never had to make
	// its invariant hold for.
	//
	// That distinction was invisible while `DeclarationId` collisions were fusing scopes, and it
	// matters now: with the collisions fixed, the corpus reports one `ScopeInBlock` violation, in
	// `drawHighlightedCountryOutlines`, a canvas drawing helper. Measured over the same corpus split
	// by `Function.Kind`: 318 components and hooks, ZERO violations; 677 functions, one.
	//
	// So the assertion below is stated over components and hooks, which is upstream's own domain,
	// and the whole-corpus figure is logged rather than asserted. Asserting it would hold this pass
	// to an invariant its reference implementation does not claim.
	if componentFunctions < 100 {
		t.Fatalf("only %d components and hooks reached; the assertion below would be a fact about "+
			"the harness rather than about the pass", componentFunctions)
	}
	if componentScopeInBlock != 0 || componentBlockInScope != 0 {
		t.Errorf("scope-involving violations remain across %d components and hooks: "+
			"ScopeInBlock=%d BlockInScope=%d", componentFunctions,
			componentScopeInBlock, componentBlockInScope)
	}
	t.Logf("whole corpus: ScopeInBlock=%d BlockInScope=%d; components and hooks (%d): "+
		"ScopeInBlock=%d BlockInScope=%d", finalScopeInBlock, finalBlockInScope,
		componentFunctions, componentScopeInBlock, componentBlockInScope)
	if widened == 0 {
		t.Error("no scope was widened, so a closed assertion would mean it was already closed")
	}
}

// TestAlignMustRunBeforeTheMerge pins the ordering, which is the finding this pass cost the most to
// establish and the one a future refactor is most likely to undo.
//
// Our merge shipped first, so landing this behind it is the natural thing to do and it is wrong: the
// merge's output is not a fixpoint of the merge once ranges are widened underneath it. Both orders
// run here side by side so that the cost of getting it backwards is a number rather than a warning.
func TestAlignMustRunBeforeTheMerge(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)
	naiveFull, naiveScope, naiveUnions := 0, 0, 0
	upstreamFull, upstreamScope, upstreamUnions := 0, 0, 0

	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		blocks := programBlockSubtrees(function)

		// The naive order: merge first, then align the merged table. Expressed by handing the
		// merged ranges back through a scope table, which is the only way to run that order now that
		// the merge-aware entry point has been deleted; see AlignReactiveScopesToBlockScopes.
		merged := MergeOverlappingReactiveScopes(function, scopes)
		naiveUnions += merged.Unions()
		mergedTable := &ReactiveScopes{
			byIdentifier: scopes.byIdentifier,
			ranges:       map[ScopeId]mutation_aliasing.MutableRange{},
			members:      scopes.members,
			order:        merged.Ids(),
		}
		for _, id := range merged.Ids() {
			mergedTable.ranges[id] = merged.RangeOf(id)
		}
		naiveAligned := AlignReactiveScopesToBlockScopes(function, mergedTable)
		naiveItems := scopeNestingItems(alignedRangeMap(naiveAligned))
		naiveFull += blockNestingViolations(naiveItems, blocks)
		naiveScope += blockNestingViolations(naiveItems, nil)

		// Upstream's order: align first, then merge over the widened ranges.
		_, upstreamMerged := AlignThenMergeReactiveScopes(function, scopes)
		upstreamUnions += upstreamMerged.Unions()
		upstreamItems := scopeNestingItems(mergedRangeMap(upstreamMerged))
		upstreamFull += blockNestingViolations(upstreamItems, blocks)
		upstreamScope += blockNestingViolations(upstreamItems, nil)
	})

	t.Logf("merge then align (naive)     full=%d scope-vs-scope=%d unions=%d",
		naiveFull, naiveScope, naiveUnions)
	t.Logf("align then merge (upstream)  full=%d scope-vs-scope=%d unions=%d",
		upstreamFull, upstreamScope, upstreamUnions)

	if upstreamFull >= naiveFull {
		t.Errorf("upstream's order leaves %d violations and the naive order leaves %d, so the "+
			"ordering finding this pass rests on no longer holds", upstreamFull, naiveFull)
	}
	if naiveScope == 0 {
		t.Error("the naive order leaves no scope-against-scope violations, so the reason it is " +
			"wrong -- that widening creates overlaps the merge is no longer there to close -- has " +
			"stopped being observable and this test measures nothing")
	}
	if upstreamScope != 0 {
		t.Errorf("upstream's order leaves %d scope-against-scope violations, which the merge "+
			"running last is supposed to close", upstreamScope)
	}
	if upstreamUnions <= naiveUnions {
		t.Errorf("the merge performs %d unions behind alignment and %d in front of it; running "+
			"second is supposed to give it strictly more work", upstreamUnions, naiveUnions)
	}
}

// TestAlignMinimisesWithoutAGuard pins the adopt-versus-minimise answer, which has differed across
// three adjacent passes and was established here by running React rather than by reading it.
//
// The rows are the ones the bundle produced on hand-built graphs; see the package comment. The point
// of pinning them synthetically is that the corpus contains no unset scope, so the corpus cannot
// distinguish a guarded implementation from an unguarded one and a future reader adding a guard
// "for safety" would be making an unmeasured change that no corpus test could see.
func TestAlignMinimisesWithoutAGuard(t *testing.T) {
	t.Parallel()

	// A construct whose terminal sits at 5 and whose fallthrough begins at 9, so an active scope
	// starting after 5 has its start minimised back to 5 and its end pushed out to 9.
	build := func(scopeRange mutation_aliasing.MutableRange) (*Function, *ReactiveScopes) {
		return buildAlignCase(t, alignCase{
			scope:        scopeRange,
			terminalId:   5,
			useAt:        7,
			fallthroughs: 9,
		})
	}

	cases := []struct {
		name  string
		start mutation_aliasing.MutableRange
		want  mutation_aliasing.MutableRange
	}{
		{"start after the construct is minimised back", mutation_aliasing.MutableRange{Start: 7, End: 15}, mutation_aliasing.MutableRange{Start: 5, End: 15}},
		{"start already earlier is left alone", mutation_aliasing.MutableRange{Start: 3, End: 15}, mutation_aliasing.MutableRange{Start: 3, End: 15}},
		{"an unset start stays zero rather than being widened", mutation_aliasing.MutableRange{Start: 0, End: 15}, mutation_aliasing.MutableRange{Start: 0, End: 15}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function, scopes := build(testCase.start)
			aligned := AlignReactiveScopesToBlockScopes(function, scopes)
			if got := aligned.RangeOf(1); got != testCase.want {
				t.Errorf("scope %v aligned to %v, want %v", testCase.start, got, testCase.want)
			}
		})
	}

	// A fully unset scope never becomes active, because activity is `end > startingId` and an end of
	// zero fails that for every block. It must come back untouched rather than widened to the
	// construct, which is what an implementation missing the activity filter would produce.
	function, scopes := build(mutation_aliasing.MutableRange{Start: 0, End: 0})
	aligned := AlignReactiveScopesToBlockScopes(function, scopes)
	if got := aligned.RangeOf(1); got != (mutation_aliasing.MutableRange{Start: 0, End: 0}) {
		t.Errorf("a fully unset scope aligned to %v; it should never become active at all", got)
	}
}

// TestAlignExcludesBranchTerminals pins the `terminal.kind !== 'branch'` exclusion, measured against
// React itself, and the structural reason it matters in this tree.
//
// It is not a corner case here: `Branch` is the largest terminal population carrying a fallthrough,
// and every fallthrough our lowering reuses is reused by a branch. So the exclusion is what makes
// upstream's "unique fallthroughs" invariant hold on our graph.
func TestAlignExcludesBranchTerminals(t *testing.T) {
	t.Parallel()

	widened := func(makeTerminal func(fallthroughBlock static_single_assignment.BlockId) Terminal) mutation_aliasing.MutableRange {
		function, scopes := buildAlignCaseWithTerminal(t, alignCase{
			scope:        mutation_aliasing.MutableRange{Start: 4, End: 6},
			terminalId:   5,
			useAt:        4,
			fallthroughs: 9,
		}, makeTerminal)
		return AlignReactiveScopesToBlockScopes(function, scopes).RangeOf(1)
	}

	ifRange := widened(func(ft static_single_assignment.BlockId) Terminal {
		return &If{Consequent: 2, Alternate: ft, Fallthrough: ft, Order: 5}
	})
	if ifRange != (mutation_aliasing.MutableRange{Start: 4, End: 9}) {
		t.Errorf("an `if` widened to %v, want [4,9); the fallthrough push did not fire", ifRange)
	}

	branchRange := widened(func(ft static_single_assignment.BlockId) Terminal {
		return &Branch{Consequent: 2, Alternate: ft, Fallthrough: ft, Order: 5}
	})
	if branchRange != (mutation_aliasing.MutableRange{Start: 4, End: 6}) {
		t.Errorf("a `branch` widened to %v, want [4,6) unchanged; upstream excludes branch "+
			"terminals from the fallthrough widening", branchRange)
	}
}

// TestAlignPushFilterDeclinesAScopeThatEndedBeforeTheTerminal covers the `end > terminal.id` test on
// the fallthrough push, which is a FILTER rather than a cost optimisation.
//
// A scope reaches that line only if it was still active at the block's first instruction, so the
// distinguishing input is a scope that is active at the top of the block and has ENDED by the time
// the sweep reaches the terminal. Without the filter such a scope is dragged forward across a
// construct it never entered.
//
// The corpus alone could not see this: the mutant removing the filter survived the whole suite, and
// only instrumenting the sweep showed the branch declines 876 of 3,269 candidates with every one of
// them a case that would otherwise have widened. That is a fixture blind spot rather than an
// equivalence, and this is the fixture.
func TestAlignPushFilterDeclinesAScopeThatEndedBeforeTheTerminal(t *testing.T) {
	t.Parallel()

	// The scope reads at 2 and ends at 3, so it is active when the block begins (its end exceeds the
	// block's first instruction at 2) and dead by the terminal at 5. The fallthrough begins at 9.
	function, scopes := buildAlignCase(t, alignCase{
		scope:        mutation_aliasing.MutableRange{Start: 2, End: 3},
		terminalId:   5,
		useAt:        2,
		fallthroughs: 9,
	})
	aligned := AlignReactiveScopesToBlockScopes(function, scopes)
	if got := aligned.RangeOf(1); got != (mutation_aliasing.MutableRange{Start: 2, End: 3}) {
		t.Errorf("a scope that ended at 3, before the terminal at 5, aligned to %v; the "+
			"`end > terminal` filter should have declined it rather than dragging its end out to "+
			"the fallthrough", got)
	}
}

// TestAlignRecordsAScopeOnceCoversTheSeenGate pins the `seen` gate on the value-block widening.
//
// The gate is what makes the fourth mechanism fire once per SCOPE rather than once per place. It is
// heavily exercised on the corpus -- 47,476 places arrive with their scope already seen -- and the
// mutant removing it still survived the corpus suite, because widening to the same value range twice
// is idempotent. The distinguishing input needs the scope to meet TWO DIFFERENT value ranges, so
// that applying the second one is not a no-op.
//
// Built synthetically rather than found, because that shape does not occur in the corpus: a scope
// whose first sighting is inside a narrow value block and whose second is inside a wider one.
func TestAlignRecordsAScopeOnceCoversTheSeenGate(t *testing.T) {
	t.Parallel()

	_, scopes := buildAlignCase(t, alignCase{
		scope:        mutation_aliasing.MutableRange{Start: 4, End: 6},
		terminalId:   5,
		useAt:        4,
		fallthroughs: 9,
	})

	// Two value-block nodes of different spans, both reached by the same scope. Only the FIRST may
	// be applied; the second must be ignored because the scope has already been seen.
	state := &alignSweepState{
		activeSet:       map[ScopeId]bool{},
		seen:            map[ScopeId]bool{},
		valueBlockNodes: map[static_single_assignment.BlockId]*valueBlockNode{},
		ranges:          map[ScopeId]mutation_aliasing.MutableRange{1: {Start: 4, End: 6}},
		scopes:          scopes,
	}
	scopedPlace := Place{Identifier: 1}
	for id, scope := range scopes.byIdentifier {
		if scope == 1 {
			scopedPlace = Place{Identifier: id}
			break
		}
	}

	narrow := &valueBlockNode{start: 3, end: 7}
	wide := &valueBlockNode{start: 1, end: 30}
	state.recordPlace(4, scopedPlace, narrow)
	afterFirst := state.ranges[1]
	state.recordPlace(5, scopedPlace, wide)
	afterSecond := state.ranges[1]

	if afterFirst != (mutation_aliasing.MutableRange{Start: 3, End: 7}) {
		t.Fatalf("the first sighting widened the scope to %v, want [3,7); the value-block "+
			"mechanism did not fire and nothing below measures the gate", afterFirst)
	}
	if afterSecond != afterFirst {
		t.Errorf("a second sighting widened the scope again, from %v to %v; the `seen` gate is "+
			"supposed to make this mechanism fire once per scope rather than once per place",
			afterFirst, afterSecond)
	}
}

// TestAlignGotoWidensAcrossSeveralOpenConstructs covers the goto back-reference, the third mechanism.
//
// A goto whose target is an open fallthrough that is NOT the innermost one is leaving several
// constructs at once, and every still-active scope has to be widened at BOTH ends to span the outer
// construct. When the target IS the innermost open construct the pop on the next block handles it,
// which is why upstream tests for it and why removing that test is not obviously wrong.
//
// This is a genuine rarity rather than a fabricated case: instrumenting the sweep found 2,642 gotos
// targeting an open fallthrough and exactly ONE of them reaching a non-innermost target. So the
// corpus exercises the branch once and no assertion over totals could see it move, which is why this
// fixture is synthetic and the measurement is recorded here.
func TestAlignGotoWidensAcrossSeveralOpenConstructs(t *testing.T) {
	t.Parallel()

	function, scopes := buildNestedGotoCase(t)
	aligned := AlignReactiveScopesToBlockScopes(function, scopes)

	// The scope reads at 21, inside the inner construct, and both constructs are open when the goto
	// at 22 jumps past the inner fallthrough to the outer one. Its start must be pulled back to the
	// OUTER construct's terminal at 5, and its end pushed out to the outer fallthrough at 40.
	got := aligned.RangeOf(1)
	if got.Start != 5 {
		t.Errorf("the scope's start is %d, want 5; the goto back-reference should have pulled it "+
			"back to the outer construct's terminal", got.Start)
	}
	if got.End != 40 {
		t.Errorf("the scope's end is %d, want 40; the goto back-reference should have pushed it "+
			"out to the outer construct's fallthrough", got.End)
	}
}

// TestAlignGotoSkipsTheInnermostOpenConstruct covers the other half of the goto back-reference: the
// `start !== activeBlockFallthroughRanges.at(-1)` test that DECLINES a goto to the innermost target.
//
// The first fixture for this mutant was wrong even though the mechanism was understood, which is the
// hazard the port brief names: a fixture written for a survivor is a hypothesis. A goto to a
// NON-innermost target takes the same path with the test and without it, so it cannot see the
// mutation at all, and the mutant survived a re-score.
//
// What separates them is that the fallthrough pop on the following block reproduces the START
// widening and not the END one. So the distinguishing input is a scope active at an
// innermost-targeting goto whose end falls SHORT of that fallthrough: upstream leaves it alone and
// the pop then widens only its start, while removing the test drags its end forward too.
func TestAlignGotoSkipsTheInnermostOpenConstruct(t *testing.T) {
	t.Parallel()

	function, scopes := buildInnermostGotoCase(t)
	aligned := AlignReactiveScopesToBlockScopes(function, scopes)

	// Run against React 7.1.1's own `alignReactiveScopesToBlockScopesHIR` on this exact graph, which
	// is how the expectation below was obtained. The first version of this test expected the start
	// to move to 5 as well, on the reasoning that the pop would still fire. React says otherwise and
	// the reason is the activity filter: the scope's end is 9 and the continuation block begins at
	// 20, so the scope is retired before the pop is reached and neither endpoint moves.
	//
	// That is worth keeping as a comment rather than only as an assertion. The pop and the goto
	// back-reference look like two chances to widen the same scope, and they are not: a scope narrow
	// enough to be skipped by the goto test is usually also narrow enough to be retired before the
	// pop.
	if got := aligned.RangeOf(1); got != (mutation_aliasing.MutableRange{Start: 7, End: 9}) {
		t.Errorf("the scope aligned to %v, want [7,9) unchanged; a goto to the INNERMOST open "+
			"fallthrough is skipped, and this scope is retired by the activity filter before the "+
			"fallthrough pop can reach it", got)
	}
}

// TestAlignDeclinesAPlaceReadOutsideItsScope covers upstream's `getPlaceScope` containment test.
//
// It is not subsumed by the activity filter, which is the plausible reading and the reason the
// mutant removing it survived at first. The filter retires a scope whose END has passed the current
// block; this declines a place whose read position its own scope never covered, which can happen at
// any point.
//
// Choosing the range is the whole difficulty and the first attempt was wrong. A scope of [2,4) read
// at 7 also comes back unchanged, and it CANNOT see this line: its end has already passed the
// block's first instruction, so the activity filter would have excluded it anyway and both spellings
// reach the same verdict by different routes. That fixture passed against the mutant.
//
// The range has to sit AFTER the read instead. A scope of [9,12) read at 7 has an end beyond the
// block start, so nothing retires it, and only the containment test declines it. The expectation is
// React's own output on that graph rather than a derivation.
func TestAlignDeclinesAPlaceReadOutsideItsScope(t *testing.T) {
	t.Parallel()

	function, scopes := buildAlignCase(t, alignCase{
		scope:        mutation_aliasing.MutableRange{Start: 9, End: 12},
		terminalId:   8,
		useAt:        7,
		fallthroughs: 20,
	})
	aligned := AlignReactiveScopesToBlockScopes(function, scopes)
	if got := aligned.RangeOf(1); got != (mutation_aliasing.MutableRange{Start: 9, End: 12}) {
		t.Errorf("a scope of [9,12) read at 7, before its own range opens, aligned to %v; the "+
			"place should never have activated it", got)
	}
}

// TestAlignFallthroughsAreUniqueOnceBranchesAreExcluded measures the structural claim the branch
// exclusion rests on, over the real corpus.
//
// Upstream raises "Expect hir blocks to have unique fallthroughs" inside the non-branch guard. If a
// non-branch terminal ever reused a fallthrough here, that invariant would be violated on our graph
// and the value-block tree could be assigned twice. The zero is the measurement; the non-zero branch
// count beside it is the control that proves the probe fired.
func TestAlignFallthroughsAreUniqueOnceBranchesAreExcluded(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)
	branchReuse, nonBranchReuse, branchesWithFallthrough := 0, 0, 0

	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		seen := map[static_single_assignment.BlockId]bool{}
		for _, block := range function.Blocks {
			if block == nil {
				continue
			}
			fallthroughId, ok := Fallthrough(block.Terminal)
			if !ok {
				continue
			}
			_, isBranch := block.Terminal.(*Branch)
			if isBranch {
				branchesWithFallthrough++
			}
			if seen[fallthroughId] {
				if isBranch {
					branchReuse++
				} else {
					nonBranchReuse++
				}
			}
			seen[fallthroughId] = true
		}
	})

	t.Logf("branch terminals carrying a fallthrough=%d; fallthrough reuses: branch=%d non-branch=%d",
		branchesWithFallthrough, branchReuse, nonBranchReuse)

	if branchesWithFallthrough == 0 || branchReuse == 0 {
		t.Fatal("no branch terminal reuses a fallthrough, so this measured nothing and the " +
			"exclusion's justification in align_scopes.go has lost its evidence")
	}
	if nonBranchReuse != 0 {
		t.Errorf("%d fallthroughs are reused by a NON-branch terminal, so upstream's unique "+
			"fallthrough invariant no longer holds here and the value-block tree can be assigned "+
			"twice for one block", nonBranchReuse)
	}
}

// TestAlignReDerivesMemberRanges is `ScopeGapPostAlignmentWidening` coming due, measured.
//
// `AssignReactiveScopes` records that React aliases every member's range to the scope's range object
// so a later widening propagates. That gap was dormant precisely because nothing widened a scope
// after it. This pass does, so the stale table and the re-derived one must now differ, and the
// direction of the difference is the check that matters: this pass only ever widens.
func TestAlignReDerivesMemberRanges(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)
	members, differing, wider, narrower := 0, 0, 0, 0

	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		aligned := AlignReactiveScopesToBlockScopes(function, scopes)
		stale, fresh := scopes.MemberRanges(), aligned.MemberRanges()
		for _, scope := range scopes.Ids() {
			for _, id := range scopes.MembersOf(scope) {
				members++
				before, after := stale.Get(id), fresh.Get(id)
				if before == after {
					continue
				}
				differing++
				if after.Start <= before.Start && after.End >= before.End {
					wider++
				} else {
					narrower++
				}
			}
		}
	})

	t.Logf("members=%d, ranges that moved after alignment=%d (wider=%d narrower=%d)",
		members, differing, wider, narrower)

	if members == 0 {
		t.Fatal("no members were seen, so this measured nothing")
	}
	if differing == 0 {
		t.Error("no member range moved, so either this pass widened nothing or MemberRanges is " +
			"not re-derived from the widened table; either way the gap is still open")
	}
	if narrower != 0 {
		t.Errorf("%d member ranges got NARROWER; this pass only ever widens, so a narrowing means "+
			"the re-derivation is reading the wrong table", narrower)
	}
}

// TestAlignPreservesScopeWidthAndMembership is Stage 4's cross-tabulation re-run behind this pass.
//
// It is the over-merge control. A pass that reached a closed assertion by collapsing every scope
// into one span would satisfy the headline number and would show up here as multi-member scopes of
// width one, which is the cell that must stay empty. Both populations are asserted non-empty as
// well, because a table with one population missing measures a collapse rather than a structure.
func TestAlignPreservesScopeWidthAndMembership(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)
	oneNarrow, oneWide, manyNarrow, manyWide := 0, 0, 0, 0
	var widths []int

	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, table *ReactiveScopes) {
		_, merged := AlignThenMergeReactiveScopes(function, table)
		for _, id := range merged.Ids() {
			scopeRange := merged.RangeOf(id)
			width := int(scopeRange.End) - int(scopeRange.Start)
			widths = append(widths, width)
			switch {
			case len(merged.MembersOf(id)) <= 1 && width <= 1:
				oneNarrow++
			case len(merged.MembersOf(id)) <= 1:
				oneWide++
			case width <= 1:
				manyNarrow++
			default:
				manyWide++
			}
		}
	})

	sort.Ints(widths)
	at := func(p float64) int { return widths[int(float64(len(widths)-1)*p)] }
	t.Logf("scopes=%d  width p50=%d p90=%d p99=%d max=%d",
		len(widths), at(0.50), at(0.90), at(0.99), widths[len(widths)-1])
	t.Logf("                width<=1   width>1")
	t.Logf("  one member     %6d    %6d", oneNarrow, oneWide)
	t.Logf("  many members   %6d    %6d", manyNarrow, manyWide)

	if manyNarrow != 0 {
		t.Errorf("%d scopes hold several values within a single instruction, which is what "+
			"catastrophic over-merging looks like", manyNarrow)
	}
	if manyWide == 0 || oneNarrow == 0 {
		t.Error("one of the two populations vanished, so this measured a collapse rather than a " +
			"structure and the empty cell above proves nothing")
	}
}

// TestAlignVoidsTheMergesComparatorVerdicts is the expiry `merge_scopes.go` predicted, measured.
//
// That file records `sortByEndDescending` as unreachable because every start position held exactly
// one scope, and `sameStartSameEnd` as unreachable because no two scopes shared both endpoints. It
// names this pass as exactly the producer that would void both, and says whoever ports it must
// re-take the measurement rather than inherit the zero. This is that measurement.
//
// The union COUNT is identical in both spellings, which is why this compares groups and ranges: a
// count-only assertion would report the reversed comparator as still equivalent and inherit a
// verdict that has actually expired.
func TestAlignVoidsTheMergesComparatorVerdicts(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)
	sharedStarts, widestStart, coincidentPairs := 0, 0, 0
	alignedDiffering, unalignedDiffering := 0, 0

	differs := func(a, b *MergedScopes) bool {
		if a.Unions() != b.Unions() || a.Len() != b.Len() {
			return true
		}
		for _, id := range a.Ids() {
			if a.RangeOf(id) != b.RangeOf(id) || a.GroupOf(id) != b.GroupOf(id) {
				return true
			}
		}
		return false
	}

	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		// The control: on the UNALIGNED table the two spellings must agree, which is what makes a
		// difference on the aligned table attributable to the comparator rather than to the local
		// copy of the sweep having drifted from the real one.
		if differs(MergeOverlappingReactiveScopes(function, scopes),
			mergeWithReversedEndOrder(function, scopes)) {
			unalignedDiffering++
		}

		aligned := AlignReactiveScopesToBlockScopes(function, scopes)
		widened := widenedTableFrom(scopes, aligned)

		starts, _ := collectScopeInfo(function, widened)
		for _, group := range starts {
			if len(group.scopes) > 1 {
				sharedStarts++
			}
			if len(group.scopes) > widestStart {
				widestStart = len(group.scopes)
			}
			for i := 1; i < len(group.scopes); i++ {
				if widened.RangeOf(group.scopes[i-1]).End == widened.RangeOf(group.scopes[i]).End {
					coincidentPairs++
				}
			}
		}

		if differs(MergeOverlappingReactiveScopes(function, widened),
			mergeWithReversedEndOrder(function, widened)) {
			alignedDiffering++
		}
	})

	t.Logf("over ALIGNED scopes: start positions holding several scopes=%d (widest=%d); "+
		"pairs sharing both endpoints=%d", sharedStarts, widestStart, coincidentPairs)
	t.Logf("reversing sortByEndDescending changes the answer in %d functions (control, unaligned: %d)",
		alignedDiffering, unalignedDiffering)

	if unalignedDiffering != 0 {
		t.Fatalf("the reversed spelling already disagrees on %d UNALIGNED functions, so it has "+
			"drifted from the real sweep and nothing below is attributable to the comparator",
			unalignedDiffering)
	}
	if sharedStarts == 0 {
		t.Error("no start position holds several scopes even after alignment, so " +
			"sortByEndDescending is still unreachable and merge_scopes.go's verdict stands")
	}
	if coincidentPairs == 0 {
		t.Error("no two scopes share both endpoints after alignment, so the sameStartSameEnd " +
			"branch is still unreachable on this corpus")
	}
	if alignedDiffering == 0 {
		t.Error("reversing sortByEndDescending changes no answer even over aligned scopes, so it " +
			"remains equivalent here and the expiry has not actually arrived")
	}
}

// TestReactReversePostorderClosesFallthroughSelfNesting pins the retired graph-ordering gap.
//
// The old reverse-postorder walk left one block subtree improperly nested in its own fallthrough.
// React's traversal visits the fallthrough first while constructing postorder, and the block-only
// control falls to zero under that order. Scope-involving residues in plain helpers are measured by
// the sibling test and are deliberately not attributed to this retired gap.
func TestReactReversePostorderClosesFallthroughSelfNesting(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)
	withScopes, withoutScopes := 0, 0

	// The whole corpus, deliberately: this is a graph invariant rather than a React-domain metric.
	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		blocks := programBlockSubtrees(function)
		_, merged := AlignThenMergeReactiveScopes(function, scopes)
		withScopes += blockNestingViolations(scopeNestingItems(mergedRangeMap(merged)), blocks)
		withoutScopes += blockNestingViolations(nil, blocks)
	})

	t.Logf("full assertion after align+merge=%d; with scope items removed entirely=%d",
		withScopes, withoutScopes)

	if withoutScopes != 0 {
		t.Errorf("block-only nesting violations=%d, want 0; the retired fallthrough-ordering gap reopened", withoutScopes)
	}
}

// TestAlignGapsAreDeclared makes closing a gap a visible event rather than a silent improvement.
func TestAlignGapsAreDeclared(t *testing.T) {
	t.Parallel()

	gaps := AlignGaps()
	if len(gaps) != 0 {
		t.Errorf("AlignGaps returned %v; a change here means a gap opened or closed and the "+
			"reasoning in align_scopes.go needs to move with it", gaps)
	}
}

// TestAlignIsASingleSweep pins termination empirically rather than resting on the argument alone.
//
// The claim is that every block is visited exactly once, so the visit count must equal the block
// count. A pass that iterated to a fixpoint would exceed it, which is the failure this catches.
//
// The denominator is blocks in functions that HAVE scopes, not blocks overall, and getting that
// wrong is how this test first failed. A function with no scopes returns before the sweep begins, so
// counting its blocks measures the early return rather than the sweep. Measured on the corpus: 49 of
// 677 functions have no scope at all, carrying 217 blocks between them, which is exactly the gap.
// Both denominators are reported so that a future reader sees the distinction rather than
// rediscovering it.
func TestAlignIsASingleSweep(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)
	blocksOverall, blocksWithScopes, visits, functionsWithoutScopes := 0, 0, 0, 0

	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		blocks := 0
		for _, block := range function.Blocks {
			if block != nil {
				blocks++
			}
		}
		blocksOverall += blocks
		if scopes.Len() == 0 {
			functionsWithoutScopes++
		} else {
			blocksWithScopes += blocks
		}
		visits += AlignReactiveScopesToBlockScopes(function, scopes).visits
	})

	t.Logf("blocks overall=%d, blocks in functions with scopes=%d, visits=%d (%d functions have "+
		"no scope and return before the sweep)",
		blocksOverall, blocksWithScopes, visits, functionsWithoutScopes)

	if blocksWithScopes == 0 {
		t.Fatal("no blocks were seen, so this measured nothing")
	}
	if visits != blocksWithScopes {
		t.Errorf("visited %d blocks over %d, so this pass is not a single sweep",
			visits, blocksWithScopes)
	}
	if functionsWithoutScopes == 0 {
		t.Error("every function has a scope, so the early return this test's denominator accounts " +
			"for is unreachable and the two denominators have stopped differing")
	}
}

// TestAlignIsDeterministic runs the pass twice over the corpus and compares.
//
// Two Go maps back this pass where upstream has insertion-ordered structures, so a nondeterminism
// here would be the shape that makes a cache non-reproducible without ever producing a wrong answer.
func TestAlignIsDeterministic(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)
	compared := 0

	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		first := AlignReactiveScopesToBlockScopes(function, scopes)
		second := AlignReactiveScopesToBlockScopes(function, scopes)
		if first.Len() != second.Len() || first.Widened() != second.Widened() {
			t.Fatalf("two runs disagree on size: %d/%d scopes, %d/%d widened",
				first.Len(), second.Len(), first.Widened(), second.Widened())
		}
		for _, id := range first.Ids() {
			if first.RangeOf(id) != second.RangeOf(id) {
				t.Fatalf("scope %d aligned to %v then %v", id, first.RangeOf(id), second.RangeOf(id))
			}
			compared++
		}
	})

	t.Logf("compared %d scope ranges across two runs", compared)
	if compared == 0 {
		t.Fatal("no scopes were compared, so this measured nothing")
	}
}

// TestAlignHandlesAnEmptyScopeTable is the ordinary answer for a function with no entangled values.
func TestAlignHandlesAnEmptyScopeTable(t *testing.T) {
	t.Parallel()

	function := &Function{}
	if got := AlignReactiveScopesToBlockScopes(function, &ReactiveScopes{}); got.Len() != 0 {
		t.Errorf("an empty scope table produced %d scopes", got.Len())
	}
	if got := AlignReactiveScopesToBlockScopes(nil, nil); got.Len() != 0 {
		t.Errorf("nil inputs produced %d scopes", got.Len())
	}
	aligned, merged := AlignThenMergeReactiveScopes(nil, nil)
	if aligned.Len() != 0 || merged.Len() != 0 {
		t.Errorf("nil inputs through the composed spelling produced %d aligned and %d merged",
			aligned.Len(), merged.Len())
	}
}
