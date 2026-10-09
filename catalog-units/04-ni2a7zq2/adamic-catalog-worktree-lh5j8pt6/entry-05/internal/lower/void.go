package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// A void initializer has no value to keep. Other expressions narrowed to undefined
// still use their declared storage, such as find's number | undefined result.
func voidInitializer(node *ast.Node) bool {
	if node.Kind != ast.KindIdentifier || node.Parent == nil || node.Parent.Kind != ast.KindVariableDeclaration {
		return false
	}
	initializer := node.Parent.AsVariableDeclaration().Initializer
	return initializer != nil && ast.SkipParentheses(initializer).Kind == ast.KindVoidExpression
}
