package lower

import (
	"context"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// An allocation's contextual view may hide the cycle. The whole-program proof
// must tag the allocation itself, while stores still dispatch on its header.
func TestGraphRegionStructuralAllocations(t *testing.T) {
	for _, name := range []string{"structural_literal", "counted_container"} {
		t.Run(name, func(t *testing.T) {
			checked, err := load.Load([]string{"../oracle/testdata/graph_regions_" + name + ".a"})
			if err != nil {
				t.Fatal(err)
			}
			program, err := Lower(context.Background(), checked)
			if err != nil {
				t.Fatal(err)
			}
			graph, counted := 0, 0
			for _, function := range program.Functions {
				walk(function.Body, func(node any) bool {
					literal, ok := node.(ir.ObjectLiteral)
					if !ok {
						return true
					}
					for _, field := range literal.Fields {
						switch field.Name {
						case "next":
							if !program.IsGraph(literal.GraphTypes) {
								t.Errorf("structural cycle allocation is counted: %v", literal.GraphTypes)
							}
							graph++
						case "held":
							if program.IsGraph(literal.GraphTypes) {
								t.Error("outside container was promoted into the graph")
							}
							counted++
						}
					}
					return true
				})
			}
			if graph == 0 || (name == "counted_container" && counted != 1) {
				t.Fatalf("missing graph/boundary allocations: %d/%d", graph, counted)
			}
		})
	}
}

func TestWeakRegionReviewStaysCounted(t *testing.T) {
	checked, err := load.Load([]string{"../oracle/testdata/weak_region_review.a"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.GraphTypes) != 0 {
		t.Fatal("Weak closed a strong cycle in graph classification")
	}
}

// Allocation-site flow IDs must select literals even when fresh checker IDs do not.
func TestGraphAllocationFlowClassification(t *testing.T) {
	for _, name := range []string{"return", "conditional", "mixed"} {
		t.Run(name, func(t *testing.T) {
			checked, err := load.Load([]string{"../oracle/testdata/graph_regions/classification_" + name + ".a"})
			if err != nil {
				t.Fatal(err)
			}
			program, err := Lower(context.Background(), checked)
			if err != nil {
				t.Fatal(err)
			}
			counted, graph := 0, 0
			for _, function := range program.Functions {
				walk(function.Body, func(node any) bool {
					literal, ok := node.(ir.ObjectLiteral)
					if !ok {
						return true
					}
					if program.IsGraph(literal.GraphTypes) {
						graph++
					} else {
						counted++
					}
					t.Logf("allocation IDs %v selected graph=%t", literal.GraphTypes, program.IsGraph(literal.GraphTypes))
					return true
				})
			}
			if len(program.GraphTypes) == 0 || counted != 0 || graph == 0 {
				t.Fatalf("want all flow allocations graph, got counted=%d graph=%d types=%v", counted, graph, program.GraphTypes)
			}
		})
	}
}
