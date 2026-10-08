package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// Literal methods are owned closure fields, unlike class prototype dispatch or
// structural signatures whose origin is erased. Every root must prove that slot.
func (l *lowering) literalMethodSlot(node *ast.Node) bool {
	symbol := l.checker.GetSymbolAtLocation(node)
	if symbol == nil {
		return false
	}
	roots := l.checker.GetRootSymbols(symbol)
	if len(roots) == 0 {
		return false
	}
	for _, root := range roots {
		if len(root.Declarations) == 0 {
			return false
		}
		for _, declaration := range root.Declarations {
			if declaration.Kind != ast.KindMethodDeclaration || declaration.Parent == nil || declaration.Parent.Kind != ast.KindObjectLiteralExpression {
				return false
			}
		}
	}
	return true
}

// A simple assignment does not read the old method or detach its receiver.
func (l *lowering) literalMethodWrite(node *ast.Node) bool {
	if !l.literalMethodSlot(node) {
		return false
	}
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	if node.Parent == nil || node.Parent.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := node.Parent.AsBinaryExpression()
	return binary.Left == node && binary.OperatorToken.Kind == ast.KindEqualsToken
}
