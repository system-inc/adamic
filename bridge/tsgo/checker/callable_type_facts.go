package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// callableTypeFacts exposes nonnullable call/construct signatures and return
// properties, leaving function/component classification to Adamic.
func (p *Program) callableTypeFacts(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "callable-type-facts" {
		return "", fmt.Errorf("unexpected callable-type-facts suffix")
	}
	parts := voidUnionParts(c.GetNonNullableType(c.GetTypeAtLocation(node)))
	out.number(uint64(len(parts)))
	for _, part := range parts {
		signatures := c.GetSignaturesOfType(part, checker.SignatureKindCall)
		out.number(uint64(len(signatures)))
		out.number(uint64(len(c.GetSignaturesOfType(part, checker.SignatureKindConstruct))))
		for _, signature := range signatures {
			result := checker.Checker_getReturnTypeOfSignature(c, signature)
			out.number(uint64(result.Flags()))
			returns := voidUnionParts(result)
			out.number(uint64(len(returns)))
			for _, ret := range returns {
				out.number(uint64(ret.Flags()))
				for _, name := range []string{"type", "props", "key"} {
					out.yes(checker.Checker_getPropertyOfType(c, ret, name) != nil)
				}
			}
		}
	}
	return out.String(), nil
}
