package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Object's call/apply adapters inspect the actual argument shape and storage.
// The checker's generic dictionary overload is not a representation conversion.
func (l *lowering) recordLibraryArgument(node *ast.Node) bool {
	at := node
	for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
		at = at.Parent
	}
	if at.Parent != nil && at.Parent.Kind == ast.KindArrayLiteralExpression {
		at = at.Parent // The dense tuple supplied to apply is unpacked by the adapter.
	}
	parent := at.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression || parent.AsCallExpression().Expression == at {
		return false
	}
	callee := ast.SkipParentheses(parent.AsCallExpression().Expression)
	if callee.Kind == ast.KindPropertyAccessExpression && (callee.Name().Text() == "call" || callee.Name().Text() == "apply") {
		callee = callee.AsPropertyAccessExpression().Expression
	}
	method, known := l.libraryMethod(callee, map[*ast.Symbol]bool{})
	return known && method.family == "Object"
}
