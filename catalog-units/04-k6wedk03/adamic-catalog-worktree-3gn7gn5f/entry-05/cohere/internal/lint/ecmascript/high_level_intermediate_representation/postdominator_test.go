package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/static_single_assignment"
	"testing"
)

// The post-dominator tree is checked against the DEFINITION rather than against the algorithm.
//
// A block B post-dominates a block A when every path from A to a `Return` passes through B. That is
// established here by deleting B from the graph and asking whether any `Return` is still reachable
// from A, using a plain reachability walk that shares no code and no bug with
// `computePostDominance`. An implementation checked against a second copy of itself proves only
// that the copy was faithful.
//
// This matters more here than it would for a helper of the same size elsewhere. A post-dominator
// tree that is built but wrong does not fail loudly: it answers "this instruction runs
// unconditionally" with confidence, and the rule consuming it turns that into a diagnostic on
// somebody's code. Cooper-Harvey-Kennedy in particular degrades quietly when its node ordering is
// wrong, because `intersect` walks two positions toward each other and a bad order makes it settle
// on a node that dominates neither input.

// reachesAReturnWithout reports whether any `Return` terminal is reachable from start without
// passing through excluded.
//
// Deliberately a naive walk over `EachSuccessor` with a visited set. It is the definition written
// out, and its only job is to disagree with the tree when the tree is wrong.
func reachesAReturnWithout(function *Function, start, excluded static_single_assignment.BlockId) bool {
	if start == excluded {
		return false
	}
	visited := map[static_single_assignment.BlockId]bool{start: true}
	stack := []static_single_assignment.BlockId{start}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		block, ok := function.Block(id)
		if !ok {
			continue
		}
		if _, isReturn := block.Terminal.(*Return); isReturn {
			return true
		}
		EachSuccessor(block.Terminal, func(next static_single_assignment.BlockId) {
			if next == excluded || visited[next] {
				return
			}
			visited[next] = true
			stack = append(stack, next)
		})
	}
	return false
}

// reachesAReturn is the same walk with nothing excluded, used to skip blocks for which the question
// is vacuous.
func reachesAReturn(function *Function, start static_single_assignment.BlockId) bool {
	return reachesAReturnWithout(function, start, InvalidBlock)
}

// postDominates answers the tree: is candidate on the post-dominator chain rising from block.
func postDominates(tree *postDominanceTree, block, candidate static_single_assignment.BlockId) bool {
	for current, steps := block, 0; steps <= len(tree.immediate)+1; steps++ {
		if current == candidate {
			return true
		}
		next, ok := tree.immediate[current]
		if !ok {
			return false
		}
		current = next
	}
	return false
}

// postDominatorShapes are the graphs the definition check runs over.
//
// `minimumPairs` is a vacuity guard, and it is not decoration. It fired on the first run here and
// caught three cases checking nothing: `a(); b(); return c;`, a sequence expression, and an optional
// chain all lower to a SINGLE block, so there is no ordered pair to ask about and every assertion
// over them passes for free. They were removed rather than given a zero threshold, because a case
// kept at zero is the vacuity the guard exists to catch, written down and blessed. The same guard
// caught two such cases on the first run of the forward dominator work.
//
// Every count below is the measured number of pairs for that shape, not a lower bound picked by
// eye. A shape that lowers differently after a change to lowering fails here rather than quietly
// checking less.
var postDominatorShapes = []struct {
	name         string
	code         string
	minimumPairs int
}{
	{name: "if", code: `function f(p) { if (p) { a(); } return b; }`, minimumPairs: 6},
	{name: "if else", code: `function f(p) { if (p) { a(); } else { b(); } return c; }`, minimumPairs: 12},
	{name: "early return", code: `function f(p) { if (p) { return 1; } a(); return 2; }`, minimumPairs: 6},
	{name: "both arms return", code: `function f(p) { if (p) { return 1; } else { return 2; } }`, minimumPairs: 9},
	{name: "nested if", code: `function f(p, q) { if (p) { if (q) { a(); } b(); } return c; }`, minimumPairs: 12},
	{name: "while", code: `function f(p) { while (p) { a(); } return b; }`, minimumPairs: 12},
	{name: "while with break", code: `function f(p) { while (p) { if (p) { break; } a(); } return b; }`, minimumPairs: 20},
	{name: "for of", code: `function f(p) { for (const x of p) { a(x); } return b; }`, minimumPairs: 12},
	{name: "for of empty body", code: `function f(p) { for (const x of p) {} return b; }`, minimumPairs: 12},
	{name: "throw in one arm", code: `function f(p) { if (p) { throw new Error(); } a(); return b; }`, minimumPairs: 4},
	{name: "loop with throw arm", code: `function f(p) { for (const x of p) { if (p) { break; } else { throw new Error(); } } return b; }`, minimumPairs: 12},
	{name: "try catch", code: `function f() { try { a(); } catch (e) { b(); } return c; }`, minimumPairs: 6},
	{name: "switch", code: `function f(p) { switch (p) { case 1: a(); break; default: b(); } return c; }`, minimumPairs: 12},
	{name: "ternary", code: `function f(p) { const x = p ? a() : b(); return x; }`, minimumPairs: 12},
	{name: "logical and", code: `function f(p) { const x = p && a(); return x; }`, minimumPairs: 12},
	{name: "do while", code: `function f(p) { do { a(); } while (p); return b; }`, minimumPairs: 12},
	{name: "labelled break", code: `function f(p) { outer: for (const x of p) { break outer; } return b; }`, minimumPairs: 20},
	{name: "two returns", code: `function f(p) { if (p) { return 1; } return 2; }`, minimumPairs: 6},
	{name: "infinite loop", code: `function f() { while (true) { a(); } }`, minimumPairs: 12},
	{name: "return inside loop", code: `function f(p) { while (p) { return 1; } return 2; }`, minimumPairs: 12},
	{name: "nested loops", code: `function f(p, q) { for (const x of p) { for (const y of q) { a(); } } return b; }`, minimumPairs: 20},
}

// TestPostDominatorsMatchTheDefinition checks every ordered pair of blocks in every shape against a
// reachability walk that shares no code with the tree.
func TestPostDominatorsMatchTheDefinition(t *testing.T) {
	t.Parallel()

	for _, shape := range postDominatorShapes {
		t.Run(shape.name, func(t *testing.T) {
			t.Parallel()
			function := lowerSource(t, shape.code)
			tree := computePostDominance(function)

			pairs := 0
			for _, from := range function.Blocks {
				// The question is vacuous for a block that cannot reach a return at all: nothing
				// post-dominates it and the walk agrees trivially. Skipping keeps the pair count
				// honest rather than inflating it with cases that cannot discriminate.
				if !reachesAReturn(function, from.Id) {
					continue
				}
				for _, candidate := range function.Blocks {
					if from.Id == candidate.Id {
						continue
					}
					pairs++
					wanted := !reachesAReturnWithout(function, from.Id, candidate.Id)
					got := postDominates(tree, from.Id, candidate.Id)
					if got != wanted {
						t.Errorf(
							"bb%d post-dominated by bb%d: tree says %v, deleting bb%d and walking says %v\n%s",
							from.Id, candidate.Id, got, candidate.Id, wanted, Print(function))
					}
				}
			}
			if pairs < shape.minimumPairs {
				t.Fatalf("only %d ordered pairs checked, want at least %d; this shape is smaller than the case assumes and would pass vacuously",
					pairs, shape.minimumPairs)
			}
		})
	}
}

// TestPostDominatorDefinitionCheckCanFail proves the check above is capable of failing.
//
// A test that never fails is indistinguishable from a test that cannot fail, and the forward
// dominator work found two of its own cases asserting nothing on the first run. Here the tree is
// deliberately corrupted by reparenting one node, and the definition check must notice.
//
// The corruption is chosen to be the kind a real bug produces: not a nil tree or an empty map,
// which any assertion catches, but a plausible tree with one edge redirected, which is exactly what
// a wrong node ordering in Cooper-Harvey-Kennedy yields.
func TestPostDominatorDefinitionCheckCanFail(t *testing.T) {
	t.Parallel()

	// A corruption is only detectable where there was structure to destroy. Self-parenting a block
	// removes every ancestor from its chain, so it changes an answer only when the chain held a
	// REAL block; a block whose immediate post-dominator is already the synthetic exit has nothing
	// above it and reparenting it is invisible.
	//
	// That is a true property of post-dominance rather than a weakness in the check, and it is why
	// this counts detections against corruptions-with-content rather than against every node. A
	// function with two returns has no real block on any chain at all, because both returns are
	// separate ways out, so `if (p) { return 1; } return 2;` legitimately offers nothing to corrupt.
	// The first version of this test demanded a detection from those shapes and failed on three of
	// them, which is how the distinction was found.
	shapesWithContent := 0

	for _, shape := range postDominatorShapes {
		// Not parallel: its subtests add to totals the test checks after them, and a parallel subtest would run
		// only after the test had returned, so the check would pass on nothing.
		t.Run(shape.name, func(t *testing.T) {
			function := lowerSource(t, shape.code)
			pristine := computePostDominance(function)

			withContent := 0
			detected := 0
			for _, block := range function.Blocks {
				parent, ok := pristine.immediate[block.Id]
				if !ok || parent == postDominatorExit {
					// Nothing real above this block, so severing its chain removes nothing.
					continue
				}
				withContent++

				// Rebuild each round so corruptions do not accumulate, then reparent this one node
				// onto itself. That leaves a structure that is still a well-formed tree, which is
				// what a wrong node ordering in Cooper-Harvey-Kennedy actually produces. A nil map
				// or an empty tree would be caught by any assertion and would prove nothing.
				tree := computePostDominance(function)
				tree.immediate[block.Id] = block.Id

				disagreements := 0
				for _, from := range function.Blocks {
					if !reachesAReturn(function, from.Id) {
						continue
					}
					for _, candidate := range function.Blocks {
						if from.Id == candidate.Id {
							continue
						}
						wanted := !reachesAReturnWithout(function, from.Id, candidate.Id)
						if postDominates(tree, from.Id, candidate.Id) != wanted {
							disagreements++
						}
					}
				}
				if disagreements > 0 {
					detected++
				}
			}

			if withContent == 0 {
				// Legitimate for a multi-return shape. Recorded rather than failed, and the
				// aggregate assertion below makes sure this is not every shape.
				t.Logf("no block has a real post-dominator, so there is nothing to corrupt here")
				return
			}
			if detected != withContent {
				t.Errorf("only %d of %d corruptions with real content were detected; the definition check is partly blind\n%s",
					detected, withContent, Print(function))
			}
			shapesWithContent++
			t.Logf("%d of %d corruptions detected", detected, withContent)
		})
	}

	if shapesWithContent == 0 {
		t.Fatal("no shape offered a corruptible post-dominator chain; the falsification check is inert")
	}
}

// TestUnconditionalBlocksOnMeasuredShapes pins the exported entry point against verdicts taken from
// React's own rule rather than from this implementation.
//
// Each `want` below is the set of blocks holding an instruction React reports on, established by
// running `eslint-plugin-react-hooks`'s `set-state-in-render` on the equivalent input. They are
// recorded as membership questions about the block a call lands in, which is what the rule asks.
func TestUnconditionalBlocksOnMeasuredShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		code string
		// wantUnconditional is whether the block holding the LAST call in the body runs on every
		// path. That is the question the rule asks and the one React's verdict answers.
		wantUnconditional bool
	}{
		// React reports: the call is in the entry block.
		{name: "straight", code: `function f() { const s = g(); s(1); return 0; }`, wantUnconditional: true},
		// React is clean: a conditional call.
		{name: "conditional", code: `function f(p) { const s = g(); if (p) { s(1); } return 0; }`, wantUnconditional: false},
		// React is clean: an early return means the tail is not on every path.
		{name: "early return", code: `function f(p) { const s = g(); if (p) { return 1; } s(1); return 0; }`, wantUnconditional: false},
		// React reports: the loop always terminates into the tail.
		{name: "after empty loop", code: `function f(p) { const s = g(); for (const x of p) {} s(1); return 0; }`, wantUnconditional: true},
		// React reports: the throwing arm is not an exit, so the break is the only way out.
		{name: "after loop with throw arm", code: `function f(p) { const s = g(); for (const x of p) { if (p) { break; } else { throw new Error(); } } s(1); return 0; }`, wantUnconditional: true},
		// React is clean: a loop body is conditional.
		{name: "inside loop", code: `function f(p) { const s = g(); for (const x of p) { s(1); } return 0; }`, wantUnconditional: false},
		// React is clean: a try body may not complete.
		{name: "inside try", code: `function f() { const s = g(); try { s(1); } catch (e) {} return 0; }`, wantUnconditional: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function := lowerSource(t, testCase.code)
			unconditional := UnconditionalBlocks(function)

			block, ok := blockHoldingLastCall(function)
			if !ok {
				t.Fatalf("no call instruction found, so this case asserts nothing\n%s", Print(function))
			}
			if got := unconditional[block]; got != testCase.wantUnconditional {
				t.Errorf("block bb%d holding the last call: unconditional = %v, want %v\n%s",
					block, got, testCase.wantUnconditional, Print(function))
			}
		})
	}
}

// blockHoldingLastCall finds the block containing the last CallExpression in evaluation order.
func blockHoldingLastCall(function *Function) (static_single_assignment.BlockId, bool) {
	best := InvalidBlock
	var bestOrder static_single_assignment.EvaluationOrder
	found := false
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			if _, isCall := instruction.Value.(*CallExpression); !isCall {
				continue
			}
			if !found || instruction.Order > bestOrder {
				best, bestOrder, found = block.Id, instruction.Order, true
			}
		}
	}
	return best, found
}

// TestUnconditionalBlocksAlwaysContainsTheEntry checks the one invariant that holds for every
// function: control begins at the entry, so the entry always runs.
//
// It is separate from the definition check because it is true even for a function with no return at
// all, where the post-dominator tree is empty and every other question is vacuous.
func TestUnconditionalBlocksAlwaysContainsTheEntry(t *testing.T) {
	t.Parallel()

	for _, shape := range postDominatorShapes {
		t.Run(shape.name, func(t *testing.T) {
			t.Parallel()
			function := lowerSource(t, shape.code)
			unconditional := UnconditionalBlocks(function)
			if !unconditional[function.Entry] {
				t.Errorf("the entry block bb%d is not unconditional\n%s", function.Entry, Print(function))
			}
		})
	}
}

// TestUnconditionalBlocksIsASubsetOfPostDominators states the relationship between the two
// functions in this file, so a change to one that silently diverges from the other is caught.
func TestUnconditionalBlocksIsASubsetOfPostDominators(t *testing.T) {
	t.Parallel()

	for _, shape := range postDominatorShapes {
		t.Run(shape.name, func(t *testing.T) {
			t.Parallel()
			function := lowerSource(t, shape.code)
			tree := computePostDominance(function)
			for id := range UnconditionalBlocks(function) {
				if !postDominates(tree, function.Entry, id) {
					t.Errorf("bb%d is reported unconditional but does not post-dominate the entry bb%d\n%s",
						id, function.Entry, Print(function))
				}
			}
		})
	}
}

// TestUnconditionalBlocksHandlesDegenerateInput checks the two shapes a caller can produce that are
// not functions at all.
func TestUnconditionalBlocksHandlesDegenerateInput(t *testing.T) {
	t.Parallel()

	if got := UnconditionalBlocks(nil); got != nil {
		t.Errorf("UnconditionalBlocks(nil) = %v, want nil", got)
	}
	if got := UnconditionalBlocks(&Function{}); got != nil {
		t.Errorf("UnconditionalBlocks on a function with no blocks = %v, want nil", got)
	}
}
