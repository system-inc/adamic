package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// clockNullableString admits only strings and null. The missing string pointer means null;
// typeOfNull preserves that interpretation, while empty strings remain allocated references.
func (l *lowering) clockNullableString(proven *checker.Type) bool {
	if !l.includesNull(proven) {
		return false
	}
	for _, member := range proven.Types() {
		if member.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNull) == 0 {
			return false
		}
	}
	return true
}

// A call can put null back into a narrowed string slot. Restore its declared meaning
// for null and undefined observations without changing the shared binary helper.
func (l *lowering) clockStringNullObservation(node *ast.Node, value ir.Expression) ir.Expression {
	observed := value
	unary, negated := value.(ir.Unary)
	if negated && unary.Operator == ir.Not {
		observed = unary.Operand
	}
	binary := node.AsBinaryExpression()
	nullComparison := ast.SkipParentheses(binary.Left).Kind == ast.KindNullKeyword || ast.SkipParentheses(binary.Right).Kind == ast.KindNullKeyword
	var held ir.Expression
	switch test := observed.(type) {
	case ir.IsNull:
		if !test.AlwaysFalse {
			return value
		}
		held = test.Value
	case ir.IsUndefined:
		held = test.Value
	default:
		return value
	}
	operand := ast.SkipParentheses(binary.Left)
	if operand.Kind == ast.KindNullKeyword || (operand.Kind == ast.KindIdentifier && operand.Text() == "undefined") {
		operand = ast.SkipParentheses(binary.Right)
	}
	at := operand
	if operand.Kind == ast.KindPropertyAccessExpression {
		at = operand.Name()
	}
	if operand.Kind == ast.KindIdentifier || operand.Kind == ast.KindPropertyAccessExpression {
		if symbol := l.checker.GetSymbolAtLocation(at); symbol != nil {
			declared := l.concrete(l.checker.GetTypeOfSymbol(symbol))
			if declared.Flags()&checker.TypeFlagsUnion != 0 && l.clockNullableString(declared) {
				test := ir.IsNull{Value: held, AlwaysFalse: !nullComparison}
				if negated {
					unary.Operand = test
					return unary
				}
				return test
			}
		}
	}
	return value
}
