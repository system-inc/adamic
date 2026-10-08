package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// mixedLogical selects one operand, evaluating the left exactly once and the right
// only on its branch. Coalesce already carries both result aliases through the
// ownership, freshness and flow analyses; Logical changes only the branch test.
func (l *lowering) mixedLogical(node *ast.Node, operator ast.Kind, left, right ir.Expression) (ir.Expression, bool, error) {
	if operator != ast.KindAmpersandAmpersandToken && operator != ast.KindBarBarToken {
		return nil, false, nil
	}
	of, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	common := left.Type()
	if common != right.Type() {
		common = ir.Union
	}
	if common == ir.Union && (l.includesNull(l.checker.GetTypeAtLocation(node.AsBinaryExpression().Left)) || l.includesNull(l.checker.GetTypeAtLocation(node.AsBinaryExpression().Right))) {
		return nil, true, l.notYet(node, "logical selection requiring a boxed null distinct from undefined")
	}
	left, right = fit(left, common), fit(right, common)
	if left.Type() != common || right.Type() != common {
		return nil, true, l.notYet(node, "logical operands requiring another representation")
	}
	logical := ir.And
	if operator == ast.KindBarBarToken {
		logical = ir.Or
	}
	value := ir.Expression(ir.Coalesce{Value: left, Fallback: right, Of: common, Logical: logical})
	if common == of {
		return value, true, nil
	}
	if common == ir.Union {
		return ir.Narrow{Value: value, To: of}, true, nil
	}
	if common.IsMaybe() && common.Present() == of {
		return fit(value, of), true, nil
	}
	return nil, true, l.notYet(node, "logical selection whose checked result requires another representation")
}
