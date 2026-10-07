package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

func init() { questionExtensions["program-imports"] = programImportsQuestion }

// The records expose syntax and module resolution. Reachability and entry
// classification are left to Adamic.
func programImportsQuestion(p *Program, out *fields, c *checker.Checker, source *ast.SourceFile, node *ast.Node, question string) (string, error) {
	if question != "program-imports" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("program-imports requires SourceFile")
	}
	files := p.Compiler.SourceFiles()
	out.number(uint64(len(files)))
	for _, file := range files {
		out.text(file.FileName())
		out.yes(file.IsDeclarationFile)
		var rows []*ast.Node
		var visit func(*ast.Node) bool
		visit = func(n *ast.Node) bool {
			if n.Kind == ast.KindImportDeclaration || n.Kind == ast.KindExportDeclaration || n.Kind == ast.KindImportEqualsDeclaration || (n.Kind == ast.KindCallExpression && (ast.IsImportCall(n) || ast.IsRequireCall(n, false))) {
				rows = append(rows, n)
			}
			n.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
		out.number(uint64(len(rows)))
		for _, n := range rows {
			var specifier *ast.Node
			typeOnly := false
			switch n.Kind {
			case ast.KindImportDeclaration:
				d := n.AsImportDeclaration()
				specifier = d.ModuleSpecifier
				typeOnly = d.ImportClause != nil && d.ImportClause.IsTypeOnly()
			case ast.KindExportDeclaration:
				d := n.AsExportDeclaration()
				specifier = d.ModuleSpecifier
				typeOnly = d.IsTypeOnly
			case ast.KindImportEqualsDeclaration:
				d := n.AsImportEqualsDeclaration()
				typeOnly = d.IsTypeOnly
				if d.ModuleReference != nil && ast.IsExternalModuleReference(d.ModuleReference) {
					specifier = d.ModuleReference.AsExternalModuleReference().Expression
				}
			case ast.KindCallExpression:
				if a := n.AsCallExpression().Arguments; a != nil && len(a.Nodes) > 0 {
					specifier = a.Nodes[0]
				}
			}
			out.text(strings.TrimPrefix(n.Kind.String(), "Kind"))
			out.yes(typeOnly)
			literal := specifier != nil && ast.IsStringLiteralLike(specifier)
			out.yes(literal)
			resolvedFile := ""
			if literal {
				if resolved := p.Compiler.GetResolvedModuleFromModuleSpecifier(file, specifier); resolved != nil && resolved.IsResolved() {
					if target := p.Compiler.GetSourceFileForResolvedModule(resolved.ResolvedFileName); target != nil {
						resolvedFile = target.FileName()
					}
				}
			}
			out.text(resolvedFile)
		}
	}
	return out.String(), nil
}
