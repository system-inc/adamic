package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// assignmentValue keeps the assignment visible to write analysis and reads its
// stored value immediately. A local has no setter that could replace that value.
func (lowering *lowering) assignmentValue(node *ast.Node) (ir.Expression, error) {
	if lowering.uninitializedInitializer(node.AsBinaryExpression().Right) {
		return nil, lowering.notYet(node, "using the result of a placeholder reset")
	}
	target := ast.SkipParentheses(node.AsBinaryExpression().Left)
	if !ast.IsIdentifier(target) {
		return nil, lowering.notYet(target, "an assignment value to a member")
	}
	body, assignmentError := lowering.assignment(node)
	if assignmentError != nil {
		return nil, assignmentError
	}
	local, known := lowering.local(target)
	if !known {
		return nil, lowering.notYet(target, "an assignment value to this binding")
	}
	lowering.result.Locals[local].ExpressionAssigned = true
	result, typeOfError := lowering.typeOf(node)
	if typeOfError != nil {
		return nil, typeOfError
	}
	var value ir.Expression = ir.Read{
		Local:   local,
		Of:      lowering.result.Locals[local].Type,
		Checked: lowering.checked(local),
	}
	if origin := lowering.result.Locals[local].Placeholder; origin != "" && !lowering.placeholderAllowsUnset(node) {
		value = ir.PlaceholderUse{
			Value:  value,
			Origin: origin,
			Use:    "assignment result",
			Path:   sourceExpression(node),
			Where:  lowering.program.Where(node),
			Of:     result,
		}
	}
	if value.Type() == ir.Union && result != ir.Union {
		value = ir.Narrow{Value: value, To: result}
	} else {
		value = fit(value, result)
	}
	if value.Type() != result {
		return nil, lowering.notYet(node, "an assignment result with different storage")
	}
	return ir.Effects{Body: body, Result: value}, nil
}
