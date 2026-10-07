package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Flags, literal values, constituents and constraints, without lint judgments.
func (p *Program) renderTypeFacts(out *fields, c *checker.Checker, node *ast.Node) error {
	root := c.GetTypeAtLocation(node)
	g := &graph{program: p, checker: c}
	out.number(g.id(root))
	if root == nil {
		out.number(0)
		return nil
	}
	queue := []*checker.Type{root}
	seen := map[*checker.Type]bool{root: true}
	bounds := map[*checker.Type]*checker.Type{}
	for i := 0; i < len(queue); i++ {
		t := queue[i]
		parts := partsOfListenerType(t)
		var bound *checker.Type
		if t.Flags()&checker.TypeFlagsInstantiable != 0 {
			bound = checker.Checker_getBaseConstraintOfType(c, t)
			bounds[t] = bound
		}
		children := append(append([]*checker.Type{}, parts...), bound)
		for _, child := range children {
			if child != nil && !seen[child] {
				seen[child] = true
				queue = append(queue, child)
			}
		}
	}
	out.number(uint64(len(queue)))
	for _, t := range queue {
		out.number(g.id(t))
		out.number(uint64(t.Flags()))
		value := ""
		held := false
		if t.Flags()&(checker.TypeFlagsNumberLiteral|checker.TypeFlagsBigIntLiteral) != 0 {
			if literal, ok := t.AsLiteralType().Value().(fmt.Stringer); ok {
				value = literal.String()
				held = true
			}
		}
		out.yes(held)
		out.text(value)
		out.number(g.id(bounds[t]))
		parts := partsOfListenerType(t)
		out.number(uint64(len(parts)))
		for _, part := range parts {
			out.number(g.id(part))
		}
	}
	return nil
}

func init() { additionalQuestions["render-type-facts"] = (*Program).renderTypeFacts }
