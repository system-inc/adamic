package wave08next

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"strconv"
)

// ProgramLoadsFields supplies program membership and all static/import-equals/
// dynamic-import/require literal resolutions. Edge selection remains native.
// Registration is deliberately pending; shared dispatcher files are not edited.
func ProgramLoadsFields(program *compiler.Program, node *ast.Node) ([]string, error) {
	if node.Kind != ast.KindSourceFile {
		return nil, fmt.Errorf("program loads requires SourceFile")
	}
	flag := func(v bool) string {
		if v {
			return "1"
		}
		return "0"
	}
	files := program.SourceFiles()
	out := []string{"1", "wave08-program-loads", strconv.Itoa(len(files))}
	for _, file := range files {
		out = append(out, file.FileName(), flag(file.IsDeclarationFile), flag(ast.IsExternalModule(file)), file.Text())
		var specifiers []*ast.Node
		roles := map[*ast.Node]string{}
		typeOnly := map[*ast.Node]bool{}
		for _, statement := range file.Statements.Nodes {
			switch statement.Kind {
			case ast.KindImportDeclaration:
				specifier := statement.AsImportDeclaration().ModuleSpecifier
				specifiers = append(specifiers, specifier)
				roles[specifier] = "ImportDeclaration"
				clause := statement.AsImportDeclaration().ImportClause
				typeOnly[specifier] = clause != nil && clause.IsTypeOnly()
			case ast.KindExportDeclaration:
				if specifier := statement.AsExportDeclaration().ModuleSpecifier; specifier != nil {
					specifiers = append(specifiers, specifier)
					roles[specifier] = "ExportDeclaration"
					typeOnly[specifier] = statement.AsExportDeclaration().IsTypeOnly
				}
			case ast.KindImportEqualsDeclaration:
				reference := statement.AsImportEqualsDeclaration().ModuleReference
				if reference != nil && ast.IsExternalModuleReference(reference) {
					specifier := reference.AsExternalModuleReference().Expression
					specifiers = append(specifiers, specifier)
					roles[specifier] = "ImportEqualsDeclaration"
					typeOnly[specifier] = statement.AsImportEqualsDeclaration().IsTypeOnly
				}
			}
		}
		var visit func(*ast.Node) bool
		visit = func(child *ast.Node) bool {
			if child.Kind == ast.KindCallExpression && (ast.IsImportCall(child) || ast.IsRequireCall(child, false)) {
				args := child.AsCallExpression().Arguments
				if args != nil && len(args.Nodes) > 0 {
					specifiers = append(specifiers, args.Nodes[0])
					roles[args.Nodes[0]] = "CallExpression"
				} else {
					specifiers = append(specifiers, child)
					roles[child] = "CallExpression"
				}
			}
			child.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
		out = append(out, strconv.Itoa(len(specifiers)))
		for _, specifier := range specifiers {
			target := ""
			if ast.IsStringLiteralLike(specifier) {
				if resolved := program.GetResolvedModuleFromModuleSpecifier(file, specifier); resolved != nil && resolved.IsResolved() {
					if found := program.GetSourceFileForResolvedModule(resolved.ResolvedFileName); found != nil {
						target = found.FileName()
					}
				}
			}
			out = append(out, strconv.Itoa(specifier.Pos()), strconv.Itoa(specifier.End()), target, roles[specifier], flag(typeOnly[specifier]), flag(ast.IsStringLiteralLike(specifier)))
		}
	}
	return out, nil
}
