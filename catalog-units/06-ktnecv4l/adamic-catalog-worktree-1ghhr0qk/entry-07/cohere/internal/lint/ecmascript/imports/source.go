// Package imports answers where a file's imports come from, in every shape the language allows.
//
// Lifted out of `internal/rules/nexus/` where three boundary rules shared it and four `structure`
// rules hand-rolled the same concern because they could not see it. Its own doc comment already
// said several rules share this, which was true and unreachable.
package imports

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// A specifier can arrive in three shapes, and a rule that guards a boundary has to see all three or
// it guards nothing: the shape it misses is the one that gets used to route around it.
//
//	import Thing from '@project/Thing'   static
//	await import('@project/Thing')       dynamic
//	require('@project/Thing')            CommonJS
//
// The TypeScript originals learned this the same way, which is why several of them share one
// visitor factory rather than each writing the static case and forgetting the other two.

// SourceVisitors builds the listeners that call report for every string specifier in a file,
// whichever of the three shapes it arrives in.
//
// **Not every rule that reads an import wants this, and reaching for it by default would widen rules
// that are correctly narrow.** Four `structure` rules watch `KindImportDeclaration` alone and are
// right to: `react-import-no-destructuring` asks about the binding form of a static import, and a
// dynamic `import('react')` has no destructuring to flag. A rule guarding a *boundary* needs all
// three shapes because the missed one routes around it; a rule asking about how a static import is
// *written* needs exactly one.
//
// The test is whether the rule is about where a module comes from or about the syntax that brings it
// in. Only the first wants this.
//
// report is handed the specifier and the node to blame. The node differs by shape on purpose: a
// static import blames the whole declaration, while a dynamic or require call blames the call, since
// that is the expression a reader has to change.
func SourceVisitors(report func(source string, node *ast.Node)) rule.Listeners {
	return rule.Listeners{
		ast.KindImportDeclaration: func(node *ast.Node) {
			declaration := node.AsImportDeclaration()
			if declaration == nil || declaration.ModuleSpecifier == nil {
				return
			}
			if !ast.IsStringLiteralLike(declaration.ModuleSpecifier) {
				return
			}
			report(declaration.ModuleSpecifier.Text(), node)
		},
		ast.KindCallExpression: func(node *ast.Node) {
			source, isImport := CallExpressionSource(node)
			if !isImport {
				return
			}
			report(source, node)
		},
	}
}

// CallExpressionSource returns the string specifier of a dynamic import or a require call.
//
// Both are call expressions, so one listener serves both rather than two listeners on the same kind,
// which the listener map could not hold anyway.
// A nil node answers rather than panics, and that guard is new with the lift. Inside the rule
// package this was only ever reached from the visitor, which never hands it nil; on a shared shelf
// any rule can call it directly, and `ast.IsImportCall` dereferences without checking. A panic in a
// shared package takes the whole run down rather than one rule's finding.
//
// Found by the first fixture written for this file, which is the argument for lifting with tests
// rather than after: the code was correct for its callers and unsafe for its new ones.
func CallExpressionSource(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	isDynamicImport := ast.IsImportCall(node)
	isRequire := ast.IsRequireCall(node, true)
	if !isDynamicImport && !isRequire {
		return "", false
	}

	call := node.AsCallExpression()
	if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return "", false
	}

	// IsRequireCall already checked this, but IsImportCall does not look at arguments at all, so a
	// computed `import(someVariable)` reaches here and has no specifier to read.
	firstArgument := call.Arguments.Nodes[0]
	if !ast.IsStringLiteralLike(firstArgument) {
		return "", false
	}
	return firstArgument.Text(), true
}

// NormalizedFileName returns a path with forward slashes, so a rule matching a path fragment finds
// it on Windows too.
func NormalizedFileName(sourceFile *ast.SourceFile) string {
	if sourceFile == nil {
		return ""
	}
	return strings.ReplaceAll(sourceFile.FileName().AsString(), "\\", "/")
}

// SpecifierNode returns the string-literal specifier inside an import-shaped node, falling
// back to the node itself when there is none to find.
//
// This exists because of where a finding lands rather than what it says. A node's Pos() includes its
// leading trivia, so reporting an import declaration anchors the finding at the first comment above
// it. In the Next wrapper files that put findings on line 1 while the author's
// `eslint-disable-next-line` sat on line 3 covering line 4, and a `-next-line` directive can only
// match the line after itself. The finding was unreachable by any suppression that could be written,
// and it read as a real finding in every count.
//
// A rule reporting through SourceVisitors receives the declaration or call, which is the right
// node to reason about and the wrong node to point at. This turns one into the other.
func SpecifierNode(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	switch node.Kind {
	case ast.KindImportDeclaration:
		if declaration := node.AsImportDeclaration(); declaration != nil && declaration.ModuleSpecifier != nil {
			return declaration.ModuleSpecifier
		}

	case ast.KindCallExpression:
		if call := node.AsCallExpression(); call != nil && call.Arguments != nil &&
			len(call.Arguments.Nodes) > 0 {
			return call.Arguments.Nodes[0]
		}
	}
	return node
}

// HasPathSegment reports whether a path contains a directory or file named exactly segment.
//
// Matching the bare substring instead would flag `@structure/internalization/Thing` for the letters
// it happens to contain, and a boundary rule that fires on a name it never meant to claim teaches
// people to disable it.
//
// Both separators are handled. Two of the three lifted implementations split on the forward slash
// alone, so a path arriving in the Windows shape answered no to every question. Nothing in this tree
// produces one today, which is why the gap was invisible; `NormalizedFileName` exists a few lines up
// for exactly this reason and those callers were not using it.
func HasPathSegment(path string, segment string) bool {
	if segment == "" {
		return false
	}
	for _, part := range strings.FieldsFunc(path, func(character rune) bool {
		return character == '/' || character == '\\'
	}) {
		if part == segment {
			return true
		}
	}
	return false
}
