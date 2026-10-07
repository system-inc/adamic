package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// symbolsInScope returns checker scope facts, not label clash judgments.
func (p *Program) symbolsInScope(c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "symbols-in-scope" || node.Kind != ast.KindLabeledStatement {
		return "", fmt.Errorf("symbols-in-scope requires a LabeledStatement")
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
