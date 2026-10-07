package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// A scalar singleton assignment returned directly can use statement IR: evaluate the
// right side once, store it, then read the same scalar before returning. No setter or
// reference conversion can make the stored value differ from the expression result.
func (l *lowering) namespaceReturnAssignment(expression *ast.Node) ([]ir.Statement, bool, error) {
	expression = ast.SkipParentheses(expression)
	if expression.Kind != ast.KindBinaryExpression || expression.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken {
		return nil, false, nil
	}
	target := ast.SkipParentheses(expression.AsBinaryExpression().Left)
	local, found := l.local(target)
	if !found || !l.namespaceMember(target) || (!ast.IsIdentifier(target) && target.Kind != ast.KindPropertyAccessExpression) {
		return nil, false, nil
	}
	of := l.result.Locals[local].Type
	if of != ir.Number && of != ir.Boolean {
		return nil, true, l.notYet(expression, "a returned namespace assignment with reference or optional storage; only scalar singleton assignment is proven")
	}
	statements, err := l.assignment(expression)
	if err != nil {
		return nil, true, err
	}
	return append(statements, ir.Return{Value: fit(l.localRead(target, local), l.function.Returns)}), true, nil
}
