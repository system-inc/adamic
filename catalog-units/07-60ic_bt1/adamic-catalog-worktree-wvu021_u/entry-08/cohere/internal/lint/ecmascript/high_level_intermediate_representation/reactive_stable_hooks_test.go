package high_level_intermediate_representation

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestStableHookPositionsAreExemptedFromReactivity pins the positional route to React's stable set.
//
// A destructured setter cannot be identified from its own type. `[t, start] = useTransition()` binds
// `start` through an `ArrayPattern` element, and asking the checker for that element's type answers
// nothing usable -- measured, `stableTypeName` returns "" for it while returning
// `TransitionStartFunction` for the same binding elsewhere. Upstream does not hit this because it
// assigns its own shape ids over hook signatures: `isStartTransitionType` tests
// `shapeId === 'BuiltInStartTransition'`. This tree asks TypeScript, and TypeScript has no name for
// a tuple element.
//
// Without the exemption the setter inherits reactivity from the hook call it was destructured out
// of, and the rule then reports it as a dependency of a memo the developer wrote as `[]`. Measured
// before this landed: six of the twelve fixtures that ungating adds were exactly this, naming
// `startTransition`, `addOptimistic`, `dispatchAction`, `start`, `start` and `setState`.
func TestStableHookPositionsAreExemptedFromReactivity(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
		stable string
	}{
		{
			name:   "useTransition start",
			source: `function useFoo() { const [t, start] = useTransition(); return [t, start]; }`,
			stable: "start",
		},
		{
			name:   "useState setter",
			source: `function useFoo() { const [s, setS] = useState(); return [s, setS]; }`,
			stable: "setS",
		},
		{
			name:   "useReducer dispatch",
			source: `function useFoo() { const [s, dispatch] = useReducer(() => {}, null); return [s, dispatch]; }`,
			stable: "dispatch",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			reactive := reactivityByName(t, testCase.source)
			if len(reactive) == 0 {
				t.Fatal("no identifier was named in the reactivity map, so this asserts nothing; " +
					"the fixture stopped lowering rather than the exemption being correct")
			}
			isReactive, seen := reactive[testCase.stable]
			if !seen {
				t.Fatalf("%q was never bound, so the fixture does not exercise the exemption",
					testCase.stable)
			}
			if isReactive {
				t.Errorf("%q is reactive; React guarantees this position is identity-stable, and "+
					"marking it makes the rule report it as a dependency of a memo written with []",
					testCase.stable)
			}
		})
	}
}

// reactivityByName reports, per named binding, whether any place naming it is reactive.
func reactivityByName(t *testing.T, source string) map[string]bool {
	t.Helper()
	result := map[string]bool{}
	probe := rule.Rule{
		Name:             "stable-hook-positions",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					if ctx.TypeChecker == nil {
						t.Fatal("the typed harness handed this probe a nil checker, so every " +
							"type lookup below would answer empty and the test would pass vacuously")
					}
					forEachFunctionLike(node, func(functionNode *ast.Node) {
						function := Lower(functionNode, ctx.TypeChecker)
						if function == nil {
							return
						}
						Construct(function)
						InferReactive(function, ctx.TypeChecker)
						for _, block := range function.Blocks {
							if block == nil {
								continue
							}
							for _, instructionId := range block.Instructions {
								instruction := function.Instructions[instructionId]
								if instruction == nil {
									continue
								}
								EachInstructionPlace(instruction,
									func(place Place, role PlaceRole) {
										identifier := function.Identifiers[place.Identifier]
										if identifier == nil || identifier.Name == "" {
											return
										}
										result[identifier.Name] = result[identifier.Name] ||
											place.Reactive
									})
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
	return result
}
