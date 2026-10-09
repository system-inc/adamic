// Package static_single_assignment is the graph maintenance and single static assignment form two
// intermediate representations share: cohere's high-level IR (the React Compiler's, in
// internal/lint/ecmascript/high_level_intermediate_representation) and Adamic's flow graph.
//
// It owns the algorithms and none of either IR's meaning. Reverse postorder, predecessors, evaluation
// order, Braun's construction, redundant phi elimination and the verifier are written here once, over
// Graph: an adapter each IR implements on its own function, block and place types. Where the two IRs
// genuinely differ, the difference is a method on Graph rather than a fork of the algorithm:
//
//   - Structural fallthroughs. cohere's terminals name the block after an if or a loop beside their
//     real edges; Adamic's have only real edges. EachEdge reports which is which.
//   - Exception edges. Adamic's MayThrow leaves its block from the middle, at the throwing
//     instruction; cohere emits no such terminal. EachEdge reports them as Exceptional, and every
//     pass reads them as real edges today. See Construct for the one place that will change (#2yz9ra9).
//   - Context stores. A binding a closure captures is defined once in cohere, as upstream defines it;
//     Adamic keeps captured variables out of the graph. IsContextStore and Contextual.
//   - The Returns place, which cohere has and Adamic does not. Returns.
//   - Nested functions, which each IR recurses into itself, around these passes.
//
// It is a module of its own, beside the TypeScript shims, because Adamic consumes it by reference
// through its cohere submodule, and Go forbids importing cohere's internal/ from another module.
package static_single_assignment

// BlockId identifies a basic block within one function.
//
// Ids are dense and assigned in creation order, which is not execution order. A function's blocks are
// held in reverse postorder once ReversePostorder has run, so a walk that wants execution order
// iterates the slice, and one that wants to address a specific block holds the id. The two are
// deliberately different types of thing: an index into the block slice moves when the order changes,
// an id does not.
type BlockId uint32

// IdentifierId names one value.
//
// Two places with the same IdentifierId are the same value: not equal, the same. That is the property
// the whole IR exists to provide, and it is what lets a later pass ask whether the object mutated here
// is the object passed there without re-deriving aliasing from syntax.
type IdentifierId uint32

// DeclarationId names one source-level binding across all the values it takes.
//
// One `let` reassigned three times is three IdentifierIds and one DeclarationId. Single-assignment
// construction mints further IdentifierIds against the same DeclarationId, so a pass that wants "this
// variable" rather than "this value" asks for the DeclarationId and keeps working afterwards.
type DeclarationId uint32

// EvaluationOrder is a monotonically increasing position in the function's evaluation.
//
// Terminals carry one as well as instructions, so "did A evaluate before B" is answerable across an
// instruction and a terminal. Assigned by MarkEvaluationOrder in reverse postorder from 1. Zero means
// unassigned.
type EvaluationOrder uint32

// Edge is what kind of edge a successor is.
type Edge uint8

const (
	// Real is an edge control takes.
	Real Edge = iota

	// Fallthrough is a structural link, not an edge: the block after an if or a loop, which a
	// structured terminal names whether or not control can reach it from here. It is never a
	// predecessor. ReversePostorder visits it first so loop bodies and conditional arms precede their
	// continuation, and keeps it as an empty placeholder when nothing real reaches it.
	Fallthrough

	// Exceptional is an edge taken when the block's last instruction throws, before it completes.
	// Every pass treats it as Real today. It is named apart because single assignment should give the
	// handler the definitions from before the throwing instruction, not the block's end (#2yz9ra9).
	Exceptional
)

// Role says whether a place reads its value or defines it.
type Role uint8

const (
	// Use reads the value.
	Use Role = iota
	// Define writes it.
	Define
)

// Graph is what the passes need of an intermediate representation: its function type F, its block
// type B and its place type P.
//
// An IR implements it on a type of its own, usually an empty struct, and calls a pass with that value
// and its function. The passes hold no IR type; every read and write of one goes through here.
type Graph[F any, B comparable, P any] interface {
	// Entry is the block control begins at.
	Entry(function F) BlockId
	// BlockBound is one past the largest block id the function has handed out or holds.
	BlockBound(function F) int
	// Block is the block with an id, and false for one the function does not hold.
	Block(function F, id BlockId) (B, bool)
	// Blocks is the function's block slice, in its current order.
	Blocks(function F) []B
	// SetBlocks replaces the block slice.
	SetBlocks(function F, blocks []B)
	// Retain forgets every block whose id keep declines, so Block no longer finds it.
	Retain(function F, keep func(id BlockId) bool)
	// Placeholder replaces a block nothing real reaches but a structured terminal still names with an
	// empty unreachable block under the same id, keeping its predecessors, and returns it. Only an IR
	// that reports Fallthrough edges is asked.
	Placeholder(function F, block B) B

	// Id is a block's id.
	Id(block B) BlockId
	// Predecessors are the blocks with a real edge into a block.
	Predecessors(block B) []BlockId
	// SetPredecessors replaces them.
	SetPredecessors(block B, predecessors []BlockId)
	// Phis are a block's merges.
	Phis(block B) []*Phi[P]
	// SetPhis replaces them.
	SetPhis(block B, phis []*Phi[P])
	// EachEdge visits a block's successors: its structural fallthrough first, if it has one, then its
	// edges in the order its terminal names them.
	EachEdge(block B, visit func(successor BlockId, edge Edge))
	// EndsInReturn reports whether a block's terminal returns from the function.
	EndsInReturn(block B) bool

	// InstructionCount is how many instructions a block holds.
	InstructionCount(function F, block B) int
	// EachInstructionPlace visits the places of a block's instruction at index, uses before
	// definitions, by pointer so a pass may rename them.
	EachInstructionPlace(function F, block B, index int, visit func(place *P, role Role))
	// IsContextStore reports whether a block's instruction at index writes a binding a closure
	// captures, which single assignment defines once rather than versioning.
	IsContextStore(function F, block B, index int) bool
	// ContextStoreDefines reports whether a place a context store defines is the instruction's own
	// result, the one definition upstream's invariant counts. Asked only of a context store.
	ContextStoreDefines(function F, block B, index int, place P) bool
	// SetInstructionOrder sets the evaluation order of a block's instruction at index.
	SetInstructionOrder(function F, block B, index int, order EvaluationOrder)
	// EachTerminalPlace visits the places of a block's terminal, by pointer.
	EachTerminalPlace(block B, visit func(place *P, role Role))
	// SetTerminalOrder sets the evaluation order of a block's terminal.
	SetTerminalOrder(block B, order EvaluationOrder)

	// Params are the function's parameters, defined on entry, in a slice whose elements a pass renames.
	Params(function F) []P
	// Returns is the place every return stores into, or nil for an IR with none.
	Returns(function F) *P
	// Declaration is the binding an identifier is a value of.
	Declaration(function F, identifier IdentifierId) DeclarationId
	// Contextual reports whether a binding is one a nested function captures.
	Contextual(function F, declaration DeclarationId) bool
	// Mint makes a new value of the original's binding and returns its id.
	Mint(function F, original IdentifierId) IdentifierId
	// Named reports whether a value carries a source name rather than being a temporary.
	Named(function F, identifier IdentifierId) bool
	// PlaceString renders a value for a message.
	PlaceString(function F, identifier IdentifierId) string

	// IdentifierOf is the value a place names.
	IdentifierOf(place P) IdentifierId
	// WithIdentifier is a place naming another value, with everything else it carries unchanged.
	WithIdentifier(place P, identifier IdentifierId) P
}
