package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

type streamImport struct {
	kind      string
	typeOnly  bool
	specifier *ast.Node
}

// Source identities, syntactic import records, and compiler module-resolution answers.
// The native rule decides which edges count, closure, entry status, and blocking order.
func (p *Program) streamProgram(out *fields, node *ast.Node) (string, error) {
	if node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("stream-program requires a SourceFile")
	}
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
		var records []streamImport
		if !file.IsDeclarationFile {
			for _, stmt := range file.Statements.Nodes {
				switch stmt.Kind {
				case ast.KindImportDeclaration:
					decl := stmt.AsImportDeclaration()
					records = append(records, streamImport{"ImportDeclaration", decl.ImportClause != nil && decl.ImportClause.IsTypeOnly(), decl.ModuleSpecifier})
				case ast.KindExportDeclaration:
					decl := stmt.AsExportDeclaration()
					records = append(records, streamImport{"ExportDeclaration", decl.IsTypeOnly, decl.ModuleSpecifier})
				case ast.KindImportEqualsDeclaration:
					decl := stmt.AsImportEqualsDeclaration()
					if decl.ModuleReference != nil && ast.IsExternalModuleReference(decl.ModuleReference) {
						records = append(records, streamImport{"ImportEqualsDeclaration", decl.IsTypeOnly, decl.ModuleReference.AsExternalModuleReference().Expression})
					}
				}
			}
			var visit func(*ast.Node) bool
			visit = func(at *ast.Node) bool {
				if at.Kind == ast.KindCallExpression && (ast.IsImportCall(at) || ast.IsRequireCall(at, false)) {
					var specifier *ast.Node
					if arguments := at.AsCallExpression().Arguments; arguments != nil && len(arguments.Nodes) > 0 {
						specifier = arguments.Nodes[0]
					}
					records = append(records, streamImport{"CallExpression", false, specifier})
				}
				at.ForEachChild(visit)
				return false
			}
			file.AsNode().ForEachChild(visit)
		}
		out.number(uint64(len(records)))
		for _, record := range records {
			out.text(record.kind)
			out.yes(record.typeOnly)
			literal := record.specifier != nil && ast.IsStringLiteralLike(record.specifier)
			out.yes(literal)
			target := ""
			value := ""
			if literal {
				value = record.specifier.Text()
				resolved := p.Compiler.GetResolvedModuleFromModuleSpecifier(file, record.specifier)
				if resolved != nil && resolved.IsResolved() {
					if next := p.Compiler.GetSourceFileForResolvedModule(resolved); next != nil {
						target = next.FileName().AsString()
					}
				}
			}
			out.text(value)
			out.text(target)
		}
	}
	return out.String(), nil
}
