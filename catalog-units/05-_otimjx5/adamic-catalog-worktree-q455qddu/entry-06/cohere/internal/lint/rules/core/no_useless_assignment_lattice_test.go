package core

import (
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// TestDeadStoreLiveSetsAreNeverWrittenInPlace holds what lets Meet hand back one of its arguments:
// nothing writes into a live set it did not just make. Two neighbours handing over one shared value
// are met, the result is transferred through a block that kills a symbol and reads another, and two
// sets neither of which holds the other are met. Every input must read as it did before.
func TestDeadStoreLiveSetsAreNeverWrittenInPlace(t *testing.T) {
	t.Parallel()
	lattice := deadStoreLiveness{reachable: map[int]bool{}, words: 1}

	shared := liveSymbols{0b011}
	met := lattice.Meet(shared, shared)
	block := &control_flow_graph.Block[deadStoreEvent]{Events: []deadStoreEvent{
		{kind: occurrenceRead, number: 2},
		{kind: occurrenceWrite, number: 0},
	}}
	transferred := lattice.Transfer(block, met)

	left, right := liveSymbols{0b001}, liveSymbols{0b010}
	merged := lattice.Meet(left, right)

	for _, check := range []struct {
		name      string
		got, want liveSymbols
	}{
		{"the shared value", shared, liveSymbols{0b011}},
		{"the meet of the shared value with itself", met, liveSymbols{0b011}},
		{"the transferred set", transferred, liveSymbols{0b110}},
		{"the left of two disjoint sets", left, liveSymbols{0b001}},
		{"the right of two disjoint sets", right, liveSymbols{0b010}},
		{"their meet", merged, liveSymbols{0b011}},
	} {
		if !slices.Equal(check.got, check.want) {
			t.Errorf("%s is %b, want %b", check.name, check.got, check.want)
		}
	}
}

// referenceLiveness is the map lattice the bitset replaced, kept here to hold the bitset to it.
type referenceLiveness struct{}

func (referenceLiveness) Bottom() map[*ast.Symbol]bool { return map[*ast.Symbol]bool{} }
func (referenceLiveness) Entry() map[*ast.Symbol]bool  { return map[*ast.Symbol]bool{} }
func (referenceLiveness) Meet(left, right map[*ast.Symbol]bool) map[*ast.Symbol]bool {
	merged := map[*ast.Symbol]bool{}
	for symbol := range left {
		merged[symbol] = true
	}
	for symbol := range right {
		merged[symbol] = true
	}
	return merged
}
func (referenceLiveness) Transfer(block *control_flow_graph.Block[deadStoreEvent], incoming map[*ast.Symbol]bool) map[*ast.Symbol]bool {
	outgoing := map[*ast.Symbol]bool{}
	for symbol := range incoming {
		outgoing[symbol] = true
	}
	for index := len(block.Events) - 1; index >= 0; index-- {
		switch event := block.Events[index]; event.kind {
		case occurrenceRead:
			outgoing[event.symbol] = true
		case occurrenceWrite:
			delete(outgoing, event.symbol)
		}
	}
	return outgoing
}
func (referenceLiveness) Equal(left, right map[*ast.Symbol]bool) bool {
	if len(left) != len(right) {
		return false
	}
	for symbol := range left {
		if !right[symbol] {
			return false
		}
	}
	return true
}

// TestDeadStoreBitsetSolvesAsTheMapDid solves each root of a file holding loops, a nested loop, a
// try with a finally, a switch that falls through and a closure, once with the bitset lattice and
// once with the map lattice it replaced, and requires the same number of rounds and the same live
// set on exit from every reachable block.
func TestDeadStoreBitsetSolvesAsTheMapDid(t *testing.T) {
	t.Parallel()
	source := "export function f(items: number[], flag: boolean) {\n" +
		"  let total = 0; let last = -1; let seen = 0;\n" +
		"  for (const item of items) {\n" +
		"    let inner = item;\n" +
		"    while (inner > 0) { inner -= 1; seen += 1; }\n" +
		"    total += item; last = item;\n" +
		"  }\n" +
		"  try { total = total * 2; } finally { last = 0; }\n" +
		"  switch (total) { case 1: seen = 1; case 2: seen = 2; break; default: seen = 3; }\n" +
		"  const later = () => seen + last;\n" +
		"  if (flag) { return total; }\n" +
		"  total = 5;\n" +
		"  return later() + total;\n" +
		"}\n"
	solved := 0
	probe := rule.Rule{Name: "dead-store-bitset", NeedsTypeChecker: true, Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{ast.KindSourceFile: func(node *ast.Node) {
			symbols, occupied := collectSymbolFacts(ctx, node)
			for _, root := range codePathRoots(ctx, node) {
				if !occupied[root.Node] {
					continue
				}
				graph := deadStoreGraph(ctx, root.Node, symbols)
				numbers := map[*ast.Symbol]int{}
				for _, block := range graph.Blocks {
					for index := range block.Events {
						event := &block.Events[index]
						if _, numbered := numbers[event.symbol]; !numbered {
							numbers[event.symbol] = len(numbers)
						}
						event.number = numbers[event.symbol]
					}
				}
				bitset := control_flow_graph.Solve[liveSymbols, deadStoreEvent](graph, control_flow_graph.Backward,
					deadStoreLiveness{reachable: map[int]bool{}, words: (len(numbers) + 63) / 64})
				reference := control_flow_graph.Solve[map[*ast.Symbol]bool, deadStoreEvent](graph,
					control_flow_graph.Backward, referenceLiveness{})
				if bitset.Rounds() != reference.Rounds() {
					t.Errorf("%v: the bitset converged in %d rounds, the map in %d", root.Node.Kind, bitset.Rounds(), reference.Rounds())
				}
				for _, block := range graph.Blocks {
					bits, ok := bitset.In(block)
					set, referenceOk := reference.In(block)
					if ok != referenceOk {
						t.Fatalf("block %d is reachable to one solution and not the other", block.Index())
					}
					for symbol, number := range numbers {
						if bits.has(number) != set[symbol] {
							t.Errorf("%v block %d: %s live %t in the bitset, %t in the map", root.Node.Kind, block.Index(),
								symbol.Name, bits.has(number), set[symbol])
						}
					}
				}
				solved++
			}
		}}
	}}
	rule_testing.RunTypedFiles(t, probe, map[string]string{"/fixture.ts": source}, "/fixture.ts")
	if solved < 2 {
		t.Fatalf("solved %d roots, want the function and its closure", solved)
	}
}
