package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// promisedShape exposes only the declared return and its promised type. The
// caller decides whether removing a return would require an explicit return.
func (p *Program) promisedShape(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if strings.Split(question, "\n")[0] != "promised-shape" {
		return p.wave06Question(out, c, node, question)
	}
	if question != "promised-shape" || !ast.IsFunctionLikeDeclaration(node) || node.Type() == nil {
		return "", fmt.Errorf("promised-shape requires an annotated function")
	}
	subject := c.GetTypeFromTypeNode(node.Type())
	if ast.GetFunctionFlags(node)&ast.FunctionFlagsAsync != 0 {
		subject = c.GetPromisedTypeOfPromise(subject)
	}
	g := &graph{program: p, checker: c, seen: make(map[*checker.Type]bool)}
	out.yes(p.Compiler.Options().GetStrictOptionValue(p.Compiler.Options().StrictNullChecks))
	out.yes(subject != nil)
	var roots []uint64
	if subject != nil {
		roots = append(roots, g.add(subject))
	}
	out.ids(roots)
	out.number(0)
	g.write(out)
	return out.String(), nil
}
