package checker

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// referenceNode exposes shelf syntax flags and the symbol a name reads. It does
// not trace aliases, classify globals, or make any rule decision.
func (p *Program) referenceNode(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "reference-node" {
		return "", fmt.Errorf("unexpected reference-node suffix")
	}
	out.yes(ast.IsDeclarationName(node))
	out.yes(ast.IsWriteAccess(node))
	symbol := c.GetSymbolAtLocation(node)
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == node {
		symbol = c.GetShorthandAssignmentValueSymbol(parent)
	} else if parent != nil && parent.Kind == ast.KindExportSpecifier {
		symbol = c.GetExportSpecifierLocalTargetSymbol(parent)
	}
	out.yes(symbol != nil)
	if symbol != nil {
		out.number(p.symbolID(symbol))
		out.number(uint64(len(symbol.Declarations)))
		for _, d := range symbol.Declarations {
			source := ast.GetSourceFileOfNode(d)
			out.text(source.FileName())
			out.yes(source.IsDeclarationFile)
		}
	}
	return out.String(), nil
}
