package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnexpectedVar = rule.Message{
	Id: "unexpectedVar",
	Description: "This is a `var` declaration, which is function-scoped and hoisted rather than " +
		"block-scoped. The binding exists from the top of the enclosing function, so it is " +
		"readable above the line that declares it and it leaks out of the block it looks like it " +
		"belongs to: a `var` inside an `if` is still in scope after the `if` ends. A `var` in a " +
		"loop is one binding shared by every iteration, which is why a closure created in the " +
		"loop sees the final value rather than its own. Use `let`, or `const` when nothing " +
		"reassigns it.",
}

// NoVar flags a var declaration, wherever one can appear.
//
//	valid:   const value = 1
//	valid:   let value = 1
//	valid:   for(let index = 0; index < 10; index++) {}
//	valid:   declare var globalThing: number
//	invalid: var value = 1
//	invalid: for(var index = 0; index < 10; index++) {}
//	invalid: for(var item of list) {}
//
// The reason this rule matters more than its findings suggest is that the two scoping rules differ
// only in cases that are already confusing: the same code usually behaves identically under both and
// diverges exactly where a reader is least likely to be checking. So a `var` is not a defect most of
// the time, and is a defect in precisely the places nobody looks.
//
// Four shapes, not one. A `var` reaches the tree either as a statement or as the initializer of a
// `for`, `for-in`, or `for-of`, where the declaration list hangs off the loop rather than off a
// statement. Measured before this was written: a listener on variable statements alone caught one of
// the four, and the three it missed are the loop forms, which carry the sharpest consequence.
//
// Ambient declarations are exempt. A `declare var` names something that exists in a runtime this file
// does not control, so the keyword describes the outside world rather than making a scoping choice.
//
// No fix. Replacing `var` with `let` is correct only when the binding is never read before its
// declaration and never relied on outside its block, and both are questions about the whole enclosing
// function rather than about this line. Changing scope silently is how a lint fix turns a working
// program into a broken one.
var NoVar = rule.Rule{
	Name: "no-var",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// reportIfVar reports when a declaration list declares with `var`.
		//
		// let and const both carry a flag, and `var` is the absence of both. Reading it as an
		// absence rather than as a keyword is what makes `using` and `await using` fall through
		// correctly, since those carry their own flags.
		reportIfVar := func(declarationList *ast.Node) {
			if declarationList == nil {
				return
			}
			blockScoped := ast.NodeFlagsLet | ast.NodeFlagsConst |
				ast.NodeFlagsUsing | ast.NodeFlagsAwaitUsing
			if declarationList.Flags&blockScoped != 0 {
				return
			}
			ctx.ReportNode(declarationList, messageUnexpectedVar)
		}

		// loopInitializer returns a loop's initializer when it declares bindings rather than
		// assigning to existing ones, since `for(index = 0; ...)` declares nothing.
		loopInitializer := func(initializer *ast.Node) *ast.Node {
			if initializer != nil && initializer.Kind == ast.KindVariableDeclarationList {
				return initializer
			}
			return nil
		}

		return rule.Listeners{
			ast.KindVariableStatement: func(node *ast.Node) {
				// An ambient declaration describes a runtime this file does not control.
				if hasDeclareModifier(node) {
					return
				}
				// Directly inside `declare global { }`, only `var` types a property of `globalThis`:
				// `let` and `const` there declare a global the runtime never puts on the object. ESLint
				// skips exactly this shape and nothing wider (a `declare module` or `declare namespace`
				// block, or a namespace nested inside the global block, still reports), so the check is
				// the immediate parent, not any ambient ancestor. Found as a false positive on Base's
				// GraphQlMetadataStorage.ts (#hzhs9f8).
				if parent := node.Parent; parent != nil && parent.Kind == ast.KindModuleBlock &&
					parent.Parent != nil && ast.IsGlobalScopeAugmentation(parent.Parent) {
					return
				}
				reportIfVar(node.AsVariableStatement().DeclarationList)
			},

			ast.KindForStatement: func(node *ast.Node) {
				reportIfVar(loopInitializer(node.AsForStatement().Initializer))
			},

			ast.KindForInStatement: func(node *ast.Node) {
				reportIfVar(loopInitializer(node.AsForInOrOfStatement().Initializer))
			},

			ast.KindForOfStatement: func(node *ast.Node) {
				reportIfVar(loopInitializer(node.AsForInOrOfStatement().Initializer))
			},
		}
	},
}

// hasDeclareModifier reports the declare keyword on a statement.
func hasDeclareModifier(node *ast.Node) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindDeclareKeyword {
			return true
		}
	}
	return false
}
