package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A literal boolean chooses a callable initializer without evaluating the other
// branch. In particular, a dead Date.now value needs no native clock intrinsic.
func (l *lowering) literalCallableBranch(node *ast.Node) *ast.Node {
	if node.Kind != ast.KindConditionalExpression || len(l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node), checker.SignatureKindCall)) == 0 {
		return nil
	}
	expression := node.AsConditionalExpression()
	switch ast.SkipParentheses(expression.Condition).Kind {
	case ast.KindTrueKeyword:
		return expression.WhenTrue
	case ast.KindFalseKeyword:
		return expression.WhenFalse
	}
	return nil
}
