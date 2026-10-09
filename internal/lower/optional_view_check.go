package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func (l *lowering) checkedOptionalView(node *ast.Node, value ir.Expression, contract load.OptionalViewContract) (ir.Expression, error) {
	if contract.Tree != nil {
		return l.checkedOptionalNested(node, value, contract.Tree)
	}
	if value.Type() != ir.Object {
		return nil, l.notYet(node, "optional view check requires own-presence object storage")
	}
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "optionalView", Type: ir.Object, Function: l.functionIndex})
	l.noteLocal(held, l.concrete(l.checker.GetTypeAtLocation(node)), node)
	read := ir.Read{Local: held, Of: ir.Object}
	body := []ir.Statement{ir.Declare{Local: held, Value: value}}
	if contract.Defined {
		message := ir.StringConstant{Index: l.constant("optional contract produced undefined at " + l.program.Where(node))}
		body = append(body, ir.Evaluate{Value: ir.Coalesce{Value: read, Panic: message, Of: ir.Object}})
	}
	if contract.Storage || contract.Nullable {
		guard := ir.ObjectCall{Method: "optionalViewStorage", Arguments: []ir.Expression{read, ir.BooleanConstant{Value: contract.Nullable}}, Returns: ir.Boolean, Readiness: l.program.Where(node)}
		body = append(body, ir.Evaluate{Value: guard})
	}
	fieldGuards := []ir.Statement{}
	for _, name := range contract.Fields {
		guard := ir.ObjectCall{Method: "optionalWritePresence", Arguments: []ir.Expression{read, ir.StringConstant{Index: l.constant(name)}}, Returns: ir.Boolean, Readiness: l.program.Where(node)}
		fieldGuards = append(fieldGuards, ir.Evaluate{Value: guard})
	}
	if contract.Nullable {
		body = append(body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: read}}, Then: fieldGuards})
	} else {
		body = append(body, fieldGuards...)
	}
	l.program.RecordOptionalLiteral(node)
	return ir.Effects{Body: body, Result: read}, nil
}
