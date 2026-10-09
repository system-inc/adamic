package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/static_single_assignment"
	"testing"
)

// TestComputeMemoizedWalksFromEscapingRoots is the core propagation, and the case a boolean model
// gets wrong.
//
// Upstream's own example: `a` is independently memoizable but does not escape, so a naive pass drops
// it -- and dropping it breaks the caching of `b`, which does escape, because `a` is one of its
// inputs. The four-level lattice exists for exactly this.
func TestComputeMemoizedWalksFromEscapingRoots(t *testing.T) {
	t.Parallel()

	graph := NewMemoizationGraph()

	// c = [b] escapes; b = [] feeds it; a = [props.a] feeds nothing that escapes.
	graph.Record(1, MemoizationMemoized, nil)                                         // a
	graph.Record(2, MemoizationMemoized, nil)                                         // b
	graph.Record(3, MemoizationMemoized, []static_single_assignment.DeclarationId{2}) // c = [b]
	graph.MarkEscaping(3)

	memoized := graph.ComputeMemoized()

	if !memoized[3] {
		t.Error("the returned value was not memoized; it is the escaping root")
	}
	if !memoized[2] {
		t.Error("a value the escaping root depends on was not memoized; that is the whole " +
			"propagation this pass performs")
	}
	if memoized[1] {
		t.Error("a value nothing reachable from a root depends on was memoized; the pass exists " +
			"to prune those")
	}
}

// TestComputeMemoizedForcesScopeDependencies covers the second reason a non-escaping value is held.
//
// Values whose mutations interleave share one scope. If any output of that scope is memoized, every
// dependency of the scope must be too -- otherwise the scope invalidates more often than necessary
// and breaks the memoization downstream of it. Upstream's second worked example.
func TestComputeMemoizedForcesScopeDependencies(t *testing.T) {
	t.Parallel()

	graph := NewMemoizationGraph()

	// Scope 1 depends on declaration 10. Declaration 2 belongs to it and escapes.
	graph.Record(2, MemoizationMemoized, nil)
	graph.Record(10, MemoizationConditional, nil)
	graph.AssociateScope(2, 1, []static_single_assignment.DeclarationId{10})
	graph.MarkEscaping(2)

	memoized := graph.ComputeMemoized()

	if !memoized[2] {
		t.Fatal("the escaping value was not memoized, so the forcing below cannot be tested")
	}
	if !memoized[10] {
		t.Error("a scope dependency was not forced; a memoized output means every dependency of " +
			"its scope must be held or the scope invalidates more often than necessary")
	}

	// The control: the same dependency with no scope association must NOT be memoized, or the case
	// above passes because Conditional values are held unconditionally.
	control := NewMemoizationGraph()
	control.Record(2, MemoizationMemoized, nil)
	control.Record(10, MemoizationConditional, nil)
	control.MarkEscaping(2)
	if control.ComputeMemoized()[10] {
		t.Fatal("an unassociated Conditional value was memoized anyway, so the case above is not " +
			"testing scope forcing")
	}
}

// TestComputeMemoizedRespectsTheLevels pins what each level does under the same graph shape.
//
// One escaping root depending on one value, varying only that value's level. A two-state model
// cannot produce these four answers.
func TestComputeMemoizedRespectsTheLevels(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name  string
		level MemoizationLevel
		want  bool
	}{
		{name: "memoized is held whenever reachable", level: MemoizationMemoized, want: true},
		{name: "conditional is held only if a dependency is", level: MemoizationConditional, want: false},
		{name: "unmemoized is held only under forcing", level: MemoizationUnmemoized, want: false},
		{name: "never is not held", level: MemoizationNever, want: false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			graph := NewMemoizationGraph()
			graph.Record(1, MemoizationMemoized, []static_single_assignment.DeclarationId{2})
			graph.Record(2, testCase.level, nil)
			graph.MarkEscaping(1)

			if got := graph.ComputeMemoized()[2]; got != testCase.want {
				t.Errorf("level %s: memoized=%t, want %t", testCase.level, got, testCase.want)
			}
		})
	}

	// Conditional WITH a memoized dependency is the arm the table above cannot show, since it
	// varies the leaf rather than the middle.
	graph := NewMemoizationGraph()
	graph.Record(1, MemoizationConditional, []static_single_assignment.DeclarationId{2})
	graph.Record(2, MemoizationMemoized, nil)
	graph.MarkEscaping(1)
	memoized := graph.ComputeMemoized()
	if !memoized[1] {
		t.Error("a Conditional root with a memoized dependency was not held; that is what " +
			"Conditional means")
	}
	if !memoized[2] {
		t.Error("its memoized dependency was not held")
	}
}

// TestComputeMemoizedTerminatesOnCycles covers the guard that makes the walk finite.
//
// A depth-first walk over a graph with a cycle recurses forever without it. The guard answers false
// for a node already on the stack, which is a temporary lie corrected on the way back up -- so the
// test asserts both that it terminates and that the answer is still right.
func TestComputeMemoizedTerminatesOnCycles(t *testing.T) {
	t.Parallel()

	graph := NewMemoizationGraph()
	// 1 -> 2 -> 3 -> 1, with 1 escaping and 3 memoized.
	graph.Record(1, MemoizationConditional, []static_single_assignment.DeclarationId{2})
	graph.Record(2, MemoizationConditional, []static_single_assignment.DeclarationId{3})
	graph.Record(3, MemoizationMemoized, []static_single_assignment.DeclarationId{1})
	graph.MarkEscaping(1)

	memoized := graph.ComputeMemoized()

	if !memoized[3] {
		t.Error("the Memoized node in the cycle was not held; its level does not depend on the walk")
	}
	if !memoized[2] {
		t.Error("a Conditional node with a memoized dependency in a cycle was not held")
	}
}

// TestComputeMemoizedIsRepeatable asserts the traversal state resets.
//
// `seen` and `memoized` live on the nodes, so a second walk over the same graph would answer from
// the first walk's leftovers if they were not cleared. That would make the pass order-dependent in a
// way no single-run test can see.
func TestComputeMemoizedIsRepeatable(t *testing.T) {
	t.Parallel()

	graph := NewMemoizationGraph()
	graph.Record(1, MemoizationMemoized, []static_single_assignment.DeclarationId{2})
	graph.Record(2, MemoizationConditional, nil)
	graph.MarkEscaping(1)

	first := graph.ComputeMemoized()
	second := graph.ComputeMemoized()

	if len(first) == 0 {
		t.Fatal("the first walk memoized nothing, so comparing the two proves nothing")
	}
	if len(first) != len(second) {
		t.Errorf("the first walk memoized %d and the second %d; traversal state is not being reset",
			len(first), len(second))
	}
	for declaration := range first {
		if !second[declaration] {
			t.Errorf("declaration %d was memoized on the first walk and not the second", declaration)
		}
	}
}

// TestComputeMemoizedHandlesUnknownDeclarations pins the conservative direction.
//
// Upstream raises an invariant for a declaration with no node. A linter cannot, and answering true
// would hold a value nothing recorded anything about. False costs granularity; true would be a
// claim the graph does not support.
func TestComputeMemoizedHandlesUnknownDeclarations(t *testing.T) {
	t.Parallel()

	graph := NewMemoizationGraph()
	graph.Record(1, MemoizationConditional, []static_single_assignment.DeclarationId{99})
	graph.MarkEscaping(1)

	memoized := graph.ComputeMemoized()
	if memoized[99] {
		t.Error("an unrecorded declaration was memoized")
	}
	if memoized[1] {
		t.Error("a Conditional value was held on the strength of an unrecorded dependency")
	}

	var nilGraph *MemoizationGraph
	if len(nilGraph.ComputeMemoized()) != 0 {
		t.Error("a nil graph memoized something")
	}
}

// TestRecordJoinsRatherThanAssigns covers the case none of the fixtures above reaches.
//
// A declaration can be an lvalue more than once -- reassigned, or destructured alongside another
// binding -- and its level is the strongest any of those assignments demanded. `Record` joins for
// that reason.
//
// Every other test in this file records each declaration exactly once, where join and assign are
// the same function. A mutation replacing the join with an assignment survived all of them.
//
// Order matters to the assertion: assigning is only visibly wrong when the WEAKER level is recorded
// second, because assigning a stronger level over a weaker one happens to agree with joining. So
// both orders are checked and the answer must be the same.
func TestRecordJoinsRatherThanAssigns(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		levels []MemoizationLevel
	}{
		{
			name:   "strong then weak, which assignment gets wrong",
			levels: []MemoizationLevel{MemoizationMemoized, MemoizationNever},
		},
		{
			name:   "weak then strong, which assignment happens to get right",
			levels: []MemoizationLevel{MemoizationNever, MemoizationMemoized},
		},
		{
			name:   "conditional then never",
			levels: []MemoizationLevel{MemoizationConditional, MemoizationNever},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			graph := NewMemoizationGraph()
			graph.Record(1, MemoizationConditional, []static_single_assignment.DeclarationId{2})
			for _, level := range testCase.levels {
				graph.Record(2, level, nil)
			}
			graph.MarkEscaping(1)

			// Whatever the order, declaration 2 was Memoized or Conditional at least once, so the
			// joined level holds it whenever it is Memoized. The strongest level in each case above
			// is Memoized except the last, so check against the join directly.
			want := MemoizationNever
			for _, level := range testCase.levels {
				want = JoinMemoizationLevels(want, level)
			}
			memoized := graph.ComputeMemoized()
			if got := memoized[2]; got != (want == MemoizationMemoized) {
				t.Errorf("recording levels %v left memoized=%t; the joined level is %s, so the "+
					"level is being assigned rather than joined", testCase.levels, got, want)
			}
		})
	}
}
