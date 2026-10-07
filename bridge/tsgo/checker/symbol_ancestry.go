package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

func init() { questionExtensions["symbol-ancestry"] = symbolAncestryQuestion }

// Declaration ancestors and file properties are raw binder facts, not lint decisions.
func symbolAncestryQuestion(p *Program, out *fields, c *checker.Checker, source *ast.SourceFile, node *ast.Node, question string) (string, error) {
	if question != "symbol-ancestry" || (node.Kind != ast.KindIdentifier && node.Kind != ast.KindStringLiteral) {
		return "", fmt.Errorf("symbol-ancestry requires an identifier or string literal")
	}
	symbol := c.GetSymbolAtLocation(node)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = c.GetAliasedSymbol(symbol)
	}
	if symbol == nil {
		out.number(0)
		out.number(0)
		return out.String(), nil
	}
	out.number(p.symbolID(symbol))
	out.number(uint64(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil {
			return "", fmt.Errorf("symbol declaration has no source file")
		}
		out.text(file.FileName())
		out.yes(file.IsDeclarationFile)
		out.yes(p.Compiler.IsSourceFileDefaultLibrary(file.Path()))
		out.yes(ast.IsExternalModule(file))
		var ancestors []*ast.Node
		for at := declaration; at != nil; at = at.Parent {
			ancestors = append(ancestors, at)
		}
		out.number(uint64(len(ancestors)))
		for _, ancestor := range ancestors {
			out.text(strings.TrimPrefix(ancestor.Kind.String(), "Kind"))
			out.number(uint64(ancestor.Pos()))
			out.number(uint64(ancestor.End()))
			out.number(uint64(ancestor.Flags))
			if ancestor.Kind == ast.KindModuleDeclaration {
				out.text(strings.TrimPrefix(ancestor.AsModuleDeclaration().Keyword.String(), "Kind"))
			} else {
				out.text("")
			}
			name := ancestor.Name()
			if name == nil {
				out.text("")
			} else {
				if name.Kind == ast.KindIdentifier || name.Kind == ast.KindPrivateIdentifier || ast.IsStringLiteralLike(name) || name.Kind == ast.KindNumericLiteral {
					out.text(name.Text())
				} else {
					if name.Pos() < 0 || name.End() > len(file.Text()) {
						return "", fmt.Errorf("declaration name outside source")
					}
					out.text(strings.TrimSpace(file.Text()[name.Pos():name.End()]))
				}
			}
		}
	}
	return out.String(), nil
}
