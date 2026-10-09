package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

func messageInvalidTypeofValue(found string) rule.Message {
	return rule.Message{
		Id: "invalidTypeofValue",
		Description: "A typeof expression can only produce one of eight strings, and \"" + found +
			"\" is not among them, so this comparison is always false and the branch behind it is " +
			"dead. The eight are undefined, object, boolean, number, string, function, symbol and " +
			"bigint. Almost every finding here is a typo or a guess at a type name that the " +
			"language does not report: typeof null is \"object\", an array is \"object\", and " +
			"there is no \"array\", \"null\" or \"date\".",
	}
}

// messageInvalidTypeofNonString shares the id of the string case, as upstream reports both as
// invalidValue, and says why: the operand is not a string at all, so no result of typeof can equal it.
func messageInvalidTypeofNonString(found string) rule.Message {
	return rule.Message{
		Id: "invalidTypeofValue",
		Description: "A typeof expression always produces a string, and `" + found + "` is not one, so " +
			"this comparison is always false (always true for != and !==) and the branch behind it is " +
			"dead or unconditional. Compare against the string instead: `typeof value === \"undefined\"`, " +
			"not `typeof value === undefined`.",
	}
}

var messageTypeofNotString = rule.Message{
	Id: "notString",
	Description: "Typeof comparisons should be to string literals. A literal is the one operand the " +
		"rule can check against the eight strings typeof produces, so with requireStringLiterals on, " +
		"anything else, a variable or a template with substitutions, is reported rather than trusted.",
}

var messageSuggestTypeofString = rule.Message{
	Id:          "suggestString",
	Description: "Use `\"undefined\"` instead of `undefined`.",
}

// ValidTypeofOptions mirrors upstream's one option object.
//
// RequireStringLiterals reports a typeof compared to anything but a string literal or another typeof.
// Default false, as upstream's: a binding holding a type name is a legitimate pattern, and only the
// stricter reading turns it into a finding.
type ValidTypeofOptions struct {
	RequireStringLiterals bool `json:"requireStringLiterals"`
}

// typeofResults are the only strings a typeof expression can produce.
var typeofResults = map[string]bool{
	"undefined": true, "object": true, "boolean": true, "number": true,
	"string": true, "function": true, "symbol": true, "bigint": true,
}

// ValidTypeof flags a typeof comparison against a value it can never equal.
//
//	valid:   typeof value === 'string'
//	valid:   typeof value === typeof other
//	valid:   typeof value === someVariable
//	invalid: typeof value === 'strng'
//	invalid: typeof value === 'array'
//	invalid: typeof value === 'null'
//	invalid: typeof value === undefined
//	invalid: typeof value === null
//
// Like use-isnan, this catches a check that cannot succeed rather than one that is merely unusual,
// so a finding is nearly always a real defect. The three most common are a misspelling, `'array'`
// (an array reports as `'object'`), and `'null'` (null also reports as `'object'`, which is the
// language's oldest wart). Any literal that is not a string, `null`, a number, a boolean, is dead
// the same way, as is the global `undefined`, the value where the string was meant.
//
// Only literal comparisons are judged by default. `typeof a === typeof b` and `typeof a === someName`
// are legitimate and unanalyzable without the checker, so they are left alone. requireStringLiterals
// reports every operand but a literal and another typeof, as upstream's notString.
//
// The global `undefined` carries upstream's one suggestion, quoting it. It is a suggestion and not a
// fix because it turns a branch that never runs into one that does. A local named `undefined` is not
// the global and is judged as any other binding, through identifierIsShadowed as use-isnan asks it.
// Nothing else gets a repair: the rule knows the comparison is dead and cannot know which of the
// eight was meant, and the two most common cases want a different check entirely, `'array'` wants
// Array.isArray and `'null'` wants `value === null`.
var ValidTypeof = rule.Rule{
	Name: "valid-typeof",
	// Asked only whether an `undefined` operand is declared in this file, so its findings key on the
	// closure's shapes. A typeof compared to undefined is rare, so the checker is consulted rarely.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, _ := rule.OptionsAs[ValidTypeofOptions](options)

		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary.OperatorToken == nil {
					return
				}
				switch binary.OperatorToken.Kind {
				case ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken,
					ast.KindExclamationEqualsToken, ast.KindExclamationEqualsEqualsToken:
				default:
					return
				}

				// Either side may hold the typeof, so the other side is the one to judge. Upstream
				// visits each typeof, so `typeof a === typeof b` is two visits that each find a typeof
				// on the other side and stay silent, which the left-first switch reaches in one.
				switch {
				case isTypeofExpression(binary.Left):
					reportUnreachableTypeofComparison(ctx, binary.Right, settings)
				case isTypeofExpression(binary.Right):
					reportUnreachableTypeofComparison(ctx, binary.Left, settings)
				}
			},
		}
	},
}

// reportUnreachableTypeofComparison judges the operand compared with a typeof, in upstream's order: a
// literal against the eight strings, then the global undefined, then, under requireStringLiterals,
// anything that is not itself a typeof.
func reportUnreachableTypeofComparison(ctx rule.Context, operand *ast.Node, settings ValidTypeofOptions) {
	operand = unwrapParentheses(operand)
	if operand == nil {
		return
	}

	switch operand.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		// A template with no substitutions is a string literal wearing backticks, and the same
		// typo is just as dead in that spelling.
		if !typeofResults[operand.Text()] {
			ctx.ReportNode(operand, messageInvalidTypeofValue(operand.Text()))
		}
		return

	case ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindRegularExpressionLiteral,
		ast.KindNullKeyword, ast.KindTrueKeyword, ast.KindFalseKeyword:
		// The rest of ESTree's Literal: no string, so no result of typeof equals it.
		ctx.ReportNode(operand, messageInvalidTypeofNonString(ctx.NodeText(operand)))
		return

	case ast.KindIdentifier:
		if operand.Text() == "undefined" && !identifierIsShadowed(ctx, operand) {
			message := messageInvalidTypeofNonString("undefined")
			if settings.RequireStringLiterals {
				message = messageTypeofNotString
			}
			ctx.ReportNodeWithSuggestions(operand, message, rule.Suggestion{
				Message: messageSuggestTypeofString,
				Fixes:   []rule.Fix{ctx.ReplaceNode(operand, `"undefined"`)},
			})
			return
		}
	}

	if settings.RequireStringLiterals && operand.Kind != ast.KindTypeOfExpression {
		ctx.ReportNode(operand, messageTypeofNotString)
	}
}

// isTypeofExpression reports a typeof operator, looking through parentheses.
func isTypeofExpression(node *ast.Node) bool {
	node = unwrapParentheses(node)
	return node != nil && node.Kind == ast.KindTypeOfExpression
}

// unwrapParentheses peels grouping, which changes nothing about what an expression is.
func unwrapParentheses(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}
