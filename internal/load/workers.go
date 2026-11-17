package load

import (
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
func (p *Program) WorkersImport() *ast.Node {
	for _, file := range p.compiler.GetSourceFiles() {
		if IsPrelude(file) || IsLibrary(file) || IsNodeLibrary(file) {
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
