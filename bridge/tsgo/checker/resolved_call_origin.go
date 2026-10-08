package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

func (p *Program) resolvedCallOrigin(out *fields, c *checker.Checker, node *ast.Node, question string) error {
	if question != "resolved-call-origin" || node.Kind != ast.KindCallExpression {
		return fmt.Errorf("resolved-call-origin requires a call")
	}
	signature := c.GetResolvedSignature(node)
	var declaration *ast.Node
	flags := uint64(0)
	if signature != nil {
		declaration = signature.Declaration()
		if t := c.GetReturnTypeOfSignature(signature); t != nil {
			flags = uint64(t.Flags())
		}
	}
	out.number(flags)
	out.yes(declaration != nil)
	if declaration != nil {
		source := ast.GetSourceFileOfNode(declaration)
		if source == nil {
			return fmt.Errorf("signature declaration has no source")
		}
		out.text(string(source.FileName()))
		out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
		out.number(uint64(declaration.Pos()))
		out.number(uint64(declaration.End()))
		out.yes(ast.IsExternalModule(source))
		out.number(uint64(ast.GetFunctionFlags(declaration)))
		out.yes(declaration.Body() != nil)
	}
	return nil
}
