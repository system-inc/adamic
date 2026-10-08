package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// An intrinsic map owns its new outer slots. Its elements also have no other
// owner when every callback exit returns a freshly allocated value directly.
// This proves a view at this site, without changing any checker type. Local
// aliases, stores and unrecognized control flow conservatively lose the proof.
func (l *lowering) freshMapReturns(node *ast.Node) []*ast.Node {
	if node.Kind != ast.KindCallExpression || !l.freshValue(node) {
		return nil
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "map" || len(call.Arguments.Nodes) != 1 {
		return nil
	}
	callback := ast.SkipParentheses(call.Arguments.Nodes[0])
	if callback.Kind != ast.KindArrowFunction && callback.Kind != ast.KindFunctionExpression {
		return nil
	}
	returns, terminal, proven := l.freshCallbackBody(callback.Body())
	if !terminal || !proven || len(returns) == 0 {
		return nil
	}
	return returns
}

func (l *lowering) freshCallbackValue(node *ast.Node) ([]*ast.Node, bool) {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindConditionalExpression {
		branch := node.AsConditionalExpression()
		left, first := l.freshCallbackValue(branch.WhenTrue)
		right, second := l.freshCallbackValue(branch.WhenFalse)
		return append(left, right...), first && second
	}
	if !l.checker.IsArrayType(l.concrete(l.checker.GetTypeAtLocation(node))) {
		return nil, false
	}
	return []*ast.Node{node}, node.Kind == ast.KindArrayLiteralExpression || l.freshValue(node)
}

// terminal records whether all paths leave. A nonterminal if may be followed
// by another return; every earlier exit must still prove freshness.
func (l *lowering) freshCallbackBody(node *ast.Node) ([]*ast.Node, bool, bool) {
	if node == nil {
		return nil, false, true
	}
	switch node.Kind {
	case ast.KindReturnStatement:
		value := node.AsReturnStatement().Expression
		if value == nil {
			return nil, true, false
		}
		returns, proven := l.freshCallbackValue(value)
		return returns, true, proven
	case ast.KindBlock:
		var returns []*ast.Node
		for _, statement := range node.AsBlock().Statements.Nodes {
			paths, terminal, proven := l.freshCallbackBody(statement)
			if !proven {
				return nil, false, false
			}
			returns = append(returns, paths...)
			if terminal {
				return returns, true, true
			}
		}
		return returns, false, true
	case ast.KindIfStatement:
		branch := node.AsIfStatement()
		left, firstExit, first := l.freshCallbackBody(branch.ThenStatement)
		right, secondExit, second := l.freshCallbackBody(branch.ElseStatement)
		return append(left, right...), firstExit && secondExit, first && second
	default:
		if ast.IsExpression(node) {
			returns, proven := l.freshCallbackValue(node)
			return returns, true, proven
		}
		return nil, false, false
	}
}
