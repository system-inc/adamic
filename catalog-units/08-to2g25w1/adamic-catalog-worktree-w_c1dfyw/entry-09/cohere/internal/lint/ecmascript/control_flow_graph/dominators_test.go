package control_flow_graph

import (
	"testing"
)

// bruteForceDominators computes dominance by the DEFINITION rather than by the algorithm: block D
// dominates block B when every path from the entry to B passes through D, established by deleting D
// from the graph and asking whether B is still reachable.
//
// This is the whole point of the file. Cooper-Harvey-Kennedy is an optimization of this definition,
// and an implementation of it can be subtly wrong — a bad reverse-postorder numbering, an intersect
// that walks the wrong side, an initialization that lets an undefined predecessor win — in ways that
// still produce a plausible-looking tree. Checking the tree against the definition is the only test
// that cannot share a bug with the thing it is testing, because it shares no code with it.
//
// It is O(blocks * edges) and would be wrong to ship inside a rule. It is exactly right in a test.
func bruteForceDominators[E any](graph *Graph[E], dominator, block *Block[E]) bool {
	if dominator == block {
		return true
	}
	if block.Index() == 0 {
		// Nothing other than the entry dominates the entry.
		return false
	}
	// Reachability from the entry with `dominator` deleted.
	seen := make([]bool, len(graph.Blocks))
	seen[dominator.Index()] = true
	stack := []*Block[E]{graph.Blocks[0]}
	if graph.Blocks[0] == dominator {
		// Deleting the entry makes everything unreachable, so the entry dominates every reachable
		// block.
		return true
	}
	seen[0] = true
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if current == block {
			// Reached without passing through `dominator`, so it does not dominate.
			return false
		}
		for _, successor := range current.Successors {
			if successor == nil || !successor.Reachable || seen[successor.Index()] {
				continue
			}
			seen[successor.Index()] = true
			stack = append(stack, successor)
		}
	}
	return true
}

// TestDominatorsMatchTheDefinition checks every ordered pair of reachable blocks in a spread of
// control-flow shapes against the brute-force definition.
//
// The shapes are chosen for the structures that break a dominator implementation rather than for
// source-language variety: a diamond (the intersect step), nested loops (back edges, where an
// undefined predecessor must not win the meet), a loop with two exits (a join whose predecessors sit
// at different reverse-postorder depths), a switch (a wide join), and try/catch/finally (the
// duplicated layout and the fork-to-handler edge, which is where this graph differs from a textbook
// one).
func TestDominatorsMatchTheDefinition(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name string
		code string
		// minimumPairs is how many reachable ordered pairs this shape must produce for the case to
		// be measuring anything. It defaults to 4 and is set explicitly for the shapes that
		// deliberately collapse to a single block.
		minimumPairs int
	}{
		// A straight-line body is ONE block: the builder starts a new one only at a fork, so
		// `a(); b(); c();` never splits. That is worth having in the table as the degenerate case
		// the analysis must not crash on, and its pair count is 1 rather than the >= 4 the other
		// shapes carry, which is why the vacuity guard below is per-case.
		{name: "straight line", code: `function f() { a(); b(); c(); }`, minimumPairs: 1},
		{name: "diamond", code: `function f(x) { if (x) a(); else b(); c(); }`},
		{name: "nested diamond", code: `function f(x, y) { if (x) { if (y) a(); else b(); } else c(); d(); }`},
		{name: "while loop", code: `function f(x) { while (x) { a(); } b(); }`},
		{name: "nested loops", code: `function f(x, y) { while (x) { while (y) { a(); } b(); } c(); }`},
		{name: "loop with two exits", code: `function f(x, y) { while (x) { if (y) break; a(); } b(); }`},
		{name: "loop with continue", code: `function f(x, y) { while (x) { if (y) continue; a(); } b(); }`},
		{name: "do while", code: `function f(x) { do { a(); } while (x); b(); }`},
		{name: "for of", code: `function f(xs) { for (const x of xs) { a(x); } b(); }`},
		{name: "switch", code: `function f(x) { switch (x) { case 0: a(); break; case 1: b(); break; default: c(); } d(); }`},
		{name: "switch fallthrough", code: `function f(x) { switch (x) { case 0: a(); case 1: b(); break; } c(); }`},
		{name: "try catch", code: `function f() { try { a(); } catch (e) { b(); } c(); }`},
		{name: "try finally", code: `function f() { try { a(); } finally { b(); } c(); }`},
		{name: "try catch finally", code: `function f() { try { a(); } catch (e) { b(); } finally { c(); } d(); }`},
		{name: "try with return", code: `function f() { try { return 1; } finally { b(); } }`},
		{name: "early return", code: `function f(x) { if (x) return; a(); }`},
		{name: "throw in branch", code: `function f(x) { if (x) throw x; a(); }`},
		{name: "short circuit", code: `function f(x, y) { if (x && y) a(); b(); }`},
		{name: "ternary", code: `function f(x) { const v = x ? a() : b(); c(v); }`},
		{name: "optional chain", code: `function f(x) { x?.a(); b(); }`},
		{name: "labelled break", code: `function f(x, y) { outer: while (x) { while (y) { break outer; } a(); } b(); }`},
		{name: "generator yield in loop", code: `function* f(x) { while (x) { a(); yield x; } b(); }`},
		{name: "infinite loop", code: `function f() { for (;;) { a(); } }`},
		// Everything after the `return` is unreachable, so exactly one block participates.
		{name: "unreachable tail", code: `function f() { return; a(); b(); }`, minimumPairs: 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			graph, _ := buildPathTestGraph(t, testCase.code)
			dominators := AnalyzeDominators(graph)

			checked := 0
			for _, dominator := range graph.Blocks {
				for _, block := range graph.Blocks {
					if !dominator.Reachable || !block.Reachable {
						continue
					}
					want := bruteForceDominators(graph, dominator, block)
					got := dominators.Dominates(dominator, block)
					if got != want {
						t.Errorf("Dominates(block %d, block %d) = %v, want %v",
							dominator.Index(), block.Index(), got, want)
					}
					checked++
				}
			}
			// A shape whose graph collapsed to fewer blocks than expected would pass this test
			// vacuously, so the count is asserted rather than assumed.
			wantPairs := testCase.minimumPairs
			if wantPairs == 0 {
				wantPairs = 4
			}
			if checked < wantPairs {
				t.Fatalf("only %d reachable pairs checked, want at least %d; the graph is smaller than this case assumes",
					checked, wantPairs)
			}
		})
	}
}

// TestDominatorsRejectABrokenTree is the failure-detection half of the pair above.
//
// A test that only ever sees a correct implementation says nothing about whether it could see an
// incorrect one. This mutates the settled dominator array — reparenting one node to the entry, which
// is the single most plausible real bug, since it is what a wrong initialization or a too-early
// convergence produces — and asserts that the definition check goes red.
//
// Without this, `TestDominatorsMatchTheDefinition` passing would be consistent with
// `bruteForceDominators` and `Dominates` both answering true for everything.
func TestDominatorsRejectABrokenTree(t *testing.T) {
	t.Parallel()
	graph, _ := buildPathTestGraph(t, `function f(x, y) { if (x) { if (y) a(); else b(); } else c(); d(); }`)
	dominators := AnalyzeDominators(graph)

	// Find one block whose immediate dominator is not already the entry, so reparenting it to the
	// entry is a real change.
	target := -1
	for _, block := range graph.Blocks {
		if !block.Reachable || block.Index() == 0 {
			continue
		}
		if dominators.immediate[block.Index()] > 0 {
			target = block.Index()
			break
		}
	}
	if target == -1 {
		t.Fatal("no block has a non-entry immediate dominator; this shape cannot detect the mutation")
	}

	original := dominators.immediate[target]
	dominators.immediate[target] = 0

	disagreements := 0
	for _, dominator := range graph.Blocks {
		for _, block := range graph.Blocks {
			if !dominator.Reachable || !block.Reachable {
				continue
			}
			if dominators.Dominates(dominator, block) != bruteForceDominators(graph, dominator, block) {
				disagreements++
			}
		}
	}
	if disagreements == 0 {
		t.Fatalf("reparenting block %d from %d to the entry changed no answer; the definition check cannot see a broken tree",
			target, original)
	}
}

// TestDominatorsOnKnownShapes pins the tree on shapes whose dominance a reader can cohere by eye,
// so a future change that keeps the brute-force check happy by breaking BOTH sides still fails here.
//
// These assert the property rather than block indices, because block numbering is construction
// order and the graph's own doc says no analysis should depend on it.
func TestDominatorsOnKnownShapes(t *testing.T) {
	t.Parallel()

	t.Run("the join after a diamond is not dominated by either arm", func(t *testing.T) {
		t.Parallel()
		graph, locations := buildPathTestGraph(t, `function f(x) { if (x) markLeft(); else markRight(); markJoin(); }`)
		dominators := AnalyzeDominators(graph)
		left := locations["markLeft"][0]
		right := locations["markRight"][0]
		join := locations["markJoin"][0]

		if dominators.Dominates(left, join) {
			t.Error("the then-arm must not dominate the join: the else-arm reaches it without passing through")
		}
		if dominators.Dominates(right, join) {
			t.Error("the else-arm must not dominate the join")
		}
		if !dominators.Dominates(graph.Blocks[0], join) {
			t.Error("the entry must dominate every reachable block")
		}
		if !dominators.Dominates(join, join) {
			t.Error("a block must dominate itself")
		}
		if dominators.StrictlyDominates(join, join) {
			t.Error("a block must not strictly dominate itself")
		}
	})

	t.Run("a loop body is dominated by its header and does not dominate the exit", func(t *testing.T) {
		t.Parallel()
		graph, locations := buildPathTestGraph(t, `function f(x) { while (x) { markBody(); } markAfter(); }`)
		dominators := AnalyzeDominators(graph)
		body := locations["markBody"][0]
		after := locations["markAfter"][0]

		if dominators.Dominates(body, after) {
			t.Error("a while body must not dominate the code after the loop: the loop may run zero times")
		}
		if !dominators.Dominates(graph.Blocks[0], body) {
			t.Error("the entry must dominate the loop body")
		}
	})

	t.Run("a do-while body DOES dominate the exit", func(t *testing.T) {
		t.Parallel()
		// The counterpart to the while case, and the one that separates a real dominator
		// computation from a syntactic guess: a `do` body runs at least once, so everything after
		// the loop is dominated by it.
		graph, locations := buildPathTestGraph(t, `function f(x) { do { markBody(); } while (x); markAfter(); }`)
		dominators := AnalyzeDominators(graph)
		body := locations["markBody"][0]
		after := locations["markAfter"][0]

		if !dominators.Dominates(body, after) {
			t.Error("a do-while body runs at least once, so it dominates the code after the loop")
		}
	})

	t.Run("an unreachable block is dominated by nothing and dominates nothing", func(t *testing.T) {
		t.Parallel()
		graph, locations := buildPathTestGraph(t, `function f() { markLive(); return; markDead(); }`)
		dominators := AnalyzeDominators(graph)
		live := locations["markLive"][0]
		dead := locations["markDead"][0]

		if dead.Reachable {
			t.Fatal("the statement after a return should be laid out unreachable")
		}
		if dominators.IsReachable(dead) {
			t.Error("an unreachable block must not participate in the dominator tree")
		}
		if dominators.Dominates(graph.Blocks[0], dead) {
			t.Error("the entry must not dominate an unreachable block")
		}
		if dominators.Dominates(dead, live) {
			t.Error("an unreachable block must dominate nothing")
		}
		if _, ok := dominators.ImmediateDominator(dead); ok {
			t.Error("an unreachable block has no immediate dominator")
		}
	})

	t.Run("the entry has no immediate dominator", func(t *testing.T) {
		t.Parallel()
		graph, _ := buildPathTestGraph(t, `function f() { a(); }`)
		dominators := AnalyzeDominators(graph)
		if _, ok := dominators.ImmediateDominator(graph.Blocks[0]); ok {
			t.Error("the entry block has no immediate dominator; returning itself would loop a caller walking upward")
		}
	})
}

// TestDominatorsAgreeWithFinalPathDominators pins this file's tree against the one
// `findFinalPathDominators` already computes in paths.go.
//
// The two are deliberately different computations over different graphs — that one adds a synthetic
// exit and excludes thrown edges — so they are not expected to agree in general. Where they must
// agree is on the entry's own dominance and on a graph with no thrown paths and one exit, where the
// synthetic exit changes nothing: there, a block on every final path is exactly a block that
// dominates the exit block.
//
// This is what makes duplicating Cooper-Harvey-Kennedy safe rather than a divergence waiting to
// happen. If either implementation drifts, this goes red.
func TestDominatorsAgreeWithFinalPathDominators(t *testing.T) {
	t.Parallel()
	for _, code := range []string{
		`function f(x) { if (x) a(); else b(); c(); }`,
		`function f(x) { if (x) a(); c(); }`,
		`function f(x) { while (x) { a(); } b(); }`,
		`function f(x) { switch (x) { case 0: a(); break; default: b(); } c(); }`,
		`function f(x, y) { while (x) { if (y) break; a(); } b(); }`,
	} {
		t.Run(code, func(t *testing.T) {
			t.Parallel()
			graph, _ := buildPathTestGraph(t, code)
			paths := AnalyzePaths(graph)
			dominators := AnalyzeDominators(graph)

			if len(graph.ThrownBlocks) != 0 {
				t.Skipf("this shape has thrown exits, where the two computations are defined to differ")
			}
			if len(graph.FinalBlocks) != 1 {
				t.Skipf("this shape has %d final blocks; the equivalence holds for one", len(graph.FinalBlocks))
			}
			exit := graph.FinalBlocks[0]

			for _, block := range graph.Blocks {
				if !block.Reachable {
					continue
				}
				onEveryFinalPath := paths.IsOnEveryFinalPath(block)
				dominatesExit := dominators.Dominates(block, exit)
				if onEveryFinalPath != dominatesExit {
					t.Errorf("block %d: IsOnEveryFinalPath = %v but Dominates(block, exit) = %v",
						block.Index(), onEveryFinalPath, dominatesExit)
				}
			}
		})
	}
}

// TestPredecessorsAreTheInverseOfSuccessors checks the edge list both ways round on every shape,
// which is the only property predecessors have to satisfy.
//
// It is written as a two-directional check rather than a count, because a predecessor array built
// with an off-by-one in its counting sort produces the right TOTAL and the wrong grouping, and a
// count-only test passes on it. This one does not.
func TestPredecessorsAreTheInverseOfSuccessors(t *testing.T) {
	t.Parallel()
	for _, code := range []string{
		`function f(x) { if (x) a(); else b(); c(); }`,
		`function f(x, y) { while (x) { while (y) { a(); } b(); } c(); }`,
		`function f(x) { switch (x) { case 0: a(); case 1: b(); break; default: c(); } d(); }`,
		`function f() { try { a(); } catch (e) { b(); } finally { c(); } d(); }`,
		// A shape with reachable edges AND an unreachable block, so the "unreachable blocks have
		// no predecessors" branch is reached without collapsing the graph to a single block.
		`function f(x) { if (x) { return; a(); } b(); }`,
	} {
		t.Run(code, func(t *testing.T) {
			t.Parallel()
			graph, _ := buildPathTestGraph(t, code)
			dominators := AnalyzeDominators(graph)

			// Every predecessor edge this reports must exist as a successor edge.
			edges := 0
			for _, block := range graph.Blocks {
				if !block.Reachable {
					if len(dominators.Predecessors(block)) != 0 {
						t.Errorf("block %d is unreachable and must have no predecessors", block.Index())
					}
					continue
				}
				for _, predecessor := range dominators.Predecessors(block) {
					found := false
					for _, successor := range predecessor.Successors {
						if successor == block {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("block %d lists block %d as a predecessor, but block %d has no successor edge to it",
							block.Index(), predecessor.Index(), predecessor.Index())
					}
					if !predecessor.Reachable {
						t.Errorf("block %d lists unreachable block %d as a predecessor", block.Index(), predecessor.Index())
					}
					edges++
				}
			}

			// And every reachable successor edge must appear as a predecessor edge. This is the
			// direction that catches a dropped group.
			expected := 0
			for _, block := range graph.Blocks {
				if !block.Reachable {
					continue
				}
				for _, successor := range block.Successors {
					if successor == nil || !successor.Reachable {
						continue
					}
					expected++
					found := false
					for _, predecessor := range dominators.Predecessors(successor) {
						if predecessor == block {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("block %d has a successor edge to block %d, which does not list it as a predecessor",
							block.Index(), successor.Index())
					}
				}
			}
			if edges != expected {
				t.Errorf("predecessor edges = %d, reachable successor edges = %d", edges, expected)
			}
			if expected == 0 {
				t.Fatal("this shape has no reachable edges; it is testing nothing")
			}
		})
	}
}

// TestReversePostOrderPlacesPredecessorsFirst checks the ordering property a forward dataflow
// depends on: every block appears after at least one predecessor, except the entry and the target of
// a back edge.
//
// A back edge is exactly an edge to a block already on the current DFS path, so the property is
// stated as "at least one predecessor appears earlier" rather than "all do", which loops violate by
// construction.
func TestReversePostOrderPlacesPredecessorsFirst(t *testing.T) {
	t.Parallel()
	for _, code := range []string{
		`function f(x) { if (x) a(); else b(); c(); }`,
		`function f(x, y) { while (x) { while (y) { a(); } b(); } c(); }`,
		`function f() { try { a(); } catch (e) { b(); } finally { c(); } d(); }`,
	} {
		t.Run(code, func(t *testing.T) {
			t.Parallel()
			graph, _ := buildPathTestGraph(t, code)
			dominators := AnalyzeDominators(graph)
			order := dominators.ReversePostOrder()

			if len(order) == 0 {
				t.Fatal("reverse postorder is empty")
			}
			if order[0] != graph.Blocks[0] {
				t.Error("reverse postorder must start at the entry block")
			}

			position := map[*Block[string]]int{}
			for index, block := range order {
				position[block] = index
			}
			for _, block := range order {
				if block.Index() == 0 {
					continue
				}
				predecessors := dominators.Predecessors(block)
				if len(predecessors) == 0 {
					t.Errorf("block %d is in reverse postorder with no predecessors", block.Index())
					continue
				}
				earlier := false
				for _, predecessor := range predecessors {
					if position[predecessor] < position[block] {
						earlier = true
						break
					}
				}
				if !earlier {
					t.Errorf("block %d appears before all of its predecessors; only a back-edge target may, and it has %d predecessors",
						block.Index(), len(predecessors))
				}
			}

			// Every reachable block appears exactly once, and no unreachable one appears.
			if len(position) != len(order) {
				t.Error("reverse postorder contains a duplicate block")
			}
			for _, block := range graph.Blocks {
				_, listed := position[block]
				if block.Reachable != listed {
					t.Errorf("block %d: reachable = %v but listed in reverse postorder = %v",
						block.Index(), block.Reachable, listed)
				}
			}
		})
	}
}
