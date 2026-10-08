package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

// symbols-in-scope v1: caller-provided symbol meaning, then every visible name,
// raw identity and export identity. Shadowing decisions remain in the rule.
func (p *Program) scopeSymbols(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) != 2 {
		return "", fmt.Errorf("symbols-in-scope requires symbol flags")
	}
	flags, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil || strconv.FormatUint(flags, 10) != parts[1] {
		return "", fmt.Errorf("invalid scope symbol flags")
	}
	symbols := c.GetSymbolsInScope(node, ast.SymbolFlags(flags))
	out.number(uint64(len(symbols)))
	for _, symbol := range symbols {
		out.text(symbol.Name)
		out.number(p.symbolID(symbol))
		out.number(p.symbolID(c.GetExportSymbolOfSymbol(symbol)))
	}
	return out.String(), nil
}
