package checker

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Symbol identity, type-position classification, and contextual construct counts.
func (p *Program) referenceContext(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "reference-context" {
		return "", fmt.Errorf("unexpected reference-context suffix")
	}
	out.number(p.symbolID(c.GetSymbolAtLocation(node)))
	out.yes(ast.IsPartOfTypeNode(node))
	own := c.GetTypeAtLocation(node)
	contextual := checker.Checker_getContextualType(c, node, checker.ContextFlagsNone)
	g := &graph{program: p, checker: c}
	out.number(g.id(own))
	out.number(g.id(contextual))
	var parts []*checker.Type
	if contextual != nil {
		if contextual.Flags()&checker.TypeFlagsUnion != 0 {
			parts = contextual.Types()
		} else {
			parts = []*checker.Type{contextual}
		}
	}
	out.number(uint64(len(parts)))
	for _, part := range parts {
		out.number(uint64(len(c.GetSignaturesOfType(part, checker.SignatureKindConstruct))))
	}
	return out.String(), nil
}
