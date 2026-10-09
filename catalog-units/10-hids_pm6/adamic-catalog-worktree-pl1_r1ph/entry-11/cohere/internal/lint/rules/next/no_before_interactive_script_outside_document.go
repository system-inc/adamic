package next

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/ecmascript/nextjs"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoBeforeInteractiveScriptOutsideDocument = rule.Message{
	Id: "noBeforeInteractiveScriptOutsideDocument",
	Description: "This loads a script with next/script's beforeInteractive strategy outside of " +
		"pages/_document.js. That strategy only has a place to run before hydration when the " +
		"framework injects it from the document, so anywhere else the script is either deferred " +
		"until after hydration or evaluated twice, and code depending on it having already run " +
		"sees a different order in development than in production. Move the script into " +
		"pages/_document.js, or pick a strategy that does not promise to run first.",
}

// NoBeforeInteractiveScriptOutsideDocument flags next/script's beforeInteractive strategy written
// anywhere other than the custom document.
//
//	valid:   pages/_document.js    import Script from 'next/script'; <Script strategy="beforeInteractive" />
//	valid:   app/layout.tsx        import Script from 'next/script'; <Script strategy="beforeInteractive" />
//	valid:   pages/index.js        import Script from 'next/script'; <Script strategy="afterInteractive" />
//	valid:   pages/index.js        function Script(){}; <Script strategy="beforeInteractive" />
//	invalid: pages/index.js        import Script from 'next/script'; <Script strategy="beforeInteractive" />
//	invalid: components/Thing.js   import Script from 'next/script'; <Script strategy="beforeInteractive" />
//
// Ported from oxc's `no_before_interactive_script_outside_document.rs`, whose corpus is seven pass
// cases and six fail cases carrying six diagnostics. The corpus is unusually strong on paths and
// unusually thin on everything else, so most of what this rule decides was measured against the
// release oxlint binary rather than read off a fixture. Every probe below states its verdict.
//
// # Which document predicate, and why this rule is the one that differs
//
// `internal/utilities/nextjs` carries two document predicates on purpose. `IsDocumentFile` is the union
// of every spelling the eight `@next/next` gating rules use, for the rules whose own upstreams
// disagree and whose corpora never vote. `IsDocumentPage` is oxc's single shared `is_document_page`
// reproduced byte for byte.
//
// **This rule is a caller of the shared helper, so `IsDocumentPage` is the faithful one**, and the
// choice is measured rather than argued. Against the release binary, with a `next/script` import and
// a `beforeInteractive` strategy in every file:
//
//	pages/_document.tsx           silent     both predicates agree
//	pages/_document/index.tsx     silent     both predicates agree
//	pages/_documentation.tsx      silent     IsDocumentPage agrees, IsDocumentFile would REPORT
//	components/_document.tsx      REPORTS    IsDocumentPage agrees, IsDocumentFile would silence
//	src/pages/user/_document.tsx  REPORTS    IsDocumentPage agrees, IsDocumentFile would silence
//	pages/index.jsx               REPORTS    both predicates agree
//
// `IsDocumentFile` disagrees with upstream on three of six. A research pass on this leaf recommended
// it, and also reported the `_documentation.tsx` direction inverted, claiming upstream exempts it
// where we would report. The opposite is true in the sense that matters: `is_document_page` tests
// `starts_with("/_document")` with no trailing dot, so upstream exempts `_documentation.tsx` and
// `IsDocumentFile`, which requires the dot, would have reported it. Both fixtures below pin the
// disagreement so a later tidy-up toward the union predicate fails rather than quietly widening.
//
// # The strategy attribute: a string literal and nothing else
//
// Three shapes were probed and all three are silent upstream, so all three are silent here:
//
//	<Script strategy={"beforeInteractive"} />          an expression container
//	<Script {...{strategy:"beforeInteractive"}} />     a spread holding the literal inline
//	<Script {...properties} />                         a spread holding it through a variable
//
// The spread cases are worth stating because a spread is not invisible in this AST the way a reader
// might assume: it parses to a `KindJsxSpreadAttribute` inside `JsxAttributes` and a rule looking
// for that kind would find it. Upstream's `attributes.iter().find(...)` matches only
// `JSXAttributeItem::Attribute`, so the spread is passed over rather than absent, and skipping it is
// the port rather than a limitation of what we can see. `<Script {...rest} strategy="beforeInteractive" />`
// reports, which is the same answer from the other direction: the spread is ignored and the real
// attribute still decides.
//
// The name is compared exactly. `<Script Strategy="beforeInteractive" />` is silent upstream, so
// this is an exact match rather than a case-insensitive one.
//
// # Which tag names count, and why the import is string-matched rather than resolved
//
// Upstream's `get_next_script_import_local_name` is a `find_map` over the module record's import
// entries, taking the local name of the **first** entry whose module request is `next/script`,
// whatever kind of entry it is. That is wider than a default import in one direction and narrower in
// another, and both were measured:
//
//	import { Foo } from 'next/script';  <Foo strategy=... />    REPORTS   a named import binds
//	import * as All from 'next/script'; <All strategy=... />    REPORTS   a namespace import binds
//	import { Foo as Bar } ...;          <Bar ... />             REPORTS   the LOCAL name is the one
//	import { Foo as Bar } ...;          <Foo ... />             silent    the imported name is not
//	import A from 'next/script';
//	import B from 'next/script';        <B ... />               silent    only the FIRST entry wins
//
// So `imports.LocalNameOfDefaultImport` is the wrong helper here, and its own doc comment is the
// reason to say so explicitly rather than silently. That comment calls upstream's string matching a
// defect and says a port resolving the import is more correct. Reaching for it on this rule would
// silence the named and namespace forms above, which upstream reports, so the invitation is declined
// and the decline is measured rather than stylistic.
//
// The match is textual, not a resolution, which cuts the other way too: a locally defined component
// named `Script` in a file whose real `next/script` import is bound to something else is silent,
// because the name that was bound is what is compared. Probed: with `import S from 'next/script'`
// alongside `function Script(){}`, a `<Script strategy="beforeInteractive" />` does not report. That
// looks like a defect and reproducing it is the port; the fixture below pins it.
//
// A member expression tag (`<Script.Inner />`) is silent, matching `get_identifier_name()`, which
// declines anything that is not a plain identifier. The kind guard is load bearing rather than
// defensive here for a second reason: `Text()` panics on a `KindPropertyAccessExpression`, which is
// what a dotted JSX tag parses to.
//
// # Both JSX kinds, and why the corpus alone would not have told us
//
// oxc emits one `JSXOpeningElement` for a self-closing element and for the opening half of a paired
// one, so a single registration covers both. Our parser splits them into `KindJsxSelfClosingElement`
// and `KindJsxOpeningElement`, so both are registered. Four of upstream's six fail cases are
// self-closing and two are paired, so the corpus does catch a port that registers only one, but the
// real reason is on our own tree: the single file in these checkouts that trips this rule writes
// both forms, two elements, one file.
//
// # Where the gates are asked
//
// The app directory gate is hoisted into `Run` and answered once, returning no listeners, because it
// is a property of the file rather than of a node. Upstream asks it at every opening element in the
// file, which is a cost rather than a decision, and fidelity is to what a rule decides. The document
// gate is file scoped too but is left where upstream puts it, after the strategy match, because
// hoisting it would change nothing observable and the ordering is easier to check against the
// original when it is written the same way.
//
// The app directory test is a substring test rather than a segment test, faithfully: any path
// containing `app/` is exempt, so `myapp/pages/x.jsx` is silent, probed and confirmed. The cost is
// real on our own trees rather than theoretical, and `nextjs.IsInApplicationDirectory` records which
// checkouts it silences.
//
// # The import scan runs before the walk
//
// Upstream reads a whole-file module record, so an import written textually after the JSX still
// binds. Probed: it reports. A listener on `KindImportDeclaration` that sets a variable would answer
// differently, because our walk is single pass and ordered, so the statements are scanned once in
// `Run` instead. That is the same decision the module record encodes, reached without one.
//
// The finding points at the `strategy` attribute node rather than at the element, taken from the
// snapshot, whose caret spans the twenty eight characters of `strategy="beforeInteractive"`. The
// eslint original reports the whole opening element, and that is a real disagreement in which oxc is
// the port target.
//
// No fix. Upstream ships none, and the repair is either a move to another file or a change to what
// the page does.
var NoBeforeInteractiveScriptOutsideDocument = rule.Rule{
	// No family prefix. The config writes `nextjs/no-before-interactive-script-outside-document` and
	// matching strips the namespace on a `/` boundary, so a prefixed name matches nothing and runs
	// on no files while its own tests stay green.
	Name: "@next/next/no-before-interactive-script-outside-document",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		fileName := ctx.SourceFile.FileName()

		// Hoisted, unlike upstream, which asks this at every opening element. Same answer, asked
		// once. Returning no listeners is this tree's spelling of upstream's `should_run`.
		if nextjs.IsInApplicationDirectory(fileName.AsString()) {
			return nil
		}

		scriptLocalName, imported := localNameOfFirstNextScriptImport(ctx.SourceFile)
		if !imported {
			// Nothing in this file can be the framework's script component, so nothing can report.
			// Upstream reaches the same conclusion later, after the strategy match, because its
			// module record costs nothing to consult; ours is a scan, so it happens once here.
			return nil
		}

		report := func(node *ast.Node) {
			tagName, attributes := jsx.ElementParts(node)
			// A member expression or namespaced tag is silent upstream, and reading text off a
			// property access panics, so the kind decides before the text is touched.
			if tagName == nil || tagName.Kind != ast.KindIdentifier || tagName.Text() != scriptLocalName {
				return
			}
			strategy := beforeInteractiveStrategyAttribute(attributes)
			if strategy == nil {
				return
			}
			if nextjs.IsDocumentPage(fileName.AsString()) {
				return
			}
			ctx.ReportNode(strategy, messageNoBeforeInteractiveScriptOutsideDocument)
		}

		return rule.Listeners{
			ast.KindJsxOpeningElement:     report,
			ast.KindJsxSelfClosingElement: report,
		}
	},
}

// localNameOfFirstNextScriptImport returns the name the file binds `next/script` to.
//
// The first import entry naming `next/script` wins whatever kind it is, which is what upstream's
// `find_map` over the module record does. A default binding, a namespace binding and a named
// binding are all entries, and within one declaration they are ordered as written, so a
// `import S, { Foo } from 'next/script'` binds `S` and a bare `import { Foo }` binds `Foo`. Each of
// those was probed against the release binary rather than inferred from the Rust.
//
// A side-effect import (`import 'next/script'`) introduces no entry and therefore binds nothing,
// which is why an empty local name answers false rather than matching a nameless tag.
func localNameOfFirstNextScriptImport(sourceFile *ast.SourceFile) (string, bool) {
	if sourceFile == nil || sourceFile.Statements == nil {
		return "", false
	}
	for _, statement := range sourceFile.Statements.Nodes {
		if statement.Kind != ast.KindImportDeclaration {
			continue
		}
		declaration := statement.AsImportDeclaration()
		if declaration == nil || declaration.ModuleSpecifier == nil {
			continue
		}
		if !ast.IsStringLiteralLike(declaration.ModuleSpecifier) {
			continue
		}
		// Compared exactly. `next/scripts` is a different module, and a substring test would bind
		// its local name to this rule.
		if declaration.ModuleSpecifier.Text() != "next/script" {
			continue
		}
		if name, found := firstBindingName(statement); found {
			return name, true
		}
	}
	return "", false
}

// firstBindingName returns the local name of an import declaration's first binding, in source order.
//
// `BindingsOf` is not used here because it answers a different question: it splits the three kinds
// into three fields, and the question this rule asks is which one came first, which the split
// discards. The order within a declaration is fixed by the grammar, so it is read directly: a
// default binding precedes named bindings and a namespace binding, and only one of the latter two
// can appear.
func firstBindingName(node *ast.Node) (string, bool) {
	declaration := node.AsImportDeclaration()
	if declaration == nil || declaration.ImportClause == nil {
		return "", false
	}
	clause := declaration.ImportClause.AsImportClause()
	if clause == nil {
		return "", false
	}

	if defaultBinding := clause.Name(); defaultBinding != nil && defaultBinding.Kind == ast.KindIdentifier {
		return defaultBinding.Text(), true
	}
	if clause.NamedBindings == nil {
		return "", false
	}

	switch clause.NamedBindings.Kind {
	case ast.KindNamespaceImport:
		name := clause.NamedBindings.AsNamespaceImport().Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.Text(), true
		}
	case ast.KindNamedImports:
		// `imports.BindingsOf` answers which shapes a statement carries and is the right call almost
		// everywhere, but it returns the three kinds as three separate fields and so discards the
		// one thing this rule needs: which binding came FIRST. Upstream's `find_map` over the module
		// record takes the first entry naming `next/script` whatever kind it is, so the order is the
		// judgment rather than an implementation detail, and a caller reassembling it from three
		// fields would be re-deriving what the grammar already fixed. Reached for directly for that
		// reason.
		elements := clause.NamedBindings.AsNamedImports().Elements
		if elements == nil {
			return "", false
		}
		for _, element := range elements.Nodes {
			// The LOCAL name, which for `{ Foo as Bar }` is `Bar`. Probed: `<Bar>` reports and
			// `<Foo>` is silent, so reading the imported name here would invert both.
			name := element.Name()
			if name != nil && name.Kind == ast.KindIdentifier {
				return name.Text(), true
			}
		}
	}
	return "", false
}

// beforeInteractiveStrategyAttribute returns the `strategy` attribute node when it is written as the
// string literal `beforeInteractive`, and nil otherwise.
//
// `jsx.StringAttributeValue` decides exactly this and declines exactly the same shapes, but it
// returns the text rather than the node, and this rule reports on the node. `jsx.AttributeName`'s
// own doc comment anticipates a third rule arriving with that need, so the loop is written here and
// the shared guard is called rather than rehandled.
//
// Upstream stops at the first attribute named `strategy` and reads its value, so a second one is
// never consulted. Written the same way here: the loop returns on the first name match whatever the
// value turns out to be.
func beforeInteractiveStrategyAttribute(attributes *ast.Node) *ast.Node {
	if attributes == nil || attributes.Kind != ast.KindJsxAttributes {
		return nil
	}
	properties := attributes.AsJsxAttributes().Properties
	if properties == nil {
		return nil
	}

	for _, property := range properties.Nodes {
		// A spread is a real node here, a `KindJsxSpreadAttribute`, rather than an absence. It is
		// skipped because upstream's `find` matches only a `JSXAttributeItem::Attribute`, and a
		// spread supplying `strategy` is silent upstream in both the inline and the variable form.
		if property.Kind != ast.KindJsxAttribute {
			continue
		}
		attributeName, named := jsx.AttributeName(property)
		if !named || attributeName != "strategy" {
			continue
		}
		// Only a plain string literal. An expression container is silent upstream, matching
		// `JSXAttributeValue::StringLiteral`.
		attribute := property.AsJsxAttribute()
		if attribute.Initializer == nil || attribute.Initializer.Kind != ast.KindStringLiteral {
			return nil
		}
		// Decoded as ESLint's parser decodes it, so `strategy="before&#73;nteractive"` is
		// beforeInteractive here as it is upstream.
		if text.UnescapeStringLiteralText(attribute.Initializer.Text()) != "beforeInteractive" {
			return nil
		}
		return property
	}
	return nil
}
