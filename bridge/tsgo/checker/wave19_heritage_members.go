package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Matching member types in source heritage order. Static/contract decisions stay native.
func (p *Program) wave19HeritageMembers(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "wave19-heritage-members" || node.Kind != ast.KindMethodDeclaration || node.Name() == nil || node.Name().Kind != ast.KindIdentifier || node.Parent == nil || !ast.IsClassLike(node.Parent) {
		return "", fmt.Errorf("wave19-heritage-members requires a named class method")
	}
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool), allReferences: true}
	var roots []uint64
	var clauses *ast.NodeList
	if node.Parent.Kind == ast.KindClassDeclaration {
		clauses = node.Parent.AsClassDeclaration().HeritageClauses
	} else {
		clauses = node.Parent.AsClassExpression().HeritageClauses
	}
	if clauses != nil {
		for _, clause := range clauses.Nodes {
			for _, typeNode := range clause.AsHeritageClause().Types.Nodes {
				subject := c.GetTypeAtLocation(typeNode)
				if subject == nil {
					continue
				}
				property := checker.Checker_getPropertyOfType(c, subject, node.Name().Text())
				if property != nil {
					roots = append(roots, g.add(c.GetTypeOfSymbolAtLocation(property, node)))
				}
			}
		}
	}
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(true)
	out.ids(roots)
	out.number(0)
	g.write(out)
	return out.String(), nil
}
