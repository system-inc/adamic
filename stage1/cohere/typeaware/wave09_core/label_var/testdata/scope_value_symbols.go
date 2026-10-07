// Raw checker operation. Registration is owned by the shared bridge integrator.
package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func scopeValueSymbols(c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "scope-value-symbols" || node.Kind != ast.KindLabeledStatement {
		return "", fmt.Errorf("scope-value-symbols requires a LabeledStatement")
	}
	out := &fields{}
	out.number(1)
	out.text(question)
	symbols := c.GetSymbolsInScope(node, ast.SymbolFlagsValue)
	out.number(uint64(len(symbols)))
	for _, symbol := range symbols {
		out.text(symbol.Name)
		out.number(uint64(symbol.Flags))
	}
	return out.String(), nil
}
