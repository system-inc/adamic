package lower

import (
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// The open-receiver cut keys every receiver a projected query opens, whatever it's rooted in (#r1ef873).
// It first keyed only Read-rooted receivers, then calls too, and each time tsc's whole program met a
// receiver outside the list that reopened under itself, uncut, for the first link's 55 minutes. So this
// pins the rule, not a list: for each kind of expression, one already open on the stack answers the cut's
// Unknown when it's asked again. Mutant: the frame keyed only where projectionValueKey names a root, as
// at 625d5643, and every kind outside Read, Call and CallClosure roots goes red.
func TestProjectionCutsEveryOpenReceiver(t *testing.T) {
	t.Parallel()
	read := ir.Read{Local: 3}
	receivers := []struct {
		name  string
		value ir.Expression
	}{
		{"property of a read", ir.Property{Object: read, Name: "left"}},
		{"call", ir.Call{Function: 1, Arguments: []ir.Expression{read}}},
		{"closure call", ir.CallClosure{Closure: read}},
		{"conditional", ir.Conditional{Condition: ir.BooleanConstant{}, WhenTrue: read, WhenNot: ir.Read{Local: 4}}},
		{"coalesce", ir.Coalesce{Value: read, Fallback: ir.Read{Local: 4}}},
		{"defined", ir.Defined{Value: read}},
		{"logical", ir.Logical{Left: read, Right: ir.Read{Local: 4}}},
		{"record call", ir.RecordCall{Method: "get", Arguments: []ir.Expression{read, ir.StringConstant{Index: 0}}}},
		{"object call", ir.ObjectCall{Method: "assign", Arguments: []ir.Expression{read}}},
		{"view element", ir.ViewElement{Array: read, Index: ir.Read{Local: 5}}},
		{"dynamic element", ir.ArrayIndex{Array: read, Index: ir.Read{Local: 5}}},
		{"property of a call", ir.Property{Object: ir.Call{Function: 2}, Name: "right"}},
		{"property of a conditional", ir.Property{Object: ir.Conditional{Condition: ir.BooleanConstant{}, WhenTrue: read, WhenNot: read}, Name: "first"}},
	}
	for _, receiver := range receivers {
		graph := &allocationFlowGraph{program: &ir.Program{}}
		_, open := graph.projectionFrame(receiver.value, 0)
		if open == "" {
			t.Errorf("%s: projectionFrame gave no value key", receiver.name)
			continue
		}
		graph.projectionOpen = map[string]bool{open: true}
		answer := graph.reachingAllocations(receiver.value, true, 0)
		if !answer.Unknown || !slices.Equal(answer.Reasons, []string{"recursive projection receiver not certified"}) {
			t.Errorf("%s: asked again while open, it answered %+v, not the open-receiver cut", receiver.name, answer)
		}
	}
}
