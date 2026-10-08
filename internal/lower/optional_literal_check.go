package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) checkedOptionalLiteral(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	if view, ok := l.program.OptionalViewSite(node); ok {
		return l.checkedOptionalView(node, value, view)
	}
	if value.Type() != ir.Object {
		return nil, l.notYet(node, "optional construction check requires own-presence object storage")
	}
	if node.AsObjectLiteralExpression().Properties.Nodes[0].Kind == ast.KindSpreadAssignment {
		literal, ok := value.(ir.ObjectLiteral)
		if !ok || literal.Spread == nil {
			return nil, l.notYet(node, "optional spread check requires a captured source before construction")
		}
	}
	body := []ir.Statement{}
	spreadKeys := -1
	if literal, ok := value.(ir.ObjectLiteral); ok && literal.Spread != nil {
		source := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "optionalSpreadSource", Type: ir.Object, Function: l.functionIndex})
		l.noteLocal(source, l.concrete(l.checker.GetTypeAtLocation(node.AsObjectLiteralExpression().Properties.Nodes[0].AsSpreadAssignment().Expression)), node)
		body = append(body, ir.Declare{Local: source, Value: literal.Spread})
		literal.Spread = ir.Read{Local: source, Of: ir.Object}
		literal.NoReuse = true
		spreadKeys = len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "optionalSpreadKeys", Type: ir.Array, Function: l.functionIndex})
		keys := ir.ObjectCall{Method: "optionalSpreadKeys", Arguments: []ir.Expression{literal.Spread}, Returns: ir.Array, Readiness: l.program.Where(node)}
		body = append(body, ir.Declare{Local: spreadKeys, Value: keys})
		value = literal
	}
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "optionalConstruction", Type: ir.Object, Function: l.functionIndex})
	l.noteLocal(held, l.concrete(l.checker.GetTypeAtLocation(node)), node)
	read := ir.Read{Local: held, Of: ir.Object}
	body = append(body, ir.Declare{Local: held, Value: value})
	if spreadKeys >= 0 {
		guard := ir.ObjectCall{Method: "optionalSpreadPresence", Arguments: []ir.Expression{read, ir.Read{Local: spreadKeys, Of: ir.Array}}, Returns: ir.Boolean, Readiness: l.program.Where(node)}
		body = append(body, ir.Evaluate{Value: guard})
	}
	for _, field := range node.AsObjectLiteralExpression().Properties.Nodes {
		if field.Kind == ast.KindSpreadAssignment {
			continue
		}
		name, known := l.methodName(field)
		if !known {
			return nil, l.notYet(field, "an optional construction field without a fixed name")
		}
		guard := ir.ObjectCall{Method: "optionalWritePresence", Arguments: []ir.Expression{read, ir.StringConstant{Index: l.constant(name)}}, Returns: ir.Boolean, Readiness: l.program.Where(node)}
		body = append(body, ir.Evaluate{Value: guard})
	}
	l.program.RecordOptionalLiteral(node)
	return ir.Effects{Body: body, Result: read}, nil
}
