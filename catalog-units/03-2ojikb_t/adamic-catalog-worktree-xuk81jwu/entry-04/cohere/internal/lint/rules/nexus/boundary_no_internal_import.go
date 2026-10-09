package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// boundaryNoInternalImportAliasedInternalText and boundaryNoInternalImportOutsideInternalText are the
// rule's messages, whose wording lives in `policy/messages/boundary-no-internal-import.json`.
var (
	boundaryNoInternalImportAliasedInternalText = policy.MessageOf("nexus/boundary-no-internal-import", "aliasedInternal")
	boundaryNoInternalImportOutsideInternalText = policy.MessageOf("nexus/boundary-no-internal-import", "outsideInternal")
)

func messageAliasedInternal() rule.Message {
	return rule.Message{
		Id:          "aliasedInternal",
		Description: boundaryNoInternalImportAliasedInternalText.Render(nil),
	}
}

func messageOutsideInternal() rule.Message {
	return rule.Message{
		Id:          "outsideInternal",
		Description: boundaryNoInternalImportOutsideInternalText.Render(nil),
	}
}

// BoundaryNoInternalImport confines an internal folder to the folder that owns it.
//
//	valid:   /a/b/Thing.ts        imports './internal/Detail'
//	valid:   /a/b/c/Thing.ts      imports '../internal/Detail'
//	invalid: /a/b/Thing.ts        imports '@structure/x/internal/Detail'
//	invalid: /a/other/Thing.ts    imports '../b/internal/Detail'
//
// The owning folder is the directory directly above the last 'internal' segment in the resolved
// import path. Files at or below that directory may import from it; everything else may not.
//
// Two details are load-bearing. The last 'internal' is what counts, not the first: a path with
// nested internal folders is owned by the innermost one, which is the stricter and correct reading.
// And the comparison resolves the relative specifier against the importing file's own directory
// rather than against the process working directory, which is what the TypeScript original did.
// That original compared two paths both made relative to cwd, so it agreed with this one whenever
// the linter ran from the repository root and silently stopped enforcing anything when it did not.
// Resolving against the importing file has no such dependence.
var BoundaryNoInternalImport = rule.Rule{
	Name: "nexus/boundary-no-internal-import",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}
		importingFile := imports.NormalizedFileName(ctx.SourceFile)

		return imports.SourceVisitors(func(source string, node *ast.Node) {
			checkInternalImport(ctx, node, importingFile, source)
		})
	},
}

// checkInternalImport decides one specifier against one importing file.
func checkInternalImport(ctx rule.Context, node *ast.Node, importingFile string, source string) {
	if !strings.Contains(source, "internal") {
		return
	}

	// An alias or a bare package name gives no distance to reason about, so it cannot be shown to
	// be inside the owning folder and is refused outright. A relative specifier is the only shape
	// that can prove it belongs.
	if !strings.HasPrefix(source, ".") {
		if imports.HasPathSegment(source, "internal") {
			ctx.ReportNode(imports.SpecifierNode(node), messageAliasedInternal())
		}
		return
	}

	resolved := tspath.NormalizePath(tspath.ResolvePath(tspath.GetDirectoryPath(importingFile), source))

	segments := strings.Split(resolved, "/")
	internalIndex := -1
	for index, segment := range segments {
		if segment == "internal" {
			internalIndex = index
		}
	}
	// Index zero would mean a path rooted at "internal" with no owner above it, which no real file
	// layout produces and which has no folder to compare against.
	if internalIndex <= 0 {
		return
	}

	owningFolder := strings.Join(segments[:internalIndex], "/")
	if !strings.HasPrefix(importingFile, owningFolder+"/") {
		ctx.ReportNode(imports.SpecifierNode(node), messageOutsideInternal())
	}
}
