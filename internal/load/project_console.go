package load

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// A host may supply console through libs, selected types, automatic type discovery, or explicit
// declaration roots. Names inside a module/namespace alone do not supply a global console.
func hasHostConsole(files []*ast.SourceFile) bool {
	globals := func(statements []*ast.Node) bool {
		for _, statement := range statements {
			if statement.Kind != ast.KindVariableStatement {
				continue
			}
			for _, declaration := range statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
				if ast.IsIdentifier(declaration.Name()) && declaration.Name().Text() == "console" {
					return true
				}
			}
		}
		return false
	}
	for _, file := range files {
		if IsPrelude(file) || !file.IsDeclarationFile {
			continue
		}
		if !ast.IsExternalModule(file) && globals(file.Statements.Nodes) {
			return true
		}
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			if ast.IsGlobalScopeAugmentation(node) {
				body := node.AsModuleDeclaration().Body
				if body != nil && body.Kind == ast.KindModuleBlock && globals(body.AsModuleBlock().Statements.Nodes) {
					return true
				}
			}
			return node.ForEachChild(visit)
		}
		if visit(file.AsNode()) {
			return true
		}
	}
	return false
}
