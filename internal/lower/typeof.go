package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// typeOfNull records what a missing pointer means before the checker type is reduced to a
// representation. A declared slot can hold null again after a call, despite a stale narrowing.
func (l *lowering) typeOfNull(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	proven := l.checker.GetTypeAtLocation(node)
	at := node
	if node.Kind == ast.KindPropertyAccessExpression {
		at = node.Name()
	}
	if node.Kind == ast.KindIdentifier || node.Kind == ast.KindPropertyAccessExpression {
		if symbol := l.checker.GetSymbolAtLocation(at); symbol != nil {
			declared := l.checker.GetTypeOfSymbol(symbol)
			if l.includesNull(declared) {
				proven = declared
			}
		}
	}
	return l.includesNull(proven) && !l.includesUndefined(proven)
}
