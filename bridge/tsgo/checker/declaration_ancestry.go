package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

// Raw ancestor locations and syntax, including the written type of an alias.
// No paths, names or kinds are filtered for a lint rule.
func (p *Program) declarationAncestry(out *fields, question string) error {
	parts := strings.Split(question, "\n")
	if len(parts) != 2 {
		return fmt.Errorf("declaration-ancestry requires a type identity")
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != parts[1] {
		return fmt.Errorf("unknown checker type identity")
	}
	symbol := p.typesByID[id-1].Symbol()
	var declarations []*ast.Node
	if symbol != nil {
		declarations = symbol.Declarations
	}
	out.number(uint64(len(declarations)))
	for _, declaration := range declarations {
		var chain []*ast.Node
		for n := declaration; n != nil; n = n.Parent {
			chain = append(chain, n)
		}
		out.number(uint64(len(chain)))
		for _, n := range chain {
			out.text(strings.TrimPrefix(n.Kind.String(), "Kind"))
			name := ""
			if named := n.Name(); named != nil {
				name = named.Text()
			}
			out.text(name)
			out.number(uint64(n.Pos()))
			out.number(uint64(n.End()))
			path := ""
			if source := ast.GetSourceFileOfNode(n); source != nil {
				path = string(source.FileName())
			}
			out.text(path)
			count := 0
			n.ForEachChild(func(child *ast.Node) bool { count++; return false })
			out.number(uint64(count))
			first, last := uint64(0), uint64(0)
			if n.Kind == ast.KindTypeAliasDeclaration && n.AsTypeAliasDeclaration().Type != nil {
				written := n.AsTypeAliasDeclaration().Type
				first, last = uint64(written.Pos()), uint64(written.End())
			}
			out.number(first)
			out.number(last)
		}
	}
	return nil
}
func (p *Program) inspectDeclarationAncestry(out *fields, c *checker.Checker, node *ast.Node, question string) (string, bool, error) {
	switch strings.Split(question, "\n")[0] {
	case "declaration-ancestry":
		if err := p.declarationAncestry(out, question); err != nil {
			return "", true, err
		}
		return out.String(), true, nil
	case "awaited-shape":
		result, err := p.awaitedShape(out, c, node, question)
		return result, true, err
	}
	return "", false, nil
}

func (p *Program) inspectContinuation(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if result, handled, err := p.inspectDeclarationAncestry(out, c, node, question); handled {
		return result, err
	}
	return "", fmt.Errorf("unsupported checker question: %s", question)
}
