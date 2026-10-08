package flow

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

// A missing edge would let ownership and SSA proofs ignore the catch path.
func TestRegExpProtocolExceptionEdges(t *testing.T) {
	for _, one := range []struct {
		value  ir.Expression
		throws bool
	}{
		{ir.RegExpNew{Invalid: true}, true},
		{ir.RegExpNew{}, false},
		{ir.RegExpCall{Method: "matchAll"}, true},
		{ir.RegExpCall{Method: "replaceAll"}, true},
		{ir.RegExpCall{Method: "replaceCallback"}, true},
		{ir.RegExpCall{Method: "symbol:replaceCallback"}, true},
		{ir.RegExpCall{Method: "symbol:matchAll"}, false},
	} {
		statement := ir.Statement(ir.Evaluate{Value: one.value})
		instruction := &Instruction{At: &statement, Expression: one.value}
		if got := CanThrow(&ir.Program{}, instruction); got != one.throws {
			t.Errorf("missing or spurious exception edge for %#v: got %v", one.value, got)
		}
	}
}
