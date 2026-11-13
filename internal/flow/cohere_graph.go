package flow

import ssa "github.com/system-inc/cohere/static_single_assignment"

// cohereGraph adapts only graph maintenance. SSA construction remains local
// until #5jxpf6r head 2; unused methods fail explicitly at that boundary.
type cohereGraph struct{}

var _ ssa.Graph[*Function, *BasicBlock, Place] = cohereGraph{}

func (cohereGraph) Entry(function *Function) ssa.BlockId { return ssa.BlockId(function.Entry) }
func (cohereGraph) BlockBound(function *Function) int {
	bound := int(function.nextBlock)
	for id := range function.blocksById {
		if int(id) >= bound {
			bound = int(id) + 1
		}
	}
	return bound
}
func (cohereGraph) Block(function *Function, id ssa.BlockId) (*BasicBlock, bool) {
	return function.Block(BlockId(id))
}
func (cohereGraph) Blocks(function *Function) []*BasicBlock            { return function.Blocks }
func (cohereGraph) SetBlocks(function *Function, blocks []*BasicBlock) { function.Blocks = blocks }
func (cohereGraph) Retain(function *Function, keep func(id ssa.BlockId) bool) {
	for id := range function.blocksById {
		if !keep(ssa.BlockId(id)) {
			delete(function.blocksById, id)
		}
	}
}
func (cohereGraph) Id(block *BasicBlock) ssa.BlockId { return ssa.BlockId(block.Id) }
func (cohereGraph) Predecessors(block *BasicBlock) []ssa.BlockId {
	predecessors := make([]ssa.BlockId, len(block.Predecessors))
	for index, id := range block.Predecessors {
		predecessors[index] = ssa.BlockId(id)
	}
	return predecessors
}
func (cohereGraph) SetPredecessors(block *BasicBlock, predecessors []ssa.BlockId) {
	block.Predecessors = block.Predecessors[:0]
	for _, id := range predecessors {
		block.Predecessors = append(block.Predecessors, BlockId(id))
	}
}
func (cohereGraph) EachEdge(block *BasicBlock, visit func(successor ssa.BlockId, edge ssa.Edge)) {
	// Preserve EachSuccessor's order and duplicates: reverse postorder reverses
	// these edges, while MarkPredecessors deduplicates them itself.
	if terminal, ok := block.Terminal.(*MayThrow); ok {
		visit(ssa.BlockId(terminal.Next), ssa.Real)
		visit(ssa.BlockId(terminal.Handler), ssa.Exceptional)
		return
	}
	EachSuccessor(block.Terminal, func(id BlockId) { visit(ssa.BlockId(id), ssa.Real) })
}
func (cohereGraph) InstructionCount(function *Function, block *BasicBlock) int {
	return len(block.Instructions)
}
func (cohereGraph) SetInstructionOrder(function *Function, block *BasicBlock, index int, order ssa.EvaluationOrder) {
	function.Instructions[block.Instructions[index]].Order = EvaluationOrder(order)
}
func (cohereGraph) SetTerminalOrder(block *BasicBlock, order ssa.EvaluationOrder) {
	switch terminal := block.Terminal.(type) {
	case *Return:
		terminal.Order = EvaluationOrder(order)
	case *Unreachable:
		terminal.Order = EvaluationOrder(order)
	case *Goto:
		terminal.Order = EvaluationOrder(order)
	case *If:
		terminal.Order = EvaluationOrder(order)
	case *MayThrow:
		terminal.Order = EvaluationOrder(order)
	case *Throw:
		terminal.Order = EvaluationOrder(order)
	case *Choose:
		terminal.Order = EvaluationOrder(order)
	case *Yield:
		terminal.Order = EvaluationOrder(order)
	}
	// Suspend is backfilled by Build, as before.
}

func (cohereGraph) Placeholder(function *Function, block *BasicBlock) *BasicBlock {
	panic("flow: cohereGraph.Placeholder unused until #5jxpf6r head 2")
}

func (cohereGraph) Phis(block *BasicBlock) []*ssa.Phi[Place] {
	panic("flow: cohereGraph.Phis unused until #5jxpf6r head 2")
}

func (cohereGraph) SetPhis(block *BasicBlock, phis []*ssa.Phi[Place]) {
	panic("flow: cohereGraph.SetPhis unused until #5jxpf6r head 2")
}

func (cohereGraph) EndsInReturn(block *BasicBlock) bool {
	panic("flow: cohereGraph.EndsInReturn unused until #5jxpf6r head 2")
}

func (cohereGraph) EachInstructionPlace(function *Function, block *BasicBlock, index int, visit func(place *Place, role ssa.Role)) {
	panic("flow: cohereGraph.EachInstructionPlace unused until #5jxpf6r head 2")
}

func (cohereGraph) IsContextStore(function *Function, block *BasicBlock, index int) bool {
	panic("flow: cohereGraph.IsContextStore unused until #5jxpf6r head 2")
}

func (cohereGraph) ContextStoreDefines(function *Function, block *BasicBlock, index int, place Place) bool {
	panic("flow: cohereGraph.ContextStoreDefines unused until #5jxpf6r head 2")
}

func (cohereGraph) EachTerminalPlace(block *BasicBlock, visit func(place *Place, role ssa.Role)) {
	panic("flow: cohereGraph.EachTerminalPlace unused until #5jxpf6r head 2")
}

func (cohereGraph) Parameters(function *Function) []Place {
	panic("flow: cohereGraph.Parameters unused until #5jxpf6r head 2")
}

func (cohereGraph) Returns(function *Function) *Place {
	panic("flow: cohereGraph.Returns unused until #5jxpf6r head 2")
}

func (cohereGraph) Declaration(function *Function, identifier ssa.IdentifierId) ssa.DeclarationId {
	panic("flow: cohereGraph.Declaration unused until #5jxpf6r head 2")
}

func (cohereGraph) Contextual(function *Function, declaration ssa.DeclarationId) bool {
	panic("flow: cohereGraph.Contextual unused until #5jxpf6r head 2")
}

func (cohereGraph) Mint(function *Function, original ssa.IdentifierId) ssa.IdentifierId {
	panic("flow: cohereGraph.Mint unused until #5jxpf6r head 2")
}

func (cohereGraph) Named(function *Function, identifier ssa.IdentifierId) bool {
	panic("flow: cohereGraph.Named unused until #5jxpf6r head 2")
}

func (cohereGraph) PlaceString(function *Function, identifier ssa.IdentifierId) string {
	panic("flow: cohereGraph.PlaceString unused until #5jxpf6r head 2")
}

func (cohereGraph) IdentifierOf(place Place) ssa.IdentifierId {
	panic("flow: cohereGraph.IdentifierOf unused until #5jxpf6r head 2")
}

func (cohereGraph) WithIdentifier(place Place, identifier ssa.IdentifierId) Place {
	panic("flow: cohereGraph.WithIdentifier unused until #5jxpf6r head 2")
}
