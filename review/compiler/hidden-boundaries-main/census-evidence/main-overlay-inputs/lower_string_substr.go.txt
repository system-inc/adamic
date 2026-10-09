package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// substr counts UTF-16 units from a relative start, with a length rather than an end.
// A helper receives each operand once in source order before normalization reads it again.
func (l *lowering) stringSubstr(node *ast.Node, value ir.Expression, written []*ast.Node) (ir.Expression, error) {
	if len(written) > 2 {
		return nil, l.notYet(node, "substr with extra arguments")
	}
	values := []ir.Expression{value}
	for _, argument := range written {
		lowered, err := l.expression(argument)
		if err != nil {
			return nil, err
		}
		if _, missing := lowered.(ir.Undefined); missing {
			lowered = fit(lowered, ir.MaybeNumber)
		}
		if lowered.Type() != ir.Number && lowered.Type() != ir.MaybeNumber {
			return nil, l.notYet(argument, "substr with a nonnumeric argument")
		}
		values = append(values, lowered)
	}
	function, reads := l.stringHelper("substr", values)
	length := ir.StringLength{Value: reads[0]}
	zero := ir.NumberConstant{Value: 0}
	argument := func(index int, fallback ir.Expression) ir.Expression {
		if index >= len(reads) {
			return fallback
		}
		value := reads[index]
		if value.Type() == ir.MaybeNumber {
			value = ir.Coalesce{Value: value, Fallback: fallback, Of: ir.Number}
		}
		return stringInteger(value)
	}
	position := argument(1, zero)
	start := ir.Conditional{
		Condition: ir.Binary{Operator: ir.Less, Left: position, Right: zero},
		WhenTrue:  ir.MathCall{Function: "max", Arguments: []ir.Expression{ir.Binary{Operator: ir.Add, Left: length, Right: position}, zero}},
		WhenNot:   ir.MathCall{Function: "min", Arguments: []ir.Expression{position, length}},
	}
	count := ir.MathCall{Function: "min", Arguments: []ir.Expression{
		ir.MathCall{Function: "max", Arguments: []ir.Expression{argument(2, length), zero}},
		ir.Binary{Operator: ir.Subtract, Left: length, Right: start},
	}}
	result := ir.StringCall{Method: "slice", Value: reads[0], Arguments: []ir.Expression{start, ir.Binary{Operator: ir.Add, Left: start, Right: count}}}
	l.result.Functions[function].Body = []ir.Statement{ir.Return{Value: result}}
	return ir.Call{Function: function, Arguments: values, Returns: ir.String}, nil
}
