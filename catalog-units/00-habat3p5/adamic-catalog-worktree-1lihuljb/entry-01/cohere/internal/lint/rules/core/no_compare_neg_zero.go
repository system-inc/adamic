package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// negativeZeroComparisonOperators is every operator ESLint's original checks for, mapped to the
// text the message quotes back.
//
// A map rather than a switch because the lookup answers two questions at once: whether this binary
// expression is a comparison the rule cares about, and what to call it in the message. `in`,
// `instanceof`, and every arithmetic operator are absent on purpose, so `x ** -0` and `x + -0` are
// not comparisons and do not report.
var negativeZeroComparisonOperators = map[ast.Kind]string{
	ast.KindLessThanToken:                "<",
	ast.KindLessThanEqualsToken:          "<=",
	ast.KindGreaterThanToken:             ">",
	ast.KindGreaterThanEqualsToken:       ">=",
	ast.KindEqualsEqualsToken:            "==",
	ast.KindEqualsEqualsEqualsToken:      "===",
	ast.KindExclamationEqualsToken:       "!=",
	ast.KindExclamationEqualsEqualsToken: "!==",
}

// NoCompareNegZero flags a comparison against the literal `-0`.
//
//	valid:   x === 0
//	valid:   Object.is(x, -0)
//	valid:   x === -1
//	valid:   x === -0n
//	invalid: x === -0
//	invalid: -0 > x
//
// Every comparison operator in JavaScript treats `-0` and `+0` as the same number, so `x === -0` is
// true for `0` as well and `x < -0` is false for both. The comparison therefore does not test what
// its author wrote it to test: it reads as a check for negative zero and is not one. The only thing
// that distinguishes the two is `Object.is(x, -0)`, which is what the author almost certainly meant.
//
// The relational operators are included alongside the equality ones for the same reason, even though
// `x < -0` looks less like a negative-zero test. Writing `-0` there still signals an intent the
// operator cannot carry, and the `-` is either meaningless or a typo for a different number.
//
// What counts as `-0` is a unary minus applied to a numeric literal whose value is zero, so
// `-0.0`, `-0x0`, and `-0e10` all report: they are the same number written differently, and the
// hazard is in the value rather than the spelling. Parentheses are transparent on both the unary
// expression and its operand, matching ESLint, whose ESTree has no node for them.
//
// A BigInt `-0n` is exempt, and not as an oversight. BigInt has no negative zero at all: `-0n` is
// `0n`, so the comparison behaves exactly as written and there is nothing to warn about. TypeScript's
// `-0 as number` and `-0 satisfies number` are also exempt, following ESLint, whose parser does not
// see the assertion but whose reading of the shape is the one we match: the operand of the minus has
// to be the numeric literal itself, and in those the minus applies to an expression.
//
// Reported on the whole comparison rather than on the `-0` operand, which is ESLint's range and the
// right one here: the defect is the comparison, not the literal, and `Object.is` replaces all of it.
//
// No fix. `Object.is(x, -0)` is the likely intent but not the certain one, since an author who
// wrote `x < -0` may have meant `x < 0`, and a rewrite that guessed would silently change a
// condition.
var NoCompareNegZero = rule.Rule{
	Name: "no-compare-neg-zero",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			// Listening on the comparison rather than on every unary minus visits one node per
			// binary expression instead of one per negation, and it is also what makes reporting
			// once for `-0 === -0` fall out for free rather than needing a de-duplication pass.
			ast.KindBinaryExpression: func(node *ast.Node) {
				binaryExpression := node.AsBinaryExpression()
				if binaryExpression == nil || binaryExpression.OperatorToken == nil {
					return
				}

				operatorText, isComparison := negativeZeroComparisonOperators[binaryExpression.OperatorToken.Kind]
				if !isComparison {
					return
				}

				if !isNegativeZeroLiteral(binaryExpression.Left) && !isNegativeZeroLiteral(binaryExpression.Right) {
					return
				}

				ctx.ReportNode(node, rule.Message{
					Id: "unexpected",
					Description: "Comparing against `-0` with `" + operatorText + "`: every comparison " +
						"operator treats `-0` and `0` as the same number, so this is true for positive " +
						"zero too and does not test what it looks like it tests. `Object.is(x, -0)` is " +
						"the only comparison that distinguishes them.",
				})
			},
		}
	},
}

// isNegativeZeroLiteral says whether an expression is a unary minus applied to a numeric literal
// whose value is zero.
//
// Parentheses are stripped at both levels because ESTree has no node for them, so `((-0))` and
// `-(0)` are the same expression to the original and have to be here too. Stripping only at the
// outer level would miss `-(0)`, which is the shape a reader is most likely to write by hand.
func isNegativeZeroLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	unwrapped := ast.SkipParentheses(node)
	if unwrapped == nil || unwrapped.Kind != ast.KindPrefixUnaryExpression {
		return false
	}
	prefixExpression := unwrapped.AsPrefixUnaryExpression()
	if prefixExpression == nil || prefixExpression.Operator != ast.KindMinusToken {
		return false
	}
	return isZeroValuedNumericLiteral(prefixExpression.Operand)
}

// isZeroValuedNumericLiteral says whether an expression is a numeric literal whose value is zero.
//
// Text is compared to "0" rather than parsed, because typescript-go has already parsed it: the
// scanner stores a numeric literal's canonical decimal value, not the source spelling, so `0x0`,
// `0.0`, `0e10`, `00`, `0_0`, and `.0` all arrive here as exactly "0". Verified by probing the
// parser rather than assumed, since the opposite assumption is the one a reader coming from ESTree
// would carry: Espree keeps `raw` alongside `value`, and a port that reached for the source text
// would need every base parsed by hand to close holes this AST does not have.
//
// A BigInt literal is KindBigIntLiteral and never reaches here, which is the exemption `-0n` needs:
// BigInt has no negative zero, so the comparison behaves as written.
func isZeroValuedNumericLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	unwrapped := ast.SkipParentheses(node)
	if unwrapped == nil || unwrapped.Kind != ast.KindNumericLiteral {
		return false
	}
	numericLiteral := unwrapped.AsNumericLiteral()
	return numericLiteral != nil && numericLiteral.Text == "0"
}
