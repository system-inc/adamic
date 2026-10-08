package fresh

import (
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// These are complete boundary inventories, not certificates extracted from IR.
// The disconnected member remains owned after a graph edge was overwritten.
func TestStep37JoinInventory(t *testing.T) {
	for _, variant := range []string{"exclusive", "kept child", "kept disconnected", "weak observer", "incomplete"} {
		t.Run(variant, func(t *testing.T) {
			snapshot := OwnershipSnapshot{
				Complete: true,
				Objects:  []OwnershipObject{{ID: 1, Region: 7}, {ID: 2, Region: 7}, {ID: 3, Region: 7}, {ID: 4}},
				Edges:    []OwnershipEdge{{From: 1, To: 2, Label: ".child"}, {From: 2, To: 1, Label: ".parent"}},
				Roots:    []OwnershipRoot{{Name: "result", Object: 1}},
			}
			want := Proven
			switch variant {
			case "kept child", "kept disconnected":
				child := 2
				if variant == "kept disconnected" {
					child = 3
				}
				snapshot.Roots = append(snapshot.Roots, OwnershipRoot{Name: "worker.cache", Object: 4})
				snapshot.Edges = append(snapshot.Edges, OwnershipEdge{From: 4, To: child, Label: ".saved"})
				want = Refused
			case "weak observer":
				snapshot.Roots = append(snapshot.Roots, OwnershipRoot{Name: "worker.weak", Object: 2, Weak: true})
				want = Refused
			case "incomplete":
				snapshot.Complete = false
				want = Unknown
			}
			answer := QueryOwnership(snapshot, "result")
			if answer.Verdict != want || !reflect.DeepEqual(answer.Members, []int{1, 2, 3}) {
				t.Fatalf("worker reference or incomplete inventory crossed join: %+v", answer)
			}
			t.Logf("%s: members=%v evidence=%v", answer.Verdict, answer.Members, answer.Evidence)
		})
	}
}

// The real extractor deliberately cannot certify dynamic graph membership.
func TestStep37JoinExtractionIsIncomplete(t *testing.T) {
	program := &ir.Program{
		GraphTypes: map[int]bool{7: true},
		Locals:     []ir.Local{{Name: "items", Type: ir.Array, Function: -1}, {Name: "work", Type: ir.Closure, Function: -1}},
		Main: []ir.Statement{
			ir.Declare{Local: 0, Value: ir.ArrayLiteral{Element: ir.Object, Elements: []ir.Expression{ir.ObjectLiteral{GraphTypes: []int{7}}}}},
			ir.Evaluate{Value: ir.ParallelMap{Moved: true, Items: ir.Read{Local: 0, Of: ir.Array}, Work: ir.Read{Local: 1, Of: ir.Closure}, Result: ir.Object}},
		},
	}
	answers := OwnershipTransfers(program)
	if len(answers) != 1 || answers[0].Answer.Verdict != Unknown || answers[0].Answer.Snapshot.Complete {
		t.Fatalf("graph inventory invented by extraction: %+v", answers)
	}
	t.Logf("extracted graph boundary: %s %v", answers[0].Answer.Verdict, answers[0].Answer.Evidence)
}
