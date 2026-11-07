package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// typeOfNull records what a missing pointer means before the checker type is reduced to a
// representation. A declared slot can hold null again after a call, despite a stale narrowing.
func (lowering *lowering) typeOfNull(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	proven := lowering.concrete(lowering.checker.GetTypeAtLocation(node))
	if node.Kind == ast.KindElementAccessExpression {
		receiver := lowering.checker.GetNonNullableType(
			lowering.concrete(lowering.checker.GetTypeAtLocation(node.AsElementAccessExpression().Expression)),
		)
		if lowering.isArrayType(receiver) && lowering.includesNull(lowering.elementTypeOfArray(receiver)) {
			return true
		}
	}
	at := node
	if node.Kind == ast.KindPropertyAccessExpression {
		at = node.Name()
	}
	if node.Kind == ast.KindIdentifier || node.Kind == ast.KindPropertyAccessExpression {
		if symbol := lowering.checker.GetSymbolAtLocation(at); symbol != nil {
			declared := lowering.concrete(lowering.checker.GetTypeOfSymbol(symbol))
			if lowering.includesNull(declared) {
				proven = declared
			}
		}
	}
	return lowering.includesNull(proven)
}
