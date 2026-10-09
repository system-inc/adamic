package flow

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

// The flow trace includes oracle fixtures. Its compiler and Node must use the same UTC policy.
func init() {
	if err := os.Setenv("TZ", "UTC"); err != nil {
		panic(err)
	}
}

// Unlike a general loop-carried range, this local Date has two straightforward writes.
// Node must observe both, and inference must assign them a range rather than decline.
func TestDateMutationRanges(t *testing.T) {
	t.Parallel()
	observed := traced(t, "../oracle/testdata/library_date_set.a")
	ranges := map[int]*MutableRanges{}
	values := map[int]map[InstructionId]map[DeclarationId]IdentifierId{}
	for function, graph := range observed.graphs {
		ranges[function] = InferMutableRanges(graph)
		values[function] = valuesBefore(graph)
	}
	mutations := 0
	for _, event := range observed.events {
		if !strings.HasPrefix(event, "mutated ") {
			continue
		}
		fields := strings.Fields(event)
		local, _ := strconv.Atoi(fields[1])
		index, _ := strconv.Atoi(fields[2])
		at := observed.marked[index]
		graph := observed.graphs[at.function]
		id, known := values[at.function][at.instruction][DeclarationId(local+1)]
		if !known {
			t.Fatalf("Node observed mutation of an untracked Date: %s", event)
		}
		interval := ranges[at.function].Get(id)
		order := graph.Instructions[at.instruction].Order
		if !interval.IsSet() || !interval.Contains(order) {
			t.Fatalf("Node observed Date mutation outside its inferred range: %s", event)
		}
		mutations++
	}
	if mutations < 2 {
		t.Fatalf("want both local Date writes observed on Node, got %d", mutations)
	}
	t.Logf("Node observed %d Date mutations, all inside their ranges", mutations)
}
