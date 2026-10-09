package next

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/ecmascript/nextjs"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoHeadElement = rule.Message{
	Id: "noHeadElement",
	Description: "This is a raw <head> element. Use the `Head` component from `next/head`, which " +
		"merges what every page contributes and deduplicates it, so two components asking for a " +
		"title do not both emit one. A hand-written <head> is emitted as-is and silently competes " +
		"with whatever the framework already put there.",
}

// NoHeadElement flags a raw <head> element outside the app directory.
//
//	valid:   <Head><title>x</title></Head>
//	valid:   a <head> in a file under an app directory
//	invalid: <head><title>x</title></head>
//
// Ported from `@next/next/no-head-element`, read against oxc's `no_head_element.rs`.
//
// # The app-directory exemption, and why it is a substring test
//
// The App Router owns the document shell itself, so a `<head>` written there is the framework's own
// metadata handling rather than a page competing with it. Upstream skips the whole file in that
// case, through what oxc spells `is_in_app_dir`, which is a plain substring test for `app/` on the
// path.
//
// That is looser than it looks and it is preserved rather than tightened. A path segment test would
// be the obvious improvement, but it would change which files the rule runs on, and upstream's
// answer is the one the rule's behaviour is defined by. Measured on this repository, the two agree:
// every `.tsx` path containing `app/` contains it as a real directory segment, so the loose test
// costs nothing here and being faithful costs nothing either.
//
// The exemption is checked once per file rather than per node, because it is a property of the file.
var NoHeadElement = rule.Rule{
	// No family prefix. The config writes `nextjs/no-head-element` and matching strips the
	// namespace on a `/` boundary, so a prefixed name matches nothing and runs on no files.
	Name: "@next/next/no-head-element",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Answering here rather than inside the listener means a file under an app directory
		// registers no listener at all, which is also how upstream spells it: `should_run` is asked
		// before the walk rather than at every element.
		if nextjs.IsInApplicationDirectory(ctx.SourceFile.FileName().AsString()) {
			return nil
		}

		report := func(node *ast.Node) {
			tagName, _ := jsx.ElementParts(node)
			if !jsx.IsIntrinsicElementNamed(tagName, "head") {
				return
			}
			ctx.ReportNode(node, messageNoHeadElement)
		}

		return rule.Listeners{
			ast.KindJsxOpeningElement:     report,
			ast.KindJsxSelfClosingElement: report,
		}
	},
}
