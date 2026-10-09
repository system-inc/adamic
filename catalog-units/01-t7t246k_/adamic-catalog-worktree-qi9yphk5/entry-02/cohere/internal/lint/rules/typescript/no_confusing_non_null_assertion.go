package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoConfusingNonNullAssertion flags a non-null assertion sitting immediately to the left of an
// operator whose spelling it can be mistaken for, or whose precedence it appears to change.
//
//	valid:   a == b!
//	valid:   a != b
//	valid:   (a + b!) == c
//	invalid: a! == b       reads as `a !== b`
//	invalid: a! = b        reads as `a != b`
//	invalid: a! in b       reads as `!(a in b)`
//
// Every case here is about how the source LOOKS rather than what it does. `a! == b` and `a !== b`
// differ by one space and mean opposite things, and `a! in b` reads to most people as a negation of
// the whole test when the `!` binds only to `a`. The code is correct as written; the next reader is
// the problem.
//
// Five operators, taken from upstream's set verbatim: `=`, `==`, `===`, `in`, `instanceof`. `!=` and
// `!==` are deliberately absent, which matters more than it looks: `a! != b` is SILENT upstream,
// measured, even though it is the shape a naive reading of "confusing bang" would flag first. There
// is nothing to confuse it with, because the assertion and the operator are already spelled apart.
//
// # Three messages, and the operator is interpolated into one of them
//
// `=` reports confusingAssign, `==` and `===` report confusingEqual, and `in` and `instanceof`
// report confusingOperator, whose text names the operator twice. That interpolation is the only
// per-finding text this rule produces, and it is asserted by equality in the fixtures rather than by
// a substring test, because a substring test cannot see a wrong operator spliced into the slot.
//
// # Suggestions, never fixes, which is upstream's decision and worth keeping
//
// Every repair here changes what the code MEANS. Removing the `!` drops an assertion the author
// wrote deliberately, and wrapping the left operand in parentheses changes precedence. Upstream
// declares `hasSuggestions` and no `fixable` for exactly that reason, so nothing is applied
// unattended, and this port carries the same distinction. A non-null assertion on the left offers
// removal; anything else offers only the wrap, because there is no `!` node to remove.
//
// `in` and `instanceof` offer BOTH suggestions, in upstream's order: remove first, wrap second.
//
// # How the token tests become a text test, and why that is not a shortcut
//
// Upstream works at the token level. It asks whether the last token of the left operand is a `!`
// punctuator, and separately whether the token after the left operand is a `)`. Neither question has
// a direct answer in our tree, and the structural equivalent turns out to be one test rather than
// two: whether the left operand's TRIMMED text ends in `!`.
//
// That reproduces both halves at once, which is the part worth recording:
//
// \ta! == b         left text "a!"        ends in bang   reports, as upstream
// \ta + b! == c     left text "a + b!"    ends in bang   reports, as upstream. The left is a
// \t                                                     binary expression rather than an
// \t                                                     assertion, which is why upstream looks at
// \t                                                     the token rather than the node kind, and
// \t                                                     why the text test has to as well.
// \t(a + b!) == c   left text "(a + b!)"  ends in paren  SILENT, and this is upstream's
// \t                                                     `tokenAfterLeft !== ')'` test arriving as
// \t                                                     a consequence rather than as a second
// \t                                                     check. Our parser keeps the parenthesized
// \t                                                     node, so its text carries the closing
// \t                                                     paren and the same test declines it.
// \ta !in b         left text "a !"       ends in bang   reports, as upstream. Whitespace inside
// \t                                                     the operand is irrelevant to both.
//
// The obvious worry is a `!` that is the last CHARACTER without being the last TOKEN. Probed, and
// the tree does not produce one: a bang inside a string, a template, or a computed key leaves the
// closing quote or bracket as the final character, so `'a!' == b`, "`a!` == b" and "x[`!`] == b" all
// end in something else and are silent here exactly as they are upstream. A trailing comment is
// excluded because the range is the TRIMMED token range rather than `Pos()` to `End()`.
//
// `a!!== b` is the one shape where the two tests could disagree and they do not: it parses as `a!`
// with the operator `!==`, so the text ends in a bang and the operator is not in the set. Silent
// both here and upstream, measured, and decided by the operator test rather than the text one.
//
// # One listener where upstream has two selectors
//
// Upstream matches `BinaryExpression, AssignmentExpression` because estree splits them. Our parser
// gives `a! = b` a KindBinaryExpression whose operator is KindEqualsToken, so a single listener
// covers both and the `=` arm is reachable from it. Probed rather than assumed.
var NoConfusingNonNullAssertion = rule.Rule{
	Name: "@typescript-eslint/no-confusing-non-null-assertion",

	// Purely syntactic. The judgment is which operator follows an assertion, and no part of it asks
	// what anything's type is.
	NeedsTypeChecker: false,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binaryExpression := node.AsBinaryExpression()
				if binaryExpression.OperatorToken == nil || binaryExpression.Left == nil {
					return
				}

				operatorText, isConfusing := confusingOperatorText(binaryExpression.OperatorToken.Kind)
				if !isConfusing {
					return
				}

				leftRange := rule.TokenRange(ctx.SourceFile, binaryExpression.Left)
				sourceText := ctx.SourceFile.Text()
				if leftRange.Pos() < 0 || leftRange.End() > len(sourceText) ||
					leftRange.End() <= leftRange.Pos() {
					return
				}
				leftText := sourceText[leftRange.Pos():leftRange.End()]
				if leftText[len(leftText)-1] != '!' {
					return
				}

				// The wrap suggestion is offered for every reported shape, so it is built once.
				wrapSuggestion := rule.Suggestion{
					Message: rule.Message{
						Id: "wrapUpLeft",
						Description: "Wrap the left-hand side in parentheses to avoid confusion " +
							"with \"" + operatorText + "\" operator.",
					},
					Fixes: []rule.Fix{
						rule.ReplaceRange(core.NewTextRange(leftRange.Pos(), leftRange.Pos()), "("),
						rule.ReplaceRange(core.NewTextRange(leftRange.End(), leftRange.End()), ")"),
					},
				}

				// Upstream branches on whether the left operand IS a non-null assertion, rather
				// than merely ending in one. `a + b! == c` ends in a bang and its left is a binary
				// expression, so there is no `!` node whose removal upstream could describe, and it
				// offers only the wrap. Removing the trailing character there would delete the bang
				// out of `b!`, which is a different edit than the one the message names.
				if binaryExpression.Left.Kind != ast.KindNonNullExpression {
					ctx.ReportNodeWithSuggestions(node, buildConfusingMessage(operatorText),
						wrapSuggestion)
					return
				}

				// The bang is the last character of the left operand, so its range is that one
				// byte. Taken from the trimmed range rather than from the node's own End(), which
				// is the same position here but would stop being so if a trailing comment were
				// ever included.
				bangRange := core.NewTextRange(leftRange.End()-1, leftRange.End())

				switch operatorText {
				case "=":
					ctx.ReportNodeWithSuggestions(node, buildConfusingMessage(operatorText),
						rule.Suggestion{
							Message: rule.Message{
								Id: "notNeedInAssign",
								Description: "Remove unnecessary non-null assertion (!) in " +
									"assignment left-hand side.",
							},
							Fixes: []rule.Fix{rule.RemoveRange(bangRange)},
						})
				case "==", "===":
					ctx.ReportNodeWithSuggestions(node, buildConfusingMessage(operatorText),
						rule.Suggestion{
							Message: rule.Message{
								Id: "notNeedInEqualTest",
								Description: "Remove unnecessary non-null assertion (!) in " +
									"equality test.",
							},
							Fixes: []rule.Fix{rule.RemoveRange(bangRange)},
						})
				default:
					// `in` and `instanceof`, which offer both repairs in upstream's order.
					ctx.ReportNodeWithSuggestions(node, buildConfusingMessage(operatorText),
						rule.Suggestion{
							Message: rule.Message{
								Id: "notNeedInOperator",
								Description: "Remove possibly unnecessary non-null assertion (!) " +
									"in the left operand of the `" + operatorText + "` operator.",
							},
							Fixes: []rule.Fix{rule.RemoveRange(bangRange)},
						}, wrapSuggestion)
				}
			},
		}
	},
}

// confusingOperatorText answers upstream's `confusingOperators` set, returning the text its messages
// quote.
//
// Five members and no more. The absent ones are the point: `!=` and `!==` already spell the
// assertion apart from the operator, so there is nothing to misread, and `<` and the compound
// assignments are not confusable with anything. Each of those was measured silent upstream rather
// than inferred from the set's shape.
func confusingOperatorText(operator ast.Kind) (operatorText string, isConfusing bool) {
	switch operator {
	case ast.KindEqualsToken:
		return "=", true
	case ast.KindEqualsEqualsToken:
		return "==", true
	case ast.KindEqualsEqualsEqualsToken:
		return "===", true
	case ast.KindInKeyword:
		return "in", true
	case ast.KindInstanceOfKeyword:
		return "instanceof", true
	}
	return "", false
}

// buildConfusingMessage picks the message for an operator, reproducing upstream's three-way split.
//
// The `in` and `instanceof` text names the operator twice, which is upstream's only interpolation in
// this rule and the reason the fixtures assert rendered text by equality.
func buildConfusingMessage(operatorText string) rule.Message {
	switch operatorText {
	case "=":
		return rule.Message{
			Id: "confusingAssign",
			Description: "Confusing combination of non-null assertion and assignment like " +
				"`a! = b`, which looks very similar to `a != b`.",
		}
	case "==", "===":
		return rule.Message{
			Id: "confusingEqual",
			Description: "Confusing combination of non-null assertion and equality test like " +
				"`a! == b`, which looks very similar to `a !== b`.",
		}
	}
	return rule.Message{
		Id: "confusingOperator",
		Description: "Confusing combination of non-null assertion and `" + operatorText +
			"` operator like `a! " + operatorText + " b`, which might be misinterpreted as " +
			"`!(a " + operatorText + " b)`.",
	}
}
