package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// writeProvenance exposes declaration ancestry and source facts, not platform
// classifications. Native rules decide which ambient declarations qualify.
func (p *Program) writeProvenance(out *fields, symbol *ast.Symbol) {
	out.number(p.symbolID(symbol))
	if symbol == nil {
		out.number(0)
		out.number(0)
		return
	}
	out.number(uint64(symbol.Flags))
	out.number(uint64(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		source := ast.GetSourceFileOfNode(declaration)
		if source == nil {
			panic("declaration has no source")
		}
		out.text(string(source.FileName()))
		out.yes(source.IsDeclarationFile)
		out.yes(p.Compiler.IsSourceFileDefaultLibrary(source.PathKey()))
		out.yes(ast.IsExternalModule(source))
		var ancestors []*ast.Node
		for node := declaration; node != nil; node = node.Parent {
			ancestors = append(ancestors, node)
		}
		out.number(uint64(len(ancestors)))
		for _, node := range ancestors {
			out.text(strings.TrimPrefix(node.Kind.String(), "Kind"))
			name := ""
			if named := node.Name(); named != nil && (named.Kind == ast.KindIdentifier || ast.IsStringLiteralLike(named)) {
				name = named.Text()
			}
			out.text(name)
			out.number(uint64(node.Pos()))
			out.number(uint64(node.End()))
			out.number(uint64(node.Flags))
			out.yes(ast.IsGlobalScopeAugmentation(node))
		}
	}
}
func (p *Program) symbolProvenance(out *fields, c *checker.Checker, node *ast.Node, question string) error {
	if question != "symbol-provenance" && question != "symbol-provenance-alias" {
		return fmt.Errorf("unexpected symbol provenance suffix")
	}
	symbol := c.GetSymbolAtLocation(node)
	if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
		if value := c.GetShorthandAssignmentValueSymbol(node.Parent); value != nil {
			symbol = value
		}
	}
	if question == "symbol-provenance-alias" && symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = c.GetAliasedSymbol(symbol)
	}
	p.writeProvenance(out, symbol)
	return nil
}
