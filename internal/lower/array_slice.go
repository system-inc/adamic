package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Undefined start means zero; undefined end means the current length. Evaluate
// both bounds before observing that length, since an operand may change it.
func (l *lowering) libraryArraySlice(node *ast.Node, array ir.Expression) (ir.Expression, bool, error) {
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) > 2 {
		return nil, true, l.notYet(node, "slice with more than two arguments")
	}
	arguments := []ir.Expression{array}
	optional := false
	for _, argument := range written {
		value, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		if value.Type() != ir.Number {
			optional = true
		}
		converted := fit(value, ir.MaybeNumber)
		if converted.Type() != ir.MaybeNumber {
			return nil, true, l.notYet(argument, "a slice bound other than number or undefined")
		}
		arguments = append(arguments, value)
	}
	if !optional {
		return ir.ArraySlice{Array: array, Arguments: arguments[1:]}, true, nil
	}
	for index := 1; index < len(arguments); index++ {
		arguments[index] = fit(arguments[index], ir.MaybeNumber)
	}
	b := l.libraryArrayBuilder(arguments)
	source := b.read(b.parameters[0])
	bounds := []ir.Expression{}
	for index, parameter := range b.parameters[1:] {
		fallback := ir.Expression(ir.NumberConstant{Value: 0})
		if index == 1 {
			fallback = ir.Length{Array: source}
		}
		bounds = append(bounds, ir.Coalesce{Value: b.read(parameter), Fallback: fallback, Of: ir.Number, UndefinedOnly: true})
	}
	return b.finish("array_slice_optional_bounds", ir.ArraySlice{Array: source, Arguments: bounds}), true, nil
}
