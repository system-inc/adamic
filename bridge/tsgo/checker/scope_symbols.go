package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Names and flags in the checker's value scope at this exact statement.
// Selection of a name and the lint judgment remain native.
func (p *Program) scopeSymbols(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "scope-symbols" || node.Kind != ast.KindLabeledStatement {
		return "", fmt.Errorf("scope-symbols requires a LabeledStatement")
	}
	symbols := c.GetSymbolsInScope(node, ast.SymbolFlagsValue)
	out.number(uint64(len(symbols)))
	for _, symbol := range symbols {
		out.text(strings.ToValidUTF8(symbol.Name, "�"))
		out.number(uint64(symbol.Flags))
	}
	return out.String(), nil
}
