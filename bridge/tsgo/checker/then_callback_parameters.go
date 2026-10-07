package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Raw apparent callback parameter types of the receiver's then signatures.
// The native rule decides whether any of these types is callable.
func (p *Program) thenCallbackParameters(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "then-callback-parameters" {
		return "", fmt.Errorf("unexpected then-callback-parameters suffix")
	}
	parts := func(t *checker.Type) []*checker.Type {
		if t.Flags()&checker.TypeFlagsUnion != 0 {
			return t.Types()
		}
		return []*checker.Type{t}
	}
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool)}
	var roots []uint64
	for _, part := range parts(checker.Checker_getApparentType(c, c.GetTypeAtLocation(node))) {
		property := checker.Checker_getPropertyOfType(c, part, "then")
		if property == nil {
			continue
		}
		for _, then := range parts(c.GetTypeOfSymbolAtLocation(property, node)) {
			for _, signature := range c.GetSignaturesOfType(then, checker.SignatureKindCall) {
				parameters := checker.Signature_parameters(signature)
				if len(parameters) == 0 {
					continue
				}
				parameter := parameters[0]
				t := checker.Checker_getApparentType(c, c.GetTypeOfSymbolAtLocation(parameter, node))
				if parameter.ValueDeclaration != nil && ast.IsParameterDeclaration(parameter.ValueDeclaration) && parameter.ValueDeclaration.AsParameterDeclaration().DotDotDotToken != nil {
					t = checker.Checker_getIndexTypeOfType(c, t, checker.Checker_numberType(c))
				}
				if t != nil {
					roots = append(roots, g.add(t))
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
