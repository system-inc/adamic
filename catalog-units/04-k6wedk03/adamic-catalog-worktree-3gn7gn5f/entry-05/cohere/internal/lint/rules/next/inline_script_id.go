package next

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageInlineScriptId = rule.Message{
	Id: "inlineScriptId",
	Description: "This renders a next/script component with inline content but no id attribute. " +
		"The framework deduplicates inline scripts by that id, so without one the same script is " +
		"re-injected on every navigation and any one-time setup it performs runs again, which shows " +
		"up as doubled analytics events or global state clobbered after a client transition. Give " +
		"the script a stable id attribute.",
}

// InlineScriptId flags a next/script component carrying inline content without an id attribute.
//
//	valid:   import Script from 'next/script'; <Script id="a">{`x()`}</Script>
//	valid:   import Script from 'next/script'; <Script src="https://example.com" />
//	valid:   import Script from 'next/script'; <Script {...{ id: "a" }}>{`x()`}</Script>
//	valid:   function Script(){}; <Script>{`x()`}</Script>
//	invalid: import Script from 'next/script'; <Script>{`x()`}</Script>
//	invalid: import Script from 'next/script'; <Script dangerouslySetInnerHTML={{__html:"x"}} />
//
// Ported from oxc's `inline_script_id.rs`, whose corpus is eight pass cases and four fail cases
// carrying four diagnostics, one per input. The corpus is strong on spreads and silent on every
// negative form of the import gate, so most of what follows was measured against the release oxlint
// binary rather than read off a fixture. Each probe states its verdict.
//
// # The import gate is a real resolution here, and this rule is the opposite of its sibling
//
// oxc anchors on the `ImportDefaultSpecifier` of `next/script` and walks
// `ctx.semantic().symbol_references(...)`, so the elements it inspects are the ones whose tag
// actually binds to that import. That makes `imports.LocalNameOfDefaultImport` the faithful helper
// rather than a convenient one, and the distinction matters because the neighbouring
// `no-before-interactive-script-outside-document` reaches the opposite conclusion in this same
// package: its upstream `find_map`s the module record and binds a named or namespace import too, so
// that rule declines the helper and matches text. Two rules over the same module, two different
// gates, written by different hands upstream. Probed here rather than assumed:
//
//	function Script(){}; <Script>{`x()`}</Script>              silent   a local component never binds
//	import S from 'next/script'; <S>{`x()`}</S>                REPORTS  the alias is what resolves
//	import { X } from 'next/script'; <X>...</X>                silent   a named import does not bind
//	import * as Script from 'next/script'; <Script>...         silent   a namespace import does not
//	import Script, { X } from 'next/script'; <Script>...       REPORTS  the default half still binds
//	import Script from 'other/script'; <Script>...             silent   compared exactly
//	<Script.Inner>{`x()`}</Script.Inner>                       silent   a member tag is not the binding
//
// `imports.LocalNameOfDefaultImport`'s doc comment invites a port to resolve the import instead of
// string matching, calling upstream's string match a defect. On this rule the invitation happens to
// describe upstream, so it is taken; on the sibling it is declined. The helper's guidance is
// therefore not a general rule in either direction, and which way it falls has to be measured per
// rule against the reference implementation.
//
// Resolution is approximated by name rather than by symbol, because we have no reference walk. The
// two answers separate only when a `next/script` default binding is shadowed by an inner
// declaration of the same name, which no corpus case and no probe here exercises. That gap is
// stated rather than papered over: a shadowed `Script` would report here and be silent upstream.
//
// A second `next/script` import in one file is handled the way oxc handles it, which is that both
// bindings arm the rule. Probed: with `import Script` and `import Other` both from `next/script`,
// two inline elements report twice. The eslint original keeps a single module-scoped `let` and so
// tracks only the last import, which is a real disagreement between the two upstreams, and oxc is
// the port target.
//
// # What counts as inline, measured rather than reasoned
//
// Either children or `dangerouslySetInnerHTML` arms the report, independently, and an `id` from any
// source disarms it. `src` does not exempt anything, which is worth pinning because the rule's
// documentation talks about external scripts and the only passing external fixture also happens to
// have no children:
//
//	<Script src="https://e.com">{`x()`}</Script>               REPORTS  src exempts nothing
//	<Script></Script>                                          silent   zero children is not inline
//	<Script>\n  </Script>                                      REPORTS  whitespace is a child
//
// The whitespace case is upstream's `!children.is_empty()` reading a JsxText node that holds only a
// newline and spaces, and our parser produces that node too, so the verdicts agree. It is pinned
// below because nothing in the corpus covers it and because `ast.IsWhitespaceOnlyJsxText` exists on
// the shelf and would be the obvious thing to reach for: it would silence this input and diverge.
//
// # Spreads are read, which is the part of this rule a reader will not expect
//
// A spread is a real `KindJsxSpreadAttribute` node rather than an absence, and unlike most rules in
// this package this one looks inside it. An object literal contributes its keys to the same name
// set the written attributes go into, so a spread can supply the `id` that disarms the rule.
// Anything else is opaque, and an opaque spread abandons the element entirely rather than merely
// being skipped: upstream spells that as a labelled `continue 'references_loop`, eslint as an early
// return citing vercel/next.js#34030 with the comment that non-checkable spreads are simply
// ignored. The difference is observable and is the corpus's one untested half:
//
//	<Script {...{ id: "a" }}>{`x()`}</Script>                  silent   the spread supplies id
//	<Script {...({ id: "a" })}>{`x()`}</Script>                silent   without_parentheses is real
//	<Script {...{ id }}>{`x()`}</Script>                       silent   shorthand is a static key
//	<Script {...{ "id": "a" }}>{`x()`}</Script>                REPORTS  a string key is not one
//	<Script {...{ ["id"]: "a" }}>{`x()`}</Script>              REPORTS  nor is a computed key
//	<Script {...{ ...o }}>{`x()`}</Script>                     REPORTS  a nested spread contributes nothing
//	<Script {...opaque}>{`x()`}</Script>                       silent   the whole element is abandoned
//	<Script {...opaque} dangerouslySetInnerHTML={{}} />        silent   abandoned before dSIH is read
//
// Upstream matches `PropertyKey::StaticIdentifier` only, which is why the string and computed keys
// report even though a reader would call them the same property. `ast.TryGetTextOfPropertyName`
// resolves all four shapes and is deliberately not used here for that reason: it would silence the
// last three and read as a correctness improvement while being a divergence.
//
// A namespaced attribute name (`x:id`) reports, matching `JSXAttributeName::Identifier`, and
// `jsx.AttributeName` already declines that shape.
//
// The finding points at the tag name, taken from the snapshot, whose caret spans the six characters
// of `Script` and the eight of `MyScript` rather than the whole element. The eslint original reports
// the whole JsxElement, and that is a disagreement in which oxc is the port target.
//
// No fix. Upstream ships none, and the repair is a name only the author can choose.
var InlineScriptId = rule.Rule{
	// No namespace prefix. The config writes `nextjs/inline-script-id` and the parity guard strips
	// the namespace on a `/` boundary, so a prefixed name would match no inventory entry and lint no
	// files while every fixture in this package stayed green.
	Name: "@next/next/inline-script-id",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		scriptNames := defaultImportNamesOf(ctx.SourceFile, "next/script")
		if len(scriptNames) == 0 {
			// Nothing in this file binds the framework's script component, so no element can
			// report. Upstream reaches the same conclusion by finding no import specifier to
			// anchor on.
			//
			// This is a cost decision rather than a discrimination, and it is recorded as one
			// because a mutation sweep will find it surviving. Deleting the early return changes
			// no verdict: the name membership test below declines every element in a file that
			// binds nothing, so no input distinguishes the two versions. What it does change is
			// whether two listeners are registered and the whole JSX tree walked for a file that
			// can never report, which on these trees is nearly every file. Inverting it, rather
			// than deleting it, is caught by twenty six lines.
			return nil
		}

		// Both JSX forms, because our parser splits what oxc emits as one JSXOpeningElement. The
		// two take different paths rather than the same one twice: only a paired element can carry
		// children, and only the self-closing form is guaranteed not to, so the children argument
		// differs and the `dangerouslySetInnerHTML` branch is the only route a self-closing element
		// has. Two of the four fail fixtures are self-closing, so a port registering one kind loses
		// half the corpus.
		return rule.Listeners{
			ast.KindJsxElement: func(node *ast.Node) {
				element := node.AsJsxElement()
				if element.OpeningElement == nil {
					return
				}
				opening := element.OpeningElement.AsJsxOpeningElement()
				childCount := 0
				if element.Children != nil {
					childCount = len(element.Children.Nodes)
				}
				reportInlineScriptWithoutId(
					ctx, scriptNames, opening.TagName, opening.Attributes, childCount > 0,
				)
			},

			ast.KindJsxSelfClosingElement: func(node *ast.Node) {
				element := node.AsJsxSelfClosingElement()
				reportInlineScriptWithoutId(
					ctx, scriptNames, element.TagName, element.Attributes, false,
				)
			},
		}
	},
}

// reportInlineScriptWithoutId decides one element and reports on its tag name.
//
// The order is upstream's: resolve the tag, collect every attribute name the element writes, let an
// `id` disarm before the inline test is asked, and only then check the two things that make an
// element inline. Asking the inline test first would be cheaper and would change nothing
// observable, but it would also stop reading like the original, which is the thing a later reader
// checks this against.
func reportInlineScriptWithoutId(
	ctx rule.Context,
	scriptNames map[string]bool,
	tagName *ast.Node,
	attributes *ast.Node,
	hasChildren bool,
) {
	// A member expression or namespaced tag is silent upstream, and reading text off a
	// KindPropertyAccessExpression, which is what a dotted JSX tag parses to, is a live panic in
	// this shim. The kind decides before the text is touched.
	if tagName == nil || tagName.Kind != ast.KindIdentifier || !scriptNames[tagName.Text()] {
		return
	}

	attributeNames, checkable := attributeNamesIncludingObjectSpreads(attributes)
	if !checkable {
		// An opaque spread abandons the element rather than being skipped, so an element carrying
		// one never reports whatever else it writes. This is vercel/next.js#34030 and both
		// upstreams implement it, by a labelled continue and by an early return respectively.
		return
	}
	if attributeNames["id"] {
		return
	}
	if !hasChildren && !attributeNames["dangerouslySetInnerHTML"] {
		return
	}
	ctx.ReportNode(tagName, messageInlineScriptId)
}

// attributeNamesIncludingObjectSpreads returns the attribute names an element writes, reading the
// static keys of object-literal spreads as if they had been written inline.
//
// The second return is false when the element carries a spread whose argument is not an object
// literal, meaning the set is not a complete answer and the caller must abandon the element rather
// than treat a missing name as absent. That distinction is the whole reason this returns two values:
// `jsx.HasAttributeNamed` deliberately skips spreads and would answer "no id" for an element whose
// id arrives through one, which inverts the verdict on two of the eight passing corpus cases.
//
// Kept local rather than lifted onto the jsx shelf. The sibling that also reads `next/script`
// attributes matches only written attributes and is silent on every spread form, so there is one
// caller, and a shared helper whose only caller is this rule would invite the sibling to adopt it
// and silently change its verdicts.
func attributeNamesIncludingObjectSpreads(attributes *ast.Node) (map[string]bool, bool) {
	names := map[string]bool{}
	if attributes == nil || attributes.Kind != ast.KindJsxAttributes {
		return names, true
	}
	properties := attributes.AsJsxAttributes().Properties
	if properties == nil {
		return names, true
	}

	for _, property := range properties.Nodes {
		switch property.Kind {
		case ast.KindJsxAttribute:
			// `jsx.AttributeName` declines a spread, a namespaced name and a nil name, which is
			// the same three-part guard upstream expresses by destructuring
			// `JSXAttributeName::Identifier`.
			if name, named := jsx.AttributeName(property); named {
				names[name] = true
			}

		case ast.KindJsxSpreadAttribute:
			expression := property.AsJsxSpreadAttribute().Expression
			// oxc calls `.without_parentheses()` before matching, so `{...({a: 1})}` is checkable
			// rather than opaque. Measured: it is silent when the parenthesized object supplies id.
			expression = ast.SkipParentheses(expression)
			if expression == nil || expression.Kind != ast.KindObjectLiteralExpression {
				return names, false
			}
			collectStaticIdentifierKeys(expression, names)
		}
	}
	return names, true
}

// collectStaticIdentifierKeys adds the plain identifier keys of an object literal to a name set.
//
// Only `PropertyKey::StaticIdentifier` counts upstream, so three shapes a reader would call the
// same property contribute nothing, and all three were measured as reporting: a string key
// (`{"id": x}`), a computed key (`{["id"]: x}`), and a nested spread (`{...other}`). Shorthand
// (`{id}`) does count, because its key is an identifier.
//
// `ast.TryGetTextOfPropertyName` resolves the string and computed forms and is not used for exactly
// that reason. Reaching for it would silence inputs upstream reports, which reads as a correctness
// improvement and is a divergence.
func collectStaticIdentifierKeys(objectLiteral *ast.Node, names map[string]bool) {
	properties := objectLiteral.AsObjectLiteralExpression().Properties
	if properties == nil {
		return
	}
	for _, property := range properties.Nodes {
		// A spread, a method, and an accessor are all not ObjectProperty upstream and contribute
		// nothing. Shorthand is a separate kind here and is included because upstream sees its key
		// as a static identifier.
		if property.Kind != ast.KindPropertyAssignment &&
			property.Kind != ast.KindShorthandPropertyAssignment {
			continue
		}
		name := property.Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			names[name.Text()] = true
		}
	}
}

// defaultImportNamesOf returns every local name the file binds as a module's default import.
//
// A set rather than a single name, because oxc anchors on each `ImportDefaultSpecifier` separately
// and walks that symbol's references, so two imports of the same module both arm the rule. Measured
// against the release binary: two `next/script` default imports in one file produce two findings.
// The eslint original overwrites one module-scoped variable and would find only the second, which
// is a disagreement between the upstreams rather than an accident of ours.
//
// This is a whole-file scan in `Run` rather than a listener that sets a variable, because our walk
// is single pass and ordered while oxc reads a module record that does not care where the import
// sits. Probed: an import written textually after the JSX still arms the rule upstream.
func defaultImportNamesOf(sourceFile *ast.SourceFile, specifier string) map[string]bool {
	names := map[string]bool{}
	if sourceFile == nil || sourceFile.Statements == nil {
		return names
	}
	for _, statement := range sourceFile.Statements.Nodes {
		if name, found := imports.LocalNameOfDefaultImport(statement, specifier); found {
			names[name] = true
		}
	}
	return names
}
