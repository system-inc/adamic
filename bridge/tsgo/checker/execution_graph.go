package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/bridge/tsgo/checker/executiongraph"
	"strings"
)

type executionEvent struct {
	hook string
	node *ast.Node
}

// executionGraph returns unlabeled syntax events and graph edges. No rule decisions live here.
func (p *Program) executionGraph(out *fields, node *ast.Node, question string) (string, error) {
	if question != "execution-graph" || !executiongraph.IsRoot(node) {
		return "", fmt.Errorf("execution-graph requires a code path root")
	}
	graph := executiongraph.Build(node, executiongraph.Hooks[executionEvent]{Expression: func(b *executiongraph.Builder[executionEvent], n *ast.Node) {
		switch n.Kind {
		case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression, ast.KindAwaitExpression:
			b.Emit(executionEvent{"expression", n})
		}
	}, Statement: func(b *executiongraph.Builder[executionEvent], n *ast.Node) {
		switch n.Kind {
		case ast.KindBlock, ast.KindForOfStatement, ast.KindVariableStatement:
			b.Emit(executionEvent{"statement", n})
		}
	}})
	out.number(uint64(len(graph.Blocks)))
	for _, block := range graph.Blocks {
		out.yes(block.Reachable)
		out.number(uint64(len(block.Events)))
		for _, event := range block.Events {
			out.text(event.hook)
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
