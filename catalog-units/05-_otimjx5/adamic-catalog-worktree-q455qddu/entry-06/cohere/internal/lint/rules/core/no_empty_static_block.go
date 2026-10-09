package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnexpectedEmptyStaticBlock = rule.Message{
	Id: "unexpectedEmptyStaticBlock",
	Description: "This static initialization block is empty, so it runs nothing at class " +
		"definition time and exists only as syntax. Unlike an empty catch, there is no reading " +
		"where that is deliberate: a static block has no parameter to bind and no error to " +
		"swallow, so an empty one is a body that was deleted or never written. Remove it, or say " +
		"in a comment why the class needs an initialization step that does nothing.",
}

// NoEmptyStaticBlock flags an empty static initialization block in a class.
//
//	valid:   class A { static { this.x = 1; } }
//	valid:   class A { static { /* nothing to do yet, see #123 */ } }
//	invalid: class A { static {} }
//
// A comment inside the block satisfies the rule, matching how `no-empty` treats its blocks: the
// point is that a reader can tell a deliberate no-op from an accident, and a comment is what makes
// that distinction visible. Consistency with the sibling rule matters more than the marginal
// strictness of rejecting the comment form.
//
// No fix. Removing the block is usually right and occasionally wrong, since a static block is also
// where someone parks work they are about to write, and this rule cannot tell which case it found.
var NoEmptyStaticBlock = rule.Rule{
	Name: "no-empty-static-block",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindClassStaticBlockDeclaration: func(node *ast.Node) {
				body := node.AsClassStaticBlockDeclaration().Body
				if body == nil || body.Kind != ast.KindBlock {
					return
				}
				if len(body.AsBlock().Statements.Nodes) > 0 {
					return
				}
				// An empty block that holds a comment is a deliberate no-op the author explained,
				// which is the same allowance no-empty makes. Reusing that rule's helper rather
				// than writing a second one: it reads the parser's own trivia, so it is immune to
				// a comment marker living inside a string earlier in the node, and two helpers
				// that almost agree would drift.
				if blockContainsComment(ctx.SourceFile, body) {
					return
				}
				ctx.ReportNode(node, messageUnexpectedEmptyStaticBlock)
			},
		}
	},
}
