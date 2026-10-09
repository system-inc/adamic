package next

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module"
	"github.com/system-inc/cohere/internal/lint/ecmascript/nextjs"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// googleFontsStylesheetPrefix is the entire font test this rule performs.
//
// Not a font-host list, not a URL parse, not a `rel` check. One case-sensitive `strings.HasPrefix`
// against one literal, which is exactly what upstream writes. `/css` rather than `/css2` so that
// both the v1 `?family=` endpoint and the v2 one match, and the scheme is part of the prefix, so an
// `http://` spelling of the same URL is silent. Every one of those was measured against the release
// binary rather than read off the constant; the fixtures below pin them.
const googleFontsStylesheetPrefix = "https://fonts.googleapis.com/css"

var messageNoPageCustomFontNotInDocument = rule.Message{
	Id: "noPageCustomFontNotInDocument",
	Description: "This is a Google Fonts stylesheet requested from a page rather than from the " +
		"custom document. The document shell renders once for the whole application, so a font " +
		"link written there is fetched once and shared; written in a page it is fetched again per " +
		"page and the framework cannot optimize it. Move the link into the custom document.",
}

var messageNoPageCustomFontOutsideDefaultExport = rule.Message{
	Id: "noPageCustomFontOutsideDefaultExport",
	Description: "This is a Google Fonts stylesheet inside the custom document but outside the " +
		"component the document default-exports. The framework rewrites font links it finds in " +
		"the exported document component, and a link rendered from a helper component is invisible " +
		"to that rewriting, so automatic font optimization silently does nothing. Inline the link " +
		"into the default-exported component.",
}

// NoPageCustomFont flags a Google Fonts stylesheet link written outside the custom document.
//
//	valid:   <link href="https://fonts.googleapis.com/css2?family=X" /> inside the default export
//	         of pages/_document.tsx
//	valid:   <link rel="apple-touch-startup-image" href="/splash.jpg" /> anywhere
//	valid:   <link href={fontUrl} /> anywhere, because the value is not a string literal
//	invalid: <link href="https://fonts.googleapis.com/css2?family=X" /> in pages/index.tsx
//	invalid: the same link in pages/_document.tsx, rendered from a helper component
//
// Ported from `@next/next/no-page-custom-font`, read against oxc's `no_page_custom_font.rs`. No
// options: oxc's rule is a unit struct with no `from_configuration`, and the eslint rule carries
// `schema: []` and never reads `context.options`.
//
// # It is not a font rule, and the name is the trap
//
// The name and both upstream messages describe a general policy about custom fonts. The executable
// rule is one string prefix on one attribute of one intrinsic element. A port that implements what
// the name says writes a different rule, and its own fixtures pass, because they were written from
// the same belief.
//
// Measured against the release binary, in a real `_document.tsx`, outside the default export:
// `<link href="https://fonts.googleapis.com/cssANYTHING" />` with **no `rel` attribute at all**
// reports, and `<link rel="preload" href="https://fonts.googleapis.com/css?family=X" />` reports.
// These are silent: a `rel="stylesheet"` pointing at `https://use.typekit.net/abc.css`, one
// pointing at `https://fonts.gstatic.com/css?family=X`, the `http://` spelling of the Google URL,
// an `HREF` spelled in capitals, an `href={variable}` expression, a spread carrying the href, and
// a capitalized `<Link>` component. So `rel` is never read, the href must be a string literal, the
// attribute name is case sensitive, and only this one host and path prefix count.
//
// # Two findings, chosen by the filename and nothing else
//
// The gate is the file, then the location, in that order, and the href test runs before both:
//
//	not the custom document          report noPageCustomFontNotInDocument
//	the custom document, inside the default export     silent
//	the custom document, outside it  report noPageCustomFontOutsideDefaultExport
//
// # The second message names <Head> and the rule never looks for one
//
// Upstream's text is "Using `<link />` outside of `<Head>`", and the rule contains no ancestor
// search for `Head` whatsoever. It asks whether the node is inside the module's default export.
// Upstream's own failing case proves the message is describing the wrong mechanism: its two links
// sit in a `Links()` helper that **is** rendered inside `<Head>`, and they report; its first
// failing case is a link correctly inside `<Head>` in a page file, and it reports too. So the
// description above says what is actually decided rather than repeating upstream's framing, and
// the message id says `OutsideDefaultExport` for the same reason. A porter who adds a `<Head>`
// ancestor check to match the text breaks both failing cases.
//
// # The document gate
//
// Upstream spells it as a bare basename test, `file_name.starts_with("_document.")`, with no
// directory component at all and no `pages` requirement. That is byte-identical to the spelling in
// `no_styled_jsx_in_document.rs:100`, so this rule takes the same shelf predicate its sibling
// takes, `nextjs.IsDocumentFile`, and inherits the same single stated divergence.
//
// Measured on the release binary with a control rule (`google-font-display`) firing on all six
// paths, so a silence here is a real answer rather than a rule that never ran:
//
//	pages/_document.tsx           silent    agrees
//	components/_document.tsx      silent    agrees, and there is no `pages` requirement
//	src/pages/user/_document.tsx  silent    agrees
//	pages/index.tsx               REPORTS   agrees
//	pages/_documentation.tsx      REPORTS   agrees, and the trailing dot is why
//	pages/_document/index.tsx     REPORTS   DIVERGES; IsDocumentFile silences it
//
// The last line is the divergence, taken deliberately. `IsDocumentFile` is the union of the four
// spellings this family uses upstream, so it also accepts the directory form that only
// `no_head_import_in_document` accepts. Reproducing this rule's private spelling would make it the
// second Go answer to a question the shelf already answers, and nothing could ever catch the two
// drifting, because each rule's corpus was written against its own spelling and passes either way.
// `nextjs.IsDocumentPage` would be wrong here in the other direction: it requires a `pages` split,
// which would silence `components/_document.tsx`, a path upstream treats as the document.
//
// # Whether it fires outside the document
//
// It does, on every file, and that is worth stating because the rule's name says "page" and its
// first message names `pages/_document.js`. There is no `should_run`, no `pages` requirement, and
// no app-directory exemption. A Google Fonts link in a component, a library file, or an App Router
// route all report `noPageCustomFontNotInDocument`. Pinned above on `components/_document.tsx` and
// on paths with no `pages` segment.
var NoPageCustomFont = rule.Rule{
	// No family prefix. The config writes `nextjs/no-page-custom-font` and matching strips the
	// namespace on a `/` boundary, so a prefixed name matches nothing and runs on no files.
	Name: "@next/next/no-page-custom-font",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// A property of the file, so it is asked once rather than at every element. Upstream asks
		// it per node only because its listener has nowhere earlier to stand.
		inDocument := nextjs.IsDocumentFile(ctx.SourceFile.FileName().AsString())

		// Collected once at rule entry rather than by a stateful visitor, because
		// `export default CustomDocument;` may sit BELOW the JSX it exports and upstream's answer
		// is order independent: it reads the module record rather than the walk. Upstream's fourth
		// passing case is exactly that shape and a source-order visitor reports it.
		defaultExportedNames := defaultExportedLocalNames(ctx.SourceFile)

		report := func(node *ast.Node) {
			tagName, attributes := jsx.ElementParts(node)
			// An intrinsic `link` only. A member-expression tag such as `<Foo.link />` is silent
			// upstream, and reading text off a property access panics, so the kind decides before
			// the text is touched.
			if !jsx.IsIntrinsicElementNamed(tagName, "link") {
				return
			}
			// Exact rather than case insensitive: `HREF` is silent upstream, measured. And
			// StringAttributeValue answers false for anything that is not a string literal
			// initializer, which is what makes `href={variable}` silent, also measured.
			href, isStringLiteral := jsx.StringAttributeValue(attributes, "href", jsx.MatchExactly)
			if !isStringLiteral || !strings.HasPrefix(href, googleFontsStylesheetPrefix) {
				return
			}

			if !inDocument {
				ctx.ReportNode(node, messageNoPageCustomFontNotInDocument)
				return
			}
			if isInsideDefaultExport(node, defaultExportedNames) {
				return
			}
			ctx.ReportNode(node, messageNoPageCustomFontOutsideDefaultExport)
		}

		// Two kinds where oxc registers one. `<link />` is a void element and is written
		// self-closing essentially always, which in our tree is a distinct kind that produces no
		// JsxOpeningElement at all. Listening only for the paired form would leave the rule silent
		// on every case in upstream's corpus. The paired `<link></link>` spelling reports upstream
		// too, measured, so both kinds are live rather than one being defensive.
		return rule.Listeners{
			ast.KindJsxOpeningElement:     report,
			ast.KindJsxSelfClosingElement: report,
		}
	},
}

// isInsideDefaultExport reports whether a JSX node is lexically inside the module's default export.
//
// Two ways to be inside, matching upstream's ancestor walk exactly:
//
//   - an ancestor is a declaration carrying both `export` and `default`, which covers
//     `export default function CustomDocument() {}` and, critically, the ANONYMOUS
//     `export default function() {}` — that one is reached without ever resolving a name, and a
//     port that walks up looking for a named function and then consults an export table reports it
//   - an ancestor is a function or class whose name is what a separate `export default <name>`
//     statement exports
//
// Only functions and classes are considered as name carriers. Any other ancestor kind is stepped
// over rather than ending the walk, which is upstream's `_ => continue`.
//
// # Where a name comes from, and the part a research pass got backwards
//
// An arrow function contributes no name OF ITS OWN and therefore falls back to the enclosing
// variable declarator, exactly like an anonymous function expression does. A recorded research pass
// read oxc's `AstKind::ArrowFunctionExpression(_) => None` as "arrows are skipped, only anonymous
// functions fall through to the declarator" and had it inverted: that `None` IS the id, and it flows
// straight into the `map_or_else` whose fallback reads the declarator. `_ => continue` is what skips
// a node; `None` does the opposite.
//
// Measured on the release binary, in a `_document.tsx`, which is what settled it:
//
//	const D = () => <link href="...css2..." />;        export default D;   SILENT
//	const Q = () => <link href="...css2..." />;        export default function D(){}   REPORTS
//	const D = function () { return <link .../> };      export default D;   SILENT
//	const D = function Inner() { return <link .../> }; export default D;   REPORTS
//
// The last line is the same mechanism seen from the other side and is the sharper case: a NAMED
// function expression takes its own name, `Inner`, and never consults the declarator, so the
// declarator's `D` is never compared and the default export never matches. The rule is not "what is
// this thing called at its binding site" but "what is its own name, or failing that its binding
// site's name", and those differ exactly here.
func isInsideDefaultExport(node *ast.Node, defaultExportedNames map[string]bool) bool {
	for ancestor := node.Parent; ancestor != nil; ancestor = ancestor.Parent {
		if isDefaultExportDeclaration(ancestor) {
			return true
		}

		name, carriesName := nameOfFunctionOrClass(ancestor)
		if !carriesName {
			continue
		}
		if defaultExportedNames[name] {
			return true
		}
	}
	return false
}

// isDefaultExportDeclaration is upstream's `AstKind::ExportDefaultDeclaration`, which our tree
// splits in two.
//
// oxc has one node kind for both spellings of a default export. We have two, and they are reached
// differently: `export default function D() {}` is a declaration carrying the `export` and
// `default` modifiers, while `export default <expression>` is a `KindExportAssignment` that carries
// **no modifiers at all**. So a check written as `module.IsDefaultExported` alone answers false for
// the whole expression form.
//
// That was not theoretical. It was written that way first and the fixture for
// `export default () => <link ... />` caught it: measured silent on the release binary, reported
// here, because the arrow's only route to the default export is through the assignment node and
// nothing was looking at it. This is the shape the brief calls out, where fidelity is to what the
// rule decides rather than to how the original obtains it, and the two kinds are one decision.
//
// `export = x` is excluded because it is the CommonJS form and is not a default export.
func isDefaultExportDeclaration(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindExportAssignment {
		return !node.AsExportAssignment().IsExportEquals
	}
	return module.IsDefaultExported(node)
}

// nameOfFunctionOrClass reads the name an ancestor contributes to the default-export comparison.
//
// The second return distinguishes "this ancestor is not a name carrier at all", which the caller
// steps over, from a carrier whose name resolution came up empty, which the caller also steps over.
// They behave identically here and are kept distinct because upstream distinguishes them and a
// later reader comparing the two will look for it.
func nameOfFunctionOrClass(node *ast.Node) (name string, carriesName bool) {
	switch node.Kind {
	case ast.KindArrowFunction:
		// No name of its own, ever, so it goes straight to the declarator fallback. See the note
		// on isInsideDefaultExport: this is the arm a research pass read as a skip.
		return nameFromEnclosingVariableDeclarator(node)

	case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
		ast.KindClassDeclaration, ast.KindClassExpression:
		if declared := node.Name(); declared != nil {
			return declared.Text(), true
		}
		return nameFromEnclosingVariableDeclarator(node)
	}
	return "", false
}

// nameFromEnclosingVariableDeclarator reads `const Foo = <this node>` and returns `Foo`.
//
// Only a plain identifier binding answers. A destructuring pattern binds no single name that could
// appear in an `export default`, which is upstream's `get_identifier_name` returning nothing.
// Measured: `const [D] = [() => <link ... />]; export default D;` reports upstream, and reports
// here.
//
// The identifier guard is kept for intent and is EQUIVALENT rather than load-bearing, which is
// recorded because a sweep will find it and a later reader should not spend the round trip twice.
// Relaxing it to `binding == nil` alone was scored and survived; the mutant was then run directly
// on the destructuring input above and produced the identical answer, so no input distinguishes
// them. A binding pattern's own text is `[D]` rather than `D`, so the map lookup misses either way.
func nameFromEnclosingVariableDeclarator(node *ast.Node) (name string, carriesName bool) {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindVariableDeclaration {
		return "", false
	}
	binding := parent.AsVariableDeclaration().Name()
	if binding == nil || binding.Kind != ast.KindIdentifier {
		return "", false
	}
	return binding.Text(), true
}

// defaultExportedLocalNames collects the local names a file default-exports by a separate statement.
//
// This is the `export default CustomDocument;` half. Gathered in one pass over the top-level
// statements before any listener runs, because the statement may appear below the JSX and the
// answer must not depend on walk order. Upstream gets order independence for free by reading its
// module record; ours is a scan, so it happens once here.
//
// Only the separate-statement form is collected. A declaration carrying the modifiers is answered
// by `module.IsDefaultExported` on the ancestor walk instead, so collecting it here as well would
// be a second answer to a question already answered.
//
// `export { D as default }` is deliberately NOT collected, and that is fidelity rather than an
// omission: upstream reads `local_export_entries` and measured on the release binary that spelling
// **reports**, so a file exporting its document that way is flagged upstream and is flagged here.
func defaultExportedLocalNames(sourceFile *ast.SourceFile) map[string]bool {
	names := map[string]bool{}
	if sourceFile == nil {
		return names
	}
	for _, statement := range sourceFile.Statements.Nodes {
		if statement.Kind != ast.KindExportAssignment {
			continue
		}
		assignment := statement.AsExportAssignment()
		// `export = x` is the CommonJS form and is not a default export.
		if assignment.IsExportEquals {
			continue
		}
		expression := ast.SkipParentheses(assignment.Expression)
		if expression != nil && expression.Kind == ast.KindIdentifier {
			names[expression.Text()] = true
		}
	}
	return names
}
