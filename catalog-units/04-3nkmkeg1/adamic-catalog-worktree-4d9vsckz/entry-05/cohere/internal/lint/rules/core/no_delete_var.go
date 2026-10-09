package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnexpectedDeleteVar = rule.Message{
	Id: "unexpectedDeleteVar",
	Description: "`delete` removes a property from an object; it cannot remove a variable. " +
		"Applied to a bare name this does nothing in strict mode and throws a SyntaxError where " +
		"strict mode is enforced, which every module in this codebase is. So the line either has " +
		"no effect or breaks the file, and neither is what it reads as doing. To drop a " +
		"reference, assign undefined; to remove a property, delete the property.",
}

// NoDeleteVar flags `delete` applied to a bare identifier.
//
//	valid:   delete object.property
//	valid:   delete object['property']
//	valid:   delete object[key]
//	invalid: delete value
//	invalid: delete (value)
//
// The rule is narrow on purpose: only a bare name is reported, because only a bare name is a
// variable. Every other target of `delete` is a property access, which is what the operator is for.
//
// Parentheses are stripped, since `delete (value)` is the same operation with the same target and
// ESTree has no node for grouping. That is the shape a formatter can introduce, so a rule that
// missed it would report the same code differently before and after formatting.
//
// No fix. The two repairs are opposite in meaning: assigning undefined keeps the binding and clears
// the value, while deleting a property removes the key. The rule cannot know which the author meant,
// and guessing wrong changes what the code does rather than how it reads.
var NoDeleteVar = rule.Rule{
	Name: "no-delete-var",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindDeleteExpression: func(node *ast.Node) {
				target := ast.SkipParentheses(node.AsDeleteExpression().Expression)
				if target == nil || target.Kind != ast.KindIdentifier {
					return
				}
				ctx.ReportNode(node, messageUnexpectedDeleteVar)
			},
		}
	},
}
