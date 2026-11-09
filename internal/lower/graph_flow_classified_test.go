package lower

import (
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// graphFlowUnclassified is an expression no rule names, as an IR type added tomorrow would be.
type graphFlowUnclassified struct{}

func (graphFlowUnclassified) Type() ir.Type { return ir.Object }

// TestGraphFlowFollowsEveryExpression holds graphFlowRules, the one table graphFlows' follow decides by
// (graphFlowStepOf), to internal/ir's source: every expression type is in exactly one class with why,
// each class keeps its shape, and a type no rule names stops rather than leaks (#y7wexf5).
func TestGraphFlowFollowsEveryExpression(t *testing.T) {
	t.Parallel()
	expressions := irOperandHolders(t)
	named := map[string]graphFlowClass{}
	counts := map[graphFlowClass]int{}
	for kind, rule := range graphFlowRules {
		named[kind.Name()] = rule.class
		counts[rule.class]++
		if kind.PkgPath() != reflect.TypeFor[ir.Read]().PkgPath() || kind.Kind() != reflect.Struct {
			t.Errorf("%v is classified for graph flow but isn't an ir struct", kind)
			continue
		}
		if _, declared := expressions[kind.Name()]; !declared {
			t.Errorf("ir.%s is classified for graph flow but isn't an expression in internal/ir", kind.Name())
		}
		if strings.TrimSpace(rule.why) == "" {
			t.Errorf("ir.%s is classified for graph flow without saying why", kind.Name())
		}
		field, allocates := kind.FieldByName("GraphTypes")
		if allocates && field.Type != reflect.TypeFor[[]int]() {
			t.Errorf("ir.%s's GraphTypes is %v, which follow can't read a site from", kind.Name(), field.Type)
		}
		if allocates != (rule.class == graphFlowPlanned) {
			t.Errorf("ir.%s: GraphTypes %t, class planned %t: an allocation carrying GraphTypes is planned, and only one is", kind.Name(), allocates, rule.class == graphFlowPlanned)
		}
		switch rule.class {
		case graphFlowPlanned:
		case graphFlowWrapper, graphFlowForwards:
			if rule.sources == nil {
				t.Errorf("ir.%s is a wrapper or forwards and names no sources", kind.Name())
			}
		case graphFlowElsewhere, graphFlowCannotHold, graphFlowStops:
			if rule.sources != nil {
				t.Errorf("ir.%s's class has nothing to enter, and it names sources", kind.Name())
			}
		default:
			t.Errorf("ir.%s has no class graphFlowStepOf knows: %d", kind.Name(), rule.class)
		}
	}
	for _, expression := range slices.Sorted(maps.Keys(expressions)) {
		if _, classified := named[expression]; !classified {
			t.Errorf("ir.%s isn't classified for graph flow: give it a rule in graphFlowRules (graph_flow_classes.go), planned, a wrapper, forwards, elsewhere, can't hold or stops, with why", expression)
		}
	}
	// The default arm: graphFlowStepOf stops at a type no rule names, never returns silently.
	if step := graphFlowStepOf(graphFlowUnclassified{}, &ir.Program{}); !step.stop {
		t.Errorf("an unclassified expression's step is %+v, want a stop", step)
	}
	t.Logf("%d expressions: %d planned, %d wrappers, %d forward, %d planned elsewhere, %d can't hold, %d stop", len(expressions), counts[graphFlowPlanned], counts[graphFlowWrapper], counts[graphFlowForwards], counts[graphFlowElsewhere], counts[graphFlowCannotHold], counts[graphFlowStops])
}

// TestGraphFlowSteps pins what follow does at the forms the four witnesses go through, and at each
// method of the types whose class depends on one.
func TestGraphFlowSteps(t *testing.T) {
	t.Parallel()
	slice := ir.ArraySlice{GraphTypes: []int{17, -1}, Array: ir.Read{Local: 0, Of: ir.Array}}
	bucket := ir.ArrayLiteral{GraphTypes: []int{-2}}
	read := ir.Read{Local: 3, Of: ir.Array}
	value := ir.Read{Local: 4, Of: ir.Object}
	// ClosureTargets reads a closure local's own record, which a plain local's is.
	program := &ir.Program{Locals: make([]ir.Local, 5)}
	for _, each := range []struct {
		name       string
		expression ir.Expression
		want       graphFlowStep
	}{
		{"toSorted's sort enters its copy", ir.ArraySort{Array: slice}, graphFlowStep{sources: graphFlowSources{operands: []ir.Expression{slice}}}},
		{"toReversed's reverse enters its copy", ir.ArrayReverse{Array: slice}, graphFlowStep{sources: graphFlowSources{operands: []ir.Expression{slice}}}},
		{"the copy is planned", slice, graphFlowStep{planned: true, site: -1}},
		{"groupBy's bucket ?? [] enters each arm, never plans the join", ir.Coalesce{Value: ir.MapGet{}, Fallback: bucket, Of: ir.Array}, graphFlowStep{sources: graphFlowSources{operands: []ir.Expression{ir.MapGet{}, bucket}}}},
		{"a conditional enters each arm", ir.Conditional{WhenTrue: bucket, WhenNot: read}, graphFlowStep{sources: graphFlowSources{operands: []ir.Expression{bucket, read}}}},
		{"a logical enters each arm", ir.Logical{Left: read, Right: bucket, Of: ir.Array}, graphFlowStep{sources: graphFlowSources{operands: []ir.Expression{read, bucket}}}},
		{"an existing bucket is planned where it was stored", ir.MapGet{}, graphFlowStep{}},
		{"Object.values is planned", ir.RecordCall{Method: "values", Returns: ir.Array, GraphTypes: []int{-3}}, graphFlowStep{planned: true, site: -3}},
		{"Object.entries is planned", ir.RecordCall{Method: "entries", Returns: ir.Array, GraphTypes: []int{-3}}, graphFlowStep{planned: true, site: -3}},
		{"a record's get is planned where it was stored", ir.RecordCall{Method: "get", Returns: ir.Object, GraphTypes: []int{-3}}, graphFlowStep{planned: true, site: -3}},
		{"a record's set is its value", ir.RecordCall{Method: "set", Arguments: []ir.Expression{read, ir.StringConstant{}, value}, Returns: ir.Object, GraphTypes: []int{-3}}, graphFlowStep{planned: true, site: -3, sources: graphFlowSources{operands: []ir.Expression{value}}}},
		{"a record call of another method stops", ir.RecordCall{Method: "fromEntries", GraphTypes: []int{-3}}, graphFlowStep{stop: true}},
		{"an allocation without a site stops", ir.RecordCall{Method: "values", Returns: ir.Array}, graphFlowStep{stop: true}},
		{"an in-place fill enters its array", ir.ArrayFill{GraphTypes: []int{-4}, Array: read}, graphFlowStep{planned: true, site: -4, sources: graphFlowSources{operands: []ir.Expression{read}}}},
		{"a new fill is planned", ir.ArrayFill{GraphTypes: []int{-4}}, graphFlowStep{planned: true, site: -4}},
		{"filter is planned", ir.ArrayVisit{Method: "filter", GraphTypes: []int{-5}}, graphFlowStep{planned: true, site: -5}},
		{"a visit of another method stops", ir.ArrayVisit{Method: "flatMap", GraphTypes: []int{-5}}, graphFlowStep{stop: true}},
		{"Object.assign is its target", ir.ObjectCall{Method: "assign", Arguments: []ir.Expression{value, read}, Returns: ir.Object}, graphFlowStep{sources: graphFlowSources{operands: []ir.Expression{value}}}},
		{"Object.freeze is its argument", ir.ObjectCall{Method: "freeze", Arguments: []ir.Expression{value}, Returns: ir.Object}, graphFlowStep{sources: graphFlowSources{operands: []ir.Expression{value}}}},
		{"Object.values of a plain object is scalars", ir.ObjectCall{Method: "values", Arguments: []ir.Expression{value}, Returns: ir.Array}, graphFlowStep{}},
		{"an Object call of another method stops", ir.ObjectCall{Method: "groupBy"}, graphFlowStep{stop: true}},
		{"a logical assignment is its read or its write", ir.LogicalAssignment{Read: read, Write: ir.Assign{Local: 3, Value: bucket}, Of: ir.Array}, graphFlowStep{sources: graphFlowSources{operands: []ir.Expression{read, bucket}}}},
		{"a logical assignment of another write stops", ir.LogicalAssignment{Read: read, Write: ir.Evaluate{Value: bucket}, Of: ir.Array}, graphFlowStep{stop: true}},
		{"a read demands its local", read, graphFlowStep{sources: graphFlowSources{locals: []int{3}}}},
		{"a call demands its target's result", ir.Call{Function: 2, Returns: ir.Object}, graphFlowStep{sources: graphFlowSources{results: []int{2}}}},
		{"a direct closure call demands its target's result", ir.CallClosure{Direct: 3, Returns: ir.Object}, graphFlowStep{sources: graphFlowSources{results: []int{2}}}},
		{"an opaque closure call's result is planned at its targets' Returns", ir.CallClosure{Closure: read, Returns: ir.Object}, graphFlowStep{}},
		{"a bound method demands its factory's result", ir.BoundMethod{Factory: 5}, graphFlowStep{sources: graphFlowSources{results: []int{5}}}},
		{"a generator's step demands its yields", ir.GeneratorNext{Function: 6, Of: ir.Object}, graphFlowStep{sources: graphFlowSources{yields: []int{6}}}},
		{"a field read is planned where it was stored", ir.Property{Object: value, Name: "next", Of: ir.Object}, graphFlowStep{}},
		{"a string can't hold a graph value", ir.StringCall{Method: "split"}, graphFlowStep{}},
		{"an iterator's result stops", ir.IteratorNext{}, graphFlowStep{stop: true}},
		{"a collection's iterator is the runtime's to plan", ir.CollectionIterator{}, graphFlowStep{}},
		{"a decoded tree stops", ir.JSONDecode{}, graphFlowStep{stop: true}},
		{"a host object stops", ir.NodeHostCall{Returns: ir.Object}, graphFlowStep{stop: true}},
		{"a host string holds no graph value", ir.NodeHostCall{Returns: ir.String}, graphFlowStep{}},
		{"a number through a wrapper is no graph value", ir.Coalesce{Value: read, Fallback: ir.NumberConstant{}, Of: ir.Number}, graphFlowStep{}},
		{"a find of numbers is no graph value", ir.ArrayVisit{Method: "find", Element: ir.Number, GraphTypes: []int{-5}}, graphFlowStep{}},
	} {
		if got := graphFlowStepOf(each.expression, program); !reflect.DeepEqual(got, each.want) {
			t.Errorf("%s: graphFlowStepOf(%T) = %+v, want %+v", each.name, each.expression, got, each.want)
		}
	}
}
