package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Save the base before the guard. The index belongs only to the present arm,
// and the ordinary lookup retains its missing-element representation.
func (l *lowering) optionalIndex(node *ast.Node, base ir.Expression) (ir.Expression, error) {
	function := -1
	if l.function != nil {
		function = l.functionIndex
	}
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "optional_index_base", Type: base.Type(), Function: function, ExpressionAssigned: true})
	read := ir.Read{Local: local, Of: base.Type()}
	value, err := l.optionalIndexValue(node, read)
	if err != nil {
		return nil, err
	}
	of := value.Type()
	return ir.Effects{Body: []ir.Statement{ir.Declare{Local: local, Value: base}}, Result: ir.Conditional{
		Condition: ir.Unary{Operator: ir.Not, Operand: ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: read}, Right: ir.IsNull{Value: read}}},
		WhenTrue:  value, WhenNot: fit(ir.Undefined{}, of), Of: of,
	}}, nil
}

func (l *lowering) optionalIndexValue(node *ast.Node, base ir.Expression) (ir.Expression, error) {
	access := node.AsElementAccessExpression()
	index, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, err
	}
	if index.Type() != ir.Number {
		return nil, l.notYet(node, "an optional index that isn't a number")
	}
	var value ir.Expression
	if base.Type() == ir.String {
		value = ir.StringIndex{Value: base, Index: index}
	} else {
		element := ir.Number
		if !base.Type().IsTypedArray() {
			element, err = l.elementType(access.Expression)
			if err != nil {
				return nil, err
			}
		}
		value = ir.ArrayIndex{Array: base, Index: index, Element: element}
	}
	return value, nil
}
