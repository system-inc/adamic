package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// The awaited type of an expression, without choosing or evaluating a lint rule.
func (p *Program) awaitedShape(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "awaited-shape" {
		return "", fmt.Errorf("unexpected awaited-shape suffix")
	}
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool)}
	t := c.GetTypeAtLocation(node)
	if t != nil {
		t = checker.Checker_getAwaitedType(c, t)
	}
	var roots []uint64
	if t != nil {
		roots = append(roots, g.add(t))
	}
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(t != nil)
	out.ids(roots)
	out.number(0)
	g.write(out)
	return out.String(), nil
}
