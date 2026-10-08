package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// Reference identity and declaration locations, including the value of shorthand and local exports.
func (p *Program) referenceSymbol(out *fields, c *checker.Checker, node *ast.Node, source *ast.SourceFile, question string) (string, error) {
	if question != "reference-symbol" {
		return p.streamSymbol(out, c, node, source, question)
	}
	if node.Kind != ast.KindIdentifier {
		return "", fmt.Errorf("reference-symbol requires an Identifier")
	}
	symbol := c.GetSymbolAtLocation(node)
	if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment && node.Parent.Name() == node {
		symbol = c.GetShorthandAssignmentValueSymbol(node.Parent)
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindExportSpecifier {
		symbol = c.GetExportSpecifierLocalTargetSymbol(node.Parent)
	}
	out.number(p.symbolID(symbol))
	count := 0
	if symbol != nil {
		count = len(symbol.Declarations)
	}
	out.number(uint64(count))
	if symbol != nil {
		for _, decl := range symbol.Declarations {
			file := ast.GetSourceFileOfNode(decl)
			if file == nil {
				return "", fmt.Errorf("reference declaration has no source")
			}
			out.text(file.FileName().AsString())
			out.yes(file.IsDeclarationFile)
			out.text(strings.TrimPrefix(decl.Kind.String(), "Kind"))
			out.number(uint64(decl.Pos()))
			out.number(uint64(decl.End()))
		}
	}
	return out.String(), nil
}
