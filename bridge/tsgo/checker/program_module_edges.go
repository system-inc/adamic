package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// One fact per module syntax occurrence. Runtime/type-only selection, entry
// classification and transitive blocking closure are exclusively native.
func (p *Program) programModuleEdges(out *fields, node *ast.Node, question string) (string, error) {
	if question != "program-module-edges" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("program-module-edges requires a SourceFile")
	}
	files := p.Compiler.GetSourceFiles()
	out.number(uint64(len(files)))
	for _, file := range files {
		out.text(file.FileName())
		out.yes(file.IsDeclarationFile)
		type edge struct {
			specifier *ast.Node
			kind      string
			typeOnly  bool
		}
		var edges []edge
		if !file.IsDeclarationFile {
			for _, statement := range file.Statements.Nodes {
				switch statement.Kind {
				case ast.KindImportDeclaration:
					d := statement.AsImportDeclaration()
					edges = append(edges, edge{d.ModuleSpecifier, "import", d.ImportClause != nil && d.ImportClause.IsTypeOnly()})
				case ast.KindExportDeclaration:
					d := statement.AsExportDeclaration()
					edges = append(edges, edge{d.ModuleSpecifier, "export", d.IsTypeOnly})
				case ast.KindImportEqualsDeclaration:
					d := statement.AsImportEqualsDeclaration()
					if d.ModuleReference != nil && ast.IsExternalModuleReference(d.ModuleReference) {
						edges = append(edges, edge{d.ModuleReference.AsExternalModuleReference().Expression, "import-equals", d.IsTypeOnly})
					}
				}
			}
			// The pinned consumer's textual fast path is preserved as an input fact.
			if strings.Contains(file.Text(), "import(") || strings.Contains(file.Text(), "require(") {
				var walk func(*ast.Node) bool
				walk = func(n *ast.Node) bool {
					if n.Kind == ast.KindCallExpression && (ast.IsImportCall(n) || ast.IsRequireCall(n, false)) {
						var arg *ast.Node
						if a := n.AsCallExpression().Arguments; a != nil && len(a.Nodes) > 0 {
							arg = a.Nodes[0]
						}
						edges = append(edges, edge{arg, "load-call", false})
					}
					n.ForEachChild(walk)
					return false
				}
				file.AsNode().ForEachChild(walk)
			}
		}
		out.number(uint64(len(edges)))
		for _, e := range edges {
			out.text(e.kind)
			out.yes(e.typeOnly)
			literal := e.specifier != nil && ast.IsStringLiteralLike(e.specifier)
			out.yes(literal)
			path := ""
			if literal {
				resolved := p.Compiler.GetResolvedModuleFromModuleSpecifier(file, e.specifier)
				if resolved != nil {
					target := p.Compiler.GetSourceFileForResolvedModule(resolved.ResolvedFileName)
					if target != nil {
						path = target.FileName()
					}
				}
			}
			out.text(path)
		}
	}
	return out.String(), nil
}
