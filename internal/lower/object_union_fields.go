package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Read the declared box before trusting a flow-narrowed field. Calls can change
// the field after the checker narrowed it, so the cast checks its actual member.
func (l *lowering) readUnionObjectField(node *ast.Node, property ir.Property, absent bool) ir.Expression {
	to := property.Of
	property.Of = ir.Union
	property.Absent = absent
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if to == ir.Union || comparedWithUndefined(node) || (parent != nil && parent.Kind == ast.KindTypeOfExpression) {
		return property
	}
	b := l.libraryArrayBuilder([]ir.Expression{property})
	held := b.read(b.parameters[0])
	matches := ir.Expression(ir.NumberCall{Function: "unionHasRepresentation", Arguments: []ir.Expression{held, ir.NumberConstant{Value: float64(to.Present())}}})
	if l.includesUndefined(l.checker.GetTypeAtLocation(node)) {
		matches = ir.Binary{Operator: ir.Or, Left: matches, Right: ir.IsUndefined{Value: held}}
	}
	message := "union member where the checker narrowed it away: a call since the narrowing put it back"
	b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}})
	return b.finish("narrowed_union_field", ir.Narrow{Value: held, To: to})
}
