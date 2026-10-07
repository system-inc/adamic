package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Program file membership and resolved module locations, including declaration
// files. Edge selection and cycle detection belong to the Adamic caller.
func (p *Program) moduleLinks(out *fields, node *ast.Node, question string) error {
	if node.Kind != ast.KindSourceFile || question != "module-links" {
		return fmt.Errorf("module-links requires SourceFile")
	}
	files := p.Compiler.SourceFiles()
	out.number(uint64(len(files)))
	for _, file := range files {
		out.text(file.FileName())
		out.yes(file.IsDeclarationFile)
		var specifiers []*ast.Node
		for _, statement := range file.Statements.Nodes {
			switch statement.Kind {
			case ast.KindImportDeclaration:
				specifiers = append(specifiers, statement.AsImportDeclaration().ModuleSpecifier)
			case ast.KindExportDeclaration:
				if s := statement.AsExportDeclaration().ModuleSpecifier; s != nil {
					specifiers = append(specifiers, s)
				}
			}
		}
		out.number(uint64(len(specifiers)))
		for _, specifier := range specifiers {
			out.number(uint64(specifier.Pos()))
			out.number(uint64(specifier.End()))
			target := ""
			if ast.IsStringLiteralLike(specifier) {
				resolved := p.Compiler.GetResolvedModuleFromModuleSpecifier(file, specifier)
				if resolved != nil && resolved.IsResolved() {
					if found := p.Compiler.GetSourceFileForResolvedModule(resolved.ResolvedFileName); found != nil {
						target = found.FileName()
					}
				}
			}
			out.text(target)
		}
	}
	return nil
}
