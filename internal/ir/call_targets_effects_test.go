package ir_test

import (
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestCallMayThrowUsesReachableTargets(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Functions: []ir.Function{{MayThrow: true}, {}, {MayThrow: true}, {}},
		MethodTargets: map[int][]int{
			0: {1, 3},
			1: {1, 2},
			3: {1, 3},
		},
	}
	cases := []struct {
		name string
		call ir.Call
		want bool
	}{
		{"direct throwing", ir.Call{Function: 0}, true},
		{"direct nonthrowing", ir.Call{Function: 1}, false},
		{"throwing later override", ir.Call{Function: 1, Virtual: 1}, true},
		{"nonthrowing overrides", ir.Call{Function: 3, Virtual: 1}, false},
		{"static throw outside target set", ir.Call{Function: 0, Virtual: 1}, false},
		{"structural accessor override", ir.Call{Function: 1, Virtual: -1}, true},
	}
	for _, each := range cases {
		if got := program.CallMayThrow(each.call); got != each.want {
			t.Errorf("%s: CallMayThrow = %t, want %t", each.name, got, each.want)
		}
	}
}

func TestDirectClosureTargetsUseEncodedIndex(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Functions: []ir.Function{{}, {MayThrow: true}, {}, {}, {MayThrow: true}},
		Locals:    []ir.Local{{ConstantClosure: 5}},
	}
	cases := []struct {
		name string
		call ir.CallClosure
		want int
	}{
		{"first function", ir.CallClosure{Direct: 1, Closure: ir.ClosureSelf{}}, 0},
		{"direct beats literal", ir.CallClosure{Direct: 2, Closure: ir.MakeClosure{Function: 4}}, 1},
		{"direct beats constant binding", ir.CallClosure{Direct: 2, Closure: ir.Read{Local: 0}}, 1},
		{"later function", ir.CallClosure{Direct: 4, Closure: ir.ClosureSelf{}}, 3},
		{"zero selects closure value", ir.CallClosure{Closure: ir.MakeClosure{Function: 4}}, 4},
	}
	for _, each := range cases {
		got := program.ClosureTargets(each.call)
		if got.Unknown || !reflect.DeepEqual(got.Functions, []int{each.want}) {
			t.Errorf("%s: ClosureTargets = %+v, want known target [%d]", each.name, got, each.want)
		}
		if got := program.ClosureMayThrow(each.call); got != program.Functions[each.want].MayThrow {
			t.Errorf("%s: ClosureMayThrow = %t, want target %d's effect", each.name, got, each.want)
		}
	}
}
