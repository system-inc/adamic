package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

// programImports supplies source filenames, resolved import targets and clause
// syntax. Deciding whether an export is used remains an Adamic judgment.
func (p *Program) programImports(out *fields, node *ast.Node, question string) (string, error) {
	if question != "program-imports" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("program-imports requires a source file")
	}
	files := p.Compiler.SourceFiles()
	out.number(uint64(len(files)))
	for _, file := range files {
		out.text(file.FileName())
		out.yes(file.IsDeclarationFile)
		var prefixes []string
		var walk func(*ast.Node)
		walk = func(n *ast.Node) {
			if n.Kind == ast.KindCallExpression {
				call := n.AsCallExpression()
				if call.Expression.Kind == ast.KindImportKeyword && call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
					argument := call.Arguments.Nodes[0]
					if argument.Kind == ast.KindTemplateExpression {
						prefixes = append(prefixes, argument.AsTemplateExpression().Head.Text())
					}
				}
			}
			n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
		}
		if !file.IsDeclarationFile && !strings.Contains(file.FileName(), "/node_modules/") {
			walk(file.AsNode())
		}
		out.number(uint64(len(prefixes)))
		for _, prefix := range prefixes {
			out.text(prefix)
		}
		specifiers := file.Imports()
		out.number(uint64(len(specifiers)))
		for _, specifier := range specifiers {
			target := ""
			if specifier != nil && ast.IsStringLiteralLike(specifier) {
				resolved := p.Compiler.GetResolvedModuleFromModuleSpecifier(file, specifier)
				if resolved != nil && resolved.IsResolved() {
					source := p.Compiler.GetSourceFileForResolvedModule(resolved.ResolvedFileName)
					if source != nil {
						target = source.FileName()
					}
				}
			}
			out.text(target)
			kind := ""
			hasDefault, namespace := false, false
			var names []string
			if specifier != nil && specifier.Parent != nil {
				parent := specifier.Parent
				kind = strings.TrimPrefix(parent.Kind.String(), "Kind")
				switch parent.Kind {
				case ast.KindImportDeclaration:
					clause := parent.AsImportDeclaration().ImportClause
					if clause != nil {
						imported := clause.AsImportClause()
						hasDefault = imported.Name() != nil
						bindings := imported.NamedBindings
						if bindings != nil {
							namespace = bindings.Kind == ast.KindNamespaceImport
							if bindings.Kind == ast.KindNamedImports {
								for _, element := range bindings.AsNamedImports().Elements.Nodes {
									name := element.AsImportSpecifier().PropertyName
									if name == nil {
										name = element.Name()
									}
									if name != nil {
										names = append(names, name.Text())
									}
								}
							}
						}
					}
				case ast.KindExportDeclaration:
					clause := parent.AsExportDeclaration().ExportClause
					namespace = clause == nil || clause.Kind != ast.KindNamedExports
					if clause != nil && clause.Kind == ast.KindNamedExports {
						for _, element := range clause.AsNamedExports().Elements.Nodes {
							name := element.AsExportSpecifier().PropertyName
							if name == nil {
								name = element.Name()
							}
							if name != nil {
								names = append(names, name.Text())
							}
						}
					}
				}
			}
			out.text(kind)
			out.yes(hasDefault)
			out.yes(namespace)
			out.number(uint64(len(names)))
			for _, name := range names {
				out.text(name)
			}
		}
	}
	return out.String(), nil
}
