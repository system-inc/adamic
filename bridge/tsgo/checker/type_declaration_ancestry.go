package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Type declaration ancestry returns original AST locations through every ancestor.
// The consumer decides which declarations, aliases, paths and arm counts matter.
func (p *Program) typeDeclarationAncestry(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "type-declaration-ancestry" && question != "type-declaration-ancestry\nawaited" {
		return "", fmt.Errorf("unexpected type-declaration-ancestry suffix")
	}
	t := c.GetTypeAtLocation(node)
	if t != nil && strings.HasSuffix(question, "\nawaited") {
		t = checker.Checker_getAwaitedType(c, t)
	}
	var parts []*checker.Type
	if t != nil {
		parts = []*checker.Type{t}
		if t.Flags()&checker.TypeFlagsUnion != 0 {
			parts = t.Types()
		}
	}
	out.number(uint64(len(parts)))
	for _, part := range parts {
		symbol := part.Symbol()
		var declarations []*ast.Node
		if symbol != nil {
			declarations = symbol.Declarations
		}
		out.number(uint64(len(declarations)))
		for _, declaration := range declarations {
			file := ast.GetSourceFileOfNode(declaration)
			if file == nil {
				return "", fmt.Errorf("type declaration has no source")
			}
			out.text(file.FileName())
			var chain []*ast.Node
			for current := declaration; current != nil; current = current.Parent {
				chain = append(chain, current)
			}
			out.number(uint64(len(chain)))
			for _, current := range chain {
				out.text(strings.TrimPrefix(current.Kind.String(), "Kind"))
				out.number(uint64(current.Pos()))
				out.number(uint64(current.End()))
				name := ""
				if current.Name() != nil {
					name = current.Name().Text()
				}
				out.text(name)
				var child *ast.Node
				switch current.Kind {
				case ast.KindTypeAliasDeclaration:
					child = current.AsTypeAliasDeclaration().Type
				case ast.KindParenthesizedType:
					child = current.AsParenthesizedTypeNode().Type
				}
				out.yes(child != nil)
				if child != nil {
					out.number(uint64(child.Pos()))
					out.number(uint64(child.End()))
				}
				count := 0
				current.ForEachChild(func(*ast.Node) bool { count++; return false })
				out.number(uint64(count))
			}
		}
	}
	return out.String(), nil
}
