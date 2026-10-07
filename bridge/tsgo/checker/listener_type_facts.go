package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Raw asserted type graph, symbol provenance, base edges and assignability.
func (p *Program) listenerTypeFacts(out *fields, c *checker.Checker, node *ast.Node) error {
	if node.Kind != ast.KindAsExpression && node.Kind != ast.KindTypeAssertionExpression {
		return fmt.Errorf("listener-type-facts requires an assertion")
	}
	expression := node.Expression()
	for expression != nil {
		switch expression.Kind {
		case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindAsExpression, ast.KindTypeAssertionExpression:
			expression = expression.Expression()
			continue
		}
		break
	}
	if expression == nil || node.Type() == nil {
		return fmt.Errorf("assertion without expression or type")
	}
	g := &graph{program: p, checker: c}
	known := c.GetTypeAtLocation(expression)
	target := c.GetTypeFromTypeNode(node.Type())
	out.yes(known != nil && target != nil && c.IsTypeAssignableTo(c.GetNonNullableType(known), c.GetNonNullableType(target)))
	if target == nil {
		out.number(0)
		out.number(0)
		return nil
	}
	out.number(g.id(target))
	queue := []*checker.Type{target}
	seen := map[*checker.Type]bool{target: true}
	bases := map[*checker.Type][]*checker.Type{}
	for i := 0; i < len(queue); i++ {
		t := queue[i]
		var parts []*checker.Type
		if t.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
			parts = t.Types()
		}
		children := parts
		if checker.Type_objectFlags(t)&(checker.ObjectFlagsInterface|checker.ObjectFlagsClass) != 0 {
			bases[t] = checker.Checker_getBaseTypes(c, t)
			children = append(append([]*checker.Type{}, children...), bases[t]...)
		}
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
		symbol := checker.Type_symbol(t)
		out.number(p.symbolID(symbol))
		name := ""
		flags := uint64(0)
		var decls []*ast.Node
		if symbol != nil {
			name = symbol.Name
			flags = uint64(symbol.Flags)
			decls = symbol.Declarations
		}
		out.text(name)
		out.number(flags)
		out.number(uint64(len(decls)))
		for _, decl := range decls {
			file := ast.GetSourceFileOfNode(decl)
			out.yes(file != nil && p.Compiler.IsSourceFileDefaultLibrary(file.Path()))
		}
		out.number(uint64(len(partsOfListenerType(t))))
		for _, part := range partsOfListenerType(t) {
			out.number(g.id(part))
		}
		out.number(uint64(len(bases[t])))
		for _, base := range bases[t] {
			out.number(g.id(base))
		}
	}
	return nil
}

func partsOfListenerType(t *checker.Type) []*checker.Type {
	if t.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		return t.Types()
	}
	return nil
}

func init() { additionalQuestions["listener-type-facts"] = (*Program).listenerTypeFacts }
