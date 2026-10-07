package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// nodeStructure returns syntax, including JSX unavailable in the Adamic parser.
// All decisions about what is a listener, mock or render stay in Adamic.
func (p *Program) nodeStructure(out *fields, source *ast.SourceFile, node *ast.Node, question string) (string, error) {
	if question != "node-structure" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("node-structure requires a SourceFile")
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
		if ast.IsIdentifier(n) || n.Kind == ast.KindStringLiteral {
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
		case ast.KindPropertyAccessExpression:
			expression = n.AsPropertyAccessExpression().Expression
			name = n.Name()
		case ast.KindCallExpression:
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
			if n.Kind == ast.KindVariableDeclaration {
				expression = n.AsVariableDeclaration().Initializer
			}
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
	}
	return out.String(), nil
}
