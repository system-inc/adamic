package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// Resolved signature identity and immutable declaration location, not a callee verdict.
func (p *Program) streamSignature(out *fields, c *checker.Checker, node *ast.Node) (string, error) {
	if node.Kind != ast.KindCallExpression {
		return "", fmt.Errorf("stream-signature requires a CallExpression")
	}
	signature := c.GetResolvedSignature(node)
	out.yes(signature != nil && signature.Declaration() != nil)
	if signature != nil && signature.Declaration() != nil {
		decl := signature.Declaration()
		file := ast.GetSourceFileOfNode(decl)
		if file == nil {
			return "", fmt.Errorf("signature declaration has no source")
		}
		out.text(file.FileName().AsString())
		out.text(strings.TrimPrefix(decl.Kind.String(), "Kind"))
		out.number(uint64(decl.Pos()))
		out.number(uint64(decl.End()))
		out.yes(ast.IsExternalModule(file))
		out.yes(decl.Body() != nil)
		flags := ast.GetFunctionFlags(decl)
		out.yes(flags&ast.FunctionFlagsGenerator != 0)
		out.yes(flags&ast.FunctionFlagsAsync != 0)
		returnType := c.GetReturnTypeOfSignature(signature)
		if returnType == nil {
			out.number(0)
		} else {
			out.number(uint64(returnType.Flags()))
		}
	}
	return out.String(), nil
}
