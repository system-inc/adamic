package high_level_intermediate_representation

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/mutation_aliasing"
)

func TestCustomHookEffectsFollowModuleBindings(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name, prefix, callee, parameters string
		frozen                           bool
	}{
		{"named", `import {useData} from './hooks';`, "useData", "", true},
		{"renamed", `import {useData as getData} from './hooks';`, "getData", "", true},
		{"local hook name", `import {getData as useData} from './hooks';`, "useData", "", true},
		{"default", `import useData from './hooks';`, "useData", "", true},
		{"namespace", `import * as Hooks from './hooks';`, "Hooks.useData", "", true},
		{"computed namespace", `import * as Hooks from './hooks';`, `Hooks['useData']`, "", true},
		{"local alias", `import {useData} from './hooks';`, "getData", "", true},
		{"module local", `function useData(value) { return value; }`, "useData", "", true},
		{"ordinary", `import {getData} from './hooks';`, "getData", "", false},
		{"shadow", `import {useData} from './hooks';`, "useData", "useData", false},
		{"namespace shadow", `import * as Hooks from './hooks';`, "Hooks.useData", "Hooks", false},
		{"unresolved", "", "useData", "", false},
		{"disabled", "// @enableAssumeHooksFollowRulesOfReact:false\nimport {useData} from './hooks';", "useData", "", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			alias := ""
			if testCase.name == "local alias" {
				alias = "const getData = useData;"
			}
			source := testCase.prefix + " function Component(" + testCase.parameters + ") {" + alias + "return " + testCase.callee + "({value: 0});}"
			visited := 0
			probe := rule.Rule{Name: "custom-hook-effects", NeedsTypeChecker: true, Run: func(ctx rule.Context, options any) rule.Listeners {
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
							var frozenResult, frozenArgument bool
							for _, effect := range effects.Get(instruction.Id) {
								if effect.Kind == mutation_aliasing.AliasingEffectCreate && effect.Into.Identifier == instruction.LValue.Identifier {
									frozenResult = effect.Value == mutation_aliasing.EffectValueFrozen
								}
								frozenArgument = frozenArgument || effect.Kind == mutation_aliasing.AliasingEffectFreeze
							}
							if frozenResult != testCase.frozen || frozenArgument != testCase.frozen {
								t.Errorf("frozen result=%t argument=%t; want both %t", frozenResult, frozenArgument, testCase.frozen)
							}
						}
					})
				}}
			}}
			rule_testing.RunTypedFiles(t, probe, map[string]string{
				"/hooks.ts":    `export function useData(value) {return value;} export function getData(value) {return value;} export default useData;`,
				"/fixture.tsx": source,
			}, "/fixture.tsx")
			if visited != 1 {
				t.Fatalf("visited %d calls, want one", visited)
			}
		})
	}
}

func TestCustomHookPropertyAliasPreservesManualMemoization(t *testing.T) {
	t.Parallel()
	const source = `import {useCallback} from 'react'; import {useData} from './hooks';
function Component() {
const data = useData();
const value = data.value;
const callback = useCallback(() => data.refresh(), [data]);
mutate(value);
return <div onClick={callback}/>;
}`
	for _, testCase := range []struct {
		name, source string
		fires        bool
	}{
		{"property alias", source, false},
		{"optional alias", strings.ReplaceAll(source, "data.value", "data.result?.items ?? []"), false},
		{"ordinary function", strings.ReplaceAll(source, "useData", "getData"), true},
		{"missing dependency", strings.ReplaceAll(source, "[data]", "[]"), true},
		{"no later mutation", strings.ReplaceAll(source, "mutate(value);", ""), false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			findings, lowered := findingsForSource(t, testCase.source)
			if !lowered || (len(findings) != 0) != testCase.fires {
				t.Fatalf("lowered=%t findings=%v; want fires=%t", lowered, findings, testCase.fires)
			}
		})
	}
}

// TestCalleeProducersAreBuiltOncePerFunction guards the quadratic this table replaced: every call
// in a function asks whether its callee is a module hook, and each ask used to walk every
// instruction to build its own table. Five calls, three of them hooks reached through an import,
// a namespace and a local alias, must share one build.
func TestCalleeProducersAreBuiltOncePerFunction(t *testing.T) {
	t.Parallel()
	calls := 0
	var builds []int
	probe := rule.Rule{Name: "callee-producers", NeedsTypeChecker: true, Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{ast.KindSourceFile: func(node *ast.Node) {
			forEachFunctionLike(node, func(node *ast.Node) {
				if node.Name() == nil || node.Name().Text() != "Component" {
					return
				}
				function := Lower(node, ctx.TypeChecker)
				Construct(function)
				producers := newCalleeProducers(function)
				inferAliasingEffects(function, producers)
				for _, instruction := range function.Instructions {
					switch instruction.Value.(type) {
					case *CallExpression, *MethodCall:
						calls++
					}
				}
				builds = append(builds, producers.builds)
			})
		}}
	}}
	rule_testing.RunTypedFiles(t, probe, map[string]string{
		"/hooks.ts": `export function useData(value) {return value;} export function getData(value) {return value;}`,
		"/fixture.tsx": `import {useData, getData} from './hooks'; import * as Hooks from './hooks';
function Component() {
	const useAlias = useData;
	const a = useData(1);
	const b = Hooks.useData(2);
	const c = useAlias(3);
	const d = getData(a);
	return Hooks.getData([b, c, d]);
}`,
	}, "/fixture.tsx")
	if calls != 5 {
		t.Fatalf("lowered %d calls, want five", calls)
	}
	if len(builds) != 1 || builds[0] != 1 {
		t.Fatalf("built the producer table %v times for one function, want once", builds)
	}
}
