package native

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestPrototypeArgumentsEscapeRegions(t *testing.T) {
	for _, method := range []string{"create", "setPrototypeOf"} {
		t.Run(method, func(t *testing.T) {
			count := 1
			if method == "setPrototypeOf" {
				count = 2
			}
			program := &ir.Program{}
			call := ir.ObjectCall{Method: method, Returns: ir.Object}
			parameters := []int{}
			for i := 0; i < count; i++ {
				program.Locals = append(program.Locals, ir.Local{Type: ir.Object, Function: 0})
				parameters = append(parameters, i)
				call.Arguments = append(call.Arguments, ir.Read{Local: i, Of: ir.Object})
			}
			program.Functions = []ir.Function{
				{Parameters: parameters, Body: []ir.Statement{ir.Evaluate{Value: call}}},
				{Returns: ir.Object, Body: []ir.Statement{ir.Return{Value: ir.ObjectLiteral{}}}},
			}
			outer := ir.Call{Function: 0}
			for range parameters {
				outer.Arguments = append(outer.Arguments, ir.Call{Function: 1, Returns: ir.Object})
			}
			program.Main = []ir.Statement{ir.Evaluate{Value: outer}}
			plan := planRegions(program)
			for position := range parameters {
				if !plan.escapes[0][position] {
					t.Fatalf("%s operand %d could be shorter lived than the receiver", method, position)
				}
			}
			if len(plan.statements) != 0 {
				t.Fatal("prototype arguments acquired a statement region")
			}
		})
	}
}
