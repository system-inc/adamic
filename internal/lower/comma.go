package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (lowering *lowering) comma(node *ast.Node) (ir.Expression, error) {
	binary := node.AsBinaryExpression()
	left, discardedValueError := lowering.discardedValue(binary.Left)
	if discardedValueError != nil {
		return nil, discardedValueError
	}
	right, expressionError := lowering.expression(binary.Right)
	if expressionError != nil {
		return nil, expressionError
	}
	return ir.Comma{Left: left, Right: right}, nil
}

// discardedValue accepts the effects already lowered as statements too: writes, increments,
// console calls and void calls. The discarded value never needs storage.
func (lowering *lowering) discardedValue(node *ast.Node) (ir.Expression, error) {
	value, expressionError := lowering.expression(node)
	if expressionError == nil {
		return value, nil
	}
	statements, statementError := lowering.expressionStatement(node)
	if statementError != nil {
		return nil, expressionError
	}
	walk(statements, func(node any) bool {
		if assign, ok := node.(ir.Assign); ok {
			lowering.result.Locals[assign.Local].ExpressionAssigned = true
		}
		if declare, ok := node.(ir.Declare); ok {
			lowering.result.Locals[declare.Local].ExpressionAssigned = true
		}
		return true
	})
	return ir.Effects{Body: statements, Result: ir.Undefined{}}, nil
}
