package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// programModules exposes loaded source text and resolved import syntax. It does
// not decide import reachability, entry status or whether any call blocks.
func (p *Program) programModules(out *fields, question string) (string, error) {
	if question != "program-modules" {
		return "", fmt.Errorf("unexpected program-modules suffix")
	}
	files := p.Compiler.SourceFiles()
	out.number(uint64(len(files)))
	for _, file := range files {
		out.text(file.FileName())
		out.yes(file.IsDeclarationFile)
		out.yes(ast.IsExternalModule(file))
		if file.IsDeclarationFile {
			out.text("")
		} else {
			out.text(file.Text())
		}
		type edge struct {
			node, specifier  *ast.Node
			typeOnly, clause bool
		}
		var edges []edge
		if !file.IsDeclarationFile {
			for _, node := range file.Statements.Nodes {
				switch node.Kind {
				case ast.KindImportDeclaration:
					d := node.AsImportDeclaration()
					edges = append(edges, edge{node: node, specifier: d.ModuleSpecifier, typeOnly: d.ImportClause != nil && d.ImportClause.IsTypeOnly(), clause: d.ImportClause != nil})
				case ast.KindExportDeclaration:
					d := node.AsExportDeclaration()
					if d.ModuleSpecifier != nil {
						edges = append(edges, edge{node: node, specifier: d.ModuleSpecifier, typeOnly: d.IsTypeOnly})
					}
				case ast.KindImportEqualsDeclaration:
					d := node.AsImportEqualsDeclaration()
					if d.ModuleReference != nil && ast.IsExternalModuleReference(d.ModuleReference) {
						edges = append(edges, edge{node: node, specifier: d.ModuleReference.AsExternalModuleReference().Expression, typeOnly: d.IsTypeOnly})
					}
				}
			}
			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool {
				if node.Kind == ast.KindCallExpression && (ast.IsImportCall(node) || ast.IsRequireCall(node, false)) {
					var specifier *ast.Node
					if args := node.AsCallExpression().Arguments; args != nil && len(args.Nodes) > 0 {
						specifier = args.Nodes[0]
					}
					edges = append(edges, edge{node: node, specifier: specifier})
				}
				node.ForEachChild(visit)
				return false
			}
			file.AsNode().ForEachChild(visit)
		}
		out.number(uint64(len(edges)))
		for _, edge := range edges {
			out.text(strings.TrimPrefix(edge.node.Kind.String(), "Kind"))
			out.number(uint64(edge.node.Pos()))
			out.number(uint64(edge.node.End()))
			out.yes(edge.typeOnly)
			out.yes(edge.clause)
			literal := edge.specifier != nil && ast.IsStringLiteralLike(edge.specifier)
			out.yes(literal)
			target := ""
			if literal {
				resolved := p.Compiler.GetResolvedModuleFromModuleSpecifier(file, edge.specifier)
				if resolved != nil && resolved.IsResolved() {
					if source := p.Compiler.GetSourceFileForResolvedModule(resolved.ResolvedFileName); source != nil {
						target = source.FileName()
					}
				}
			}
			out.text(target)
		}
	}
	return out.String(), nil
}
