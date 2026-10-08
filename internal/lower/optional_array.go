package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// optionalArrayElement saves the receiver once and evaluates the index only
// when it is present. The ordinary array read retains its missing-index result
// and the usual readiness check when the checker excludes undefined.
func (l *lowering) optionalArrayElement(node *ast.Node, array ir.Expression) (ir.Expression, error) {
	access := node.AsElementAccessExpression()
	element, err := l.elementType(access.Expression)
	if err != nil {
		return nil, err
	}
	position, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, err
	}
	if position.Type() != ir.Number {
		return nil, l.notYet(node, "an array index that isn't a number")
	}
	function := -1
	if l.function != nil {
		function = l.functionIndex
	}
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "optional_array_receiver", Type: ir.Array, Function: function, ExpressionAssigned: true})
	read := ir.Read{Local: local, Of: ir.Array}
	indexed := ir.ArrayIndex{Array: read, Index: position, Element: element}
	result := ir.Effects{Body: []ir.Statement{ir.Declare{Local: local, Value: array}}, Result: ir.Conditional{
		Condition: ir.Truthy{Value: read}, WhenTrue: indexed, WhenNot: fit(ir.Undefined{}, indexed.Type()), Of: indexed.Type(),
	}}
	return l.defined(node, result), nil
}
