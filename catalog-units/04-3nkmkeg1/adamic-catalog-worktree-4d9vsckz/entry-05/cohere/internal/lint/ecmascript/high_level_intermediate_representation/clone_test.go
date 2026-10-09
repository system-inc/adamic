package high_level_intermediate_representation

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestCloneFunctionSharesNothingAPassWrites runs the whole preserve-manual-memoization pipeline over a
// copy and checks the original against a second, untouched lowering of the same function.
//
// The pipeline is the hardest writer in the package: it marks places reactive, erases memo calls,
// splices inlined bodies in with new blocks and identifiers, removes dead code, adds scope terminals
// and rewrites some of them to labels. Anything the copy still shares with the original shows up as
// a difference between the original and the fresh lowering. The copy's findings are checked against
// the findings of a function lowered for the pipeline directly, so a copy that drops a field fails
// as well as one that shares one.
//
// The source holds what reaches the copy's edge cases: phis from a branch and a loop, a try, nested
// functions with captures, a memo callback that gets inlined, a hook call with a callback nested in
// its scope, and a finding upstream reports.
func TestCloneFunctionSharesNothingAPassWrites(t *testing.T) {
	t.Parallel()
	const source = `import * as React from 'react';
declare function makeArray(): unknown[];
declare const store: { subscribe: (listener: () => void) => () => void; get: () => number };
function useEverything(properties: { items: number[]; flag: boolean }) {
  let x: unknown[] = [];
  x.push(properties);
  let total = 0;
  for (const item of properties.items) {
    if (item > 2) { total += item; } else { total -= 1; }
  }
  try { x.push(total); } catch (error) { x = []; }
  const doubled = React.useMemo(() => properties.items.map((item) => item * 2), [properties.items]);
  const value = React.useSyncExternalStore(
    React.useCallback((onStoreChange: () => void) => { x.push(onStoreChange); return store.subscribe(onStoreChange); }, [x]),
    () => store.get(),
  );
  x = makeArray();
  return [value, x, doubled, properties.flag ? total : 0];
}`
	visited := 0
	probe := rule.Rule{Name: "clone", NeedsTypeChecker: true, Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{ast.KindSourceFile: func(node *ast.Node) {
			forEachFunctionLike(node, func(node *ast.Node) {
				original := Lower(node, ctx.TypeChecker)
				if original == nil {
					return
				}
				Construct(original)
				untouched := Lower(node, ctx.TypeChecker)
				Construct(untouched)
				forgetIdentifierSlabs(original)
				forgetIdentifierSlabs(untouched)
				if !reflect.DeepEqual(original, untouched) {
					t.Fatal("two lowerings of one function differ, so this test cannot tell a shared write from noise")
				}

				clone := CloneFunction(original)
				if !reflect.DeepEqual(clone, original) {
					t.Fatal("the copy differs from the function it copies")
				}
				fromClone := AnalyzePreservedManualMemoization(clone, ctx.TypeChecker)

				if !reflect.DeepEqual(original, untouched) {
					t.Error("the pipeline run over the copy changed the original")
				}
				direct := Lower(node, ctx.TypeChecker)
				Construct(direct)
				fromLowering := AnalyzePreservedManualMemoization(direct, ctx.TypeChecker)
				if len(fromLowering) == 0 {
					t.Error("the fixture produced no finding, so the comparison below proves nothing about the copy")
				}
				if !reflect.DeepEqual(fromClone, fromLowering) {
					t.Errorf("findings from the copy %v differ from a fresh lowering's %v", fromClone, fromLowering)
				}
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
}

// forgetIdentifierSlabs drops the chunk NewIdentifier carves from, in a function and its nested ones.
//
// The chunk is where identifiers are allocated, not part of the graph: a copy starts without one, and
// comparing it would report a difference no pass can see. The identifiers themselves stay where they
// are, so the graph compares exactly as before.
func forgetIdentifierSlabs(function *Function) {
	function.identifierSlab = nil
	for _, nested := range function.Functions {
		forgetIdentifierSlabs(nested)
	}
}
