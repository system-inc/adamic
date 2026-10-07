// Post-dominance, and the set of blocks that must execute on the way out of a function.
//
// This is a port of React's `Dominator.ts` and `ComputeUnconditionalBlocks.ts`, transcribed by oxc
// at `oxc_react_compiler/src/react_compiler_hir/dominator.rs`. Cooper, Harvey and Kennedy, "A
// Simple, Fast Dominance Algorithm", run over the REVERSED graph with one synthetic exit.
//
// # Why this is here rather than borrowed from internal/lint/ecmascript/control_flow_graph
//
// `controlflow` has both pieces and neither is reachable. `AnalyzeDominators` is generic over
// `control_flow_graph.Graph[E]` and `control_flow_graph.Block[E]`, whose `index`, `final` and `thrown` fields are
// unexported and set only by `control_flow_graph.Build`, so a caller holding a `high_level_intermediate_representation.Function` cannot
// construct one. `PathAnalysis.IsOnEveryFinalPath` computes semantically the right thing over the
// right kind of graph, and is likewise only obtainable from a graph this package does not produce.
// `ssa.go`'s header states the general form of this: nothing in that package applies to this graph
// without a conversion that does not exist.
//
// It is also NOT the same question `ssa_verify.go`'s `computeDominance` answers, and the difference
// is the whole point rather than an implementation detail. That one is forward dominance: does
// every path from the entry REACH this block. This is post-dominance: does every path from the
// entry PASS THROUGH this block on its way to a return.
//
// Measured on React's own rule, which is why the distinction is recorded rather than asserted:
//
//	function Component(props) {
//	  const [x, setX] = useState(0);
//	  if (props.c) { return null; }
//	  setX(1);          // clean upstream
//	  return x;
//	}
//
// The entry block dominates the block holding `setX(1)`, so a forward-dominance test marks it
// unconditional and reports. It does not post-dominate the entry, because the early return leaves
// by another path, and upstream is silent. Reusing `computeDominance` here would have produced a
// false positive that no fixture built from the imported corpus contains, since every upstream
// invalid case puts the call on a path that both tests agree about.
package high_level_intermediate_representation

// UnconditionalBlocks returns the blocks that execute on every non-throwing path through function.
//
// This is upstream's `computeUnconditionalBlocks`: walk the post-dominator chain from the entry
// block to the synthetic exit, collecting every real block on the way. A block on that chain lies
// on every path from the entry to a return, so an instruction in it runs unconditionally.
//
// # Throwing paths are excluded, deliberately, and it is observable
//
// The synthetic exit is fed by `Return` terminals only, never by `Throw`. Upstream passes
// `includeThrowsAsExitNode: false` from this caller and true from others, and the choice decides a
// real fixture:
//
//	for (const _ of props) {
//	  if (props.cond) { break; } else { throw new Error('bye!'); }
//	}
//	setState(true);     // REPORTS upstream
//
// If a throw counted as an exit, the throwing arm would be a second way out and the code after the
// loop would stop post-dominating the entry. Because it does not count, the only way to leave is
// through the `break`, so the call after the loop is unconditional and reports. Measured on React's
// rule, which reports it, and it is the reason this parameter is fixed here rather than exposed:
// no caller in this tree wants the other answer, and a flag nothing sets is a flag nothing tests.
//
// A block that cannot reach any return at all is not on the chain and is therefore never
// unconditional, which is the right answer: an infinite loop's body runs, but nothing after it does.
//
// The result is nil for a nil function and for one with no blocks. Membership is the only supported
// question; the ordering of the underlying walk is not exposed because it carries no meaning.
func UnconditionalBlocks(function *Function) map[BlockId]bool {
	if function == nil || len(function.Blocks) == 0 {
		return nil
	}

	tree := computePostDominance(function)
	unconditional := make(map[BlockId]bool, len(function.Blocks))

	// Upstream asserts against re-entering a block, calling it a non-terminating loop. A cycle
	// cannot occur in a post-dominator tree, which is a tree by construction, but the walk is
	// bounded anyway rather than trusting that: this runs over lowered user code, and a panic in a
	// linter is worse than a conservative answer.
	current := function.Entry
	for step := 0; step <= len(function.Blocks); step++ {
		if current == postDominatorExit || unconditional[current] {
			break
		}
		unconditional[current] = true
		next, ok := tree.immediate[current]
		if !ok {
			break
		}
		current = next
	}
	return unconditional
}

// postDominatorExit is the synthetic block every return flows into.
//
// `InvalidBlock` is zero and lowering never mints block zero, so zero is free to mean the exit here
// and cannot collide with a real id.
const postDominatorExit BlockId = InvalidBlock

// postDominanceTree holds the immediate post-dominator of each block.
//
// A block absent from `immediate` reaches no return, so it post-dominates nothing and nothing
// reports it. That is a real state rather than an error: a function whose every path throws has one.
type postDominanceTree struct {
	immediate map[BlockId]BlockId
}

// computePostDominance runs Cooper-Harvey-Kennedy over the reversed graph.
//
// The algorithm needs its nodes in reverse postorder OF THE GRAPH IT WALKS, and the graph it walks
// here is the reverse of `Function.Blocks`' order rather than that order itself. So the traversal
// below computes its own ordering from the exit backwards, instead of reusing `Function.Blocks`,
// which is in reverse postorder of the FORWARD graph and is the wrong order for this.
//
// Getting that wrong does not fail loudly. Cooper-Harvey-Kennedy's `intersect` walks two pointers
// toward each other by comparing ordering positions, so a wrong order makes it walk the wrong way
// and either loop or settle on a node that does not dominate either input. The result is a
// plausible tree that is quietly incorrect, which is why the order is derived here rather than
// borrowed.
func computePostDominance(function *Function) *postDominanceTree {
	// Successors in the reversed graph are predecessors in the forward one, and vice versa. The
	// exit is a successor of every returning block.
	reverseSuccessors := make(map[BlockId][]BlockId, len(function.Blocks)+1)
	for _, block := range function.Blocks {
		if _, isReturn := block.Terminal.(*Return); isReturn {
			reverseSuccessors[postDominatorExit] = append(reverseSuccessors[postDominatorExit], block.Id)
		}
		for _, predecessorId := range block.Predecessors {
			reverseSuccessors[block.Id] = append(reverseSuccessors[block.Id], predecessorId)
		}
	}

	order := reversePostorderFrom(postDominatorExit, reverseSuccessors)
	position := make(map[BlockId]int, len(order))
	for index, id := range order {
		position[id] = index
	}

	immediate := make([]int, len(order))
	for index := range immediate {
		immediate[index] = -1
	}
	tree := &postDominanceTree{immediate: make(map[BlockId]BlockId, len(order))}
	if len(order) == 0 {
		return tree
	}
	immediate[0] = 0

	// The two pointers walk toward the root, which sits at position zero, so the larger position is
	// always the one further from it and is the one that steps.
	intersect := func(a, b int) int {
		for a != b {
			for a > b {
				if immediate[a] == -1 || immediate[a] == a {
					return b
				}
				a = immediate[a]
			}
			for b > a {
				if immediate[b] == -1 || immediate[b] == b {
					return a
				}
				b = immediate[b]
			}
		}
		return a
	}

	// Predecessors in the reversed graph, which the fixed point reads.
	reversePredecessors := make(map[BlockId][]BlockId, len(order))
	for from, targets := range reverseSuccessors {
		for _, to := range targets {
			reversePredecessors[to] = append(reversePredecessors[to], from)
		}
	}

	for changed := true; changed; {
		changed = false
		for index := 1; index < len(order); index++ {
			newImmediate := -1
			for _, predecessorId := range reversePredecessors[order[index]] {
				predecessorIndex, known := position[predecessorId]
				if !known || immediate[predecessorIndex] == -1 {
					continue
				}
				if newImmediate == -1 {
					newImmediate = predecessorIndex
				} else {
					newImmediate = intersect(predecessorIndex, newImmediate)
				}
			}
			if newImmediate != -1 && immediate[index] != newImmediate {
				immediate[index] = newImmediate
				changed = true
			}
		}
	}

	for index, id := range order {
		if index == 0 || immediate[index] == -1 {
			continue
		}
		tree.immediate[id] = order[immediate[index]]
	}
	return tree
}

// reversePostorderFrom returns the nodes reachable from root, in reverse postorder.
//
// Iterative rather than recursive for the same reason `ReversePostorder` is: the depth is whatever
// the source contains, and a linter must not be bounded by the goroutine stack.
func reversePostorderFrom(root BlockId, successors map[BlockId][]BlockId) []BlockId {
	visited := make(map[BlockId]bool, len(successors))
	postorder := make([]BlockId, 0, len(successors))

	type frame struct {
		id   BlockId
		next int
	}
	stack := []*frame{{id: root}}
	visited[root] = true

	for len(stack) > 0 {
		top := stack[len(stack)-1]
		children := successors[top.id]
		if top.next < len(children) {
			child := children[top.next]
			top.next++
			if !visited[child] {
				visited[child] = true
				stack = append(stack, &frame{id: child})
			}
			continue
		}
		postorder = append(postorder, top.id)
		stack = stack[:len(stack)-1]
	}

	for left, right := 0, len(postorder)-1; left < right; left, right = left+1, right-1 {
		postorder[left], postorder[right] = postorder[right], postorder[left]
	}
	return postorder
}

// ControlDominators reports, for a block, whether it is control-dependent on a branch whose test
// satisfies isControlValue.
//
// This is upstream's `createControlDominators` from the React compiler's `Dominator.ts`. It answers
// a different question from `UnconditionalBlocks` over the same tree, and the difference is why both
// live here: that one asks whether a block runs on every path, this one asks which branches decide
// whether it runs at all.
//
// `set-state-in-effect` is the caller. A setter inside `if (previousReference.current !== value)` is
// exempt upstream because the branch that decides whether it runs tests a ref, which is React's
// evidence that the effect is synchronizing against something render cannot see. The exemption has
// two halves and this is the second: the first, value taint, is the setter's own arguments being
// ref-derived, and it needs no control flow at all.
//
// The returned predicate caches per block, because a rule asks it once per setter call and several
// setters commonly share a block.
//
// Nil-safe: a nil function yields a predicate that answers false, which is the correct answer for a
// function with no blocks rather than a special case.
func ControlDominators(function *Function, isControlValue func(place Place) bool) func(block BlockId) bool {
	if function == nil || isControlValue == nil {
		return func(BlockId) bool { return false }
	}

	tree := computePostDominance(function)
	cache := map[BlockId]bool{}

	return func(block BlockId) bool {
		if answer, known := cache[block]; known {
			return answer
		}

		answer := false
		for frontierBlock := range postDominatorFrontier(function, tree, block) {
			source := blockById(function, frontierBlock)
			if source == nil {
				continue
			}
			if terminalTestSatisfies(source.Terminal, isControlValue) {
				answer = true
				break
			}
		}

		cache[block] = answer
		return answer
	}
}

// terminalTestSatisfies asks whether a terminal branches on a value the caller cares about.
//
// Three terminal kinds carry a test, matching upstream's switch over `if`, `branch` and `switch`. A
// switch is asked about its subject and about every case test, because `switch (reference.current)`
// and `case reference.current:` are both a branch on a ref, and upstream checks both.
//
// Every other terminal leaves control unconditionally, so a block downstream of one is not control
// dependent on it and there is nothing to test.
func terminalTestSatisfies(terminal Terminal, isControlValue func(place Place) bool) bool {
	switch typed := terminal.(type) {
	case *If:
		return isControlValue(typed.Test)
	case *Branch:
		return isControlValue(typed.Test)
	case *Switch:
		if isControlValue(typed.Test) {
			return true
		}
		for _, kase := range typed.Cases {
			if kase.Test != nil && isControlValue(*kase.Test) {
				return true
			}
		}
	}
	return false
}

// postDominatorFrontier is the set of blocks that decide whether target runs.
//
// Upstream's `postDominatorFrontier`. A predecessor of anything target post-dominates, which target
// does NOT post-dominate, is a block where control could have gone elsewhere: that is exactly a
// branch target depends on.
//
// The walk covers target itself as well as its post-dominated set, because a branch immediately
// above target is a frontier block and target does not post-dominate itself in `postDominatorsOf`'s
// result.
func postDominatorFrontier(function *Function, tree *postDominanceTree, target BlockId) map[BlockId]bool {
	postDominated := postDominatorsOf(function, tree, target)

	frontier := map[BlockId]bool{}
	visited := map[BlockId]bool{}

	walk := func(id BlockId) {
		if visited[id] {
			return
		}
		visited[id] = true

		block := blockById(function, id)
		if block == nil {
			return
		}
		for _, predecessor := range block.Predecessors {
			if !postDominated[predecessor] {
				frontier[predecessor] = true
			}
		}
	}

	for id := range postDominated {
		walk(id)
	}
	walk(target)

	return frontier
}

// postDominatorsOf is every block target post-dominates.
//
// Upstream's `postDominatorsOf`, walking predecessors from target and keeping a block when its own
// immediate post-dominator is target or is already known to be post-dominated by it. A block with no
// entry in the tree reaches no return, and upstream falls back to the block itself there, which this
// reproduces rather than skipping: the fallback makes such a block its own post-dominator, so it
// joins the set only if it IS the target.
func postDominatorsOf(function *Function, tree *postDominanceTree, target BlockId) map[BlockId]bool {
	result := map[BlockId]bool{}
	visited := map[BlockId]bool{}

	queue := []BlockId{target}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		block := blockById(function, current)
		if block == nil {
			continue
		}
		for _, predecessor := range block.Predecessors {
			immediate, known := tree.immediate[predecessor]
			if !known {
				immediate = predecessor
			}
			if immediate == target || result[immediate] {
				result[predecessor] = true
			}
			queue = append(queue, predecessor)
		}
	}

	return result
}
