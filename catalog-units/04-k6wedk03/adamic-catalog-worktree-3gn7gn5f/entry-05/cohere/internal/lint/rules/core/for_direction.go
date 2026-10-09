package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageIncorrectDirection = rule.Message{
	Id: "incorrectDirection",
	Description: "The update clause of this loop moves the counter away from the condition, so the " +
		"condition can never become false and the loop runs forever. Nothing about this fails to " +
		"compile and nothing throws: the program simply stops making progress, which is why it is " +
		"usually found by a hung process rather than by reading the loop.",
}

// ForDirection flags a for loop whose update moves the counter away from its condition.
//
//	valid:   for(let i = 0; i < 10; i++)
//	valid:   for(let i = 10; i > 0; i--)
//	valid:   for(let i = 0; i < 10; i += step)   // step is not statically known
//	invalid: for(let i = 0; i < 10; i--)
//	invalid: for(let i = 10; i > 0; i++)
//	invalid: for(let i = 0; i < 10; i -= 1)
//
// The counter may sit on either side of the condition, and which side it is on flips what counts as
// wrong: `i < 10` wants an increment, `10 > i` wants one too, and a rule checking only the left
// operand is silent on half the loops anyone writes. Both sides are checked, matching ESLint.
//
// A report needs exactly one expression modifying the counter. `for(...; ...; i++, i--)` is
// ambiguous rather than wrong, and ESLint deliberately says nothing about it: the loop may be
// doing something deliberate that no syntactic reading can settle. The count is what makes that
// distinction, so a sequence expression is walked to count modifications rather than to find the
// first one.
//
// The compound-assignment arm asks for a statically known sign, and that is the only place this
// rule can be uncertain. `i += step` is not reported because step could be negative, and reporting
// it would flag a correct loop. ESLint reaches for its scope analysis here and resolves a `const`
// binding to its value; ours reads only literals and their unary negations, which is a deliberate
// narrowing recorded below.
var ForDirection = rule.Rule{
	Name: "for-direction",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindForStatement: func(node *ast.Node) {
				forStatement := node.AsForStatement()
				if forStatement.Condition == nil || forStatement.Incrementor == nil {
					return
				}

				condition := ast.SkipParentheses(forStatement.Condition)
				if condition == nil || condition.Kind != ast.KindBinaryExpression {
					return
				}
				binary := condition.AsBinaryExpression()
				if binary.OperatorToken == nil {
					return
				}

				var wrongDirectionOnLeft, wrongDirectionOnRight int
				switch binary.OperatorToken.Kind {
				case ast.KindLessThanToken, ast.KindLessThanEqualsToken:
					wrongDirectionOnLeft, wrongDirectionOnRight = -1, 1
				case ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken:
					wrongDirectionOnLeft, wrongDirectionOnRight = 1, -1
				default:
					return
				}

				// Both operands are considered, and at most one report is emitted. A condition
				// like `i < i` would otherwise report twice about one loop.
				for _, side := range []struct {
					operand        *ast.Node
					wrongDirection int
				}{
					{binary.Left, wrongDirectionOnLeft},
					{binary.Right, wrongDirectionOnRight},
				} {
					operand := ast.SkipParentheses(side.operand)
					if operand == nil || operand.Kind != ast.KindIdentifier {
						continue
					}

					modifications := counterModifications(forStatement.Incrementor, operand.Text())
					if len(modifications) != 1 {
						continue
					}
					if modificationDirection(modifications[0], operand.Text()) == side.wrongDirection {
						ctx.ReportNode(node, messageIncorrectDirection)
						return
					}
				}
			},
		}
	},
}

// counterModifications collects every expression in the update clause that modifies the counter.
//
// The count is what the rule acts on rather than the first hit, because two modifications make the
// loop ambiguous rather than wrong. A comma expression is flattened recursively so `i++, i--` is
// seen as two rather than as one comma expression that happens to increment.
func counterModifications(node *ast.Node, counter string) []*ast.Node {
	if node == nil {
		return nil
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindPostfixUnaryExpression:
		if isIdentifierNamed(node.AsPostfixUnaryExpression().Operand, counter) {
			return []*ast.Node{node}
		}

	case ast.KindPrefixUnaryExpression:
		operator := node.AsPrefixUnaryExpression().Operator
		if (operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken) &&
			isIdentifierNamed(node.AsPrefixUnaryExpression().Operand, counter) {
			return []*ast.Node{node}
		}

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken == nil {
			return nil
		}
		if binary.OperatorToken.Kind == ast.KindCommaToken {
			return append(
				counterModifications(binary.Left, counter),
				counterModifications(binary.Right, counter)...,
			)
		}
		if ast.IsAssignmentOperator(binary.OperatorToken.Kind) && isIdentifierNamed(binary.Left, counter) {
			return []*ast.Node{node}
		}
	}

	return nil
}

// modificationDirection reads which way one modifying expression moves the counter.
//
// Zero means unknown, and unknown is silence. That is the direction this rule has to err in: a
// wrong finding here tells someone their working loop is an infinite loop.
func modificationDirection(node *ast.Node, counter string) int {
	switch node.Kind {
	case ast.KindPostfixUnaryExpression:
		return incrementDirection(node.AsPostfixUnaryExpression().Operator)

	case ast.KindPrefixUnaryExpression:
		return incrementDirection(node.AsPrefixUnaryExpression().Operator)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken == nil || !isIdentifierNamed(binary.Left, counter) {
			return 0
		}

		// Only += and -= carry a direction. A plain `i = something` says nothing about which way
		// the counter moves, and the other compound operators (*=, <<=) are not this rule's
		// business, which is ESLint's behavior rather than an omission.
		switch binary.OperatorToken.Kind {
		case ast.KindPlusEqualsToken:
			return staticSign(binary.Right)
		case ast.KindMinusEqualsToken:
			return -staticSign(binary.Right)
		}
	}
	return 0
}

// incrementDirection maps ++ and -- to a direction, and anything else to unknown.
//
// The fallthrough is unreachable and stays anyway. The prefix caller filters to ++ and -- before a
// node reaches here, and the postfix caller does not need to: JavaScript has exactly two postfix
// operators and both are handled. So no fixture can kill a mutant that changes what the
// fallthrough returns, which a mutation sweep reports the same way it reports a missing test.
//
// Kept because the alternative is a function that is total only by grammar, and a reader would
// have to know the grammar to see that. Recorded because "no fixture measures this" and "nothing
// can" are the same output and only the second one is fine to leave.
func incrementDirection(operator ast.Kind) int {
	switch operator {
	case ast.KindPlusPlusToken:
		return 1
	case ast.KindMinusMinusToken:
		return -1
	}
	return 0
}

// staticSign reads the sign of a statically known numeric operand, or 0 when it is not known.
//
// # A deliberate narrowing from ESLint, and the direction it errs in
//
// ESLint calls getStaticValue with the scope, which resolves a `const step = -1` binding to its
// value and reports `for(let i = 0; i < 10; i += step)`. This reads only literals and their unary
// negations, so that loop is not reported here.
//
// The narrowing is a false negative rather than a false positive, which is the survivable direction
// for this rule specifically: a missed infinite loop is found by the process hanging, while a wrong
// finding tells someone their working loop is broken and costs them the time to prove it is not.
//
// Recorded rather than quietly dropped, because it is a real gap: it is the one shape where our
// answer differs from the gate's, and if the differential ever shows a for-direction disagreement
// this comment is where to start.
func staticSign(node *ast.Node) int {
	node = ast.SkipParentheses(node)
	if node == nil {
		return 0
	}

	switch node.Kind {
	case ast.KindNumericLiteral:
		return numericLiteralSign(node.Text())

	case ast.KindBigIntLiteral:
		return numericLiteralSign(node.Text())

	case ast.KindTrueKeyword:
		// Number(true) is 1, and ESLint accepts booleans here. `i += true` is nobody's code, but
		// matching the set of types keeps a needless divergence out of the differential.
		return 1

	case ast.KindFalseKeyword:
		return 0

	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		switch unary.Operator {
		case ast.KindMinusToken:
			return -staticSign(unary.Operand)
		case ast.KindPlusToken:
			return staticSign(unary.Operand)
		}
	}

	return 0
}

// numericLiteralSign reports 1 for a nonzero literal and 0 for a zero one.
//
// The parser has already normalized the text, which is the whole reason this is three lines rather
// than a scanner. `0e5`, `0x0`, and `0.0` all arrive as "0"; `1e2` arrives as "100". So every base,
// exponent, and separator question is answered before the rule sees the literal, and the only
// thing left to ask is whether the digits are a zero.
//
// Written the long way first, with branches for hex prefixes and exponents, and every one of those
// branches was unreachable. Found by mutating the exponent branch to return the opposite answer and
// watching two fixtures that should have failed stay green: a mutant nothing can kill is either a
// missing fixture or dead code, and here it was dead code. The probe that settled it printed the
// literal text the rule actually receives rather than the text in the source.
//
// A literal is never negative in the AST either: `-1` is a unary minus applied to `1`, which
// staticSign handles.
func numericLiteralSign(text string) int {
	for _, character := range text {
		if character != '0' && character != '.' {
			return 1
		}
	}
	return 0
}

// isIdentifierNamed reports an identifier with exactly this name, through any parentheses.
func isIdentifierNamed(node *ast.Node, name string) bool {
	node = ast.SkipParentheses(node)
	return node != nil && node.Kind == ast.KindIdentifier && node.Text() == name
}
