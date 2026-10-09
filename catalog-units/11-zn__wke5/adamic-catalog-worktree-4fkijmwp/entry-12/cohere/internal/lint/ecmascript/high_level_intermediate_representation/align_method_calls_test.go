package high_level_intermediate_representation

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestAlignMethodCallScopesRemovesTheInBetweenState asserts the pass's whole contract.
//
// After it runs, a `MethodCall` has either both sides scoped together or neither side scoped. The
// before assertion keeps that result from passing vacuously, and the call count distinguishes a
// lowering failure from a fixture that simply no longer reaches the in-between state.
func TestAlignMethodCallScopesRemovesTheInBetweenState(t *testing.T) {
	t.Parallel()

	// A known primitive-returning method is the live shape after call-result allocation became
	// faithful. Its result does not allocate, while its PropertyLoad result is scoped, so alignment
	// must remove the property's scope.
	//
	// Unknown and known-mutable calls are no longer suitable controls here: upstream's mayAllocate
	// rule gives their results a scope, and the disjoint pass joins that result with the property
	// before this pass runs. The population measurement below records which other arms remain absent.
	const source = `
		function contains(items: string[], selected: string) {
			return items.includes(selected);
		}
	`
	before, after, calls := methodCallScopeStates(t, source)
	if calls == 0 {
		t.Fatal("no method call was reached, so this test asserts nothing about the pass; the " +
			"fixture stopped exercising it rather than the pass being correct")
	}
	if before == 0 {
		t.Fatal("no method call was in the in-between state before the pass, so a clean result " +
			"after it proves nothing")
	}
	if after != 0 {
		t.Errorf("%d method call(s) still have one side scoped and the other not, down from %d; "+
			"the pass exists to make that count zero", after, before)
	}
	t.Logf("method calls: %d reached, %d in the in-between state before, %d after", calls, before, after)
}

// TestAlignMethodCallScopesReachesTheCorpus is the population check.
//
// A pass that fires on nothing passes every assertion above by never being exercised. This is the
// control for that, and the count is asserted rather than logged because a drop to zero would mean
// the pass stopped being reachable rather than stopped being needed.
func TestAlignMethodCallScopesReachesTheCorpus(t *testing.T) {
	t.Parallel()

	functions, changed := 0, 0
	forEachCorpusFunctionWithChecker(t, 400, func(function *Function, checker *shimchecker.Checker) {
		InferReactive(function, checker)
		DropManualMemoization(function)
		ranges := InferMutableRanges(function)
		disjoint := FindDisjointMutableValuesWithRanges(function, ranges)
		scopes := AssignReactiveScopesWithSets(function, ranges, disjoint)
		functions++
		if AlignMethodCallScopes(function, scopes) != scopes {
			changed++
		}
	})
	if functions < 100 {
		t.Fatalf("only %d functions reached; the corpus walk is not finding real source and the "+
			"count below would be a fact about the harness", functions)
	}
	// Measured. A move here is the corpus changing or the pass changing what it touches; both are
	// worth a look rather than an adjustment.
	//
	// 363 with the two frozen-propagation edges in `effects.go` and `ranges.go`. A frozen value no
	// longer widens its range across a conditional mutation, so more method calls reach this pass
	// with their two sides in different scopes and more get aligned. The pass changing what it
	// touches is the second cause this comment names, and it is this one.
	//
	// 365 with the frozen-capture rule in `ranges.go`, for the same reason as the move to 363: a
	// narrower range leaves more method calls with their two sides in different scopes.
	//
	// 368 with the destination half of the frozen-capture rule. A capture into a frozen or primitive
	// destination stops widening, so more method calls reach this pass with their two sides in
	// different scopes. Every upstream-referenced number is unmoved by that change: the scope
	// oracle's per-fixture dump is byte-identical across all 48 scored fixtures, and the board and
	// the other two oracles do not move either.
	//
	// 62 after call-result allocation began following upstream's non-primitive default. Unknown and
	// known-mutable MethodCalls now allocate a result that the disjoint pass joins with the property,
	// so they reach this pass already aligned. The remaining 177 property-only calls are known
	// primitive-returning methods, spread across these 62 functions.
	//
	// 60 after reverse postorder began matching React's fallthrough-first traversal. Loop bodies now
	// precede their continuations, which settles two call/property pairs before this pass sees them.
	// State freezing exposes Menu's findIndex property/result pair: 61. See STATE_EFFECTS.md.
	// Effect-hook signatures expose 79 useEffect pairs in 48 more functions: EFFECT_HOOK_EFFECTS.md.
	// Immutable alias-edge refinement exposes four more functions, listed in IMMUTABLE_ALIAS_EDGES.md.
	//
	// 109 when the walk moved to the frozen corpus (pinnedCorpusFiles), 2026-10-01. A corpus move, not
	// a pass move: over the live tree the analysis at 04501c3 and at ddc86c0 both gave 109, so the four
	// functions left with edits to the live directory before the input was frozen.
	const knownChanged = 109
	if changed != knownChanged {
		t.Errorf("the pass changed %d of %d functions, want %d", changed, functions, knownChanged)
	}
	t.Logf("the pass rewrote scopes in %d of %d corpus functions", changed, functions)
}

// methodCallScopeStates counts method calls whose two sides disagree about scopes.
func methodCallScopeStates(t *testing.T, source string) (before, after, calls int) {
	t.Helper()
	probe := rule.Rule{
		Name:             "method-call-scope-states",
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
						InferReactive(function, ctx.TypeChecker)
						DropManualMemoization(function)
						ranges := InferMutableRanges(function)
						disjoint := FindDisjointMutableValuesWithRanges(function, ranges)
						scopes := AssignReactiveScopesWithSets(function, ranges, disjoint)
						aligned := AlignMethodCallScopes(function, scopes)

						disagreements := func(table *ReactiveScopes) int {
							count := 0
							for _, instruction := range function.Instructions {
								if instruction == nil {
									continue
								}
								call, isMethodCall := instruction.Value.(*MethodCall)
								if !isMethodCall {
									continue
								}
								lvalue := table.ScopeOf(instruction.LValue.Identifier)
								property := table.ScopeOf(call.Property.Identifier)
								if lvalue != property {
									count++
								}
							}
							return count
						}
						for _, instruction := range function.Instructions {
							if instruction == nil {
								continue
							}
							if _, isMethodCall := instruction.Value.(*MethodCall); isMethodCall {
								calls++
							}
						}
						before += disagreements(scopes)
						after += disagreements(aligned)
					})
				},
			}
		},
	}
	rule_testing.RunTypedFiles(t, probe, map[string]string{
		"/react.d.ts":  reactiveDeclarations,
		"/fixture.tsx": source,
	}, "/fixture.tsx")
	return before, after, calls
}
