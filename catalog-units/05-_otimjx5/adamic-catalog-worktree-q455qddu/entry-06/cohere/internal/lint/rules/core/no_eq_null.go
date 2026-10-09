package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoEqNull = rule.Message{
	Id: "unexpected",
	Description: "This compares against `null` with `==` or `!=`, which also matches `undefined`. " +
		"That coercion is usually what the author wanted and is never what the code says, so a " +
		"reader cannot tell a deliberate both-values check from a typo. Write `=== null` when only " +
		"null is meant, or `== null` deliberately somewhere the convention says so.",
}

// NoEqNull flags a loose comparison against the `null` literal.
//
//	valid:   if (x === null) { }
//	valid:   if (null === f()) { }
//	valid:   if (x == undefined) { }
//	invalid: if (x == null) { }
//	invalid: if (x != null) { }
//	invalid: do {} while (null == x)
//
// Either side counts, because `null == x` and `x == null` are the same comparison written two ways
// and upstream tests both. Only `==` and `!=` count: `===` and `!==` are the repair the rule asks
// for, and `<`, `>`, `+` and the rest are different operators entirely.
//
// # The literal, not the value
//
// Upstream's predicate is `node.right.raw === "null"`, which is the SOURCE TEXT of the operand
// rather than anything about its value. So `x == undefined` is clean even though the comparison it
// performs is identical, and `let n = null; x == n` is clean because the operand is an identifier.
// Both measured against the installed build. That looks like an oversight and is the rule: it is a
// spelling rule about a literal, and widening it to "anything that is nullish" would be a different
// rule with a different false-positive surface.
//
// # Parentheses, which are a real divergence in the AST and not in the behaviour
//
// Upstream reads `node.right` directly with no parenthesis skip, which reads as though
// `x == (null)` must be clean. It is not: espree does not produce a node for a parenthesized
// expression at all, so `node.right` there IS the literal. typescript-go does produce one, so
// reading `binary.Right` directly would go silent on a case upstream reports.
//
// Measured against ESLint 10.8.1 with `sourceType: "script"`:
//
//	if (x == (null)) { }   reports, columns 5 to 16
//	if ((null) == x) { }   reports, columns 5 to 16
//
// So the skip is required for fidelity rather than being an improvement on upstream, and the span
// stays on the whole binary expression, parentheses included, which is what those columns say.
var NoEqNull = rule.Rule{
	Name: "no-eq-null",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				// The operator token is optional on the node type even though the parser always
				// supplies one for a binary expression, and `SkipParentheses` below dereferences,
				// so both nil tests are here rather than trusted away.
				if binary.OperatorToken == nil || binary.Left == nil || binary.Right == nil {
					return
				}
				switch binary.OperatorToken.Kind {
				case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken:
				default:
					return
				}

				if !isNullLiteralOperand(binary.Left) && !isNullLiteralOperand(binary.Right) {
					return
				}

				ctx.ReportNode(node, messageNoEqNull)
			},
		}
	},
}

// isNullLiteralOperand answers whether an operand is written as the `null` literal.
//
// The parenthesis skip is what makes `x == (null)` report, matching the installed build. See the
// rule's doc comment for why that is fidelity rather than a widening: espree gives parentheses no
// node, so upstream's direct field read already sees through them.
func isNullLiteralOperand(operand *ast.Node) bool {
	if operand == nil {
		return false
	}
	return ast.SkipParentheses(operand).Kind == ast.KindNullKeyword
}
