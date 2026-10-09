package structure

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/ecmascript/scope"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// reactHookNoDestructuringText is the rule's message, whose wording lives in
// `policy/messages/react-hook-no-destructuring.json`.
var reactHookNoDestructuringText = policy.MessageOf("structure/react-hook-no-destructuring", "noDestructuringInHook")

// messageNoDestructuringInHook is the finding, rendered when it is reported so the text comes from the current catalog.
func messageNoDestructuringInHook() rule.Message {
	return rule.Message{Id: reactHookNoDestructuringText.Id, Description: reactHookNoDestructuringText.Render(nil)}
}

// ReactHookNoDestructuring flags object destructuring inside a custom hook's body.
//
//	valid:   function useThing() { const result = useQuery(); return result.data; }
//	invalid: function useThing() { const { data } = useQuery(); return data; }
//
// Only object destructuring, matching the original. Array destructuring is exempt and that is the
// convention rather than an oversight: `const [value, setValue] = React.useState()` has no property
// names to preserve, so there is nothing for the rule to protect.
//
// The enclosing-hook test walks every enclosing function rather than only the nearest, so a
// destructure inside a callback inside a hook still counts. That is the original's behavior and it
// is what makes the rule hold: a hook's callbacks are part of the hook.
//
// The original gates on `isReactFile || isTypeScriptFile`, which together mean any .ts or .tsx file
// and so exclude nothing this linter would ever see. Reproduced as no gate at all rather than as
// two flags whose disjunction is always true, since a guard that cannot fail reads as a guard that
// might.
var ReactHookNoDestructuring = rule.Rule{
	Name: "structure/react-hook-no-destructuring",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()

				name := declaration.Name()
				if name == nil || name.Kind != ast.KindObjectBindingPattern {
					return
				}

				// No initializer, no report, and this is load-bearing rather than defensive.
				//
				// It is what exempts `for(const { data } of items)`, whose declaration has no
				// initializer because the loop supplies the value. That is a real destructure
				// inside a hook and the original does not report it either: its check is on a
				// VariableDeclarator's `init`, which ESTree leaves null in a for-of head.
				//
				// Reproduced rather than improved on. Reporting it would be defensible and would
				// put us one finding ahead of the gate on a shape nobody has decided about, which
				// is a different thing from the one deliberate divergence already ruled on.
				// Recorded here so a future reader knows it was seen rather than missed.
				if declaration.Initializer == nil {
					return
				}

				if !isInsideCustomHook(node) {
					return
				}

				ctx.ReportNode(name, messageNoDestructuringInHook())
			},
		}
	},
}

// isInsideCustomHook reports whether any enclosing function is named like a hook.
//
// Every enclosing function is tested rather than only the nearest, so a destructure inside a
// callback inside a hook counts. The nearest function is often an anonymous callback with no name
// to test at all, and stopping there would silence the rule wherever it matters most.
func isInsideCustomHook(node *ast.Node) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		name := scope.NameOf(current)
		if name != "" && react.IsHookName(name) {
			return true
		}
	}
	return false
}
