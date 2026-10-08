package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// globalBinding supplies declaration-file origins, including the local merged
// symbol and shorthand value side. Native rules choose the provenance policy.
func (p *Program) globalBinding(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if node.Kind != ast.KindIdentifier || (question != "global-binding\nnode" && question != "global-binding\nvalue") {
		return "", fmt.Errorf("global-binding requires Identifier and node/value mode")
	}
	symbol := c.GetSymbolAtLocation(node)
	if question == "global-binding\nvalue" && node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
		symbol = c.GetShorthandAssignmentValueSymbol(node.Parent)
	}
	write := func(symbol *ast.Symbol) {
		out.yes(symbol != nil)
		var declarations []*ast.Node
		if symbol != nil {
			declarations = symbol.Declarations
		}
		out.number(uint64(len(declarations)))
		for _, declaration := range declarations {
			source := ast.GetSourceFileOfNode(declaration)
			out.yes(source != nil)
			out.yes(source != nil && source.IsDeclarationFile)
		}
	}
	write(symbol)
	write(node.LocalSymbol())
	return out.String(), nil
}
