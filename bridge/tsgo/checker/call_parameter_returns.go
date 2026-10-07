package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// callParameterReturns exposes parameter callback returns for every overload.
func (p *Program) callParameterReturns(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "call-parameter-returns" {
		return "", fmt.Errorf("unexpected call-parameter-returns suffix")
	}
	var callee *ast.Node
	var arguments *ast.NodeList
	kind := checker.SignatureKindCall
	switch node.Kind {
	case ast.KindCallExpression:
		callee = node.AsCallExpression().Expression
		arguments = node.AsCallExpression().Arguments
	case ast.KindNewExpression:
		callee = node.AsNewExpression().Expression
		arguments = node.AsNewExpression().Arguments
		kind = checker.SignatureKindConstruct
	default:
		return "", fmt.Errorf("call-parameter-returns requires a call or new expression")
	}
	var signatures []*checker.Signature
	for _, part := range voidUnionParts(c.GetTypeAtLocation(callee)) {
		signatures = append(signatures, c.GetSignaturesOfType(part, kind)...)
	}
	out.number(uint64(len(signatures)))
	count := 0
	if arguments != nil {
		count = len(arguments.Nodes)
	}
	out.number(uint64(count))
	for index := 0; index < count; index++ {
		var returns []*checker.Type
		for _, signature := range signatures {
			parameters := checker.Signature_parameters(signature)
			if index < len(parameters) {
				returns = append(returns, voidSignatureReturns(c, c.GetTypeOfSymbolAtLocation(parameters[index], callee))...)
			}
		}
		writeVoidReturnTypes(out, returns)
	}
	return out.String(), nil
}
