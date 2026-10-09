package high_level_intermediate_representation

import (
	"testing"

	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/static_single_assignment"
)

// TestPruneUnusedScopesPrunesSomethingAndKeepsSomething is the baseline before any specific claim.
//
// A pass that prunes everything and a pass that prunes nothing both satisfy any assertion phrased
// only about the pruned side. Both directions are asserted first, on real source, so the cases below
// are known to be discriminating rather than assumed to be.
func TestPruneUnusedScopesPrunesSomethingAndKeepsSomething(t *testing.T) {
	t.Parallel()

	functions, scopes, pruned, kept := 0, 0, 0, 0

	forEachCorpusFunctionWithChecker(t, 200, func(function *Function, checker *shimchecker.Checker) {
		tree, dependencies := prunableTree(t, function, checker)
		if tree == nil {
			return
		}
		functions++
		pruned += PruneUnusedScopes(tree, dependencies)

		VisitReactiveFunction(tree, ReactiveVisitor{
			Scope: func(scope *ReactiveScopeBlock, traverse func()) {
				scopes++
				if !scope.Pruned {
					kept++
				}
				traverse()
			},
		})
	})

	if functions < 50 {
		t.Fatalf("only %d functions reached the pass; the corpus walk is not reaching real source",
			functions)
	}
	if scopes == 0 {
		t.Fatal("the corpus produced no scopes, so pruning none proves nothing")
	}
	if pruned == 0 {
		t.Error("no scope was pruned across the whole corpus; the rule reads the pruned set and " +
			"would see every scope as live")
	}
	if kept == 0 {
		t.Error("every scope was pruned; a pass that prunes unconditionally passes any assertion " +
			"phrased about the pruned side alone")
	}

	t.Logf("functions=%d scopes=%d pruned=%d kept=%d", functions, scopes, pruned, kept)
}

// TestPruneUnusedScopesKeepsScopesHoldingAReturn pins the condition that is not obvious.
//
// A scope containing a `return` is kept even with no declarations at all, because the value being
// memoized is early-returned from inside it and `propagateEarlyReturns` rewrites that shape later.
// Pruning it here would delete a scope a downstream pass still needs, and nothing about the
// resulting tree would look wrong.
func TestPruneUnusedScopesKeepsScopesHoldingAReturn(t *testing.T) {
	t.Parallel()

	// A scope with no declarations at all, so only the return condition can keep it.
	withReturn := &ReactiveScopeBlock{Scope: 1, Instructions: ReactiveBlock{
		&ReactiveTerminalStatement{Terminal: &ReactiveReturn{}},
	}}
	tree := &ReactiveFunction{Body: ReactiveBlock{withReturn}}
	if PruneUnusedScopes(tree, &ScopeDependencies{}) != 0 {
		t.Error("a scope holding a return was pruned; the memoized value is early-returned from " +
			"inside it and propagateEarlyReturns needs the scope standing")
	}
	if withReturn.Pruned {
		t.Error("the scope was marked pruned")
	}

	// The control: the same scope without the return must prune, or the case above passes for the
	// wrong reason.
	withoutReturn := &ReactiveScopeBlock{Scope: 1}
	control := &ReactiveFunction{Body: ReactiveBlock{withoutReturn}}
	if PruneUnusedScopes(control, &ScopeDependencies{}) != 1 {
		t.Fatal("the control scope, identical but for the return, was not pruned; the case above " +
			"is therefore not testing the return condition")
	}

	// A return nested inside a terminal still counts, which is the arm a shallow walk would miss.
	nested := &ReactiveScopeBlock{Scope: 1, Instructions: ReactiveBlock{
		&ReactiveTerminalStatement{Terminal: &ReactiveIf{
			Consequent: ReactiveBlock{
				&ReactiveTerminalStatement{Terminal: &ReactiveReturn{}},
			},
		}},
	}}
	nestedTree := &ReactiveFunction{Body: ReactiveBlock{nested}}
	if PruneUnusedScopes(nestedTree, &ScopeDependencies{}) != 0 {
		t.Error("a scope whose return sits inside an if was pruned; the walk is not descending " +
			"into terminals")
	}
}

// TestPruneUnusedScopesKeepsScopesWithOwnDeclarations covers `hasOwnDeclaration`.
//
// The distinction that matters is between a declaration a scope produced and one that bubbled up
// from a scope nested inside it. Only the first keeps the scope, and telling them apart is the
// entire reason `ScopeDependencies` records an origin.
func TestPruneUnusedScopesKeepsScopesWithOwnDeclarations(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name        string
		origin      ScopeId
		originKnown bool
		wantPruned  bool
	}{
		{name: "declares its own value", origin: 1, originKnown: true, wantPruned: false},
		{name: "declaration bubbled up from within", origin: 2, originKnown: true, wantPruned: true},
		{name: "origin unknown keeps the scope", originKnown: false, wantPruned: false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			dependencies := &ScopeDependencies{
				declarations: map[ScopeId][]static_single_assignment.IdentifierId{1: {7}},
			}
			if testCase.originKnown {
				dependencies.declarationOrigin = map[static_single_assignment.IdentifierId]ScopeId{7: testCase.origin}
			}

			scope := &ReactiveScopeBlock{Scope: 1}
			tree := &ReactiveFunction{Body: ReactiveBlock{scope}}
			PruneUnusedScopes(tree, dependencies)

			if scope.Pruned != testCase.wantPruned {
				t.Errorf("pruned=%t, want %t", scope.Pruned, testCase.wantPruned)
			}
		})
	}
}

// TestPruneUnusedScopesKeepsScopesWithReassignments covers the arm the corpus barely exercises.
//
// Only 2 of 1,576 corpus scopes carry a reassignment, so corpus coverage says almost nothing here.
func TestPruneUnusedScopesKeepsScopesWithReassignments(t *testing.T) {
	t.Parallel()

	dependencies := &ScopeDependencies{
		reassignments: map[ScopeId][]static_single_assignment.IdentifierId{1: {7}},
	}
	scope := &ReactiveScopeBlock{Scope: 1}
	tree := &ReactiveFunction{Body: ReactiveBlock{scope}}

	if PruneUnusedScopes(tree, dependencies) != 0 {
		t.Error("a scope carrying a reassignment was pruned")
	}

	// The control, so this is known to be the reassignment doing the work.
	bare := &ReactiveScopeBlock{Scope: 1}
	control := &ReactiveFunction{Body: ReactiveBlock{bare}}
	if PruneUnusedScopes(control, &ScopeDependencies{}) != 1 {
		t.Fatal("the control scope was not pruned, so the case above is not testing reassignments")
	}
}

// prunableTree builds the tree this pass runs on, at the pipeline stage that delivers it.
func prunableTree(t *testing.T, function *Function,
	checker *shimchecker.Checker) (*ReactiveFunction, *ScopeDependencies) {
	t.Helper()

	ranges := InferMutableRanges(function)
	set := FindDisjointMutableValuesWithRanges(function, ranges)
	scopes := AssignReactiveScopesWithSets(function, ranges, set)
	aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
	identity := MergedScopeIdentity{Aligned: aligned, Merged: merged}
	BuildReactiveScopeTerminals(function, scopes, identity)
	dependencies := CollectScopeDependenciesWithHoistable(function, scopes, identity, ranges)

	tree, _ := BuildReactiveFunction(function)
	if tree == nil {
		return nil, nil
	}
	MergeReactiveScopesThatInvalidateTogether(tree, function, dependencies, checker)
	return tree, dependencies
}
