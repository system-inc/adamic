package flow

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestArrayHolesExceptionEdges(t *testing.T) {
	for _, value := range []ir.Expression{
		ir.ArrayHoles{Length: ir.NumberConstant{Value: -1}, Element: ir.Number},
		ir.ArraySetLength{Array: ir.Read{Local: 0, Of: ir.Array}, Length: ir.NumberConstant{Value: -1}},
	} {
		var statement ir.Statement = ir.Evaluate{Value: value}
		instruction := &Instruction{At: &statement, Expression: value}
		if !CanThrow(&ir.Program{}, instruction) {
			t.Fatalf("%T must preserve the exception successor for catch liveness", value)
		}
	}
}
