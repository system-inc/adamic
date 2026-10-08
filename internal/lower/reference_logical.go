package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Only these representations prove truthiness from presence. A string may be
// empty, and a boxed union may contain a falsy primitive.
func (l *lowering) truthyReference(proven *checker.Type) bool {
	proven = l.concrete(proven)
	of, known := l.representation(l.checker.GetNonNullableType(proven))
	return known && (of == ir.Array || of == ir.Object || of == ir.Map || of == ir.Closure)
}

func (l *lowering) referenceLogical(node *ast.Node, left, right ir.Expression) (ir.Expression, bool, error) {
	binary := node.AsBinaryExpression()
	operator := binary.OperatorToken.Kind
	if (operator != ast.KindAmpersandAmpersandToken && operator != ast.KindBarBarToken) || !l.truthyReference(l.checker.GetTypeAtLocation(binary.Left)) {
		return nil, false, nil
	}
	of, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	return ir.Coalesce{Value: left, Fallback: fit(right, of), Of: of, ReferenceAnd: operator == ast.KindAmpersandAmpersandToken}, true, nil
}

// The left of a reference && supplies only its absent members to the result.
// Its populated elements are not viewed as the right operand's elements.
func (l *lowering) referenceAndTest(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := parent.AsBinaryExpression()
	return binary.Left == node && binary.OperatorToken.Kind == ast.KindAmpersandAmpersandToken && l.truthyReference(l.checker.GetTypeAtLocation(node))
}

// A slot lookup can give undefined even when its populated slots hold null.
// The pointer alone cannot distinguish them. typeof retains slot presence in
// the backend, and ??/|| consume both absences without returning either one.
func (l *lowering) nullableReferenceObservation(node *ast.Node) error {
	proven := l.concrete(l.checker.GetTypeAtLocation(node))
	if !l.includesNull(proven) || !l.includesUndefined(proven) || !l.truthyReference(proven) {
		return nil
	}
	at := node
	for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
		at = at.Parent
	}
	if parent := at.Parent; parent != nil {
		if parent.Kind == ast.KindTypeOfExpression {
			return nil
		}
		if parent.Kind == ast.KindBinaryExpression {
			binary := parent.AsBinaryExpression()
			if binary.Left == at && (binary.OperatorToken.Kind == ast.KindQuestionQuestionToken || binary.OperatorToken.Kind == ast.KindBarBarToken) {
				return nil
			}
		}
	}
	return l.notYet(node, "a reference holding both null and undefined without a distinct absence tag; consume it with ?? or ||, or inspect it with typeof")
}
