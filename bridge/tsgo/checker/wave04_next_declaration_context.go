package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// wave04NextDeclarationContext exposes raw declaration ancestry and source flags.
// It makes no judgment about globals, process members or timer functions.
func (p *Program) wave04NextDeclarationContext(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "wave04-next-declaration-context" {
		return "", fmt.Errorf("unexpected declaration context suffix")
	}
	symbol := c.GetSymbolAtLocation(node)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = c.GetAliasedSymbol(symbol)
	}
	out.yes(symbol != nil)
	if symbol != nil {
		out.number(p.symbolID(symbol))
		out.number(uint64(symbol.Flags))
		out.text(symbol.Name)
		out.number(uint64(len(symbol.Declarations)))
		for _, declaration := range symbol.Declarations {
			p.wave04NextWriteContext(out, declaration)
		}
	}
	return out.String(), nil
}

func (p *Program) wave04NextWriteContext(out *fields, node *ast.Node) {
	source := ast.GetSourceFileOfNode(node)
	out.text(source.FileName())
	out.yes(source.IsDeclarationFile)
	out.yes(p.Compiler.IsSourceFileDefaultLibrary(source.Path()))
	out.yes(ast.IsExternalModule(source))
	var ancestors []*ast.Node
	for current := node; current != nil; current = current.Parent {
		ancestors = append(ancestors, current)
	}
	out.number(uint64(len(ancestors)))
	for _, ancestor := range ancestors {
		out.text(strings.TrimPrefix(ancestor.Kind.String(), "Kind"))
		out.number(uint64(ancestor.Pos()))
		out.number(uint64(ancestor.End()))
		out.number(uint64(ancestor.Flags))
		name := ""
		if declared := ancestor.Name(); declared != nil && (ast.IsIdentifier(declared) || ast.IsStringLiteralLike(declared) || declared.Kind == ast.KindNumericLiteral) {
			name = declared.Text()
		}
		out.text(name)
		out.yes(ancestor.Kind == ast.KindModuleDeclaration && ast.IsGlobalScopeAugmentation(ancestor))
	}
}
