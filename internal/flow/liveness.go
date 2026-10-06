package flow

// LiveOut is, for every instruction, the variables whose current value something may still read
// after it: classic backward liveness over the graph, by variable rather than by value, so it can be
// run on a graph before or after single assignment. Reuse in place reads it (internal/native): a
// value whose variable isn't live after an instruction is dead there, and its memory may be taken.
//
// Only the variables the graph tracks are here (flow.go): a global or a captured variable is never
// in a set, and a caller must treat it as always live.
func LiveOut(function *Function) map[InstructionId]map[DeclarationId]bool {
	declaration := func(place Place) DeclarationId { return function.Identifiers[place.Identifier].Declaration }
	liveIn := map[BlockId]map[DeclarationId]bool{}
	result := map[InstructionId]map[DeclarationId]bool{}
	for changed := true; changed; {
		changed = false
		// Backward, so reverse postorder reversed reaches a fixed point in few passes.
		for index := len(function.Blocks) - 1; index >= 0; index-- {
			block := function.Blocks[index]
			live := map[DeclarationId]bool{}
			EachSuccessor(block.Terminal, func(successor BlockId) {
				for variable := range liveIn[successor] {
					live[variable] = true
				}
			})
			for position := len(block.Instructions) - 1; position >= 0; position-- {
				id := block.Instructions[position]
				after := make(map[DeclarationId]bool, len(live))
				for variable := range live {
					after[variable] = true
				}
				result[id] = after
				instruction := function.Instructions[id]
				for _, define := range instruction.Defines {
					delete(live, declaration(define))
				}
				// An instruction that throws never gives its variables their values, so on the way to
				// the handler what was live there is still live before it: in local = f() with a catch
				// that reads local, local's old value must outlive the call.
				if throws, ok := block.Terminal.(*MayThrow); ok && position == len(block.Instructions)-1 {
					for variable := range liveIn[throws.Handler] {
						live[variable] = true
					}
				}
				for _, use := range instruction.Uses {
					live[declaration(use)] = true
				}
			}
			if len(live) != len(liveIn[block.Id]) {
				liveIn[block.Id] = live
				changed = true
			}
		}
	}
	return result
}
