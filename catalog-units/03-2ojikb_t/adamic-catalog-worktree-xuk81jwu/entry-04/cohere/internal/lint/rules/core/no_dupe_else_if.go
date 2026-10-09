package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnexpectedDuplicateElseIf = rule.Message{
	Id: "unexpected",
	Description: "This branch can never execute, because an earlier condition in the same " +
		"if-else-if chain already covers it. The code inside is unreachable and nothing says so: " +
		"it compiles, it type-checks, and it silently never runs, which usually means either the " +
		"condition was meant to be different or the branch was meant to come first.",
}

// NoDupeElseIf flags a branch in an if-else-if chain whose condition an earlier branch subsumes.
//
//	valid:   if(a) {} else if(b) {}
//	valid:   if(a) {} else if(a && b) {}          // narrower, so reachable
//	invalid: if(a) {} else if(a) {}
//	invalid: if(a || b) {} else if(a) {}
//	invalid: if(a && b) {} else if(b && a) {}     // && is commutative here
//
// This is subsumption rather than equality, and the difference is most of the rule. `if(a || b)`
// followed by `else if(a)` is unreachable even though the two conditions are not the same
// expression, because everything the second matches the first already matched.
//
// The algorithm is ESLint's and is worth stating in one line, since the nested loops obscure it: a
// condition is written as an OR of AND-groups, and a branch is unreachable when every one of its
// OR-groups has been covered by some earlier branch. A group is covered when an earlier group is a
// subset of it, since a smaller AND is a weaker requirement and therefore matches more.
//
// Both operands of a commutative operator are tried when comparing, so `a && b` equals `b && a`.
// That is only sound because these are conditions read for truthiness, where reordering cannot
// change the result. Outside a boolean context it would be wrong, which is why the equality used
// here is local to this rule rather than a general one.
//
// A chain of length n costs O(n squared) comparisons, which the original accepts and so do we. An
// if-else-if chain long enough to matter is a defect this rule is not the one to report.
var NoDupeElseIf = rule.Rule{
	Name: "no-dupe-else-if",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindIfStatement: func(node *ast.Node) {
				test := node.AsIfStatement().Expression
				if test == nil {
					return
				}

				// An `&&` condition is checked whole and also per conjunct. `if(a) {} else if(a && b)`
				// is reachable, but the chain `if(a && b) {} else if(a && b && c)` is not, and
				// catching the second needs the parts as well as the whole.
				conditionsToCheck := []*ast.Node{test}
				if isLogicalOperator(test, ast.KindAmpersandAmpersandToken) {
					conditionsToCheck = append(conditionsToCheck, splitByOperator(test, ast.KindAmpersandAmpersandToken)...)
				}

				// Each condition becomes a list of OR-groups, each group a list of AND-operands.
				remaining := make([][][]*ast.Node, 0, len(conditionsToCheck))
				for _, condition := range conditionsToCheck {
					orGroups := splitByOperator(condition, ast.KindBarBarToken)
					groups := make([][]*ast.Node, 0, len(orGroups))
					for _, orGroup := range orGroups {
						groups = append(groups, splitByOperator(orGroup, ast.KindAmpersandAmpersandToken))
					}
					remaining = append(remaining, groups)
				}

				// Walk up the chain of enclosing else branches. Only an `else if` counts: this
				// statement must be the alternate of its parent, not its consequent.
				for current := node; ; {
					parent := current.Parent
					if parent == nil || parent.Kind != ast.KindIfStatement {
						return
					}
					if parent.AsIfStatement().ElseStatement != current {
						return
					}
					current = parent

					earlierTest := current.AsIfStatement().Expression
					if earlierTest == nil {
						return
					}

					earlierGroups := make([][]*ast.Node, 0, 4)
					for _, orGroup := range splitByOperator(earlierTest, ast.KindBarBarToken) {
						earlierGroups = append(earlierGroups, splitByOperator(orGroup, ast.KindAmpersandAmpersandToken))
					}

					// Drop every OR-group this earlier condition covers.
					for conditionIndex, groups := range remaining {
						kept := groups[:0]
						for _, group := range groups {
							if !anyGroupIsSubsetOf(ctx.SourceFile, earlierGroups, group) {
								kept = append(kept, group)
							}
						}
						remaining[conditionIndex] = kept
					}

					// A condition with nothing left is fully covered, so this branch is dead.
					for _, groups := range remaining {
						if len(groups) == 0 {
							ctx.ReportNode(test, messageUnexpectedDuplicateElseIf)
							return
						}
					}
				}
			},
		}
	},
}

// anyGroupIsSubsetOf reports whether some earlier AND-group is a subset of this one.
//
// The direction is the part that reads backwards and is correct: a smaller AND is a weaker
// condition, so if an earlier branch required only `a` and this one requires `a && b`, everything
// this one matches the earlier one already matched.
func anyGroupIsSubsetOf(sourceFile *ast.SourceFile, earlierGroups [][]*ast.Node, group []*ast.Node) bool {
	for _, earlierGroup := range earlierGroups {
		if isAndSubset(sourceFile, earlierGroup, group) {
			return true
		}
	}
	return false
}

// isAndSubset reports whether every operand of `subset` appears in `superset`.
func isAndSubset(sourceFile *ast.SourceFile, subset []*ast.Node, superset []*ast.Node) bool {
	for _, operand := range subset {
		found := false
		for _, candidate := range superset {
			if conditionsAreEqual(sourceFile, operand, candidate) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// conditionsAreEqual compares two conditions, treating && and || as commutative.
//
// Falls through to the package's token oracle for anything else, which is what makes `a.b` and
// `a?.b` distinct and two differently written literals distinct, while ignoring comments and
// whitespace.
func conditionsAreEqual(sourceFile *ast.SourceFile, left *ast.Node, right *ast.Node) bool {
	left = ast.SkipParentheses(left)
	right = ast.SkipParentheses(right)
	if left == nil || right == nil {
		return false
	}
	if left.Kind != right.Kind {
		return false
	}

	if left.Kind == ast.KindBinaryExpression {
		leftBinary := left.AsBinaryExpression()
		rightBinary := right.AsBinaryExpression()
		if leftBinary.OperatorToken != nil && rightBinary.OperatorToken != nil &&
			leftBinary.OperatorToken.Kind == rightBinary.OperatorToken.Kind &&
			(leftBinary.OperatorToken.Kind == ast.KindBarBarToken ||
				leftBinary.OperatorToken.Kind == ast.KindAmpersandAmpersandToken) {
			return (conditionsAreEqual(sourceFile, leftBinary.Left, rightBinary.Left) &&
				conditionsAreEqual(sourceFile, leftBinary.Right, rightBinary.Right)) ||
				(conditionsAreEqual(sourceFile, leftBinary.Left, rightBinary.Right) &&
					conditionsAreEqual(sourceFile, leftBinary.Right, rightBinary.Left))
		}
	}

	return hasSameTokens(sourceFile, left, right)
}

// splitByOperator flattens a chain of one logical operator into its operands.
//
// `a || b || c` is a left-nested tree, so a rule reading only the top node sees two operands rather
// than three and misses `else if(b)`. Flattening is what makes the group comparison work on the
// operands people actually wrote.
func splitByOperator(node *ast.Node, operator ast.Kind) []*ast.Node {
	node = ast.SkipParentheses(node)
	if node == nil {
		return nil
	}
	if isLogicalOperator(node, operator) {
		binary := node.AsBinaryExpression()
		return append(
			splitByOperator(binary.Left, operator),
			splitByOperator(binary.Right, operator)...,
		)
	}
	return []*ast.Node{node}
}

// isLogicalOperator reports a binary expression joined by exactly this operator.
func isLogicalOperator(node *ast.Node, operator ast.Kind) bool {
	node = ast.SkipParentheses(node)
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := node.AsBinaryExpression()
	return binary.OperatorToken != nil && binary.OperatorToken.Kind == operator
}
