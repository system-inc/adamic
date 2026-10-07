// Syntax and compiler flags only. No rule predicates or findings are computed here.
package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/bridge/tsgo/checker/process_flow"
	"strings"
)

func syntaxNodes(source *ast.SourceFile) ([]*ast.Node, map[*ast.Node]uint64) {
	var nodes []*ast.Node
	ids := map[*ast.Node]uint64{}
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		ids[node] = uint64(len(nodes) + 1)
		nodes = append(nodes, node)
		node.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
	}
	visit(source.AsNode())
	return nodes, ids
}
func nodeKind(node *ast.Node) string {
	if node == nil {
		return ""
	}
	return strings.TrimPrefix(node.Kind.String(), "Kind")
}
func (p *Program) syntaxProjection(node *ast.Node, question string) (string, error) {
	if question != "syntax-projection" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("syntax-projection requires a SourceFile")
	}
	source := node.AsSourceFile()
	nodes, ids := syntaxNodes(source)
	out := &fields{}
	out.number(1)
	out.text(question)
	out.text(source.FileName())
	out.yes(source.IsDeclarationFile)
	out.yes(ast.IsExternalModule(source))
	out.number(uint64(len(nodes)))
	for _, n := range nodes {
		out.text(nodeKind(n))
		out.number(uint64(n.Pos()))
		out.number(uint64(n.End()))
		out.number(uint64(scanner.GetTokenPosOfNode(n, source, false)))
		out.number(ids[n.Parent])
		text := ""
		switch n.Kind {
		case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateHead, ast.KindTemplateMiddle, ast.KindTemplateTail, ast.KindRegularExpressionLiteral:
			text = n.Text()
		}
		out.text(strings.ToValidUTF8(text, "�"))
		out.number(uint64(n.Flags))
		flags := ast.FunctionFlagsNormal
		if ast.IsFunctionLike(n) {
			flags = ast.GetFunctionFlags(n)
		}
		out.number(uint64(flags))
		out.number(ids[process_flow.RootOf(n)])
		out.yes(process_flow.IsRoot(n))
		out.yes(ast.IsGlobalScopeAugmentation(n))
		out.number(ids[n.Name()])
		out.number(ids[n.Body()])
		initializer := (*ast.Node)(nil)
		switch n.Kind {
		case ast.KindVariableDeclaration, ast.KindParameter, ast.KindBindingElement, ast.KindPropertyDeclaration, ast.KindPropertySignature, ast.KindPropertyAssignment, ast.KindEnumMember, ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement, ast.KindJsxAttribute:
			initializer = n.Initializer()
		}
		out.number(ids[initializer])
		expression := (*ast.Node)(nil)
		switch n.Kind {
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindParenthesizedExpression, ast.KindCallExpression, ast.KindNewExpression, ast.KindExpressionWithTypeArguments, ast.KindComputedPropertyName, ast.KindNonNullExpression, ast.KindTypeAssertionExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindTypeOfExpression, ast.KindSpreadAssignment, ast.KindSpreadElement, ast.KindTemplateSpan, ast.KindDeleteExpression, ast.KindVoidExpression, ast.KindAwaitExpression, ast.KindYieldExpression, ast.KindPartiallyEmittedExpression, ast.KindIfStatement, ast.KindDoStatement, ast.KindWhileStatement, ast.KindWithStatement, ast.KindForInStatement, ast.KindForOfStatement, ast.KindSwitchStatement, ast.KindCaseClause, ast.KindExpressionStatement, ast.KindReturnStatement, ast.KindThrowStatement, ast.KindExternalModuleReference, ast.KindExportAssignment, ast.KindDecorator, ast.KindJsxExpression, ast.KindJsxSpreadAttribute:
			expression = n.Expression()
		}
		if n.Kind == ast.KindTaggedTemplateExpression {
			expression = n.AsTaggedTemplateExpression().Tag
		}
		out.number(ids[expression])
		operator := ""
		switch n.Kind {
		case ast.KindBinaryExpression:
			operator = nodeKind(n.AsBinaryExpression().OperatorToken)
		case ast.KindPrefixUnaryExpression:
			operator = strings.TrimPrefix(n.AsPrefixUnaryExpression().Operator.String(), "Kind")
		}
		out.text(operator)
		var children []uint64
		n.ForEachChild(func(child *ast.Node) bool { children = append(children, ids[child]); return false })
		out.ids(children)
		var arguments []uint64
		if n.Kind == ast.KindCallExpression || n.Kind == ast.KindNewExpression {
			for _, argument := range n.Arguments() {
				arguments = append(arguments, ids[argument])
			}
		}
		out.ids(arguments)
	}
	return out.String(), nil
}
