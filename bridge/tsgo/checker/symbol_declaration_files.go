package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func init() { questionExtensions["symbol-declaration-files"] = symbolDeclarationFilesQuestion }

// Preserve binder declaration order and alias symbols; this question supplies no lint verdict.
func symbolDeclarationFilesQuestion(p *Program, out *fields, c *checker.Checker, source *ast.SourceFile, node *ast.Node, question string) (string, error) {
	if question != "symbol-declaration-files" || node.Kind != ast.KindIdentifier {
		return "", fmt.Errorf("symbol-declaration-files requires an identifier")
	}
	symbol := c.GetSymbolAtLocation(node)
	if symbol == nil {
		out.number(0)
		return out.String(), nil
	}
	out.number(uint64(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil {
			return "", fmt.Errorf("symbol declaration has no source file")
		}
		out.text(file.FileName())
		out.yes(file.IsDeclarationFile)
	}
	return out.String(), nil
}
