package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Nullable strings need the existing union's distinct null sentinel. A raw string
// pointer cannot distinguish null from undefined or safely stand in for a string.
func nullableStringUnion(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	text, null := false, false
	for _, member := range proven.Types() {
		switch {
		case member.Flags()&checker.TypeFlagsStringLike != 0:
			text = true
		case member.Flags()&checker.TypeFlagsNull != 0:
			null = true
		case member.Flags()&checker.TypeFlagsUndefined != 0:
		default:
			return false
		}
	}
	return text && null
}

// Read the stored tag once before narrowing; alias writes can invalidate the
// checker's field fact just as captured writes can invalidate a local fact.
func (l *lowering) checkedNullableString(node *ast.Node, read ir.Expression) ir.Expression {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		node, parent = parent, parent.Parent
	}
	if comparedWithUndefined(node) || (parent != nil && parent.Kind == ast.KindTypeOfExpression) {
		return read
	}
	if parent != nil && parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindQuestionQuestionToken && binary.Left == node {
			return read
		}
	}
	b := l.libraryArrayBuilder([]ir.Expression{read})
	held := b.read(b.parameters[0])
	matches := ir.Expression(ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: held}, Right: ir.StringConstant{Index: l.constant("string")}})
	if l.includesUndefined(l.checker.GetTypeAtLocation(node)) {
		matches = ir.Binary{Operator: ir.Or, Left: matches, Right: ir.IsUndefined{Value: held}}
	}
	message := "union member where the checker narrowed it away: a call since the narrowing put it back"
	b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}})
	result := b.finish("narrowed_union_member", ir.Narrow{Value: held, To: ir.String})
	l.result.Functions[b.function].CheckedUnionNarrow = true
	return result
}
