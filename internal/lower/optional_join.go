package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Only receiver-optional intrinsic join is supported here. Its arguments belong
// inside the present branch, and the receiver must be held before that branch.
func (l *lowering) isOptionalJoin(node *ast.Node) bool {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if call.QuestionDotToken != nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	if access.QuestionDotToken == nil || access.Name().Text() != "join" {
		return false
	}
	of, known := l.representation(l.checker.GetTypeAtLocation(access.Expression))
	return known && of == ir.Array && l.libraryMember(callee)
}

func (l *lowering) optionalJoin(node *ast.Node) (ir.Expression, error) {
	receiver := ast.SkipParentheses(node.AsCallExpression().Expression).AsPropertyAccessExpression().Expression
	value, _, err := l.arrayMethod(node, receiver, "join")
	if err != nil {
		return nil, err
	}
	join := value.(ir.ArrayJoin)
	function := -1
	if l.function != nil {
		function = l.functionIndex
	}
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "optional_join_receiver", Type: ir.Array, Function: function, ExpressionAssigned: true})
	read := ir.Read{Local: local, Of: ir.Array}
	array := join.Array
	join.Array = read
	return ir.Effects{Body: []ir.Statement{ir.Declare{Local: local, Value: array}}, Result: ir.Conditional{
		Condition: ir.Truthy{Value: read}, WhenTrue: join, WhenNot: ir.Undefined{}, Of: ir.String,
	}}, nil
}
