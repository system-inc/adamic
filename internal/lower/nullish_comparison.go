package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// An untagged nullable pointer cannot distinguish a stored null from a missing slot.
// The checker exposes both possibilities for nullable map and indexed array reads.
func (l *lowering) nullishReferenceComparison(node *ast.Node, left, right ir.Expression) error {
	isNullish := func(value ir.Expression) bool {
		switch value.(type) {
		case ir.Null, ir.Undefined:
			return true
		}
		return false
	}
	value, operand := left, node.AsBinaryExpression().Left
	if isNullish(left) {
		value, operand = right, node.AsBinaryExpression().Right
	} else if !isNullish(right) {
		return nil
	}
	if value.Type() == ir.Union || !value.Type().IsReference() {
		return nil
	}
	proven := l.checker.GetTypeAtLocation(operand)
	if l.includesNull(proven) && l.includesUndefined(proven) {
		return l.notYet(node, "null and undefined comparison without a tagged reference slot; store null in an explicitly tagged object and test missing separately")
	}
	return nil
}
