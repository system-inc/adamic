package core

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnexpectedConcat = rule.Message{
	Id: "unexpectedConcat",
	// "One literal" rather than "a constant": a template with a substitution reports too, as
	// `${name}` + ' suffix' does, and joining that produces a template, which is one literal and
	// not a constant. The earlier wording said "constant" and was false on two of the four ahra
	// findings.
	Description: "These two literals are joined at runtime where the source could have written one " +
		"literal. Nothing is computed between them, so the concatenation only adds an operator for a " +
		"reader to follow and a chance for the two halves to drift apart.",
}

// NoUselessConcat flags a `+` whose two adjacent operands are both literal strings.
//
//	valid:   var a = 1 + 1;
//	valid:   var a = 'foo' + bar;
//	valid:   var foo = 'foo' +\n 'bar';
//	valid:   var string = (number + 1) + 'px';
//	valid:   (1 + +2) + `b`
//	invalid: 'a' + 'b'
//	invalid: foo + 'a' + 'b'
//	invalid: `a` + `b`
//
// # The operands are the ADJACENT ones, not the node's own children
//
// Concatenation associates left, so `foo + 'a' + 'b'` parses as `(foo + 'a') + 'b'` and the outer
// node's left child is a whole binary expression rather than a literal. Upstream walks INTO that
// child to find the rightmost thing actually next to the operator, and symmetrically walks into the
// right child to find its leftmost. Without that descent the rule reports only the innermost `+` of
// a chain and misses `foo + 'a' + 'b'` entirely, which is one of upstream's own reporting cases.
//
// The descent stops at anything that is not itself a `+` concatenation, which is what makes
// `(1 + +2) + ` + "`b`" + ` clean: the left operand's rightmost element is `+2`, a unary expression, not a
// literal.
//
// # Parentheses have to be skipped, and the corpus says so rather than my guessing
//
// ESTree produces no node for a parenthesized expression, so upstream's `node.left` is already the
// literal inside the parentheses. typescript-go produces one, so without skipping, every
// parenthesized operand reads as a non-literal and goes silent. This is not a hypothetical drawn
// from the brief's warning: upstream's corpus writes `(foo + 'a') + ('b' + 'c')` and asserts TWO
// findings, and measured against the installed build at 10.8.1, `('a') + 'b'`, `'a' + ('b')` and
// `('a') + ('b')` all report.
//
// The skip runs before the descent AND inside it, because either can be wrapped.
//
// # What counts as a string literal
//
// Upstream's `astUtils.isStringLiteral` is a string-valued `Literal` or ANY `TemplateLiteral`,
// including one carrying substitutions. Measured: a template with a substitution added to a
// string reports. That is wider than it first reads, since such a template is not a constant,
// and reproducing it is deliberate. A numeric literal is not a string literal, which is why
// `1 + '1'` and a template added to a number are both clean.
//
// This parser splits templates into two kinds, so both are named: a template with no substitutions
// and a template expression with them.
//
// # The same-line requirement
//
// Upstream compares the left operand's END line against the right operand's START line, so a
// concatenation broken across lines is left alone as a deliberate layout choice. Written here as a
// scan for a line break in the gap between the two operands, which is the same question asked the
// other way round. The two spellings were compared over ten shapes rather than assumed equivalent,
// including multiline templates on either side and a block comment carrying a newline between the
// operands, and they agree on every one. The multiline template is the case that could have
// separated them, because a template ENDING on line 2 keeps the operator on line 2.
//
// # Where the finding points
//
// At the `+` OPERATOR, not at the expression. Upstream overrides the location to the operator
// token's own span, which is why its corpus asserts a one-character range at column 5 for
// `'a' + 'b'`. Reporting the node would point at column 1 and no message-id fixture could see it.
//
// # No fixer, matching upstream
//
// Upstream declares none and marks the rule frozen. Joining the two literals is not always a
// spelling change: the halves may use different quote styles, one may be a template carrying a
// substitution, and the line break the author chose is often what makes a long string readable.
var NoUselessConcat = rule.Rule{
	Name: "no-useless-concat",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				expression := node.AsBinaryExpression()
				if !noUselessConcatIsConcatenation(node) {
					return
				}

				left := noUselessConcatRightmostOfLeft(expression.Left)
				right := noUselessConcatLeftmostOfRight(expression.Right)
				if !noUselessConcatIsStringLiteral(left) || !noUselessConcatIsStringLiteral(right) {
					return
				}

				leftRange := rule.TokenRange(ctx.SourceFile, left)
				rightRange := rule.TokenRange(ctx.SourceFile, right)
				if !noUselessConcatOnSameLine(ctx.SourceFile.Text(), leftRange.End(), rightRange.Pos()) {
					return
				}

				// The operator token rather than the expression, which is upstream's own location
				// override. `TokenRange` strips the leading trivia the raw token position carries,
				// so this is the single `+` character and matches upstream's one-column span.
				ctx.ReportRange(rule.TokenRange(ctx.SourceFile, expression.OperatorToken),
					messageUnexpectedConcat)
			},
		}
	},
}

// noUselessConcatIsConcatenation answers whether a node is a `+` binary expression.
//
// The operator is checked rather than assumed: this listener sees every binary expression, and
// `'a' - 'b'`, `1 * '2'` and `'a' === 'b'` all have two operands that could pass the literal test.
func noUselessConcatIsConcatenation(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return false
	}
	operator := node.AsBinaryExpression().OperatorToken
	return operator != nil && operator.Kind == ast.KindPlusToken
}

// noUselessConcatRightmostOfLeft descends the left operand to the element adjacent to the operator.
//
// Upstream's `getLeft`. Left-associativity means `foo + 'a' + 'b'` has a binary expression on the
// left of the outer `+`, and the thing actually beside that operator is that expression's own right
// operand. The parenthesis skip runs at every step, since `(foo + 'a') + 'b'` wraps the very node
// being descended and upstream's corpus asserts two findings on the parenthesized form.
func noUselessConcatRightmostOfLeft(node *ast.Node) *ast.Node {
	current := ast.SkipParentheses(node)
	for noUselessConcatIsConcatenation(current) {
		current = ast.SkipParentheses(current.AsBinaryExpression().Right)
	}
	return current
}

// noUselessConcatLeftmostOfRight descends the right operand to the element adjacent to the operator.
//
// Upstream's `getRight`, the mirror of the function above. It matters less often than its twin
// because `+` associates left, so a plain chain never puts a concatenation on the right, but
// `'a' + ('b' + 'c')` does and it is in upstream's corpus.
func noUselessConcatLeftmostOfRight(node *ast.Node) *ast.Node {
	current := ast.SkipParentheses(node)
	for noUselessConcatIsConcatenation(current) {
		current = ast.SkipParentheses(current.AsBinaryExpression().Left)
	}
	return current
}

// noUselessConcatIsStringLiteral is upstream's `astUtils.isStringLiteral`.
//
// A string-valued literal or any template literal. This parser gives a template two kinds depending
// on whether it carries substitutions, and BOTH count, which is upstream's own behaviour rather
// than a widening: measured against the installed build, a template carrying a substitution added
// to a string reports, even though such a template is not a constant.
//
// A numeric literal is deliberately absent. `1 + '1'` and a template added to a number are both
// upstream clean cases, and admitting numbers here would report both.
func noUselessConcatIsStringLiteral(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateExpression:
		return true
	}
	return false
}

// noUselessConcatOnSameLine reports whether the gap between two offsets holds no line break.
//
// Upstream asks whether the left operand's end line equals the right operand's start line. Asking
// whether the text between them contains a line break is the same question and needs no line map.
// The two spellings were compared over ten shapes rather than assumed equal, including a template
// ending on a later line than it starts, which is the case that could have separated them.
func noUselessConcatOnSameLine(text string, from int, to int) bool {
	if from < 0 || to > len(text) || from > to {
		return false
	}
	return !strings.ContainsAny(text[from:to], "\n\r")
}
