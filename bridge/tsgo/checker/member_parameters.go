package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// Parameter source text is a compiler fact. Native callers decide whether lists match.
func writeMemberParameters(out *fields, node *ast.Node) {
	function := ast.IsFunctionLike(node)
	out.yes(function)
	if !function {
		return
	}
	parameters := node.ParameterList()
	count := 0
	if parameters != nil {
		count = len(parameters.Nodes)
	}
	out.number(uint64(count))
	if parameters != nil {
		for _, parameter := range parameters.Nodes {
			source := ast.GetSourceFileOfNode(parameter)
			first := scanner.GetTokenPosOfNode(parameter, source, false)
			out.text(source.Text()[first:parameter.End()])
		}
	}
}
func (p *Program) memberParameters(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "member-parameters" {
		return "", fmt.Errorf("unexpected member parameter suffix")
	}
	symbol := c.GetSymbolAtLocation(node)
	count := 0
	if symbol != nil {
		count = len(symbol.Declarations)
	}
	out.number(uint64(count))
	if symbol != nil {
		for _, declaration := range symbol.Declarations {
			writeMemberParameters(out, declaration)
			var includes *ast.Symbol
			if declaration.Parent != nil {
				includes = checker.Checker_getPropertyOfType(c, c.GetTypeAtLocation(declaration.Parent), "includes")
			}
			count = 0
			if includes != nil {
				count = len(includes.Declarations)
			}
			out.number(uint64(count))
			if includes != nil {
				for _, candidate := range includes.Declarations {
					writeMemberParameters(out, candidate)
				}
			}
		}
	}
	return out.String(), nil
}
