package core

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	cfg "github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"strings"
)

func Wave01AtomicSolverCapture() string {
	var out strings.Builder
	a, b, c := &ast.Symbol{}, &ast.Symbol{}, &ast.Symbol{}
	for mode := 0; mode < 16; mode++ {
		events := [][]atomicEvent{{{kind: atomicRead, symbol: a}}, nil, {{kind: atomicRead, symbol: b}}, nil, {{kind: atomicWrite, symbol: a}, {kind: atomicWrite, symbol: b}}, {{kind: atomicRead, symbol: c}, {kind: atomicSuspend}}}
		if mode&1 != 0 {
			events[1] = []atomicEvent{{kind: atomicSuspend}}
		}
		if mode&2 != 0 {
			events[2] = append(events[2], atomicEvent{kind: atomicSuspend})
		}
		if mode&8 != 0 {
			events[3] = []atomicEvent{{kind: atomicRead, symbol: a}}
		}
		next := [][]int{{1, 2}, {3}, {3}, {4}, nil, {4}}
		if mode&4 != 0 {
			next[3] = []int{1, 4}
		}
		graph := cfg.Wave01Graph(events, next)
		solution := cfg.Solve(graph, cfg.Forward, outdatedReadLattice{})
		for i, block := range graph.Blocks {
			state, reachable := solution.In(block)
			fmt.Fprintf(&out, "%d %d %t %t %t %t %t %t\n", mode, i, reachable, state.fresh[a], state.fresh[b], state.outdated[a], state.outdated[b], state.outdated[c])
		}
	}
	return out.String()
}
