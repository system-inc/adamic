package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// programModuleEdges returns resolved import edges and computed-load locations.
// It deliberately does not compute entry status or the blocking closure.
func (p *Program) programModuleEdges(out *fields, _ *checker.Checker, node *ast.Node, question string) error {
	if question != "program-module-edges" || node.Kind != ast.KindSourceFile {
		return fmt.Errorf("program-module-edges requires a source file")
	}
	sources := p.Compiler.SourceFiles()
	out.number(uint64(len(sources)))
	for _, source := range sources {
		out.text(string(source.FileName()))
		out.yes(source.IsDeclarationFile)
		var edges []string
		computed := false
		add := func(specifier *ast.Node) {
			if specifier == nil || !ast.IsStringLiteralLike(specifier) {
				return
			}
			resolved := p.Compiler.GetResolvedModuleFromModuleSpecifier(source, specifier)
			if resolved == nil || !resolved.IsResolved() {
				return
			}
			target := p.Compiler.GetSourceFileForResolvedModule(resolved)
			if target != nil && target != source {
				edges = append(edges, string(target.FileName()))
			}
		}
		if !source.IsDeclarationFile {
			for _, statement := range source.Statements.Nodes {
				switch statement.Kind {
				case ast.KindImportDeclaration:
					decl := statement.AsImportDeclaration()
					if decl.ImportClause == nil || !decl.ImportClause.IsTypeOnly() {
						add(decl.ModuleSpecifier)
					}
				case ast.KindExportDeclaration:
					decl := statement.AsExportDeclaration()
					if !decl.IsTypeOnly {
						add(decl.ModuleSpecifier)
					}
				case ast.KindImportEqualsDeclaration:
					decl := statement.AsImportEqualsDeclaration()
					if !decl.IsTypeOnly && decl.ModuleReference != nil && ast.IsExternalModuleReference(decl.ModuleReference) {
						add(decl.ModuleReference.AsExternalModuleReference().Expression)
					}
				}
			}
			if strings.Contains(source.Text(), "import(") || strings.Contains(source.Text(), "require(") {
				var visit func(*ast.Node) bool
				visit = func(child *ast.Node) bool {
					if child.Kind == ast.KindCallExpression && (ast.IsImportCall(child) || ast.IsRequireCall(child, false)) {
						args := child.AsCallExpression().Arguments
						if args != nil && len(args.Nodes) > 0 && ast.IsStringLiteralLike(args.Nodes[0]) {
							add(args.Nodes[0])
						} else {
							computed = true
						}
					}
					child.ForEachChild(visit)
					return false
				}
				source.AsNode().ForEachChild(visit)
			}
		}
		out.yes(computed)
		out.number(uint64(len(edges)))
		for _, edge := range edges {
			out.text(edge)
		}
	}
	return nil
}
