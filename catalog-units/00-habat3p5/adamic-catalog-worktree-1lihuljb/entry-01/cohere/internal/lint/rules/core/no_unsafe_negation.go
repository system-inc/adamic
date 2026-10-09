package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// relationalOperatorText maps the operators this rule always checks to the text the message quotes.
//
// `in` and `instanceof` and nothing else. They are the two whose precedence sits below `!` while
// reading like they should sit above it, and they are what oxc's `is_relational` means, despite the
// name suggesting the `<`/`>` family. That family is `is_compare`, is a different question, and is
// off by default. Reading the predicates rather than inferring them from their names is the whole
// defence against implementing the option inverted.
var relationalOperatorText = map[ast.Kind]string{
	ast.KindInKeyword:         "in",
	ast.KindInstanceOfKeyword: "instanceof",
}

// orderingOperatorText maps the four ordering comparisons, checked only when the option is on.
var orderingOperatorText = map[ast.Kind]string{
	ast.KindLessThanToken:          "<",
	ast.KindLessThanEqualsToken:    "<=",
	ast.KindGreaterThanToken:       ">",
	ast.KindGreaterThanEqualsToken: ">=",
}

// NoUnsafeNegationOptions configures which operators are checked.
type NoUnsafeNegationOptions struct {
	// EnforceForOrderingRelations extends the rule to `<`, `<=`, `>`, and `>=`.
	//
	// Off by default, matching both upstreams. `!a < b` is the same precedence mistake as `!a in b`
	// and is a much rarer intent, but unlike `in` and `instanceof` it does have a legitimate
	// reading: comparing a boolean against a number coerces cleanly and someone may mean it. So the
	// default is the two operators where the mistake is close to certain, and the ordering family
	// is opt-in.
	EnforceForOrderingRelations bool `json:"enforceForOrderingRelations"`
}

// NoUnsafeNegation flags a `!` applied to the left operand of a relational operator.
//
//	valid:   !(key in object)
//	valid:   (!key) in object
//	valid:   a in b
//	valid:   !a === b                       // equality is not this rule's business
//	valid:   !a < b                         // unless enforceForOrderingRelations is on
//	invalid: !key in object
//	invalid: !obj instanceof Ctor
//	invalid: !a < b                         // with enforceForOrderingRelations on
//
// `!` has higher precedence than `in` and `instanceof`, so `!key in object` parses as
// `(!key) in object`: the negation runs first, producing `true` or `false`, and the `in` then asks
// whether the *string* `"true"` or `"false"` is a key of the object. That is a well-formed
// expression, it type-checks, and it is essentially never what anybody wrote it to mean. The author
// meant `!(key in object)`.
//
// Parentheses are the whole discrimination and this rule deliberately does not skip them.
//
// `(!a) in b` is clean and `!(a) in b` reports, which is the reverse of what the house reflex
// produces. `ast.SkipParentheses` is used in ten rules in this package, and the reasoning for it
// there does not transfer: those rules ask what an expression *is*, and parentheses do not change
// that. This rule asks what the author *said*, and parentheses are exactly how an author says it.
// Somebody who writes `(!a) in b` has stated the grouping the rule would otherwise warn they might
// not have intended, so warning them would be telling them their explicit answer is a mistake.
//
// Our AST gives that for free rather than needing a check: typescript-go keeps a
// ParenthesizedExpression node, so `(!a) in b` has a parenthesized left and falls out of the kind
// test with no paren logic anywhere in this file. Verified by probing the parser, because it is the
// kind of thing that is true of ESTree in the opposite direction (Espree has no paren node at all,
// which is why ESLint has to call `isParenthesised` explicitly).
//
// Reported on the left operand rather than on the whole comparison, matching upstream's label: the
// defect is the negation, the reader needs their eye on the `!`, and the operator they need to
// compare its precedence against is named in the message.
//
// Two suggestions, and no fix. This is a deliberate divergence from oxc, which declares `fix` and
// auto-applies the first of these.
//
// The rewrite `!a in b` to `!(a in b)` changes what the program does at runtime. That is the point
// of it, and it is almost certainly the change the author wanted, but "almost certainly" is the
// test a Fix fails: a Fix is applied unattended and must preserve meaning. The evidence that the
// intent is genuinely unrecoverable is that the other reading has its own repair, `(!a) in b`,
// which preserves behaviour exactly and merely makes it explicit. A rule that had to choose between
// them by guessing is a rule that should not be choosing, so both are offered and a human ranks
// them. ESLint reaches the same conclusion, declaring `fixable: null` with these same two
// suggestions, one labelled as changing behaviour and one as preserving it.
var NoUnsafeNegation = rule.Rule{
	Name: "no-unsafe-negation",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		enforceForOrderingRelations := false
		if parsed, ok := rule.OptionsAs[NoUnsafeNegationOptions](options); ok {
			enforceForOrderingRelations = parsed.EnforceForOrderingRelations
		}

		return rule.Listeners{
			// Listening on the comparison rather than on every `!` is what makes `!!a in b` report
			// once: only the outer negation is the left operand, and walking up from a unary would
			// have to rediscover that.
			ast.KindBinaryExpression: func(node *ast.Node) {
				binaryExpression := node.AsBinaryExpression()
				if binaryExpression == nil || binaryExpression.OperatorToken == nil ||
					binaryExpression.Left == nil || binaryExpression.Right == nil {
					return
				}

				operatorText, isChecked := relationalOperatorText[binaryExpression.OperatorToken.Kind]
				if !isChecked && enforceForOrderingRelations {
					operatorText, isChecked = orderingOperatorText[binaryExpression.OperatorToken.Kind]
				}
				if !isChecked {
					return
				}

				// Not SkipParentheses, and the doc comment above argues why at length.
				left := binaryExpression.Left
				if left.Kind != ast.KindPrefixUnaryExpression {
					return
				}
				prefixExpression := left.AsPrefixUnaryExpression()
				if prefixExpression == nil || prefixExpression.Operator != ast.KindExclamationToken ||
					prefixExpression.Operand == nil {
					return
				}

				// Sliced from the source rather than reconstructed from the nodes, because the
				// operand may be anything: `(y=>{if(!/s/ in(l)){}})` has a regex literal there, and
				// only the original text gives it back unchanged.
				//
				// Through rule.TokenRange rather than node.Pos, and that is not a detail. Pos sits
				// before leading trivia, so slicing an operand with it picks up the space in
				// `! a <= b` and the rewrite comes out as `!( a <= b)`. Same trimming ReplaceNode
				// does, for the same reason, and the same reason a comment before the right operand
				// would otherwise be moved inside the parentheses.
				sourceText := ctx.SourceFile.Text()
				operandRange := rule.TokenRange(ctx.SourceFile, prefixExpression.Operand)
				rightRange := rule.TokenRange(ctx.SourceFile, binaryExpression.Right)
				negationRange := rule.TokenRange(ctx.SourceFile, left)
				operandText := sourceText[operandRange.Pos():operandRange.End()]
				rightText := sourceText[rightRange.Pos():rightRange.End()]
				negationText := sourceText[negationRange.Pos():negationRange.End()]

				ctx.ReportNodeWithSuggestions(left, rule.Message{
					Id:          "unexpected",
					Description: describeUnsafeNegation(operatorText),
				}, rule.Suggestion{
					Message: rule.Message{
						Id: "negateTheComparison",
						Description: "Negate the whole `" + operatorText + "` comparison, which is " +
							"what the parentheses in `!(a " + operatorText + " b)` say and is " +
							"almost certainly the intent. This changes what the code does.",
					},
					Fixes: []rule.Fix{ctx.ReplaceNode(node,
						"!("+operandText+" "+operatorText+" "+rightText+")")},
				}, rule.Suggestion{
					Message: rule.Message{
						Id: "parenthesiseTheNegation",
						Description: "Wrap the negation in parentheses to say the `!` was meant to " +
							"apply to the left operand alone. This preserves what the code does " +
							"today and only makes it explicit.",
					},
					Fixes: []rule.Fix{ctx.ReplaceNode(left, "("+negationText+")")},
				})
			},
		}
	},
}

// describeUnsafeNegation names the operator in the message.
//
// The operator is the load-bearing half: a reader who is told the precedence is wrong still has to
// know which operator lost, and `in` and `<=` fail for the same reason but read very differently.
func describeUnsafeNegation(operatorText string) string {
	return "The `!` binds tighter than the `" + operatorText + "` after it, so this negates the " +
		"left operand and then compares the resulting boolean, rather than negating the " +
		"comparison. `!key " + operatorText + " object` runs the `!` first and hands `" +
		operatorText + "` a `true` or a `false`, which is almost never the question being asked, " +
		"and it is a valid expression either way so nothing else reports it."
}
