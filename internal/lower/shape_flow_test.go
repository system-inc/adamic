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

func TestShapeProjectionJoinsInitialAndStoredValues(t *testing.T) {
	first := ir.ObjectLiteral{GraphTypes: []int{-2}}
	second := ir.ObjectLiteral{GraphTypes: []int{-3}}
	receiver := ir.Read{Local: 0, Of: ir.Object}
	program := &ir.Program{Locals: []ir.Local{{Type: ir.Object}}, Main: []ir.Statement{
		ir.Declare{Local: 0, Value: ir.ObjectLiteral{GraphTypes: []int{-1}, Fields: []ir.Field{{Name: "child", Value: first}}}},
		ir.SetProperty{Object: receiver, Name: "child", Value: second},
	}}
	result := newAllocationFlowGraph(program).ReachingAllocations(ir.Property{Object: receiver, Name: "child", Of: ir.Object})
	if result.Unknown || !reflect.DeepEqual(result.Sites, []int{-3, -2}) {
		t.Fatalf("field join lost reaching shapes: %+v", result)
	}
	program.Main = append(program.Main, ir.SetProperty{Object: ir.Read{Local: 99, Of: ir.Object}, Name: "child", Value: second})
	result = newAllocationFlowGraph(program).ReachingAllocations(ir.Property{Object: receiver, Name: "child", Of: ir.Object})
	if !result.Unknown || !reflect.DeepEqual(result.Sites, []int{-3, -2}) {
		t.Fatalf("opaque store must retain known shapes and unknown: %+v", result)
	}
}
func TestShapeProjectionCyclesAndMissingFieldsStayUnknown(t *testing.T) {
	receiver := ir.Read{Local: 0, Of: ir.Object}
	property := ir.Property{Object: receiver, Name: "child", Of: ir.Object}
	program := &ir.Program{Locals: []ir.Local{{Type: ir.Object}}, Main: []ir.Statement{
		ir.Declare{Local: 0, Value: ir.ObjectLiteral{GraphTypes: []int{-1}, Fields: []ir.Field{{Name: "child", Value: ir.ObjectLiteral{GraphTypes: []int{-2}}}}}},
		ir.SetProperty{Object: receiver, Name: "child", Value: property},
	}}
	if result := newAllocationFlowGraph(program).ReachingAllocations(property); !result.Unknown || !reflect.DeepEqual(result.Sites, []int{-2}) {
		t.Fatalf("recursive store should terminate conservatively: %+v", result)
	}
	program.Main = append(program.Main, ir.SetProperty{Object: receiver, Name: "missing", Value: ir.ObjectLiteral{GraphTypes: []int{-3}}})
	if result := newAllocationFlowGraph(program).ReachingAllocations(ir.Property{Object: receiver, Name: "missing", Of: ir.Object}); !result.Unknown {
		t.Fatal("store without path readiness certified an initially missing field")
	}
}
func TestShapeProjectionConstantElements(t *testing.T) {
	receiver := ir.Read{Local: 0, Of: ir.Array}
	element := ir.ArrayIndex{Array: receiver, Index: ir.NumberConstant{}, Element: ir.Object}
	program := &ir.Program{Locals: []ir.Local{{Type: ir.Array}}, Main: []ir.Statement{ir.Declare{Local: 0, Value: ir.ArrayLiteral{GraphTypes: []int{-1}, Element: ir.Object, Elements: []ir.Expression{ir.ObjectLiteral{GraphTypes: []int{-2}}}}}}}
	if result := newAllocationFlowGraph(program).ReachingAllocations(element); result.Unknown || !reflect.DeepEqual(result.Sites, []int{-2}) {
		t.Fatalf("constant element lost allocation: %+v", result)
	}
	program.Main = append(program.Main, ir.Evaluate{Value: ir.ArrayPush{Array: receiver}})
	if result := newAllocationFlowGraph(program).ReachingAllocations(element); !result.Unknown {
		t.Fatal("mutable element certified")
	}
}
