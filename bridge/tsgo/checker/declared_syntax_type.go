package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Returns an annotation node identity, never an inferred lint classification.
func (p *Program) declaredSyntaxType(node *ast.Node, question string) (string, error) {
	if question != "declared-syntax-type" {
		return "", fmt.Errorf("declared-syntax-type takes no arguments")
	}
	var annotation *ast.Node
	switch node.Kind {
	case ast.KindParameter:
		annotation = node.AsParameterDeclaration().Type
	case ast.KindVariableDeclaration:
		annotation = node.AsVariableDeclaration().Type
	case ast.KindPropertySignature:
		annotation = node.AsPropertySignatureDeclaration().Type
	case ast.KindTypeAliasDeclaration:
		annotation = node.AsTypeAliasDeclaration().Type
	case ast.KindParenthesizedType:
		annotation = node.AsParenthesizedTypeNode().Type
	}
	_, ids := syntaxNodes(ast.GetSourceFileOfNode(node))
	out := &fields{}
	out.number(1)
	out.text(question)
	out.number(ids[annotation])
	return out.String(), nil
}
