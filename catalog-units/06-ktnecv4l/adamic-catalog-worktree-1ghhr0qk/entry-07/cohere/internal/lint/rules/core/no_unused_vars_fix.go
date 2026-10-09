package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/ecmascript/tokens"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// The removal of an unused import, typescript-eslint 8.71.0's getImportFixer (src/rules/no-unused-vars.ts),
// ported against its rows (#e1zk9s0). Under `enableAutofixRemoval: {imports: true}` it is a fix the edit
// engine applies; otherwise, the default, upstream offers the same edit as a suggestion, and so does this.
// Upstream computes no removal for any other kind of binding, and neither does this.
//
// # Report order is part of the answer
//
// ESLint calls a fix function when the finding is reported, so upstream's question "is every binding of
// this declaration unused" (areAllSpecifiersUnused) sees only the bindings reported so far. In
// `import { A, B, C } from 'm'` with all three unused, A's and B's removals take the specifier and C's,
// the last, takes the declaration. The edit engine then applies the earliest-starting of overlapping
// fixes, the declaration's, which is ESLint's own rule for one pass. So the bindings reported so far are
// passed in, in report order, rather than the set the whole file will report.
//
// # Two divergences, both refusals to remove a binding something reads
//
//   - Upstream's report attaches the removal to every finding on an import, usedIgnoredVar included, and
//     counts that finding as "reported" when it asks the question above. Under reportUsedIgnorePattern
//     with autofix on, that deletes an import the file reads: `import { _a } from 'm'; console.log(_a);`
//     becomes `console.log(_a);`, which no longer compiles. Measured on the installed rule. Here the
//     removal is offered only on an unused finding, and only unused findings count as reported.
//   - Upstream removes a namespace import's whole declaration unconditionally, which deletes a used
//     default beside it. Here that declaration goes only when every binding in it is reported unused.

// noUnusedVarsRemovalMessage renders the two suggestion messages upstream's removal carries:
// removeUnusedImportDeclaration when the edit takes the whole declaration, removeUnusedVar otherwise.
func noUnusedVarsRemovalMessage(id string, name string) rule.Message {
	if id == "removeUnusedImportDeclaration" {
		return rule.Message{Id: id, Description: "Remove the import declaration: nothing it binds is used."}
	}
	return rule.Message{Id: id, Description: "Remove '" + name + "' from the import, keeping the bindings beside it that are used."}
}

// importRemoval is the edit that removes an unused import binding, and which of upstream's two messages
// names it. It declines (false) where upstream's fixer does: a binding with more than one declaration, or
// a declaration whose tokens are not the shape the edit needs.
func importRemoval(ctx rule.Context, candidate candidateBinding, symbol *ast.Symbol, reportedUnused map[*ast.Node]bool) (rule.Fix, string, bool) {
	// Upstream fixes only a variable with exactly one definition, since cleaning up several is a
	// different edit it declines to guess at.
	if symbol != nil && len(rule.DeclarationsIn(ctx.SourceFile, symbol)) > 1 {
		return rule.Fix{}, "", false
	}

	declaration := candidate.declaration
	switch declaration.Kind {
	case ast.KindImportEqualsDeclaration:
		// One binding per declaration, so the declaration goes.
		return removeNodeWithTrailingNewline(ctx.SourceFile, declaration), "removeUnusedImportDeclaration", true

	case ast.KindNamespaceImport:
		// Upstream removes the declaration here unconditionally, on the comment that a namespace import
		// shares its declaration with nothing. It can share it with a default, `import Used, * as
		// Unused from 'm'`, and there the unconditional removal deletes the used default. The same
		// refusal as above: the declaration goes only when every binding in it is reported unused.
		importDeclaration := enclosingImportDeclaration(declaration)
		if importDeclaration == nil || !allReported(importDeclarationBindings(importDeclaration), reportedUnused) {
			return rule.Fix{}, "", false
		}
		return removeNodeWithTrailingNewline(ctx.SourceFile, importDeclaration), "removeUnusedImportDeclaration", true

	case ast.KindImportClause:
		// The default binding.
		importDeclaration := enclosingImportDeclaration(declaration)
		if importDeclaration == nil {
			return rule.Fix{}, "", false
		}
		bindings := importDeclarationBindings(importDeclaration)
		if len(bindings) == 1 || allReported(bindings, reportedUnused) {
			return removeNodeWithTrailingNewline(ctx.SourceFile, importDeclaration), "removeUnusedImportDeclaration", true
		}
		// `import Unused, { Used } from 'm'`: the default and the comma after it.
		list := tokens.Of(ctx.SourceFile, importDeclaration)
		name := rule.TokenRange(ctx.SourceFile, candidate.name)
		comma, found := list.After(name.End())
		if !found || comma.Kind != ast.KindCommaToken {
			return rule.Fix{}, "", false
		}
		return rule.RemoveRange(core.NewTextRange(min(name.Pos(), comma.Start), max(name.End(), comma.End))), "removeUnusedVar", true

	case ast.KindImportSpecifier:
		importDeclaration := enclosingImportDeclaration(declaration)
		if importDeclaration == nil {
			return rule.Fix{}, "", false
		}
		bindings := importDeclarationBindings(importDeclaration)
		if len(bindings) == 1 || allReported(bindings, reportedUnused) {
			return removeNodeWithTrailingNewline(ctx.SourceFile, importDeclaration), "removeUnusedImportDeclaration", true
		}
		list := tokens.Of(ctx.SourceFile, importDeclaration)
		statement := rule.TokenRange(ctx.SourceFile, importDeclaration)

		usedNamed := 0
		for _, binding := range bindings {
			if binding.Parent != nil && binding.Parent.Kind == ast.KindImportSpecifier && !reportedUnused[binding] {
				usedNamed++
			}
		}
		if usedNamed == 0 {
			// `import Used, { Unused, ... } from 'm'`: the braces and the comma before them.
			leftCurly, found := list.FirstBetween(statement.Pos(), statement.End(), ast.KindOpenBraceToken)
			if !found {
				return rule.Fix{}, "", false
			}
			comma, found := list.Before(leftCurly.Start)
			if !found || comma.Kind != ast.KindCommaToken {
				return rule.Fix{}, "", false
			}
			rightCurly, found := list.FirstBetween(statement.Pos(), statement.End(), ast.KindCloseBraceToken)
			if !found {
				return rule.Fix{}, "", false
			}
			return rule.RemoveRange(core.NewTextRange(comma.Start, rightCurly.End)), "removeUnusedVar", true
		}

		// One specifier and a comma: the one before it where there is one, which leaves the nicer text,
		// else the one after, for a specifier first in its list.
		specifier := rule.TokenRange(ctx.SourceFile, declaration)
		comma, found := list.Before(specifier.Pos())
		if !found || comma.Kind != ast.KindCommaToken {
			comma, found = list.After(specifier.End())
			if !found || comma.Kind != ast.KindCommaToken {
				return rule.Fix{}, "", false
			}
		}
		return rule.RemoveRange(core.NewTextRange(min(specifier.Pos(), comma.Start), max(specifier.End(), comma.End))), "removeUnusedVar", true
	}
	return rule.Fix{}, "", false
}

// enclosingImportDeclaration is the import declaration an import clause, namespace import or specifier
// belongs to.
func enclosingImportDeclaration(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindImportDeclaration {
			return current
		}
	}
	return nil
}

// importDeclarationBindings is every name an import declaration binds, upstream's
// getDeclaredVariables(importDecl): the default, the namespace, and each named specifier's local name.
func importDeclarationBindings(importDeclaration *ast.Node) []*ast.Node {
	shapes := imports.BindingsOf(importDeclaration)
	var bindings []*ast.Node
	if shapes.Default != nil {
		bindings = append(bindings, shapes.Default)
	}
	if shapes.Namespace != nil {
		if name := shapes.Namespace.Name(); name != nil {
			bindings = append(bindings, name)
		}
	}
	for _, specifier := range shapes.Named {
		if name := specifier.Name(); name != nil {
			bindings = append(bindings, name)
		}
	}
	return bindings
}

// allReported is upstream's areAllSpecifiersUnused, asked of the bindings reported so far.
func allReported(bindings []*ast.Node, reportedUnused map[*ast.Node]bool) bool {
	for _, binding := range bindings {
		if !reportedUnused[binding] {
			return false
		}
	}
	return true
}

// removeNodeWithTrailingNewline is upstream's helper of the same name: when the node is the only thing on
// its lines, the lines go, newline included; otherwise just the node's own text.
func removeNodeWithTrailingNewline(sourceFile *ast.SourceFile, node *ast.Node) rule.Fix {
	nodeRange := rule.TokenRange(sourceFile, node)
	sourceText := sourceFile.Text()
	lineStarts := scanner.GetECMALineStarts(sourceFile)
	startLine := scanner.ComputeLineOfPosition(lineStarts, nodeRange.Pos())
	endLine := scanner.ComputeLineOfPosition(lineStarts, nodeRange.End())
	lineRangeStart := int(lineStarts[startLine])
	lineRangeEnd := len(sourceText)
	if endLine+1 < len(lineStarts) {
		lineRangeEnd = int(lineStarts[endLine+1])
	}
	if sourceText[nodeRange.Pos():nodeRange.End()] == text.TrimWhitespace(sourceText[lineRangeStart:lineRangeEnd]) {
		return rule.RemoveRange(core.NewTextRange(lineRangeStart, lineRangeEnd))
	}
	return rule.RemoveRange(nodeRange)
}
