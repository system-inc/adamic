package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Raw declared and instantiated signature identities, without contract judgments.
func (p *Program) genericSignatureShape(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "generic-signature-shape" || (node.Kind != ast.KindCallExpression && node.Kind != ast.KindNewExpression) {
		return "", fmt.Errorf("generic-signature-shape requires a call")
	}
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool), allReferences: true}
	out.number(uint64(len(node.Arguments())))
	for _, argument := range node.Arguments() {
		out.number(uint64(argument.Pos()))
		out.number(uint64(argument.End()))
		out.number(uint64(argument.Kind))
	}
	out.number(uint64(len(node.TypeArguments())))
	signature := c.GetResolvedSignature(node)
	var declared *checker.Signature
	if signature != nil {
		declared = signature.Target()
	}
	out.yes(declared != nil)
	var roots []uint64
	if declared != nil {
		ids := []uint64{}
		for _, t := range declared.TypeParameters() {
			ids = append(ids, g.add(t))
		}
		out.ids(ids)
		for _, s := range []*checker.Signature{declared, signature} {
			out.yes(s.HasRestParameter())
			ids = nil
			for _, parameter := range s.Parameters() {
				ids = append(ids, g.add(checker.Checker_getTypeOfSymbol(c, parameter)))
			}
			out.ids(ids)
			element := uint64(0)
			if s.HasRestParameter() && len(s.Parameters()) > 0 {
				t := checker.Checker_getTypeOfSymbol(c, s.Parameters()[len(s.Parameters())-1])
				element = g.add(checker.Checker_getIndexTypeOfType(c, t, checker.Checker_numberType(c)))
			}
			out.number(element)
		}
		out.number(g.add(checker.Checker_getReturnTypeOfSignature(c, declared)))
		out.number(g.add(checker.Checker_getContextualType(c, node, checker.ContextFlagsNone)))
	}
	out.yes(true)
	out.yes(true)
	out.ids(roots)
	out.number(0)
	g.write(out)
	return out.String(), nil
}
