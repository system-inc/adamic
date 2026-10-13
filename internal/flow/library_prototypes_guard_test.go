package flow

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestPrototypeMutationHasExceptionalEdge(t *testing.T) {
	t.Parallel()
	for _, method := range []string{"create", "setPrototypeOf", "getPrototypeOf"} {
		call := ir.ObjectCall{Method: method, Returns: ir.Object}
		var statement ir.Statement = ir.Evaluate{Value: call}
		instruction := Instruction{At: &statement, Expression: call}
		if CanThrow(&ir.Program{}, &instruction) != (method == "setPrototypeOf") {
			t.Fatalf("wrong exceptional edge for %s", method)
		}
	}
}
