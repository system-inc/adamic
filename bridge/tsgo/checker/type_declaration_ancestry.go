// Raw declaration ancestry, including the type reached by awaiting a value.
package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (p *Program) typeDeclarationAncestry(c *checker.Checker, node *ast.Node, question string) (string, error) {
	if strings.Split(question, "\n")[0] != "type-declaration-ancestry" {
		return p.processQuestions(c, node, question)
	}
	split := strings.Split(question, "\n")
	if len(split) != 2 || (split[1] != "raw" && split[1] != "awaited") {
		return "", fmt.Errorf("type-declaration-ancestry requires raw or awaited")
	}
	value := c.GetTypeAtLocation(node)
	if value != nil && split[1] == "awaited" {
		value = checker.Checker_getAwaitedType(c, value)
	}
	var parts []*checker.Type
	if value != nil {
		parts = []*checker.Type{value}
		if value.Flags()&checker.TypeFlagsUnion != 0 {
			parts = value.Types()
		}
	}
	identities := map[*ast.Node]uint64{}
	var nodes []*ast.Node
	var add func(*ast.Node) uint64
	add = func(n *ast.Node) uint64 {
		if n == nil {
			return 0
		}
		if id := identities[n]; id != 0 {
			return id
		}
		id := uint64(len(nodes) + 1)
		identities[n] = id
		nodes = append(nodes, n)
		add(n.Parent)
		if n.Kind == ast.KindTypeAliasDeclaration {
			add(n.AsTypeAliasDeclaration().Type)
		}
		if n.Kind == ast.KindParenthesizedType {
			add(n.AsParenthesizedTypeNode().Type)
		}
		return id
	}
	var declarations []uint64
	for _, part := range parts {
		if symbol := part.Symbol(); symbol != nil {
			for _, declaration := range symbol.Declarations {
				declarations = append(declarations, add(declaration))
			}
		}
	}
	out := &fields{}
	out.number(1)
	out.text("type-declaration-ancestry")
	out.ids(declarations)
	out.number(uint64(len(nodes)))
	for _, n := range nodes {
		out.number(identities[n])
		out.number(identities[n.Parent])
		out.text(strings.TrimPrefix(n.Kind.String(), "Kind"))
		name := ""
		if named := n.Name(); named != nil {
			name = named.Text()
		}
		out.text(name)
		file := ""
		if source := ast.GetSourceFileOfNode(n); source != nil {
			file = source.FileName()
		}
		out.text(file)
		var typed *ast.Node
		if n.Kind == ast.KindTypeAliasDeclaration {
			typed = n.AsTypeAliasDeclaration().Type
		}
		if n.Kind == ast.KindParenthesizedType {
			typed = n.AsParenthesizedTypeNode().Type
		}
		out.number(identities[typed])
		count := 0
		if n.Kind == ast.KindUnionType {
			count = len(n.AsUnionTypeNode().Types.Nodes)
		}
		out.number(uint64(count))
	}
	return out.String(), nil
}
