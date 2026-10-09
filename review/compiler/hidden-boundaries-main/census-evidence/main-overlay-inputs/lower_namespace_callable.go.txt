package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

func (l *lowering) namespaceMergedFunction(node *ast.Node) bool {
	symbol := l.symbol(node)
	if symbol == nil || l.namespaceDeclaration(node) == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindFunctionDeclaration {
			return true
		}
	}
	return false
}

// A closure represents the callable identity, but has no attached object storage.
// Only uses that cannot observe that omitted container are admitted.
func (l *lowering) callableNamespaceUse(node *ast.Node) bool {
	if called(node) {
		return true
	}
	use := node
	for use.Parent != nil && use.Parent.Kind == ast.KindParenthesizedExpression {
		use = use.Parent
	}
	parent := use.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindTypeOfExpression {
		return true
	}
	if parent.Kind == ast.KindPropertyAccessExpression && parent.Expression() == use && l.namespaceMember(parent) {
		return true
	}
	if parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken || binary.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken {
			return l.namespaceFixedFunction(binary.Left) && l.namespaceFixedFunction(binary.Right)
		}
	}
	return false
}

func (l *lowering) namespaceFixedFunction(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindIdentifier && node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	symbol := l.symbol(node)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindFunctionDeclaration && (declaration.Parent.Kind == ast.KindModuleBlock || l.namespaceMergedFunction(node)) {
			return true
		}
	}
	return false
}

func callableNamespaceIntrinsic(name string) bool {
	switch name {
	case "name", "length", "prototype", "caller", "arguments", "call", "apply", "bind", "toString":
		return true
	}
	return false
}
