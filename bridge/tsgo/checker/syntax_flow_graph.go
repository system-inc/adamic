package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	flow "github.com/system-inc/adamic/bridge/tsgo/checker/wave_01_next_flow"
)

// A generic evaluation-order graph with only raw identifier read/write events.
// Consumers classify the identifiers and analyze paths in native Adamic.
func (p *Program) syntaxFlowGraph(out *fields, _ *checker.Checker, _ *ast.SourceFile, node *ast.Node, question string) error {
	if question != "syntax-flow-graph" || !flow.IsRoot(node) {
		return fmt.Errorf("syntax-flow-graph requires a code path root")
	}
	type event struct {
		write bool
		node  *ast.Node
	}
	graph := flow.Build(node, flow.Hooks[event]{
		Read:  func(b *flow.Builder[event], node *ast.Node) { b.Emit(event{node: node}) },
		Write: func(b *flow.Builder[event], node *ast.Node) { b.Emit(event{write: true, node: node}) },
	})
	final, thrown := map[int]bool{}, map[int]bool{}
	for _, block := range graph.FinalBlocks {
		final[block.Index()] = true
	}
	for _, block := range graph.ThrownBlocks {
		thrown[block.Index()] = true
	}
	out.yes(graph.EndReachable)
	out.number(uint64(len(graph.Blocks)))
	for _, block := range graph.Blocks {
		out.yes(block.Reachable)
		out.yes(final[block.Index()])
		out.yes(thrown[block.Index()])
		out.number(uint64(len(block.Successors)))
		for _, next := range block.Successors {
			out.number(uint64(next.Index()))
		}
		out.number(uint64(len(block.Events)))
		for _, event := range block.Events {
			out.yes(event.write)
			out.text(strings.TrimPrefix(event.node.Kind.String(), "Kind"))
			out.number(uint64(event.node.Pos()))
			out.number(uint64(event.node.End()))
		}
	}
	return nil
}
func init() { additionalQuestions["syntax-flow-graph"] = (*Program).syntaxFlowGraph }
