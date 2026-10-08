package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// placeholderDeclaration admits only literal slot initializers. A shadowed
// undefined binding is an ordinary assertion, as are calls and standalone values.
func (l *lowering) placeholderDeclaration(node *ast.Node) bool {
	if !l.uninitializedInitializer(node) {
		return false
	}
	for node.Parent != nil && (node.Parent.Kind == ast.KindParenthesizedExpression || node.Parent.Kind == ast.KindAsExpression) {
		node = node.Parent
	}
	if node.Parent == nil {
		return false
	}
	switch parent := node.Parent; parent.Kind {
	case ast.KindVariableDeclaration:
		return parent.AsVariableDeclaration().Initializer == node
	case ast.KindPropertyDeclaration:
		return parent.AsPropertyDeclaration().Initializer == node
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Initializer == node
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		return binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Right == node
	}
	return false
}

// placeholderZero carries no semantic value; the slot's readiness flag guards it.
func placeholderZero(of ir.Type) ir.Expression {
	if of.IsReference() {
		return ir.Undefined{Of: of}
	}
	if of == ir.MaybeBoolean {
		return ir.MaybeOf{Of: of}
	}
	return zeroValue(of)
}
