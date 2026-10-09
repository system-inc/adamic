package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// boundaryNoProjectImportText is the rule's message, whose wording lives in
// `policy/messages/boundary-no-project-import.json`.
var boundaryNoProjectImportText = policy.MessageOf("nexus/boundary-no-project-import", "forbiddenProjectImport")

func messageForbiddenProjectImport() rule.Message {
	return rule.Message{
		Id:          "forbiddenProjectImport",
		Description: boundaryNoProjectImportText.Render(nil),
	}
}

// BoundaryNoProjectImportOptions configures which library is guarded and what it may still read.
//
// Both fields are options rather than constants because base and structure each guard their own
// directory with their own whitelist, and this rule is meant to know about neither.
type BoundaryNoProjectImportOptions struct {
	// LibraryDirectory is the path fragment identifying the guarded library, matched against the
	// normalized filename. "/libraries/base/", say. Files outside it are never visited.
	LibraryDirectory string `json:"libraryDirectory"`

	// Allowed are the exact '@project/...' specifiers this library may import.
	Allowed []string `json:"allowed"`
}

// BoundaryNoProjectImport blocks '@project/*' imports inside a framework library.
//
//	valid:   import { thing } from '@structure/source/Thing'   (inside /libraries/base/)
//	valid:   import { settings } from '@project/ProjectSettings' when named in Allowed
//	invalid: import { helper } from '@project/source/Helper'
//	invalid: const x = require('@project/source/Helper')
//
// A few project files exist precisely for the framework to read. ProjectSettings is the usual one:
// the project answers questions the framework's own interface asks. Those are named in Allowed
// rather than reached through a runtime path read, which would dodge this rule while giving up
// types, refactoring, and go-to-definition.
//
// The whitelist match is exact, not a prefix. A prefix would let '@project/ProjectSettings' admit
// '@project/ProjectSettingsSecret', so a whitelist entry would quietly widen over time as files
// were added beside the one that was actually approved.
//
// This rule declines every file outside the guarded directory before any node is visited, and a
// configuration with an empty LibraryDirectory declines everything. That last case matters: it is
// how a rule registered without its required option fails closed rather than guarding the whole
// tree by accident.
var BoundaryNoProjectImport = rule.Rule{
	Name: "nexus/boundary-no-project-import",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		settings, hasSettings := rule.OptionsAs[BoundaryNoProjectImportOptions](options)
		if !hasSettings || settings.LibraryDirectory == "" {
			// Failing closed is right, and silent is not: a rule enabled without the directory it guards
			// checks nothing, so the skip names the missing option (#pa7k7zv).
			ctx.Skip("no libraryDirectory is configured")
			return nil
		}
		if !strings.Contains(imports.NormalizedFileName(ctx.SourceFile), settings.LibraryDirectory) {
			return nil
		}

		allowed := make(map[string]bool, len(settings.Allowed))
		for _, specifier := range settings.Allowed {
			allowed[specifier] = true
		}

		return imports.SourceVisitors(func(source string, node *ast.Node) {
			if isProjectAlias(source) && !allowed[source] {
				ctx.ReportNode(imports.SpecifierNode(node), messageForbiddenProjectImport())
			}
		})
	},
}

// isProjectAlias reports whether a specifier reaches through the '@project' alias.
//
// The prefix has to end at a boundary. A bare prefix test would flag an unrelated package named
// "@projector", and a boundary rule that fires on a name it never meant to claim teaches people to
// disable it.
func isProjectAlias(source string) bool {
	return source == "@project" || strings.HasPrefix(source, "@project/")
}
