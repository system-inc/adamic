package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/bridge/tsgo/checker/outputflow"
	"strings"
)

type outputSyntaxEvent struct {
	site string
	node *ast.Node
}

func (p *Program) outputFlow(_ *checker.Checker, node *ast.Node, question string) (string, error) {
	const mode = "output-flow"
	if question != mode || !outputflow.IsRoot(node) {
		return "", fmt.Errorf("output-flow requires a code path root without suffix")
	}
	graph := outputflow.Build(node, outputflow.Hooks[outputSyntaxEvent]{
		Expression: func(b *outputflow.Builder[outputSyntaxEvent], n *ast.Node) {
			if n.Kind == ast.KindCallExpression || n.Kind == ast.KindNewExpression || n.Kind == ast.KindTaggedTemplateExpression || n.Kind == ast.KindAwaitExpression {
				b.Emit(outputSyntaxEvent{"expression", n})
			}
		},
		Statement: func(b *outputflow.Builder[outputSyntaxEvent], n *ast.Node) { b.Emit(outputSyntaxEvent{"statement", n}) },
	})
	out := &fields{}
	out.number(1)
	out.text(mode)
	out.number(uint64(len(graph.Blocks)))
	for _, block := range graph.Blocks {
		out.yes(block.Reachable)
		out.number(uint64(len(block.Events)))
		for _, event := range block.Events {
			out.text(event.site)
			out.text(strings.TrimPrefix(event.node.Kind.String(), "Kind"))
			out.number(uint64(event.node.Pos()))
			out.number(uint64(event.node.End()))
		}
		out.number(uint64(len(block.Successors)))
		for _, next := range block.Successors {
			out.number(uint64(next.Index()))
		}
	}
	return out.String(), nil
}
