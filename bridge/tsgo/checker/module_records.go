package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

// moduleRecords exposes source metadata and syntactic import records with resolutions.
func (p *Program) moduleRecords(out *fields, node *ast.Node, question string) (string, error) {
	if question != "module-records" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("module-records requires SourceFile")
	}
	files := p.Compiler.GetSourceFiles()
	out.number(uint64(len(files)))
	for _, source := range files {
		out.text(source.FileName())
		out.yes(source.IsDeclarationFile)
		out.text(source.Text())
		type record struct {
			kind      string
			typeOnly  bool
			specifier *ast.Node
		}
		var records []record
		for _, statement := range source.Statements.Nodes {
			switch statement.Kind {
			case ast.KindImportDeclaration:
				d := statement.AsImportDeclaration()
				records = append(records, record{"import", d.ImportClause != nil && d.ImportClause.IsTypeOnly(), d.ModuleSpecifier})
			case ast.KindExportDeclaration:
				d := statement.AsExportDeclaration()
				records = append(records, record{"export", d.IsTypeOnly, d.ModuleSpecifier})
			case ast.KindImportEqualsDeclaration:
				d := statement.AsImportEqualsDeclaration()
				if d.ModuleReference != nil && ast.IsExternalModuleReference(d.ModuleReference) {
					records = append(records, record{"import-equals", d.IsTypeOnly, d.ModuleReference.AsExternalModuleReference().Expression})
				}
			}
		}
		if strings.Contains(source.Text(), "import(") || strings.Contains(source.Text(), "require(") {
			var visit func(*ast.Node) bool
			visit = func(current *ast.Node) bool {
				if current.Kind == ast.KindCallExpression && (ast.IsImportCall(current) || ast.IsRequireCall(current, false)) {
					var specifier *ast.Node
					if list := current.AsCallExpression().Arguments; list != nil && len(list.Nodes) > 0 {
						specifier = list.Nodes[0]
					}
					records = append(records, record{"call", false, specifier})
				}
				current.ForEachChild(visit)
				return false
			}
			source.AsNode().ForEachChild(visit)
		}
		out.number(uint64(len(records)))
		for _, r := range records {
			out.text(r.kind)
			out.yes(r.typeOnly)
			literal := r.specifier != nil && ast.IsStringLiteralLike(r.specifier)
			out.yes(literal)
			target := ""
			if literal {
				if resolved := p.Compiler.GetResolvedModuleFromModuleSpecifier(source, r.specifier); resolved != nil && resolved.IsResolved() {
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
