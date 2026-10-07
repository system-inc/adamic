package checker

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (p *Program) resolvedCallDeclaration(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "resolved-call-declaration" || (node.Kind != ast.KindCallExpression && node.Kind != ast.KindNewExpression && node.Kind != ast.KindTaggedTemplateExpression) {
		return "", fmt.Errorf("resolved-call-declaration requires a call-like expression")
	}
	signature := c.GetResolvedSignature(node)
	out.yes(signature != nil)
	if signature != nil {
		declaration := signature.Declaration()
		out.yes(declaration != nil)
		if declaration != nil {
			p.contextDeclaration(out, declaration)
		}
		result := checker.Checker_getReturnTypeOfSignature(c, signature)
		out.yes(result != nil)
		if result != nil {
			out.number(uint64(result.Flags()))
		}
	}
	return out.String(), nil
}
