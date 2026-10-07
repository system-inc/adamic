package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (p *Program) resolvedCallDeclaration(c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "resolved-call-declaration" || node.Kind != ast.KindCallExpression {
		return "", fmt.Errorf("resolved-call-declaration requires a call")
	}
	signature := c.GetResolvedSignature(node)
	out := &fields{}
	out.number(1)
	out.text(question)
	present := signature != nil && signature.Declaration() != nil
	out.yes(present)
	if present {
		p.writeProcessDeclaration(out, signature.Declaration())
		flags := uint64(0)
		if result := c.GetReturnTypeOfSignature(signature); result != nil {
			flags = uint64(result.Flags())
		}
		out.number(flags)
	}
	return out.String(), nil
}
