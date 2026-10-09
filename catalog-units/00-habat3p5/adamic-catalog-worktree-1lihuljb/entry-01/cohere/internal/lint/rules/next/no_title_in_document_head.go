package next

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoTitleInDocumentHead = rule.Message{
	Id: "noTitleInDocumentHead",
	Description: "A <title> here is emitted once for every page in the application, because the " +
		"`Head` from `next/document` renders the shared document shell rather than a page. A page " +
		"that then sets its own title competes with this one and which wins depends on order. Set " +
		"the title per page, with `Head` from `next/head`.",
}

// NoTitleInDocumentHead flags a <title> written directly inside a component imported from
// `next/document`.
//
//	valid:   import Head from 'next/head';     <Head><title>x</title></Head>
//	valid:   import { Head } from 'next/document';  <Head></Head>
//	invalid: import { Head } from 'next/document';  <Head><title>x</title></Head>
//
// Ported from oxc's `no_title_in_document_head.rs`, read against `@next/next`'s own rule. There is
// no option surface: oxc's struct is a unit and eslint declares `schema: []`.
//
// # The import is the gate, and there is deliberately no path gate
//
// Five rules in this package ask whether the file is a document or an application-directory file
// before they run. This one does not, and the absence is upstream's and is correct. The rule cannot
// be reached until a file has imported from `next/document`, which is a far tighter gate than any
// filename test: a `<Head>` from `@react-email/components` or a local component of that name can
// never satisfy it. Adding `nextjs.IsDocumentFile` here would look like a tidy-up and would silently
// delete the rule's main true positive, since `next/document`'s `Head` used *outside* a document
// file is exactly the mistake being caught, and both upstreams catch it there.
//
// # Only the FIRST named specifier of a declaration is watched, and that is a defect reproduced
//
// oxc reaches for the import's specifiers with `find_map`, which stops at the first one matching the
// named arm, and then watches only that binding. Every later named specifier in the same declaration
// is invisible. Measured against the release binary:
//
//	import { Head, Html } from 'next/document';  <Head><title>x</title></Head>   reports
//	import { Html, Head } from 'next/document';  <Head><title>x</title></Head>   silent
//	import { Html, Head } from 'next/document';  <Html><title>x</title></Html>   reports
//	import { A, Head, C } from 'next/document';  <Head><title>x</title></Head>   silent
//
// So the real violation goes unreported whenever `Head` is not written first, which is the ordering
// `next/document` users actually write, since `Html` conventionally comes first. Upstream's own
// second passing fixture is `import Document, { Html, Head }` with an empty `<Head>`, and it is
// clean for two reasons at once rather than the one it looks like: the `Head` is empty AND the
// watched binding is `Html`. That fixture cannot see this defect and neither can the other two.
//
// It is reproduced rather than corrected because the harness compares against oxc, and a port that
// helpfully watched every named specifier would report on inputs oxc is silent on. The intuitive
// reading is that the rule watches whichever specifier is named `Head`; it does not check the
// imported name at all, so `import { Html }` arms it and `<Html><title/></Html>` reports. Both
// halves are pinned by fixtures so neither is quietly corrected back.
//
// A separate import declaration gets its own first specifier, so two declarations watch two
// bindings.
//
// # References resolve through the checker rather than by name
//
// oxc walks `symbol_references` off the specifier's symbol, which is shadow-aware for free. Matching
// the local binding's text instead would be right in every fixture upstream ships and wrong on a
// local declaration of the same name inside a function. The checker answers this directly: asking
// `GetSymbolAtLocation` on the tag identifier and comparing the resolved declaration against the
// specifier node by identity reproduces upstream exactly, including the aliased form, and a shadowed
// `Head` resolves to its own variable declaration rather than to the import. Measured on all three
// shapes before the rule was written.
var NoTitleInDocumentHead = rule.Rule{
	// No family prefix. The config writes `nextjs/no-title-in-document-head` and matching strips the
	// namespace on a `/` boundary, so a prefixed name matches nothing and runs on no files.
	Name:             "@next/next/no-title-in-document-head",
	NeedsTypeChecker: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// The specifier nodes this file watches, at most one per `next/document` declaration.
		var watched []*ast.Node

		collect := func(node *ast.Node) {
			declaration := node.AsImportDeclaration()
			if declaration == nil || declaration.ModuleSpecifier == nil {
				return
			}
			// Inert against every input the parser can produce, and kept anyway. Probed across the
			// double-quoted, single-quoted and backtick spellings: all three answer yes, and the
			// `import x = require(...)` form is a different node kind that never reaches this
			// listener, so nothing can take the false branch. Mutating it away survives the fixture
			// set for that reason rather than for a missing case. It stays because `Text()` on a
			// non-literal is not safe and the guard states which shapes this line assumes.
			//
			// One measured divergence lives here and it is the parser's rather than the rule's:
			// oxc rejects a backtick module specifier as a syntax error, so upstream has no verdict
			// on `import { Head } from ` + "`next/document`" + ` at all, while our parser accepts it
			// and this rule would report. Not pinned as a fixture, because a fixture would assert a
			// behaviour the reference implementation cannot express either way.
			if !ast.IsStringLiteralLike(declaration.ModuleSpecifier) {
				return
			}
			if declaration.ModuleSpecifier.Text() != "next/document" {
				return
			}
			// Only the first named specifier, which is upstream's `find_map` and the defect
			// documented above. `BindingsOf` returns the named specifiers in source order, and its
			// Default and Namespace fields are deliberately not read: those are different specifier
			// arms upstream never matches.
			named := imports.BindingsOf(node).Named
			if len(named) == 0 {
				return
			}
			watched = append(watched, named[0])
		}

		reportTitleChildren := func(element *ast.Node, headTagName *ast.Node) {
			// Direct children only, which is upstream's single pass over `jsx_element.children`. A
			// title one level down, or inside an expression container, is not a child element here.
			//
			// The kind half of this guard is inert and kept for the same reason as the tag guard
			// above: a KindJsxOpeningElement exists only as the OpeningElement of a
			// KindJsxElement, so the parent handed in here is always that kind and nothing can
			// reach the false branch. Probed rather than reasoned: every opening element across a
			// paired tag, a fragment child, a member tag and an expression container reported the
			// same parent kind. Mutating the kind test away survives the fixture set, and that
			// survivor is subsumed by the parse shape rather than unseen. The nil test is the half
			// that does work, since AsJsxElement would dereference it.
			if element == nil || element.Kind != ast.KindJsxElement {
				return
			}
			children := element.AsJsxElement().Children
			if children == nil {
				return
			}
			for _, child := range children.Nodes {
				// Both opening forms, because our AST gives a self-closing `<title />` its own kind
				// while oxc sees one `JSXElement` either way. Reading only the paired form would be
				// silent on the self-closing child and every upstream fixture would still pass.
				var tagName *ast.Node
				switch child.Kind {
				case ast.KindJsxElement:
					tagName, _ = jsx.ElementParts(child.AsJsxElement().OpeningElement)
				case ast.KindJsxSelfClosingElement:
					tagName, _ = jsx.ElementParts(child)
				default:
					continue
				}
				if !jsx.IsIntrinsicElementNamed(tagName, "title") {
					continue
				}
				// Upstream reports inside the child loop rather than breaking, so two titles under
				// one element produce two findings at the same span. Pinned by a fixture.
				// The finding points at the `Head` tag NAME, which is upstream's
				// `jsx_opening_element.name.span()`, not at the `<title>` that eslint blames and not
				// at the whole opening element. Asserted by a span fixture, which is how the first
				// version of this line was caught pointing at the title.
				ctx.ReportNode(headTagName, messageNoTitleInDocumentHead)
			}
		}

		return rule.Listeners{
			ast.KindImportDeclaration: collect,
			ast.KindJsxOpeningElement: func(node *ast.Node) {
				if len(watched) == 0 {
					return
				}
				// Through the shelf rather than the accessor, which is what
				// `TestRulePackagesDoNotReachPastWrappedAccessors` asks for and what caught the
				// first version of this line. Only the paired form is listened for, because a
				// self-closing `<Head />` has no children for the title scan to find, so the
				// utility's second answer is unused here rather than unneeded.
				tagName, _ := jsx.ElementParts(node)
				// A member-expression tag (`<Head.Sub>`) is not the binding being used as an
				// element, and upstream declines it by requiring the reference's own parent to be
				// the opening element.
				//
				// This guard changes no verdict and is kept deliberately. Mutating it away survives
				// the whole fixture set, and the survivor is equivalent rather than a blind spot: a
				// member tag parses to KindPropertyAccessExpression, and resolving one answers the
				// PROPERTY's symbol, never the imported binding's specifier node, so the identity
				// test below already declines every such tag. Four shapes were run against the
				// mutant to look for a distinguishing input, `<Head.Sub>`, `<NS.Head>` over a
				// shorthand object, `<All.Head>` over a namespace import, and a namespace re-export,
				// and all four reported zero with the guard removed. It stays because reading a tag
				// name's text before checking its kind panics outright on that kind, which is a
				// documented live trap in this tree, and because it states the intent at the line.
				if tagName == nil || tagName.Kind != ast.KindIdentifier {
					return
				}
				if !resolvesToWatchedSpecifier(ctx, tagName, watched) {
					return
				}
				reportTitleChildren(node.Parent, tagName)
			},
		}
	},
}

// resolvesToWatchedSpecifier reports whether a tag identifier binds to one of the watched import
// specifiers, by declaration node identity rather than by name.
//
// Identity is the whole point. Comparing the declaration's KIND instead would answer true for any
// import specifier in the file, and comparing the identifier's text would answer true for a local
// component shadowing the import. Both pass every fixture upstream ships.
func resolvesToWatchedSpecifier(ctx rule.Context, tagName *ast.Node, watched []*ast.Node) bool {
	if ctx.TypeChecker == nil {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(tagName)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		for _, specifier := range watched {
			if declaration == specifier {
				return true
			}
		}
	}
	return false
}
