// An unlabelled evaluation graph. Consumers attach meaning to its syntax events.
package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/bridge/tsgo/checker/process_flow"
)

func (p *Program) syntaxControlFlow(node *ast.Node, question string) (string, error) {
	if question != "syntax-control-flow" || !process_flow.IsRoot(node) {
		return "", fmt.Errorf("syntax-control-flow requires a code path root")
	}
	source := ast.GetSourceFileOfNode(node)
	_, ids := syntaxNodes(source)
	type event struct {
		hook uint64
		node *ast.Node
	}
	graph := process_flow.Build(node, process_flow.Hooks[event]{
		Expression: func(b *process_flow.Builder[event], n *ast.Node) { b.Emit(event{0, n}) },
		Statement:  func(b *process_flow.Builder[event], n *ast.Node) { b.Emit(event{1, n}) },
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
		for _, e := range block.Events {
			out.number(e.hook)
			out.number(ids[e.node])
		}
	}
	return out.String(), nil
}
