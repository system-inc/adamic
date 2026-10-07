package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// Raw symbol declarations and named ancestors. No NodeJS or lint predicates.
func (p *Program) symbolDeclarationPaths(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "symbol-declaration-paths" {
		return "", fmt.Errorf("unexpected symbol-declaration-paths suffix")
	}
	symbol := c.GetSymbolAtLocation(node)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = c.GetAliasedSymbol(symbol)
	}
	out.yes(symbol != nil)
	if symbol == nil {
		return out.String(), nil
	}
	out.number(uint64(symbol.Flags))
	out.number(uint64(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		out.text(file.FileName())
		out.yes(file.IsDeclarationFile)
		out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
		out.number(uint64(declaration.Pos()))
		out.number(uint64(declaration.End()))
		out.number(uint64(declaration.Flags))
		name := declaration.Name()
		text := ""
		if name != nil {
			text = processScalarName(name)
		}
		out.text(text)
		var ancestors []*ast.Node
		for current := declaration.Parent; current != nil; current = current.Parent {
			ancestors = append(ancestors, current)
		}
		out.number(uint64(len(ancestors)))
		for _, ancestor := range ancestors {
			out.text(strings.TrimPrefix(ancestor.Kind.String(), "Kind"))
			name := ancestor.Name()
			out.yes(name != nil)
			if name != nil {
				out.text(strings.TrimPrefix(name.Kind.String(), "Kind"))
				out.text(processScalarName(name))
			}
		}
	}
	return out.String(), nil
}

// Binding and computed names have no scalar text. Their kind remains in the wire.
func processScalarName(name *ast.Node) string {
	switch name.Kind {
	case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral:
		return name.Text()
	default:
		return ""
	}
}
