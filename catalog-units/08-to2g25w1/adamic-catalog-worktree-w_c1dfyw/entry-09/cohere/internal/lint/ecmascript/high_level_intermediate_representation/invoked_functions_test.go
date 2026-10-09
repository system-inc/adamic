package high_level_intermediate_representation

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// invokedFor lowers a source and returns which nested functions are assumed invoked.
func invokedFor(t *testing.T, source string) (*Function, AssumedInvokedFunctions) {
	t.Helper()
	var outer *Function
	var invoked AssumedInvokedFunctions
	probe := rule.Rule{
		Name:             "assumed-invoked",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					if ctx.TypeChecker == nil {
						return
					}
					forEachFunctionLike(node, func(functionNode *ast.Node) {
						if outer != nil {
							return
						}
						lowered := Lower(functionNode, ctx.TypeChecker)
						if lowered == nil {
							return
						}
						Construct(lowered)
						outer = lowered
						invoked = CollectAssumedInvokedFunctions(lowered)
					})
				},
			}
		},
	}
	rule_testing.RunTypedFiles(t, probe, map[string]string{
		"/react.d.ts":  reactiveDeclarations,
		"/fixture.tsx": source,
	}, "/fixture.tsx")
	if outer == nil {
		t.Fatal("nothing lowered, so this test measured nothing")
	}
	return outer, invoked
}

// TestAssumedInvokedFunctionsRecognisesEachCallShape covers every arm that admits a function.
//
// # The arms overlap in this lowering, measured rather than assumed
//
// A mutation sweep removing any SINGLE arm leaves every case below passing, which reads as three
// dead arms. It is not: removing the whole of `collectInvocations` fails all four positive cases, so
// each is genuinely detected. What the sweep actually shows is that this lowering routes one source
// shape through several arms -- `callback()` reaches the hook-argument arm as well as the direct-call
// arm, and a returned callback reaches the return arm as well.
//
// The arms are kept separate anyway, because they are upstream's and because the overlap is a
// property of how this tree lowers calls rather than a property of the analysis. A source shape that
// lowers differently -- and `getAssumedInvokedFunctions` exists to handle several -- would lose
// coverage silently if the arms were collapsed to whichever one happens to fire today.
//
// The negative case is the one that matters. A false positive here is the dangerous direction: it
// lets the hoistable analysis descend into a callback that may never run, and a read hoisted out of
// an uncalled function moves a load to somewhere it can throw. So a fixture that declares a callback
// and never calls it must report nothing, and it is asserted rather than assumed.
func TestAssumedInvokedFunctionsRecognisesEachCallShape(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
		want   int
	}{
		{
			name: "a direct call",
			source: `
				function useHook(x) {
					const callback = () => [x.y.z];
					return callback();
				}
			`,
			want: 1,
		},
		{
			name: "an argument to a hook",
			source: `
				import {useMemo} from 'react';
				function useHook(x) { return useMemo(() => [x.y.z], [x]); }
			`,
			want: 1,
		},
		{
			// The callback reaches the hook as an argument and is NOT returned, so the hook arm is
			// the only one that can admit it. Measured: a mutation removing that arm survives the
			// `useMemo` case above, because there the callback is also the returned value.
			name: "an argument to a hook, where nothing else could admit it",
			source: `
				import {useEffect} from 'react';
				function useHook(x) {
					useEffect(() => { console.log(x.y.z); }, [x]);
					return 1;
				}
			`,
			want: 1,
		},
		{
			name: "a directly returned function",
			source: `
				function useHook(x) { return () => [x.y.z]; }
			`,
			want: 1,
		},
		{
			name: "declared and never called",
			source: `
				function useHook(x) {
					const callback = () => [x.y.z];
					return 1;
				}
			`,
			want: 0,
		},
		{
			name: "stored in an object, which is not a call",
			source: `
				function useHook(x) {
					const callback = () => [x.y.z];
					return {callback};
				}
			`,
			want: 0,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function, invoked := invokedFor(t, testCase.source)
			if len(function.Functions) == 0 {
				t.Fatal("the fixture lowered no nested function, so nothing here could be invoked")
			}
			if len(invoked) != testCase.want {
				t.Errorf("assumed %d of %d nested functions invoked, want %d",
					len(invoked), len(function.Functions), testCase.want)
			}
		})
	}
}
