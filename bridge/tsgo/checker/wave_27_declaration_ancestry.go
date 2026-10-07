package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// wave-27-declaration-ancestry exposes symbol identity and raw ancestor metadata. It
// makes no platform-membership, timeout, stream or lint judgment.
func (p *Program) wave27DeclarationAncestry(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "wave-27-declaration-ancestry" {
		return "", fmt.Errorf("unexpected wave-27-declaration-ancestry suffix")
	}
	symbol := c.GetSymbolAtLocation(node)
	out.number(p.symbolID(symbol))
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = c.GetAliasedSymbol(symbol)
	}
	out.number(p.symbolID(symbol))
	out.yes(symbol != nil)
	if symbol != nil {
		out.number(uint64(symbol.Flags))
		out.number(uint64(len(symbol.Declarations)))
		for _, declaration := range symbol.Declarations {
			source := ast.GetSourceFileOfNode(declaration)
			if source == nil {
				return "", fmt.Errorf("declaration has no source")
			}
			out.text(source.FileName())
			out.yes(source.IsDeclarationFile)
			out.yes(p.Compiler.IsSourceFileDefaultLibrary(source.Path()))
			out.yes(ast.IsExternalModule(source))
			out.number(uint64(declaration.Pos()))
			out.number(uint64(declaration.End()))
			var ancestors []*ast.Node
			for current := declaration; current != nil; current = current.Parent {
				ancestors = append(ancestors, current)
			}
			out.number(uint64(len(ancestors)))
			for _, current := range ancestors {
				out.text(strings.TrimPrefix(current.Kind.String(), "Kind"))
				name, nameKind := "", ""
				if n := current.Name(); n != nil {
					nameKind = strings.TrimPrefix(n.Kind.String(), "Kind")
					switch n.Kind {
					case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNoSubstitutionTemplateLiteral:
						name = n.Text()
					}
				}
				out.text(nameKind)
				out.text(name)
				out.number(uint64(current.Flags))
				out.yes(ast.IsGlobalScopeAugmentation(current))
			}
		}
	}
	return out.String(), nil
}
