package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// Symbol lineage exposes declaration names and container ancestry, not lint decisions.
func (p *Program) symbolLineage(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "symbol-lineage" && question != "symbol-lineage\nraw" {
		return "", fmt.Errorf("unexpected symbol-lineage suffix")
	}
	symbol := c.GetSymbolAtLocation(node)
	if question == "symbol-lineage" && symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = c.GetAliasedSymbol(symbol)
	}
	out.yes(symbol != nil)
	if symbol != nil {
		out.number(p.symbolID(symbol))
		out.number(uint64(len(symbol.Declarations)))
		for _, declaration := range symbol.Declarations {
			p.writeDeclaration(out, declaration)
			name := ""
			if declaration.Name() != nil {
				name = declaration.Name().Text()
			}
			out.text(name)
			var ancestors []*ast.Node
			for current := declaration.Parent; current != nil; current = current.Parent {
				ancestors = append(ancestors, current)
			}
			out.number(uint64(len(ancestors)))
			for _, ancestor := range ancestors {
				out.text(strings.TrimPrefix(ancestor.Kind.String(), "Kind"))
				out.number(uint64(ancestor.Flags))
				out.yes(ancestor.Kind == ast.KindSourceFile && ast.IsExternalModule(ancestor.AsSourceFile()))
				out.yes(ast.IsGlobalScopeAugmentation(ancestor))
			}
		}
	}
	t := c.GetTypeAtLocation(node)
	count := 0
	if t != nil {
		count = len(c.GetSignaturesOfType(c.GetNonNullableType(t), checker.SignatureKindCall))
	}
	out.number(uint64(count))
	return out.String(), nil
}
