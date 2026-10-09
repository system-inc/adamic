package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUselessEmptyExport = rule.Message{
	Id: "uselessEmptyExport",
	Description: "An `export {}` exists to make a script into a module, and this file is already " +
		"a module because something else in it imports or exports. The statement therefore " +
		"changes nothing, and it reads as though it were load-bearing to anyone deciding whether " +
		"they may delete it. Remove it.",
}

// NoUselessEmptyExport flags an `export {}` in a file that is already a module.
//
//	valid:   export {}
//	valid:   export = 3; export {}
//	valid:   declare module '_'
//	invalid: export const value = 1; export {}
//	invalid: import '_'; export {}
//
// Ported from oxc's `no_useless_empty_export`, which is what the gate runs.
//
// # The whole rule is which statements make a file a module
//
// A lone `export {}` is the standard way to force module scope on a file that would otherwise be a
// script sharing the global namespace, so deleting one is a semantic change rather than a tidy-up.
// The rule fires only once something *else* in the file has already made it a module, at which
// point the empty export is genuinely inert.
//
// The set of statements that disarm it is the entire rule, it is not derivable from the phrase "any
// top-level import or export", and every member of it below was measured against the release binary
// rather than reasoned from the source. Four results are the opposite of the intuitive reading:
//
//	declare const a = 2; export {}          REPORTS NOT — an ambient declaration exports nothing
//	declare module 'x' {} export {}         REPORTS NOT — likewise, and it is the corpus's first pass case
//	interface I {} export {}                REPORTS NOT — a bare type declaration is local
//	export = 3; export {}                   REPORTS NOT — see the export-assignment section below
//
// against the ones that do disarm it:
//
//	export const _ = {}; export {}          reports — an export modifier on any declaration
//	export type A = 1; export {}            reports — a type-only export still exports a name
//	export declare const a = 2; export {}   reports — `declare` does not cancel `export`
//	export namespace N {} export {}         reports
//	const _ = {}; export { _ }; export {}   reports — a named export clause carrying specifiers
//	export * from '_'; export {}            reports — a star re-export
//	export default 3; export {}             reports
//	import '_'; export {}                   reports — even a bare side-effect import
//	import type { A } from '_'; export {}   reports
//	import _ = require('_'); export {}       reports — an external import-equals, top level only
//	import x = ns.value; export {}          REPORTS NOT — an internal import-equals is not external
//
// # `export default` and `export =` are one node kind here, and they fall opposite ways
//
// ESTree keeps these apart as `ExportDefaultDeclaration` and `TSExportAssignment`. typescript-go
// parses both to `KindExportAssignment` and separates them with the `IsExportEquals` flag alone.
// That distinction is load-bearing rather than cosmetic, because oxc counts a default export (its
// module record's `export_default`) and counts `export =` nowhere at all: an export assignment is
// absent from all six collections the original tests. So `export default 3; export {}` reports and
// `export = 3; export {}` does not, and a port reading only the node kind gets one of them wrong.
//
// The corpus looks like it says something stronger and does not. Its pass case
// `export * from '_'; export = {}` reads as evidence that an export assignment suppresses the rule
// even alongside a star export. It is not evidence about this rule at all: that input is a
// TypeScript semantic error, "an export assignment cannot be used in a module with other exported
// elements", and oxlint stops before the rule ever runs. Measured directly, the silence is decided
// above the rule the same way a suppression comment would decide it. The rule-level truth is the
// simpler one: an export assignment contributes nothing, so `import '_'; export = 3; export {}`
// does report, on the strength of the import alone.
//
// # An import-equals counts only at the top level, and only when external
//
// The original reaches past its module record for exactly one case, and it scans
// `nodes.program().body` rather than the whole tree. That top-level restriction is real and
// measured: `namespace Foo { import Bar = require('_'); } export {}` is silent even though the
// nested reference is external. The corpus writes the internal-reference form of the same shape as
// a pass case, which would pass either way, so only the probe separates them.
//
// # A `.d.ts` file is exempt entirely
//
// In a declaration file an `export {}` is how the module is kept properly encapsulated, so it is
// load-bearing even when other exports are present. The original gates on the file extension before
// doing any work, and its second tester block is four inputs that all report as `.ts` and all stay
// silent as `.d.ts`. Those four are the only thing pinning the gate, so they are kept here twice,
// once per extension.
var NoUselessEmptyExport = rule.Rule{
	Name: "@typescript-eslint/no-useless-empty-export",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				sourceFile := node.AsSourceFile()
				if isDeclarationFileName(sourceFile.FileName().AsString()) {
					return
				}

				emptyExports := []*ast.Node{}
				makesFileAModule := false
				for _, statement := range sourceFile.Statements.Nodes {
					if isEmptyExportStatement(statement) {
						emptyExports = append(emptyExports, statement)
						continue
					}
					if statementMakesFileAModule(statement) {
						makesFileAModule = true
					}
				}
				if !makesFileAModule {
					return
				}
				for _, emptyExport := range emptyExports {
					ctx.ReportNodeWithFixes(emptyExport, messageUselessEmptyExport,
						ctx.RemoveNode(emptyExport))
				}
			},
		}
	},
}

// isEmptyExportStatement reports whether a statement is an `export {}` carrying nothing.
//
// The module-specifier test is the condition a reading of the original does not suggest. A
// re-export spelled `export {} from './x'` is also an export declaration with an empty clause, and
// it is not useless: it still requests the module, which runs its side effects. Measured, upstream
// stays silent on it in both directions, as the anchor and as the disarmer, because a module
// specifier puts it in the requested-modules collection and out of the empty-export shape at once.
//
// `IsTypeOnly` is deliberately not consulted. `export type {}` reports the same as `export {}`,
// measured, because a type-only empty clause exports exactly as many names as an empty one.
func isEmptyExportStatement(statement *ast.Node) bool {
	if statement.Kind != ast.KindExportDeclaration {
		return false
	}
	declaration := statement.AsExportDeclaration()
	if declaration.ModuleSpecifier != nil {
		return false
	}
	// The nil test and the cast below are both kept, and both are unreachable defensively rather
	// than load-bearing. That is stated here because a later reader will otherwise reinstate the
	// kind test this file used to carry, or delete the nil test as dead, and each was measured.
	//
	// A clause is either NamedExports or NamespaceExport, and the module-specifier return above
	// already declined every input that could arrive here as either a nil clause or a
	// NamespaceExport. Probed on the well-formed `export * from '_'` and `export * as ns from '_'`
	// and on the malformed `export *`, `export *;`, `export * from;`, and `export * as ns;`: every
	// one carries a module specifier, because our parser synthesizes one during error recovery
	// rather than leaving it nil. No input produces a nil clause with no module specifier, so
	// nothing can reach the line below with a clause that is not an empty-capable NamedExports.
	//
	// Mutants removing each of these survived the whole fixture set, and both are equivalent rather
	// than fixture blind spots. The nil test is the one worth keeping anyway: reaching
	// AsNamedExports on a nil clause panics, verified directly, so this is the guard standing
	// between a future parser change and a linter that crashes on a star export.
	clause := declaration.ExportClause
	if clause == nil {
		return false
	}
	return len(clause.AsNamedExports().Elements.Nodes) == 0
}

// statementMakesFileAModule reports whether a top-level statement already puts the file in module
// scope, which is what makes a sibling `export {}` inert.
//
// The arms below are the collections of oxc's module record that any statement can reach, flattened
// into one pass over the statement list. Reading them as "an import or an export"
// over-reports: `declare const`, `declare module`, and a bare `interface` are none of the three, and
// each is a pass case that a looser predicate breaks.
func statementMakesFileAModule(statement *ast.Node) bool {
	switch statement.Kind {
	case ast.KindImportDeclaration:
		// Any import at all, including a bare `import '_'` that binds no name. It requests a
		// module, which is enough on its own.
		return true

	case ast.KindExportDeclaration:
		// A non-empty `export { _ }`, or a re-export with a module specifier. The empty form was
		// already claimed as the anchor above and never reaches here.
		return true

	case ast.KindExportAssignment:
		// `export default` counts and `export =` does not. See the doc comment above: these share
		// a node kind here and fall opposite ways, so the flag is the whole decision.
		return !statement.AsExportAssignment().IsExportEquals

	case ast.KindImportEqualsDeclaration:
		// Only an external reference, `import _ = require('_')`. An internal alias such as
		// `import x = ns.value` resolves inside the file and exports nothing.
		return statement.AsImportEqualsDeclaration().ModuleReference.Kind ==
			ast.KindExternalModuleReference
	}

	// Everything else counts only if it carries the export keyword, which covers `export const`,
	// `export type`, `export interface`, `export class`, `export enum`, `export namespace`, and
	// `export declare const`. The modifier is read rather than the kind, so a declaration form
	// nobody thought to enumerate is handled by the same line.
	return module.IsExported(statement)
}

// isDeclarationFileName reports whether a path names a TypeScript declaration file.
//
// Written here rather than reached for because the question is a suffix test on a string and the
// shelf's nearest neighbours all answer questions about nodes. If a `.d.ts` predicate lands on a
// shelf later this should move to it.
func isDeclarationFileName(fileName string) bool {
	const declarationSuffix = ".d.ts"
	if len(fileName) < len(declarationSuffix) {
		return false
	}
	return fileName[len(fileName)-len(declarationSuffix):] == declarationSuffix
}
