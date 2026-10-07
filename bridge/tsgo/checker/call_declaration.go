package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

func init() { questionExtensions["call-declaration"] = callDeclarationQuestion }
func callDeclarationQuestion(p *Program, out *fields, c *checker.Checker, source *ast.SourceFile, node *ast.Node, question string) (string, error) {
	if question != "call-declaration" || node.Kind != ast.KindCallExpression {
		return "", fmt.Errorf("call-declaration requires CallExpression")
	}
	signature := c.GetResolvedSignature(node)
	if signature == nil || signature.Declaration() == nil {
		out.yes(false)
		return out.String(), nil
	}
	declaration := signature.Declaration()
	file := ast.GetSourceFileOfNode(declaration)
	if file == nil {
		return "", fmt.Errorf("signature declaration has no file")
	}
	out.yes(true)
	out.text(file.FileName())
	out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
	out.number(uint64(declaration.Pos()))
	out.number(uint64(declaration.End()))
	out.yes(ast.IsExternalModule(file))
	out.yes(declaration.Body() != nil)
	out.yes(ast.GetFunctionFlags(declaration)&ast.FunctionFlagsGenerator != 0)
	out.yes(ast.GetFunctionFlags(declaration)&ast.FunctionFlagsAsync != 0)
	result := c.GetReturnTypeOfSignature(signature)
	if result == nil {
		out.number(0)
	} else {
		out.number(uint64(result.Flags()))
	}
	return out.String(), nil
}
