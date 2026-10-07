package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// wave04NextResolvedSignatureDeclaration exposes the selected declaration and body,
// not whether the callee writes or blocks standard streams.
func (p *Program) wave04NextResolvedSignatureDeclaration(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "wave04-next-resolved-signature-declaration" {
		return "", fmt.Errorf("unexpected resolved declaration suffix")
	}
	switch node.Kind {
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression:
	default:
		return "", fmt.Errorf("resolved declaration requires a call-like node")
	}
	signature := c.GetResolvedSignature(node)
	out.yes(signature != nil)
	if signature != nil {
		returned := c.GetReturnTypeOfSignature(signature)
		out.number(uint64(returned.Flags()))
		declaration := signature.Declaration()
		out.yes(declaration != nil)
		if declaration != nil {
			p.wave04NextWriteContext(out, declaration)
			out.number(uint64(ast.GetFunctionFlags(declaration)))
			body := declaration.Body()
			out.yes(body != nil)
			if body != nil {
				out.number(uint64(body.Pos()))
				out.number(uint64(body.End()))
			}
		}
	}
	return out.String(), nil
}
