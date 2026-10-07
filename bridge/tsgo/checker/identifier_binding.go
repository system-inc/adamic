package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

func init() { questionExtensions["identifier-binding"] = identifierBindingQuestion }

// A raw symbol identity and its declarations, preserving aliases and binder order.
func identifierBindingQuestion(p *Program, out *fields, c *checker.Checker, source *ast.SourceFile, node *ast.Node, question string) (string, error) {
	if question != "identifier-binding" || node.Kind != ast.KindIdentifier {
		return "", fmt.Errorf("identifier-binding requires an identifier")
	}
	symbol := c.GetSymbolAtLocation(node)
	if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
		if value := c.GetShorthandAssignmentValueSymbol(node.Parent); value != nil {
			symbol = value
		}
	}
	out.yes(symbol != nil)
	if symbol == nil {
		return out.String(), nil
	}
	out.number(p.symbolID(symbol))
	out.number(uint64(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil {
			return "", fmt.Errorf("binding declaration has no source file")
		}
		out.text(file.FileName())
		out.yes(file.IsDeclarationFile)
		out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
		out.number(uint64(declaration.Pos()))
		out.number(uint64(declaration.End()))
	}
	return out.String(), nil
}
