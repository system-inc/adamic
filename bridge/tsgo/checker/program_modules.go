package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

type moduleFact struct {
	kind      string
	typeOnly  bool
	specifier *ast.Node
}

// programModules exposes source files and raw static/dynamic module references, not import closure judgments.
func (p *Program) programModules(out *fields, question string) (string, error) {
	if question != "program-modules" {
		return "", fmt.Errorf("unexpected module question suffix")
	}
	files := p.Compiler.SourceFiles()
	out.number(uint64(len(files)))
	for _, file := range files {
		out.text(file.FileName())
		out.yes(file.IsDeclarationFile)
		out.yes(ast.IsExternalModule(file))
		out.text(file.Text())
		var refs []moduleFact
		for _, stmt := range file.Statements.Nodes {
			switch stmt.Kind {
			case ast.KindImportDeclaration:
				d := stmt.AsImportDeclaration()
				refs = append(refs, moduleFact{"import", d.ImportClause != nil && d.ImportClause.IsTypeOnly(), d.ModuleSpecifier})
			case ast.KindExportDeclaration:
				d := stmt.AsExportDeclaration()
				refs = append(refs, moduleFact{"export", d.IsTypeOnly, d.ModuleSpecifier})
			case ast.KindImportEqualsDeclaration:
				d := stmt.AsImportEqualsDeclaration()
				if d.ModuleReference != nil && ast.IsExternalModuleReference(d.ModuleReference) {
					refs = append(refs, moduleFact{"import-equals", d.IsTypeOnly, d.ModuleReference.AsExternalModuleReference().Expression})
				}
			}
		}
		var visit func(*ast.Node) bool
		visit = func(n *ast.Node) bool {
			if n.Kind == ast.KindCallExpression && (ast.IsImportCall(n) || ast.IsRequireCall(n, false)) {
				var specifier *ast.Node
				if args := n.AsCallExpression().Arguments; args != nil && len(args.Nodes) > 0 {
					specifier = args.Nodes[0]
				}
				kind := "require-call"
				if ast.IsImportCall(n) {
					kind = "import-call"
				}
				refs = append(refs, moduleFact{kind, false, specifier})
			}
			n.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
		out.number(uint64(len(refs)))
		for _, ref := range refs {
			out.text(ref.kind)
			out.yes(ref.typeOnly)
			literal := ref.specifier != nil && ast.IsStringLiteralLike(ref.specifier)
			out.yes(literal)
			target := ""
			if literal {
				resolved := p.Compiler.GetResolvedModuleFromModuleSpecifier(file, ref.specifier)
				if resolved != nil && resolved.IsResolved() {
					source := p.Compiler.GetSourceFileForResolvedModule(resolved.ResolvedFileName)
					if source != nil {
						target = source.FileName()
					}
				}
			}
			out.text(target)
		}
	}
	return out.String(), nil
}
