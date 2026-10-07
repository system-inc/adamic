package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// namespaceBinding exposes the unresolved binding's own declarations.
func (p *Program) namespaceBinding(c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "namespace-binding" || node.Kind != ast.KindIdentifier {
		return "", fmt.Errorf("namespace-binding requires an Identifier")
	}
	out := &fields{}
	out.number(1)
	out.text(question)
	symbol := c.GetSymbolAtLocation(node)
	out.yes(symbol != nil)
	if symbol != nil {
		out.number(uint64(symbol.Flags))
		out.number(uint64(len(symbol.Declarations)))
		for _, d := range symbol.Declarations {
			out.text(strings.TrimPrefix(d.Kind.String(), "Kind"))
			source := ast.GetSourceFileOfNode(d)
			out.yes(source == ast.GetSourceFileOfNode(node))
			out.yes(ast.IsTypeOnlyImportOrExportDeclaration(d))
		}
	}
	return out.String(), nil
}
