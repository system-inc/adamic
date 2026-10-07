package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Compiler access classification and binding identity, with no lint predicate.
func (p *Program) syntaxReferenceFacts(c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "syntax-reference-facts" {
		return "", fmt.Errorf("syntax-reference-facts takes no arguments")
	}
	out := &fields{}
	out.number(1)
	out.text(question)
	out.yes(ast.IsDeclarationName(node))
	out.yes(ast.IsWriteAccess(node))
	keyword := ""
	if node.Kind == ast.KindMetaProperty {
		keyword = stringsKind(node.AsMetaProperty().KeywordToken)
	}
	out.text(keyword)
	var propertyName *ast.Node
	switch node.Kind {
	case ast.KindBindingElement:
		propertyName = node.AsBindingElement().PropertyName
	case ast.KindExportSpecifier:
		propertyName = node.AsExportSpecifier().PropertyName
	}
	identity := uint64(0)
	if propertyName != nil {
		_, ids := syntaxNodes(ast.GetSourceFileOfNode(node))
		identity = ids[propertyName]
	}
	out.number(identity)
	symbol := c.GetSymbolAtLocation(node)
	if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment && node.Parent.Name() == node {
		symbol = c.GetShorthandAssignmentValueSymbol(node.Parent)
	}
	if node.Parent != nil && node.Parent.Kind == ast.KindExportSpecifier {
		symbol = c.GetExportSpecifierLocalTargetSymbol(node.Parent)
	}
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
	out.text(name)
	out.number(uint64(len(declarations)))
	for _, declaration := range declarations {
		p.writeProcessDeclaration(out, declaration)
	}
	return out.String(), nil
}
func stringsKind(kind ast.Kind) string { return kind.String()[4:] }
