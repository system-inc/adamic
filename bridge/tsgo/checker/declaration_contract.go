package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// declarationContract returns signature facts, including absence and declaration
// identity. Native callers decide whether a default can be reached.
func (p *Program) declarationContract(c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "declaration-contract" || node.Kind != ast.KindParameter || node.Parent == nil {
		return "", fmt.Errorf("declaration-contract requires a Parameter")
	}
	out := &fields{}
	out.number(1)
	out.text(question)
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool)}
	function := node.Parent
	position := -1
	for i, parameter := range function.Parameters() {
		if parameter == node {
			position = i
			break
		}
	}
	var roots []uint64
	own := c.GetSignatureFromDeclaration(function)
	ownType := (*checker.Type)(nil)
	if own != nil && position >= 0 && position < len(checker.Signature_parameters(own)) {
		ownType = checker.Checker_getTypeOfSymbol(c, checker.Signature_parameters(own)[position])
	}
	out.number(g.add(ownType))
	contextual := checker.Checker_getContextualType(c, function, checker.ContextFlagsNone)
	var signatures []*checker.Signature
	if contextual != nil {
		signatures = c.GetSignaturesOfType(contextual, checker.SignatureKindCall)
	}
	out.number(uint64(len(signatures)))
	for _, signature := range signatures {
		out.yes(checker.Signature_declaration(signature) == function)
		parameters := checker.Signature_parameters(signature)
		present := position >= 0 && position < len(parameters) && parameters[position] != nil
		out.yes(present)
		var flags uint64
		rest := false
		t := checker.Checker_undefinedType(c)
		if present {
			parameter := parameters[position]
			flags = uint64(parameter.Flags)
			declaration := parameter.ValueDeclaration
			rest = declaration != nil && declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().DotDotDotToken != nil
			t = checker.Checker_getTypeOfSymbol(c, parameter)
		}
		out.number(flags)
		out.yes(rest)
		roots = append(roots, g.add(t))
	}
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(true)
	out.ids(roots)
	out.number(0)
	g.write(out)
	return out.String(), nil
}
