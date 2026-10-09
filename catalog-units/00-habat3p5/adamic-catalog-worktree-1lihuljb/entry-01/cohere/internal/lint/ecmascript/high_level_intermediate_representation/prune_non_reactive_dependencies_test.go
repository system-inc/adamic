package high_level_intermediate_representation

import (
	"testing"

	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/static_single_assignment"
)

// TestPruneNonReactiveDependenciesKeepsAndPrunes is the baseline, and it matters more here than
// elsewhere.
//
// This pass reads a seed it does not compute: `Place.Reactive`, written by `InferReactive`. A caller
// that forgets to run that pass gets an empty seed and prunes EVERY dependency, which is a total
// loss of the rule's comparison surface and would still satisfy any assertion phrased about pruning
// alone. So both directions are asserted, on real source, before any specific claim.
func TestPruneNonReactiveDependenciesKeepsAndPrunes(t *testing.T) {
	t.Parallel()

	functions, before, after, removed := 0, 0, 0, 0

	forEachCorpusFunctionWithChecker(t, 200, func(function *Function, checker *shimchecker.Checker) {
		InferReactive(function, checker)
		tree, dependencies := prunableTree(t, function, checker)
		if tree == nil {
			return
		}
		functions++
		for _, scope := range dependencies.Ids() {
			before += len(dependencies.DependenciesOf(scope))
		}
		removed += PruneNonReactiveDependencies(tree, function, dependencies)
		for _, scope := range dependencies.Ids() {
			after += len(dependencies.DependenciesOf(scope))
		}
	})

	if functions < 50 {
		t.Fatalf("only %d functions reached the pass; the corpus walk is not reaching real source",
			functions)
	}
	if before == 0 {
		t.Fatal("no dependencies were collected, so pruning none proves nothing")
	}
	if after == 0 {
		t.Error("every dependency was pruned; that is what an empty reactive seed looks like, and " +
			"it destroys the set the rule compares")
	}
	if removed != before-after {
		t.Errorf("the pass reported %d removals and the table shrank by %d", removed, before-after)
	}

	t.Logf("functions=%d dependenciesBefore=%d after=%d removed=%d", functions, before, after, removed)
}

// TestPruneNonReactiveDependenciesPropagatesToScopeOutputs covers the half `InferReactive` cannot do.
//
// A value can be non-reactive in itself and reactive in practice: if the scope producing it has any
// reactive dependency, the value re-evaluates whenever that dependency changes. That is a fact about
// scopes rather than about values, so the seed cannot carry it and this pass adds it.
//
// Without the propagation an enclosing scope depending on an inner scope's output prunes it as
// non-reactive, and the rule then compares a set missing a dependency that genuinely changes.
func TestPruneNonReactiveDependenciesPropagatesToScopeOutputs(t *testing.T) {
	t.Parallel()

	// Inner scope: depends on reactive value 1, declares value 2.
	// Outer scope: depends on value 2, which is only reactive because the inner scope's is.
	inner := &ReactiveScopeBlock{Scope: 1}
	outer := &ReactiveScopeBlock{Scope: 2}
	tree := &ReactiveFunction{Body: ReactiveBlock{inner, outer}}

	function := &Function{
		Identifiers: []*Identifier{{Id: 0}, {Id: 1}, {Id: 2}},
		Blocks: []*BasicBlock{{
			Id:           1,
			Instructions: []InstructionId{0},
		}},
		Instructions: []*Instruction{{
			LValue: Place{Identifier: 1, Reactive: true},
			Value:  &LoadLocal{Place: Place{Identifier: 1, Reactive: true}},
		}},
	}
	dependencies := &ScopeDependencies{
		dependencies: map[ScopeId][]ReactiveScopeDependency{
			1: {{Identifier: 1}},
			2: {{Identifier: 2}},
		},
		declarations: map[ScopeId][]static_single_assignment.IdentifierId{1: {2}},
	}

	PruneNonReactiveDependencies(tree, function, dependencies)

	if len(dependencies.DependenciesOf(1)) != 1 {
		t.Fatalf("the inner scope's reactive dependency was pruned, leaving %d; the seed is not "+
			"reaching this fixture and the propagation below cannot be tested",
			len(dependencies.DependenciesOf(1)))
	}
	if len(dependencies.DependenciesOf(2)) != 1 {
		t.Error("the outer scope's dependency was pruned; the inner scope has a reactive " +
			"dependency, so its declared output is reactive in practice and must survive")
	}
}

// TestPruneNonReactiveDependenciesDoesNotPropagateFromAnInertScope is the control on the case above.
//
// The propagation is conditional: a scope whose dependencies were ALL pruned has nothing driving its
// outputs, so those outputs stay non-reactive. Dropping that condition makes the propagation
// unconditional, which keeps every dependency and turns the pass into a no-op that still passes a
// test asserting only that something survives.
func TestPruneNonReactiveDependenciesDoesNotPropagateFromAnInertScope(t *testing.T) {
	t.Parallel()

	inner := &ReactiveScopeBlock{Scope: 1}
	outer := &ReactiveScopeBlock{Scope: 2}
	tree := &ReactiveFunction{Body: ReactiveBlock{inner, outer}}

	// No place is reactive, so the inner scope's dependency prunes and it propagates nothing.
	function := &Function{
		Identifiers: []*Identifier{{Id: 0}, {Id: 1}, {Id: 2}},
		Blocks: []*BasicBlock{{
			Id:           1,
			Instructions: []InstructionId{0},
		}},
		Instructions: []*Instruction{{
			LValue: Place{Identifier: 1},
			Value:  &LoadLocal{Place: Place{Identifier: 1}},
		}},
	}
	dependencies := &ScopeDependencies{
		dependencies: map[ScopeId][]ReactiveScopeDependency{
			1: {{Identifier: 1}},
			2: {{Identifier: 2}},
		},
		declarations: map[ScopeId][]static_single_assignment.IdentifierId{1: {2}},
	}

	removed := PruneNonReactiveDependencies(tree, function, dependencies)
	if removed != 2 {
		t.Errorf("removed %d dependencies, want 2; a scope with no reactive dependency propagates "+
			"nothing, so the outer scope's dependency must prune too", removed)
	}
	if len(dependencies.DependenciesOf(1)) != 0 || len(dependencies.DependenciesOf(2)) != 0 {
		t.Errorf("dependencies survived: inner=%d outer=%d",
			len(dependencies.DependenciesOf(1)), len(dependencies.DependenciesOf(2)))
	}
}

// TestPruneNonReactiveDependenciesVisitsInnermostFirst pins the traversal order.
//
// The propagation runs outward: an inner scope's outputs become reactive, and an enclosing scope
// depending on them must see that before deciding its own. Visiting outermost-first asks the
// enclosing scope first, when the inner scope has not yet contributed, so its dependency prunes as
// non-reactive.
//
// The two flat fixtures above cannot catch that -- a mutation reversing the order survived them,
// because siblings in one block are visited left to right either way. Nesting is where the order
// shows, and it is not exotic: 115 of 1,202 corpus scopes sit inside another.
func TestPruneNonReactiveDependenciesVisitsInnermostFirst(t *testing.T) {
	t.Parallel()

	// The inner scope sits INSIDE the outer one. Inner depends on reactive value 1 and declares
	// value 2; outer depends on value 2, which is reactive only by propagation from inner.
	inner := &ReactiveScopeBlock{Scope: 1}
	outer := &ReactiveScopeBlock{Scope: 2, Instructions: ReactiveBlock{inner}}
	tree := &ReactiveFunction{Body: ReactiveBlock{outer}}

	function := &Function{
		Identifiers: []*Identifier{{Id: 0}, {Id: 1}, {Id: 2}},
		Blocks: []*BasicBlock{{
			Id:           1,
			Instructions: []InstructionId{0},
		}},
		Instructions: []*Instruction{{
			LValue: Place{Identifier: 1, Reactive: true},
			Value:  &LoadLocal{Place: Place{Identifier: 1, Reactive: true}},
		}},
	}
	dependencies := &ScopeDependencies{
		dependencies: map[ScopeId][]ReactiveScopeDependency{
			1: {{Identifier: 1}},
			2: {{Identifier: 2}},
		},
		declarations: map[ScopeId][]static_single_assignment.IdentifierId{1: {2}},
	}

	PruneNonReactiveDependencies(tree, function, dependencies)

	if len(dependencies.DependenciesOf(1)) != 1 {
		t.Fatalf("the inner scope's reactive dependency was pruned, leaving %d; the seed is not "+
			"reaching this fixture", len(dependencies.DependenciesOf(1)))
	}
	if len(dependencies.DependenciesOf(2)) != 1 {
		t.Error("the enclosing scope's dependency was pruned; the inner scope propagates its " +
			"output as reactive, so visiting outermost-first asks before that has happened")
	}
}
