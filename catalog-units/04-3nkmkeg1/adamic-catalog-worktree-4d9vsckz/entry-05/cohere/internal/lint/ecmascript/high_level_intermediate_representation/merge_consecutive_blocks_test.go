package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/static_single_assignment"
	"testing"
)

// TestMergeCollapsesTheRegionTheSpliceCreated is the pass's reason for existing.
//
// The splice cuts one block into three joined by plain gotos. Nothing about that region is a real
// branch, so leaving it split gives block structure to straight-line code -- and several passes in
// this package key on block structure. The assertion is that the region is one block again and that
// every instruction survived the join in order.
func TestMergeCollapsesTheRegionTheSpliceCreated(t *testing.T) {
	t.Parallel()

	function, inlined := inlinedFixture(t, singleReturnIifeSource, false)
	if function == nil || inlined != 1 {
		t.Fatalf("the fixture spliced %d calls, want 1", inlined)
	}
	if len(function.Blocks) < 2 {
		t.Fatalf("the splice left %d block(s); with one there is nothing to merge and this test "+
			"asserts nothing", len(function.Blocks))
	}

	var before []InstructionId
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		before = append(before, block.Instructions...)
	}

	merged := MergeConsecutiveBlocks(function)
	if merged == 0 {
		t.Fatal("nothing merged; the spliced region is joined by plain gotos with one predecessor " +
			"each, which is exactly the shape this pass collapses")
	}
	if len(function.Blocks) != 1 {
		t.Errorf("the function has %d blocks after merging, want 1; the spliced region is "+
			"straight-line code and should be one block", len(function.Blocks))
	}

	var after []InstructionId
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		after = append(after, block.Instructions...)
	}
	if len(after) != len(before) {
		t.Errorf("the merge changed the instruction count from %d to %d; a merge moves "+
			"instructions and must not drop or duplicate any", len(before), len(after))
	}
	for index := range after {
		if index < len(before) && after[index] != before[index] {
			t.Errorf("instruction %d of the merged block is %d, was %d at that position before; "+
				"the merge reordered evaluation", index, after[index], before[index])
			break
		}
	}
}

// TestMergeLeavesTheGraphResolvable pins that no reference outlives the block it names.
//
// A merged-away block is deleted from the id index, so every terminal that named it and every
// fallthrough that pointed at it has to be repointed. A missed one is not a crash: `Block` returns
// false and passes skip the edge, so the function analyses as though the edge were absent.
//
// # This runs on the SINGLE-return fixture, and the reason is a measurement
//
// It was written first against the multi-return one, on the assumption that a bigger graph is a
// better test. Measured, that fixture merges NOTHING, and correctly so: the labeled form's cut
// block ends in a `Label` rather than a `Goto`, the label's target is a fallthrough-protected
// block, and the continuation has two predecessors because both returns jump to it. All three of
// upstream's conditions decline it, which is the pass working rather than failing.
//
// So the merge only has something to do on the single-return path -- and that is also the path
// where it matters, because that is the one that gives straight-line code block boundaries it did
// not have.
//
// The CALLER branches, which is the second thing this fixture needs. A caller that does not branch
// merges all the way down to one block ending in a `Return`, and a `Return` names no block -- so
// the repointing this test is about has nothing left to check and the vacuity guard at the bottom
// fires. The branch leaves real references behind for the merge to get wrong.
func TestMergeLeavesTheGraphResolvable(t *testing.T) {
	t.Parallel()

	const source = `
		function Component(properties: {items: Array<number>; ready: boolean}) {
			if (!properties.ready) {
				return null;
			}
			const built = (() => {
				const out = [];
				out.push(properties.items);
				return out;
			})();
			return [built];
		}
	`
	function, inlined := inlinedFixture(t, source, false)
	if function == nil || inlined != 1 {
		t.Fatalf("the fixture spliced %d calls, want 1", inlined)
	}
	if MergeConsecutiveBlocks(function) == 0 {
		t.Fatal("nothing merged, so the repointing this test guards never ran")
	}

	held := map[static_single_assignment.BlockId]bool{}
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		held[block.Id] = true
		if _, found := function.Block(block.Id); !found {
			t.Errorf("block %d is in Blocks but was deleted from the id index", block.Id)
		}
	}
	if !held[function.Entry] {
		t.Errorf("the entry names block %d, which the function no longer holds", function.Entry)
	}

	references := 0
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		EachBlockReferencePointer(block.Terminal, func(reference *static_single_assignment.BlockId) {
			references++
			if !held[*reference] {
				t.Errorf("block %d names block %d, which was merged away and not repointed",
					block.Id, *reference)
			}
		})
		if target, has := Fallthrough(block.Terminal); has && HasBlock(target) && !held[target] {
			t.Errorf("block %d has a fallthrough naming block %d, which was merged away; "+
				"fallthroughs are not edges, so MarkPredecessors does not repair them",
				block.Id, target)
		}
		for _, phi := range block.Phis {
			for _, entry := range phi.Operands {
				if !held[entry.Predecessor] {
					t.Errorf("the phi in block %d takes an operand from block %d, which was "+
						"merged away", block.Id, entry.Predecessor)
				}
			}
		}
	}
	if references == 0 {
		t.Error("no terminal names a block, so this test asserted nothing")
	}
}

// TestMergeLeavesTheLabeledSpliceAlone pins the measurement the test above records.
//
// The labeled form is a real branch, not straight-line code given boundaries: two returns jump to
// one continuation, so that continuation is a join. A merge pass that collapsed it would fold one
// arm's instructions into the block before the label and run them unconditionally. Asserted
// directly, because "nothing merged" is otherwise indistinguishable from a pass that did not run.
func TestMergeLeavesTheLabeledSpliceAlone(t *testing.T) {
	t.Parallel()

	function, inlined := inlinedFixture(t, multipleReturnIifeSource, false)
	if function == nil || inlined != 1 {
		t.Fatalf("the fixture spliced %d calls, want 1", inlined)
	}
	blocks := len(function.Blocks)
	joins := 0
	for _, block := range function.Blocks {
		if block != nil && len(block.Predecessors) > 1 {
			joins++
		}
	}
	if joins == 0 {
		t.Fatal("the labeled splice produced no join block, so the two returns are not both " +
			"reaching the continuation and this test asserts nothing")
	}

	if merged := MergeConsecutiveBlocks(function); merged != 0 {
		t.Errorf("merged %d blocks out of the labeled splice, want 0; its continuation is a real "+
			"join and folding it would run one arm's instructions unconditionally", merged)
	}
	if len(function.Blocks) != blocks {
		t.Errorf("the block count moved from %d to %d", blocks, len(function.Blocks))
	}
}

// TestMergeLeavesAJoinBlockAlone is the guard on the single-predecessor condition.
//
// A join is a block several edges arrive at. Folding one into a single predecessor makes the other
// predecessors' control flow skip it entirely, and makes the join's instructions run
// unconditionally on the one path that absorbed them. The graph still builds.
//
// # BOTH arms are spliced, and getting to that fixture took two measurements
//
// Written first against a plain `if`/`else`, this could not see the mutation that removes the
// single-predecessor condition. Measured: that fixture merges NOTHING either way, because the
// join's predecessors are the two arms and each arm's own predecessor is the `If` block -- the pass
// declines on "the predecessor must end in a `Goto`" long before it reaches the count, so the
// mutant's extra merges never happen.
//
// Splicing into ONE arm was the second attempt and it also survived, for a subtler version of the
// same reason: the join's predecessors are then the spliced arm's tail and the `If` block itself,
// and the `If` is still what the `Goto` condition rejects. Measured output was byte-identical with
// and without the mutant.
//
// Splicing into BOTH arms is what discriminates. Every predecessor of the join then ends in a
// `Goto`, so the count is the only condition left standing between the pass and swallowing the
// join -- which is the point: a guard is only tested by an input where it is the guard that
// decides.
func TestMergeLeavesAJoinBlockAlone(t *testing.T) {
	t.Parallel()

	const source = `
		function Component(properties: {flag: boolean; items: Array<number>}) {
			let out = [];
			if (properties.flag) {
				out = (() => {
					const inner = [];
					inner.push(properties.items);
					return inner;
				})();
			} else {
				out = (() => {
					const other = [];
					other.push(1);
					return other;
				})();
			}
			return [out];
		}
	`
	function, inlined := inlinedFixture(t, source, false)
	if function == nil || inlined != 2 {
		t.Fatalf("the fixture spliced %d calls, want 2; both arms must be spliced or the join's "+
			"predecessors do not all end in a Goto and the count condition is not what decides",
			inlined)
	}

	joins := map[static_single_assignment.BlockId]int{}
	branches := 0
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		if len(block.Predecessors) > 1 {
			joins[block.Id] = len(block.Predecessors)
		}
		if _, isIf := block.Terminal.(*If); isIf {
			branches++
		}
	}
	if len(joins) == 0 {
		t.Fatal("the fixture produced no join block, so the condition this test guards is never " +
			"reached and the assertions below pass vacuously")
	}
	if branches == 0 {
		t.Fatal("the fixture lowered no branch")
	}
	// Every predecessor of every join must end in a `Goto`, or some OTHER condition is what
	// declines the merge and this test is measuring that one instead. Asserted rather than assumed,
	// because two earlier fixtures failed exactly here and read as passing.
	for id := range joins {
		join, found := function.Block(id)
		if !found || join == nil {
			continue
		}
		for _, predecessorId := range join.Predecessors {
			predecessor, found := function.Block(predecessorId)
			if !found || predecessor == nil {
				continue
			}
			if _, isGoto := predecessor.Terminal.(*Goto); !isGoto {
				t.Fatalf("block %d predecessor %d ends in %T, not a Goto; the merge declines it "+
					"on the terminal condition rather than on the predecessor count, so this "+
					"fixture cannot see a change to the count condition",
					id, predecessorId, predecessor.Terminal)
			}
		}
	}

	if merged := MergeConsecutiveBlocks(function); merged == 0 {
		t.Fatal("nothing merged at all, so this fixture cannot distinguish a pass that respects " +
			"the join from one that never ran")
	}

	for id, predecessors := range joins {
		if _, found := function.Block(id); !found {
			t.Errorf("block %d had %d predecessors and was merged away; the other predecessors "+
				"now skip its instructions and the one that absorbed them runs them "+
				"unconditionally", id, predecessors)
		}
	}
	after := 0
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		if _, isIf := block.Terminal.(*If); isIf {
			after++
		}
	}
	if after != branches {
		t.Errorf("the function had %d If terminals before merging and %d after; a real branch "+
			"was flattened", branches, after)
	}
}
