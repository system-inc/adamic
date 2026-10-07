package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

// Source files and module-specifier resolutions only. Native code chooses the edges.
func (p *Program) programModuleResolution(out *fields, node *ast.Node, question string) (string, error) {
	if question != "program-module-resolution" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("program-module-resolution requires a SourceFile")
	}
	files := p.Compiler.SourceFiles()
	out.number(uint64(len(files)))
	for _, file := range files {
		out.text(file.FileName())
		out.yes(file.IsDeclarationFile)
		text := ""
		if !file.IsDeclarationFile {
			text = file.Text()
		}
		out.text(text)
		var specifiers []*ast.Node
		if !file.IsDeclarationFile {
			var visit func(*ast.Node) bool
			visit = func(n *ast.Node) bool {
				var specifier *ast.Node
				switch n.Kind {
				case ast.KindImportDeclaration:
					specifier = n.AsImportDeclaration().ModuleSpecifier
				case ast.KindExportDeclaration:
					specifier = n.AsExportDeclaration().ModuleSpecifier
				case ast.KindImportEqualsDeclaration:
					r := n.AsImportEqualsDeclaration().ModuleReference
					if r != nil && ast.IsExternalModuleReference(r) {
						specifier = r.AsExternalModuleReference().Expression
					}
				case ast.KindCallExpression:
					if ast.IsImportCall(n) || ast.IsRequireCall(n, false) {
						a := n.AsCallExpression().Arguments
						if a != nil && len(a.Nodes) > 0 {
							specifier = a.Nodes[0]
						}
					}
				}
				if specifier != nil && ast.IsStringLiteralLike(specifier) {
					specifiers = append(specifiers, specifier)
				}
				n.ForEachChild(visit)
				return false
			}
			file.AsNode().ForEachChild(visit)
		}
		out.number(uint64(len(specifiers)))
		for _, specifier := range specifiers {
			out.text(strings.TrimPrefix(specifier.Kind.String(), "Kind"))
			out.number(uint64(specifier.Pos()))
			out.number(uint64(specifier.End()))
			target := ""
			resolved := p.Compiler.GetResolvedModuleFromModuleSpecifier(file, specifier)
			if resolved != nil && resolved.IsResolved() {
				if source := p.Compiler.GetSourceFileForResolvedModule(resolved.ResolvedFileName); source != nil {
					target = source.FileName()
				}
			}
			out.text(target)
		}
	}
	return out.String(), nil
}
