package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestOptionalAfterCallPreservesEarlierThrow(t *testing.T) {
	t.Parallel()
	l := &lowering{result: &ir.Program{Functions: []ir.Function{{MayThrow: true}}}}
	statements := []ir.Statement{ir.Evaluate{Value: ir.Concat{Parts: []ir.Expression{
		ir.Call{Function: 0, Returns: ir.String},
		ir.Defined{Value: ir.Undefined{Of: ir.String}, Message: "type invariant"},
	}}}}

	if !l.throwsOut(statements) {
		t.Fatal("nonthrowing Defined erased an earlier throw")
	}
}
