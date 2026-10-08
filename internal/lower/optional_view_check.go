package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) checkedOptionalView(node *ast.Node, value ir.Expression, names []string) (ir.Expression, error) {
	if value.Type() != ir.Object {
		return nil, l.notYet(node, "optional view check requires own-presence object storage")
	}
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "optionalView", Type: ir.Object, Function: l.functionIndex})
	l.noteLocal(held, l.concrete(l.checker.GetTypeAtLocation(node)), node)
	read := ir.Read{Local: held, Of: ir.Object}
	body := []ir.Statement{ir.Declare{Local: held, Value: value}}
	for _, name := range names {
		guard := ir.ObjectCall{Method: "optionalWritePresence", Arguments: []ir.Expression{read, ir.StringConstant{Index: l.constant(name)}}, Returns: ir.Boolean, Readiness: l.program.Where(node)}
		body = append(body, ir.Evaluate{Value: guard})
	}
	l.program.RecordOptionalLiteral(node)
	return ir.Effects{Body: body, Result: read}, nil
}
