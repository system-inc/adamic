package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) checkedJSONStringifyUse(node *ast.Node, value ir.Expression) ir.Expression {
	if value.Type() != ir.String || !l.includesUndefined(l.concrete(l.checker.GetTypeAtLocation(node))) || !l.program.RequiresJSONStringifySite(node) {
		return value
	}
	return ir.Coalesce{Value: value, Panic: ir.StringConstant{Index: l.constant("JSON.stringify result is undefined: " + l.program.Where(node))}, Of: ir.String}
}
