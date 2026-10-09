package next

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoSyncScripts = rule.Message{
	Id: "noSyncScripts",
	Description: "This <script> has a src and neither `async` nor `defer`, so the browser stops " +
		"parsing the document, fetches the script, and runs it before continuing. Everything below " +
		"it waits on the network. Add `defer` to run it after parsing, `async` if it does not " +
		"depend on the document, or use the `Script` component from `next/script`.",
}

// NoSyncScripts flags a <script src> written without `async` or `defer`.
//
//	valid:   <script src="/a.js" defer />
//	valid:   <script src="/a.js" async />
//	valid:   <script>inline</script>
//	invalid: <script src="/a.js" />
//
// Ported from `@next/next/no-sync-scripts`, read against oxc's `no_sync_scripts.rs`.
//
// The rule keys on the presence of the attributes rather than their values, which is upstream's
// choice and the correct one: `async` and `defer` are boolean attributes, so the idiomatic spelling
// is a bare `async` with no initializer. A port that read the value would find nothing there and
// report a script that is already asynchronous.
var NoSyncScripts = rule.Rule{
	// No family prefix. The config writes `nextjs/no-sync-scripts` and matching strips the
	// namespace on a `/` boundary, so a prefixed name matches nothing and runs on no files while
	// its own tests pass.
	Name: "@next/next/no-sync-scripts",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		report := func(node *ast.Node) {
			tagName, attributes := jsx.ElementParts(node)
			if !jsx.IsIntrinsicElementNamed(tagName, "script") {
				return
			}
			// A <script> with no src is inline. It blocks the parser too, but it has no network
			// fetch to defer, and `async` and `defer` have no meaning on it, so upstream leaves it
			// alone rather than asking for an attribute that would do nothing.
			if !jsx.HasAttributeNamed(attributes, "src", jsx.MatchExactly) {
				return
			}
			if jsx.HasAttributeNamed(attributes, "async", jsx.MatchExactly) || jsx.HasAttributeNamed(attributes, "defer", jsx.MatchExactly) {
				return
			}
			ctx.ReportNode(node, messageNoSyncScripts)
		}

		return rule.Listeners{
			ast.KindJsxOpeningElement:     report,
			ast.KindJsxSelfClosingElement: report,
		}
	},
}
