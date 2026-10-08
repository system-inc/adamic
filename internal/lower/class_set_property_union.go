package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// A narrowed access still reads the declared union slot. Calls through aliases can
// invalidate the checker's narrowing, so check the runtime member before unpacking.
func (l *lowering) unionFieldRead(node *ast.Node, property ir.Property) (ir.Expression, bool) {
	field := l.checker.GetSymbolAtLocation(node.Name())
	if field == nil {
		return nil, false
	}
	declared, known := l.representation(l.checker.GetTypeOfSymbol(field))
	if !known || declared != ir.Union {
		return nil, false
	}
	observed := property.Of
	property.Of = ir.Union
	property.Absent = field.Flags&ast.SymbolFlagsOptional != 0
	if observed == ir.Union {
		return property, true
	}
	name := "object"
	switch observed.Present() {
	case ir.Number:
		name = "number"
	case ir.Boolean:
		name = "boolean"
	case ir.String:
		name = "string"
	case ir.Closure:
		name = "function"
	}
	b := l.libraryArrayBuilder([]ir.Expression{property})
	held := b.read(b.parameters[0])
	matches := ir.Expression(ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: held}, Right: ir.StringConstant{Index: l.constant(name)}})
	if name == "object" {
		// typeof deliberately groups arrays, maps and plain objects. Their native
		// layouts differ, so test the representation tag before any pointer cast.
		matches = ir.Binary{Operator: ir.Equal, Left: ir.NumberCall{Function: "unionFieldKind", Arguments: []ir.Expression{held, ir.NumberConstant{Value: float64(observed)}}}, Right: ir.BooleanConstant{Value: true}}
	}
	if l.includesUndefined(l.checker.GetTypeAtLocation(node)) {
		matches = ir.Binary{Operator: ir.Or, Left: matches, Right: ir.IsUndefined{Value: held}}
	}
	b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant("union field member where the checker narrowed it away: a call since the narrowing put it back")}}}})
	return b.finish("narrowed_union_field", ir.Narrow{Value: held, To: observed}), true
}
