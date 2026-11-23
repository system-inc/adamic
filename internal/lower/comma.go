package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) comma(node *ast.Node) (ir.Expression, error) {
	binary := node.AsBinaryExpression()
	left, err := l.discardedValue(binary.Left)
	if err != nil {
		return nil, err
	}
	right, err := l.expression(binary.Right)
	if err != nil {
		return nil, err
	}
	return ir.Comma{Left: left, Right: right}, nil
}

// discardedValue accepts the effects already lowered as statements too: writes, increments,
// console calls and void calls. The discarded value never needs storage.
func (l *lowering) discardedValue(node *ast.Node) (ir.Expression, error) {
	l.inspectorRecordBuiltUse(node)
	// A postfix update whose value is discarded is the increment alone: its value would be the old one held in
	// postfix's helper, which nothing reads, so it tries the statement first and keeps the cheaper lowering
	// (#f3n0gyp). Only where the statement can't lower does it take the expression's road and its stop.
	if ast.SkipParentheses(node).Kind == ast.KindPostfixUnaryExpression {
		if statements, err := l.expressionStatement(node); err == nil {
			return l.discardedEffects(statements), nil
		}
	}
	value, err := l.expression(node)
	if err == nil {
		return value, nil
	}
	statements, statementErr := l.expressionStatement(node)
	if statementErr != nil {
		return nil, err
	}
	return l.discardedEffects(statements), nil
}

// discardedEffects is statements as a value nothing reads: every local they write is marked as written by an
// expression, and the value is undefined.
func (l *lowering) discardedEffects(statements []ir.Statement) ir.Expression {
	walk(statements, func(node any) bool {
		if assign, ok := node.(ir.Assign); ok {
			l.result.Locals[assign.Local].ExpressionAssigned = true
		}
		if declare, ok := node.(ir.Declare); ok {
			l.result.Locals[declare.Local].ExpressionAssigned = true
		}
		return true
	})
	return ir.Effects{Body: statements, Result: ir.Undefined{}}
}
