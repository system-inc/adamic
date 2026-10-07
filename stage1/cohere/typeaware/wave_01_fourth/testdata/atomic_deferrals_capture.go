package core

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"strings"
)

func Wave01AtomicDeferralsCapture() string {
	var out strings.Builder
	node := func(start, end int) *ast.Node { return &ast.Node{Loc: core.NewTextRange(start, end)} }
	for mode := 0; mode < 32; mode++ {
		kept := []atomicEvent{{kind: atomicRead, target: node(0, 2)}, {kind: atomicRead, target: node(4, 6)}, {kind: atomicRead, target: node(8, 10)}}
		if mode&4 != 0 {
			kept[2].target = nil
		}
		start := 20
		limit := 30
		if mode&8 != 0 {
			start = 3
			limit = 12
		}
		end := 20
		if mode&16 != 0 {
			end = 2
		}
		a, b := node(3, 12), node(0, 20)
		deferrals := []deferredAtomicEvent{{within: node(start, limit), originalIndex: mode & 3, event: atomicEvent{target: a}}, {within: node(0, end), originalIndex: 0, event: atomicEvent{target: b}}}
		ids := map[*ast.Node]int{kept[0].target: 1, kept[1].target: 2, kept[2].target: 3, a: 4, b: 5}
		fmt.Fprintf(&out, "%d", mode)
		for _, event := range placeDeferrals(kept, deferrals) {
			fmt.Fprintf(&out, " %d", ids[event.target])
		}
		out.WriteByte('\n')
	}
	return out.String()
}
