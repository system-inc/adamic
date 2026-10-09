package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// Arrows inherit this, but a nested ordinary function, object method or class
// owns a separate receiver. Its body is checked by that construct's lowering.
func namespaceOwnThis(declaration *ast.Node) bool {
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil || ast.IsTypeNode(node) {
			return false
		}
		if ast.IsClassLike(node) || ast.IsFunctionLike(node) && node.Kind != ast.KindArrowFunction {
			return false
		}
		if node.Kind == ast.KindThisKeyword {
			return true
		}
		return node.ForEachChild(visit)
	}
	return visit(declaration.Body())
}
