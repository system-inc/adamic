package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

// runtimeModules emits the program's runtime import edges and computed loads.
// Reachability to any named module and lint policy remain native judgments.
func (p *Program) runtimeModules(node *ast.Node, question string) (string, error) {
	if question != "runtime-modules" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("runtime-modules requires an unsuffixed source")
	}
	out := &fields{}
	out.number(1)
	out.text(question)
	files := p.Compiler.SourceFiles()
	out.number(uint64(len(files)))
	for _, file := range files {
		out.text(file.FileName().AsString())
		out.yes(file.IsDeclarationFile)
		var edges []string
		computed := false
		add := func(specifier *ast.Node) {
			if specifier == nil || !ast.IsStringLiteralLike(specifier) {
				return
			}
			resolved := p.Compiler.GetResolvedModuleFromModuleSpecifier(file, specifier)
			if resolved == nil || !resolved.IsResolved() {
				return
			}
			target := p.Compiler.GetSourceFileForResolvedModule(resolved)
			if target != nil && target != file {
				edges = append(edges, target.FileName().AsString())
			}
		}
		if !file.IsDeclarationFile {
			for _, statement := range file.Statements.Nodes {
				switch statement.Kind {
				case ast.KindImportDeclaration:
					d := statement.AsImportDeclaration()
					if d.ImportClause == nil || !d.ImportClause.IsTypeOnly() {
						add(d.ModuleSpecifier)
					}
				case ast.KindExportDeclaration:
					d := statement.AsExportDeclaration()
					if !d.IsTypeOnly {
						add(d.ModuleSpecifier)
					}
				case ast.KindImportEqualsDeclaration:
					d := statement.AsImportEqualsDeclaration()
					if !d.IsTypeOnly && d.ModuleReference != nil && ast.IsExternalModuleReference(d.ModuleReference) {
						add(d.ModuleReference.AsExternalModuleReference().Expression)
					}
				}
			}
			if strings.Contains(file.Text(), "import(") || strings.Contains(file.Text(), "require(") {
				var visit func(*ast.Node) bool
				visit = func(n *ast.Node) bool {
					if n.Kind == ast.KindCallExpression && (ast.IsImportCall(n) || ast.IsRequireCall(n, false)) {
						args := n.AsCallExpression().Arguments
						if args != nil && len(args.Nodes) > 0 && ast.IsStringLiteralLike(args.Nodes[0]) {
							add(args.Nodes[0])
						} else {
							computed = true
						}
					}
					n.ForEachChild(visit)
					return false
				}
				file.AsNode().ForEachChild(visit)
			}
		}
		out.yes(computed)
		out.number(uint64(len(edges)))
		for _, edge := range edges {
			out.text(edge)
		}
	}
	return out.String(), nil
}
