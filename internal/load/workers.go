package load

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// WorkerKV is the prelude export's symbol identity, never a program's spelling.
func WorkerKV(types *checker.Checker, t *checker.Type) bool {
	module := types.TryFindAmbientModule("adamic/workers")
	if module == nil || t == nil || t.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	for _, symbol := range types.GetExportsOfModule(module) {
		if symbol.Name != "KvNamespace" || t.Symbol() != symbol {
			continue
		}
		for _, d := range symbol.Declarations {
			if IsPrelude(ast.GetSourceFileOfNode(d)) {
				return true
			}
		}
	}
	return false
}

// WorkersImport finds the first source import, including type-only dependencies.
// Every program lowers through it, so a file whose text never names the module is
// skipped before any node is visited, and only a file that names it is walked: an
// import or export declaration, an import-equals' external module reference, or an
// import type anywhere in a type. Two spellings are not found here, and both fail
// closed elsewhere: import("adamic/workers") as a call, which stops as a dynamic
// import, and a declare module 'adamic/workers' augmentation, whose members no
// registration names.
func (p *Program) WorkersImport() *ast.Node {
	for _, file := range p.compiler.GetSourceFiles() {
		if IsPrelude(file) || IsLibrary(file) || IsNodeLibrary(file) || !strings.Contains(file.Text(), "adamic/workers") {
			continue
		}
		var found *ast.Node
		var visit ast.Visitor
		visit = func(n *ast.Node) bool {
			if n.Kind == ast.KindStringLiteral && n.Text() == "adamic/workers" {
				parent := n.Parent
				if parent != nil && (parent.Kind == ast.KindImportDeclaration || parent.Kind == ast.KindExportDeclaration || parent.Kind == ast.KindExternalModuleReference || parent.Kind == ast.KindLiteralType && parent.Parent != nil && parent.Parent.Kind == ast.KindImportType) {
					found = n
					return true
				}
			}
			return n.ForEachChild(visit)
		}
		file.AsNode().ForEachChild(visit)
		if found != nil {
			return found
		}
	}
	return nil
}
