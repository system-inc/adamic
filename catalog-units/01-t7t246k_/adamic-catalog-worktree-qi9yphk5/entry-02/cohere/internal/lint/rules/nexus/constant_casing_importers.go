package nexus

import (
	"crypto/sha256"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module_roots"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// importedNames is what the rest of the program takes from one file.
//
// everything is set by any import that can reach a name without spelling it: a namespace import, an
// `export *`, a dynamic `import()`, a `require`, an import type. Those make "nothing imports this
// name" unprovable, so they make it false.
type importedNames struct {
	everything bool
	names      map[string]bool
}

// importerIndex is, for each file in the program, the names any other file imports from it, plus
// the files that are alive without an importer.
type importerIndex struct {
	byPath map[tspath.PathKey]*importedNames
	roots  *module_roots.Set
}

// importerCache holds one index per program, built the first time a camelCase export asks.
//
// A package variable for the reason `boundary_no_project_theme_value.go` gives at length: the file
// cache is per file and carries no mutex, the program does not exist at registration, and the index
// is a fact about the whole run. Built lazily, so a tree with no camelCase export pays nothing.
var importerCache struct {
	sync.Mutex
	program rule.ProgramIdentity
	index   *importerIndex
}

// importerIndexFor returns the run's index, or nil when there is no program to read, which the
// caller treats as "cannot prove nobody imports it".
//
// It takes the program rather than the rule context so the one read of the context's program sits
// in the rule's own file, beside the ProgramReads declaration it obliges.
func importerIndexFor(program rule.Program) *importerIndex {
	if program == nil {
		return nil
	}
	importerCache.Lock()
	defer importerCache.Unlock()
	if importerCache.program == program.Identity() && importerCache.index != nil {
		return importerCache.index
	}
	index := buildImporterIndex(program)
	importerCache.program = program.Identity()
	importerCache.index = index
	return index
}

// buildImporterIndex reads every module specifier in the program once.
//
// `SourceFile.Imports()` is the binder's own list of every specifier a file names, static or
// dynamic, so an import form nobody enumerated here still lands in the index, as "everything". The
// specific forms below only narrow that to named imports where the syntax says exactly which names.
func buildImporterIndex(program rule.Program) *importerIndex {
	index := &importerIndex{byPath: map[tspath.PathKey]*importedNames{}}
	var projectFiles []*ast.SourceFile
	for _, importer := range program.SourceFiles() {
		if importer == nil {
			continue
		}
		if !strings.Contains(importer.FileName().AsString(), "/node_modules/") && !importer.IsDeclarationFile {
			projectFiles = append(projectFiles, importer)
		}
		for _, specifier := range importer.Imports() {
			if specifier == nil || !ast.IsStringLiteralLike(specifier) {
				continue
			}
			resolved := program.ResolveModule(importer, specifier)
			if !resolved.IsResolved() {
				continue
			}
			target := program.GetSourceFileForResolvedModule(resolved.ResolvedFileName)
			if target == nil || target == importer {
				continue
			}
			entry := index.byPath[target.PathKey()]
			if entry == nil {
				entry = &importedNames{names: map[string]bool{}}
				index.byPath[target.PathKey()] = entry
			}
			recordImportedNames(entry, specifier)
		}
	}
	index.roots = module_roots.Build(projectFiles)
	return index
}

// constantCasingFingerprint is the rule's program fingerprint (#f96cnry): every file's entry in the
// importer index, the names others take from it and whether something takes everything, and the
// directories a computed `import()` roots. A file's verdict reads its own entry and whether it is a
// root, which is a fact about its name and those directories, so an edit anywhere that changes no
// import and no computed import directory leaves this alone, and the file's findings replay. The
// rest of what the verdict reads, the checker's answers about the file, is the file's type
// fingerprint's to cover.
//
// The options are not read. The rule's one option, frameworkConstantNames, judges a declaration inside
// the file and chooses nothing the index reads, and the findings cache's key already holds every config
// file, so a change of options re-runs the rule without the fingerprint having to move.
func constantCasingFingerprint(program rule.Program, options any) [sha256.Size]byte {
	hash := sha256.New()
	if index := importerIndexFor(program); index != nil {
		for _, path := range slices.Sorted(maps.Keys(index.byPath)) {
			entry := index.byPath[path]
			hash.Write([]byte(rule.FingerprintPath(program, string(path))))
			hash.Write([]byte{0})
			if entry.everything {
				hash.Write([]byte{1})
			}
			for _, name := range slices.Sorted(maps.Keys(entry.names)) {
				hash.Write([]byte(name))
				hash.Write([]byte{2})
			}
			hash.Write([]byte{3})
		}
		for _, directory := range index.roots.DynamicImportDirectories() {
			hash.Write([]byte(rule.FingerprintPath(program, directory)))
			hash.Write([]byte{4})
		}
	}
	var fingerprint [sha256.Size]byte
	copy(fingerprint[:], hash.Sum(nil))
	return fingerprint
}

// recordImportedNames adds what one specifier's statement takes from its target.
func recordImportedNames(entry *importedNames, specifier *ast.Node) {
	statement := specifier.Parent
	if statement == nil {
		entry.everything = true
		return
	}
	switch statement.Kind {
	case ast.KindImportDeclaration:
		// `import './x'` runs the module for its effects, and BindingsOf answers it empty: no name.
		bindings := imports.BindingsOf(statement)
		if bindings.Default != nil {
			entry.names["default"] = true
		}
		if bindings.Namespace != nil {
			entry.everything = true
			return
		}
		// The name the source module exports, never the local alias: `{ x as y }` takes x.
		for _, element := range bindings.Named {
			if importedName := imports.ImportedNameOf(element); importedName != "" {
				entry.names[importedName] = true
			}
		}
	case ast.KindExportDeclaration:
		exportClause := statement.AsExportDeclaration().ExportClause
		if exportClause == nil || exportClause.Kind != ast.KindNamedExports {
			entry.everything = true
			return
		}
		for _, element := range exportClause.AsNamedExports().Elements.Nodes {
			exportSpecifier := element.AsExportSpecifier()
			if exportSpecifier.PropertyName != nil {
				entry.names[exportSpecifier.PropertyName.Text()] = true
			} else if exportSpecifier.Name() != nil {
				entry.names[exportSpecifier.Name().Text()] = true
			}
		}
	default:
		entry.everything = true
	}
}

// provablyUnimported reports whether nothing in the program can import this name from this file,
// and the file is not one a framework or test runner loads without an import.
//
// Every doubt answers false, because a false here only keeps the older PascalCase advice, while a
// wrong true tells the author to drop an export something reads.
//
// Several names because one binding can leave its file under more than one: `export { x, x as Y }`.
// It is unimported only when every one of them is.
func (index *importerIndex) provablyUnimported(sourceFile *ast.SourceFile, names ...string) bool {
	if index == nil || sourceFile == nil || len(names) == 0 {
		return false
	}
	if isRoot, _ := index.roots.IsRoot(sourceFile.FileName().AsString()); isRoot {
		return false
	}
	entry := index.byPath[sourceFile.PathKey()]
	if entry == nil {
		return true
	}
	if entry.everything {
		return false
	}
	for _, name := range names {
		if entry.names[name] {
			return false
		}
	}
	return true
}

// exportedNamesOf returns the names a constant leaves its file under, or none when it stays local.
//
// Two ways out. The modifier (`export const X`) exports it as X. A local export clause
// (`export { X }`, `export { X as Y }`) exports it as whatever the clause says, and the clause is
// matched to the binding through the checker's local target symbol rather than by spelling, so
// `export { X } from './other'` (another module's X) and a same-spelled binding in an inner scope
// never count. `export default X` and `export { X as default }` do not export the spelling: an
// importer names a default whatever it likes, so the local name reaches nobody, which is also where
// the TypeScript original lands.
//
// Missing this cost a false finding: `PensieveCommandBoundary.ts` declares
// `const IdentitySafePensieveCommandPaths` and exports it on line 39 with `export { ... }`, and the
// rule called it file-local and told it to become camelCase, which would have renamed an export.
func exportedNamesOf(ctx rule.Context, statement *ast.Node, name *ast.Node) []string {
	if module.IsExported(statement) {
		return []string{name.Text()}
	}
	clauses := localExportClausesFor(ctx)
	if len(clauses) == 0 || ctx.TypeChecker == nil {
		return nil
	}
	// A nil here would match every clause whose own target is unresolved. A const's declaration name
	// always resolves, so no fixture reaches this; it is the guard that keeps nil from equalling nil.
	declared := ctx.TypeChecker.GetSymbolAtLocation(name)
	if declared == nil {
		return nil
	}
	var names []string
	for _, specifier := range clauses {
		exportedName := specifier.AsExportSpecifier().Name()
		if exportedName == nil || exportedName.Text() == "default" {
			continue
		}
		if ctx.TypeChecker.GetExportSpecifierLocalTargetSymbol(specifier) == declared {
			names = append(names, exportedName.Text())
		}
	}
	return names
}

// localExportClausesFor returns every specifier of every `export { ... }` in the file that has no
// `from`, read once per file. Most files have none, and then no constant pays a checker call.
//
// Leaving out the `from` clauses is cost, not correctness, and a mutation confirms it: their local
// target is the other module's binding, so the checker comparison already rejects them. The filter
// keeps them from costing a checker call per constant.
func localExportClausesFor(ctx rule.Context) []*ast.Node {
	return rule.Cached(ctx.FileCache, "nexus/constant-casing/local-export-clauses", func() []*ast.Node {
		var specifiers []*ast.Node
		if ctx.SourceFile == nil {
			return specifiers
		}
		for _, statement := range ctx.SourceFile.Statements.Nodes {
			if statement.Kind != ast.KindExportDeclaration {
				continue
			}
			declaration := statement.AsExportDeclaration()
			if declaration.ModuleSpecifier != nil || declaration.ExportClause == nil ||
				declaration.ExportClause.Kind != ast.KindNamedExports {
				continue
			}
			specifiers = append(specifiers, declaration.ExportClause.AsNamedExports().Elements.Nodes...)
		}
		return specifiers
	})
}
