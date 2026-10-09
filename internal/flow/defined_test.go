package flow

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A catch must keep its reads live across a failed required read. An invariant
// panic has no successor, and an optional phantom member cannot fail its read.
func TestDefinedExceptionEdges(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		value  ir.Expression
		throws bool
	}{
		{"required read", ir.Defined{Value: ir.Read{Local: 0, Of: ir.Object}, Message: "TypeError: Cannot read properties of undefined (reading 'label')"}, true},
		{"invariant", ir.Defined{Value: ir.Read{Local: 0, Of: ir.Object}, Message: "undefined where the checker narrowed it away"}, false},
		{"required phantom", ir.PhantomMember{Value: ir.Read{Local: 0, Of: ir.Object}, Name: "marker"}, true},
		{"optional phantom", ir.PhantomMember{Value: ir.Read{Local: 0, Of: ir.Object}, Name: "marker", Optional: true}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var statement ir.Statement = ir.Evaluate{Value: test.value}
			instruction := &Instruction{At: &statement, Expression: test.value}
			if got := CanThrow(&ir.Program{}, instruction); got != test.throws {
				t.Fatalf("catch successor: got %t, want %t", got, test.throws)
			}
		})
	}
}
