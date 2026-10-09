package rule

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// DeclarationsIn returns the declarations of symbol that sit in sourceFile, in the checker's order.
//
// A symbol's declarations can live in any file the checker reaches, and a rule keyed on the shapes of
// its imports must not read into another file's bodies (TypeReach). Most rules that ask for a
// declaration only want the one their own file holds: ESLint's core rules see a single file, and its
// scope manager knows no declaration outside it. Reading through this says so in a way the type-reach
// guard can check, since it trusts this function's result to be the rule's own file (#9bjjk4a).
//
// The common case, every declaration in the file, returns the checker's own slice without copying.
func DeclarationsIn(sourceFile *ast.SourceFile, symbol *ast.Symbol) []*ast.Node {
	if symbol == nil || sourceFile == nil {
		return nil
	}
	for index, declaration := range symbol.Declarations {
		if ast.GetSourceFileOfNode(declaration) == sourceFile {
			continue
		}
		own := make([]*ast.Node, 0, len(symbol.Declarations)-1)
		own = append(own, symbol.Declarations[:index]...)
		for _, later := range symbol.Declarations[index+1:] {
			if ast.GetSourceFileOfNode(later) == sourceFile {
				own = append(own, later)
			}
		}
		return own
	}
	return symbol.Declarations
}

// IsDeclaredOnlyInDeclarationFiles reports whether symbol has declarations and every one is in a
// declaration file: a global from a lib or an installed package, never a binding a source file
// declares. It answers about files and hands back no node, so a rule can ask it without reading another
// file's syntax, which the type-reach guard trusts.
func IsDeclaredOnlyInDeclarationFiles(symbol *ast.Symbol) bool {
	return symbol != nil && len(symbol.Declarations) > 0 && !IsDeclaredInASourceFile(symbol)
}

// IsDeclaredInASourceFile reports whether any declaration of symbol is in a source file rather than a
// declaration file: something this program declares, where a symbol the checker made with no declaration
// at all, like `undefined`, is not. It answers about files and hands back no node, which the type-reach
// guard trusts.
func IsDeclaredInASourceFile(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		sourceFile := ast.GetSourceFileOfNode(declaration)
		if sourceFile == nil || !sourceFile.IsDeclarationFile {
			return true
		}
	}
	return false
}
