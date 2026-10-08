package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Restrict a backward search to starts at or before position by keeping the
// prefix ending at position plus the needle length. All lengths are UTF-16.
// Capture operands once before using them for clamping and searching.
func (l *lowering) stringLastIndexOf(node *ast.Node, value ir.Expression, written []*ast.Node) (ir.Expression, error) {
	if len(written) < 1 || len(written) > 2 {
		return nil, l.notYet(node, "lastIndexOf with missing search or extra arguments")
	}
	search, err := l.expression(written[0])
	if err != nil {
		return nil, err
	}
	if search.Type() != ir.String || l.includesUndefined(l.checker.GetTypeAtLocation(written[0])) || l.includesNull(l.checker.GetTypeAtLocation(written[0])) {
		return nil, l.notYet(written[0], "lastIndexOf with a search other than a present string")
	}
	values := []ir.Expression{value, search}
	if len(written) == 2 {
		position, err := l.expression(written[1])
		if err != nil {
			return nil, err
		}
		position = fit(position, ir.MaybeNumber)
		if position.Type() != ir.MaybeNumber {
			return nil, l.notYet(written[1], "lastIndexOf with a position other than number or undefined")
		}
		values = append(values, position)
	}
	function, reads := l.stringHelper("last_index_of", values)
	l.result.Functions[function].Returns = ir.Number
	length := ir.StringLength{Value: reads[0]}
	position := ir.Expression(length)
	if len(reads) > 2 {
		given := ir.Coalesce{Value: reads[2], Fallback: length, Of: ir.Number, UndefinedOnly: true}
		position = ir.Conditional{Condition: ir.NumberCall{Function: "isNaN", Arguments: []ir.Expression{given}}, WhenTrue: length, WhenNot: ir.MathCall{Function: "trunc", Arguments: []ir.Expression{given}}}
	}
	zero := ir.NumberConstant{Value: 0}
	position = ir.MathCall{Function: "min", Arguments: []ir.Expression{ir.MathCall{Function: "max", Arguments: []ir.Expression{position, zero}}, length}}
	end := ir.MathCall{Function: "min", Arguments: []ir.Expression{length, ir.Binary{Operator: ir.Add, Left: position, Right: ir.StringLength{Value: reads[1]}}}}
	prefix := ir.StringCall{Method: "slice", Value: reads[0], Arguments: []ir.Expression{zero, end}}
	result := ir.StringCall{Method: "lastIndexOf", Value: prefix, Arguments: []ir.Expression{reads[1]}}
	l.result.Functions[function].Body = []ir.Statement{ir.Return{Value: result}}
	return ir.Call{Function: function, Arguments: values, Returns: ir.Number}, nil
}
