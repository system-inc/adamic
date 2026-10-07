package checker

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// handlerParameterTypes preserves raw first-parameter and first type-argument
// flags, including signatures with empty tuple rests and non-array rest types.
func (p *Program) handlerParameterTypes(out *fields, c *checker.Checker, node *ast.Node, question string) error {
	if question != "handler-parameter-types" || (node.Kind != ast.KindArrowFunction && node.Kind != ast.KindFunctionExpression) {
		return fmt.Errorf("handler-parameter-types requires a function expression")
	}
	var signatures []*checker.Signature
	for _, part := range wave24UnionMembers(c.GetTypeAtLocation(node)) {
		signatures = append(signatures, c.GetSignaturesOfType(part, checker.SignatureKindCall)...)
	}
	out.number(uint64(len(signatures)))
	for _, signature := range signatures {
		parameters := checker.Signature_parameters(signature)
		out.number(uint64(len(parameters)))
		var t *checker.Type
		rest := false
		if len(parameters) > 0 {
			parameter := parameters[0]
			t = checker.Checker_getTypeOfSymbol(c, parameter)
			d := parameter.ValueDeclaration
			rest = d != nil && d.Kind == ast.KindParameter && d.AsParameterDeclaration().DotDotDotToken != nil
		}
		out.yes(t != nil)
		flags := uint64(0)
		arrayOrTuple := false
		var arguments []*checker.Type
		if t != nil {
			flags = uint64(t.Flags())
			arrayOrTuple = checker.Checker_isArrayType(c, t) || checker.IsTupleType(t)
			if arrayOrTuple {
				arguments = checker.Checker_getTypeArguments(c, t)
			}
		}
		out.number(flags)
		out.yes(rest)
		out.yes(arrayOrTuple)
		out.number(uint64(len(arguments)))
		first := uint64(0)
		if len(arguments) > 0 {
			first = uint64(arguments[0].Flags())
		}
		out.number(first)
	}
	return nil
}
