package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Raw compiler facts only. The native consumer decides findings and edits.
func (p *Program) exportSymbolChain(c *checker.Checker, node *ast.Node, question string) (string, error) {
	const mode = "export-symbol-chain"
	if question != mode {
		return "", fmt.Errorf("unexpected export facts suffix")
	}
	out := &fields{}
	out.number(1)
	out.text(mode)
	if node.Kind != ast.KindIdentifier && node.Kind != ast.KindStringLiteral {
		return "", fmt.Errorf("export-symbol-chain requires a name")
	}
	symbol := c.GetSymbolAtLocation(node)
	var chain []*ast.Symbol
	seen := map[*ast.Symbol]bool{}
	for symbol != nil && !seen[symbol] {
		seen[symbol] = true
		chain = append(chain, symbol)
		if symbol.Flags&ast.SymbolFlagsAlias == 0 {
			break
		}
		next := checker.SkipAlias(symbol, c)
		if next == symbol {
			break
		}
		symbol = next
	}
	out.number(uint64(len(chain)))
	for _, symbol := range chain {
		out.number(uint64(symbol.Flags))
		out.number(uint64(len(symbol.Declarations)))
		for _, declaration := range symbol.Declarations {
			out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
			out.yes(declaration.IsTypeOnly())
			var ancestors []*ast.Node
			for parent := declaration.Parent; parent != nil; parent = parent.Parent {
				ancestors = append(ancestors, parent)
				if parent.Kind == ast.KindSourceFile {
					break
				}
			}
			out.number(uint64(len(ancestors)))
			for _, parent := range ancestors {
				out.text(strings.TrimPrefix(parent.Kind.String(), "Kind"))
				out.yes(parent.IsTypeOnly())
			}
		}
	}
	return out.String(), nil
}
