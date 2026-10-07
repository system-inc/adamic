package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strconv"
	"strings"
)

// Raw signatures of an identity, including parameter/rest types, never a lint decision.
func (p *Program) wave19TypeSignatures(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) != 2 {
		return "", fmt.Errorf("wave19-type-signatures requires a type identity")
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || id == 0 || id > uint64(len(p.typesByID)) || strconv.FormatUint(id, 10) != parts[1] {
		return "", fmt.Errorf("unknown checker type identity")
	}
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool), allReferences: true}
	signatures := c.GetSignaturesOfType(p.typesByID[id-1], checker.SignatureKindCall)
	out.number(uint64(len(signatures)))
	var roots []uint64
	for _, signature := range signatures {
		result := g.add(c.GetReturnTypeOfSignature(signature))
		roots = append(roots, result)
		out.number(result)
		parameters := signature.Parameters()
		var ids []uint64
		for _, parameter := range parameters {
			ids = append(ids, g.add(c.GetTypeOfSymbolAtLocation(parameter, node)))
		}
		roots = append(roots, ids...)
		out.ids(ids)
		rest := false
		var element uint64
		if len(parameters) > 0 {
			declaration := parameters[0].ValueDeclaration
			rest = declaration != nil && declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().DotDotDotToken != nil
			if rest {
				subject := checker.Checker_getApparentType(c, c.GetTypeOfSymbolAtLocation(parameters[0], node))
				element = g.add(checker.Checker_getIndexTypeOfType(c, subject, checker.Checker_numberType(c)))
			}
		}
		roots = append(roots, element)
		out.yes(rest)
		out.number(element)
	}
	out.text(p.wave19ContractTypes(g, roots))
	return out.String(), nil
}
