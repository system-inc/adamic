package control_flow_graph

// Test-only overlay constructs graphs; Solve and the rule lattice stay unchanged.
func Wave01Graph[E any](events [][]E, successors [][]int) *Graph[E] {
	g := &Graph[E]{}
	for i, event := range events {
		g.Blocks = append(g.Blocks, &Block[E]{index: int32(i), Events: event})
	}
	for i, next := range successors {
		for _, n := range next {
			g.Blocks[i].Successors = append(g.Blocks[i].Successors, g.Blocks[n])
		}
	}
	pending := []int{0}
	for len(pending) > 0 {
		i := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if g.Blocks[i].Reachable {
			continue
		}
		g.Blocks[i].Reachable = true
		pending = append(pending, successors[i]...)
	}
	return g
}
