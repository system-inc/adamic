package ir_test

import (
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestMixedOptionalFixedLayoutPreservesMaybeNumber(t *testing.T) {
	t.Parallel()
	for _, targets := range [][]int{{0, 1}, {1, 0}} {
		program := ir.Program{
			Locals: []ir.Local{{Type: ir.Number}, {Type: ir.MaybeNumber}},
			Functions: []ir.Function{
				{Closure: true, Parameters: []int{0}},
				{Closure: true, Parameters: []int{1}},
			},
			FunctionTypeTargets: map[int][]int{1: targets},
		}
		layout := program.ClosureArgumentLayout(ir.CallClosure{FunctionType: 1})
		if want := []ir.Type{ir.MaybeNumber}; !reflect.DeepEqual(layout.Fixed, want) {
			t.Errorf("targets %v: fixed layout %v, want %v (MaybeNumber must preserve undefined)", targets, layout.Fixed, want)
		}
	}
}
