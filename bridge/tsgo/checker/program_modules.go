package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

func (p *Program) programModules(_ *checker.Checker, node *ast.Node, question string) (string, error) {
	const mode = "program-modules"
	if question != mode || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("program-modules requires a SourceFile without suffix")
	}
	out := &fields{}
	out.number(1)
	out.text(mode)
	files := p.Compiler.SourceFiles()
	out.number(uint64(len(files)))
	for _, source := range files {
		out.text(source.FileName())
		out.yes(source.IsDeclarationFile)
		type specifier struct {
			kind string
			only bool
			node *ast.Node
		}
		var list []specifier
		for _, statement := range source.Statements.Nodes {
			switch statement.Kind {
			case ast.KindImportDeclaration:
				d := statement.AsImportDeclaration()
				list = append(list, specifier{"import", d.ImportClause != nil && d.ImportClause.IsTypeOnly(), d.ModuleSpecifier})
			case ast.KindExportDeclaration:
				d := statement.AsExportDeclaration()
				if d.ModuleSpecifier != nil {
					list = append(list, specifier{"export", d.IsTypeOnly, d.ModuleSpecifier})
				}
			case ast.KindImportEqualsDeclaration:
				d := statement.AsImportEqualsDeclaration()
				if d.ModuleReference != nil && ast.IsExternalModuleReference(d.ModuleReference) {
					list = append(list, specifier{"import-equals", d.IsTypeOnly, d.ModuleReference.AsExternalModuleReference().Expression})
				}
			}
		}
		// Preserve raw call-import positions; the native consumer applies the exact
		// production text fast-path and declaration-file exclusions.
		if strings.Contains(source.Text(), "import(") || strings.Contains(source.Text(), "require(") {
			var visit func(*ast.Node)
			visit = func(n *ast.Node) {
				if n.Kind == ast.KindCallExpression && (ast.IsImportCall(n) || ast.IsRequireCall(n, false)) {
					var arg *ast.Node
					if a := n.AsCallExpression().Arguments; a != nil && len(a.Nodes) > 0 {
						arg = a.Nodes[0]
					}
					list = append(list, specifier{"call", false, arg})
				}
				n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
			}
			visit(source.AsNode())
		}
		out.number(uint64(len(list)))
		for _, s := range list {
			out.text(s.kind)
			out.yes(s.only)
			literal := s.node != nil && ast.IsStringLiteralLike(s.node)
			out.yes(literal)
			target := ""
			if literal {
				resolved := p.Compiler.GetResolvedModuleFromModuleSpecifier(source, s.node)
				if resolved != nil && resolved.IsResolved() {
					if file := p.Compiler.GetSourceFileForResolvedModule(resolved.ResolvedFileName); file != nil {
						target = file.FileName()
					}
				}
			}
			out.text(target)
		}
	}
	return out.String(), nil
}
