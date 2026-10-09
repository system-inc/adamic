package high_level_intermediate_representation

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/mutation_aliasing"
)

func TestStateEffectsFollowImportOrigin(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name   string
		prefix string
		callee string
		frozen bool
	}{
		{"named", `import {useState} from 'react';`, "useState", true},
		{"renamed", `import {useState as state} from 'react';`, "state", true},
		{"namespace", `import * as Hooks from 'react';`, "Hooks.useState", true},
		{"default", `import Hooks from 'react';`, "Hooks.useState", true},
		{"computed namespace", `import * as Hooks from 'react';`, `Hooks['useState']`, true},
		{"constant alias", `import {useState} from 'react'; const state = useState;`, "state", true},
		{"constant namespace alias", `import * as React from 'react'; const Hooks = React;`, "Hooks.useState", true},
		{"reexport", `import {state} from './named';`, "state", true},
		{"star reexport", `import {useState} from './star';`, "useState", true},
		{"namespace reexport", `import * as Hooks from './star';`, "Hooks.useState", true},
		{"local shadow", `import {useState} from 'react';`, "useState", false},
		{"namespace shadow", `import * as React from 'react';`, "React.useState", false},
		{"wrong package", `import {useState} from './other';`, "useState", false},
		{"overridden star export", `import {useState} from './overridden';`, "useState", false},
		{"mutable alias", `import {useState} from 'react'; let state = useState;`, "state", false},
		{"typed impostor", `import {useState as real} from 'react'; const useState: typeof real = value => [value, () => {}];`, "useState", false},
		{"unresolved", "", "useState", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			parameters := ""
			if testCase.name == "local shadow" {
				parameters = "useState: (value: unknown) => unknown"
			} else if testCase.name == "namespace shadow" {
				parameters = "React: {useState(value: unknown): unknown}"
			}
			source := testCase.prefix + ` function Component(` + parameters + `) { return ` + testCase.callee + `({value: 0}); }`
			visited := 0
			probe := rule.Rule{Name: "state-effects-origin", NeedsTypeChecker: true, Run: func(ctx rule.Context, options any) rule.Listeners {
				return rule.Listeners{ast.KindSourceFile: func(node *ast.Node) {
					forEachFunctionLike(node, func(node *ast.Node) {
						if node.Name() == nil || node.Name().Text() != "Component" {
							return
						}
						function := Lower(node, ctx.TypeChecker)
						Construct(function)
						effects := InferAliasingEffects(function)
						for _, instruction := range function.Instructions {
							switch instruction.Value.(type) {
							case *CallExpression, *MethodCall:
							default:
								continue
							}
							visited++
							var origin ModuleExportOrigin
							switch call := instruction.Value.(type) {
							case *CallExpression:
								origin = call.CalleeOrigin
							case *MethodCall:
								origin = call.CalleeOrigin
							}
							if (origin.Module == "react" && origin.Export == "useState") != testCase.frozen {
								t.Errorf("origin=%+v; want React state=%t", origin, testCase.frozen)
							}
							wantFrozen := testCase.frozen || testCase.name == "wrong package" ||
								testCase.name == "overridden star export" || testCase.name == "typed impostor"
							var frozenResult, frozenArgument bool
							for _, effect := range effects.Get(instruction.Id) {
								if effect.Kind == mutation_aliasing.AliasingEffectCreate && effect.Into.Identifier == instruction.LValue.Identifier {
									frozenResult = effect.Value == mutation_aliasing.EffectValueFrozen
								}
								frozenArgument = frozenArgument || effect.Kind == mutation_aliasing.AliasingEffectFreeze
							}
							if frozenResult != wantFrozen || frozenArgument != wantFrozen {
								t.Errorf("result frozen=%t, argument frozen=%t; want both %t", frozenResult, frozenArgument, wantFrozen)
							}
						}
					})
				}}
			}}
			rule_testing.RunTypedFiles(t, probe, map[string]string{
				"/react.d.ts":    reactiveDeclarations + `declare module 'react' { const React: {useState: typeof useState}; export default React; }`,
				"/named.ts":      `export {useState as state} from 'react';`,
				"/star.ts":       `export * from 'react';`,
				"/other.ts":      `export function useState(value: unknown) { return [value]; }`,
				"/overridden.ts": `export * from 'react'; export {useState} from './other';`,
				"/fixture.tsx":   source,
			}, "/fixture.tsx")
			if visited != 1 {
				t.Fatalf("visited %d calls, want exactly one", visited)
			}
		})
	}
}

func TestStateMethodCallDoesNotMakeSetterReactive(t *testing.T) {
	t.Parallel()
	const source = `import * as React from 'react';
export function PD(properties: {ids: string[]; suggested?: {personId: string}}) {
const [selectedValue] = React.useState<string>(properties.suggested ? properties.suggested.personId : 'new');
const [newName] = React.useState('');
const [kept, setKept] = React.useState<Set<string>>(function() { return new Set(properties.ids); });
const canGo = kept.size > 0 && selectedValue === 'new' && newName.trim().length > 0;
const toggle = React.useCallback(function toggle(id: string) {
setKept(function(previous) { const next = new Set(previous); next.add(id); return next; });
}, []);
return <div onClick={() => toggle('a')}>{canGo ? 'y' : 'n'}</div>;
}`
	for _, testCase := range []struct {
		name   string
		source string
		fires  bool
	}{
		{"method read", source, false},
		{"property read", strings.ReplaceAll(source, "newName.trim().length", "newName.length"), false},
		{"default import method read", strings.ReplaceAll(source, "import * as React", "import React"), false},
		{"unpreserved property call", `import {useMemo} from 'react'; function Component({propA}) { return useMemo(() => propA.x(), [propA.x]); }`, true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var findings, markers int
			probe := rule.Rule{Name: "state-method-memoization", NeedsTypeChecker: true, Run: func(ctx rule.Context, options any) rule.Listeners {
				return rule.Listeners{ast.KindSourceFile: func(node *ast.Node) {
					forEachFunctionLike(node, func(node *ast.Node) {
						function := Lower(node, ctx.TypeChecker)
						Construct(function)
						findings += len(AnalyzePreservedManualMemoization(function, ctx.TypeChecker))
						for _, instruction := range function.Instructions {
							if _, found := instruction.Value.(*StartMemoize); found {
								markers++
							}
						}
					})
				}}
			}}
			declarations := strings.ReplaceAll(reactiveDeclarations, "initial?: S", "initial?: S | (() => S)") + `declare module 'react' {
export function useCallback<T extends (...args: any[]) => any>(callback: T, deps: unknown[]): T;
const React: {useState: typeof useState; useCallback: typeof useCallback}; export default React;
}`
			rule_testing.RunTypedFiles(t, probe, map[string]string{"/react.d.ts": declarations, "/fixture.tsx": testCase.source}, "/fixture.tsx")
			if markers == 0 {
				t.Fatal("the fixture did not exercise a memo marker")
			}
			if (findings != 0) != testCase.fires {
				t.Fatalf("findings=%d, want fires=%t", findings, testCase.fires)
			}
		})
	}
}

func TestStateSetterDependencyIsPrunedAfterScopeSeparation(t *testing.T) {
	t.Parallel()
	const source = `import {useCallback, useEffect, useState} from 'react';
let someGlobal = {};
function Component() {
const [state, setState] = useState(someGlobal);
const setGlobal = useCallback(() => { someGlobal.value = true; }, []);
useEffect(() => { setGlobal(); }, []);
useEffect(() => { setState(someGlobal.value); }, [someGlobal]);
return <div>{String(state)}</div>;
}`
	visited := 0
	probe := rule.Rule{Name: "state-setter-dependency-pruning", NeedsTypeChecker: true, Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{ast.KindSourceFile: func(node *ast.Node) {
			forEachFunctionLike(node, func(node *ast.Node) {
				function := Lower(node, ctx.TypeChecker)
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
				identity := MergedScopeIdentity{Aligned: aligned, Merged: merged}
				BuildReactiveScopeTerminals(function, scopes, identity)
				flattened := FlattenReactiveLoops(function)
				dependencies := CollectScopeDependenciesWithHoistable(function, scopes, identity, ranges)
				setterDependencies := func() int {
					count := 0
					for _, scope := range dependencies.Ids() {
						for _, dependency := range dependencies.DependenciesOf(scope) {
							if identifierName(function, dependency.Identifier) == "setState" {
								count++
							}
						}
					}
					return count
				}
				if count := setterDependencies(); count != 1 {
					t.Errorf("raw setter dependencies=%d, want one outside the state scope", count)
				}
				tree, _ := BuildReactiveFunctionWithFlattenedScopes(function, flattened)
				MergeReactiveScopesThatInvalidateTogether(tree, function, dependencies, ctx.TypeChecker)
				PruneNonEscapingScopesWithScopes(tree, function, dependencies, scopes, ctx.TypeChecker)
				PruneUnusedScopes(tree, dependencies)
				PruneAlwaysInvalidatingScopes(tree, function, dependencies)
				PruneNonReactiveDependencies(tree, function, dependencies)
				if count := setterDependencies(); count != 0 {
					t.Errorf("final setter dependencies=%d, want zero for a stable setter", count)
				}
				visited++
			})
		}}
	}}
	rule_testing.RunTypedFiles(t, probe, map[string]string{
		"/react.d.ts":  reactiveDeclarations,
		"/fixture.tsx": source,
	}, "/fixture.tsx")
	if visited != 1 {
		t.Fatalf("visited %d functions, want one", visited)
	}
}
