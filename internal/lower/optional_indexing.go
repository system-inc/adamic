package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Save the base before the guard. The index belongs only to the present arm,
// and the ordinary lookup retains its missing-element representation.
func (lowering *lowering) optionalIndex(node *ast.Node, base ir.Expression) (ir.Expression, error) {
	function := -1
	if lowering.function != nil {
		function = lowering.functionIndex
	}
	local := len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "optional_index_base", Type: base.Type(), Function: function, ExpressionAssigned: true},
	)
	read := ir.Read{Local: local, Of: base.Type()}
	value, optionalIndexValueError := lowering.optionalIndexValue(node, read)
	if optionalIndexValueError != nil {
		return nil, optionalIndexValueError
	}
	of := value.Type()
	return ir.Effects{Body: []ir.Statement{ir.Declare{Local: local, Value: base}}, Result: ir.Conditional{
		Condition: ir.Unary{
			Operator: ir.Not,
			Operand:  ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: read}, Right: ir.IsNull{Value: read}},
		},
		WhenTrue: value, WhenNot: fit(ir.Undefined{}, of), Of: of,
	}}, nil
}

func (lowering *lowering) optionalIndexValue(node *ast.Node, base ir.Expression) (ir.Expression, error) {
	access := node.AsElementAccessExpression()
	index, expressionError := lowering.expression(access.ArgumentExpression)
	if expressionError != nil {
		return nil, expressionError
	}
	if index.Type() != ir.Number {
		return nil, lowering.notYet(node, "an optional index that isn't a number")
	}
	var value ir.Expression
	if base.Type() == ir.String {
		value = ir.StringIndex{Value: base, Index: index}
	} else {
		element := ir.Number
		if !base.Type().IsTypedArray() {
			var elementTypeError error
			element, elementTypeError = lowering.elementType(access.Expression)
			if elementTypeError != nil {
				return nil, elementTypeError
			}
		}
		value = ir.ArrayIndex{Array: base, Index: index, Element: element}
	}
	return value, nil
}
