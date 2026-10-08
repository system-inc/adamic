package lower

import (
	"context"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"testing"
)

func TestShapeDynamicArrayFrontier(t *testing.T) {
	for _, name := range []string{"proven-dynamic", "nonconforming-dynamic"} {
		t.Run(name, func(t *testing.T) {
			checked, err := load.Load([]string{"../../stage3/interface-downcasts/lane3/" + name + ".a"})
			if err != nil {
				t.Fatal(err)
			}
			program, err := Lower(context.Background(), checked)
			if err != nil {
				t.Fatal(err)
			}
			reads := 0
			inspect := func(node any) bool {
				if cast, ok := node.(ir.CheckedCast); ok {
					got := newAllocationFlowGraph(program).ReachingAllocations(cast.Value)
					if !got.Unknown {
						t.Fatal("uninitialized temporary frontier was lost")
					}
					t.Logf("cast reaching=%+v", got)
				}
				if index, ok := node.(ir.ArrayIndex); ok {
					got := newAllocationFlowGraph(program).ReachingAllocations(index)
					if got.Unknown || len(got.Sites) == 0 {
						t.Fatalf("finite dynamic index lost its selected allocations: %+v", got)
					}
					t.Logf("index reaching=%+v", got)
				}
				if property, ok := node.(ir.Property); ok && property.View != "" {
					reads++
				}
				return true
			}
			walk(program.Main, inspect)
			for _, function := range program.Functions {
				walk(function.Body, inspect)
			}
			t.Logf("checked reads=%d", reads)
			if reads == 0 {
				t.Fatal("uninitialized temporary frontier erased field checks")
			}
			if name == "nonconforming-dynamic" && reads == 0 {
				t.Fatal("finite dynamic keys erased the wrong selected field type")
			}
		})
	}
}
