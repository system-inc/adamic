package high_level_intermediate_representation

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// hookFlattening is what FlattenScopesWithHooksOrUse did to one function.
type hookFlattening struct {
	// pruned is how many scopes it returned as pruned.
	pruned int
	// labels is how many scope terminals it rewrote to a label.
	labels int
	// scopesLeft is how many scope terminals are still in the graph afterwards, pruned or not.
	scopesLeft int
}

// flattenHooksIn runs the passes upstream runs before the hook flatten, then the flatten, over the
// one function in source.
func flattenHooksIn(t *testing.T, source string) hookFlattening {
	t.Helper()
	var result hookFlattening
	visited := 0
	probe := rule.Rule{Name: "hook-flatten", NeedsTypeChecker: true, Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{ast.KindSourceFile: func(node *ast.Node) {
			forEachFunctionLike(node, func(node *ast.Node) {
				function := Lower(node, ctx.TypeChecker)
				if function == nil {
					return
				}
				Construct(function)
				OutlineFunctions(function)
				InferReactive(function, ctx.TypeChecker)
				DropManualMemoization(function)
				if InlineImmediatelyInvokedFunctionExpressionsIncludingMemoCallbacks(function) > 0 {
					MergeConsecutiveBlocks(function)
				}
				EliminateDeadCode(function)
				ranges := InferMutableRanges(function)
				disjoint := FindDisjointMutableValuesWithRanges(function, ranges)
				scopes := AssignReactiveScopesWithSets(function, ranges, disjoint)
				aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
				BuildReactiveScopeTerminals(function, scopes, MergedScopeIdentity{Aligned: aligned, Merged: merged})
				FlattenReactiveLoops(function)
				labelsBefore := 0
				for _, block := range function.Blocks {
					if _, isLabel := block.Terminal.(*Label); isLabel {
						labelsBefore++
					}
				}
				result.pruned = len(FlattenScopesWithHooksOrUse(function))
				for _, block := range function.Blocks {
					switch block.Terminal.(type) {
					case *Label:
						result.labels++
					case *Scope:
						result.scopesLeft++
					}
				}
				result.labels -= labelsBefore
				visited++
			})
		}}
	}}
	rule_testing.RunTypedFiles(t, probe, map[string]string{
		"/react.d.ts": reactiveDeclarations + `
declare module 'react' {
export function useCallback<T extends (...args: any[]) => any>(callback: T, deps: unknown[]): T;
export function useSyncExternalStore<T>(subscribe: (onStoreChange: () => void) => () => void, getSnapshot: () => T): T;
}`,
		"/fixture.tsx": source,
	}, "/fixture.tsx")
	if visited != 1 {
		t.Fatalf("visited %d functions, want one", visited)
	}
	return result
}

// TestFlattenScopesWithHooksOrUseLabelsALoneHookCall pins upstream's first branch: a scope whose body
// is the hook call and nothing else becomes a label, not a pruned scope.
//
// The two branches are indistinguishable to preserve-manual-memoization on every fixture measured,
// because neither leaves a live scope behind, so a mutation folding the label branch into the pruned
// one survives the rule's tests. This counts them instead.
func TestFlattenScopesWithHooksOrUseLabelsALoneHookCall(t *testing.T) {
	t.Parallel()
	got := flattenHooksIn(t, `import {useState} from 'react';
function useFoo() { const [value] = useState(() => ({count: 0})); return value; }`)
	if got.labels != 1 || got.pruned != 0 || got.scopesLeft != 0 {
		t.Fatalf("labels=%d pruned=%d scopes left=%d, want the one hook scope rewritten to a label",
			got.labels, got.pruned, got.scopesLeft)
	}
}

// TestFlattenScopesWithHooksOrUsePrunesAScopeHoldingMore pins the second branch and what it keeps: the
// hook call's scope holds a callback's scope, so it is pruned rather than labelled, and the callback's
// scope stays a scope of its own. That surviving inner scope is what preserve-manual-memoization reads.
func TestFlattenScopesWithHooksOrUsePrunesAScopeHoldingMore(t *testing.T) {
	t.Parallel()
	got := flattenHooksIn(t, `import * as React from 'react';
declare function makeObserver(): { subscribe: (listener: () => void) => () => void; get: () => number };
function useFoo() {
  const [observer] = React.useState(() => makeObserver());
  return React.useSyncExternalStore(
    React.useCallback((onStoreChange: () => void) => observer.subscribe(onStoreChange), [observer]),
    () => observer.get(),
  );
}`)
	// Two hook calls, and each one's scope holds its method property as well as the call, so both are
	// pruned rather than labelled.
	if got.pruned != 2 || got.labels != 0 {
		t.Fatalf("pruned=%d labels=%d, want both hook calls' scopes pruned", got.pruned, got.labels)
	}
	if got.scopesLeft <= got.pruned {
		t.Fatalf("scopes left=%d with %d pruned, want the callback's scope to survive inside the hook's",
			got.scopesLeft, got.pruned)
	}
}
