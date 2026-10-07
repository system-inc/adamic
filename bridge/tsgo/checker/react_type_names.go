package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Raw type metadata only. Native validators compare these names themselves.
func (p *Program) reactTypeNames(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "react-type-names" {
		return "", fmt.Errorf("unexpected react-type-names suffix")
	}
	t := c.GetTypeAtLocation(node)
	out.yes(t != nil)
	aliasName, symbolName := "", ""
	if t != nil {
		if alias := checker.Type_alias(t); alias != nil && alias.Symbol() != nil {
			aliasName = alias.Symbol().Name
		}
		if symbol := checker.Type_symbol(t); symbol != nil {
			symbolName = symbol.Name
		}
	}
	out.text(string([]rune(aliasName)))
	out.text(string([]rune(symbolName)))
	return out.String(), nil
}
