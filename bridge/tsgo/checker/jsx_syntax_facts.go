package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"unicode/utf16"
)

// Raw numeric syntax, source ranges and symbol declarations. No lint judgments.
func (p *Program) jsxSyntaxFacts(out *fields, c *checker.Checker, source *ast.SourceFile, node *ast.Node, question string) error {
	if question != "jsx-syntax-facts" || node != source.AsNode() {
		return fmt.Errorf("jsx-syntax-facts requires a source root")
	}
	nodes := []*ast.Node{nil}
	ids := map[*ast.Node]uint64{nil: 0}
	add := func(n *ast.Node) uint64 {
		if id, ok := ids[n]; ok {
			return id
		}
		id := uint64(len(nodes))
		ids[n] = id
		nodes = append(nodes, n)
		return id
	}
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		switch n.Kind {
		case ast.KindJsxElement, ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment, ast.KindCallExpression:
			add(n)
		}
		n.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(node)
	rootCount := len(nodes)
	records := &fields{}
	for at := 1; at < len(nodes); at++ {
		n := nodes[at]
		file := ast.GetSourceFileOfNode(n)
		records.number(uint64(n.Kind))
		records.number(uint64(scanner.SkipTrivia(file.Text(), n.Pos())))
		records.number(uint64(n.End()))
		records.text(string(file.FileName()))
		records.number(add(n.Parent))
		text := ""
		switch n.Kind {
		case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindRegularExpressionLiteral:
			text = n.Text()
		}
		units := utf16.Encode([]rune(text))
		records.number(uint64(len(units)))
		for _, u := range units {
			records.number(uint64(u))
		}
		var expression, initializer, opening, attributes, module, property *ast.Node
		var items, arguments []*ast.Node
		switch n.Kind {
		case ast.KindPropertyAccessExpression:
			expression = n.AsPropertyAccessExpression().Expression
		case ast.KindCallExpression:
			v := n.AsCallExpression()
			expression = v.Expression
			if v.Arguments != nil {
				arguments = v.Arguments.Nodes
			}
		case ast.KindParenthesizedExpression:
			expression = n.AsParenthesizedExpression().Expression
		case ast.KindVariableDeclaration:
			initializer = n.AsVariableDeclaration().Initializer
		case ast.KindBindingElement:
			property = n.AsBindingElement().PropertyName
		case ast.KindImportSpecifier:
			property = n.AsImportSpecifier().PropertyName
		case ast.KindImportDeclaration:
			module = n.AsImportDeclaration().ModuleSpecifier
		case ast.KindJsxElement:
			v := n.AsJsxElement()
			opening = v.OpeningElement
			if v.Children != nil {
				items = v.Children.Nodes
			}
		case ast.KindJsxOpeningElement:
			v := n.AsJsxOpeningElement()
			expression = v.TagName
			attributes = v.Attributes
		case ast.KindJsxSelfClosingElement:
			v := n.AsJsxSelfClosingElement()
			expression = v.TagName
			attributes = v.Attributes
		case ast.KindJsxAttributes:
			v := n.AsJsxAttributes()
			if v.Properties != nil {
				items = v.Properties.Nodes
			}
		case ast.KindArrayLiteralExpression:
			v := n.AsArrayLiteralExpression()
			if v.Elements != nil {
				items = v.Elements.Nodes
			}
		}
		for _, v := range []*ast.Node{n.Name(), expression, initializer, opening, attributes, module, property} {
			records.number(add(v))
		}
		records.number(uint64(len(items)))
		for _, v := range items {
			records.number(add(v))
		}
		records.number(uint64(len(arguments)))
		for _, v := range arguments {
			records.number(add(v))
		}
		var symbol *ast.Symbol
		if n.Kind == ast.KindIdentifier {
			ancestor := n.Parent
			for ancestor != nil && ancestor.Kind == ast.KindPropertyAccessExpression {
				ancestor = ancestor.Parent
			}
			if n.Text() == "createElement" || (ancestor != nil && (ancestor.Kind == ast.KindJsxOpeningElement || ancestor.Kind == ast.KindJsxSelfClosingElement)) {
				symbol = c.GetSymbolAtLocation(n)
			}
		}
		records.yes(symbol != nil)
		if symbol == nil {
			records.number(0)
		} else {
			records.number(uint64(len(symbol.Declarations)))
			for _, v := range symbol.Declarations {
				records.number(add(v))
			}
		}
	}
	out.number(uint64(rootCount))
	out.number(uint64(len(nodes)))
	out.WriteString(records.String())
	return nil
}
