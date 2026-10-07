package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Raw accessor identity and declaration count, without shorthand-value resolution.
func (p *Program) rawSymbolDeclarationCount(out *fields, c *checker.Checker, node *ast.Node, question string) error {
	if question != "raw-symbol-declaration-count" || node.Kind != ast.KindIdentifier {
		return fmt.Errorf("raw-symbol-declaration-count requires an Identifier")
	}
	symbol := c.GetSymbolAtLocation(node)
	out.number(p.symbolID(symbol))
	count := 0
	if symbol != nil {
		count = len(symbol.Declarations)
	}
	out.number(uint64(count))
	return nil
}
