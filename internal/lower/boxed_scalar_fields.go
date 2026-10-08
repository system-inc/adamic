package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// A field keeps its declared union even when flow analysis narrows a read to a
// scalar. Evaluate that stored value once and verify its tag before unboxing.
func (l *lowering) checkedBoxedScalarField(node *ast.Node, read ir.Expression, narrowed ir.Type) ir.Expression {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		node, parent = parent, parent.Parent
	}
	if comparedWithUndefined(node) || (parent != nil && parent.Kind == ast.KindTypeOfExpression) {
		return read
	}
	if parent != nil && parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
			// Equality observes the actual tag, including a stale flow fact.
			return read
		}
	}
	name := "number"
	if narrowed.Present() == ir.Boolean {
		name = "boolean"
	}
	b := l.libraryArrayBuilder([]ir.Expression{read})
	held := b.read(b.parameters[0])
	matches := ir.Expression(ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: held}, Right: ir.StringConstant{Index: l.constant(name)}})
	if narrowed.IsMaybe() {
		matches = ir.Binary{Operator: ir.Or, Left: matches, Right: ir.IsUndefined{Value: held}}
	}
	message := "union member where the checker narrowed it away: a call since the narrowing put it back"
	b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}})
	result := b.finish("narrowed_boxed_scalar_field", ir.Narrow{Value: held, To: narrowed})
	l.result.Functions[b.function].CheckedUnionNarrow = true
	return result
}
