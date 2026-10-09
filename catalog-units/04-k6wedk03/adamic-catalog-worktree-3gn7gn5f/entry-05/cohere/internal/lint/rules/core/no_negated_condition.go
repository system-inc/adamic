package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoNegatedCondition = rule.Message{
	Id: "unexpectedNegated",
	Description: "This condition is negated and has an else branch, so the reader has to invert it " +
		"mentally to work out which branch runs when. Swapping the two branches and dropping the " +
		"negation says the same thing without the double take.",
}

// NoNegatedCondition flags a negated test that has an unconditional else beside it.
//
//	valid:   if (!a) {}
//	valid:   if (!a) {} else if (b) {}
//	valid:   if (a != b) {}
//	invalid: if (!a) { f(); } else { g(); }
//	invalid: if (a !== b) { f(); } else { g(); }
//	invalid: !a ? b : c
//
// # A negation alone is not the finding, and this is the whole rule
//
// `if (!a) {}` is clean. The complaint is a negated test paired with an else that the reader must
// hold in their head as "the positive case", which is why the else has to be UNCONDITIONAL: an
// `else if` starts a new question rather than closing this one, so `if (!a) {} else if (b) {}` is
// clean and `if (!a) {} else {}` is not. Upstream's `hasElseWithoutCondition` is exactly that test,
// and its corpus writes both directions.
//
// A conditional expression has no such escape, since `? :` always has both arms, so it is judged on
// the negation alone.
//
// # Three negations count and no others
//
// `!x`, `x != y`, `x !== y`. Upstream names them literally, and the omissions matter more than the
// inclusions: `==` and `===` are clean, and so is any other operator. There is no attempt to see
// through a negation written some other way, so `if (a === false)` and `if (!(a && b))`... the
// second of those IS a `!` and does report, which is worth stating because it reads as a compound
// condition rather than as a negated one.
//
// # The span is the whole statement, not the test
//
// Upstream reports `node`, which for an `if` is the statement including both branches. Measured
// against the installed 10.8.1 build: `if (!a) {;} else {;}` reports columns 1 to 21, the entire
// text. A port reporting the test would pass every message-id fixture while pointing at `!a`.
//
// # No fixer, deliberately
//
// `meta.fixable` is unset and the rule is `frozen` upstream. Swapping two branches is a change to
// what the code says rather than to how it is spelled, and the repair is not even mechanical: the
// negation has to be removed correctly, which for `!(a && b)` means either double-negating or
// applying De Morgan.
var NoNegatedCondition = rule.Rule{
	Name: "no-negated-condition",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindIfStatement: func(node *ast.Node) {
				statement := node.AsIfStatement()
				if !noNegatedConditionHasElseWithoutCondition(statement) {
					return
				}
				if noNegatedConditionIsNegated(statement.Expression) {
					ctx.ReportNode(node, messageNoNegatedCondition)
				}
			},
			ast.KindConditionalExpression: func(node *ast.Node) {
				if noNegatedConditionIsNegated(node.AsConditionalExpression().Condition) {
					ctx.ReportNode(node, messageNoNegatedCondition)
				}
			},
		}
	},
}

// noNegatedConditionHasElseWithoutCondition is upstream's `hasElseWithoutCondition`.
//
// An else that is itself an `if` is an `else if`, which upstream treats as a continuation rather
// than as the positive branch. Our parser represents `else if` the same way, as an ElseStatement
// holding an IfStatement, so the test reads identically.
func noNegatedConditionHasElseWithoutCondition(statement *ast.IfStatement) bool {
	return statement.ElseStatement != nil &&
		statement.ElseStatement.Kind != ast.KindIfStatement
}

// noNegatedConditionIsNegated is upstream's `isNegatedIf`: a `!` prefix or a `!=` / `!==` operator.
//
// # The parenthesis unwrap is load bearing and no imported fixture can see it
//
// Upstream's parser folds parentheses away, so `if ((!a))` arrives at its test as a bare unary.
// Ours keeps a `KindParenthesizedExpression`, so a port without this loop declines every
// parenthesized form and loses the finding silently. Measured against the installed 10.8.1 build,
// all four of these REPORT upstream:
//
//	if ((!a)) { f(); } else { g(); }          columns 1 to 33
//	if ((a !== b)) { f(); } else { g(); }     columns 1 to 38
//	if (((!a))) { f(); } else { g(); }        columns 1 to 35, so the unwrap has to loop
//	(!a) ? b : c                              columns 1 to 13
//
// Upstream's corpus writes none of them, and it cannot: its parser deleted the node before the test
// was written. So this is invisible to the imported fixtures and is pinned by cases of our own.
//
// The loop rather than a single step is what `((!a))` needs. `ast.SkipParentheses` is deliberately
// not used: it dereferences its argument, and an `if` with no condition is not reachable here but a
// conditional's is read through the same helper, so the nil check stays local.
func noNegatedConditionIsNegated(test *ast.Node) bool {
	if test == nil {
		return false
	}
	for test.Kind == ast.KindParenthesizedExpression {
		test = test.AsParenthesizedExpression().Expression
		if test == nil {
			return false
		}
	}
	if test.Kind == ast.KindPrefixUnaryExpression {
		return test.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	}
	if test.Kind == ast.KindBinaryExpression {
		switch test.AsBinaryExpression().OperatorToken.Kind {
		case ast.KindExclamationEqualsToken, ast.KindExclamationEqualsEqualsToken:
			return true
		}
	}
	return false
}
