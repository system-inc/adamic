package imports

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Declarations lists every import declaration in a file, in source order: the ones at the top of the
// file and the ones inside an ambient module body, as in `declare module '*.vue' { import { X } from
// 'vue' }`.
//
// ESLint visits ImportDeclaration wherever the tree holds one, so a rule that walks only the file's
// top-level statements misses the second kind and looks complete. That gap shipped once, in
// `@typescript-eslint/consistent-type-imports` against TanStack's Vue shims (#7g5r6vt).
//
// A dotted namespace (`declare module A.B { }`) nests one module declaration's body inside another,
// so the walk follows the chain down to the block that holds the statements.
func Declarations(sourceFile *ast.SourceFile) []*ast.Node {
	if sourceFile == nil {
		return nil
	}
	return appendDeclarations(nil, sourceFile.Statements)
}

func appendDeclarations(declarations []*ast.Node, statements *ast.NodeList) []*ast.Node {
	if statements == nil {
		return declarations
	}
	for _, statement := range statements.Nodes {
		switch statement.Kind {
		case ast.KindImportDeclaration:
			declarations = append(declarations, statement)
		case ast.KindModuleDeclaration:
			body := statement.Body()
			for body != nil && body.Kind == ast.KindModuleDeclaration {
				body = body.Body()
			}
			if body != nil && body.Kind == ast.KindModuleBlock {
				declarations = appendDeclarations(declarations, body.AsModuleBlock().Statements)
			}
		}
	}
	return declarations
}
