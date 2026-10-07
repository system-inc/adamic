package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (p *Program) callDeclaration(c *checker.Checker, node *ast.Node, question string) (string, error) {
	const mode = "call-declaration"
	if question != mode || node.Kind != ast.KindCallExpression {
		return "", fmt.Errorf("call-declaration requires a CallExpression without suffix")
	}
	out := &fields{}
	out.number(1)
	out.text(mode)
	signature := c.GetResolvedSignature(node)
	out.yes(signature != nil && signature.Declaration() != nil)
	if signature != nil && signature.Declaration() != nil {
		p.writePlatformDeclaration(out, signature.Declaration())
		t := c.GetReturnTypeOfSignature(signature)
		flags := uint64(0)
		if t != nil {
			flags = uint64(t.Flags())
		}
		out.number(flags)
	}
	return out.String(), nil
}
