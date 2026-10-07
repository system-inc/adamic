package checker

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Ordinary, local merged, and shorthand value symbols with raw declarations.
func (p *Program) inspectGlobalBindingFacts(c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "global-binding-facts" {
		return p.inspectTypeDeclarationAncestry(c, node, question)
	}
	if node.Kind != ast.KindIdentifier {
		return "", fmt.Errorf("global-binding-facts requires an Identifier")
	}
	out := &fields{}
	out.number(1)
	out.text(question)
	p.writeSymbolDetails(out, c.GetSymbolAtLocation(node))
	p.writeSymbolDetails(out, node.LocalSymbol())
	var shorthand *ast.Symbol
	if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
		shorthand = c.GetShorthandAssignmentValueSymbol(node.Parent)
	}
	p.writeSymbolDetails(out, shorthand)
	return out.String(), nil
}
