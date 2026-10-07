package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// bindingOrigin preserves the raw symbol and the symbol a shorthand/export reads.
// It returns compiler metadata, with no global/reference or lint judgment.
func (p *Program) bindingOrigin(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "binding-origin" || node.Kind != ast.KindIdentifier {
		return "", fmt.Errorf("binding-origin requires an Identifier and no suffix")
	}
	original := c.GetSymbolAtLocation(node)
	read := original
	if parent := node.Parent; parent != nil {
		if parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == node {
			read = c.GetShorthandAssignmentValueSymbol(parent)
		}
		if parent.Kind == ast.KindExportSpecifier {
			read = c.GetExportSpecifierLocalTargetSymbol(parent)
		}
	}
	for _, symbol := range []*ast.Symbol{original, read} {
		out.number(p.symbolID(symbol))
		if symbol == nil {
			out.number(0)
			continue
		}
		out.number(uint64(len(symbol.Declarations)))
		for _, declaration := range symbol.Declarations {
			source := ast.GetSourceFileOfNode(declaration)
			if source == nil {
				return "", fmt.Errorf("binding-origin declaration without source")
			}
			out.text(source.FileName())
			out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
			out.number(uint64(declaration.Pos()))
			out.number(uint64(declaration.End()))
			out.yes(source.IsDeclarationFile)
		}
	}
	return out.String(), nil
}
