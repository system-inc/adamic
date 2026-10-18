package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// assignmentValue keeps the assignment visible to write analysis and reads its
// stored value immediately. A local has no setter that could replace that value.
func (l *lowering) assignmentValue(node *ast.Node) (ir.Expression, error) {
	target := ast.SkipParentheses(node.AsBinaryExpression().Left)
	if !ast.IsIdentifier(target) {
		return nil, l.notYet(target, "an assignment value to a member")
	}
	body, err := l.assignment(node)
	if err != nil {
		return nil, err
	}
	local, known := l.local(target)
	if !known {
		return nil, l.notYet(target, "an assignment value to this binding")
	}
	l.result.Locals[local].ExpressionAssigned = true
	result, err := l.typeOf(node)
	if err != nil {
		return nil, err
	}
	var value ir.Expression = ir.Read{Local: local, Of: l.result.Locals[local].Type, Checked: l.checked(local)}
	if value.Type() == ir.Union && result != ir.Union {
		value = ir.Narrow{Value: value, To: result}
	} else {
		value = fit(value, result)
	}
	if value.Type() != result {
		return nil, l.notYet(node, "an assignment result with different storage")
	}
	return ir.Effects{Body: body, Result: value}, nil
}
