package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// substr's second argument counts UTF-16 units rather than naming an end.
// Capture all arguments once, then use the existing slice representation and
// integer conversion. Missing length means the remaining units; NaN means zero.
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
		lowered = fit(lowered, ir.MaybeNumber)
		if lowered.Type() != ir.MaybeNumber {
			return nil, l.notYet(argument, "a substr argument other than number or undefined")
		}
		values = append(values, lowered)
	}
	function, reads := l.stringHelper("substr", values)
	zero := ir.NumberConstant{Value: 0}
	length := ir.StringLength{Value: reads[0]}
	start := ir.Expression(zero)
	if len(reads) > 1 {
		start = stringInteger(ir.Coalesce{Value: reads[1], Fallback: zero, Of: ir.Number, UndefinedOnly: true})
	}
	relative := ir.Conditional{Condition: ir.Binary{Operator: ir.Less, Left: start, Right: zero}, WhenTrue: ir.Binary{Operator: ir.Add, Left: length, Right: start}, WhenNot: start}
	from := ir.MathCall{Function: "min", Arguments: []ir.Expression{ir.MathCall{Function: "max", Arguments: []ir.Expression{relative, zero}}, length}}
	remaining := ir.Binary{Operator: ir.Subtract, Left: length, Right: from}
	count := ir.Expression(remaining)
	if len(reads) > 2 {
		count = stringInteger(ir.Coalesce{Value: reads[2], Fallback: remaining, Of: ir.Number, UndefinedOnly: true})
	}
	count = ir.MathCall{Function: "min", Arguments: []ir.Expression{ir.MathCall{Function: "max", Arguments: []ir.Expression{count, zero}}, remaining}}
	end := ir.Binary{Operator: ir.Add, Left: from, Right: count}
	result := ir.StringCall{Method: "slice", Value: reads[0], Arguments: []ir.Expression{from, end}}
	l.result.Functions[function].Body = []ir.Statement{ir.Return{Value: result}}
	return ir.Call{Function: function, Arguments: values, Returns: ir.String}, nil
}
