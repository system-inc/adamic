// Graph maintenance: the passes Build runs to bring a freshly built function into the state every
// other pass assumes. Lifted from cohere's graph.go (high_level_intermediate_representation at
// 715ba94); what changed is said where it changed.
//
// Three invariants are established here and every consumer may rely on them:
//
//  1. Function.Blocks is in reverse postorder, and unreachable blocks have been removed.
//  2. Every block's Predecessors is exactly the set of blocks with a real edge into it.
//  3. Every instruction and terminal has a nonzero, monotonically increasing EvaluationOrder.
//
// They are re-established by calling Finalize, which any pass that restructures the graph must do.
package flow

import ssa "github.com/system-inc/cohere/static_single_assignment"

// Finalize brings a freshly lowered function into the state every pass assumes.
//
// Order matters: reverse postorder first, because it drops unreachable blocks and the predecessor
// edges must not name a block that is gone; evaluation order last, because it walks the blocks in
// their final order. (cohere's also recurses into nested functions; an Adamic closure is a function
// of its own in the IR, and is built and finalized on its own.)
func Finalize(function *Function) {
	ssa.ReversePostorder(cohereGraph{}, function)
	ssa.MarkPredecessors(cohereGraph{}, function)
	ssa.MarkEvaluationOrder(cohereGraph{}, function)
}
