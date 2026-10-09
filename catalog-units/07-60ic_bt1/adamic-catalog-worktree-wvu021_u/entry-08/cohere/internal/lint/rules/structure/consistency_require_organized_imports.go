package structure

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// consistencyRequireOrganizedImportsText is the rule's message, whose wording lives in
// `policy/messages/consistency-require-organized-imports.json`.
var consistencyRequireOrganizedImportsText = policy.MessageOf("structure/consistency-require-organized-imports", "importsNotOrganized")

// messageImportsNotOrganized is the finding, rendered when it is reported so the text comes from the current catalog.
func messageImportsNotOrganized() rule.Message {
	return rule.Message{
		Id:          consistencyRequireOrganizedImportsText.Id,
		Description: consistencyRequireOrganizedImportsText.Render(nil),
	}
}

// ConsistencyRequireOrganizedImports flags a file whose import section is not in canonical form.
//
//	valid:   // Dependencies - Frameworks
//	         import React from 'react';
//
//	         // Dependencies - Utilities
//	         import { titleCase } from '@structure/source/utilities/Strings';
//
//	invalid: import { titleCase } from '@structure/source/utilities/Strings';
//	         import React from 'react';
//
// # What canonical means, exactly
//
// Every import is classified into one of twenty-three groups (`groupOrder`). The groups are emitted
// in that fixed order, each preceded by `// Dependencies - <Group>`, with one blank line between
// consecutive groups and none before the first. An empty group emits nothing at all. Inside a
// group, imports sort by their module specifier string, byte-comparison ascending, with one
// exception: inside Frameworks, `react` sorts ahead of everything regardless of spelling, because
// it is the import every React file has and putting it first makes the block scannable.
//
// A `'use client'` or `'use server'` directive, when it is the first statement, is re-emitted with a
// fixed trailing comment and a blank line under it, above every group header.
//
// # This is a house rule, so there is no oxc corpus
//
// The authority is `libraries/structure/code-quality/lint/rules/ConsistencyRequireOrganizedImportsRule.ts`
// in the ahra tree, and every classification and mechanic below was pinned by running that rule
// through ESLint's Linter API on constructed inputs rather than by reading it. Where this comment
// states a behavior, a probe established it.
//
// A second copy of the rule exists at `projects/www-ahra-ai/.../ConsistencyRequireOrganizedImportsRule.ts`
// and it is behind: it has no file-description handling at all. The copy under `libraries/structure`
// is the live one and is what this ports.
//
// # Why this reports and does not repair
//
// The original ships a fixer that rewrites the whole import section, and three of its rewrites are
// wrong in ways that are invisible until they have already been applied. All three were measured on
// the real rule:
//
//	a disable comment above the first import is DELETED. `importSectionStart` is set to
//	  `firstImport.range[0]`, which is past the comment, so the preservation filter skips it as
//	  preamble while the replaced range still covers it, because an ESTree range starts at trivia.
//	  Input `// eslint-disable-next-line no-restricted-imports` above `import a from 'alpha'`
//	  came back as a bare space. That silently un-suppresses whatever the author was silencing.
//	a trailing comment migrates to a DIFFERENT import. `import z from 'zebra'; // note about zebra`
//	  followed by `import a from 'alpha'` came back with `// note about zebra` sitting above the
//	  alpha import, because a trailing comment reads as a leading comment of the next node.
//	a statement written between two imports is MOVED BELOW BOTH. Input
//	  `import a from 'alpha'; console.log('...'); import b from 'beta';` came back with the call
//	  after both imports, which changes evaluation order.
//
// The brief's standing rule is that a port is not obliged to carry a defect it can see, and that
// reporting without a fix is the subset that can be shown correct. A rewrite of an entire import
// block is applied unattended across three thousand files, so a wrong one is the most expensive
// shape of defect available here. The finding says the section is wrong; a human reorders it.
//
// # The finding points at the first import declaration
//
// Measured on the original: the report is anchored on `imports[0]` in every reporting case,
// including the ones where the import that is out of place is somewhere else entirely, and
// including the `'use client'` cases where the thing that is wrong is the directive's comment. The
// span is that declaration's own text.
var ConsistencyRequireOrganizedImports = rule.Rule{
	Name: "structure/consistency-require-organized-imports",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				checkImportOrganization(ctx, node)
			},
		}
	},
}

// checkImportOrganization compares a file's import section against its canonical rendering.
//
// A KindSourceFile listener rather than a per-import one, because the judgment is about the section
// as a whole: no single import is wrong on its own, and reporting per-import would produce one
// finding per file times however many imports it holds.
func checkImportOrganization(ctx rule.Context, sourceFile *ast.Node) {
	statements := sourceFile.AsSourceFile().Statements
	if statements == nil {
		return
	}

	importDeclarations := []*ast.Node{}
	for _, statement := range statements.Nodes {
		if statement.Kind == ast.KindImportDeclaration {
			importDeclarations = append(importDeclarations, statement)
		}
	}
	// A file with no imports has no section to organize. The original returns here too, and it is
	// load-bearing rather than an optimization: without it every file in the tree that imports
	// nothing would render an empty section and compare unequal to itself.
	if len(importDeclarations) == 0 {
		return
	}

	sourceFileNode := ctx.SourceFile
	sourceText := sourceFileNode.Text()
	fileComments := comments.ForFile(ctx)

	directive, directiveNode := leadingDirective(statements)

	sectionStart := importSectionStart(sourceFileNode, sourceText, fileComments, importDeclarations[0], directiveNode)

	descriptionBlock := fileDescriptionBlock(sourceFileNode, fileComments, importDeclarations[0], directiveNode)

	classified := classifyDeclarations(sourceFileNode, sourceText, fileComments, importDeclarations, sectionStart, descriptionBlock)

	interleaved, sectionEnd := interleavedStatements(sourceFileNode, sourceText, fileComments, statements, importDeclarations, sectionStart)

	expected := renderSection(directive, descriptionBlock, classified, interleaved)

	if sectionEnd > len(sourceText) {
		return
	}
	if sourceText[sectionStart:sectionEnd] == expected {
		return
	}

	/*
	 * The fix is the rendering the comparison just used, written back over the span it was compared
	 * against. Nothing is recomputed: a fix derived from a second rendering could disagree with the
	 * verdict that produced it, and a rule that reports one thing and writes another is worse than
	 * one that only reports.
	 *
	 * `canRenderSafely` is what makes this shippable. Two of the original fixer's three defects are
	 * repaired at the rendering (a trailing comment now travels with the import it trails, and a
	 * comment above the section is outside the replaced span to begin with), and the third cannot
	 * be: a statement written between two imports has no correct home in a section whose imports are
	 * being reordered. Those files are reported without a fix.
	 */
	if !canRenderSafely(interleaved) || suppressionPrecedesSection(sourceText, fileComments, sectionStart) {
		ctx.ReportNode(importDeclarations[0], messageImportsNotOrganized())
		return
	}

	ctx.ReportNodeWithFixes(importDeclarations[0], messageImportsNotOrganized(),
		rule.ReplaceRange(core.NewTextRange(sectionStart, sectionEnd), expected))
}
