package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// preferenceBinding exposes ordinary and read symbols with declaration-file metadata.
func (p *Program) preferenceBinding(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "preference-binding" || node.Kind != ast.KindIdentifier {
		return "", fmt.Errorf("preference-binding requires an Identifier")
	}
	plain := c.GetSymbolAtLocation(node)
	read := plain
	if node.Parent != nil {
		if node.Parent.Kind == ast.KindShorthandPropertyAssignment && node.Parent.Name() == node {
			read = c.GetShorthandAssignmentValueSymbol(node.Parent)
		}
		if node.Parent.Kind == ast.KindExportSpecifier {
			read = c.GetExportSpecifierLocalTargetSymbol(node.Parent)
		}
	}
	for _, symbol := range []*ast.Symbol{plain, read} {
		out.number(p.symbolID(symbol))
		var declarations []*ast.Node
		if symbol != nil {
			declarations = symbol.Declarations
		}
		out.number(uint64(len(declarations)))
		for _, declaration := range declarations {
			source := ast.GetSourceFileOfNode(declaration)
			out.yes(source != nil && source.IsDeclarationFile)
		}
	}
	return out.String(), nil
}
