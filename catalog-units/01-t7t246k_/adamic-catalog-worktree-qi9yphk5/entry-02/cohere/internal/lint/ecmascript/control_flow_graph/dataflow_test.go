package control_flow_graph

import (
	"sort"
	"strings"
	"testing"
)

// reachingNames is a Forward "may" analysis: the set of event names that could have been seen on
// some path to this point. Meet is union, so it is the simplest lattice that still exercises a real
// join, and its answer on any shape is checkable by reading the source.
type reachingNames struct{}

func (reachingNames) Bottom() map[string]bool { return map[string]bool{} }
func (reachingNames) Entry() map[string]bool  { return map[string]bool{} }

func (reachingNames) Meet(left, right map[string]bool) map[string]bool {
	merged := make(map[string]bool, len(left)+len(right))
	for name := range left {
		merged[name] = true
	}
	for name := range right {
		merged[name] = true
	}
	return merged
}

func (reachingNames) Transfer(block *Block[string], incoming map[string]bool) map[string]bool {
	outgoing := make(map[string]bool, len(incoming)+len(block.Events))
	for name := range incoming {
		outgoing[name] = true
	}
	for _, event := range block.Events {
		outgoing[event] = true
	}
	return outgoing
}

func (reachingNames) Equal(left, right map[string]bool) bool {
	if len(left) != len(right) {
		return false
	}
	for name := range left {
		if !right[name] {
			return false
		}
	}
	return true
}

// mustReachNames is the "must" counterpart: the names seen on EVERY path to this point. Meet is
// intersection, so Bottom has to be the universe rather than the empty set, which is the case the
// LatticeOf doc warns reads backward.
type mustReachNames struct{ universe map[string]bool }

func (m mustReachNames) Bottom() map[string]bool {
	all := make(map[string]bool, len(m.universe))
	for name := range m.universe {
		all[name] = true
	}
	return all
}

func (mustReachNames) Entry() map[string]bool { return map[string]bool{} }

func (mustReachNames) Meet(left, right map[string]bool) map[string]bool {
	merged := map[string]bool{}
	for name := range left {
		if right[name] {
			merged[name] = true
		}
	}
	return merged
}

func (mustReachNames) Transfer(block *Block[string], incoming map[string]bool) map[string]bool {
	outgoing := make(map[string]bool, len(incoming)+len(block.Events))
	for name := range incoming {
		outgoing[name] = true
	}
	for _, event := range block.Events {
		outgoing[event] = true
	}
	return outgoing
}

func (mustReachNames) Equal(left, right map[string]bool) bool {
	return reachingNames{}.Equal(left, right)
}

func sortedNames(set map[string]bool) string {
	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ",")
}

// TestForwardMayAnalysis checks that a union-meet forward dataflow reaches the answer a reader gets
// from the source.
func TestForwardMayAnalysis(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name string
		code string
		at   string
		want string
	}{
		{
			name: "both arms of a diamond reach the join",
			code: `function f(x) { if (x) markLeft(); else markRight(); markJoin(); }`,
			at:   "markJoin",
			want: "markLeft,markRight",
		},
		{
			name: "an if with no else contributes only on one path",
			code: `function f(x) { if (x) markThen(); markAfter(); }`,
			at:   "markAfter",
			want: "markThen",
		},
		{
			name: "a loop body reaches the code after it",
			code: `function f(x) { while (x) { markBody(); } markAfter(); }`,
			at:   "markAfter",
			want: "markBody",
		},
		{
			// The back edge is what makes this a fixed point rather than one pass: on the second
			// iteration `markBody` has already run, so it reaches its own block. A solver that
			// stopped after one sweep would answer "markFirst" alone.
			name: "a loop body reaches itself through the back edge",
			code: `function f(x) { while (x) { markFirst(); markBody(); } }`,
			at:   "markBody",
			want: "markBody,markFirst",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			graph, locations := buildPathTestGraph(t, testCase.code)
			solution := Solve[map[string]bool, string](graph, Forward, reachingNames{})
			block := locations[testCase.at][0]
			incoming, ok := solution.In(block)
			if !ok {
				t.Fatalf("block holding %q is not in the solution", testCase.at)
			}
			// The block's own events are in Out, not In, so the event named by `at` is expected to
			// be absent from In unless a back edge brings it round.
			if got := sortedNames(incoming); got != testCase.want {
				t.Errorf("names reaching %q = %q, want %q", testCase.at, got, testCase.want)
			}
		})
	}
}

// TestForwardMustAnalysis checks the intersection-meet direction, which is the shape
// `constructor-super` and `no-this-before-super` describe in their doc comments: has `super()`
// definitely been called on every path to here.
func TestForwardMustAnalysis(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name string
		code string
		at   string
		want string
	}{
		{
			name: "an if with both arms establishes the name",
			code: `function f(x) { if (x) markSuper(); else markSuper(); markAfter(); }`,
			at:   "markAfter",
			want: "markSuper",
		},
		{
			name: "an if with one arm does not",
			code: `function f(x) { if (x) markSuper(); markAfter(); }`,
			at:   "markAfter",
			want: "",
		},
		{
			name: "a while body never establishes anything, because it may run zero times",
			code: `function f(x) { while (x) { markSuper(); } markAfter(); }`,
			at:   "markAfter",
			want: "",
		},
		{
			name: "a do-while body does establish it, because it runs at least once",
			code: `function f(x) { do { markSuper(); } while (x); markAfter(); }`,
			at:   "markAfter",
			want: "markSuper",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			graph, locations := buildPathTestGraph(t, testCase.code)
			universe := map[string]bool{}
			for _, block := range graph.Blocks {
				for _, event := range block.Events {
					universe[event] = true
				}
			}
			solution := Solve[map[string]bool, string](graph, Forward, mustReachNames{universe: universe})
			block := locations[testCase.at][0]
			incoming, ok := solution.In(block)
			if !ok {
				t.Fatalf("block holding %q is not in the solution", testCase.at)
			}
			if got := sortedNames(incoming); got != testCase.want {
				t.Errorf("names definitely reaching %q = %q, want %q", testCase.at, got, testCase.want)
			}
		})
	}
}

// liveNames is a Backward "may" analysis: the set of event names that appear anywhere later on some
// path. It is the shape of liveness with the symbol table replaced by names, so it exercises the
// backward direction and the exit seeding without dragging in a checker.
type liveNames struct{}

func (liveNames) Bottom() map[string]bool { return map[string]bool{} }
func (liveNames) Entry() map[string]bool  { return map[string]bool{} }

func (liveNames) Meet(left, right map[string]bool) map[string]bool {
	return reachingNames{}.Meet(left, right)
}

func (liveNames) Transfer(block *Block[string], incoming map[string]bool) map[string]bool {
	outgoing := make(map[string]bool, len(incoming)+len(block.Events))
	for name := range incoming {
		outgoing[name] = true
	}
	for _, event := range block.Events {
		outgoing[event] = true
	}
	return outgoing
}

func (liveNames) Equal(left, right map[string]bool) bool {
	return reachingNames{}.Equal(left, right)
}

// TestBackwardAnalysis checks the backward direction, including that the exit seeding reaches every
// terminal block.
func TestBackwardAnalysis(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name string
		code string
		at   string
		want string
	}{
		{
			name: "both arms of a diamond are live before the branch",
			code: `function f(x) { markBefore(); if (x) markLeft(); else markRight(); }`,
			at:   "markBefore",
			want: "markLeft,markRight",
		},
		{
			name: "nothing is live after the last statement",
			code: `function f(x) { if (x) markLeft(); markLast(); }`,
			at:   "markLast",
			want: "",
		},
		{
			name: "a loop body is live before the loop",
			code: `function f(x) { markBefore(); while (x) { markBody(); } }`,
			at:   "markBefore",
			want: "markBody",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			graph, locations := buildPathTestGraph(t, testCase.code)
			solution := Solve[map[string]bool, string](graph, Backward, liveNames{})
			block := locations[testCase.at][0]
			// For a Backward analysis, In is the value on EXIT from the block: what is live after it.
			outgoing, ok := solution.In(block)
			if !ok {
				t.Fatalf("block holding %q is not in the solution", testCase.at)
			}
			if got := sortedNames(outgoing); got != testCase.want {
				t.Errorf("names live after %q = %q, want %q", testCase.at, got, testCase.want)
			}
		})
	}
}

// TestSolveExcludesUnreachableBlocks pins the exclusion every existing consumer already wanted, and
// which `no-useless-assignment`'s doc comment explains in its own terms: an unreachable block's
// events are the same source positions laid out a second time.
func TestSolveExcludesUnreachableBlocks(t *testing.T) {
	t.Parallel()
	graph, locations := buildPathTestGraph(t, `function f() { markLive(); return; markDead(); }`)
	solution := Solve[map[string]bool, string](graph, Forward, reachingNames{})

	dead := locations["markDead"][0]
	if dead.Reachable {
		t.Fatal("the statement after a return should be laid out unreachable")
	}
	if _, ok := solution.In(dead); ok {
		t.Error("an unreachable block must not appear in the solution")
	}
	if _, ok := solution.Out(dead); ok {
		t.Error("an unreachable block must not appear in the solution")
	}

	live := locations["markLive"][0]
	outgoing, ok := solution.Out(live)
	if !ok {
		t.Fatal("a reachable block must appear in the solution")
	}
	if outgoing["markDead"] {
		t.Error("an unreachable block's events must not contribute to a reachable block's value")
	}
}

// TestSolveSeedsEveryTerminalBlockBackward pins the property `Solve`'s backward seeding rests on.
//
// `try { return 1; } finally { beta(); }` lays the `finally` out TWICE. One copy sits on the normal
// path; the other sits on the abrupt path and is a TERMINAL block with no reachable successors. If
// that copy were in neither FinalBlocks nor ThrownBlocks, a backward analysis seeding only from
// those two sets would start it at Bottom rather than Entry and propagate a value nothing
// established.
//
// Measured: it IS in an exit set, on this shape and on five others probed alongside it, so the
// seeding is complete without a third clause. This asserts that rather than assuming it, because it
// is a property of the BUILDER — if `markFinal`/`markThrown` ever stop being called on the abrupt
// copy, the gap opens silently in `Solve` and is named here instead.
func TestSolveSeedsEveryTerminalBlockBackward(t *testing.T) {
	t.Parallel()
	graph, locations := buildPathTestGraph(t, `
		function f() {
			try { markTry(); return 1; } finally { markFinally(); }
		}
	`)

	copies := locations["markFinally"]
	if len(copies) < 2 {
		t.Fatalf("the finally block should be laid out twice, got %d copies", len(copies))
	}

	inExitSet := map[int]bool{}
	for _, block := range graph.FinalBlocks {
		inExitSet[block.Index()] = true
	}
	for _, block := range graph.ThrownBlocks {
		inExitSet[block.Index()] = true
	}

	// The shape has to actually contain a terminal block for this to be measuring anything.
	terminals := 0
	for _, block := range graph.Blocks {
		if !block.Reachable {
			continue
		}
		hasSuccessor := false
		for _, successor := range block.Successors {
			if successor != nil && successor.Reachable {
				hasSuccessor = true
				break
			}
		}
		if !hasSuccessor {
			terminals++
			if !inExitSet[block.Index()] {
				t.Errorf("reachable terminal block %d is in neither FinalBlocks nor ThrownBlocks, so a backward analysis would not seed it",
					block.Index())
			}
		}
	}
	if terminals == 0 {
		t.Fatal("this shape has no reachable terminal block; it is testing nothing")
	}

	// And every reachable block, terminal or not, has a value.
	solution := Solve[map[string]bool, string](graph, Backward, liveNames{})
	for _, block := range graph.Blocks {
		if !block.Reachable {
			continue
		}
		if _, ok := solution.In(block); !ok {
			t.Errorf("reachable block %d has no backward solution value", block.Index())
		}
	}
}

// TestSolveConvergesInFewRoundsOnAcyclicGraphs is the measurement that says the reverse-postorder
// iteration is doing its job, in BOTH directions.
//
// A forward analysis over an acyclic graph in reverse postorder settles in exactly two rounds: one
// to propagate everything, one to observe nothing moved. A backward analysis over the same graph
// settles in two rounds only if the order is reverse postorder REVERSED; walking it forward instead
// still converges, because a fixed point does not depend on the order it is reached in, but it takes
// an extra round per level of depth.
//
// The backward half is here because a mutation reversing that reversal SURVIVED every value
// assertion in this file. Measured: with the order un-reversed, `markBefore(); if (x) ... ` goes
// 2 rounds to 3, `markBefore(); while (x) ...` goes 3 to 4, and a doubly-nested loop goes 4 to 6,
// while the settled live sets are bit-for-bit identical in all three. The answer is genuinely
// unchanged, so no value assertion can ever see it, and the round count is the only observable that
// separates the two versions. That makes this a real test rather than a performance nicety.
func TestSolveConvergesInFewRoundsOnAcyclicGraphs(t *testing.T) {
	t.Parallel()
	graph, _ := buildPathTestGraph(t, `
		function f(a, b, c) {
			if (a) markOne(); else markTwo();
			if (b) markThree(); else markFour();
			if (c) markFive(); else markSix();
			markSeven();
		}
	`)
	if len(graph.Blocks) < 8 {
		t.Fatalf("this shape has only %d blocks; 2 rounds would be unremarkable", len(graph.Blocks))
	}
	if rounds := Solve[map[string]bool, string](graph, Forward, reachingNames{}).Rounds(); rounds != 2 {
		t.Errorf("an acyclic forward analysis in reverse postorder should settle in 2 rounds, got %d", rounds)
	}
	if rounds := Solve[map[string]bool, string](graph, Backward, liveNames{}).Rounds(); rounds != 2 {
		t.Errorf("an acyclic backward analysis in reversed reverse postorder should settle in 2 rounds, got %d", rounds)
	}
}

// TestSolveBackwardOrderIsReversed pins the round counts on the cyclic shapes where the two orders
// diverge most, which is where a future change to the ordering would show up first.
//
// The values are asserted alongside the counts, so this fails loudly in the one direction that would
// matter — an ordering change that also moved an answer — rather than only in the cheap one.
func TestSolveBackwardOrderIsReversed(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		code       string
		wantRounds int
	}{
		{code: `function f(x) { markBefore(); while (x) { markBody(); } }`, wantRounds: 3},
		{
			code:       `function f(x, y) { markA(); while (x) { markB(); while (y) { markC(); } markD(); } markE(); }`,
			wantRounds: 4,
		},
	} {
		t.Run(testCase.code, func(t *testing.T) {
			t.Parallel()
			graph, _ := buildPathTestGraph(t, testCase.code)
			solution := Solve[map[string]bool, string](graph, Backward, liveNames{})
			if rounds := solution.Rounds(); rounds != testCase.wantRounds {
				t.Errorf("backward analysis settled in %d rounds, want %d; the iteration order is not reversed reverse postorder",
					rounds, testCase.wantRounds)
			}
			// Every reachable block still has a value, so a wrong count is not being read off a
			// solver that quietly stopped early.
			for _, block := range graph.Blocks {
				if !block.Reachable {
					continue
				}
				if _, ok := solution.In(block); !ok {
					t.Errorf("reachable block %d has no value", block.Index())
				}
			}
		})
	}
}
