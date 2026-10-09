package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// forbiddenArgumentsProperties are the two properties of `arguments` this rule refuses.
//
// A set rather than two comparisons because the property name is quoted back in the message, so the
// lookup answers both questions at once: whether to report, and what to call it.
var forbiddenArgumentsProperties = map[string]bool{
	"callee": true,
	"caller": true,
}

// NoCaller flags `arguments.callee` and `arguments.caller`.
//
//	valid:   var x = arguments.length
//	valid:   var x = arguments
//	valid:   var x = arguments[0]
//	valid:   var x = arguments[caller]
//	invalid: var x = arguments.callee
//	invalid: var x = arguments.caller
//
// Both properties throw in strict mode, and every module and class body is strict, so code reaching
// for them either already fails or is running in the shrinking part of the language where it does
// not. They also defeat the optimizations engines apply to functions whose callers are statically
// known, which is why the restriction exists rather than being merely stylistic.
//
// # The subscript form is not the same question
//
// `arguments[caller]` reads a *variable* named caller and gets whatever property that names, which
// is a different access and may not be either of these. Upstream matches the static member form
// only, and the fourth clean case pins that: a rule reading the source text, or one treating a
// computed access as equivalent, reports code that does nothing of the kind.
//
// `arguments[0]` is the same shape and is ordinary.
//
// # No fix, and none is possible
//
// There is no mechanical replacement. `arguments.callee` is usually repaired by naming the function
// and referring to it, which requires choosing a name, and `arguments.caller` has no replacement at
// all: the information it carried is not available another way. A rule offering a fix here would be
// offering to guess.
var NoCaller = rule.Rule{
	Name: "no-caller",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				access := node.AsPropertyAccessExpression()

				name := access.Name()
				if name == nil || !forbiddenArgumentsProperties[name.Text()] {
					return
				}
				if !isIdentifierNamed(access.Expression, "arguments") {
					return
				}

				// The property rather than the whole access, matching upstream, so the finding
				// points at `callee` rather than at `arguments.callee`. That also puts it where a
				// suppressing comment on the line above can reach it.
				ctx.ReportNode(name, rule.Message{
					Id: "noCaller",
					Description: fmt.Sprintf(
						"This uses `arguments.%s`. Both properties throw in strict mode, which "+
							"every module and class body already is, and they prevent engines from "+
							"optimizing functions whose callers would otherwise be statically "+
							"known. Name the function and refer to it by name instead.",
						name.Text()),
				})
			},
		}
	},
}
