package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// A scalar singleton assignment returned directly can use statement IR: evaluate the
// right side once, store it, then read the same scalar before returning. No setter or
// reference conversion can make the stored value differ from the expression result.
func (lowering *lowering) namespaceReturnAssignment(expression *ast.Node) ([]ir.Statement, bool, error) {
	expression = ast.SkipParentheses(expression)
	if expression.Kind != ast.KindBinaryExpression ||
		expression.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken {
		return nil, false, nil
	}
	target := ast.SkipParentheses(expression.AsBinaryExpression().Left)
	local, found := lowering.local(target)
	if !found ||
		!lowering.namespaceMember(target) ||
		(!ast.IsIdentifier(target) && target.Kind != ast.KindPropertyAccessExpression) {
		return nil, false, nil
	}
	of := lowering.result.Locals[local].Type
	if of != ir.Number && of != ir.Boolean {
		return nil, true, lowering.notYet(
			expression,
			"a returned namespace assignment with reference or optional storage; only scalar singleton assignment is proven",
		)
	}
	statements, assignmentError := lowering.assignment(expression)
	if assignmentError != nil {
		return nil, true, assignmentError
	}
	value, localReadError := lowering.localRead(target, local)
	if localReadError != nil {
		return nil, true, localReadError
	}
	return append(statements, ir.Return{Value: fit(value, lowering.function.Returns)}), true, nil
}
