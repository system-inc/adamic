package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// declaredCallSignature exposes signature origins, not contextual-rule judgments.
func (p *Program) declaredCallSignature(c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "declared-call-signature" || (node.Kind != ast.KindCallExpression && node.Kind != ast.KindNewExpression) {
		return "", fmt.Errorf("declared-call-signature requires a call without suffix")
	}
	out := &fields{}
	out.number(1)
	out.text("declared-call-signature")
	resolved := c.GetResolvedSignature(node)
	var declared *checker.Signature
	if resolved != nil {
		declared = resolved.Target()
	}
	out.yes(declared != nil && len(declared.TypeParameters()) > 0)
	if declared != nil && len(declared.TypeParameters()) > 0 {
		g := &graph{program: p}
		var typeParameters, parameters, instantiated []uint64
		for _, t := range declared.TypeParameters() {
			typeParameters = append(typeParameters, g.id(t))
		}
		for _, s := range declared.Parameters() {
			parameters = append(parameters, g.id(checker.Checker_getTypeOfSymbol(c, s)))
		}
		for _, s := range resolved.Parameters() {
			instantiated = append(instantiated, g.id(checker.Checker_getTypeOfSymbol(c, s)))
		}
		out.ids(typeParameters)
		out.ids(parameters)
		out.ids(instantiated)
		out.yes(declared.HasRestParameter())
		out.number(g.id(c.GetReturnTypeOfSignature(declared)))
		out.number(g.id(checker.Checker_getContextualType(c, node, checker.ContextFlagsNone)))
	}
	return out.String(), nil
}
