package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// Raw resolved signature metadata, not a callee-write or blocking verdict.
func (p *Program) wave27CallDeclaration(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "wave-27-call-declaration" || node.Kind != ast.KindCallExpression {
		return "", fmt.Errorf("call-declaration requires a CallExpression")
	}
	signature := c.GetResolvedSignature(node)
	var declaration *ast.Node
	if signature != nil {
		declaration = signature.Declaration()
	}
	out.yes(declaration != nil)
	if declaration != nil {
		source := ast.GetSourceFileOfNode(declaration)
		if source == nil {
			return "", fmt.Errorf("signature declaration has no source")
		}
		out.text(source.FileName())
		out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
		out.number(uint64(declaration.Pos()))
		out.number(uint64(declaration.End()))
		out.number(uint64(ast.GetFunctionFlags(declaration)))
		out.yes(declaration.Body() != nil)
		out.yes(ast.IsExternalModule(source))
		t := c.GetReturnTypeOfSignature(signature)
		flags := uint64(0)
		if t != nil {
			flags = uint64(t.Flags())
		}
		out.number(flags)
	}
	return out.String(), nil
}
