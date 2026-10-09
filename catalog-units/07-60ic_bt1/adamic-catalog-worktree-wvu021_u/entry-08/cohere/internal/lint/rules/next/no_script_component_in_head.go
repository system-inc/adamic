package next

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoScriptComponentInHead = rule.Message{
	Id: "noScriptComponentInHead",
	Description: "This puts a <Script> inside <Head>. The head component only forwards plain markup " +
		"into the document head, so the script loader never mounts and the loading strategy is " +
		"dropped: nothing decides when the script runs and nothing tracks whether it already did. " +
		"Move the <Script> outside of <Head> and let it sit in the page body.",
}

// NoScriptComponentInHead flags a <Script> written as a direct child of next/head's <Head>.
//
//	valid:   import Head from 'next/head'; <Head><title>a</title></Head>
//	valid:   import Head from 'next/head'; <Head><div><Script /></div></Head>
//	valid:   const Head = (p) => p.children; <Head><Script></Script></Head>
//	invalid: import Head from 'next/head'; <Head><Script></Script></Head>
//	invalid: import Head from 'next/head'; <Head><Script src="/a.js" /></Head>
//
// Ported from `@next/next/no-script-component-in-head`, read against oxc's
// `no_script_component_in_head.rs` and measured on the release binary, because the corpus is one
// pass case and one fail case and answers almost none of the questions the rule actually decides.
//
// # The two halves are asymmetric, and that asymmetry is the rule
//
// The `<Head>` side is resolved. oxc enters on the import declaration, finds `next/head`'s default
// specifier, and walks that symbol's references, so an aliased `import H from 'next/head'` reports
// and a `Head` shadowed by a local binding does not.
//
// The `<Script>` side is a bare string comparison on the child's tag text. `next/script` is never
// resolved, and oxc's own source shows this is not an oversight the way the eslint original's is:
// the eslint plugin carries a dead `if (node.source.value !== 'next/script') return` inside a
// visitor that does nothing afterwards. Both land in the same place.
//
// **The consequences run in both directions and both were measured rather than reasoned about.** A
// locally defined component named `Script` inside a real `next/head` `<Head>` reports, and
// `import S from 'next/script'` used as `<S>` inside `<Head>` is silent. Those look like defects and
// reproducing them is the port. `imports.LocalNameOfDefaultImport` sits one line away and its own
// doc comment invites resolving the script side, calling upstream's string match a defect that a
// port using the helper would improve on. Taking that invitation would flip both verdicts above and
// no imported fixture could see it, because the corpus writes `Script` as the local name of a real
// `next/script` import, where resolving and not resolving agree. The fixtures below pin both
// directions so the improvement cannot be made by accident.
//
// # Why the checker rather than a name
//
// Matching the tag text against the import's local name passes every case upstream ships and every
// obvious one a porter would invent. It fails on shadowing, which is silent upstream and which
// nothing in the corpus writes: `const Head = ...` inside the function shadows the import, and the
// element is then not the framework's head at all. The checker answers which declaration a tag
// binds to, so the rule anchors on the import clause the listener already saw and compares node
// identity, the technique `no_class_assign.go` established.
//
// An import binding's symbol declares at the `ImportClause`, not at the identifier, which is why the
// anchor is the default binding's parent rather than the binding itself. Measured with a probe; the
// identifier compares equal to nothing.
//
// # Depth, kinds, and count
//
// Only direct children, one level. `<Head><div><Script /></div></Head>` is silent upstream and is
// pinned silent here, so a later deep walk is a test failure rather than a quiet widening.
//
// Both element kinds are scanned as children. oxc sees one `JSXElement` node whose opening element
// may be self closing; our parser splits that into `KindJsxElement` and `KindJsxSelfClosingElement`,
// and the self closing form is the one real code writes. A port reading only `KindJsxElement`
// children passes both upstream fixtures and misses `<Script src="..." />` entirely.
//
// One finding per matching child rather than one per `<Head>`, and the finding points at the `<Head>`
// opening element's name node rather than at the `<Script>` that caused it. Both taken from the
// snapshot, whose caret sits under `Head` for four columns.
//
// No fix. The repair moves an element out of a subtree and has to choose where it lands, which
// changes what the page does.
var NoScriptComponentInHead = rule.Rule{
	// No family prefix. The config writes `nextjs/no-script-component-in-head` and matching strips
	// the namespace on a `/` boundary, so a prefixed name matches nothing and runs on no files while
	// its own tests pass.
	Name: "@next/next/no-script-component-in-head",

	// Shadowing is the whole reason. See the doc above: a local `Head` is textually identical to the
	// imported one and is silent upstream, and only resolution separates them.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// The clause each `next/head` default import declares. A file can write more than one such
		// import, so this is a set rather than a single anchor.
		headImportClauses := map[*ast.Node]bool{}

		return rule.Listeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				if _, imported := imports.LocalNameOfDefaultImport(node, "next/head"); !imported {
					return
				}
				defaultBinding := imports.BindingsOf(node).Default
				if defaultBinding == nil || defaultBinding.Parent == nil {
					return
				}
				headImportClauses[defaultBinding.Parent] = true
			},

			ast.KindJsxElement: func(node *ast.Node) {
				element := node.AsJsxElement()
				if element == nil || element.OpeningElement == nil {
					return
				}
				tagName, _ := jsx.ElementParts(element.OpeningElement)
				if !bindsToOneOf(ctx, tagName, headImportClauses) {
					return
				}
				if element.Children == nil {
					return
				}

				for _, child := range element.Children.Nodes {
					if isScriptComponent(child) {
						ctx.ReportNode(tagName, messageNoScriptComponentInHead)
					}
				}
			},
		}
	},
}

// bindsToOneOf reports whether a JSX tag resolves to a binding declared by one of these clauses.
//
// A member expression tag (`<Foo.Head>`) parses as a property access rather than an identifier and
// is declined before any text is read, which matters twice: it is silent upstream, and `Text()`
// panics on that kind.
//
// **The identifier test here is defense in depth rather than a discrimination, and that was
// measured rather than assumed.** Removing it survived the whole fixture set, so the branch was
// probed directly: a JSX tag name can only be `KindIdentifier`, `KindPropertyAccessExpression`, or
// `KindJsxNamespacedName`, and the checker resolves neither of the latter two to an import clause,
// so the map lookup already declines them. It is kept because the exhaustion is a property of the
// parser rather than of this rule, and because `Text()` is reached below on the child tag where the
// same guard is genuinely load bearing. Do not read the surviving mutant as a fixture gap.
func bindsToOneOf(ctx rule.Context, tagName *ast.Node, clauses map[*ast.Node]bool) bool {
	if tagName == nil || tagName.Kind != ast.KindIdentifier || len(clauses) == 0 {
		return false
	}
	// A nil checker is not hypothetical and it does not degrade gracefully. The declaration above
	// asks for one, but the engine leaves it nil when the program could not be built, and the plain
	// test harness never has one at all. Dereferencing it panicked outright rather than going
	// silent, which takes the whole run down instead of dropping one rule's findings. Found by the
	// fixture written to assert the silence.
	if ctx.TypeChecker == nil {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(tagName)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	return clauses[symbol.Declarations[0]]
}

// isScriptComponent reports whether a JSX child is an element whose tag is literally `Script`.
//
// Text comparison rather than resolution, deliberately, and the doc on the rule says why. Both
// element kinds are read because our parser splits what oxc keeps as one node.
func isScriptComponent(child *ast.Node) bool {
	if child == nil {
		return false
	}

	var tagName *ast.Node
	switch child.Kind {
	case ast.KindJsxElement:
		element := child.AsJsxElement()
		if element == nil || element.OpeningElement == nil {
			return false
		}
		tagName, _ = jsx.ElementParts(element.OpeningElement)
	case ast.KindJsxSelfClosingElement:
		tagName, _ = jsx.ElementParts(child)
	default:
		return false
	}

	// The identifier check is the same guard as above and for the same two reasons: `<Foo.Script>`
	// is silent upstream, and reading text off a property access panics.
	return tagName != nil && tagName.Kind == ast.KindIdentifier && tagName.Text() == "Script"
}
