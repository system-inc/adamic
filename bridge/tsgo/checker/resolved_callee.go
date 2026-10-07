package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// resolvedCallee exposes signature declaration and return flags without inspecting its body.
func (p *Program) resolvedCallee(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "resolved-callee" || node.Kind != ast.KindCallExpression {
		return "", fmt.Errorf("resolved-callee requires CallExpression")
	}
	signature := c.GetResolvedSignature(node)
	out.yes(signature != nil)
	if signature == nil {
		return out.String(), nil
	}
	t := c.GetReturnTypeOfSignature(signature)
	flags := uint64(0)
	if t != nil {
		flags = uint64(t.Flags())
	}
	out.number(flags)
	declaration := signature.Declaration()
	out.yes(declaration != nil)
	if declaration == nil {
		return out.String(), nil
	}
	source := ast.GetSourceFileOfNode(declaration)
	if source == nil {
		return "", fmt.Errorf("signature declaration has no source")
	}
	out.text(source.FileName())
	out.yes(ast.IsExternalModule(source))
	out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
	out.number(uint64(declaration.Pos()))
	out.number(uint64(declaration.End()))
	out.number(uint64(ast.GetFunctionFlags(declaration)))
	out.yes(declaration.Body() != nil)
	return out.String(), nil
}
