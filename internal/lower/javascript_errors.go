package lower

import (
	"math"

	"github.com/system-inc/adamic/internal/ir"
)

// checkedRepeat throws JavaScript's RangeErrors using the existing ownership and
// exception graph, before the C helper can panic. Allocation failures and Adamic
// soundness checks remain terminal. Both receiver and count are evaluated once.
func (l *lowering) checkedRepeat(value, count ir.Expression) ir.Expression {
	// Keep already-proven literal operations unchanged, including their counts.
	if text, known := value.(ir.StringConstant); known {
		if n, known := count.(ir.NumberConstant); known && !math.IsInf(n.Value, 0) && math.Trunc(n.Value) >= 0 && float64(len(l.result.Strings[text.Index]))*math.Trunc(n.Value) <= 536870888 {
			return ir.StringCall{Method: "repeat", Value: value, Arguments: []ir.Expression{count}, FailuresChecked: true}
		}
	}
	b := l.libraryArrayBuilder([]ir.Expression{value, count})
	receiver, raw := b.read(b.parameters[0]), b.read(b.parameters[1])
	// ToIntegerOrInfinity converts NaN to zero and truncates finite fractions.
	integer := b.declare("repeat_integer", ir.Conditional{
		Condition: ir.NumberCall{Function: "isNaN", Arguments: []ir.Expression{raw}},
		WhenTrue:  ir.NumberConstant{}, WhenNot: ir.MathCall{Function: "trunc", Arguments: []ir.Expression{raw}},
	})
	n := b.read(integer)
	badCount := ir.Binary{Operator: ir.Or,
		Left:  ir.Binary{Operator: ir.Less, Left: n, Right: ir.NumberConstant{}},
		Right: ir.Unary{Operator: ir.Not, Operand: ir.NumberCall{Function: "isFinite", Arguments: []ir.Expression{n}}},
	}
	throwing := func(message ir.Expression) []ir.Statement {
		return []ir.Statement{ir.Throw{Value: ir.MakeError{Name: ir.StringConstant{Index: l.constant("RangeError")}, Message: message}}}
	}
	// V8 formats the original count, not its truncated integer, in this diagnostic.
	countMessage := ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: l.constant("Invalid count value: ")}, ir.NumberToString{Value: raw}}}
	b.body = append(b.body, ir.If{Condition: badCount, Then: throwing(countMessage)})
	// This is the same V8 UTF-16 limit enforced by the runtime, not a byte length.
	tooLong := ir.Binary{Operator: ir.Greater, Left: ir.Binary{Operator: ir.Multiply, Left: ir.StringLength{Value: receiver}, Right: n}, Right: ir.NumberConstant{Value: 536870888}}
	b.body = append(b.body, ir.If{Condition: tooLong, Then: throwing(ir.StringConstant{Index: l.constant("Invalid string length")})})
	return b.finish("javascript_checked_repeat", ir.StringCall{Method: "repeat", Value: receiver, Arguments: []ir.Expression{n}, FailuresChecked: true})
}
