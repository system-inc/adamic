package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

// wave04NextProgramModules exposes module syntax and resolved literal targets.
// The native rule computes imported files and the blocking reachability closure.
func (p *Program) wave04NextProgramModules(out *fields, node *ast.Node, question string) (string, error) {
	if question != "wave04-next-program-modules" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("program modules requires a source file")
	}
	files := p.Compiler.SourceFiles()
	out.number(uint64(len(files)))
	for _, file := range files {
		out.text(file.FileName())
		out.yes(file.IsDeclarationFile)
		text := ""
		if !file.IsDeclarationFile {
			text = file.Text()
		}
		out.text(text)
		type item struct {
			kind                   string
			typeOnly               bool
			argument               *ast.Node
			calleeKind, calleeName string
			arguments              int
		}
		var items []item
		if !file.IsDeclarationFile {
			for _, statement := range file.Statements.Nodes {
				switch statement.Kind {
				case ast.KindImportDeclaration:
					declaration := statement.AsImportDeclaration()
					only := declaration.ImportClause != nil && declaration.ImportClause.IsTypeOnly()
					items = append(items, item{kind: "ImportDeclaration", typeOnly: only, argument: declaration.ModuleSpecifier})
				case ast.KindExportDeclaration:
					declaration := statement.AsExportDeclaration()
					items = append(items, item{kind: "ExportDeclaration", typeOnly: declaration.IsTypeOnly, argument: declaration.ModuleSpecifier})
				case ast.KindImportEqualsDeclaration:
					declaration := statement.AsImportEqualsDeclaration()
					var argument *ast.Node
					if declaration.ModuleReference != nil && ast.IsExternalModuleReference(declaration.ModuleReference) {
						argument = declaration.ModuleReference.AsExternalModuleReference().Expression
					}
					items = append(items, item{kind: "ImportEqualsDeclaration", typeOnly: declaration.IsTypeOnly, argument: argument})
				}
			}
			var walk func(*ast.Node)
			walk = func(current *ast.Node) {
				if current.Kind == ast.KindCallExpression {
					call := current.AsCallExpression()
					callee := call.Expression
					if callee.Kind == ast.KindImportKeyword || (callee.Kind == ast.KindIdentifier && callee.Text() == "require") {
						var argument *ast.Node
						count := 0
						if call.Arguments != nil {
							count = len(call.Arguments.Nodes)
							if count > 0 {
								argument = call.Arguments.Nodes[0]
							}
						}
						name := ""
						if callee.Kind == ast.KindIdentifier {
							name = callee.Text()
						}
						items = append(items, item{kind: "CallExpression", argument: argument, calleeKind: strings.TrimPrefix(callee.Kind.String(), "Kind"), calleeName: name, arguments: count})
					}
				}
				current.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
			}
			walk(file.AsNode())
		}
		out.number(uint64(len(items)))
		for _, entry := range items {
			out.text(entry.kind)
			out.yes(entry.typeOnly)
			out.text(entry.calleeKind)
			out.text(entry.calleeName)
			out.number(uint64(entry.arguments))
			kind, text, target := "", "", ""
			if entry.argument != nil {
				kind = strings.TrimPrefix(entry.argument.Kind.String(), "Kind")
				if ast.IsStringLiteralLike(entry.argument) {
					text = entry.argument.Text()
					resolved := p.Compiler.GetResolvedModuleFromModuleSpecifier(file, entry.argument)
					if resolved != nil && resolved.IsResolved() {
						source := p.Compiler.GetSourceFileForResolvedModule(resolved.ResolvedFileName)
						if source != nil {
							target = source.FileName()
						}
					}
				}
			}
			out.text(kind)
			out.text(text)
			out.text(target)
		}
	}
	return out.String(), nil
}
