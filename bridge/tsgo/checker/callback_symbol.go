package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// callbackSymbol supplies identity and declaration count without rendering names.
func (p *Program) callbackSymbol(c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "callback-symbol" || node.Kind != ast.KindIdentifier {
		return "", fmt.Errorf("callback-symbol requires an Identifier")
	}
	out := &fields{}
	out.number(1)
	out.text(question)
	symbol := c.GetSymbolAtLocation(node)
	out.number(p.symbolID(symbol))
	count := 0
	if symbol != nil {
		count = len(symbol.Declarations)
	}
	out.number(uint64(count))
	return out.String(), nil
}
