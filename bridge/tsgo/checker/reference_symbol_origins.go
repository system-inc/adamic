package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// Identity and declaration locations of the value a reference reads. The
// shorthand/export accessors answer checker facts, without any lint decision.
func (p *Program) referenceSymbolOrigins(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "reference-symbol-origins" || node.Kind != ast.KindIdentifier {
		return "", fmt.Errorf("reference-symbol-origins requires an identifier and no arguments")
	}
	symbol := c.GetSymbolAtLocation(node)
	if parent := node.Parent; parent != nil {
		if parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == node {
			symbol = c.GetShorthandAssignmentValueSymbol(parent)
		}
		if parent.Kind == ast.KindExportSpecifier {
			symbol = c.GetExportSpecifierLocalTargetSymbol(parent)
		}
	}
	out.number(p.symbolID(symbol))
	if symbol != nil {
		out.number(uint64(len(symbol.Declarations)))
		for _, declaration := range symbol.Declarations {
			file := ast.GetSourceFileOfNode(declaration)
			if file == nil {
				return "", fmt.Errorf("reference declaration has no source")
			}
			out.text(file.FileName())
			out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
			out.number(uint64(declaration.Pos()))
			out.number(uint64(declaration.End()))
			out.yes(file.IsDeclarationFile)
		}
	}
	return out.String(), nil
}
