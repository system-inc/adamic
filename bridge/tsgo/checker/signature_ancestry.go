package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"strings"
)

// Signature ancestry is declaration syntax, with no rule-specific judgment.
func (p *Program) signatureAncestry(out *fields, c *checker.Checker, node *ast.Node) error {
	if node.Kind != ast.KindCallExpression {
		return fmt.Errorf("signature-ancestry requires a CallExpression")
	}
	signature := c.GetResolvedSignature(node)
	var declaration *ast.Node
	if signature != nil {
		declaration = signature.Declaration()
	}
	out.yes(declaration != nil)
	if declaration == nil {
		return nil
	}
	var ancestors []*ast.Node
	for current := declaration; current != nil; current = current.Parent {
		ancestors = append(ancestors, current)
	}
	out.number(uint64(len(ancestors)))
	for _, current := range ancestors {
		out.text(strings.TrimPrefix(current.Kind.String(), "Kind"))
		name, kind := "", ""
		if current.Kind != ast.KindSourceFile {
			if n := current.Name(); n != nil {
				name = n.Text()
				kind = strings.TrimPrefix(n.Kind.String(), "Kind")
			}
		}
		out.text(name)
		out.text(kind)
	}
	return nil
}

func init() { additionalQuestions["signature-ancestry"] = (*Program).signatureAncestry }
