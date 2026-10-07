package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	checking "github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

type Wave01Heritage struct {
	Method, StaticMember, IdentifierName, ClassParent bool
	Members                                           []int
	Expected                                          int
}

func Wave01AwaitHeritageCapture(ctx rule.Context, n *ast.Node) []Wave01Heritage {
	c := Wave01Heritage{Method: ast.IsMethodDeclaration(n), StaticMember: ast.IsStatic(n), Members: []int{}}
	name := n.Name()
	c.IdentifierName = name != nil && ast.IsIdentifier(name)
	parent := n.Parent
	c.ClassParent = parent != nil && ast.IsClassLike(parent)
	if c.IdentifierName && c.ClassParent {
		clauses := checking.GetHeritageClauses(parent)
		if clauses != nil {
			for _, clause := range clauses.Nodes {
				for _, t := range clause.AsHeritageClause().Types.Nodes {
					identity := 0
					typ := ctx.TypeChecker.GetTypeAtLocation(t)
					if typ != nil {
						member := ctx.TypeChecker.GetPropertyOfType(typ, name.Text())
						if member != nil {
							identity = int(ctx.TypeChecker.GetTypeOfSymbolAtLocation(member, n).Id())
						}
					}
					c.Members = append(c.Members, identity)
				}
			}
		}
	}
	if t := requireAwaitHeritageMemberType(ctx, n); t != nil {
		c.Expected = int(t.Id())
	}
	return []Wave01Heritage{c}
}
