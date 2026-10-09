package unused_exports

import (
	"context"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module_roots"
	"github.com/system-inc/cohere/internal/types/program"
)

// Unreferenced is one exported declaration that nothing in the program uses.
type Unreferenced struct {
	// FileName is where the declaration lives.
	FileName string

	// Name is what was declared.
	Name string

	// Kind is what sort of thing it is, for grouping in the report.
	Kind string

	// Range is the declaration's own span.
	Range TextSpan

	// Intentional is set when the author marked this as deliberately unused.
	Intentional intentionalReason

	// WholeFileUnused is set when nothing in the file is referenced, so the reader can be told about
	// the file rather than handed a list of its exports one at a time.
	WholeFileUnused bool
}

// FileUnused is a file nothing imports.
type FileUnused struct {
	FileName    string
	Exports     int
	Intentional intentionalReason
}

// Report is what one `--unused` run found.
type Report struct {
	Unreferenced []Unreferenced
	Files        []FileUnused
	Unreachable  []Unreachable

	// FilesAnalyzed is the population the findings were drawn from, printed alongside them because a
	// finding count with no denominator cannot be told apart from a broken probe.
	FilesAnalyzed int

	// RootedFiles is how many files were spared by the root set, which is the number that says
	// whether the root set is doing anything. A root set that spares zero files on a Next.js tree is
	// broken, and a report that does not state the number cannot reveal that.
	RootedFiles int

	// ExportsAnalyzed is how many exported declarations were examined.
	ExportsAnalyzed int

	// Islands are the clusters the transitive closure found dead together. Empty unless the closure
	// was asked for, which is the point: the flat sections above make a weaker claim that does not
	// depend on the root set being complete.
	Islands []Island

	// DeclarationsAnalyzed is the closure's population — every declaration, not only the exported
	// ones — printed for the same reason as the others: a cluster count with no denominator cannot be
	// told apart from a broken inventory.
	DeclarationsAnalyzed int

	// ClosurePasses is the longest chain THIS traversal followed, which is a lower bound on the
	// graph's depth rather than a property of it: a worklist reaches a symbol by whichever path it
	// pops first, so the number moves between runs while the finding does not. Reported because a
	// closure that stops at depth 1 on a real tree has failed to propagate, and that failure produces
	// a bigger report rather than an error.
	ClosurePasses int
}

// referenceIndex is the edge set: which declarations reference which symbols.
//
// It replaced a plain `map[*ast.Symbol]bool` that recorded only THAT a symbol was used. The walk is
// identical and costs the same; what changed is what it stores. Recording the edge is what makes
// the transitive question answerable, and the transitive question is the one that finds an island —
// an abandoned export plus the private world that existed only to serve it.
type referenceIndex struct {
	// referencedBy maps a symbol to the declarations that reference it. A symbol present with an
	// empty slice is impossible; a symbol referenced only from module top level appears with
	// `nil` in its slice, which is the root marker rather than a missing entry.
	referencedBy map[*ast.Symbol][]*ast.Symbol

	// referenced is the flat question the original set answered, kept because the flat report is
	// genuinely useful and a reader may trust it more than the closure.
	referenced map[*ast.Symbol]bool
}

// record adds one edge. A nil `from` means the reference sits at module top level, which is a ROOT
// and not an absent referrer — the distinction the whole closure turns on.
func (index *referenceIndex) record(to *ast.Symbol, from *ast.Symbol) {
	index.referenced[to] = true
	index.referencedBy[to] = append(index.referencedBy[to], from)
}

// enclosingDeclarationNode walks up from a reference to the declaration that contains it, or nil
// when the reference sits at module top level.
//
// # Why the walk stops where it does
//
// The set of kinds is the set of things that can THEMSELVES be dead. A reference inside an `if`
// block belongs to the function containing the block, because deleting the function deletes the
// reference; the block is not a unit anyone deletes on its own. Function, class, interface, type,
// enum, variable, method, accessor, constructor, property, module and export-assignment are the
// units, and everything else is passed through.
//
// # What was probed rather than assumed
//
// The three shapes where "what declaration am I inside" gets strange were measured on a fixture
// built for exactly this, not reasoned about:
//
//	JSX          `<div title={String(usedInJsx())}>` attributes both `String` and `usedInJsx` to the
//	             enclosing function. A JSX expression container does not break the parent chain.
//	decorators   `@usedByDecorator() export class DecoratedClass {}` attributes the decorator's
//	             reference to `DecoratedClass`, because the decorator is a child of the class node.
//	             That is the answer this analysis wants: if the class is dead its decorator's
//	             reference should die with it, rather than escaping to module level as a root.
//	default      `export default function defaultFn()` attributes to `defaultFn`. The anonymous
//	exports      `export default arrow` form lands on the ExportAssignment, which has no name — see
//	             `symbolOfDeclaration` for why that case still resolves.
//
// Returning nil at `KindSourceFile` rather than falling out of the loop is deliberate: a module-top
// -level reference must be distinguishable from a walk that ran off the end of a detached subtree,
// and both would otherwise produce the same nil.
func enclosingDeclarationNode(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindFunctionDeclaration,
			ast.KindClassDeclaration,
			ast.KindInterfaceDeclaration,
			ast.KindTypeAliasDeclaration,
			ast.KindEnumDeclaration,
			ast.KindVariableDeclaration,
			ast.KindMethodDeclaration,
			ast.KindPropertyDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor,
			ast.KindConstructor,
			ast.KindExportAssignment,
			ast.KindModuleDeclaration:
			return current
		case ast.KindSourceFile:
			return nil
		}
	}
	return nil
}

// symbolOfDeclaration returns a declaration node's own symbol without asking the checker.
//
// # Why this is not a second GetSymbolAtLocation
//
// The checker call is the expensive primitive in this phase, and resolving "who is doing the
// referencing" the obvious way — walk to the enclosing declaration, take its name, ask the checker —
// roughly doubles how many are made. `node.Symbol()` is populated by the binder when the program is
// built, which is exactly the situation this phase runs in, so the answer is already sitting on the
// node.
//
// Probed on the fixture rather than assumed, because a nil here would silently sever every edge and
// read as a clean codebase: 18 of 18 enclosing declarations agreed between `node.Symbol()` and a
// second `GetSymbolAtLocation` on the declaration's name, with zero disagreements and zero nils from
// the free path.
//
// The free path is also strictly BETTER on one shape. `export default arrow` produces an anonymous
// ExportAssignment with no name for `GetSymbolAtLocation` to be given, so the checker path returns
// nil and drops the edge, while `node.Symbol()` returns the `default` symbol regardless. The cheaper
// call is the more complete one, which is not the usual direction and is why it is written down.
func symbolOfDeclaration(node *ast.Node) *ast.Symbol {
	if node == nil {
		return nil
	}
	return node.Symbol()
}

// FindUnreferenced builds a program-wide reverse index and reports the exported declarations
// nothing uses.
//
// # Why a program-wide index rather than per-rule symbol comparison
//
// `internal/rule/rule.go` argues at length against putting a references-to-a-binding helper on the
// shelf, and that argument is correct for every rule shipped so far. It does not govern this,
// and the comment says so itself in its own last paragraph: "Find-all-references answers 'where is
// this used', and these rules already know where to look, since they walk one file. They need 'is
// this the same binding', which is a symbol comparison. Reaching for a reverse index here would be
// answering a forward question with the wrong instrument."
//
// This is the forward question. "Where is this used" is the entire output, the scope is the program
// rather than one file, and there is no cheaper tier that is correct: name matching cannot tell an
// unused `Button` from a used one in another file, and per-file symbol comparison cannot answer a
// whole-program question at all. The cost the rule comment refuses to hide behind an import is
// stated here instead — this is a phase, run when asked for, not a rule that every file pays for.
//
// # What the index is keyed on
//
// Symbol pointers, established by probe on this substrate before anything was built on it: 42 of 42
// imported names resolved through `GetAliasedSymbol` to a declaration in another file, with zero nil
// symbols. Pointer identity is what makes a reference in one file and a declaration in another the
// same fact.
func FindUnreferenced(ctx context.Context, graph *program.Graph, files []*ast.SourceFile, roots *module_roots.Set, deep bool) (*Report, error) {
	report := &Report{FilesAnalyzed: len(files)}

	// The closure's inputs, collected during the passes below rather than in a third walk. Both stay
	// empty when `deep` is false, so the flat report pays nothing for a feature it does not use.
	var inventory []declarationRecord
	rootSymbols := map[*ast.Symbol]bool{}

	// Pass one: every symbol that anything anywhere refers to, and WHO refers to it.
	//
	// Built from identifier positions rather than from import statements, because a reference is a
	// use wherever it appears — a call, a type position, a JSX tag, a re-export — and enumerating the
	// forms would miss whichever one nobody thought of. Resolving every identifier is more work and
	// it cannot have that class of gap.
	//
	// The referrer is resolved without a second checker call, which is the difference between this
	// costing what the set cost and costing twice that. See `symbolOfDeclaration`.
	index := &referenceIndex{
		referencedBy: make(map[*ast.Symbol][]*ast.Symbol, len(files)*8),
		referenced:   make(map[*ast.Symbol]bool, len(files)*8),
	}

	for _, file := range files {
		checker, release := graph.CheckerForFile(ctx, file)
		if checker == nil {
			release()
			continue
		}

		var visit func(node *ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node == nil {
				return false
			}
			if node.Kind == ast.KindIdentifier {
				// A declaration's own name is not a use of it. Without this the index marks
				// everything referenced and the report finds nothing — the exact broken-probe shape
				// that reads as a clean codebase.
				if !isDeclarationName(node) {
					if symbol := checker.GetSymbolAtLocation(node); symbol != nil {
						// Resolved once per reference, not once per edge, because the alias and its
						// target share a referrer.
						from := symbolOfDeclaration(enclosingDeclarationNode(node))
						index.record(symbol, from)
						// An import binding is an alias; the thing actually used is what it points
						// at. Recording only the alias leaves every real declaration looking unused.
						if symbol.Flags&ast.SymbolFlagsAlias != 0 {
							if aliased := checker.GetAliasedSymbol(symbol); aliased != nil {
								index.record(aliased, from)
							}
						}
					}
				}
			}
			node.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
		release()
	}

	// Pass two: every exported declaration, asked whether the index knows it.
	for _, file := range files {
		fileName := file.FileName()
		if isRoot, _ := roots.IsRoot(fileName.AsString()); isRoot {
			report.RootedFiles++
			// A spared file's declarations are roots of the closure, not merely absent from the flat
			// report. A `page.tsx` is loaded by the framework with no import anywhere, so everything
			// it reaches is alive, and omitting these would darken every component the app renders.
			if deep {
				collectRootSymbols(file, rootSymbols)
			}
			continue
		}

		checker, release := graph.CheckerForFile(ctx, file)
		if checker == nil {
			release()
			continue
		}

		text := file.Text()
		fileExports := 0
		fileUnusedExports := 0
		var fileFindings []Unreferenced

		// The closure needs EVERY declaration, not only the exported ones. The private helper that
		// existed to serve an abandoned export is the second half of an island, and a report that
		// inventoried only exports could never find it.
		if deep {
			inventory = append(inventory, fileDeclarations(file, text)...)
		}

		for _, statement := range file.Statements.Nodes {
			for _, declaration := range exportedDeclarations(statement) {
				name := declaration.name
				if name == nil {
					continue
				}
				fileExports++
				report.ExportsAnalyzed++

				symbol := checker.GetSymbolAtLocation(name)
				if symbol == nil {
					// A name that does not resolve is a question this analysis cannot answer, and
					// answering it anyway is how live code gets reported as dead. Skipped and
					// counted rather than guessed.
					continue
				}
				if index.referenced[symbol] {
					continue
				}

				fileUnusedExports++
				fileFindings = append(fileFindings, Unreferenced{
					FileName:    fileName.AsString(),
					Name:        name.Text(),
					Kind:        declaration.kind,
					Range:       TextSpan{Position: statement.Pos(), End: statement.End()},
					Intentional: findIntentionalMarker(leadingText(text, statement)),
				})
			}
		}

		// A file where every export is unused reads better as one finding about the file than as a
		// list of its parts. That is the case the brief calls the insight: an abandoned direction
		// drags a small private world behind it, and naming the file is what makes the world visible.
		if fileExports > 0 && fileUnusedExports == fileExports {
			intentional := intentionalReason{}
			for _, finding := range fileFindings {
				if finding.Intentional.Present {
					intentional = finding.Intentional
					break
				}
			}
			report.Files = append(report.Files, FileUnused{
				FileName:    fileName.AsString(),
				Exports:     fileExports,
				Intentional: intentional,
			})
		} else {
			report.Unreferenced = append(report.Unreferenced, fileFindings...)
		}

		release()
	}

	if deep {
		report.Islands, report.ClosurePasses = computeIslands(index, inventory, rootSymbols)
		report.DeclarationsAnalyzed = len(inventory)
	}

	return report, nil
}

// exportedDeclaration is one exported name and what kind of thing it names.
type exportedDeclaration struct {
	name *ast.Node
	kind string
}

// exportedDeclarations returns the names a statement exports.
//
// Only `export` declarations are considered. A module-private helper that nothing calls is a real
// finding and a different one: it is decidable within the file, needs no index, and belongs to a
// lint rule rather than to a whole-program report. Keeping this to exports is what keeps the claim
// this report makes a whole-program claim.
func exportedDeclarations(statement *ast.Node) []exportedDeclaration {
	if !hasExportModifier(statement) {
		return nil
	}

	switch statement.Kind {
	case ast.KindFunctionDeclaration:
		return one(statement.Name(), "function")
	case ast.KindClassDeclaration:
		return one(statement.Name(), "class")
	case ast.KindInterfaceDeclaration:
		return one(statement.Name(), "interface")
	case ast.KindTypeAliasDeclaration:
		return one(statement.Name(), "type")
	case ast.KindEnumDeclaration:
		return one(statement.Name(), "enum")
	case ast.KindVariableStatement:
		declarationList := statement.AsVariableStatement().DeclarationList
		if declarationList == nil {
			return nil
		}
		var declarations []exportedDeclaration
		for _, declaration := range declarationList.AsVariableDeclarationList().Declarations.Nodes {
			name := declaration.Name()
			// Only a plain identifier. A destructuring pattern exports several names whose
			// individual liveness this pass does not separate, and reporting the whole statement
			// because one name is unused would be wrong.
			if name != nil && name.Kind == ast.KindIdentifier {
				declarations = append(declarations, exportedDeclaration{name: name, kind: "constant"})
			}
		}
		return declarations
	}
	return nil
}

func one(name *ast.Node, kind string) []exportedDeclaration {
	if name == nil {
		return nil
	}
	return []exportedDeclaration{{name: name, kind: kind}}
}

// hasExportModifier reports whether a statement carries `export`.
//
// A default export is deliberately excluded: it is imported under whatever name the consumer
// chooses, so the name at the declaration site is not the name at the use site, and the symbol
// identity this index is keyed on does not carry across reliably. Reporting one would risk a false
// positive on the most load-bearing kind of export there is.
func hasExportModifier(statement *ast.Node) bool {
	modifiers := statement.Modifiers()
	if modifiers == nil {
		return false
	}
	hasExport := false
	for _, modifier := range modifiers.Nodes {
		switch modifier.Kind {
		case ast.KindExportKeyword:
			hasExport = true
		case ast.KindDefaultKeyword:
			return false
		case ast.KindDeclareKeyword:
			// An ambient declaration describes something that exists elsewhere. Nothing here can
			// establish whether it is used.
			return false
		}
	}
	return hasExport
}

// isDeclarationName reports whether an identifier is the name being declared rather than a use of
// something declared elsewhere.
func isDeclarationName(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	if parent.Name() == node {
		switch parent.Kind {
		case ast.KindFunctionDeclaration,
			ast.KindClassDeclaration,
			ast.KindInterfaceDeclaration,
			ast.KindTypeAliasDeclaration,
			ast.KindEnumDeclaration,
			ast.KindVariableDeclaration,
			ast.KindParameter,
			ast.KindPropertyDeclaration,
			ast.KindMethodDeclaration,
			ast.KindEnumMember,
			ast.KindPropertySignature,
			ast.KindMethodSignature,
			ast.KindImportSpecifier,
			ast.KindImportClause,
			ast.KindNamespaceImport,
			ast.KindExportSpecifier,
			ast.KindBindingElement,
			ast.KindTypeParameter,
			ast.KindModuleDeclaration:
			return true
		}
	}
	return false
}

// leadingText returns the trivia before a statement, where a marker comment lives.
func leadingText(text string, statement *ast.Node) string {
	start := statement.Pos()
	end := statement.End()
	if start < 0 || start > len(text) {
		return ""
	}
	// The statement's Pos() is the start of its leading trivia in this parser, so the span from Pos
	// to the first non-trivia character is the comment block. Taking the whole statement head is
	// simpler and cannot miss a marker written on the same line.
	head := end
	if head > len(text) {
		head = len(text)
	}
	region := text[start:head]
	if newline := strings.LastIndex(region[:min(len(region), 400)], "\n"); newline >= 0 {
		return region[:newline]
	}
	return region[:min(len(region), 400)]
}
