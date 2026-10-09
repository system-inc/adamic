package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoNewStatement = rule.Message{
	Id: "noNewStatement",
	Description: "This constructs an object and then throws it away, so the only thing the line " +
		"can accomplish is whatever the constructor does on the side. A constructor that works by " +
		"side effect is doing a function's job under a name that promises a value, and the next " +
		"reader cannot tell the discarded result from a mistake. Call a function instead, or keep " +
		"the object and use it.",
}

// NoNew flags a `new` expression used as a whole statement.
//
//	valid:   var a = new Date()
//	valid:   var a; if (a === new Date()) { a = false; }
//	invalid: new Date()
//
// # The anchor is the statement, not the expression
//
// Upstream selects `ExpressionStatement > NewExpression` and then reports `node.parent`, so the
// finding covers the statement rather than the expression, and the trailing semicolon is inside the
// span. Measured on eslint 10.8.1: `new Date;` reports columns 1 through 10, which is one past the
// semicolon.
//
// Anchoring on the statement here rather than on the `new` gets that span by construction instead of
// by walking back up to the parent, and it also makes the "is this the whole statement" question the
// listener's own rather than something to re-derive.
//
// # Parentheses, which are a parser difference rather than a judgment
//
// The upstream selector demands a DIRECT child, and in ESTree it gets one however many parentheses
// are written, because that tree discards them. Ours does not: `(new Date())` parses as a
// parenthesized expression wrapping the `new`. So the direct-child reading, ported literally, goes
// silent on a case upstream reports.
//
// Measured on eslint 10.8.1 rather than assumed, since the corpus writes no parenthesized form:
//
//	(new Date());      reports, columns 1 through 14
//	(new Date);        reports, columns 1 through 12
//	((new Date()));    reports, columns 1 through 16
//
// The span stays the statement's in every one, which is what `ast.SkipParentheses` on the
// expression reproduces exactly: the parens are seen through for the decision and kept in the span.
//
// # What stays silent, and why each is a different question
//
// A `new` anywhere other than the top of a statement expression is fine, because its value is being
// used: `var a = new Date()` binds it, `a === new Date()` compares it, `new Date().getTime()` calls
// through it, and `void new Date();` and `new Date() + 1;` are statements whose expression is an
// operator rather than the `new`. A comma expression is the sharpest of these, and it is silent:
// `new Date(), new Foo();` reports nothing on the same build, because the statement's expression is
// the comma. That reads like a gap and it is upstream's, reproduced rather than improved on.
//
// Nesting reports once rather than twice. `new new Foo();` is one finding, because only the outer
// `new` is the statement's expression and the inner one is its callee.
var NoNew = rule.Rule{
	Name: "no-new",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindExpressionStatement: func(node *ast.Node) {
				// See the doc above: our tree keeps parentheses that ESTree discards, so the
				// direct-child reading needs this to match upstream on `(new Date());`.
				expression := ast.SkipParentheses(node.AsExpressionStatement().Expression)
				if expression == nil || expression.Kind != ast.KindNewExpression {
					return
				}
				ctx.ReportNode(node, messageNoNewStatement)
			},
		}
	},
}
