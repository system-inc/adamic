package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// PlatformSymbol exposes declaration ancestry, without classifying platform APIs.
func (p *Program) platformSymbol(c *checker.Checker, node *ast.Node, question string) (string, error) {
	const mode = "platform-symbol"
	if question != mode || (node.Kind != ast.KindIdentifier && node.Kind != ast.KindStringLiteral) {
		return "", fmt.Errorf("platform-symbol requires a name without suffix")
	}
	out := &fields{}
	out.number(1)
	out.text(mode)
	symbol := c.GetSymbolAtLocation(node)
	p.writePlatformSymbol(out, symbol)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = c.GetAliasedSymbol(symbol)
	}
	p.writePlatformSymbol(out, symbol)
	return out.String(), nil
}
func (p *Program) writePlatformSymbol(out *fields, symbol *ast.Symbol) {
	out.yes(symbol != nil)
	if symbol == nil {
		return
	}
	out.number(p.symbolID(symbol))
	out.number(uint64(symbol.Flags))
	out.number(uint64(len(symbol.Declarations)))
	for _, declaration := range symbol.Declarations {
		p.writePlatformDeclaration(out, declaration)
	}
}
func (p *Program) writePlatformDeclaration(out *fields, declaration *ast.Node) {
	source := ast.GetSourceFileOfNode(declaration)
	out.text(source.FileName())
	out.yes(source.IsDeclarationFile)
	out.yes(p.Compiler.IsSourceFileDefaultLibrary(source.Path()))
	out.yes(ast.IsExternalModule(source))
	out.number(uint64(declaration.Flags))
	out.number(uint64(ast.GetFunctionFlags(declaration)))
	var chain []*ast.Node
	for at := declaration; at != nil; at = at.Parent {
		chain = append(chain, at)
	}
	out.number(uint64(len(chain)))
	for _, at := range chain {
		out.text(strings.TrimPrefix(at.Kind.String(), "Kind"))
		name, nameKind := "", ""
		if declared := at.Name(); declared != nil {
			nameKind = strings.TrimPrefix(declared.Kind.String(), "Kind")
			if declared.Kind == ast.KindIdentifier || declared.Kind == ast.KindPrivateIdentifier || declared.Kind == ast.KindNumericLiteral || ast.IsStringLiteralLike(declared) {
				name = declared.Text()
			}
		}
		out.text(name)
		out.text(nameKind)
		out.number(uint64(at.Pos()))
		out.number(uint64(at.End()))
		out.yes(ast.IsGlobalScopeAugmentation(at))
	}
	body := declaration.Body()
	out.yes(body != nil)
	if body != nil {
		out.text(strings.TrimPrefix(body.Kind.String(), "Kind"))
		out.number(uint64(body.Pos()))
		out.number(uint64(body.End()))
	}
}
