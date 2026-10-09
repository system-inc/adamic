package high_level_intermediate_representation

import (
	"fmt"
	"github.com/system-inc/cohere/static_single_assignment"
	"testing"
)

func TestImmutableSourcesDoNotWidenThroughMutableFallbacks(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name    string
		binding string
	}{
		{"property", "const items = props.items ?? [];"},
		{"local alias", "const alias = props; const items = alias ?? [];"},
	} {
		for _, functionName := range []string{"Component", "ordinary"} {
			t.Run(testCase.name+"/"+functionName, func(t *testing.T) {
				t.Parallel()
				function, ranges := rangesFor(t, fmt.Sprintf("function %s(props) {%s mutate(items); return props;}", functionName, testCase.binding))
				var mutationOrder static_single_assignment.EvaluationOrder
				calls := 0
				for _, instruction := range function.Instructions {
					if instruction != nil {
						if _, ok := instruction.Value.(*CallExpression); ok {
							mutationOrder = instruction.Order
							calls++
						}
					}
				}
				if calls != 1 {
					t.Fatalf("found %d calls, want one conditional mutation", calls)
				}
				foundSource, foundFallback, sourceWidened, fallbackWidened := false, false, false, false
				for _, identifier := range function.Identifiers {
					if identifier == nil {
						continue
					}
					switch identifier.Name {
					case "props":
						foundSource = true
						sourceWidened = sourceWidened || ranges.Get(identifier.Id).Contains(mutationOrder)
					case "items":
						foundFallback = true
						fallbackWidened = fallbackWidened || ranges.Get(identifier.Id).Contains(mutationOrder)
					}
				}
				if !foundSource || !foundFallback {
					t.Fatal("source or fallback was not exercised")
				}
				if wantWidened := functionName == "ordinary"; fallbackWidened != wantWidened {
					t.Errorf("fallback widened=%t, want %t", fallbackWidened, wantWidened)
				}
				if wantWidened := functionName == "ordinary"; sourceWidened != wantWidened {
					t.Errorf("source widened=%t, want %t", sourceWidened, wantWidened)
				}
			})
		}
	}
}

func TestManualMemoizationPreservesFrozenPropertyFallback(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name   string
		source string
		fires  bool
	}{
		{"fallback", `import {useMemo} from 'react'; import {useProjects} from './hooks';
function Component() {
const request = useProjects();
const result = useMemo(() => {
const projects = request.data ?? [];
const items = [];
projects.forEach(project => items.push(project));
return items;
}, [request.data]);
return <div>{result}</div>;
}`, false},
		{"no fallback", `import {useMemo} from 'react'; import {useProjects} from './hooks';
function Component() {
const request = useProjects();
const result = useMemo(() => {
const projects = request.data;
const items = [];
projects.forEach(project => items.push(project));
return items;
}, [request.data]);
return <div>{result}</div>;
}`, false},
		{"no iteration", `import {useMemo} from 'react'; import {useProjects} from './hooks';
function Component() {const request = useProjects(); const result = useMemo(() => request.data ?? [], [request.data]); return <div>{result}</div>;}`, false},
		{"unpreserved dependency", `import {useMemo} from 'react'; function Component({propA}) {return useMemo(() => propA.x(), [propA.x]);}`, true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			findings, lowered := findingsForSource(t, testCase.source)
			if !lowered {
				t.Fatal("fixture did not lower")
			}
			if (len(findings) != 0) != testCase.fires {
				t.Fatalf("findings=%v, want fires=%t", findings, testCase.fires)
			}
		})
	}
}
