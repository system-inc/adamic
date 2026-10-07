// All module-load syntax and compiler resolutions. Closure decisions belong to the consumer.
package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func (p *Program) programImports(node *ast.Node, question string) (string, error) {
	if question != "program-imports" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("program-imports requires a SourceFile")
	}
	out := &fields{}
	out.number(1)
	out.text(question)
	files := p.Compiler.SourceFiles()
	out.number(uint64(len(files)))
	for _, source := range files {
		out.text(source.FileName())
		out.yes(source.IsDeclarationFile)
		text := ""
		if !source.IsDeclarationFile {
			text = source.Text()
		}
		out.text(text)
		type entry struct {
			kind      string
			typed     bool
			specifier *ast.Node
		}
		var entries []entry
		for _, statement := range source.Statements.Nodes {
			switch statement.Kind {
			case ast.KindImportDeclaration:
				d := statement.AsImportDeclaration()
				entries = append(entries, entry{nodeKind(statement), d.ImportClause != nil && d.ImportClause.IsTypeOnly(), d.ModuleSpecifier})
			case ast.KindExportDeclaration:
				d := statement.AsExportDeclaration()
				entries = append(entries, entry{nodeKind(statement), d.IsTypeOnly, d.ModuleSpecifier})
			case ast.KindImportEqualsDeclaration:
				d := statement.AsImportEqualsDeclaration()
				var specifier *ast.Node
				if d.ModuleReference != nil && ast.IsExternalModuleReference(d.ModuleReference) {
					specifier = d.ModuleReference.AsExternalModuleReference().Expression
				}
				entries = append(entries, entry{nodeKind(statement), d.IsTypeOnly, specifier})
			}
		}
		var visit func(*ast.Node) bool
		visit = func(n *ast.Node) bool {
			if n.Kind == ast.KindCallExpression && (ast.IsImportCall(n) || ast.IsRequireCall(n, false)) {
				var specifier *ast.Node
				args := n.AsCallExpression().Arguments
				if args != nil && len(args.Nodes) > 0 {
					specifier = args.Nodes[0]
				}
				entries = append(entries, entry{"CallExpression", false, specifier})
			}
			n.ForEachChild(visit)
			return false
		}
		source.AsNode().ForEachChild(visit)
		out.number(uint64(len(entries)))
		for _, e := range entries {
			out.text(e.kind)
			out.yes(e.typed)
			literal := e.specifier != nil && ast.IsStringLiteralLike(e.specifier)
			out.yes(literal)
			target := ""
			if literal {
				resolved := p.Compiler.GetResolvedModuleFromModuleSpecifier(source, e.specifier)
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
