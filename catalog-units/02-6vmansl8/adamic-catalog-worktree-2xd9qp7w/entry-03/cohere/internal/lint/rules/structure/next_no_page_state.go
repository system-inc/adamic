package structure

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// nextNoPageStateText is the rule's message, whose wording lives in
// `policy/messages/next-no-page-state.json`.
var nextNoPageStateText = policy.MessageOf("structure/next-no-page-state", "pageStateRemounts")

// messagePageStateRemounts is the finding, rendered when it is reported so the text comes from the current catalog.
func messagePageStateRemounts() rule.Message {
	return rule.Message{Id: nextNoPageStateText.Id, Description: nextNoPageStateText.Render(nil)}
}

// pageStateHookNames are the hooks that hold state across renders.
//
// Only these two. `useRef` also survives a render and is deliberately absent: a ref is already
// understood to be a mutable box rather than state, and flagging it would report every measured
// element and every timer handle on the page.
var pageStateHookNames = map[string]bool{
	"useState":   true,
	"useReducer": true,
}

// NextNoPageState flags useState or useReducer called directly in a Next.js page file.
//
//	valid:   state held in a provider in layout.tsx
//	valid:   React.useState inside a component the page imports
//	invalid: React.useState in page.tsx
//	invalid: useState in page.tsx
//
// The failure this catches is invisible rather than loud, which is why a lint rule is the only
// thing that can catch it: the code is correct React, it compiles, it works on first load, and it
// silently loses the value on a navigation the developer did not think to test.
//
// Both spellings are matched. `React.useState` is what this codebase writes, since
// `react-import-no-destructuring` forbids the bare import, but the bare form is checked too because
// a page that violated that rule would otherwise slip past this one as well, and two rules failing
// together is exactly when a page most needs the finding.
//
// # What a clean run on this tree proves, which is nothing
//
// The tree contains zero occurrences of either hook in any page.tsx, so this rule reporting zero is
// not evidence that it works. The shape is absent, and a rule that never runs and a rule that runs
// correctly produce the same output here.
//
// Recorded rather than left implicit, because every other rule ported tonight had a real population
// behind its zero and this one does not. The fixtures and a planted control through the real binary
// are the whole of the evidence for it.
var NextNoPageState = rule.Rule{
	Name: "structure/next-no-page-state",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// IsReactFile is redundant here and kept for symmetry with the original, which tests both.
		// IsPageFile already requires a .tsx or .jsx suffix, so no file can satisfy it and fail the
		// React test. A mutation sweep cannot kill the first half, which is what a check subsumed
		// by the one beside it looks like from inside a green suite.
		fileContext := FileContextFor(ctx.SourceFile.FileName().AsString())
		if !fileContext.IsReactFile || !fileContext.IsPageFile {
			return nil
		}

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callee := ast.SkipParentheses(node.AsCallExpression().Expression)
				if callee == nil {
					return
				}

				switch callee.Kind {
				case ast.KindIdentifier:
					if pageStateHookNames[callee.Text()] {
						ctx.ReportNode(callee, messagePageStateRemounts())
					}

				case ast.KindPropertyAccessExpression:
					// Dotted access only. `React['useState'](1)` is not matched, following the
					// original, whose computed guard excludes it. Nobody writes a hook call that
					// way, and matching it would mean deciding what a computed key resolves to.
					//
					// The receiver must be React. `Store.useState` is somebody's own accessor and
					// has nothing to do with the App Router's remounting.
					if react.IsNamespacedMember(callee, func(name string) bool { return pageStateHookNames[name] }) {
						ctx.ReportNode(callee, messagePageStateRemounts())
					}
				}
			},
		}
	},
}
