package structure

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// The rule's messages, one handle per id, whose wording lives in
// `policy/messages/network-no-forbidden-import.json`.
var (
	networkNoForbiddenImportNoDirectTanStackQueryText = policy.MessageOf("structure/network-no-forbidden-import", "noDirectTanStackQuery")
	networkNoForbiddenImportNoDirectApolloText        = policy.MessageOf("structure/network-no-forbidden-import", "noDirectApollo")
	networkNoForbiddenImportNoDirectGraphqlImportText = policy.MessageOf("structure/network-no-forbidden-import", "noDirectGraphqlImport")
)

// messageNoDirectTanStackQuery is the finding, rendered when it is reported so the text comes from the current catalog.
func messageNoDirectTanStackQuery() rule.Message {
	return rule.Message{
		Id:          networkNoForbiddenImportNoDirectTanStackQueryText.Id,
		Description: networkNoForbiddenImportNoDirectTanStackQueryText.Render(nil),
	}
}

// messageNoDirectApollo is the finding, rendered when it is reported so the text comes from the current catalog.
func messageNoDirectApollo() rule.Message {
	return rule.Message{
		Id:          networkNoForbiddenImportNoDirectApolloText.Id,
		Description: networkNoForbiddenImportNoDirectApolloText.Render(nil),
	}
}

// messageNoDirectGraphqlImport is the finding, rendered when it is reported so the text comes from the current catalog.
func messageNoDirectGraphqlImport() rule.Message {
	return rule.Message{
		Id:          networkNoForbiddenImportNoDirectGraphqlImportText.Id,
		Description: networkNoForbiddenImportNoDirectGraphqlImportText.Render(nil),
	}
}

// NetworkNoForbiddenImport flags a direct import of a library NetworkService wraps or replaces.
//
//	valid:   import { useGraphQlQuery } from '@structure/source/services/network/NetworkService'
//	valid:   import { QueryClient } from '@tanstack/react-query'   (inside source/services/network/ only)
//	invalid: import { useQuery } from '@tanstack/react-query'
//	invalid: import { gql } from '@apollo/client'
//	invalid: import { graphql } from '../generated/graphql'
//
// Three message ids because the three imports are forbidden for three different reasons, and a
// reader who sees the wrong one goes looking for the wrong fix. TanStack is wrapped, Apollo is
// replaced, and the generated `graphql` is a type-information problem rather than a client problem.
//
// The three tests are deliberately different shapes, matching the original. TanStack is an exact
// match with NetworkService's module exempt, Apollo is a prefix match so every package under the scope is caught,
// and the generated import only reports when the `graphql` specifier is actually named, since a
// generated path is a normal thing to import types from.
//
// No fix. Each of the three wants a different replacement, and choosing it means knowing what the
// call site was doing with the import.
var NetworkNoForbiddenImport = rule.Rule{
	Name: "structure/network-no-forbidden-import",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// TanStack is allowed where it is configured: NetworkService's module, by where it lives. It
		// once also exempted any Providers.tsx, after Structure's providers stopped importing
		// TanStack, which left a hole the ban was meant to close.
		fileContext := FileContextFor(ctx.SourceFile.FileName().AsString())
		isTanStackExempt := fileContext.IsNetworkServiceFile

		return rule.Listeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				declaration := node.AsImportDeclaration()
				if declaration.ModuleSpecifier == nil || !ast.IsStringLiteralLike(declaration.ModuleSpecifier) {
					return
				}
				source := declaration.ModuleSpecifier.Text()

				if source == "@tanstack/react-query" && !isTanStackExempt {
					ctx.ReportNode(node, messageNoDirectTanStackQuery())
				}

				// A prefix match rather than an exact one, so every package in the scope is
				// caught rather than only the client entry point.
				if strings.HasPrefix(source, "@apollo/") {
					ctx.ReportNode(node, messageNoDirectApollo())
				}

				// Only when the graphql specifier is actually named. A generated path is an
				// ordinary place to import types from, so the path alone is not the defect.
				if strings.Contains(source, "/generated") && importsGraphqlSpecifier(declaration) {
					ctx.ReportNode(node, messageNoDirectGraphqlImport())
				}
			},
		}
	},
}

// importsGraphqlSpecifier reports whether a named import brings in `graphql`.
//
// The imported name is what counts rather than the local alias, so `import { graphql as query }`
// still reports: the alias changes what the call site says and not what was imported.
func importsGraphqlSpecifier(declaration *ast.ImportDeclaration) bool {
	if declaration.ImportClause == nil {
		return false
	}
	// `import { graphql as query }` carries the original name and the alias in two places, which
	// `ImportedNameOf` reads; asking for the local name would miss the aliased form.
	for _, element := range imports.BindingsOf(declaration.AsNode()).Named {
		if imports.ImportedNameOf(element) == "graphql" {
			return true
		}
	}
	return false
}
