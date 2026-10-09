package fresh

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestCheckedJSONPreservesIdentityAndOperandWrites(t *testing.T) {
	read := ir.Read{Local: 1, Of: ir.Object}
	check := func(value ir.Expression) ir.CheckedJSON {
		return ir.CheckedJSON{Value: value, Contract: &ir.JSONContract{Kind: "object", Name: "Config"}, Of: ir.Object}
	}
	program := &ir.Program{Locals: []ir.Local{{Name: "parameter", Type: ir.Union, Function: 0}, {Name: "holder", Type: ir.Object, Function: -1}}, Functions: []ir.Function{{Name: "convert", Parameters: []int{0}, Returns: ir.Object, Body: []ir.Statement{ir.Return{Value: check(ir.Read{Local: 0, Of: ir.Union})}}}}}
	for _, wrapped := range []bool{false, true} {
		t.Run(map[bool]string{false: "identity", true: "operand_write"}[wrapped], func(t *testing.T) {
			write := ir.SetProperty{Object: read, Name: "next", Value: ir.Call{Function: 0, Arguments: []ir.Expression{ir.Box{Value: read}}, Returns: ir.Object}, Site: 1}
			program.Main = []ir.Statement{ir.Declare{Local: 1, Value: ir.ObjectLiteral{}}}
			if wrapped {
				program.Main = append(program.Main, ir.Evaluate{Value: check(ir.Effects{Body: []ir.Statement{write}, Result: read})})
			} else {
				program.Main = append(program.Main, write)
			}
			writes := ProveWrites(program)
			if len(writes) != 1 || writes[0].Kind != WriteField || writes[0].Proven {
				t.Fatalf("validation must preserve the self-reaching value and its write: %+v", writes)
			}
		})
	}
	program.Main = []ir.Statement{ir.Evaluate{Value: check(ir.ObjectLiteral{})}}
	if writes := ProveWrites(program); len(writes) != 0 {
		t.Fatalf("validation must record no source write: %+v", writes)
	}
}
