package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

// A fresh, private integer induction variable and a current length comparison establish
// both bounds. Effects before the read invalidate that iteration's comparison. Effects
// after it are allowed: the next iteration compares the current length again.
func (l *lowering) collectionRangeProven(node *ast.Node) bool {
	lookup := load.CollectionLookup(l.checker, node)
	if lookup == nil || lookup.Kind != ast.KindElementAccessExpression {
		return false
	}
	access := lookup.AsElementAccessExpression()
	elements := l.typeArguments(l.checker.GetTypeAtLocation(access.Expression))
	if len(elements) != 1 || l.includesUndefined(elements[0]) || l.includesNull(elements[0]) {
		return false
	}
	index := ast.SkipParentheses(access.ArgumentExpression)
	if !ast.IsIdentifier(index) {
		return false
	}
	for parent := lookup.Parent; parent != nil; parent = parent.Parent {
		if ast.IsFunctionLike(parent) {
			break
		}
		if parent.Kind != ast.KindForStatement {
			continue
		}
		loop := parent.AsForStatement()
		if loop.Initializer == nil || loop.Initializer.Kind != ast.KindVariableDeclarationList || loop.Initializer.Flags&ast.NodeFlagsLet == 0 || len(loop.Initializer.AsVariableDeclarationList().Declarations.Nodes) != 1 || loop.Condition == nil || loop.Incrementor == nil {
			continue
		}
		declaration := loop.Initializer.AsVariableDeclarationList().Declarations.Nodes[0].AsVariableDeclaration()
		if declaration.Initializer == nil || declaration.Initializer.Kind != ast.KindNumericLiteral || declaration.Initializer.Text() != "0" || !l.collectionSameOperand(declaration.Name(), index) {
			continue
		}
		condition := ast.SkipParentheses(loop.Condition)
		if condition.Kind != ast.KindBinaryExpression {
			continue
		}
		comparison := condition.AsBinaryExpression()
		length := ast.SkipParentheses(comparison.Right)
		if comparison.OperatorToken.Kind != ast.KindLessThanToken || !l.collectionSameOperand(comparison.Left, index) || length.Kind != ast.KindPropertyAccessExpression || length.Name().Text() != "length" || !l.collectionSameOperand(length.AsPropertyAccessExpression().Expression, access.Expression) {
			continue
		}
		increment := ast.SkipParentheses(loop.Incrementor)
		if increment.Kind != ast.KindPostfixUnaryExpression || increment.AsPostfixUnaryExpression().Operator != ast.KindPlusPlusToken || !l.collectionSameOperand(increment.AsPostfixUnaryExpression().Operand, index) {
			continue
		}
		if !collectionPrefixUnchanged(loop.Statement, loop.Condition.End(), lookup.Pos()) {
			continue
		}
		private := true
		var inspect ast.Visitor
		inspect = func(n *ast.Node) bool {
			if ast.IsFunctionLike(n) {
				var captured ast.Visitor
				captured = func(child *ast.Node) bool {
					if ast.IsIdentifier(child) && l.collectionSameOperand(child, index) {
						private = false
					}
					child.ForEachChild(captured)
					return false
				}
				n.ForEachChild(captured)
				return false
			}
			if n.Kind == ast.KindBinaryExpression && ast.IsAssignmentOperator(n.AsBinaryExpression().OperatorToken.Kind) && l.collectionSameOperand(n.AsBinaryExpression().Left, index) {
				private = false
			}
			if n.Kind == ast.KindPostfixUnaryExpression && l.collectionSameOperand(n.AsPostfixUnaryExpression().Operand, index) {
				private = false
			}
			if n.Kind == ast.KindPrefixUnaryExpression {
				unary := n.AsPrefixUnaryExpression()
				if (unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken) && l.collectionSameOperand(unary.Operand, index) {
					private = false
				}
			}
			n.ForEachChild(inspect)
			return false
		}
		loop.Statement.ForEachChild(inspect)
		if private {
			return true
		}
	}
	return false
}
