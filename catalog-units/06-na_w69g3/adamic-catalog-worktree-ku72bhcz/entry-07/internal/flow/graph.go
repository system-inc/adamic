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

// Finalize brings a freshly lowered function into the state every pass assumes.
//
// Order matters: reverse postorder first, because it drops unreachable blocks and the predecessor
// edges must not name a block that is gone; evaluation order last, because it walks the blocks in
// their final order. (cohere's also recurses into nested functions; an Adamic closure is a function
// of its own in the IR, and is built and finalized on its own.)
func Finalize(function *Function) {
	ReversePostorder(function)
	MarkPredecessors(function)
	MarkEvaluationOrder(function)
}

// ReversePostorder reorders Function.Blocks into reverse postorder and drops unreachable blocks.
//
// # Why unreachable blocks are dropped here and kept in controlflow
//
// `controlflow` keeps them, because a rule may want to ask what would have run after a `return`.
// This drops them, because a value defined only in a block control never reaches is a value no
// analysis should see: single-assignment construction over an unreachable definition produces a phi
// operand from a predecessor that cannot execute, and every effect derived from it is fiction.
//
// A pass wanting the unreachable code should read the AST, or the other graph. This one is for
// reasoning about what runs.
//
// The traversal is upstream's `getReversePostorderedBlocks`, less its structural fallthroughs: it
// visits the real successors in reverse order, and postorder is reversed at the end. cohere's
// terminals name a fallthrough (the block after an if or a loop) beside their real edges, and visit it
// first so loop bodies come before their continuation; Adamic's terminals have only real edges, so
// that part, and the placeholder block it keeps for a fallthrough nothing reaches, isn't here. The
// order is still a reverse postorder, which is what single assignment needs. Mutable ranges, when
// they're lifted, will want the loop-body-first order back, and with it a fallthrough on If.
func ReversePostorder(function *Function) {
	if function == nil {
		return
	}
	postorder := make([]*BasicBlock, 0, len(function.Blocks))

	// The three sets are dense slices over block ids rather than maps, and the walk keeps its frames
	// and their successors in two flat buffers rather than a frame and two slices per block. The
	// traversal is unchanged; only where its bookkeeping lives moved, because this runs for every
	// lowered function and the per-block allocations were a million objects on a cold ahra run
	// (#rwsffzm). Ids are bounded by `nextBlock`, and by the largest id in the table for a function
	// built by hand without `NewFunction`.
	bound := int(function.nextBlock)
	for id := range function.blocksById {
		if int(id) >= bound {
			bound = int(id) + 1
		}
	}
	visited := make([]bool, bound)
	used := make([]bool, bound)
	inRange := func(id BlockId) bool { return int(id) < bound }

	// isUsed is always true here: it told cohere's walk a structural fallthrough from a real edge,
	// and Adamic's graph has only real edges. It's kept so the walk reads as cohere's does.
	type successor struct {
		id     BlockId
		isUsed bool
	}
	type frame struct {
		block *BasicBlock
		// successors are this frame's entries in the shared buffer, from start to end. A child's
		// entries are appended after them and truncated away when the child is popped, so the
		// buffer is a stack in step with the frames.
		start, end, next int
		ownsAppend       bool
	}
	var successors []successor
	var real []BlockId
	collect := func(next BlockId) { real = append(real, next) }
	enter := func(id BlockId, isUsed bool) (frame, bool) {
		if !inRange(id) {
			return frame{}, false
		}
		wasUsed := used[id]
		wasVisited := visited[id]
		visited[id] = true
		if isUsed {
			used[id] = true
		}
		if wasVisited && (wasUsed || !isUsed) {
			return frame{}, false
		}

		block, ok := function.Block(id)
		if !ok {
			return frame{}, false
		}
		start := len(successors)
		real = real[:0]
		EachSuccessor(block.Terminal, collect)
		for index := len(real) - 1; index >= 0; index-- {
			successors = append(successors, successor{id: real[index], isUsed: isUsed})
		}
		return frame{block: block, start: start, end: len(successors), next: start, ownsAppend: !wasVisited}, true
	}

	entry, ok := enter(function.Entry, true)
	if !ok {
		return
	}
	stack := []frame{entry}

	for len(stack) > 0 {
		top := &stack[len(stack)-1]
		if top.next == top.end {
			if top.ownsAppend {
				postorder = append(postorder, top.block)
			}
			successors = successors[:top.start]
			stack = stack[:len(stack)-1]
			continue
		}
		next := successors[top.next]
		top.next++
		if child, ok := enter(next.id, next.isUsed); ok {
			stack = append(stack, child)
		}
	}

	blocks := make([]*BasicBlock, 0, len(postorder))
	for index := len(postorder) - 1; index >= 0; index-- {
		block := postorder[index]
		if inRange(block.Id) && used[block.Id] {
			blocks = append(blocks, block)
		}
	}
	function.Blocks = blocks

	for id := range function.blocksById {
		if !inRange(id) || !used[id] {
			delete(function.blocksById, id)
		}
	}
}

// MarkPredecessors recomputes every block's Predecessors from the real edges.
//
// This is the field single-assignment construction needs: Braun's algorithm reads a value at a
// block join by asking each predecessor what it holds, and a missing or spurious predecessor is a
// wrong phi rather than a crash.
func MarkPredecessors(function *Function) {
	for _, block := range function.Blocks {
		block.Predecessors = block.Predecessors[:0]
	}
	for _, block := range function.Blocks {
		seen := map[BlockId]bool{}
		EachSuccessor(block.Terminal, func(id BlockId) {
			if seen[id] {
				return
			}
			seen[id] = true
			successor, ok := function.Block(id)
			if !ok {
				return
			}
			successor.Predecessors = append(successor.Predecessors, block.Id)
		})
	}
}

// MarkEvaluationOrder assigns each instruction and terminal its position in evaluation order.
//
// Numbering follows Function.Blocks, which is reverse postorder, so a forward analysis sees
// increasing numbers along any acyclic path. Across a back edge the number decreases, which is the
// correct and expected signal that a loop was traversed.
//
// Starts at 1: zero means unassigned, and a pass that reads an order of zero has found a block the
// finalizer did not reach.
func MarkEvaluationOrder(function *Function) {
	order := EvaluationOrder(1)
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			function.Instructions[instructionId].Order = order
			order++
		}
		setTerminalOrder(block.Terminal, order)
		order++
	}
}

func setTerminalOrder(terminal Terminal, order EvaluationOrder) {
	switch t := terminal.(type) {
	case *Return:
		t.Order = order
	case *Unreachable:
		t.Order = order
	case *Goto:
		t.Order = order
	case *If:
		t.Order = order
	case *MayThrow:
		t.Order = order
	case *Throw:
		t.Order = order
	case *Choose:
		t.Order = order
	}
}
