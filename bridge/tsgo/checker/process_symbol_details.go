// Symbol identity and declaration provenance, without lint-specific classification.
package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

func (p *Program) writeProcessDeclaration(out *fields, n *ast.Node) {
	source := ast.GetSourceFileOfNode(n)
	if source == nil {
		panic("declaration has no source")
	}
	out.text(source.FileName())
	out.yes(source.IsDeclarationFile)
	out.yes(ast.IsExternalModule(source))
	out.yes(p.Compiler.IsSourceFileDefaultLibrary(source.Path()))
	out.text(nodeKind(n))
	out.number(uint64(n.Pos()))
	out.number(uint64(n.End()))
	out.number(uint64(n.Flags))
	flags := ast.FunctionFlagsNormal
	if ast.IsFunctionLike(n) {
		flags = ast.GetFunctionFlags(n)
	}
	out.number(uint64(flags))
	out.yes(n.Body() != nil)
	var ancestors []*ast.Node
	for parent := n.Parent; parent != nil; parent = parent.Parent {
		ancestors = append(ancestors, parent)
	}
	out.number(uint64(len(ancestors)))
	for _, parent := range ancestors {
		out.text(nodeKind(parent))
		name := ""
		if parent.Name() != nil && (parent.Name().Kind == ast.KindIdentifier || ast.IsStringLiteralLike(parent.Name()) || parent.Name().Kind == ast.KindNumericLiteral) {
			name = parent.Name().Text()
		}
		out.text(name)
		out.number(uint64(parent.Flags))
		out.yes(ast.IsGlobalScopeAugmentation(parent))
	}
}
func (p *Program) processSymbolDetails(c *checker.Checker, node *ast.Node, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) != 2 || (parts[1] != "own" && parts[1] != "alias" && parts[1] != "reference") {
		return "", fmt.Errorf("process-symbol-details requires own, alias or reference")
	}
	symbol := c.GetSymbolAtLocation(node)
	if parts[1] == "reference" && node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
		if value := c.GetShorthandAssignmentValueSymbol(node.Parent); value != nil {
			symbol = value
		}
	}
	if parts[1] == "alias" && symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = c.GetAliasedSymbol(symbol)
	}
	out := &fields{}
	out.number(1)
	out.text(parts[0])
	out.number(p.symbolID(symbol))
	flags := uint64(0)
	name := ""
	var declarations []*ast.Node
	if symbol != nil {
		flags = uint64(symbol.Flags)
		name = symbol.Name
		declarations = symbol.Declarations
	}
	out.number(flags)
	out.text(strings.ToValidUTF8(name, "�"))
	out.number(uint64(len(declarations)))
	for _, declaration := range declarations {
		p.writeProcessDeclaration(out, declaration)
	}
	return out.String(), nil
}
