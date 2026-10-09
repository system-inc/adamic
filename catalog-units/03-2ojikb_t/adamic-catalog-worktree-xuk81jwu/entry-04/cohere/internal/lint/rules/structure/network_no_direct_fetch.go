package structure

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// networkNoDirectFetchText is the rule's message, whose wording lives in
// `policy/messages/network-no-direct-fetch.json`.
var networkNoDirectFetchText = policy.MessageOf("structure/network-no-direct-fetch", "noDirectFetch")

// messageNoDirectFetch is the finding, rendered when it is reported so the text comes from the current catalog.
func messageNoDirectFetch() rule.Message {
	return rule.Message{Id: networkNoDirectFetchText.Id, Description: networkNoDirectFetchText.Render(nil)}
}

// NetworkNoDirectFetch flags a direct call to fetch.
//
//	valid:   networkService.request(url)
//	valid:   fetch(url)                     (inside NetworkService.ts only)
//	invalid: fetch(url)
//	invalid: window.fetch(url)
//	invalid: globalThis.fetch(url)
//
// Three spellings reach the same primitive, enumerated before the listener rather than after: the
// bare call, and the two global objects it hangs off. A rule catching only the bare form would be
// trivially worked around by writing `window.fetch`, which is not a workaround anyone would think of
// deliberately but is exactly what a copied snippet contains.
//
// The exemption is by path rather than by an inline comment, because the file that implements the
// wrapper is a fixed and knowable place rather than a judgment call, and a rule that forbids the
// thing it asks people to use is a rule people learn to disable.
//
// No fix. Replacing a fetch with a NetworkService call means choosing a method, a response shape,
// and an error path, which is the work the rule is asking for rather than something to generate.
var NetworkNoDirectFetch = rule.Rule{
	Name: "structure/network-no-direct-fetch",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// The implementation is the one place the raw primitive is allowed, so it declines the file
		// before any node is visited.
		if FileContextFor(ctx.SourceFile.FileName().AsString()).IsNetworkServiceFile {
			return nil
		}

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if isFetchCallee(node.AsCallExpression().Expression) {
					ctx.ReportNode(node, messageNoDirectFetch())
				}
			},
		}
	},
}

// fetchGlobalObjects are the globals fetch can be reached through.
var fetchGlobalObjects = map[string]bool{"window": true, "globalThis": true}

// isFetchCallee reports whether a callee names fetch, bare or through a known global.
//
// Only the two documented globals count. A `service.fetch(...)` is somebody's own method that
// happens to share the name, and reading every `.fetch` as the primitive would flag the wrapper
// calls this rule exists to encourage.
func isFetchCallee(callee *ast.Node) bool {
	callee = ast.SkipParentheses(callee)
	if callee == nil {
		return false
	}

	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text() == "fetch"

	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		name := access.Name()
		if name == nil || name.Text() != "fetch" {
			return false
		}
		object := ast.SkipParentheses(access.Expression)
		return object != nil && object.Kind == ast.KindIdentifier && fetchGlobalObjects[object.Text()]
	}
	return false
}
