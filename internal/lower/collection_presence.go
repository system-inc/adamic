package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Same receiver and key mean the same bindings or the same primitive literal. No property
// getters, computed receivers, conversions or calls can run while reproducing these operands.
func (l *lowering) collectionSameOperand(left, right *ast.Node) bool {
	left, right = ast.SkipParentheses(left), ast.SkipParentheses(right)
	if ast.IsIdentifier(left) && ast.IsIdentifier(right) {
		return l.symbol(left) != nil && l.symbol(left) == l.symbol(right)
	}
	if left.Kind != right.Kind {
		return false
	}
	switch left.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral:
		return left.Text() == right.Text()
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return true
	}
	return false
}
func (l *lowering) collectionHasMatches(condition, lookup *ast.Node) bool {
	condition = ast.SkipParentheses(condition)
	if condition.Kind != ast.KindCallExpression {
		return false
	}
	has := condition.AsCallExpression()
	get := lookup.AsCallExpression()
	callee := ast.SkipParentheses(has.Expression)
	receiver := ast.SkipParentheses(get.Expression)
	return callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "has" && callee.AsPropertyAccessExpression().QuestionDotToken == nil && len(has.Arguments.Nodes) == 1 && l.collectionSameOperand(callee.AsPropertyAccessExpression().Expression, receiver.AsPropertyAccessExpression().Expression) && l.collectionSameOperand(has.Arguments.Nodes[0], get.Arguments.Nodes[0])
}

// An unknown call can write through an alias. Conservatively reject all intervening calls,
// constructors, assignments and updates, even ones that might affect an unrelated map.
func collectionPrefixUnchanged(root *ast.Node, start, end int) bool {
	unchanged := true
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if !unchanged || node.End() <= start || node.Pos() >= end {
			return false
		}
		if ast.IsFunctionLike(node) {
			return false
		}
		switch node.Kind {
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression, ast.KindNewExpression, ast.KindPostfixUnaryExpression, ast.KindDeleteExpression, ast.KindAwaitExpression, ast.KindYieldExpression:
			if node.Pos() >= start {
				unchanged = false
				return true
			}
		case ast.KindPrefixUnaryExpression:
			op := node.AsPrefixUnaryExpression().Operator
			if op == ast.KindPlusPlusToken || op == ast.KindMinusMinusToken {
				unchanged = false
				return true
			}
		case ast.KindBinaryExpression:
			if ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
				unchanged = false
				return true
			}
		}
		node.ForEachChild(visit)
		return false
	}
	root.ForEachChild(visit)
	return unchanged
}
func (l *lowering) collectionPresenceProven(node *ast.Node) bool {
	lookup := load.CollectionLookup(l.checker, node)
	if lookup == nil {
		return false
	}
	receiver := ast.SkipParentheses(lookup.AsCallExpression().Expression).AsPropertyAccessExpression().Expression
	arguments := l.typeArguments(l.checker.GetTypeAtLocation(receiver))
	if len(arguments) != 2 || l.includesUndefined(arguments[1]) || l.includesNull(arguments[1]) {
		return false
	}
	child := lookup
	for parent := lookup.Parent; parent != nil; child, parent = parent, parent.Parent {
		if ast.IsFunctionLike(parent) {
			break
		}
		if parent.Kind == ast.KindBlock {
			for _, statement := range parent.AsBlock().Statements.Nodes {
				if statement.End() > lookup.Pos() || statement.Kind != ast.KindIfStatement {
					continue
				}
				guard := statement.AsIfStatement()
				condition := ast.SkipParentheses(guard.Expression)
				if guard.ElseStatement != nil || condition.Kind != ast.KindPrefixUnaryExpression || condition.AsPrefixUnaryExpression().Operator != ast.KindExclamationToken {
					continue
				}
				terminal := guard.ThenStatement
				if terminal.Kind == ast.KindBlock && len(terminal.AsBlock().Statements.Nodes) == 1 {
					terminal = terminal.AsBlock().Statements.Nodes[0]
				}
				if terminal.Kind != ast.KindReturnStatement && terminal.Kind != ast.KindThrowStatement {
					continue
				}
				if l.collectionHasMatches(condition.AsPrefixUnaryExpression().Operand, lookup) && collectionPrefixUnchanged(parent, statement.End(), lookup.Pos()) {
					return true
				}
			}
		}
		if parent.Kind == ast.KindIfStatement && child == parent.AsIfStatement().ThenStatement {
			branch := parent.AsIfStatement()
			if l.collectionHasMatches(branch.Expression, lookup) && collectionPrefixUnchanged(branch.ThenStatement, branch.Expression.End(), lookup.Pos()) {
				return true
			}
		}
	}
	return false
}
func (l *lowering) collectionRequiredRead(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	if !l.program.CheckedCollectionRead(node) {
		return value, nil
	}
	of, known := l.representation(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node)))
	if !known || (!value.Type().IsMaybe() && !value.Type().IsReference()) || (value.Type() != of && value.Type().Present() != of) {
		return nil, l.notYet(node, "a checked collection lookup without a compatible presence representation")
	}
	return l.collectionPresentValue(node, value, of, l.collectionPresenceProven(node)), nil
}
func (l *lowering) collectionPresentValue(node *ast.Node, value ir.Expression, of ir.Type, proven bool) ir.Expression {
	if proven && !value.Type().IsMaybe() {
		return value
	}
	if proven {
		// The fallback is unreachable by the independent presence proof. This reuses the
		// existing scalar-pair extraction in both backends without a runtime panic check.
		fallback := ir.Expression(ir.NumberConstant{Value: 0})
		if of == ir.Boolean {
			fallback = ir.BooleanConstant{Value: false}
		}
		return ir.Coalesce{Value: value, Fallback: fallback, Of: of}
	}
	message := "collection lookup failed: " + sourceExpression(node) + " is undefined"
	return ir.Coalesce{Value: value, Panic: ir.StringConstant{Index: l.constant(message)}, Of: of}
}

// Only an ordinary collection lookup receives the ruling's checked presence contract.
func (l *lowering) collectionNonNull(node *ast.Node) bool {
	return node.Kind == ast.KindNonNullExpression && load.CollectionLookup(l.checker, node.AsNonNullExpression().Expression) != nil
}
