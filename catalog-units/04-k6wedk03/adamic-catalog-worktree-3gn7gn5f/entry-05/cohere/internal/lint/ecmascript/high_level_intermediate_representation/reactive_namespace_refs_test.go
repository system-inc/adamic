package high_level_intermediate_representation

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

func TestRefStabilityRecognizesNamespaceCalls(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name     string
		source   string
		reactive bool
	}{
		{
			name: "named import",
			source: `import {useRef} from 'react';
function useFoo() { const reference = useRef(null); return reference; }`,
		},
		{
			name: "namespace import",
			source: `import * as React from 'react';
function useFoo() { const reference = React.useRef(null); return reference; }`,
		},
		{
			name: "namespace result copied through locals",
			source: `import * as React from 'react';
function useFoo() { const original = React.useRef(null); const copied = original;
const reference = copied; return reference; }`,
		},
		{
			name: "custom hook result remains reactive",
			source: `import * as React from 'react';
function useFoo() { function useCustomRef() { return React.useRef(null); }
const reference = useCustomRef(); return reference; }`,
			reactive: true,
		},
		{
			name: "conditional choice remains reactive",
			source: `import * as React from 'react';
function useFoo(condition: boolean) { const first = React.useRef(null);
const second = React.useRef(null); const reference = condition ? first : second;
return reference; }`,
			reactive: true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			reactivity := reactivityByName(t, testCase.source)
			actual, found := reactivity["reference"]
			if !found {
				t.Fatal("the typed fixture did not produce the reference binding")
			}
			if actual != testCase.reactive {
				t.Fatalf("reference reactive=%t, want %t", actual, testCase.reactive)
			}
		})
	}
}

func TestManualMemoizationPreservesNamespaceRefCallbacks(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
		fires  bool
	}{
		{
			name: "named ref callback",
			source: `import {useRef, useCallback} from 'react';
function useFoo() { const reference = useRef(0);
return useCallback((value: number) => { reference.current = value; }, []); }`,
		},
		{
			name: "namespace ref callback",
			source: `import * as React from 'react';
function useFoo() { const reference = React.useRef(0);
return React.useCallback((value: number) => { reference.current = value; }, []); }`,
		},
		{
			name: "property call dependency mismatch",
			source: `import {useMemo} from 'react';
function Component({propA}) { return useMemo(() => propA.x(), [propA.x]); }`,
			fires: true,
		},
		{
			name: "reactive dependency supplied",
			source: `import * as React from 'react';
function Component(props: { value: number }) {
const result = React.useMemo(() => [props.value], [props.value]); return <div>{result}</div>; }`,
		},
		{
			name: "state setter after reactive initialization",
			source: `import * as React from 'react';
function Component(props: { values: number[] }) {
const [values, setValues] = React.useState(props.values);
const [busy, setBusy] = React.useState(false);
const callback = React.useCallback(() => { setBusy(true); setValues([]); }, []);
return <button onClick={callback}>{values.length}{busy}</button>;
}`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var visited, findings, markers int
			probe := rule.Rule{
				Name:             "namespace-ref-memoization",
				NeedsTypeChecker: true,
				Run: func(ctx rule.Context, options any) rule.Listeners {
					return rule.Listeners{
						ast.KindSourceFile: func(node *ast.Node) {
							if ctx.TypeChecker == nil {
								t.Fatal("the fixture requires a type checker")
							}
							forEachFunctionLike(node, func(functionNode *ast.Node) {
								function := Lower(functionNode, ctx.TypeChecker)
								if function == nil {
									t.Fatal("the fixture did not lower")
								}
								visited++
								Construct(function)
								findings += len(AnalyzePreservedManualMemoization(function, ctx.TypeChecker))
								for _, instruction := range function.Instructions {
									if _, isMarker := instruction.Value.(*StartMemoize); isMarker {
										markers++
									}
								}
							})
						},
					}
				},
			}
			rule_testing.RunTypedFiles(t, probe, map[string]string{
				"/react.d.ts": reactiveDeclarations + `
declare module 'react' {
export function useCallback<T extends (...args: any[]) => any>(callback: T, deps: unknown[]): T;
}`,
				"/fixture.tsx": testCase.source,
			}, "/fixture.tsx")
			if visited == 0 || markers == 0 {
				t.Fatal("the fixture did not exercise manual memoization analysis")
			}
			if (findings != 0) != testCase.fires {
				t.Fatalf("findings=%d, want fires=%t", findings, testCase.fires)
			}
		})
	}
}

func TestStableTupleHookPositionsRecognizeNamespaceCalls(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
	}{
		{
			name: "named useState",
			source: `import {useState} from 'react';
function useFoo() { const [value, update] = useState(0); return [value, update]; }`,
		},
		{
			name: "namespace useState",
			source: `import * as React from 'react';
function useFoo() { const [value, update] = React.useState(0); return [value, update]; }`,
		},
		{
			name: "namespace useReducer",
			source: `import * as React from 'react';
function useFoo() { const [value, update] = React.useReducer((state: number, action: number) => state + action, 0);
return [value, update]; }`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			reactivity := reactivityByName(t, testCase.source)
			for name, expected := range map[string]bool{"value": true, "update": false} {
				actual, found := reactivity[name]
				if !found {
					t.Fatalf("the fixture did not produce %s", name)
				}
				if actual != expected {
					t.Errorf("%s reactive=%t, want %t", name, actual, expected)
				}
			}
		})
	}
}
