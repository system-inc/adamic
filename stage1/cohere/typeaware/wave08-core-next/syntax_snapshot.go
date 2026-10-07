package wave08core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
)

// SyntaxSnapshotFields returns raw syntax fields and identifier symbol identities.
// It builds no control-flow graph and computes no escape or lint verdict.
func SyntaxSnapshotFields(c *checker.Checker, node *ast.Node, id func(*ast.Symbol) uint64) []string {
	var nodes []*ast.Node
	var walk func(*ast.Node) bool
	walk = func(n *ast.Node) bool { nodes = append(nodes, n); n.ForEachChild(walk); return false }
	walk(node)
	out := []string{"1", "wave08-syntax-snapshot", strconv.Itoa(len(nodes))}
	symbols := map[*ast.Symbol]bool{}
	var ordered []*ast.Symbol
	number := func(n int) { out = append(out, strconv.Itoa(n)) }
	span := func(n *ast.Node) {
		if n == nil {
			number(-1)
			number(-1)
		} else {
			number(n.Pos())
			number(n.End())
		}
	}
	for _, n := range nodes {
		number(int(n.Kind))
		span(n)
		span(n.Name())
		var initializer, body *ast.Node
		switch n.Kind {
		case ast.KindVariableDeclaration, ast.KindParameter, ast.KindBindingElement, ast.KindPropertyDeclaration, ast.KindPropertySignature, ast.KindPropertyAssignment, ast.KindEnumMember, ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement, ast.KindJsxAttribute:
			initializer = n.Initializer()
		}
		if ast.IsFunctionLike(n) {
			body = n.Body()
		}
		span(initializer)
		span(body)
		number(int(n.Flags))
		number(int(n.ModifierFlags()))
		ff := ast.FunctionFlagsNormal
		if ast.IsFunctionLike(n) {
			ff = ast.GetFunctionFlags(n)
		}
		number(int(ff))
		if ast.IsFunctionLike(n) && n.Type() != nil {
			number(1)
		} else {
			number(0)
		}
		var s *ast.Symbol
		if n.Kind == ast.KindIdentifier {
			s = c.GetSymbolAtLocation(n)
			if n.Parent != nil && n.Parent.Kind == ast.KindShorthandPropertyAssignment {
				if v := c.GetShorthandAssignmentValueSymbol(n.Parent); v != nil {
					s = v
				}
			}
		}
		out = append(out, strconv.FormatUint(id(s), 10))
		if s != nil && !symbols[s] {
			symbols[s] = true
			ordered = append(ordered, s)
		}
	}
	number(len(ordered))
	for _, s := range ordered {
		out = append(out, strconv.FormatUint(id(s), 10), s.Name)
		number(len(s.Declarations))
		for _, decl := range s.Declarations {
			f := ast.GetSourceFileOfNode(decl)
			path := ""
			if f != nil {
				path = f.FileName()
			}
			out = append(out, path)
			number(int(decl.Kind))
			span(decl)
		}
	}
	return out
}
