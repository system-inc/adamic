package wave10next

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

// syntaxFields exposes AST field membership. It does not build control flow or
// classify lint behavior. Byte spans identify children in the native parse tree.
func syntaxFields(out *fields, node *ast.Node) {
	out.yes(ast.IsStatement(node))
	out.yes(ast.IsOptionalChain(node))
	out.yes(ast.IsOptionalChain(node) && ast.IsOutermostOptionalChain(node))
	out.yes(ast.IsOptionalChainRoot(node))
	roles := map[string][]*ast.Node{}
	one := func(name string, n *ast.Node) {
		if n != nil {
			roles[name] = []*ast.Node{n}
		}
	}
	list := func(name string, n *ast.NodeList) {
		if n != nil {
			roles[name] = n.Nodes
		}
	}
	one("name", node.Name())
	one("type", node.Type())
	one("body", node.Body())
	switch node.Kind {
	case ast.KindVariableDeclaration, ast.KindParameter, ast.KindBindingElement, ast.KindPropertyDeclaration, ast.KindPropertySignature, ast.KindPropertyAssignment, ast.KindEnumMember, ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement, ast.KindJsxAttribute:
		one("initializer", node.Initializer())
	}
	if node.FunctionLikeData() != nil {
		list("parameters", node.ParameterList())
		roles["typeParameters"] = node.TypeParameters()
	}
	switch node.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		roles["typeParameters"] = node.TypeParameters()
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression, ast.KindTypeReference, ast.KindExpressionWithTypeArguments, ast.KindImportType, ast.KindTypeQuery, ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement:
		roles["typeArguments"] = node.TypeArguments()
	}
	roles["decorators"] = node.Decorators()
	switch node.Kind {
	case ast.KindSourceFile:
		list("statements", node.AsSourceFile().Statements)
	case ast.KindClassStaticBlockDeclaration:
		one("body", node.AsClassStaticBlockDeclaration().Body)
	case ast.KindBlock:
		list("statements", node.AsBlock().Statements)
	case ast.KindVariableStatement:
		one("declarations", node.AsVariableStatement().DeclarationList)
	case ast.KindVariableDeclarationList:
		list("declarations", node.AsVariableDeclarationList().Declarations)
	case ast.KindExpressionStatement:
		one("expression", node.AsExpressionStatement().Expression)
	case ast.KindIfStatement:
		s := node.AsIfStatement()
		one("expression", s.Expression)
		one("then", s.ThenStatement)
		one("else", s.ElseStatement)
	case ast.KindWhileStatement:
		s := node.AsWhileStatement()
		one("expression", s.Expression)
		one("statement", s.Statement)
	case ast.KindDoStatement:
		s := node.AsDoStatement()
		one("expression", s.Expression)
		one("statement", s.Statement)
	case ast.KindForStatement:
		s := node.AsForStatement()
		one("initializer", s.Initializer)
		one("condition", s.Condition)
		one("incrementor", s.Incrementor)
		one("statement", s.Statement)
	case ast.KindForInStatement, ast.KindForOfStatement:
		s := node.AsForInOrOfStatement()
		one("initializer", s.Initializer)
		one("expression", s.Expression)
		one("statement", s.Statement)
		one("await", s.AwaitModifier)
	case ast.KindSwitchStatement:
		s := node.AsSwitchStatement()
		one("expression", s.Expression)
		one("caseBlock", s.CaseBlock)
	case ast.KindCaseBlock:
		list("clauses", node.AsCaseBlock().Clauses)
	case ast.KindCaseClause, ast.KindDefaultClause:
		s := node.AsCaseOrDefaultClause()
		one("expression", s.Expression)
		list("statements", s.Statements)
	case ast.KindTryStatement:
		s := node.AsTryStatement()
		one("try", s.TryBlock)
		one("catch", s.CatchClause)
		one("finally", s.FinallyBlock)
	case ast.KindCatchClause:
		s := node.AsCatchClause()
		one("variable", s.VariableDeclaration)
		one("body", s.Block)
	case ast.KindLabeledStatement:
		s := node.AsLabeledStatement()
		one("label", s.Label)
		one("statement", s.Statement)
	case ast.KindReturnStatement:
		one("expression", node.AsReturnStatement().Expression)
	case ast.KindThrowStatement:
		one("expression", node.AsThrowStatement().Expression)
	case ast.KindBreakStatement:
		one("label", node.AsBreakStatement().Label)
	case ast.KindContinueStatement:
		one("label", node.AsContinueStatement().Label)
	case ast.KindWithStatement:
		s := node.AsWithStatement()
		one("expression", s.Expression)
		one("statement", s.Statement)
	case ast.KindExportAssignment:
		one("expression", node.AsExportAssignment().Expression)
	case ast.KindParenthesizedExpression:
		one("expression", node.AsParenthesizedExpression().Expression)
	case ast.KindBinaryExpression:
		s := node.AsBinaryExpression()
		one("left", s.Left)
		one("right", s.Right)
		one("operator", s.OperatorToken)
	case ast.KindConditionalExpression:
		s := node.AsConditionalExpression()
		one("condition", s.Condition)
		one("then", s.WhenTrue)
		one("else", s.WhenFalse)
	case ast.KindPrefixUnaryExpression:
		one("expression", node.AsPrefixUnaryExpression().Operand)
	case ast.KindPostfixUnaryExpression:
		one("expression", node.AsPostfixUnaryExpression().Operand)
	case ast.KindPropertyAccessExpression:
		one("expression", node.AsPropertyAccessExpression().Expression)
	case ast.KindElementAccessExpression:
		s := node.AsElementAccessExpression()
		one("expression", s.Expression)
		one("argument", s.ArgumentExpression)
	case ast.KindCallExpression:
		s := node.AsCallExpression()
		one("expression", s.Expression)
		list("arguments", s.Arguments)
	case ast.KindNewExpression:
		s := node.AsNewExpression()
		one("expression", s.Expression)
		list("arguments", s.Arguments)
	case ast.KindTaggedTemplateExpression:
		s := node.AsTaggedTemplateExpression()
		one("expression", s.Tag)
		one("template", s.Template)
	case ast.KindYieldExpression:
		one("expression", node.AsYieldExpression().Expression)
	case ast.KindClassDeclaration, ast.KindClassExpression:
		s := node.ClassLikeData()
		list("heritage", s.HeritageClauses)
		list("members", s.Members)
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		list("elements", node.AsBindingPattern().Elements)
	case ast.KindBindingElement:
		s := node.AsBindingElement()
		one("propertyName", s.PropertyName)
		one("rest", s.DotDotDotToken)
	case ast.KindObjectLiteralExpression:
		list("properties", node.AsObjectLiteralExpression().Properties)
	case ast.KindArrayLiteralExpression:
		list("elements", node.AsArrayLiteralExpression().Elements)
	case ast.KindPropertyAssignment:
		one("initializer", node.AsPropertyAssignment().Initializer)
	case ast.KindShorthandPropertyAssignment:
		one("initializer", node.AsShorthandPropertyAssignment().ObjectAssignmentInitializer)
	}
	out.number(uint64(len(roles)))
	for name, nodes := range roles {
		out.text(name)
		out.number(uint64(len(nodes)))
		for _, child := range nodes {
			out.text(strings.TrimPrefix(child.Kind.String(), "Kind"))
			out.number(uint64(child.Pos()))
			out.number(uint64(child.End()))
		}
	}
}
