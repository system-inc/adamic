package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// wave19ResolvedCallee exposes the chosen signature declaration and its source/body metadata.
// It does not decide whether a call prints, exits, blocks, or should be followed.
func (p *Program) wave19ResolvedCallee(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "wave19-resolved-callee" || node.Kind != ast.KindCallExpression {
		return "", fmt.Errorf("wave19-resolved-callee requires a CallExpression and no suffix")
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
	source := ast.GetSourceFileOfNode(declaration)
	if source == nil {
		return "", fmt.Errorf("signature declaration has no source")
	}
	out.text(source.FileName())
	out.yes(source.IsDeclarationFile)
	out.yes(ast.IsExternalModule(source))
	out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
	out.number(uint64(declaration.Pos()))
	out.number(uint64(declaration.End()))
	out.number(uint64(ast.GetFunctionFlags(declaration)))
	body := declaration.Body()
	out.yes(body != nil)
	if body != nil {
		out.text(strings.TrimPrefix(body.Kind.String(), "Kind"))
		out.number(uint64(body.Pos()))
		out.number(uint64(body.End()))
		out.text(source.Text())
	}
	return out.String(), nil
}
