package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Exposes declared and instantiated signatures separately, without choosing a demand.
func (p *Program) wave19GenericCall(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "wave19-generic-call" || (node.Kind != ast.KindCallExpression && node.Kind != ast.KindNewExpression) {
		return "", fmt.Errorf("wave19-generic-call requires a call or new expression")
	}
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool), allReferences: true}
	var roots []uint64
	resolved := c.GetResolvedSignature(node)
	out.yes(len(node.TypeArguments()) != 0)
	var declared *checker.Signature
	if resolved != nil {
		declared = resolved.Target()
	}
	var parameters, declaredTypes, resolvedTypes []uint64
	var rest, declaredReturn uint64
	hasRest := false
	if declared != nil {
		for _, t := range declared.TypeParameters() {
			parameters = append(parameters, g.add(t))
		}
		for _, parameter := range declared.Parameters() {
			declaredTypes = append(declaredTypes, g.add(checker.Checker_getTypeOfSymbol(c, parameter)))
		}
		hasRest = declared.HasRestParameter()
		if hasRest && len(declared.Parameters()) > 0 {
			t := checker.Checker_getTypeOfSymbol(c, declared.Parameters()[len(declared.Parameters())-1])
			rest = g.add(checker.Checker_getIndexTypeOfType(c, t, checker.Checker_numberType(c)))
		}
		declaredReturn = g.add(c.GetReturnTypeOfSignature(declared))
	}
	if resolved != nil {
		for _, parameter := range resolved.Parameters() {
			resolvedTypes = append(resolvedTypes, g.add(checker.Checker_getTypeOfSymbol(c, parameter)))
		}
	}
	contextual := g.add(checker.Checker_getContextualType(c, node, checker.ContextFlagsNone))
	roots = append(roots, parameters...)
	roots = append(roots, declaredTypes...)
	roots = append(roots, resolvedTypes...)
	roots = append(roots, rest, declaredReturn, contextual)
	out.ids(parameters)
	out.ids(declaredTypes)
	out.yes(hasRest)
	out.number(rest)
	out.ids(resolvedTypes)
	out.number(declaredReturn)
	out.number(contextual)
	out.text(p.wave19ContractTypes(g, roots))
	return out.String(), nil
}
