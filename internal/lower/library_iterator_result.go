package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// IteratorYieldResult permits an omitted done flag. Its boolean test is false on
// both omission and explicit false; preserve the optional value everywhere else.
func libraryIteratorDoneTruth(value ir.Expression) ir.Expression {
	if call, ok := value.(ir.RegExpCall); ok && call.Method == "iteratorDone" && call.Returns == ir.MaybeBoolean {
		return ir.Coalesce{Value: value, Fallback: ir.BooleanConstant{Value: false}, Of: ir.Boolean}
	}
	return value
}

// The runtime's internal next closure represents an inherited collection method.
// Only a bound call may read that slot; detached reads and copies stay refused.
func (l *lowering) libraryIteratorNextClosure(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "next" {
		return nil, false, nil
	}
	receiver := callee.AsPropertyAccessExpression().Expression
	if !l.libraryIteratorType(l.checker.GetTypeAtLocation(receiver)) {
		return nil, false, nil
	}
	if len(call.Arguments.Nodes) != 0 {
		return nil, true, l.notYet(node, "collection iterator next with arguments")
	}
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	return ir.Property{Object: value, Name: "next", Of: ir.Closure}, true, nil
}
