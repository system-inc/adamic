package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Branded void has an undefined payload, independently of the pointer family
// chosen for a union's storage. Evaluate and check it before choosing that family.
func (l *lowering) fitMapEntry(node *ast.Node, value ir.Expression, target ir.Type) ir.Expression {
	if target.IsReference() && l.phantomUndefined(l.checker.GetTypeAtLocation(node)) {
		return ir.Narrow{Value: value, To: target, Undefined: true, UndefinedWhere: sourceExpression(node)}
	}
	return fit(value, target)
}
