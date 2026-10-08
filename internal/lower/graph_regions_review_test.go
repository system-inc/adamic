package lower

import (
	"context"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// Preserve the area cycle refusal even when a structural view hides the back-reference.
func TestGraphRegionStructuralCyclesAreRefused(t *testing.T) {
	for _, name := range []string{"structural_literal", "counted_container"} {
		t.Run(name, func(t *testing.T) {
			checked, err := load.Load([]string{"../oracle/testdata/graph_regions_" + name + ".a"})
			if err != nil {
				t.Fatal(err)
			}
			_, err = Lower(context.Background(), checked)
			requireCycleRefusal(t, err)
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

// Flow-derived graph metadata cannot bypass the ruled cycle refusal.
func TestGraphAllocationFlowCyclesAreRefused(t *testing.T) {
	for _, name := range []string{"return", "conditional", "mixed", "override"} {
		t.Run(name, func(t *testing.T) {
			checked, err := load.Load([]string{"../oracle/testdata/graph_regions/classification_" + name + ".a"})
			if err != nil {
				t.Fatal(err)
			}
			_, err = Lower(context.Background(), checked)
			requireCycleRefusal(t, err)
		})
	}
}
