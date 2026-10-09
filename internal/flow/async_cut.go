package flow

// SuspensionGraph is the narrow view state cutting needs from a flow graph. The shared
// single assignment adapter supplies this view; cutting never reads SSA implementation files.
type SuspensionGraph interface {
	EntryBlock() BlockId
	BlockOrder() []BlockId
	Successors(BlockId, func(BlockId))
	Suspension(BlockId) (BlockId, BlockId, bool)
}

type suspensionGraph struct{ function *Function }

func (g suspensionGraph) EntryBlock() BlockId { return g.function.Entry }
func (g suspensionGraph) BlockOrder() []BlockId {
	result := make([]BlockId, 0, len(g.function.Blocks))
	for _, block := range g.function.Blocks {
		result = append(result, block.Id)
	}
	return result
}
func (g suspensionGraph) Successors(id BlockId, visit func(BlockId)) {
	block, ok := g.function.Block(id)
	if !ok {
		panic("flow: state cutting reached a missing block")
	}
	EachSuccessor(block.Terminal, visit)
}
func (g suspensionGraph) Suspension(id BlockId) (BlockId, BlockId, bool) {
	block, _ := g.function.Block(id)
	terminal, ok := block.Terminal.(*Suspend)
	if !ok {
		return 0, 0, false
	}
	return terminal.Fulfilled, terminal.Rejected, true
}

type AsyncRegions struct {
	Roots  []BlockId
	Blocks [][]BlockId
	Owner  map[BlockId]int
}

func CutAsync(function *Function) AsyncRegions { return CutSuspensionGraph(suspensionGraph{function}) }

// CutSuspensionGraph promotes joins reached from different roots until each block belongs
// to exactly one region. Only Suspend yields; crossing a synchronous region edge is a jump.
func CutSuspensionGraph(graph SuspensionGraph) AsyncRegions {
	order := graph.BlockOrder()
	roots := map[BlockId]bool{graph.EntryBlock(): true}
	for _, id := range order {
		if fulfilled, rejected, ok := graph.Suspension(id); ok {
			roots[fulfilled] = true
			roots[rejected] = true
		}
	}
	for {
		owners := map[BlockId]BlockId{}
		changed := false
		for _, root := range order {
			if !roots[root] {
				continue
			}
			pending := []BlockId{root}
			visited := map[BlockId]bool{}
			for len(pending) > 0 {
				id := pending[len(pending)-1]
				pending = pending[:len(pending)-1]
				if visited[id] || (id != root && roots[id]) {
					continue
				}
				visited[id] = true
				if owner, ok := owners[id]; ok && owner != root {
					roots[id] = true
					changed = true
					continue
				}
				owners[id] = root
				if _, _, suspends := graph.Suspension(id); !suspends {
					graph.Successors(id, func(next BlockId) { pending = append(pending, next) })
				}
			}
		}
		if changed {
			continue
		}
		result := AsyncRegions{Owner: map[BlockId]int{}}
		indices := map[BlockId]int{}
		for _, id := range order {
			if roots[id] {
				indices[id] = len(result.Roots)
				result.Roots = append(result.Roots, id)
				result.Blocks = append(result.Blocks, nil)
			}
		}
		for _, id := range order {
			owner, ok := owners[id]
			if !ok {
				panic("flow: block has no suspension region")
			}
			index := indices[owner]
			result.Owner[id] = index
			result.Blocks[index] = append(result.Blocks[index], id)
		}
		return result
	}
}
