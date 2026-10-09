package next

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/ecmascript/nextjs"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoStyledJsxInDocument = rule.Message{
	Id: "noStyledJsxInDocument",
	Description: "This is styled-jsx inside the custom document. The document shell renders once, " +
		"outside the React tree that styled-jsx instruments, so the styles it collects here are " +
		"never injected and the rule that looks correct in a page silently produces nothing. Move " +
		"the styles into a page or a component, or emit plain CSS with a raw <style> element.",
}

// NoStyledJsxInDocument flags a styled-jsx `<style jsx>` element inside a Next.js custom document.
//
//	valid:   <style>{"body{color:red}"}</style>            in pages/_document.tsx
//	valid:   <style {...{jsx: true}}></style>              in pages/_document.tsx
//	valid:   <style jsx>{`p{color:orange}`}</style>        in pages/index.jsx
//	invalid: <style jsx>{"body{color:red}"}</style>        in pages/_document.tsx
//
// Ported from `@next/next/no-styled-jsx-in-document`, read against oxc's
// `no_styled_jsx_in_document.rs`. No options: oxc's rule is a unit struct with no
// `from_configuration`, and the eslint rule carries `schema: []` and never reads `context.options`.
//
// # What styled-jsx is, to this rule
//
// Three string comparisons and nothing more. The name suggests the rule resolves the `styled-jsx`
// package, reads an import, or analyses a tagged template, and it does none of that. It asks
// whether an intrinsic `<style>` carries an attribute literally named `jsx`, and it asks that by
// **presence rather than value**, because `<style jsx>` is written bare with no initializer. A port
// reading the value would find nothing and report nothing.
//
// Measured against the release binary in a `pages/_document.tsx`, which is the authority for every
// line of this paragraph. These report: `<style jsx>`, `<style jsx global>`, `<style jsx="true">`,
// `<style jsx={true}>`, and the self-closing `<style jsx />`. These are silent: `<style JSX>` (the
// comparison is case sensitive), `<style global>`, `<Style jsx>`, `<style.x jsx>`, `<div jsx>`, and
// `<style data-jsx>`.
//
// # The spread is visible and still exempt
//
// `<style {...{jsx: true}}></style>` is **silent upstream**, and upstream's own corpus pins the
// weaker form of this with a `nonce` spread. That is worth stating because a spread is not invisible
// to us the way it is to oxc's attribute loop: it arrives as a real `KindJsxSpreadAttribute` node
// and its object literal is right there to read. Resolving it would be a strict improvement on what
// the rule catches and a divergence from what the rule decides, so it is declined deliberately.
// `jsx.HasAttributeNamed` already answers this correctly, because `jsx.AttributeName` returns false
// for any node that is not a `KindJsxAttribute`.
//
// # The document gate, and where it diverges from this rule's own upstream spelling
//
// This is the one place the port knowingly disagrees with the rule it was ported from, and it is a
// disagreement inherited on purpose from `internal/utilities/nextjs`.
//
// oxc spells the gate inside this rule as a basename test: `file_name.starts_with("_document.")`,
// with no directory component at all. The word `pages` appears in this rule's diagnostic text, its
// doc comment and a doc example, and **never in its executable code**, so `components/_document.tsx`
// fires upstream. `nextjs.IsDocumentFile` agrees on that and on every other path this rule's own
// spelling decides, including the trailing dot that keeps `_documentation.tsx` silent.
//
// It disagrees on exactly one input. The shared helper is the union of the four spellings the family
// uses upstream, so it also accepts the directory form `pages/_document/index.tsx`, which only
// `no_head_import_in_document` accepts. Pinned against the release binary: that path is **silent**
// under this rule upstream and reports here. Both answers are on the fixture list below, the second
// marked as the divergence it is.
//
// It is taken anyway, because the alternative is worse than the divergence. Eight rules in this
// family gate on one question, upstream answers it four ways, and one upstream rule answers it two
// ways nine lines apart. Reproducing this rule's spelling privately would make it the second Go
// answer to a question the shelf already answers, and nothing would ever catch the two drifting,
// since each rule's corpus was written against its own spelling and passes either way. A stated
// divergence on one path is cheaper than a silent disagreement that no guard can see.
var NoStyledJsxInDocument = rule.Rule{
	// No family prefix. The config writes `nextjs/no-styled-jsx-in-document` and matching strips
	// the namespace on a `/` boundary, so a prefixed name matches nothing and runs on no files.
	Name:       "@next/next/no-styled-jsx-in-document",
	NoListener: rule.NoListenerDeclinesIrrelevantFiles,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Asked once rather than per element, because it is a property of the file. Upstream asks
		// it per node only because its listener has nowhere earlier to stand.
		if !nextjs.IsDocumentFile(ctx.SourceFile.FileName().AsString()) {
			return nil
		}

		report := func(node *ast.Node) {
			tagName, attributes := jsx.ElementParts(node)
			if !jsx.IsIntrinsicElementNamed(tagName, "style") {
				return
			}
			// Exact rather than case insensitive: upstream compares the attribute name bytewise,
			// and `<style JSX>` is silent there.
			if !jsx.HasAttributeNamed(attributes, "jsx", jsx.MatchExactly) {
				return
			}
			ctx.ReportNode(node, messageNoStyledJsxInDocument)
		}

		// Two kinds where oxc registers one. A self-closing element still produces a
		// `JSXOpeningElement` in oxc's tree and is a distinct kind in ours, so listening only for
		// the paired form would make `<style jsx />` silent, which upstream reports.
		return rule.Listeners{
			ast.KindJsxOpeningElement:     report,
			ast.KindJsxSelfClosingElement: report,
		}
	},
}
