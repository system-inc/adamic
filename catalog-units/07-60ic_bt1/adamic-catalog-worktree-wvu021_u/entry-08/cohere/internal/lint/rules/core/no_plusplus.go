package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoPlusplusOptions configures the one exemption.
//
// Upstream's schema is a single object with one boolean, `allowForLoopAfterthoughts`, defaulting to
// false. Our config layer unwraps the severity tuple before dispatch, so the decoder receives that
// object directly rather than upstream's one-element array.
type NoPlusplusOptions struct {
	// AllowForLoopAfterthoughts exempts the update expression of a `for` loop, which is the one
	// place `i++` is idiomatic rather than incidental. Absent means false, which is upstream's
	// default.
	AllowForLoopAfterthoughts bool `json:"allowForLoopAfterthoughts"`
}

// noPlusplusUnexpectedUnaryOp builds the finding, naming the operator.
//
// The Id is fixed and only the Description moves, so `ExpectFindings` can count these while the
// rendered text names the operator the way upstream's `{{operator}}` interpolation does.
func noPlusplusUnexpectedUnaryOp(operator string) rule.Message {
	return rule.Message{
		Id: "unexpectedUnaryOp",
		Description: fmt.Sprintf(
			"Unary operator `%s` used. It mutates its operand and evaluates to a different value "+
				"depending on which side it sits, so `a[i++]` and `a[++i]` read almost identically "+
				"and index different elements. Automatic semicolon insertion compounds it: a line "+
				"ending in a value followed by a line starting with `++` joins into one statement. "+
				"Write the assignment out as `x += 1`, which does one thing and reads the same "+
				"wherever it appears.",
			operator,
		),
	}
}

// NoPlusplus flags the `++` and `--` operators.
//
//	valid:   var foo = 0; foo += 1;
//	valid:   for (i = 0; i < l; i++) {}     (with allowForLoopAfterthoughts: true)
//	invalid: var foo = 0; foo++;
//	invalid: for (i = 0; i < l; i++) {}
//
// Ported from `no-plusplus` in ESLint, read from the clone at `lib/rules/no-plusplus.js`. One
// option, one message, no fixer: `meta` carries `type`, `defaultOptions`, `docs`, `schema` and
// `messages`, and the string `fixable` does not appear in the file.
//
// The whole 23-case corpus was extracted from upstream's own tester and replayed against the rule
// through the ESLint Linter API before any code was written. It agreed on all 23.
//
// # The afterthought exemption is recursive, and that is the whole of the port
//
// With `allowForLoopAfterthoughts`, upstream exempts the update expression of a `for` loop. But the
// update may be a SEQUENCE expression, and the exemption reaches through it to any depth:
//
//	for (;; i++) {}                          exempt, the update itself
//	for (;; foo(), i++) {}                   exempt, an operand of the update
//	for (;; foo(), (bar(), (baz(), i++))) {}  exempt, nested sequences
//
// Upstream expresses this as `isForLoopAfterthought` recursing while the parent is a
// `SequenceExpression`. Reproduced as the same recursion rather than as a "somewhere inside a for
// update" containment test, and the difference is measurable: containment would also exempt
// `for (;; foo(i++)) {}`, where the operator is an ARGUMENT rather than an operand, and upstream
// reports that. The corpus writes it.
//
// # Our parser keeps parentheses, and here that costs nothing
//
// `KindParenthesizedExpression` is a real node here and espree folds it away, so the recursion above
// would stop at a paren that upstream never sees. `for (;; (i++)) {}` is the shape, and it is not in
// the corpus. The walk skips parentheses as well as sequences for that reason; a fixture covers it,
// written from reading our AST rather than from upstream's cases.
var NoPlusplus = rule.Rule{
	Name: "no-plusplus",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		allowForLoopAfterthoughts := false
		if resolved, isNoPlusplusOptions := rule.OptionsAs[NoPlusplusOptions](options); isNoPlusplusOptions {
			allowForLoopAfterthoughts = resolved.AllowForLoopAfterthoughts
		}

		return rule.Listeners{
			ast.KindPostfixUnaryExpression: func(node *ast.Node) {
				expression := node.AsPostfixUnaryExpression()
				if expression == nil {
					return
				}
				if allowForLoopAfterthoughts && noPlusplusIsForLoopAfterthought(node) {
					return
				}
				ctx.ReportNode(node, noPlusplusUnexpectedUnaryOp(
					noPlusplusOperatorText(expression.Operator)))
			},

			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				expression := node.AsPrefixUnaryExpression()
				if expression == nil {
					return
				}
				// A prefix `+`, `-`, `!`, `~`, `void` and friends are all prefix unary expressions
				// here, and only the two update operators are this rule's subject. Upstream has a
				// dedicated `UpdateExpression` node and so never needs this filter.
				if expression.Operator != ast.KindPlusPlusToken &&
					expression.Operator != ast.KindMinusMinusToken {
					return
				}
				if allowForLoopAfterthoughts && noPlusplusIsForLoopAfterthought(node) {
					return
				}
				ctx.ReportNode(node, noPlusplusUnexpectedUnaryOp(
					noPlusplusOperatorText(expression.Operator)))
			},
		}
	},
}

// noPlusplusIsForLoopAfterthought reproduces upstream's `isForLoopAfterthought`.
//
// Walks up through sequence expressions to see whether this node IS the update expression of a
// `for` statement. Identity against `statement.Incrementor` rather than containment, which is what
// keeps `for (;; foo(i++)) {}` reporting: there the operator is an argument to a call that is the
// update, not an operand of the update.
func noPlusplusIsForLoopAfterthought(node *ast.Node) bool {
	current := node
	for current != nil {
		parent := current.Parent
		if parent == nil {
			return false
		}

		// Parentheses are transparent here and invisible upstream, whose parser folds them away
		// before the rule runs. Skipping them keeps `for (;; (i++)) {}` exempt, matching what
		// upstream would do with the same source.
		if parent.Kind == ast.KindParenthesizedExpression {
			current = parent
			continue
		}

		if parent.Kind == ast.KindBinaryExpression {
			binary := parent.AsBinaryExpression()
			// A comma expression is our spelling of upstream's SequenceExpression, and it is the
			// only binary operator this recursion follows.
			if binary != nil && binary.OperatorToken != nil &&
				binary.OperatorToken.Kind == ast.KindCommaToken {
				current = parent
				continue
			}
			return false
		}

		if parent.Kind == ast.KindForStatement {
			statement := parent.AsForStatement()
			return statement != nil && statement.Incrementor == current
		}
		return false
	}
	return false
}

// noPlusplusOperatorText renders the operator for the message.
func noPlusplusOperatorText(operator ast.Kind) string {
	if operator == ast.KindMinusMinusToken {
		return "--"
	}
	return "++"
}

// DefaultNoPlusplusOptions is the unconfigured answer.
//
// Upstream's default is false, so the zero value happens to be right. Written explicitly anyway,
// because a future option defaulting to true would silently invert through the generic decoder and
// this is the line where that would have to be noticed.
func DefaultNoPlusplusOptions() NoPlusplusOptions {
	return NoPlusplusOptions{AllowForLoopAfterthoughts: false}
}

// DecodeNoPlusplusOptions reads this rule's configuration from the config layer.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` so a bare `"error"` configuration, which arrives
// as empty input, resolves to the documented default instead of erroring.
func DecodeNoPlusplusOptions(raw []byte) (any, error) {
	options := DefaultNoPlusplusOptions()
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}
