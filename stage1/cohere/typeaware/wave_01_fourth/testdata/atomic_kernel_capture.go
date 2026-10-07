package core

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"strings"
)

// Test-only overlay exposes observations from the unchanged production lattice.
func Wave01AtomicKernelCapture() string {
	var out strings.Builder
	first := &ast.Symbol{Name: "first"}
	second := &ast.Symbol{Name: "second"}
	apply := func(state atomicReadState, kind atomicEventKind, symbol *ast.Symbol) bool {
		stale := false
		applyAtomicTransfer(&control_flow_graph.Block[atomicEvent]{Events: []atomicEvent{{kind: kind, symbol: symbol}}}, state, func(atomicEvent) { stale = true })
		return stale
	}
	lattice := outdatedReadLattice{}
	for mode := 0; mode < 16; mode++ {
		left := lattice.Bottom()
		right := lattice.Bottom()
		apply(left, atomicRead, first)
		if mode&1 != 0 {
			apply(left, atomicSuspend, nil)
		}
		if mode&2 != 0 {
			apply(left, atomicRead, first)
		}
		apply(right, atomicRead, second)
		if mode&4 != 0 {
			apply(right, atomicSuspend, nil)
		}
		joined := lattice.Meet(left, right)
		if mode&8 != 0 {
			apply(joined, atomicSuspend, nil)
		}
		fmt.Fprintf(&out, "%d %t %t %t %t\n", mode, apply(joined, atomicWrite, first), apply(joined, atomicWrite, second), left.outdated[first], right.outdated[second])
	}
	return out.String()
}
