package checker

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// The checker's awaited type graph; no declaration or lint interpretation.
func (p *Program) awaitedTypeShape(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "awaited-type-shape" {
		return "", fmt.Errorf("unexpected awaited type suffix")
	}
	value := c.GetTypeAtLocation(node)
	if value != nil {
		value = checker.Checker_getAwaitedType(c, value)
	}
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool)}
	var roots []uint64
	if value != nil {
		roots = append(roots, g.add(value))
	}
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(value != nil)
	out.ids(roots)
	out.number(0)
	g.write(out)
	return out.String(), nil
}
