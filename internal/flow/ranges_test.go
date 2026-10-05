package flow

import (
	"strconv"
	"strings"
	"sync"
	"testing"
)

// Every mutation Node makes falls inside the mutable range of the value it mutates. A range claims
// its value is settled from its end on, and this checks that claim against what runs: the trace
// reports each instruction after which a tracked variable's object (the same object it held) printed
// differently, and the value that variable held there must have a range containing that
// instruction. A mutation the alias graph missed (through a value it didn't know was the same, or a
// part of the same) lands outside the range and fails here.
//
// A range the pass left unset (a loop-carried value whose writes would invert it, which cohere
// names RangeGapLoopCarriedInversion) claims nothing, so a mutation there is the pass declining, not
// a wrong answer; those are counted and logged, and a consumer must decline on them too.
func TestEveryMutationIsInItsRange(t *testing.T) {
	t.Parallel()
	var lock sync.Mutex
	var mutations, declined, invalid int
	t.Run("programs", func(t *testing.T) {
		for _, path := range programs(t) {
			t.Run(path, func(t *testing.T) {
				t.Parallel()
				run := traced(t, path)
				ranges := map[int]*MutableRanges{}
				values := map[int]map[InstructionId]map[DeclarationId]IdentifierId{}
				for function, graph := range run.graphs {
					ranges[function] = InferMutableRanges(graph)
					values[function] = valuesBefore(graph)
					if bad := ValidateMutableRanges(ranges[function]); len(bad) > 0 {
						t.Errorf("function %d (%s): invalid ranges for %v", function, graph.Name, bad)
						lock.Lock()
						invalid += len(bad)
						lock.Unlock()
					}
				}
				seen, skipped := 0, 0
				for _, event := range run.events {
					if !strings.HasPrefix(event, "mutated ") {
						continue
					}
					fields := strings.Fields(event)
					local, _ := strconv.Atoi(fields[1])
					index, _ := strconv.Atoi(fields[2])
					at := run.marked[index]
					graph := run.graphs[at.function]
					value, ok := values[at.function][at.instruction][DeclarationId(local+1)]
					if !ok {
						t.Errorf("function %d (%s): %s was mutated at instruction %d, where no value of it reaches", at.function, graph.Name, graph.Program.Locals[local].Name, at.instruction)
						continue
					}
					seen++
					order := graph.Instructions[at.instruction].Order
					mutable := ranges[at.function].Get(value)
					switch {
					case !mutable.IsSet():
						skipped++
					case !mutable.Contains(order):
						t.Errorf("function %d (%s): %s was mutated at instruction %d (order %d), outside its range [%d, %d)",
							at.function, graph.Name, graph.PlaceString(Place{Identifier: value}), at.instruction, order, mutable.Start, mutable.End)
					}
				}
				lock.Lock()
				mutations += seen
				declined += skipped
				lock.Unlock()
			})
		}
	})
	// A run that saw no mutation proves nothing about ranges.
	if mutations == 0 {
		t.Errorf("no mutation was seen in any program")
	}
	t.Logf("%d mutations checked, %d on a range the pass left unset, %d invalid ranges", mutations, declined, invalid)
}

// valuesBefore is, for every instruction, the value each variable holds just before it, found by a
// forward walk over the constructed graph: within a block, the last definition; on entry, the block's
// phi, or what its walked predecessors agree on (a disagreement would have needed a phi).
func valuesBefore(graph *Function) map[InstructionId]map[DeclarationId]IdentifierId {
	result := map[InstructionId]map[DeclarationId]IdentifierId{}
	exits := map[BlockId]map[DeclarationId]IdentifierId{}
	for _, block := range graph.Blocks {
		current := map[DeclarationId]IdentifierId{}
		if block.Id == graph.Entry {
			for _, parameter := range graph.Params {
				current[graph.Identifiers[parameter.Identifier].Declaration] = parameter.Identifier
			}
		}
		for _, predecessor := range block.Predecessors {
			for variable, value := range exits[predecessor] {
				current[variable] = value
			}
		}
		for _, phi := range block.Phis {
			current[graph.Identifiers[phi.Place.Identifier].Declaration] = phi.Place.Identifier
		}
		for _, id := range block.Instructions {
			before := map[DeclarationId]IdentifierId{}
			for variable, value := range current {
				before[variable] = value
			}
			result[id] = before
			for _, define := range graph.Instructions[id].Defines {
				current[graph.Identifiers[define.Identifier].Declaration] = define.Identifier
			}
		}
		exits[block.Id] = current
	}
	return result
}
