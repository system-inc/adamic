package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// jsxStructure supplies syntax roles missing from preference-structure.
// IDs use the same preorder; no rule predicates or diagnostics are supplied.
func (p *Program) jsxStructure(out *fields, source *ast.SourceFile, node *ast.Node, question string) (string, error) {
	if question != "jsx-structure" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("jsx-structure requires a SourceFile")
	}
	if len(source.Diagnostics()) != 0 {
		return "", fmt.Errorf("source has parse diagnostics")
	}
	ids := map[*ast.Node]uint64{}
	var nodes []*ast.Node
	var visit func(*ast.Node)
	visit = func(n *ast.Node) {
		if n == nil || ids[n] != 0 {
			return
		}
		ids[n] = uint64(len(nodes) + 1)
		nodes = append(nodes, n)
		n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(node)
	out.number(uint64(len(nodes)))
	for _, n := range nodes {
		var name, value, tag, attrs, opening, body *ast.Node
		name = n.Name()
		text, token := "", ""
		switch n.Kind {
		case ast.KindBigIntLiteral, ast.KindRegularExpressionLiteral:
			text = n.Text()
		case ast.KindJsxAttribute:
			value = n.AsJsxAttribute().Initializer
		case ast.KindJsxOpeningElement:
			tag = n.AsJsxOpeningElement().TagName
			attrs = n.AsJsxOpeningElement().Attributes
		case ast.KindJsxSelfClosingElement:
			tag = n.AsJsxSelfClosingElement().TagName
			attrs = n.AsJsxSelfClosingElement().Attributes
		case ast.KindJsxElement:
			opening = n.AsJsxElement().OpeningElement
		case ast.KindReturnStatement:
			value = n.AsReturnStatement().Expression
		case ast.KindExportAssignment:
			value = n.AsExportAssignment().Expression
		case ast.KindImportDeclaration:
			value = n.AsImportDeclaration().ModuleSpecifier
		case ast.KindHeritageClause:
			token = strings.TrimPrefix(n.AsHeritageClause().Token.String(), "Kind")
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			body = n.Body()
		}
		for _, role := range []*ast.Node{name, value, tag, attrs, opening, body} {
			out.number(ids[role])
		}
		out.text(text)
		out.text(token)
		modifierEnd := 0
		async := false
		if modifiers := n.Modifiers(); modifiers != nil {
			for _, modifier := range modifiers.Nodes {
				modifierEnd = modifier.End()
				async = async || modifier.Kind == ast.KindAsyncKeyword
			}
		}
		generator := false
		if n.Kind == ast.KindFunctionDeclaration {
			generator = n.AsFunctionDeclaration().AsteriskToken != nil
		}
		if n.Kind == ast.KindFunctionExpression {
			generator = n.AsFunctionExpression().AsteriskToken != nil
		}
		out.number(uint64(modifierEnd))
		out.yes(async && generator)
	}
	return out.String(), nil
}
