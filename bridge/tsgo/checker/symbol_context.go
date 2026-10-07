package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Declaration ancestry is raw binder/AST information. Interpret it in Adamic.
func (p *Program) contextDeclaration(out *fields, node *ast.Node) {
	source := ast.GetSourceFileOfNode(node)
	path := ""
	declared, library, external := false, false, false
	if source != nil {
		path = source.FileName()
		declared = source.IsDeclarationFile
		library = p.Compiler.IsSourceFileDefaultLibrary(source.Path())
		external = ast.IsExternalModule(source)
	}
	out.text(path)
	out.yes(declared)
	out.yes(library)
	out.yes(external)
	count := uint64(0)
	for n := node; n != nil; n = n.Parent {
		count++
	}
	out.number(count)
	for n := node; n != nil; n = n.Parent {
		out.text(strings.TrimPrefix(n.Kind.String(), "Kind"))
		out.number(uint64(n.Pos()))
		out.number(uint64(n.End()))
		out.number(uint64(n.Flags))
		name := ""
		if declarationName := n.Name(); declarationName != nil {
			switch declarationName.Kind {
			case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral,
				ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindNoSubstitutionTemplateLiteral:
				name = declarationName.Text()
			}
		}
		out.text(name)
		out.yes(ast.IsGlobalScopeAugmentation(n))
	}
	body := node.Body()
	out.yes(body != nil)
	if body != nil {
		out.number(uint64(body.Pos()))
		out.number(uint64(body.End()))
		out.text(strings.TrimPrefix(body.Kind.String(), "Kind"))
	}
	flags := uint64(0)
	if ast.IsFunctionLike(node) {
		flags = uint64(ast.GetFunctionFlags(node))
	}
	out.number(flags)
}
func (p *Program) contextSymbol(out *fields, symbol *ast.Symbol) {
	out.number(p.symbolID(symbol))
	if symbol == nil {
		out.number(0)
		out.number(0)
		return
	}
	out.number(uint64(symbol.Flags))
	out.number(uint64(len(symbol.Declarations)))
	for _, node := range symbol.Declarations {
		p.contextDeclaration(out, node)
	}
}
func (p *Program) symbolContext(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "symbol-context" {
		return "", fmt.Errorf("symbol-context has no suffix")
	}
	symbol := c.GetSymbolAtLocation(node)
	p.contextSymbol(out, symbol)
	resolved := symbol
	if resolved != nil && resolved.Flags&ast.SymbolFlagsAlias != 0 {
		resolved = c.GetAliasedSymbol(resolved)
	}
	p.contextSymbol(out, resolved)
	var shorthand *ast.Symbol
	if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
		shorthand = c.GetShorthandAssignmentValueSymbol(node.Parent)
	}
	out.number(p.symbolID(shorthand))
	return out.String(), nil
}
