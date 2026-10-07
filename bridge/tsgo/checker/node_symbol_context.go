package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// nodeSymbolContext supplies symbol identities and raw declaration ancestry.
// Global-scope, ambient and library decisions belong to the native rule.
func (p *Program) nodeSymbolContext(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "node-symbol-context" && question != "node-symbol-context\nfollow-alias" {
		return "", fmt.Errorf("unexpected symbol-context suffix")
	}
	symbol := c.GetSymbolAtLocation(node)
	if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
		if value := c.GetShorthandAssignmentValueSymbol(node.Parent); value != nil {
			symbol = value
		}
	}
	if question == "node-symbol-context\nfollow-alias" && symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = c.GetAliasedSymbol(symbol)
	}
	out.yes(symbol != nil)
	if symbol == nil {
		return out.String(), nil
	}
	out.number(p.symbolID(symbol))
	out.number(uint64(symbol.Flags))
	out.text(symbol.Name)
	out.number(uint64(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		source := ast.GetSourceFileOfNode(declaration)
		out.text(source.FileName())
		out.yes(source.IsDeclarationFile)
		out.yes(p.Compiler.IsSourceFileDefaultLibrary(source.Path()))
		out.yes(ast.IsExternalModule(source))
		var chain []*ast.Node
		for current := declaration; current != nil; current = current.Parent {
			chain = append(chain, current)
		}
		out.number(uint64(len(chain)))
		for _, current := range chain {
			out.text(strings.TrimPrefix(current.Kind.String(), "Kind"))
			out.number(uint64(current.Pos()))
			out.number(uint64(current.End()))
			out.number(uint64(current.Flags))
			name := ""
			if n := current.Name(); n != nil && (n.Kind == ast.KindIdentifier || n.Kind == ast.KindStringLiteral || n.Kind == ast.KindNumericLiteral) {
				name = n.Text()
			}
			out.text(name)
			out.yes(current.Kind == ast.KindModuleDeclaration && ast.IsGlobalScopeAugmentation(current))
		}
	}
	return out.String(), nil
}
