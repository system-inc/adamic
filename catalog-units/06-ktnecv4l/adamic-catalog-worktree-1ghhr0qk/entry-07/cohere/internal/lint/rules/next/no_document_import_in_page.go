package next

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/nextjs"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoDocumentImportInPage = rule.Message{
	Id: "noDocumentImportInPage",
	Description: "This file imports `next/document` and is not the custom document. The pieces " +
		"that module exports only mean anything while the document shell is being rendered, so " +
		"anywhere else they either render nothing or break the page they are on. Move the code " +
		"that needs them into `pages/_document`, and use `next/head` for the ordinary case of a " +
		"page contributing to the head.",
}

// NoDocumentImportInPage flags an import of `next/document` outside the custom document file.
//
//	valid:   `pages/_document.tsx`      import Document from "next/document"
//	valid:   `components/Thing.tsx`     import Head from "next/head"
//	invalid: `components/Thing.tsx`     import Document from "next/document"
//	invalid: `pages/About.tsx`          import { Html } from "next/document"
//
// Ported from `@next/next/no-document-import-in-page`, read against oxc's
// `no_document_import_in_page.rs` and its shared `is_document_page` helper, with every judgment
// below confirmed against the release oxlint binary rather than modelled.
//
// # The rule reads the specifier and nothing else
//
// The name says "Document import", which invites a port to look for a default binding named
// `Document`, or for a class extending it. Upstream looks for neither. The whole body is an
// equality test on the module specifier of a static import declaration, so the import clause is
// never examined and the binding's name is irrelevant. Upstream's own corpus makes the point three
// times over: one pass case renames the binding to `NextDocument`, and two more rename it to
// `NDocument` while naming a local class `Document`, and all three turn on the path alone.
//
// Measured on the release binary, every one of these reports, and the corpus pins none of them:
//
//	import { Html, Main } from "next/document"
//	import * as Document from "next/document"
//	import "next/document"
//	import type Document from "next/document"
//
// The type-only form is worth calling out because it is the one a porter is most likely to filter
// out as harmless. Upstream has no `import_kind` check, so it reports, and so does this.
//
// # What it does not see, and why that is deliberate rather than an oversight
//
// Upstream registers the import declaration alone, so a specifier arriving any other way is
// invisible to it. Measured silent on the release binary:
//
//	const p = import("next/document")     a dynamic import
//	const D = require("next/document")    a CommonJS require
//	export { Html } from "next/document"  a re-export, which is not an import declaration at all
//
// `imports.SourceVisitors` on our shelf would catch the first two, and its own doc comment draws
// exactly the line this rule sits on: a rule guarding a boundary wants all three shapes, a rule
// asking about the syntax that brings a module in wants one. By meaning this rule looks like the
// first kind. By implementation upstream made it the second, and the port target is the behaviour
// rather than the intent, so the narrowing is reproduced and stated here so a later reader does not
// widen it as a bug fix. The re-export is silent for a plainer reason: it is a different node.
//
// # The path gate, which is the only interesting part
//
// Upstream computes the gate in `should_run`, before the walk, by asking the shared helper whether
// the file is the document and negating the answer. That is a per-file property, so it is answered
// here once when the listeners are built rather than at every import, which is the same shape
// `no-head-element` uses for its own file-level exemption and the same shape upstream uses.
//
// **`nextjs.IsDocumentPage` is used here rather than the neighbouring `nextjs.IsDocumentFile`, and
// choosing the other one would break the corpus.** IsDocumentFile is the union of the four spellings
// upstream drifted into across this family, and it is the right answer for the sibling rules whose
// corpora never vote on the difference. This rule's corpus votes: its fourth failing case is
// `src/pages/user/_document.tsx`, a file named exactly `_document.tsx` that upstream deliberately
// reports because it is not the immediate child of a `pages` directory, and IsDocumentFile exempts
// it. Both predicates and their disagreement are documented at `internal/utilities/nextjs/paths.go`.
//
// Fidelity is to what a rule decides, and here the decision is pinned by a test upstream wrote.
//
// # Where the finding points
//
// At the whole import declaration, matching upstream's `import_decl.span`, confirmed on the binary
// at offset 20 length 37 for a file whose import sits under a comment: the span starts at the
// `import` keyword and excludes the leading trivia. `ctx.ReportNode` routes through `TokenRange`
// and does the same, so the shelf's `imports.SpecifierNode` is not needed to avoid the
// leading-trivia anchor it exists to solve.
var NoDocumentImportInPage = rule.Rule{
	// No family prefix. The config writes `nextjs/no-document-import-in-page` and matching strips
	// the namespace on a `/` boundary, so a prefixed name matches nothing and runs on no files.
	Name: "@next/next/no-document-import-in-page",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// The document file is allowed to import the module it exists to implement, so it registers
		// no listener at all rather than being filtered per import. Upstream spells this the same
		// way, in `should_run`.
		if nextjs.IsDocumentPage(ctx.SourceFile.FileName().AsString()) {
			return nil
		}

		return rule.Listeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				declaration := node.AsImportDeclaration()
				if declaration == nil || declaration.ModuleSpecifier == nil {
					return
				}
				if !ast.IsStringLiteralLike(declaration.ModuleSpecifier) {
					return
				}
				// An equality test, not a prefix or a contains. `next/documents`,
				// `next/document/foo`, `@next/document` and `./next/document` are all silent
				// upstream, measured, and a port reaching for a looser comparison passes every
				// imported fixture while flagging all four.
				if declaration.ModuleSpecifier.Text() != "next/document" {
					return
				}
				ctx.ReportNode(node, messageNoDocumentImportInPage)
			},
		}
	},
}
