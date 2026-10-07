package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// awaitedShape exposes the checker's awaited type without any lint policy.
func (p *Program) awaitedShape(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "awaited-shape" {
		return "", fmt.Errorf("unexpected awaited-shape suffix")
	}
	t := c.GetTypeAtLocation(node)
	if t != nil {
		t = checker.Checker_getAwaitedType(c, t)
	}
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool)}
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(t != nil)
	var roots []uint64
	if t != nil {
		roots = append(roots, g.add(t))
	}
	out.ids(roots)
	out.number(0)
	g.write(out)
	return out.String(), nil
}
