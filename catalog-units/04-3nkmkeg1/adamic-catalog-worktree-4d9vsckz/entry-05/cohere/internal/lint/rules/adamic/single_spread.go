package adamic

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/checking/flow"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// singleSpreadText is the rule's message, whose wording lives in `policy/messages/single-spread.json`.
var singleSpreadText = policy.MessageOf("adamic/single-spread", "laterSpread")

/*
 * SingleSpread allows an object literal one spread, as its first member (#drbrp8c).
 *
 *     invalid: { ...a, ...b }                 b may carry an x of any type, copied over a's
 *     invalid: { x: 1, ...b }                 the same, over the field
 *     valid:   { ...tree, left: newLeft }     fields after the spread always win
 *     valid:   { ...base, ...(flag ? { a: 1 } : {}) }   a literal carries exactly its type's keys
 *
 * # The hole
 *
 * A value's type lists the properties it promises, not every property it has: `{ y: number }` may be a
 * value with an `x: string` that width subtyping let through. A spread copies what the value has. So a
 * spread after anything else can overwrite a property with one its type never mentioned, and the result's
 * type still says what the earlier member said: `merged.x` typed `number` holds a string (probes h06 and
 * h07 on #drbrp8c, `merged.x.toFixed is not a function` on Node). One leading spread followed by fields is
 * safe, because a field always overwrites (h08). Adamic 0.1 allows exactly that form; its native objects
 * hold only their type's fields, so beyond the type hole the two backends would also disagree.
 *
 * # What is exempt
 *
 * A spread of an object literal, or of a conditional whose branches are object literals, carries exactly
 * the keys its type names, so it overwrites nothing unexpected. It is not reported, and it counts as a
 * field: a spread after it can still overwrite what it wrote.
 *
 * JSX spread attributes have the same hole and are left out: JSX is no part of Adamic 0.1.
 *
 * # No fix
 *
 * The repair is a different construction of the object, and only the author knows which keys it needs.
 */
var SingleSpread = rule.Rule{
	Name: "adamic/single-spread",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				// `({ x, ...rest } = point)` is a pattern: its rest collects, it copies over nothing.
				if flow.IsDestructuringTarget(node) {
					return
				}
				before := ""
				for _, member := range node.AsObjectLiteralExpression().Properties.Nodes {
					if member.Kind != ast.KindSpreadAssignment {
						if before == "" {
							before = "a field"
						}
						continue
					}
					if spreadsALiteral(member.Expression()) {
						if before == "" {
							before = "a spread object literal"
						}
						continue
					}
					if before != "" {
						ctx.ReportNode(member, rule.Message{
							Id:          "laterSpread",
							Description: singleSpreadText.Render(map[string]string{"what": before}),
						})
						continue
					}
					before = "another spread"
				}
			},
		}
	},
}

// spreadsALiteral is an object literal, or a conditional whose branches both are, through parentheses.
func spreadsALiteral(expression *ast.Node) bool {
	expression = ast.SkipParentheses(expression)
	switch expression.Kind {
	case ast.KindObjectLiteralExpression:
		return true
	case ast.KindConditionalExpression:
		conditional := expression.AsConditionalExpression()
		return spreadsALiteral(conditional.WhenTrue) && spreadsALiteral(conditional.WhenFalse)
	}
	return false
}
