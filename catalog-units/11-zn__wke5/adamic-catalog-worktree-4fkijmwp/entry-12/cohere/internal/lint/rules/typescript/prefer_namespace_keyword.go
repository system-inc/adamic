package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messagePreferNamespaceKeyword = rule.Message{
	Id: "preferNamespaceKeyword",
	Description: "This declares a namespace with the `module` keyword, which TypeScript has " +
		"deprecated and plans to make a parse error. It means the same thing as `namespace` and " +
		"reads as an ECMAScript module, which it is not. Write `namespace` instead.",
}

// PreferNamespaceKeyword flags a namespace declared with the `module` keyword.
//
//	valid:   namespace Example {}
//	valid:   declare namespace Example {}
//	valid:   declare module 'foo' {}
//	valid:   declare module 'foo';
//	valid:   declare global {}
//	invalid: module Example {}
//	invalid: declare module Example {}
//	invalid: module A.B {}
//
// Ported from `@typescript-eslint/prefer-namespace-keyword` by way of oxc's
// `typescript/prefer-namespace-keyword`. No options: oxc declares none and the corpus carries no
// case with a second tuple element.
//
// # What the rule decides, and why our version has to work much harder than oxc's
//
// The judgment is one line long: a namespace written with `module` should be written with
// `namespace`. Everything below exists because oxc gets its exemptions from its grammar and we have
// to perform them by hand.
//
// oxc parses these into three unrelated node types. `TSNamespaceDeclaration` is the
// identifier-named form and carries an explicit `kind` field holding `Module` or `Namespace`.
// `TSExternalModuleDeclaration` is the quoted form, `declare module 'foo'`. `TSGlobalDeclaration`
// is `declare global`. Its rule anchors on the first of those, so the other two are exempt by
// never arriving, and its `kind` field answers the keyword question outright.
//
// typescript-go folds all three into one `KindModuleDeclaration`, and it does not expose the
// keyword. TypeScript's own `NodeFlagsNamespace` is what records "written with the namespace
// keyword", and it is not reachable through our shim: `ast.NodeFlagsNamespace` fails to compile,
// which was checked with a probe rather than with a grep, because a grep cannot see through the
// shim's type aliases. Measured on the parse, `module foo {}` and `namespace foo {}` both come back
// with flags `0x0`, and `declare module foo {}` and `declare namespace foo {}` both come back with
// `NodeFlagsAmbient` alone. Nothing on the node distinguishes them.
//
// So the keyword is recovered from the source text, and each of oxc's three free exemptions becomes
// a guard below with a fixture standing over it.
//
// # Why the keyword is read with the scanner rather than searched for
//
// `declare /* module */ module foo {}` is an upstream fix vector, and it exists to punish a rule
// that searches its source for the text `module`: the first hit is inside the comment. Asking the
// scanner for the token at a position skips trivia by construction, so the comment is never a
// candidate. oxc needs `find_next_token_within` for the same reason.
//
// # Which nesting is exempt, and which is not
//
// The two nestings look alike in source and parse differently, and the corpus contains both.
//
// `module A.B {}` is one statement with one `module` keyword. It parses as a ModuleDeclaration for
// `A` whose body is another ModuleDeclaration for `B`, so the tree holds two nodes where the source
// holds one keyword. Upstream reports it once, and the snapshot pins that. Every link after the
// head is skipped.
//
// `declare module foo { module bar {} }` is two statements with two keywords, nested through a
// ModuleBlock. Upstream reports it twice, which the snapshot also pins: six diagnostics across five
// failing inputs, and this is the input supplying the sixth.
//
// So the exemption is keyed on the PARENT being a ModuleDeclaration, never on the node being
// nested. A guard reading "skip anything inside another module" would silence the second case and
// still pass every fixture that only covers the first.
//
// That exemption is also load-bearing beyond fidelity, which the upstream fix vector
// `declare module X.Y.module { let x: 'module'; }` demonstrates. The third link of that chain is
// named `module`, so the token beginning it is the identifier `module`. A keyword test applied to
// a nested link would read that name as the keyword and rewrite a name segment into `namespace`,
// producing source that parses, does not mean the same thing, and is applied unattended.
//
// # The span, which is not the node
//
// oxc reports the `TSNamespaceDeclaration` span. `declare` is inside that node and `export` is not,
// because an exported declaration is wrapped in a separate node there. Measured on the release
// binary: `declare module foo {}` reports at column 1, and `export module foo {}` reports at column
// 8 rather than column 1.
//
// typescript-go puts both modifiers inside the declaration, so matching that means starting the
// span after an `export` modifier and before a `declare` one. The corpus contains no exported
// declaration, so nothing upstream covers this and only the fixtures written here hold it.
var PreferNamespaceKeyword = rule.Rule{
	Name: "@typescript-eslint/prefer-namespace-keyword",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Defensive at the boundary, and deliberately untested. Mutating this guard away survives
		// the whole fixture set, and no fixture can change that: rule_testing.Run always builds a real
		// program, so no input reachable through the harness produces a nil SourceFile. That makes
		// it unreachable through the harness rather than a fixture blind spot, and a test written
		// for it would assert nothing. Kept because everything below reads the source text, and
		// because 46 of the 184 rule files here open the same way.
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindModuleDeclaration: func(node *ast.Node) {
				// A quoted name is an ambient external module: `declare module 'foo'` describes
				// somebody else's package and is not a namespace at all, so rewriting it would
				// change what the file declares. oxc gives this its own node type and never sees
				// it here.
				//
				// Measured on the release binary, the quoted name is what exempts it rather than
				// the `declare`: `module "foo" {}` is silent and `declare module foo {}` reports.
				name := node.Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					return
				}

				// A link after the head of a dotted chain. `module A.B {}` holds one keyword and
				// two nodes, so only the head is judged. The parent test is what separates this
				// from a module nested through a ModuleBlock, which upstream reports.
				if node.Parent != nil && node.Parent.Kind == ast.KindModuleDeclaration {
					return
				}

				// The token that opens the declaration proper, past any modifiers and any trivia
				// between them.
				keyword := moduleDeclarationKeywordRange(ctx.SourceFile, node)
				sourceText := ctx.SourceFile.Text()
				if keyword.End() > len(sourceText) {
					return
				}

				// `declare global {}` opens with `global`, not with `module`, so this test exempts
				// it without naming it. Keying the exemption on the identifier would be wrong:
				// `module global {}` reports upstream, measured on the release binary.
				if sourceText[keyword.Pos():keyword.End()] != "module" {
					return
				}

				ctx.ReportRangeWithFixes(
					moduleDeclarationReportRange(ctx.SourceFile, node),
					messagePreferNamespaceKeyword,
					rule.ReplaceRange(keyword, "namespace"),
				)
			},
		}
	},
}

// moduleDeclarationKeywordRange is the span of the `module` or `namespace` token that opens a
// declaration.
//
// The token begins after the last modifier, and asking the scanner for the token at that position
// skips whatever trivia sits between, which is what makes `declare /* module */ module foo {}` land
// on the keyword rather than on the comment.
//
// A modifier list on a nested link of a dotted chain holds a synthetic zero-width export keyword
// (measured: the `B` of `module A.B {}` reports `[KindExportKeyword 9:9]`), so the walk takes the
// furthest modifier end rather than assuming the list is meaningful. That shape cannot reach here
// while the parent guard above stands, and the walk is written not to depend on it.
func moduleDeclarationKeywordRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	start := node.Pos()
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier.End() > start {
				start = modifier.End()
			}
		}
	}
	return scanner.GetRangeOfTokenAtPosition(sourceFile, start)
}

// moduleDeclarationReportRange is the span oxc reports: the declaration including a `declare`
// modifier and excluding an `export` one.
//
// oxc gets this from its tree rather than from a decision, because `export` wraps the declaration
// in a separate node there while `declare` is a field on the declaration itself. Reproducing it
// here means naming the one modifier kind that sits outside the span upstream.
func moduleDeclarationReportRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	start := node.Pos()
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier.Kind == ast.KindExportKeyword && modifier.End() > start {
				start = modifier.End()
			}
		}
	}
	return scanner.GetRangeOfTokenAtPosition(sourceFile, start).WithEnd(node.End())
}
