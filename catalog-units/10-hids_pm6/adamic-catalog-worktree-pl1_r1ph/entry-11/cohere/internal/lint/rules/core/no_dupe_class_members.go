package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/classmembers"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var NoDupeClassMembers = rule.Rule{
	Name: "no-dupe-class-members",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// The collision judgment lives in `classmembers` rather than here, because
		// `@typescript-eslint/no-dupe-class-members` needs the same answer and a rule package may
		// not import another rule package. This rule owns the wording and the message id; the
		// utility owns which members collide.
		check := func(members *ast.NodeList) {
			classmembers.ForEachDuplicate(members, func(name *ast.Node, key classmembers.Key) {
				ctx.ReportNode(name, rule.Message{
					Id: "noDupeClassMembers",
					Description: fmt.Sprintf(
						"This class already declares a member named %s. The later declaration wins "+
							"silently, so the earlier one is dead code that reads as live, and "+
							"nothing in the language or at runtime tells the two apart.", key.Name),
				})
			})
		}

		return rule.Listeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				check(node.AsClassDeclaration().Members)
			},
			ast.KindClassExpression: func(node *ast.Node) {
				check(node.AsClassExpression().Members)
			},
		}
	},
}
