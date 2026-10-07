package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// bindingOrigin preserves declaration order and both ordinary and shorthand symbols.
func (p *Program) bindingOrigin(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "binding-origin" || node.Kind != ast.KindIdentifier {
		return "", fmt.Errorf("binding-origin requires an Identifier")
	}
	plain := c.GetSymbolAtLocation(node)
	var shorthand *ast.Symbol
	if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
		shorthand = c.GetShorthandAssignmentValueSymbol(node.Parent)
	}
	for _, symbol := range []*ast.Symbol{plain, shorthand} {
		out.yes(symbol != nil)
		var declarations []*ast.Node
		if symbol != nil {
			declarations = symbol.Declarations
		}
		out.number(uint64(len(declarations)))
		for _, d := range declarations {
			f := ast.GetSourceFileOfNode(d)
			if f == nil {
				return "", fmt.Errorf("declaration has no source")
			}
			out.text(f.FileName())
			out.yes(f.IsDeclarationFile)
			out.text(strings.TrimPrefix(d.Kind.String(), "Kind"))
			out.number(uint64(d.Pos()))
			out.number(uint64(d.End()))
		}
	}
	return out.String(), nil
}
