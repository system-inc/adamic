package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// ReadSymbol supplies the raw binding read by shorthand and local export sites.
func (p *Program) readSymbol(c *checker.Checker, node *ast.Node, question string) (string, error) {
	const mode = "read-symbol"
	if question != mode || node.Kind != ast.KindIdentifier {
		return "", fmt.Errorf("read-symbol requires an Identifier without suffix")
	}
	symbol := c.GetSymbolAtLocation(node)
	if parent := node.Parent; parent != nil {
		if parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == node {
			symbol = c.GetShorthandAssignmentValueSymbol(parent)
		} else if parent.Kind == ast.KindExportSpecifier {
			symbol = c.GetExportSpecifierLocalTargetSymbol(parent)
		}
	}
	out := &fields{}
	out.number(1)
	out.text(mode)
	p.writePlatformSymbol(out, symbol)
	return out.String(), nil
}
