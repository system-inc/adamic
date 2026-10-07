// Graph maintenance: the passes lowering runs to bring a freshly built function into the state
// every other pass assumes.
//
// Three invariants are established here and every consumer may rely on them:
//
//  1. Function.Blocks is in reverse postorder, and unreachable blocks have been removed.
//  2. Every block's Predecessors is exactly the set of blocks with a real edge into it.
//  3. Every instruction and terminal has a nonzero, monotonically increasing EvaluationOrder.
//
// They are re-established by calling Finalize, which any pass that restructures the graph must do.
package high_level_intermediate_representation

// Finalize brings a freshly lowered function into the state every pass assumes.
//
// Order matters: reverse postorder first, because it drops unreachable blocks and the predecessor
// edges must not name a block that is gone; evaluation order last, because it walks the blocks in
// their final order.
func Finalize(function *Function) {
	ReversePostorder(function)
	MarkPredecessors(function)
	MarkEvaluationOrder(function)
	for _, nested := range function.Functions {
		Finalize(nested)
	}
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
// The traversal is upstream's `getReversePostorderedBlocks`: visit a structural fallthrough first,
// then the real successors in reverse order. Postorder is reversed at the end, so that puts loop
// bodies and conditional arms before their continuation in the final slice. That ordering is more
// than presentation. Evaluation order is assigned from this slice, and mutable-range inference
// must see a mutation in a loop body before a read after the loop.
//
// A fallthrough is visited initially as structural rather than executable. If no real edge reaches
// it, upstream retains its block id as an empty Unreachable block; if a real edge reaches it later,
// the same block is revisited as executable. Keeping those two states separate prevents a
// fallthrough link from manufacturing a control-flow edge while still preserving the structured
// terminal's required target.
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
	usedFallthroughs := make([]bool, bound)
	inRange := func(id BlockId) bool { return int(id) < bound }

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
		if fallthroughBlock, ok := Fallthrough(block.Terminal); ok {
			if isUsed && inRange(fallthroughBlock) {
				usedFallthroughs[fallthroughBlock] = true
			}
			successors = append(successors, successor{id: fallthroughBlock, isUsed: false})
		}
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
		switch {
		case inRange(block.Id) && used[block.Id]:
			blocks = append(blocks, block)
		case inRange(block.Id) && usedFallthroughs[block.Id]:
			placeholder := &BasicBlock{
				Id:           block.Id,
				Kind:         block.Kind,
				Terminal:     &Unreachable{},
				Predecessors: append([]BlockId(nil), block.Predecessors...),
			}
			blocks = append(blocks, placeholder)
			function.blocksById[block.Id] = placeholder
		}
	}
	function.Blocks = blocks

	for id := range function.blocksById {
		if !inRange(id) || (!used[id] && !usedFallthroughs[id]) {
			delete(function.blocksById, id)
		}
	}
}

// MarkPredecessors recomputes every block's Predecessors from the real edges.
//
// Fallthroughs are NOT edges and do not produce a predecessor. See EachSuccessor.
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
	case *Throw:
		t.Order = order
	case *Unreachable:
		t.Order = order
	case *Unsupported:
		t.Order = order
	case *Goto:
		t.Order = order
	case *If:
		t.Order = order
	case *Branch:
		t.Order = order
	case *Switch:
		t.Order = order
	case *While:
		t.Order = order
	case *DoWhile:
		t.Order = order
	case *For:
		t.Order = order
	case *ForOf:
		t.Order = order
	case *ForIn:
		t.Order = order
	case *Logical:
		t.Order = order
	case *Ternary:
		t.Order = order
	case *Optional:
		t.Order = order
	case *Sequence:
		t.Order = order
	case *Label:
		t.Order = order
	case *Try:
		t.Order = order
	case *MaybeThrow:
		t.Order = order
	case *Scope:
		t.Order = order
	}
}
