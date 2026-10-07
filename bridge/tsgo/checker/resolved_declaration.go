package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// resolvedDeclaration returns the resolved signature's declaration and source,
// function flags and return flags. Native code decides whether to follow it.
func (p *Program) resolvedDeclaration(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "resolved-declaration" || node.Kind != ast.KindCallExpression {
		return "", fmt.Errorf("resolved-declaration requires a call")
	}
	signature := c.GetResolvedSignature(node)
	var declaration *ast.Node
	if signature != nil {
		declaration = signature.Declaration()
	}
	out.yes(declaration != nil)
	if declaration != nil {
		file := ast.GetSourceFileOfNode(declaration)
		out.text(file.FileName())
		out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
		out.number(uint64(declaration.Pos()))
		out.number(uint64(declaration.End()))
		out.yes(file.IsDeclarationFile)
		out.yes(ast.IsExternalModule(file))
		out.yes(declaration.Body() != nil)
		flags := ast.GetFunctionFlags(declaration)
		out.yes(flags&ast.FunctionFlagsAsync != 0)
		out.yes(flags&ast.FunctionFlagsGenerator != 0)
		var returnFlags uint64
		if t := c.GetReturnTypeOfSignature(signature); t != nil {
			returnFlags = uint64(t.Flags())
		}
		out.number(returnFlags)
		if file.IsDeclarationFile {
			out.text("")
		} else {
			out.text(file.Text())
		}
	}
	return out.String(), nil
}
