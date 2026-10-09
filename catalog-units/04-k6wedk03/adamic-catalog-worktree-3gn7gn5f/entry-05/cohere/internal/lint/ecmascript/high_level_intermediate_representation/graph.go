// Graph maintenance: the passes lowering runs to bring a freshly built function into the state
// every other pass assumes.
//
// Three invariants are established here and every consumer may rely on them:
//
//  1. Function.Blocks is in reverse postorder, and unreachable blocks have been removed.
//  2. Every block's Predecessors is exactly the set of blocks with a real edge into it.
//  3. Every instruction and terminal has a nonzero, monotonically increasing EvaluationOrder.
//
// They are re-established by calling Finalize, which any pass that restructures the graph must do.
//
// The passes themselves are static_single_assignment's, which this IR shares with Adamic's flow graph.
// What lives here is ssaGraph, this IR's side of that module's Graph: how it reads and writes a
// Function, a BasicBlock and a Place.
package high_level_intermediate_representation

import "github.com/system-inc/cohere/static_single_assignment"

// Finalize brings a freshly lowered function into the state every pass assumes.
//
// Order matters: reverse postorder first, because it drops unreachable blocks and the predecessor
// edges must not name a block that is gone; evaluation order last, because it walks the blocks in
// their final order.
func Finalize(function *Function) {
	ReversePostorder(function)
	MarkPredecessors(function)
	MarkEvaluationOrder(function)
	for _, nested := range function.Functions {
		Finalize(nested)
	}
}

// ReversePostorder reorders Function.Blocks into reverse postorder and drops unreachable blocks.
//
// Upstream's `getReversePostorderedBlocks`, a structural fallthrough visited first so loop bodies and
// conditional arms precede their continuation, and a fallthrough nothing real reaches kept as an empty
// Unreachable block. static_single_assignment.ReversePostorder carries the reasoning.
func ReversePostorder(function *Function) {
	if function == nil {
		return
	}
	static_single_assignment.ReversePostorder(ssaGraph{}, function)
}

// MarkPredecessors recomputes every block's Predecessors from the real edges.
//
// Fallthroughs are NOT edges and do not produce a predecessor. See EachSuccessor.
//
// This is the field single-assignment construction needs: Braun's algorithm reads a value at a
// block join by asking each predecessor what it holds, and a missing or spurious predecessor is a
// wrong phi rather than a crash.
func MarkPredecessors(function *Function) {
	static_single_assignment.MarkPredecessors(ssaGraph{}, function)
}

// MarkEvaluationOrder assigns each instruction and terminal its position in evaluation order.
//
// Numbering follows Function.Blocks, which is reverse postorder, so a forward analysis sees
// increasing numbers along any acyclic path. Across a back edge the number decreases, which is the
// correct and expected signal that a loop was traversed.
//
// Starts at 1: zero means unassigned, and a pass that reads an order of zero has found a block the
// finalizer did not reach.
func MarkEvaluationOrder(function *Function) {
	static_single_assignment.MarkEvaluationOrder(ssaGraph{}, function)
}

func setTerminalOrder(terminal Terminal, order static_single_assignment.EvaluationOrder) {
	switch t := terminal.(type) {
	case *Return:
		t.Order = order
	case *Throw:
		t.Order = order
	case *Unreachable:
		t.Order = order
	case *Unsupported:
		t.Order = order
	case *Goto:
		t.Order = order
	case *If:
		t.Order = order
	case *Branch:
		t.Order = order
	case *Switch:
		t.Order = order
	case *While:
		t.Order = order
	case *DoWhile:
		t.Order = order
	case *For:
		t.Order = order
	case *ForOf:
		t.Order = order
	case *ForIn:
		t.Order = order
	case *Logical:
		t.Order = order
	case *Ternary:
		t.Order = order
	case *Optional:
		t.Order = order
	case *Sequence:
		t.Order = order
	case *Label:
		t.Order = order
	case *Try:
		t.Order = order
	case *MaybeThrow:
		t.Order = order
	case *Scope:
		t.Order = order
	}
}

// ssaGraph is this IR's side of static_single_assignment.Graph.
//
// Where the two IRs that share the module differ, this is cohere's answer: a structured terminal's
// fallthrough is reported as one, MaybeThrow's handler as an exceptional edge, a StoreContext as a
// context store, and the Returns place is the function's own. Each method is a field read or the
// visitor this package already has, so the passes see exactly what they read before they moved.
type ssaGraph struct{}

func (ssaGraph) Entry(function *Function) static_single_assignment.BlockId { return function.Entry }

func (ssaGraph) BlockBound(function *Function) int {
	// Ids are bounded by `nextBlock`, and by the largest id in the table for a function built by
	// hand without `NewFunction`.
	bound := int(function.nextBlock)
	for id := range function.blocksById {
		if int(id) >= bound {
			bound = int(id) + 1
		}
	}
	return bound
}

func (ssaGraph) Block(function *Function, id static_single_assignment.BlockId) (*BasicBlock, bool) {
	return function.Block(id)
}

func (ssaGraph) Blocks(function *Function) []*BasicBlock { return function.Blocks }

func (ssaGraph) SetBlocks(function *Function, blocks []*BasicBlock) { function.Blocks = blocks }

func (ssaGraph) Retain(function *Function, keep func(id static_single_assignment.BlockId) bool) {
	for id := range function.blocksById {
		if !keep(id) {
			delete(function.blocksById, id)
		}
	}
}

func (ssaGraph) Placeholder(function *Function, block *BasicBlock) *BasicBlock {
	placeholder := &BasicBlock{
		Id:           block.Id,
		Kind:         block.Kind,
		Terminal:     &Unreachable{},
		Predecessors: append([]static_single_assignment.BlockId(nil), block.Predecessors...),
	}
	function.blocksById[block.Id] = placeholder
	return placeholder
}

func (ssaGraph) Id(block *BasicBlock) static_single_assignment.BlockId { return block.Id }

func (ssaGraph) Predecessors(block *BasicBlock) []static_single_assignment.BlockId {
	return block.Predecessors
}

func (ssaGraph) SetPredecessors(block *BasicBlock, predecessors []static_single_assignment.BlockId) {
	block.Predecessors = predecessors
}

func (ssaGraph) Phis(block *BasicBlock) []*Phi { return block.Phis }

func (ssaGraph) SetPhis(block *BasicBlock, phis []*Phi) { block.Phis = phis }

// EachEdge reports the structural fallthrough first, then EachSuccessor's real edges in its order,
// with MaybeThrow's handler marked exceptional. Try's handler is a real edge: it leaves the block
// holding the Try, before the try body runs, so nothing in that block can throw on the way.
func (ssaGraph) EachEdge(block *BasicBlock, visit func(successor static_single_assignment.BlockId, edge static_single_assignment.Edge)) {
	if fallthroughBlock, ok := Fallthrough(block.Terminal); ok {
		visit(fallthroughBlock, static_single_assignment.Fallthrough)
	}
	handler := InvalidBlock
	if maybeThrow, ok := block.Terminal.(*MaybeThrow); ok {
		handler = maybeThrow.Handler
	}
	EachSuccessor(block.Terminal, func(successor static_single_assignment.BlockId) {
		if HasBlock(handler) && successor == handler {
			visit(successor, static_single_assignment.Exceptional)
			return
		}
		visit(successor, static_single_assignment.Real)
	})
}

func (ssaGraph) EndsInReturn(block *BasicBlock) bool {
	_, returns := block.Terminal.(*Return)
	return returns
}

func (ssaGraph) InstructionCount(function *Function, block *BasicBlock) int {
	return len(block.Instructions)
}

func (ssaGraph) EachInstructionPlace(function *Function, block *BasicBlock, index int,
	visit func(place *Place, role static_single_assignment.Role)) {
	EachInstructionPlacePointer(function.Instructions[block.Instructions[index]], func(place *Place, role PlaceRole) {
		visit(place, ssaRole(role))
	})
}

func (ssaGraph) IsContextStore(function *Function, block *BasicBlock, index int) bool {
	_, isContextStore := function.Instructions[block.Instructions[index]].Value.(*StoreContext)
	return isContextStore
}

func (ssaGraph) ContextStoreDefines(function *Function, block *BasicBlock, index int, place Place) bool {
	return place.Identifier == function.Instructions[block.Instructions[index]].LValue.Identifier
}

func (ssaGraph) SetInstructionOrder(function *Function, block *BasicBlock, index int, order static_single_assignment.EvaluationOrder) {
	function.Instructions[block.Instructions[index]].Order = order
}

func (ssaGraph) EachTerminalPlace(block *BasicBlock, visit func(place *Place, role static_single_assignment.Role)) {
	EachTerminalPlacePointer(block.Terminal, func(place *Place, role PlaceRole) {
		visit(place, ssaRole(role))
	})
}

func (ssaGraph) SetTerminalOrder(block *BasicBlock, order static_single_assignment.EvaluationOrder) {
	setTerminalOrder(block.Terminal, order)
}

func (ssaGraph) Params(function *Function) []Place { return function.Params }

func (ssaGraph) Returns(function *Function) *Place { return &function.Returns }

func (ssaGraph) Declaration(function *Function, identifier static_single_assignment.IdentifierId) static_single_assignment.DeclarationId {
	return function.Identifiers[identifier].Declaration
}

func (ssaGraph) Contextual(function *Function, declaration static_single_assignment.DeclarationId) bool {
	return function.ContextDeclarations[declaration]
}

// Mint creates a new value carrying the same source binding as the original.
//
// The DeclarationId is preserved, which is the whole point: after renaming, "which variable is
// this" is answered by the declaration and "which value is this" by the identifier, and a pass can
// ask either.
func (ssaGraph) Mint(function *Function, original static_single_assignment.IdentifierId) static_single_assignment.IdentifierId {
	source := function.Identifiers[original]
	return function.NewIdentifier(source.Name, source.Node, source.Declaration).Id
}

func (ssaGraph) Named(function *Function, identifier static_single_assignment.IdentifierId) bool {
	return function.Identifiers[identifier].Name != ""
}

func (ssaGraph) PlaceString(function *Function, identifier static_single_assignment.IdentifierId) string {
	return function.PlaceString(Place{Identifier: identifier})
}

func (ssaGraph) IdentifierOf(place Place) static_single_assignment.IdentifierId {
	return place.Identifier
}

// WithIdentifier keeps the place's effect, reactivity and range, which is what every renamed place
// and phi operand carried before the passes moved.
func (ssaGraph) WithIdentifier(place Place, identifier static_single_assignment.IdentifierId) Place {
	place.Identifier = identifier
	return place
}

// ssaRole is the module's two roles from this IR's three: a receiver is read, as every pass here has
// always treated it.
func ssaRole(role PlaceRole) static_single_assignment.Role {
	if role == PlaceRoleDefine {
		return static_single_assignment.Define
	}
	return static_single_assignment.Use
}
