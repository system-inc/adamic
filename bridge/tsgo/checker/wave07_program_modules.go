package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

// wave07ProgramModules exposes source files and syntax-level import resolutions.
// Entry status, transitive closure, blocking, and findings are native decisions.
func (p *Program) wave07ProgramModules(root *ast.Node, question string) (string, error) {
	if question != "wave07-program-modules" || root.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("wave07-program-modules requires SourceFile")
	}
	out := &fields{}
	out.number(1)
	out.text(question)
	files := p.Compiler.GetSourceFiles()
	out.number(uint64(len(files)))
	for _, file := range files {
		out.text(file.FileName().AsString())
		out.yes(file.IsDeclarationFile)
		text := ""
		if !file.IsDeclarationFile {
			text = file.Text()
		}
		out.text(text)
		var nodes []*ast.Node
		var specifiers []*ast.Node
		var typeOnly, importCall, requireCall []bool
		add := func(node, specifier *ast.Node, only, imports, requires bool) {
			nodes = append(nodes, node)
			specifiers = append(specifiers, specifier)
			typeOnly = append(typeOnly, only)
			importCall = append(importCall, imports)
			requireCall = append(requireCall, requires)
		}
		if !file.IsDeclarationFile {
			for _, node := range file.Statements.Nodes {
				switch node.Kind {
				case ast.KindImportDeclaration:
					d := node.AsImportDeclaration()
					add(node, d.ModuleSpecifier, d.ImportClause != nil && d.ImportClause.IsTypeOnly(), false, false)
				case ast.KindExportDeclaration:
					d := node.AsExportDeclaration()
					add(node, d.ModuleSpecifier, d.IsTypeOnly, false, false)
				case ast.KindImportEqualsDeclaration:
					d := node.AsImportEqualsDeclaration()
					var specifier *ast.Node
					if d.ModuleReference != nil && ast.IsExternalModuleReference(d.ModuleReference) {
						specifier = d.ModuleReference.AsExternalModuleReference().Expression
					}
					add(node, specifier, d.IsTypeOnly, false, false)
				}
			}
			var walk func(*ast.Node)
			walk = func(node *ast.Node) {
				if node.Kind == ast.KindCallExpression && (ast.IsImportCall(node) || ast.IsRequireCall(node, false)) {
					var specifier *ast.Node
					if a := node.AsCallExpression().Arguments; a != nil && len(a.Nodes) > 0 {
						specifier = a.Nodes[0]
					}
					add(node, specifier, false, ast.IsImportCall(node), ast.IsRequireCall(node, false))
				}
				node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
			}
			walk(file.AsNode())
		}
		out.number(uint64(len(nodes)))
		for index, node := range nodes {
			out.text(strings.TrimPrefix(node.Kind.String(), "Kind"))
			out.yes(typeOnly[index])
			out.yes(importCall[index])
			out.yes(requireCall[index])
			specifier := specifiers[index]
			literal := specifier != nil && ast.IsStringLiteralLike(specifier)
			out.yes(literal)
			target := ""
			if literal {
				resolved := p.Compiler.GetResolvedModuleFromModuleSpecifier(file, specifier)
				if resolved != nil && resolved.IsResolved() {
					if next := p.Compiler.GetSourceFileForResolvedModule(resolved); next != nil {
						target = next.FileName().AsString()
					}
				}
			}
			out.text(target)
		}
	}
	return out.String(), nil
}
