package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Raw resolution facts: own and aliased qualifier declarations, followed by the
// export-normalized identity of the first matching symbol in scope. No lint
// decision, namespace containment test, diagnostic or repair runs here.
func (p *Program) scopeExportSymbols(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) != 3 {
		return "", fmt.Errorf("scope-export-symbols requires flags and name")
	}
	flags, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil || strconv.FormatUint(flags, 10) != parts[1] {
		return "", fmt.Errorf("invalid scope symbol flags")
	}
	if node.Kind != ast.KindIdentifier && node.Kind != ast.KindQualifiedName && node.Kind != ast.KindPropertyAccessExpression {
		return "", fmt.Errorf("scope-export-symbols requires an entity name")
	}
	symbol := c.GetSymbolAtLocation(node)
	p.writeSymbolDetails(out, symbol)
	var alias *ast.Symbol
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		alias = c.GetAliasedSymbol(symbol)
	}
	p.writeSymbolDetails(out, alias)
	var scoped *ast.Symbol
	for _, candidate := range c.GetSymbolsInScope(node, ast.SymbolFlags(flags)) {
		if candidate.Name == parts[2] {
			scoped = c.GetExportSymbolOfSymbol(candidate)
			break
		}
	}
	out.number(p.symbolID(scoped))
	return out.String(), nil
}
