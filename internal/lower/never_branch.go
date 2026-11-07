package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) isNever(node *ast.Node) bool {
	return l.concrete(l.checker.GetTypeAtLocation(node)).Flags()&checker.TypeFlagsNever != 0
}

// A never arm has effects but no value. The typed placeholder exists only to
// satisfy the backend's branch ABI; the proven non-returning arm precedes it.
func (l *lowering) neverValue(node *ast.Node, of ir.Type) (ir.Expression, error) {
	effect, discardedValueError := l.discardedValue(node)
	if discardedValueError != nil {
		return nil, discardedValueError
	}
	return ir.Comma{Left: effect, Right: unreachablePlaceholder(of)}, nil
}

// unreachablePlaceholder is the typed value held where no value is ever made, after what never returns:
// a number's 0, a boolean's false, and undefined as of holds it (a string's null pointer).
func unreachablePlaceholder(of ir.Type) ir.Expression {
	switch of {
	case ir.Number:
		return ir.NumberConstant{}
	case ir.Boolean:
		return ir.BooleanConstant{}
	}
	return fit(ir.Undefined{}, of)
}

func (l *lowering) logicalNever(node *ast.Node) (ir.Expression, error) {
	binary := node.AsBinaryExpression()
	left, expressionError := l.expression(binary.Left)
	if expressionError != nil {
		return nil, expressionError
	}
	of, typeOfError := l.typeOf(node)
	if typeOfError != nil {
		return nil, typeOfError
	}
	right, neverValueError := l.neverValue(binary.Right, of)
	if neverValueError != nil {
		return nil, neverValueError
	}
	return ir.Logical{
		Left:       left,
		Right:      right,
		Of:         of,
		KeepTruthy: binary.OperatorToken.Kind == ast.KindBarBarToken,
	}, nil
}

func (l *lowering) conditionalNever(node *ast.Node, condition ir.Expression) (ir.Expression, error) {
	conditional := node.AsConditionalExpression()
	of, typeOfError := l.typeOf(node)
	if typeOfError != nil {
		return nil, typeOfError
	}
	branch := func(arm *ast.Node) (ir.Expression, error) {
		if l.isNever(arm) {
			return l.neverValue(arm, of)
		}
		value, expressionError := l.expression(arm)
		if expressionError != nil {
			return nil, expressionError
		}
		return fit(value, of), nil
	}
	whenTrue, branchError := branch(conditional.WhenTrue)
	if branchError != nil {
		return nil, branchError
	}
	whenNot, branchError := branch(conditional.WhenFalse)
	if branchError != nil {
		return nil, branchError
	}
	return ir.Conditional{Condition: condition, WhenTrue: whenTrue, WhenNot: whenNot, Of: of}, nil
}
