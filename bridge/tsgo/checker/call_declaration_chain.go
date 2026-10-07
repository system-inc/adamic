package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// callDeclarationChain exposes declaration ancestry and emit options, never a lint verdict.
func (p *Program) callDeclarationChain(c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "call-declaration-chain" || node.Kind != ast.KindCallExpression {
		return "", fmt.Errorf("call-declaration-chain requires a CallExpression")
	}
	out := &fields{}
	out.number(1)
	out.text(question)
	out.number(uint64(p.Compiler.Options().GetEmitModuleKind()))
	var chain []*ast.Node
	if signature := c.GetResolvedSignature(node); signature != nil {
		for d := signature.Declaration(); d != nil; d = d.Parent {
			chain = append(chain, d)
		}
	}
	out.number(uint64(len(chain)))
	for _, d := range chain {
		out.text(strings.TrimPrefix(d.Kind.String(), "Kind"))
		name, kind := "", ""
		if n := d.Name(); n != nil {
			name = n.Text()
			kind = strings.TrimPrefix(n.Kind.String(), "Kind")
		}
		out.text(name)
		out.text(kind)
	}
	return out.String(), nil
}
