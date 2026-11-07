package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// A field keeps its declared union even when flow analysis narrows a read to a
// primitive. Evaluate that stored value once and verify its tag before unboxing.
func (lowering *lowering) checkedBoxedScalarField(node *ast.Node, read ir.Expression, narrowed ir.Type) ir.Expression {
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
	} else if narrowed == ir.String {
		name = "string"
	}
	builder := lowering.libraryArrayBuilder([]ir.Expression{read})
	held := builder.read(builder.parameters[0])
	matches := ir.Expression(ir.Binary{
		Operator: ir.Equal,
		Left:     ir.TypeOf{Value: held},
		Right:    ir.StringConstant{Index: lowering.constant(name)},
	})
	if narrowed.IsMaybe() {
		matches = ir.Binary{Operator: ir.Or, Left: matches, Right: ir.IsUndefined{Value: held}}
	}
	message := "union member where the checker narrowed it away: a call since the narrowing put it back"
	builder.body = append(builder.body, ir.If{
		Condition: ir.Unary{Operator: ir.Not, Operand: matches},
		Then:      []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: lowering.constant(message)}}},
	})
	result := builder.finish("narrowed_boxed_scalar_field", ir.Narrow{Value: held, To: narrowed})
	lowering.result.Functions[builder.function].CheckedUnionNarrow = true
	return result
}

// Strict equality compares the actual tags; it never consumes a narrowed slot.
func unionEqualityObservation(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return false
	}
	operator := parent.AsBinaryExpression().OperatorToken.Kind
	return operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindExclamationEqualsEqualsToken
}
