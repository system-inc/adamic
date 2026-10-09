package high_level_intermediate_representation

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/static_single_assignment"
)

// TestValidatePreservedManualMemoizationFiresAndStaysSilent is the baseline.
//
// A rule that reports on everything and one that reports on nothing both satisfy an assertion
// phrased about a single side. This one has a further trap: its inputs are markers constructed by a
// pass that produced zero of them until recently, so a silent run is the expected shape of a broken
// setup rather than of a clean program.
func TestValidatePreservedManualMemoizationFiresAndStaysSilent(t *testing.T) {
	t.Parallel()

	scopes := &ReactiveScopes{byIdentifier: map[static_single_assignment.IdentifierId]ScopeId{1: 7}}

	// The value's scope did not survive: it is memoized in source and not in output.
	lost := &ReactiveFunction{Body: ReactiveBlock{
		memoStatement(1, &StartMemoize{ManualMemoId: 1}),
		memoStatement(2, &FinishMemoize{ManualMemoId: 1, Value: Place{Identifier: 1}}),
	}}
	findings := ValidatePreservedManualMemoization(lost, &Function{}, scopes)
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1; a value whose scope did not survive is the headline "+
			"case this rule exists for", len(findings))
	}
	if findings[0].Kind != PreserveManualMemoizationValueUnmemoized {
		t.Errorf("finding kind is %s, want the unmemoized-value condition", findings[0].Kind)
	}

	// The same program with the scope surviving must be silent, or the case above passes on a rule
	// that reports unconditionally.
	preserved := &ReactiveFunction{Body: ReactiveBlock{
		&ReactiveScopeBlock{Scope: 7},
		memoStatement(1, &StartMemoize{ManualMemoId: 1}),
		memoStatement(2, &FinishMemoize{ManualMemoId: 1, Value: Place{Identifier: 1}}),
	}}
	if findings := ValidatePreservedManualMemoization(preserved, &Function{}, scopes); len(findings) != 0 {
		t.Errorf("got %d findings on a program whose scope survived, want 0", len(findings))
	}
}

// TestConditionalOptionalArgumentPreservesManualMemoization is the end-to-end regression for the
// final clean-corpus false positive. The optional join is evaluated before the allocating call's
// singleton scope, so its intermediate `.x` store is not a dependency and the written `.x.y` path
// remains valid.
func TestConditionalOptionalArgumentPreservesManualMemoization(t *testing.T) {
	t.Parallel()

	t.Run("matching dependency", func(t *testing.T) {
		t.Parallel()
		findings, lowered := findingsForSource(t, `
			// @validatePreserveExistingMemoizationGuarantees
			import {useMemo} from 'react';
			import {identity} from 'shared-runtime';
			function Component({propA, propB}) {
				return useMemo(() => ({
					value: identity(propB?.x.y),
					other: propA,
				}), [propA, propB.x.y]);
			}
		`)
		if !lowered {
			t.Fatal("optional allocation fixture did not lower")
		}
		if len(findings) != 0 {
			t.Fatalf("optional allocation produced %d preserved-memoization findings, want 0: %+v",
				len(findings), findings)
		}
	})

	// The negative assertion above would also pass if memo markers disappeared, every reactive scope
	// was pruned, or validation stopped walking scope dependencies. Pair it with an upstream error
	// shape through the same harness and require the scope-exit dependency finding specifically.
	t.Run("property call control", func(t *testing.T) {
		t.Parallel()
		findings, lowered := findingsForSource(t, `
			// @validatePreserveExistingMemoizationGuarantees @validateExhaustiveMemoizationDependencies:false
			import {useMemo} from 'react';
			function Component({propA}) {
				return useMemo(() => propA.x(), [propA.x]);
			}
		`)
		if !lowered {
			t.Fatal("property-call control did not lower")
		}
		if len(findings) != 1 {
			t.Fatalf("property-call control produced %d findings, want 1: %+v", len(findings), findings)
		}
		if findings[0].Kind != PreserveManualMemoizationValueUnmemoized {
			t.Errorf("property-call control finding is %s, want dependency mismatch", findings[0].Kind)
		}
		if findings[0].Order != 0 {
			t.Errorf("property-call control finding has order %d, want 0 from scope-exit dependency comparison",
				findings[0].Order)
		}
	})
}

// TestValidatePreservedManualMemoizationAcceptsAMergedScope is why the merge pass was owed.
//
// A scope absorbed by a merge survived under its survivor's identity. Reading only the survivor's
// own id would report every value that was in an absorbed scope -- a false positive on exactly the
// programs the merge improved.
func TestValidatePreservedManualMemoizationAcceptsAMergedScope(t *testing.T) {
	t.Parallel()

	scopes := &ReactiveScopes{byIdentifier: map[static_single_assignment.IdentifierId]ScopeId{1: 8}}

	// Scope 9 survived and absorbed scope 8, which is where the value lives.
	tree := &ReactiveFunction{Body: ReactiveBlock{
		&ReactiveScopeBlock{Scope: 9, Merged: []ScopeId{8}},
		memoStatement(1, &StartMemoize{ManualMemoId: 1}),
		memoStatement(2, &FinishMemoize{ManualMemoId: 1, Value: Place{Identifier: 1}}),
	}}
	if findings := ValidatePreservedManualMemoization(tree, &Function{}, scopes); len(findings) != 0 {
		t.Errorf("got %d findings; the value's scope was absorbed by a surviving scope, so its "+
			"memoization was preserved under the survivor's identity", len(findings))
	}

	// The control: the same survivor without the merge record must report, or the case above passes
	// because surviving scopes are accepted wholesale.
	unmerged := &ReactiveFunction{Body: ReactiveBlock{
		&ReactiveScopeBlock{Scope: 9},
		memoStatement(1, &StartMemoize{ManualMemoId: 1}),
		memoStatement(2, &FinishMemoize{ManualMemoId: 1, Value: Place{Identifier: 1}}),
	}}
	if findings := ValidatePreservedManualMemoization(unmerged, &Function{}, scopes); len(findings) != 1 {
		t.Fatalf("the control returned %d findings, want 1; the case above is not testing the "+
			"merged-scope acceptance", len(findings))
	}
}

// TestValidatePreservedManualMemoizationPairsMarkersById covers the nesting case.
//
// `ManualMemoId` exists because memo calls nest and, once callbacks are inlined, the markers do not
// form a simple stack in instruction order. Pairing by proximity lets an inner finish close an outer
// block, and the outer block's own finish is then dropped as unopened -- so its value is never
// checked and a real lost memoization goes unreported.
func TestValidatePreservedManualMemoizationPairsMarkersById(t *testing.T) {
	t.Parallel()

	scopes := &ReactiveScopes{byIdentifier: map[static_single_assignment.IdentifierId]ScopeId{
		1: 7, // outer value, scope did not survive
		2: 8, // inner value, scope did not survive
	}}

	// Outer opens, inner opens and closes, outer closes. Both values must be reported.
	tree := &ReactiveFunction{Body: ReactiveBlock{
		memoStatement(1, &StartMemoize{ManualMemoId: 1}),
		memoStatement(2, &StartMemoize{ManualMemoId: 2}),
		memoStatement(3, &FinishMemoize{ManualMemoId: 2, Value: Place{Identifier: 2}}),
		memoStatement(4, &FinishMemoize{ManualMemoId: 1, Value: Place{Identifier: 1}}),
	}}

	findings := ValidatePreservedManualMemoization(tree, &Function{}, scopes)
	if len(findings) != 2 {
		t.Errorf("got %d findings, want 2; a boolean open-block flag lets the inner finish close "+
			"the outer block, so the outer value is never checked", len(findings))
	}
}

// TestValidatePreservedManualMemoizationSkipsPrunedAndUnopened covers the two early returns.
//
// A pruned memo block was deliberately discarded, so there is nothing to preserve. A finish with no
// matching start belongs to a block whose dependencies were invalid, which upstream records no state
// for -- validating it would report against a block that was never opened.
func TestValidatePreservedManualMemoizationSkipsPrunedAndUnopened(t *testing.T) {
	t.Parallel()

	scopes := &ReactiveScopes{byIdentifier: map[static_single_assignment.IdentifierId]ScopeId{1: 7}}

	pruned := &ReactiveFunction{Body: ReactiveBlock{
		memoStatement(1, &StartMemoize{ManualMemoId: 1}),
		memoStatement(2, &FinishMemoize{ManualMemoId: 1, Value: Place{Identifier: 1}, Pruned: true}),
	}}
	if findings := ValidatePreservedManualMemoization(pruned, &Function{}, scopes); len(findings) != 0 {
		t.Errorf("got %d findings on a pruned memo block, want 0", len(findings))
	}

	unopened := &ReactiveFunction{Body: ReactiveBlock{
		memoStatement(1, &FinishMemoize{ManualMemoId: 1, Value: Place{Identifier: 1}}),
	}}
	if findings := ValidatePreservedManualMemoization(unopened, &Function{}, scopes); len(findings) != 0 {
		t.Errorf("got %d findings on a finish with no start, want 0", len(findings))
	}
}

// TestValidatePreservedManualMemoizationIgnoresUnscopedValues pins upstream's proxy.
//
// An identifier with no scope is upstream's proxy for a primitive, a global, or another guaranteed
// non-allocating value. Those need no memoization, so a rule reporting them would fire on every
// `useMemo` returning a number.
func TestValidatePreservedManualMemoizationIgnoresUnscopedValues(t *testing.T) {
	t.Parallel()

	// No entry for identifier 1, so ScopeOf answers zero.
	scopes := &ReactiveScopes{byIdentifier: map[static_single_assignment.IdentifierId]ScopeId{}}

	tree := &ReactiveFunction{Body: ReactiveBlock{
		memoStatement(1, &StartMemoize{ManualMemoId: 1}),
		memoStatement(2, &FinishMemoize{ManualMemoId: 1, Value: Place{Identifier: 1}}),
	}}
	if findings := ValidatePreservedManualMemoization(tree, &Function{}, scopes); len(findings) != 0 {
		t.Errorf("got %d findings for an unscoped value; no scope is the proxy for a "+
			"non-allocating value, which needs no memoization", len(findings))
	}

	if findings := ValidatePreservedManualMemoization(tree, &Function{}, nil); len(findings) != 0 {
		t.Errorf("got %d findings with no scope table at all, want 0", len(findings))
	}
	if findings := ValidatePreservedManualMemoization(nil, &Function{}, scopes); findings != nil {
		t.Error("a nil tree produced findings")
	}
}

// memoStatement wraps a memo marker as a tree statement at the given order.
func memoStatement(order static_single_assignment.EvaluationOrder, value InstructionValue) ReactiveStatement {
	return &ReactiveInstructionStatement{Instruction: &ReactiveInstruction{
		Order: order,
		Value: &ReactiveInstructionValue{Value: value},
	}}
}

// TestInferredDependencyComparisonIsBuiltAndGatedOnTruncation pins the reason the third firing
// condition is wired but switched off.
//
// The comparison and its normalization are exercised here rather than left dark, because a pass with
// no caller and no test is the shape this package keeps finding declared and never constructed. What
// is asserted is the state of the INPUT, since that is what the decision rests on.
func TestInferredDependencyComparisonIsBuiltAndGatedOnTruncation(t *testing.T) {
	t.Parallel()

	const source = `
		import {useMemo} from 'react';
		import {sum} from 'shared-runtime';
		function Component({propA, propB}) {
			const x = propB.x.y;
			return useMemo(() => sum(propA.x, x), [propA.x, x]);
		}
	`

	withPath, withoutPath, comparisons := inferredDependencyShapes(t, source)
	if withPath+withoutPath == 0 {
		t.Fatal("no scope dependency was collected, so this test asserts nothing about path depth; " +
			"the fixture stopped exercising the collector rather than the collector being correct")
	}

	// # What this test held, and what it holds now
	//
	// It was written when the source wrote `propA.x` and `propB.x.y` and the inferred side carried
	// only bare roots, and its own message said to re-measure "if the hoistable analysis landed".
	// It has: `CollectHoistablePropertyLoads` populates the tree, the collector descends into
	// nested functions and into callbacks assumed to be invoked, and the memo marker's own operands
	// are no longer read as dependencies of the enclosing scope.
	//
	// On this fixture that is now 2 with a path and 0 truncated, and `propA.x` is one of the two
	// goldens the marker change recovered -- upstream's compiled output for it is
	// `$[0] !== propA.x || $[1] !== x`, so the deep path is the right answer rather than a
	// regression.
	//
	// The floor is what the assertion became. A drop back to zero would mean truncation returned,
	// which is the state this test was built to make visible.
	if withPath == 0 {
		t.Error("no inferred dependency carries a path on a fixture that writes `propA.x` and " +
			"`propB.x.y`; path depth was recovered here and losing it again is the regression " +
			"this test exists to catch")
	}

	// The comparison runs and disagrees, which is the correct answer for a truncated input and is
	// what makes turning it on a regression rather than a fix.
	if comparisons == 0 {
		t.Error("the comparison was never reached, so its wiring is dead rather than gated")
	}
	t.Logf("inferred dependencies: %d with a path, %d truncated; %d comparisons made",
		withPath, withoutPath, comparisons)
}

// inferredDependencyShapes reports the path depth of every inferred dependency, and how many
// comparisons against a written dependency the validator would make.
func inferredDependencyShapes(t *testing.T, source string) (withPath, withoutPath, comparisons int) {
	t.Helper()
	probe := rule.Rule{
		Name:             "inferred-dependency-shapes",
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
						aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
						identity := MergedScopeIdentity{Aligned: aligned, Merged: merged}
						BuildReactiveScopeTerminals(function, scopes, identity)
						dependencies := CollectScopeDependenciesWithHoistable(
							function, scopes, identity, ranges)

						var written []ManualMemoDependency
						for _, instruction := range function.Instructions {
							if instruction == nil {
								continue
							}
							if marker, isStart := instruction.Value.(*StartMemoize); isStart {
								written = append(written, marker.Deps...)
							}
						}

						for _, scope := range scopes.Ids() {
							for _, inferred := range dependencies.DependenciesOf(scope) {
								normalized, ok := dependencies.NormalizeInferredDependency(
									function, inferred)
								if !ok {
									continue
								}
								if len(normalized.Path) > 0 {
									withPath++
								} else {
									withoutPath++
								}
								for _, source := range written {
									CompareManualMemoDependencies(normalized, source)
									comparisons++
								}
							}
						}
					})
				},
			}
		},
	}
	rule_testing.RunTypedFiles(t, probe, map[string]string{
		"/react.d.ts":  reactiveDeclarations,
		"/fixture.tsx": source,
	}, "/fixture.tsx")
	return withPath, withoutPath, comparisons
}
