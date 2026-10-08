package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

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

// The admitted namespace exports have fixed function identities. Compare their
// declaration symbols without manufacturing a callable object or its properties.
// Readiness checks still precede the comparison and preserve source order.
func (l *lowering) namespaceFunctionReference(node *ast.Node) *ast.Symbol {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindIdentifier && node.Kind != ast.KindPropertyAccessExpression {
		return nil
	}
	symbol := l.symbol(node)
	if symbol == nil {
		return nil
	}
	function := false
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindFunctionDeclaration && (declaration.Parent.Kind == ast.KindModuleBlock || l.namespaceMergedFunction(node)) {
			function = true
		}
	}
	if !function {
		return nil
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		for receiver := node.Expression(); receiver != nil; {
			receiver = ast.SkipParentheses(receiver)
			if l.namespaceDeclaration(receiver) == nil {
				return nil
			}
			if receiver.Kind == ast.KindIdentifier {
				break
			}
			if receiver.Kind != ast.KindPropertyAccessExpression {
				return nil
			}
			receiver = receiver.Expression()
		}
	}
	return symbol
}
func (l *lowering) namespaceFunctionIdentityOperands(node *ast.Node) (*ast.Symbol, *ast.Symbol, bool) {
	if node.Kind != ast.KindBinaryExpression {
		return nil, nil, false
	}
	binary := node.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken && binary.OperatorToken.Kind != ast.KindExclamationEqualsEqualsToken {
		return nil, nil, false
	}
	left, right := l.namespaceFunctionReference(binary.Left), l.namespaceFunctionReference(binary.Right)
	return left, right, left != nil && right != nil
}
func (l *lowering) namespaceFunctionIdentity(node *ast.Node) (ir.Expression, bool) {
	left, right, known := l.namespaceFunctionIdentityOperands(node)
	if !known {
		return nil, false
	}
	binary := node.AsBinaryExpression()
	equal := left == right
	if binary.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken {
		equal = !equal
	}
	value := ir.Expression(ir.BooleanConstant{Value: equal})
	value = l.namespaceReadyValue(ast.SkipParentheses(binary.Right), value)
	value = l.namespaceReadyValue(ast.SkipParentheses(binary.Left), value)
	return value, true
}
