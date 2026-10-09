package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// sys.ts binds its fixed stat options with as const. This proves the same exact
// object layout as a plain literal, without permitting annotations or other casts.
func (l *lowering) nodeFSDirectoryStatOptions(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if !ast.IsIdentifier(node) {
		return false
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}
	variable := declaration.AsVariableDeclaration()
	if variable.Type != nil || variable.Initializer == nil {
		return false
	}
	initializer := ast.SkipParentheses(variable.Initializer)
	if initializer.Kind != ast.KindAsExpression {
		return false
	}
	as := initializer.AsAsExpression()
	return as.Type.Kind == ast.KindTypeReference && ast.IsIdentifier(as.Type.AsTypeReferenceNode().TypeName) &&
		as.Type.AsTypeReferenceNode().TypeName.Text() == "const" && l.exactObject(as.Expression, 0)
}
