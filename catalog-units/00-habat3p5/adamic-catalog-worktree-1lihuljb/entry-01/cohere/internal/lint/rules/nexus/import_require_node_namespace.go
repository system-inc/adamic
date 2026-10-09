// Package nexus holds the rules that describe how Nexus code reads, ported from the TypeScript
// originals in ahra's libraries/structure/libraries/nexus/code-quality/lint/rules.
//
// The judgment in these rules was made once, in prose, in those files. Porting carries the prose
// across deliberately: a rule that says only what is wrong gets disabled the first time it is
// inconvenient, while a rule that says why gets fixed.
package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// Node built-in modules, bare (no "node:" prefix).
var nodeBuiltins = map[string]bool{
	"assert": true, "async_hooks": true, "buffer": true, "child_process": true,
	"cluster": true, "console": true, "constants": true, "crypto": true,
	"dgram": true, "diagnostics_channel": true, "dns": true, "domain": true,
	"events": true, "fs": true, "http": true, "http2": true, "https": true,
	"inspector": true, "module": true, "net": true, "os": true, "path": true,
	"perf_hooks": true, "process": true, "punycode": true, "querystring": true,
	"readline": true, "repl": true, "stream": true, "string_decoder": true,
	"sys": true, "timers": true, "tls": true, "trace_events": true, "tty": true,
	"url": true, "util": true, "v8": true, "vm": true, "wasi": true,
	"worker_threads": true, "zlib": true,
}

// Subpath exports whose alias cannot be derived from the path alone.
var subpathAliases = map[string]string{
	"fs/promises":     "NodeFileSystemPromises",
	"stream/promises": "NodeStreamPromises",
	"timers/promises": "NodeTimersPromises",
}

// Module names whose expansion is not a mechanical PascalCase of the path. An alias exists to be
// read, and "NodeFs" tells a reader less than "NodeFileSystem" does.
var moduleNameExpansions = map[string]string{
	"fs": "FileSystem", "os": "OperatingSystem", "vm": "VirtualMachine",
	"tls": "Tls", "tty": "Tty", "net": "Net", "dns": "Dns", "dgram": "Dgram",
	"url": "Url", "sys": "Sys", "v8": "V8",
}

// bareModuleName strips the "node:" prefix when present.
func bareModuleName(source string) string {
	return strings.TrimPrefix(source, "node:")
}

// isNodeBuiltin reports whether a specifier names a Node built-in, in either spelling.
func isNodeBuiltin(source string) bool {
	bare := bareModuleName(source)
	if nodeBuiltins[bare] {
		return true
	}
	_, isSubpath := subpathAliases[bare]
	return isSubpath
}

// expectedAlias derives the namespace alias a built-in must be imported under.
func expectedAlias(source string) string {
	bare := bareModuleName(source)

	if alias, isSubpath := subpathAliases[bare]; isSubpath {
		return alias
	}
	if expansion, hasExpansion := moduleNameExpansions[bare]; hasExpansion {
		return "Node" + expansion
	}

	// "child_process" becomes "NodeChildProcess".
	var builder strings.Builder
	builder.WriteString("Node")
	for _, part := range strings.Split(bare, "_") {
		if part == "" {
			continue
		}
		builder.WriteString(strings.ToUpper(part[:1]))
		builder.WriteString(part[1:])
	}
	return builder.String()
}

// The three messages interpolate what the TypeScript original interpolates: the specifier as
// written, the alias it should be imported under, and for the alias arm the alias it actually has.
// An earlier port carried a fixed example, "import * as NodeFileSystem from 'node:fs'", in the
// namespace message, which told `import NodePath from 'node:path'` to import a different module
// under a different name. The reasons are cohere's own and stay; the specifics are the original's.

// importRequireNodeNamespaceRequireNamespaceImportText is the rule's `requireNamespaceImport` message,
// whose wording lives in `policy/messages/import-require-node-namespace.json`.
var importRequireNodeNamespaceRequireNamespaceImportText = policy.MessageOf("nexus/import-require-node-namespace", "requireNamespaceImport")

func messageRequireNamespaceImport(source string, expected string, bare string) rule.Message {
	return rule.Message{
		Id: "requireNamespaceImport",
		Description: importRequireNodeNamespaceRequireNamespaceImportText.Render(map[string]string{
			"source":   source,
			"expected": expected,
			"bare":     bare,
		}),
	}
}

// importRequireNodeNamespaceRequireNodePrefixText is the rule's `requireNodePrefix` message,
// whose wording lives in `policy/messages/import-require-node-namespace.json`.
var importRequireNodeNamespaceRequireNodePrefixText = policy.MessageOf("nexus/import-require-node-namespace", "requireNodePrefix")

func messageRequireNodePrefix(source string, expected string, bare string) rule.Message {
	return rule.Message{
		Id: "requireNodePrefix",
		Description: importRequireNodeNamespaceRequireNodePrefixText.Render(map[string]string{
			"source":   source,
			"expected": expected,
			"bare":     bare,
		}),
	}
}

// importRequireNodeNamespaceRequireCorrectAliasText is the rule's `requireCorrectAlias` message,
// whose wording lives in `policy/messages/import-require-node-namespace.json`.
var importRequireNodeNamespaceRequireCorrectAliasText = policy.MessageOf("nexus/import-require-node-namespace", "requireCorrectAlias")

func messageRequireCorrectAlias(expected string, actual string, bare string) rule.Message {
	return rule.Message{
		Id: "requireCorrectAlias",
		Description: importRequireNodeNamespaceRequireCorrectAliasText.Render(map[string]string{
			"expected": expected,
			"actual":   actual,
			"bare":     bare,
		}),
	}
}

// ImportRequireNodeNamespace enforces namespace imports with a Node-prefixed alias for Node
// built-ins, in both the bare and "node:" spellings.
//
//	valid:   import * as NodeFileSystem from 'node:fs'
//	invalid: import fs from 'node:fs'
//	invalid: import { readFileSync } from 'node:fs'
//	invalid: import * as fs from 'node:fs'
//	invalid: import * as NodeFileSystem from 'fs'
//
// Fixes are deliberately absent for the namespace case, and that absence is the interesting part.
// Rewriting the import line is only half the job: a default or named import binds names the body
// goes on using, and those bindings do not survive the change. Turning `import fs from 'node:fs'`
// into a namespace import leaves every `fs.existsSync` call unresolved. Both spellings typecheck
// before the fix and fail to compile after it, which is the worst shape a fixer can have.
//
// The TypeScript original followed each binding through scope and re-rooted it at the namespace.
// That needs scope analysis this rule does not have yet, so the finding is reported without a fix
// rather than with one that breaks the build. Reporting honestly beats fixing wrongly.
var ImportRequireNodeNamespace = rule.Rule{
	Name: "nexus/import-require-node-namespace",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				declaration := node.AsImportDeclaration()
				if declaration == nil || declaration.ModuleSpecifier == nil {
					return
				}

				source := declaration.ModuleSpecifier.Text()
				if !isNodeBuiltin(source) {
					return
				}

				bare := bareModuleName(source)
				expected := expectedAlias(source)

				namespaceImport, hasNamespaceImport := namespaceImportOf(declaration)
				if !hasNamespaceImport {
					ctx.ReportNode(node, messageRequireNamespaceImport(source, expected, bare))
					return
				}

				// The prefix and the alias are independent defects, so a line can carry both and
				// must report both. Returning after the first would hide the second until the
				// first was fixed and the gate run again.
				//
				// Each finding points where the original's does: the prefix at the whole declaration
				// (its `node`), with the repair still confined to the specifier inside it, and the
				// alias at the `* as alias` specifier (its `namespaceSpecifier`) rather than at the
				// name alone.
				if !strings.HasPrefix(source, "node:") {
					ctx.ReportNodeWithFixes(
						node,
						messageRequireNodePrefix(source, expected, bare),
						ctx.ReplaceNode(declaration.ModuleSpecifier, "'node:"+bare+"'"),
					)
				}

				if alias := namespaceImport.Name(); alias != nil && alias.Text() != expected {
					// No fix: renaming the alias without renaming its references through scope
					// would leave the file uncompilable, the same trap as the namespace case above.
					ctx.ReportNode(namespaceImport, messageRequireCorrectAlias(expected, alias.Text(), bare))
				}
			},
		}
	},
}

// namespaceImportOf returns the `* as alias` node of a namespace import, and whether the declaration
// is one at all. A declaration with no import clause (a bare side-effect import) is neither.
func namespaceImportOf(declaration *ast.ImportDeclaration) (*ast.Node, bool) {
	if declaration.ImportClause == nil {
		return nil, false
	}
	clause := declaration.ImportClause.AsImportClause()
	if clause == nil || clause.NamedBindings == nil {
		return nil, false
	}
	if clause.NamedBindings.Kind != ast.KindNamespaceImport {
		return nil, false
	}
	return clause.NamedBindings, true
}
