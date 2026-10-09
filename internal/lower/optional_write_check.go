package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Holding the receiver preserves JavaScript evaluation order even when the
// value expression changes the variable from which the receiver was read.
func (l *lowering) checkedOptionalWrite(target *ast.Node, store ir.SetProperty) []ir.Statement {
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "optionalWriteReceiver", Type: ir.Object, Function: l.functionIndex})
	l.noteLocal(held, l.concrete(l.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression)), target)
	receiver := store.Object
	store.Object = ir.Read{Local: held, Of: ir.Object}
	guard := ir.ObjectCall{Method: "optionalWritePresence", Arguments: []ir.Expression{store.Object, ir.StringConstant{Index: l.constant(store.Name)}}, Returns: ir.Boolean, Readiness: l.program.Where(target)}
	l.program.RecordOptionalWrite(target)
	return []ir.Statement{ir.Block{Body: []ir.Statement{ir.Declare{Local: held, Value: receiver}, store, ir.Evaluate{Value: guard}}}}
}
