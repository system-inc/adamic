package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// The checker-selected signature declaration, raw flags and source text.
func (p *Program) resolvedCallTarget(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "resolved-call-target" || node.Kind != ast.KindCallExpression {
		return "", fmt.Errorf("resolved-call-target requires a CallExpression")
	}
	signature := c.GetResolvedSignature(node)
	out.yes(signature != nil)
	if signature == nil {
		return out.String(), nil
	}
	result := c.GetReturnTypeOfSignature(signature)
	flags := uint64(0)
	if result != nil {
		flags = uint64(result.Flags())
	}
	out.number(flags)
	declaration := signature.Declaration()
	out.yes(declaration != nil)
	if declaration == nil {
		return out.String(), nil
	}
	out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
	file := ast.GetSourceFileOfNode(declaration)
	out.text(file.FileName())
	out.yes(ast.IsExternalModule(file))
	out.number(uint64(declaration.Pos()))
	out.number(uint64(declaration.End()))
	out.number(uint64(ast.GetFunctionFlags(declaration)))
	body := declaration.Body()
	out.yes(body != nil)
	if body != nil {
		out.text(file.Text())
		out.text(strings.TrimPrefix(body.Kind.String(), "Kind"))
		out.number(uint64(body.Pos()))
		out.number(uint64(body.End()))
	}
	return out.String(), nil
}
