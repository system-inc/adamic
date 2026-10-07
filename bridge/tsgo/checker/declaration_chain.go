package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// declarationChain supplies symbol identity and raw declaration ancestry, not lint decisions.
func (p *Program) declarationChain(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "declaration-chain" {
		return "", fmt.Errorf("unexpected declaration-chain suffix")
	}
	symbol := c.GetSymbolAtLocation(node)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = c.GetAliasedSymbol(symbol)
	}
	out.number(p.symbolID(symbol))
	if symbol == nil {
		out.number(0)
		out.text("")
		out.number(0)
		return out.String(), nil
	}
	out.number(uint64(symbol.Flags))
	out.text(strings.ToValidUTF8(symbol.Name, "�"))
	out.number(uint64(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		source := ast.GetSourceFileOfNode(declaration)
		out.yes(source != nil)
		if source == nil {
			continue
		}
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
			name := ""
			if named := current.Name(); named != nil {
				switch named.Kind {
				case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindNoSubstitutionTemplateLiteral:
					name = named.Text()
				}
			}
			out.text(name)
			out.number(uint64(current.Pos()))
			out.number(uint64(current.End()))
			out.number(uint64(current.Flags))
			out.yes(ast.IsGlobalScopeAugmentation(current))
		}
	}
	return out.String(), nil
}
