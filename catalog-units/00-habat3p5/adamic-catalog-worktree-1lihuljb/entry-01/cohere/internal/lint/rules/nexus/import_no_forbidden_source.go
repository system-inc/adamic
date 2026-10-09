package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// forbiddenSource is one package we do not import, and what to import instead.
//
// A table rather than a branch per package: the original repeated the same report-and-rewrite four
// times for static imports and four more for call expressions, which is where the next entry gets
// added to one list and forgotten in the other.
type forbiddenSource struct {
	Specifier   string
	Replacement string
	// Message is the entry's message, whose wording lives in
	// `policy/messages/import-no-forbidden-source.json`.
	Message policy.MessageHandle

	// Owners are where the sanctioned wrapper lives, matched as path substrings. The wrapper imports
	// the forbidden source by design, so a file there is exempt, and reporting it would also have the
	// fix rewrite the wrapper into an import of itself. A package with no wrapper has no owner.
	Owners []string
}

var forbiddenSources = []forbiddenSource{
	{
		Specifier:   "next/navigation",
		Replacement: "@structure/source/router/Navigation",
		// Navigation spans its router directory: its useRouter hook imports Next's own.
		Owners:  []string{"/source/router/"},
		Message: policy.MessageOf("nexus/import-no-forbidden-source", "forbiddenNavigationImport"),
	},
	{
		Specifier:   "next/link",
		Replacement: "@structure/source/components/navigation/Link",
		// A single file, since the components beside it must still use the wrapper.
		Owners:  []string{"/source/components/navigation/Link.tsx"},
		Message: policy.MessageOf("nexus/import-no-forbidden-source", "forbiddenLinkImport"),
	},
	{
		Specifier:   "next/image",
		Replacement: "@structure/source/components/images/Image",
		Owners:      []string{"/source/components/images/Image.tsx"},
		Message:     policy.MessageOf("nexus/import-no-forbidden-source", "forbiddenImageImport"),
	},
	{
		Specifier:   "framer-motion",
		Replacement: "motion/react",
		Message:     policy.MessageOf("nexus/import-no-forbidden-source", "forbiddenMotionImport"),
	},
}

// ImportNoForbiddenSource replaces framework-coupled packages with the framework-independent
// equivalents shipped from @structure/source.
//
//	valid:   import Link from '@structure/source/components/navigation/Link'
//	invalid: import Link from 'next/link'
//	invalid: const { motion } = require('framer-motion')
//	invalid: const module = await import('next/image')
//
// The fix rewrites only the specifier string, never the import clause, and that restraint is the
// interesting decision. The TypeScript original rebuilds the whole declaration for the next/image
// case so a default import becomes a named one, which means the fix is guessing at the replacement
// module's export shape from inside a rule that cannot see it. Rewriting the string is the part
// that is always correct; if the bindings also need to change, the compiler says so immediately and
// a human makes that call with the module in front of them.
var ImportNoForbiddenSource = rule.Rule{
	Name: "nexus/import-no-forbidden-source",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// The finding is anchored on the specifier, which is both the thing that gets rewritten and
		// the thing a reader has to change.
		//
		// Anchoring on the enclosing declaration instead is what this rule did first, and it is a
		// defect that hides in plain sight: a node's Pos() includes its leading trivia, so an import
		// preceded by comments reports at the first comment rather than at the import. In the Next
		// wrapper files that put the finding on line 1 while the author's
		// `eslint-disable-next-line` sat on line 3 covering line 4. A `-next-line` directive can
		// only match the line after itself, so the finding was unreachable by any suppression that
		// could be written, and it reads as a real finding in every count. Report the node the
		// original reports.
		if ctx.SourceFile == nil {
			return nil
		}
		fileName := strings.ReplaceAll(ctx.SourceFile.FileName().AsString(), `\`, "/")
		report := func(specifierNode *ast.Node, source string) {
			for _, forbidden := range forbiddenSources {
				if source != forbidden.Specifier {
					continue
				}
				for _, owner := range forbidden.Owners {
					if strings.Contains(fileName, owner) {
						return
					}
				}
				ctx.ReportNodeWithFixes(
					specifierNode,
					rule.Message{Id: forbidden.Message.Id, Description: forbidden.Message.Render(nil)},
					ctx.ReplaceNode(specifierNode, "'"+forbidden.Replacement+"'"),
				)
				return
			}
		}

		return rule.Listeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				declaration := node.AsImportDeclaration()
				if declaration == nil || declaration.ModuleSpecifier == nil {
					return
				}
				if !ast.IsStringLiteralLike(declaration.ModuleSpecifier) {
					return
				}
				report(declaration.ModuleSpecifier, declaration.ModuleSpecifier.Text())
			},

			ast.KindCallExpression: func(node *ast.Node) {
				source, isImport := imports.CallExpressionSource(node)
				if !isImport {
					return
				}
				call := node.AsCallExpression()
				if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				report(call.Arguments.Nodes[0], source)
			},
		}
	},
}
