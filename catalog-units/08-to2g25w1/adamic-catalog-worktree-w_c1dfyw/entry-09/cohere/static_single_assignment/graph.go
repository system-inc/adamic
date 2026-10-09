// Graph maintenance: the passes an IR runs to bring a freshly built function into the state every
// other pass assumes.
//
// Three invariants are established here and every consumer may rely on them:
//
//  1. The block slice is in reverse postorder, and unreachable blocks have been removed.
//  2. Every block's predecessors are exactly the set of blocks with a real edge into it.
//  3. Every instruction and terminal has a nonzero, monotonically increasing EvaluationOrder.
//
// An IR re-establishes them by calling the three passes in that order (its Finalize), which any pass
// that restructures the graph must do.
package static_single_assignment

import "sync"

// ReversePostorder reorders the function's blocks into reverse postorder and drops unreachable blocks.
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
// terminal's required target. An IR whose terminals have only real edges reports no Fallthrough, and
// the walk is then a plain reverse postorder.
func ReversePostorder[G Graph[F, B, P], F any, B comparable, P any](graph G, function F) {
	scratch := reversePostorderScratchPool.Get().(*reversePostorderScratch)
	postorder := scratch.postorder[:0]

	// The three sets are dense slices over block ids rather than maps, and the walk keeps its frames
	// and their successors in two flat buffers rather than a frame and two slices per block. The
	// traversal is unchanged; only where its bookkeeping lives moved, because this runs for every
	// lowered function and the per-block allocations were a million objects on a cold ahra run
	// (#rwsffzm). Ids are bounded by BlockBound.
	bound := graph.BlockBound(function)
	scratch.bound = bound
	visited := clearedFlags(scratch.visited, bound)
	scratch.used = clearedFlags(scratch.used, bound)
	scratch.usedFallthroughs = clearedFlags(scratch.usedFallthroughs, bound)
	used, usedFallthroughs := scratch.used, scratch.usedFallthroughs
	inRange := func(id BlockId) bool { return int(id) < bound }

	type successor = reversePostorderSuccessor
	type frame = reversePostorderFrame
	successors := scratch.successors[:0]
	stack := scratch.stack[:0]
	// The buffers go back to the pool however this returns. They hold ids only, so a pooled buffer
	// never keeps a function's blocks alive.
	defer func() {
		scratch.postorder, scratch.successors, scratch.stack, scratch.visited = postorder, successors, stack, visited
		reversePostorderScratchPool.Put(scratch)
	}()
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

		block, ok := graph.Block(function, id)
		if !ok {
			return frame{}, false
		}
		start := len(successors)
		scratch.real = scratch.real[:0]
		scratch.hasFallthrough = false
		graph.EachEdge(block, scratch.collect)
		if scratch.hasFallthrough {
			if isUsed && inRange(scratch.fallthroughBlock) {
				usedFallthroughs[scratch.fallthroughBlock] = true
			}
			successors = append(successors, successor{id: scratch.fallthroughBlock, isUsed: false})
		}
		for index := len(scratch.real) - 1; index >= 0; index-- {
			successors = append(successors, successor{id: scratch.real[index], isUsed: isUsed})
		}
		return frame{block: id, start: start, end: len(successors), next: start, ownsAppend: !wasVisited}, true
	}

	entry, ok := enter(graph.Entry(function), true)
	if !ok {
		return
	}
	stack = append(stack, entry)

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

	blocks := make([]B, 0, len(postorder))
	for index := len(postorder) - 1; index >= 0; index-- {
		id := postorder[index]
		switch {
		case inRange(id) && used[id]:
			block, _ := graph.Block(function, id)
			blocks = append(blocks, block)
		case inRange(id) && usedFallthroughs[id]:
			block, _ := graph.Block(function, id)
			blocks = append(blocks, graph.Placeholder(function, block))
		}
	}
	graph.SetBlocks(function, blocks)
	graph.Retain(function, scratch.keep)
}

// reversePostorderSuccessor is one edge ReversePostorder will follow, and whether it is executable.
type reversePostorderSuccessor struct {
	id     BlockId
	isUsed bool
}

// reversePostorderFrame is one block on ReversePostorder's walk stack.
type reversePostorderFrame struct {
	block BlockId
	// successors are this frame's entries in the shared buffer, from start to end. A child's entries
	// are appended after them and truncated away when the child is popped, so the buffer is a stack in
	// step with the frames.
	start, end, next int
	ownsAppend       bool
}

// reversePostorderScratch is ReversePostorder's working memory, kept between calls. It runs once per
// lowered function, and its stack, its postorder and its three sets were 22 MB of a cold ahra run,
// allocated and dropped every call (#p4h0p54). Only the ordered block list it returns needs to be new.
//
// It holds block ids rather than an IR's blocks, so one pool serves every IR. The walk looks each kept
// block up once, when it assembles the result.
//
// The two callbacks the walk hands to Graph live here too, made once with the scratch. A callback
// handed to a Graph method escapes, since the compiler cannot see through a type parameter's method,
// and so does every local it captures: made per call, they were about eight allocations a call and a
// million on a cold ahra run. These read and write only the scratch.
type reversePostorderScratch struct {
	postorder                       []BlockId
	visited, used, usedFallthroughs []bool
	bound                           int
	successors                      []reversePostorderSuccessor
	stack                           []reversePostorderFrame

	// real and the fallthrough are what collect gathers from one block's edges.
	real             []BlockId
	fallthroughBlock BlockId
	hasFallthrough   bool

	collect func(next BlockId, edge Edge)
	keep    func(id BlockId) bool
}

var reversePostorderScratchPool = sync.Pool{New: func() any {
	scratch := new(reversePostorderScratch)
	scratch.collect = func(next BlockId, edge Edge) {
		if edge == Fallthrough {
			scratch.fallthroughBlock, scratch.hasFallthrough = next, true
			return
		}
		scratch.real = append(scratch.real, next)
	}
	scratch.keep = func(id BlockId) bool {
		return int(id) < scratch.bound && (scratch.used[id] || scratch.usedFallthroughs[id])
	}
	return scratch
}}

// edgeScratch gathers one block's edges, for a pass that reads them in a loop of its own rather than
// in a callback. Pooled with its callback for the reason reversePostorderScratch says.
type edgeScratch struct {
	edges   []edgeTo
	collect func(successor BlockId, edge Edge)
}

// edgeTo is one gathered edge.
type edgeTo struct {
	successor BlockId
	edge      Edge
}

var edgeScratchPool = sync.Pool{New: func() any {
	scratch := new(edgeScratch)
	scratch.collect = func(successor BlockId, edge Edge) {
		scratch.edges = append(scratch.edges, edgeTo{successor: successor, edge: edge})
	}
	return scratch
}}

// edgesOf gathers a block's edges into the scratch, replacing what it held.
func edgesOf[G Graph[F, B, P], F any, B comparable, P any](graph G, block B, scratch *edgeScratch) []edgeTo {
	scratch.edges = scratch.edges[:0]
	graph.EachEdge(block, scratch.collect)
	return scratch.edges
}

// clearedFlags is flags resized to length, all false, reusing its array when it is long enough.
func clearedFlags(flags []bool, length int) []bool {
	if cap(flags) < length {
		return make([]bool, length)
	}
	flags = flags[:length]
	clear(flags)
	return flags
}

// MarkPredecessors recomputes every block's predecessors from the real edges.
//
// Fallthroughs are NOT edges and do not produce a predecessor; exceptional edges are real ones.
//
// This is the field single-assignment construction needs: Braun's algorithm reads a value at a
// block join by asking each predecessor what it holds, and a missing or spurious predecessor is a
// wrong phi rather than a crash.
func MarkPredecessors[G Graph[F, B, P], F any, B comparable, P any](graph G, function F) {
	blocks := graph.Blocks(function)
	for _, block := range blocks {
		graph.SetPredecessors(block, graph.Predecessors(block)[:0])
	}
	scratch := edgeScratchPool.Get().(*edgeScratch)
	defer edgeScratchPool.Put(scratch)
	for _, block := range blocks {
		id := graph.Id(block)
		seen := map[BlockId]bool{}
		for _, edge := range edgesOf(graph, block, scratch) {
			if edge.edge == Fallthrough || seen[edge.successor] {
				continue
			}
			seen[edge.successor] = true
			successor, ok := graph.Block(function, edge.successor)
			if !ok {
				continue
			}
			graph.SetPredecessors(successor, append(graph.Predecessors(successor), id))
		}
	}
}

// MarkEvaluationOrder assigns each instruction and terminal its position in evaluation order.
//
// Numbering follows the block slice, which is reverse postorder, so a forward analysis sees
// increasing numbers along any acyclic path. Across a back edge the number decreases, which is the
// correct and expected signal that a loop was traversed.
//
// Starts at 1: zero means unassigned, and a pass that reads an order of zero has found a block the
// finalizer did not reach.
func MarkEvaluationOrder[G Graph[F, B, P], F any, B comparable, P any](graph G, function F) {
	order := EvaluationOrder(1)
	for _, block := range graph.Blocks(function) {
		count := graph.InstructionCount(function, block)
		for index := 0; index < count; index++ {
			graph.SetInstructionOrder(function, block, index, order)
			order++
		}
		graph.SetTerminalOrder(block, order)
		order++
	}
}
