package high_level_intermediate_representation

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/static_single_assignment"
)

// TestPrunedScopesAreRecordedRatherThanCompared pins upstream's split into two visitor methods.
//
// `visitScope` runs `validateInferredDep` over the scope's dependencies; `visitPrunedScope` does not,
// it only adds the id to `prunedScopes` (`ValidatePreservedManualMemoization.ts:413` and `:441`).
//
// Running one arm for both is not a shape difference. A pruned scope still carries the dependency
// table entry it was assigned, so comparing it re-reports whatever its surviving twin already agreed
// about. Measured on the fixture below: two scopes each carry `propA.x` and `x`, the second is
// pruned, and upstream emits a single guard naming exactly that pair.
//
// The fixture is a program run through the real pipeline rather than a tree built by hand, because a
// hand-built tree asserts what someone believed the pipeline produces. It needs a type checker:
// without one `DropManualMemoization` never recognises the call, no marker is inserted, and nothing
// is pruned -- a first version of this test asserted against an untyped lowering and reported zero
// pruned scopes, which would have passed whichever arm ran.
//
// # What this does NOT catch, stated because a guard that cannot fail is worse than none
//
// It pins the SHAPE the defect needs, not the behaviour. Every call site passes nil for the
// dependency table, so `compareInferredDependencies` returns immediately and moving it back above
// the `Pruned` check changes no output -- measured, that mutation survives. The behavioural assertion
// becomes possible when the third firing condition is ungated, and belongs in the score test at that
// point.
//
// Kept anyway because the shape is the load-bearing half: a live and a pruned scope carrying the
// same dependencies is what makes the arm split matter, and if a future change stops producing that
// shape the reason this code exists has quietly evaporated.
func TestPrunedScopesAreRecordedRatherThanCompared(t *testing.T) {
	t.Parallel()

	const source = `
		import {useCallback} from 'react';
		function sum(a, b) { return a + b; }
		function Component({propA, propB}) {
			const x = propB.x.y;
			return useCallback(() => { return sum(propA.x, x); }, [propA.x, x]);
		}
	`

	live, pruned := 0, 0
	var sharedDependencies int
	probe := rule.Rule{
		Name:             "pruned-scope-split",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					if ctx.TypeChecker == nil {
						return
					}
					forEachFunctionLike(node, func(functionNode *ast.Node) {
						function := Lower(functionNode, ctx.TypeChecker)
						if function == nil {
							return
						}
						Construct(function)
						live, pruned, sharedDependencies = prunedScopeShape(function, ctx.TypeChecker)
					})
				},
			}
		},
	}
	rule_testing.RunTypedFiles(t, probe, map[string]string{
		"/react.d.ts":  reactiveDeclarations,
		"/fixture.tsx": source,
	}, "/fixture.tsx")

	if pruned == 0 {
		t.Fatal("no scope was pruned on this fixture, so the split this test pins is never " +
			"exercised and it would pass whichever arm ran")
	}
	if live == 0 {
		t.Fatal("every scope was pruned, so there is no surviving twin for a pruned scope to " +
			"duplicate and the finding this pins does not arise")
	}
	if sharedDependencies == 0 {
		t.Error("no dependency appears in both a live and a pruned scope, which is the duplication " +
			"that made comparing pruned scopes report twice; the fixture stopped exercising it")
	}
	t.Logf("scopes: %d live, %d pruned; %d dependencies carried by both",
		live, pruned, sharedDependencies)
}

// prunedScopeShape reports how many scopes survive, how many are pruned, and how many dependencies
// are carried by both a live and a pruned scope.
func prunedScopeShape(function *Function, checker *shimchecker.Checker) (int, int, int) {
	InferReactive(function, checker)
	DropManualMemoization(function)
	ranges := InferMutableRanges(function)
	set := FindDisjointMutableValuesWithRanges(function, ranges)
	scopes := AssignReactiveScopesWithSets(function, ranges, set)
	aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
	identity := MergedScopeIdentity{Aligned: aligned, Merged: merged}
	BuildReactiveScopeTerminals(function, scopes, identity)
	dependencies := CollectScopeDependenciesWithHoistable(function, scopes, identity, ranges)

	tree, _ := BuildReactiveFunction(function)
	if tree == nil {
		return 0, 0, 0
	}
	MergeReactiveScopesThatInvalidateTogether(tree, function, dependencies, checker)
	PruneNonEscapingScopesWithScopes(tree, function, dependencies, scopes, checker)
	PruneUnusedScopes(tree, dependencies)
	PruneAlwaysInvalidatingScopes(tree, function, dependencies)
	PruneNonReactiveDependencies(tree, function, dependencies)

	live, pruned := 0, 0
	liveKeys := map[static_single_assignment.IdentifierId]bool{}
	prunedKeys := map[static_single_assignment.IdentifierId]bool{}
	VisitReactiveFunction(tree, ReactiveVisitor{
		Scope: func(block *ReactiveScopeBlock, traverse func()) {
			target := liveKeys
			if block.Pruned {
				pruned++
				target = prunedKeys
			} else {
				live++
			}
			for _, dependency := range dependencies.DependenciesOf(block.Scope) {
				target[dependency.Identifier] = true
			}
			traverse()
		},
	})
	shared := 0
	for id := range prunedKeys {
		if liveKeys[id] {
			shared++
		}
	}
	return live, pruned, shared
}
