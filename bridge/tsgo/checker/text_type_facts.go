package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"unicode/utf8"
)

// Raw literal values, constraints and template-literal type texts/holes.
func (p *Program) textTypeFacts(out *fields, c *checker.Checker, node *ast.Node) error {
	root := c.GetTypeAtLocation(node)
	g := &graph{program: p, checker: c}
	out.number(g.id(root))
	if root == nil {
		out.number(0)
		out.number(0)
		return nil
	}
	bound := checker.Checker_getBaseConstraintOfType(c, root)
	if bound == nil {
		bound = root
	}
	out.number(g.id(bound))
	queue := []*checker.Type{root}
	seen := map[*checker.Type]bool{root: true}
	bounds := map[*checker.Type]*checker.Type{}
	holes := map[*checker.Type][]*checker.Type{}
	for i := 0; i < len(queue); i++ {
		t := queue[i]
		bounds[t] = checker.Checker_getBaseConstraintOfType(c, t)
		parts := partsOfListenerType(t)
		if t.Flags()&checker.TypeFlagsTemplateLiteral != 0 {
			holes[t] = t.AsTemplateLiteralType().Types()
		}
		children := append(append(append([]*checker.Type{}, parts...), holes[t]...), bounds[t])
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
		out.yes(t.Flags()&checker.TypeFlagsIntrinsic != 0 && t.AsIntrinsicType().IntrinsicName() == "error")
		kind, value := "", ""
		if t.Flags()&checker.TypeFlagsLiteral != 0 {
			switch v := t.AsLiteralType().Value().(type) {
			case string:
				kind = "string"
				value = v
			case bool:
				kind = "boolean"
				if v {
					value = "true"
				} else {
					value = "false"
				}
			case fmt.Stringer:
				kind = "number"
				value = v.String()
			}
		}
		if !utf8.ValidString(value) {
			return fmt.Errorf("text-type-facts refuses non-UTF-8 literal values")
		}
		out.text(kind)
		out.text(value)
		out.number(g.id(bounds[t]))
		parts := partsOfListenerType(t)
		out.number(uint64(len(parts)))
		for _, part := range parts {
			out.number(g.id(part))
		}
		var texts []string
		if t.Flags()&checker.TypeFlagsTemplateLiteral != 0 {
			texts = t.AsTemplateLiteralType().Texts()
		}
		out.number(uint64(len(texts)))
		for _, text := range texts {
			if !utf8.ValidString(text) {
				return fmt.Errorf("text-type-facts refuses non-UTF-8 template texts")
			}
			out.text(text)
		}
		out.number(uint64(len(holes[t])))
		for _, hole := range holes[t] {
			out.number(g.id(hole))
		}
	}
	return nil
}
func init() { additionalQuestions["text-type-facts"] = (*Program).textTypeFacts }
