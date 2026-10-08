package lower

import (
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestShapeMissingValueProducers(t *testing.T) {
	for _, test := range []struct {
		name    string
		program ir.Program
		query   ir.Expression
		reason  string
	}{
		{name: "local", program: ir.Program{Locals: []ir.Local{{}}, Main: []ir.Statement{
			ir.Declare{Local: 0}, ir.Assign{Local: 0, Value: ir.ObjectLiteral{GraphTypes: []int{-1}}},
		}}, query: ir.Read{Local: 0}, reason: "local 0 has a producer without a value"},
		{name: "result", program: ir.Program{Functions: []ir.Function{{Body: []ir.Statement{
			ir.Return{}, ir.Return{Value: ir.ObjectLiteral{GraphTypes: []int{-1}}},
		}}}}, query: ir.Call{Function: 0}, reason: "function 0 returns without a value"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := newAllocationFlowGraph(&test.program).ReachingAllocations(test.query)
			if !got.Unknown || !slices.Equal(got.Sites, []int{-1}) || !slices.Contains(got.Reasons, test.reason) {
				t.Fatalf("nonempty allocations must retain their missing-value frontier: %+v; want %q", got, test.reason)
			}
			if slices.Contains(got.Reasons, "missing expression") {
				t.Fatalf("producer context was lost: %+v", got)
			}
		})
	}
}
