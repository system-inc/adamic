package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// preferenceStructure extends raw syntax roles for property access and binding writes.
// All decisions about references, globals and writes stay in Adamic.
func (p *Program) preferenceStructure(out *fields, source *ast.SourceFile, node *ast.Node, question string) (string, error) {
	if question != "preference-structure" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("preference-structure requires a SourceFile")
	}
	if len(source.Diagnostics()) != 0 {
		return "", fmt.Errorf("source has parse diagnostics")
	}
	ids := map[*ast.Node]uint64{}
	nodes := []*ast.Node{}
	var visit func(*ast.Node)
	visit = func(n *ast.Node) {
		if n == nil {
			return
		}
		if ids[n] != 0 {
			return
		}
		ids[n] = uint64(len(nodes) + 1)
		nodes = append(nodes, n)
		n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(node)
	module := p.Compiler.Options().GetEmitModuleKind()
	out.number(uint64(module))
	out.number(uint64(len(nodes)))
	for _, n := range nodes {
		out.text(strings.TrimPrefix(n.Kind.String(), "Kind"))
		out.number(uint64(n.Pos()))
		out.number(uint64(n.End()))
		out.number(uint64(scanner.GetTokenPosOfNode(n, source, false)))
		out.number(ids[n.Parent])
		out.number(uint64(n.Flags))
		text := ""
		if ast.IsIdentifier(n) || n.Kind == ast.KindStringLiteral || n.Kind == ast.KindNoSubstitutionTemplateLiteral || n.Kind == ast.KindNumericLiteral || n.Kind == ast.KindTemplateHead || n.Kind == ast.KindTemplateMiddle || n.Kind == ast.KindTemplateTail || n.Kind == ast.KindPrivateIdentifier {
			text = n.Text()
		}
		out.text(text)
		var children []uint64
		n.ForEachChild(func(child *ast.Node) bool { children = append(children, ids[child]); return false })
		out.ids(children)
		var expression, name, typ, body, left, right, yes, no *ast.Node
		var args, parameters []*ast.Node
		operator := ""
		spread := false
		switch n.Kind {
		case ast.KindTaggedTemplateExpression:
			expression = n.AsTaggedTemplateExpression().Tag
			name = n.AsTaggedTemplateExpression().Template
		case ast.KindTemplateExpression:
			name = n.AsTemplateExpression().Head
		case ast.KindTemplateSpan:
			expression = n.AsTemplateSpan().Expression
			name = n.AsTemplateSpan().Literal
		case ast.KindComputedPropertyName:
			expression = n.AsComputedPropertyName().Expression
		case ast.KindExportSpecifier:
			name = n.Name()
			typ = n.AsExportSpecifier().PropertyName
		case ast.KindExportDeclaration:
			expression = n.AsExportDeclaration().ModuleSpecifier
		case ast.KindClassDeclaration, ast.KindClassExpression, ast.KindMethodDeclaration, ast.KindMethodSignature, ast.KindPropertySignature, ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport:
			name = n.Name()
		case ast.KindElementAccessExpression:
			expression = n.AsElementAccessExpression().Expression
			name = n.AsElementAccessExpression().ArgumentExpression
		case ast.KindExpressionWithTypeArguments:
			expression = n.AsExpressionWithTypeArguments().Expression
		case ast.KindPrefixUnaryExpression:
			expression = n.AsPrefixUnaryExpression().Operand
			operator = strings.TrimPrefix(n.AsPrefixUnaryExpression().Operator.String(), "Kind")
		case ast.KindPostfixUnaryExpression:
			expression = n.AsPostfixUnaryExpression().Operand
			operator = strings.TrimPrefix(n.AsPostfixUnaryExpression().Operator.String(), "Kind")
		case ast.KindShorthandPropertyAssignment:
			name = n.Name()
			expression = n.AsShorthandPropertyAssignment().ObjectAssignmentInitializer
		case ast.KindForInStatement, ast.KindForOfStatement:
			expression = n.AsForInOrOfStatement().Initializer
		case ast.KindPropertyAssignment, ast.KindPropertyDeclaration, ast.KindBindingElement, ast.KindEnumMember:
			name = n.Name()
			expression = n.Initializer()
			if n.Kind == ast.KindBindingElement {
				typ = n.AsBindingElement().PropertyName
				spread = n.AsBindingElement().DotDotDotToken != nil
			}
		case ast.KindPropertyAccessExpression:
			expression = n.AsPropertyAccessExpression().Expression
			name = n.Name()
		case ast.KindCallExpression, ast.KindNewExpression:
			expression = n.Expression()
			args = n.Arguments()
		case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindSatisfiesExpression:
			expression = n.Expression()
			if n.Kind == ast.KindAsExpression || n.Kind == ast.KindTypeAssertionExpression {
				typ = n.Type()
			}
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
			parameters = n.Parameters()
			body = n.Body()
			name = n.Name()
		case ast.KindParameter, ast.KindVariableDeclaration:
			name = n.Name()
			expression = n.Initializer()
		case ast.KindBinaryExpression:
			b := n.AsBinaryExpression()
			left = b.Left
			right = b.Right
			operator = strings.TrimPrefix(b.OperatorToken.Kind.String(), "Kind")
		case ast.KindConditionalExpression:
			b := n.AsConditionalExpression()
			yes = b.WhenTrue
			no = b.WhenFalse
		case ast.KindJsxExpression:
			b := n.AsJsxExpression()
			expression = b.Expression
			spread = b.DotDotDotToken != nil
		}
		for _, role := range []*ast.Node{expression, name, typ, body, left, right, yes, no} {
			out.number(ids[role])
		}
		var a, b []uint64
		for _, v := range args {
			a = append(a, ids[v])
		}
		for _, v := range parameters {
			b = append(b, ids[v])
		}
		out.ids(a)
		out.ids(b)
		out.text(operator)
		out.yes(spread)
		out.yes(ast.IsDeclarationName(n))
	}
	return out.String(), nil
}
