package high_level_intermediate_representation

import (
	"testing"

	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/static_single_assignment"
)

// TestPruneAlwaysInvalidatingPrunesSomethingAndKeepsSomething is the baseline.
//
// Both directions before any specific claim, because a pass that prunes everything and one that
// prunes nothing each satisfy an assertion phrased about only one side.
func TestPruneAlwaysInvalidatingPrunesSomethingAndKeepsSomething(t *testing.T) {
	t.Parallel()

	functions, scopes, pruned, kept := 0, 0, 0, 0

	forEachCorpusFunctionWithChecker(t, 200, func(function *Function, checker *shimchecker.Checker) {
		tree, dependencies := prunableTree(t, function, checker)
		if tree == nil {
			return
		}
		functions++
		pruned += PruneAlwaysInvalidatingScopes(tree, function, dependencies)

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
	if kept == 0 {
		t.Error("every scope was pruned; a pass that prunes unconditionally satisfies any " +
			"assertion phrased about the pruned side alone")
	}

	t.Logf("functions=%d scopes=%d pruned=%d kept=%d", functions, scopes, pruned, kept)
}

// TestPruneAlwaysInvalidatingSeedsFromAllocationsOnly covers the edge case upstream calls out.
//
// A function call may return a primitive, so upstream optimistically assumes it does and an
// unmemoized call does not prune downstream memoization. Only guaranteed allocations do. That is a
// deliberate imprecision and the arm most likely to be "fixed" by someone reasoning from first
// principles, so it gets a case rather than only a comment.
func TestPruneAlwaysInvalidatingSeedsFromAllocationsOnly(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		value      InstructionValue
		wantPruned bool
	}{
		{name: "array literal allocates", value: &ArrayExpression{}, wantPruned: true},
		{name: "object literal allocates", value: &ObjectExpression{}, wantPruned: true},
		{name: "jsx allocates", value: &JsxExpression{}, wantPruned: true},
		{name: "jsx fragment allocates", value: &JsxFragment{}, wantPruned: true},
		{name: "new allocates", value: &NewExpression{}, wantPruned: true},
		{name: "a call may return a primitive", value: &CallExpression{}, wantPruned: false},
		{name: "a primitive does not allocate", value: &Primitive{}, wantPruned: false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function := &Function{Identifiers: []*Identifier{{Id: 0}, {Id: 1}}}
			allocation := Place{Identifier: 1}

			// The allocation sits OUTSIDE any scope, which is what makes it unmemoized; a scope
			// then depends on it and must prune.
			scope := &ReactiveScopeBlock{Scope: 1}
			tree := &ReactiveFunction{Body: ReactiveBlock{
				&ReactiveInstructionStatement{Instruction: &ReactiveInstruction{
					LValue: &allocation,
					Value:  &ReactiveInstructionValue{Value: testCase.value},
				}},
				scope,
			}}
			dependencies := &ScopeDependencies{
				dependencies: map[ScopeId][]ReactiveScopeDependency{1: {{Identifier: 1}}},
			}

			pruned := PruneAlwaysInvalidatingScopes(tree, function, dependencies)
			if (pruned == 1) != testCase.wantPruned {
				t.Errorf("pruned=%d, want pruned=%t", pruned, testCase.wantPruned)
			}
			if scope.Pruned != testCase.wantPruned {
				t.Errorf("scope.Pruned=%t, want %t", scope.Pruned, testCase.wantPruned)
			}
		})
	}
}

// TestPruneAlwaysInvalidatingRespectsWithinScope pins the distinction the pass turns on.
//
// An allocation inside a scope is memoized by that scope, so it is always-invalidating in principle
// and not unmemoized in fact. Losing that distinction makes every array literal anywhere prune every
// scope reading it, which is the failure mode with no fixture of its own unless one is written.
func TestPruneAlwaysInvalidatingRespectsWithinScope(t *testing.T) {
	t.Parallel()

	function := &Function{Identifiers: []*Identifier{{Id: 0}, {Id: 1}}}
	allocation := Place{Identifier: 1}

	// The allocation is INSIDE a scope, so it is memoized and must not make the consumer prune.
	producer := &ReactiveScopeBlock{Scope: 1, Instructions: ReactiveBlock{
		&ReactiveInstructionStatement{Instruction: &ReactiveInstruction{
			LValue: &allocation,
			Value:  &ReactiveInstructionValue{Value: &ArrayExpression{}},
		}},
	}}
	consumer := &ReactiveScopeBlock{Scope: 2}
	tree := &ReactiveFunction{Body: ReactiveBlock{producer, consumer}}
	dependencies := &ScopeDependencies{
		dependencies: map[ScopeId][]ReactiveScopeDependency{2: {{Identifier: 1}}},
	}

	if pruned := PruneAlwaysInvalidatingScopes(tree, function, dependencies); pruned != 0 {
		t.Errorf("pruned %d scope(s); an allocation inside a scope is memoized by it and must not "+
			"make a consumer prune", pruned)
	}

	// The control: the same allocation outside a scope must prune the consumer, or the case above
	// passes for the wrong reason.
	outside := Place{Identifier: 1}
	consumerAlone := &ReactiveScopeBlock{Scope: 2}
	controlTree := &ReactiveFunction{Body: ReactiveBlock{
		&ReactiveInstructionStatement{Instruction: &ReactiveInstruction{
			LValue: &outside,
			Value:  &ReactiveInstructionValue{Value: &ArrayExpression{}},
		}},
		consumerAlone,
	}}
	if pruned := PruneAlwaysInvalidatingScopes(controlTree, function, dependencies); pruned != 1 {
		t.Fatalf("the control pruned %d, want 1; the case above is therefore not testing the "+
			"within-scope distinction", pruned)
	}
}

// TestPruneAlwaysInvalidatingPropagatesOutward covers the channel that made the declaration pruning
// owed.
//
// A scope pruned here stops memoizing what it declares, so those values become unmemoized and a
// scope depending on them prunes in turn. That chain is the reason a stale declaration in the table
// prunes a scope upstream keeps -- the deny direction that made `PruneDeclarationsLastUsedBefore`
// stop being optional.
func TestPruneAlwaysInvalidatingPropagatesOutward(t *testing.T) {
	t.Parallel()

	function := &Function{Identifiers: []*Identifier{{Id: 0}, {Id: 1}, {Id: 2}}}
	seed := Place{Identifier: 1}

	first := &ReactiveScopeBlock{Scope: 1}
	second := &ReactiveScopeBlock{Scope: 2}
	tree := &ReactiveFunction{Body: ReactiveBlock{
		&ReactiveInstructionStatement{Instruction: &ReactiveInstruction{
			LValue: &seed,
			Value:  &ReactiveInstructionValue{Value: &ArrayExpression{}},
		}},
		first,
		second,
	}}
	// The first scope depends on the unmemoized allocation and declares value 2, which is itself an
	// allocation. The second depends on value 2 and can only prune if the first propagated.
	dependencies := &ScopeDependencies{
		dependencies: map[ScopeId][]ReactiveScopeDependency{
			1: {{Identifier: 1}},
			2: {{Identifier: 2}},
		},
		declarations: map[ScopeId][]static_single_assignment.IdentifierId{1: {2}},
	}

	// Value 2 must be known always-invalidating for the propagation to have anything to carry.
	allocationTwo := Place{Identifier: 2}
	first.Instructions = ReactiveBlock{
		&ReactiveInstructionStatement{Instruction: &ReactiveInstruction{
			LValue: &allocationTwo,
			Value:  &ReactiveInstructionValue{Value: &ObjectExpression{}},
		}},
	}

	pruned := PruneAlwaysInvalidatingScopes(tree, function, dependencies)
	if !first.Pruned {
		t.Fatal("the first scope did not prune, so the propagation below cannot be tested")
	}
	if !second.Pruned {
		t.Error("the second scope did not prune; the first stopped memoizing what it declares, so " +
			"that value is unmemoized and its consumer must prune in turn")
	}
	if pruned != 2 {
		t.Errorf("pruned %d, want 2", pruned)
	}
}
