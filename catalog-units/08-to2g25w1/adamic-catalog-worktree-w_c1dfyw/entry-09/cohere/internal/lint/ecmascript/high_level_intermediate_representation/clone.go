// Copying a finished graph, so a pass that rewrites one can start from the shared lowering.
package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/static_single_assignment"
	"maps"
	"slices"
)

// CloneFunction returns a function that shares no storage a pass can write with its argument.
//
// The shared lowering (`ForFunction`) is read-only by contract, and the preserve-manual-memoization
// pipeline rewrites the graph it is handed from its first pass to its last. Lowering the function a
// second time was the only way to give it a graph of its own, and that paid for the control-flow
// builder, reverse postorder and single-assignment construction a second time. A copy pays for the
// finished tables only.
//
// What is shared and why: `*ast.Node` is the program, which no pass writes, and the reactive scopes
// live in their own side table. Everything else is copied,
// blocks, instructions, identifiers, phis and nested functions, and the instruction values and
// terminals go through the same reflective copy the inliner uses, for the reason `deepCopyAny` gives.
//
// `blocksById` is rebuilt from the copied blocks rather than copied, because it aliases `Blocks`: a
// copied map would point at the original's blocks, and the first block a pass rewrote would be
// rewritten in the cached graph.
func CloneFunction(function *Function) *Function {
	if function == nil {
		return nil
	}
	clone := &Function{
		ContextDeclarations: maps.Clone(function.ContextDeclarations),
		Node:                function.Node,
		Name:                function.Name,
		Kind:                function.Kind,
		Params:              append([]Place(nil), function.Params...),
		Returns:             function.Returns,
		Context:             append([]Place(nil), function.Context...),
		Entry:               function.Entry,
		Outlined:            maps.Clone(function.Outlined),
		IsAsync:             function.IsAsync,
		IsGenerator:         function.IsGenerator,
		nextBlock:           function.nextBlock,
	}

	copiedBlocks := make(map[*BasicBlock]*BasicBlock, len(function.blocksById))
	copyBlock := func(block *BasicBlock) *BasicBlock {
		if block == nil {
			return nil
		}
		if copied, found := copiedBlocks[block]; found {
			return copied
		}
		copied := &BasicBlock{
			Id:           block.Id,
			Kind:         block.Kind,
			Instructions: append([]InstructionId(nil), block.Instructions...),
			Terminal:     copyTerminal(block.Terminal),
			Predecessors: append([]static_single_assignment.BlockId(nil), block.Predecessors...),
		}
		if block.Phis != nil {
			copied.Phis = make([]*Phi, len(block.Phis))
			for index, phi := range block.Phis {
				if phi != nil {
					copied.Phis[index] = &Phi{Place: phi.Place, Operands: slices.Clone(phi.Operands)}
				}
			}
		}
		copiedBlocks[block] = copied
		return copied
	}
	if function.Blocks != nil {
		clone.Blocks = make([]*BasicBlock, len(function.Blocks))
		for index, block := range function.Blocks {
			clone.Blocks[index] = copyBlock(block)
		}
	}
	if function.blocksById != nil {
		clone.blocksById = make(map[static_single_assignment.BlockId]*BasicBlock, len(function.blocksById))
		for id, block := range function.blocksById {
			clone.blocksById[id] = copyBlock(block)
		}
	}

	// Instructions and identifiers are copied into one backing slice each, so the copy costs two
	// allocations for its tables rather than one per entry. The backing is never appended to, so the
	// pointers into it stay put, and an identifier or instruction a pass mints later is allocated on
	// its own as it is in a lowering.
	if function.Instructions != nil {
		clone.Instructions = make([]*Instruction, len(function.Instructions))
		backing := make([]Instruction, len(function.Instructions))
		for index, instruction := range function.Instructions {
			if instruction == nil {
				continue
			}
			backing[index] = *instruction
			backing[index].Value = copyInstructionValue(instruction.Value)
			clone.Instructions[index] = &backing[index]
		}
	}
	if function.Identifiers != nil {
		clone.Identifiers = make([]*Identifier, len(function.Identifiers))
		backing := make([]Identifier, len(function.Identifiers))
		for index, identifier := range function.Identifiers {
			if identifier == nil {
				continue
			}
			backing[index] = *identifier
			clone.Identifiers[index] = &backing[index]
		}
	}
	if function.Functions != nil {
		clone.Functions = make([]*Function, len(function.Functions))
		for index, nested := range function.Functions {
			clone.Functions[index] = CloneFunction(nested)
		}
	}
	return clone
}
