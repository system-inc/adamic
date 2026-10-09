package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoSelfCompareComparingToSelf = rule.Message{
	Id: "comparingToSelf",
	Description: "Both sides of this comparison are written identically, so the answer does not " +
		"depend on the data: it is true, or it is false, every time. The two shapes this is almost " +
		"always hiding are a typo in one operand, and a deliberate `x !== x` reaching for a NaN " +
		"check, which `Number.isNaN(x)` says out loud instead.",
}

// noSelfCompareOperators are the eight comparison operators upstream watches.
//
// Upstream spells them as a set of source strings; here they are token kinds, which is the same set
// read off the parser rather than off the text. Deliberately NOT the whole relational family: `in`
// and `instanceof` are comparisons in the grammar and are absent from upstream's set, because
// `x in x` and `x instanceof x` are not tautologies, and adding them because they look like they
// belong is a false-positive class the imported corpus cannot see.
var noSelfCompareOperators = map[ast.Kind]bool{
	ast.KindEqualsEqualsEqualsToken:      true,
	ast.KindEqualsEqualsToken:            true,
	ast.KindExclamationEqualsEqualsToken: true,
	ast.KindExclamationEqualsToken:       true,
	ast.KindGreaterThanToken:             true,
	ast.KindLessThanToken:                true,
	ast.KindGreaterThanEqualsToken:       true,
	ast.KindLessThanEqualsToken:          true,
}

// NoSelfCompare flags a comparison whose two sides are written with the same tokens.
//
//	valid:   if (x === y) { }
//	valid:   if (1 === 2) { }
//	valid:   foo.bar.baz === foo.bar.qux
//	valid:   class C { #field; foo() { this.#field === this['#field']; } }
//	invalid: if (x === x) { }
//	invalid: x !== x
//	invalid: foo.bar().baz.qux >= foo.bar ().baz .qux
//
// # Token equality, not structural equality, and not source text
//
// Upstream's predicate is `hasSameTokens`, which asks the source code for both operands' token
// streams and compares them pairwise on type and value. Two consequences the corpus pins:
//
// `foo.bar().baz.qux >= foo.bar ().baz .qux` REPORTS. The two sides differ in whitespace and are
// the same tokens, so a port comparing raw source text is silent on upstream's own reporting case.
//
// `this.#field === this['#field']` is CLEAN. The two sides read the same characters where it
// matters and are different tokens: a private name is one token, while a subscript is a bracket, a
// string and a bracket. A port comparing text with the punctuation stripped reports it, and it is a
// real distinction rather than a curiosity, because the two access different things entirely.
//
// The shelf already holds this oracle as `hasSameTokens`, written for `prefer-spread` against the
// same upstream helper and carrying one edge case a fixture found there: an empty `[]` and an
// empty `[ ]` are the same tokens. Reusing it rather than writing a ninth spelling of token
// equality is the whole reason it is on the shelf.
//
// # This rule reports a tautology, not a side effect
//
// `x() === x()` is in neither list upstream, and upstream reports it, because two identical calls
// are identical tokens. That is upstream saying the shape is worth a look even when the two sides
// could evaluate differently; it is a deliberate over-report rather than an oversight, and it is
// reproduced rather than improved on.
var NoSelfCompare = rule.Rule{
	Name: "no-self-compare",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binaryExpression := node.AsBinaryExpression()
				if binaryExpression == nil || binaryExpression.OperatorToken == nil {
					return
				}
				if !noSelfCompareOperators[binaryExpression.OperatorToken.Kind] {
					return
				}
				if hasSameTokens(ctx.SourceFile, binaryExpression.Left, binaryExpression.Right) {
					ctx.ReportNode(node, messageNoSelfCompareComparingToSelf)
				}
			},
		}
	},
}
