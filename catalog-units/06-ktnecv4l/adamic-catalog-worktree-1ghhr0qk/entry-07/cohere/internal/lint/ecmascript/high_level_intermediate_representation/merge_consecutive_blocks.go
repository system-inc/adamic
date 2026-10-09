// Merging blocks that always execute one after the other.
//
// This is React's `mergeConsecutiveBlocks` (`HIR/MergeConsecutiveBlocks.ts`), which upstream calls
// on the line after `inlineImmediatelyInvokedFunctionExpressions` (`Pipeline.ts:180`). The two are
// a pair rather than two independent passes, and the ordering is not decoration.
//
// # Why the inliner needs it, measured rather than assumed
//
// The splice cuts one block into three: the code before the call, the copied body, and the
// continuation. On the single-return path all three are joined by plain `Goto`s, so the region is
// straight-line code that has been given block boundaries it did not have before.
//
// Those boundaries are not inert here. `DropManualMemoization` rewrites `useMemo(fn, deps)` into a
// zero-argument call of `fn`, which is structurally an IIFE, so the inliner splices it -- correctly,
// and upstream does the same. But it leaves `StartMemoize` in the first block and `FinishMemoize`
// in the third, and scope construction reads block structure. Measured on the vendored corpus with
// the splice and no merge: the goldens this rule fires on fell from 15 to 6, and every one of the
// nine losses was a fixture where a splice had happened. With the merge the region is one block
// again and the markers sit where they sat before.
package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/static_single_assignment"
	"slices"
)

// MergeConsecutiveBlocks collapses each block into its predecessor where control always flows from
// one to the other, and reports how many blocks were merged away.
//
// The condition is upstream's, and all four parts are load-bearing:
//
//   - The block has exactly one predecessor, so nothing else can reach it.
//   - The predecessor ends in a plain `Goto`, so it always transfers control here.
//
// # The first two overlap on every graph reached so far, and the count is kept anyway
//
// A mutation sweep could not kill the predecessor-count condition. Three fixtures were built to try
// -- a plain branch, a splice in one arm, a splice in both -- and all three produced byte-identical
// graphs with the condition removed. The reason is the walk order rather than the condition being
// wrong: blocks are visited in reverse postorder, so by the time a join is reached its predecessors
// have already absorbed their own predecessors and carry the join's terminal, and the `Goto`
// condition rejects it before the count is consulted.
//
// That is an accident of how lowering shapes a graph, not a property this pass should rely on. A
// join whose resolved predecessor still ends in a `Goto` would be swallowed without the count, and
// swallowing a join makes its other predecessors skip its instructions entirely. The condition is
// upstream's, it is cheap, and correctness here should not rest on a traversal order that a later
// pass is free to change. It is kept, and the sweep's result is recorded rather than dressed up as
// a pass.
//   - Both blocks are ordinary statement blocks. A value or loop block is named by a high-level
//     terminal that expects to find it, and merging it away breaks that terminal's structure.
//   - The block is not some terminal's fallthrough. Merging across a fallthrough would pull the
//     predecessor's instructions into a region the fallthrough was meant to sit outside of.
//
// # Phis in a merged block become loads
//
// A block with one predecessor can still hold phis, left behind by an earlier construction. Each
// has exactly one operand by definition, so it is not a merge at all: it is an alias, and after the
// blocks are joined there is no join for it to sit at. Upstream rewrites each into a `LoadLocal` of
// its single operand, appended to the predecessor before the merged instructions, and this does the
// same. Dropping them instead would leave the phi's value undefined at every later read.
func MergeConsecutiveBlocks(function *Function) int {
	if function == nil {
		return 0
	}

	// Where each merged-away block's contents ended up. Transitive, because a block merged into a
	// predecessor that is itself later merged has to resolve all the way down.
	mergedInto := map[static_single_assignment.BlockId]static_single_assignment.BlockId{}
	resolve := func(id static_single_assignment.BlockId) static_single_assignment.BlockId {
		current := id
		for {
			next, moved := mergedInto[current]
			if !moved || next == current {
				return current
			}
			current = next
		}
	}

	// Blocks that are some terminal's fallthrough. Collected as the walk goes, matching upstream:
	// a fallthrough named by a LATER block does not protect an earlier one, because the walk is in
	// reverse postorder and a fallthrough is always laid out after the terminal that names it.
	fallthroughs := map[static_single_assignment.BlockId]bool{}
	merged := 0
	survivors := make([]*BasicBlock, 0, len(function.Blocks))

	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		if target, has := Fallthrough(block.Terminal); has {
			fallthroughs[target] = true
		}
		if len(block.Predecessors) != 1 || block.Kind != BlockKindBlock ||
			fallthroughs[block.Id] {
			survivors = append(survivors, block)
			continue
		}
		predecessor, found := function.Block(resolve(block.Predecessors[0]))
		if !found || predecessor == nil {
			survivors = append(survivors, block)
			continue
		}
		if _, isGoto := predecessor.Terminal.(*Goto); !isGoto ||
			predecessor.Kind != BlockKindBlock {
			// The predecessor is not guaranteed to transfer control here, so the two are not
			// consecutive.
			survivors = append(survivors, block)
			continue
		}

		for _, phi := range block.Phis {
			operand, single := singlePhiOperand(phi)
			if !single {
				// A block with one predecessor whose phi has several operands is a graph whose
				// predecessor list and phis disagree. Upstream asserts here. Skipping the phi is
				// the conservative answer: picking an operand arbitrarily would make the merged
				// block read a value from an edge that does not exist, which produces a plausible
				// function that is wrong.
				continue
			}
			function.AddInstruction(predecessor, &Instruction{
				LValue: phi.Place,
				Value:  &LoadLocal{Place: operand},
			})
		}

		predecessor.Instructions = append(predecessor.Instructions, block.Instructions...)
		predecessor.Terminal = block.Terminal
		mergedInto[block.Id] = predecessor.Id
		delete(function.blocksById, block.Id)
		merged++
	}

	if merged == 0 {
		return 0
	}
	function.Blocks = survivors

	// A phi operand keyed by a block that no longer exists reads from an edge that is gone. Rekey
	// to whatever the predecessor was merged into. The walk is over a copy, because rekeying
	// deletes from and inserts into the sorted slice it would otherwise be ranging.
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, phi := range block.Phis {
			for _, entry := range slices.Clone(phi.Operands) {
				mapped := resolve(entry.Predecessor)
				if mapped == entry.Predecessor {
					continue
				}
				phi.Operands.Delete(entry.Predecessor)
				phi.Operands.Set(mapped, entry.Place)
			}
		}
	}

	// Fallthroughs are not edges, so `MarkPredecessors` does not repair them and a stale one names
	// a block the function no longer holds.
	//
	// A mutation sweep could not kill this loop either, and the measurement says why rather than
	// leaving it as a shrug: a merged-away block is absorbed INTO its only predecessor, so the only
	// terminal that named it is that predecessor's, and the merge overwrites that terminal in the
	// same step. Nothing else can name it, because "nothing else can reach it" is the condition
	// that let it merge.
	//
	// It survives as protection against the one shape that would break that reasoning -- a
	// fallthrough pointing into the middle of a merged run -- which the fallthrough guard above is
	// what currently prevents. Two guards for one hazard, and removing either alone is invisible.
	// Said plainly so the next person does not spend the afternoon proving it twice.
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		EachBlockReferencePointer(block.Terminal, func(reference *static_single_assignment.BlockId) {
			*reference = resolve(*reference)
		})
	}
	if function.Entry != resolve(function.Entry) {
		function.Entry = resolve(function.Entry)
	}

	MarkPredecessors(function)
	return merged
}

// singlePhiOperand returns a phi's only operand, and whether it had exactly one.
func singlePhiOperand(phi *Phi) (Place, bool) {
	if len(phi.Operands) != 1 {
		return Place{}, false
	}
	return phi.Operands[0].Place, true
}
