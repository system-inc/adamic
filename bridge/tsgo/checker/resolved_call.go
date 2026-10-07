package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// resolvedCall exposes signature declarations and return flags without following a callee.
func (p *Program) resolvedCall(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "resolved-call" || node.Kind != ast.KindCallExpression {
		return "", fmt.Errorf("resolved-call requires a CallExpression")
	}
	sig := c.GetResolvedSignature(node)
	out.yes(sig != nil)
	if sig != nil {
		declaration := sig.Declaration()
		out.yes(declaration != nil)
		if declaration != nil {
			file := ast.GetSourceFileOfNode(declaration)
			out.text(file.FileName())
			out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
			out.number(uint64(declaration.Pos()))
			out.number(uint64(declaration.End()))
		}
		result := c.GetReturnTypeOfSignature(sig)
		out.yes(result != nil)
		if result != nil {
			out.number(uint64(result.Flags()))
		}
	}
	return out.String(), nil
}
