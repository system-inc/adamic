package flow

import (
	"fmt"
	"slices"
	"testing"
)

type graphOracleBlock struct {
	id           BlockId
	predecessors []BlockId
	instructions []EvaluationOrder
	terminal     EvaluationOrder
}

type graphOracleSnapshot struct {
	blocks       []graphOracleBlock
	instructions []EvaluationOrder
}

func snapshotGraphOracle(function *Function) graphOracleSnapshot {
	var snapshot graphOracleSnapshot
	for _, instruction := range function.Instructions {
		snapshot.instructions = append(snapshot.instructions, instruction.Order)
	}
	for _, block := range function.Blocks {
		predecessors := slices.Clone(block.Predecessors)
		slices.Sort(predecessors)
		entry := graphOracleBlock{id: block.Id, predecessors: predecessors, terminal: TerminalOrder(block.Terminal)}
		for _, id := range block.Instructions {
			entry.instructions = append(entry.instructions, function.Instructions[id].Order)
		}
		snapshot.blocks = append(snapshot.blocks, entry)
	}
	return snapshot
}

func compareGraphOracle(t *testing.T, where string, got, want graphOracleSnapshot) {
	t.Helper()
	if len(got.blocks) != len(want.blocks) {
		t.Fatalf("%s: block count: Build=%d reference=%d", where, len(got.blocks), len(want.blocks))
	}
	for index, block := range got.blocks {
		reference := want.blocks[index]
		if block.id != reference.id {
			t.Fatalf("%s: block order at %d: Build=bb%d reference=bb%d", where, index, block.id, reference.id)
		}
		if !slices.Equal(block.predecessors, reference.predecessors) {
			t.Fatalf("%s: bb%d predecessors: Build=%v reference=%v", where, block.id, block.predecessors, reference.predecessors)
		}
		if !slices.Equal(block.instructions, reference.instructions) {
			t.Fatalf("%s: bb%d instruction evaluation orders: Build=%v reference=%v", where, block.id, block.instructions, reference.instructions)
		}
		if block.terminal != reference.terminal {
			t.Fatalf("%s: bb%d terminal evaluation order: Build=%d reference=%d", where, block.id, block.terminal, reference.terminal)
		}
	}
	for id, order := range got.instructions {
		if order != want.instructions[id] {
			t.Fatalf("%s: instruction %d evaluation order: Build=%d reference=%d", where, id, order, want.instructions[id])
		}
	}
}

func TestGraphOracle(t *testing.T) {
	var functions, blocks, mayThrowFunctions, mayThrows int
	for _, path := range programs(t) {
		program := lowered(t, path)
		for index := -1; index < len(program.Functions); index++ {
			function := Build(program, index)
			where := fmt.Sprintf("%s, function %d (%s)", path, index, function.Name)
			before := snapshotGraphOracle(function)
			functions++
			blocks += len(function.Blocks)
			hasMayThrow := false
			for _, block := range function.Blocks {
				if _, ok := block.Terminal.(*MayThrow); ok {
					hasMayThrow = true
					mayThrows++
				}
				block.Predecessors = nil
				for _, id := range block.Instructions {
					function.Instructions[id].Order = 0
				}
				// Clear precisely the terminal kinds written by today's pass. Suspend's
				// Build backfill is retained because Finalize does not write it.
				referenceSetTerminalOrder(block.Terminal, 0)
			}
			if hasMayThrow {
				mayThrowFunctions++
			}
			// Remove the pass's ordering while retaining the same reachable block set.
			slices.SortFunc(function.Blocks, func(a, b *BasicBlock) int {
				if a.Id < b.Id {
					return -1
				}
				if a.Id > b.Id {
					return 1
				}
				return 0
			})
			referenceFinalize(function)
			compareGraphOracle(t, where, before, snapshotGraphOracle(function))
		}
	}
	if mayThrowFunctions == 0 {
		t.Fatal("corpus has no functions with MayThrow handler edges")
	}
	t.Logf("graph oracle: %d functions, %d blocks, %d MayThrow terminals in %d functions", functions, blocks, mayThrows, mayThrowFunctions)
}
