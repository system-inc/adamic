package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Admit one concrete object representation plus both nullish tags. Collections,
// closures, weak handles and erased structural primitives need their own proofs.
func (l *lowering) nullableObjectUnion(proven *checker.Type) bool {
	proven = l.concrete(proven)
	if proven.Flags()&checker.TypeFlagsUnion == 0 || !l.includesNull(proven) || !l.includesUndefined(proven) {
		return false
	}
	present := l.checker.GetNonNullableType(proven)
	held, known := l.representation(present)
	return known && held == ir.Object && present.Flags()&checker.TypeFlagsObject != 0
}

func unionObservation(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		node, parent = parent, parent.Parent
	}
	if comparedWithUndefined(node) || parent != nil && parent.Kind == ast.KindTypeOfExpression {
		return true
	}
	if parent != nil && parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
			return true
		case ast.KindQuestionQuestionToken:
			return binary.Left == node
		}
	}
	return false
}

func (l *lowering) checkedNullableObject(node *ast.Node, read ir.Expression) ir.Expression {
	observed := l.concrete(l.checker.GetTypeAtLocation(node))
	narrowed, known := l.representation(observed)
	contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	if contextual != nil && l.nullableObjectUnion(contextual) {
		return read
	}
	if !known || narrowed == ir.Union || unionObservation(node) {
		return read
	}
	b := l.libraryArrayBuilder([]ir.Expression{read})
	held := b.read(b.parameters[0])
	null := ir.Expression(ir.IsNull{Value: held})
	matches := ir.Expression(ir.Binary{Operator: ir.And,
		Left:  ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: held}, Right: ir.StringConstant{Index: l.constant("object")}},
		Right: ir.Unary{Operator: ir.Not, Operand: null}})
	if l.includesNull(observed) {
		matches = ir.Binary{Operator: ir.Or, Left: matches, Right: null}
	}
	if l.includesUndefined(observed) {
		matches = ir.Binary{Operator: ir.Or, Left: matches, Right: ir.IsUndefined{Value: held}}
	}
	message := "union member where the checker narrowed it away: a call since the narrowing put it back"
	b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}}}})
	result := ir.Expression(ir.Narrow{Value: held, To: ir.Object})
	if l.includesNull(observed) {
		result = ir.Conditional{Condition: null, WhenTrue: ir.Null{Of: ir.Object}, WhenNot: result, Of: ir.Object}
	}
	call := b.finish("narrowed_nullable_object", result)
	l.result.Functions[b.function].CheckedUnionNarrow = true
	return call
}

// A legacy object-or-null pointer uses NULL for null. Normalize it when a
// contextual tagged slot receives it, evaluating the source only once.
func (l *lowering) nullableObjectBoundary(node *ast.Node, value ir.Expression) ir.Expression {
	if value == nil || value.Type() != ir.Object {
		return value
	}
	own := l.concrete(l.checker.GetTypeAtLocation(node))
	if !l.includesNull(own) || l.includesUndefined(own) {
		return value
	}
	contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	if contextual == nil {
		return value
	}
	target, known := l.representation(contextual)
	if !known || target != ir.Union {
		return value
	}
	if _, literal := value.(ir.Null); literal {
		return value
	}
	b := l.libraryArrayBuilder([]ir.Expression{value})
	held := b.read(b.parameters[0])
	return b.finish("nullable_object_to_union", ir.Conditional{Condition: ir.IsNull{Value: held}, WhenTrue: ir.Box{Value: ir.Null{}}, WhenNot: ir.Box{Value: held}, Of: ir.Union})
}

// Optional access skips both nullish tags; only an actual object reaches the read.
func (l *lowering) optionalNullableObject(read ir.Expression) ir.Expression {
	b := l.libraryArrayBuilder([]ir.Expression{read})
	held := b.read(b.parameters[0])
	absent := ir.Binary{Operator: ir.Or, Left: ir.IsNull{Value: held}, Right: ir.IsUndefined{Value: held}}
	return b.finish("optional_nullable_object", ir.Conditional{Condition: absent, WhenTrue: ir.Undefined{Of: ir.Object}, WhenNot: ir.Narrow{Value: held, To: ir.Object}, Of: ir.Object})
}
