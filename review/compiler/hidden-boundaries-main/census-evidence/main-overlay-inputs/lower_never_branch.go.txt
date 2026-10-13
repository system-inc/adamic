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
	effect, err := l.discardedValue(node)
	if err != nil {
		return nil, err
	}
	var unreachable ir.Expression = fit(ir.Undefined{}, of)
	switch of {
	case ir.Number:
		unreachable = ir.NumberConstant{}
	case ir.Boolean:
		unreachable = ir.BooleanConstant{}
	}
	return ir.Comma{Left: effect, Right: unreachable}, nil
}

func (l *lowering) logicalNever(node *ast.Node) (ir.Expression, error) {
	binary := node.AsBinaryExpression()
	left, err := l.expression(binary.Left)
	if err != nil {
		return nil, err
	}
	of, err := l.typeOf(node)
	if err != nil {
		return nil, err
	}
	right, err := l.neverValue(binary.Right, of)
	if err != nil {
		return nil, err
	}
	return ir.Logical{Left: left, Right: right, Of: of, KeepTruthy: binary.OperatorToken.Kind == ast.KindBarBarToken}, nil
}

func (l *lowering) conditionalNever(node *ast.Node, condition ir.Expression) (ir.Expression, error) {
	conditional := node.AsConditionalExpression()
	of, err := l.typeOf(node)
	if err != nil {
		return nil, err
	}
	branch := func(arm *ast.Node) (ir.Expression, error) {
		if l.isNever(arm) {
			return l.neverValue(arm, of)
		}
		value, err := l.expression(arm)
		if err != nil {
			return nil, err
		}
		return fit(value, of), nil
	}
	whenTrue, err := branch(conditional.WhenTrue)
	if err != nil {
		return nil, err
	}
	whenNot, err := branch(conditional.WhenFalse)
	if err != nil {
		return nil, err
	}
	return ir.Conditional{Condition: condition, WhenTrue: whenTrue, WhenNot: whenNot, Of: of}, nil
}
