package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

// typeProjection returns raw property, index and call-signature identities.
func (p *Program) typeProjection(c *checker.Checker, node *ast.Node, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if question == "type-projection\nheritage" {
		var clauses *ast.NodeList
		switch node.Kind {
		case ast.KindClassDeclaration:
			clauses = node.AsClassDeclaration().HeritageClauses
		case ast.KindClassExpression:
			clauses = node.AsClassExpression().HeritageClauses
		default:
			return "", fmt.Errorf("heritage projection requires a class")
		}
		g := &graph{program: p}
		var ids []uint64
		if clauses != nil {
			for _, clause := range clauses.Nodes {
				for _, typeNode := range clause.AsHeritageClause().Types.Nodes {
					ids = append(ids, g.id(c.GetTypeAtLocation(typeNode)))
				}
			}
		}
		out := &fields{}
		out.number(1)
		out.text("type-projection")
		out.ids(ids)
		return out.String(), nil
	}
	if len(parts) < 3 || len(parts) > 4 {
		return "", fmt.Errorf("type-projection requires identity and operation")
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != parts[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	subject := p.typesByID[id-1]
	g := &graph{program: p}
	out := &fields{}
	out.number(1)
	out.text("type-projection")
	switch parts[2] {
	case "property":
		if len(parts) != 4 {
			return "", fmt.Errorf("property projection requires a name")
		}
		property := checker.Checker_getPropertyOfType(c, subject, parts[3])
		member := uint64(0)
		if property != nil {
			member = g.id(c.GetTypeOfSymbolAtLocation(property, node))
		}
		out.ids([]uint64{member, g.id(checker.Checker_getIndexTypeOfType(c, subject, checker.Checker_stringType(c)))})
	case "index":
		if len(parts) != 3 {
			return "", fmt.Errorf("unexpected index suffix")
		}
		out.ids([]uint64{g.id(checker.Checker_getIndexTypeOfType(c, subject, checker.Checker_numberType(c)))})
	case "signatures":
		if len(parts) != 3 {
			return "", fmt.Errorf("unexpected signatures suffix")
		}
		signatures := c.GetSignaturesOfType(subject, checker.SignatureKindCall)
		out.number(uint64(len(signatures)))
		for _, sig := range signatures {
			out.number(g.id(c.GetReturnTypeOfSignature(sig)))
			var ids []uint64
			for _, parameter := range sig.Parameters() {
				ids = append(ids, g.id(c.GetTypeOfSymbolAtLocation(parameter, node)))
			}
			out.ids(ids)
			out.yes(sig.HasRestParameter())
		}
	default:
		return "", fmt.Errorf("unknown type projection operation")
	}
	return out.String(), nil
}
