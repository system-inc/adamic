package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Intrinsic calls are recognized before generic Object coercion so sibling
// dispatch cannot fold the inherited iterator family's tag to Object.
func (l *lowering) libraryIteratorObservation(node *ast.Node) (ir.Expression, bool, error) {
	if node.Kind != ast.KindCallExpression {
		return nil, false, nil
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	var receiver *ast.Node
	if l.isLibraryGlobal(callee, "String") && len(call.Arguments.Nodes) == 1 {
		receiver = call.Arguments.Nodes[0]
	}
	if callee.Kind == ast.KindPropertyAccessExpression {
		access := callee.AsPropertyAccessExpression()
		if (callee.Name().Text() == "toString" || callee.Name().Text() == "toLocaleString") && len(call.Arguments.Nodes) == 0 && l.libraryMember(callee) {
			receiver = access.Expression
		}
		if callee.Name().Text() == "call" && len(call.Arguments.Nodes) == 1 {
			method := ast.SkipParentheses(access.Expression)
			if method.Kind == ast.KindPropertyAccessExpression && method.Name().Text() == "toString" {
				proto := ast.SkipParentheses(method.AsPropertyAccessExpression().Expression)
				if proto.Kind == ast.KindPropertyAccessExpression && proto.Name().Text() == "prototype" && l.isLibraryGlobal(proto.AsPropertyAccessExpression().Expression, "Object") {
					receiver = call.Arguments.Nodes[0]
				}
			}
		}
	}
	if receiver == nil || !l.libraryIteratorTagType(l.checker.GetTypeAtLocation(receiver)) {
		return nil, false, nil
	}
	if err := l.optionalCall(node); err != nil {
		return nil, true, err
	}
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	return l.libraryIteratorString(value), true, nil
}
