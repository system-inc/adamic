// Package flow is a control-flow graph of one Adamic function, built from the typed IR, for the
// analyses that need to know which value reaches where: single assignment first, then the aliasing
// and mutable ranges that reuse in place and arenas are built on (docs/memory.md, #5jck546).
//
// It is not a second lowering. Every semantic decision (JavaScript's evaluation order, the snapshots,
// the inserted checks) was made once, by internal/lower, and the oracle holds it there. A block here
// holds instructions that each point back at the IR statement or condition they came from, and say
// only which locals that piece of the program reads and which one it defines. That is all single
// assignment asks.
//
// The analyses are lifted from cohere's high-level IR (cohere/internal/lint/ecmascript/
// high_level_intermediate_representation), which ports the React Compiler's. Names that mean the same
// thing there keep cohere's spelling, so a lifted pass reads side by side with its original; ssa.go,
// ssa_eliminate.go, ssa_verify.go and graph.go say what changed from cohere's and why.
//
// # Which locals are values here
//
// A local the graph tracks is one only its own function can write: not a global (any call may write
// one) and not a captured variable (it lives in a cell any closure holding it may write). Those two
// are left out of the graph entirely, as cohere leaves a LoadGlobal unrenamed: there is no point in
// this function at which their value is known, so a phi over one would be an invention.
package flow

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// BlockId identifies a basic block within one Function. Ids are dense and in creation order;
// Function.Blocks is in reverse postorder once Finalize has run. 0 is InvalidBlock.
type BlockId uint32

// InvalidBlock is the zero BlockId, never a real block's.
const InvalidBlock BlockId = 0

// InstructionId indexes Function.Instructions.
type InstructionId uint32

// IdentifierId names one value. Two places with the same IdentifierId are the same value.
type IdentifierId uint32

// EvaluationOrder is a position in the function's evaluation, assigned by MarkEvaluationOrder in
// reverse postorder from 1; zero means unassigned. A back edge shows as a decrease.
type EvaluationOrder uint32

// DeclarationId names one variable across every value it takes: the IR's local index, plus one, so
// that zero is never a variable. A temporary the effect inference mints (infer.go) has zero.
type DeclarationId uint32

// Function is one IR function (or the program's top level) as a control-flow graph.
type Function struct {
	// Name is the IR function's name, or "main" for the top level, for reading only.
	Name string

	// Program is the IR program the function is from: its locals' types, and the statements its
	// instructions point back at.
	Program *ir.Program

	// Params are the parameters the graph tracks, in order, defined on entry.
	Params []Place

	// Entry is the block control begins at.
	Entry BlockId

	// Blocks are the basic blocks, in reverse postorder after Finalize.
	Blocks []*BasicBlock

	// Instructions is the flat instruction table. Blocks hold indices into it.
	Instructions []*Instruction

	// Identifiers names every value. Indexed by IdentifierId.
	Identifiers []*Identifier

	blocksById map[BlockId]*BasicBlock
	nextBlock  BlockId
}

// BasicBlock is a straight-line run of instructions ending in exactly one terminal.
type BasicBlock struct {
	Id BlockId

	// Instructions are ids into Function.Instructions, in evaluation order.
	Instructions []InstructionId

	// Terminal is how control leaves. Never nil in a finished function.
	Terminal Terminal

	// Predecessors are the blocks with an edge into this one, maintained by MarkPredecessors.
	Predecessors []BlockId

	// Phis are the merges single assignment placed: empty until Construct runs.
	Phis []*Phi
}

// Phi is a merge: the value of Place at the top of a block, given which predecessor control came
// from. Read the operands in a fixed order through PhiOperandsInOrder, never by ranging the map.
type Phi struct {
	Place    Place
	Operands map[BlockId]Place
}

// Instruction is one piece of an IR statement: the locals it reads, then the one it defines, if any.
type Instruction struct {
	Id InstructionId

	// Order is where this evaluates relative to everything else in the function.
	Order EvaluationOrder

	// At is the IR statement this came from, where it stands in its slice, and Part which piece of it
	// this is: 0 for the statement itself (or its condition, a for...of's iterable, a switch's value),
	// 1 for a for...of binding its next element, and 1, 2 and on for a switch's case tests in order.
	// The pair names the same point the JavaScript backend's trace marks (javascript.Options.Mark).
	At   *ir.Statement
	Part int

	// Expression, when set, is what this instruction evaluates.
	Expression ir.Expression

	// Uses are the tracked locals read, each once per read, in the order the IR holds them. Nothing
	// inside an expression writes a local, so their order doesn't change which value each one reads.
	Uses []Place

	// Defines are the locals this writes: a declaration, an assignment, or a for...of's binding (one,
	// or a pattern's several).
	Defines []Place
}

// Identifier names one value and the variable it belongs to.
type Identifier struct {
	Id          IdentifierId
	Declaration DeclarationId

	// Name is the variable's name as written, for reading only.
	Name string
}

// Place is a reference to a value at one point.
type Place struct {
	Identifier IdentifierId
}

// PlaceRole says whether a place reads its value or defines it.
type PlaceRole uint8

const (
	PlaceRoleUse PlaceRole = iota
	PlaceRoleDefine
)

// Terminal is how control leaves a block.
type Terminal interface{ terminal() }

type (
	// Goto continues at Block.
	Goto struct {
		Block BlockId
		Order EvaluationOrder
	}

	// If continues at Consequent or Alternate, on what the block's last instruction decided: a
	// condition, a switch's case test, or (for a for...of's head) whether there's a next element.
	If struct {
		Consequent, Alternate BlockId
		Order                 EvaluationOrder
	}

	// Return leaves the function. Its value, if any, is read by the block's last instruction.
	Return struct{ Order EvaluationOrder }

	// Unreachable ends a path that never continues: a panic.
	Unreachable struct{ Order EvaluationOrder }

	// MayThrow ends a block whose last instruction can throw (a throw statement aside, which goes
	// straight to its handler): control goes on at Next, or, when it throws, at Handler, the catch or
	// finally around it, or the block that throws out of the function.
	MayThrow struct {
		Next, Handler BlockId
		Order         EvaluationOrder
	}

	// Throw leaves the function by an exception: its caller's handler takes it.
	Throw struct{ Order EvaluationOrder }

	// Choose goes on at any of Blocks: the end of a finally, which goes on wherever whatever entered
	// it was going (after the try, a rethrow, a break, a continue, a return).
	Choose struct {
		Blocks []BlockId
		Order  EvaluationOrder
	}
)

func (*Goto) terminal()        {}
func (*If) terminal()          {}
func (*Return) terminal()      {}
func (*Unreachable) terminal() {}
func (*MayThrow) terminal()    {}
func (*Throw) terminal()       {}
func (*Choose) terminal()      {}

// Suspend has two explicit resumption edges, independent of the operand call's synchronous effects.
type Suspend struct {
	Fulfilled, Rejected BlockId
	Local               int
	Of                  ir.Type
	Order               EvaluationOrder
}

func (*Suspend) terminal() {}

// NewFunction makes an empty function ready to build.
func NewFunction(name string) *Function {
	return &Function{Name: name, blocksById: map[BlockId]*BasicBlock{}, nextBlock: 1}
}

// NewBlock appends a block and returns it.
func (f *Function) NewBlock() *BasicBlock {
	block := &BasicBlock{Id: f.nextBlock}
	f.nextBlock++
	f.Blocks = append(f.Blocks, block)
	f.blocksById[block.Id] = block
	return block
}

// Block returns the block with the given id, and false for one that isn't there (InvalidBlock, or
// one Finalize removed as unreachable).
func (f *Function) Block(id BlockId) (*BasicBlock, bool) {
	block, ok := f.blocksById[id]
	return block, ok
}

// NewIdentifier mints a new value of a variable.
func (f *Function) NewIdentifier(name string, declaration DeclarationId) *Identifier {
	identifier := &Identifier{Id: IdentifierId(len(f.Identifiers)), Declaration: declaration, Name: name}
	f.Identifiers = append(f.Identifiers, identifier)
	return identifier
}

// AddInstruction appends an instruction to the table and to a block.
func (f *Function) AddInstruction(block *BasicBlock, instruction *Instruction) InstructionId {
	id := InstructionId(len(f.Instructions))
	instruction.Id = id
	f.Instructions = append(f.Instructions, instruction)
	block.Instructions = append(block.Instructions, id)
	return id
}

// PlaceString renders a place as its variable's name and its value's id.
func (f *Function) PlaceString(place Place) string {
	identifier := f.Identifiers[place.Identifier]
	return fmt.Sprintf("%s$%d", identifier.Name, identifier.Id)
}

// EachInstructionPlacePointer visits an instruction's places, uses first, so a pass may rewrite them.
func EachInstructionPlacePointer(instruction *Instruction, visit func(place *Place, role PlaceRole)) {
	for index := range instruction.Uses {
		visit(&instruction.Uses[index], PlaceRoleUse)
	}
	for index := range instruction.Defines {
		visit(&instruction.Defines[index], PlaceRoleDefine)
	}
}

// EachInstructionPlace visits an instruction's places, uses first.
func EachInstructionPlace(instruction *Instruction, visit func(place Place, role PlaceRole)) {
	EachInstructionPlacePointer(instruction, func(place *Place, role PlaceRole) { visit(*place, role) })
}

// EachTerminalPlace visits a terminal's places: none, as EachTerminalPlacePointer says.
func EachTerminalPlace(terminal Terminal, visit func(place Place, role PlaceRole)) {}

// EachTerminalPlacePointer visits a terminal's places. Adamic's terminals hold none: whatever a
// branch tests or a return returns is read by an instruction just before it.
func EachTerminalPlacePointer(terminal Terminal, visit func(place *Place, role PlaceRole)) {}

// EachSuccessor visits every block control can go to next.
func EachSuccessor(terminal Terminal, visit func(block BlockId)) {
	switch terminal := terminal.(type) {
	case *Goto:
		visit(terminal.Block)
	case *If:
		visit(terminal.Consequent)
		visit(terminal.Alternate)
	case *Suspend:
		visit(terminal.Fulfilled)
		visit(terminal.Rejected)
	case *MayThrow:
		visit(terminal.Next)
		visit(terminal.Handler)
	case *Choose:
		for _, block := range terminal.Blocks {
			visit(block)
		}
	case *Return, *Unreachable, *Throw:
	default:
		panic(fmt.Sprintf("flow: no successors known for %T", terminal))
	}
}

// TerminalOrder is a terminal's position in evaluation order.
func TerminalOrder(terminal Terminal) EvaluationOrder {
	switch terminal := terminal.(type) {
	case *Goto:
		return terminal.Order
	case *If:
		return terminal.Order
	case *Return:
		return terminal.Order
	case *Unreachable:
		return terminal.Order
	case *Suspend:
		return terminal.Order
	case *MayThrow:
		return terminal.Order
	case *Throw:
		return terminal.Order
	case *Choose:
		return terminal.Order
	}
	return 0
}
