package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Heritage types retain the source clause order; no member or contract is selected.
func (p *Program) heritageTypes(out *fields, c *checker.Checker, node *ast.Node, question string) error {
	if question != "heritage-types" || !ast.IsClassLike(node) {
		return fmt.Errorf("heritage-types requires a class")
	}
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool)}
	var roots []uint64
	var clauses *ast.NodeList
	if node.Kind == ast.KindClassDeclaration {
		clauses = node.AsClassDeclaration().HeritageClauses
	} else {
		clauses = node.AsClassExpression().HeritageClauses
	}
	if clauses != nil {
		for _, part := range clauses.Nodes {
			for _, t := range part.AsHeritageClause().Types.Nodes {
				if value := c.GetTypeAtLocation(t); value != nil {
					roots = append(roots, g.add(value))
				}
			}
		}
	}
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(true)
	out.ids(roots)
	out.number(0)
	g.write(out)
	return nil
}
