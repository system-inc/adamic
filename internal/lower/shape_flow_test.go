package lower

import (
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestShapeFlowJoinsArgumentsAssignmentsAndReturns(t *testing.T) {
	first := ir.ObjectLiteral{GraphTypes: []int{123, -1}}
	second := ir.ObjectLiteral{GraphTypes: []int{123, -2}}
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.Object}, {Type: ir.Object}},
		Main: []ir.Statement{
			ir.Declare{Local: 0, Value: first}, ir.Assign{Local: 0, Value: second},
			ir.Evaluate{Value: ir.Call{Function: 0, Arguments: []ir.Expression{ir.Read{Local: 0, Of: ir.Object}}, Returns: ir.Object}},
		},
		Functions: []ir.Function{{Parameters: []int{1}, Returns: ir.Object, Body: []ir.Statement{ir.Return{Value: ir.Read{Local: 1, Of: ir.Object}}}}},
	}
	flow := newAllocationFlowGraph(program)
	result := flow.ReachingAllocations(ir.Call{Function: 0, Returns: ir.Object})
	if result.Unknown || !reflect.DeepEqual(result.Sites, []int{-2, -1}) {
		t.Fatalf("join lost a reaching allocation: %+v", result)
	}
	if len(program.GraphTypes) != 0 {
		t.Fatal("query changed graph ownership")
	}
}

func TestShapeFlowUnknownRetainsKnownAllocations(t *testing.T) {
	flow := newAllocationFlowGraph(&ir.Program{})
	cases := []ir.Expression{
		ir.Read{Local: 0, Of: ir.Object},
		ir.ObjectLiteral{},
		ir.Property{Object: ir.ObjectLiteral{GraphTypes: []int{-3}}, Name: "host", Of: ir.Object},
	}
	for _, opaque := range cases {
		result := flow.ReachingAllocations(ir.Conditional{WhenTrue: ir.ObjectLiteral{GraphTypes: []int{-1}}, WhenNot: opaque, Of: ir.Object})
		if !result.Unknown || !reflect.DeepEqual(result.Sites, []int{-1}) || len(result.Reasons) == 0 {
			t.Fatalf("unknown frontier disappeared: %+v", result)
		}
	}
}

func TestShapeFlowEmptyCycleIsUnknown(t *testing.T) {
	program := &ir.Program{Locals: []ir.Local{{Type: ir.Object}}, Main: []ir.Statement{ir.Assign{Local: 0, Value: ir.Read{Local: 0, Of: ir.Object}}}}
	result := newAllocationFlowGraph(program).ReachingAllocations(ir.Read{Local: 0, Of: ir.Object})
	if !result.Unknown || len(result.Sites) != 0 {
		t.Fatalf("empty cycle must not prove conformance: %+v", result)
	}
}

func TestShapeFlowOmittedAndCallbackArgumentsAreUnknown(t *testing.T) {
	for _, callback := range []bool{false, true} {
		program := &ir.Program{
			Locals:    []ir.Local{{Type: ir.Object}},
			Functions: []ir.Function{{Parameters: []int{0}, Returns: ir.Object, Body: []ir.Statement{ir.Return{Value: ir.Read{Local: 0, Of: ir.Object}}}}},
			Main:      []ir.Statement{ir.Evaluate{Value: ir.Call{Function: 0, Arguments: []ir.Expression{ir.ObjectLiteral{GraphTypes: []int{-1}}}, Returns: ir.Object}}},
		}
		if callback {
			program.Main = append(program.Main, ir.Evaluate{Value: ir.MakeClosure{Function: 0}})
		} else {
			program.Main = append(program.Main, ir.Evaluate{Value: ir.Call{Function: 0, Returns: ir.Object}})
		}
		result := newAllocationFlowGraph(program).ReachingAllocations(ir.Read{Local: 0, Of: ir.Object})
		if !result.Unknown || !reflect.DeepEqual(result.Sites, []int{-1}) {
			t.Fatalf("open parameter was treated as closed: %+v", result)
		}
	}
}
