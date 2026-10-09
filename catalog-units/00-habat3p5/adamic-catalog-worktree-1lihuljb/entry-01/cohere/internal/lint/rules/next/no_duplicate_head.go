package next

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoDuplicateHead = rule.Message{
	Id: "noDuplicateHead",
	Description: "This file renders the imported `Head` component more than once. Next.js merges " +
		"one document head, so a second `Head` does not add to the first, it competes with it: " +
		"whichever the framework resolves last wins and the other one's tags are silently " +
		"dropped. Render a single `Head` and put every tag inside it.",
}

// NoDuplicateHead flags a file that renders an imported `Head` component more than once.
//
//	valid:   import { Head } from 'next/document'; <Html><Head/></Html>
//	valid:   import { Head } from 'next/document'; <Html><Head><meta/></Head></Html>
//	invalid: import Head from 'next/head'; <Html><Head/><Head/><Head/></Html>
//	invalid: import Head from 'next/head'; <Html><Head>a</Head><body/><Head>b</Head></Html>
//
// Ported from `@next/next/no-duplicate-head`, read against oxc's `no_duplicate_head.rs` and against
// `@next/eslint-plugin-next`'s own `no-duplicate-head.js`. The two sources implement substantially
// different rules and oxc is followed throughout, because oxc is what the gate being replaced runs.
// Every behaviour below was measured on the release oxlint binary rather than inferred, because the
// imported corpus is four cases and pins almost none of it.
//
// # This is not a JSX rule upstream, and that changes what it counts
//
// The name suggests walking JSX and counting elements called `Head`. oxc does not do that. It
// anchors on `ImportDeclaration`, finds the specifier whose local name is `Head`, resolves it to a
// symbol, and walks the semantic reference table counting reads whose parent is a JSX opening
// element. The two mechanisms agree only when every `Head` in the file is both imported and used
// directly as a tag, which is every case the corpus ships, so the corpus cannot tell them apart.
//
// Three consequences follow, and a JSX-counting port gets all three wrong:
//
//   - **The module specifier is never checked.** `import { Head } from './my-own-widgets'` arms the
//     rule exactly as `next/document` does. Measured: reports. This is upstream firing well outside
//     the scope its own message describes, and it is reproduced rather than narrowed, because
//     narrowing it would make us silent on files the gate reports.
//   - **A `Head` that is not an import never counts.** `const Head = () => null` used twice is
//     silent, and so is a `Head` shadowed inside a function while an unused import sits at the root.
//     Both measured.
//   - **A tag that is not a plain identifier never counts.** `<Head.Sub/><Head.Sub/>` is silent,
//     because the reference's parent is the member expression rather than the opening element.
//     Measured, and it is also why the identifier guard below is load-bearing rather than defensive:
//     reading `Text()` off that kind panics.
//
// # Which name matches, and the recorded research had this backwards
//
// oxc reads `specifiers.iter().find(|s| s.name() == "Head")`, and `ImportDeclarationSpecifier::name`
// is defined as `self.local().name` at `oxc_ast/src/ast_impl/js.rs:2034`. It is the **local** name,
// not the imported one. So `import { Head as PageHead } from 'next/document'` with two `<PageHead/>`
// is **silent** (no specifier is locally named `Head`), and `import { Header as Head }` with two
// `<Head/>` **reports** (the local name is `Head`). Both measured on the binary, and both are the
// opposite of what this rule's own research pass recorded from reading the accessor's name. The
// fixtures below pin both directions, because this is the axis nothing upstream tests and the
// intuitive reading is the wrong one.
//
// Matching the local name is also why the specifier's *declaration* is what identity is compared
// against below rather than its imported name: a default import, a namespace import and a named one
// all bind a local `Head`, and all three arm the rule. The namespace form was measured reporting.
//
// # One finding per file, not one per pair and not one per extra copy
//
// oxc accumulates labels and calls `ctx.diagnostic` once, so three copies produce **one** finding
// carrying three underlines and four copies produce one carrying four. Measured on the binary at
// three and at four. This is worth stating because both other readings are natural designs that
// agree with the corpus: the corpus's only multi-copy case is the three-copy fail, and a per-extra
// reading would give it two while a pairwise reading would give it three. `jsx-no-duplicate-props`
// in this tree answers a structurally similar question the *other* way, once per extra occurrence,
// so its shape is worth copying here and its count is not.
//
// Our harness carries one range per finding rather than a label list, so the reported range is the
// first occurrence, which is the position oxc renders as the diagnostic's own: the snapshot for the
// three-copy fail prints `9:19`, the first `<Head`. The range is the tag name alone, four bytes,
// not the element.
//
// # Two element kinds where oxc has one
//
// oxc listens on `JSXOpeningElement` and gets both shapes, because its parser gives a self-closing
// element an opening element carrying a flag. Our parser gives it a distinct `JsxSelfClosingElement`
// containing no opening element, and **every** `<Head />` in the corpus is self-closing, so
// listening on the literal translation of oxc's node would silence one of two fail cases and both
// pass cases while leaving the rule looking implemented.
//
// A closing tag does not count in either tree. `<Head>x</Head>` is one occurrence, measured silent,
// and a pair of them reports with two labels rather than four. Ours falls out of the same shape:
// `KindJsxClosingElement` is not listened on.
//
// # Whole file, any depth, any component
//
// The count spans the file. Two `Head` in two unrelated components report, and so do two nested at
// different depths under different parents. Both measured. ESLint's rule filters the *direct
// children* of one returned element and requires a class extending the `next/document` default
// import, so it is silent on both; that is the sharpest disagreement between the two upstreams and
// oxc's answer is the one taken. The nested case is pinned by a fixture below because nothing
// upstream covers it and the narrower reading is the one a reader expects.
//
// # No path gate exists, and that is upstream's defect reproduced
//
// The message names `pages/_document.js` and the rule runs on every file. Verified twice: the rule
// file contains no `file_path`, `file_name`, `_document`, or `is_document_page`, and a file at an
// arbitrary path reports on the binary. This is the inverse of the failure mode a broken-shut gate
// produces: the rule is over-broad rather than inert, and any file importing something named `Head`
// and rendering it twice reports with a message about a file the author may not have.
//
// It is reproduced rather than corrected, and the reason is that correcting it is not a narrowing we
// can defend as fidelity: a `pages/_document` test would silence inputs the gate currently reports,
// which is a difference the differential harness would read as ours being wrong. The four upstream
// spellings of that path test disagree with each other anyway. Recorded here rather than fixed so
// the next reader finds the decision rather than re-deriving it.
var NoDuplicateHead = rule.Rule{
	// No family prefix. The config writes `nextjs/no-duplicate-head` and matching strips the
	// namespace on a `/` boundary, so a prefixed name matches nothing and runs on no files.
	Name: "@next/next/no-duplicate-head",
	// Symbol identity is what oxc's reference walk answers, and nothing syntactic reproduces it:
	// the shadow cases and the member-tag case both turn on which binding a tag resolves to rather
	// than on how it is spelled. Two siblings in this package resolve JSX tags the same way.
	NeedsTypeChecker: true,
	NoListener:       rule.NoListenerDeclinesIrrelevantFiles,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Every declaration node that binds a local name `Head` at the module root. Keyed by node
		// so identity is compared rather than kind: two files can each declare a `Head` and a tag
		// resolving to the wrong one must not count.
		headDeclarations := map[*ast.Node]bool{}

		for _, statement := range ctx.SourceFile.Statements.Nodes {
			// Root scope only, which is oxc's `symbol_scope_id != root_scope_id` guard. Walking
			// statements rather than the whole tree is that guard: an import is only legal at the
			// module root, so a nested one cannot be reached from here in valid source.
			if statement.Kind != ast.KindImportDeclaration {
				continue
			}
			// A type-only import binds no value and is silent upstream, measured. Its specifier
			// still parses, so the clause has to be asked rather than the binding assumed.
			if declaration := statement.AsImportDeclaration(); declaration != nil &&
				declaration.ImportClause != nil {
				if clause := declaration.ImportClause.AsImportClause(); clause != nil &&
					clause.IsTypeOnly() {
					continue
				}
			}

			// `BindingsOf` answers with the local *identifier* for a default and a namespace import
			// and with the *specifier node* for a named one, which is the right answer to the
			// question it is asked and the wrong key for this map. The checker resolves a tag to
			// the node that declares the binding, and the three forms declare in three different
			// places: a default import's declaration is the enclosing `ImportClause`, a namespace
			// import's is the `NamespaceImport` node, and a named one's is the `ImportSpecifier`.
			// Keying on the identifier matched none of the first two, which is how this was found:
			// both upstream fail cases use a default import and both went silent while the invented
			// named-import cases passed.
			bindings := imports.BindingsOf(statement)
			declarationOf := func(binding *ast.Node) *ast.Node {
				if binding == nil {
					return nil
				}
				switch binding {
				case bindings.Default:
					// The identifier's parent, which is the clause the checker names.
					return binding.Parent
				case bindings.Namespace:
					// Already the `NamespaceImport` node rather than the identifier inside it.
					return binding
				}
				return binding
			}
			for _, binding := range append([]*ast.Node{bindings.Default, bindings.Namespace}, bindings.Named...) {
				if binding == nil {
					continue
				}
				// A named specifier and a namespace import carry their local name in `Name()`; a
				// default binding is already that identifier. `ImportedNameOf` is deliberately not
				// used: oxc matches the specifier's LOCAL name, so an aliased import is matched on
				// its alias and never on the name the source module exports.
				local := binding
				if binding.Kind == ast.KindImportSpecifier || binding.Kind == ast.KindNamespaceImport {
					local = binding.Name()
				}
				if local == nil || local.Kind != ast.KindIdentifier || local.Text() != "Head" {
					continue
				}
				if declaration := declarationOf(binding); declaration != nil {
					headDeclarations[declaration] = true
				}
			}
		}

		// A cost guard rather than a discrimination, and stated as such because a mutant removing it
		// survives the whole fixture set and reads like a blind spot. It is genuinely equivalent: an
		// empty map makes `resolvesToDeclaration` answer false for every tag, so `occurrences` never
		// leaves zero and the report is unreachable. No input can distinguish the two versions,
		// which is why no fixture was added for it. What it buys is real, though: a file importing
		// nothing called `Head` is most files, and this skips walking their whole tree.
		if len(headDeclarations) == 0 {
			return nil
		}

		// Gathered across the whole file and judged once, because the finding is one per file
		// rather than one per occurrence and the first occurrence is what it points at. There is no
		// exit hook and the walk is pre-order, so the source file listener does both: it fires
		// before its children, which is what lets the report happen after the walk it runs itself.
		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				var first *ast.Node
				occurrences := 0

				var visit func(current *ast.Node) bool
				visit = func(current *ast.Node) bool {
					switch current.Kind {
					case ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement:
						tagName, _ := jsx.ElementParts(current)
						// A member-expression tag is declined before any text is read. Upstream is
						// silent on it because the reference's parent is the property access rather
						// than the opening element, and reading `Text()` off that kind panics here.
						if tagName != nil && tagName.Kind == ast.KindIdentifier &&
							tagName.Text() == "Head" && resolvesToDeclaration(ctx, tagName, headDeclarations) {
							occurrences++
							if first == nil {
								first = tagName
							}
						}
					}
					current.ForEachChild(visit)
					return false
				}
				node.ForEachChild(visit)

				// One `Head` is the point of the component. Two or more is the finding, and it is
				// one finding however many there are.
				if occurrences > 1 {
					ctx.ReportNode(first, messageNoDuplicateHead)
				}
			},
		}
	},
}

// resolvesToDeclaration reports whether a JSX tag binds to one of the gathered import declarations.
//
// Node identity rather than declaration kind, which is the distinction that survives a shadow: a
// `const Head` inside a function and an imported `Head` at the root are two declarations and a tag
// under the function resolves to the inner one. Comparing kinds, or comparing text, would count it.
func resolvesToDeclaration(ctx rule.Context, tagName *ast.Node, declarations map[*ast.Node]bool) bool {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(tagName)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declarations[declaration] {
			return true
		}
	}
	return false
}
