package fresh_test

import (
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

// Reading arity executes its receiver, including writes made while producing a callback.
func TestFunctionLengthReceiverStillJudgesCycleWrites(t *testing.T) {
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.Object, Function: 0}},
		Functions: []ir.Function{
			{Name: "make", Parameters: []int{0}, Returns: ir.Closure, Body: []ir.Statement{
				ir.SetProperty{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "next", Value: ir.Read{Local: 0, Of: ir.Object}, Site: 1},
				ir.Return{Value: ir.MakeClosure{Function: 1}},
			}},
			{Name: "callback", Closure: true, Returns: ir.Number, Body: []ir.Statement{ir.Return{Value: ir.NumberConstant{}}}},
		},
		Main: []ir.Statement{ir.Evaluate{Value: ir.ClosureLength{Value: ir.Call{Function: 0, Arguments: []ir.Expression{ir.ObjectLiteral{}}, Returns: ir.Closure}}}},
	}
	for _, write := range fresh.ProveWrites(program) {
		if write.Site == 1 && write.Kind == fresh.WriteField && !write.Proven {
			return
		}
	}
	t.Fatal("function length lost its receiver's cycle-closing write")
}
