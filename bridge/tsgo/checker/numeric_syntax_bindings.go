package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// Numeric parser nodes and raw symbol declaration syntax. No lint predicates,
// React inference, scope classification, or diagnostic decisions live here.
func (p *Program) numericSyntaxBindings(out *fields, c *checker.Checker, root *ast.Node, question string) (string, error) {
	if question != "numeric-syntax-bindings" || root.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("numeric-syntax-bindings requires a SourceFile")
	}
	type binding struct {
		node        *ast.Node
		initializer *ast.Node
	}
	nodes := []*ast.Node{}
	ids := map[*ast.Node]uint64{}
	children := map[*ast.Node][]*ast.Node{}
	var add func(*ast.Node) uint64
	add = func(node *ast.Node) uint64 {
		if node == nil {
			return 0
		}
		if id := ids[node]; id != 0 {
			return id
		}
		id := uint64(len(nodes) + 1)
		ids[node] = id
		nodes = append(nodes, node)
		node.ForEachChild(func(child *ast.Node) bool { children[node] = append(children[node], child); add(child); return false })
		return id
	}
	add(root)
	localCount := len(nodes)
	declarations := map[*ast.Node][]binding{}
	hasSymbol := map[*ast.Node]bool{}
	for _, node := range nodes[:localCount] {
		if node.Kind != ast.KindIdentifier {
			continue
		}
		symbol := c.GetSymbolAtLocation(node)
		if symbol == nil {
			continue
		}
		hasSymbol[node] = true
		for _, declaration := range symbol.Declarations {
			var initializer *ast.Node
			if declaration.Kind == ast.KindVariableDeclaration {
				initializer = declaration.AsVariableDeclaration().Initializer
			}
			if declaration.Kind == ast.KindBindingElement && declaration.Parent != nil && declaration.Parent.Parent != nil && declaration.Parent.Parent.Kind == ast.KindVariableDeclaration {
				initializer = declaration.Parent.Parent.AsVariableDeclaration().Initializer
			}
			add(initializer)
			declarations[node] = append(declarations[node], binding{declaration, initializer})
		}
	}
	list := func(items []*ast.Node) {
		out.number(uint64(len(items)))
		for _, node := range items {
			out.number(ids[node])
		}
	}
	kind := func(node *ast.Node) uint64 {
		if node == nil {
			return 0
		}
		return uint64(node.Kind)
	}
	text := func(node *ast.Node) string {
		if node == nil {
			return ""
		}
		switch node.Kind {
		case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindNoSubstitutionTemplateLiteral:
			return node.Text()
		}
		return ""
	}
	out.number(uint64(len(nodes)))
	for _, node := range nodes {
		source := ast.GetSourceFileOfNode(node)
		if source == nil {
			return "", fmt.Errorf("syntax node has no source")
		}
		out.number(uint64(node.Kind))
		out.number(uint64(max(0, node.Pos())))
		out.number(uint64(max(0, node.End())))
		out.number(uint64(max(0, scanner.GetRangeOfTokenAtPosition(source, max(0, node.Pos())).Pos())))
		out.text(text(node))
		out.number(ids[node.Parent])
		list(children[node])
		out.number(ids[node.Name()])
		var expression, initializer, property, opening, tag, attributes, left, right *ast.Node
		var arguments, parameters []*ast.Node
		operator := ast.KindUnknown
		rest := false
		switch node.Kind {
		case ast.KindVariableDeclaration, ast.KindParameter, ast.KindBindingElement, ast.KindPropertyAssignment, ast.KindJsxAttribute:
			initializer = node.Initializer()
		}
		if ast.IsFunctionLike(node) && node.ParameterList() != nil {
			parameters = node.ParameterList().Nodes
		}
		switch node.Kind {
		case ast.KindCallExpression:
			call := node.AsCallExpression()
			expression = call.Expression
			if call.Arguments != nil {
				arguments = call.Arguments.Nodes
			}
		case ast.KindPropertyAccessExpression:
			expression = node.AsPropertyAccessExpression().Expression
		case ast.KindBinaryExpression:
			b := node.AsBinaryExpression()
			left = b.Left
			right = b.Right
			operator = b.OperatorToken.Kind
		case ast.KindTemplateSpan:
			expression = node.AsTemplateSpan().Expression
		case ast.KindParenthesizedExpression:
			expression = node.AsParenthesizedExpression().Expression
		case ast.KindJsxExpression:
			expression = node.AsJsxExpression().Expression
		case ast.KindJsxElement:
			opening = node.AsJsxElement().OpeningElement
		case ast.KindJsxOpeningElement:
			j := node.AsJsxOpeningElement()
			tag = j.TagName
			attributes = j.Attributes
		case ast.KindJsxSelfClosingElement:
			j := node.AsJsxSelfClosingElement()
			tag = j.TagName
			attributes = j.Attributes
		case ast.KindImportSpecifier:
			property = node.AsImportSpecifier().PropertyName
		case ast.KindParameter:
			rest = node.AsParameterDeclaration().DotDotDotToken != nil
		}
		for _, value := range []*ast.Node{initializer, expression, property, opening, tag, attributes, left, right} {
			out.number(ids[value])
		}
		out.number(uint64(operator))
		out.yes(rest)
		list(arguments)
		list(parameters)
		out.yes(hasSymbol[node])
		out.number(uint64(len(declarations[node])))
		for _, entry := range declarations[node] {
			declaration := entry.node
			source := ast.GetSourceFileOfNode(declaration)
			out.number(uint64(declaration.Kind))
			out.text(source.FileName().AsString())
			out.number(ids[entry.initializer])
			imported := declaration.Name()
			if declaration.Kind == ast.KindImportSpecifier && declaration.AsImportSpecifier().PropertyName != nil {
				imported = declaration.AsImportSpecifier().PropertyName
			}
			out.number(kind(imported))
			out.text(text(imported))
			parent := declaration.Parent
			var grand, great *ast.Node
			if parent != nil {
				grand = parent.Parent
			}
			if grand != nil {
				great = grand.Parent
			}
			out.number(kind(parent))
			out.number(kind(grand))
			out.number(kind(great))
			var module *ast.Node
			for owner := declaration; owner != nil; owner = owner.Parent {
				if owner.Kind == ast.KindImportDeclaration {
					module = owner.AsImportDeclaration().ModuleSpecifier
					break
				}
			}
			out.number(kind(module))
			out.text(text(module))
		}
	}
	return out.String(), nil
}
