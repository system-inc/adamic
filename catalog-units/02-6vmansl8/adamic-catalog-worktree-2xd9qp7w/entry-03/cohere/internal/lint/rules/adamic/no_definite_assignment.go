package adamic

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// noDefiniteAssignmentText is the rule's message, whose wording lives in
// `policy/messages/no-definite-assignment.json`.
var noDefiniteAssignmentText = policy.MessageOf("adamic/no-definite-assignment", "definiteAssignment")

/*
 * NoDefiniteAssignment reports a definite assignment assertion, `field!: T` or `let x!: T` (#drbrp8c).
 *
 *     invalid: class Account { balance!: number }     read before it is set, it is undefined
 *     invalid: let total!: number;
 *     valid:   class Account { balance = 0 }
 *     valid:   class Account { balance: number | undefined }
 *
 * # The hole
 *
 * The `!` tells tsc the binding is assigned before it is read and turns off the check that would say
 * otherwise. Nothing verifies it: `new Account().describe()` reads `this.balance.toFixed(2)` on undefined
 * (probe h09 on #drbrp8c, accepted by tsc 6.0.3 under strict, a TypeError on Node). Adamic 0.1 refuses it
 * with the non-null `!` it resembles; no-non-null-assertion reports only the expression form, `value!`,
 * so this is the declaration form's rule.
 *
 * The finding sits on the declared name, which is what the `!` makes a claim about.
 *
 * # No fix
 *
 * Removing the `!` leaves a compile error until the author assigns the binding or widens its type.
 */
var NoDefiniteAssignment = rule.Rule{
	Name: "adamic/no-definite-assignment",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		report := func(name *ast.Node) {
			if name != nil {
				ctx.ReportNode(name, rule.Message{
					Id:          "definiteAssignment",
					Description: noDefiniteAssignmentText.Render(nil),
				})
			}
		}
		return rule.Listeners{
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				if postfix := node.PostfixToken(); postfix != nil && postfix.Kind == ast.KindExclamationToken {
					report(node.Name())
				}
			},
			ast.KindVariableDeclaration: func(node *ast.Node) {
				if node.AsVariableDeclaration().ExclamationToken != nil {
					report(node.Name())
				}
			},
		}
	},
}
