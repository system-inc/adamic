package next

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/nextjs"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoHeadImportInDocument = rule.Message{
	Id: "noHeadImportInDocument",
	Description: "This is the custom document and it imports `next/head`. That component collects " +
		"what individual pages contribute to the head and merges it during the page render, " +
		"which is a phase the document has already finished by the time it runs, so what it " +
		"gathers here is silently dropped or duplicated into the shell. The document has its own " +
		"`Head`, exported from `next/document`, and that is the one that writes the shell.",
}

// NoHeadImportInDocument flags an import of `next/head` inside the Next.js custom document.
//
//	valid:   `pages/_document.tsx`   import Document, { Head } from "next/document"
//	valid:   `pages/index.tsx`       import Head from "next/head"
//	invalid: `pages/_document.tsx`   import Head from "next/head"
//
// Ported from `@next/next/no-head-import-in-document`, read against oxc's
// `no_head_import_in_document.rs`, with every judgment below confirmed against the release oxlint
// binary rather than modelled from the Rust.
//
// # Two halves, and the name describes only one of them
//
// The name reads as though the rule reasons about a `Head` component, so a port is drawn toward
// resolving the local binding, or looking for `<Head>` in the tree, or checking whether the import
// is used at all. Upstream does none of it. The whole body is one equality test on the module
// specifier of a static import declaration, and it reports unconditionally on a match. Every
// upstream fixture carries a `<Head />` element in a document class, and deleting all of that JSX
// changes no verdict: it is decoration around a one-line rule.
//
// So the local name is irrelevant and the clause shape is irrelevant. Measured reporting on the
// release binary, and no upstream fixture writes any of them:
//
//	import Whatever from "next/head"     a renamed default
//	import * as Head from "next/head"    a namespace import
//	import "next/head"                   a bare side effect import, binding nothing
//	import type Head from "next/head"    type only, erased at compile time
//
// The type-only case is the one most likely to be filtered out as harmless by a porter reasoning
// about what the rule ought to mean. Upstream carries no import-kind check, so it reports, and
// improving on that here would be a silent divergence in a rule whose corpus could never see it.
//
// The specifier is compared for equality rather than by prefix or containment. Measured silent:
// `next/head/dist/index`, `next/head/`, `@next/head` and `./next/head` are all left alone, and a
// port reaching for a looser comparison passes all seven upstream fixtures while firing on four
// things upstream permits.
//
// # What it does not see, which is narrowness reproduced rather than missed
//
// Upstream registers the import declaration alone, so a specifier arriving any other way is
// invisible. Measured silent on the release binary:
//
//	const Head = require("next/head")     a CommonJS require
//	const Head = import("next/head")      a dynamic import
//	export { default } from "next/head"   a re-export, which is a different node entirely
//
// `imports.SourceVisitors` on our shelf would catch the first two, and its own doc comment draws
// exactly the line this rule sits on: a rule guarding a boundary needs all three shapes because the
// one it misses is the one used to route around it, while a rule asking about the syntax that
// brings a module in needs one. By its name this rule sounds like the first kind. By implementation
// upstream made it the second, and the port target is the behaviour rather than the intent, so the
// narrowing is reproduced and stated here so a later reader does not widen it as a bug fix.
//
// The rule also reports once per import declaration with no deduplication, so a document importing
// the module twice produces two findings. No upstream fixture writes the module twice; measured on
// the binary, which reports at both column 1 and column 34 for a file holding two.
//
// # The file gate, which is where the whole rule lives
//
// Upstream answers this in `should_run`, before the walk, so a file that is not the document
// registers nothing. That is a per-file property and it is answered here the same way, when the
// listeners are built, which is also the shape `no-head-element` and `no-document-import-in-page`
// already use for their own file-level questions.
//
// **This rule's upstream gate is its own spelling, and it disagrees with itself inside one
// function.** The basename arm tests `starts_with("_document.")`, with a trailing dot, and the
// directory arm tests `starts_with("_document")` on the parent, without one. So a file named
// `_documentation.tsx` is not the document and a directory named `_documentation/` is. One rule,
// two answers to the same question, and it is upstream's behaviour rather than a reading error:
// both were measured reporting and not reporting on the release binary before they were written
// down. The tidier answer is the wrong one and the fixtures pin both directions.
//
// `nextjs.IsDocumentFile` is used rather than the neighbouring `nextjs.IsDocumentPage`, and the
// choice is not a preference. `IsDocumentFile` is that helper's union of the spellings this family
// drifted into, and its body is this rule's `should_run` transcribed: basename beginning with
// `_document.`, or basename beginning with `index` while the parent begins with `_document`. Read
// against the Rust and then probed on the binary across thirteen paths, the two agree on every one,
// including the three that a reader would expect to go the other way:
//
//	pages/_documentation/index.tsx    reports, the missing dot on the directory arm
//	pages/_document/index.helper.tsx  reports, `starts_with("index")` rather than equality
//	source/components/_document.tsx   reports, oxc has no `pages` requirement at all
//
// The last is a real divergence from ESLint, which bails unless the path contains `pages`. oxc is
// the port target and it is preserved, which means a file named `_document.tsx` written anywhere in
// a tree for unrelated reasons is linted as a Next document.
//
// `IsDocumentPage` would be wrong here on all three: it is oxc's shared `is_document_page`, kept
// byte for byte for the one sibling whose corpus votes on the difference. This rule's corpus does
// not vote, and its own `should_run` is what decides.
//
// # What the fixtures below cannot reach, measured rather than assumed
//
// The gate resolves both path separators, because upstream reads the path through Rust's `Path` and
// a document arriving in the Windows shape is the same document. **No rule fixture can exercise
// that**, and the way that was established is worth more than the fact: a mutation blinding
// `splitPath` to the backslash survived the whole table, two Windows-shaped fixtures were added to
// kill it, and it survived again. The second survival is what sent me to read instead of guess.
//
// `rule_testing.Run` passes the file name through `tspath.NormalizePath` before the program is built, so
// a rule handed `pages\_document.tsx` sees `/pages/_document.tsx`. Probed directly: every backslash
// is a forward slash by the time `ctx.SourceFile.FileName()` answers. So the two fixtures were exact
// duplicates of their forward-slash twins, asserting nothing, and they were deleted rather than left
// to read as coverage. The separator handling is covered where it can be, in
// `internal/utilities/nextjs`'s own tests, which call the predicate directly and bypass the harness.
//
// A mutant that survives after the fixture written to kill it is not a fixture problem by default.
//
// # Where the finding points
//
// At the whole import declaration, matching upstream's `import_decl.span`. Measured on the binary
// against a file whose import sits under two comment lines: offset 24, length 28, so the span
// starts at the `import` keyword and excludes the leading trivia. `ctx.ReportNode` routes through
// TokenRange and already strips it, so `imports.SpecifierNode` is not reached for here. That helper
// exists to solve the leading-trivia anchor, and using it would move the finding off upstream's
// span onto the specifier for no gain.
var NoHeadImportInDocument = rule.Rule{
	// No family prefix. The config writes `nextjs/no-head-import-in-document` and matching strips
	// the namespace on a `/` boundary, so a prefixed name matches nothing and runs on no files.
	Name:       "@next/next/no-head-import-in-document",
	NoListener: rule.NoListenerDeclinesIrrelevantFiles,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Every file that is not the document registers no listener at all rather than testing the
		// path at every import, which is how upstream spells it too.
		if !nextjs.IsDocumentFile(ctx.SourceFile.FileName().AsString()) {
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
				// Equality, never a prefix or a contains. See the deep-path and trailing-slash
				// cases in the fixtures, both measured silent upstream.
				if declaration.ModuleSpecifier.Text() != "next/head" {
					return
				}
				// Unconditional. The import clause is never examined, so a type-only import and an
				// import binding nothing both report.
				ctx.ReportNode(node, messageNoHeadImportInDocument)
			},
		}
	},
}
