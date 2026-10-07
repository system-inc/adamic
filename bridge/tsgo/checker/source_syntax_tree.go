package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"strings"
)

// Unclassified compiler syntax. All component inference and lint judgments belong to Adamic.
func (p *Program) sourceSyntaxTree(out *fields, node *ast.Node, question string) (string, error) {
	if question != "source-syntax-tree" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("source-syntax-tree requires a SourceFile")
	}
	sf := node.AsSourceFile()
	nodes := []*ast.Node{}
	indices := map[*ast.Node]uint64{}
	parents := map[*ast.Node]*ast.Node{}
	children := map[*ast.Node][]*ast.Node{}
	var visit func(*ast.Node, *ast.Node)
	visit = func(current, parent *ast.Node) {
		indices[current] = uint64(len(nodes) + 1)
		nodes = append(nodes, current)
		parents[current] = parent
		current.ForEachChild(func(child *ast.Node) bool {
			children[current] = append(children[current], child)
			visit(child, current)
			return false
		})
	}
	visit(node, nil)
	out.number(uint64(len(nodes)))
	list := func(items []*ast.Node) {
		out.number(uint64(len(items)))
		for _, item := range items {
			out.number(indices[item])
		}
	}
	for _, current := range nodes {
		out.text(strings.TrimPrefix(current.Kind.String(), "Kind"))
		out.number(uint64(max(0, current.Pos())))
		out.number(uint64(max(0, current.End())))
		out.number(uint64(max(0, scanner.GetRangeOfTokenAtPosition(sf, max(0, current.Pos())).Pos())))
		text := ""
		switch current.Kind {
		case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindNoSubstitutionTemplateLiteral:
			text = current.Text()
		}
		out.text(text)
		out.number(indices[parents[current]])
		list(children[current])
		out.number(indices[current.Name()])
		out.number(indices[current.Type()])
		var initializer *ast.Node
		switch current.Kind {
		case ast.KindVariableDeclaration, ast.KindParameter, ast.KindBindingElement, ast.KindPropertyDeclaration, ast.KindPropertySignature, ast.KindPropertyAssignment, ast.KindEnumMember, ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement, ast.KindJsxAttribute:
			initializer = current.Initializer()
		}
		out.number(indices[initializer])
		out.number(indices[current.Body()])
		out.number(uint64(current.Flags))
		flags := uint64(0)
		if ast.IsFunctionLike(current) {
			flags = uint64(ast.GetFunctionFlags(current))
		}
		out.number(flags)
		expression, left, right, yes, no := (*ast.Node)(nil), (*ast.Node)(nil), (*ast.Node)(nil), (*ast.Node)(nil), (*ast.Node)(nil)
		operator := ""
		rest := false
		arguments := []*ast.Node{}
		parameters := []*ast.Node{}
		if ast.IsFunctionLike(current) {
			if list := current.ParameterList(); list != nil {
				parameters = list.Nodes
			}
		}
		switch current.Kind {
		case ast.KindHeritageClause:
			operator = strings.TrimPrefix(current.AsHeritageClause().Token.String(), "Kind")
		case ast.KindCallExpression:
			call := current.AsCallExpression()
			expression = call.Expression
			if call.Arguments != nil {
				arguments = call.Arguments.Nodes
			}
		case ast.KindPropertyAccessExpression:
			expression = current.AsPropertyAccessExpression().Expression
		case ast.KindElementAccessExpression:
			access := current.AsElementAccessExpression()
			expression = access.Expression
			right = access.ArgumentExpression
		case ast.KindBinaryExpression:
			binary := current.AsBinaryExpression()
			left = binary.Left
			right = binary.Right
			operator = strings.TrimPrefix(binary.OperatorToken.Kind.String(), "Kind")
		case ast.KindConditionalExpression:
			conditional := current.AsConditionalExpression()
			expression = conditional.Condition
			yes = conditional.WhenTrue
			no = conditional.WhenFalse
		case ast.KindParameter:
			rest = current.AsParameterDeclaration().DotDotDotToken != nil
		case ast.KindParenthesizedExpression:
			expression = current.AsParenthesizedExpression().Expression
		case ast.KindNonNullExpression:
			expression = current.AsNonNullExpression().Expression
		case ast.KindReturnStatement:
			expression = current.AsReturnStatement().Expression
		case ast.KindComputedPropertyName:
			expression = current.AsComputedPropertyName().Expression
		case ast.KindTypeReference:
			expression = current.AsTypeReferenceNode().TypeName
		case ast.KindParenthesizedType:
			expression = current.AsParenthesizedTypeNode().Type
		case ast.KindExpressionWithTypeArguments:
			expression = current.AsExpressionWithTypeArguments().Expression
		}
		out.number(indices[expression])
		out.number(indices[left])
		out.number(indices[right])
		out.number(indices[yes])
		out.number(indices[no])
		out.text(operator)
		out.yes(rest)
		list(arguments)
		list(parameters)
	}
	return out.String(), nil
}
