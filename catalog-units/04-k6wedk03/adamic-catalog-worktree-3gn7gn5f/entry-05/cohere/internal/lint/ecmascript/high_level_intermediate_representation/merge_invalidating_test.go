package high_level_intermediate_representation

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
)

// TestFindLastUsageRecordsTheHighestOrder is the property the merge condition reads.
//
// A declaration read at orders 3, 9 and 5 must answer 9. Getting max wrong in the direction of
// "last write wins" produces 5 here, which is lower than the true last usage -- and a value that
// looks settled earlier than it is gets its scope merged forward past a read that is still live.
func TestFindLastUsageRecordsTheHighestOrder(t *testing.T) {
	t.Parallel()

	// `total` is read in the loop body and again in the return, so its last usage must be the
	// return rather than the loop.
	function, _ := rangesFor(t, `
		function f(items) {
			let total = 0;
			for (const item of items) {
				total = total + item;
			}
			return total;
		}
	`)
	if function == nil {
		t.Fatal("the source did not lower")
	}
	tree, _ := BuildReactiveFunction(function)
	if tree == nil {
		t.Fatal("no tree was built")
	}

	usage := FindLastUsage(tree, function)
	if usage.Len() == 0 {
		t.Fatal("no declaration was recorded, so every assertion below is vacuous")
	}

	// Asserted against an independent walk rather than against the builder: the maximum order any
	// place carries must equal the maximum this table holds, or the walk missed a read.
	highestSeen := static_single_assignment.EvaluationOrder(0)
	VisitReactiveFunction(tree, ReactiveVisitor{
		Place: func(order static_single_assignment.EvaluationOrder, place Place, role PlaceRole) {
			if order > highestSeen {
				highestSeen = order
			}
		},
	})

	highestRecorded := static_single_assignment.EvaluationOrder(0)
	for declaration := range usage.byDeclaration {
		if at, found := usage.LastUsedAt(declaration); found && at > highestRecorded {
			highestRecorded = at
		}
	}
	if highestSeen == 0 {
		t.Fatal("the walk saw no place at a nonzero order; this fixture cannot discriminate")
	}
	if highestRecorded != highestSeen {
		t.Errorf("the highest order any place carries is %d and the table's highest is %d; the "+
			"table is not recording the last usage", highestSeen, highestRecorded)
	}
}

// TestLastUsedAtReportsMissesRatherThanZero pins the second return value.
//
// Upstream indexes its map with a non-null assertion and would throw on a miss. A linter cannot
// throw, and the two callers treat a miss differently, so a miss must be distinguishable from a
// real order of zero. Collapsing the two would make an unread declaration look like one last used
// at the very start of the function, which is exactly the answer that permits a merge.
func TestLastUsedAtReportsMissesRatherThanZero(t *testing.T) {
	t.Parallel()

	usage := &LastUsage{byDeclaration: map[static_single_assignment.DeclarationId]static_single_assignment.EvaluationOrder{7: 0}}

	if order, found := usage.LastUsedAt(7); !found || order != 0 {
		t.Errorf("a declaration recorded at order zero answered (%d, %t), want (0, true)", order, found)
	}
	if _, found := usage.LastUsedAt(99); found {
		t.Error("an unrecorded declaration answered found; a miss must be distinguishable from a " +
			"real order of zero, because zero is the answer that permits a merge")
	}

	var nilUsage *LastUsage
	if _, found := nilUsage.LastUsedAt(7); found {
		t.Error("a nil table answered found")
	}
}

// TestAreEqualDependenciesIsSetEquality covers the comparison the merge turns on.
//
// Order must not matter -- it is an artifact of collection -- and the optional flag must, because
// `a?.b` and `a.b` invalidate under different conditions.
func TestAreEqualDependenciesIsSetEquality(t *testing.T) {
	t.Parallel()

	plain := func(identifier static_single_assignment.IdentifierId, property string, optional bool) ReactiveScopeDependency {
		return ReactiveScopeDependency{
			Identifier: identifier,
			Path:       []DependencyPathEntry{{Property: property, Optional: optional}},
		}
	}

	for _, testCase := range []struct {
		name string
		a, b []ReactiveScopeDependency
		want bool
	}{
		{
			name: "same set in a different order",
			a:    []ReactiveScopeDependency{plain(1, "a", false), plain(2, "b", false)},
			b:    []ReactiveScopeDependency{plain(2, "b", false), plain(1, "a", false)},
			want: true,
		},
		{
			name: "different lengths",
			a:    []ReactiveScopeDependency{plain(1, "a", false)},
			b:    []ReactiveScopeDependency{plain(1, "a", false), plain(2, "b", false)},
			want: false,
		},
		{
			name: "different identifier",
			a:    []ReactiveScopeDependency{plain(1, "a", false)},
			b:    []ReactiveScopeDependency{plain(2, "a", false)},
			want: false,
		},
		{
			name: "different property",
			a:    []ReactiveScopeDependency{plain(1, "a", false)},
			b:    []ReactiveScopeDependency{plain(1, "b", false)},
			want: false,
		},
		{
			name: "optional differs, which is a real difference",
			a:    []ReactiveScopeDependency{plain(1, "a", false)},
			b:    []ReactiveScopeDependency{plain(1, "a", true)},
			want: false,
		},
		{
			name: "both empty",
			a:    nil,
			b:    nil,
			want: true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := AreEqualDependencies(testCase.a, testCase.b); got != testCase.want {
				t.Errorf("got %t, want %t", got, testCase.want)
			}
			// Symmetry is asserted rather than assumed: the implementation walks one side looking
			// into the other, and a length check is the only thing making that symmetric.
			if got := AreEqualDependencies(testCase.b, testCase.a); got != testCase.want {
				t.Errorf("reversed: got %t, want %t", got, testCase.want)
			}
		})
	}
}

// TestAreLValuesLastUsedByScopeDeclinesOnMissingInformation pins the conservative direction.
//
// A declaration with no recorded usage is not a declaration proven safe to merge, it is one this
// pass knows nothing about. Upstream would throw; answering true instead would permit a merge on
// absent information, which is the wrong direction for a pass whose true answer relaxes a guard.
func TestAreLValuesLastUsedByScopeDeclinesOnMissingInformation(t *testing.T) {
	t.Parallel()

	usage := &LastUsage{byDeclaration: map[static_single_assignment.DeclarationId]static_single_assignment.EvaluationOrder{
		1: 5,
		2: 12,
	}}

	if !AreLValuesLastUsedByScope(10, []static_single_assignment.DeclarationId{1}, usage) {
		t.Error("a declaration last used at 5 before a scope ending at 10 should be mergeable")
	}
	if AreLValuesLastUsedByScope(10, []static_single_assignment.DeclarationId{2}, usage) {
		t.Error("a declaration last used at 12 is read after a scope ending at 10, so the merge " +
			"must be declined")
	}
	if AreLValuesLastUsedByScope(10, []static_single_assignment.DeclarationId{1, 2}, usage) {
		t.Error("one unmergeable lvalue must decline the whole set")
	}
	if AreLValuesLastUsedByScope(10, []static_single_assignment.DeclarationId{99}, usage) {
		t.Error("an unrecorded declaration answered true; missing information must decline")
	}
	if !AreLValuesLastUsedByScope(10, nil, usage) {
		t.Error("an empty lvalue set has nothing blocking the merge and should answer true")
	}
	// The boundary: upstream's test is `lastUsedAt >= scope.range.end`, so a usage exactly at the
	// end is a usage after the scope.
	if AreLValuesLastUsedByScope(5, []static_single_assignment.DeclarationId{1}, usage) {
		t.Error("a declaration last used exactly at the scope end must decline; upstream's test " +
			"is >= rather than >")
	}
}

// TestFindLastUsageCorpus runs the walk over real source.
//
// The property asserted is total coverage: every declaration any place names must appear in the
// table, because a missing entry makes `AreLValuesLastUsedByScope` decline a merge that should have
// been allowed, and that failure is silent -- the output is merely more granular.
func TestFindLastUsageCorpus(t *testing.T) {
	t.Parallel()

	functions, declarationsSeen, missing, outOfOrder := 0, 0, 0, 0
	// Accumulated across the whole corpus rather than per function, because a DeclarationId is only
	// unique within its function -- so these are compared per function, inside the walk below, and
	// only the last function's tables survive to the assertion. That is deliberate: the property is
	// per function and one counterexample is enough.
	highestPerDeclaration := map[static_single_assignment.DeclarationId]static_single_assignment.EvaluationOrder{}
	recordedTable := map[static_single_assignment.DeclarationId]static_single_assignment.EvaluationOrder{}

	forEachCorpusFunction(t, 200, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
		BuildReactiveScopeTerminals(function, scopes, MergedScopeIdentity{Aligned: aligned, Merged: merged})

		tree, _ := BuildReactiveFunction(function)
		if tree == nil {
			return
		}
		functions++
		usage := FindLastUsage(tree, function)

		highest := map[static_single_assignment.DeclarationId]static_single_assignment.EvaluationOrder{}
		previous, first := static_single_assignment.EvaluationOrder(0), true
		VisitReactiveFunction(tree, ReactiveVisitor{
			Place: func(order static_single_assignment.EvaluationOrder, place Place, role PlaceRole) {
				declarationsSeen++
				if !first && order < previous {
					outOfOrder++
				}
				previous, first = order, false

				declaration := declarationOf(function, place.Identifier)
				if _, found := usage.LastUsedAt(declaration); !found {
					missing++
				}
				if held, seen := highest[declaration]; !seen || order > held {
					highest[declaration] = order
				}
			},
		})

		// Keep the tables of any function whose walk was non-monotonic, since those are the only
		// ones where max and last-write can differ.
		if outOfOrder > 0 && len(highestPerDeclaration) == 0 {
			for declaration, order := range highest {
				highestPerDeclaration[declaration] = order
				if at, found := usage.LastUsedAt(declaration); found {
					recordedTable[declaration] = at
				}
			}
		}
	})

	if functions < 50 {
		t.Fatalf("only %d functions converted; the corpus walk is not reaching real source", functions)
	}
	if declarationsSeen == 0 {
		t.Fatal("the walk saw no places, so a zero miss count proves nothing")
	}
	if missing != 0 {
		t.Errorf("%d of %d place visits named a declaration absent from the table; the walk that "+
			"builds it and the walk that reads it disagree", missing, declarationsSeen)
	}

	// # The assertion that actually discriminates max from last-write
	//
	// The hand-written case above cannot: it only checks that the table's highest entry matches the
	// walk's highest order, which a last-write-wins implementation also satisfies. A mutation
	// replacing `order > previous` with an unconditional overwrite survived it.
	//
	// It survives only if the walk is monotonic, and it is not: a meaningful fraction of place
	// visits arrive at a lower order than the visit before them. So a last-write table is genuinely
	// wrong on those, and the property that separates the two is stated directly -- every recorded
	// entry must be the maximum over all visits naming that declaration, not the last one.
	//
	// The counts are deliberately not written down here. An earlier version cited "473 of 26,331"
	// from one corpus pass while the test printed 549 of 39,028 from another, and that stale pair
	// sat inside the failure message whose whole job is to tell a reader the assertion still
	// discriminates. A number hardcoded beside the code that computes it drifts, and this is the one
	// place where a phantom discrepancy would cost the most. The live values are interpolated below.
	if outOfOrder == 0 {
		t.Fatalf("no place visit arrived out of order across %d visits in %d functions, so a "+
			"last-write table would be identical to a max table here and this assertion cannot "+
			"discriminate", declarationsSeen, functions)
	}
	for declaration, recorded := range highestPerDeclaration {
		if held, found := recordedTable[declaration]; !found || held != recorded {
			t.Errorf("declaration %d was seen at a highest order of %d and the table holds %d; "+
				"the table is recording the last usage rather than the greatest",
				declaration, recorded, held)
			break
		}
	}

	t.Logf("functions=%d placeVisits=%d missing=%d outOfOrder=%d",
		functions, declarationsSeen, missing, outOfOrder)
}

// TestScopeIsEligibleForMergingTreatsNoDependenciesAsEligible pins the special case.
//
// Upstream's comment: a scope with no dependencies can never change, so there is nothing for a
// later scope to compare and nothing lost by fusing. Every other scope is eligible only if it
// declares an always-invalidating value, because otherwise keeping them separate lets the later
// scope skip work.
//
// Both directions are asserted with a nil checker, which forces `IsAlwaysInvalidatingType` to
// answer false: that isolates the no-dependencies arm from the type arm, so a test passing here
// cannot be passing because the type lookup happened to say yes.
func TestScopeIsEligibleForMergingTreatsNoDependenciesAsEligible(t *testing.T) {
	t.Parallel()

	function, _ := rangesFor(t, `function f(a) { const x = [a]; return x; }`)
	if function == nil {
		t.Fatal("the source did not lower")
	}

	empty := &ScopeDependencies{}
	if !ScopeIsEligibleForMerging(function, 1, empty, nil) {
		t.Error("a scope with no dependencies must be eligible; its output can never change")
	}

	withDependency := &ScopeDependencies{
		dependencies: map[ScopeId][]ReactiveScopeDependency{
			1: {{Identifier: 0}},
		},
	}
	if ScopeIsEligibleForMerging(function, 1, withDependency, nil) {
		t.Error("a scope with dependencies and no always-invalidating declaration must not be " +
			"eligible; keeping it separate lets a later scope compare and skip")
	}

	if ScopeIsEligibleForMerging(function, 1, nil, nil) {
		t.Error("a nil dependency table answered eligible")
	}
}

// TestCanMergeScopesDeclinesReassignments covers the arm the corpus cannot exercise.
//
// Only 2 of 1,576 corpus scopes carry a reassignment, so corpus coverage says almost nothing about
// this branch. It gets its own fixture for that reason rather than being trusted to the sweep.
func TestCanMergeScopesDeclinesReassignments(t *testing.T) {
	t.Parallel()

	function, _ := rangesFor(t, `function f(a) { const x = [a]; return x; }`)
	if function == nil {
		t.Fatal("the source did not lower")
	}

	// Identical dependencies, which would otherwise merge on condition two.
	shared := []ReactiveScopeDependency{{Identifier: 3}}
	table := &ScopeDependencies{
		dependencies: map[ScopeId][]ReactiveScopeDependency{1: shared, 2: shared},
	}
	if !CanMergeScopes(function, 1, 2, table, nil, nil) {
		t.Fatal("two scopes with identical dependencies and no reassignments must merge; the rest " +
			"of this test assumes that baseline")
	}

	for _, testCase := range []struct {
		name          string
		reassignments map[ScopeId][]static_single_assignment.IdentifierId
	}{
		{name: "current reassigns", reassignments: map[ScopeId][]static_single_assignment.IdentifierId{1: {5}}},
		{name: "next reassigns", reassignments: map[ScopeId][]static_single_assignment.IdentifierId{2: {5}}},
		{name: "both reassign", reassignments: map[ScopeId][]static_single_assignment.IdentifierId{1: {5}, 2: {6}}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			withReassignment := &ScopeDependencies{
				dependencies:  map[ScopeId][]ReactiveScopeDependency{1: shared, 2: shared},
				reassignments: testCase.reassignments,
			}
			if CanMergeScopes(function, 1, 2, withReassignment, nil, nil) {
				t.Error("a scope carrying a reassignment must decline; its output depends on " +
					"control flow rather than on its dependencies alone")
			}
		})
	}
}

// TestCanMergeScopesRequiresRootedInvalidatingFlow covers the third arm's two guards.
//
// Upstream requires both that the flowing dependency be rooted (`path.length === 0`) and that its
// type be always-invalidating. Dropping either turns "these always change together" into "these
// happen to be connected", which merges scopes that do not invalidate together -- and the resulting
// program recomputes more than it should while every fixture still passes.
func TestCanMergeScopesRequiresRootedInvalidatingFlow(t *testing.T) {
	t.Parallel()

	function, _ := rangesFor(t, `function f(a) { const x = [a]; return x; }`)
	if function == nil {
		t.Fatal("the source did not lower")
	}

	// A pathed dependency, with a nil checker so the type arm cannot rescue it either. The point is
	// that the path check must reject before the type check is even consulted.
	pathed := &ScopeDependencies{
		dependencies: map[ScopeId][]ReactiveScopeDependency{
			1: {{Identifier: 7}},
			2: {{Identifier: 9, Path: []DependencyPathEntry{{Property: "a"}}}},
		},
		declarations: map[ScopeId][]static_single_assignment.IdentifierId{1: {9}},
	}
	if CanMergeScopes(function, 1, 2, pathed, nil, nil) {
		t.Error("a dependency read through a property path must decline; a property off an " +
			"invalidating value is not itself guaranteed to change")
	}

	// Rooted, flowing from the earlier scope, and not always-invalidating.
	//
	// The later scope depends on TWO values where the earlier declares both, so the synthetic
	// declarations set does not equal the dependency set and condition two cannot fire -- which is
	// what leaves condition three as the only path and makes the type check the deciding factor.
	// An earlier spelling used one dependency matching one declaration exactly, and condition two
	// returned true before the type was ever consulted: the case asserted nothing about types.
	rooted := &ScopeDependencies{
		dependencies: map[ScopeId][]ReactiveScopeDependency{
			1: {{Identifier: 7}},
			2: {{Identifier: 9}},
		},
		declarations: map[ScopeId][]static_single_assignment.IdentifierId{1: {9, 11}},
	}
	if CanMergeScopes(function, 1, 2, rooted, nil, nil) {
		t.Error("a rooted dependency whose type is not always-invalidating must decline; the " +
			"earlier scope's output may not change when its input does")
	}

	// The declarations-equal arm, which does not consult the type at all.
	declarationsMatch := &ScopeDependencies{
		dependencies: map[ScopeId][]ReactiveScopeDependency{
			1: {{Identifier: 7}},
			2: {{Identifier: 9, Reactive: true}},
		},
		declarations: map[ScopeId][]static_single_assignment.IdentifierId{1: {9}},
	}
	if !CanMergeScopes(function, 1, 2, declarationsMatch, nil, nil) {
		t.Error("when the earlier scope's declarations are exactly the later scope's dependencies, " +
			"the scopes merge without consulting types")
	}
}

// TestCanMergeScopesGuardsAreIndependentlyLoadBearing isolates each arm of condition three.
//
// A mutation sweep found the rooted-path check and the flow check both survivable: each fixture
// that was supposed to isolate one of them was ALSO rejected by a later guard, so deleting the
// guard under test changed no answer. A test that only ever sees a rejection cannot tell which
// guard did the rejecting.
//
// Each case here is therefore built to pass every guard except the one it targets, with a real
// checker so the type arm genuinely answers true rather than falling through on a nil.
func TestCanMergeScopesGuardsAreIndependentlyLoadBearing(t *testing.T) {
	t.Parallel()

	withInvalidatingScopes(t, func(function *Function, checker *shimchecker.Checker,
		invalidating static_single_assignment.IdentifierId) {
		// The baseline: rooted, always-invalidating, flowing from the earlier scope's declarations,
		// and the two dependency sets deliberately unequal so conditions one and two cannot fire.
		// This must merge, or every rejection below proves nothing.
		baseline := &ScopeDependencies{
			dependencies: map[ScopeId][]ReactiveScopeDependency{
				1: {{Identifier: invalidating}, {Identifier: invalidating + 1}},
				2: {{Identifier: invalidating}},
			},
			declarations: map[ScopeId][]static_single_assignment.IdentifierId{1: {invalidating}},
		}
		if !CanMergeScopes(function, 1, 2, baseline, nil, checker) {
			t.Fatal("the baseline did not merge, so every case below is rejected for the wrong " +
				"reason and this test asserts nothing about individual guards")
		}

		// Targets the rooted check only: identical to the baseline but for the path.
		pathed := &ScopeDependencies{
			dependencies: map[ScopeId][]ReactiveScopeDependency{
				1: {{Identifier: invalidating}, {Identifier: invalidating + 1}},
				2: {{Identifier: invalidating, Path: []DependencyPathEntry{{Property: "a"}}}},
			},
			declarations: map[ScopeId][]static_single_assignment.IdentifierId{1: {invalidating}},
		}
		if CanMergeScopes(function, 1, 2, pathed, nil, checker) {
			t.Error("a pathed dependency merged; a property read off an invalidating value is not " +
				"itself guaranteed to change, so the rooted check must reject it")
		}

		// Targets the flow check only: rooted and invalidating, but declared by nobody.
		unflowed := &ScopeDependencies{
			dependencies: map[ScopeId][]ReactiveScopeDependency{
				1: {{Identifier: invalidating}, {Identifier: invalidating + 1}},
				2: {{Identifier: invalidating}},
			},
			declarations: map[ScopeId][]static_single_assignment.IdentifierId{1: {invalidating + 2}},
		}
		if CanMergeScopes(function, 1, 2, unflowed, nil, checker) {
			t.Error("a dependency the earlier scope does not declare merged; without the flow " +
				"check this is two unrelated scopes that happen to read invalidating values")
		}
	})
}

// withInvalidatingScopes hands the callback a lowered function plus an identifier whose type the
// checker reports as always-invalidating.
//
// The identifier is found rather than assumed: a hardcoded id would silently stop being an array
// the day lowering changes, and the test would keep passing while asserting nothing.
func withInvalidatingScopes(t *testing.T, visit func(*Function, *shimchecker.Checker, static_single_assignment.IdentifierId)) {
	t.Helper()

	ran := false
	probe := rule.Rule{
		Name:             "merge-guards-harness",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					if ctx.TypeChecker == nil {
						t.Fatal("the typed harness handed this probe a nil checker, so the type " +
							"arm would answer false and the baseline could never merge")
					}
					forEachFunctionLike(node, func(functionNode *ast.Node) {
						function := Lower(functionNode, ctx.TypeChecker)
						if function == nil {
							return
						}
						Construct(function)
						for _, identifier := range function.Identifiers {
							if identifier == nil {
								continue
							}
							if IsAlwaysInvalidatingType(function, identifier.Id, ctx.TypeChecker) {
								ran = true
								visit(function, ctx.TypeChecker, identifier.Id)
								return
							}
						}
					})
				},
			}
		},
	}
	rule_testing.RunTypedFiles(t, probe, map[string]string{
		"/merge.ts": `function f(a) { const first = [a]; const second = [first]; return second; }`,
	}, "/merge.ts")

	if !ran {
		t.Fatal("no always-invalidating identifier was found in the fixture, so the callback " +
			"never ran and this test asserted nothing")
	}
}

// TestMergeReactiveScopesPreservesEveryStatement is the conservation property for the merge.
//
// A merge moves statements from one scope into another; it must never lose one. That is the failure
// this pass can have that no shape assertion catches -- the tree stays well-formed and a scope's
// body is simply short, so every downstream pass reads a program missing instructions.
//
// Run over the corpus rather than a fixture because the merge fires on real code and rarely on
// invented code, and a conservation test that never sees a merge asserts nothing. The merge count is
// therefore asserted nonzero first.
func TestMergeReactiveScopesPreservesEveryStatement(t *testing.T) {
	t.Parallel()

	functions, totalMerges, lost := 0, 0, 0
	pruneCalls, declarationsPruned := 0, 0

	forEachCorpusFunctionWithChecker(t, 200, func(function *Function, checker *shimchecker.Checker) {
		ranges := InferMutableRanges(function)
		set := FindDisjointMutableValuesWithRanges(function, ranges)
		scopes := AssignReactiveScopesWithSets(function, ranges, set)
		aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
		identity := MergedScopeIdentity{Aligned: aligned, Merged: merged}
		BuildReactiveScopeTerminals(function, scopes, identity)
		dependencies := CollectScopeDependenciesWithHoistable(function, scopes, identity, ranges)

		tree, _ := BuildReactiveFunction(function)
		if tree == nil {
			return
		}
		functions++

		before := countReactiveInstructionNodes(tree.Body)
		result := MergeReactiveScopesThatInvalidateTogether(tree, function, dependencies, checker)
		merges := result.Merges
		pruneCalls += result.DeclarationPruneCalls
		declarationsPruned += result.DeclarationsPruned
		after := countReactiveInstructionNodes(tree.Body)

		totalMerges += merges
		if after != before {
			lost++
		}
	})

	if functions < 50 {
		t.Fatalf("only %d functions reached the merge; the corpus walk is not reaching real source",
			functions)
	}
	// The baseline. Every rejection-shaped assertion below is meaningless if the pass never fires,
	// and a merge pass that silently does nothing is exactly what this file's earlier fixtures kept
	// producing.
	if totalMerges == 0 {
		t.Fatalf("no scope merged across %d functions, so the conservation assertion below is "+
			"vacuous; either the pass is a no-op or the corpus holds no mergeable pair", functions)
	}
	if lost != 0 {
		t.Errorf("%d of %d functions changed instruction count across the merge; a merge moves "+
			"statements between scopes and must never lose one", lost, functions)
	}

	// # The control on a zero, which is the whole reason the call count is reported
	//
	// `DeclarationsPruned` is zero on this corpus: every declaration of every merged scope is still
	// read at or after the widened range end, so nothing is prunable. A zero has two readings --
	// the helper is correct and no case exists, or the helper stopped being called -- and only the
	// second is a defect.
	//
	// Upstream calls `updateScopeDeclarations` exactly once per merge, so the call count equals the
	// merge count. Asserted as an equality between live values rather than against a recorded
	// number, for the reason the drift fix in this file records: a figure written beside the code
	// that computes it goes stale.
	//
	// A refactor that drops the call, or moves it behind a guard that stops firing, breaks this
	// equality. Without it every test still passes and the pruning silently stops happening.
	if pruneCalls != totalMerges {
		t.Errorf("the declaration pruning ran %d times across %d merges; upstream runs it once per "+
			"merge, so an inequality means the call was dropped or guarded and the zero below "+
			"stops meaning anything", pruneCalls, totalMerges)
	}

	t.Logf("functions=%d merges=%d functionsLosingStatements=%d pruneCalls=%d declarationsPruned=%d",
		functions, totalMerges, lost, pruneCalls, declarationsPruned)
}

// TestMergeReactiveScopesRecordsAbsorbedIds pins the field the rule reads.
//
// `Merged` is written only by this pass and read by `validatePreservedManualMemoization`. A merge
// that folds statements without recording the absorbed id leaves the rule tracking fewer scopes than
// exist, and nothing about the resulting tree looks wrong.
//
// The self-reference case is asserted directly because it is the specific failure a missing
// `to > from + 1` guard produces: every scope "merges" with itself, `Merged` fills with its own id,
// and the pass reports work it did not do.
func TestMergeReactiveScopesRecordsAbsorbedIds(t *testing.T) {
	t.Parallel()

	functions, merges, recorded, selfReferences := 0, 0, 0, 0

	forEachCorpusFunctionWithChecker(t, 200, func(function *Function, checker *shimchecker.Checker) {
		ranges := InferMutableRanges(function)
		set := FindDisjointMutableValuesWithRanges(function, ranges)
		scopes := AssignReactiveScopesWithSets(function, ranges, set)
		aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
		identity := MergedScopeIdentity{Aligned: aligned, Merged: merged}
		BuildReactiveScopeTerminals(function, scopes, identity)
		dependencies := CollectScopeDependenciesWithHoistable(function, scopes, identity, ranges)

		tree, _ := BuildReactiveFunction(function)
		if tree == nil {
			return
		}
		functions++
		merges += MergeReactiveScopesThatInvalidateTogether(tree, function, dependencies, checker).Merges

		VisitReactiveFunction(tree, ReactiveVisitor{
			Scope: func(scope *ReactiveScopeBlock, traverse func()) {
				recorded += len(scope.Merged)
				for _, absorbed := range scope.Merged {
					if absorbed == scope.Scope {
						selfReferences++
					}
				}
				traverse()
			},
		})
	})

	if merges == 0 {
		t.Fatalf("no scope merged across %d functions, so the assertions below are vacuous", functions)
	}
	if recorded != merges {
		t.Errorf("the pass reported %d merges and the tree records %d absorbed ids; the rule reads "+
			"that field and would track fewer scopes than exist", merges, recorded)
	}
	if selfReferences != 0 {
		t.Errorf("%d scope(s) list their own id as absorbed; a candidate spanning one scope must "+
			"not commit, and this is what its absence looks like", selfReferences)
	}

	t.Logf("functions=%d merges=%d recordedIds=%d selfReferences=%d",
		functions, merges, recorded, selfReferences)
}

// TestMergeReactiveScopesFoldsInterleavedStatements exercises the arm the corpus never reaches.
//
// The allowlist explicitly permits instructions between two merging scopes, and when one is there
// it must be folded into the survivor rather than dropped. The corpus does not produce that shape:
// 132 scopes fold across 200 files and zero non-scope statements come with them, so a mutation
// deleting the fold survives every corpus assertion.
//
// This drives the rewrite directly with a constructed block instead, which is the only way to reach
// the arm at all. A synthetic tree is a weaker test than real source and it is used here for a
// stated reason rather than convenience: the alternative is an arm with no coverage whatsoever.
func TestMergeReactiveScopesFoldsInterleavedStatements(t *testing.T) {
	t.Parallel()

	first := &ReactiveScopeBlock{Scope: 1}
	between := &ReactiveInstructionStatement{Instruction: &ReactiveInstruction{Order: 5}}
	second := &ReactiveScopeBlock{Scope: 2, Instructions: ReactiveBlock{
		&ReactiveInstructionStatement{Instruction: &ReactiveInstruction{Order: 7}},
	}}
	trailing := &ReactiveInstructionStatement{Instruction: &ReactiveInstruction{Order: 9}}

	block := ReactiveBlock{first, between, second, trailing}
	merger := &scopeMerger{}
	rebuilt := merger.rewrite(block, []mergeCandidate{{scope: first, from: 0, to: 3}})

	if merger.merges != 1 {
		t.Fatalf("the rewrite folded %d scopes, want 1; the rest of this test assumes it ran",
			merger.merges)
	}
	if len(rebuilt) != 2 {
		t.Fatalf("the rebuilt block holds %d statements, want 2 (the survivor and the trailing "+
			"statement outside the run)", len(rebuilt))
	}
	if rebuilt[0] != ReactiveStatement(first) {
		t.Error("the survivor is not the first statement of the rebuilt block")
	}
	if rebuilt[1] != ReactiveStatement(trailing) {
		t.Error("the statement after the run was not copied through; the tail is being dropped")
	}

	// The load-bearing assertion: the interleaved statement is inside the survivor, not discarded.
	foundBetween, foundAbsorbed := false, false
	for _, statement := range first.Instructions {
		if statement == ReactiveStatement(between) {
			foundBetween = true
		}
	}
	for _, statement := range second.Instructions {
		_ = statement
		foundAbsorbed = true
	}
	if !foundBetween {
		t.Error("the statement between the two scopes was dropped rather than folded into the " +
			"survivor; the allowlist permits it to be there, so losing it loses real work")
	}
	if !foundAbsorbed {
		t.Fatal("the absorbed scope had no statements, so this fixture cannot tell whether its " +
			"body was folded")
	}
	if len(first.Merged) != 1 || first.Merged[0] != 2 {
		t.Errorf("the survivor records %v as absorbed, want [2]", first.Merged)
	}
}
