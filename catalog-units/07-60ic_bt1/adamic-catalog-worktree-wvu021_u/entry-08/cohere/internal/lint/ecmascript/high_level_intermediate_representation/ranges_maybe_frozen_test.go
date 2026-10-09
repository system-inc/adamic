package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/static_single_assignment"
	"strings"
	"testing"
)

func TestManualMemoizationPreservesMixedValueClosure(t *testing.T) {
	t.Parallel()
	const source = `import React from 'react';
function Component(props) {
 const items = props.items ?? [];
 const [state,setState] = React.useState(0);
 function findItem(id) { return items.find(item => item.id === id); }
 const item = findItem(props.id);
 consume(item);
 const callback = React.useCallback(() => setState(1), []);
 return <div onClick={callback}>{state}</div>;
}`
	// Every verdict below is React Compiler's, from the copy bundled in eslint-plugin-react-hooks 7.1.1
	// run without the lint gate: "mutable", "global" and "primitive mutable" raise "Existing
	// memoization could not be preserved", "missing dependency" raises a memo dependency error, and
	// the other four compile clean.
	//
	// The three marked reactivityGap are findings upstream makes and this rule does not. Upstream's
	// callback scope keeps `setState` as a reactive dependency, because `setState` shares a mutable
	// alias set with `items` and InferReactivePlaces marks a whole set at once. `InferReactive` keys on
	// the identifier instead (`ReactiveGapMutation`, `ReactiveGapAliasing`). These rows used to fire
	// only because the scope holding `setState` stayed live and handed its reactivity to every
	// declaration in it; it is the `useState` call's scope, which upstream's
	// `flattenScopesWithHooksOrUseHIR` prunes, and a pruned scope hands on nothing. They assert the
	// silence so that closing the gap (#pdxkp8z) fails here and flips them to fires.
	for _, testCase := range []struct {
		name, source  string
		fires         bool
		reactivityGap bool
	}{
		{"mixed", source, false, false},
		{"conditional", strings.Replace(source, "props.items ?? []", "props.flag ? props.items : []", 1), false, false},
		{"frozen", strings.Replace(source, "props.items ?? []", "props.items", 1), false, false},
		{"mutable", strings.Replace(source, "props.items ?? []", "getItems() ?? []", 1), true, true},
		{"global", strings.Replace(source, "props.items ?? []", "props.flag ? externalItems : []", 1), true, true},
		{"primitive mutable", strings.Replace(source, "props.items ?? []", "props.flag ? [] : null", 1), true, true},
		{"missing dependency", strings.Replace(source, "() => setState(1)", "() => [props.value, setState(1)]", 1), true, false},
		{"direct method", strings.Replace(source, "function findItem(id) { return items.find(item => item.id === id); }\n const item = findItem(props.id);", "const item = items.find(item => item.id === props.id);", 1), false, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			findings, lowered := findingsForSource(t, testCase.source)
			reported := testCase.fires && !testCase.reactivityGap
			if !lowered || (len(findings) > 0) != reported {
				t.Fatalf("lowered=%t findings=%v; want reported=%t (upstream fires=%t, reactivity gap=%t)",
					lowered, findings, reported, testCase.fires, testCase.reactivityGap)
			}
		})
	}
}

func TestReadOnlyClosureEffectsRejectUnprovenBodies(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name, body string
		readOnly   bool
	}{
		{"property", "return values.name;", true},
		{"callback", "return values.find(value => value.id === id);", true},
		{"unknown call", "return consume(values);", false},
		{"parameter mutation", "values.name = id; return values;", false},
		{"global assignment", "externalValue = values; return values;", false},
		{"impure builtin", "return Math.random();", false},
		{"nested mutation", "return values.find(value => {value.id = id; return true});", false},
		{"object method", "return {run() {consume(values)}};", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function, _ := rangesFor(t, "function helper(values, id) {"+testCase.body+"}")
			if got := hasReadOnlyClosureEffects(function, map[*Function]bool{}); got != testCase.readOnly {
				t.Fatalf("read-only=%t, want %t", got, testCase.readOnly)
			}
		})
	}
}

func TestClosureRefExclusionFollowsDerivedValues(t *testing.T) {
	t.Parallel()
	function, _ := rangesFor(t, `import React from 'react'; function Component(props) {
const reference = React.useRef(null);
const copied = reference;
const current = copied.current;
const nested = current.value;
const selected = props.flag ? nested : null;
const unrelated = props.value;
return {reference, copied, current, nested, selected, unrelated};
}`)
	refs := refDerivedValues(function)
	active := map[static_single_assignment.IdentifierId]bool{}
	for _, instruction := range function.Instructions {
		if instruction != nil {
			EachInstructionPlace(instruction, func(place Place, role PlaceRole) {
				active[place.Identifier] = true
			})
		}
	}
	for name, want := range map[string]bool{
		"reference": true, "copied": true, "current": true,
		"nested": true, "selected": true, "unrelated": false,
	} {
		found := false
		for _, identifier := range function.Identifiers {
			if identifier != nil && active[identifier.Id] && identifier.Name == name {
				found = true
				if refs[identifier.Id] != want {
					t.Errorf("%s ref-derived=%t, want %t", name, refs[identifier.Id], want)
				}
			}
		}
		if !found {
			t.Errorf("missing %s", name)
		}
	}
}
