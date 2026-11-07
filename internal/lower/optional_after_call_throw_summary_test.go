package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestOptionalAfterCallPreservesEarlierThrow(test *testing.T) {
	test.Parallel()
	lowering := &lowering{result: &ir.Program{Functions: []ir.Function{{MayThrow: true}}}}
	statements := []ir.Statement{ir.Evaluate{Value: ir.Concat{Parts: []ir.Expression{
		ir.Call{Function: 0, Returns: ir.String},
		ir.Defined{Value: ir.Undefined{Of: ir.String}, Message: "type invariant"},
	}}}}

	if !lowering.throwsOut(statements) {
		test.Fatal("nonthrowing Defined erased an earlier throw")
	}
}
