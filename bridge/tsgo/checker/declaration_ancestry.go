package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// declarationAncestry exposes raw symbol and ancestor metadata, never a lint judgment.
// Registration is pending in the shared Inspect dispatch; this file is isolated.
func (p *Program) declarationAncestry(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if (question != "declaration-ancestry" && question != "declaration-ancestry\nalias") || node.Kind != ast.KindIdentifier {
		return "", fmt.Errorf("declaration-ancestry requires an Identifier and either no suffix or alias")
	}
	symbol := c.GetSymbolAtLocation(node)
	if question == "declaration-ancestry\nalias" && symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = c.GetAliasedSymbol(symbol)
	}
	out.yes(symbol != nil)
	if symbol == nil {
		return out.String(), nil
	}
	out.number(p.symbolID(symbol))
	out.number(uint64(symbol.Flags))
	out.number(uint64(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil {
			return "", fmt.Errorf("declaration has no source")
		}
		out.text(file.FileName())
		out.yes(file.IsDeclarationFile)
		out.yes(p.Compiler.IsSourceFileDefaultLibrary(file.Path()))
		out.yes(ast.IsExternalModule(file))
		var chain []*ast.Node
		for ancestor := declaration; ancestor != nil; ancestor = ancestor.Parent {
			chain = append(chain, ancestor)
		}
		out.number(uint64(len(chain)))
		for _, ancestor := range chain {
			out.text(strings.TrimPrefix(ancestor.Kind.String(), "Kind"))
			name := ""
			if identifier := ancestor.Name(); identifier != nil && identifier.Kind == ast.KindIdentifier {
				name = identifier.Text()
			}
			out.text(name)
			out.number(uint64(ancestor.Pos()))
			out.number(uint64(ancestor.End()))
			out.number(uint64(ancestor.Flags))
			out.yes(ancestor.Kind == ast.KindModuleDeclaration && ast.IsGlobalScopeAugmentation(ancestor))
		}
	}
	return out.String(), nil
}
