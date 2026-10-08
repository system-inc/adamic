package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// symbolDeclarationSyntax supplies raw declaration trees and ancestors. It does
// not classify React bindings or decide whether any rule should report.
func (p *Program) symbolDeclarationSyntax(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "symbol-declaration-syntax" {
		return "", fmt.Errorf("invalid symbol-declaration-syntax question")
	}
	symbol := c.GetSymbolAtLocation(node)
	out.yes(symbol != nil)
	var nodes []*ast.Node
	ids := map[*ast.Node]int{}
	var add func(*ast.Node)
	add = func(n *ast.Node) {
		if n == nil {
			return
		}
		if _, found := ids[n]; found {
			return
		}
		ids[n] = len(nodes)
		nodes = append(nodes, n)
		n.ForEachChild(func(child *ast.Node) bool { add(child); return false })
	}
	if symbol != nil {
		for _, declaration := range symbol.Declarations {
			root := declaration
			for depth := 0; depth < 3 && root.Parent != nil && root.Parent.Kind != ast.KindSourceFile; depth++ {
				root = root.Parent
			}
			add(root)
		}
	}
	count := 0
	if symbol != nil {
		count = len(symbol.Declarations)
	}
	out.number(uint64(count))
	if symbol != nil {
		for _, declaration := range symbol.Declarations {
			out.number(uint64(ids[declaration]))
		}
	}
	out.number(uint64(len(nodes)))
	for _, n := range nodes {
		source := ast.GetSourceFileOfNode(n)
		file := ""
		if source != nil {
			file = source.FileName().AsString()
		}
		out.text(file)
		out.number(uint64(n.Kind))
		out.text(strings.TrimPrefix(n.Kind.String(), "Kind"))
		out.number(uint64(n.Pos()))
		out.number(uint64(n.End()))
		text := ""
		switch n.Kind {
		case ast.KindIdentifier, ast.KindPrivateIdentifier, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral,
			ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindRegularExpressionLiteral:
			text = n.Text()
		}
		out.text(text)
		parent := uint64(0)
		if id, found := ids[n.Parent]; found {
			parent = uint64(id + 1)
		}
		out.number(parent)
		list := uint64(0)
		if n.Kind == ast.KindCallExpression && n.AsCallExpression().Arguments != nil {
			list = uint64(len(n.AsCallExpression().Arguments.Nodes))
		}
		out.number(list)
		initializer := uint64(0)
		if n.Kind == ast.KindVariableDeclaration && n.AsVariableDeclaration().Initializer != nil {
			initializer = uint64(ids[n.AsVariableDeclaration().Initializer] + 1)
		}
		out.number(initializer)
		var children []uint64
		n.ForEachChild(func(child *ast.Node) bool { children = append(children, uint64(ids[child])); return false })
		out.ids(children)
	}
	return out.String(), nil
}
