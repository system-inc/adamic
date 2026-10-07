package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// Raw symbol identity, declaration flags and alias metadata. No lint decision.
func (p *Program) wave08Symbol(out *fields, c *checker.Checker, node *ast.Node, question string) error {
	if question != "wave08-symbol" {
		return fmt.Errorf("unexpected wave08-symbol suffix")
	}
	symbol := c.GetSymbolAtLocation(node)
	if node.Kind == ast.KindExportSpecifier {
		symbol = node.Symbol()
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
		symbol = c.GetShorthandAssignmentValueSymbol(node.Parent)
	}
	out.number(p.symbolID(symbol))
	out.number(uint64(0)) // Reserved for versioned symbol attributes.
	if symbol == nil {
		out.number(0)
		out.number(0)
		out.yes(false)
		out.text("")
		out.number(0)
		out.number(0)
		return nil
	}
	out.number(uint64(symbol.Flags))
	out.number(uint64(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		out.text(file.FileName())
		out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
		out.number(uint64(declaration.Pos()))
		out.number(uint64(declaration.End()))
		out.number(uint64(declaration.Flags))
	}
	out.yes(c.GetTypeOnlyAliasDeclaration(symbol) != nil)
	target := symbol
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		target = c.GetAliasedSymbol(symbol)
	}
	if target == nil || c.IsUnknownSymbol(target) {
		out.text("")
		out.number(0)
		out.number(0)
		return nil
	}
	out.text(target.Name)
	out.number(uint64(target.Flags))
	out.number(p.symbolID(target))
	declaration := target.ValueDeclaration
	out.yes(declaration != nil)
	if declaration != nil {
		file := ast.GetSourceFileOfNode(declaration)
		out.text(file.FileName())
		out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
		out.number(uint64(declaration.Flags))
		out.number(uint64(declaration.ModifierFlags()))
		out.yes(declaration.Kind == ast.KindExportAssignment && declaration.AsExportAssignment().IsExportEquals)
	}
	return nil
}
