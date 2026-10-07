package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Generic target and instantiated signature types are raw checker facts.
// Native code decides which argument positions demand promises.
func (p *Program) genericCallSignature(out *fields, c *checker.Checker, node *ast.Node, question string) error {
	if question != "generic-call-signature" || (node.Kind != ast.KindCallExpression && node.Kind != ast.KindNewExpression) {
		return fmt.Errorf("generic-call-signature requires a call or new expression")
	}
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool)}
	resolved := c.GetResolvedSignature(node)
	var declared *checker.Signature
	if resolved != nil {
		declared = resolved.Target()
	}
	out.yes(declared != nil)
	var roots []uint64
	if declared != nil {
		var parameters []uint64
		for _, t := range declared.TypeParameters() {
			parameters = append(parameters, g.add(t))
		}
		out.ids(parameters)
		out.yes(declared.HasRestParameter())
		for _, sig := range []*checker.Signature{declared, resolved} {
			var ids []uint64
			for _, symbol := range sig.Parameters() {
				ids = append(ids, g.add(checker.Checker_getTypeOfSymbol(c, symbol)))
			}
			out.ids(ids)
		}
		roots = append(roots, g.add(checker.Checker_getReturnTypeOfSignature(c, declared)))
	}
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(declared != nil)
	out.ids(roots)
	out.number(0)
	g.write(out)
	return nil
}
