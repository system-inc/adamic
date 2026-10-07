package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Emit module kind and the resolved signature declaration's ancestor chain.
func (p *Program) signatureContext(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "signature-context" || node.Kind != ast.KindCallExpression {
		return "", fmt.Errorf("signature-context requires a CallExpression")
	}
	out.number(uint64(p.Compiler.Options().GetEmitModuleKind()))
	var declarations []*ast.Node
	if signature := c.GetResolvedSignature(node); signature != nil {
		for d := signature.Declaration(); d != nil; d = d.Parent {
			declarations = append(declarations, d)
		}
	}
	out.number(uint64(len(declarations)))
	for _, d := range declarations {
		out.text(strings.TrimPrefix(d.Kind.String(), "Kind"))
		name := ""
		nameKind := ""
		if n := d.Name(); n != nil {
			name = n.Text()
			nameKind = strings.TrimPrefix(n.Kind.String(), "Kind")
		}
		out.text(strings.ToValidUTF8(name, "�"))
		out.text(nameKind)
	}
	return out.String(), nil
}
