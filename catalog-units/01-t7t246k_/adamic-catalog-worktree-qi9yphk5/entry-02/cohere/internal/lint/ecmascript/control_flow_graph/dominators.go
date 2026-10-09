package control_flow_graph

// This file is a LOCAL ADDITION, not vendored. See the change list in cfg.go's header.
//
// It exposes two things the graph already computed and threw away: the predecessor edges that
// `findFinalPathDominators` built into a scratch array, and the immediate-dominator array that
// same function wrote and dropped after reading one chain out of it.
//
// # Why this is an addition rather than a rewrite of paths.go
//
// `findFinalPathDominators` is a correct Cooper-Harvey-Kennedy implementation and it is load-
// bearing for `PathAnalysis.IsOnEveryFinalPath`, which `rules_of_hooks` consumes. Editing it to
// hand its array outward would put a second caller's requirements on a function whose current
// caller needs one boolean per block, and the two want different graphs: it dominates over a graph
// with ONE SYNTHETIC EXIT reached only by non-throwing final blocks, deliberately excluding thrown
// exits, because that is the question ESLint asks. A consumer asking "what dominates this block" in
// general does not want thrown paths excluded.
//
// So the algorithm is written a second time here over the real successor graph, and the two are
// pinned against each other by `TestDominatorsAgreeWithFinalPathDominators`. That test is the
// reason duplicating it is safe rather than a divergence waiting to happen: if either drifts, it
// goes red.

// Predecessors returns the reachable blocks with an edge into block, in Blocks order.
//
// The graph stores only successors, because that is the direction the builder lays edges down in.
// Every backward analysis needs the other direction, and every consumer that has needed it so far
// has rebuilt it: `findFinalPathDominators` builds a counting-sorted predecessor array and discards
// it, and the liveness solver in `no-useless-assignment` avoids needing one only by iterating every
// block every round.
//
// The result is computed once by `AnalyzeDominators` and shared, so this is a slice read rather
// than a walk. It is owned by the analysis and must not be mutated: two callers see the same slice.
//
// Unreachable blocks have no predecessors here even when a successor edge points at them. That
// mirrors what every existing consumer already does — each one skips `!successor.Reachable` at its
// own loop — and it is the definition dominance needs, since a block nothing arrives at cannot
// dominate anything.
func (d *DominatorAnalysis[E]) Predecessors(block *Block[E]) []*Block[E] {
	if d == nil || block == nil || block.Index() >= len(d.predecessorStart)-1 {
		return nil
	}
	start := d.predecessorStart[block.Index()]
	end := d.predecessorStart[block.Index()+1]
	return d.predecessors[start:end]
}

// DominatorAnalysis holds the dominator tree of one graph's reachable subgraph, plus the
// predecessor edges it was computed from.
//
// It is separate from PathAnalysis rather than folded into it because the two answer different
// questions over different graphs. PathAnalysis dominates over a graph with a synthetic exit and no
// thrown edges, to answer "is this on every non-throwing final path". This dominates over the real
// reachable successor graph, to answer "does control have to pass through A to reach B". A consumer
// wanting the second and given the first would get thrown paths silently excluded.
type DominatorAnalysis[E any] struct {
	graph *Graph[E]

	// immediate[i] is the block index of block i's immediate dominator, or -1 when block i is
	// unreachable. The entry block is its own immediate dominator, which is the convention
	// Cooper-Harvey-Kennedy uses as its fixed point and which `Dominates` accounts for.
	immediate []int

	// rpoPosition[i] is block i's index in reverse postorder, or -1 when unreachable. It is kept
	// because `Dominates` walks the tree upward and needs no ordering, but `ReversePostOrder`
	// hands it out and the intersect step needed it anyway.
	rpoPosition []int

	// reversePostOrder is the reachable blocks in reverse postorder: every block appears after at
	// least one predecessor, except the entry and the targets of back edges. It is the iteration
	// order a forward dataflow wants, and the order the liveness solver in `no-useless-assignment`
	// documents itself as approximating with a descending-index walk.
	reversePostOrder []*Block[E]

	// predecessors is one flat array of every reachable in-edge, grouped by target block, with
	// predecessorStart[i]..predecessorStart[i+1] delimiting block i's group. One allocation rather
	// than one slice per block, which is the same counting-sort shape `findFinalPathDominators`
	// already uses internally.
	predecessors     []*Block[E]
	predecessorStart []int
}

// AnalyzeDominators computes the dominator tree and predecessor edges of graph's reachable
// subgraph.
//
// Unreachable blocks are excluded entirely rather than given a dominator. There is no meaningful
// answer for them: a block control never arrives at is not dominated by anything, and treating the
// entry as its dominator would let a query about dead code come back with a confident true. Every
// accessor here answers false or nil for one, and `IsReachable` is the way to ask before believing
// a false.
func AnalyzeDominators[E any](graph *Graph[E]) *DominatorAnalysis[E] {
	if graph == nil || len(graph.Blocks) == 0 {
		return &DominatorAnalysis[E]{}
	}

	blockCount := len(graph.Blocks)
	analysis := &DominatorAnalysis[E]{
		graph:            graph,
		immediate:        make([]int, blockCount),
		rpoPosition:      make([]int, blockCount),
		predecessorStart: make([]int, blockCount+1),
	}

	// Counting sort over in-degree, so the predecessor groups land in one array.
	//
	// The entry block's own reachability is not read from Blocks[0].Reachable. Every other block's
	// is, but the entry is reachable by construction — `Build` sets it before laying anything out —
	// and reading the field here would make this depend on a field the builder sets rather than on
	// the definition.
	counts := make([]int, blockCount+1)
	for _, block := range graph.Blocks {
		if !block.Reachable {
			continue
		}
		for _, successor := range block.Successors {
			if successor != nil && successor.Reachable {
				counts[successor.Index()+1]++
			}
		}
	}
	for index := 1; index <= blockCount; index++ {
		counts[index] += counts[index-1]
	}
	copy(analysis.predecessorStart, counts)
	analysis.predecessors = make([]*Block[E], counts[blockCount])
	cursor := make([]int, blockCount)
	copy(cursor, counts[:blockCount])
	for _, block := range graph.Blocks {
		if !block.Reachable {
			continue
		}
		for _, successor := range block.Successors {
			if successor == nil || !successor.Reachable {
				continue
			}
			analysis.predecessors[cursor[successor.Index()]] = block
			cursor[successor.Index()]++
		}
	}

	for index := range analysis.immediate {
		analysis.immediate[index] = -1
		analysis.rpoPosition[index] = -1
	}

	// Depth-first postorder over reachable successors, then reversed.
	//
	// Written with an explicit stack rather than recursion. `findFinalPathDominators` recurses, and
	// on a deeply nested source file that is a stack the graph's depth controls; this runs over
	// every code path root in every file in the tree, so the depth is worth not trusting.
	postOrder := make([]*Block[E], 0, blockCount)
	type frame struct {
		block     *Block[E]
		successor int
	}
	visited := make([]bool, blockCount)
	stack := []frame{{block: graph.Blocks[0]}}
	visited[0] = true
	for len(stack) > 0 {
		top := &stack[len(stack)-1]
		if top.successor < len(top.block.Successors) {
			successor := top.block.Successors[top.successor]
			top.successor++
			if successor == nil || !successor.Reachable || visited[successor.Index()] {
				continue
			}
			visited[successor.Index()] = true
			stack = append(stack, frame{block: successor})
			continue
		}
		postOrder = append(postOrder, top.block)
		stack = stack[:len(stack)-1]
	}

	analysis.reversePostOrder = make([]*Block[E], len(postOrder))
	for index, block := range postOrder {
		reversed := len(postOrder) - 1 - index
		analysis.reversePostOrder[reversed] = block
		analysis.rpoPosition[block.Index()] = reversed
	}

	// Cooper-Harvey-Kennedy. The entry is its own immediate dominator, and every other reachable
	// block starts undefined and is refined by intersecting its already-processed predecessors
	// until nothing moves.
	analysis.immediate[0] = 0
	intersect := func(left, right int) int {
		for left != right {
			for analysis.rpoPosition[left] > analysis.rpoPosition[right] {
				left = analysis.immediate[left]
			}
			for analysis.rpoPosition[right] > analysis.rpoPosition[left] {
				right = analysis.immediate[right]
			}
		}
		return left
	}
	for changed := true; changed; {
		changed = false
		for _, block := range analysis.reversePostOrder[1:] {
			index := block.Index()
			newImmediate := -1
			for _, predecessor := range analysis.Predecessors(block) {
				predecessorIndex := predecessor.Index()
				if analysis.immediate[predecessorIndex] == -1 {
					continue
				}
				if newImmediate == -1 {
					newImmediate = predecessorIndex
				} else {
					newImmediate = intersect(predecessorIndex, newImmediate)
				}
			}
			if newImmediate != -1 && analysis.immediate[index] != newImmediate {
				analysis.immediate[index] = newImmediate
				changed = true
			}
		}
	}

	return analysis
}

// IsReachable reports whether block participates in this analysis at all.
//
// Every other accessor answers false or nil for an unreachable block, and that answer is
// indistinguishable from a real negative. This is how a caller tells the two apart.
func (d *DominatorAnalysis[E]) IsReachable(block *Block[E]) bool {
	if d == nil || block == nil || block.Index() >= len(d.immediate) {
		return false
	}
	return d.immediate[block.Index()] != -1
}

// ImmediateDominator returns the block that immediately dominates block: the last block every path
// from the entry to block passes through, other than block itself.
//
// The second result is false for an unreachable block and for the entry block, which has no
// immediate dominator. Returning the entry as its own would make a caller walking the tree upward
// loop forever, and callers do walk it upward.
func (d *DominatorAnalysis[E]) ImmediateDominator(block *Block[E]) (*Block[E], bool) {
	if !d.IsReachable(block) {
		return nil, false
	}
	index := block.Index()
	if index == 0 {
		return nil, false
	}
	return d.graph.Blocks[d.immediate[index]], true
}

// Dominates reports whether every path from the graph entry to block passes through dominator.
//
// A block dominates itself, which is the standard convention and the one Cooper-Harvey-Kennedy's
// fixed point produces. `StrictlyDominates` is the version that excludes it.
//
// Both blocks must be reachable. An unreachable block is dominated by nothing and dominates
// nothing, so either being unreachable answers false.
func (d *DominatorAnalysis[E]) Dominates(dominator, block *Block[E]) bool {
	if !d.IsReachable(dominator) || !d.IsReachable(block) {
		return false
	}
	target := dominator.Index()
	// Walking the tree upward rather than comparing depth intervals. The chain is bounded by the
	// tree's height and this is not on a hot path; a depth-interval encoding would need a second
	// pass to build and would have to be kept honest against the tree.
	for index := block.Index(); ; index = d.immediate[index] {
		if index == target {
			return true
		}
		if index == 0 {
			return false
		}
	}
}

// StrictlyDominates reports whether dominator dominates block and is not block.
func (d *DominatorAnalysis[E]) StrictlyDominates(dominator, block *Block[E]) bool {
	return dominator != block && d.Dominates(dominator, block)
}

// ReversePostOrder returns the reachable blocks in reverse postorder, entry first.
//
// This is the iteration order a forward dataflow wants: every block appears after at least one of
// its predecessors, except the entry and the targets of back edges, so one pass propagates as far
// as the acyclic structure allows and the fixed point converges in the number of rounds the loop
// nesting depth sets rather than the block count.
//
// A backward dataflow wants this reversed, which `Solve` does rather than making each caller
// remember to.
func (d *DominatorAnalysis[E]) ReversePostOrder() []*Block[E] {
	if d == nil {
		return nil
	}
	return d.reversePostOrder
}
