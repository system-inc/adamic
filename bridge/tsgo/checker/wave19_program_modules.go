package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

type wave19ModuleSpecifier struct {
	kind     string
	typeOnly bool
	node     *ast.Node
}

// wave19ProgramModules serializes raw program files and resolved import specifiers.
// Native Adamic must compute entry points, closures and any lint decisions.
func (p *Program) wave19ProgramModules(out *fields, node *ast.Node, question string) (string, error) {
	if question != "wave19-program-modules" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("wave19-program-modules requires a SourceFile and no suffix")
	}
	files := p.Compiler.SourceFiles()
	out.number(uint64(len(files)))
	for _, source := range files {
		out.text(source.FileName())
		out.yes(source.IsDeclarationFile)
		out.yes(ast.IsExternalModule(source))
		out.yes(p.Compiler.IsSourceFileDefaultLibrary(source.Path()))
		var specifiers []wave19ModuleSpecifier
		if !source.IsDeclarationFile {
			var walk func(*ast.Node)
			walk = func(current *ast.Node) {
				switch current.Kind {
				case ast.KindImportDeclaration:
					declaration := current.AsImportDeclaration()
					specifiers = append(specifiers, wave19ModuleSpecifier{"ImportDeclaration", declaration.ImportClause != nil && declaration.ImportClause.IsTypeOnly(), declaration.ModuleSpecifier})
				case ast.KindExportDeclaration:
					declaration := current.AsExportDeclaration()
					if declaration.ModuleSpecifier != nil {
						specifiers = append(specifiers, wave19ModuleSpecifier{"ExportDeclaration", declaration.IsTypeOnly, declaration.ModuleSpecifier})
					}
				case ast.KindImportEqualsDeclaration:
					declaration := current.AsImportEqualsDeclaration()
					if declaration.ModuleReference != nil && ast.IsExternalModuleReference(declaration.ModuleReference) {
						specifiers = append(specifiers, wave19ModuleSpecifier{"ImportEqualsDeclaration", declaration.IsTypeOnly, declaration.ModuleReference.AsExternalModuleReference().Expression})
					}
				case ast.KindCallExpression:
					if ast.IsImportCall(current) || ast.IsRequireCall(current, false) {
						kind := "require"
						if ast.IsImportCall(current) {
							kind = "import"
						}
						var argument *ast.Node
						call := current.AsCallExpression()
						if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
							argument = call.Arguments.Nodes[0]
						}
						specifiers = append(specifiers, wave19ModuleSpecifier{kind, false, argument})
					}
				}
				current.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
			}
			walk(source.AsNode())
		}
		out.number(uint64(len(specifiers)))
		for _, specifier := range specifiers {
			out.text(specifier.kind)
			out.yes(specifier.typeOnly)
			kind, text, resolved := "", "", ""
			if specifier.node != nil {
				kind = strings.TrimPrefix(specifier.node.Kind.String(), "Kind")
				if ast.IsStringLiteralLike(specifier.node) {
					text = specifier.node.Text()
					module := p.Compiler.GetResolvedModuleFromModuleSpecifier(source, specifier.node)
					if module != nil && module.IsResolved() {
						if target := p.Compiler.GetSourceFileForResolvedModule(module.ResolvedFileName); target != nil {
							resolved = target.FileName()
						}
					}
				}
			}
			out.text(kind)
			out.text(text)
			out.text(resolved)
		}
	}
	return out.String(), nil
}
