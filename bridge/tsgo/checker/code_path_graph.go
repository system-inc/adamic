package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	codepath "github.com/system-inc/adamic/bridge/tsgo/code_path_graph"
)

// The syntax substrate emits every expression/statement in evaluation order,
// never classified writes, exits, blocking calls, messages or lint verdicts.
func (p *Program) codePathGraph(out *fields, node *ast.Node, question string) (string, error) {
	if question != "code-path-graph" || !codepath.IsRoot(node) {
		return "", fmt.Errorf("code-path-graph requires a code path root")
	}
	type event struct {
		node      *ast.Node
		statement bool
	}
	g := codepath.Build(node, codepath.Hooks[event]{Expression: func(b *codepath.Builder[event], n *ast.Node) { b.Emit(event{node: n}) }, Statement: func(b *codepath.Builder[event], n *ast.Node) { b.Emit(event{node: n, statement: true}) }})
	out.number(uint64(len(g.Blocks)))
	for _, block := range g.Blocks {
		out.yes(block.Reachable)
		out.number(uint64(len(block.Successors)))
		for _, next := range block.Successors {
			out.number(uint64(next.Index()))
		}
		out.number(uint64(len(block.Events)))
		for _, e := range block.Events {
			out.yes(e.statement)
			out.text(strings.TrimPrefix(e.node.Kind.String(), "Kind"))
			out.number(uint64(e.node.Pos()))
			out.number(uint64(e.node.End()))
		}
	}
	return out.String(), nil
}
