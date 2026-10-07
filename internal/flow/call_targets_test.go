package flow_test

import (
	"github.com/system-inc/adamic/internal/flow"
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestCanThrowUsesEveryCallTarget(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Functions: []ir.Function{{}, {MayThrow: true}}, MethodTargets: map[int][]int{0: {0, 1}}, Locals: []ir.Local{{ConstantClosure: 1}, {}}}
	for _, test := range []struct {
		name       string
		expression ir.Expression
		throws     bool
	}{
		{"direct", ir.Call{Function: 0}, false},
		{"override", ir.Call{Function: 0, Virtual: 1}, true},
		{"literal", ir.CallClosure{Closure: ir.MakeClosure{Function: 0}}, false},
		{"const", ir.CallClosure{Closure: ir.Read{Local: 0}}, false},
		{"unknown", ir.CallClosure{Closure: ir.Read{Local: 1}}, true},
		{"map", ir.ArrayMap{Callback: ir.MakeClosure{Function: 1}}, true},
		{"visit", ir.ArrayVisit{Callback: ir.MakeClosure{Function: 1}}, true},
		{"reduce", ir.ArrayReduce{Callback: ir.MakeClosure{Function: 1}}, true},
		{"from", ir.ArrayFrom{Callback: ir.MakeClosure{Function: 1}}, true},
		{"map and set", ir.MapForEach{Callback: ir.MakeClosure{Function: 1}}, true},
		{"named comparator", ir.ArraySort{Comparator: 1}, true},
		{"closure comparator", ir.ArraySort{Callback: ir.MakeClosure{Function: 1}}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			var statement ir.Statement = ir.Evaluate{Value: test.expression}
			instruction := &flow.Instruction{At: &statement, Expression: test.expression}
			if got := flow.CanThrow(program, instruction); got != test.throws {
				t.Fatalf("CanThrow=%v, want %v", got, test.throws)
			}
		})
	}
}
