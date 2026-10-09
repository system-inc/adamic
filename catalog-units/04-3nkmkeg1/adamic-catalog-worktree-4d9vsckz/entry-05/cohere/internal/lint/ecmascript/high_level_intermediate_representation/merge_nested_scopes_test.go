package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/static_single_assignment"
	"testing"
)

func TestNestedScopeMergeMatchesEnclosingDependencies(t *testing.T) {
	t.Parallel()
	const input static_single_assignment.IdentifierId = 1
	for _, testCase := range []struct {
		name                      string
		outer, inner              []ReactiveScopeDependency
		prunedOuter, prunedInner  bool
		conditional, prunedBridge bool
		removed                   int
	}{
		{name: "empty", removed: 1},
		{name: "same input", outer: []ReactiveScopeDependency{{Identifier: input}}, inner: []ReactiveScopeDependency{{Identifier: input}}, removed: 1},
		{name: "different input", outer: []ReactiveScopeDependency{{Identifier: input}}},
		{name: "pruned outer", prunedOuter: true},
		{name: "pruned inner", prunedInner: true},
		{name: "conditional", conditional: true, removed: 1},
		{name: "pruned bridge", prunedBridge: true, removed: 1},
		{name: "different optionality", outer: []ReactiveScopeDependency{{Identifier: input, Path: []DependencyPathEntry{{Property: "value"}}}}, inner: []ReactiveScopeDependency{{Identifier: input, Path: []DependencyPathEntry{{Property: "value", Optional: true}}}}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function := NewFunction(nil, "helper", FunctionKindOther)
			function.NewIdentifier("condition", nil, 0)
			function.NewIdentifier("input", nil, 1)
			leaf := &ReactiveInstruction{Order: 7, Value: &ReactiveInstructionValue{Value: &Primitive{Value: 1}}}
			inner := &ReactiveScopeBlock{Scope: 2, Pruned: testCase.prunedInner, Instructions: ReactiveBlock{&ReactiveInstructionStatement{Instruction: leaf}}}
			body := ReactiveBlock{inner}
			if testCase.conditional {
				body = ReactiveBlock{&ReactiveTerminalStatement{Terminal: &ReactiveIf{Consequent: body}}}
			}
			if testCase.prunedBridge {
				body = ReactiveBlock{&ReactiveScopeBlock{Scope: 3, Pruned: true, Instructions: body}}
			}
			outer := &ReactiveScopeBlock{Scope: 1, Pruned: testCase.prunedOuter, Instructions: body}
			tree := &ReactiveFunction{Body: ReactiveBlock{outer}}
			dependencies := &ScopeDependencies{dependencies: map[ScopeId][]ReactiveScopeDependency{
				1: testCase.outer, 2: testCase.inner, 3: {{Identifier: input}},
			}}
			result := MergeReactiveScopesThatInvalidateTogether(tree, function, dependencies, nil)
			if result.NestedScopesRemoved != testCase.removed || result.Merges != 0 {
				t.Fatalf("merge result=%+v, want %d nested removals and no adjacent merges", result, testCase.removed)
			}
			instructions, innerScopes := 0, 0
			VisitReactiveFunction(tree, ReactiveVisitor{
				Instruction: func(instruction *ReactiveInstruction, traverse func()) {
					if instruction != leaf {
						t.Error("instruction replaced")
					}
					instructions++
					traverse()
				},
				Scope: func(scope *ReactiveScopeBlock, traverse func()) {
					if scope.Scope == 2 {
						innerScopes++
					}
					if len(scope.Merged) != 0 {
						t.Error("nested removal recorded as an adjacent merge")
					}
					traverse()
				},
			})
			if instructions != 1 || innerScopes != 1-testCase.removed {
				t.Fatalf("instructions=%d inner scopes=%d", instructions, innerScopes)
			}
		})
	}
}

func TestManualMemoizationChecksNestedScopeAfterDependencyPruning(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name, parameters, arguments, dependencies string
		fires                                     bool
	}{
		{"constant inputs", "", "", "enabled", true},
		{"reactive inputs", "props", "props.theme", "enabled", false},
		{"missing dependency", "props", "props.theme", "", true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			source := `import {useCallback} from 'react'; function Component(` + testCase.parameters + `) {
const theme=mergeTheme(` + testCase.arguments + `);
const enabled=theme.enabled===true;
const callback=useCallback(value=>[value,enabled],[` + testCase.dependencies + `]);
mutate(theme.colors);
return <div onClick={callback}/>;
}`
			findings, lowered := findingsForSource(t, source)
			if !lowered || (len(findings) > 0) != testCase.fires {
				t.Fatalf("lowered=%t findings=%v; want fires=%t", lowered, findings, testCase.fires)
			}
		})
	}
}
