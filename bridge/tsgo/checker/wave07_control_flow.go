package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	flow "github.com/system-inc/adamic/bridge/tsgo/wave07_flow"
	"strings"
)

// wave07ControlFlow serializes structural blocks and unclassified syntax hooks.
// It does not recognize writes, exits, blocking calls, or any lint condition.
func (p *Program) wave07ControlFlow(root *ast.Node, question string) (string, error) {
	if question != "wave07-control-flow" || !flow.IsRoot(root) {
		return "", fmt.Errorf("wave07-control-flow requires a code path root")
	}
	graph := flow.Build(root, flow.Hooks[*ast.Node]{
		Expression: func(b *flow.Builder[*ast.Node], node *ast.Node) { b.Emit(node) },
		Statement:  func(b *flow.Builder[*ast.Node], node *ast.Node) { b.Emit(node) },
	})
	out := &fields{}
	out.number(1)
	out.text(question)
	out.number(uint64(len(graph.Blocks)))
	for _, block := range graph.Blocks {
		out.yes(block.Reachable)
		var successors []uint64
		for _, next := range block.Successors {
			if next != nil {
				successors = append(successors, uint64(next.Index()))
			}
		}
		out.ids(successors)
		out.number(uint64(len(block.Events)))
		for _, event := range block.Events {
			out.text(strings.TrimPrefix(event.Kind.String(), "Kind"))
			out.number(uint64(event.Pos()))
			out.number(uint64(event.End()))
		}
	}
	return out.String(), nil
}
