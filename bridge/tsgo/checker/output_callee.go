package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// outputCallee exports the resolved signature declaration, never a lint verdict.
func (p *Program) outputCallee(c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "output-callee" || node.Kind != ast.KindCallExpression {
		return "", fmt.Errorf("output-callee requires an unsuffixed call")
	}
	out := &fields{}
	out.number(1)
	out.text(question)
	signature := c.GetResolvedSignature(node)
	var declaration *ast.Node
	if signature != nil {
		declaration = signature.Declaration()
	}
	out.yes(declaration != nil)
	if declaration != nil {
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil {
			return "", fmt.Errorf("resolved declaration has no source")
		}
		out.text(file.FileName().AsString())
		out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
		out.number(uint64(declaration.Pos()))
		out.number(uint64(declaration.End()))
		out.yes(ast.IsExternalModule(file))
		flags := ast.GetFunctionFlags(declaration)
		out.yes(flags&ast.FunctionFlagsGenerator != 0)
		out.yes(flags&ast.FunctionFlagsAsync != 0)
		returned := c.GetReturnTypeOfSignature(signature)
		flagsOfReturn := uint64(0)
		if returned != nil {
			flagsOfReturn = uint64(returned.Flags())
		}
		out.number(flagsOfReturn)
		body := declaration.Body()
		out.yes(body != nil)
		if body != nil {
			out.text(strings.TrimPrefix(body.Kind.String(), "Kind"))
			out.number(uint64(body.Pos()))
			out.number(uint64(body.End()))
		}
	}
	return out.String(), nil
}
